import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("loop reassignment regression", () => {
  test("loop expression updates existing binding with =", () => {
    const I = new InterpreterV10();
    const env = I.globals;

    // a = 5
    I.evaluate(AST.assignmentExpression("=", AST.identifier("a"), AST.integerLiteral(5)), env);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("guard"), AST.integerLiteral(0)), env);

    const loop = AST.loopExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("guard"),
          AST.binaryExpression("+", AST.identifier("guard"), AST.integerLiteral(1)),
        ),
        AST.ifExpression(
          AST.binaryExpression(">=", AST.identifier("guard"), AST.integerLiteral(20)),
          AST.blockExpression([AST.raiseStatement(AST.stringLiteral("guard exceeded"))]),
          [],
        ),
        AST.ifExpression(
          AST.binaryExpression("<=", AST.identifier("a"), AST.integerLiteral(0)),
          AST.blockExpression([AST.breakStatement(undefined, AST.identifier("a"))]),
          [],
        ),
        AST.assignmentExpression("=", AST.identifier("a"), AST.binaryExpression("-", AST.identifier("a"), AST.integerLiteral(1))),
      ]),
    );

    const loopResult = I.evaluate(loop, env);
    expect(loopResult).toEqual({ kind: "i32", value: 0n });
    expect(I.evaluate(AST.identifier("a"), env)).toEqual({ kind: "i32", value: 0n });
  });
});
