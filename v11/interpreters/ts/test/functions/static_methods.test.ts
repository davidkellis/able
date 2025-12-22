import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - static methods", () => {
  test("call static method on struct definition", () => {
    const I = new Interpreter();
    const pointDef = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(pointDef);

    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "origin",
          [],
          AST.blockExpression([
            AST.returnStatement(
              AST.structLiteral([
                AST.structFieldInitializer(AST.integerLiteral(0), "x"),
                AST.structFieldInitializer(AST.integerLiteral(0), "y"),
              ], false, "Point")
            )
          ])
        )
      ]
    );
    I.evaluate(methods);

    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("Point"), "origin"), []);
    const v = I.evaluate(call);
    expect(v.kind).toBe('struct_instance');
  });
});


