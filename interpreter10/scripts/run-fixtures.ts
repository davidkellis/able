import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, TypeChecker, V10 } from "../index";
import type {
  PackageSummary,
  TypecheckerDiagnostic,
} from "../src/typechecker/diagnostics";
import { mapSourceFile } from "../src/parser/tree-sitter-mapper";
import { getTreeSitterParser } from "../src/parser/tree-sitter-loader";
import { formatTypecheckerDiagnostic, printPackageSummaries } from "./typecheck-utils";
import { resolveTypecheckMode, type TypecheckMode } from "./typecheck-mode";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../fixtures/ast");
const TYPECHECK_MODE = resolveTypecheckMode(process.env.ABLE_TYPECHECK_FIXTURES);
const TYPECHECK_BASELINE_PATH = path.join(FIXTURE_ROOT, "typecheck-baseline.json");
const WRITE_TYPECHECK_BASELINE = process.argv.includes("--write-typecheck-baseline");

type Manifest = {
  description?: string;
  entry?: string;
  setup?: string[];
  skipTargets?: string[];
  expect?: {
    result?: { kind: string; value?: unknown };
    stdout?: string[];
    errors?: string[];
    typecheckDiagnostics?: string[];
  };
};

type FixtureResult = { name: string; description?: string };

async function main() {
  const fixtures = await collectFixtures(FIXTURE_ROOT);
  if (fixtures.length === 0) {
    console.log("No fixtures found.");
    return;
  }

  const results: FixtureResult[] = [];
  const typecheckBaseline = new Map<string, string[]>();

  for (const fixtureDir of fixtures) {
    const manifest = await readManifest(fixtureDir);
    if (manifest.skipTargets?.includes("ts")) {
      continue;
    }
    const relativeName = path.relative(FIXTURE_ROOT, fixtureDir).split(path.sep).join("/");
    const interpreter = new V10.InterpreterV10();
    ensurePrint(interpreter);
    installRuntimeStubs(interpreter);
    const stdout: string[] = [];
    let evaluationError: unknown;
    const entry = manifest.entry ?? "module.json";
    const moduleAst = await loadModuleFromFixture(fixtureDir, entry);
    const setupModules: AST.Module[] = [];
    if (manifest.setup) {
      for (const setupFile of manifest.setup) {
        const setupPath = path.join(fixtureDir, setupFile);
        const setupModule = await loadModuleFromPath(setupPath);
        setupModules.push(setupModule);
      }
    }

    const typecheckDiagnostics: TypecheckerDiagnostic[] = [];
    let packageSummaries = new Map<string, PackageSummary>();
    if (TYPECHECK_MODE !== "off") {
      const session = new TypeChecker.TypecheckerSession();
      for (const setupModule of setupModules) {
        const { diagnostics } = session.checkModule(setupModule);
        typecheckDiagnostics.push(...diagnostics);
      }
      const { diagnostics } = session.checkModule(moduleAst);
      typecheckDiagnostics.push(...diagnostics);
      packageSummaries = session.getPackageSummaries();
    }

    const formattedDiagnostics = maybeReportTypecheckDiagnostics(
      fixtureDir,
      TYPECHECK_MODE,
      manifest.expect?.typecheckDiagnostics ?? null,
      typecheckDiagnostics,
      packageSummaries,
    );
    if (TYPECHECK_MODE !== "off") {
      typecheckBaseline.set(relativeName, formattedDiagnostics);
    }

    let result: V10.V10Value | undefined;
    interceptStdout(stdout, () => {
      try {
        for (const setupModule of setupModules) {
          interpreter.evaluate(setupModule);
        }
        result = interpreter.evaluate(moduleAst);
      } catch (err) {
        evaluationError = err;
      }
    });

    assertExpectations(fixtureDir, manifest.expect, result, stdout, evaluationError, typecheckDiagnostics);
    results.push({ name: relativeName, description: manifest.description });
  }

  if (TYPECHECK_MODE !== "off") {
    const baselineObject = Object.fromEntries(
      [...typecheckBaseline.entries()].sort(([a], [b]) => a.localeCompare(b)),
    );
    if (WRITE_TYPECHECK_BASELINE) {
      await fs.writeFile(TYPECHECK_BASELINE_PATH, `${JSON.stringify(baselineObject, null, 2)}\n`, "utf8");
    } else {
      let existingBaseline: Record<string, string[]> | null = null;
      try {
        const raw = await fs.readFile(TYPECHECK_BASELINE_PATH, "utf8");
        existingBaseline = JSON.parse(raw) as Record<string, string[]>;
      } catch (err: any) {
        if (err.code === "ENOENT") {
          throw new Error(
            `typechecker baseline missing at ${TYPECHECK_BASELINE_PATH}. Run with --write-typecheck-baseline to generate it.`,
          );
        }
        throw err;
      }
      const differences = diffBaselineMaps(baselineObject, existingBaseline ?? {});
      if (differences.length > 0) {
        const message = [`Typecheck baseline mismatch:`].concat(differences).join("\n  ");
        throw new Error(message);
      }
    }
  } else if (WRITE_TYPECHECK_BASELINE) {
    console.warn("typechecker: skipping baseline write because ABLE_TYPECHECK_FIXTURES is off");
  }

  for (const res of results) {
    const desc = res.description ? ` - ${res.description}` : "";
    console.log(`✓ ${res.name}${desc}`);
  }
  console.log(`Executed ${results.length} fixture(s).`);
}

