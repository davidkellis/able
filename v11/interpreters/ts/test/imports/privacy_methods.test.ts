import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - method privacy", () => {
  test("private static method is not accessible; public static is", () => {
    const I = new InterpreterV10();
    const Point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"), AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y")],
      "named"
    );
    I.evaluate(Point);

    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "hidden_static",
          [],
          AST.blockExpression([
            AST.returnStatement(
              AST.structLiteral([
                AST.structFieldInitializer(AST.integerLiteral(0), "x"),
                AST.structFieldInitializer(AST.integerLiteral(0), "y"),
              ], false, "Point")
            )
          ]),
          undefined,
          undefined,
          undefined,
          false,
          true,
        ),
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
        ),
      ]
    );
    I.evaluate(methods);

    // Private static method access should throw
    const callHidden = AST.functionCall(AST.memberAccessExpression(AST.identifier("Point"), "hidden_static"), []);
    expect(() => I.evaluate(callHidden)).toThrow();

    // Public static method call should work
    const callOrigin = AST.functionCall(AST.memberAccessExpression(AST.identifier("Point"), "origin"), []);
    const v = I.evaluate(callOrigin);
    expect(v.kind).toBe("struct_instance");
  });

  test("private instance method is not accessible; public instance is", () => {
    const I = new InterpreterV10();
    const C = AST.structDefinition("Counter", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")], "named");
    I.evaluate(C);

    const meths = AST.methodsDefinition(
      AST.simpleTypeExpression("Counter"),
      [
        AST.functionDefinition(
          "hidden", // instance method (explicit self supported via bound_method injection regardless of shorthand flag)
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.integerLiteral(1))
          ]),
          undefined,
          undefined,
          undefined,
          false,
          true,
        ),
        AST.functionDefinition(
          "get",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.memberAccessExpression(AST.identifier("self"), "value"))
          ])
        ),
      ]
    );
    I.evaluate(meths);

    const inst = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(5), "value")], false, "Counter");
    const obj = I.evaluate(inst);
    // Private instance method should throw on access
    const callHidden = AST.functionCall(AST.memberAccessExpression(inst, "hidden"), []);
    expect(() => I.evaluate(callHidden)).toThrow();

    // Public instance method should work
    const callGet = AST.functionCall(AST.memberAccessExpression(inst, "get"), []);
    const got = I.evaluate(callGet);
    expect(got).toEqual({ kind: "i32", value: 5n });
  });
});


