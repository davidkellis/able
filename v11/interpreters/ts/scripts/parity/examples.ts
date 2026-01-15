import { spawn } from "node:child_process";
import { promises as fs } from "node:fs";
import * as fsSync from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, V11 } from "../../index";
import { Environment } from "../../src/interpreter/environment";
import { TypecheckerSession } from "../../src/typechecker";
import type { DiagnosticLocation } from "../../src/typechecker/diagnostics";
import { ensurePrint, installRuntimeStubs, interceptStdout, formatRuntimeErrorMessage } from "../fixture-utils";
import { ModuleLoader, type Program } from "../module-loader";
import { collectModuleSearchPaths, type ModuleSearchPath } from "../module-search-paths";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../../../");
const TS_INTERPRETER_ROOT = path.resolve(REPO_ROOT, "interpreters/ts");
const GO_INTERPRETER_ROOT = path.resolve(REPO_ROOT, "interpreters/go");
export const DEFAULT_EXAMPLES_ROOT = path.resolve(TS_INTERPRETER_ROOT, "testdata/examples");
const DEFAULT_PARITY_DEPS_ROOT = path.resolve(TS_INTERPRETER_ROOT, "testdata/examples/deps/vendor");
const GO_WORKDIR = GO_INTERPRETER_ROOT;
const STDLIB_PATH = path.resolve(REPO_ROOT, "stdlib/src");
if (fsExists(DEFAULT_PARITY_DEPS_ROOT)) {
  ensureEnvPath("ABLE_MODULE_PATHS", DEFAULT_PARITY_DEPS_ROOT);
}
const ABLE_PATH_ENV = process.env.ABLE_PATH ?? "";
const ABLE_MODULE_PATHS_ENV = process.env.ABLE_MODULE_PATHS ?? "";
const MODULE_SEARCH_PATHS = computeModuleSearchPaths();
const GO_STD_ENV = fsExists(STDLIB_PATH) ? STDLIB_PATH : "";

export type CanonicalDiagnostic = {
  packageName: string;
  path: string;
  line: number;
  column: number;
  message: string;
};

export type TSExampleOutcome = {
  stdout: string[];
  diagnostics: string[];
  exitCode: number;
  error?: string;
};

export type GoExampleOutcome = {
  stdout: string[];
  diagnostics: string[];
  exitCode: number;
  error?: string;
};

export type ExampleParityDiff =
  | { kind: "exit-code"; message: string; expected: number; actual: number }
  | { kind: "stdout" | "diagnostics"; message: string; diff: ArrayDiff }
  | { kind: "error-presence"; message: string; expected?: string; actual?: string };

export type GoAbleRunner = {
  binaryPath: string;
  cleanup: () => Promise<void>;
};

export type ArrayDiff = {
  expected: string[];
  actual: string[];
  onlyInTS: string[];
  onlyInGo: string[];
};

export async function collectExamples(root: string = DEFAULT_EXAMPLES_ROOT): Promise<string[]> {
  const entries = await fs.readdir(root, { withFileTypes: true });
  const results: string[] = [];
  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const mainPath = path.join(root, entry.name, "main.able");
    if (await fileExists(mainPath)) {
      results.push(mainPath);
    }
  }
  results.sort((a, b) => a.localeCompare(b));
  return results;
}

