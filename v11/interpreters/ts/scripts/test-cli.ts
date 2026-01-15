import fs from "node:fs";
import { promises as fsPromises } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, TypeChecker, V11 } from "../index";
import type { TypecheckerDiagnostic } from "../src/typechecker/diagnostics";
import { callCallableValue } from "../src/interpreter/functions";
import { makeIntegerValue } from "../src/interpreter/numeric";
import { ExitSignal } from "../src/interpreter/signals";
import { ParserDiagnosticError } from "../src/parser/diagnostics";

import { ModuleLoader, type Program } from "./module-loader";
import { collectModuleSearchPaths, type ModuleSearchPath } from "./module-search-paths";
import { buildExecutionSearchPaths, loadManifestContext } from "./module-deps";
import { discoverRoot, indexSourceFiles } from "./module-utils";
import { normalizeRepoRelativePath } from "./path-utils";
import { buildRuntimeDiagnostic, formatParserDiagnostic, formatRuntimeDiagnostic, formatTypecheckerDiagnostic, printPackageSummaries } from "./typecheck-utils";
import type { TypecheckMode } from "./typecheck-mode";

export type TestReporterFormat = "doc" | "progress" | "tap" | "json";

export type TestCliFilters = {
  includePaths: string[];
  excludePaths: string[];
  includeNames: string[];
  excludeNames: string[];
  includeTags: string[];
  excludeTags: string[];
};

export type TestRunOptions = {
  failFast: boolean;
  repeat: number;
  parallelism: number;
  shuffleSeed?: number;
};

export type TestCliConfig = {
  targets: string[];
  filters: TestCliFilters;
  run: TestRunOptions;
  reporterFormat: TestReporterFormat;
  listOnly: boolean;
  dryRun: boolean;
};

type HarnessFailure = {
  message: string;
  details: string | null;
};

type RootTestPlan = {
  rootDir: string;
  rootName: string;
  entryFile: string;
  testFiles: string[];
  packageNames: Set<string>;
};

export type TestLoadResult = {
  programs: Program[];
  modules: Program["modules"];
};

type ModuleDiagnosticEntry = {
  packageName: string;
  diagnostic: TypecheckerDiagnostic;
};

type TestCliEnv = {
  ablePathEnv: string;
  ableModulePathsEnv: string;
};

const TEST_FILE_SUFFIXES = [".test.able", ".spec.able"];

export async function resolveTestTargets(targets: string[]): Promise<string[]> {
  const rawTargets = targets.length > 0 ? targets : ["."];
  const resolved: string[] = [];
  const seen = new Set<string>();

  for (const target of rawTargets) {
    const abs = path.resolve(process.cwd(), target);
    let stats: fs.Stats;
    try {
      stats = await fsPromises.stat(abs);
    } catch (error) {
      throw new Error(`unable to access ${abs}: ${(error as Error).message}`);
    }
    if (stats.isFile()) {
      if (isTestFile(abs)) {
        if (!seen.has(abs)) {
          seen.add(abs);
          resolved.push(abs);
        }
      } else {
        const dir = path.dirname(abs);
        if (!seen.has(dir)) {
          seen.add(dir);
          resolved.push(dir);
        }
      }
      continue;
    }
    if (!stats.isDirectory()) {
      throw new Error(`unsupported test target: ${abs}`);
    }
    if (!seen.has(abs)) {
      seen.add(abs);
      resolved.push(abs);
    }
  }

  return resolved;
}

export async function collectTestFiles(targets: string[]): Promise<string[]> {
  const found = new Set<string>();
  for (const target of targets) {
    const stats = await fsPromises.stat(target);
    if (stats.isFile()) {
      if (isTestFile(target)) {
        found.add(path.resolve(target));
      }
      continue;
    }
    if (stats.isDirectory()) {
      await walkTestFiles(target, found);
    }
  }
  return [...found].sort();
}

export async function loadTestPrograms(
  testFiles: string[],
  env: TestCliEnv,
): Promise<TestLoadResult | null> {
  let plans: RootTestPlan[];
  try {
    plans = await buildRootTestPlans(testFiles);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    return null;
  }

  if (plans.length === 0) {
    return { programs: [], modules: [] };
  }

  const searchPaths = await resolveTestSearchPaths(plans, env);
  if (!searchPaths) {
    return null;
  }

  const programs: Program[] = [];
  for (const plan of plans) {
    const includePackages = new Set<string>([
      "able.test.harness",
      "able.test.protocol",
      "able.test.reporters",
      ...plan.packageNames,
    ]);
    const loader = new ModuleLoader(searchPaths);
    try {
      const program = await loader.load(plan.entryFile, { includePackages: [...includePackages] });
      programs.push(program);
    } catch (error) {
      if (error instanceof ParserDiagnosticError) {
        console.error(formatParserDiagnostic(error.diagnostic, { absolutePath: true }));
        return null;
      }
      console.error(`failed to load tests from ${plan.entryFile}: ${extractErrorMessage(error)}`);
      return null;
    }
  }

  return {
    programs,
    modules: mergeTestModules(programs),
  };
}

