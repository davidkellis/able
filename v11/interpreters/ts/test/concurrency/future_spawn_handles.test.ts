import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import type { RuntimeValue } from "../../src/interpreter";

import { appendToTrace, drainScheduler, expectErrorValue, expectStructInstance, flushScheduler } from "./future_spawn.helpers";

describe("v11 interpreter - future & spawn handles", () => {
  test("future handle supports status, value, and cancel", async () => {
    const I = new Interpreter();

    const add1 = AST.functionDefinition(
      "add1",
      [AST.functionParameter("x")],
      AST.blockExpression([
        AST.returnStatement(
          AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(1))
        ),
      ])
    );
    I.evaluate(add1);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("handle"),
        AST.spawnExpression(
          AST.functionCall(AST.identifier("add1"), [AST.integerLiteral(10)])
        )
      )
    );

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "status"),
      []
    );
    const pendingStatus = I.evaluate(statusCall) as any;
    expect(pendingStatus.kind).toBe("struct_instance");
    expect(pendingStatus.def.id.name).toBe("Pending");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "value"),
      []
    );
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 11n });

    const resolvedStatus = I.evaluate(statusCall) as any;
    expect(resolvedStatus.kind).toBe("struct_instance");
    expect(resolvedStatus.def.id.name).toBe("Resolved");

    const cancelCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
      []
    );
    I.evaluate(cancelCall);
    const statusAfterCancel = I.evaluate(statusCall) as any;
    expect(statusAfterCancel.kind).toBe("struct_instance");
    expect(statusAfterCancel.def.id.name).toBe("Resolved");

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("pending"),
        AST.spawnExpression(AST.blockExpression([AST.integerLiteral(5)]))
      )
    );
    const cancelPending = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("pending"), "cancel"),
      []
    );
    I.evaluate(cancelPending);
    const pendingStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("pending"), "status"),
      []
    );
    await flushScheduler();
    const cancelledStatus = I.evaluate(pendingStatusCall) as any;
    expect(cancelledStatus.kind).toBe("struct_instance");
    expect(cancelledStatus.def.id.name).toBe("Cancelled");

    const cancelledValue = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("pending"), "value"),
      []
    );
    const cancelledResult = expectErrorValue(I.evaluate(cancelledValue));
    expect(cancelledResult.message).toMatch(/Future cancelled/);
    const futureErr = expectStructInstance(cancelledResult.value, "FutureError");
    const futureErrDetails = (futureErr.values as Map<string, any>).get("details");
    expect(futureErrDetails).toBeDefined();
    expect(futureErrDetails?.kind).toBe("String");
    expect((futureErrDetails as any).value).toMatch(/cancelled/);

    // ! operator should propagate the FutureError
    const propagateCancelled = AST.propagationExpression(cancelledValue);
    let cancelledThrown = false;
    try {
      I.evaluate(propagateCancelled);
    } catch (e: any) {
      cancelledThrown = true;
      expect(e).toBeInstanceOf(Error);
      expect(e?.value?.message ?? "").toMatch(/Future cancelled/);
    }
    expect(cancelledThrown).toBe(true);
  });

  test("future failure surfaces FutureStatus::Failed and FutureError", () => {
    const I = new Interpreter();

    const boom = AST.functionDefinition(
      "boom",
      [],
      AST.blockExpression([
        AST.raiseStatement(AST.stringLiteral("kaboom")),
      ])
    );
    I.evaluate(boom);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("failure"),
        AST.spawnExpression(AST.functionCall(AST.identifier("boom"), []))
      )
    );

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("failure"), "status"),
      []
    );
    const initialStatus = I.evaluate(statusCall) as any;
    expect(initialStatus.kind).toBe("struct_instance");
    expect(initialStatus.def.id.name).toBe("Pending");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("failure"), "value"),
      []
    );
    const failureValue = expectErrorValue(I.evaluate(valueCall));
    expect(failureValue.message).toMatch(/Future failed/);
    const futureErr = expectStructInstance(failureValue.value, "FutureError");
    const details = (futureErr.values as Map<string, any>).get("details");
    expect(details).toBeDefined();
    expect(details?.kind).toBe("String");
    expect((details as any).value).toMatch(/kaboom/);

    const failedStatus = I.evaluate(statusCall) as any;
    expect(failedStatus.kind).toBe("struct_instance");
    expect(failedStatus.def.id.name).toBe("Failed");
    const failedMap = failedStatus.values as Map<string, any>;
    const statusErr = failedMap.get("error") as any;
    expect(statusErr.kind).toBe("struct_instance");
    expect(statusErr.def.id.name).toBe("FutureError");
    const statusDetails = (statusErr.values as Map<string, any>).get("details");
    expect(statusDetails.kind).toBe("String");
    expect(statusDetails.value).toMatch(/kaboom/);

    const propagate = AST.propagationExpression(valueCall);
    let threw = false;
    try {
      I.evaluate(propagate);
    } catch (e: any) {
      threw = true;
      expect(e).toBeInstanceOf(Error);
      expect(e?.value?.message ?? "").toMatch(/kaboom/);
    }
    expect(threw).toBe(true);
  });

  test("spawn returns memoized future handle", () => {
    const I = new Interpreter();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("count"),
        AST.integerLiteral(0)
      )
    );

    const nextFn = AST.functionDefinition(
      "next",
      [],
      AST.blockExpression([
        AST.assignmentExpression(
          "+=",
          AST.identifier("count"),
          AST.integerLiteral(1)
        ),
        AST.returnStatement(AST.identifier("count")),
      ])
    );
    I.evaluate(nextFn);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("future"),
        AST.spawnExpression(AST.functionCall(AST.identifier("next"), []))
      )
    );

    const futureStatus = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future"), "status"),
      []
    );
    const futurePending = I.evaluate(futureStatus) as any;
    expect(futurePending.kind).toBe("struct_instance");
    expect(futurePending.def.id.name).toBe("Pending");

    const futureValue = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future"), "value"),
      []
    );
    expect(I.evaluate(futureValue)).toEqual({ kind: "i32", value: 1n });
    const futureResolved = I.evaluate(futureStatus) as any;
    expect(futureResolved.kind).toBe("struct_instance");
    expect(futureResolved.def.id.name).toBe("Resolved");

    // Memoized value reused
    expect(I.evaluate(futureValue)).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 1n });

    const boomFn = AST.functionDefinition(
      "boom",
      [],
      AST.blockExpression([
        AST.raiseStatement(AST.stringLiteral("boom")),
      ])
    );
    I.evaluate(boomFn);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("bad"),
        AST.spawnExpression(AST.functionCall(AST.identifier("boom"), []))
      )
    );

    const badValue = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("bad"), "value"),
      []
    );
    const badResult = expectErrorValue(I.evaluate(badValue));
    expect(badResult.message).toMatch(/boom/);
    const futureErr = expectStructInstance(badResult.value, "FutureError");
    const futureErrDetails = (futureErr.values as Map<string, any>).get("details");
    expect(futureErrDetails).toBeDefined();
    expect(futureErrDetails?.kind).toBe("String");
    expect((futureErrDetails as any).value).toMatch(/boom/);

    const propagateFailure = AST.propagationExpression(badValue);
    let futureThrown = false;
    try {
      I.evaluate(propagateFailure);
    } catch (e: any) {
      futureThrown = true;
      expect(e).toBeInstanceOf(Error);
      expect(e?.value?.message ?? "").toMatch(/boom/);
    }
    expect(futureThrown).toBe(true);

    const badStatus = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("bad"), "status"),
      []
    );
    const failedStatus = I.evaluate(badStatus) as any;
    expect(failedStatus.kind).toBe("struct_instance");
    expect(failedStatus.def.id.name).toBe("Failed");
    const failedMap = failedStatus.values as Map<string, any>;
    const futureError = failedMap.get("error") as any;
    expect(futureError.kind).toBe("struct_instance");
    expect(futureError.def.id.name).toBe("FutureError");
    const detailsMap = futureError.values as Map<string, any>;
    const detailsVal = detailsMap.get("details");
    expect(detailsVal.kind).toBe("String");
    expect(detailsVal.value).toMatch(/boom/);
  });

  test("future handle progresses without explicit join", async () => {
    const I = new Interpreter();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("counter"),
        AST.integerLiteral(0)
      )
    );

    const bump = AST.functionDefinition(
      "bump",
      [],
      AST.blockExpression([
        AST.assignmentExpression(
          "+=",
          AST.identifier("counter"),
          AST.integerLiteral(1)
        ),
        AST.returnStatement(AST.identifier("counter")),
      ])
    );
    I.evaluate(bump);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("handle"),
        AST.spawnExpression(AST.functionCall(AST.identifier("bump"), []))
      )
    );

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "status"),
      []
    );
    const initial = I.evaluate(statusCall) as any;
    expect(initial.kind).toBe("struct_instance");
    expect(initial.def.id.name).toBe("Pending");

    await flushScheduler();

    const resolved = I.evaluate(statusCall) as any;
    expect(resolved.kind).toBe("struct_instance");
    expect(resolved.def.id.name).toBe("Resolved");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "value"),
      []
    );
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(AST.identifier("counter"))).toEqual({ kind: "i32", value: 1n });
  });

  test("future cancel before start surfaces FutureError", async () => {
    const I = new Interpreter();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("flag"),
        AST.integerLiteral(0)
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("cancelled"),
        AST.spawnExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              "=",
              AST.identifier("flag"),
              AST.integerLiteral(1)
            ),
            AST.integerLiteral(42),
          ])
        )
      )
    );

    const cancelCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("cancelled"), "cancel"),
      []
    );
    I.evaluate(cancelCall);

    await flushScheduler();

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("cancelled"), "status"),
      []
    );
    const status = I.evaluate(statusCall) as any;
    expect(status.kind).toBe("struct_instance");
    expect(status.def.id.name).toBe("Cancelled");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("cancelled"), "value"),
      []
    );
    const cancelledResult = expectErrorValue(I.evaluate(valueCall));
    expect(cancelledResult.message).toMatch(/Future cancelled/);
    const errPayload = expectStructInstance(cancelledResult.value, "FutureError");
    const details = (errPayload.values as Map<string, any>).get("details");
    expect(details).toBeDefined();
    expect(details?.kind).toBe("String");
    expect((details as any).value).toMatch(/cancelled/);
    expect(I.evaluate(AST.identifier("flag"))).toEqual({ kind: "i32", value: 0n });
  });

  test("future errors expose cause via err.cause()", async () => {
    const I = new Interpreter();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("handle"),
        AST.spawnExpression(
          AST.blockExpression([
            AST.raiseStatement(
              AST.structLiteral(
                [AST.structFieldInitializer(AST.stringLiteral("boom"), "details")],
                false,
                "FutureError",
              ),
            ),
          ]),
        ),
      ),
    );

    await flushScheduler();

    I.evaluate(
      AST.functionDefinition(
        "describe_error",
        [AST.functionParameter("err", AST.simpleTypeExpression("Error"))],
        AST.blockExpression([
          AST.assignmentExpression(
            ":=",
            AST.identifier("cause"),
            AST.functionCall(AST.memberAccessExpression(AST.identifier("err"), "cause"), []),
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
            AST.functionCall(AST.memberAccessExpression(AST.identifier("err"), "message"), []),
            AST.stringLiteral("|"),
            AST.identifier("flag"),
          ]),
        ]),
        AST.simpleTypeExpression("String"),
      ),
    );
    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "value"),
      [],
    );
    const errorValue = expectErrorValue(I.evaluate(valueCall));
    expect(errorValue.cause).toEqual(errorValue.value);
    const causeSummary = AST.orElseExpression(
      AST.propagationExpression(valueCall),
      AST.blockExpression([
        AST.functionCall(AST.identifier("describe_error"), [AST.identifier("err")]),
      ]),
      "err",
    );

    const summary = I.evaluate(causeSummary) as RuntimeValue;
    expect(summary.kind).toBe("String");
    expect(summary.value).toBe("Future failed: boom|has-cause");
  });

  test("future cancel after resolve is no-op", async () => {
    const I = new Interpreter();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("handle"),
        AST.spawnExpression(AST.blockExpression([AST.integerLiteral(5)]))
      )
    );

    await flushScheduler();

    const cancelCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
      []
    );
    I.evaluate(cancelCall);

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "status"),
      []
    );
    const status = I.evaluate(statusCall) as any;
    expect(status.kind).toBe("struct_instance");
    expect(status.def.id.name).toBe("Resolved");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "value"),
      []
    );
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 5n });
  });

});
