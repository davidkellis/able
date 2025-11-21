import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

const methodCall = (receiver: AST.Expression, name: string, args: AST.Expression[] = []) =>
  AST.functionCall(AST.memberAccessExpression(receiver, name), args);

const readInteger = (value: any): number => Number(value?.value ?? 0);

describe("array helpers", () => {
  test("push/pop/get/set/size/clear", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("arr"), AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)])));

    I.evaluate(methodCall(AST.identifier("arr"), "push", [AST.integerLiteral(3)]));
    const sizeAfterPush = I.evaluate(methodCall(AST.identifier("arr"), "size")) as any;
    expect(readInteger(sizeAfterPush)).toBe(3);

    const getMiddle = I.evaluate(methodCall(AST.identifier("arr"), "get", [AST.integerLiteral(1)])) as any;
    expect(readInteger(getMiddle)).toBe(2);

    const setResult = I.evaluate(methodCall(AST.identifier("arr"), "set", [AST.integerLiteral(0), AST.integerLiteral(9)])) as any;
    expect(setResult.kind).toBe("nil");

    const setError = I.evaluate(methodCall(AST.identifier("arr"), "set", [AST.integerLiteral(5), AST.integerLiteral(1)])) as any;
    expect(setError.kind).toBe("error");

    const popped = I.evaluate(methodCall(AST.identifier("arr"), "pop")) as any;
    expect(readInteger(popped)).toBe(3);

    I.evaluate(methodCall(AST.identifier("arr"), "clear"));
    const sizeAfterClear = I.evaluate(methodCall(AST.identifier("arr"), "size")) as any;
    expect(readInteger(sizeAfterClear)).toBe(0);
  });
});
