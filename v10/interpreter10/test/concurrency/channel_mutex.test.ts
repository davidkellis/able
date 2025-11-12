import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";
import { RaiseSignal } from "../../src/interpreter/signals";

const call = (name: string, args = []) =>
  AST.functionCall(AST.identifier(name), args);

const memberCall = (object: string, member: string, args = []) =>
  AST.functionCall(AST.memberAccessExpression(AST.identifier(object), member), args);

describe("channel helpers", () => {
  test("unbuffered send blocks until a receiver arrives", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("ch"),
        call("__able_channel_new", [AST.integerLiteral(0)]),
      ),
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("sender"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_send", [AST.identifier("ch"), AST.stringLiteral("value")]),
            AST.stringLiteral("done"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const statusPending = I.evaluate(memberCall("sender", "status")) as any;
    expect(statusPending.kind).toBe("struct_instance");
    expect(statusPending.def.id.name).toBe("Pending");

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("receiver"),
        AST.procExpression(
          AST.blockExpression([call("__able_channel_receive", [AST.identifier("ch")])]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const senderStatus = I.evaluate(memberCall("sender", "status")) as any;
    expect(senderStatus.def.id.name).toBe("Resolved");
    const senderValue = I.evaluate(memberCall("sender", "value")) as any;
    expect(senderValue).toEqual({ kind: "string", value: "done" });

    const receiverResult = I.evaluate(memberCall("receiver", "value")) as any;
    expect(receiverResult).toEqual({ kind: "string", value: "value" });
  });

  test("buffered channel blocks when capacity is exhausted", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("ch"),
        call("__able_channel_new", [AST.integerLiteral(1)]),
      ),
    );

    // First send fills the buffer immediately.
    I.evaluate(call("__able_channel_send", [AST.identifier("ch"), AST.stringLiteral("first")]));

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("blocked"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_send", [AST.identifier("ch"), AST.stringLiteral("second")]),
            AST.stringLiteral("sent"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const blockedStatus = I.evaluate(memberCall("blocked", "status")) as any;
    expect(blockedStatus.def.id.name).toBe("Pending");

    const firstValue = I.evaluate(call("__able_channel_receive", [AST.identifier("ch")])) as any;
    expect(firstValue).toEqual({ kind: "string", value: "first" });

    I.evaluate(call("proc_flush"));

    const blockedStatusAfter = I.evaluate(memberCall("blocked", "status")) as any;
    expect(blockedStatusAfter.def.id.name).toBe("Resolved");
    const blockedReturn = I.evaluate(memberCall("blocked", "value")) as any;
    expect(blockedReturn).toEqual({ kind: "string", value: "sent" });

    const secondValue = I.evaluate(call("__able_channel_receive", [AST.identifier("ch")])) as any;
    expect(secondValue).toEqual({ kind: "string", value: "second" });
  });

  test("closing a channel wakes waiting receivers and errors senders", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("ch"),
        call("__able_channel_new", [AST.integerLiteral(0)]),
      ),
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("waitingReceiver"),
        AST.procExpression(
          AST.blockExpression([call("__able_channel_receive", [AST.identifier("ch")])]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    I.evaluate(call("__able_channel_close", [AST.identifier("ch")]));
    I.evaluate(call("proc_flush"));

    const receiverStatus = I.evaluate(memberCall("waitingReceiver", "status")) as any;
    expect(receiverStatus.def.id.name).toBe("Resolved");
    const receiverValue = I.evaluate(memberCall("waitingReceiver", "value")) as any;
    expect(receiverValue).toEqual({ kind: "nil", value: null });

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("failingSender"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_send", [AST.identifier("ch"), AST.stringLiteral("payload")]),
            AST.stringLiteral("unreachable"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const senderStatus = I.evaluate(memberCall("failingSender", "status")) as any;
    expect(senderStatus.def.id.name).toBe("Failed");
    const senderError = I.evaluate(memberCall("failingSender", "value")) as any;
    expect(senderError.kind).toBe("error");
    expect(senderError.message).toContain("send on closed channel");
  });

  test("cancelling a blocked sender removes it from waiters", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("ch"),
        call("__able_channel_new", [AST.integerLiteral(0)]),
      ),
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("blockedSender"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_send", [AST.identifier("ch"), AST.stringLiteral("first")]),
            AST.stringLiteral("unreachable"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const sendStatus = I.evaluate(memberCall("blockedSender", "status")) as any;
    expect(sendStatus.def.id.name).toBe("Pending");

    I.evaluate(memberCall("blockedSender", "cancel"));
    I.evaluate(call("proc_flush"));

    const cancelledStatus = I.evaluate(memberCall("blockedSender", "status")) as any;
    expect(cancelledStatus.def.id.name).toBe("Cancelled");
    const cancelledValue = I.evaluate(memberCall("blockedSender", "value")) as any;
    expect(cancelledValue.kind).toBe("error");
    expect(cancelledValue.message).toContain("cancelled");

    const drain = I.evaluate(call("__able_channel_try_receive", [AST.identifier("ch")])) as any;
    expect(drain).toEqual({ kind: "nil", value: null });

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("receiver2"),
        AST.procExpression(
          AST.blockExpression([call("__able_channel_receive", [AST.identifier("ch")])]),
        ),
      ),
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("sender2"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_send", [AST.identifier("ch"), AST.stringLiteral("second")]),
            AST.stringLiteral("ok"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const sender2Status = I.evaluate(memberCall("sender2", "status")) as any;
    expect(sender2Status.def.id.name).toBe("Resolved");
    const sender2Value = I.evaluate(memberCall("sender2", "value")) as any;
    expect(sender2Value).toEqual({ kind: "string", value: "ok" });

    const receiver2Value = I.evaluate(memberCall("receiver2", "value")) as any;
    expect(receiver2Value).toEqual({ kind: "string", value: "second" });

    I.evaluate(call("proc_flush"));
  });

  test("nil channel send blocks until cancellation", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("blockedSender"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_send", [AST.integerLiteral(0), AST.stringLiteral("value")]),
            AST.stringLiteral("unreachable"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const pendingStatus = I.evaluate(memberCall("blockedSender", "status")) as any;
    expect(pendingStatus.def.id.name).toBe("Pending");

    I.evaluate(memberCall("blockedSender", "cancel"));
    I.evaluate(call("proc_flush"));

    const cancelledStatus = I.evaluate(memberCall("blockedSender", "status")) as any;
    expect(cancelledStatus.def.id.name).toBe("Cancelled");
    const cancelledValue = I.evaluate(memberCall("blockedSender", "value")) as any;
    expect(cancelledValue.kind).toBe("error");
    expect(cancelledValue.message.toLowerCase()).toContain("cancel");
  });

  test("nil channel receive blocks until cancellation", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("blockedReceiver"),
        AST.procExpression(
          AST.blockExpression([
            call("__able_channel_receive", [AST.integerLiteral(0)]),
            AST.stringLiteral("unreachable"),
          ]),
        ),
      ),
    );

    I.evaluate(call("proc_flush"));

    const pendingStatus = I.evaluate(memberCall("blockedReceiver", "status")) as any;
    expect(pendingStatus.def.id.name).toBe("Pending");

    I.evaluate(memberCall("blockedReceiver", "cancel"));
    I.evaluate(call("proc_flush"));

    const cancelledStatus = I.evaluate(memberCall("blockedReceiver", "status")) as any;
    expect(cancelledStatus.def.id.name).toBe("Cancelled");
    const cancelledValue = I.evaluate(memberCall("blockedReceiver", "value")) as any;
    expect(cancelledValue.kind).toBe("error");
    expect(cancelledValue.message.toLowerCase()).toContain("cancel");
  });

  test("send on closed channel surfaces ChannelSendOnClosed struct", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("ch"),
        call("__able_channel_new", [AST.integerLiteral(0)]),
      ),
    );

    I.evaluate(call("__able_channel_close", [AST.identifier("ch")]));

    try {
      I.evaluate(call("__able_channel_send", [AST.identifier("ch"), AST.integerLiteral(1)]));
      throw new Error("expected send to raise");
    } catch (err) {
      expect(err).toBeInstanceOf(RaiseSignal);
      const signal = err as RaiseSignal;
      expect(signal.value.kind).toBe("error");
      expect(signal.value.message).toContain("send on closed channel");
      const payload = signal.value.value;
      expect(payload?.kind).toBe("struct_instance");
      expect(payload?.def.id.name).toBe("ChannelSendOnClosed");
    }
  });
});

describe("mutex helpers", () => {
  test("lock/unlock errors when reentered outside procs", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("mutex"),
        call("__able_mutex_new"),
      ),
    );

    I.evaluate(call("__able_mutex_lock", [AST.identifier("mutex")]));

    expect(() =>
      I.evaluate(call("__able_mutex_lock", [AST.identifier("mutex")])),
    ).toThrow("Mutex already locked");

    I.evaluate(call("__able_mutex_unlock", [AST.identifier("mutex")]));
    I.evaluate(call("__able_mutex_unlock", [AST.identifier("mutex")]));
  });
});
