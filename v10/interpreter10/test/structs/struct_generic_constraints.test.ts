import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - struct generic constraints", () => {
  test("struct literal enforces type parameter interface constraints", () => {
    const I = new InterpreterV10();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const pointDef = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named"
    );
    I.evaluate(pointDef);

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
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(pointMethods);

    const boxDef = AST.structDefinition(
      "Box",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])]
    );
    I.evaluate(boxDef);

    const pointLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point");

    const okLiteral = AST.structLiteral(
      [AST.structFieldInitializer(pointLiteral, "value")],
      false,
      "Box",
      undefined,
      [AST.simpleTypeExpression("Point")]
    );
    const okVal = I.evaluate(okLiteral);
    expect(okVal.kind).toBe("struct_instance");

    const badLiteral = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(9), "value")],
      false,
      "Box",
      undefined,
      [AST.simpleTypeExpression("i32")]
    );
    expect(() => I.evaluate(badLiteral)).toThrow(/does not satisfy interface 'Show': missing method 'to_string'/);
  });

  test("struct literal rejects missing type arguments", () => {
    const I = new InterpreterV10();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const boxDef = AST.structDefinition(
      "WithConstraint",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])]
    );
    I.evaluate(boxDef);

    const literal = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(1), "value")],
      false,
      "WithConstraint"
    );
    expect(() => I.evaluate(literal)).toThrow(/Type arguments count mismatch/);
  });
});

