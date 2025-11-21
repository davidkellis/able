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
        AST.procExpression(
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
        AST.procExpression(
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
      call("proc_flush", []),
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
            { kind: "string", value: "payload" },
            { kind: "string", value: "sent" },
            { kind: "string", value: "payload" },
            { kind: "string", value: "sent" },
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
        AST.procExpression(
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
        result: { kind: "string", value: "A1B1A2B2" },
      },
    },
  },
  {
    name: "concurrency/await_timer_default",
    module: AST.module([
      AST.structDefinition("CancelReg", [], "named"),
      AST.methodsDefinition(AST.simpleTypeExpression("CancelReg"), [
        AST.functionDefinition(
          "cancel",
          [AST.functionParameter("self")],
          AST.blockExpression([AST.nilLiteral()]),
          AST.simpleTypeExpression("void"),
        ),
      ]),
      AST.structDefinition(
        "Timer",
        [AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "ready")],
        "named",
      ),
      AST.methodsDefinition(AST.simpleTypeExpression("Timer"), [
        AST.functionDefinition(
          "is_ready",
          [AST.functionParameter("self")],
          AST.blockExpression([AST.memberAccessExpression(AST.identifier("self"), AST.identifier("ready"))]),
          AST.simpleTypeExpression("bool"),
        ),
        AST.functionDefinition(
          "register",
          [AST.functionParameter("self"), AST.functionParameter("_waker")],
          AST.blockExpression([AST.structLiteral([], false, "CancelReg")]),
          AST.simpleTypeExpression("CancelReg"),
        ),
        AST.functionDefinition(
          "commit",
          [AST.functionParameter("self")],
          AST.blockExpression([AST.stringLiteral("timer")]),
          AST.simpleTypeExpression("string"),
        ),
        AST.functionDefinition(
          "is_default",
          [AST.functionParameter("self")],
          AST.blockExpression([AST.booleanLiteral(false)]),
          AST.simpleTypeExpression("bool"),
        ),
      ]),
      AST.structDefinition(
        "DefaultArm",
        [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "value")],
        "named",
      ),
      AST.methodsDefinition(AST.simpleTypeExpression("DefaultArm"), [
        AST.functionDefinition(
          "is_ready",
          [AST.functionParameter("self")],
          AST.blockExpression([AST.booleanLiteral(false)]),
          AST.simpleTypeExpression("bool"),
        ),
        AST.functionDefinition(
          "register",
          [AST.functionParameter("self"), AST.functionParameter("_waker")],
          AST.blockExpression([AST.structLiteral([], false, "CancelReg")]),
          AST.simpleTypeExpression("CancelReg"),
        ),
        AST.functionDefinition(
          "commit",
          [AST.functionParameter("self")],
          AST.blockExpression([
            AST.memberAccessExpression(AST.identifier("self"), AST.identifier("value")),
          ]),
          AST.simpleTypeExpression("string"),
        ),
        AST.functionDefinition(
          "is_default",
          [AST.functionParameter("self")],
          AST.blockExpression([AST.booleanLiteral(true)]),
          AST.simpleTypeExpression("bool"),
        ),
      ]),
      AST.functionDefinition(
        "default_arm",
        [AST.functionParameter("value", AST.simpleTypeExpression("string"))],
        AST.blockExpression([
          AST.structLiteral(
            [AST.structFieldInitializer(AST.identifier("value"), "value")],
            false,
            "DefaultArm",
          ),
        ]),
        AST.simpleTypeExpression("DefaultArm"),
      ),
      assign(
        ":=",
        "timer",
        AST.structLiteral([AST.structFieldInitializer(AST.booleanLiteral(true), "ready")], false, "Timer"),
      ),
      assign(
        ":=",
        "runner",
        AST.procExpression(
          AST.blockExpression([
            assign(
              ":=",
              "first",
              AST.awaitExpression(AST.arrayLiteral([call("default_arm", [AST.stringLiteral("fallback")])])),
            ),
            assign(":=", "second", AST.awaitExpression(AST.arrayLiteral([AST.identifier("timer")]))),
            assign(":=", "third", AST.awaitExpression(AST.arrayLiteral([AST.identifier("timer")]))),
            AST.arrayLiteral([
              AST.binaryExpression("==", AST.identifier("first"), AST.stringLiteral("fallback")),
              AST.binaryExpression("==", AST.identifier("second"), AST.stringLiteral("timer")),
              AST.binaryExpression("==", AST.identifier("third"), AST.stringLiteral("timer")),
            ]),
          ]),
        ),
      ),
      call("proc_flush", []),
      call(AST.memberAccessExpression(AST.identifier("runner"), AST.identifier("value")), []),
    ]),
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
];

export default awaitFixtures;
