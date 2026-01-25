import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - rethrow", () => {
  test("inner rescue rethrows to outer rescue", () => {
    const I = new Interpreter();
    const inner = AST.rescueExpression(
      AST.blockExpression([AST.raiseStatement(AST.stringLiteral("oops"))]),
      [AST.matchClause(AST.wildcardPattern(), AST.blockExpression([AST.rethrowStatement()]))]
    );
    const outer = AST.rescueExpression(
      inner,
      [AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("handled"))]
    );
    expect(I.evaluate(outer)).toEqual({ kind: 'String', value: 'handled' });
  });
});

