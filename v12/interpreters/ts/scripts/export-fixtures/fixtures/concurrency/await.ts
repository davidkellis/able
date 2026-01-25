import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const call = (callee: string | any, args: any[] = []) =>
  AST.functionCall(typeof callee === "string" ? AST.identifier(callee) : callee, args);

const assign = (op: ":=" | "=", left: string | any, right: any) =>
  AST.assignmentExpression(op, typeof left === "string" ? AST.identifier(left) : left, right);

const awaitFixtures: Fixture[] = [
  {
    name: "concurrency/await_channel_arms",
    module: AST.module([
      assign(":=", "ch", call("__able_channel_new", [AST.integerLiteral(0)])),
      assign(":=", "receiver_result", AST.stringLiteral("pending")),
      assign(":=", "sender_result", AST.stringLiteral("pending")),
      assign(
        ":=",
        "receiver",
        AST.spawnExpression(
          AST.blockExpression([
            assign(
              "=",
              "receiver_result",
              AST.awaitExpression(
                AST.arrayLiteral([
                  call("__able_channel_await_try_recv", [
                    AST.identifier("ch"),
                    AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                  ]),
                ]),
              ),
            ),
            AST.identifier("receiver_result"),
          ]),
        ),
      ),
      assign(
        ":=",
        "sender",
        AST.spawnExpression(
          AST.blockExpression([
            assign(
              "=",
              "sender_result",
              AST.awaitExpression(
                AST.arrayLiteral([
                  call("__able_channel_await_try_send", [
                    AST.identifier("ch"),
                    AST.stringLiteral("payload"),
                    AST.lambdaExpression([], AST.stringLiteral("sent")),
                  ]),
                ]),
              ),
            ),
            AST.identifier("sender_result"),
          ]),
        ),
      ),
      call("future_flush", []),
      AST.arrayLiteral([
        AST.identifier("receiver_result"),
        AST.identifier("sender_result"),
        call(AST.memberAccessExpression(AST.identifier("receiver"), AST.identifier("value")), []),
        call(AST.memberAccessExpression(AST.identifier("sender"), AST.identifier("value")), []),
      ]),
    ]),
    manifest: {
      description: "await send/receive arms resolve via channel await helpers and return callback payloads",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "String", value: "payload" },
            { kind: "String", value: "sent" },
            { kind: "String", value: "payload" },
            { kind: "String", value: "sent" },
          ],
        },
      },
    },
  },
  {
    name: "concurrency/await_channel_fairness",
    module: AST.module([
      assign(":=", "ch1", call("__able_channel_new", [AST.integerLiteral(2)])),
      assign(":=", "ch2", call("__able_channel_new", [AST.integerLiteral(2)])),
      call("__able_channel_send", [AST.identifier("ch1"), AST.stringLiteral("A1")]),
      call("__able_channel_send", [AST.identifier("ch2"), AST.stringLiteral("B1")]),
      call("__able_channel_send", [AST.identifier("ch1"), AST.stringLiteral("A2")]),
      call("__able_channel_send", [AST.identifier("ch2"), AST.stringLiteral("B2")]),
      assign(":=", "trace", AST.stringLiteral("")),
      assign(
        ":=",
        "runner",
        AST.spawnExpression(
          AST.blockExpression([
            assign(
              "=",
              "trace",
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch1"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch2"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                  ]),
                ),
              ),
            ),
            assign(
              "=",
              "trace",
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch1"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch2"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                  ]),
                ),
              ),
            ),
            assign(
              "=",
              "trace",
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch1"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch2"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                  ]),
                ),
              ),
            ),
            assign(
              "=",
              "trace",
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch1"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                    call("__able_channel_await_try_recv", [
                      AST.identifier("ch2"),
                      AST.lambdaExpression([AST.functionParameter("v")], AST.identifier("v")),
                    ]),
                  ]),
                ),
              ),
            ),
            AST.identifier("trace"),
          ]),
        ),
      ),
      call(AST.memberAccessExpression(AST.identifier("runner"), AST.identifier("value")), []),
    ]),
    manifest: {
      description: "await rotates ready channel arms round-robin instead of preferring the first entry",
      expect: {
        result: { kind: "String", value: "A1B1A2B2" },
      },
    },
  },
  {
    name: "concurrency/await_timer_default",
    module: AST.module(
      [
        assign(
          ":=",
          "runner",
          AST.spawnExpression(
            AST.blockExpression([
              assign(
                ":=",
                "first",
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call(AST.memberAccessExpression(AST.identifier("Await"), "default"), [
                      AST.lambdaExpression([], AST.stringLiteral("fallback")),
                    ]),
                  ]),
                ),
              ),
              assign(
                ":=",
                "second",
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call(AST.memberAccessExpression(AST.identifier("Await"), "sleep"), [
                      call(AST.memberAccessExpression(AST.integerLiteral(0), "seconds"), []),
                      AST.lambdaExpression([], AST.stringLiteral("timer")),
                    ]),
                  ]),
                ),
              ),
              assign(
                ":=",
                "third",
                AST.awaitExpression(
                  AST.arrayLiteral([
                    call(AST.memberAccessExpression(AST.identifier("Await"), "sleep_ms"), [
                      AST.integerLiteral(0),
                      AST.lambdaExpression([], AST.stringLiteral("timer")),
                    ]),
                  ]),
                ),
              ),
              AST.arrayLiteral([
                AST.binaryExpression("==", AST.identifier("first"), AST.stringLiteral("fallback")),
                AST.binaryExpression("==", AST.identifier("second"), AST.stringLiteral("timer")),
                AST.binaryExpression("==", AST.identifier("third"), AST.stringLiteral("timer")),
              ]),
            ]),
          ),
        ),
        call("future_flush", []),
        call(AST.memberAccessExpression(AST.identifier("runner"), AST.identifier("value")), []),
      ],
      [AST.importStatement(["able", "concurrency"], false, [AST.importSelector("Await")])],
    ),
    manifest: {
      description: "await honours the default arm and a ready timer awaitable without blocking",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "bool", value: true },
            { kind: "bool", value: true },
            { kind: "bool", value: true },
          ],
        },
      },
    },
  },
  {
    name: "concurrency/await_channel_send_cancel",
    module: AST.module([
      assign(":=", "ch", call("__able_channel_new", [AST.integerLiteral(0)])),
      assign(":=", "hits", AST.integerLiteral(0)),
      assign(
        ":=",
        "runner",
        AST.spawnExpression(
          AST.blockExpression([
            assign(
              "=",
              "result",
              AST.awaitExpression(
                AST.arrayLiteral([
                  call("__able_channel_await_try_send", [
                    AST.identifier("ch"),
                    AST.stringLiteral("payload"),
                    AST.lambdaExpression(
                      [],
                      AST.blockExpression([
                        assign(
                          "=",
                          "hits",
                          AST.binaryExpression("+", AST.identifier("hits"), AST.integerLiteral(1)),
                        ),
                        AST.stringLiteral("sent"),
                      ]),
                    ),
                  ]),
                ]),
              ),
            ),
            AST.identifier("result"),
          ]),
        ),
      ),
      call("future_flush", []),
      call(AST.memberAccessExpression(AST.identifier("runner"), AST.identifier("cancel")), []),
      call("future_flush", []),
      assign(":=", "late_send", call("__able_channel_try_send", [AST.identifier("ch"), AST.stringLiteral("late")])),
      AST.arrayLiteral([
        AST.binaryExpression("==", AST.identifier("hits"), AST.integerLiteral(0)),
        AST.binaryExpression("==", AST.identifier("late_send"), AST.booleanLiteral(false)),
      ]),
    ]),
    manifest: {
      description: "await send arm cancels registrations when the awaiting task is cancelled",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "bool", value: true },
            { kind: "bool", value: true },
          ],
        },
      },
    },
  },
  {
    name: "concurrency/await_channel_receive_cancel",
    module: AST.module([
      assign(":=", "ch", call("__able_channel_new", [AST.integerLiteral(0)])),
      assign(":=", "hits", AST.integerLiteral(0)),
      assign(
        ":=",
        "runner",
        AST.spawnExpression(
          AST.blockExpression([
            assign(
              "=",
              "result",
              AST.awaitExpression(
                AST.arrayLiteral([
                  call("__able_channel_await_try_recv", [
                    AST.identifier("ch"),
                    AST.lambdaExpression(
                      [AST.functionParameter("v")],
                      AST.blockExpression([
                        assign(
                          "=",
                          "hits",
                          AST.binaryExpression("+", AST.identifier("hits"), AST.integerLiteral(1)),
                        ),
                        AST.identifier("v"),
                      ]),
                    ),
                  ]),
                ]),
              ),
            ),
            AST.identifier("result"),
          ]),
        ),
      ),
      call("future_flush", []),
      call(AST.memberAccessExpression(AST.identifier("runner"), AST.identifier("cancel")), []),
      call("future_flush", []),
      assign(":=", "late_send", call("__able_channel_try_send", [AST.identifier("ch"), AST.stringLiteral("late")])),
      AST.arrayLiteral([
        AST.binaryExpression("==", AST.identifier("hits"), AST.integerLiteral(0)),
        AST.binaryExpression("==", AST.identifier("late_send"), AST.booleanLiteral(false)),
      ]),
    ]),
    manifest: {
      description: "await receive arm cancels registrations when the awaiting task is cancelled",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "bool", value: true },
            { kind: "bool", value: true },
          ],
        },
      },
    },
  },
  {
    name: "concurrency/await_mutex_cancel",
    module: AST.module([
      assign(":=", "mtx", call("__able_mutex_new")),
      call("__able_mutex_lock", [AST.identifier("mtx")]),
      assign(":=", "hits", AST.integerLiteral(0)),
      assign(
        ":=",
        "runner",
        AST.spawnExpression(
          AST.blockExpression([
            assign(
              "=",
              "result",
              AST.awaitExpression(
                AST.arrayLiteral([
                  call("__able_mutex_await_lock", [
                    AST.identifier("mtx"),
                    AST.lambdaExpression(
                      [],
                      AST.blockExpression([
                        assign(
                          "=",
                          "hits",
                          AST.binaryExpression("+", AST.identifier("hits"), AST.integerLiteral(1)),
                        ),
                        AST.stringLiteral("locked"),
                      ]),
                    ),
                  ]),
                ]),
              ),
            ),
            AST.identifier("result"),
          ]),
        ),
      ),
      call("future_flush", []),
      call(AST.memberAccessExpression(AST.identifier("runner"), AST.identifier("cancel")), []),
      call("future_flush", []),
      call("__able_mutex_unlock", [AST.identifier("mtx")]),
      assign(":=", "lock_again", call("__able_mutex_lock", [AST.identifier("mtx")])),
      call("__able_mutex_unlock", [AST.identifier("mtx")]),
      AST.arrayLiteral([
        AST.binaryExpression("==", AST.identifier("hits"), AST.integerLiteral(0)),
        AST.binaryExpression("==", AST.identifier("lock_again"), AST.nilLiteral()),
      ]),
    ]),
    manifest: {
      description: "await mutex lock cancels registrations when the awaiting task is cancelled",
      expect: {
        result: {
          kind: "array",
          elements: [
            { kind: "bool", value: true },
            { kind: "bool", value: true },
          ],
        },
      },
    },
  },
];

export default awaitFixtures;
