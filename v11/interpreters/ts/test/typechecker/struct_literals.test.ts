import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker struct literals", () => {
  test("reports literal overflow inside struct field", () => {
    const checker = new TypeChecker();
    const structDef = AST.structDefinition(
      "Packet",
      [AST.structFieldDefinition(AST.simpleTypeExpression("u8"), "value")],
      "named",
    );
    const literal = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(512), "value")],
      false,
      "Packet",
    );
    const module = AST.module([structDef, literal as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("reports type mismatch for named struct field", () => {
    const checker = new TypeChecker();
    const structDef = AST.structDefinition(
      "Carrier",
      [AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "flag")],
      "named",
    );
    const literal = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(1), "flag")],
      false,
      "Carrier",
    );
    const module = AST.module([structDef, literal as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("struct field 'flag' expects type bool");
  });

  test("reports mismatch for positional struct fields", () => {
    const checker = new TypeChecker();
    const structDef = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
      ],
      "positional",
    );
    const literal = AST.structLiteral(
      [
        AST.structFieldInitializer(AST.integerLiteral(1)),
        AST.structFieldInitializer(AST.stringLiteral("oops")),
      ],
      true,
      "Point",
    );
    const module = AST.module([structDef, literal as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("struct field '#1' expects type i32");
  });
});
