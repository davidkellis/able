import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10, Environment } from "../../src/interpreter";

describe("v11 interpreter - assignments & blocks", () => {
  test(":= defines in current scope; = reassigns outer; redeclare errors", () => {
    const I = new InterpreterV10();
    const env = I.globals;

    // x := 1
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(1)), env);
    expect(env.get("x")).toEqual({ kind: "i32", value: 1n });

    // x = 2
    I.evaluate(AST.assignmentExpression("=", AST.identifier("x"), AST.integerLiteral(2)), env);
    expect(env.get("x")).toEqual({ kind: "i32", value: 2n });

    // Redeclare x in same scope should error
    expect(() => I.evaluate(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(3)), env)).toThrow();
  });

  test("block creates a new scope; inner := shadows; outer unchanged", () => {
    const I = new InterpreterV10();
    const env = I.globals;
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("y"), AST.integerLiteral(10)), env);

    const block = AST.blockExpression([
      AST.assignmentExpression(":=", AST.identifier("y"), AST.integerLiteral(20)),
      AST.identifier("y"),
    ]);
    const result = I.evaluate(block, env);
    expect(result).toEqual({ kind: "i32", value: 20n }); // inner value
    expect(env.get("y")).toEqual({ kind: "i32", value: 10n }); // outer unchanged
  });

  test("= creates a binding when none exists", () => {
    const I = new InterpreterV10();
    const env = I.globals;
    I.evaluate(AST.assignmentExpression("=", AST.identifier("fresh"), AST.integerLiteral(42)), env);
    expect(env.get("fresh")).toEqual({ kind: "i32", value: 42n });
    // subsequent = behaves like reassignment
    I.evaluate(AST.assignmentExpression("=", AST.identifier("fresh"), AST.integerLiteral(7)), env);
    expect(env.get("fresh")).toEqual({ kind: "i32", value: 7n });
  });
});

