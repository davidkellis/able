import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - break statement", () => {
  test("break exits while loop", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("i"), AST.integerLiteral(0)));
    const loop = AST.whileLoop(
      AST.booleanLiteral(true),
      AST.blockExpression([
        AST.assignmentExpression("=", AST.identifier("i"), AST.binaryExpression("+", AST.identifier("i"), AST.integerLiteral(1))),
        AST.ifExpression(
          AST.binaryExpression("==", AST.identifier("i"), AST.integerLiteral(3)),
          AST.blockExpression([AST.breakStatement(undefined, AST.integerLiteral(0))]),
          []
        )
      ])
    );
    I.evaluate(loop);
    expect(I.evaluate(AST.identifier("i"))).toEqual({ kind: 'i32', value: 3 });
  });
});

