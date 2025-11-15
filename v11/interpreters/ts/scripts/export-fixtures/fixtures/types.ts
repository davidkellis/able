import { AST } from "../../context";
import type { Fixture } from "../../types";

const typesFixtures: Fixture[] = [
  {
      name: "types/generic_type_expression",
      module: AST.module([
        AST.structDefinition(
          "Box",
          [
            AST.structFieldDefinition(
              AST.genericTypeExpression(
                AST.simpleTypeExpression("Array"),
                [AST.simpleTypeExpression("i32")],
              ),
              "values",
            ),
          ],
          "named",
        ),
        AST.assign(
          "box",
          AST.structLiteral(
            [
              AST.structFieldInitializer(
                AST.arrayLiteral([
                  AST.integerLiteral(1),
                  AST.integerLiteral(2),
                  AST.integerLiteral(3),
                ]),
                "values",
              ),
            ],
            false,
            "Box",
          ),
        ),
        AST.memberAccessExpression(AST.identifier("box"), "values"),
      ]),
      manifest: {
        description: "Struct field uses generic type annotation",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 1 },
              { kind: "i32", value: 2 },
              { kind: "i32", value: 3 },
            ],
          },
        },
      },
    },

  {
      name: "types/function_type_expression",
      module: AST.module([
        AST.fn(
          "apply",
          [
            AST.param("value", AST.simpleTypeExpression("i32")),
            AST.param(
              "cb",
              AST.functionTypeExpression(
                [AST.simpleTypeExpression("i32")],
                AST.simpleTypeExpression("i32"),
              ),
            ),
          ],
          [AST.call("cb", AST.identifier("value"))],
          AST.simpleTypeExpression("i32"),
        ),
        AST.fn(
          "double",
          [AST.param("n", AST.simpleTypeExpression("i32"))],
          [
            AST.binaryExpression(
              "*",
              AST.identifier("n"),
              AST.integerLiteral(2),
            ),
          ],
          AST.simpleTypeExpression("i32"),
        ),
        AST.call("apply", AST.integerLiteral(3), AST.identifier("double")),
      ]),
      manifest: {
        description: "Function parameter uses arrow type annotation",
        expect: {
          result: { kind: "i32", value: 6 },
        },
      },
    },

  {
      name: "types/nullable_type_expression",
      module: AST.module([
        AST.fn(
          "maybe_identity",
          [
            AST.param(
              "value",
              AST.nullableTypeExpression(AST.simpleTypeExpression("string")),
            ),
          ],
          [AST.identifier("value")],
          AST.nullableTypeExpression(AST.simpleTypeExpression("string")),
        ),
        AST.call("maybe_identity", AST.stringLiteral("ready")),
      ]),
      manifest: {
        description: "Function parameter and return use nullable type",
        expect: {
          result: { kind: "string", value: "ready" },
        },
      },
    },

  {
      name: "types/result_type_expression",
      module: AST.module([
        AST.fn(
          "always_ok",
          [],
          [AST.integerLiteral(7)],
          AST.resultTypeExpression(AST.simpleTypeExpression("i32")),
        ),
        AST.call("always_ok"),
      ]),
      manifest: {
        description: "Function returns a result-wrapped type",
        expect: {
          result: { kind: "i32", value: 7 },
        },
      },
    },

  {
      name: "types/union_type_expression",
      module: AST.module([
        AST.fn(
          "identity_union",
          [
            AST.param(
              "value",
              AST.unionTypeExpression([
                AST.simpleTypeExpression("string"),
                AST.simpleTypeExpression("i32"),
              ]),
            ),
          ],
          [AST.identifier("value")],
          AST.unionTypeExpression([
            AST.simpleTypeExpression("string"),
            AST.simpleTypeExpression("i32"),
          ]),
        ),
        AST.call("identity_union", AST.stringLiteral("hello")),
      ]),
      manifest: {
        description: "Function uses union type in parameter and return",
        expect: {
          result: { kind: "string", value: "hello" },
        },
      },
    },

  {
    name: "types/generic_where_constraint",
    module: AST.module([
      AST.fn(
        "choose_first",
        [
          AST.param("first", AST.simpleTypeExpression("T")),
          AST.param("second", AST.simpleTypeExpression("U")),
        ],
        [AST.identifier("first")],
        AST.simpleTypeExpression("T"),
        [
          AST.genericParameter("T"),
          AST.genericParameter("U"),
        ],
        [
          AST.whereClauseConstraint("T", [
            AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
            AST.interfaceConstraint(AST.simpleTypeExpression("Clone")),
          ]),
          AST.whereClauseConstraint("U", [
            AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
          ]),
        ],
      ),
      AST.callT(
        "choose_first",
        [
          AST.simpleTypeExpression("string"),
          AST.simpleTypeExpression("i32"),
        ],
        AST.stringLiteral("winner"),
        AST.integerLiteral(1),
      ),
    ]),
    manifest: {
      description: "Function where clause constrains generic parameters",
      expect: {
        result: { kind: "string", value: "winner" },
      },
    },
  },

  {
    name: "types/type_alias_definition",
    module: AST.module([
      AST.typeAliasDefinition("UserID", AST.simpleTypeExpression("u64")),
      AST.typeAliasDefinition(
        "Box",
        AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
        [AST.genericParameter("T")],
        [
          AST.whereClauseConstraint("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Display"))]),
        ],
      ),
      AST.integerLiteral(10),
    ]),
    manifest: {
      description: "Module can declare plain and generic type aliases",
      expect: {
        result: { kind: "i32", value: 10 },
      },
    },
  },
];

export default typesFixtures;
