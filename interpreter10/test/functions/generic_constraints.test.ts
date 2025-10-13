import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - generic where-constraints (minimal runtime checks)", () => {
  test("fn constrained by Interface is enforced at call site", () => {
    const I = new InterpreterV10();

    // interface Show { fn to_string(self: Self) -> string }
    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature("to_string", [AST.functionParameter("self", AST.simpleTypeExpression("Self"))], AST.simpleTypeExpression("string")),
    ]);
    I.evaluate(show);

    // struct Point { x: i32, y: i32 } with inherent method to_string(self) -> string
    const Point = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named"
    );
    I.evaluate(Point);
    const pointMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "to_string",
          [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.stringInterpolation([
                AST.stringLiteral("Point("),
                AST.memberAccessExpression(AST.identifier("self"), "x"),
                AST.stringLiteral(", "),
                AST.memberAccessExpression(AST.identifier("self"), "y"),
                AST.stringLiteral(")"),
              ])
            ),
          ])
        ),
      ]
    );
    I.evaluate(pointMethods);

    // fn show_val<T where T: Show>(x: T) -> string { x.to_string() }
    const showVal = AST.functionDefinition(
      "show_val",
      [AST.functionParameter("x")],
      AST.blockExpression([
        AST.returnStatement(AST.functionCall(AST.memberAccessExpression(AST.identifier("x"), "to_string"), [])),
      ]),
      AST.simpleTypeExpression("string"),
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])],
    );
    I.evaluate(showVal);

    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point");

    const ok = AST.functionCall(AST.identifier("show_val"), [p], [AST.simpleTypeExpression("Point")]);
    const okVal = I.evaluate(ok);
    expect(okVal).toEqual({ kind: "string", value: "Point(1, 2)" });

    // call with unconstrained type (i32) should fail before body executes
    const badCall = AST.functionCall(AST.identifier("show_val"), [AST.integerLiteral(3)], [AST.simpleTypeExpression("i32")]);
    expect(() => I.evaluate(badCall)).toThrow(/does not satisfy interface 'Show': missing method 'to_string'/);
  });

  test("mismatched type argument count is rejected", () => {
    const I = new InterpreterV10();
    const id = AST.functionDefinition(
      "id",
      [AST.functionParameter("x")],
      AST.blockExpression([AST.returnStatement(AST.identifier("x"))]),
      undefined,
      [AST.genericParameter("T")]
    );
    I.evaluate(id);
    const tooFew = AST.functionCall(AST.identifier("id"), [AST.integerLiteral(1)], []);
    expect(() => I.evaluate(tooFew)).toThrow(/Type arguments count mismatch/);
    const tooMany = AST.functionCall(AST.identifier("id"), [AST.integerLiteral(1)], [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")]);
    expect(() => I.evaluate(tooMany)).toThrow(/Type arguments count mismatch/);
  });
});


