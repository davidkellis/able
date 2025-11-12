import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - destructuring assignment", () => {
  test("array destructuring with rest", () => {
    const I = new InterpreterV10();
    const arr = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2), AST.integerLiteral(3)]);
    const pat = AST.arrayPattern([AST.identifier("a"), AST.identifier("b")], AST.identifier("rest"));
    I.evaluate(AST.assignmentExpression(":=", pat as any, arr));
    expect(I.evaluate(AST.identifier("a"))).toEqual({ kind: 'i32', value: 1 });
    expect(I.evaluate(AST.identifier("b"))).toEqual({ kind: 'i32', value: 2 });
    const restVal = I.evaluate(AST.identifier("rest"));
    expect(restVal).toEqual({ kind: 'array', elements: [{ kind: 'i32', value: 3 }] });
  });

  test("struct destructuring named fields", () => {
    const I = new InterpreterV10();
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);
    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(10), "x"),
      AST.structFieldInitializer(AST.integerLiteral(20), "y"),
    ], false, "Point");
    const pat = AST.structPattern([
      AST.structPatternField(AST.identifier("mx"), "x"),
      AST.structPatternField(AST.identifier("my"), "y"),
    ], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", pat as any, p));
    expect(I.evaluate(AST.identifier("mx"))).toEqual({ kind: 'i32', value: 10 });
    expect(I.evaluate(AST.identifier("my"))).toEqual({ kind: 'i32', value: 20 });
  });
});


