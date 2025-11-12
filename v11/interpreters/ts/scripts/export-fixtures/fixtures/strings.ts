import { AST } from "../../context";
import type { Fixture } from "../../types";

const stringsFixtures: Fixture[] = [
  {
      name: "strings/interpolation_basic",
      module: AST.module([
        AST.assign("x", AST.int(2)),
        AST.stringInterpolation([
          AST.stringLiteral("x = "),
          AST.identifier("x"),
          AST.stringLiteral(", sum = "),
          AST.binaryExpression("+", AST.integerLiteral(3), AST.integerLiteral(4)),
        ]),
      ]),
      manifest: {
        description: "Interpolates literals and expressions",
        expect: {
          result: { kind: "string", value: "x = 2, sum = 7" },
        },
      },
    },

  {
      name: "strings/interpolation_struct_to_string",
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
              "to_string",
              [AST.functionParameter("self")],
              AST.blockExpression([
                AST.returnStatement(
                  AST.stringInterpolation([
                    AST.stringLiteral("Point("),
                    AST.memberAccessExpression(AST.identifier("self"), "x"),
                    AST.stringLiteral(","),
                    AST.memberAccessExpression(AST.identifier("self"), "y"),
                    AST.stringLiteral(")"),
                  ]),
                ),
              ]),
            ),
          ],
        ),
        AST.assign(
          "p",
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.integerLiteral(1), "x"),
              AST.structFieldInitializer(AST.integerLiteral(2), "y"),
            ],
            false,
            "Point",
          ),
        ),
        AST.stringInterpolation([
          AST.stringLiteral("P= "),
          AST.identifier("p"),
        ]),
      ]),
      manifest: {
        description: "Uses to_string method when interpolating struct instances",
        expect: {
          result: { kind: "string", value: "P= Point(1,2)" },
        },
      },
    },
];

export default stringsFixtures;
