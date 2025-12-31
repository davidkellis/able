#!/usr/bin/env bun
import fs from "node:fs";
import { promises as fsPromises } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST, TypeChecker, V11 } from "../index";
import type { PackageSummary, TypecheckerDiagnostic } from "../src/typechecker/diagnostics";
import { ensureConsolePrint, installRuntimeStubs } from "./runtime-stubs";
import { formatTypecheckerDiagnostic, printPackageSummaries } from "./typecheck-utils";
import { resolveTypecheckMode, type TypecheckMode } from "./typecheck-mode";
import { ModuleLoader, type Program } from "./module-loader";
import { callCallableValue } from "../src/interpreter/functions";
import { numericToNumber } from "../src/interpreter/numeric";
import { collectModuleSearchPaths, type ModuleSearchPath } from "./module-search-paths";
import { buildExecutionSearchPaths, loadManifestContext } from "./module-deps";
import { ExitSignal } from "../src/interpreter/signals";
import {
  buildDiscoveryRequest,
  buildRunOptions,
  buildTestPlan,
  callHarnessDiscover,
  callHarnessRun,
  collectTestFiles,
  evaluateTestModules,
  loadTestPrograms,
  maybeTypecheckTestModules,
  resolveTestTargets,
  type TestCliConfig,
  type TestReporterFormat,
  type TestRunOptions,
  type TestCliFilters,
} from "./test-cli";
import {
  createTestReporter,
  emitTestPlanList,
  type TestEventState,
} from "./test-cli-reporters";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CLI_VERSION = process.env.ABLE_TS_VERSION ?? "able-ts dev";
const rawTypecheckMode = process.env.ABLE_TYPECHECK_FIXTURES;
const TYPECHECK_MODE = resolveTypecheckMode(rawTypecheckMode !== undefined ? rawTypecheckMode : "strict");
const ABLE_PATH_ENV = process.env.ABLE_PATH ?? "";
const ABLE_MODULE_PATHS_ENV = process.env.ABLE_MODULE_PATHS ?? "";

type CLICommand = "run" | "check" | "test";

type ModuleDiagnosticEntry = {
  packageName: string;
  diagnostic: TypecheckerDiagnostic;
};

async function main() {
  const argv = process.argv.slice(2);
  if (argv.length === 0) {
    printUsage();
    process.exitCode = 1;
    return;
  }

  const first = argv[0];
  if (isHelpFlag(first)) {
    printUsage();
    return;
  }
  if (isVersionFlag(first)) {
    printVersion();
    return;
  }

  const { command, args } = extractCommand(argv);
  switch (command) {
    case "run":
      await handleRunCommand(args);
      return;
    case "check":
      await handleCheckCommand(args);
      return;
    case "test":
      await handleTestCommand(args);
      return;
    default:
      printUsage();
      process.exitCode = 1;
  }
}

function extractCommand(argv: string[]): { command: CLICommand; args: string[] } {
  const candidate = argv[0] ?? "";
  if (isCommand(candidate)) {
    return { command: candidate, args: argv.slice(1) };
  }
  return { command: "run", args: argv };
}

function isCommand(value: string | undefined): value is CLICommand {
  return value === "run" || value === "check" || value === "test";
}

function isHelpFlag(value: string | undefined): boolean {
  return value === "--help" || value === "-h" || value === "help";
}

function isVersionFlag(value: string | undefined): boolean {
  return value === "--version" || value === "-V" || value === "version";
}

async function handleRunCommand(args: string[]): Promise<void> {
  const entry = args[0];
  if (!entry) {
    console.error("able run requires a path to an Able module (file or directory containing main.able)");
    process.exitCode = 1;
    return;
  }
  const programArgs = args.slice(1);
  const entryPath = await resolveEntryPath(entry);
  if (!entryPath) {
    process.exitCode = 1;
    return;
  }
  const program = await loadProgram(entryPath);
  if (!program) {
    process.exitCode = 1;
    return;
  }
  const session = new TypeChecker.TypecheckerSession();
  const ok = maybeTypecheckProgram(session, program.modules, "run");
  if (!ok) {
    return;
  }

  const interpreter = new V11.Interpreter({ args: programArgs });
  ensureConsolePrint(interpreter);
  installRuntimeStubs(interpreter);

  const evaluated = await evaluateProgram(interpreter, program.modules);
  if (!evaluated) {
    return;
  }

  await invokeEntryMain(interpreter, program.entry);
}

