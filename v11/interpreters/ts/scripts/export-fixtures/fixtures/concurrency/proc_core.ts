import { AST } from "../../../context";
import type { Fixture } from "../../../types";

export const procCoreFixtures: Fixture[] = [
] = [
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
            result: { kind: "i32", value: 1n },
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
            result: { kind: "String", value: "Pending:Resolved:Resolved:AB:done" },
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
            result: { kind: "String", value: "A1B2" },
          },
        },
      },
];

export default procCoreFixtures;
