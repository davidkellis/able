import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - impl generic constraints", () => {
  function registerShowInterfaceAndPoint(interpreter: InterpreterV10) {
    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "to_String",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String")
      ),
    ]);
    interpreter.evaluate(show);

    const pointDef = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named"
    );
    interpreter.evaluate(pointDef);

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
          ]),
          AST.simpleTypeExpression("String")
        ),
      ]
    );
    interpreter.evaluate(pointMethods);
  }

  test("impl method dispatch enforces constraints before call", () => {
    const I = new InterpreterV10();
    registerShowInterfaceAndPoint(I);

    const wrapperDef = AST.structDefinition(
      "Wrapper",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")]
    );
    I.evaluate(wrapperDef);

    const implShowForWrapper = AST.implementationDefinition(
      "Show",
      AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]),
      [
        AST.functionDefinition(
          "to_String",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(
                  AST.memberAccessExpression(AST.identifier("self"), "value"),
                  "to_String"
                ),
                []
              )
            ),
          ]),
          AST.simpleTypeExpression("String")
        ),
      ],
      undefined,
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))])]
    );
    I.evaluate(implShowForWrapper);

    const pointLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.integerLiteral(1), "x"),
      AST.structFieldInitializer(AST.integerLiteral(2), "y"),
    ], false, "Point");

    const wrapperPointLiteral = AST.structLiteral(
      [AST.structFieldInitializer(pointLiteral, "value")],
      false,
      "Wrapper",
      undefined,
      [AST.simpleTypeExpression("Point")]
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("good"), wrapperPointLiteral));

    const callGood = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("good"), "to_String"),
      []
    );
    expect(I.evaluate(callGood)).toEqual({ kind: "String", value: "Point(1, 2)" });

    const wrapperI32Literal = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(7), "value")],
      false,
      "Wrapper",
      undefined,
      [AST.simpleTypeExpression("i32")]
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("bad"), wrapperI32Literal));

    const callBad = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("bad"), "to_String"),
      []
    );
    expect(() => I.evaluate(callBad)).toThrow(/does not satisfy interface 'Show': missing method 'to_String'/);
  });
});

