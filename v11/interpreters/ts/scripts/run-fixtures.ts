import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, TypeChecker, V11 } from "../index";
import { Environment } from "../src/interpreter/environment";
import type {
  PackageSummary,
  TypecheckerDiagnostic,
} from "../src/typechecker/diagnostics";
import { formatTypecheckerDiagnostic, printPackageSummaries } from "./typecheck-utils";
import { resolveTypecheckMode, type TypecheckMode } from "./typecheck-mode";
import { ModuleLoader, type Program } from "./module-loader";
import { collectModuleSearchPaths, type ModuleSearchPath } from "./module-search-paths";
import { buildExecutionSearchPaths, loadManifestContext } from "./module-deps";
import {
  collectFixtures,
  readManifest,
  loadModuleFromFixture,
  loadModuleFromPath,
  ensurePrint,
  installRuntimeStubs,
  interceptStdout,
  formatValue,
  extractErrorMessage,
  formatRuntimeErrorMessage,
  type Manifest,
} from "./fixture-utils";
import { startRunTimeout } from "./test-timeouts";
import { serializeMapEntries } from "../src/interpreter/maps";
import { ExitSignal } from "../src/interpreter/signals";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../../fixtures/ast");
const TYPECHECK_MODE = resolveTypecheckMode(process.env.ABLE_TYPECHECK_FIXTURES);
const TYPECHECK_BASELINE_PATH = path.join(FIXTURE_ROOT, "typecheck-baseline.json");
const WRITE_TYPECHECK_BASELINE = process.argv.includes("--write-typecheck-baseline");
const FIXTURE_FILTER = process.env.ABLE_FIXTURE_FILTER?.trim() ?? null;
const BASELINE_ENABLED = TYPECHECK_MODE !== "off" && !FIXTURE_FILTER;
const STDLIB_ROOT = path.resolve(__dirname, "../../../stdlib/src");
const KERNEL_ROOT = path.resolve(__dirname, "../../../kernel/src");
const KERNEL_ENTRY = path.join(KERNEL_ROOT, "kernel.able");
const STDLIB_STRING_ENTRY = path.join(STDLIB_ROOT, "text", "string.able");
const STDLIB_CONCURRENCY_ENTRY = path.join(STDLIB_ROOT, "concurrency", "await.able");
const STDLIB_HASH_MAP_ENTRY = path.join(STDLIB_ROOT, "collections", "hash_map.able");
const stdlibLoader = new ModuleLoader([STDLIB_ROOT, KERNEL_ROOT]);
let stdlibStringProgram: Program | null = null;
let stdlibConcurrencyProgram: Program | null = null;
let stdlibHashMapProgram: Program | null = null;
let kernelProgram: Program | null = null;
const EXEC_ROOT = path.resolve(__dirname, "../../../fixtures/exec");

type FixtureResult = { name: string; description?: string };

