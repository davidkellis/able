import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker type aliases", () => {
  test("allows typed bindings to reference simple aliases", () => {
    const alias = AST.typeAliasDefinition(AST.identifier("UserID"), AST.simpleTypeExpression("i32"));
    const binding = AST.assignmentExpression(
      ":=",
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("UserID")),
      AST.integerLiteral(42),
    );
    const module = AST.module([alias, binding]);
    const checker = new TypeChecker();

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("instantiates generic aliases with explicit type arguments", () => {
    const alias = AST.typeAliasDefinition(
      AST.identifier("Box"),
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const binding = AST.assignmentExpression(
      ":=",
      AST.typedPattern(
        AST.identifier("values"),
        AST.genericTypeExpression(AST.simpleTypeExpression("Box"), [AST.simpleTypeExpression("string")]),
      ),
      AST.arrayLiteral([AST.stringLiteral("a"), AST.stringLiteral("b")]),
    );
    const module = AST.module([alias, binding]);
    const checker = new TypeChecker();

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("records aliases in package summaries", () => {
    const alias = AST.typeAliasDefinition(AST.identifier("UserID"), AST.simpleTypeExpression("i32"));
    const checker = new TypeChecker();

    const result = checker.checkModule(AST.module([alias]));
    expect(result.summary?.symbols["UserID"]?.type).toContain("type alias");
  });
});
