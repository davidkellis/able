import path from "node:path";

import { normalizeRepoRelativePath, repoRoot } from "./path-utils";
import type {
  DiagnosticSeverity,
  DiagnosticLocation,
  DiagnosticNote,
  PackageSummary,
  TypecheckerDiagnostic,
} from "../src/typechecker/diagnostics";
import type * as AST from "../src/ast";
import type { ParserDiagnostic } from "../src/parser/diagnostics";
import { RaiseSignal } from "../src/interpreter/signals";
import type { RuntimeValue } from "../src/interpreter/values";
import { getRuntimeDiagnosticContext } from "../src/interpreter/runtime_diagnostics";

const REPO_ROOT = repoRoot();

export type RuntimeDiagnostic = {
  severity: DiagnosticSeverity;
  message: string;
  code?: string;
  location?: DiagnosticLocation;
  notes?: DiagnosticNote[];
};

export function formatTypecheckerDiagnostic(
  diag: TypecheckerDiagnostic,
  options?: { packageName?: string; absolutePath?: boolean },
): string {
  const location = formatDiagnosticLocation(diag.location, { absolutePath: options?.absolutePath });
  const severityPrefix = diag.severity === "warning" ? "warning: " : "";
  let message = diag.message;
  if (options?.packageName && options.packageName !== "<anonymous>") {
    message = `${options.packageName}: ${message}`;
  }
  const base = location
    ? `${severityPrefix}typechecker: ${location} ${message}`
    : `${severityPrefix}typechecker: ${message}`;
  return formatDiagnosticWithNotes(base, diag.notes, { absolutePath: options?.absolutePath });
}

export function formatParserDiagnostic(
  diag: ParserDiagnostic,
  options?: { absolutePath?: boolean },
): string {
  const location = formatDiagnosticLocation(diag.location, { absolutePath: options?.absolutePath });
  const severityPrefix = diag.severity === "warning" ? "warning: " : "";
  let message = diag.message;
  if (message.startsWith("parser:")) {
    message = message.replace(/^parser:\s*/, "");
  }
  const base = location
    ? `${severityPrefix}parser: ${location} ${message}`
    : `${severityPrefix}parser: ${message}`;
  return formatDiagnosticWithNotes(base, diag.notes, { absolutePath: options?.absolutePath });
}

export function formatRuntimeDiagnostic(
  diag: RuntimeDiagnostic,
  options?: { absolutePath?: boolean },
): string {
  const location = formatDiagnosticLocation(diag.location, { absolutePath: options?.absolutePath });
  const severityPrefix = diag.severity === "warning" ? "warning: " : "";
  let message = diag.message;
  if (message.startsWith("runtime:")) {
    message = message.replace(/^runtime:\s*/, "");
  }
  const base = location
    ? `${severityPrefix}runtime: ${location} ${message}`
    : `${severityPrefix}runtime: ${message}`;
  return formatDiagnosticWithNotes(base, diag.notes, { absolutePath: options?.absolutePath });
}

export function buildRuntimeDiagnostic(error: unknown): RuntimeDiagnostic {
  const context = getRuntimeDiagnosticContext(error);
  const message = extractRuntimeMessage(error);
  let location = nodeToDiagnosticLocation(context?.node);
  if (!location && context?.callStack?.length) {
    for (let i = context.callStack.length - 1; i >= 0; i -= 1) {
      const candidate = nodeToDiagnosticLocation(context.callStack[i]?.node ?? undefined);
      if (candidate) {
        location = candidate;
        break;
      }
    }
  }
  const notes = buildRuntimeNotes(context, location);
  return {
    severity: "error",
    message,
    location,
    notes: notes.length > 0 ? notes : undefined,
  };
}

export function printPackageSummaries(summaries: Map<string, PackageSummary>): void {
  if (summaries.size === 0) {
    return;
  }
  const entries = [...summaries.values()].sort((a, b) => a.name.localeCompare(b.name));
  if (entries.length === 0) {
    return;
  }
  console.warn("---- package export summary ----");
  for (const summary of entries.map(augmentKernelSummary)) {
    const label =
      summary.visibility && summary.visibility === "private"
        ? `${summary.name} (private)`
        : summary.name;
    const structs = formatSummaryList(summary.structs);
    const interfaces = formatSummaryList(summary.interfaces);
    const functions = formatSummaryList(summary.functions);
    console.warn(
      `package ${label} exports: structs=${structs}; interfaces=${interfaces}; functions=${functions}; impls=${summary.implementations.length}; method sets=${summary.methodSets.length}`,
    );
  }
}

function formatSummaryList(record?: Record<string, unknown>): string {
  if (!record) {
    return "-";
  }
  const names = Object.keys(record).sort();
  if (names.length === 0) {
    return "-";
  }
  return names.join(", ");
}

function formatDiagnosticLocation(
  location: DiagnosticLocation | undefined,
  options?: { absolutePath?: boolean },
): string | null {
  if (!location) {
    return null;
  }
  const { path: filePath, line, column } = location;
  const normalizedPath = filePath ? normalizePath(filePath, options?.absolutePath ?? false) : null;
  if (normalizedPath && line && column) {
    return `${normalizedPath}:${line}:${column}`;
  }
  if (normalizedPath && line) {
    return `${normalizedPath}:${line}`;
  }
  if (normalizedPath) {
    return normalizedPath;
  }
  if (line && column) {
    return `line ${line}, column ${column}`;
  }
  if (line) {
    return `line ${line}`;
  }
  return null;
}

