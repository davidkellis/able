import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - method-level generic constraints (minimal runtime checks)", () => {
  test("instance method with constrained type parameter is enforced", () => {
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
        // fn accept_show<T where T: Show>(self: Point, x: T) -> string { x.to_string() }
        AST.functionDefinition(
          "accept_show",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Point")),
            AST.functionParameter("x"),
          ],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(AST.memberAccessExpression(AST.identifier("x"), "to_string"), [])
            ),
          ]),
          AST.simpleTypeExpression("string"),
          [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])]
        ),
      ]
    );
    I.evaluate(pointMethods);

    const pVal = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point");

    // OK: x is Point which implements Show
    const ok = AST.functionCall(
      AST.memberAccessExpression(pVal, "accept_show"),
      [pVal],
      [AST.simpleTypeExpression("Point")]
    );
    const okVal = I.evaluate(ok);
    expect(okVal).toEqual({ kind: "string", value: "Point(1, 2)" });

    // Not OK: x is i32 which does not implement Show
    const bad = AST.functionCall(
      AST.memberAccessExpression(pVal, "accept_show"),
      [AST.integerLiteral(3)],
      [AST.simpleTypeExpression("i32")]
    );
    expect(() => I.evaluate(bad)).toThrow(/does not satisfy interface 'Show': missing method 'to_string'/);
  });
});


