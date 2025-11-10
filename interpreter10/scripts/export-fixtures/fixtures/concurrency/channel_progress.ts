import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const channelProgressFixtures: Fixture[] = [
  {
    name: "concurrency/channel_send_receive_progress",
    module: AST.module([
      AST.assign(
        "handle",
        AST.functionCall(AST.identifier("__able_channel_new"), [
          AST.integerLiteral(2),
        ]),
      ),
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
        AST.simpleTypeExpression("string"),
      ),
      AST.assign(
        "sender_a",
        AST.procExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              ":=",
              AST.identifier("sent"),
              AST.bool(false),
            ),
            AST.assignmentExpression(
              ":=",
              AST.identifier("attempts"),
              AST.integerLiteral(0),
            ),
            AST.whileLoop(
              AST.binaryExpression(
                "&&",
                AST.binaryExpression(
                  "==",
                  AST.identifier("sent"),
                  AST.bool(false),
                ),
                AST.binaryExpression(
                  "<",
                  AST.identifier("attempts"),
                  AST.integerLiteral(32),
                ),
              ),
              AST.blockExpression([
                AST.ifExpression(
                  AST.functionCall(AST.identifier("__able_channel_try_send"), [
                    AST.identifier("handle"),
                    AST.integerLiteral(10),
                  ]),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("sent"),
                      AST.bool(true),
                    ),
                  ]),
                  [
                    AST.orClause(
                      AST.blockExpression([
                        AST.assignmentExpression(
                          "+=",
                          AST.identifier("attempts"),
                          AST.integerLiteral(1),
                        ),
                        AST.functionCall(AST.identifier("proc_yield"), []),
                      ]),
                    ),
                  ],
                ),
              ]),
            ),
            AST.identifier("sent"),
          ]),
        ),
      ),
      AST.assign(
        "sender_b",
        AST.procExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              ":=",
              AST.identifier("sent"),
              AST.bool(false),
            ),
            AST.assignmentExpression(
              ":=",
              AST.identifier("attempts"),
              AST.integerLiteral(0),
            ),
            AST.whileLoop(
              AST.binaryExpression(
                "&&",
                AST.binaryExpression(
                  "==",
                  AST.identifier("sent"),
                  AST.bool(false),
                ),
                AST.binaryExpression(
                  "<",
                  AST.identifier("attempts"),
                  AST.integerLiteral(32),
                ),
              ),
              AST.blockExpression([
                AST.ifExpression(
                  AST.functionCall(AST.identifier("__able_channel_try_send"), [
                    AST.identifier("handle"),
                    AST.integerLiteral(20),
                  ]),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("sent"),
                      AST.bool(true),
                    ),
                  ]),
                  [
                    AST.orClause(
                      AST.blockExpression([
                        AST.assignmentExpression(
                          "+=",
                          AST.identifier("attempts"),
                          AST.integerLiteral(1),
                        ),
                        AST.functionCall(AST.identifier("proc_yield"), []),
                      ]),
                    ),
                  ],
                ),
              ]),
            ),
            AST.identifier("sent"),
          ]),
        ),
      ),
      AST.assign(
        "receiver",
        AST.procExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              ":=",
              AST.identifier("received"),
              AST.integerLiteral(0),
            ),
            AST.assignmentExpression(
              ":=",
              AST.identifier("attempts"),
              AST.integerLiteral(0),
            ),
            AST.assignmentExpression(
              ":=",
              AST.identifier("sum"),
              AST.integerLiteral(0),
            ),
            AST.whileLoop(
              AST.binaryExpression(
                "&&",
                AST.binaryExpression(
                  "<",
                  AST.identifier("received"),
                  AST.integerLiteral(2),
                ),
                AST.binaryExpression(
                  "<",
                  AST.identifier("attempts"),
                  AST.integerLiteral(64),
                ),
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  ":=",
                  AST.identifier("value"),
                  AST.functionCall(AST.identifier("__able_channel_try_receive"), [
                    AST.identifier("handle"),
                  ]),
                ),
                AST.ifExpression(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("value"),
                    AST.nilLiteral(),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "+=",
                      AST.identifier("attempts"),
                      AST.integerLiteral(1),
                    ),
                    AST.functionCall(AST.identifier("proc_yield"), []),
                  ]),
                  [
                    AST.orClause(
                      AST.blockExpression([
                        AST.assignmentExpression(
                          "+=",
                          AST.identifier("sum"),
                          AST.identifier("value"),
                        ),
                        AST.assignmentExpression(
                          "+=",
                          AST.identifier("received"),
                          AST.integerLiteral(1),
                        ),
                      ]),
                    ),
                  ],
                ),
              ]),
            ),
            AST.identifier("sum"),
          ]),
        ),
      ),
      AST.assign("sender_a_value", AST.bool(false)),
      AST.assign("sender_b_value", AST.bool(false)),
      AST.assign("receiver_sum", AST.integerLiteral(-1)),
      AST.assign(
        "status_sender_a",
        AST.functionCall(AST.identifier("status_name"), [
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("sender_a"), "status"),
            [],
          ),
        ]),
      ),
      AST.assign(
        "status_sender_b",
        AST.functionCall(AST.identifier("status_name"), [
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("sender_b"), "status"),
            [],
          ),
        ]),
      ),
      AST.assign(
        "status_receiver",
        AST.functionCall(AST.identifier("status_name"), [
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("receiver"), "status"),
            [],
          ),
        ]),
      ),
      AST.assign("flushes", AST.integerLiteral(0)),
      AST.whileLoop(
        AST.binaryExpression(
          "&&",
          AST.binaryExpression(
            "||",
            AST.binaryExpression(
              "==",
              AST.identifier("status_sender_a"),
              AST.stringLiteral("Pending"),
            ),
            AST.binaryExpression(
              "||",
              AST.binaryExpression(
                "==",
                AST.identifier("status_sender_b"),
                AST.stringLiteral("Pending"),
              ),
              AST.binaryExpression(
                "==",
                AST.identifier("status_receiver"),
                AST.stringLiteral("Pending"),
              ),
            ),
          ),
          AST.binaryExpression(
            "<",
            AST.identifier("flushes"),
            AST.integerLiteral(16),
          ),
        ),
        AST.blockExpression([
          AST.functionCall(AST.identifier("proc_flush"), []),
          AST.assignmentExpression(
            "+=",
            AST.identifier("flushes"),
            AST.integerLiteral(1),
          ),
          AST.assignmentExpression(
            "=",
            AST.identifier("status_sender_a"),
            AST.functionCall(AST.identifier("status_name"), [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("sender_a"), "status"),
                [],
              ),
            ]),
          ),
          AST.assignmentExpression(
            "=",
            AST.identifier("status_sender_b"),
            AST.functionCall(AST.identifier("status_name"), [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("sender_b"), "status"),
                [],
              ),
            ]),
          ),
          AST.assignmentExpression(
            "=",
            AST.identifier("status_receiver"),
            AST.functionCall(AST.identifier("status_name"), [
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("receiver"), "status"),
                [],
              ),
            ]),
          ),
        ]),
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("status_sender_a"),
          AST.stringLiteral("Resolved"),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("sender_a_value"),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("sender_a"), "value"),
              [],
            ),
          ),
        ]),
        [],
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("status_sender_b"),
          AST.stringLiteral("Resolved"),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("sender_b_value"),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("sender_b"), "value"),
              [],
            ),
          ),
        ]),
        [],
      ),
      AST.ifExpression(
        AST.binaryExpression(
          "==",
          AST.identifier("status_receiver"),
          AST.stringLiteral("Resolved"),
        ),
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("receiver_sum"),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("receiver"), "value"),
              [],
            ),
          ),
        ]),
        [],
      ),
      AST.arrayLiteral([
        AST.binaryExpression(
          "==",
          AST.identifier("status_sender_a"),
          AST.stringLiteral("Resolved"),
        ),
        AST.binaryExpression(
          "==",
          AST.identifier("status_sender_b"),
          AST.stringLiteral("Resolved"),
        ),
        AST.binaryExpression(
          "==",
          AST.identifier("status_receiver"),
          AST.stringLiteral("Resolved"),
        ),
        AST.binaryExpression(
          "==",
          AST.identifier("sender_a_value"),
          AST.bool(true),
        ),
        AST.binaryExpression(
          "==",
          AST.identifier("sender_b_value"),
          AST.bool(true),
        ),
        AST.binaryExpression(
          "==",
          AST.identifier("receiver_sum"),
          AST.integerLiteral(30),
        ),
        AST.binaryExpression(
          ">",
          AST.identifier("flushes"),
          AST.integerLiteral(0),
        ),
      ]),
    ]),
    manifest: {
      description:
        "Concurrent channel senders and receiver make progress under cooperative scheduling",
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
          ],
        },
      },
    },
  },
];

export default channelProgressFixtures;