export async function evaluateExampleTS(entryPath: string): Promise<TSExampleOutcome> {
  const loader = new ModuleLoader(MODULE_SEARCH_PATHS);
  const program = await loader.load(entryPath);

  const session = new TypecheckerSession();
  const diagnostics: CanonicalDiagnostic[] = [];
  for (const mod of program.modules) {
    const result = session.checkModule(mod.module);
    const packageName = result.summary?.name ?? mod.packageName ?? "";
    for (const diag of result.diagnostics) {
      const message = diag.severity === "warning" ? `warning: ${diag.message}` : diag.message;
      diagnostics.push(canonicalDiagnostic(packageName, diag.location, message));
    }
  }

  const canonicalDiagnostics = canonicalizeDiagnostics(diagnostics);
  if (canonicalDiagnostics.length > 0) {
    return {
      stdout: [],
      diagnostics: canonicalDiagnostics,
      exitCode: 1,
    };
  }

  const interpreter = new V11.Interpreter();
  ensurePrint(interpreter);
  installRuntimeStubs(interpreter);

  const stdout: string[] = [];
  let runtimeError: string | undefined;
  await interceptStdout(stdout, async () => {
    try {
      for (const mod of program.modules) {
        await interpreter.evaluateAsTask(mod.module);
      }
      await invokeEntryMain(interpreter, program.entry);
    } catch (err) {
      runtimeError = formatRuntimeErrorMessage(err);
    }
  });

  return {
    stdout,
    diagnostics: canonicalDiagnostics,
    exitCode: runtimeError ? 1 : 0,
    error: runtimeError,
  };
}

export async function evaluateExampleGo(entryPath: string, runner: GoAbleRunner): Promise<GoExampleOutcome> {
  const mergedModulePaths = [ABLE_MODULE_PATHS_ENV, GO_STD_ENV].filter(Boolean).join(path.delimiter);
  return new Promise((resolve, reject) => {
    const child = spawn(runner.binaryPath, ["run", entryPath], {
      cwd: GO_WORKDIR,
      env: {
        ...process.env,
        ABLE_MODULE_PATHS: mergedModulePaths,
      },
      stdio: ["ignore", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";
    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString();
    });
    child.on("error", reject);
    child.on("close", (code) => {
      const parsed = parseGoDiagnostics(stderr);
      resolve({
        stdout: splitLines(stdout),
        diagnostics: canonicalizeDiagnostics(parsed.diagnostics),
        exitCode: code ?? 0,
        error: parsed.errors.length > 0 ? parsed.errors.join("\n") : undefined,
      });
    });
  });
}

export function compareExampleOutcomes(example: string, ts: TSExampleOutcome, go: GoExampleOutcome): ExampleParityDiff | null {
  if (ts.exitCode !== go.exitCode) {
    return {
      kind: "exit-code",
      message: `exit code mismatch for ${example}`,
      expected: ts.exitCode,
      actual: go.exitCode,
    };
  }

  const stdoutDiff = diffArrays(ts.stdout, go.stdout);
  if (stdoutDiff) {
    return {
      kind: "stdout",
      message: `stdout mismatch for ${example}`,
      diff: stdoutDiff,
    };
  }

  const diagDiff = diffArrays(ts.diagnostics, go.diagnostics);
  if (diagDiff) {
    return {
      kind: "diagnostics",
      message: `diagnostics mismatch for ${example}`,
      diff: diagDiff,
    };
  }

  if (!!ts.error !== !!go.error) {
    return {
      kind: "error-presence",
      message: `error mismatch for ${example}`,
      expected: ts.error,
      actual: go.error,
    };
  }

  return null;
}

export async function buildGoAbleRunner(): Promise<GoAbleRunner> {
  const binDir = await fs.mkdtemp(path.join(os.tmpdir(), "able-go-cli-"));
  const binaryName = process.platform === "win32" ? "able.exe" : "able";
  const binaryPath = path.join(binDir, binaryName);

  const env = { ...process.env };
  const goCache = await fs.mkdtemp(path.join(os.tmpdir(), "able-go-build-cache-"));
  env.GOCACHE = goCache;

  await runGoBuild(binaryPath, env);
  await fs.rm(goCache, { recursive: true, force: true }).catch(() => {});

  return {
    binaryPath,
    cleanup: async () => {
      await fs.rm(binDir, { recursive: true, force: true });
    },
  };
}

async function runGoBuild(outputPath: string, env: NodeJS.ProcessEnv): Promise<void> {
  await new Promise<void>((resolve, reject) => {
    const child = spawn(
      "go",
      ["build", "-o", outputPath, "./cmd/able"],
      {
        cwd: GO_WORKDIR,
        env,
        stdio: ["ignore", "pipe", "pipe"],
      },
    );

    let stderr = "";
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString();
    });
    child.on("error", reject);
    child.on("close", (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`go build failed (${code}): ${stderr}`));
      }
    });
  });
}

