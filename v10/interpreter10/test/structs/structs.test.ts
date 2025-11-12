import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - structs", () => {
  test("named struct literal and member access", () => {
    const I = new InterpreterV10();
    // struct Point { x: i32, y: i32 }
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);
    // p := Point { x: 2, y: 3 }
    const pLit = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(2), "x"),
      AST.structFieldInitializer(AST.integerLiteral(3), "y"),
    ], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), pLit));
    // p.x == 2
    const px = AST.memberAccessExpression(AST.identifier("p"), "x");
    expect(I.evaluate(px)).toEqual({ kind: 'i32', value: 2 });
  });

  test("positional struct literal and index access", () => {
    const I = new InterpreterV10();
    // struct Color (i32, i32, i32)
    const colorDef = AST.structDefinition(
      "Color",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32")), AST.structFieldDefinition(AST.simpleTypeExpression("i32")), AST.structFieldDefinition(AST.simpleTypeExpression("i32"))],
      "positional"
    );
    I.evaluate(colorDef);
    const cLit = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(255)),
      AST.structFieldInitializer(AST.integerLiteral(0)),
      AST.structFieldInitializer(AST.integerLiteral(128)),
    ], true, "Color");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("c"), cLit));
    const g = AST.memberAccessExpression(AST.identifier("c"), AST.integerLiteral(1));
    expect(I.evaluate(g)).toEqual({ kind: 'i32', value: 0 });
  });
});


