import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker duplicate declarations", () => {
  test("reports previous declaration location for top-level functions", () => {
    const first = AST.functionDefinition(
      "greet",
      [AST.functionParameter("_", AST.simpleTypeExpression("string"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("hi"))]),
      AST.simpleTypeExpression("string"),
    );
    first.origin = "first.able";
    const second = AST.functionDefinition(
      "greet",
      [AST.functionParameter("_", AST.simpleTypeExpression("string"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("ho"))]),
      AST.simpleTypeExpression("string"),
    );
    second.origin = "second.able";
    const module = AST.module([first, second]);

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("previous declaration at first.able");
  });
});
