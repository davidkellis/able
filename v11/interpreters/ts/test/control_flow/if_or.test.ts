import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - if/elsif", () => {
  const I = new Interpreter();

  test("if true selects first branch", () => {
    const expr = AST.ifExpression(
      AST.booleanLiteral(true),
      AST.blockExpression([AST.integerLiteral(1)]),
      []
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'i32', value: 1n });
  });

  test("if false with elsif condition and else", () => {
    const expr = AST.ifExpression(
      AST.booleanLiteral(false),
      AST.blockExpression([AST.integerLiteral(1)]),
      [
        AST.elseIfClause(AST.booleanLiteral(false), AST.blockExpression([AST.integerLiteral(2)])),
      ],
      AST.blockExpression([AST.integerLiteral(3)])
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'i32', value: 3n });
  });
});

