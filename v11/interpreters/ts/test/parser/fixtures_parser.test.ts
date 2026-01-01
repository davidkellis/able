import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as AST from "../../src/ast";
import { mapSourceFile } from "../../src/parser/tree-sitter-mapper";
type FixtureEntry = {
  name: string;
  sourcePath: string;
};

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../../../");
const FIXTURE_ROOT = path.join(REPO_ROOT, "fixtures", "ast");
const TREE_SITTER_ROOT = path.join(REPO_ROOT, "parser", "tree-sitter-able");
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
      if (!tree) throw new Error(`failed to parse ${fixture.sourcePath}`);

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
    if (!tree) throw new Error("failed to parse comments fixture");
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);
  });

  test("rejects prefix-style match expressions", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = `fn main() {
  match 1 {
    case _ => 2
  }
}
`;

    const tree = parser.parse(source);
    if (!tree) throw new Error("failed to parse prefix match sample");
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(true);
    expect(() => mapSourceFile(tree.rootNode, source, "<inline>")).toThrow(/syntax errors present/);
  });

  test("splits unparenthesized generic interface arguments", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = "impl Foo Array i64 for Bar {}";
    const tree = parser.parse(source);
    if (!tree) throw new Error("failed to parse impl fixture");
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);
    const module = mapSourceFile(tree.rootNode, source, "<inline>");
    const impl = module.body[0] as AST.ImplementationDefinition;
    expect(impl.type).toBe("ImplementationDefinition");
    expect(impl.interfaceArgs?.length ?? 0).toBe(2);

    const okSource = "impl Foo (Array i64) for Bar {}";
    const okTree = parser.parse(okSource);
    if (!okTree) throw new Error("failed to parse impl fixture");
    expect(okTree.rootNode.type).toBe("source_file");
    expect(okTree.rootNode.hasError).toBe(false);
    const okModule = mapSourceFile(okTree.rootNode, okSource, "<inline>");
    const okImpl = okModule.body[0] as AST.ImplementationDefinition;
    expect(okImpl.type).toBe("ImplementationDefinition");
    expect(okImpl.interfaceArgs?.length ?? 0).toBe(1);
    expect(okImpl.interfaceArgs?.[0]?.type).toBe("GenericTypeExpression");
  });

  test("parses verbose anonymous function syntax", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = "adder := fn(x: i32) -> i32 { x + 1 }\n";
    const tree = parser.parse(source);
    if (!tree) throw new Error("failed to parse verbose lambda fixture");
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);
    const module = mapSourceFile(tree.rootNode, source, "<inline>");
    const assignment = module.body[0] as AST.AssignmentExpression;
    expect(assignment.type).toBe("AssignmentExpression");
    const lambda = assignment.right as AST.LambdaExpression;
    expect(lambda.type).toBe("LambdaExpression");
    expect(lambda.isVerboseSyntax).toBe(true);
  });

  test("range operators map inclusivity correctly", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = `exclusive := 0...5
inclusive := 0..5
`;
    const tree = parser.parse(source);
    if (!tree) throw new Error("failed to parse range fixture");
    const module = mapSourceFile(tree.rootNode, source, "<inline>");
    const exclusive = module.body[0];
    const inclusive = module.body[1];
    if (exclusive?.type !== "AssignmentExpression" || inclusive?.type !== "AssignmentExpression") {
      throw new Error("expected assignment expressions for range test");
    }
    const exclusiveRange = exclusive.right as AST.RangeExpression;
    const inclusiveRange = inclusive.right as AST.RangeExpression;
    expect(exclusiveRange.inclusive).toBe(false);
    expect(inclusiveRange.inclusive).toBe(true);
  });

  test("parses loop expression statements with break/continue", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = `counter := 3
loop {
  counter = counter - 1
  if counter < 0 {
    break
  }
}
counter
`;

    const tree = parser.parse(source);
    if (!tree) throw new Error("failed to parse loop stmt fixture");
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);

    const loopNodes = (tree.rootNode as any).descendantsOfType?.(["loop_expression"]) ?? [];
    expect(loopNodes.length).toBe(1);
    let ancestor = loopNodes[0]?.parent;
    let sawExpressionStatement = false;
    while (ancestor) {
      if (ancestor.type === "expression_statement") {
        sawExpressionStatement = true;
        break;
      }
      ancestor = ancestor.parent;
    }
    expect(sawExpressionStatement).toBe(true);
  });

  test("maps '=' assignments inside loops as reassignments", async () => {
    const { Parser, Language } = await import("web-tree-sitter");
    await Parser.init();
    const parser = new Parser();
    const language = await Language.load(WASM_PATH);
    parser.setLanguage(language);

    const source = `fn main() {
  a = 5
  loop {
    if a <= 0 { break }
    a = a - 1
  }
}
`;

    const tree = parser.parse(source);
    if (!tree) throw new Error("failed to parse loop assignment fixture");
    const module = mapSourceFile(tree.rootNode, source, "<inline>");
    const fn = module.body.find((stmt): stmt is AST.FunctionDefinition => stmt?.type === "FunctionDefinition");
    if (!fn || fn.body?.type !== "BlockExpression") throw new Error("expected function definition with block body");

    const [initAssign, loopStmt] = fn.body.body;
    if (initAssign?.type !== "AssignmentExpression") throw new Error("expected initial assignment");
    expect(initAssign.operator).toBe("=");

    if (loopStmt?.type !== "LoopExpression") throw new Error("expected loop expression statement");
    const loopAssignments = loopStmt.body.body.filter(
      (stmt): stmt is AST.AssignmentExpression => stmt?.type === "AssignmentExpression",
    );
    const decrementAssign = loopAssignments.find(
      (stmt) => stmt.left.type === "Identifier" && stmt.left.name === "a",
    );
    if (!decrementAssign) {
      throw new Error("expected loop assignment to reassign 'a'");
    }
    expect(decrementAssign.operator).toBe("=");
    expect(decrementAssign.right.type).toBe("BinaryExpression");
    const binary = decrementAssign.right as AST.BinaryExpression;
    expect(binary.operator).toBe("-");
    expect(binary.left.type).toBe("Identifier");
    expect((binary.left as AST.Identifier).name).toBe("a");
    expect(binary.right.type).toBe("IntegerLiteral");
  });
});
