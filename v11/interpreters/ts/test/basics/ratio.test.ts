import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

const ratioLiteral = (num: number, den: number) =>
  AST.structLiteral(
    [
      AST.structFieldInitializer(AST.integerLiteral(BigInt(num), "i64"), "num"),
      AST.structFieldInitializer(AST.integerLiteral(BigInt(den), "i64"), "den"),
    ],
    false,
    "Ratio",
  );

describe("v11 interpreter - Ratio support", () => {
  test("arithmetic with Ratio stays exact", () => {
    const I = new InterpreterV10();
    I.ensureRatioStruct();
    const half = ratioLiteral(1, 2);
    const quarter = ratioLiteral(1, 4);
    const expr = AST.binaryExpression("+", half, quarter);
    const value = I.evaluate(expr);
    expect(value.kind).toBe("struct_instance");
    if (value.kind !== "struct_instance" || !(value.values instanceof Map)) {
      throw new Error("expected Ratio instance");
    }
    expect(value.values.get("num")).toEqual({ kind: "i64", value: 3n });
    expect(value.values.get("den")).toEqual({ kind: "i64", value: 4n });
  });

  test("Ratio mixes with integers and floats", () => {
    const I = new InterpreterV10();
    I.ensureRatioStruct();
    const third = ratioLiteral(1, 3);
    const expr = AST.binaryExpression("+", third, AST.floatLiteral(0.5));
    const value = I.evaluate(expr);
    expect(value.kind).toBe("struct_instance");
    if (value.kind !== "struct_instance" || !(value.values instanceof Map)) {
      throw new Error("expected Ratio instance");
    }
    // 1/3 + 1/2 = 5/6
    expect(value.values.get("num")).toEqual({ kind: "i64", value: 5n });
    expect(value.values.get("den")).toEqual({ kind: "i64", value: 6n });
  });

  test("Ratio division by zero raises", () => {
    const I = new InterpreterV10();
    I.ensureRatioStruct();
    const half = ratioLiteral(1, 2);
    const zero = ratioLiteral(0, 1);
    expect(() => I.evaluate(AST.binaryExpression("/", half, zero))).toThrow(/division by zero/i);
  });

  test("builtin float->Ratio conversion preserves fraction", () => {
    const I = new InterpreterV10();
    I.ensureRatioStruct();
    const expr = AST.functionCall(AST.identifier("__able_ratio_from_float"), [AST.floatLiteral(0.25)]);
    const value = I.evaluate(expr);
    expect(value.kind).toBe("struct_instance");
    if (value.kind !== "struct_instance" || !(value.values instanceof Map)) {
      throw new Error("expected Ratio instance");
    }
    expect(value.values.get("num")).toEqual({ kind: "i64", value: 1n });
    expect(value.values.get("den")).toEqual({ kind: "i64", value: 4n });
  });
});
