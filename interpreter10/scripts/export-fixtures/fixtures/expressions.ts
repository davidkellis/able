import { AST } from "../../context";
import type { Fixture } from "../../types";

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
          result: { kind: "i32", value: -5 },
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
          result: { kind: "i32", value: 1 },
        },
      },
    },

  {
      name: "expressions/integer_suffix",
      module: AST.module([AST.integerLiteral(42, "i64")]),
      manifest: {
        description: "Evaluates an integer literal with an i64 suffix",
        expect: {
          result: { kind: "i32", value: 42 },
        },
      },
    },

  {
      name: "expressions/float_suffix",
      module: AST.module([AST.floatLiteral(3.5, "f32")]),
      manifest: {
        description: "Evaluates a float literal with an f32 suffix",
        expect: {
          result: { kind: "f64", value: 3.5 },
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
          result: { kind: "i32", value: 1 },
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
          result: { kind: "i32", value: 3 },
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
          result: { kind: "i32", value: 7 },
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
          result: { kind: "i32", value: 42 },
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
          result: { kind: "i32", value: 2 },
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
          result: { kind: "string", value: "ok" },
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
          result: { kind: "string", value: "safe" },
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
          result: { kind: "string", value: "body" },
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
          result: { kind: "i32", value: 3 },
        },
      },
    },
];

export default expressionsFixtures;
