import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - for loop", () => {
  test("sum over array", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("sum"), AST.integerLiteral(0)));
    const arr = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2), AST.integerLiteral(3)]);
    const loop = AST.forLoop(
      AST.identifier("x"),
      arr,
      AST.blockExpression([
        AST.assignmentExpression("=", AST.identifier("sum"), AST.binaryExpression("+", AST.identifier("sum"), AST.identifier("x")))
      ])
    );
    I.evaluate(loop);
    expect(I.evaluate(AST.identifier("sum"))).toEqual({ kind: 'i32', value: 6 });
  });

  test("count down using range", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("last"), AST.integerLiteral(0)));
    const rng = AST.rangeExpression(AST.integerLiteral(3), AST.integerLiteral(1), true);
    const loop = AST.forLoop(
      AST.identifier("i"),
      rng,
      AST.blockExpression([
        AST.assignmentExpression("=", AST.identifier("last"), AST.identifier("i"))
      ])
    );
    I.evaluate(loop);
    expect(I.evaluate(AST.identifier("last"))).toEqual({ kind: 'i32', value: 1 });
  });

  test("continue skips matching elements", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("sum"), AST.integerLiteral(0)));
    const arr = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2), AST.integerLiteral(3)]);
    const loop = AST.forLoop(
      AST.identifier("x"),
      arr,
      AST.blockExpression([
        AST.ifExpression(
          AST.binaryExpression("==", AST.identifier("x"), AST.integerLiteral(2)),
          AST.blockExpression([AST.continueStatement()])
        ),
        AST.assignmentExpression(
          "=",
          AST.identifier("sum"),
          AST.binaryExpression("+", AST.identifier("sum"), AST.identifier("x"))
        ),
      ])
    );
    I.evaluate(loop);
    expect(I.evaluate(AST.identifier("sum"))).toEqual({ kind: 'i32', value: 4 });
  });
});

