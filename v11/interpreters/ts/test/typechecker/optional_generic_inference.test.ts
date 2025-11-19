import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker optional generics - redeclaration guards", () => {
  test("reports struct definition that reuses an inferred generic name", () => {
    const structDef = AST.structDefinition(
      "T",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );
    const fn = AST.fn(
      "wrap",
      [AST.functionParameter("value", AST.simpleTypeExpression("T"))],
      [structDef, AST.returnStatement(AST.identifier("value"))],
      AST.simpleTypeExpression("T"),
    );
    const module = AST.module([fn]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("cannot redeclare inferred type parameter 'T'");
  });

  test("reports type alias that reuses an inferred generic name", () => {
    const alias = AST.typeAliasDefinition("T", AST.simpleTypeExpression("i64"));
    const fn = AST.fn(
      "convert",
      [AST.functionParameter("value", AST.simpleTypeExpression("T"))],
      [alias, AST.returnStatement(AST.identifier("value"))],
      AST.simpleTypeExpression("T"),
    );
    const module = AST.module([fn]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("cannot redeclare inferred type parameter 'T'");
  });
});
