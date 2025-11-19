import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - loop expression", () => {
  test("loop expression returns break payload", () => {
    const I = new InterpreterV10();
    const loop = AST.loopExpression(
      AST.blockExpression([AST.breakStatement(undefined, AST.integerLiteral(42))]),
    );
    const result = I.evaluate(loop);
    expect(result).toEqual({ kind: "i32", value: 42n });
  });

  test("loop expression returns nil when break omits payload", () => {
    const I = new InterpreterV10();
    const loop = AST.loopExpression(AST.blockExpression([AST.breakStatement()]));
    const result = I.evaluate(loop);
    expect(result).toEqual({ kind: "nil", value: null });
  });

  test("loop expression honors continue before breaking", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("count"), AST.integerLiteral(0)));
    const loop = AST.loopExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("count"),
          AST.binaryExpression("+", AST.identifier("count"), AST.integerLiteral(1)),
        ),
        AST.ifExpression(
          AST.binaryExpression("<", AST.identifier("count"), AST.integerLiteral(3)),
          AST.blockExpression([AST.continueStatement()]),
          [],
        ),
        AST.breakStatement(undefined, AST.identifier("count")),
      ]),
    );

    const loopResult = I.evaluate(loop);
    expect(loopResult).toEqual({ kind: "i32", value: 3n });
  });

  test("loop expression can be used as a standalone statement", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("counter"), AST.integerLiteral(3)));

    const loop = AST.loopExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("counter"),
          AST.binaryExpression("-", AST.identifier("counter"), AST.integerLiteral(1)),
        ),
        AST.ifExpression(
          AST.binaryExpression("<", AST.identifier("counter"), AST.integerLiteral(0)),
          AST.blockExpression([AST.breakStatement()]),
          [],
        ),
      ]),
    );

    const block = AST.blockExpression([loop, AST.identifier("counter")]);
    const result = I.evaluate(block);
    expect(result).toEqual({ kind: "i32", value: -1n });
  });
});

