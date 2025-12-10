import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - error handling", () => {
  test("raise and rescue with pattern", () => {
    const I = new InterpreterV10();
    // raise a simple number, rescue wildcard
    const expr = AST.rescueExpression(
      AST.blockExpression([AST.raiseStatement(AST.integerLiteral(42))]),
      [AST.matchClause(AST.wildcardPattern(), AST.integerLiteral(7))]
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'i32', value: 7n });
  });

  test("or-else binds error", () => {
    const I = new InterpreterV10();
    const expr = AST.orElseExpression(
      AST.propagationExpression(AST.blockExpression([AST.raiseStatement(AST.stringLiteral("x"))])),
      AST.blockExpression([AST.stringLiteral("handled")]),
      "e"
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'String', value: 'handled' });
  });

  test("ensure runs even on error", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("flag"), AST.stringLiteral("")));
    const expr = AST.ensureExpression(
      AST.rescueExpression(
        AST.blockExpression([AST.raiseStatement(AST.stringLiteral("oops"))]),
        [AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("rescued"))]
      ),
      AST.blockExpression([AST.assignmentExpression("=", AST.identifier("flag"), AST.stringLiteral("done"))])
    );
    expect(I.evaluate(expr)).toEqual({ kind: 'String', value: 'rescued' });
    expect(I.evaluate(AST.identifier("flag"))).toEqual({ kind: 'String', value: 'done' });
  });
});

