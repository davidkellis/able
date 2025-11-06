import { AST } from "../../context";
import type { Fixture } from "../../types";

const privacyFixtures: Fixture[] = [
  {
      name: "privacy/private_static_method",
      module: AST.module([
        AST.structDefinition(
          "Point",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
          ],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Point"),
          [
            AST.functionDefinition(
              "hidden_static",
              [],
              AST.blockExpression([
                AST.returnStatement(
                  AST.structLiteral(
                    [
                      AST.structFieldInitializer(AST.integerLiteral(0), "x"),
                      AST.structFieldInitializer(AST.integerLiteral(0), "y"),
                    ],
                    false,
                    "Point",
                  ),
                ),
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
                  AST.structLiteral(
                    [
                      AST.structFieldInitializer(AST.integerLiteral(0), "x"),
                      AST.structFieldInitializer(AST.integerLiteral(0), "y"),
                    ],
                    false,
                    "Point",
                  ),
                ),
              ]),
            ),
          ],
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("Point"), "hidden_static"),
          [],
        ),
      ]),
      manifest: {
        description: "Calling a private static method raises an error",
        expect: {
          errors: ["Method 'hidden_static' on Point is private"],
        },
      },
    },
];

export default privacyFixtures;
