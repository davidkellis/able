import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - rethrow", () => {
  test("inner rescue rethrows to outer rescue", () => {
    const I = new InterpreterV10();
    const inner = AST.rescueExpression(
      AST.blockExpression([AST.raiseStatement(AST.stringLiteral("oops"))]),
      [AST.matchClause(AST.wildcardPattern(), AST.blockExpression([AST.rethrowStatement()]))]
    );
    const outer = AST.rescueExpression(
      inner,
      [AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("handled"))]
    );
    expect(I.evaluate(outer)).toEqual({ kind: 'string', value: 'handled' });
  });
});

