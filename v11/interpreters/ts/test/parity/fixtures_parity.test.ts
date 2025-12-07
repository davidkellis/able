import { describe, test, afterAll } from "bun:test";
import path from "node:path";

import {
  collectFixtures,
  readManifest,
} from "../../scripts/fixture-utils";
import {
  DEFAULT_FIXTURE_ROOT,
  buildGoFixtureRunner,
  compareFixtureOutcomes,
  evaluateFixtureGo,
  evaluateFixtureTS,
  formatFixtureDiff,
  getMaxFixturesOverride,
  shouldSkipFixture,
} from "../../scripts/parity/fixtures";

const FIXTURE_ROOT = DEFAULT_FIXTURE_ROOT;
const MAX_FIXTURES = getMaxFixturesOverride(undefined);

const goFixtureRunnerPromise = buildGoFixtureRunner();

afterAll(async () => {
  const runner = await goFixtureRunnerPromise;
  await runner.cleanup().catch(() => {});
});

describe("fixture parity", async () => {
  const fixtures = await collectFixtures(FIXTURE_ROOT);
  const selected =
    typeof MAX_FIXTURES === "number"
      ? fixtures.slice(0, Math.min(MAX_FIXTURES, fixtures.length))
      : fixtures;
  const goFixtureRunner = await goFixtureRunnerPromise;

  for (const fixtureDir of selected) {
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
      const tsOutcome = await evaluateFixtureTS(fixtureDir, manifest ?? null, entry);
      const goOutcome = await evaluateFixtureGo(goFixtureRunner, fixtureDir, entry);
      const diff = compareFixtureOutcomes(tsOutcome, goOutcome, relativeName, manifest ?? null);
      if (diff) {
        throw new Error(formatFixtureDiff(diff));
      }
    }, { timeout: 20000 });
  }
});