async function collectFixtures(root: string): Promise<string[]> {
  const dirs: string[] = [];
  async function walk(current: string) {
    const entries = await fs.readdir(current, { withFileTypes: true });
    let hasModule = false;
    for (const entry of entries) {
      if (entry.isFile() && entry.name.endsWith(".json") && entry.name === "module.json") {
        hasModule = true;
      }
    }
    if (hasModule) {
      dirs.push(current);
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        await walk(path.join(current, entry.name));
      }
    }
  }
  await walk(root);
  return dirs.sort();
}

async function readManifest(dir: string): Promise<Manifest> {
  const manifestPath = path.join(dir, "manifest.json");
  try {
    const contents = await fs.readFile(manifestPath, "utf8");
    return JSON.parse(contents);
  } catch (err: any) {
    if (err.code === "ENOENT") return {};
    throw err;
  }
}

async function readModule(filePath: string): Promise<AST.Module> {
  const raw = JSON.parse(await fs.readFile(filePath, "utf8"));
  const module = hydrateNode(raw) as AST.Module;
  annotateModuleOrigin(module, filePath);
  return module;
}

async function loadModuleFromFixture(dir: string, relativePath: string): Promise<AST.Module> {
  const absolute = path.join(dir, relativePath);
  return loadModuleFromPath(absolute);
}

async function loadModuleFromPath(filePath: string): Promise<AST.Module> {
  if (filePath.endsWith(".json")) {
    const directory = path.dirname(filePath);
    const base = path.basename(filePath, ".json");
    const candidates = [path.join(directory, `${base}.able`)];
    if (base === "module") {
      candidates.push(path.join(directory, "source.able"));
    }
    for (const candidate of candidates) {
      const fromSource = await parseModuleFromSource(candidate);
      if (fromSource) {
        return fromSource;
      }
    }
  }
  return readModule(filePath);
}

async function parseModuleFromSource(sourcePath: string): Promise<AST.Module | null> {
  if (!(await fileExists(sourcePath))) {
    return null;
  }
  try {
    const source = await fs.readFile(sourcePath, "utf8");
    const parser = await getTreeSitterParser();
    const tree = parser.parse(source);
    if (tree.rootNode.type !== "source_file") {
      throw new Error(`tree-sitter returned unexpected root ${tree.rootNode.type}`);
    }
    if ((tree.rootNode as unknown as { hasError?: boolean }).hasError) {
      throw new Error("tree-sitter reported syntax errors");
    }
    return mapSourceFile(tree.rootNode, source, sourcePath);
  } catch (error) {
    console.warn(`Failed to parse ${sourcePath} via tree-sitter; falling back to module.json.`, error);
    return null;
  }
}

