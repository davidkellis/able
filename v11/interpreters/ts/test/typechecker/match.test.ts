import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker match/rescue", () => {
  test("accepts truthy match guards", () => {
    const checker = new TypeChecker();
    const matchExpr = AST.matchExpression(AST.integerLiteral(1), [
      AST.matchClause(
        AST.identifier("value"),
        AST.identifier("value"),
        AST.integerLiteral(0),
      ),
    ]);
    const module = AST.module([matchExpr as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("accepts truthy rescue guards", () => {
    const checker = new TypeChecker();
    const rescueExpr = AST.rescueExpression(AST.stringLiteral("ok"), [
      AST.matchClause(
        AST.identifier("err"),
        AST.identifier("err"),
        AST.integerLiteral(1),
      ),
    ]);
    const module = AST.module([rescueExpr as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });
});
