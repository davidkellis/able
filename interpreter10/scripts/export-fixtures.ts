import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST } from "../index";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../fixtures/ast");

interface Fixture {
  name: string; // folder relative to FIXTURE_ROOT
  module: AST.Module;
  setupModules?: Record<string, AST.Module>;
  manifest?: FixtureManifest;
}

interface FixtureManifest {
  description: string;
  entry?: string;
  expect?: Record<string, unknown>;
  setup?: string[];
}

const fixtures: Fixture[] = [
  {
    name: "basics/string_literal",
    module: AST.module([AST.str("hello")]),
    manifest: {
      description: "Evaluates a simple string literal module",
      expect: {
        result: { kind: "string", value: "hello" },
      },
    },
  },
  {
    name: "basics/bool_literal",
    module: AST.module([AST.bool(true)]),
    manifest: {
      description: "Evaluates a boolean literal",
      expect: {
        result: { kind: "bool", value: true },
      },
    },
  },
  {
    name: "basics/nil_literal",
    module: AST.module([AST.nil()]),
    manifest: {
      description: "Evaluates the nil literal",
      expect: {
        result: { kind: "nil" },
      },
    },
  },
  {
    name: "basics/char_literal",
    module: AST.module([AST.chr("a")]),
    manifest: {
      description: "Evaluates a character literal",
      expect: {
        result: { kind: "char", value: "a" },
      },
    },
  },
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
    name: "functions/lambda_expression",
    module: AST.module([
      AST.assign(
        "adder",
        AST.lambdaExpression(
          [AST.param("x"), AST.param("y")],
          AST.bin("+", AST.id("x"), AST.id("y")),
        ),
      ),
      AST.functionCall(AST.id("adder"), [AST.int(2), AST.int(3)]),
    ]),
    manifest: {
      description: "Lambda expression returns computed sum",
      expect: {
        result: { kind: "i32", value: 5 },
      },
    },
  },
  {
    name: "functions/trailing_lambda_call",
    module: AST.module([
      AST.functionDefinition(
        "for_each",
        [AST.param("items"), AST.param("callback")],
        AST.blockExpression([
          AST.forIn(
            "item",
            AST.id("items"),
            AST.functionCall(AST.id("callback"), [AST.id("item")]),
          ),
        ]),
        undefined,
        undefined,
        undefined,
        false,
        false,
      ),
      AST.assign("numbers", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
      AST.assign("total", AST.int(0)),
      AST.functionCall(
        AST.id("for_each"),
        [
          AST.id("numbers"),
          AST.lambdaExpression(
            [AST.param("n")],
            AST.assign("total", AST.id("n"), "+="),
          ),
        ],
        undefined,
        true,
      ),
      AST.id("total"),
    ]),
    manifest: {
      description: "Trailing lambda iterates array and accumulates values",
      expect: {
        result: { kind: "i32", value: 6 },
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
        result: { kind: "i32", value: 7 },
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
        result: { kind: "i32", value: 6 },
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
        result: { kind: "i32", value: 3 },
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
        result: { kind: "i32", value: 12 },
      },
    },
  },
  {
    name: "control/iterator_for_loop",
    module: AST.module([
      AST.assign("sum", AST.integerLiteral(0)),
      AST.assign(
        "iter",
        AST.iteratorLiteral([
          AST.forLoop(
            AST.identifier("item"),
            AST.arrayLiteral([
              AST.integerLiteral(1),
              AST.integerLiteral(2),
              AST.integerLiteral(3),
            ]),
            AST.blockExpression([
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                [AST.identifier("item")],
              ),
            ]),
          ),
        ]),
      ),
      AST.forLoop(
        AST.identifier("value"),
        AST.identifier("iter"),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("sum"),
            AST.binaryExpression(
              "+",
              AST.identifier("sum"),
              AST.identifier("value"),
            ),
          ),
        ]),
      ),
      AST.identifier("sum"),
    ]),
    manifest: {
      description: "For loop drives iterator literal lazily",
      expect: {
        result: { kind: "i32", value: 6 },
      },
    },
  },
  {
    name: "control/iterator_while_loop",
    module: AST.module([
      AST.assign("count", AST.integerLiteral(0)),
      AST.assign(
        "iter",
        AST.iteratorLiteral([
          AST.whileLoop(
            AST.binaryExpression("<", AST.identifier("count"), AST.integerLiteral(3)),
            AST.blockExpression([
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                [AST.identifier("count")],
              ),
              AST.assignmentExpression(
                "=",
                AST.identifier("count"),
                AST.binaryExpression(
                  "+",
                  AST.identifier("count"),
                  AST.integerLiteral(1),
                ),
              ),
            ]),
          ),
        ]),
      ),
      AST.assign("total", AST.integerLiteral(0)),
      AST.forLoop(
        AST.identifier("n"),
        AST.identifier("iter"),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("total"),
            AST.binaryExpression(
              "+",
              AST.identifier("total"),
              AST.identifier("n"),
            ),
          ),
        ]),
      ),
      AST.arrayLiteral([
        AST.identifier("total"),
        AST.identifier("count"),
      ]),
    ]),
    manifest: {
      description: "While loop inside iterator literal updates captured state lazily",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "i32", value: 3 },
            { kind: "i32", value: 3 },
          ],
        },
      },
    },
  },
  {
    name: "control/iterator_if_match",
    module: AST.module([
      AST.assign("calls", AST.integerLiteral(0)),
      AST.functionDefinition(
        "tick",
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("calls"),
            AST.binaryExpression(
              "+",
              AST.identifier("calls"),
              AST.integerLiteral(1),
            ),
          ),
          AST.returnStatement(AST.booleanLiteral(true)),
        ]),
      ),
      AST.assign("subject_calls", AST.integerLiteral(0)),
      AST.assign("guard_calls", AST.integerLiteral(0)),
      AST.functionDefinition(
        "get_subject",
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("subject_calls"),
            AST.binaryExpression(
              "+",
              AST.identifier("subject_calls"),
              AST.integerLiteral(1),
            ),
          ),
          AST.returnStatement(AST.integerLiteral(1)),
        ]),
      ),
      AST.functionDefinition(
        "guard_check",
        [AST.functionParameter("value")],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("guard_calls"),
            AST.binaryExpression(
              "+",
              AST.identifier("guard_calls"),
              AST.integerLiteral(1),
            ),
          ),
          AST.returnStatement(AST.booleanLiteral(true)),
        ]),
      ),
      AST.assign(
        "iter",
        AST.iteratorLiteral([
          AST.ifExpression(
            AST.functionCall(AST.identifier("tick"), []),
            AST.blockExpression([
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                [AST.integerLiteral(10)],
              ),
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                [AST.integerLiteral(20)],
              ),
            ]),
          ),
          AST.matchExpression(
            AST.functionCall(AST.identifier("get_subject"), []),
            [
              AST.matchClause(
                AST.literalPattern(AST.integerLiteral(1)),
                AST.blockExpression([
                  AST.functionCall(
                    AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                    [AST.integerLiteral(30)],
                  ),
                  AST.functionCall(
                    AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                    [AST.integerLiteral(40)],
                  ),
                  AST.integerLiteral(0),
                ]),
                AST.functionCall(AST.identifier("guard_check"), [AST.integerLiteral(1)]),
              ),
            ],
          ),
        ]),
      ),
      AST.assign("total", AST.integerLiteral(0)),
      AST.forLoop(
        AST.identifier("value"),
        AST.identifier("iter"),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("total"),
            AST.binaryExpression(
              "+",
              AST.identifier("total"),
              AST.identifier("value"),
            ),
          ),
        ]),
      ),
      AST.arrayLiteral([
        AST.identifier("total"),
        AST.identifier("calls"),
        AST.identifier("subject_calls"),
        AST.identifier("guard_calls"),
      ]),
    ]),
    manifest: {
      description: "Iterator literal preserves single evaluation of if conditions and match guards",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "i32", value: 100 },
            { kind: "i32", value: 1 },
            { kind: "i32", value: 1 },
            { kind: "i32", value: 1 },
          ],
        },
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
  {
    name: "control/while_sum",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.assign("i", AST.int(0)),
      AST.assign("limit", AST.int(3)),
      AST.wloop(
        AST.bin("<", AST.id("i"), AST.id("limit")),
        AST.block(
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("i")), "="),
          AST.assign("i", AST.bin("+", AST.id("i"), AST.int(1)), "="),
        ),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "Sums integers using a while loop",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "control/if_stdout",
    module: AST.module([
      AST.ifExpression(
        AST.bool(true),
        AST.block(
          AST.call("print", AST.str("branch")),
        ),
      ),
      AST.str("done"),
    ]),
    manifest: {
      description: "If expression emits stdout",
      expect: {
        stdout: ["branch"],
        result: { kind: "string", value: "done" },
      },
    },
  },
  {
    name: "control/if_else_branch",
    module: AST.module([
      AST.ifExpression(
        AST.bool(false),
        AST.block(AST.call("print", AST.str("true"))),
        [AST.orClause(AST.block(AST.call("print", AST.str("false"))))],
      ),
      AST.str("after"),
    ]),
    manifest: {
      description: "Else branch executes when condition false",
      expect: {
        stdout: ["false"],
        result: { kind: "string", value: "after" },
      },
    },
  },
  {
    name: "control/if_or_else",
    module: AST.module([
      AST.assign("score", AST.int(85)),
      AST.assign(
        "grade",
        AST.ifExpression(
          AST.bin(">=", AST.id("score"), AST.int(90)),
          AST.block(AST.str("A")),
          [
            AST.orClause(
              AST.block(AST.str("B")),
              AST.bin(">=", AST.id("score"), AST.int(80)),
            ),
            AST.orClause(AST.block(AST.str("C or lower"))),
          ],
        ),
      ),
      AST.id("grade"),
    ]),
    manifest: {
      description: "If-or chain picks first matching clause with default fallback",
      expect: {
        result: { kind: "string", value: "B" },
      },
    },
  },
  {
    name: "control/for_sum",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.assign("items", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
      AST.forIn("n", AST.id("items"), AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "=")),
      AST.id("sum"),
    ]),
    manifest: {
      description: "For loop iterates over array",
      expect: {
        result: { kind: "i32", value: 6 },
      },
    },
  },
  {
    name: "control/for_continue",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.assign("items", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
      AST.forIn(
        "n",
        AST.id("items"),
        AST.ifExpression(
          AST.bin("==", AST.id("n"), AST.int(2)),
          AST.block(AST.continueStatement()),
        ),
        AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "="),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "For loop continue skips matching elements",
      expect: {
        result: { kind: "i32", value: 4 },
      },
    },
  },
  {
    name: "control/for_range_break",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.forIn(
        "n",
        AST.range(AST.int(0), AST.int(5), false),
        AST.block(
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "="),
          AST.ifExpression(
            AST.bin(">=", AST.id("n"), AST.int(2)),
            AST.block(AST.brk(undefined, AST.id("sum"))),
          ),
        ),
      ),
    ]),
    manifest: {
      description: "For loop over range with break",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "control/range_inclusive",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.forIn(
        "n",
        AST.range(AST.int(0), AST.int(5), true),
        AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "="),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "Inclusive range includes upper bound during iteration",
      expect: {
        result: { kind: "i32", value: 15 },
      },
    },
  },
  {
    name: "patterns/array_destructuring",
    module: AST.module([
      AST.assign("arr", AST.arr(AST.int(1), AST.int(2), AST.int(3), AST.int(4))),
      AST.assign(AST.arrP([AST.id("first"), AST.id("second")], AST.id("rest")), AST.id("arr")),
      AST.assign(AST.arrP([AST.id("third"), AST.id("fourth")]), AST.id("rest")),
      AST.bin("+", AST.id("first"), AST.bin("+", AST.id("second"), AST.id("third"))),
    ]),
    manifest: {
      description: "Array destructuring assignment extracts prefix and rest",
      expect: {
        result: { kind: "i32", value: 6 },
      },
    },
  },
  {
    name: "patterns/struct_positional_destructuring",
    module: AST.module([
      AST.structDefinition(
        "Pair",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
          AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
        ],
        "positional",
      ),
      AST.assign(
        "pair",
        AST.structLiteral(
          [
            AST.structFieldInitializer(AST.integerLiteral(4)),
            AST.structFieldInitializer(AST.integerLiteral(8)),
          ],
          true,
          "Pair",
        ),
      ),
      AST.assign(
        AST.structPattern(
          [
            AST.structPatternField(AST.identifier("first")),
            AST.structPatternField(AST.identifier("second")),
          ],
          true,
          "Pair",
        ),
        AST.identifier("pair"),
      ),
      AST.binaryExpression(
        "+",
        AST.identifier("first"),
        AST.identifier("second"),
      ),
    ]),
    manifest: {
      description: "Positional struct destructuring assignment binds tuple fields",
      expect: {
        result: { kind: "i32", value: 12 },
      },
    },
  },
  {
    name: "patterns/nested_struct_destructuring",
    module: AST.module([
      AST.structDefinition(
        "Point",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
        ],
        "named",
      ),
      AST.structDefinition(
        "Wrapper",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("Point"), "left"),
          AST.structFieldDefinition(AST.simpleTypeExpression("Point"), "right"),
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
        "wrapper",
        AST.structLiteral(
          [
            AST.structFieldInitializer(
              AST.structLiteral(
                [
                  AST.structFieldInitializer(AST.integerLiteral(1), "x"),
                  AST.structFieldInitializer(AST.integerLiteral(2), "y"),
                ],
                false,
                "Point",
              ),
              "left",
            ),
            AST.structFieldInitializer(
              AST.structLiteral(
                [
                  AST.structFieldInitializer(AST.integerLiteral(3), "x"),
                  AST.structFieldInitializer(AST.integerLiteral(4), "y"),
                ],
                false,
                "Point",
              ),
              "right",
            ),
            AST.structFieldInitializer(
              AST.arrayLiteral([
                AST.integerLiteral(10),
                AST.integerLiteral(20),
                AST.integerLiteral(30),
              ]),
              "values",
            ),
          ],
          false,
          "Wrapper",
        ),
      ),
      AST.assign(
        AST.structPattern(
          [
            AST.structPatternField(
              AST.structPattern(
                [
                  AST.structPatternField(AST.identifier("left_x"), "x"),
                  AST.structPatternField(AST.identifier("left_y"), "y"),
                ],
                false,
                "Point",
              ),
              "left",
            ),
            AST.structPatternField(
              AST.structPattern(
                [
                  AST.structPatternField(AST.wildcardPattern(), "x"),
                  AST.structPatternField(AST.identifier("right_y"), "y"),
                ],
                false,
                "Point",
              ),
              "right",
            ),
            AST.structPatternField(
              AST.arrayPattern(
                [AST.identifier("first_value")],
                "rest_values",
              ),
              "values",
            ),
          ],
          false,
          "Wrapper",
        ),
        AST.identifier("wrapper"),
      ),
      AST.binaryExpression(
        "+",
        AST.binaryExpression(
          "+",
          AST.binaryExpression(
            "+",
            AST.identifier("left_x"),
            AST.identifier("right_y"),
          ),
          AST.identifier("first_value"),
        ),
        AST.index(AST.identifier("rest_values"), AST.integerLiteral(0)),
      ),
    ]),
    manifest: {
      description: "Nested struct and array patterns destructure composite value",
      expect: {
        result: { kind: "i32", value: 35 },
      },
    },
  },
  {
    name: "patterns/for_array_pattern",
    module: AST.module([
      AST.assign("pairs", AST.arr(
        AST.arr(AST.int(1), AST.int(2)),
        AST.arr(AST.int(3), AST.int(4)),
      )),
      AST.assign("sum", AST.int(0)),
      AST.forIn(
        AST.arrP([AST.id("x"), AST.id("y")]),
        AST.id("pairs"),
        AST.block(
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("x")), "="),
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("y")), "="),
        ),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "For-in loop destructures array elements",
      expect: {
        result: { kind: "i32", value: 10 },
      },
    },
  },
  {
    name: "patterns/typed_assignment",
    module: AST.module([
      AST.assign("value", AST.int(42)),
      AST.assign(
        AST.typedP(AST.id("n"), AST.ty("i32")),
        AST.id("value"),
      ),
      AST.id("n"),
    ]),
    manifest: {
      description: "Typed pattern enforces simple type on assignment",
      expect: {
        result: { kind: "i32", value: 42 },
      },
    },
  },
  {
    name: "patterns/typed_assignment_error",
    module: AST.module([
      AST.assign("value", AST.str("nope")),
      AST.assign(
        AST.typedP(AST.id("n"), AST.ty("i32")),
        AST.id("value"),
      ),
    ]),
    manifest: {
      description: "Typed pattern mismatch raises error",
      expect: {
        errors: ["Typed pattern mismatch in assignment"],
      },
    },
  },
  {
    name: "structs/named_literal",
    module: AST.module([
      AST.structDefinition(
        "Point",
        [
          AST.fieldDef(AST.ty("i32"), "x"),
          AST.fieldDef(AST.ty("i32"), "y"),
        ],
        "named",
      ),
      AST.assign(
        "point",
        AST.structLiteral(
          [
            AST.fieldInit(AST.int(3), "x"),
            AST.fieldInit(AST.int(4), "y"),
          ],
          false,
          "Point",
        ),
      ),
      AST.member(AST.id("point"), "x"),
    ]),
    manifest: {
      description: "Named struct literal evaluates and exposes fields",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "structs/positional_literal",
    module: AST.module([
      AST.structDefinition(
        "Pair",
        [
          AST.fieldDef(AST.ty("i32")),
          AST.fieldDef(AST.ty("i32")),
        ],
        "positional",
      ),
      AST.assign(
        "pair",
        AST.structLiteral(
          [
            AST.fieldInit(AST.int(7)),
            AST.fieldInit(AST.int(9)),
          ],
          true,
          "Pair",
        ),
      ),
      AST.member(AST.id("pair"), AST.int(1)),
    ]),
    manifest: {
      description: "Positional struct literal supports numeric member access",
      expect: {
        result: { kind: "i32", value: 9 },
      },
    },
  },
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
  {
    name: "imports/dynimport_selector_alias",
    setupModules: {
      "package.json": AST.module(
        [
          AST.functionDefinition(
            "f",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))]),
          ),
          AST.functionDefinition(
            "hidden",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
        ],
        [],
        AST.packageStatement(["dynp"]),
      ),
    },
    module: AST.module([
      AST.dynImportStatement(["dynp"], false, [AST.importSelector("f", "ff")]),
      AST.dynImportStatement(["dynp"], false, undefined, "D"),
      AST.assignmentExpression(
        ":=",
        AST.identifier("x"),
        AST.call(AST.identifier("ff")),
      ),
      AST.assignmentExpression(
        ":=",
        AST.identifier("y"),
        AST.call(AST.memberAccessExpression(AST.identifier("D"), "f")),
      ),
      AST.binaryExpression("+", AST.identifier("x"), AST.identifier("y")),
    ]),
    manifest: {
      description: "Dyn import selector and alias return callable references",
      expect: {
        result: { kind: "i32", value: 22 },
      },
    },
  },
  {
    name: "imports/dynimport_wildcard",
    setupModules: {
      "package.json": AST.module(
        [
          AST.functionDefinition(
            "f",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))]),
          ),
          AST.functionDefinition(
            "hidden",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
        ],
        [],
        AST.packageStatement(["dynp"]),
      ),
    },
    module: AST.module([
      AST.dynImportStatement(["dynp"], true),
      AST.assignmentExpression(
        ":=",
        AST.identifier("value"),
        AST.call(AST.identifier("f")),
      ),
      AST.identifier("value"),
    ]),
    manifest: {
      description: "Dyn import wildcard exposes public symbols",
      expect: {
        result: { kind: "i32", value: 11 },
      },
    },
  },
  {
    name: "imports/dynimport_private_selector_error",
    setupModules: {
      "package.json": AST.module(
        [
          AST.functionDefinition(
            "hidden",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
        ],
        [],
        AST.packageStatement(["dynp"]),
      ),
    },
    module: AST.module([
      AST.dynImportStatement(["dynp"], false, [AST.importSelector("hidden")]),
    ]),
    manifest: {
      description: "Dyn import selector rejects private symbols",
      expect: {
        errors: ["dynimport error: function 'hidden' is private"],
      },
    },
  },
  {
    name: "concurrency/proc_cancel_value",
    module: AST.module([
      AST.assign(
        "handle",
        AST.procExpression(AST.blockExpression([AST.integerLiteral(0)])),
      ),
      AST.assign(
        "_cancelResult",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
          [],
        ),
      ),
      AST.assign(
        "result",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("handle"), "value"),
          [],
        ),
      ),
      AST.identifier("result"),
    ]),
    manifest: {
      description: "Proc cancellation before start returns an error value",
      expect: {
        result: { kind: "error" },
      },
    },
  },
  {
    name: "concurrency/future_memoization",
    module: AST.module([
      AST.assign("count", AST.integerLiteral(0)),
      AST.assign(
        "future",
        AST.spawnExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              "+=",
              AST.identifier("count"),
              AST.integerLiteral(1),
            ),
            AST.integerLiteral(1),
          ]),
        ),
      ),
      AST.assign(
        "first",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("future"), "value"),
          [],
        ),
      ),
      AST.assign(
        "second",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("future"), "value"),
          [],
        ),
      ),
      AST.identifier("count"),
    ]),
    manifest: {
      description: "Future value memoises results and runs the task only once",
      expect: {
        result: { kind: "i32", value: 1 },
      },
    },
  },
  {
    name: "concurrency/proc_cancelled_outside_error",
    module: AST.module([
      AST.functionCall(AST.identifier("proc_cancelled"), []),
    ]),
    manifest: {
      description: "proc_cancelled raises when called outside of proc/spawn",
      expect: {
        errors: ["proc_cancelled must be called inside an asynchronous task"],
      },
    },
  },
  {
    name: "concurrency/proc_yield_flush",
    module: AST.module([
      AST.assign("stage", AST.integerLiteral(0)),
      AST.assign("trace", AST.stringLiteral("")),
      AST.assign(
        "handle",
        AST.procExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage"),
                AST.integerLiteral(0),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage"),
                  AST.integerLiteral(1),
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
              ]),
              [],
            ),
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage"),
                AST.integerLiteral(1),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("B"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage"),
                  AST.integerLiteral(2),
                ),
              ]),
              [],
            ),
            AST.stringLiteral("done"),
          ]),
        ),
      ),
      AST.assign(
        "status_before",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("handle"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "status_mid",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("handle"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.assign(
        "result",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("handle"), "value"),
          [],
        ),
      ),
      AST.assign(
        "status_after",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("handle"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.stringInterpolation([
        AST.identifier("status_before"),
        AST.stringLiteral(":"),
        AST.identifier("status_mid"),
        AST.stringLiteral(":"),
        AST.identifier("status_after"),
        AST.stringLiteral(":"),
        AST.identifier("trace"),
        AST.stringLiteral(":"),
        AST.identifier("result"),
      ]),
    ]),
    manifest: {
      description: "proc_yield cooperates with proc_flush to resume the task",
      expect: {
        result: { kind: "string", value: "Pending:Resolved:Resolved:AB:done" },
      },
    },
  },
  {
    name: "concurrency/proc_cancelled_helper",
    module: AST.module([
      AST.assign("trace", AST.stringLiteral("")),
      AST.assign(
        "handle",
        AST.procExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              "=",
              AST.identifier("trace"),
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.stringLiteral("A"),
              ),
            ),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
              [],
            ),
            AST.ifExpression(
              AST.functionCall(AST.identifier("proc_cancelled"), []),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("C"),
                  ),
                ),
              ]),
              [],
            ),
            AST.integerLiteral(0),
          ]),
        ),
      ),
      AST.assign(
        "_result",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("handle"), "value"),
          [],
        ),
      ),
      AST.identifier("trace"),
    ]),
    manifest: {
      description: "Proc uses proc_cancelled() after yielding to observe cancellation flag",
      expect: {
        result: { kind: "string", value: "AC" },
      },
    },
  },
  {
    name: "concurrency/fairness_proc_round_robin",
    module: AST.module([
      AST.assign("trace", AST.stringLiteral("")),
      AST.assign("stage_a", AST.integerLiteral(0)),
      AST.assign("stage_b", AST.integerLiteral(0)),
      AST.assign(
        "worker",
        AST.procExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_a"),
                AST.integerLiteral(0),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A1"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_a"),
                  AST.integerLiteral(1),
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
              ]),
              [],
            ),
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_a"),
                AST.integerLiteral(1),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A2"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_a"),
                  AST.integerLiteral(2),
                ),
              ]),
              [],
            ),
            AST.integerLiteral(0),
          ]),
        ),
      ),
      AST.assign(
        "other",
        AST.procExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_b"),
                AST.integerLiteral(0),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("B1"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_b"),
                  AST.integerLiteral(1),
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
              ]),
              [],
            ),
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_b"),
                AST.integerLiteral(1),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("B2"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_b"),
                  AST.integerLiteral(2),
                ),
              ]),
              [],
            ),
            AST.integerLiteral(0),
          ]),
        ),
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "status_worker",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("worker"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.assign(
        "status_other",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("other"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.stringInterpolation([
        AST.identifier("trace"),
        AST.stringLiteral(":"),
        AST.identifier("status_worker"),
        AST.stringLiteral(":"),
        AST.identifier("status_other"),
      ]),
    ]),
    manifest: {
      description: "Serial executor yields alternate between procs when proc_yield is used",
      expect: {
        result: { kind: "string", value: "A1B1A2B2:Resolved:Resolved" },
      },
    },
  },
  {
    name: "concurrency/fairness_proc_future",
    module: AST.module([
      AST.assign("trace", AST.stringLiteral("")),
      AST.assign("stage_proc", AST.integerLiteral(0)),
      AST.assign("stage_future", AST.integerLiteral(0)),
      AST.assign(
        "worker",
        AST.procExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_proc"),
                AST.integerLiteral(0),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A1"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_proc"),
                  AST.integerLiteral(1),
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
              ]),
              [],
            ),
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_proc"),
                AST.integerLiteral(1),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A2"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_proc"),
                  AST.integerLiteral(2),
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
              ]),
              [],
            ),
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_proc"),
                AST.integerLiteral(2),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A3"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_proc"),
                  AST.integerLiteral(3),
                ),
              ]),
              [],
            ),
            AST.integerLiteral(0),
          ]),
        ),
      ),
      AST.assign(
        "future",
        AST.spawnExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage_future"),
                AST.integerLiteral(0),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("B1"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage_future"),
                  AST.integerLiteral(1),
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
                AST.integerLiteral(0),
              ]),
              [
                AST.orClause(
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("trace"),
                      AST.binaryExpression(
                        "+",
                        AST.identifier("trace"),
                        AST.stringLiteral("B2"),
                      ),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("stage_future"),
                      AST.integerLiteral(2),
                    ),
                    AST.integerLiteral(0),
                  ]),
                  AST.binaryExpression(
                    "==",
                    AST.identifier("stage_future"),
                    AST.integerLiteral(1),
                  ),
                ),
              ],
            ),
            AST.integerLiteral(0),
          ]),
        ),
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "worker_status",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("worker"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.assign(
        "future_status",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("future"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.assign(
        "future_result",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("future"), "value"),
          [],
        ),
      ),
      AST.stringInterpolation([
        AST.identifier("trace"),
        AST.stringLiteral(":"),
        AST.identifier("worker_status"),
        AST.stringLiteral(":"),
        AST.identifier("future_status"),
        AST.stringLiteral(":"),
        AST.identifier("future_result"),
      ]),
    ]),
    manifest: {
      description: "Proc and future alternate via proc_yield under the serial executor",
      expect: {
        result: { kind: "string", value: "A1B1A2B2A3:Resolved:Resolved:0" },
      },
    },
  },
  {
    name: "concurrency/channel_basic_ops",
    module: AST.module([
      AST.structDefinition(
        "Channel",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "capacity"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle"),
        ],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Channel"),
        [
          AST.functionDefinition(
            "new",
            [AST.functionParameter("capacity", AST.simpleTypeExpression("i32"))],
            AST.blockExpression([
              AST.assignmentExpression(
                ":=",
                AST.identifier("handle"),
                AST.functionCall(AST.identifier("__able_channel_new"), [
                  AST.identifier("capacity"),
                ]),
              ),
              AST.returnStatement(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.identifier("capacity"), "capacity"),
                    AST.structFieldInitializer(AST.identifier("handle"), "handle"),
                  ],
                  false,
                  "Channel",
                ),
              ),
            ]),
            AST.simpleTypeExpression("Channel"),
          ),
          AST.functionDefinition(
            "send",
            [AST.functionParameter("self"), AST.functionParameter("value")],
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("__able_channel_send"),
                [
                  AST.memberAccessExpression(AST.identifier("self"), "handle"),
                  AST.identifier("value"),
                ],
              ),
            ]),
          ),
          AST.functionDefinition(
            "receive",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.returnStatement(
                AST.functionCall(
                  AST.identifier("__able_channel_receive"),
                  [AST.memberAccessExpression(AST.identifier("self"), "handle")],
                ),
              ),
            ]),
          ),
          AST.functionDefinition(
            "try_send",
            [AST.functionParameter("self"), AST.functionParameter("value")],
            AST.blockExpression([
              AST.returnStatement(
                AST.functionCall(
                  AST.identifier("__able_channel_try_send"),
                  [
                    AST.memberAccessExpression(AST.identifier("self"), "handle"),
                    AST.identifier("value"),
                  ],
                ),
              ),
            ]),
          ),
          AST.functionDefinition(
            "try_receive",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.returnStatement(
                AST.functionCall(
                  AST.identifier("__able_channel_try_receive"),
                  [AST.memberAccessExpression(AST.identifier("self"), "handle")],
                ),
              ),
            ]),
          ),
          AST.functionDefinition(
            "close",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("__able_channel_close"),
                [AST.memberAccessExpression(AST.identifier("self"), "handle")],
              ),
            ]),
          ),
          AST.functionDefinition(
            "is_closed",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.returnStatement(
                AST.functionCall(
                  AST.identifier("__able_channel_is_closed"),
                  [AST.memberAccessExpression(AST.identifier("self"), "handle")],
                ),
              ),
            ]),
          ),
        ],
      ),
      AST.assign(
        "channel",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("Channel"), "new"),
          [AST.integerLiteral(1)],
        ),
      ),
      AST.assign("score", AST.integerLiteral(0)),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("channel"), "send"),
        [AST.integerLiteral(11)],
      ),
      AST.assign(
        "first",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "receive"),
          [],
        ),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("first"),
          AST.integerLiteral(11),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "try_success",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "try_send"),
          [AST.integerLiteral(21)],
        ),
      ),
      AST.ifExpression(
        AST.identifier("try_success"),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "try_receive_value",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "try_receive"),
          [],
        ),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("try_receive_value"),
          AST.integerLiteral(21),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "second_try",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "try_send"),
          [AST.integerLiteral(31)],
        ),
      ),
      AST.ifExpression(
        AST.identifier("second_try"),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "try_fail",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "try_send"),
          [AST.integerLiteral(41)],
        ),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("try_fail"),
          AST.booleanLiteral(false),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "second",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "receive"),
          [],
        ),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("second"),
          AST.integerLiteral(31),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("channel"), "close"),
        [],
      ),
      AST.assign(
        "closed",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "is_closed"),
          [],
        ),
      ),
      AST.ifExpression(
        AST.identifier("closed"),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "final_receive",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "receive"),
          [],
        ),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("final_receive"),
          AST.nilLiteral(),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.assign(
        "try_receive_nil",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "try_receive"),
          [],
        ),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("try_receive_nil"),
          AST.nilLiteral(),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.binaryExpression(
        "==",
        AST.identifier("score"),
        AST.integerLiteral(9),
      ),
    ]),
    manifest: {
      description:
        "Channel.new, send/receive, non-blocking ops, and close/is_closed behave as expected",
      expect: {
        result: { kind: "bool", value: true },
      },
    },
  },
  {
    name: "concurrency/channel_receive_loop",
    module: AST.module([
      AST.structDefinition(
        "Channel",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "capacity"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle"),
        ],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Channel"),
        [
          AST.functionDefinition(
            "new",
            [AST.functionParameter("capacity", AST.simpleTypeExpression("i32"))],
            AST.blockExpression([
              AST.assignmentExpression(
                ":=",
                AST.identifier("handle"),
                AST.functionCall(AST.identifier("__able_channel_new"), [
                  AST.identifier("capacity"),
                ]),
              ),
              AST.returnStatement(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.identifier("capacity"), "capacity"),
                    AST.structFieldInitializer(AST.identifier("handle"), "handle"),
                  ],
                  false,
                  "Channel",
                ),
              ),
            ]),
            AST.simpleTypeExpression("Channel"),
          ),
          AST.functionDefinition(
            "send",
            [AST.functionParameter("self"), AST.functionParameter("value")],
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("__able_channel_send"),
                [
                  AST.memberAccessExpression(AST.identifier("self"), "handle"),
                  AST.identifier("value"),
                ],
              ),
            ]),
          ),
          AST.functionDefinition(
            "receive",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.returnStatement(
                AST.functionCall(
                  AST.identifier("__able_channel_receive"),
                  [AST.memberAccessExpression(AST.identifier("self"), "handle")],
                ),
              ),
            ]),
          ),
          AST.functionDefinition(
            "close",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("__able_channel_close"),
                [AST.memberAccessExpression(AST.identifier("self"), "handle")],
              ),
            ]),
          ),
        ],
      ),
      AST.assign(
        "channel",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("Channel"), "new"),
          [AST.integerLiteral(3)],
        ),
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("channel"), "send"),
        [AST.integerLiteral(2)],
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("channel"), "send"),
        [AST.integerLiteral(3)],
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("channel"), "close"),
        [],
      ),
      AST.assign("sum", AST.integerLiteral(0)),
      AST.assign("value", AST.nilLiteral()),
      AST.whileLoop(
        AST.booleanLiteral(true),
        AST.blockExpression([
          AST.assignmentExpression(
            ":=",
            AST.identifier("value"),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("channel"), "receive"),
              [],
            ),
          ),
          AST.ifExpression(
            AST.binaryExpression(
              "==",
              AST.identifier("value"),
              AST.nilLiteral(),
            ),
            AST.blockExpression([AST.breakStatement()]),
            [AST.orClause(
              AST.blockExpression([
                AST.assignmentExpression(
                  "+=",
                  AST.identifier("sum"),
                  AST.identifier("value"),
                ),
              ]),
            )],
          ),
        ]),
      ),
      AST.identifier("sum"),
    ]),
    manifest: {
      description: "Channel.receive drains buffered values and returns nil after close",
      expect: {
        result: { kind: "i32", value: 5 },
      },
    },
  },
  {
    name: "concurrency/channel_send_on_closed_error",
    module: AST.module([
      AST.assign(
        "handle",
        AST.functionCall(AST.identifier("__able_channel_new"), [AST.integerLiteral(0)]),
      ),
      AST.functionCall(
        AST.identifier("__able_channel_close"),
        [AST.identifier("handle")],
      ),
      AST.functionCall(
        AST.identifier("__able_channel_send"),
        [AST.identifier("handle"), AST.integerLiteral(1)],
      ),
    ]),
    manifest: {
      description: "Sending on a closed channel raises an error",
      expect: {
        errors: ["send on closed channel"],
      },
    },
  },
  {
    name: "concurrency/channel_nil_send_cancel",
    module: AST.module([
      AST.assign(
        "handle",
        AST.procExpression(
          AST.blockExpression([
            AST.functionCall(
              AST.identifier("__able_channel_send"),
              [AST.integerLiteral(0), AST.stringLiteral("value")],
            ),
            AST.integerLiteral(1),
          ]),
        ),
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
        [],
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "status",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("handle"), "status"),
          [],
        ),
      ),
      AST.matchExpression(
        AST.identifier("status"),
        [
          AST.matchClause(
            AST.structPattern([], false, "Cancelled"),
            AST.stringLiteral("Cancelled"),
          ),
          AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
        ],
      ),
    ]),
    manifest: {
      description: "Nil channel send blocks until the proc is cancelled",
      expect: {
        result: { kind: "string", value: "Cancelled" },
      },
    },
  },
  {
    name: "concurrency/channel_nil_receive_cancel",
    module: AST.module([
      AST.assign(
        "handle",
        AST.procExpression(
          AST.blockExpression([
            AST.functionCall(
              AST.identifier("__able_channel_receive"),
              [AST.integerLiteral(0)],
            ),
            AST.integerLiteral(1),
          ]),
        ),
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
        [],
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "status",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("handle"), "status"),
          [],
        ),
      ),
      AST.matchExpression(
        AST.identifier("status"),
        [
          AST.matchClause(
            AST.structPattern([], false, "Cancelled"),
            AST.stringLiteral("Cancelled"),
          ),
          AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
        ],
      ),
    ]),
    manifest: {
      description: "Nil channel receive blocks until the proc is cancelled",
      expect: {
        result: { kind: "string", value: "Cancelled" },
      },
    },
  },
  {
    name: "concurrency/mutex_locking",
    module: AST.module([
      AST.structDefinition(
        "Mutex",
        [AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle")],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Mutex"),
        [
          AST.functionDefinition(
            "new",
            [],
            AST.blockExpression([
              AST.assignmentExpression(
                ":=",
                AST.identifier("handle"),
                AST.functionCall(AST.identifier("__able_mutex_new"), []),
              ),
              AST.returnStatement(
                AST.structLiteral(
                  [AST.structFieldInitializer(AST.identifier("handle"), "handle")],
                  false,
                  "Mutex",
                ),
              ),
            ]),
            AST.simpleTypeExpression("Mutex"),
          ),
          AST.functionDefinition(
            "lock",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("__able_mutex_lock"),
                [AST.memberAccessExpression(AST.identifier("self"), "handle")],
              ),
            ]),
          ),
          AST.functionDefinition(
            "unlock",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("__able_mutex_unlock"),
                [AST.memberAccessExpression(AST.identifier("self"), "handle")],
              ),
            ]),
          ),
        ],
      ),
      AST.assign(
        "mutex",
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("Mutex"), "new"),
          [],
        ),
      ),
      AST.assign("trace", AST.stringLiteral("")),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("mutex"), "lock"),
        [],
      ),
      AST.assignmentExpression(
        "=",
        AST.identifier("trace"),
        AST.binaryExpression(
          "+",
          AST.identifier("trace"),
          AST.stringLiteral("A"),
        ),
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("mutex"), "unlock"),
        [],
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("mutex"), "lock"),
        [],
      ),
      AST.assignmentExpression(
        "=",
        AST.identifier("trace"),
        AST.binaryExpression(
          "+",
          AST.identifier("trace"),
          AST.stringLiteral("B"),
        ),
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("mutex"), "unlock"),
        [],
      ),
      AST.identifier("trace"),
    ]),
    manifest: {
      description: "Mutex lock/unlock methods mutate in sequence",
      expect: {
        result: { kind: "string", value: "AB" },
      },
    },
  },
  {
    name: "concurrency/mutex_contention",
    module: AST.module([
      AST.assign(
        "mutex",
        AST.functionCall(AST.identifier("__able_mutex_new"), []),
      ),
      AST.assign("trace", AST.stringLiteral("")),
      AST.functionCall(
        AST.identifier("__able_mutex_lock"),
        [AST.identifier("mutex")],
      ),
      AST.assignmentExpression(
        "=",
        AST.identifier("trace"),
        AST.binaryExpression(
          "+",
          AST.identifier("trace"),
          AST.stringLiteral("A"),
        ),
      ),
      AST.assign(
        "worker",
        AST.procExpression(
          AST.blockExpression([
            AST.functionCall(
              AST.identifier("__able_mutex_lock"),
              [AST.identifier("mutex")],
            ),
            AST.assignmentExpression(
              "=",
              AST.identifier("trace"),
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.stringLiteral("C"),
              ),
            ),
            AST.functionCall(
              AST.identifier("__able_mutex_unlock"),
              [AST.identifier("mutex")],
            ),
            AST.nilLiteral(),
          ]),
        ),
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "status_initial",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("worker"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.assignmentExpression(
        "=",
        AST.identifier("trace"),
        AST.binaryExpression(
          "+",
          AST.identifier("trace"),
          AST.stringLiteral("B"),
        ),
      ),
      AST.functionCall(
        AST.identifier("__able_mutex_unlock"),
        [AST.identifier("mutex")],
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign(
        "status_final",
        AST.matchExpression(
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("worker"), "status"),
            [],
          ),
          [
            AST.matchClause(
              AST.structPattern([], false, "Pending"),
              AST.stringLiteral("Pending"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Resolved"),
              AST.stringLiteral("Resolved"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Cancelled"),
              AST.stringLiteral("Cancelled"),
            ),
            AST.matchClause(
              AST.structPattern([], false, "Failed"),
              AST.stringLiteral("Failed"),
            ),
            AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Other")),
          ],
        ),
      ),
      AST.stringInterpolation([
        AST.identifier("status_initial"),
        AST.stringLiteral(":"),
        AST.identifier("status_final"),
        AST.stringLiteral(":"),
        AST.identifier("trace"),
      ]),
    ]),
    manifest: {
      description: "Mutex contention ensures the waiting proc resumes only after unlock",
      expect: {
        result: { kind: "string", value: "Pending:Resolved:ABC" },
      },
    },
  },
  {
    name: "stdlib/channel_mutex_helpers",
    module: AST.module([
      AST.structDefinition(
        "Channel",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "capacity"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle"),
        ],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Channel"),
        [
          AST.functionDefinition(
            "new",
            [],
            AST.blockExpression([
              AST.assignmentExpression(
                ":=",
                AST.identifier("handle"),
                AST.functionCall(AST.identifier("__able_channel_new"), [
                  AST.integerLiteral(0),
                ]),
              ),
              AST.returnStatement(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.integerLiteral(0), "capacity"),
                    AST.structFieldInitializer(AST.identifier("handle"), "handle"),
                  ],
                  false,
                  "Channel",
                ),
              ),
            ]),
            AST.simpleTypeExpression("Channel"),
          ),
        ],
      ),
      AST.structDefinition(
        "Mutex",
        [AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle")],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Mutex"),
        [
          AST.functionDefinition(
            "new",
            [],
            AST.blockExpression([
              AST.assignmentExpression(
                ":=",
                AST.identifier("handle"),
                AST.functionCall(AST.identifier("__able_mutex_new"), []),
              ),
              AST.returnStatement(
                AST.structLiteral(
                  [AST.structFieldInitializer(AST.identifier("handle"), "handle")],
                  false,
                  "Mutex",
                ),
              ),
            ]),
            AST.simpleTypeExpression("Mutex"),
          ),
        ],
      ),
      AST.fn(
        "channel_handle",
        [AST.param("capacity", AST.simpleTypeExpression("i32"))],
        [AST.ret(AST.call("__able_channel_new", AST.id("capacity")))],
        AST.simpleTypeExpression("i64"),
      ),
      AST.fn(
        "mutex_handle",
        [],
        [AST.ret(AST.call("__able_mutex_new"))],
        AST.simpleTypeExpression("i64"),
      ),
      AST.assign(
        "channel_handle_value",
        AST.call("channel_handle", AST.integerLiteral(0)),
      ),
      AST.assign(
        "channel_instance",
        AST.call(
          AST.memberAccessExpression(AST.identifier("Channel"), "new"),
        ),
      ),
      AST.assign(
        "mutex_instance",
        AST.call(
          AST.memberAccessExpression(AST.identifier("Mutex"), "new"),
        ),
      ),
      AST.assign(
        "mutex_handle_value",
        AST.call("mutex_handle"),
      ),
      AST.assign("score", AST.integerLiteral(0)),
      AST.ifExpression(
        AST.binaryExpression(
          "!=",
          AST.memberAccessExpression(AST.identifier("channel_instance"), "handle"),
          AST.integerLiteral(0),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.memberAccessExpression(AST.identifier("channel_instance"), "capacity"),
          AST.integerLiteral(0),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "!=",
          AST.identifier("channel_handle_value"),
          AST.integerLiteral(0),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "!=",
          AST.memberAccessExpression(AST.identifier("mutex_instance"), "handle"),
          AST.integerLiteral(0),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "!=",
          AST.identifier("mutex_handle_value"),
          AST.integerLiteral(0),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "+=",
            AST.identifier("score"),
            AST.integerLiteral(1),
          ),
        ]),
        [],
      ),
      AST.binaryExpression(
        "==",
        AST.identifier("score"),
        AST.integerLiteral(5),
      ),
    ]),
    manifest: {
      description: "Channel.new and Mutex.new expose runtime handles via stdlib helpers",
      expect: {
        result: { kind: "bool", value: true },
      },
    },
  },
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
      skipTargets: ["ts"],
    },
  },
  {
    name: "errors/rescue_guard",
    module: AST.module([
      AST.rescue(
        AST.block(AST.raise(AST.str("boom"))),
        [
          AST.mc(AST.litP(AST.str("ignore")), AST.str("ignored")),
          AST.mc(
            AST.id("msg"),
            AST.block(
              AST.ifExpression(
                AST.bin("==", AST.id("msg"), AST.str("boom")),
                AST.block(AST.str("handled")),
              ),
            ),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Rescue guard selects matching clause",
      expect: {
        result: { kind: "nil" },
      },
    },
  },
  {
    name: "errors/raise_manifest",
    module: AST.module([
      AST.raise(AST.str("boom")),
    ]),
    manifest: {
      description: "Fixture raises error",
      expect: {
        errors: ["boom"],
      },
    },
  },
  {
    name: "errors/rescue_catch",
    module: AST.module([
      AST.rescue(
        AST.block(AST.raise(AST.str("boom"))),
        [
          AST.mc(
            AST.id("err"),
            AST.block(AST.call("print", AST.id("err")), AST.str("handled")),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Rescue expression catches raise",
      expect: {
        stdout: ["[error]"],
        result: { kind: "string", value: "handled" },
      },
    },
  },
  {
    name: "errors/rescue_typed_pattern",
    module: AST.module([
      AST.rescue(
        AST.block(AST.raise(AST.str("boom"))),
        [
          AST.matchClause(
            AST.typedPattern(AST.identifier("err"), AST.simpleTypeExpression("Error")),
            AST.stringLiteral("caught"),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Typed pattern catches raised error",
      expect: {
        result: { kind: "string", value: "caught" },
      },
    },
  },
  {
    name: "errors/or_else_handler",
    module: AST.module([
      AST.orElseExpression(
        AST.propagationExpression(
          AST.blockExpression([AST.raiseStatement(AST.stringLiteral("boom"))]),
        ),
        AST.blockExpression([AST.stringLiteral("handled")]),
        "err",
      ),
    ]),
    manifest: {
      description: "Or else handler runs when propagation raises",
      expect: {
        result: { kind: "string", value: "handled" },
      },
    },
  },
  {
    name: "errors/ensure_runs",
    module: AST.module([
      AST.ensureExpression(
        AST.rescueExpression(
          AST.blockExpression([AST.raiseStatement(AST.stringLiteral("oops"))]),
          [AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("rescued"))],
        ),
        AST.blockExpression([AST.call("print", AST.stringLiteral("ensure"))]),
      ),
    ]),
    manifest: {
      description: "Ensure block executes regardless of rescue",
      expect: {
        stdout: ["ensure"],
        result: { kind: "string", value: "rescued" },
      },
    },
  },
];

const SOURCE_EXCLUDE = new Set([
  "match/struct_positional_pattern",
  "patterns/struct_positional_destructuring",
]);

async function main() {
  for (const fixture of fixtures) {
    await writeFixture(fixture);
  }
  console.log(`Wrote ${fixtures.length} fixture(s) to ${FIXTURE_ROOT}`);
}

async function writeFixture(fixture: Fixture) {
  const outDir = path.join(FIXTURE_ROOT, fixture.name);
  await fs.mkdir(outDir, { recursive: true });

  normalizeModule(fixture.module);

  if (fixture.setupModules) {
    for (const [fileName, module] of Object.entries(fixture.setupModules)) {
      normalizeModule(module);
      const filePath = path.join(outDir, fileName);
      await fs.writeFile(filePath, stringify(module), "utf8");
    }
  }

  const modulePath = path.join(outDir, "module.json");
  await fs.writeFile(modulePath, stringify(fixture.module), "utf8");

  if (!SOURCE_EXCLUDE.has(fixture.name)) {
    const sourcePath = path.join(outDir, "source.able");
    const source = moduleToSource(fixture.module);
    await fs.writeFile(sourcePath, source.endsWith("\n") ? source : `${source}\n`, "utf8");
  } else {
    const existing = path.join(outDir, "source.able");
    await fs.rm(existing, { force: true });
  }

  if (fixture.manifest) {
    const manifestPath = path.join(outDir, "manifest.json");
    const entry = fixture.manifest.entry ?? "module.json";
    const setup = fixture.manifest.setup ?? (fixture.setupModules ? Object.keys(fixture.setupModules) : undefined);
    const manifest = { ...fixture.manifest, entry, ...(setup ? { setup } : {}) };
    await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2), "utf8");
  }
}

function stringify(value: unknown): string {
  return JSON.stringify(
    value,
    (_key, val) => (typeof val === "bigint" ? val.toString() : val),
    2,
  );
}

function normalizeModule(module: AST.Module): void {
  // no-op; method shorthand metadata must be set explicitly in fixtures.
}

const INDENT = "  ";

function moduleToSource(module: AST.Module): string {
  const lines: string[] = [];
  if (module.package) {
    lines.push(`package ${module.package.namePath.map(printIdentifier).join(".")}`);
    lines.push("");
  }
  if (module.imports && module.imports.length > 0) {
    for (const imp of module.imports) {
      lines.push(printImport(imp));
    }
    lines.push("");
  }
  for (const stmt of module.body) {
    lines.push(printStatement(stmt, 0));
  }
  return lines
    .map((line) => line.replace(/\s+$/g, ""))
    .join("\n")
    .replace(/\n{3,}/g, "\n\n")
    .trimEnd();
}

function printImport(imp: AST.ImportStatement): string {
  const path = imp.packagePath.map(printIdentifier).join(".");
  if (imp.isWildcard) {
    return `import ${path}.*`;
  }
  if (imp.selectors && imp.selectors.length > 0) {
    const selectors = imp.selectors
      .map((sel) => (sel.alias ? `${printIdentifier(sel.name)} as ${printIdentifier(sel.alias)}` : printIdentifier(sel.name)))
      .join(", ");
    return `import ${path}.{${selectors}}`;
  }
  if (imp.alias) {
    return `import ${path} as ${printIdentifier(imp.alias)}`;
  }
  return `import ${path}`;
}

function printDynImport(imp: AST.DynImportStatement, level: number): string {
  const path = imp.packagePath.map(printIdentifier).join(".");
  if (imp.isWildcard) {
    return `${indent(level)}dynimport ${path}.*`;
  }
  if (imp.selectors && imp.selectors.length > 0) {
    const selectors = imp.selectors
      .map((sel) => (sel.alias ? `${printIdentifier(sel.name)} as ${printIdentifier(sel.alias)}` : printIdentifier(sel.name)))
      .join(", ");
    return `${indent(level)}dynimport ${path}.{${selectors}}`;
  }
  if (imp.alias) {
    return `${indent(level)}dynimport ${path} as ${printIdentifier(imp.alias)}`;
  }
  return `${indent(level)}dynimport ${path}`;
}

function printStatement(stmt: AST.Statement, level: number): string {
  switch (stmt.type) {
    case "FunctionDefinition":
      return printFunctionDefinition(stmt, level);
    case "StructDefinition":
      return printStructDefinition(stmt, level);
    case "UnionDefinition":
      return printUnionDefinition(stmt, level);
    case "InterfaceDefinition":
      return printInterfaceDefinition(stmt, level);
    case "ImplementationDefinition":
      return printImplementationDefinition(stmt, level);
    case "MethodsDefinition":
      return printMethodsDefinition(stmt, level);
    case "ReturnStatement":
      return `${indent(level)}return${stmt.argument ? ` ${printExpression(stmt.argument, level)}` : ""}`;
    case "RaiseStatement":
      return `${indent(level)}raise ${printExpression(stmt.expression, level)}`;
    case "RethrowStatement":
      return `${indent(level)}rethrow`;
    case "BreakStatement": {
      const label = stmt.label ? ` '${printIdentifier(stmt.label)}` : "";
      const value = stmt.value ? ` ${printExpression(stmt.value, level)}` : "";
      return `${indent(level)}break${label}${value}`;
    }
    case "ContinueStatement":
      return `${indent(level)}continue`;
    case "WhileLoop":
      return `${indent(level)}while ${printExpression(stmt.condition, level)} ${printBlock(stmt.body, level)}`;
    case "ForLoop":
      return `${indent(level)}for ${printPattern(stmt.pattern)} in ${printExpression(stmt.iterable, level)} ${printBlock(stmt.body, level)}`;
    case "YieldStatement":
      return `${indent(level)}yield${stmt.argument ? ` ${printExpression(stmt.argument, level)}` : ""}`;
    case "PreludeStatement":
      return `${indent(level)}prelude ${stmt.target} {\n${indent(level + 1)}${stmt.code}\n${indent(level)}}`;
    case "ExternFunctionBody":
      return printExternFunction(stmt, level);
    case "DynImportStatement":
      return printDynImport(stmt, level);
    default:
      if (isExpression(stmt)) {
        return `${indent(level)}${printExpression(stmt, level)}`;
      }
      return `${indent(level)}/* unsupported ${stmt.type} */`;
  }
}

function printFunctionDefinition(fn: AST.FunctionDefinition, level: number): string {
  let header = `${indent(level)}${fn.isPrivate ? "private " : ""}fn`;
  if (fn.isMethodShorthand) {
    header += " #";
  } else {
    header += " ";
  }
  header += printIdentifier(fn.id);
  if (fn.genericParams && fn.genericParams.length > 0) {
    header += `<${fn.genericParams.map(printGenericParameter).join(", " )}>`;
  }
  header += `(${fn.params.map(printFunctionParameter).join(", ")})`;
  if (fn.returnType) {
    header += ` -> ${printTypeExpression(fn.returnType)}`;
  }
  if (fn.whereClause && fn.whereClause.length > 0) {
    header += ` where ${fn.whereClause.map(printWhereClause).join(", ")}`;
  }
  return `${header} ${printBlock(fn.body, level)}`;
}

function printStructDefinition(def: AST.StructDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("struct");
  header.push(printIdentifier(def.id));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  const prefix = `${indent(level)}${header.join(" ")}`;
  const whereSuffix = def.whereClause && def.whereClause.length > 0 ? ` where ${def.whereClause.map(printWhereClause).join(", ")}` : "";
  if (def.kind === "positional") {
    const types = (def.fields ?? []).map((field) => printTypeExpression(field.fieldType)).join(", ");
    return `${prefix}${whereSuffix}(${types})`;
  }
  if (def.kind === "named") {
    const fieldList = def.fields ?? [];
    const fields = fieldList.map((field, index) => {
      const suffix = index === fieldList.length - 1 ? "" : ",";
      return `${indent(level + 1)}${field.isPrivate ? "private " : ""}${printIdentifier(field.name!)}: ${printTypeExpression(field.fieldType)}${suffix}`;
    });
    return `${prefix}${whereSuffix} {\n${fields.join("\n")}\n${indent(level)}}`;
  }
  return `${prefix}${whereSuffix} {}`;
}

function printUnionDefinition(def: AST.UnionDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("union");
  header.push(printIdentifier(def.id));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  const suffix = def.whereClause && def.whereClause.length > 0 ? ` where ${def.whereClause.map(printWhereClause).join(", ")}` : "";
  const variants = def.variants && def.variants.length > 0 ? ` = ${def.variants.map(printTypeExpression).join(" | ")}` : "";
  return `${indent(level)}${header.join(" ")}${suffix}${variants}`;
}

function printInterfaceDefinition(def: AST.InterfaceDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("interface");
  header.push(printIdentifier(def.id));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  if (def.whereClause && def.whereClause.length > 0) {
    header.push(`where ${def.whereClause.map(printWhereClause).join(", ")}`);
  }
  if (def.baseInterfaces && def.baseInterfaces.length > 0) {
    header.push(`= ${def.baseInterfaces.map(printTypeExpression).join(" + ")}`);
  }
  const lines = [`${indent(level)}${header.join(" ")}`];
  if (def.signatures && def.signatures.length > 0) {
    lines.push(`${indent(level)}{`);
    for (const sig of def.signatures) {
      lines.push(`${indent(level + 1)}${printFunctionSignature(sig)}`);
    }
    lines.push(`${indent(level)}}`);
  }
  return lines.join("\n");
}

function printImplementationDefinition(def: AST.ImplementationDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("impl");
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  if (def.interfaceName) {
    header.push(printIdentifier(def.interfaceName));
    if (def.interfaceArgs && def.interfaceArgs.length > 0) {
      header.push(def.interfaceArgs.map(printTypeExpression).join(" "));
    }
  }
  if (def.targetType) {
    header.push("for");
    header.push(printTypeExpression(def.targetType));
  }
  if (def.whereClause && def.whereClause.length > 0) {
    header.push(`where ${def.whereClause.map(printWhereClause).join(", ")}`);
  }
  const lines = [`${indent(level)}${header.join(" ")}`];
  if (def.definitions && def.definitions.length > 0) {
    lines.push(`${indent(level)}{`);
    for (const inner of def.definitions) {
      lines.push(printFunctionDefinition(inner, level + 1));
    }
    lines.push(`${indent(level)}}`);
  }
  return lines.join("\n");
}

function printMethodsDefinition(def: AST.MethodsDefinition, level: number): string {
  const header: string[] = [];
  header.push("methods");
  header.push(printTypeExpression(def.targetType));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  if (def.whereClause && def.whereClause.length > 0) {
    header.push(`where ${def.whereClause.map(printWhereClause).join(", ")}`);
  }
  const lines = [`${indent(level)}${header.join(" ")} {`];
  if (def.definitions) {
    for (const inner of def.definitions) {
      lines.push(printFunctionDefinition(inner, level + 1));
    }
  }
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printExternFunction(externFn: AST.ExternFunctionBody, level: number): string {
  const signature = printFunctionDefinition(externFn.signature, level);
  const header = `${indent(level)}extern ${externFn.target} ${signature.trimStart()}`;
  const body = externFn.body.split("\n").map((line) => `${indent(level + 1)}${line}`).join("\n");
  return `${header} {\n${body}\n${indent(level)}}`;
}

function printExpression(expr: AST.Expression, level: number): string {
  switch (expr.type) {
    case "StringLiteral":
      return `"${expr.value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
    case "IntegerLiteral":
      return expr.integerType ? `${expr.value.toString()}_${expr.integerType}` : expr.value.toString();
    case "FloatLiteral":
      return expr.floatType ? `${expr.value}_${expr.floatType}` : expr.value.toString();
    case "BooleanLiteral":
      return String(expr.value);
    case "NilLiteral":
      return "nil";
    case "CharLiteral":
      return `'${expr.value}'`;
    case "Identifier":
      return printIdentifier(expr);
    case "ArrayLiteral":
      return `[${expr.elements.map((el) => printExpression(el, level)).join(", ")}]`;
    case "AssignmentExpression":
      if (expr.right.type === "MatchExpression") {
        return `${printAssignmentLeft(expr.left)} ${expr.operator} (${printMatchExpression(expr.right, level)})`;
      }
      return `${printAssignmentLeft(expr.left)} ${expr.operator} ${printExpression(expr.right, level)}`;
    case "BinaryExpression":
      return `${printBinaryOperand(expr.left, expr.operator, "left", level)} ${expr.operator} ${printBinaryOperand(expr.right, expr.operator, "right", level)}`;
    case "UnaryExpression":
      return `${expr.operator}${printExpression(expr.operand, level)}`;
    case "FunctionCall":
      return printFunctionCall(expr, level);
    case "BlockExpression":
      return `do ${printBlock(expr, level)}`;
    case "LambdaExpression":
      return printLambda(expr, level);
    case "MemberAccessExpression":
      return `${printExpression(expr.object, level)}.${printMember(expr.member)}`;
    case "ImplicitMemberExpression":
      return `.${printIdentifier(expr.member)}`;
    case "IndexExpression":
      return `${printExpression(expr.object, level)}[${printExpression(expr.index, level)}]`;
    case "RangeExpression":
      return `${printExpression(expr.start, level)} ${expr.inclusive ? "..." : ".."} ${printExpression(expr.end, level)}`;
    case "ProcExpression":
      return expr.expression.type === "BlockExpression"
        ? `proc ${printBlock(expr.expression, level)}`
        : `proc ${printExpression(expr.expression, level)}`;
    case "SpawnExpression":
      return expr.expression.type === "BlockExpression"
        ? `spawn ${printBlock(expr.expression, level)}`
        : `spawn ${printExpression(expr.expression, level)}`;
    case "StructLiteral":
      return printStructLiteral(expr, level);
    case "IfExpression":
      return printIfExpression(expr, level);
    case "MatchExpression":
      return printMatchExpression(expr, level);
    case "PropagationExpression":
      return `${printExpression(expr.expression, level)}!`;
    case "OrElseExpression":
      return `${printExpression(expr.expression, level)} else ${printHandlingBlock(expr.handler, expr.errorBinding, level)}`;
    case "EnsureExpression":
      return `${printExpression(expr.tryExpression, level)} ensure ${printBlock(expr.ensureBlock, level)}`;
    case "RescueExpression":
      return `${printExpression(expr.monitoredExpression, level)} rescue ${printRescueBlock(expr.clauses, level)}`;
    case "IteratorLiteral":
      return printIteratorLiteral(expr, level);
    case "TopicReferenceExpression":
      return "%";
    case "PlaceholderExpression":
      return expr.index ? `@${expr.index}` : "@";
    case "BreakpointExpression":
      return `breakpoint '${printIdentifier(expr.label)} ${printBlock(expr.body, level)}`;
    case "StringInterpolation":
      return printStringInterpolation(expr, level);
    default:
      return "/* expression */";
  }
}

function printStructLiteral(lit: AST.StructLiteral, level: number): string {
  const base = lit.structType ? printIdentifier(lit.structType) : "";
  if (lit.isPositional) {
    const values = lit.fields.map((field) => printExpression(field.value!, level)).join(", ");
    return `${base}(${values})`;
  }
  const fields = lit.fields.map((field) => {
    if (field.isShorthand && field.name) {
      return printIdentifier(field.name);
    }
    if (field.name) {
      return `${printIdentifier(field.name)}: ${printExpression(field.value!, level)}`;
    }
    return printExpression(field.value!, level);
  });
  const spreads =
    lit.functionalUpdateSources && lit.functionalUpdateSources.length > 0
      ? lit.functionalUpdateSources.map((src) => `..${printExpression(src, level)}`)
      : [];
  const items = [...spreads, ...fields].join(", ");
  return `${base} { ${items} }`;
}

function printIteratorLiteral(lit: AST.IteratorLiteral, level: number): string {
  const block: AST.BlockExpression = {
    type: "BlockExpression",
    body: lit.body ?? [],
  };
  return `Iterator ${printBlock(block, level)}`;
}

function printHandlingBlock(block: AST.BlockExpression, binding: AST.Identifier | undefined, level: number): string {
  const lines = ["{"];
  if (binding) {
    lines.push(`${indent(level + 1)}| ${printIdentifier(binding)} |`);
  }
  for (const stmt of block.body ?? []) {
    lines.push(printStatement(stmt, level + 1));
  }
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printRescueBlock(clauses: AST.MatchClause[], level: number): string {
  const lines = ["{"];
  clauses.forEach((clause, index) => {
    const suffix = index === clauses.length - 1 ? "" : ",";
    lines.push(`${indent(level + 1)}${printMatchClause(clause, level + 1)}${suffix}`);
  });
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printStringInterpolation(interp: AST.StringInterpolation, level: number): string {
  const parts = interp.parts
    .map((part) => {
      if (typeof part === "string") {
        return part.replace(/`/g, "\\`").replace(/\$/g, "\\$");
      }
      if (part.type === "StringLiteral") {
        return part.value.replace(/`/g, "\\`").replace(/\$/g, "\\$");
      }
      return `${"${"}${printExpression(part as AST.Expression, level)}${"}"}`;
    })
    .join("");
  return `\`${parts}\``;
}

function printFunctionCall(call: AST.FunctionCall, level: number): string {
  const callee = printExpression(call.callee, level);
  const typeArgs = call.typeArguments && call.typeArguments.length > 0 ? `<${call.typeArguments.map(printTypeExpression).join(", ")}>` : "";
  if (call.isTrailingLambda && call.arguments.length > 0) {
    const trailing = call.arguments[call.arguments.length - 1];
    if (trailing.type === "LambdaExpression") {
      const precedingArgs = call.arguments.slice(0, -1).map((arg) => printExpression(arg, level)).join(", ");
      const callPart = `${callee}${typeArgs}(${precedingArgs})`.replace(/\(\)/, "");
      const spacer = callPart.length > 0 ? " " : "";
      return `${callPart}${spacer}${printLambda(trailing, level)}`.trim();
    }
  }
  const args = call.arguments.map((arg) => printExpression(arg, level)).join(", ");
  return `${callee}${typeArgs}(${args})`;
}

function printLambda(lambda: AST.LambdaExpression, level: number): string {
  const params = lambda.params.map((param) => printPattern(param.name)).join(", ");
  let result = "{";
  if (params.length > 0) {
    result += ` ${params}`;
  }
  if (lambda.returnType) {
    result += ` -> ${printTypeExpression(lambda.returnType)}`;
  }
  const bodyExpr = lambda.body.type === "BlockExpression"
    ? printBlock(lambda.body, level)
    : printExpression(lambda.body, level);
  result += ` => ${bodyExpr}`;
  if (!bodyExpr.endsWith("}")) {
    result += "";
  }
  result += "}";
  return result;
}

function printBinaryOperand(expr: AST.Expression, parentOperator: string, side: "left" | "right", level: number): string {
  const rendered = printExpression(expr, level);
  if (expr.type !== "BinaryExpression") {
    return rendered;
  }
  const parentPrecedence = getBinaryPrecedence(parentOperator);
  const childPrecedence = getBinaryPrecedence(expr.operator);
  if (parentPrecedence === -1 || childPrecedence === -1) {
    return rendered;
  }
  if (side === "left") {
    if (childPrecedence < parentPrecedence || (childPrecedence === parentPrecedence && isRightAssociative(parentOperator))) {
      return `(${rendered})`;
    }
  } else {
    if (childPrecedence < parentPrecedence || (childPrecedence === parentPrecedence && !isRightAssociative(parentOperator))) {
      return `(${rendered})`;
    }
  }
  return rendered;
}

function getBinaryPrecedence(operator: string): number {
  switch (operator) {
    case "||":
      return 1;
    case "&&":
      return 2;
    case "|":
      return 3;
    case "\\xor":
      return 4;
    case "&":
      return 5;
    case "==":
    case "!=":
      return 6;
    case ">":
    case "<":
    case ">=":
    case "<=":
      return 7;
    case "<<":
    case ">>":
      return 8;
    case "+":
    case "-":
      return 9;
    case "*":
    case "/":
    case "%":
      return 10;
    case "**":
      return 11;
    default:
      return -1;
  }
}

function isRightAssociative(operator: string): boolean {
  return operator === "**";
}

function printIfExpression(expr: AST.IfExpression, level: number): string {
  const parts: string[] = [];
  parts.push(`if ${printExpression(expr.ifCondition, level)} ${printBlock(expr.ifBody, level)}`);
  for (const clause of expr.orClauses ?? []) {
    if (clause.condition) {
      parts.push(`or ${printExpression(clause.condition, level)} ${printBlock(clause.body, level)}`);
    } else {
      parts.push(`else ${printBlock(clause.body, level)}`);
    }
  }
  return parts.join("\n");
}

function printMatchExpression(expr: AST.MatchExpression, level: number): string {
  const subject = printExpression(expr.subject, level);
  const lines = [`${subject} match {`];
  const clauses = expr.clauses ?? [];
  clauses.forEach((clause, index) => {
    const suffix = index === clauses.length - 1 ? "" : ",";
    lines.push(`${indent(level + 1)}${printMatchClause(clause, level + 1)}${suffix}`);
  });
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printMatchClause(clause: AST.MatchClause, level: number): string {
  const pattern = printPattern(clause.pattern);
  const guard = clause.guard ? ` if ${printExpression(clause.guard, level)}` : "";
  const body = clause.body.type === "BlockExpression" ? printBlock(clause.body, level).trim() : printExpression(clause.body, level);
  return `case ${pattern}${guard} => ${body}`;
}

function printBlock(block: AST.BlockExpression, level: number): string {
  const lines = ["{"];
  for (const stmt of block.body ?? []) {
    lines.push(printStatement(stmt, level + 1));
  }
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printAssignmentLeft(left: AST.Pattern | AST.MemberAccessExpression | AST.IndexExpression): string {
  if (left.type === "MemberAccessExpression" || left.type === "IndexExpression") {
    return printExpression(left, 0);
  }
  return printPattern(left);
}

function printPattern(pattern: AST.Pattern): string {
  switch (pattern.type) {
    case "Identifier":
      return printIdentifier(pattern);
    case "WildcardPattern":
      return "_";
    case "LiteralPattern":
      return printExpression(pattern.literal, 0);
    case "StructPattern":
      if (pattern.isPositional) {
        const fields = pattern.fields.map((field) => printPattern(field.pattern)).join(", ");
        const prefix = pattern.structType ? printIdentifier(pattern.structType) : "";
        return prefix ? `${prefix} (${fields})` : `(${fields})`;
      }
      const namedFields = pattern.fields.map((field) => {
        if (field.fieldName) {
          return `${printIdentifier(field.fieldName)}: ${printPattern(field.pattern)}`;
        }
        return printPattern(field.pattern);
      });
      return `${pattern.structType ? `${printIdentifier(pattern.structType)} ` : ""}{ ${namedFields.join(", ")} }`;
    case "ArrayPattern":
      const elements = pattern.elements.map(printPattern).join(", ");
      const rest = pattern.restPattern ? `, ...${printPattern(pattern.restPattern)}` : "";
      return `[${elements}${rest}]`;
    case "TypedPattern":
      return `${printPattern(pattern.pattern)}: ${printTypeExpression(pattern.typeAnnotation)}`;
    default:
      return "_";
  }
}

function printFunctionParameter(param: AST.FunctionParameter): string {
  if (param.paramType) {
    return `${printPattern(param.name)}: ${printTypeExpression(param.paramType)}`;
  }
  return printPattern(param.name);
}

function printGenericParameter(param: AST.GenericParameter): string {
  if (param.constraints && param.constraints.length > 0) {
    return `${printIdentifier(param.name)}: ${param.constraints.map((c) => printTypeExpression(c.interfaceType)).join(" + ")}`;
  }
  return printIdentifier(param.name);
}

function printWhereClause(clause: AST.WhereClauseConstraint): string {
  return `${printIdentifier(clause.typeParam)}: ${clause.constraints.map((c) => printTypeExpression(c.interfaceType)).join(" + ")}`;
}

function printTypeExpression(typeExpr: AST.TypeExpression): string {
  switch (typeExpr.type) {
    case "SimpleTypeExpression":
      return printIdentifier(typeExpr.name);
    case "GenericTypeExpression":
      return `${printTypeExpression(typeExpr.base)} ${typeExpr.arguments.map(printTypeExpression).join(" ")}`;
    case "FunctionTypeExpression":
      return `(${typeExpr.paramTypes.map(printTypeExpression).join(", ")}) -> ${printTypeExpression(typeExpr.returnType)}`;
    case "NullableTypeExpression":
      return `?${printTypeExpression(typeExpr.innerType)}`;
    case "ResultTypeExpression":
      return `!${printTypeExpression(typeExpr.innerType)}`;
    case "UnionTypeExpression":
      return typeExpr.members.map(printTypeExpression).join(" | ");
    case "WildcardTypeExpression":
      return "_";
    default:
      return "/* type */";
  }
}

function printFunctionSignature(sig: AST.FunctionSignature): string {
  const parts: string[] = [];
  parts.push("fn");
  parts.push(printIdentifier(sig.name));
  if (sig.genericParams && sig.genericParams.length > 0) {
    parts.push(`<${sig.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  parts.push(`(${sig.params.map(printFunctionParameter).join(", ")})`);
  if (sig.returnType) {
    parts.push(`-> ${printTypeExpression(sig.returnType)}`);
  }
  if (sig.whereClause && sig.whereClause.length > 0) {
    parts.push(`where ${sig.whereClause.map(printWhereClause).join(", ")}`);
  }
  if (sig.defaultImpl) {
    parts.push(printBlock(sig.defaultImpl, 0));
  }
  return parts.join(" ");
}

function printIdentifier(id: AST.Identifier | string | undefined): string {
  if (!id) return "";
  if (typeof id === "string") return id;
  return id.name;
}

function printMember(member: AST.Identifier | AST.IntegerLiteral): string {
  return member.type === "Identifier" ? printIdentifier(member) : member.value.toString();
}

function indent(level: number): string {
  return INDENT.repeat(level);
}

function isExpression(node: AST.Statement): node is AST.Expression {
  return (node as AST.Expression).type !== undefined;
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
