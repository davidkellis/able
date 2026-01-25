import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker, TypecheckerSession } from "../../src/typechecker";

describe("typechecker duplicate declarations", () => {
  test("reports previous declaration location for top-level functions", () => {
    const first = AST.functionDefinition(
      "greet",
      [AST.functionParameter("_", AST.simpleTypeExpression("String"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("hi"))]),
      AST.simpleTypeExpression("String"),
    );
    first.origin = "first.able";
    const second = AST.functionDefinition(
      "greet",
      [AST.functionParameter("_", AST.simpleTypeExpression("String"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("ho"))]),
      AST.simpleTypeExpression("String"),
    );
    second.origin = "second.able";
    const module = AST.module([first, second]);

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("previous declaration at first.able");
  });

  test("allows same symbol name across packages", () => {
    const moduleA = AST.module(
      [AST.structDefinition("Thing", [], "named")],
      [],
      AST.packageStatement(["dep"]),
    );
    const moduleB = AST.module(
      [AST.structDefinition("Thing", [], "named")],
      [],
      AST.packageStatement(["app"]),
    );

    const session = new TypecheckerSession();
    expect(session.checkModule(moduleA).diagnostics).toEqual([]);
    expect(session.checkModule(moduleB).diagnostics).toEqual([]);
  });
});