export function maybeTypecheckTestModules(
  session: TypeChecker.TypecheckerSession,
  modules: Program["modules"],
  typecheckMode: TypecheckMode,
): boolean {
  if (typecheckMode === "off") {
    return true;
  }
  const diagnostics: ModuleDiagnosticEntry[] = [];
  const seen = new Set<string>();
  for (const mod of modules) {
    if (seen.has(mod.packageName)) {
      continue;
    }
    seen.add(mod.packageName);
    const result = session.checkModule(mod.module);
    if (result.diagnostics.length === 0) {
      continue;
    }
    const packageName =
      result.summary?.name ?? mod.packageName ?? resolveModulePackageName(mod.module);
    if (isStdlibPackage(packageName)) {
      continue;
    }
    const filtered = result.diagnostics.filter((diag) => !isStdlibDiagnostic(diag));
    if (filtered.length === 0) {
      continue;
    }
    for (const diag of filtered) {
      diagnostics.push({ packageName, diagnostic: diag });
    }
  }
  if (diagnostics.length === 0) {
    return true;
  }
  emitDiagnostics(diagnostics);
  printPackageSummaries(session.getPackageSummaries());

  if (typecheckMode === "strict") {
    process.exitCode = 2;
    return false;
  }

  console.warn("typechecker: proceeding despite diagnostics because ABLE_TYPECHECK_FIXTURES=warn");
  return true;
}

export async function evaluateTestModules(interpreter: V11.Interpreter, modules: Program["modules"]): Promise<boolean> {
  const evaluated = new Set<string>();
  for (const mod of modules) {
    if (evaluated.has(mod.packageName)) {
      continue;
    }
    evaluated.add(mod.packageName);
    try {
      await interpreter.evaluateAsTask(mod.module);
    } catch (error) {
      if (error instanceof ExitSignal) {
        process.exitCode = error.code;
        return false;
      }
      console.error(formatRuntimeFailure(error));
      process.exitCode = 2;
      return false;
    }
  }
  return true;
}

export function buildDiscoveryRequest(interpreter: V11.Interpreter, config: TestCliConfig): V11.RuntimeValue {
  const def = getStructDef(interpreter, "able.test.protocol", "DiscoveryRequest");
  return interpreter.makeNamedStructInstance(def, [
    ["include_paths", makeStringArray(interpreter, config.filters.includePaths)],
    ["exclude_paths", makeStringArray(interpreter, config.filters.excludePaths)],
    ["include_names", makeStringArray(interpreter, config.filters.includeNames)],
    ["exclude_names", makeStringArray(interpreter, config.filters.excludeNames)],
    ["include_tags", makeStringArray(interpreter, config.filters.includeTags)],
    ["exclude_tags", makeStringArray(interpreter, config.filters.excludeTags)],
    ["list_only", { kind: "bool", value: config.listOnly }],
  ]);
}

export function buildRunOptions(interpreter: V11.Interpreter, run: TestRunOptions): V11.RuntimeValue {
  const def = getStructDef(interpreter, "able.test.protocol", "RunOptions");
  const shuffleSeed =
    run.shuffleSeed === undefined ? { kind: "nil", value: null } : makeIntegerValue("i64", BigInt(run.shuffleSeed));
  return interpreter.makeNamedStructInstance(def, [
    ["shuffle_seed", shuffleSeed],
    ["fail_fast", { kind: "bool", value: run.failFast }],
    ["parallelism", makeIntegerValue("i32", BigInt(run.parallelism))],
    ["repeat", makeIntegerValue("i32", BigInt(run.repeat))],
  ]);
}

export function buildTestPlan(interpreter: V11.Interpreter, descriptors: V11.RuntimeValue): V11.RuntimeValue {
  const def = getStructDef(interpreter, "able.test.protocol", "TestPlan");
  return interpreter.makeNamedStructInstance(def, [["descriptors", descriptors]]);
}

