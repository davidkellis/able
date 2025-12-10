import { AST } from "../../context";
import type { Fixture } from "../../types";

const pipeChain = (initial: AST.Expression, steps: AST.Expression[]): AST.Expression => {
  return steps.reduce((subject, step) => AST.binaryExpression("|>", subject, step), initial);
};

const expressionsFixtures: Fixture[] = [
  {
      name: "expressions/array_literal_empty",
      module: AST.module([AST.arr()]),
      manifest: {
        description: "Evaluates an empty array literal",
        expect: {
          result: { kind: "array", elements: [] },
        },
      },
    },

  {
      name: "expressions/unary_negation",
      module: AST.module([AST.un("-", AST.int(5))]),
      manifest: {
        description: "Negates an integer literal",
        expect: {
          result: { kind: "i32", value: -5n },
        },
      },
    },

  {
      name: "expressions/block_expression",
      module: AST.module([
        AST.block(
          AST.assign("x", AST.int(1)),
          AST.identifier("x"),
        ),
      ]),
      manifest: {
        description: "Evaluates a block expression and returns its final expression",
        expect: {
          result: { kind: "i32", value: 1n },
        },
      },
    },

  {
      name: "expressions/integer_suffix",
      module: AST.module([AST.integerLiteral(42, "i64")]),
      manifest: {
        description: "Evaluates an integer literal with an i64 suffix",
        expect: {
          result: { kind: "i64", value: 42n },
        },
      },
    },

  {
      name: "expressions/float_suffix",
      module: AST.module([AST.floatLiteral(3.5, "f32")]),
      manifest: {
        description: "Evaluates a float literal with an f32 suffix",
        expect: {
          result: { kind: "f32", value: 3.5 },
        },
      },
    },

  {
      name: "expressions/assignment_declare",
      module: AST.module([
        AST.assign("value", AST.int(1)),
        AST.id("value"),
      ]),
      manifest: {
        description: "Declaration assignment binds a new identifier",
        expect: {
          result: { kind: "i32", value: 1n },
        },
      },
    },

  {
      name: "expressions/reassignment",
      module: AST.module([
        AST.assign("value", AST.int(1)),
        AST.assign("value", AST.bin("+", AST.id("value"), AST.int(2)), "="),
        AST.id("value"),
      ]),
      manifest: {
        description: "Reassignment updates an existing binding",
        expect: {
          result: { kind: "i32", value: 3n },
        },
      },
    },

  {
      name: "expressions/compound_assignment",
      module: AST.module([
        AST.assign("value", AST.int(2)),
        AST.assign("value", AST.int(5), "+="),
        AST.id("value"),
      ]),
      manifest: {
        description: "Compound assignment applies operator to existing value",
        expect: {
          result: { kind: "i32", value: 7n },
        },
      },
    },

  {
      name: "expressions/breakpoint_value",
      module: AST.module([
        AST.assign(
          "result",
          AST.breakpointExpression(
            "exit",
            AST.block(
              AST.breakStatement("exit", AST.int(42)),
            ),
          ),
        ),
        AST.id("result"),
      ]),
      manifest: {
        description: "Breakpoint expression returns value from labeled break",
        expect: {
          result: { kind: "i32", value: 42n },
        },
      },
    },

  {
      name: "expressions/index_access",
      module: AST.module([
        AST.assign("arr", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
        AST.index(AST.id("arr"), AST.int(1)),
      ]),
      manifest: {
        description: "Index expression reads array element by position",
        expect: {
          result: { kind: "i32", value: 2n },
        },
      },
    },

  {
      name: "expressions/map_literal_spread",
      module: AST.module([
        AST.assign(
          "defaults",
          AST.mapLit([
            AST.mapEntry(AST.stringLiteral("accept"), AST.stringLiteral("application/json")),
          ]),
        ),
        AST.assign(
          "headers",
          AST.mapLit([
            AST.mapEntry(AST.stringLiteral("content-type"), AST.stringLiteral("application/json")),
            AST.mapSpread(AST.identifier("defaults")),
            AST.mapEntry(AST.stringLiteral("authorization"), AST.stringLiteral("Bearer abc")),
          ]),
        ),
        AST.identifier("headers"),
      ]),
      manifest: {
        description: "Map literal supports spreads and overrides",
        expect: {
          result: {
            kind: "hash_map",
            entries: [
              { key: { kind: "String", value: "content-type" }, value: { kind: "String", value: "application/json" } },
              { key: { kind: "String", value: "accept" }, value: { kind: "String", value: "application/json" } },
              { key: { kind: "String", value: "authorization" }, value: { kind: "String", value: "Bearer abc" } },
            ],
          },
        },
      },
    },

  {
      name: "expressions/or_else_success",
      module: AST.module([
        AST.assign("value", AST.stringLiteral("ok")),
        AST.assign(
          "result",
          AST.orElseExpression(
            AST.propagationExpression(AST.identifier("value")),
            AST.blockExpression([AST.stringLiteral("fallback")]),
          ),
        ),
        AST.identifier("result"),
      ]),
      manifest: {
        description: "Propagation succeeds without invoking handler",
        expect: {
          result: { kind: "String", value: "ok" },
        },
      },
    },

  {
      name: "expressions/rescue_success",
      module: AST.module([
        AST.rescueExpression(
          AST.blockExpression([AST.stringLiteral("safe")]),
          [
            AST.matchClause(
              AST.wildcardPattern(),
              AST.stringLiteral("handled"),
            ),
          ],
        ),
      ]),
      manifest: {
        description: "Rescue expression returns original value when no error is raised",
        expect: {
          result: { kind: "String", value: "safe" },
        },
      },
    },

  {
      name: "expressions/ensure_success",
      module: AST.module([
        AST.ensureExpression(
          AST.blockExpression([AST.stringLiteral("body")]),
          AST.blockExpression([
            AST.functionCall(AST.identifier("print"), [AST.stringLiteral("ensure")]),
          ]),
        ),
      ]),
      manifest: {
        description: "Ensure block runs even when try expression succeeds",
        expect: {
          stdout: ["ensure"],
          result: { kind: "String", value: "body" },
        },
      },
    },

  {
      name: "expressions/int_addition",
      module: AST.module([
        AST.assign("a", AST.int(1)),
        AST.assign("b", AST.int(2)),
        AST.bin("+", AST.id("a"), AST.id("b")),
      ]),
      manifest: {
        description: "Adds two integers",
        expect: {
          result: { kind: "i32", value: 3n },
        },
      },
    },

  {
      name: "pipes/topic_placeholder",
      module: AST.module([
        AST.fn(
          "add",
          [AST.param("x", AST.simpleTypeExpression("i32")), AST.param("y", AST.simpleTypeExpression("i32"))],
          [AST.ret(AST.bin("+", AST.id("x"), AST.id("y")))],
          AST.simpleTypeExpression("i32"),
        ),
        AST.fn(
          "double",
          [AST.param("x", AST.simpleTypeExpression("i32"))],
          [AST.ret(AST.bin("*", AST.id("x"), AST.int(2)))],
          AST.simpleTypeExpression("i32"),
        ),
        AST.structDefinition(
          "Box",
          [AST.fieldDef(AST.simpleTypeExpression("i32"), "value")],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Box"),
          [
            AST.fn(
              "double",
              [],
              [AST.ret(AST.bin("*", AST.implicitMemberExpression("value"), AST.int(2)))],
              AST.simpleTypeExpression("i32"),
              undefined,
              undefined,
              true,
            ),
          ],
        ),
        AST.assign("value", AST.int(5)),
        AST.assign(
          "arithmetic",
          AST.binaryExpression("|>", AST.id("value"), AST.bin("+", AST.placeholderExpression(), AST.int(3))),
        ),
        AST.assign(
          "pipePlaceholder",
          AST.binaryExpression("|>", AST.id("value"), AST.call(AST.id("add"), AST.placeholderExpression(), AST.int(4))),
        ),
        AST.assign(
          "lambdaResult",
          AST.call(AST.call(AST.id("add"), AST.placeholderExpression(), AST.int(4)), AST.id("value")),
        ),
        AST.assign("ufcs", AST.binaryExpression("|>", AST.id("value"), AST.id("double"))),
        AST.assign(
          "boxed",
          AST.binaryExpression(
            "|>",
            AST.structLiteral([AST.fieldInit(AST.id("value"), "value")], false, "Box"),
            AST.implicitMemberExpression("double"),
          ),
        ),
        AST.assign(
          "total",
          AST.bin(
            "+",
            AST.bin(
              "+",
              AST.bin("+", AST.bin("+", AST.id("arithmetic"), AST.id("pipePlaceholder")), AST.id("lambdaResult")),
              AST.id("ufcs"),
            ),
            AST.id("boxed"),
          ),
        ),
        AST.id("total"),
      ]),
      manifest: {
        description: "Pipelines cover placeholder (@) callables, UFCS, and method steps",
        expect: {
          result: { kind: "i32", value: 46n },
        },
      },
    },

  {
      name: "pipes/member_topic",
      module: AST.module([
        AST.structDefinition(
          "Box",
          [AST.fieldDef(AST.simpleTypeExpression("i32"), "value")],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Box"),
          [
            AST.fn(
              "increment",
              [],
              [AST.ret(AST.bin("+", AST.implicitMemberExpression("value"), AST.int(1)))],
              AST.simpleTypeExpression("i32"),
              undefined,
              undefined,
              true,
            ),
            AST.fn(
              "double",
              [],
              [AST.ret(AST.bin("*", AST.implicitMemberExpression("value"), AST.int(2)))],
              AST.simpleTypeExpression("i32"),
              undefined,
              undefined,
              true,
            ),
          ],
        ),
        AST.assign(
          "box",
          AST.structLiteral([AST.fieldInit(AST.int(5), "value")], false, "Box"),
        ),
        AST.assign("first", AST.binaryExpression("|>", AST.id("box"), AST.implicitMemberExpression("increment"))),
        AST.assign("second", AST.binaryExpression("|>", AST.id("box"), AST.implicitMemberExpression("double"))),
        AST.stringInterpolation([AST.id("first"), AST.stringLiteral(","), AST.id("second")]),
      ]),
      manifest: {
        description: "Pipe uses pipe/apply to call implicit methods without topic placeholder",
        expect: {
          result: { kind: "String", value: "6,10" },
        },
      },
    },

  {
      name: "pipes/multi_stage_chain",
      module: AST.module([
        AST.fn(
          "add",
          [
            AST.param("left", AST.simpleTypeExpression("i32")),
            AST.param("right", AST.simpleTypeExpression("i32")),
          ],
          [AST.ret(AST.bin("+", AST.id("left"), AST.id("right")))],
          AST.simpleTypeExpression("i32"),
        ),
        AST.fn(
          "toBox",
          [AST.param("value", AST.simpleTypeExpression("i32"))],
          [
            AST.ret(
              AST.structLiteral(
                [AST.fieldInit(AST.id("value"), "value")],
                false,
                "Box",
              ),
            ),
          ],
          AST.simpleTypeExpression("Box"),
        ),
        AST.structDefinition(
          "Box",
          [AST.fieldDef(AST.simpleTypeExpression("i32"), "value")],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Box"),
          [
            AST.fn(
              "augment",
              [
                AST.param("self", AST.simpleTypeExpression("Box")),
                AST.param("delta", AST.simpleTypeExpression("i32")),
              ],
              [
                AST.ret(
                  AST.structLiteral(
                    [
                      AST.fieldInit(
                        AST.bin(
                          "+",
                          AST.member(AST.id("self"), "value"),
                          AST.id("delta"),
                        ),
                        "value",
                      ),
                    ],
                    false,
                    "Box",
                  ),
                ),
              ],
              AST.simpleTypeExpression("Self"),
            ),
            AST.fn(
              "double",
              [AST.param("self", AST.simpleTypeExpression("Box"))],
              [AST.ret(AST.bin("*", AST.member(AST.id("self"), "value"), AST.int(2)))],
              AST.simpleTypeExpression("i32"),
            ),
          ],
        ),
        AST.fn(
          "augmentBy4",
          [AST.param("box", AST.simpleTypeExpression("Box"))],
          [
            AST.ret(
              AST.call(AST.memberAccessExpression(AST.id("box"), "augment"), AST.int(4)),
            ),
          ],
          AST.simpleTypeExpression("Box"),
        ),
        AST.fn(
          "doubleBox",
          [AST.param("box", AST.simpleTypeExpression("Box"))],
          [AST.ret(AST.call(AST.memberAccessExpression(AST.id("box"), "double")))],
          AST.simpleTypeExpression("i32"),
        ),
        AST.assign("start", AST.int(5)),
        AST.assign(
          "result",
          pipeChain(AST.id("start"), [
            AST.bin("*", AST.placeholderExpression(), AST.int(2)),
            AST.call(AST.id("add"), AST.placeholderExpression(), AST.int(3)),
            AST.id("toBox"),
            AST.id("augmentBy4"),
            AST.id("doubleBox"),
          ]),
        ),
        AST.id("result"),
      ]),
      manifest: {
        description: "Multi-stage pipeline mixing placeholder-built callables and bound methods",
        expect: {
          result: { kind: "i32", value: 34n },
        },
      },
    },

  {
      name: "pipes/low_precedence_pipe",
      module: AST.module([
        AST.fn(
          "inc",
          [AST.param("x", AST.simpleTypeExpression("i32"))],
          [AST.ret(AST.bin("+", AST.id("x"), AST.int(1)))],
          AST.simpleTypeExpression("i32"),
        ),
        AST.fn(
          "bool_to_int",
          [AST.param("value", AST.simpleTypeExpression("bool"))],
          [
            AST.ret(
              AST.ifExpression(
                AST.id("value"),
                AST.block(AST.ret(AST.int(1))),
                [],
                AST.block(AST.ret(AST.int(0))),
              ),
            ),
          ],
          AST.simpleTypeExpression("i32"),
        ),
        AST.assign("a", AST.int(0)),
        AST.assign("b", AST.int(0)),
        AST.assign("res1", AST.binaryExpression("|>>", AST.assign("a", AST.int(5), "="), AST.id("inc"))),
        AST.assign("res2", AST.assign("b", AST.binaryExpression("|>", AST.int(1), AST.id("inc")), "=")),
        AST.assign("flag", AST.binaryExpression("|>", AST.bin("||", AST.bool(false), AST.bool(true)), AST.id("bool_to_int"))),
        AST.arr(AST.id("res1"), AST.id("a"), AST.id("res2"), AST.id("b"), AST.id("flag")),
      ]),
      manifest: {
        description: "Low-precedence pipe (|>>) runs after assignment; |> binds above assignment but below ||",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 6n },
              { kind: "i32", value: 5n },
              { kind: "i32", value: 2n },
              { kind: "i32", value: 2n },
              { kind: "i32", value: 1n },
            ],
          },
        },
      },
    },
];

export default expressionsFixtures;
