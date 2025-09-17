import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - breakpoint expression", () => {
  test("evaluates body and returns last value", () => {
    const I = new InterpreterV10();
    const body = AST.blockExpression([
      AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(1)),
      AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(2))
    ]);
    const bp = AST.breakpointExpression("dbg", body);
    expect(I.evaluate(bp)).toEqual({ kind: 'i32', value: 3 });
  });
});


