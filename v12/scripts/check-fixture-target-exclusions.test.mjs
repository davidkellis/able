import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { validateFixtureTargetExclusions } from "./check-fixture-target-exclusions.mjs";

const tempRoots = [];

test.afterEach(async () => {
  await Promise.all(
    tempRoots.splice(0).map((root) =>
      fs.rm(root, { force: true, recursive: true }),
    ),
  );
});

test("repository policy accepts retired metadata and has no active exclusions", async () => {
  const result = await validateFixtureTargetExclusions();
  assert.deepEqual(result.errors, []);
  assert.equal(result.summary.activeExclusionCount, 0);
  assert.equal(result.summary.retiredExclusionCount, 1);
  assert.equal(result.summary.allowlistCount, 0);
});

test("rejects an unclassified active target exclusion", async () => {
  const fixture = await fixtureTree({
    manifest: { skipTargets: ["go"] },
  });
  const result = await validateFixtureTargetExclusions(fixture);
  assert.deepEqual(result.errors, [
    "sample: active target go exclusion is not classified and allowlisted",
  ]);
});

test("accepts an active exclusion with classification and reason", async () => {
  const fixture = await fixtureTree({
    manifest: { skipTargets: ["go"] },
    allowlist: [
      {
        fixture: "sample",
        target: "go",
        classification: "known-correctness-gap",
        reason: "Reproduced parity defect with a tracked follow-up.",
      },
    ],
  });
  const result = await validateFixtureTargetExclusions(fixture);
  assert.deepEqual(result.errors, []);
  assert.equal(result.summary.activeExclusionCount, 1);
  assert.equal(result.summary.allowlistCount, 1);
});

test("rejects stale allowlist entries", async () => {
  const fixture = await fixtureTree({
    manifest: {},
    allowlist: [
      {
        fixture: "sample",
        target: "go",
        classification: "toolchain-limitation",
        reason: "Temporary toolchain constraint.",
      },
    ],
  });
  const result = await validateFixtureTargetExclusions(fixture);
  assert.deepEqual(result.errors, ["stale allowlist entry: sample (go)"]);
});

test("retired target metadata does not require an allowlist entry", async () => {
  const fixture = await fixtureTree({
    manifest: { skipTargets: ["ts"] },
  });
  const result = await validateFixtureTargetExclusions(fixture);
  assert.deepEqual(result.errors, []);
  assert.equal(result.summary.retiredExclusionCount, 1);
});

test("rejects duplicate and unknown skip targets", async () => {
  const fixture = await fixtureTree({
    manifest: { skipTargets: ["go", "go", "lua"] },
  });
  const result = await validateFixtureTargetExclusions(fixture);
  assert.deepEqual(result.errors, [
    "sample: active target go exclusion is not classified and allowlisted",
    "sample: duplicate skipTargets entry go",
    "sample: unknown skip target lua",
  ]);
});

test("rejects invalid allowlist classifications and empty reasons", async () => {
  const fixture = await fixtureTree({
    manifest: { skipTargets: ["go"] },
    allowlist: [
      {
        fixture: "sample",
        target: "go",
        classification: "benchmark-specific-exception",
        reason: "",
      },
    ],
  });
  const result = await validateFixtureTargetExclusions(fixture);
  assert.deepEqual(result.errors, [
    "allowlist entry sample has invalid classification benchmark-specific-exception",
    "allowlist entry sample needs a non-empty reason",
  ]);
});

async function fixtureTree({ manifest, allowlist = [] }) {
  const root = await fs.mkdtemp(
    path.join(os.tmpdir(), "able-fixture-target-exclusions-"),
  );
  tempRoots.push(root);
  const fixtureRoot = path.join(root, "fixtures");
  const sampleRoot = path.join(fixtureRoot, "sample");
  const policyPath = path.join(root, "policy.json");
  await fs.mkdir(sampleRoot, { recursive: true });
  await writeJson(path.join(sampleRoot, "manifest.json"), manifest);
  await writeJson(policyPath, {
    schema_version: 1,
    active_targets: ["go"],
    retired_targets: ["ts"],
    allowed_classifications: [
      "intentional-contract-exclusion",
      "known-correctness-gap",
      "platform-limitation",
      "toolchain-limitation",
    ],
    allowlist,
  });
  return { fixtureRoot, policyPath };
}

async function writeJson(filePath, value) {
  await fs.writeFile(filePath, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}
