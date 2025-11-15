import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - index & member assignment", () => {
  test("array index mutation", () => {
    const I = new InterpreterV10();
    const arr = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)]);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("a"), arr));
    I.evaluate(AST.assignmentExpression("=", AST.indexExpression(AST.identifier("a"), AST.integerLiteral(1) as any), AST.integerLiteral(9)) as any);
    expect(I.evaluate(AST.identifier("a"))).toEqual({ kind: 'array', elements: [{ kind: 'i32', value: 1n }, { kind: 'i32', value: 9n }] });
  });

  test("named struct field mutation", () => {
    const I = new InterpreterV10();
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);
    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(0), "x"),
      AST.structFieldInitializer(AST.integerLiteral(0), "y"),
    ], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), p));
    I.evaluate(AST.assignmentExpression("=", AST.memberAccessExpression(AST.identifier("p"), "x"), AST.integerLiteral(5)));
    expect(I.evaluate(AST.memberAccessExpression(AST.identifier("p"), "x"))).toEqual({ kind: 'i32', value: 5n });
  });
});


