import { AST } from "../../../context";
import type { Fixture } from "../../../types";

export const procSchedulingPart1: Fixture[] = [
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
          AST.assign("stage_a", AST.integerLiteral(0)),
          AST.assign("stage_b", AST.integerLiteral(0)),
          AST.assign("worker_second_safe", AST.bool(false)),
          AST.assign("other_second_safe", AST.bool(false)),
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
                      AST.identifier("worker_second_safe"),
                      AST.binaryExpression(
                        ">=",
                        AST.identifier("stage_b"),
                        AST.integerLiteral(1),
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
                      AST.identifier("other_second_safe"),
                      AST.binaryExpression(
                        ">=",
                        AST.identifier("stage_a"),
                        AST.integerLiteral(1),
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
          AST.arrayLiteral([
            AST.binaryExpression(
              "==",
              AST.identifier("stage_a"),
              AST.integerLiteral(2),
            ),
            AST.binaryExpression(
              "==",
              AST.identifier("stage_b"),
              AST.integerLiteral(2),
            ),
            AST.identifier("worker_second_safe"),
            AST.identifier("other_second_safe"),
            AST.binaryExpression(
              "==",
              AST.identifier("status_worker"),
              AST.stringLiteral("Resolved"),
            ),
            AST.binaryExpression(
              "==",
              AST.identifier("status_other"),
              AST.stringLiteral("Resolved"),
            ),
          ]),
        ]),
        manifest: {
          description: "Yielding procs make progress without one jumping ahead of the other",
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
              ],
            },
          },
        },
      },

  {
        name: "concurrency/fairness_proc_future",
        module: AST.module([
          AST.assign("stage_proc", AST.integerLiteral(0)),
          AST.assign("stage_future", AST.integerLiteral(0)),
          AST.assign("worker_stage2_safe", AST.bool(false)),
          AST.assign("worker_stage3_safe", AST.bool(false)),
          AST.assign("future_stage2_safe", AST.bool(false)),
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
                      AST.identifier("worker_stage2_safe"),
                      AST.binaryExpression(
                        ">=",
                        AST.identifier("stage_future"),
                        AST.integerLiteral(1),
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
                      AST.identifier("worker_stage3_safe"),
                      AST.binaryExpression(
                        ">=",
                        AST.identifier("stage_future"),
                        AST.integerLiteral(2),
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
                          AST.identifier("future_stage2_safe"),
                          AST.binaryExpression(
                            ">=",
                            AST.identifier("stage_proc"),
                            AST.integerLiteral(2),
                          ),
                        ),
                        AST.integerLiteral(0),
                        AST.assignmentExpression(
                          "=",
                          AST.identifier("stage_future"),
                          AST.integerLiteral(2),
                        ),
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
          AST.arrayLiteral([
            AST.binaryExpression(
              "==",
              AST.identifier("stage_proc"),
              AST.integerLiteral(3),
            ),
            AST.binaryExpression(
              "==",
              AST.identifier("stage_future"),
              AST.integerLiteral(2),
            ),
            AST.identifier("worker_stage2_safe"),
            AST.identifier("future_stage2_safe"),
            AST.identifier("worker_stage3_safe"),
            AST.binaryExpression(
              "==",
              AST.identifier("worker_status"),
              AST.stringLiteral("Resolved"),
            ),
            AST.binaryExpression(
              "==",
              AST.identifier("future_status"),
              AST.stringLiteral("Resolved"),
            ),
            AST.binaryExpression(
              "==",
              AST.identifier("future_result"),
              AST.integerLiteral(0),
            ),
          ]),
        ]),
        manifest: {
          description: "Proc and future both advance between yields without overtaking one another",
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
      name: "concurrency/proc_error_cause",
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
                [
                  AST.orClause(
                    AST.blockExpression([AST.stringLiteral("has-cause")]),
                    AST.booleanLiteral(true),
                  ),
                ],
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
          AST.simpleTypeExpression("string"),
        ),
        AST.assignmentExpression(
          ":=",
          AST.identifier("handle"),
          AST.procExpression(
            AST.blockExpression([
              AST.raiseStatement(AST.stringLiteral("boom")),
            ]),
          ),
        ),
        AST.functionCall(AST.identifier("proc_flush"), []),
        AST.assignmentExpression(
          ":=",
          AST.identifier("summary"),
          AST.orElseExpression(
            AST.propagationExpression(
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("handle"), "value"),
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
        description: "Proc handle errors expose ProcError causes to handlers",
        expect: {
          result: { kind: "string", value: "Proc failed: boom|has-cause" },
        },
      },
    },
];

export default procSchedulingPart1;
