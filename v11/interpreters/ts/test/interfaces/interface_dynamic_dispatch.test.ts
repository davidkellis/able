import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - interface dynamic dispatch", () => {
  test("value stored as interface dispatches to underlying impl", () => {
    const I = new InterpreterV10();

    const pointDef = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named"
    );
    I.evaluate(pointDef);

    const pointImpl = AST.implementationDefinition(
      "Display",
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
    I.evaluate(pointImpl);

    const assign = AST.assignmentExpression(
      ":=",
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("Display")),
      AST.structLiteral([
        AST.structFieldInitializer(AST.integerLiteral(2), "x"),
        AST.structFieldInitializer(AST.integerLiteral(3), "y"),
      ], false, "Point")
    );
    I.evaluate(assign);

    const call = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("value"), "to_string"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "string", value: "Point(2, 3)" });
  });

  test("array of interface values uses most specific impl", () => {
    const I = new InterpreterV10();

    const show = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "describe",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("string")
      ),
    ]);
    I.evaluate(show);

    const fancyDef = AST.structDefinition(
      "Fancy",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(fancyDef);

    const basicDef = AST.structDefinition(
      "Basic",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "label")],
      "named"
    );
    I.evaluate(basicDef);

    const fancyImpl = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("Fancy"),
      [
        AST.functionDefinition(
          "describe",
          [AST.functionParameter("self", AST.simpleTypeExpression("Fancy"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("fancy")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(fancyImpl);

    const basicImpl = AST.implementationDefinition(
      "Show",
      AST.simpleTypeExpression("Basic"),
      [
        AST.functionDefinition(
          "describe",
          [AST.functionParameter("self", AST.simpleTypeExpression("Basic"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("basic")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(basicImpl);

    const unionImpl = AST.implementationDefinition(
      "Show",
      AST.unionTypeExpression([
        AST.simpleTypeExpression("Fancy"),
        AST.simpleTypeExpression("Basic"),
      ]),
      [
        AST.functionDefinition(
          "describe",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("union")),
          ]),
          AST.simpleTypeExpression("string")
        ),
      ]
    );
    I.evaluate(unionImpl);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("items"),
        AST.arrayLiteral([
          AST.structLiteral([
            AST.structFieldInitializer(AST.stringLiteral("f"), "label"),
          ], false, "Fancy"),
          AST.structLiteral([
            AST.structFieldInitializer(AST.stringLiteral("b"), "label"),
          ], false, "Basic"),
        ])
      )
    );

    I.evaluate(AST.assignmentExpression(":=", AST.identifier("buffer"), AST.stringLiteral("")));
    I.evaluate(
      AST.forLoop(
        AST.typedPattern(AST.identifier("item"), AST.simpleTypeExpression("Show")),
        AST.identifier("items"),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("buffer"),
            AST.binaryExpression(
              "+",
              AST.identifier("buffer"),
              AST.functionCall(AST.memberAccessExpression(AST.identifier("item"), "describe"), [])
            ),
          ),
        ])
      )
    );
    expect(I.evaluate(AST.identifier("buffer"))).toEqual({ kind: "string", value: "fancybasic" });
  });
});
