import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - import alias selectors", () => {
  test("import selector with alias binds in module env", () => {
    const I = new InterpreterV10();
    // Predefine a function in globals
    I.evaluate(AST.functionDefinition("foo", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(42))])));

    const imp = AST.importStatement(["core" as any], false, [AST.importSelector("foo", "bar")]);
    const mod = AST.module([
      imp,
      AST.assignmentExpression(":=", AST.identifier("x"), AST.functionCall(AST.identifier("bar"), []))
    ]);
    const res = I.evaluate(mod as any);
    expect(I.evaluate(AST.identifier("x"))).toEqual({ kind: 'i32', value: 42 });
    expect(res).toEqual({ kind: 'i32', value: 42 });
  });
});


