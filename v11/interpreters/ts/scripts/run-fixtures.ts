import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import type { AST } from "../index";
import { TypeChecker, V10 } from "../index";
import type {
  PackageSummary,
  TypecheckerDiagnostic,
} from "../src/typechecker/diagnostics";
import { formatTypecheckerDiagnostic, printPackageSummaries } from "./typecheck-utils";
import { resolveTypecheckMode, type TypecheckMode } from "./typecheck-mode";
import { ModuleLoader, type Program } from "./module-loader";
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
  type Manifest,
} from "./fixture-utils";
import { startRunTimeout } from "./test-timeouts";
import { serializeMapEntries } from "../src/interpreter/maps";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../../fixtures/ast");
const TYPECHECK_MODE = resolveTypecheckMode(process.env.ABLE_TYPECHECK_FIXTURES);
const TYPECHECK_BASELINE_PATH = path.join(FIXTURE_ROOT, "typecheck-baseline.json");
const WRITE_TYPECHECK_BASELINE = process.argv.includes("--write-typecheck-baseline");
const FIXTURE_FILTER = process.env.ABLE_FIXTURE_FILTER?.trim() ?? null;
const BASELINE_ENABLED = TYPECHECK_MODE !== "off" && !FIXTURE_FILTER;
const STDLIB_ROOT = path.resolve(__dirname, "../../../stdlib/src");
const STDLIB_STRING_ENTRY = path.join(STDLIB_ROOT, "text", "string.able");
const STDLIB_CONCURRENCY_ENTRY = path.join(STDLIB_ROOT, "concurrency", "await.able");
const stdlibLoader = new ModuleLoader([STDLIB_ROOT]);
let stdlibStringProgram: Program | null = null;
let stdlibConcurrencyProgram: Program | null = null;

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
      const interpreter = new V10.InterpreterV10();
      ensurePrint(interpreter);
      installRuntimeStubs(interpreter);

      const stdout: string[] = [];
      let evaluationError: unknown;

      const entry = manifest.entry ?? "module.json";
      const moduleAst = await loadModuleFromFixture(fixtureDir, entry);
      const needsStdlibString = moduleImportsStdlibString(moduleAst);
      const needsStdlibConcurrency = moduleImportsStdlibConcurrency(moduleAst);
      const stdlibPrograms: Program[] = [];
      if (needsStdlibString) {
        stdlibPrograms.push(await ensureStdlibStringProgram());
      }
      if (needsStdlibConcurrency) {
        stdlibPrograms.push(await ensureStdlibConcurrencyProgram());
      }
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
        const seenPkgs = new Set<string>();
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

      let result: V10.V10Value | undefined;
      interceptStdout(stdout, () => {
        try {
          for (const stdlibProgram of stdlibPrograms) {
            evaluateProgram(interpreter, stdlibProgram);
          }
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
    console.log(`Executed ${results.length} fixture(s).`);
  } finally {
    clearTimeouts();
  }
}

function moduleImportsStdlibString(module: AST.Module): boolean {
  return moduleImportsPackage(module, "able.text.String") || moduleImportsPackage(module, "able.text.string");
}

function moduleImportsStdlibConcurrency(module: AST.Module): boolean {
  return moduleImportsPackage(module, "able.concurrency");
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

function evaluateProgram(interpreter: V10.InterpreterV10, program: Program): void {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    interpreter.evaluate(mod.module);
  }
  interpreter.evaluate(program.entry.module);
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
      if (result.kind !== "hash_map") {
        throw new Error(`Fixture ${dir} expected hash_map entries but result was ${result.kind}`);
      }
      compareMapEntries(dir, expect.result.entries, result);
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

function compareMapEntries(dir: string, expected: { key: unknown; value: unknown }[], value: V10.V10Value): void {
  const entries = serializeMapEntries(value as Extract<V10.V10Value, { kind: "hash_map" }>);
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

function normalizeValueForExpect(value: V10.V10Value): unknown {
  switch (value.kind) {
    case "String":
    case "char":
    case "bool":
    case "i32":
    case "f64":
      return { kind: value.kind, value: value.value };
    case "nil":
      return { kind: "nil", value: null };
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
