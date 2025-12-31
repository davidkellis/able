import { AST, V11 } from "../index";
import { callCallableValue } from "../src/interpreter/functions";
import { numericToNumber } from "../src/interpreter/numeric";
import { mapSourceFile } from "../src/parser/tree-sitter-mapper";
import { getTreeSitterParser } from "../src/parser/tree-sitter-loader";

import type { TestCliConfig, TestReporterFormat } from "./test-cli";

type TestDescriptorData = {
  frameworkId: string;
  modulePath: string;
  testId: string;
  displayName: string;
  tags: string[];
  metadata: Array<{ key: string; value: string }>;
  location: { modulePath: string; line: number; column: number } | null;
};

type FailureData = {
  message: string;
  details: string | null;
  location: { modulePath: string; line: number; column: number } | null;
};

type TestEventData =
  | { kind: "case_started"; descriptor: TestDescriptorData }
  | { kind: "case_passed"; descriptor: TestDescriptorData; durationMs: number }
  | { kind: "case_failed"; descriptor: TestDescriptorData; durationMs: number; failure: FailureData }
  | { kind: "case_skipped"; descriptor: TestDescriptorData; reason: string | null }
  | { kind: "framework_error"; message: string };

export type TestEventState = {
  total: number;
  failed: number;
  skipped: number;
  frameworkErrors: number;
};

export type ReporterBundle = {
  reporter: V11.RuntimeValue;
  finish?: () => void;
};

const CLI_REPORTER_PACKAGE = "able_test_cli";
const CLI_REPORTER_SOURCE = `
package able_test_cli

import able.test.protocol.{Reporter, TestEvent}

struct CliReporter { emit_fn: TestEvent -> void }
struct CliCompositeReporter { inner: Reporter, emit_fn: TestEvent -> void }

fn CliReporter(emit_fn: TestEvent -> void) -> CliReporter {
  CliReporter { emit_fn }
}

fn CliCompositeReporter(inner: Reporter, emit_fn: TestEvent -> void) -> CliCompositeReporter {
  CliCompositeReporter { inner, emit_fn }
}

impl Reporter for CliReporter {
  fn emit(self: Self, event: TestEvent) -> void {
    self.emit_fn(event)
  }
}

impl Reporter for CliCompositeReporter {
  fn emit(self: Self, event: TestEvent) -> void {
    self.inner.emit(event)
    self.emit_fn(event)
  }
}
`;

export async function createTestReporter(
  interpreter: V11.Interpreter,
  format: TestReporterFormat,
  state: TestEventState,
): Promise<ReporterBundle | null> {
  const emitHandler = createEventHandler(interpreter, format, state);
  const emitFn = interpreter.makeNativeFunction("__able_test_cli_emit", 1, (_ctx, [event]) => {
    if (event) {
      emitHandler(event);
    }
    return { kind: "nil", value: null };
  });

  if (format === "json" || format === "tap") {
    const cli = await ensureCliReporterModule(interpreter);
    if (!cli) {
      return null;
    }
    if (format === "tap") {
      console.log("TAP version 13");
    }
    const reporterFn = getCallableFromPackage(cli, "CliReporter");
    if (!reporterFn) {
      console.error("able test: missing CliReporter helper");
      return null;
    }
    const reporter = callCallableValue(interpreter as any, reporterFn, [emitFn], interpreter.globals);
    return { reporter };
  }

  const innerReporter = createStdlibReporter(interpreter, format);
  if (!innerReporter) {
    return null;
  }

  const cli = await ensureCliReporterModule(interpreter);
  if (!cli) {
    return null;
  }
  const compositeFn = getCallableFromPackage(cli, "CliCompositeReporter");
  if (!compositeFn) {
    console.error("able test: missing CliCompositeReporter helper");
    return null;
  }

  const reporter = callCallableValue(interpreter as any, compositeFn, [innerReporter, emitFn], interpreter.globals);
  let finish: (() => void) | undefined;
  if (format === "progress") {
    finish = () => finishProgressReporter(interpreter, innerReporter);
  }
  return { reporter, finish };
}

export function emitTestPlanList(interpreter: V11.Interpreter, descriptors: V11.RuntimeValue, config: TestCliConfig): void {
  const items = decodeDescriptorArray(interpreter, descriptors);
  if (config.reporterFormat === "json") {
    console.log(JSON.stringify(items, null, 2));
    return;
  }
  if (items.length === 0) {
    console.log("able test: no tests found");
    return;
  }
  for (const item of items) {
    const tags = item.tags.length > 0 ? item.tags.join(",") : "-";
    const modulePath = item.modulePath || "-";
    const metadata = config.dryRun ? formatMetadata(item.metadata) : null;
    const parts = [
      item.frameworkId,
      modulePath,
      item.testId,
      item.displayName,
      `tags=${tags}`,
    ];
    if (metadata) {
      parts.push(`metadata=${metadata}`);
    }
    console.log(parts.join(" | "));
  }
}

