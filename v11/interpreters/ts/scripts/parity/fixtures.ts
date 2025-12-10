import { spawn } from "node:child_process";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { V10 } from "../../index";
import type { AST } from "../../index";
import {
  ensurePrint,
  installRuntimeStubs,
  interceptStdout,
  loadModuleFromFixture,
  loadModuleFromPath,
  readManifest,
  extractErrorMessage,
  type Manifest,
} from "../fixture-utils";
import { TypecheckerSession } from "../../src/typechecker";
import { formatTypecheckerDiagnostic } from "../typecheck-utils";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(fileURLToPath(new URL("../../../../..", import.meta.url)));
const NORMALIZED_REPO_ROOT = REPO_ROOT.replace(/\\/g, "/");
export const DEFAULT_FIXTURE_ROOT = path.resolve(REPO_ROOT, "v11/fixtures/ast");

export type NormalizedValue = {
  kind: string;
  value?: string;
  bool?: boolean;
};

export type TSOutcome = {
  result?: NormalizedValue;
  stdout: string[];
  error?: string;
  diagnostics?: string[];
};

export type GoOutcome = {
  result?: NormalizedValue;
  stdout?: string[];
  error?: string;
  diagnostics?: string[];
  typecheckMode?: string;
  skipped?: boolean;
};

export type GoFixtureRunner = {
  binaryPath: string;
  cleanup: () => Promise<void>;
};

export type ArrayDiff = {
  expected: string[];
  actual: string[];
  onlyInTS: string[];
  onlyInGo: string[];
};

export type FixtureParityDiff =
  | { kind: "go-skipped"; message: string }
  | { kind: "runtime-error"; message: string; expectedError: string; actualError: string }
  | { kind: "result-presence"; message: string; expectedHasResult: boolean; actualHasResult: boolean }
  | { kind: "result-kind"; message: string; expectedKind: string; actualKind: string }
  | { kind: "result-bool"; message: string; expected?: boolean; actual?: boolean }
  | { kind: "result-value"; message: string; expected?: string; actual?: string }
  | { kind: "stdout" | "diagnostics"; message: string; diff: ArrayDiff };

const MAX_FIXTURES = resolveMaxFixtures(process.env.ABLE_PARITY_MAX_FIXTURES);

export function resolveMaxFixtures(raw: string | undefined): number | undefined {
  if (!raw) return undefined;
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) return undefined;
  return parsed;
}

export function getMaxFixturesOverride(cliValue: number | undefined): number | undefined {
  if (typeof cliValue === "number") {
    return Number.isFinite(cliValue) && cliValue > 0 ? Math.floor(cliValue) : undefined;
  }
  return MAX_FIXTURES;
}

export async function buildGoFixtureRunner(): Promise<GoFixtureRunner> {
  const runnerDir = await fs.mkdtemp(path.join(os.tmpdir(), "able-go-fixture-bin-"));
  const binaryName = process.platform === "win32" ? "fixture.exe" : "fixture";
  const binaryPath = path.join(runnerDir, binaryName);

  const buildEnv = { ...process.env };
  const goCache = await fs.mkdtemp(path.join(os.tmpdir(), "able-go-build-cache-"));
  buildEnv.GOCACHE = goCache;

  await runGoBuild(binaryPath, buildEnv);
  await fs.rm(goCache, { recursive: true, force: true }).catch(() => {});

  return {
    binaryPath,
    cleanup: async () => {
      await fs.rm(runnerDir, { recursive: true, force: true });
    },
  };
}

async function runGoBuild(outputPath: string, env: NodeJS.ProcessEnv): Promise<void> {
  const cliPath = path.resolve(REPO_ROOT, "v11/interpreters/go");
  await new Promise<void>((resolve, reject) => {
    const child = spawn(
      "go",
      ["build", "-o", outputPath, "./cmd/fixture"],
      {
        cwd: cliPath,
        env,
        stdio: ["ignore", "pipe", "pipe"],
      },
    );

    let stderr = "";
    let stdout = "";
    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString();
    });
    child.on("error", reject);
    child.on("close", (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`go build failed (${code}): ${stderr || stdout}`));
      }
    });
  });
}

