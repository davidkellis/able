import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker iterator annotations", () => {
  test("reports diagnostic when yields violate the annotation", () => {
    const checker = new TypeChecker();
    const iterator = AST.iteratorLiteral(
      [
        AST.assignmentExpression(":=", AST.identifier("value"), AST.integerLiteral(1)),
        AST.yieldStatement(AST.identifier("value")),
      ],
      undefined,
      AST.simpleTypeExpression("String"),
    );
    const module = AST.module([iterator]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("iterator annotation");
  });

  test("accepts matching annotated yields", () => {
    const checker = new TypeChecker();
    const iterator = AST.iteratorLiteral(
      [
        AST.assignmentExpression(":=", AST.identifier("value"), AST.stringLiteral("ok")),
        AST.yieldStatement(AST.identifier("value")),
      ],
      undefined,
      AST.simpleTypeExpression("String"),
    );
    const module = AST.module([iterator]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("reports diagnostic when for-loop pattern annotation conflicts with iterator elements", () => {
    const checker = new TypeChecker();
    const iterator = AST.iteratorLiteral(
      [AST.yieldStatement(AST.integerLiteral(1))],
      undefined,
      AST.simpleTypeExpression("i32"),
    );
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("String")),
      iterator,
      AST.blockExpression([AST.identifier("value") as unknown as AST.Statement]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("for-loop pattern expects type String");
  });

  test("reports diagnostic when for-loop iterable is not iterable", () => {
    const checker = new TypeChecker();
    const loop = AST.forLoop(
      AST.identifier("value"),
      AST.integerLiteral(1),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("for-loop iterable must be array, range, String, or iterator");
  });

  test("reports diagnostic when array literal element type conflicts with typed pattern", () => {
    const checker = new TypeChecker();
    const arrayLiteral = AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)]);
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("String")),
      arrayLiteral,
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("for-loop pattern expects type String");
  });

  test("accepts typed pattern when array literal matches element type", () => {
    const checker = new TypeChecker();
    const arrayLiteral = AST.arrayLiteral([AST.stringLiteral("a"), AST.stringLiteral("b")]);
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("String")),
      arrayLiteral,
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("accepts String iterable when typed pattern expects u8", () => {
    const checker = new TypeChecker();
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("b"), AST.simpleTypeExpression("u8")),
      AST.stringLiteral("ok"),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("reports mismatch when String iterable is bound to non-byte typed pattern", () => {
    const checker = new TypeChecker();
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("b"), AST.simpleTypeExpression("String")),
      AST.stringLiteral("ok"),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
  });

  test("reports diagnostic when range expression element type conflicts with typed pattern", () => {
    const checker = new TypeChecker();
    const range = AST.rangeExpression(AST.integerLiteral(0), AST.integerLiteral(3), true);
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("String")),
      range,
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("for-loop pattern expects type String");
  });

  test("accepts typed pattern when range expression matches element type", () => {
    const checker = new TypeChecker();
    const range = AST.rangeExpression(AST.integerLiteral(0), AST.integerLiteral(3), true);
    const loop = AST.forLoop(
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("i32")),
      range,
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("accepts implicit generator binding inside iterator literals", () => {
    const checker = new TypeChecker();
    const iterator = AST.iteratorLiteral([
      AST.forLoop(
        AST.identifier("item"),
        AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)]),
        AST.blockExpression([
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("gen"), AST.identifier("yield")),
            [AST.identifier("item")],
          ) as unknown as AST.Statement,
        ]),
      ),
    ]);
    const module = AST.module([iterator]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

});
