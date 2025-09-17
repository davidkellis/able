import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";

describe("v10 interpreter - match", () => {
  test("match with identifier/wildcard/literal", () => {
    const I = new InterpreterV10();
    const subject = AST.integerLiteral(2);
    const m = AST.matchExpression(subject, [
      AST.matchClause(AST.literalPattern(AST.integerLiteral(1)), AST.integerLiteral(10)),
      AST.matchClause(AST.identifier("x"), AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(5))),
    ]);
    expect(I.evaluate(m)).toEqual({ kind: 'i32', value: 7 });
  });

  test("match struct named fields with guard", () => {
    const I = new InterpreterV10();
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);
    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point");
    const subject = p;
    const pat = AST.structPattern([
      AST.structPatternField(AST.identifier("a"), "x"),
      AST.structPatternField(AST.identifier("b"), "y"),
    ], false, "Point");
    const clause = AST.matchClause(pat, AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")), AST.binaryExpression(">", AST.identifier("b"), AST.identifier("a")));
    const m = AST.matchExpression(subject, [clause]);
    expect(I.evaluate(m)).toEqual({ kind: 'i32', value: 3 });
  });
});