async function invokeEntryMain(interpreter: V11.Interpreter, entry: Program["entry"]): Promise<void> {
  const packageBucket = interpreter.packageRegistry.get(entry.packageName);
  if (!packageBucket) {
    console.error(`runtime error: entry package '${entry.packageName}' is not available at runtime`);
    process.exitCode = 1;
    return;
  }
  const mainValue = packageBucket.get("main");
  if (!mainValue) {
    console.error("entry module does not define a main function");
    process.exitCode = 1;
    return;
  }

  try {
    const callNode = AST.functionCall(AST.identifier("main"), []);
    callCallableValue(interpreter as any, mainValue, [], interpreter.globals, callNode);
  } catch (error) {
    if (error instanceof ExitSignal) {
      process.exitCode = error.code;
      return;
    }
    console.error(`runtime error: ${extractErrorMessage(error)}`);
    process.exitCode = 1;
  }
}

async function handleCheckCommand(args: string[]): Promise<void> {
  const entry = args[0];
  if (!entry) {
    console.error("able check requires a path to an Able module");
    process.exitCode = 1;
    return;
  }
  const entryPath = await resolveEntryPath(entry);
  if (!entryPath) {
    process.exitCode = 1;
    return;
  }
  const program = await loadProgram(entryPath);
  if (!program) {
    process.exitCode = 1;
    return;
  }
  const session = new TypeChecker.TypecheckerSession();
  const ok = maybeTypecheckProgram(session, program.modules, "check");
  if (!ok) {
    return;
  }
  console.log("typecheck: ok");
}

async function handleTestCommand(args: string[]): Promise<void> {
  let config: TestCliConfig;
  try {
    config = parseTestArguments(args);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    process.exitCode = 1;
    return;
  }

  let targets: string[];
  try {
    targets = await resolveTestTargets(config.targets);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    process.exitCode = 1;
    return;
  }

  let testFiles: string[];
  try {
    testFiles = await collectTestFiles(targets);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    process.exitCode = 1;
    return;
  }

  if (testFiles.length === 0) {
    console.log("able test: no test modules found");
    process.exitCode = 0;
    return;
  }

  const loadResult = await loadTestPrograms(testFiles, {
    ablePathEnv: ABLE_PATH_ENV,
    ableModulePathsEnv: ABLE_MODULE_PATHS_ENV,
  });
  if (!loadResult) {
    process.exitCode = 2;
    return;
  }

  const session = new TypeChecker.TypecheckerSession();
  const typecheckOk = maybeTypecheckTestModules(session, loadResult.modules, TYPECHECK_MODE);
  if (!typecheckOk) {
    return;
  }

  const interpreter = new V11.Interpreter();
  ensureConsolePrint(interpreter);
  installRuntimeStubs(interpreter);

  const evaluated = evaluateTestModules(interpreter, loadResult.modules);
  if (!evaluated) {
    return;
  }

  const discoveryRequest = buildDiscoveryRequest(interpreter, config);
  const discoveryResult = callHarnessDiscover(interpreter, discoveryRequest);
  if (!discoveryResult) {
    process.exitCode = 2;
    return;
  }

  if (config.listOnly || config.dryRun) {
    emitTestPlanList(interpreter, discoveryResult, config);
    process.exitCode = 0;
    return;
  }

  if (arrayLength(interpreter, discoveryResult) === 0) {
    console.log("able test: no tests to run");
    process.exitCode = 0;
    return;
  }

  const eventState: TestEventState = {
    total: 0,
    failed: 0,
    skipped: 0,
    frameworkErrors: 0,
  };

  const reporter = await createTestReporter(interpreter, config.reporterFormat, eventState);
  if (!reporter) {
    process.exitCode = 2;
    return;
  }

  let runOptions: V11.RuntimeValue;
  let testPlan: V11.RuntimeValue;
  try {
    runOptions = buildRunOptions(interpreter, config.run);
    testPlan = buildTestPlan(interpreter, discoveryResult);
  } catch (error) {
    console.error(`able test: ${extractErrorMessage(error)}`);
    process.exitCode = 2;
    return;
  }
  const runResult = callHarnessRun(interpreter, testPlan, runOptions, reporter.reporter);
  if (runResult) {
    process.exitCode = 2;
    return;
  }

  if (reporter.finish) {
    reporter.finish();
  }

  if (eventState.frameworkErrors > 0) {
    process.exitCode = 2;
    return;
  }
  if (eventState.failed > 0) {
    process.exitCode = 1;
    return;
  }
  process.exitCode = 0;
}

