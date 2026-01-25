import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - bitshift range checks", () => {
  test("left/right shift reject out-of-range counts", () => {
    const I = new Interpreter();
    const bad1 = AST.binaryExpression(".<<", AST.integerLiteral(1), AST.integerLiteral(32));
    expect(() => I.evaluate(bad1 as any)).toThrow();
    const bad2 = AST.binaryExpression(".>>", AST.integerLiteral(1), AST.integerLiteral(33));
    expect(() => I.evaluate(bad2 as any)).toThrow();
  });

  test("compound shift also enforces range", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(1)));
    expect(() => I.evaluate(AST.assignmentExpression(".<<=", AST.identifier("x"), AST.integerLiteral(32)))).toThrow();
    // valid shift
    I.evaluate(AST.assignmentExpression(".<<=", AST.identifier("x"), AST.integerLiteral(3)));
    expect(I.evaluate(AST.identifier("x"))).toEqual({ kind: "i32", value: 8n });
  });
});
