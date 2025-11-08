import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectFixtures, readManifest } from "./fixture-utils";
import {
  DEFAULT_FIXTURE_ROOT,
  buildGoFixtureRunner,
  compareFixtureOutcomes,
  evaluateFixtureGo,
  evaluateFixtureTS,
  formatFixtureDiff,
  getMaxFixturesOverride,
  shouldSkipFixture,
  type FixtureParityDiff,
  type GoFixtureRunner,
  type GoOutcome as FixtureGoOutcome,
  type TSOutcome as FixtureTSOutcome,
} from "./parity/fixtures";
import {
  DEFAULT_EXAMPLES_ROOT,
  buildGoAbleRunner,
  collectExamples as collectExampleEntries,
  compareExampleOutcomes,
  evaluateExampleGo,
  evaluateExampleTS,
  formatExampleDiff,
  type ExampleParityDiff,
  type GoAbleRunner,
  type GoExampleOutcome,
  type TSExampleOutcome,
} from "./parity/examples";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../");
const DEFAULT_REPORT_PATH = path.join(REPO_ROOT, "tmp", "parity-report.json");

type SuiteName = "fixtures" | "examples";

type SuiteEntry<TDiff, TSOutcome, GoOutcome> = {
  name: string;
  status: "ok" | "failed" | "skipped";
  reason?: string;
  diff?: TDiff;
  tsOutcome?: TSOutcome;
  goOutcome?: GoOutcome;
};

type SuiteReport<TSuite extends SuiteName, TDiff, TSOutcome, GoOutcome> = {
  suite: TSuite;
  root: string;
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  entries: Array<SuiteEntry<TDiff, TSOutcome, GoOutcome>>;
};

type FixtureSuiteReport = SuiteReport<
  "fixtures",
  FixtureParityDiff,
  FixtureTSOutcome,
  FixtureGoOutcome
>;
type ExampleSuiteReport = SuiteReport<
  "examples",
  ExampleParityDiff,
  TSExampleOutcome,
  GoExampleOutcome
>;

type ParityReport = {
  suites: Array<FixtureSuiteReport | ExampleSuiteReport>;
};

async function writeReport(report: ParityReport, requestedPath?: string): Promise<string> {
  const targetPath = requestedPath ?? DEFAULT_REPORT_PATH;
  await fs.mkdir(path.dirname(targetPath), { recursive: true });
  await fs.writeFile(targetPath, `${JSON.stringify(report, null, 2)}\n`, "utf8");
  return targetPath;
}

type CLIOptions = {
  suites: SuiteName[];
  fixturesRoot: string;
  examplesRoot: string;
  maxFixtures?: number;
  jsonOutput: boolean;
  reportPath?: string;
};

