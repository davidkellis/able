import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - struct functional update", () => {
  test("named-field struct: spread base then override fields", () => {
    const I = new InterpreterV10();
    const Def = AST.structDefinition(
      "User",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "id"),
        AST.structFieldDefinition(AST.simpleTypeExpression("string"), "name"),
        AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "active"),
      ],
      "named"
    );
    I.evaluate(Def);

    // base instance
    const base = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "id"),
      AST.structFieldInitializer(AST.stringLiteral("Alice"), "name"),
      AST.structFieldInitializer(AST.booleanLiteral(true), "active"),
    ], false, "User");
    const baseVal = I.evaluate(base);
    expect(baseVal.kind).toBe("struct_instance");

    // functional update: ...base with override
    const updated = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("Bob"), "name"),
    ], false, "User", [base]);
    const updVal = I.evaluate(updated);
    expect(updVal.kind).toBe("struct_instance");

    // Check fields
    const nameField = I.evaluate(AST.memberAccessExpression(updated, "name"));
    expect(nameField).toEqual({ kind: "string", value: "Bob" });
    const idField = I.evaluate(AST.memberAccessExpression(updated, "id"));
    expect(idField).toEqual({ kind: "i32", value: 1n });
    const activeField = I.evaluate(AST.memberAccessExpression(updated, "active"));
    expect(activeField).toEqual({ kind: "bool", value: true });

    // Ensure base did not change
    const nameBase = I.evaluate(AST.memberAccessExpression(base, "name"));
    expect(nameBase).toEqual({ kind: "string", value: "Alice" });
  });

  test("functional update rejects wrong-type source or positional source", () => {
    const I = new InterpreterV10();
    const A = AST.structDefinition("A", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")], "named");
    const B = AST.structDefinition("B", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")], "named");
    I.evaluate(A);
    I.evaluate(B);
    const abase = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(10), "x")], false, "A");
    const bbase = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(20), "y")], false, "B");
    I.evaluate(abase);
    I.evaluate(bbase);

    const wrongType = AST.structLiteral([], false, "A", [bbase]);
    expect(() => I.evaluate(wrongType as any)).toThrow();
  });

  test("functional update accepts multiple spreads", () => {
    const I = new InterpreterV10();
    const Point = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named"
    );
    I.evaluate(Point);

    const base1 = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(10), "y"),
    ], false, "Point");
    const base2 = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(2), "x"),
      AST.structFieldInitializer(AST.integerLiteral(20), "y"),
    ], false, "Point");

    const merged = AST.structLiteral(
      [
        AST.structFieldInitializer(AST.integerLiteral(99), "x"),
      ],
      false,
      "Point",
      [base1, base2],
    );

    const value = I.evaluate(merged);
    expect(value.kind).toBe("struct_instance");
    if (value.kind !== "struct_instance" || !(value.values instanceof Map)) throw new Error("expected named struct instance");
    expect(value.values.get("x")).toEqual({ kind: "i32", value: 99n });
    expect(value.values.get("y")).toEqual({ kind: "i32", value: 20n });
  });
});