export function diffArrays(expected: string[], actual: string[]): ArrayDiff | null {
  if (arraysEqual(expected, actual)) {
    return null;
  }
  const onlyInTS = expected.filter((value) => !actual.includes(value));
  const onlyInGo = actual.filter((value) => !expected.includes(value));
  return { expected, actual, onlyInTS, onlyInGo };
}

function arraysEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

export function formatExampleDiff(diff: ExampleParityDiff): string {
  switch (diff.kind) {
    case "exit-code":
      return `${diff.message}: ts=${diff.expected}, go=${diff.actual}`;
    case "error-presence":
      return `${diff.message}: ts=${diff.expected ?? "<none>"}, go=${diff.actual ?? "<none>"}`;
    case "stdout":
    case "diagnostics":
      return [
        diff.message,
        `  expected=${JSON.stringify(diff.diff.expected)}`,
        `  actual=${JSON.stringify(diff.diff.actual)}`,
        diff.diff.onlyInTS.length > 0 ? `  only-in-ts=${JSON.stringify(diff.diff.onlyInTS)}` : null,
        diff.diff.onlyInGo.length > 0 ? `  only-in-go=${JSON.stringify(diff.diff.onlyInGo)}` : null,
      ]
        .filter(Boolean)
        .join("\n");
    default:
      return diff.message;
  }
}

async function invokeEntryMain(interpreter: V11.Interpreter, entry: Program["entry"]): Promise<void> {
  const packageBucket = interpreter.packageRegistry.get(entry.packageName);
  if (!packageBucket) {
    throw new Error(`entry package '${entry.packageName}' is not available at runtime`);
  }
  const mainValue = packageBucket.get("main");
  if (!mainValue) {
    throw new Error("entry module does not define a main function");
  }
  const callEnv = new Environment(interpreter.globals);
  callEnv.define("main", mainValue);
  const callNode = AST.functionCall(AST.identifier("main"), []);
  await interpreter.evaluateAsTask(callNode, callEnv);
}

async function fileExists(candidate: string): Promise<boolean> {
  try {
    const stats = await fs.stat(candidate);
    return stats.isFile();
  } catch {
    return false;
  }
}

function computeModuleSearchPaths(): ModuleSearchPath[] {
  const extras: ModuleSearchPath[] = [];
  if (fsExists(DEFAULT_PARITY_DEPS_ROOT)) {
    extras.push({ path: DEFAULT_PARITY_DEPS_ROOT });
  }
  return collectModuleSearchPaths({
    cwd: process.cwd(),
    ablePathEnv: ABLE_PATH_ENV,
    ableModulePathsEnv: ABLE_MODULE_PATHS_ENV,
    extras,
    probeFrom: [process.cwd(), __dirname, path.dirname(process.execPath)],
  });
}

function fsExists(target: string): boolean {
  try {
    return fsSync.statSync(target).isDirectory();
  } catch {
    return false;
  }
}

function ensureEnvPath(name: string, extra: string): void {
  if (!extra) return;
  const resolved = path.resolve(extra);
  const current = process.env[name] ?? "";
  const parts = current
    .split(path.delimiter)
    .map((segment) => segment.trim())
    .filter((segment) => segment.length > 0);
  if (parts.some((segment) => path.resolve(segment) === resolved)) {
    return;
  }
  parts.push(resolved);
  process.env[name] = parts.join(path.delimiter);
}

function splitLines(output: string): string[] {
  return output
    .split(/\r?\n/)
    .map((line) => line.trimEnd())
    .filter((line) => line.length > 0);
}

type ParsedDiagnostics = {
  diagnostics: CanonicalDiagnostic[];
  errors: string[];
};

