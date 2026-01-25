import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker interface self pattern scope", () => {
  test("self pattern placeholders are in scope for interface method signatures", () => {
    const signature = AST.functionSignature(
      "map",
      [
        AST.functionParameter(
          "self",
          AST.genericTypeExpression(AST.simpleTypeExpression("C"), [AST.simpleTypeExpression("A")]),
        ),
        AST.functionParameter(
          "f",
          AST.functionTypeExpression([AST.simpleTypeExpression("A")], AST.simpleTypeExpression("B")),
        ),
      ],
      AST.genericTypeExpression(AST.simpleTypeExpression("C"), [AST.simpleTypeExpression("B")]),
      [AST.genericParameter("B")],
    );
    const iface = AST.interfaceDefinition(
      "Enumerable",
      [signature],
      [AST.genericParameter("A")],
      AST.genericTypeExpression(AST.simpleTypeExpression("C"), [AST.wildcardTypeExpression()]),
    );
    const moduleAst = AST.module([iface]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([]);

    const genericNames = (signature.genericParams ?? [])
      .map((param) => param?.name?.name)
      .filter((name): name is string => Boolean(name));
    expect(genericNames).toEqual(["B"]);
  });
});
