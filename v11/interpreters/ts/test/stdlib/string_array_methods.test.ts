import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

const methodCall = (receiver: AST.Expression, name: string, args: AST.Expression[] = []) =>
  AST.functionCall(AST.memberAccessExpression(receiver, name), args);

const readInteger = (value: any): number => Number(value?.value ?? 0);

describe("array helpers", () => {
  test("native array helpers are absent without stdlib import", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("arr"), AST.arrayLiteral([AST.integerLiteral(1)])));
    expect(() => I.evaluate(methodCall(AST.identifier("arr"), "size"))).toThrow();
    expect(() => I.evaluate(methodCall(AST.identifier("arr"), "push", [AST.integerLiteral(2)]))).toThrow();
  });
});

describe("String helpers", () => {
  test("native String helpers are absent without stdlib import", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("s"), AST.stringLiteral("hi")));
    expect(() => I.evaluate(methodCall(AST.identifier("s"), "len_bytes"))).toThrow();
    expect(() => I.evaluate(methodCall(AST.identifier("s"), "split", [AST.stringLiteral("")]))).toThrow();
  });
});