function maybeTypecheckProgram(
  session: TypeChecker.TypecheckerSession,
  modules: Program["modules"],
  command: "run" | "check",
): boolean {
  if (TYPECHECK_MODE === "off") {
    return true;
  }
  const diagnostics: ModuleDiagnosticEntry[] = [];
  for (const mod of modules) {
    const result = session.checkModule(mod.module);
    if (result.diagnostics.length === 0) {
      continue;
    }
    const packageName =
      result.summary?.name ?? mod.packageName ?? resolveModulePackageName(mod.module);
    for (const diag of result.diagnostics) {
      diagnostics.push({ packageName, diagnostic: diag });
    }
  }
  if (diagnostics.length === 0) {
    return true;
  }
  emitDiagnostics(diagnostics);
  printPackageSummaries(session.getPackageSummaries());

  if (command === "check") {
    process.exitCode = 1;
    return false;
  }
  if (TYPECHECK_MODE === "strict") {
    process.exitCode = 1;
    return false;
  }

  console.warn("typechecker: proceeding despite diagnostics because ABLE_TYPECHECK_FIXTURES=warn");
  return true;
}

async function loadProgram(entryPath: string): Promise<Program | null> {
  const searchPaths = await resolveSearchPaths(entryPath);
  if (!searchPaths) {
    return null;
  }
  const loader = new ModuleLoader(searchPaths);
  try {
    return await loader.load(entryPath);
  } catch (error) {
    console.error(`failed to load program: ${extractErrorMessage(error)}`);
    return null;
  }
}

async function resolveEntryPath(input: string): Promise<string | null> {
  const candidate = path.resolve(process.cwd(), input);
  try {
    const stats = await fsPromises.stat(candidate);
    if (stats.isDirectory()) {
      const mainPath = path.join(candidate, "main.able");
      try {
        await fsPromises.access(mainPath);
        return mainPath;
      } catch {
        console.error(`unable to locate main.able inside ${candidate}`);
        return null;
      }
    }
    return candidate;
  } catch (error) {
    console.error(`unable to access ${candidate}: ${(error as Error).message}`);
    return null;
  }
}

async function evaluateProgram(interpreter: V11.Interpreter, modules: Program["modules"]): Promise<boolean> {
  for (const mod of modules) {
    try {
      interpreter.evaluate(mod.module);
    } catch (error) {
      if (error instanceof ExitSignal) {
        process.exitCode = error.code;
        return false;
      }
      console.error(`runtime error: ${extractErrorMessage(error)}`);
      process.exitCode = 1;
      return false;
    }
  }
  return true;
}

function fsExistsSync(target: string): boolean {
  try {
    fs.statSync(target);
    return true;
  } catch {
    return false;
  }
}

async function resolveSearchPaths(entryPath: string): Promise<ModuleSearchPath[] | null> {
  const entryDir = path.dirname(path.resolve(entryPath));
  let manifestRoot: string | null = null;
  let extras: ModuleSearchPath[] = [];
  try {
    const manifestContext = await loadManifestContext(entryDir);
    manifestRoot = manifestContext.manifest?.path ? path.dirname(manifestContext.manifest.path) : null;
    extras = buildExecutionSearchPaths(manifestContext.manifest, manifestContext.lock);
  } catch (error) {
    console.error(extractErrorMessage(error));
    return null;
  }
  const probeFrom = [
    entryDir,
    manifestRoot ?? undefined,
    process.cwd(),
    path.dirname(fileURLToPath(import.meta.url)),
    path.dirname(process.execPath),
  ].filter(Boolean) as string[];

  return collectModuleSearchPaths({
    cwd: entryDir ?? process.cwd(),
    ablePathEnv: ABLE_PATH_ENV,
    ableModulePathsEnv: ABLE_MODULE_PATHS_ENV,
    extras,
    probeFrom,
  });
}