async function fileExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

function hydrateNode(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(hydrateNode);
  if (value && typeof value === "object") {
    const node = value as Record<string, unknown>;
    if (typeof node.type === "string") {
      switch (node.type) {
        case "IntegerLiteral":
          if (typeof node.value === "string") node.value = BigInt(node.value);
          break;
        case "FloatLiteral":
          if (typeof node.value === "string") node.value = Number(node.value);
          break;
        case "BooleanLiteral":
          if (typeof node.value === "string") node.value = node.value === "true";
          break;
        case "ArrayLiteral":
          node.elements = hydrateNode(node.elements) as unknown[];
          break;
        case "Module":
          node.imports = hydrateNode(node.imports) as unknown[];
          node.body = hydrateNode(node.body) as unknown[];
          break;
        default:
          for (const [key, val] of Object.entries(node)) {
            node[key] = hydrateNode(val) as never;
          }
          return node;
      }
    }
    for (const [key, val] of Object.entries(node)) {
      if (key !== "type") node[key] = hydrateNode(val) as never;
    }
    return node;
  }
  return value;
}

function annotateModuleOrigin(node: unknown, origin: string, seen = new Set<object>()) {
  if (!node || typeof node !== "object") {
    return;
  }
  if (seen.has(node as object)) {
    return;
  }
  seen.add(node as object);

  if (Array.isArray(node)) {
    for (const entry of node) {
      annotateModuleOrigin(entry, origin, seen);
    }
    return;
  }

  const candidate = node as Partial<AST.AstNode>;
  if (typeof candidate.type === "string" && typeof candidate.origin !== "string") {
    candidate.origin = origin;
  }

  for (const value of Object.values(node)) {
    annotateModuleOrigin(value, origin, seen);
  }
}

function assertExpectations(
  dir: string,
  expect: Manifest["expect"],
  result: V10.V10Value | undefined,
  stdout: string[],
  evaluationError: unknown,
  _typecheckDiagnostics: TypecheckerDiagnostic[],
) {
  if (!expect) {
    if (evaluationError) {
      throw evaluationError;
    }
    return;
  }
  if (expect.errors && expect.errors.length > 0) {
    if (!evaluationError) {
      throw new Error(`Fixture ${dir} expected errors ${JSON.stringify(expect.errors)} but evaluation succeeded`);
    }
    const message = extractErrorMessage(evaluationError);
    if (!expect.errors.includes(message)) {
      throw new Error(`Fixture ${dir} expected error message in ${JSON.stringify(expect.errors)}, got '${message}'`);
    }
    return;
  }
  if (evaluationError) {
    throw evaluationError;
  }
  if (expect.stdout) {
    if (JSON.stringify(stdout) !== JSON.stringify(expect.stdout)) {
      throw new Error(`Fixture ${dir} expected stdout ${JSON.stringify(expect.stdout)}, got ${JSON.stringify(stdout)}`);
    }
  }
  if (expect.result) {
    const { kind, value } = expect.result;
    if (!result) {
      throw new Error(`Fixture ${dir} expected result kind ${kind}, but evaluation produced no value`);
    }
    if (result.kind !== kind) {
      throw new Error(`Fixture ${dir} expected result kind ${kind}, got ${result.kind}`);
    }
    if (value !== undefined) {
      switch (result.kind) {
        case "string":
        case "bool":
        case "char":
        case "i32":
        case "f64":
          if ((result as any).value !== value) {
            throw new Error(`Fixture ${dir} expected value ${value}, got ${(result as any).value}`);
          }
          break;
        default:
          // For now only support primitive comparisons.
          break;
      }
    }
  }
}

