import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - destructuring assignment", () => {
  test("array destructuring with rest", () => {
    const I = new InterpreterV10();
    const arr = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2), AST.integerLiteral(3)]);
    const pat = AST.arrayPattern([AST.identifier("a"), AST.identifier("b")], AST.identifier("rest"));
    I.evaluate(AST.assignmentExpression(":=", pat as any, arr));
    expect(I.evaluate(AST.identifier("a"))).toEqual({ kind: 'i32', value: 1n });
    expect(I.evaluate(AST.identifier("b"))).toEqual({ kind: 'i32', value: 2n });
    const restVal = I.evaluate(AST.identifier("rest"));
    expect(restVal).toEqual({ kind: 'array', elements: [{ kind: 'i32', value: 3n }] });
  });

  test("destructuring assignment with = declares bindings when missing", () => {
    const I = new InterpreterV10();
    const arr = AST.arrayLiteral([AST.integerLiteral(4), AST.integerLiteral(5)]);
    const pat = AST.arrayPattern([AST.identifier("left"), AST.identifier("right")]);
    I.evaluate(AST.assignmentExpression("=", pat as any, arr));
    expect(I.evaluate(AST.identifier("left"))).toEqual({ kind: "i32", value: 4n });
    expect(I.evaluate(AST.identifier("right"))).toEqual({ kind: "i32", value: 5n });
  });

  test(":= requires at least one new binding in destructuring", () => {
    const I = new InterpreterV10();
    const first = AST.arrayLiteral([AST.integerLiteral(8), AST.integerLiteral(9)]);
    const pat = AST.arrayPattern([AST.identifier("m"), AST.identifier("n")]);
    I.evaluate(AST.assignmentExpression(":=", pat as any, first));
    expect(() => I.evaluate(AST.assignmentExpression(":=", pat as any, first))).toThrow();
  });

  test(":= allows reassignment in pattern when at least one name is new", () => {
    const I = new InterpreterV10();
    const initial = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)]);
    const pat = AST.arrayPattern([AST.identifier("x"), AST.identifier("y")]);
    I.evaluate(AST.assignmentExpression(":=", pat as any, initial));
    const mixed = AST.arrayPattern([AST.identifier("x"), AST.identifier("z")]);
    const arr = AST.arrayLiteral([AST.integerLiteral(3), AST.integerLiteral(4)]);
    I.evaluate(AST.assignmentExpression(":=", mixed as any, arr));
    expect(I.evaluate(AST.identifier("x"))).toEqual({ kind: "i32", value: 3n });
    expect(I.evaluate(AST.identifier("z"))).toEqual({ kind: "i32", value: 4n });
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
    expect(I.evaluate(AST.identifier("mx"))).toEqual({ kind: 'i32', value: 10n });
    expect(I.evaluate(AST.identifier("my"))).toEqual({ kind: 'i32', value: 20n });
  });
});