export function shouldSkipFixture(manifest: Manifest | null): boolean {
  if (!manifest) return false;
  return Boolean(manifest.skipTargets?.includes("ts") || manifest.skipTargets?.includes("go"));
}

export async function evaluateFixtureTS(dir: string, manifest: Manifest | null, entry: string): Promise<TSOutcome> {
  const interpreter = new V10.InterpreterV10();
  ensurePrint(interpreter);
  installRuntimeStubs(interpreter);

  const stdout: string[] = [];
  const entryModule = await loadModuleFromFixture(dir, entry);
  const setupModules: AST.Module[] = [];
  if (manifest?.setup) {
    for (const setupFile of manifest.setup) {
      const setupModule = await loadModuleFromPath(path.join(dir, setupFile));
      setupModules.push(setupModule);
    }
  }

  let evaluationError: unknown;
  let result: NormalizedValue | undefined;
  const diagnostics: string[] = [];
  const typecheckMode = (process.env.ABLE_TYPECHECK_FIXTURES ?? "off").toLowerCase();

  if (typecheckMode !== "off") {
    const session = new TypecheckerSession();
    const collectDiagnostics = (module: AST.Module) => {
      const { diagnostics: moduleDiagnostics } = session.checkModule(module);
      for (const diag of moduleDiagnostics) {
        diagnostics.push(formatTypecheckerDiagnostic(diag));
      }
    };
    for (const module of setupModules) {
      collectDiagnostics(module);
    }
    collectDiagnostics(entryModule);
  }

  interceptStdout(stdout, () => {
    try {
      for (const module of setupModules) {
        interpreter.evaluate(module);
      }
      const value = interpreter.evaluate(entryModule);
      if (value) {
        result = normalizeTSValue(value);
      }
    } catch (err) {
      evaluationError = err;
    }
  });

  const outcome: TSOutcome = { stdout };
  if (evaluationError) {
    outcome.error = extractErrorMessage(evaluationError);
  }
  if (result) {
    outcome.result = result;
  }
  if (diagnostics.length > 0) {
    outcome.diagnostics = diagnostics;
  }
  return outcome;
}