async function main() {
  const clearTimeouts = startRunTimeout("run-fixtures");
  try {
    const fixtures = await collectFixtures(FIXTURE_ROOT);
    if (fixtures.length === 0) {
      console.log("No fixtures found.");
      return;
    }

    const results: FixtureResult[] = [];
    const typecheckBaseline = new Map<string, string[]>();
    const existingBaseline = BASELINE_ENABLED
      ? await readExistingBaseline(TYPECHECK_BASELINE_PATH)
      : null;
    if (BASELINE_ENABLED && !existingBaseline && !WRITE_TYPECHECK_BASELINE) {
      throw new Error(
        `typechecker baseline missing at ${TYPECHECK_BASELINE_PATH}. Run with --write-typecheck-baseline to generate it.`,
      );
    }
    if (FIXTURE_FILTER && TYPECHECK_MODE !== "off" && !BASELINE_ENABLED) {
      console.warn(
        `typechecker: skipping baseline enforcement because ABLE_FIXTURE_FILTER=${FIXTURE_FILTER}`,
      );
    }

    for (const fixtureDir of fixtures) {
      const manifest = await readManifest(fixtureDir);
      if (manifest.skipTargets?.includes("ts")) {
        continue;
      }

      const relativeName = path.relative(FIXTURE_ROOT, fixtureDir).split(path.sep).join("/");
      if (FIXTURE_FILTER && !relativeName.includes(FIXTURE_FILTER)) {
        continue;
      }
      const interpreter = new V11.Interpreter();
      ensurePrint(interpreter);
      installRuntimeStubs(interpreter);

      const stdout: string[] = [];
      let evaluationError: unknown;

      const entry = manifest.entry ?? "module.json";
      const moduleAst = await loadModuleFromFixture(fixtureDir, entry);
      const needsStdlibString = moduleImportsStdlibString(moduleAst);
      const needsStdlibConcurrency = moduleImportsStdlibConcurrency(moduleAst);
      const needsStdlibHashMap = moduleImportsStdlibHashMap(moduleAst);
      let needsKernel = moduleImportsKernel(moduleAst);
      const stdlibPrograms: Program[] = [];
      if (needsStdlibString) {
        stdlibPrograms.push(await ensureStdlibStringProgram());
      }
      if (needsStdlibConcurrency) {
        stdlibPrograms.push(await ensureStdlibConcurrencyProgram());
      }
      if (needsStdlibHashMap) {
        stdlibPrograms.push(await ensureStdlibHashMapProgram());
      }
      const setupModules: AST.Module[] = [];
      if (manifest.setup) {
        for (const setupFile of manifest.setup) {
          const setupPath = path.join(fixtureDir, setupFile);
          const setupModule = await loadModuleFromPath(setupPath);
          if (!needsKernel && moduleImportsKernel(setupModule)) {
            needsKernel = true;
          }
          setupModules.push(setupModule);
        }
      }
      const includeKernel =
        needsKernel && !stdlibPrograms.some((program) => programHasPackage(program, "able.kernel"));
      const preludePrograms: Program[] = [];
      if (includeKernel) {
        preludePrograms.push(await ensureKernelProgram());
      }

      const typecheckDiagnostics: TypecheckerDiagnostic[] = [];
      let packageSummaries = new Map<string, PackageSummary>();
      if (TYPECHECK_MODE !== "off") {
        const session = new TypeChecker.TypecheckerSession();
        const seenPkgs = new Set<string>();
        for (const preludeProgram of preludePrograms) {
          for (const mod of preludeProgram.modules) {
            if (seenPkgs.has(mod.packageName)) continue;
            seenPkgs.add(mod.packageName);
            const { diagnostics } = session.checkModule(mod.module);
            typecheckDiagnostics.push(...diagnostics);
          }
          if (!seenPkgs.has(preludeProgram.entry.packageName)) {
            const { diagnostics } = session.checkModule(preludeProgram.entry.module);
            typecheckDiagnostics.push(...diagnostics);
            seenPkgs.add(preludeProgram.entry.packageName);
          }
        }
        for (const stdlibProgram of stdlibPrograms) {
          for (const mod of stdlibProgram.modules) {
            if (seenPkgs.has(mod.packageName)) continue;
            seenPkgs.add(mod.packageName);
            const { diagnostics } = session.checkModule(mod.module);
            typecheckDiagnostics.push(...diagnostics);
          }
          if (!seenPkgs.has(stdlibProgram.entry.packageName)) {
            const { diagnostics } = session.checkModule(stdlibProgram.entry.module);
            typecheckDiagnostics.push(...diagnostics);
            seenPkgs.add(stdlibProgram.entry.packageName);
          }
        }
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
    if (BASELINE_ENABLED) {
      typecheckBaseline.set(relativeName, formattedDiagnostics);
    }
    if (BASELINE_ENABLED && existingBaseline && !WRITE_TYPECHECK_BASELINE) {
      enforceTypecheckBaseline(relativeName, TYPECHECK_MODE, formattedDiagnostics, existingBaseline);
    }

      let result: V11.RuntimeValue | undefined;
      await interceptStdout(stdout, async () => {
        try {
          for (const preludeProgram of preludePrograms) {
            await evaluateProgram(interpreter, preludeProgram);
          }
          for (const stdlibProgram of stdlibPrograms) {
            await evaluateProgram(interpreter, stdlibProgram);
          }
          for (const setupModule of setupModules) {
            await interpreter.evaluateAsTask(setupModule);
          }
          result = await interpreter.evaluateAsTask(moduleAst);
        } catch (err) {
          evaluationError = err;
        }
      });

      assertExpectations(fixtureDir, manifest.expect, result, stdout, evaluationError, typecheckDiagnostics, interpreter);
      results.push({ name: relativeName, description: manifest.description });
    }

    if (BASELINE_ENABLED) {
      const baselineObject = Object.fromEntries(
        [...typecheckBaseline.entries()].sort(([a], [b]) => a.localeCompare(b)),
      );
      if (WRITE_TYPECHECK_BASELINE) {
        await fs.writeFile(TYPECHECK_BASELINE_PATH, `${JSON.stringify(baselineObject, null, 2)}\n`, "utf8");
      } else {
        const differences = diffBaselineMaps(baselineObject, existingBaseline ?? {});
        if (differences.length > 0) {
          const message = [`Typecheck baseline mismatch:`].concat(differences).join("\n  ");
          throw new Error(message);
        }
      }
    } else if (WRITE_TYPECHECK_BASELINE) {
      const reason = TYPECHECK_MODE === "off" ? "ABLE_TYPECHECK_FIXTURES is off" : "ABLE_FIXTURE_FILTER is set";
      console.warn(`typechecker: skipping baseline write because ${reason}`);
    }

    for (const res of results) {
      const desc = res.description ? ` - ${res.description}` : "";
      console.log(`✓ ${res.name}${desc}`);
    }
    const execFixtures = await collectExecFixtures(EXEC_ROOT);
    for (const execDir of execFixtures) {
      const relativeName = path.relative(EXEC_ROOT, execDir).split(path.sep).join("/");
      if (FIXTURE_FILTER && !relativeName.includes(FIXTURE_FILTER)) {
        continue;
      }
      const res = await runExecFixture(execDir);
      results.push(res);
      const desc = res.description ? ` - ${res.description}` : "";
      console.log(`✓ ${res.name}${desc}`);
    }
    console.log(`Executed ${results.length} fixture(s).`);
  } finally {
    clearTimeouts();
  }
}

async function collectExecFixtures(root: string): Promise<string[]> {
  const dirs: string[] = [];
  async function walk(current: string) {
    const entries = await fs.readdir(current, { withFileTypes: true });
    let hasManifest = false;
    for (const entry of entries) {
      if (entry.isFile() && entry.name === "manifest.json") {
        hasManifest = true;
        break;
      }
    }
    if (hasManifest) {
      dirs.push(current);
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        await walk(path.join(current, entry.name));
      }
    }
  }
  try {
    await walk(root);
  } catch (err: any) {
    if (err && err.code === "ENOENT") {
      return [];
    }
    throw err;
  }
  return dirs.sort();
}

