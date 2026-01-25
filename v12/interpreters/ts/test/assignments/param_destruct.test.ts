import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - destructuring function parameters", () => {
  test("array destructuring params", () => {
    const I = new Interpreter();
    const fn = AST.functionDefinition(
      "sum2",
      [AST.functionParameter(AST.arrayPattern([AST.identifier("a"), AST.identifier("b")]))],
      AST.blockExpression([
        AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))
      ])
    );
    I.evaluate(fn);
    const call = AST.functionCall(AST.identifier("sum2"), [AST.arrayLiteral([AST.integerLiteral(3), AST.integerLiteral(4)])]);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 7n });
  });

  test("struct destructuring params", () => {
    const I = new Interpreter();
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);
    const fn = AST.functionDefinition(
      "sumPoint",
      [AST.functionParameter(AST.structPattern([AST.structPatternField(AST.identifier("x"), "x"), AST.structPatternField(AST.identifier("y"), "y")], false, "Point"))],
      AST.blockExpression([
        AST.returnStatement(AST.binaryExpression("+", AST.identifier("x"), AST.identifier("y")))
      ])
    );
    I.evaluate(fn);
    const arg = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(5), "x"),
      AST.structFieldInitializer(AST.integerLiteral(6), "y"),
    ], false, "Point");
    const call = AST.functionCall(AST.identifier("sumPoint"), [arg]);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 11n });
  });
});


