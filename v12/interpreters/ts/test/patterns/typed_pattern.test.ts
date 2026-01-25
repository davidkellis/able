import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - typed patterns", () => {
  test("typed pattern in match filters by type", () => {
    const I = new Interpreter();
    const pat = AST.typedPattern(AST.identifier("s"), AST.simpleTypeExpression("String"));
    const clause = AST.matchClause(pat, AST.identifier("s"));
    const m = AST.matchExpression(AST.stringLiteral("ok"), [clause]);
    expect(I.evaluate(m)).toEqual({ kind: 'String', value: 'ok' });
  });

  test("typed pattern in assignment enforces type", () => {
    const I = new Interpreter();
    const pat = AST.typedPattern(AST.identifier("n"), AST.simpleTypeExpression("i32"));
    I.evaluate(AST.assignmentExpression(":=", pat as any, AST.integerLiteral(5)));
    expect(I.evaluate(AST.identifier("n"))).toEqual({ kind: 'i32', value: 5n });
  });

  test("typed assignment widens integer values to annotated type", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("value"), AST.integerLiteral(5)) as any);
    const pat = AST.typedPattern(AST.identifier("wide"), AST.simpleTypeExpression("i64"));
    I.evaluate(AST.assignmentExpression(":=", pat as any, AST.identifier("value")));
    expect(I.evaluate(AST.identifier("wide"))).toEqual({ kind: 'i64', value: 5n });
  });
});

