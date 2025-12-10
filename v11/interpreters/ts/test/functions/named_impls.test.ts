import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - named impls as callable namespaces", () => {
  test("define two named impls and call their static methods via name", () => {
    const I = new InterpreterV10();

    // interface Monoid for T { fn id() -> Self; fn op(self: Self, other: Self) -> Self }
    const monoid = AST.interfaceDefinition("Monoid", [
      AST.functionSignature("id", [], AST.simpleTypeExpression("Self")),
      AST.functionSignature("op", [AST.functionParameter("a", AST.simpleTypeExpression("Self")), AST.functionParameter("b", AST.simpleTypeExpression("Self"))], AST.simpleTypeExpression("Self")),
    ]);
    I.evaluate(monoid);

    // Named impls for i32
    const sumName = AST.identifier("Sum");
    const prodName = AST.identifier("Product");
    const implSum = AST.implementationDefinition(
      "Monoid",
      AST.simpleTypeExpression("i32"),
      [
        AST.functionDefinition("id", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(0))])),
        AST.functionDefinition("op", [AST.functionParameter("a"), AST.functionParameter("b")], AST.blockExpression([
          AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))
        ])),
      ],
      sumName,
    );
    const implProd = AST.implementationDefinition(
      "Monoid",
      AST.simpleTypeExpression("i32"),
      [
        AST.functionDefinition("id", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))])),
        AST.functionDefinition("op", [AST.functionParameter("a"), AST.functionParameter("b")], AST.blockExpression([
          AST.returnStatement(AST.binaryExpression("*", AST.identifier("a"), AST.identifier("b")))
        ])),
      ],
      prodName,
    );
    I.evaluate(implSum);
    I.evaluate(implProd);

    // Call static id via named impl
    const sumNs = I.evaluate(AST.identifier("Sum"));
    expect(sumNs.kind).toBe("impl_namespace");
    const sumId = I.evaluate(AST.memberAccessExpression(AST.identifier("Sum"), "id"));
    expect(sumId.kind).toBe("function");
    const sumVal = I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("Sum"), "id"), []));
    expect(sumVal).toEqual({ kind: "i32", value: 0n });

    const prodVal = I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("Product"), "id"), []));
    expect(prodVal).toEqual({ kind: "i32", value: 1n });

    const ifaceName = I.evaluate(AST.memberAccessExpression(AST.identifier("Sum"), "interface"));
    expect(ifaceName).toEqual({ kind: "String", value: "Monoid" });
    const targetName = I.evaluate(AST.memberAccessExpression(AST.identifier("Sum"), "target"));
    expect(targetName).toEqual({ kind: "String", value: "i32" });
  });
});