export function callHarnessDiscover(
  interpreter: V11.Interpreter,
  request: V11.RuntimeValue,
): V11.RuntimeValue | null {
  const discover = getCallableSymbol(interpreter, "able.test.harness", "discover_all");
  if (!discover) {
    console.error("able test: unable to find able.test.harness.discover_all");
    return null;
  }
  try {
    const result = callCallableValue(interpreter as any, discover, [request], interpreter.globals);
    const failure = extractFailure(interpreter, result);
    if (failure) {
      console.error(`able test: ${formatFailure(failure)}`);
      return null;
    }
    if (result.kind !== "array" && !isKernelArray(result)) {
      console.error("able test: discovery returned unexpected result");
      return null;
    }
    return result;
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    return null;
  }
}

export async function callHarnessRun(
  interpreter: V11.Interpreter,
  plan: V11.RuntimeValue,
  options: V11.RuntimeValue,
  reporter: V11.RuntimeValue,
): Promise<HarnessFailure | null> {
  const runPlan = getCallableSymbol(interpreter, "able.test.harness", "run_plan");
  if (!runPlan) {
    console.error("able test: unable to find able.test.harness.run_plan");
    return { message: "missing able.test.harness.run_plan", details: null };
  }
  const callEnv = new V11.Environment(interpreter.globals);
  const runPlanIdent = "__able_test_run_plan";
  const planIdent = "__able_test_plan";
  const optionsIdent = "__able_test_options";
  const reporterIdent = "__able_test_reporter";
  callEnv.define(runPlanIdent, runPlan);
  callEnv.define(planIdent, plan);
  callEnv.define(optionsIdent, options);
  callEnv.define(reporterIdent, reporter);
  const callNode = AST.functionCall(AST.identifier(runPlanIdent), [
    AST.identifier(planIdent),
    AST.identifier(optionsIdent),
    AST.identifier(reporterIdent),
  ]);
  try {
    const result = await interpreter.evaluateAsTask(callNode, callEnv);
    if (result.kind === "nil") {
      return null;
    }
    const failure = extractFailure(interpreter, result);
    if (failure) {
      console.error(`able test: ${formatFailure(failure)}`);
      return failure;
    }
    console.error("able test: run_plan returned unexpected result");
    return { message: "run_plan returned unexpected result", details: null };
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    return { message: extractErrorMessage(error), details: null };
  }
}

function isTestFile(pathname: string): boolean {
  return TEST_FILE_SUFFIXES.some((suffix) => pathname.endsWith(suffix));
}

async function walkTestFiles(dir: string, found: Set<string>): Promise<void> {
  const entries = await fsPromises.readdir(dir, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === "quarantine" || entry.name === "node_modules" || entry.name === ".git") {
        continue;
      }
      await walkTestFiles(fullPath, found);
      continue;
    }
    if (entry.isFile() && isTestFile(entry.name)) {
      found.add(path.resolve(fullPath));
    }
  }
}

async function buildRootTestPlans(testFiles: string[]): Promise<RootTestPlan[]> {
  const roots = new Map<string, RootTestPlan>();
  for (const file of testFiles) {
    const { rootDir, rootName } = await discoverRoot(file);
    const existing = roots.get(rootDir);
    if (existing) {
      if (existing.rootName !== rootName) {
        throw new Error(`conflicting package roots at ${rootDir}`);
      }
      existing.testFiles.push(file);
      continue;
    }
    roots.set(rootDir, {
      rootDir,
      rootName,
      entryFile: file,
      testFiles: [file],
      packageNames: new Set(),
    });
  }

  for (const plan of roots.values()) {
    const { fileToPackage } = await indexSourceFiles(plan.rootDir, plan.rootName);
    for (const file of plan.testFiles) {
      const pkgName = fileToPackage.get(file);
      if (!pkgName) {
        throw new Error(`unable to resolve package for test file ${file}`);
      }
      plan.packageNames.add(pkgName as string);
    }
  }

  return [...roots.values()];
}

async function resolveTestSearchPaths(
  plans: RootTestPlan[],
  env: TestCliEnv,
): Promise<ModuleSearchPath[] | null> {
  const extras: ModuleSearchPath[] = [];
  const probeFrom = new Set<string>();

  for (const plan of plans) {
    probeFrom.add(plan.rootDir);
    try {
      const manifestContext = await loadManifestContext(plan.rootDir);
      if (manifestContext.manifest?.path) {
        probeFrom.add(path.dirname(manifestContext.manifest.path));
      }
      extras.push(...buildExecutionSearchPaths(manifestContext.manifest, manifestContext.lock));
    } catch (error) {
      console.error(extractErrorMessage(error));
      return null;
    }
  }

  probeFrom.add(process.cwd());
  probeFrom.add(path.dirname(fileURLToPath(import.meta.url)));
  probeFrom.add(path.dirname(process.execPath));

  return collectModuleSearchPaths({
    cwd: process.cwd(),
    ablePathEnv: env.ablePathEnv,
    ableModulePathsEnv: env.ableModulePathsEnv,
    extras,
    probeFrom: [...probeFrom],
  });
}