function normalizePath(filePath: string, absolutePath: boolean): string {
  const absolute = path.isAbsolute(filePath) ? filePath : path.resolve(REPO_ROOT, filePath);
  if (absolutePath) {
    return absolute.split(path.sep).join("/");
  }
  const relative = normalizeRepoRelativePath(absolute);
  return relative === "" ? path.basename(absolute) : relative;
}

function formatDiagnosticWithNotes(
  base: string,
  notes: DiagnosticNote[] | undefined,
  options?: { absolutePath?: boolean },
): string {
  const extra = formatDiagnosticNotes(notes, options);
  if (extra.length === 0) {
    return base;
  }
  return [base, ...extra].join("\n");
}

function formatDiagnosticNotes(
  notes: DiagnosticNote[] | undefined,
  options?: { absolutePath?: boolean },
): string[] {
  if (!notes || notes.length === 0) {
    return [];
  }
  return notes.map((note) => {
    const location = formatDiagnosticLocation(note.location, { absolutePath: options?.absolutePath });
    if (location) {
      return `note: ${location} ${note.message}`;
    }
    return `note: ${note.message}`;
  });
}

function buildRuntimeNotes(
  context: ReturnType<typeof getRuntimeDiagnosticContext>,
  primary: DiagnosticLocation | undefined,
): DiagnosticNote[] {
  if (!context?.callStack?.length) {
    return [];
  }
  const notes: DiagnosticNote[] = [];
  for (let i = context.callStack.length - 1; i >= 0 && notes.length < 8; i -= 1) {
    const frame = context.callStack[i];
    const location = nodeToDiagnosticLocation(frame?.node ?? undefined);
    if (!location || locationsEqual(location, primary)) {
      continue;
    }
    notes.push({ message: "called from here", location });
  }
  return notes;
}

function nodeToDiagnosticLocation(node?: AST.AstNode | null): DiagnosticLocation | undefined {
  if (!node) return undefined;
  const span = node.span;
  if (!span && !node.origin) {
    return undefined;
  }
  if (!span) {
    return node.origin ? { path: node.origin } : undefined;
  }
  return {
    path: node.origin,
    line: span.start.line,
    column: span.start.column,
    endLine: span.end.line,
    endColumn: span.end.column,
  };
}

function extractRuntimeMessage(error: unknown): string {
  const errorValue = extractRuntimeErrorValue(error);
  if (errorValue) {
    const message = runtimeErrorValueMessage(errorValue);
    if (message) {
      return message;
    }
  }
  if (!error) return "";
  if (typeof error === "string") return error;
  if (error instanceof Error) return error.message;
  if (typeof error === "object" && error) {
    const anyErr = error as Record<string, unknown>;
    if (typeof anyErr.message === "string") return anyErr.message;
  }
  return String(error);
}

function extractRuntimeErrorValue(error: unknown): RuntimeValue | null {
  if (error instanceof RaiseSignal) return error.value;
  if (typeof error === "object" && error) {
    const anyErr = error as Record<string, unknown>;
    if ("value" in anyErr) {
      const value = anyErr.value as RuntimeValue | undefined;
      if (value && typeof value === "object" && "kind" in value) {
        return value;
      }
    }
  }
  return null;
}

function runtimeErrorValueMessage(value: RuntimeValue): string | null {
  if (value.kind === "error") {
    return value.message;
  }
  if (value.kind === "interface_value" && value.interfaceName === "Error" && value.value.kind === "error") {
    return value.value.message;
  }
  return null;
}

function locationsEqual(
  left: DiagnosticLocation | undefined,
  right: DiagnosticLocation | undefined,
): boolean {
  if (!left || !right) {
    return false;
  }
  return (
    left.path === right.path &&
    left.line === right.line &&
    left.column === right.column
  );
}

function augmentKernelSummary(summary: PackageSummary): PackageSummary {
  if (summary.name !== "kernel.kernel") {
    return summary;
  }
  if (summary.functions && Object.keys(summary.functions).length > 0) {
    return summary;
  }
  const functions = Object.fromEntries(
    [
      "__able_String_from_builtin",
      "__able_String_to_builtin",
      "__able_array_capacity",
      "__able_array_clone",
      "__able_array_new",
      "__able_array_read",
      "__able_array_reserve",
      "__able_array_set_len",
      "__able_array_size",
      "__able_array_with_capacity",
      "__able_array_write",
      "__able_await_default",
      "__able_await_sleep_ms",
      "__able_channel_await_try_recv",
      "__able_channel_await_try_send",
      "__able_channel_close",
      "__able_channel_is_closed",
      "__able_channel_new",
      "__able_channel_receive",
      "__able_channel_send",
      "__able_channel_try_receive",
      "__able_channel_try_send",
      "__able_char_from_codepoint",
      "__able_char_to_codepoint",
      "__able_hasher_create",
      "__able_hasher_finish",
      "__able_hasher_write",
      "__able_mutex_await_lock",
      "__able_mutex_lock",
      "__able_mutex_new",
      "__able_mutex_unlock",
    ].map((name) => [name, {}]),
  );
  return { ...summary, functions };
}
