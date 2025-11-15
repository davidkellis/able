import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker typed patterns", () => {
  test("does not emit diagnostics when annotation mismatches subject in assignment", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("string")),
        AST.integerLiteral(1),
      ) as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("reports diagnostic when := introduces no new bindings", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(1)) as unknown as AST.Statement,
      AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(2)) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain(":=");
  });

  test("assignment with = declares bindings when missing", () => {
    const checker = new TypeChecker();
    const structDef = AST.structDefinition(
      "Pair",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "first"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "second"),
      ],
      "named",
    );
    const literal = AST.structLiteral(
      [
        AST.structFieldInitializer(AST.integerLiteral(1), "first"),
        AST.structFieldInitializer(AST.integerLiteral(2), "second"),
      ],
      false,
      "Pair",
    );
    const pattern = AST.structPattern(
      [
        AST.structPatternField(AST.identifier("a"), "first"),
        AST.structPatternField(AST.identifier("b"), "second"),
      ],
      false,
      "Pair",
    );
    const module = AST.module([structDef, AST.assignmentExpression("=", pattern as any, literal) as any]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("reports diagnostic when literal does not fit annotated integer type", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("u8")),
        AST.integerLiteral(300),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 300 does not fit in u8");
  });

  test("typed array pattern adopts integer literals when they fit target type", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("values"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("u8")]),
        ),
        AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("typed array pattern reports literal overflow for nested elements", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("values"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("u8")]),
        ),
        AST.arrayLiteral([AST.integerLiteral(300)]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 300 does not fit in u8");
  });

  test("typed map pattern adopts nested literal structures", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("headers"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Map"), [
            AST.simpleTypeExpression("string"),
            AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("u8")]),
          ]),
        ),
        AST.mapLit([
          AST.mapEntry(AST.stringLiteral("ok"), AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)])),
        ]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("typed map pattern surfaces literal overflow within nested elements", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("headers"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Map"), [
            AST.simpleTypeExpression("string"),
            AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("u8")]),
          ]),
        ),
        AST.mapLit([
          AST.mapEntry(AST.stringLiteral("bad"), AST.arrayLiteral([AST.integerLiteral(512)])),
        ]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("typed range pattern reports literal overflow for endpoints", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("window"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Range"), [AST.simpleTypeExpression("u8")]),
        ),
        AST.rangeExpression(AST.integerLiteral(0), AST.integerLiteral(512), true),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("typed range pattern adopts literals when they fit target type", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("window"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Range"), [AST.simpleTypeExpression("u8")]),
        ),
        AST.rangeExpression(AST.integerLiteral(1), AST.integerLiteral(10), true),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("typed iterator pattern surfaces literal overflow for yielded values", () => {
    const checker = new TypeChecker();
    const iteratorLiteral = AST.iteratorLiteral([AST.yieldStatement(AST.integerLiteral(512))]);
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("iter"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Iterator"), [AST.simpleTypeExpression("u8")]),
        ),
        iteratorLiteral,
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("typed proc pattern reports literal overflow for result expressions", () => {
    const checker = new TypeChecker();
    const procBlock = AST.blockExpression([AST.integerLiteral(512) as unknown as AST.Statement]);
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("handle"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Proc"), [AST.simpleTypeExpression("u8")]),
        ),
        AST.procExpression(procBlock),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("typed future pattern reports literal overflow for spawn bodies", () => {
    const checker = new TypeChecker();
    const futureBlock = AST.blockExpression([AST.integerLiteral(512) as unknown as AST.Statement]);
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(
          AST.identifier("task"),
          AST.genericTypeExpression(AST.simpleTypeExpression("Future"), [AST.simpleTypeExpression("u8")]),
        ),
        AST.spawnExpression(futureBlock),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });
});