function createStdlibReporter(
  interpreter: V11.Interpreter,
  format: "doc" | "progress",
): V11.RuntimeValue | null {
  const writeLine = interpreter.makeNativeFunction("__able_test_cli_line", 1, (_ctx, [line]) => {
    if (!line) return { kind: "nil", value: null };
    console.log(runtimeValueToString(interpreter, line));
    return { kind: "nil", value: null };
  });
  const reporterName = format === "doc" ? "DocReporter" : "ProgressReporter";
  const reporterFn = getCallableSymbol(interpreter, "able.test.reporters", reporterName);
  if (!reporterFn) {
    console.error(`able test: unable to find able.test.reporters.${reporterName}`);
    return null;
  }
  try {
    return callCallableValue(interpreter as any, reporterFn, [writeLine], interpreter.globals);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    return null;
  }
}

function finishProgressReporter(interpreter: V11.Interpreter, reporter: V11.RuntimeValue): void {
  const methods = interpreter.inherentMethods.get("ProgressReporter");
  const finish = methods?.get("finish");
  if (!finish) {
    return;
  }
  try {
    callCallableValue(interpreter as any, finish, [reporter], interpreter.globals);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
  }
}

function createEventHandler(
  interpreter: V11.Interpreter,
  format: TestReporterFormat,
  state: TestEventState,
): (event: V11.RuntimeValue) => void {
  let tapIndex = 0;
  return (eventValue: V11.RuntimeValue) => {
    let event: TestEventData | null;
    try {
      event = decodeTestEvent(interpreter, eventValue);
    } catch (error) {
      console.error(`able test: ${extractErrorMessage(error)}`);
      return;
    }
    if (!event) return;
    recordTestEvent(state, event);

    if (format === "json") {
      console.log(JSON.stringify(serializeEvent(event)));
      return;
    }
    if (format === "tap") {
      switch (event.kind) {
        case "case_passed":
          tapIndex += 1;
          console.log(`ok ${tapIndex} - ${event.descriptor.displayName}`);
          return;
        case "case_failed":
          tapIndex += 1;
          console.log(`not ok ${tapIndex} - ${event.descriptor.displayName}`);
          emitTapFailure(event.failure);
          return;
        case "case_skipped": {
          tapIndex += 1;
          const reason = event.reason ?? "skipped";
          console.log(`ok ${tapIndex} - ${event.descriptor.displayName} # SKIP ${reason}`);
          return;
        }
        case "framework_error":
          console.log(`Bail out! ${event.message}`);
          return;
        case "case_started":
          return;
        default:
          return;
      }
    }
  };
}

function emitTapFailure(failure: FailureData): void {
  const lines: string[] = [];
  lines.push("  ---");
  lines.push(`  message: ${sanitizeTapValue(failure.message)}`);
  if (failure.details) {
    lines.push(`  details: ${sanitizeTapValue(failure.details)}`);
  }
  if (failure.location) {
    lines.push(
      `  location: ${sanitizeTapValue(
        `${failure.location.modulePath}:${failure.location.line}:${failure.location.column}`,
      )}`,
    );
  }
  lines.push("  ...");
  for (const line of lines) {
    console.log(line);
  }
}

function sanitizeTapValue(value: string): string {
  return value.replace(/\r?\n/g, "\\n");
}

function recordTestEvent(state: TestEventState, event: TestEventData): void {
  switch (event.kind) {
    case "case_passed":
      state.total += 1;
      return;
    case "case_failed":
      state.total += 1;
      state.failed += 1;
      return;
    case "case_skipped":
      state.total += 1;
      state.skipped += 1;
      return;
    case "framework_error":
      state.frameworkErrors += 1;
      return;
    default:
      return;
  }
}

function serializeEvent(event: TestEventData): Record<string, unknown> {
  switch (event.kind) {
    case "case_started":
      return { event: "case_started", descriptor: event.descriptor };
    case "case_passed":
      return { event: "case_passed", descriptor: event.descriptor, duration_ms: event.durationMs };
    case "case_failed":
      return {
        event: "case_failed",
        descriptor: event.descriptor,
        duration_ms: event.durationMs,
        failure: event.failure,
      };
    case "case_skipped":
      return { event: "case_skipped", descriptor: event.descriptor, reason: event.reason };
    case "framework_error":
      return { event: "framework_error", message: event.message };
    default:
      return { event: "unknown" };
  }
}