function interceptStdout(buffer: string[], fn: () => void) {
  const original = console.log;
  console.log = (...args: unknown[]) => {
    buffer.push(args.join(" "));
  };
  try {
    fn();
  } finally {
    console.log = original;
  }
}

function ensurePrint(interpreter: V10.InterpreterV10) {
  const globals = interpreter.globals ?? (interpreter as any).globals;
  if (!globals) return;
  try {
    globals.define("print", {
      kind: "native_function",
      name: "print",
      arity: 1,
      impl: (_interp: any, args: any[]) => {
        console.log(args.map(formatValue).join(" "));
        return { kind: "nil", value: null };
      },
    });
  } catch {
    // ignore redefinition
  }
}

function installRuntimeStubs(interpreter: V10.InterpreterV10) {
  const globals = interpreter.globals ?? (interpreter as any).globals;
  if (!globals) return;

  const defineStub = (
    name: string,
    arity: number,
    impl: (interp: V10.InterpreterV10, args: V10.V10Value[]) => V10.V10Value | null,
  ) => {
    try {
      globals.define(
        name,
        interpreter.makeNativeFunction(name, arity, (innerInterp, args) => {
          const result = impl(innerInterp as V10.InterpreterV10, args);
          return result ?? { kind: "nil", value: null };
        }),
      );
    } catch {
      // ignore redefinition attempts
    }
  };

  const hasGlobal = (name: string): boolean => {
    try {
      globals.get(name);
      return true;
    } catch {
      return false;
    }
  };

  let nextHandle = 1;
  const makeHandle = (): V10.V10Value => ({ kind: "i32", value: nextHandle++ });

  type ChannelState = {
    capacity: number;
    queue: V10.V10Value[];
    closed: boolean;
  };

  const channels = new Map<number, ChannelState>();
  type MutexState = {
    locked: boolean;
  };
  const mutexes = new Map<number, MutexState>();

  const toNumber = (value: V10.V10Value): number => {
    if (value.kind === "i32" || value.kind === "f64") return Number(value.value ?? 0);
    return Number((value as any).value ?? value ?? 0);
  };
  const toHandle = (value: V10.V10Value): number => toNumber(value);

  const checkCancelled = (interp: V10.InterpreterV10): boolean => {
    try {
      const cancelled = interp.procCancelled();
      return cancelled.kind === "bool" && cancelled.value;
    } catch {
      return false;
    }
  };

  const blockOnNilChannel = (interp: V10.InterpreterV10): V10.V10Value | null => {
    if (checkCancelled(interp)) {
      return { kind: "nil", value: null };
    }
    interp.procYield();
    return null;
  };

  if (!hasGlobal("__able_channel_new")) defineStub("__able_channel_new", 1, (_interp, [capacityArg]) => {
    const capacity = Math.max(0, Math.trunc(toNumber(capacityArg)));
    const handleValue = makeHandle();
    const handle = toHandle(handleValue);
    channels.set(handle, { capacity, queue: [], closed: false });
    return handleValue;
  });

  if (!hasGlobal("__able_channel_send")) defineStub("__able_channel_send", 2, (interp, [handleArg, value]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) {
      const blocked = blockOnNilChannel(interp);
      if (blocked) return blocked;
      return { kind: "nil", value: null };
    }
    if (channel.closed) {
      throw new Error("send on closed channel");
    }
    if (channel.capacity === 0) {
      channel.queue = [value];
    } else if (channel.queue.length < channel.capacity) {
      channel.queue.push(value);
    } else {
      // exceed capacity: overwrite most recent slot to keep fixture deterministic
      channel.queue[channel.queue.length - 1] = value;
    }
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_receive")) defineStub("__able_channel_receive", 1, (interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) {
      const blocked = blockOnNilChannel(interp);
      if (blocked) return blocked;
      return { kind: "nil", value: null };
    }
    if (channel.queue.length > 0) {
      return channel.queue.shift()!;
    }
    if (channel.closed) {
      return { kind: "nil", value: null };
    }
    const blocked = blockOnNilChannel(interp);
    if (blocked) return blocked;
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_try_send")) defineStub("__able_channel_try_send", 2, (_interp, [handleArg, value]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) return { kind: "bool", value: false };
    if (channel.closed) throw new Error("send on closed channel");
    if (channel.capacity === 0) {
      channel.queue = [value];
      return { kind: "bool", value: true };
    }
    if (channel.queue.length < channel.capacity) {
      channel.queue.push(value);
      return { kind: "bool", value: true };
    }
    return { kind: "bool", value: false };
  });

  if (!hasGlobal("__able_channel_try_receive")) defineStub("__able_channel_try_receive", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (!channel) return { kind: "nil", value: null };
    if (channel.queue.length > 0) {
      return channel.queue.shift()!;
    }
    return channel.closed ? { kind: "nil", value: null } : { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_close")) defineStub("__able_channel_close", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    if (channel) channel.closed = true;
    return { kind: "nil", value: null };
  });

  if (!hasGlobal("__able_channel_is_closed")) defineStub("__able_channel_is_closed", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    const channel = channels.get(handle);
    return { kind: "bool", value: channel ? channel.closed : false };
  });

  if (!hasGlobal("__able_mutex_new")) defineStub("__able_mutex_new", 0, () => makeHandle());
  if (!hasGlobal("__able_mutex_lock")) defineStub("__able_mutex_lock", 1, (interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    let state = mutexes.get(handle);
    if (!state) {
      state = { locked: false };
      mutexes.set(handle, state);
    }
    if (!state.locked) {
      state.locked = true;
      return { kind: "nil", value: null };
    }
    if (checkCancelled(interp)) {
      return { kind: "nil", value: null };
    }
    interp.procYield();
    return null;
  });
  if (!hasGlobal("__able_mutex_unlock")) defineStub("__able_mutex_unlock", 1, (_interp, [handleArg]) => {
    const handle = toHandle(handleArg);
    let state = mutexes.get(handle);
    if (!state) {
      state = { locked: false };
      mutexes.set(handle, state);
    }
    state.locked = false;
    return { kind: "nil", value: null };
  });
}

