import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("TypeChecker iterables", () => {
  test("for-loop typed pattern rejects mismatched Iterable element types", () => {
    const displayInterface = AST.interfaceDefinition("Display", [
      AST.functionSignature(
        "describe",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string"),
      ),
    ]);

    const iterableInterface = AST.interfaceDefinition(
      "Iterable",
      [
        AST.functionSignature(
          "iterator",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("Iterator"),
        ),
      ],
      [AST.genericParameter("T")],
    );

    const iterableType = AST.genericTypeExpression(AST.simpleTypeExpression("Iterable"), [
      AST.simpleTypeExpression("string"),
    ]);

    const makeItemsFn = AST.functionDefinition("make_items", [], AST.blockExpression([AST.nilLiteral()]), iterableType);

    const bindItems = AST.assignmentExpression(
      ":=",
      AST.typedPattern(AST.identifier("items"), iterableType),
      AST.functionCall(AST.identifier("make_items"), []),
    );

    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("item"), AST.simpleTypeExpression("Display")),
      AST.identifier("items"),
      AST.blockExpression([AST.identifier("item")]),
    );

    const moduleAst = AST.module(
      [displayInterface, iterableInterface, makeItemsFn, bindItems, loop],
      [],
      AST.packageStatement(["app"]),
    );

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("for-loop pattern expects type Display");
    expect(diagnostics[0]?.message).toContain("got string");
  });
});
