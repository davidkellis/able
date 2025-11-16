import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - compound assignments", () => {
  test("identifier targets for arithmetic and bitwise", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(2)));
    I.evaluate(AST.assignmentExpression("+=", AST.identifier("x"), AST.integerLiteral(3)));
    expect(I.evaluate(AST.identifier("x"))).toEqual({ kind: 'i32', value: 5n });
    I.evaluate(AST.assignmentExpression("<<=", AST.identifier("x"), AST.integerLiteral(1)));
    expect(I.evaluate(AST.identifier("x"))).toEqual({ kind: 'i32', value: 10n });
  });

  test("struct field and array index compound assignment", () => {
    const I = new InterpreterV10();
    const Def = AST.structDefinition("Point", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")], "named");
    I.evaluate(Def);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point")));
    I.evaluate(AST.assignmentExpression("+=", AST.memberAccessExpression(AST.identifier("p"), "x"), AST.integerLiteral(4)));
    const px = I.evaluate(AST.memberAccessExpression(AST.identifier("p"), "x"));
    expect(px).toEqual({ kind: 'i32', value: 5n });

    I.evaluate(AST.assignmentExpression(":=", AST.identifier("arr"), AST.arrayLiteral([AST.integerLiteral(3), AST.integerLiteral(4)])));
    I.evaluate(AST.assignmentExpression("*=", AST.indexExpression(AST.identifier("arr"), AST.integerLiteral(1)), AST.integerLiteral(2)));
    const a1 = I.evaluate(AST.indexExpression(AST.identifier("arr"), AST.integerLiteral(1)));
    expect(a1).toEqual({ kind: 'i32', value: 8n });
  });
});