function maybeReportTypecheckDiagnostics(
  dir: string,
  mode: TypecheckMode,
  expected: string[] | null,
  actual: TypecheckerDiagnostic[],
  summaries: Map<string, PackageSummary>,
): string[] {
  if (mode === "off") {
    return [];
  }

  const formattedActual = actual.map(formatTypecheckerDiagnostic);

  if (expected && expected.length > 0) {
    if (formattedActual.length === 0) {
      if (mode === "strict") {
        throw new Error(`Fixture ${dir} expected typechecker diagnostics ${JSON.stringify(expected)} but none were produced`);
      }
      console.warn(
        `typechecker: fixture ${dir} expected diagnostics ${JSON.stringify(expected)} but checker returned none (mode=${mode})`,
      );
      return formattedActual;
    }
    const actualKeys = formattedActual.map(toDiagnosticKey);
    const expectedKeys = expected.map(toDiagnosticKey);
    const allExpectedPresent = expectedKeys.every(expectedKey =>
      actualKeys.some(actualKey => diagnosticKeyMatches(actualKey, expectedKey)),
    );
    if (!allExpectedPresent) {
      printPackageSummaries(summaries);
      const message = `Fixture ${dir} expected typechecker diagnostics ${JSON.stringify(expected)}, got ${JSON.stringify(formattedActual)}`;
      if (mode === "strict") {
        throw new Error(message);
      }
      console.warn(`typechecker: ${message}`);
      return formattedActual;
    }
    return formattedActual;
  }

  if (formattedActual.length === 0) {
    return formattedActual;
  }

  printPackageSummaries(summaries);

  for (const entry of formattedActual) {
    console.warn(entry);
  }

  if (mode === "strict") {
    throw new Error(`Fixture ${dir} produced typechecker diagnostics in strict mode`);
  }
  return formattedActual;
}