async function copyReportArtifacts(sourcePath: string, silent: boolean): Promise<void> {
  const targets = collectReportTargets();
  for (const target of targets) {
    if (path.resolve(target) === path.resolve(sourcePath)) {
      continue;
    }
    try {
      await fs.mkdir(path.dirname(target), { recursive: true });
      await fs.copyFile(sourcePath, target);
      if (!silent) {
        console.log(`Parity report copied to ${relativeFromRepo(target)}`);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      console.error(`Failed to copy parity report to ${target}: ${message}`);
    }
  }
}

function collectReportTargets(): string[] {
  const targets = new Set<string>();
  const explicit = process.env.ABLE_PARITY_REPORT_DEST?.trim();
  if (explicit) {
    targets.add(path.resolve(explicit));
  }
  const artifactsDir = process.env.CI_ARTIFACTS_DIR?.trim();
  if (artifactsDir) {
    targets.add(path.resolve(path.join(artifactsDir, "parity-report.json")));
  }
  return [...targets];
}

function relativeFromRepo(candidate: string): string {
  const resolved = path.resolve(candidate);
  if (!resolved.startsWith(REPO_ROOT)) {
    return resolved;
  }
  return path.relative(REPO_ROOT, resolved) || ".";
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const report: ParityReport = { suites: [] };
  const silentLogs = options.jsonOutput;

  if (options.suites.includes("fixtures")) {
    const fixtureSuite = await runFixtureSuite({
      root: options.fixturesRoot,
      maxFixtures: options.maxFixtures,
    });
    report.suites.push(fixtureSuite);
  }

  if (options.suites.includes("examples")) {
    const exampleSuite = await runExampleSuite({
      root: options.examplesRoot,
    });
    report.suites.push(exampleSuite);
  }

  const finalReportPath = await writeReport(report, options.reportPath);
  if (!silentLogs) {
    console.log(`JSON parity report written to ${relativeFromRepo(finalReportPath)}`);
  }
  await copyReportArtifacts(finalReportPath, silentLogs);

  if (options.jsonOutput) {
    console.log(JSON.stringify(report, null, 2));
  } else {
    printHumanReport(report);
  }

  const hasFailures = report.suites.some((suite) => suite.failed > 0);
  process.exitCode = hasFailures ? 1 : 0;
}

function parseArgs(argv: string[]): CLIOptions {
  const suites: SuiteName[] = [];
  let fixturesRoot = DEFAULT_FIXTURE_ROOT;
  let examplesRoot = DEFAULT_EXAMPLES_ROOT;
  let maxFixtures: number | undefined;
  let jsonOutput = false;
  let reportPath: string | undefined;

  const takeValue = (arr: string[], index: number, flag: string): string => {
    const value = arr[index + 1];
    if (!value) {
      throw new Error(`Missing value for ${flag}`);
    }
    arr[index + 1] = "";
    return value;
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (!arg) continue;
    if (arg === "--suite") {
      appendSuiteValue(suites, takeValue(argv, i, "--suite"));
      continue;
    }
    if (arg.startsWith("--suite=")) {
      appendSuiteValue(suites, arg.slice("--suite=".length));
      continue;
    }
    if (arg === "--fixtures-root") {
      fixturesRoot = path.resolve(REPO_ROOT, takeValue(argv, i, "--fixtures-root"));
      continue;
    }
    if (arg.startsWith("--fixtures-root=")) {
      fixturesRoot = path.resolve(REPO_ROOT, arg.slice("--fixtures-root=".length));
      continue;
    }
    if (arg === "--examples-root") {
      examplesRoot = path.resolve(REPO_ROOT, takeValue(argv, i, "--examples-root"));
      continue;
    }
    if (arg.startsWith("--examples-root=")) {
      examplesRoot = path.resolve(REPO_ROOT, arg.slice("--examples-root=".length));
      continue;
    }
    if (arg === "--max-fixtures") {
      maxFixtures = parsePositiveInt(takeValue(argv, i, "--max-fixtures"));
      continue;
    }
    if (arg.startsWith("--max-fixtures=")) {
      maxFixtures = parsePositiveInt(arg.slice("--max-fixtures=".length));
      continue;
    }
    if (arg === "--json") {
      jsonOutput = true;
      continue;
    }
    if (arg === "--report") {
      reportPath = path.resolve(REPO_ROOT, takeValue(argv, i, "--report"));
      continue;
    }
    if (arg.startsWith("--report=")) {
      reportPath = path.resolve(REPO_ROOT, arg.slice("--report=".length));
      continue;
    }
    if (arg === "--help" || arg === "-h") {
      printHelp();
      process.exit(0);
    }
    throw new Error(`Unknown option: ${arg}`);
  }

  const uniqueSuites =
    suites.length > 0 ? dedupeSuites(suites) : (["fixtures", "examples"] as SuiteName[]);

  return {
    suites: uniqueSuites,
    fixturesRoot,
    examplesRoot,
    maxFixtures,
    jsonOutput,
    reportPath,
  };
}

function appendSuiteValue(target: SuiteName[], raw: string): void {
  const resolved = resolveSuite(raw);
  if (resolved === "all") {
    target.push("fixtures", "examples");
  } else {
    target.push(resolved);
  }
}

function resolveSuite(value: string): SuiteName | "all" {
  const normalized = value.trim().toLowerCase();
  if (normalized === "fixtures" || normalized === "examples") {
    return normalized;
  }
  if (normalized === "all") {
    return "all";
  }
  throw new Error(`Unknown suite: ${value}`);
}

function dedupeSuites(values: SuiteName[]): SuiteName[] {
  const set = new Set<SuiteName>();
  for (const value of values) {
    if (value === "fixtures") {
      set.add("fixtures");
    } else if (value === "examples") {
      set.add("examples");
    }
  }
  return [...set];
}

function parsePositiveInt(raw: string): number {
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new Error(`Expected positive integer, received '${raw}'`);
  }
  return parsed;
}

function printHelp(): void {
  console.log(`Usage: bun run scripts/run-parity.ts [options]

Options:
  --suite <fixtures|examples|all>   Select parity suite(s). Repeatable. Default: fixtures + examples.
  --fixtures-root <path>            Override fixtures root (default: ${DEFAULT_FIXTURE_ROOT}).
  --examples-root <path>            Override curated examples root (default: ${DEFAULT_EXAMPLES_ROOT}).
  --max-fixtures <n>                Limit number of fixtures processed (respects env ABLE_PARITY_MAX_FIXTURES when omitted).
  --json                            Emit JSON report to stdout (parity JSON is still written to tmp/parity-report.json).
  --report <path>                   Override the JSON report path (default: tmp/parity-report.json under the repo root).
                                    Reports are also copied when ABLE_PARITY_REPORT_DEST or CI_ARTIFACTS_DIR is set.
  -h, --help                        Show this message.`);
}

type FixtureSuiteOptions = {
  root: string;
  maxFixtures?: number;
};

