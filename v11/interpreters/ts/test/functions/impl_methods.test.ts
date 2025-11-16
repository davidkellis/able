import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - impl methods via ImplementationDefinition", () => {
  test("impl adds method available on struct instances", () => {
    const I = new InterpreterV10();

    // struct Point { x: i32, y: i32 }
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);

    // Define a dummy interface and an impl for Point providing method 'sum'
    const iface = AST.interfaceDefinition("Adder", [], [], undefined, undefined, undefined, false);
    I.evaluate(iface);

    const impl = AST.implementationDefinition(
      "Adder",
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "sum",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(
              AST.binaryExpression(
                "+",
                AST.memberAccessExpression(AST.identifier("self"), "x"),
                AST.memberAccessExpression(AST.identifier("self"), "y")
              )
            ),
          ])
        )
      ]
    );
    I.evaluate(impl);

    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(4), "x"),
      AST.structFieldInitializer(AST.integerLiteral(6), "y"),
    ], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), p));

    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("p"), "sum"), []);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 10n });
  });
});