function formatValue(value: V10.V10Value): string {
  switch (value.kind) {
    case "string":
    case "char":
      return String(value.value);
    case "bool":
      return value.value ? "true" : "false";
    case "i32":
    case "f64":
      return String(value.value);
    default:
      return `[${value.kind}]`;
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

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});

function diffBaselineMaps(
  actual: Record<string, string[]>,
  expected: Record<string, string[]>,
): string[] {
  const differences: string[] = [];
  const allKeys = new Set([...Object.keys(actual), ...Object.keys(expected)]);
  for (const key of [...allKeys].sort()) {
    const actualValues = actual[key] ?? [];
    const expectedValues = expected[key] ?? [];
    if (actual[key] === undefined && expected[key] !== undefined && expectedValues.length > 0) {
      differences.push(`${key}: expected ${JSON.stringify(expectedValues)} but fixture was not executed`);
      continue;
    }
    if (actual[key] !== undefined && expected[key] === undefined && actualValues.length === 0) {
      continue;
    }
    const actualKeys = actualValues.map(toDiagnosticKey);
    const expectedKeys = expectedValues.map(toDiagnosticKey);
    const allExpectedPresent = expectedKeys.every(expectedKey =>
      actualKeys.some(actualKey => diagnosticKeyMatches(actualKey, expectedKey)),
    );
    if (!allExpectedPresent) {
      differences.push(
        `${key}: expected ${JSON.stringify(expectedValues)} but got ${JSON.stringify(actualValues)}`,
      );
    }
  }
  return differences;
}

function toDiagnosticKey(entry: string): string {
  const trimmed = entry.startsWith("typechecker: ") ? entry.slice("typechecker: ".length) : entry;
  const firstSpace = trimmed.indexOf(" ");
  if (firstSpace === -1) {
    return trimmed;
  }
  const location = trimmed.slice(0, firstSpace);
  let message = trimmed.slice(firstSpace + 1);
  if (!message.startsWith("typechecker:")) {
    message = `typechecker: ${message}`;
  }
  const segments = location.split(":");
  if (segments.length < 2) {
    return `${location}|${message}`;
  }
  let pathSegments = [...segments];
  let line = 0;
  const takeNumeric = () => {
    if (pathSegments.length === 0) return undefined;
    const candidate = pathSegments[pathSegments.length - 1];
    const parsed = Number.parseInt(candidate, 10);
    if (Number.isNaN(parsed)) {
      return undefined;
    }
    pathSegments = pathSegments.slice(0, -1);
    return parsed;
  };
  const columnMaybe = takeNumeric();
  const lineMaybe = takeNumeric();
  if (typeof lineMaybe === "number") {
    line = lineMaybe;
  } else if (typeof columnMaybe === "number") {
    line = columnMaybe;
  }
  const pathPart = pathSegments.join(":");
  const normalizedPath = pathPart === "typechecker" || pathPart === "" ? "" : pathPart;
  return `${normalizedPath}:${line}|${message}`;
}

function diagnosticKeyMatches(actual: string, expected: string): boolean {
  const [actualPrefix, actualMessage] = actual.split("|", 2);
  const [expectedPrefix, expectedMessage] = expected.split("|", 2);
  let normalizedExpectedMessage = expectedMessage;
  if (!expectedMessage.startsWith("typechecker:") && actualMessage.startsWith("typechecker:")) {
    normalizedExpectedMessage = `typechecker: ${expectedMessage}`;
  }
  if (actualMessage !== normalizedExpectedMessage) {
    return false;
  }
  const [actualPath, actualLine] = actualPrefix.split(":", 2);
  const [expectedPath, expectedLine] = expectedPrefix.split(":", 2);
  if (expectedPath && expectedPath !== actualPath) {
    return false;
  }
  if (expectedLine && expectedLine !== "" && expectedLine !== "0" && expectedLine !== actualLine) {
    return false;
  }
  return true;
}
