import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

const asStatement = (node: AST.Node): AST.Statement => node as unknown as AST.Statement;

describe("typechecker numeric promotion", () => {
  test("promotes signed operands to the wider width", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("i64")),
          AST.binaryExpression("+", AST.integerLiteral(1, "i16"), AST.integerLiteral(2, "i64")),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("promotes mixed signed/unsigned integers to a compatible signed type", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("i32")),
          AST.binaryExpression("+", AST.integerLiteral(1, "i8"), AST.integerLiteral(2, "u16")),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("falls back to u128 when mixing u128 with smaller signed integers", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.typedPattern(AST.identifier("sum"), AST.simpleTypeExpression("u128")),
          AST.binaryExpression("+", AST.integerLiteral(1, "u128"), AST.integerLiteral(2, "i32")),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("promotes mixed i128/u64 without exceeding available widths", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.typedPattern(AST.identifier("sum"), AST.simpleTypeExpression("i128")),
          AST.binaryExpression("+", AST.integerLiteral(1, "i128"), AST.integerLiteral(2, "u64")),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("bitwise operations require integer operands", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.identifier("value"),
          AST.binaryExpression(".&", AST.integerLiteral(1, "i32"), AST.stringLiteral("nope")),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("requires integer operands");
  });

  test("comparison diagnostics mention numeric requirement when not comparing strings", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(AST.assignmentExpression(":=", AST.identifier("foo"), AST.stringLiteral("hey"))),
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.identifier("flag"),
          AST.binaryExpression("<", AST.identifier("foo"), AST.integerLiteral(1)),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("requires numeric operands");
  });

  test("/ promotes integer operands to f64", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("f64")),
          AST.binaryExpression("/", AST.integerLiteral(5), AST.integerLiteral(2)),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("// and %% require integer operands", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.identifier("badQuot"),
          AST.binaryExpression("//", AST.integerLiteral(5), AST.floatLiteral(2.0)),
        ),
      ),
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.identifier("badRem"),
          AST.binaryExpression("%%", AST.floatLiteral(5.0), AST.integerLiteral(2)),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(2);
    expect(diagnostics[0]?.message).toContain("requires integer operands");
    expect(diagnostics[1]?.message).toContain("requires integer operands");
  });

  test("/% yields DivMod struct with integer type argument", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      asStatement(
        AST.assignmentExpression(
          ":=",
          AST.typedPattern(
            AST.identifier("pair"),
            AST.genericTypeExpression(AST.simpleTypeExpression("DivMod"), [AST.simpleTypeExpression("i32")]),
          ),
          AST.binaryExpression("/%", AST.integerLiteral(7), AST.integerLiteral(3)),
        ),
      ),
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });
});
