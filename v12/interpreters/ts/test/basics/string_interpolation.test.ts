import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - String interpolation", () => {
  test("interpolates literals and expressions", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(2)));
    const str = AST.stringInterpolation([
      AST.stringLiteral("x = "),
      AST.identifier("x"),
      AST.stringLiteral(", sum = "),
      AST.binaryExpression("+", AST.integerLiteral(3), AST.integerLiteral(4)),
    ]);
    expect(I.evaluate(str)).toEqual({ kind: 'String', value: 'x = 2, sum = 7' });
  });

  test("uses to_String method on struct instances when available", () => {
    const I = new Interpreter();
    const Def = AST.structDefinition("Point", [
      AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
      AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
    ], "named");
    I.evaluate(Def);
    const Meths = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.functionDefinition("to_String", [AST.functionParameter("self")], AST.blockExpression([
        AST.returnStatement(AST.stringInterpolation([
          AST.stringLiteral("Point("),
          AST.memberAccessExpression(AST.identifier("self"), "x"),
          AST.stringLiteral(","),
          AST.memberAccessExpression(AST.identifier("self"), "y"),
          AST.stringLiteral(")"),
        ]))
      ]))
    ]);
    I.evaluate(Meths);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point")));
    const s = AST.stringInterpolation([
      AST.stringLiteral("P= "),
      AST.identifier("p"),
    ]);
    expect(I.evaluate(s)).toEqual({ kind: 'String', value: 'P= Point(1,2)' });
  });
});


