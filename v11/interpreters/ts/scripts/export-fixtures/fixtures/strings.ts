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

  {
      name: "strings/string_methods",
      module: AST.module(
        [
          AST.assign("s", AST.stringLiteral("héllo")),
          AST.arr(
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "len_bytes"), []),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "len_chars"), []),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "len_graphemes"), []),
            AST.propagationExpression(
              AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "substring"), [
                AST.integerLiteral(1),
                AST.integerLiteral(3),
              ]),
            ),
            AST.propagationExpression(
              AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "substring"), [
                AST.integerLiteral(2),
                AST.nil(),
              ]),
            ),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "split"), [AST.stringLiteral("l")]),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "split"), [AST.stringLiteral("")]),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("s"), "replace"),
              [AST.stringLiteral("l"), AST.stringLiteral("L")],
            ),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "starts_with"), [AST.stringLiteral("hé")]),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("s"), "ends_with"), [AST.stringLiteral("lo")]),
          ),
        ],
        [AST.importStatement(["able", "text", "string"], true)],
      ),
      manifest: {
        description: "string helpers cover length, substring, split, replace, and prefix/suffix checks",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "u64", value: "6" },
              { kind: "u64", value: "5" },
              { kind: "u64", value: "5" },
              { kind: "string", value: "éll" },
              { kind: "string", value: "llo" },
              {
                kind: "array",
                elements: [
                  { kind: "string", value: "hé" },
                  { kind: "string", value: "" },
                  { kind: "string", value: "o" },
                ],
              },
              {
                kind: "array",
                elements: [
                  { kind: "string", value: "h" },
                  { kind: "string", value: "é" },
                  { kind: "string", value: "l" },
                  { kind: "string", value: "l" },
                  { kind: "string", value: "o" },
                ],
              },
              { kind: "string", value: "héLLo" },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
            ],
          },
        },
      },
    },
];

export default stringsFixtures;
