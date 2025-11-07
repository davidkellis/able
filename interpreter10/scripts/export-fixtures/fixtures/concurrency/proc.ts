import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const procConcurrencyFixtures: Fixture[] = [
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
        name: "concurrency/proc_flush_fairness",
        module: AST.module([
          AST.assign("trace", AST.stringLiteral("")),
          AST.assign("firstStage", AST.integerLiteral(0)),
          AST.assign("secondStage", AST.integerLiteral(0)),
          AST.assign(
            "first",
            AST.procExpression(
              AST.blockExpression([
                AST.ifExpression(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("firstStage"),
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
                      AST.identifier("firstStage"),
                      AST.integerLiteral(1),
                    ),
                    AST.functionCall(AST.identifier("proc_yield"), []),
                  ]),
                  [],
                ),
                AST.ifExpression(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("firstStage"),
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
                      AST.identifier("firstStage"),
                      AST.integerLiteral(2),
                    ),
                  ]),
                  [],
                ),
              ]),
            ),
          ),
          AST.assign(
            "second",
            AST.procExpression(
              AST.blockExpression([
                AST.ifExpression(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("secondStage"),
                    AST.integerLiteral(0),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("trace"),
                      AST.binaryExpression(
                        "+",
                        AST.identifier("trace"),
                        AST.stringLiteral("1"),
                      ),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("secondStage"),
                      AST.integerLiteral(1),
                    ),
                    AST.functionCall(AST.identifier("proc_yield"), []),
                  ]),
                  [],
                ),
                AST.ifExpression(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("secondStage"),
                    AST.integerLiteral(1),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("trace"),
                      AST.binaryExpression(
                        "+",
                        AST.identifier("trace"),
                        AST.stringLiteral("2"),
                      ),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("secondStage"),
                      AST.integerLiteral(2),
                    ),
                  ]),
                  [],
                ),
              ]),
            ),
          ),
          AST.functionCall(AST.identifier("proc_flush"), []),
          AST.identifier("trace"),
        ]),
        manifest: {
          description: "proc_flush drains the queue in creation order to keep scheduling fair",
          expect: {
            result: { kind: "string", value: "A1B2" },
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
];

export default procConcurrencyFixtures;
