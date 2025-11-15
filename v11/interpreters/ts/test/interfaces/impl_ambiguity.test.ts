import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - impl ambiguity/coherence", () => {
  test("reject multiple unnamed impls for same (Interface, Type)", () => {
    const I = new InterpreterV10();
    const iface = AST.interfaceDefinition("M", [
      AST.functionSignature("id", [], AST.simpleTypeExpression("Self")),
    ]);
    I.evaluate(iface);

    const impl1 = AST.implementationDefinition(
      "M",
      AST.simpleTypeExpression("i32"),
      [AST.functionDefinition("id", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(0))]))]
    );
    const impl2 = AST.implementationDefinition(
      "M",
      AST.simpleTypeExpression("i32"),
      [AST.functionDefinition("id", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]))]
    );

    I.evaluate(impl1);
    expect(() => I.evaluate(impl2)).toThrow();
  });

  test("allow multiple named impls for same (Interface, Type)", () => {
    const I = new InterpreterV10();
    const iface = AST.interfaceDefinition("N", [
      AST.functionSignature("id", [], AST.simpleTypeExpression("Self")),
    ]);
    I.evaluate(iface);

    const A = AST.implementationDefinition(
      "N",
      AST.simpleTypeExpression("i32"),
      [AST.functionDefinition("id", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(7))]))],
      "A"
    );
    const B = AST.implementationDefinition(
      "N",
      AST.simpleTypeExpression("i32"),
      [AST.functionDefinition("id", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(9))]))],
      "B"
    );

    I.evaluate(A);
    I.evaluate(B);

    const aId = I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("A"), "id"), []));
    const bId = I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("B"), "id"), []));
    expect(aId).toEqual({ kind: "i32", value: 7n });
    expect(bId).toEqual({ kind: "i32", value: 9n });
  });
});


