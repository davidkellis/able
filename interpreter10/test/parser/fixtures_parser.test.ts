import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
type FixtureEntry = {
  name: string;
  sourcePath: string;
};

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const FIXTURE_ROOT = path.join(REPO_ROOT, "fixtures", "ast");
const TREE_SITTER_ROOT = path.join(REPO_ROOT, "parser10", "tree-sitter-able");
const WASM_PATH = path.join(TREE_SITTER_ROOT, "tree-sitter-able.wasm");

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

async function findFixtureParityViolations(): Promise<string[]> {
  const violations: string[] = [];

  async function walk(dir: string) {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    let hasModule = false;
    let hasSource = false;
    for (const entry of entries) {
      if (!entry.isFile()) continue;
      if (entry.name === "module.json") hasModule = true;
      if (entry.name === "source.able") hasSource = true;
    }
    if (hasModule && !hasSource) {
      violations.push(`${path.relative(FIXTURE_ROOT, dir)} missing source.able`);
    } else if (!hasModule && hasSource) {
      violations.push(`${path.relative(FIXTURE_ROOT, dir)} missing module.json`);
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        await walk(path.join(dir, entry.name));
      }
    }
  }

  await walk(FIXTURE_ROOT);
  return violations;
}

describe("tree-sitter Able grammar", () => {
  test("every fixture exports module.json and source.able", async () => {
    const violations = await findFixtureParityViolations();
    if (violations.length > 0) {
      throw new Error(`Fixture parity violations detected:\\n${violations.join("\\n")}`);
    }
  });

  test("parses every fixture source without errors", async () => {
    const fixtures = await collectFixtures();
    expect(fixtures.length).toBeGreaterThan(0);

    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    for (const fixture of fixtures) {
      const source = await fs.readFile(fixture.sourcePath, "utf8");
      const tree = parser.parse(source);

      expect(tree.rootNode.type).toBe("source_file");
      expect(tree.rootNode.hasError).toBe(false);
    }
  });

  test("parses modules that contain comments", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = `## comment before package
package demo

## standalone comment
fn main() -> void {
  ## inner comment
}
`;

    const tree = parser.parse(source);
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);
  });
});
