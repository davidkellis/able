import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";
import { primitiveType } from "../../src/typechecker/types";

describe("typechecker union normalization", () => {
  test("warns on redundant union members (including nullable expansion)", () => {
    const alias = AST.typeAliasDefinition(
      AST.identifier("MaybeInt"),
      AST.unionTypeExpression([
        AST.nullableTypeExpression(AST.simpleTypeExpression("i32")),
        AST.simpleTypeExpression("i32"),
      ]),
    );
    const module = AST.module([alias]);
    const checker = new TypeChecker();
    const result = checker.checkModule(module);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.severity).toBe("warning");
    expect(result.diagnostics[0]?.message).toContain("redundant union member i32");
  });

  test("collapses duplicate unions to a single member", () => {
    const checker = new TypeChecker();
    const normalizeUnionType = (checker as any).normalizeUnionType.bind(checker);
    const normalized = normalizeUnionType([primitiveType("i32"), primitiveType("i32")]);
    expect(normalized.kind).toBe("primitive");
  });

  test("collapses nil unions to nullable types", () => {
    const checker = new TypeChecker();
    const normalizeUnionType = (checker as any).normalizeUnionType.bind(checker);
    const normalized = normalizeUnionType([primitiveType("nil"), primitiveType("i32")]);
    expect(normalized.kind).toBe("nullable");
  });
});