function mergeTestModules(programs: Program[]): Program["modules"] {
  const modules: Program["modules"] = [];
  const seen = new Set<string>();
  for (const program of programs) {
    for (const mod of program.modules) {
      if (seen.has(mod.packageName)) {
        continue;
      }
      seen.add(mod.packageName);
      modules.push(mod);
    }
  }
  return modules;
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

function coerceToString(interpreter: V11.Interpreter, value: V11.RuntimeValue | undefined): string {
  if (!value) return "";
  if (value.kind === "String") return value.value;
  try {
    return interpreter.valueToString(value);
  } catch {
    return String(value);
  }
}

function extractFailure(interpreter: V11.Interpreter, value: V11.RuntimeValue): HarnessFailure | null {
  if (!value || value.kind !== "struct_instance") {
    return null;
  }
  if (value.def.id.name !== "Failure") {
    return null;
  }
  const message = coerceToString(interpreter, structField(value, "message"));
  const detailsValue = structField(value, "details");
  const details = detailsValue.kind === "nil" ? null : coerceToString(interpreter, detailsValue);
  return { message, details };
}

function isKernelArray(value: V11.RuntimeValue): value is Extract<V11.RuntimeValue, { kind: "struct_instance" }> {
  return value.kind === "struct_instance" && value.def.id.name === "Array";
}

function formatFailure(failure: HarnessFailure): string {
  const detail = failure.details ? ` (${failure.details})` : "";
  return `${failure.message}${detail}`;
}

function getStructDef(
  interpreter: V11.Interpreter,
  packageName: string,
  name: string,
): AST.StructDefinition {
  const pkg = interpreter.packageRegistry.get(packageName);
  if (!pkg) {
    throw new Error(`unable to locate package ${packageName}`);
  }
  const value = pkg.get(name);
  if (!value || value.kind !== "struct_def") {
    throw new Error(`unable to locate ${name} in ${packageName}`);
  }
  return value.def;
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

function makeStringArray(interpreter: V11.Interpreter, values: string[]): V11.RuntimeValue {
  return interpreter.makeArrayValue(
    values.map((value) => ({ kind: "String", value })),
  );
}

function resolveModulePackageName(module: AST.Module | undefined | null): string {
  const segments =
    module?.package?.namePath
      ?.map((identifier) => identifier?.name)
      .filter((segment): segment is string => Boolean(segment)) ?? [];
  if (segments.length === 0) {
    return "<anonymous>";
  }
  return segments.join(".");
}

function isStdlibPackage(packageName: string | undefined | null): boolean {
  if (!packageName) return false;
  return packageName === "able" || packageName.startsWith("able.");
}

function isStdlibDiagnostic(diag: TypecheckerDiagnostic): boolean {
  const rawPath = diag.location?.path;
  if (!rawPath) return false;
  const relative = normalizeRepoRelativePath(rawPath);
  if (relative.startsWith("..")) return false;
  return relative.startsWith("v11/stdlib/") || relative.startsWith("v11/kernel/");
}

function emitDiagnostics(diags: ModuleDiagnosticEntry[]): void {
  const seen = new Set<string>();
  for (const entry of diags) {
    const formatted = formatTypecheckerDiagnostic(entry.diagnostic, {
      packageName: entry.packageName,
      absolutePath: true,
    });
    if (seen.has(formatted)) {
      continue;
    }
    seen.add(formatted);
    console.error(formatted);
  }
}

function formatRuntimeFailure(error: unknown): string {
  return formatRuntimeDiagnostic(buildRuntimeDiagnostic(error), { absolutePath: true });
}

function extractErrorMessage(err: unknown): string {
  if (!err) return "";
  if (typeof err === "string") return err;
  if (err instanceof Error) {
    if (process.env.ABLE_TRACE_ERRORS) {
      const anyErr = err as any;
      const stack = err.stack ?? err.message;
      if (anyErr.value && typeof anyErr.value === "object" && "message" in anyErr.value) {
        return `${stack}\nRaised: ${String(anyErr.value.message)}`;
      }
      return stack;
    }
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