async function runExecFixture(dir: string): Promise<FixtureResult> {
  const manifest = await readManifest(dir);
  const entryRel = manifest.entry ?? "main.able";
  const entryPath = path.join(dir, entryRel);
  const searchPaths = await resolveExecSearchPaths(dir, entryPath, manifest);
  const loader = new ModuleLoader(searchPaths);
  const program = await loader.load(entryPath);

  const interpreter = new V11.Interpreter();
  installRuntimeStubs(interpreter);

  if (manifest.expect?.typecheckDiagnostics !== undefined) {
    const typecheckDiagnostics: TypecheckerDiagnostic[] = [];
    let packageSummaries = new Map<string, PackageSummary>();
    if (TYPECHECK_MODE !== "off") {
      const session = new TypeChecker.TypecheckerSession();
      const seenPkgs = new Set<string>();
      for (const mod of program.modules) {
        if (seenPkgs.has(mod.packageName)) continue;
        seenPkgs.add(mod.packageName);
        const { diagnostics } = session.checkModule(mod.module);
        typecheckDiagnostics.push(...diagnostics);
      }
      if (!seenPkgs.has(program.entry.packageName)) {
        const { diagnostics } = session.checkModule(program.entry.module);
        typecheckDiagnostics.push(...diagnostics);
      }
      packageSummaries = session.getPackageSummaries();
    }
    maybeReportTypecheckDiagnostics(
      dir,
      TYPECHECK_MODE,
      manifest.expect?.typecheckDiagnostics ?? null,
      typecheckDiagnostics,
      packageSummaries,
    );
  }

  const stdout: string[] = [];
  const nilValue: V11.RuntimeValue = { kind: "nil", value: null };
  interpreter.globals.define(
    "print",
    interpreter.makeNativeFunction("print", 1, (_interp, args) => {
      const parts = args.map((value) => formatValue(value));
      stdout.push(parts.join(" "));
      return nilValue;
    }),
  );
  let exitCode = 0;
  let runtimeError: unknown;
  let exitSignaled = false;

  try {
    const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
    for (const mod of nonEntry) {
      await interpreter.evaluateAsTask(mod.module);
    }
    await interpreter.evaluateAsTask(program.entry.module);
    const pkg = interpreter.packageRegistry.get(program.entry.packageName);
    const mainFn = pkg?.get("main");
    if (!mainFn) {
      throw new Error("entry module missing main");
    }
    const callEnv = new Environment(interpreter.globals);
    callEnv.define("main", mainFn);
    const callNode = AST.functionCall(AST.identifier("main"), []);
    await interpreter.evaluateAsTask(callNode, callEnv);
  } catch (err) {
    if (err instanceof ExitSignal) {
      exitSignaled = true;
      exitCode = err.code;
    } else {
      exitCode = 1;
      runtimeError = err;
    }
  }

  const expected = manifest.expect ?? {};
  if (runtimeError) {
    if (expected.exit === undefined || exitCode !== expected.exit) {
      throw runtimeError instanceof Error ? runtimeError : new Error(String(runtimeError));
    }
  }
  if (expected.stdout) {
    if (JSON.stringify(stdout) !== JSON.stringify(expected.stdout)) {
      throw new Error(
        `stdout mismatch for ${dir}: expected ${JSON.stringify(expected.stdout)}, got ${JSON.stringify(stdout)}`,
      );
    }
  }
  if (expected.stderr) {
    const actualErr = runtimeError ? [formatRuntimeErrorMessage(runtimeError)] : [];
    if (JSON.stringify(actualErr) !== JSON.stringify(expected.stderr)) {
      throw new Error(
        `stderr mismatch for ${dir}: expected ${JSON.stringify(expected.stderr)}, got ${JSON.stringify(actualErr)}`,
      );
    }
  }
  if (expected.exit !== undefined) {
    if (exitCode !== expected.exit) {
      throw new Error(`exit code mismatch for ${dir}: expected ${expected.exit}, got ${exitCode}`);
    }
  } else if (exitSignaled) {
    throw new Error(`exit code mismatch for ${dir}: expected default exit, got ${exitCode}`);
  } else if (runtimeError) {
    throw runtimeError instanceof Error ? runtimeError : new Error(String(runtimeError));
  }

  return { name: `exec/${path.basename(dir)}`, description: manifest.description };
}