async function runFixtureSuite(options: FixtureSuiteOptions): Promise<FixtureSuiteReport> {
  const fixtures = await collectFixtures(options.root);
  const limit = getMaxFixturesOverride(options.maxFixtures);
  const selected =
    typeof limit === "number" ? fixtures.slice(0, Math.min(limit, fixtures.length)) : fixtures;

  if (selected.length === 0) {
    return {
      suite: "fixtures",
      root: options.root,
      total: 0,
      passed: 0,
      failed: 0,
      skipped: 0,
      entries: [],
    };
  }

  const runner = await buildGoFixtureRunner();
  const entries: FixtureSuiteReport["entries"] = [];
  let skipped = 0;
  let failed = 0;

  try {
    for (const fixtureDir of selected) {
      const manifest = await readManifest(fixtureDir);
      const relativeName = path.relative(options.root, fixtureDir).split(path.sep).join("/");

      if (shouldSkipFixture(manifest)) {
        skipped += 1;
        entries.push({
          name: relativeName,
          status: "skipped",
          reason: "skipTargets includes ts or go",
        });
        continue;
      }

      const entryFile = manifest?.entry ?? "module.json";
      let tsOutcome: FixtureTSOutcome | undefined;
      let goOutcome: FixtureGoOutcome | undefined;

      try {
        tsOutcome = await evaluateFixtureTS(fixtureDir, manifest, entryFile);
        goOutcome = await evaluateFixtureGo(runner, fixtureDir, entryFile);
      } catch (err) {
        failed += 1;
        entries.push({
          name: relativeName,
          status: "failed",
          reason: err instanceof Error ? err.message : String(err),
        });
        continue;
      }

      const diff = compareFixtureOutcomes(tsOutcome, goOutcome, relativeName, manifest ?? null);
      if (diff) {
        failed += 1;
        entries.push({
          name: relativeName,
          status: "failed",
          reason: formatFixtureDiff(diff),
          diff,
          tsOutcome,
          goOutcome,
        });
      } else {
        entries.push({ name: relativeName, status: "ok" });
      }
    }
  } finally {
    await runner.cleanup().catch(() => {});
  }

  const total = entries.length;
  const passed = total - failed - skipped;

  return {
    suite: "fixtures",
    root: options.root,
    total,
    passed,
    failed,
    skipped,
    entries,
  };
}

type ExampleSuiteOptions = {
  root: string;
};

async function runExampleSuite(options: ExampleSuiteOptions): Promise<ExampleSuiteReport> {
  const entriesPaths = await collectExampleEntries(options.root);
  if (entriesPaths.length === 0) {
    return {
      suite: "examples",
      root: options.root,
      total: 0,
      passed: 0,
      failed: 0,
      skipped: 0,
      entries: [],
    };
  }

  const runner = await buildGoAbleRunner();
  const entries: ExampleSuiteReport["entries"] = [];
  let failed = 0;

  try {
    for (const entryPath of entriesPaths) {
      const relative = path.relative(options.root, entryPath).split(path.sep).join("/");
      let tsOutcome: TSExampleOutcome | undefined;
      let goOutcome: GoExampleOutcome | undefined;
      try {
        tsOutcome = await evaluateExampleTS(entryPath);
        goOutcome = await evaluateExampleGo(entryPath, runner);
      } catch (err) {
        failed += 1;
        entries.push({
          name: relative,
          status: "failed",
          reason: err instanceof Error ? err.message : String(err),
        });
        continue;
      }

      const diff = compareExampleOutcomes(relative, tsOutcome, goOutcome);
      if (diff) {
        failed += 1;
        entries.push({
          name: relative,
          status: "failed",
          reason: formatExampleDiff(diff),
          diff,
          tsOutcome,
          goOutcome,
        });
      } else {
        entries.push({ name: relative, status: "ok" });
      }
    }
  } finally {
    await runner.cleanup().catch(() => {});
  }

  const total = entries.length;
  const skipped = 0;
  const passed = total - failed;

  return {
    suite: "examples",
    root: options.root,
    total,
    passed,
    failed,
    skipped,
    entries,
  };
}

function printHumanReport(report: ParityReport): void {
  for (const suite of report.suites) {
    const title = suite.suite === "fixtures" ? "AST Fixtures" : "Curated Examples";
    console.log(`=== ${title} (${suite.root}) ===`);
    console.log(
      `total=${suite.total} passed=${suite.passed} failed=${suite.failed} skipped=${suite.skipped}`,
    );
    if (suite.failed > 0) {
      for (const entry of suite.entries) {
        if (entry.status !== "failed") continue;
        console.log(`  ✗ ${entry.name}`);
        if (entry.reason) {
          for (const line of entry.reason.split("\n")) {
            console.log(`    ${line}`);
          }
        }
      }
    } else if (suite.passed > 0) {
      console.log("  ✓ all entries passed");
    } else {
      console.log("  (no entries)");
    }
    console.log();
  }
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : err);
  process.exitCode = 1;
});
