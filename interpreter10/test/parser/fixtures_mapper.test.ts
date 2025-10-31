import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { mapSourceFile } from "../../src/parser/tree-sitter-mapper";
import { getTreeSitterParser } from "../../src/parser/tree-sitter-loader";

type FixtureEntry = {
  name: string;
  sourcePath: string;
  modulePath: string;
};

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const FIXTURE_ROOT = path.join(REPO_ROOT, "fixtures", "ast");

async function collectFixtures(): Promise<FixtureEntry[]> {
  const fixtures: FixtureEntry[] = [];

  async function walk(dir: string) {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    let hasModule = false;
    let hasSource = false;
    for (const entry of entries) {
      if (!entry.isFile()) continue;
      if (entry.name === "module.json") hasModule = true;
      if (entry.name === "source.able") hasSource = true;
    }
    if (hasModule && hasSource) {
      fixtures.push({
        name: path.relative(FIXTURE_ROOT, dir),
        sourcePath: path.join(dir, "source.able"),
        modulePath: path.join(dir, "module.json"),
      });
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        await walk(path.join(dir, entry.name));
      }
    }
  }

  await walk(FIXTURE_ROOT);
  return fixtures.sort((a, b) => a.name.localeCompare(b.name));
}

describe("tree-sitter Able mapper", () => {
  test("produces ASTs matching module.json fixtures", async () => {
    const fixtures = await collectFixtures();
    expect(fixtures.length).toBeGreaterThan(0);

    const parser = await getTreeSitterParser();

    for (const fixture of fixtures) {
      const source = await fs.readFile(fixture.sourcePath, "utf8");
      const expected = JSON.parse(await fs.readFile(fixture.modulePath, "utf8"));

      const tree = parser.parse(source);
      expect(tree.rootNode.type).toBe("source_file");
      expect(tree.rootNode.hasError).toBe(false);

      const mapped = mapSourceFile(tree.rootNode, source);
      expect(mapped).toEqual(expected);
    }
  });
});
