import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - primitives", () => {
  const I = new InterpreterV10();

  test("string, bool, char, nil", () => {
    expect(I.evaluate(AST.stringLiteral("hi")).kind).toBe("string");
    expect(I.evaluate(AST.booleanLiteral(true))).toEqual({ kind: "bool", value: true });
    expect(I.evaluate(AST.charLiteral("x"))).toEqual({ kind: "char", value: "x" });
    expect(I.evaluate(AST.nilLiteral())).toEqual({ kind: "nil", value: null });
  });

  test("int defaults to i32 and float to f64", () => {
    expect(I.evaluate(AST.integerLiteral(123))).toEqual({ kind: "i32", value: 123 });
    expect(I.evaluate(AST.floatLiteral(1.5))).toEqual({ kind: "f64", value: 1.5 });
  });

  test("array literal evaluates elements", () => {
    const arr = AST.arrayLiteral([AST.integerLiteral(1), AST.stringLiteral("a")]);
    expect(I.evaluate(arr)).toEqual({ kind: "array", elements: [{ kind: "i32", value: 1 }, { kind: "string", value: "a" }] });
  });

  test("identifier and assignment (:= and =)", () => {
    const env = I.globals;
    // declare x := 1
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(1)), env);
    expect(env.get("x")).toEqual({ kind: "i32", value: 1 });
    // reassign x = 2
    I.evaluate(AST.assignmentExpression("=", AST.identifier("x"), AST.integerLiteral(2)), env);
    expect(env.get("x")).toEqual({ kind: "i32", value: 2 });
  });
});


