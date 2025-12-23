import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - methods & impls", () => {
  test("inherent method on struct instance", () => {
    const I = new Interpreter();
    // struct Point { x: i32, y: i32 }
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);

    // methods for Point
    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "sum",
          [AST.functionParameter("self" )],
          AST.blockExpression([
            AST.returnStatement(AST.binaryExpression("+", AST.memberAccessExpression(AST.identifier("self"), "x"), AST.memberAccessExpression(AST.identifier("self"), "y")))
          ])
        )
      ]
    );
    I.evaluate(methods);

    // p := Point { x: 2, y: 3 }
    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(2), "x"),
      AST.structFieldInitializer(AST.integerLiteral(3), "y"),
    ], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), p));

    // p.sum()
    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("p"), "sum"), []);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 5n });
  });

  test("methods are exported as callable functions", () => {
    const I = new Interpreter();
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named",
    );
    I.evaluate(pointDef);

    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Point"),
      [
        AST.fn(
          "norm",
          [],
          AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
          AST.simpleTypeExpression("i32"),
          undefined,
          undefined,
          true,
        ),
        AST.fn(
          "origin",
          [],
          AST.blockExpression([
            AST.returnStatement(
              AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(0), "x")], false, "Point"),
            ),
          ]),
        ),
      ],
    );
    I.evaluate(methods);

    const pointValue = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(3), "x")], false, "Point");
    const callAsMethod = AST.functionCall(AST.memberAccessExpression(pointValue, "norm"), []);
    const callAsFunction = AST.functionCall(AST.identifier("norm"), [pointValue]);
    const staticCall = AST.functionCall(AST.memberAccessExpression(AST.identifier("Point"), "origin"), []);

    expect(I.evaluate(callAsMethod)).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(callAsFunction)).toEqual({ kind: "i32", value: 1n });

    const origin = I.evaluate(staticCall);
    expect(origin.kind).toBe("struct_instance");
  });
});
