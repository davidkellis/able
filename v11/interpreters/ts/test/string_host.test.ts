import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { Interpreter } from "../src/interpreter";

const call = (name: string, args = []) =>
  AST.functionCall(AST.identifier(name), args);

describe("String host builtins", () => {
  test("__able_String_from_builtin returns UTF-8 bytes", () => {
    const I = new Interpreter();
    const result = I.evaluate(call("__able_String_from_builtin", [AST.stringLiteral("Hi")])) as any;
    expect(result.kind).toBe("array");
    expect(result.elements.map((el: any) => Number(el.value))).toEqual([72, 105]);
  });

  test("__able_String_to_builtin decodes UTF-8 arrays", () => {
    const I = new Interpreter();
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("bytes"),
        AST.arrayLiteral([AST.integerLiteral(0xE2), AST.integerLiteral(0x82), AST.integerLiteral(0xAC)]),
      ),
    );
    const result = I.evaluate(call("__able_String_to_builtin", [AST.identifier("bytes")])) as any;
    expect(result).toEqual({ kind: "String", value: "€" });
  });

  test("__able_char_from_codepoint builds chars", () => {
    const I = new Interpreter();
    const charVal = I.evaluate(call("__able_char_from_codepoint", [AST.integerLiteral(0x1F600)])) as any;
    expect(charVal).toEqual({ kind: "char", value: "😀" });
  });
});
