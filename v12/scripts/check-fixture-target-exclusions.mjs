#!/usr/bin/env node

// Prevents active runtime coverage from being silently disabled by fixture
// manifest metadata. Retired targets remain readable historical context.

import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const SCRIPT_PATH = fileURLToPath(import.meta.url);
const SCRIPT_DIR = path.dirname(SCRIPT_PATH);
const V12_ROOT = path.resolve(SCRIPT_DIR, "..");
const DEFAULT_FIXTURE_ROOT = path.join(V12_ROOT, "fixtures", "ast");
const DEFAULT_POLICY_PATH = path.join(
  V12_ROOT,
  "fixture-target-exclusion-policy.json",
);

export async function validateFixtureTargetExclusions({
  fixtureRoot = DEFAULT_FIXTURE_ROOT,
  policyPath = DEFAULT_POLICY_PATH,
} = {}) {
  const errors = [];
  const policy = await readJson(policyPath, "target-exclusion policy", errors);
  const parsedPolicy = validatePolicy(policy, errors);
  const manifestPaths = await collectManifestPaths(fixtureRoot, errors);
  const exclusions = [];

  for (const manifestPath of manifestPaths) {
    const relativeDir = normalizeRelative(
      path.relative(fixtureRoot, path.dirname(manifestPath)),
    );
    const manifest = await readJson(
      manifestPath,
      `fixture manifest ${relativeDir}`,
      errors,
    );
    if (!manifest) {
      continue;
    }
    const rawTargets = manifest.skipTargets ?? [];
    if (!Array.isArray(rawTargets)) {
      errors.push(`${relativeDir}: skipTargets must be an array`);
      continue;
    }
    const seen = new Set();
    for (const rawTarget of rawTargets) {
      if (typeof rawTarget !== "string" || rawTarget.trim() === "") {
        errors.push(`${relativeDir}: skipTargets entries must be non-empty strings`);
        continue;
      }
      const target = rawTarget.trim();
      if (seen.has(target)) {
        errors.push(`${relativeDir}: duplicate skipTargets entry ${target}`);
        continue;
      }
      seen.add(target);
      exclusions.push({ fixture: relativeDir, target });
    }
  }

  const knownTargets = new Set([
    ...parsedPolicy.activeTargets,
    ...parsedPolicy.retiredTargets,
  ]);
  const activeTargets = new Set(parsedPolicy.activeTargets);
  const exclusionKeys = new Set(
    exclusions.map(({ fixture, target }) => exclusionKey(fixture, target)),
  );
  const allowlistByKey = new Map();

  for (const entry of parsedPolicy.allowlist) {
    const key = exclusionKey(entry.fixture, entry.target);
    if (allowlistByKey.has(key)) {
      errors.push(`duplicate allowlist entry: ${entry.fixture} (${entry.target})`);
      continue;
    }
    allowlistByKey.set(key, entry);
    if (!activeTargets.has(entry.target)) {
      errors.push(
        `allowlist entry ${entry.fixture} uses non-active target ${entry.target}`,
      );
    }
    if (!parsedPolicy.allowedClassifications.has(entry.classification)) {
      errors.push(
        `allowlist entry ${entry.fixture} has invalid classification ${entry.classification}`,
      );
    }
    if (typeof entry.reason !== "string" || entry.reason.trim() === "") {
      errors.push(`allowlist entry ${entry.fixture} needs a non-empty reason`);
    }
    if (!exclusionKeys.has(key)) {
      errors.push(`stale allowlist entry: ${entry.fixture} (${entry.target})`);
    }
  }

  for (const { fixture, target } of exclusions) {
    if (!knownTargets.has(target)) {
      errors.push(`${fixture}: unknown skip target ${target}`);
      continue;
    }
    if (activeTargets.has(target)) {
      const entry = allowlistByKey.get(exclusionKey(fixture, target));
      if (!entry) {
        errors.push(
          `${fixture}: active target ${target} exclusion is not classified and allowlisted`,
        );
      }
    }
  }

  errors.sort();
  return {
    errors,
    summary: {
      manifestCount: manifestPaths.length,
      exclusionCount: exclusions.length,
      activeExclusionCount: exclusions.filter(({ target }) =>
        activeTargets.has(target),
      ).length,
      retiredExclusionCount: exclusions.filter(({ target }) =>
        parsedPolicy.retiredTargets.includes(target),
      ).length,
      allowlistCount: parsedPolicy.allowlist.length,
    },
  };
}

