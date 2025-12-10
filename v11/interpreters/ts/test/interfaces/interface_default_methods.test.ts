import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - interface default methods", () => {
  test("default method is used when impl omits definition", () => {
    const I = new InterpreterV10();

    const speakable = AST.interfaceDefinition("Speakable", [
      AST.functionSignature(
        "speak",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String"),
        undefined,
        undefined,
        AST.blockExpression([
          AST.returnStatement(AST.stringLiteral("default")),
        ])
      ),
    ]);
    I.evaluate(speakable);

    const botDef = AST.structDefinition(
      "Bot",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "name")],
      "named"
    );
    I.evaluate(botDef);

    const impl = AST.implementationDefinition(
      "Speakable",
      AST.simpleTypeExpression("Bot"),
      []
    );
    I.evaluate(impl);

    const botLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("Beep"), "name"),
    ], false, "Bot");

    const call = AST.functionCall(
      AST.memberAccessExpression(botLiteral, "speak"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "String", value: "default" });
  });

  test("impl overrides default when method provided", () => {
    const I = new InterpreterV10();

    const speakable = AST.interfaceDefinition("Speakable", [
      AST.functionSignature(
        "speak",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String"),
        undefined,
        undefined,
        AST.blockExpression([
          AST.returnStatement(AST.stringLiteral("default")),
        ])
      ),
    ]);
    I.evaluate(speakable);

    const botDef = AST.structDefinition(
      "Bot",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "name")],
      "named"
    );
    I.evaluate(botDef);

    const impl = AST.implementationDefinition(
      "Speakable",
      AST.simpleTypeExpression("Bot"),
      [
        AST.functionDefinition(
          "speak",
          [AST.functionParameter("self", AST.simpleTypeExpression("Bot"))],
          AST.blockExpression([
            AST.returnStatement(AST.stringLiteral("custom")),
          ]),
          AST.simpleTypeExpression("String"),
        ),
      ]
    );
    I.evaluate(impl);

    const botLiteral = AST.structLiteral([
      AST.structFieldInitializer(AST.stringLiteral("Beep"), "name"),
    ], false, "Bot");

    const call = AST.functionCall(
      AST.memberAccessExpression(botLiteral, "speak"),
      []
    );
    expect(I.evaluate(call)).toEqual({ kind: "String", value: "custom" });
  });
});

