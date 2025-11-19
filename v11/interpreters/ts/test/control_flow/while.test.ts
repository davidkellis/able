import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - while loop", () => {
  test("while increments counter until condition fails", () => {
    const I = new InterpreterV10();
    // i := 0
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("i"), AST.integerLiteral(0)));
    // while (i < 3) { i = i + 1 }
    const loop = AST.whileLoop(
      AST.binaryExpression("<", AST.identifier("i"), AST.integerLiteral(3)),
      AST.blockExpression([
        AST.assignmentExpression("=", AST.identifier("i"), AST.binaryExpression("+", AST.identifier("i"), AST.integerLiteral(1)))
      ])
    );
    const loopResult = I.evaluate(loop);
    expect(loopResult).toEqual({ kind: "nil", value: null });
    // expect i == 3
    expect(I.evaluate(AST.identifier("i"))).toEqual({ kind: 'i32', value: 3n });
  });

  test("while loop returns break payload", () => {
    const I = new InterpreterV10();
    const loop = AST.whileLoop(
      AST.booleanLiteral(true),
      AST.blockExpression([
        AST.breakStatement(undefined, AST.integerLiteral(42)),
      ]),
    );
    const loopResult = I.evaluate(loop);
    expect(loopResult).toEqual({ kind: "i32", value: 42n });
  });
});

