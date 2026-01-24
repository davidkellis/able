import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const mutexConcurrencyFixtures: Fixture[] = [
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
            result: { kind: "String", value: "AB" },
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
            AST.spawnExpression(
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
          AST.functionCall(AST.identifier("future_flush"), []),
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
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("worker"), "value"),
            [],
          ),
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
          description: "Mutex contention ensures the waiting task resumes only after unlock",
          expect: {
            result: { kind: "String", value: "Pending:Resolved:ABC" },
          },
        },
      },
];

export default mutexConcurrencyFixtures;
