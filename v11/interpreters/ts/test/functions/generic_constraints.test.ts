import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - generic where-constraints (minimal runtime checks)", () => {
  test("built-in Display/Clone interfaces are available without declarations", () => {
    const I = new InterpreterV10();
    const chooseFirst = AST.functionDefinition(
      "choose_first",
      [
        AST.functionParameter("first", AST.simpleTypeExpression("T")),
        AST.functionParameter("second", AST.simpleTypeExpression("U")),
      ],
      AST.blockExpression([AST.returnStatement(AST.identifier("first"))]),
      AST.simpleTypeExpression("T"),
      [AST.genericParameter("T"), AST.genericParameter("U")],
      [
        AST.whereClauseConstraint("T", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
          AST.interfaceConstraint(AST.simpleTypeExpression("Clone")),
        ]),
        AST.whereClauseConstraint("U", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
        ]),
      ],
    );
    I.evaluate(chooseFirst);
    const call = AST.functionCall(
      AST.identifier("choose_first"),
      [AST.stringLiteral("winner"), AST.integerLiteral(1)],
      [AST.simpleTypeExpression("String"), AST.simpleTypeExpression("i32")],
    );
    const value = I.evaluate(call);
    expect(value).toEqual({ kind: "String", value: "winner" });
  });

  test("fn constrained by Interface is enforced at call site", () => {
    const I = new InterpreterV10();

    // interface Show { fn to_string(self: Self) -> string }
    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature("to_String", [AST.functionParameter("self", AST.simpleTypeExpression("Self"))], AST.simpleTypeExpression("String")),
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
          "to_String",
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
        AST.returnStatement(AST.functionCall(AST.memberAccessExpression(AST.identifier("x"), "to_String"), [])),
      ]),
      AST.simpleTypeExpression("String"),
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])],
    );
    I.evaluate(showVal);

    const p = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point");

    const ok = AST.functionCall(AST.identifier("show_val"), [p], [AST.simpleTypeExpression("Point")]);
    const okVal = I.evaluate(ok);
    expect(okVal).toEqual({ kind: "String", value: "Point(1, 2)" });

    // call with unconstrained type (i32) should fail before body executes
    const badCall = AST.functionCall(AST.identifier("show_val"), [AST.integerLiteral(3)], [AST.simpleTypeExpression("i32")]);
    expect(() => I.evaluate(badCall)).toThrow(/does not satisfy interface 'Show': missing method 'to_String'/);
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
    const inferred = I.evaluate(tooFew);
    expect(inferred).toEqual({ kind: "i32", value: 1n });
    const tooMany = AST.functionCall(AST.identifier("id"), [AST.integerLiteral(1)], [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")]);
    expect(() => I.evaluate(tooMany)).toThrow(/Type arguments count mismatch/);
  });
});