function resolveEnvPathList(raw: string | undefined, baseDir: string): string {
  if (!raw) return "";
  const resolved = raw
    .split(path.delimiter)
    .map((segment) => segment.trim())
    .filter(Boolean)
    .map((segment) => (path.isAbsolute(segment) ? segment : path.resolve(baseDir, segment)));
  return resolved.join(path.delimiter);
}

async function resolveExecSearchPaths(
  dir: string,
  entryPath: string,
  manifest: Manifest,
): Promise<ModuleSearchPath[]> {
  const entryDir = path.dirname(path.resolve(entryPath));
  let manifestRoot: string | null = null;
  let extras: ModuleSearchPath[] = [];
  const context = await loadManifestContext(entryDir);
  manifestRoot = context.manifest?.path ? path.dirname(context.manifest.path) : null;
  extras = buildExecutionSearchPaths(context.manifest, context.lock);

  const env = manifest.env ?? {};
  const ablePathEnv = resolveEnvPathList(env.ABLE_PATH ?? process.env.ABLE_PATH, dir);
  const ableModulePathsEnv = resolveEnvPathList(
    env.ABLE_MODULE_PATHS ?? process.env.ABLE_MODULE_PATHS,
    dir,
  );
  const probeFrom = [entryDir, manifestRoot, process.cwd(), __dirname, path.dirname(process.execPath)].filter(
    Boolean,
  ) as string[];

  return collectModuleSearchPaths({
    cwd: entryDir ?? process.cwd(),
    ablePathEnv,
    ableModulePathsEnv,
    extras,
    probeFrom,
  });
}

