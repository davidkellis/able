import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - unary/binary ops and ranges", () => {
  const I = new InterpreterV10();

  test("unary", () => {
    expect(I.evaluate(AST.unaryExpression('-', AST.integerLiteral(5)))).toEqual({ kind: 'i32', value: -5 });
    expect(I.evaluate(AST.unaryExpression('!', AST.booleanLiteral(false)))).toEqual({ kind: 'bool', value: true });
    expect(I.evaluate(AST.unaryExpression('~', AST.integerLiteral(0)))).toEqual({ kind: 'i32', value: -1 });
  });

  test("arithmetic & string +", () => {
    const add = AST.binaryExpression('+', AST.integerLiteral(2), AST.integerLiteral(3));
    expect(I.evaluate(add)).toEqual({ kind: 'i32', value: 5 });
    const fadd = AST.binaryExpression('+', AST.floatLiteral(1.5), AST.floatLiteral(2.25));
    expect(I.evaluate(fadd)).toEqual({ kind: 'f64', value: 3.75 });
    const s = AST.binaryExpression('+', AST.stringLiteral('a'), AST.stringLiteral('b'));
    expect(I.evaluate(s)).toEqual({ kind: 'string', value: 'ab' });
  });

  test("logical && || with short-circuit", () => {
    const and = AST.binaryExpression('&&', AST.booleanLiteral(true), AST.booleanLiteral(false));
    expect(I.evaluate(and)).toEqual({ kind: 'bool', value: false });
    const or = AST.binaryExpression('||', AST.booleanLiteral(false), AST.booleanLiteral(true));
    expect(I.evaluate(or)).toEqual({ kind: 'bool', value: true });
  });

  test("comparisons", () => {
    expect(I.evaluate(AST.binaryExpression('>', AST.integerLiteral(3), AST.integerLiteral(2)))).toEqual({ kind: 'bool', value: true });
    expect(I.evaluate(AST.binaryExpression('==', AST.stringLiteral('a'), AST.stringLiteral('a')))).toEqual({ kind: 'bool', value: true });
  });

  test("bitwise", () => {
    expect(I.evaluate(AST.binaryExpression('&', AST.integerLiteral(6), AST.integerLiteral(3)))).toEqual({ kind: 'i32', value: 2 });
    expect(I.evaluate(AST.binaryExpression('|', AST.integerLiteral(4), AST.integerLiteral(1)))).toEqual({ kind: 'i32', value: 5 });
  });

  test("range expression", () => {
    const r1 = I.evaluate(AST.rangeExpression(AST.integerLiteral(1), AST.integerLiteral(3), true));
    expect(r1).toEqual({ kind: 'range', start: 1, end: 3, inclusive: true });
    const r2 = I.evaluate(AST.rangeExpression(AST.floatLiteral(0.0), AST.floatLiteral(1.0), false));
    expect(r2).toEqual({ kind: 'range', start: 0, end: 1, inclusive: false });
  });
});


