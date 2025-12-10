import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

const call = (name: string, args = []) =>
  AST.functionCall(AST.identifier(name), args);

describe("String host builtins", () => {
  test("__able_String_from_builtin returns UTF-8 bytes", () => {
    const I = new InterpreterV10();
    const result = I.evaluate(call("__able_String_from_builtin", [AST.stringLiteral("Hi")])) as any;
    expect(result.kind).toBe("array");
    expect(result.elements.map((el: any) => Number(el.value))).toEqual([72, 105]);
  });

  test("__able_String_to_builtin decodes UTF-8 arrays", () => {
    const I = new InterpreterV10();
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
    const I = new InterpreterV10();
    const charVal = I.evaluate(call("__able_char_from_codepoint", [AST.integerLiteral(0x1F600)])) as any;
    expect(charVal).toEqual({ kind: "char", value: "😀" });
  });
});

describe("hasher host builtins", () => {
  test("__able_hasher_create/write/finish round trips hash state", () => {
    const I = new InterpreterV10();

    I.evaluate(AST.assignmentExpression(":=", AST.identifier("hasher"), call("__able_hasher_create")));
    I.evaluate(call("__able_hasher_write", [AST.identifier("hasher"), AST.stringLiteral("abc")]));

    const result = I.evaluate(call("__able_hasher_finish", [AST.identifier("hasher")])) as any;
    expect(result.kind).toBe("i32");
    const hash = Number(result.value);
    expect(hash >>> 0).toBe(0x1A47E90B);
  });
});