function moduleImportsStdlibString(module: AST.Module): boolean {
  return moduleImportsPackage(module, "able.text.String") || moduleImportsPackage(module, "able.text.string");
}

function moduleImportsStdlibConcurrency(module: AST.Module): boolean {
  return moduleImportsPackage(module, "able.concurrency");
}

function moduleImportsStdlibHashMap(module: AST.Module): boolean {
  return moduleImportsPackage(module, "able.collections.hash_map");
}

function moduleImportsKernel(module: AST.Module): boolean {
  return moduleImportsPackage(module, "able.kernel");
}

function moduleImportsPackage(module: AST.Module, pkgName: string): boolean {
  return Array.isArray((module as any)?.imports)
    ? (module as any).imports.some((imp: any) => {
        if (!imp || imp.type !== "ImportStatement" || !Array.isArray(imp.packagePath)) return false;
        const pkg = imp.packagePath.map((segment: any) => segment?.name ?? "").filter(Boolean).join(".");
        return pkg === pkgName;
      })
    : false;
}

async function ensureStdlibStringProgram(): Promise<Program> {
  if (!stdlibStringProgram) {
    stdlibStringProgram = await stdlibLoader.load(STDLIB_STRING_ENTRY);
  }
  return stdlibStringProgram;
}

async function ensureStdlibConcurrencyProgram(): Promise<Program> {
  if (!stdlibConcurrencyProgram) {
    stdlibConcurrencyProgram = await stdlibLoader.load(STDLIB_CONCURRENCY_ENTRY);
  }
  return stdlibConcurrencyProgram;
}

async function ensureStdlibHashMapProgram(): Promise<Program> {
  if (!stdlibHashMapProgram) {
    stdlibHashMapProgram = await stdlibLoader.load(STDLIB_HASH_MAP_ENTRY);
  }
  return stdlibHashMapProgram;
}

async function ensureKernelProgram(): Promise<Program> {
  if (!kernelProgram) {
    const kernelModule = await loadModuleFromPath(KERNEL_ENTRY);
    kernelModule.package = AST.packageStatement(["able", "kernel"]);
    kernelProgram = {
      entry: {
        packageName: "able.kernel",
        module: kernelModule,
        files: [KERNEL_ENTRY],
        imports: [],
        dynImports: [],
      },
      modules: [],
    };
  }
  return kernelProgram;
}

function programHasPackage(program: Program, name: string): boolean {
  if (program.entry.packageName === name) return true;
  return program.modules.some((mod) => mod.packageName === name);
}

async function evaluateProgram(interpreter: V11.Interpreter, program: Program): Promise<void> {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    await interpreter.evaluateAsTask(mod.module);
  }
  await interpreter.evaluateAsTask(program.entry.module);
}

async function readExistingBaseline(filePath: string): Promise<Record<string, string[]> | null> {
  try {
    const raw = await fs.readFile(filePath, "utf8");
    return JSON.parse(raw) as Record<string, string[]>;
  } catch (err: any) {
    if (err && err.code === "ENOENT") return null;
    throw err;
  }
}