async function ensureCliReporterModule(interpreter: V11.Interpreter): Promise<Map<string, V11.RuntimeValue> | null> {
  const existing = interpreter.packageRegistry.get(CLI_REPORTER_PACKAGE);
  if (existing) {
    return existing;
  }
  let moduleAst: AST.Module;
  try {
    moduleAst = await parseInlineModule(CLI_REPORTER_SOURCE, `<${CLI_REPORTER_PACKAGE}>`);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    return null;
  }
  try {
    interpreter.evaluate(moduleAst);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    return null;
  }
  return interpreter.packageRegistry.get(CLI_REPORTER_PACKAGE) ?? null;
}

async function parseInlineModule(source: string, origin: string): Promise<AST.Module> {
  const parser = await getTreeSitterParser();
  const tree = parser.parse(source);
  if (tree.rootNode.type !== "source_file") {
    throw new Error(`parser: unexpected root ${tree.rootNode.type}`);
  }
  const root = tree.rootNode as unknown as { hasError?: boolean };
  if (root.hasError) {
    throw new Error("parser: inline module has syntax errors");
  }
  const moduleAst = mapSourceFile(tree.rootNode, source, origin);
  if (!moduleAst) {
    throw new Error("parser: failed to map inline module");
  }
  return moduleAst;
}

function formatMetadata(entries: Array<{ key: string; value: string }>): string {
  if (entries.length === 0) {
    return "-";
  }
  return entries.map((entry) => `${entry.key}=${entry.value}`).join(",");
}

function decodeTestEvent(interpreter: V11.Interpreter, value: V11.RuntimeValue): TestEventData | null {
  if (value.kind !== "struct_instance") {
    return null;
  }
  const rawName = value.def.id.name ?? "";
  const tag = rawName.includes(".") ? rawName.split(".").pop() ?? rawName : rawName;
  switch (tag) {
    case "case_started":
      return {
        kind: "case_started",
        descriptor: decodeDescriptor(interpreter, structField(value, "descriptor")),
      };
    case "case_passed":
      return {
        kind: "case_passed",
        descriptor: decodeDescriptor(interpreter, structField(value, "descriptor")),
        durationMs: decodeNumber(structField(value, "duration_ms"), "duration_ms"),
      };
    case "case_failed":
      return {
        kind: "case_failed",
        descriptor: decodeDescriptor(interpreter, structField(value, "descriptor")),
        durationMs: decodeNumber(structField(value, "duration_ms"), "duration_ms"),
        failure: decodeFailure(interpreter, structField(value, "failure")),
      };
    case "case_skipped":
      return {
        kind: "case_skipped",
        descriptor: decodeDescriptor(interpreter, structField(value, "descriptor")),
        reason: decodeOptionalString(interpreter, structField(value, "reason")),
      };
    case "framework_error":
      return {
        kind: "framework_error",
        message: decodeString(interpreter, structField(value, "message")),
      };
    default:
      return null;
  }
}

function decodeDescriptorArray(
  interpreter: V11.Interpreter,
  descriptors: V11.RuntimeValue,
): TestDescriptorData[] {
  const items: TestDescriptorData[] = [];
  const arrayValue = coerceArrayValue(interpreter, descriptors, "descriptor array");
  for (const entry of arrayValue.elements) {
    items.push(decodeDescriptor(interpreter, entry));
  }
  return items;
}

function decodeDescriptor(interpreter: V11.Interpreter, value: V11.RuntimeValue): TestDescriptorData {
  if (!value || value.kind !== "struct_instance") {
    throw new Error("expected TestDescriptor struct");
  }
  return {
    frameworkId: decodeString(interpreter, structField(value, "framework_id")),
    modulePath: decodeString(interpreter, structField(value, "module_path")),
    testId: decodeString(interpreter, structField(value, "test_id")),
    displayName: decodeString(interpreter, structField(value, "display_name")),
    tags: decodeStringArray(interpreter, structField(value, "tags")),
    metadata: decodeMetadataArray(interpreter, structField(value, "metadata")),
    location: decodeLocation(interpreter, structField(value, "location")),
  };
}

function decodeFailure(interpreter: V11.Interpreter, value: V11.RuntimeValue): FailureData {
  if (!value || value.kind !== "struct_instance") {
    throw new Error("expected Failure struct");
  }
  return {
    message: decodeString(interpreter, structField(value, "message")),
    details: decodeOptionalString(interpreter, structField(value, "details")),
    location: decodeLocation(interpreter, structField(value, "location")),
  };
}

