import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - generic type arguments (accepted, no typecheck)", () => {
  test("function call with typeArguments is ignored at runtime but accepted", () => {
    const I = new InterpreterV10();
    const id = AST.functionDefinition(
      "id",
      [AST.functionParameter("x")],
      AST.blockExpression([AST.returnStatement(AST.identifier("x"))])
    );
    I.evaluate(id);
    const call = AST.functionCall(AST.identifier("id"), [AST.integerLiteral(9)], [AST.simpleTypeExpression("i32")]);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 9n });
  });
});