function assertExpectations(
  dir: string,
  expect: Manifest["expect"],
  result: V11.RuntimeValue | undefined,
  stdout: string[],
  evaluationError: unknown,
  _typecheckDiagnostics: TypecheckerDiagnostic[],
  interpreter: V11.Interpreter,
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
  if (!result) {
    throw new Error(`Fixture ${dir} produced no result`);
  }
  if (expect.result) {
    if (result.kind !== expect.result.kind) {
      throw new Error(
        `Fixture ${dir} expected result kind ${expect.result.kind}, got ${result.kind} (${JSON.stringify(result)})`,
      );
    }
    if (expect.result.value !== undefined) {
      const formatted = formatValue(result);
      if (formatted !== String(expect.result.value)) {
        throw new Error(
          `Fixture ${dir} expected value ${JSON.stringify(expect.result.value)}, got ${formatted}`,
        );
      }
    }
    if (expect.result.entries) {
      compareMapEntries(interpreter, dir, expect.result.entries, result);
    }
  }
  if (expect.stdout) {
    const normalize = (values: string[]) =>
      values.map((value) => {
        if (typeof value !== "string") return value;
        if (value.startsWith('"') && value.endsWith('"')) {
          try {
            const parsed = JSON.parse(value);
            if (
              typeof parsed === "string" ||
              typeof parsed === "number" ||
              typeof parsed === "boolean" ||
              parsed === null
            ) {
              return parsed;
            }
          } catch {
            // ignore JSON parse failures and fall back to raw value
          }
        }
        return value;
      });
    const normalizedActual = normalize(stdout);
    const normalizedExpected = normalize(expect.stdout);
    if (JSON.stringify(normalizedActual) !== JSON.stringify(normalizedExpected)) {
      throw new Error(
        `Fixture ${dir} expected stdout ${JSON.stringify(normalizedExpected)}, got ${JSON.stringify(normalizedActual)}`,
      );
    }
  }
}

function compareMapEntries(
  interpreter: V11.Interpreter,
  dir: string,
  expected: { key: unknown; value: unknown }[],
  value: V11.RuntimeValue,
): void {
  const entries = serializeMapEntries(interpreter, value);
  if (entries.length !== expected.length) {
    throw new Error(
      `Fixture ${dir} expected ${expected.length} map entries, got ${entries.length}`,
    );
  }
  for (let i = 0; i < entries.length; i += 1) {
    const actualEntry = entries[i]!;
    const normalizedKey = normalizeValueForExpect(actualEntry.key);
    const normalizedValue = normalizeValueForExpect(actualEntry.value);
    if (JSON.stringify(normalizedKey) !== JSON.stringify(expected[i]?.key)) {
      throw new Error(
        `Fixture ${dir} expected map key ${JSON.stringify(expected[i]?.key)}, got ${JSON.stringify(normalizedKey)}`,
      );
    }
    if (JSON.stringify(normalizedValue) !== JSON.stringify(expected[i]?.value)) {
      throw new Error(
        `Fixture ${dir} expected map value ${JSON.stringify(expected[i]?.value)}, got ${JSON.stringify(normalizedValue)}`,
      );
    }
  }
}

