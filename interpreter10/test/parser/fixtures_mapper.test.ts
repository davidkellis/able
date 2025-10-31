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

function normalizeModule(value: unknown): void {
  if (Array.isArray(value)) {
    for (const item of value) {
      normalizeModule(item);
    }
    return;
  }
  if (!value || typeof value !== "object") {
    return;
  }
  const record = value as Record<string, unknown>;
  for (const key of Object.keys(record)) {
    const entry = record[key];
    if (key === "isShorthand" && entry === false) {
      delete record[key];
      continue;
    }
    normalizeModule(entry);
  }
}

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
      normalizeModule(mapped);
      normalizeModule(expected);
      expect(mapped).toEqual(expected);
    }
  });

  test("ignores comments when mapping modules", async () => {
    const parser = await getTreeSitterParser();
    const source = `## leading comment
package demo

## another comment
fn main() -> void {
  ## inside block
}
`;

    const tree = parser.parse(source);
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);

    const module = mapSourceFile(tree.rootNode, source);
    expect(module.package?.namePath?.map(id => id.name)).toEqual(["demo"]);
    expect(module.body).toHaveLength(1);
    expect(module.body[0]?.type).toBe("FunctionDefinition");
  });

  test("ignores comments inside struct literals with spreads", async () => {
    const parser = await getTreeSitterParser();
    const source = `struct Point {
  x: i32,
  y: i32,
}

fn update(base: Point) -> Point {
  Point {
    ## spread previous fields
    ...base,
    ## override with trailing comment
    x: base.x + 1 ## trailing
  }
}
`;

    const tree = parser.parse(source);
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);

    const module = mapSourceFile(tree.rootNode, source);
    const fn = module.body.find(stmt => stmt.type === "FunctionDefinition");
    expect(fn?.type).toBe("FunctionDefinition");
    const bodyStatements = fn && "body" in fn ? fn.body.body : [];
    expect(bodyStatements?.length).toBeGreaterThan(0);
    const literal = bodyStatements?.[bodyStatements.length - 1];
    expect(literal?.type).toBe("StructLiteral");
    if (literal?.type !== "StructLiteral") {
      throw new Error("expected struct literal expression");
    }
    expect(literal.functionalUpdateSources?.length).toBe(1);
    expect(literal.functionalUpdateSources?.[0]?.type).toBe("Identifier");
    expect(literal.functionalUpdateSources?.[0]?.name).toBe("base");
    expect(literal.fields).toHaveLength(1);
    expect(literal.fields[0]?.name?.name).toBe("x");
  });

  test.skip("ignores comments inside struct patterns", async () => {
    // TODO: Enable once the tree-sitter grammar accepts comments within pattern bodies.
    const parser = await getTreeSitterParser();
    const source = `struct Point {
  x: i32,
  y: i32,
}

fn project(point: Point) -> i32 {
  match point {
    Point {
      x: px, ## capture x
      y: py ## capture y
    } => px + py,
    _ => 0 ## fallback branch
  }
}
`;

    const tree = parser.parse(source);
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);

    const module = mapSourceFile(tree.rootNode, source);
    const fn = module.body.find(stmt => stmt.type === "FunctionDefinition");
    expect(fn?.type).toBe("FunctionDefinition");
    const bodyStatements = fn && "body" in fn ? fn.body.body : [];
    const matchExpr = bodyStatements?.find(stmt => stmt?.type === "MatchExpression");
    expect(matchExpr?.type).toBe("MatchExpression");
    if (matchExpr?.type !== "MatchExpression") {
      throw new Error("expected match expression");
    }
    expect(matchExpr.clauses).toHaveLength(2);
    const structClause = matchExpr.clauses[0];
    expect(structClause?.pattern?.type).toBe("StructPattern");
    if (structClause?.pattern?.type !== "StructPattern") {
      throw new Error("expected struct pattern");
    }
    expect(structClause.pattern.fields).toHaveLength(2);
    const names = structClause.pattern.fields.map(field => field.fieldName?.name ?? field.pattern.type);
    expect(names).toEqual(["x", "y"]);
  });
});
