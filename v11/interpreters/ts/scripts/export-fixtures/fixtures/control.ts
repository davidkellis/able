import { AST } from "../../context";
import type { Fixture } from "../../types";

const controlFixtures: Fixture[] = [
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
          result: { kind: "i32", value: 6n },
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
              { kind: "i32", value: 3n },
              { kind: "i32", value: 3n },
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
          AST.simpleTypeExpression("bool"),
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
          AST.simpleTypeExpression("bool"),
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
              { kind: "i32", value: 100n },
              { kind: "i32", value: 1n },
              { kind: "i32", value: 1n },
              { kind: "i32", value: 1n },
            ],
          },
        },
      },
    },

  {
    name: "control/loop_expression",
    module: AST.module([
      AST.assign("counter", AST.integerLiteral(0)),
      AST.assign(
        "result",
        AST.loopExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              "=",
              AST.identifier("counter"),
              AST.binaryExpression(
                "+",
                AST.identifier("counter"),
                AST.integerLiteral(1),
              ),
            ),
            AST.ifExpression(
              AST.binaryExpression(
                ">=",
                AST.identifier("counter"),
                AST.integerLiteral(3),
              ),
              AST.blockExpression([
                AST.breakStatement(undefined, AST.identifier("counter")),
              ]),
            ),
          ]),
        ),
      ),
      AST.identifier("result"),
    ]),
    manifest: {
      description: "Loop expression returns break payload",
      expect: {
        result: { kind: "i32", value: 3n },
      },
    },
  },

  {
      name: "control/iterator_annotation_binding",
      module: AST.module([
        AST.assign("sum", AST.integerLiteral(0)),
        AST.assign(
          "iter",
          AST.iteratorLiteral(
            [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("driver"), "yield"),
                [AST.integerLiteral(2)],
              ),
            ],
            "driver",
            AST.simpleTypeExpression("i32"),
          ),
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
        description: "Iterator literal preserves element type annotations and custom binding names",
        expect: {
          result: { kind: "i32", value: 2n },
        },
      },
    },

  {
      name: "control/iterator_lazy_next",
      module: AST.module([
        AST.assign("count", AST.integerLiteral(0)),
        AST.assign(
          "iter",
          AST.iteratorLiteral([
            AST.assignmentExpression(
              "=",
              AST.identifier("count"),
              AST.binaryExpression("+", AST.identifier("count"), AST.integerLiteral(1)),
            ),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("gen"), "yield"),
              [AST.identifier("count")],
            ),
            AST.assignmentExpression(
              "=",
              AST.identifier("count"),
              AST.binaryExpression("+", AST.identifier("count"), AST.integerLiteral(1)),
            ),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("gen"), "yield"),
              [AST.identifier("count")],
            ),
          ]),
        ),
        AST.assign("before", AST.identifier("count")),
        AST.assign("first", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("second", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("third", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("after", AST.identifier("count")),
        AST.arrayLiteral([
          AST.identifier("before"),
          AST.identifier("first"),
          AST.identifier("second"),
          AST.identifier("third"),
          AST.identifier("after"),
        ]),
      ]),
      manifest: {
        description: "Iterator.next drives generator lazily and returns IteratorEnd when exhausted",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 0n },
              { kind: "i32", value: 1n },
              { kind: "i32", value: 2n },
              { kind: "iterator_end" },
              { kind: "i32", value: 2n },
            ],
          },
        },
      },
    },

  {
      name: "control/iterator_for_body_next",
      module: AST.module([
        AST.assign(
          "iter",
          AST.iteratorLiteral([
            AST.forLoop(
              AST.identifier("item"),
              AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2), AST.integerLiteral(3)]),
              AST.blockExpression([
                AST.functionCall(
                  AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                  [AST.identifier("item")],
                ),
              ]),
            ),
          ]),
        ),
        AST.assign("first", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("second", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("third", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("done", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.arrayLiteral([
          AST.identifier("first"),
          AST.identifier("second"),
          AST.identifier("third"),
          AST.identifier("done"),
        ]),
      ]),
      manifest: {
        description: "For loop inside iterator literal resumes correctly when driving next() manually",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 1n },
              { kind: "i32", value: 2n },
              { kind: "i32", value: 3n },
              { kind: "iterator_end" },
            ],
          },
        },
      },
    },

  {
      name: "control/iterator_stop_next",
      module: AST.module([
        AST.assign(
          "iter",
          AST.iteratorLiteral([
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("gen"), "yield"),
              [AST.integerLiteral(1)],
            ),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("gen"), "stop"),
              [],
            ),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("gen"), "yield"),
              [AST.integerLiteral(99)],
            ),
          ]),
        ),
        AST.assign("first", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("second", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.assign("third", AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), [])),
        AST.arrayLiteral([
          AST.identifier("first"),
          AST.identifier("second"),
          AST.identifier("third"),
        ]),
      ]),
      manifest: {
        description: "gen.stop terminates iteration and memoizes IteratorEnd for subsequent next() calls",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 1n },
              { kind: "iterator_end" },
              { kind: "iterator_end" },
            ],
          },
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
          result: { kind: "i32", value: 3n },
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
          result: { kind: "i32", value: 6n },
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
          result: { kind: "i32", value: 4n },
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
          result: { kind: "i32", value: 3n },
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
          result: { kind: "i32", value: 15n },
        },
      },
    },
];

export default controlFixtures;