function normalizeValueForExpect(value: V11.RuntimeValue): unknown {
  switch (value.kind) {
    case "String":
    case "char":
    case "bool":
    case "i32":
    case "f64":
      return { kind: value.kind, value: value.value };
    case "nil":
      return { kind: "nil", value: null };
    case "void":
      return { kind: "void" };
    default:
      return { kind: value.kind };
  }
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

  const formattedActual = dedupeDiagnostics(actual.map(formatTypecheckerDiagnostic));
  const expectedNormalized = expected ? dedupeDiagnostics(expected) : null;

  if (expectedNormalized && expectedNormalized.length > 0) {
    if (formattedActual.length === 0) {
      if (mode === "strict") {
        throw new Error(
          `Fixture ${dir} expected typechecker diagnostics ${JSON.stringify(expectedNormalized)} but none were produced`,
        );
      }
      console.warn(
        `typechecker: fixture ${dir} expected diagnostics ${JSON.stringify(expectedNormalized)} but checker returned none (mode=${mode})`,
      );
      return formattedActual;
    }
    const actualKeys = formattedActual.map(toDiagnosticKey);
    const expectedKeys = expectedNormalized.map(toDiagnosticKey);
    const allExpectedPresent = expectedKeys.every(expectedKey =>
      actualKeys.some(actualKey => diagnosticKeyMatches(actualKey, expectedKey)),
    );
    if (!allExpectedPresent) {
      printPackageSummaries(summaries);
      const message = `Fixture ${dir} expected typechecker diagnostics ${JSON.stringify(expectedNormalized)}, got ${JSON.stringify(formattedActual)}`;
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

function enforceTypecheckBaseline(
  rel: string,
  mode: TypecheckMode,
  actual: string[],
  baseline: Record<string, string[]> | null,
) {
  if (mode === "off" || !baseline) return;
  const expected = dedupeDiagnostics(baseline[rel] ?? []);
  const dedupedActual = dedupeDiagnostics(actual);
  if (expected.length === 0) {
    if (dedupedActual.length > 0) {
      throw new Error(`typechecker diagnostics mismatch for ${rel}: expected none, got ${JSON.stringify(dedupedActual)}`);
    }
    return;
  }
  const actualKeys = dedupedActual.map(toDiagnosticKey);
  const expectedKeys = expected.map(toDiagnosticKey);
  if (expectedKeys.length !== actualKeys.length) {
    throw new Error(`typechecker diagnostics mismatch for ${rel}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(dedupedActual)}`);
  }
  for (let i = 0; i < expectedKeys.length; i += 1) {
    if (!diagnosticKeyMatches(actualKeys[i], expectedKeys[i])) {
      throw new Error(`typechecker diagnostics mismatch for ${rel}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(dedupedActual)}`);
    }
  }
}

function diffBaselineMaps(
  actual: Record<string, string[]>,
  expected: Record<string, string[]>,
): string[] {
  const differences: string[] = [];
  const allKeys = new Set([...Object.keys(actual), ...Object.keys(expected)]);
  for (const key of [...allKeys].sort()) {
    const actualValues = dedupeDiagnostics(actual[key] ?? []);
    const expectedValues = dedupeDiagnostics(expected[key] ?? []);
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
  let trimmed = entry;
  let severityPrefix = "";
  if (trimmed.startsWith("warning: ")) {
    severityPrefix = "warning: ";
    trimmed = trimmed.slice("warning: ".length);
  }
  trimmed = trimmed.startsWith("typechecker: ") ? trimmed.slice("typechecker: ".length) : trimmed;
  const firstSpace = trimmed.indexOf(" ");
  if (firstSpace === -1) {
    return `${severityPrefix}${trimmed}`;
  }
  const location = trimmed.slice(0, firstSpace);
  let message = trimmed.slice(firstSpace + 1);
  if (!message.startsWith("typechecker:")) {
    message = `typechecker: ${message}`;
  }
  if (severityPrefix) {
    message = `${severityPrefix}${message}`;
  }
  const segments = location.split(":");
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
  if (segments.length >= 2) {
    const columnMaybe = takeNumeric();
    const lineMaybe = takeNumeric();
    if (typeof lineMaybe === "number") {
      line = lineMaybe;
    } else if (typeof columnMaybe === "number") {
      line = columnMaybe;
    }
  } else {
    pathSegments = [location];
  }
  let pathPart = pathSegments.join(":");
  if (pathPart !== "") {
    pathPart = pathPart.replace(/\\/g, "/");
    while (pathPart.startsWith("../")) {
      pathPart = pathPart.slice(3);
    }
    if (pathPart === "typechecker") {
      pathPart = "";
    }
  }
  return `${pathPart}:${line}|${message}`;
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

function dedupeDiagnostics(entries: string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const entry of entries) {
    const key = toDiagnosticKey(entry);
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(entry);
  }
  return result;
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