function parseTestArguments(args: string[]): TestCliConfig {
  const filters: TestCliFilters = {
    includePaths: [],
    excludePaths: [],
    includeNames: [],
    excludeNames: [],
    includeTags: [],
    excludeTags: [],
  };
  const run: TestRunOptions = {
    failFast: false,
    repeat: 1,
    parallelism: 1,
  };
  let reporterFormat: TestReporterFormat = "doc";
  let listOnly = false;
  let dryRun = false;
  const targets: string[] = [];
  let shuffleSeed: number | undefined;

  for (let i = 0; i < args.length; i += 1) {
    const arg = args[i]!;
    switch (arg) {
      case "--list":
        listOnly = true;
        break;
      case "--dry-run":
        dryRun = true;
        listOnly = true;
        break;
      case "--path":
        filters.includePaths.push(expectFlagValue(arg, args[++i]));
        break;
      case "--exclude-path":
        filters.excludePaths.push(expectFlagValue(arg, args[++i]));
        break;
      case "--name":
        filters.includeNames.push(expectFlagValue(arg, args[++i]));
        break;
      case "--exclude-name":
        filters.excludeNames.push(expectFlagValue(arg, args[++i]));
        break;
      case "--tag":
        filters.includeTags.push(expectFlagValue(arg, args[++i]));
        break;
      case "--exclude-tag":
        filters.excludeTags.push(expectFlagValue(arg, args[++i]));
        break;
      case "--format":
        reporterFormat = parseReporterFormat(expectFlagValue(arg, args[++i]));
        break;
      case "--fail-fast":
        run.failFast = true;
        break;
      case "--repeat":
        run.repeat = parsePositiveInteger(expectFlagValue(arg, args[++i]), arg, 1);
        break;
      case "--parallel":
        run.parallelism = parsePositiveInteger(expectFlagValue(arg, args[++i]), arg, 1);
        break;
      case "--shuffle": {
        const maybeSeed = args[i + 1];
        if (maybeSeed && !maybeSeed.startsWith("-")) {
          shuffleSeed = parsePositiveInteger(maybeSeed, arg, 0);
          i += 1;
        } else {
          shuffleSeed = generateShuffleSeed();
        }
        break;
      }
      default:
        if (arg.startsWith("-")) {
          throw new Error(`unknown able test flag '${arg}'`);
        }
        targets.push(arg);
        break;
    }
  }

  run.shuffleSeed = shuffleSeed;

  return {
    targets,
    filters,
    run,
    reporterFormat,
    listOnly,
    dryRun,
  };
}

function expectFlagValue(flag: string, value: string | undefined): string {
  if (!value || value.startsWith("-")) {
    throw new Error(`${flag} expects a value`);
  }
  return value;
}

function parseReporterFormat(value: string): TestReporterFormat {
  if (value === "doc" || value === "progress" || value === "tap" || value === "json") {
    return value;
  }
  throw new Error(`unknown --format value '${value}' (expected doc, progress, tap, or json)`);
}

function parsePositiveInteger(value: string, flag: string, min: number): number {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || parsed < min) {
    throw new Error(`${flag} expects an integer >= ${min}`);
  }
  return parsed;
}

function arrayLength(interpreter: V11.Interpreter, value: V11.RuntimeValue): number {
  if (!value) return 0;
  if (value.kind === "array") {
    return value.elements.length;
  }
  if (value.kind === "struct_instance" && value.def.id.name === "Array") {
    const lengthValue = structField(value, "length");
    return numericToNumber(lengthValue, "array length", { requireSafeInteger: true });
  }
  return 0;
}

function structField(
  value: Extract<V11.RuntimeValue, { kind: "struct_instance" }>,
  field: string,
): V11.RuntimeValue {
  if (value.values instanceof Map) {
    return value.values.get(field) ?? { kind: "nil", value: null };
  }
  const fieldIndex = value.def.fields.findIndex((entry) => entry.name?.name === field);
  if (fieldIndex >= 0 && Array.isArray(value.values)) {
    return value.values[fieldIndex] ?? { kind: "nil", value: null };
  }
  return { kind: "nil", value: null };
}

function generateShuffleSeed(): number {
  const now = Date.now();
  return Number(now.toString().slice(-9));
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

function printUsage(): void {
  const script = path.relative(process.cwd(), path.join(__dirname, "run-module.ts"));
  console.log(`Able CLI (Bun prototype)

Usage:
  bun run ${script} run <path>      Typecheck and execute an Able module (default command)
  bun run ${script} check <path>    Typecheck without executing
  bun run ${script} test [paths]    Discover and run Able tests

Options:
  --help, -h        Show this message
  --version, -V     Print CLI version
  --list            Discover tests and print plan
  --dry-run         Discover tests and skip execution (implies --list)
  --path <value>    Include tests whose module path contains value (repeatable)
  --exclude-path <value>
  --name <value>    Include tests whose name contains value (repeatable)
  --exclude-name <value>
  --tag <value>     Include tests with tag (repeatable)
  --exclude-tag <value>
  --format <fmt>    Reporter format: doc, progress, tap, json
  --shuffle [seed]  Shuffle execution order with optional seed
  --fail-fast       Stop after first failure
  --parallel <n>    Requested parallelism (framework may ignore)
  --repeat <count>  Repeat entire plan (default 1)

Environment:
  ABLE_TYPECHECK_FIXTURES=warn|strict|off   Controls typecheck enforcement (default strict for CLI)`);
}

function printVersion(): void {
  console.log(CLI_VERSION);
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

await main();
