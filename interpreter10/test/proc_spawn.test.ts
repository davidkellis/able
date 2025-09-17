import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - proc & spawn (sync placeholder)", () => {
  test("proc wraps a call and returns its result", () => {
    const I = new InterpreterV10();
    const fn = AST.functionDefinition(
      "add1",
      [AST.functionParameter("x")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(1)))])
    );
    I.evaluate(fn);
    const pr = AST.procExpression(AST.functionCall(AST.identifier("add1"), [AST.integerLiteral(10)]));
    expect(I.evaluate(pr)).toEqual({ kind: 'i32', value: 11 });
  });

  test("spawn wraps a block and returns last value", () => {
    const I = new InterpreterV10();
    const sp = AST.spawnExpression(AST.blockExpression([AST.integerLiteral(7)]));
    expect(I.evaluate(sp)).toEqual({ kind: 'i32', value: 7 });
  });
});


