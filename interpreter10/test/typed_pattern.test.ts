import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - typed patterns", () => {
  test("typed pattern in match filters by type", () => {
    const I = new InterpreterV10();
    const pat = AST.typedPattern(AST.identifier("s"), AST.simpleTypeExpression("string"));
    const clause = AST.matchClause(pat, AST.identifier("s"));
    const m = AST.matchExpression(AST.stringLiteral("ok"), [clause]);
    expect(I.evaluate(m)).toEqual({ kind: 'string', value: 'ok' });
  });

  test("typed pattern in assignment enforces type", () => {
    const I = new InterpreterV10();
    const pat = AST.typedPattern(AST.identifier("n"), AST.simpleTypeExpression("i32"));
    I.evaluate(AST.assignmentExpression(":=", pat as any, AST.integerLiteral(5)));
    expect(I.evaluate(AST.identifier("n"))).toEqual({ kind: 'i32', value: 5 });
  });
});


