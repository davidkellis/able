#!/usr/bin/env node

// Verifies that every exec fixture directory is captured in the coverage index
// and that seeded entries have corresponding fixture directories.

import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const EXEC_ROOT = path.join(ROOT, "fixtures", "exec");
const COVERAGE_PATH = path.join(EXEC_ROOT, "coverage-index.json");
const VALID_STATUSES = new Set(["seeded", "planned"]);

async function main() {
  const errors = [];

  const coverage = JSON.parse(await fs.readFile(COVERAGE_PATH, "utf8"));
  const byId = new Map();
  for (const entry of coverage) {
    if (!entry?.id || typeof entry.id !== "string") {
      errors.push("coverage entry missing string id");
      continue;
    }
    if (byId.has(entry.id)) {
      errors.push(`duplicate coverage id: ${entry.id}`);
      continue;
    }
    if (!VALID_STATUSES.has(entry.status)) {
      errors.push(`invalid status for ${entry.id}: ${entry.status}`);
    }
    byId.set(entry.id, entry);
  }

  const dirEntries = await fs.readdir(EXEC_ROOT, { withFileTypes: true });
  const fixtureDirs = dirEntries.filter((ent) => ent.isDirectory());

  for (const dirent of fixtureDirs) {
    const dirPath = path.join(EXEC_ROOT, dirent.name);
    const manifestPath = path.join(dirPath, "manifest.json");
    const hasManifest = await exists(manifestPath);
    if (!hasManifest) {
      // Skip non-fixture directories (e.g., future support files) silently.
      continue;
    }
    const id = `exec/${dirent.name}`;
    const entry = byId.get(id);
    if (!entry) {
      errors.push(`fixture directory missing from coverage index: ${id}`);
      continue;
    }
    if (entry.status !== "seeded") {
      errors.push(`fixture ${id} exists but coverage status is ${entry.status} (expected seeded)`);
    }
  }

  for (const [id, entry] of byId.entries()) {
    if (entry.status !== "seeded") {
      continue;
    }
    const dirName = id.replace(/^exec\//, "");
    const dirPath = path.join(EXEC_ROOT, dirName);
    if (!(await exists(dirPath))) {
      errors.push(`coverage marks ${id} as seeded but directory is missing`);
    }
  }

  if (errors.length > 0) {
    console.error("exec coverage check failed:");
    for (const err of errors) {
      console.error(`- ${err}`);
    }
    process.exit(1);
  }

  const seededCount = coverage.filter((c) => c.status === "seeded").length;
  const plannedCount = coverage.filter((c) => c.status === "planned").length;
  const fixtureCount = fixtureDirs.length;
  console.log(`exec coverage ok (seeded: ${seededCount}, planned: ${plannedCount}, fixture dirs: ${fixtureCount})`);
}

async function exists(p) {
  try {
    await fs.access(p);
    return true;
  } catch {
    return false;
  }
}

await main();
