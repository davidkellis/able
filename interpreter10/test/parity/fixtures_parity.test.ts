import { describe, test, expect, afterAll } from "bun:test";
import { spawn } from "node:child_process";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { V10 } from "../../index";
import type { AST } from "../../index";
import {
  collectFixtures,
  readManifest,
  loadModuleFromFixture,
  loadModuleFromPath,
  ensurePrint,
  installRuntimeStubs,
  interceptStdout,
  extractErrorMessage,
  type Manifest,
} from "../../scripts/fixture-utils";
import { TypecheckerSession } from "../../src/typechecker";
import { formatTypecheckerDiagnostic } from "../../scripts/typecheck-utils";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../../fixtures/ast");
const MAX_FIXTURES = (() => {
  const raw = process.env.ABLE_PARITY_MAX_FIXTURES;
  if (!raw) return undefined;
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) return undefined;
  return parsed;
})();

if (!process.env.ABLE_TYPECHECK_FIXTURES) {
  process.env.ABLE_TYPECHECK_FIXTURES = "off";
}

type NormalizedValue = {
  kind: string;
  value?: string;
  bool?: boolean;
};

type TSOutcome = {
  result?: NormalizedValue;
  stdout: string[];
  error?: string;
  diagnostics?: string[];
};

type GoOutcome = {
  result?: NormalizedValue;
  stdout?: string[];
  error?: string;
  diagnostics?: string[];
  typecheckMode?: string;
  skipped?: boolean;
};

type GoFixtureRunner = {
  binaryPath: string;
  cleanup: () => Promise<void>;
};

const goFixtureRunnerPromise = buildGoFixtureRunner();

afterAll(async () => {
  const runner = await goFixtureRunnerPromise;
  await runner.cleanup().catch(() => {});
});

async function buildGoFixtureRunner(): Promise<GoFixtureRunner> {
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
  const cliPath = path.resolve(__dirname, "../../../interpreter10-go");
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

async function evaluateTS(dir: string, manifest: Manifest, entry: string): Promise<TSOutcome> {
  const interpreter = new V10.InterpreterV10();
  ensurePrint(interpreter);
  installRuntimeStubs(interpreter);

  const stdout: string[] = [];
  const entryModule = await loadModuleFromFixture(dir, entry);
  const setupModules = [];
  if (manifest.setup) {
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

function normalizeTSValue(value: V10.V10Value): NormalizedValue {
  switch (value.kind) {
    case "string":
    case "char":
      return { kind: value.kind, value: String(value.value) };
    case "bool":
      return { kind: "bool", bool: !!value.value };
    case "i32":
    case "f64":
      return { kind: value.kind, value: String(value.value) };
    case "nil":
      return { kind: "nil" };
    default:
      return { kind: value.kind };
  }
}

async function evaluateGo(dir: string, entry: string): Promise<GoOutcome> {
  const runner = await goFixtureRunnerPromise;
  return new Promise((resolve, reject) => {
    const cliPath = path.resolve(__dirname, "../../../interpreter10-go");
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

function shouldSkipFixture(manifest: Manifest | null): boolean {
  if (!manifest) return false;
  if (manifest.skipTargets?.includes("ts") || manifest.skipTargets?.includes("go")) {
    return true;
  }
  return false;
}

function compareOutcomes(ts: TSOutcome, go: GoOutcome, fixture: string, manifest: Manifest | null) {
  if (go.skipped) {
    throw new Error(`Go fixture runner reported skipped for ${fixture}`);
  }
  if (ts.error || go.error) {
    const expectedError = ts.error ?? "<none>";
    const actualError = go.error ?? "<none>";
    if (!actualError.includes(expectedError)) {
      throw new Error(formatDiff("runtime error", fixture, expectedError, actualError));
    }
    return;
  }
  if (ts.result || go.result) {
    if (!go.result || !ts.result) {
      throw new Error(formatDiff("result presence", fixture, JSON.stringify(ts.result), JSON.stringify(go.result)));
    }
    if (go.result.kind !== ts.result.kind) {
      throw new Error(formatDiff("result kind", fixture, ts.result.kind, go.result.kind));
    }
    if (go.result.bool !== undefined || ts.result.bool !== undefined) {
      if ((go.result.bool ?? false) !== (ts.result.bool ?? false)) {
        throw new Error(formatDiff("bool result", fixture, String(ts.result.bool ?? false), String(go.result.bool ?? false)));
      }
    } else {
      const expectedValue = ts.result.value ?? "";
      const actualValue = go.result.value ?? "";
      if (expectedValue !== actualValue) {
        throw new Error(formatDiff("result value", fixture, expectedValue, actualValue));
      }
    }
  }

  const stdoutDiff = diffArrays(ts.stdout, go.stdout ?? [], "stdout", fixture);
  if (stdoutDiff) throw new Error(stdoutDiff);

  const expectedDiagnostics = manifest?.expect?.typecheckDiagnostics ?? [];
  if (expectedDiagnostics.length > 0) {
    const diagDiff = diffArrays(ts.diagnostics ?? [], go.diagnostics ?? [], "diagnostics", fixture);
    if (diagDiff) throw new Error(diagDiff);
  }
}

function diffArrays(expected: string[], actual: string[], title: string, fixture: string): string | null {
  if (arraysEqual(expected, actual)) {
    return null;
  }
  const onlyInExpected = expected.filter((value) => !actual.includes(value));
  const onlyInActual = actual.filter((value) => !expected.includes(value));
  return [
    `Mismatch in ${title} for ${fixture}`,
    `  expected=${JSON.stringify(expected)}`,
    `  actual=${JSON.stringify(actual)}`,
    onlyInExpected.length > 0 ? `  only-in-ts=${JSON.stringify(onlyInExpected)}` : null,
    onlyInActual.length > 0 ? `  only-in-go=${JSON.stringify(onlyInActual)}` : null,
  ]
    .filter(Boolean)
    .join("\n");
}

function arraysEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

function formatDiff(kind: string, fixture: string, expected: string, actual: string): string {
  return [`Mismatch in ${kind} for ${fixture}`, `  expected=${expected}`, `  actual=${actual}`].join("\n");
}

describe("fixture parity", async () => {
  const fixtures = (await collectFixtures(FIXTURE_ROOT)).slice(0, MAX_FIXTURES);

  for (const fixtureDir of fixtures) {
    const manifest = await readManifest(fixtureDir);
    const relativeName = path.relative(FIXTURE_ROOT, fixtureDir).split(path.sep).join("/");

    if (shouldSkipFixture(manifest)) {
      test(relativeName, () => {
        // Fixture skipped for TS or Go; no parity check required.
      });
      continue;
    }

    const entry = manifest?.entry ?? "module.json";

    test(relativeName, async () => {
      const tsOutcome = await evaluateTS(fixtureDir, manifest, entry);
      const goOutcome = await evaluateGo(fixtureDir, entry);
      compareOutcomes(tsOutcome, goOutcome, relativeName, manifest ?? null);
    }, { timeout: 20000 });
  }
});
