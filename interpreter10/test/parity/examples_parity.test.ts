import { describe, test, afterAll } from "bun:test";
import { spawn } from "node:child_process";
import { promises as fs } from "node:fs";
import * as fsSync from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { V10 } from "../../index";
import { ensurePrint, installRuntimeStubs, interceptStdout } from "../../scripts/fixture-utils";
import { ModuleLoader, type Program } from "../../scripts/module-loader";
import { TypecheckerSession } from "../../src/typechecker";
import type { DiagnosticLocation } from "../../src/typechecker/diagnostics";
import { callCallableValue } from "../../src/interpreter/functions";

type CanonicalDiagnostic = {
  packageName: string;
  path: string;
  line: number;
  column: number;
  message: string;
};

type TSOutcome = {
  stdout: string[];
  diagnostics: CanonicalDiagnostic[];
  exitCode: number;
  error?: string;
};

type GoOutcome = {
  stdout: string[];
  diagnostics: CanonicalDiagnostic[];
  exitCode: number;
  error?: string;
};

type GoAbleRunner = {
  binaryPath: string;
  cleanup: () => Promise<void>;
};

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const EXAMPLES_ROOT = path.resolve(REPO_ROOT, "interpreter10/testdata/examples");
const GO_WORKDIR = path.resolve(REPO_ROOT, "interpreter10-go");
const STDLIB_PATH = path.resolve(REPO_ROOT, "stdlib/v10/src");
const MODULE_SEARCH_PATHS = computeModuleSearchPaths();
const GO_STD_ENV = fsExists(STDLIB_PATH) ? STDLIB_PATH : "";

const goAbleRunnerPromise = buildGoAbleRunner();

afterAll(async () => {
  const runner = await goAbleRunnerPromise;
  await runner.cleanup().catch(() => {});
});

describe("examples parity (testdata)", async () => {
  const examples = await collectExamples(EXAMPLES_ROOT);
  if (examples.length === 0) {
    test("no examples found", () => {
      throw new Error("expected at least one example under interpreter10/testdata/examples");
    });
    return;
  }

  for (const entryPath of examples) {
    const relative = path.relative(EXAMPLES_ROOT, entryPath).split(path.sep).join("/");
    test(relative, async () => {
      const tsOutcome = await evaluateExampleTS(entryPath);
      const goOutcome = await evaluateExampleGo(entryPath);
      compareOutcomes(relative, tsOutcome, goOutcome);
    }, { timeout: 20000 });
  }
});

