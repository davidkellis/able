import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - functions & lambdas", () => {
  test("define and call function returning sum", () => {
    const I = new InterpreterV10();
    const int = AST.integerLiteral;

    const sumFn = AST.functionDefinition(
      "sum",
      [AST.functionParameter("a"), AST.functionParameter("b")],
      AST.blockExpression([
        AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))
      ])
    );

    // define function
    I.evaluate(sumFn);
    // call sum(2, 3)
    const call = AST.functionCall(AST.identifier("sum"), [int(2), int(3)]);
    expect(I.evaluate(call)).toEqual({ kind: "i32", value: 5 });
  });

  test("lambda expression capturing outer variable (closure)", () => {
    const I = new InterpreterV10();
    // x := 10
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(10)));
    // lam = { |a| => a + x }
    const lam = AST.lambdaExpression(
      [AST.functionParameter("a")],
      AST.binaryExpression("+", AST.identifier("a"), AST.identifier("x"))
    );
    const lamVal = I.evaluate(lam);
    const call = AST.functionCall(lamVal as any, [AST.integerLiteral(5)]);
    // our functionCall expects callee as Expression; supply identifier bound to lambda:
    // bind lam in env
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("lam"), lam));
    const call2 = AST.functionCall(AST.identifier("lam"), [AST.integerLiteral(5)]);
    expect(I.evaluate(call2)).toEqual({ kind: "i32", value: 15 });
  });

  test("typed parameters are checked at runtime (minimal)", () => {
    const I = new InterpreterV10();
    const fn = AST.functionDefinition(
      "add",
      [AST.functionParameter(AST.identifier("a"), AST.simpleTypeExpression("i32")), AST.functionParameter(AST.identifier("b"), AST.simpleTypeExpression("i32"))],
      AST.blockExpression([
        AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))
      ])
    );
    I.evaluate(fn);

    const ok = AST.functionCall(AST.identifier("add"), [AST.integerLiteral(1), AST.integerLiteral(2)]);
    expect(I.evaluate(ok)).toEqual({ kind: 'i32', value: 3 });

    const bad = AST.functionCall(AST.identifier("add"), [AST.integerLiteral(1), AST.stringLiteral("x")]);
    expect(() => I.evaluate(bad)).toThrow();
  });
});
