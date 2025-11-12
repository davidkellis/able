import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - methods & impls", () => {
  test("inherent method on struct instance", () => {
    const I = new InterpreterV10();
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
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 5 });
  });
});