function validatePolicy(policy, errors) {
  const fallback = {
    activeTargets: [],
    retiredTargets: [],
    allowedClassifications: new Set(),
    allowlist: [],
  };
  if (!policy) {
    return fallback;
  }
  if (policy.schema_version !== 1) {
    errors.push("target-exclusion policy schema_version must be 1");
  }
  const activeTargets = stringArray(
    policy.active_targets,
    "active_targets",
    errors,
  );
  const retiredTargets = stringArray(
    policy.retired_targets,
    "retired_targets",
    errors,
  );
  const classifications = stringArray(
    policy.allowed_classifications,
    "allowed_classifications",
    errors,
  );
  const activeSet = new Set(activeTargets);
  for (const target of retiredTargets) {
    if (activeSet.has(target)) {
      errors.push(`target ${target} cannot be both active and retired`);
    }
  }
  const allowlist = Array.isArray(policy.allowlist) ? policy.allowlist : [];
  if (!Array.isArray(policy.allowlist)) {
    errors.push("allowlist must be an array");
  }
  const normalizedAllowlist = [];
  for (const entry of allowlist) {
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      errors.push("allowlist entries must be objects");
      continue;
    }
    const fixture =
      typeof entry.fixture === "string"
        ? normalizeRelative(entry.fixture.trim())
        : "";
    const target = typeof entry.target === "string" ? entry.target.trim() : "";
    const classification =
      typeof entry.classification === "string"
        ? entry.classification.trim()
        : "";
    if (!fixture || fixture === "." || fixture.startsWith("../")) {
      errors.push("allowlist entries need a fixture path within the fixture root");
    }
    if (!target) {
      errors.push(`allowlist entry ${fixture || "<unknown>"} needs a target`);
    }
    if (!classification) {
      errors.push(
        `allowlist entry ${fixture || "<unknown>"} needs a classification`,
      );
    }
    normalizedAllowlist.push({
      fixture,
      target,
      classification,
      reason: entry.reason,
    });
  }
  return {
    activeTargets,
    retiredTargets,
    allowedClassifications: new Set(classifications),
    allowlist: normalizedAllowlist,
  };
}

function stringArray(value, label, errors) {
  if (!Array.isArray(value)) {
    errors.push(`${label} must be an array`);
    return [];
  }
  const result = [];
  const seen = new Set();
  for (const entry of value) {
    if (typeof entry !== "string" || entry.trim() === "") {
      errors.push(`${label} entries must be non-empty strings`);
      continue;
    }
    const normalized = entry.trim();
    if (seen.has(normalized)) {
      errors.push(`${label} contains duplicate ${normalized}`);
      continue;
    }
    seen.add(normalized);
    result.push(normalized);
  }
  return result;
}

async function collectManifestPaths(root, errors) {
  const manifests = [];
  async function visit(dir) {
    let entries;
    try {
      entries = await fs.readdir(dir, { withFileTypes: true });
    } catch (error) {
      errors.push(`cannot read fixture root ${dir}: ${error.message}`);
      return;
    }
    for (const entry of entries) {
      const entryPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        await visit(entryPath);
      } else if (entry.isFile() && entry.name === "manifest.json") {
        manifests.push(entryPath);
      }
    }
  }
  await visit(root);
  manifests.sort();
  return manifests;
}

async function readJson(filePath, label, errors) {
  try {
    return JSON.parse(await fs.readFile(filePath, "utf8"));
  } catch (error) {
    errors.push(`cannot read ${label}: ${error.message}`);
    return null;
  }
}

function normalizeRelative(value) {
  return value.split(path.sep).join("/").replace(/^\.\/+/, "");
}

function exclusionKey(fixture, target) {
  return `${fixture}\0${target}`;
}

function parseArgs(args) {
  let fixtureRoot = DEFAULT_FIXTURE_ROOT;
  let policyPath = DEFAULT_POLICY_PATH;
  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    if (arg === "--fixture-root" && index + 1 < args.length) {
      fixtureRoot = path.resolve(args[index + 1]);
      index += 1;
    } else if (arg === "--policy" && index + 1 < args.length) {
      policyPath = path.resolve(args[index + 1]);
      index += 1;
    } else {
      throw new Error(`unknown or incomplete argument: ${arg}`);
    }
  }
  return { fixtureRoot, policyPath };
}

async function main() {
  let options;
  try {
    options = parseArgs(process.argv.slice(2));
  } catch (error) {
    console.error(`fixture target-exclusion check failed:\n- ${error.message}`);
    process.exitCode = 1;
    return;
  }
  const { errors, summary } = await validateFixtureTargetExclusions(options);
  if (errors.length > 0) {
    console.error("fixture target-exclusion check failed:");
    for (const error of errors) {
      console.error(`- ${error}`);
    }
    process.exitCode = 1;
    return;
  }
  console.log(
    "fixture target exclusions ok " +
      `(manifests: ${summary.manifestCount}, ` +
      `active: ${summary.activeExclusionCount}, ` +
      `retired: ${summary.retiredExclusionCount}, ` +
      `allowlisted: ${summary.allowlistCount})`,
  );
}

if (process.argv[1] && path.resolve(process.argv[1]) === SCRIPT_PATH) {
  await main();
}
