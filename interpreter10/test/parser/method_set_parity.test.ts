import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import type { MethodsDefinition, WhereClauseConstraint } from "../../src/ast";
import { mapSourceFile } from "../../src/parser/tree-sitter-mapper";
import { getTreeSitterParser } from "../../src/parser/tree-sitter-loader";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const FIXTURE_DIR = path.join(
  REPO_ROOT,
  "fixtures",
  "ast",
  "functions",
  "method_set_where_constraint_ok",
);

function findMethodsDefinition(module: any): MethodsDefinition | null {
  for (const stmt of module.body ?? []) {
    if (stmt && stmt.type === "MethodsDefinition") {
      return stmt as MethodsDefinition;
    }
  }
  return null;
}

describe("tree-sitter Able mapper - method-set coverage", () => {
  test("preserves where-clause constraints on methods definitions", async () => {
    const parser = await getTreeSitterParser();
    const source = await fs.readFile(path.join(FIXTURE_DIR, "source.able"), "utf8");
    const tree = parser.parse(source);
    expect(tree.rootNode.type).toBe("source_file");
    expect(tree.rootNode.hasError).toBe(false);

    const module = mapSourceFile(tree.rootNode, source);
    const methods = findMethodsDefinition(module);
    if (!methods) {
      throw new Error("expected fixture to include a MethodsDefinition");
    }

    expect(methods.targetType?.type).toBe("SimpleTypeExpression");
    expect(methods.targetType?.name?.name).toBe("Wrapper");

    const whereClause = methods.whereClause ?? [];
    expect(whereClause.length).toBe(1);
    const constraint = whereClause[0] as WhereClauseConstraint;
    expect(constraint.typeParam?.name).toBe("Self");
    expect(constraint.constraints?.length).toBe(1);

    const interfaceConstraint = constraint.constraints?.[0];
    expect(interfaceConstraint?.type).toBe("InterfaceConstraint");
    expect(interfaceConstraint?.interfaceType?.type).toBe("SimpleTypeExpression");
    expect(interfaceConstraint?.interfaceType?.name?.name).toBe("Display");

    expect(methods.definitions.length).toBe(1);
    const describeFn = methods.definitions[0];
    expect(describeFn.id.name).toBe("describe");
  });
});
