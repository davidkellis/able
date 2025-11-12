import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - index read", () => {
  test("reads element by index", () => {
    const I = new InterpreterV10();
    const arr = AST.arrayLiteral([AST.integerLiteral(10), AST.integerLiteral(20), AST.integerLiteral(30)]);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("a"), arr));
    const v = I.evaluate(AST.indexExpression(AST.identifier("a"), AST.integerLiteral(1)));
    expect(v).toEqual({ kind: 'i32', value: 20 });
  });
});


