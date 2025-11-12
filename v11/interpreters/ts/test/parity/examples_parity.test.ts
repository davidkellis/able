import { describe, test, afterAll } from "bun:test";
import path from "node:path";

import {
  DEFAULT_EXAMPLES_ROOT,
  buildGoAbleRunner,
  collectExamples,
  compareExampleOutcomes,
  evaluateExampleGo,
  evaluateExampleTS,
  formatExampleDiff,
} from "../../scripts/parity/examples";

const EXAMPLES_ROOT = DEFAULT_EXAMPLES_ROOT;
const EXTRA_MODULE_ROOT = path.resolve(EXAMPLES_ROOT, "deps/vendor");
appendEnvPath("ABLE_MODULE_PATHS", EXTRA_MODULE_ROOT);

const goAbleRunnerPromise = buildGoAbleRunner();

afterAll(async () => {
  const runner = await goAbleRunnerPromise;
  await runner.cleanup().catch(() => {});
});

describe("examples parity (testdata)", async () => {
  const examples = await collectExamples(EXAMPLES_ROOT);
  if (examples.length === 0) {
    test("no examples found", () => {
      throw new Error("expected at least one example under interpreters/ts/testdata/examples");
    });
    return;
  }

  for (const entryPath of examples) {
    const relative = path.relative(EXAMPLES_ROOT, entryPath).split(path.sep).join("/");
    test(relative, async () => {
      const runner = await goAbleRunnerPromise;
      const tsOutcome = await evaluateExampleTS(entryPath);
      const goOutcome = await evaluateExampleGo(entryPath, runner);
      const diff = compareExampleOutcomes(relative, tsOutcome, goOutcome);
      if (diff) {
        throw new Error(formatExampleDiff(diff));
      }
    }, { timeout: 20000 });
  }
});

function appendEnvPath(name: string, extra: string): void {
  if (!extra) return;
  const resolved = path.resolve(extra);
  const current = process.env[name] ?? "";
  const parts = current.split(path.delimiter).filter((segment) => segment.length > 0);
  if (parts.some((segment) => path.resolve(segment) === resolved)) {
    return;
  }
  parts.push(resolved);
  process.env[name] = parts.join(path.delimiter);
}
