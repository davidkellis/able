import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - module", () => {
  test("module evaluates definitions and top-level statements", () => {
    const I = new InterpreterV10();
    const fn = AST.functionDefinition(
      "add1",
      [AST.functionParameter("x")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(1)))])
    );
    const mod = AST.module([
      fn,
      AST.assignmentExpression(":=", AST.identifier("y"), AST.functionCall(AST.identifier("add1"), [AST.integerLiteral(4)])),
    ]);
    const result = I.evaluate(mod as any);
    expect(I.evaluate(AST.identifier("y"))).toEqual({ kind: 'i32', value: 5 });
    expect(result).toEqual({ kind: 'i32', value: 5 });
  });
});


