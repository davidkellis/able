import { AST } from "../../../context";
import type { Fixture } from "../../../types";

export const futureSchedulingPart2: Fixture[] = [
  {
      name: "concurrency/future_error_cause",
      module: AST.module([
        AST.functionDefinition(
          "describe_error",
          [AST.functionParameter("err", AST.simpleTypeExpression("Error"))],
          AST.blockExpression([
            AST.assignmentExpression(
              ":=",
              AST.identifier("cause"),
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("err"), "cause"),
                [],
              ),
            ),
            AST.assignmentExpression(
              ":=",
              AST.identifier("flag"),
              AST.ifExpression(
                AST.binaryExpression(
                  "==",
                  AST.identifier("cause"),
                  AST.nilLiteral(),
                ),
                AST.blockExpression([AST.stringLiteral("no-error")]),
                [],
                AST.blockExpression([AST.stringLiteral("has-cause")]),
              ),
            ),
            AST.stringInterpolation([
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("err"), "message"),
                [],
              ),
              AST.stringLiteral("|"),
              AST.identifier("flag"),
            ]),
          ]),
          AST.simpleTypeExpression("String"),
        ),
        AST.assignmentExpression(
          ":=",
          AST.identifier("future"),
          AST.spawnExpression(
            AST.blockExpression([
              AST.raiseStatement(AST.stringLiteral("boom")),
            ]),
          ),
        ),
        AST.functionCall(AST.identifier("future_flush"), []),
        AST.assignmentExpression(
          ":=",
          AST.identifier("summary"),
          AST.orElseExpression(
            AST.propagationExpression(
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("future"), "value"),
                [],
              ),
            ),
            AST.blockExpression([
              AST.functionCall(
                AST.identifier("describe_error"),
                [AST.identifier("err")],
              ),
            ]),
            "err",
          ),
        ),
        AST.identifier("summary"),
      ]),
      manifest: {
        description: "Future value failures expose FutureError causes to handlers",
        expect: {
          result: { kind: "String", value: "Future failed: boom|has-cause" },
        },
      },
    },

  {
      name: "concurrency/future_cancel_nested",
      module: AST.module([
        AST.functionDefinition(
          "status_name",
          [AST.functionParameter("status")],
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("status"),
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
              ],
            ),
          ]),
          AST.simpleTypeExpression("String"),
        ),
        AST.assign("iterations", AST.integerLiteral(2048)),
        AST.assign("future_started", AST.bool(false)),
        AST.assign("future_iteration", AST.integerLiteral(0)),
        AST.assign("future_completed", AST.bool(false)),
        AST.assign("outer_waited", AST.bool(false)),
        AST.assign("outer_summary", AST.stringLiteral("")),
        AST.assign(
          "future",
          AST.spawnExpression(
            AST.blockExpression([
              AST.assignmentExpression("=", AST.identifier("future_started"), AST.bool(true)),
              AST.whileLoop(
                AST.binaryExpression("<", AST.identifier("future_iteration"), AST.identifier("iterations")),
                AST.blockExpression([
                  AST.assignmentExpression(
                    "=",
                    AST.identifier("future_iteration"),
                    AST.binaryExpression("+", AST.identifier("future_iteration"), AST.integerLiteral(1)),
                  ),
                  AST.functionCall(AST.identifier("future_yield"), []),
                ]),
              ),
              AST.assignmentExpression("=", AST.identifier("future_completed"), AST.bool(true)),
              AST.stringLiteral("done"),
            ]),
          ),
        ),
        AST.functionCall(AST.identifier("future_flush"), []),
        AST.assign(
          "status_before",
          AST.functionCall(
            AST.identifier("status_name"),
            [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("future"), "status"),
                [],
              ),
            ],
          ),
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("future"), "cancel"),
          [],
        ),
        AST.functionCall(AST.identifier("future_flush"), []),
        AST.assign(
          "status_after",
          AST.functionCall(
            AST.identifier("status_name"),
            [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("future"), "status"),
                [],
              ),
            ],
          ),
        ),
        AST.assign(
          "outer",
          AST.spawnExpression(
            AST.blockExpression([
              AST.assignmentExpression("=", AST.identifier("outer_waited"), AST.bool(true)),
              AST.assignmentExpression(
                ":=",
                AST.identifier("value"),
                AST.functionCall(
                  AST.memberAccessExpression(AST.identifier("future"), "value"),
                  [],
                ),
              ),
              AST.assignmentExpression(
                ":=",
                AST.identifier("summary"),
                AST.matchExpression(
                  AST.identifier("value"),
                  [
                    AST.matchClause(
                      AST.typedPattern(
                        AST.identifier("err"),
                        AST.simpleTypeExpression("Error"),
                      ),
                      AST.functionCall(
                        AST.memberAccessExpression(AST.identifier("err"), "message"),
                        [],
                      ),
                    ),
                    AST.matchClause(
                      AST.identifier("other"),
                      AST.stringInterpolation([
                        AST.stringLiteral("ok:"),
                        AST.identifier("other"),
                      ]),
                    ),
                  ],
                ),
              ),
              AST.assignmentExpression("=", AST.identifier("outer_summary"), AST.identifier("summary")),
              AST.identifier("summary"),
            ]),
          ),
        ),
        AST.assign(
          "outer_result",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("outer"), "value"),
            [],
          ),
        ),
        AST.arrayLiteral([
          AST.identifier("future_started"),
          AST.binaryExpression(">", AST.identifier("future_iteration"), AST.integerLiteral(0)),
          AST.unaryExpression("!", AST.identifier("future_completed")),
          AST.binaryExpression("==", AST.identifier("status_before"), AST.stringLiteral("Pending")),
          AST.binaryExpression("==", AST.identifier("status_after"), AST.stringLiteral("Cancelled")),
          AST.binaryExpression("==", AST.identifier("outer_result"), AST.stringLiteral("Future cancelled")),
          AST.binaryExpression("==", AST.identifier("outer_summary"), AST.stringLiteral("Future cancelled")),
          AST.identifier("outer_waited"),
        ]),
      ]),
      manifest: {
        description: "Awaiting a spawned future observes cancellation results and statuses",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
            ],
          },
        },
      },
    },

  {
      name: "concurrency/future_time_slicing",
      module: AST.module([
        AST.assign("iterations", AST.integerLiteral(4096)),
        AST.assign("counter", AST.integerLiteral(0)),
        AST.functionDefinition(
          "status_name",
          [AST.functionParameter("status")],
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("status"),
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
              ],
            ),
          ]),
          AST.simpleTypeExpression("String"),
        ),
        AST.assign(
          "handle",
          AST.spawnExpression(
            AST.blockExpression([
              AST.assignmentExpression(":=", AST.identifier("i"), AST.integerLiteral(0)),
              AST.whileLoop(
                AST.binaryExpression("<", AST.identifier("i"), AST.identifier("iterations")),
                AST.blockExpression([
                  AST.assignmentExpression("=", AST.identifier("counter"), AST.identifier("i")),
                  AST.assignmentExpression(
                    "=",
                    AST.identifier("i"),
                    AST.binaryExpression("+", AST.identifier("i"), AST.integerLiteral(1)),
                  ),
                ]),
              ),
              AST.integerLiteral(123),
            ]),
          ),
        ),
        AST.assign(
          "first_status",
          AST.functionCall(
            AST.identifier("status_name"),
            [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("handle"), "status"),
                [],
              ),
            ],
          ),
        ),
        AST.assign(
          "pending_observed",
          AST.binaryExpression("==", AST.identifier("first_status"), AST.stringLiteral("Pending")),
        ),
        AST.assign("flushes", AST.integerLiteral(0)),
        AST.assign("current_status", AST.identifier("first_status")),
        AST.whileLoop(
          AST.binaryExpression(
            "&&",
            AST.binaryExpression("==", AST.identifier("current_status"), AST.stringLiteral("Pending")),
            AST.binaryExpression("<", AST.identifier("flushes"), AST.integerLiteral(12)),
          ),
          AST.blockExpression([
            AST.functionCall(AST.identifier("future_flush"), []),
            AST.assignmentExpression("+=", AST.identifier("flushes"), AST.integerLiteral(1)),
            AST.assignmentExpression(
              "=",
              AST.identifier("current_status"),
              AST.functionCall(
                AST.identifier("status_name"),
                [
                  AST.functionCall(
                    AST.memberAccessExpression(AST.identifier("handle"), "status"),
                    [],
                  ),
                ],
              ),
            ),
            AST.assignmentExpression(
              "=",
              AST.identifier("pending_observed"),
              AST.binaryExpression(
                "||",
                AST.identifier("pending_observed"),
                AST.binaryExpression("==", AST.identifier("current_status"), AST.stringLiteral("Pending")),
              ),
            ),
          ]),
        ),
        AST.assign("final_status", AST.identifier("current_status")),
        AST.assign(
          "value_result",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("handle"), "value"),
            [],
          ),
        ),
        AST.arrayLiteral([
          AST.binaryExpression(
            "==",
            AST.identifier("counter"),
            AST.binaryExpression("-", AST.identifier("iterations"), AST.integerLiteral(1)),
          ),
          AST.binaryExpression("==", AST.identifier("final_status"), AST.stringLiteral("Resolved")),
          AST.binaryExpression("==", AST.identifier("value_result"), AST.integerLiteral(123)),
          AST.binaryExpression(
            "==",
            AST.identifier("pending_observed"),
            AST.binaryExpression(">", AST.identifier("flushes"), AST.integerLiteral(0)),
          ),
        ]),
      ]),
      manifest: {
        description: "Long-running future without explicit yields still resolves via automatic time slicing",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
            ],
          },
        },
      },
    },
  {
      name: "concurrency/future_time_slicing",
      module: AST.module([
        AST.assign("iterations", AST.integerLiteral(4096)),
        AST.assign("counter", AST.integerLiteral(0)),
        AST.functionDefinition(
          "status_name",
          [AST.functionParameter("status")],
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("status"),
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
              ],
            ),
          ]),
          AST.simpleTypeExpression("String"),
        ),
        AST.assign(
          "future",
          AST.spawnExpression(
            AST.blockExpression([
              AST.assignmentExpression(":=", AST.identifier("i"), AST.integerLiteral(0)),
              AST.whileLoop(
                AST.binaryExpression("<", AST.identifier("i"), AST.identifier("iterations")),
                AST.blockExpression([
                  AST.assignmentExpression("=", AST.identifier("counter"), AST.identifier("i")),
                  AST.assignmentExpression(
                    "=",
                    AST.identifier("i"),
                    AST.binaryExpression("+", AST.identifier("i"), AST.integerLiteral(1)),
                  ),
                ]),
              ),
              AST.integerLiteral(123),
            ]),
          ),
        ),
        AST.assign(
          "first_status",
          AST.functionCall(
            AST.identifier("status_name"),
            [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("future"), "status"),
                [],
              ),
            ],
          ),
        ),
        AST.assign(
          "pending_observed",
          AST.binaryExpression("==", AST.identifier("first_status"), AST.stringLiteral("Pending")),
        ),
        AST.assign("flushes", AST.integerLiteral(0)),
        AST.assign("current_status", AST.identifier("first_status")),
        AST.whileLoop(
          AST.binaryExpression(
            "&&",
            AST.binaryExpression("==", AST.identifier("current_status"), AST.stringLiteral("Pending")),
            AST.binaryExpression("<", AST.identifier("flushes"), AST.integerLiteral(12)),
          ),
          AST.blockExpression([
            AST.functionCall(AST.identifier("future_flush"), []),
            AST.assignmentExpression("+=", AST.identifier("flushes"), AST.integerLiteral(1)),
            AST.assignmentExpression(
              "=",
              AST.identifier("current_status"),
              AST.functionCall(
                AST.identifier("status_name"),
                [
                  AST.functionCall(
                    AST.memberAccessExpression(AST.identifier("future"), "status"),
                    [],
                  ),
                ],
              ),
            ),
            AST.assignmentExpression(
              "=",
              AST.identifier("pending_observed"),
              AST.binaryExpression(
                "||",
                AST.identifier("pending_observed"),
                AST.binaryExpression("==", AST.identifier("current_status"), AST.stringLiteral("Pending")),
              ),
            ),
          ]),
        ),
        AST.assign("final_status", AST.identifier("current_status")),
        AST.assign(
          "value_result",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("future"), "value"),
            [],
          ),
        ),
        AST.arrayLiteral([
          AST.binaryExpression(
            "==",
            AST.identifier("counter"),
            AST.binaryExpression("-", AST.identifier("iterations"), AST.integerLiteral(1)),
          ),
          AST.binaryExpression("==", AST.identifier("final_status"), AST.stringLiteral("Resolved")),
          AST.binaryExpression("==", AST.identifier("value_result"), AST.integerLiteral(123)),
          AST.binaryExpression(
            "==",
            AST.identifier("pending_observed"),
            AST.binaryExpression(">", AST.identifier("flushes"), AST.integerLiteral(0)),
          ),
        ]),
      ]),
      manifest: {
        description: "Long-running future without explicit yields still resolves via automatic time slicing",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
              { kind: "bool", value: true },
            ],
          },
        },
      },
    },
];

export default futureSchedulingPart2;