async function collectExamples(root: string): Promise<string[]> {
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

async function evaluateExampleTS(entryPath: string): Promise<TSOutcome> {
  const loader = new ModuleLoader(MODULE_SEARCH_PATHS);
  const program = await loader.load(entryPath);

  const session = new TypecheckerSession();
  const diagnostics: CanonicalDiagnostic[] = [];
  for (const mod of program.modules) {
    const result = session.checkModule(mod.module);
    const packageName = result.summary?.name ?? mod.packageName ?? "";
    for (const diag of result.diagnostics) {
      diagnostics.push(canonicalDiagnostic(packageName, diag.location, diag.message));
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

  const interpreter = new V10.InterpreterV10();
  ensurePrint(interpreter);
  installRuntimeStubs(interpreter);

  const stdout: string[] = [];
  let runtimeError: string | undefined;
  interceptStdout(stdout, () => {
    try {
      for (const mod of program.modules) {
        interpreter.evaluate(mod.module);
      }
      invokeEntryMain(interpreter, program.entry);
    } catch (err) {
      runtimeError = err instanceof Error ? err.message : String(err);
    }
  });

  return {
    stdout,
    diagnostics: canonicalDiagnostics,
    exitCode: runtimeError ? 1 : 0,
    error: runtimeError,
  };
}

function invokeEntryMain(interpreter: V10.InterpreterV10, entry: Program["entry"]): void {
  const packageBucket = interpreter.packageRegistry.get(entry.packageName);
  if (!packageBucket) {
    throw new Error(`entry package '${entry.packageName}' is not available at runtime`);
  }
  const mainValue = packageBucket.get("main");
  if (!mainValue) {
    throw new Error("entry module does not define a main function");
  }
  callCallableValue(interpreter as unknown as V10.InterpreterV10, mainValue, [], interpreter.globals);
}

async function evaluateExampleGo(entryPath: string): Promise<GoOutcome> {
  const runner = await goAbleRunnerPromise;
  return new Promise((resolve, reject) => {
    const child = spawn(runner.binaryPath, ["run", entryPath], {
      cwd: GO_WORKDIR,
      env: {
        ...process.env,
        ABLE_STD_LIB: GO_STD_ENV,
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

function compareOutcomes(example: string, ts: TSOutcome, go: GoOutcome): void {
  if (ts.exitCode !== go.exitCode) {
    throw new Error(`exit code mismatch for ${example}: ts=${ts.exitCode}, go=${go.exitCode}`);
  }

  const stdoutDiff = diffArrays(ts.stdout, go.stdout);
  if (stdoutDiff) {
    throw new Error(`stdout mismatch for ${example}:\n${stdoutDiff}`);
  }

  const diagDiff = diffArrays(ts.diagnostics, go.diagnostics);
  if (diagDiff) {
    throw new Error(`diagnostics mismatch for ${example}:\n${diagDiff}`);
  }

  if (!!ts.error !== !!go.error) {
    throw new Error(`error mismatch for ${example}: ts=${ts.error ?? "<none>"}, go=${go.error ?? "<none>"}`);
  }
}

function diffArrays(tsValues: string[], goValues: string[]): string | null {
  if (arraysEqual(tsValues, goValues)) {
    return null;
  }
  const tsOnly = tsValues.filter((value) => !goValues.includes(value));
  const goOnly = goValues.filter((value) => !tsValues.includes(value));
  const parts = [
    `  ts=${JSON.stringify(tsValues)}`,
    `  go=${JSON.stringify(goValues)}`,
    tsOnly.length > 0 ? `  only-in-ts=${JSON.stringify(tsOnly)}` : null,
    goOnly.length > 0 ? `  only-in-go=${JSON.stringify(goOnly)}` : null,
  ].filter(Boolean);
  return parts.join("\n");
}

function arraysEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

function splitLines(output: string): string[] {
  return output
    .split(/\r?\n/)
    .map((line) => line.trimEnd())
    .filter((line) => line.length > 0);
}

function parseGoDiagnostics(stderr: string): { diagnostics: CanonicalDiagnostic[]; errors: string[] } {
  const diagnostics: CanonicalDiagnostic[] = [];
  const errors: string[] = [];
  const lines = stderr
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.length > 0 && line !== "typecheck: ok" && !line.startsWith("---- package export summary ----"));

  for (const line of lines) {
    const match = /^([^:]+): typechecker: (.+)$/.exec(line);
    if (!match) {
      errors.push(line);
      continue;
    }
    const [, pkg, rest] = match;
    const { message, location } = splitMessageAndLocation(rest);
    diagnostics.push(canonicalDiagnostic(pkg, location, message));
  }
  return { diagnostics, errors };
}

function splitMessageAndLocation(rest: string): { message: string; location?: DiagnosticLocation } {
  const trimmed = rest.trim();
  const parenIndex = trimmed.lastIndexOf(" (");
  if (parenIndex === -1 || !trimmed.endsWith(")")) {
    return { message: trimmed };
  }
  const message = trimmed.slice(0, parenIndex).trim();
  const locationRaw = trimmed.slice(parenIndex + 2, trimmed.length - 1);
  return { message, location: parseLocation(locationRaw) };
}

function parseLocation(raw: string): DiagnosticLocation {
  const cleaned = raw.trim();
  if (cleaned === "") return {};
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
    path: filePath,
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

async function fileExists(candidate: string): Promise<boolean> {
  try {
    const stats = await fs.stat(candidate);
    return stats.isFile();
  } catch {
    return false;
  }
}

function computeModuleSearchPaths(): string[] {
  const paths = new Set<string>();
  const rawEnv = process.env.ABLE_MODULE_PATHS ?? "";
  for (const part of rawEnv.split(path.delimiter)) {
    const trimmed = part.trim();
    if (trimmed) {
      paths.add(path.resolve(trimmed));
    }
  }
  if (fsExists(STDLIB_PATH)) {
    paths.add(STDLIB_PATH);
  }
  return [...paths];
}

function fsExists(target: string): boolean {
  try {
    return fsSync.statSync(target).isDirectory();
  } catch {
    return false;
  }
}

async function buildGoAbleRunner(): Promise<GoAbleRunner> {
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
