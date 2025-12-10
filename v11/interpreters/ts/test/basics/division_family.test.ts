import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - division operators", () => {
  test("/ promotes integer operands to f64", () => {
    const I = new InterpreterV10();
    const expr = AST.binaryExpression("/", AST.integerLiteral(5), AST.integerLiteral(2));
    expect(I.evaluate(expr)).toEqual({ kind: "f64", value: 2.5 });
  });

  test("// and % follow Euclidean semantics for integers", () => {
    const I = new InterpreterV10();
    const quotient = AST.binaryExpression("//", AST.integerLiteral(-5), AST.integerLiteral(3));
    const remainder = AST.binaryExpression("%", AST.integerLiteral(-5), AST.integerLiteral(3));
    expect(I.evaluate(quotient)).toEqual({ kind: "i32", value: -2n });
    expect(I.evaluate(remainder)).toEqual({ kind: "i32", value: 1n });
  });

  test("/% returns DivMod struct with quotient and remainder", () => {
    const I = new InterpreterV10();
    const expr = AST.binaryExpression("/%", AST.integerLiteral(7), AST.integerLiteral(3));
    const value = I.evaluate(expr);
    expect(value.kind).toBe("struct_instance");
    if (value.kind !== "struct_instance" || !(value.values instanceof Map)) {
      throw new Error("expected named struct instance");
    }
    expect(value.def.id.name).toBe("DivMod");
    expect(value.typeArguments?.[0]).toEqual(AST.simpleTypeExpression("i32"));
    expect(value.values.get("quotient")).toEqual({ kind: "i32", value: 2n });
    expect(value.values.get("remainder")).toEqual({ kind: "i32", value: 1n });
  });

  test("division family rejects zero divisors", () => {
    const I = new InterpreterV10();
    expect(() =>
      I.evaluate(AST.binaryExpression("%", AST.integerLiteral(4), AST.integerLiteral(0))),
    ).toThrow(/division by zero/i);
    expect(() =>
      I.evaluate(AST.binaryExpression("//", AST.integerLiteral(4), AST.integerLiteral(0))),
    ).toThrow(/division by zero/i);
    expect(() =>
      I.evaluate(AST.binaryExpression("/%", AST.integerLiteral(4), AST.integerLiteral(0))),
    ).toThrow(/division by zero/i);
    expect(() =>
      I.evaluate(AST.binaryExpression("/", AST.integerLiteral(4), AST.integerLiteral(0))),
    ).toThrow(/division by zero/i);
  });
});
