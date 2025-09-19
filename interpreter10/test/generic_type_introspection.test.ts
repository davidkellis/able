import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - generic type arg introspection (bind T_type)", () => {
  test("function can read T_type bound as string", () => {
    const I = new InterpreterV10();
    const showT = AST.functionDefinition(
      "showT",
      [AST.functionParameter("x")],
      AST.blockExpression([
        AST.returnStatement(AST.identifier("T_type")),
      ]),
      AST.simpleTypeExpression("string"),
      [AST.genericParameter("T")]
    );
    I.evaluate(showT);

    const call1 = AST.functionCall(AST.identifier("showT"), [AST.integerLiteral(1)], [AST.simpleTypeExpression("i32")]);
    expect(I.evaluate(call1)).toEqual({ kind: "string", value: "i32" });

    const call2 = AST.functionCall(AST.identifier("showT"), [AST.floatLiteral(1.5)], [
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("i32")])
    ]);
    expect(I.evaluate(call2)).toEqual({ kind: "string", value: "Array<i32>" });
  });
});


