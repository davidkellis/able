import { AST } from "../../context";
import type { Fixture } from "../../types";

const matchFixtures: Fixture[] = [
  {
      name: "match/identifier_literal",
      module: AST.module([
        AST.matchExpression(
          AST.integerLiteral(2),
          [
            AST.matchClause(
              AST.literalPattern(AST.integerLiteral(1)),
              AST.integerLiteral(10),
            ),
            AST.matchClause(
              AST.identifier("x"),
              AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(5)),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Match falls through literal clause and binds identifier",
        expect: {
          result: { kind: "i32", value: 7n },
        },
      },
    },

  {
      name: "match/guard_clause",
      module: AST.module([
        AST.matchExpression(
          AST.integerLiteral(3),
          [
            AST.matchClause(
              AST.identifier("value"),
              AST.binaryExpression(
                "*",
                AST.identifier("value"),
                AST.integerLiteral(2),
              ),
              AST.binaryExpression(
                ">",
                AST.identifier("value"),
                AST.integerLiteral(2),
              ),
            ),
            AST.matchClause(
              AST.wildcardPattern(),
              AST.integerLiteral(0),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Match guard executes only when predicate passes",
        expect: {
          result: { kind: "i32", value: 6n },
        },
      },
    },

  {
      name: "match/wildcard_pattern",
      module: AST.module([
        AST.matchExpression(
          AST.integerLiteral(0),
          [
            AST.matchClause(
              AST.literalPattern(AST.integerLiteral(1)),
              AST.stringLiteral("One"),
            ),
            AST.matchClause(
              AST.wildcardPattern(),
              AST.stringLiteral("Other"),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Wildcard fallback handles unmatched cases",
        expect: {
          result: { kind: "string", value: "Other" },
        },
      },
    },

  {
      name: "match/struct_guard",
      module: AST.module([
        AST.structDefinition(
          "Point",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
          ],
          "named",
        ),
        AST.matchExpression(
          AST.structLiteral([
            AST.structFieldInitializer(AST.integerLiteral(1), "x"),
            AST.structFieldInitializer(AST.integerLiteral(2), "y"),
          ], false, "Point"),
          [
            AST.matchClause(
              AST.structPattern([
                AST.structPatternField(AST.identifier("a"), "x"),
                AST.structPatternField(AST.identifier("b"), "y"),
              ], false, "Point"),
              AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")),
              AST.binaryExpression(">", AST.identifier("b"), AST.identifier("a")),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Match struct pattern binds fields and guard filters clauses",
        expect: {
          result: { kind: "i32", value: 3n },
        },
      },
    },

  {
      name: "match/struct_positional_pattern",
      module: AST.module([
        AST.structDefinition(
          "Pair",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
          ],
          "positional",
        ),
        AST.matchExpression(
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.integerLiteral(4)),
              AST.structFieldInitializer(AST.integerLiteral(8)),
            ],
            true,
            "Pair",
          ),
          [
            AST.matchClause(
              AST.structPattern(
                [
                  AST.structPatternField(AST.identifier("first")),
                  AST.structPatternField(AST.identifier("second")),
                ],
                true,
                "Pair",
              ),
              AST.binaryExpression(
                "+",
                AST.identifier("first"),
                AST.identifier("second"),
              ),
            ),
            AST.matchClause(
              AST.wildcardPattern(),
              AST.integerLiteral(0),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Positional struct match destructures tuple-style structs",
        expect: {
          result: { kind: "i32", value: 12n },
        },
      },
    },
];

export default matchFixtures;
