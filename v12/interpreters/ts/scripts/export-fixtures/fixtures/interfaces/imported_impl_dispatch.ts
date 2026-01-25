import { AST } from "../../../context";
import type { Fixture } from "../../../types";

export const importedImplDispatchFixture: Fixture = {
  name: "interfaces/imported_impl_dispatch",
  setupModules: {
    "package.json": AST.module(
      [
        AST.structDefinition(
          "Talker",
          [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "name")],
          "named",
        ),
        AST.interfaceDefinition(
          "Speaker",
          [
            AST.functionSignature(
              "speak",
              [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
              AST.simpleTypeExpression("String"),
            ),
          ],
        ),
        AST.implementationDefinition(
          "Speaker",
          AST.simpleTypeExpression("Talker"),
          [
            AST.functionDefinition(
              "speak",
              [AST.functionParameter("self", AST.simpleTypeExpression("Talker"))],
              AST.blockExpression([AST.memberAccessExpression(AST.id("self"), "name")]),
              AST.simpleTypeExpression("String"),
            ),
          ],
        ),
      ],
      [],
      AST.packageStatement(["speaklib"]),
    ),
  },
  module: AST.module(
    [
      AST.assign(
        "talker",
        AST.structLiteral(
          [AST.structFieldInitializer(AST.stringLiteral("hi"), "name")],
          false,
          "Talker",
        ),
      ),
      AST.functionCall(AST.memberAccessExpression(AST.id("talker"), "speak"), []),
    ],
    [AST.importStatement(["speaklib"], true)],
  ),
  manifest: {
    description: "Wildcard imports bring interface impls into scope for method dispatch",
    setup: ["package.json"],
    expect: {
      result: { kind: "String", value: "hi" },
    },
  },
};
