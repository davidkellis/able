import { describe, expect, test } from "bun:test";
import { formatType, primitiveType, type TypeInfo } from "../../src/typechecker/types";

describe("TypeInfo formatting", () => {
  test("formats struct with type arguments using spec spacing", () => {
    const arrayOfI32: TypeInfo = {
      kind: "struct",
      name: "Array",
      typeArguments: [primitiveType("i32")],
    };
    expect(formatType(arrayOfI32)).toBe("Array i32");
  });

  test("formats nullable and result types", () => {
    const nullableString: TypeInfo = { kind: "nullable", inner: primitiveType("String") };
    const resultString: TypeInfo = { kind: "result", inner: primitiveType("String") };
    expect(formatType(nullableString)).toBe("String?");
    expect(formatType(resultString)).toBe("Result String");
  });

  test("formats unions and nested generics", () => {
    const optionStruct: TypeInfo = {
      kind: "struct",
      name: "Option",
      typeArguments: [{ kind: "nullable", inner: primitiveType("char") }],
    };
    const union: TypeInfo = {
      kind: "union",
      members: [primitiveType("i32"), optionStruct],
    };
    expect(formatType(union)).toBe("i32 | Option char?");
  });
});