function decodeLocation(
  interpreter: V11.Interpreter,
  value: V11.RuntimeValue | undefined,
): { modulePath: string; line: number; column: number } | null {
  if (!value || value.kind === "nil") {
    return null;
  }
  if (value.kind !== "struct_instance") {
    return null;
  }
  return {
    modulePath: decodeString(interpreter, structField(value, "module_path")),
    line: decodeNumber(structField(value, "line"), "line"),
    column: decodeNumber(structField(value, "column"), "column"),
  };
}

function decodeMetadataArray(
  interpreter: V11.Interpreter,
  value: V11.RuntimeValue | undefined,
): Array<{ key: string; value: string }> {
  if (!value) {
    return [];
  }
  const arrayValue = value.kind === "nil" ? null : coerceArrayValue(interpreter, value, "metadata array");
  if (!arrayValue) return [];
  const entries: Array<{ key: string; value: string }> = [];
  for (const entry of arrayValue.elements) {
    if (!entry || entry.kind !== "struct_instance") continue;
    const key = decodeString(interpreter, structField(entry, "key"));
    const val = decodeString(interpreter, structField(entry, "value"));
    entries.push({ key, value: val });
  }
  return entries;
}

function decodeStringArray(
  interpreter: V11.Interpreter,
  value: V11.RuntimeValue | undefined,
): string[] {
  if (!value) return [];
  const arrayValue = value.kind === "nil" ? null : coerceArrayValue(interpreter, value, "string array");
  if (!arrayValue) return [];
  return arrayValue.elements.map((entry) => decodeString(interpreter, entry));
}

function decodeString(interpreter: V11.Interpreter, value: V11.RuntimeValue | undefined): string {
  if (!value) return "";
  if (value.kind === "String") return value.value;
  return runtimeValueToString(interpreter, value);
}

function decodeOptionalString(
  interpreter: V11.Interpreter,
  value: V11.RuntimeValue | undefined,
): string | null {
  if (!value || value.kind === "nil") {
    return null;
  }
  return decodeString(interpreter, value);
}

function decodeNumber(value: V11.RuntimeValue | undefined, label: string): number {
  if (!value) return 0;
  return numericToNumber(value, label, { requireSafeInteger: true });
}

function coerceArrayValue(
  interpreter: V11.Interpreter,
  value: V11.RuntimeValue,
  label: string,
): Extract<V11.RuntimeValue, { kind: "array" }> {
  if (value.kind === "array") {
    return value;
  }
  if (value.kind === "struct_instance" && value.def.id.name === "Array") {
    const handleValue = structField(value, "storage_handle");
    const handle = Math.trunc(numericToNumber(handleValue, "array handle", { requireSafeInteger: true }));
    if (handle) {
      const state = interpreter.arrayStates.get(handle);
      if (state) {
        return interpreter.makeArrayValue(state.values, state.capacity);
      }
    }
  }
  throw new Error(`expected ${label}`);
}

function structField(value: Extract<V11.RuntimeValue, { kind: "struct_instance" }>, field: string): V11.RuntimeValue {
  if (value.values instanceof Map) {
    return value.values.get(field) ?? { kind: "nil", value: null };
  }
  const fieldIndex = value.def.fields.findIndex((entry) => entry.name?.name === field);
  if (fieldIndex >= 0 && Array.isArray(value.values)) {
    return value.values[fieldIndex] ?? { kind: "nil", value: null };
  }
  return { kind: "nil", value: null };
}

function getCallableSymbol(
  interpreter: V11.Interpreter,
  packageName: string,
  name: string,
): V11.RuntimeValue | null {
  const pkg = interpreter.packageRegistry.get(packageName);
  if (!pkg) return null;
  return getCallableFromPackage(pkg, name);
}

function getCallableFromPackage(
  pkg: Map<string, V11.RuntimeValue>,
  name: string,
): V11.RuntimeValue | null {
  const value = pkg.get(name);
  if (!value) return null;
  if (
    value.kind === "function" ||
    value.kind === "function_overload" ||
    value.kind === "native_function" ||
    value.kind === "bound_method" ||
    value.kind === "native_bound_method" ||
    value.kind === "partial_function"
  ) {
    return value;
  }
  return null;
}

function runtimeValueToString(interpreter: V11.Interpreter, value: V11.RuntimeValue): string {
  try {
    return interpreter.valueToString(value);
  } catch {
    return String(value);
  }
}

function extractErrorMessage(err: unknown): string {
  if (!err) return "";
  if (typeof err === "string") return err;
  if (err instanceof Error) {
    const anyErr = err as any;
    if (anyErr.value && typeof anyErr.value === "object" && "message" in anyErr.value) {
      return String(anyErr.value.message);
    }
    return err.message;
  }
  if (typeof err === "object" && err) {
    const anyErr = err as any;
    if (typeof anyErr.message === "string") return anyErr.message;
  }
  return String(err);
}