function parseGoDiagnostics(stderr: string): ParsedDiagnostics {
  const diagnostics: CanonicalDiagnostic[] = [];
  const errors: string[] = [];
  const lines = stderr
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.length > 0 && line !== "typecheck: ok" && !line.startsWith("---- package export summary ----"));

  for (const line of lines) {
    let trimmed = line;
    let severityPrefix = "";
    if (trimmed.startsWith("warning: ")) {
      severityPrefix = "warning: ";
      trimmed = trimmed.slice("warning: ".length);
    }
    if (!trimmed.startsWith("typechecker: ")) {
      errors.push(line);
      continue;
    }
    const rest = trimmed.slice("typechecker: ".length);
    const { message, location } = splitMessageAndLocation(rest);
    const pkgMatch = /^([^:]+): (typechecker:.*)$/.exec(message);
    const pkg = pkgMatch ? pkgMatch[1] : "";
    const finalMessage = severityPrefix + (pkgMatch ? pkgMatch[2] : message);
    diagnostics.push(canonicalDiagnostic(pkg, location, finalMessage));
  }
  return { diagnostics, errors };
}

function splitMessageAndLocation(rest: string): { message: string; location?: DiagnosticLocation } {
  const trimmed = rest.trim();
  if (trimmed === "") {
    return { message: "" };
  }
  if (trimmed.startsWith("line ")) {
    const match = /^line\s+\d+(?:,\s*column\s+\d+)?/i.exec(trimmed);
    if (match) {
      const location = parseLocation(match[0]);
      const message = trimmed.slice(match[0].length).trim();
      return { message, location };
    }
  }
  const firstSpace = trimmed.indexOf(" ");
  if (firstSpace === -1) {
    return { message: trimmed };
  }
  const locationRaw = trimmed.slice(0, firstSpace).trim();
  const location = parseLocation(locationRaw);
  if (location.path || location.line || location.column) {
    const message = trimmed.slice(firstSpace + 1).trim();
    return { message, location };
  }
  return { message: trimmed };
}

function parseLocation(raw: string): DiagnosticLocation {
  const cleaned = raw.trim();
  if (cleaned === "") return {};
  const lineMatch = /^line\s+(\d+)(?:,\s*column\s+(\d+))?$/i.exec(cleaned);
  if (lineMatch) {
    const line = Number.parseInt(lineMatch[1] ?? "0", 10);
    const column = Number.parseInt(lineMatch[2] ?? "0", 10);
    return {
      path: "",
      line: Number.isFinite(line) ? line : 0,
      column: Number.isFinite(column) ? column : 0,
    };
  }
  const segments = cleaned.split(":");
  const numbers: number[] = [];
  while (segments.length > 0) {
    const tail = segments[segments.length - 1];
    const value = Number.parseInt(tail, 10);
    if (Number.isFinite(value)) {
      numbers.unshift(value);
      segments.pop();
    } else {
      break;
    }
  }
  const filePath = segments.join(":");
  const [line, column] = numbers;
  return {
    path: normalizePath(filePath),
    line: line ?? 0,
    column: column ?? 0,
  };
}

function canonicalDiagnostic(
  packageName: string,
  location: DiagnosticLocation | undefined,
  message: string,
): CanonicalDiagnostic {
  return {
    packageName: packageName || "",
    path: location?.path ? normalizePath(location.path) : "",
    line: location?.line ?? 0,
    column: location?.column ?? 0,
    message: message.trim(),
  };
}

function canonicalizeDiagnostics(diags: CanonicalDiagnostic[]): string[] {
  const keys = diags.map((diag) =>
    [
      diag.packageName,
      diag.path,
      String(diag.line),
      String(diag.column),
      diag.message,
    ].join("|"),
  );
  keys.sort((a, b) => a.localeCompare(b));
  return keys;
}

function normalizePath(target: string): string {
  if (!target) return "";
  const absolute = path.isAbsolute(target) ? target : path.resolve(target);
  const relative = path.relative(REPO_ROOT, absolute);
  const final = relative.startsWith("..") ? absolute : relative;
  return final.split(path.sep).join("/");
}
