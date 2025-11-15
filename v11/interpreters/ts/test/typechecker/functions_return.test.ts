import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker function returns", () => {
  test("reports literal overflow when function body does not fit annotated return type", () => {
    const checker = new TypeChecker();
    const fn = AST.functionDefinition(
      "make_byte",
      [],
      AST.blockExpression([AST.returnStatement(AST.integerLiteral(512))]),
      AST.simpleTypeExpression("u8"),
    );
    const module = AST.module([fn]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("allows function bodies to return widened integer values", () => {
    const checker = new TypeChecker();
    const fn = AST.functionDefinition(
      "make_big",
      [],
      AST.blockExpression([
        AST.assignmentExpression(":=", AST.identifier("value"), AST.integerLiteral(1)) as unknown as AST.Statement,
        AST.returnStatement(AST.identifier("value")),
      ]),
      AST.simpleTypeExpression("i64"),
    );
    const module = AST.module([fn]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });
});