export async function evaluateFixtureGo(
  runner: GoFixtureRunner,
  dir: string,
  entry: string,
): Promise<GoOutcome> {
  return new Promise((resolve, reject) => {
    const cliPath = path.resolve(REPO_ROOT, "v11/interpreters/go");
    const child = spawn(runner.binaryPath, ["--dir", dir, "--entry", entry, "--executor", "serial"], {
      cwd: cliPath,
      env: {
        ...process.env,
        ABLE_TYPECHECK_FIXTURES: process.env.ABLE_TYPECHECK_FIXTURES ?? "off",
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
      if (code !== 0) {
        return reject(new Error(`fixture runner exited with code ${code}: ${stderr}`));
      }
      try {
        const parsed = JSON.parse(stdout) as GoOutcome;
        resolve(parsed);
      } catch (err) {
        reject(new Error(`failed to parse go output: ${err}\nstdout:\n${stdout}\nstderr:\n${stderr}`));
      }
    });
  });
}

export function compareFixtureOutcomes(
  ts: TSOutcome,
  go: GoOutcome,
  fixture: string,
  _manifest: Manifest | null,
): FixtureParityDiff | null {
  if (go.skipped) {
    return { kind: "go-skipped", message: `Go fixture runner reported skipped for ${fixture}` };
  }
  if (ts.error || go.error) {
    const expectedError = ts.error ?? "<none>";
    const actualError = go.error ?? "<none>";
    if (!actualError.includes(expectedError)) {
      return {
        kind: "runtime-error",
        message: `Mismatch in runtime error for ${fixture}`,
        expectedError,
        actualError,
      };
    }
    return null;
  }
  if (ts.result || go.result) {
    if (!go.result || !ts.result) {
      return {
        kind: "result-presence",
        message: `Mismatch in result presence for ${fixture}`,
        expectedHasResult: Boolean(ts.result),
        actualHasResult: Boolean(go.result),
      };
    }
    if (go.result.kind !== ts.result.kind) {
      return {
        kind: "result-kind",
        message: `Mismatch in result kind for ${fixture}`,
        expectedKind: ts.result.kind,
        actualKind: go.result.kind,
      };
    }
    if (go.result.bool !== undefined || ts.result.bool !== undefined) {
      if ((go.result.bool ?? false) !== (ts.result.bool ?? false)) {
        return {
          kind: "result-bool",
          message: `Mismatch in boolean result for ${fixture}`,
          expected: ts.result.bool,
          actual: go.result.bool,
        };
      }
    } else {
      const expectedValue = ts.result.value ?? "";
      const actualValue = go.result.value ?? "";
      if (expectedValue !== actualValue) {
        return {
          kind: "result-value",
          message: `Mismatch in result value for ${fixture}`,
          expected: expectedValue,
          actual: actualValue,
        };
      }
    }
  }

  const stdoutDiff = diffArrays(ts.stdout, go.stdout ?? []);
  if (stdoutDiff) {
    return {
      kind: "stdout",
      message: `Mismatch in stdout for ${fixture}`,
      diff: stdoutDiff,
    };
  }

  const canonicalTsDiagnostics = canonicalizeDiagnostics(ts.diagnostics);
  const canonicalGoDiagnostics = canonicalizeDiagnostics(go.diagnostics);
  const diagDiff = diffArrays(canonicalTsDiagnostics ?? [], canonicalGoDiagnostics ?? []);
  if (diagDiff) {
    return {
      kind: "diagnostics",
      message: `Mismatch in diagnostics for ${fixture}`,
      diff: diagDiff,
    };
  }

  return null;
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

export function formatFixtureDiff(diff: FixtureParityDiff): string {
  switch (diff.kind) {
    case "go-skipped":
      return diff.message;
    case "runtime-error":
      return [diff.message, `  expected=${diff.expectedError}`, `  actual=${diff.actualError}`].join("\n");
    case "result-presence":
      return [
        diff.message,
        `  expected=${String(diff.expectedHasResult)}`,
        `  actual=${String(diff.actualHasResult)}`,
      ].join("\n");
    case "result-kind":
      return [diff.message, `  expected=${diff.expectedKind}`, `  actual=${diff.actualKind}`].join("\n");
    case "result-bool":
      return [diff.message, `  expected=${String(diff.expected)}`, `  actual=${String(diff.actual)}`].join("\n");
    case "result-value":
      return [diff.message, `  expected=${diff.expected ?? ""}`, `  actual=${diff.actual ?? ""}`].join("\n");
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

function canonicalizeDiagnostics(entries?: string[]): string[] | undefined {
  if (!entries || entries.length === 0) {
    return entries;
  }
  return entries.map((entry) => {
    let normalized = entry.replace(/\\/g, "/");
    if (normalized.includes(NORMALIZED_REPO_ROOT)) {
      const rootPattern = new RegExp(escapeRegExp(`${NORMALIZED_REPO_ROOT}/`), "g");
      normalized = normalized.replace(rootPattern, "");
    }
    normalized = normalized.replace(/(\.\.\/)+fixtures\//g, "fixtures/");
    return normalized;
  });
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

const INTEGER_VALUE_KINDS = new Set<string>([
  "i8",
  "i16",
  "i32",
  "i64",
  "i128",
  "u8",
  "u16",
  "u32",
  "u64",
  "u128",
]);
const FLOAT_VALUE_KINDS = new Set<string>(["f32", "f64"]);

function normalizeTSValue(value: V10.V10Value): NormalizedValue {
  switch (value.kind) {
    case "String":
    case "char":
      return { kind: value.kind, value: String(value.value) };
    case "bool":
      return { kind: "bool", bool: !!value.value };
    case "nil":
      return { kind: "nil" };
    default:
      if (INTEGER_VALUE_KINDS.has(value.kind) || FLOAT_VALUE_KINDS.has(value.kind)) {
        return { kind: value.kind, value: value.value !== undefined ? String(value.value) : undefined };
      }
      return { kind: value.kind };
  }
}

export { readManifest };
