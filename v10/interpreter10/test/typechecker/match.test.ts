import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker match/rescue", () => {
  test("reports diagnostic when match guard is not bool", () => {
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
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("match guard must be bool");
  });

  test("reports diagnostic when rescue guard is not bool", () => {
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
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("rescue guard must be bool");
  });
});
