import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - if/or", () => {
  const I = new InterpreterV10();

  test("if true selects first branch", () => {
    const expr = AST.ifExpression(
      AST.booleanLiteral(true),
      AST.blockExpression([AST.integerLiteral(1)]),
      []
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'i32', value: 1n });
  });

  test("if false with or condition and else", () => {
    const expr = AST.ifExpression(
      AST.booleanLiteral(false),
      AST.blockExpression([AST.integerLiteral(1)]),
      [
        AST.orClause(AST.blockExpression([AST.integerLiteral(2)]), AST.booleanLiteral(false)),
        AST.orClause(AST.blockExpression([AST.integerLiteral(3)])),
      ]
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'i32', value: 3n });
  });
});


