import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - for loop destructuring", () => {
  test("array destructuring in loop", () => {
    const I = new InterpreterV10();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("sum"), AST.integerLiteral(0)));
    const loop = AST.forLoop(
      AST.arrayPattern([AST.identifier("a"), AST.identifier("b")]),
      AST.arrayLiteral([
        AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2)]),
        AST.arrayLiteral([AST.integerLiteral(3), AST.integerLiteral(4)]),
      ]),
      AST.blockExpression([
        AST.assignmentExpression("=", AST.identifier("sum"), AST.binaryExpression("+", AST.identifier("sum"), AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b"))))
      ])
    );
    I.evaluate(loop);
    expect(I.evaluate(AST.identifier("sum"))).toEqual({ kind: 'i32', value: 10 });
  });

  test("struct destructuring in loop", () => {
    const I = new InterpreterV10();
    const P = AST.structDefinition("P", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")], "named");
    I.evaluate(P);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("sum"), AST.integerLiteral(0)));
    const arr = AST.arrayLiteral([
      AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x"), AST.structFieldInitializer(AST.integerLiteral(2), "y")], false, "P"),
      AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(3), "x"), AST.structFieldInitializer(AST.integerLiteral(4), "y")], false, "P"),
    ]);
    const loop = AST.forLoop(
      AST.structPattern([AST.structPatternField(AST.identifier("x"), "x"), AST.structPatternField(AST.identifier("y"), "y")], false, "P"),
      arr,
      AST.blockExpression([
        AST.assignmentExpression("=", AST.identifier("sum"), AST.binaryExpression("+", AST.identifier("sum"), AST.binaryExpression("+", AST.identifier("x"), AST.identifier("y"))))
      ])
    );
    I.evaluate(loop);
    expect(I.evaluate(AST.identifier("sum"))).toEqual({ kind: 'i32', value: 10 });
  });
});


