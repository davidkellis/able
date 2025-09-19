import { describe, expect, test } from "bun:test";
import * as AST from "../src/ast";
import { InterpreterV10 } from "../src/interpreter";
import type { V10Value } from "../src/interpreter";

const flushScheduler = () => new Promise<void>(resolve => setTimeout(resolve, 0));

const expectErrorValue = (value: V10Value) => {
  expect(value.kind).toBe("error");
  if (value.kind !== "error") throw new Error("expected error value");
  return value;
};

const expectStructInstance = (value: V10Value | undefined, structName: string) => {
  expect(value && value.kind === "struct_instance" && value.def.id.name === structName).toBe(true);
  if (!value || value.kind !== "struct_instance" || value.def.id.name !== structName) {
    throw new Error(`expected struct_instance ${structName}`);
  }
  return value;
};

describe("v10 interpreter - proc & spawn handles", () => {
  test("proc handle supports status, value, and cancel", async () => {
    const I = new InterpreterV10();

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
        AST.procExpression(
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
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 11 });

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
        AST.procExpression(AST.blockExpression([AST.integerLiteral(5)]))
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
    expect(cancelledResult.message).toMatch(/Proc cancelled/);
    const procErr = expectStructInstance(cancelledResult.value, "ProcError");
    const procErrDetails = (procErr.values as Map<string, any>).get("details");
    expect(procErrDetails).toBeDefined();
    expect(procErrDetails?.kind).toBe("string");
    expect((procErrDetails as any).value).toMatch(/cancelled/);

    // ! operator should propagate the ProcError
    const propagateCancelled = AST.propagationExpression(cancelledValue);
    let cancelledThrown = false;
    try {
      I.evaluate(propagateCancelled);
    } catch (e: any) {
      cancelledThrown = true;
      expect(e).toBeInstanceOf(Error);
      expect(e?.value?.message ?? "").toMatch(/Proc cancelled/);
    }
    expect(cancelledThrown).toBe(true);
  });

  test("proc failure surfaces ProcStatus::Failed and ProcError", () => {
    const I = new InterpreterV10();

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
        AST.procExpression(AST.functionCall(AST.identifier("boom"), []))
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
    expect(failureValue.message).toMatch(/Proc failed/);
    const procErr = expectStructInstance(failureValue.value, "ProcError");
    const details = (procErr.values as Map<string, any>).get("details");
    expect(details).toBeDefined();
    expect(details?.kind).toBe("string");
    expect((details as any).value).toMatch(/kaboom/);

    const failedStatus = I.evaluate(statusCall) as any;
    expect(failedStatus.kind).toBe("struct_instance");
    expect(failedStatus.def.id.name).toBe("Failed");
    const failedMap = failedStatus.values as Map<string, any>;
    const statusErr = failedMap.get("error") as any;
    expect(statusErr.kind).toBe("struct_instance");
    expect(statusErr.def.id.name).toBe("ProcError");
    const statusDetails = (statusErr.values as Map<string, any>).get("details");
    expect(statusDetails.kind).toBe("string");
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
    const I = new InterpreterV10();

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
    expect(I.evaluate(futureValue)).toEqual({ kind: "i32", value: 1 });
    const futureResolved = I.evaluate(futureStatus) as any;
    expect(futureResolved.kind).toBe("struct_instance");
    expect(futureResolved.def.id.name).toBe("Resolved");

    // Memoized value reused
    expect(I.evaluate(futureValue)).toEqual({ kind: "i32", value: 1 });
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 1 });

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
    const futureErr = expectStructInstance(badResult.value, "ProcError");
    const futureErrDetails = (futureErr.values as Map<string, any>).get("details");
    expect(futureErrDetails).toBeDefined();
    expect(futureErrDetails?.kind).toBe("string");
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
    const procError = failedMap.get("error") as any;
    expect(procError.kind).toBe("struct_instance");
    expect(procError.def.id.name).toBe("ProcError");
    const detailsMap = procError.values as Map<string, any>;
    const detailsVal = detailsMap.get("details");
    expect(detailsVal.kind).toBe("string");
    expect(detailsVal.value).toMatch(/boom/);
  });

  test("proc handle progresses without explicit join", async () => {
    const I = new InterpreterV10();

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
        AST.procExpression(AST.functionCall(AST.identifier("bump"), []))
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
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 1 });
    expect(I.evaluate(AST.identifier("counter"))).toEqual({ kind: "i32", value: 1 });
  });

  test("proc cancel before start surfaces ProcError", async () => {
    const I = new InterpreterV10();

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
        AST.procExpression(
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
    expect(cancelledResult.message).toMatch(/Proc cancelled/);
    const errPayload = expectStructInstance(cancelledResult.value, "ProcError");
    const details = (errPayload.values as Map<string, any>).get("details");
    expect(details).toBeDefined();
    expect(details?.kind).toBe("string");
    expect((details as any).value).toMatch(/cancelled/);
    expect(I.evaluate(AST.identifier("flag"))).toEqual({ kind: "i32", value: 0 });
  });

  test("proc cancel after resolve is no-op", async () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("handle"),
        AST.procExpression(AST.blockExpression([AST.integerLiteral(5)]))
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
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 5 });
  });

  test("future resolves after scheduler tick", async () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("count"),
        AST.integerLiteral(0)
      )
    );

    const inc = AST.functionDefinition(
      "inc",
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
    I.evaluate(inc);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("future"),
        AST.spawnExpression(AST.functionCall(AST.identifier("inc"), []))
      )
    );

    await flushScheduler();

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future"), "status"),
      []
    );
    const status = I.evaluate(statusCall) as any;
    expect(status.kind).toBe("struct_instance");
    expect(status.def.id.name).toBe("Resolved");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future"), "value"),
      []
    );
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 1 });
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 1 });
  });

  test("proc yield allows interleaving tasks", async () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("trace"),
        AST.stringLiteral("")
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("stage"),
        AST.integerLiteral(0)
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("slow"),
        AST.procExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage"),
                AST.integerLiteral(0)
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("A")
                  )
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("stage"),
                  AST.integerLiteral(1)
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
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
                        AST.stringLiteral("C")
                      )
                    ),
                  ])
                ),
              ]
            ),
            AST.integerLiteral(1),
          ])
        )
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("fast"),
        AST.procExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              "=",
              AST.identifier("trace"),
              AST.binaryExpression(
                "+",
                AST.identifier("trace"),
                AST.stringLiteral("B")
              )
            ),
            AST.integerLiteral(2),
          ])
        )
      )
    );

    const slowStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("slow"), "status"),
      []
    );
    const slowValueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("slow"), "value"),
      []
    );
    const fastStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("fast"), "status"),
      []
    );

    await flushScheduler();

    const slowValue = I.evaluate(slowValueCall);
    expect(slowValue).toEqual({ kind: "i32", value: 1 });

    const traceVal = I.evaluate(AST.identifier("trace"));
    expect(traceVal).toEqual({ kind: "string", value: "ABC" });

    const slowStatus = I.evaluate(slowStatusCall) as any;
    expect(slowStatus.kind).toBe("struct_instance");
    expect(slowStatus.def.id.name).toBe("Resolved");

    const fastStatus = I.evaluate(fastStatusCall) as any;
    expect(fastStatus.kind).toBe("struct_instance");
    expect(fastStatus.def.id.name).toBe("Resolved");
  });

  test("proc task observes cancellation cooperatively", () => {
    const I = new InterpreterV10();

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("trace"),
        AST.stringLiteral("")
      )
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("saw_cancel"),
        AST.booleanLiteral(false)
      )
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("stage"),
        AST.integerLiteral(0)
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("handle"),
        AST.procExpression(
          AST.blockExpression([
            AST.assignmentExpression(
              "=",
              AST.identifier("stage"),
              AST.binaryExpression(
                "+",
                AST.identifier("stage"),
                AST.integerLiteral(1)
              )
            ),
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("stage"),
                AST.integerLiteral(1)
              ),
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("w")
                  )
                ),
                AST.functionCall(AST.identifier("proc_yield"), []),
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
                        AST.stringLiteral("x")
                      )
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("saw_cancel"),
                      AST.functionCall(AST.identifier("proc_cancelled"), [])
                    ),
                    AST.integerLiteral(0),
                  ])
                ),
              ]
            ),
            AST.integerLiteral(0),
          ])
        )
      )
    );

    const handleVal = I.evaluate(AST.identifier("handle"));
    (I as any).runProcHandle(handleVal);
    expect(I.evaluate(AST.identifier("trace"))).toEqual({ kind: "string", value: "w" });

    handleVal.cancelRequested = true;
    (I as any).runProcHandle(handleVal);

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "status"),
      []
    );
    const status = I.evaluate(statusCall) as any;
    expect(status.kind).toBe("struct_instance");
    expect(status.def.id.name).toBe("Cancelled");

    const valueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("handle"), "value"),
      []
    );
    const valueResult = I.evaluate(valueCall);
    expect(valueResult.kind).toBe("error");
    expect(valueResult.message).toMatch(/cancelled/i);

    const sawCancel = I.evaluate(AST.identifier("saw_cancel"));
    expect(sawCancel).toEqual({ kind: "bool", value: true });

    const trace = I.evaluate(AST.identifier("trace"));
    expect(trace.kind).toBe("string");
    expect(trace.value).toBe("wx");
  });
});
