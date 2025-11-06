import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const channelConcurrencyFixtures: Fixture[] = [
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
                AST.simpleTypeExpression("bool"),
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
                AST.simpleTypeExpression("bool"),
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
];

export default channelConcurrencyFixtures;
