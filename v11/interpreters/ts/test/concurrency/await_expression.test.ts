import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import { flushScheduler } from "./proc_spawn.helpers";

describe("v11 interpreter - await expression", () => {
  test("await resolves manual awaitable once waker fires", () => {
    const I = new Interpreter();

    I.evaluate(AST.assignmentExpression(":=", AST.identifier("last_waker"), AST.nilLiteral()));

    const manualAwaitableStruct = AST.structDefinition(
      "ManualAwaitable",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "ready"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value"),
        AST.structFieldDefinition(
          AST.nullableTypeExpression(AST.simpleTypeExpression("AwaitWaker")),
          "waker",
        ),
      ],
      "named",
    );
    I.evaluate(manualAwaitableStruct);

    const manualRegistrationStruct = AST.structDefinition(
      "ManualRegistration",
      [AST.structFieldDefinition(AST.simpleTypeExpression("ManualAwaitable"), "owner")],
      "named",
    );
    I.evaluate(manualRegistrationStruct);

    const isReadyFn = AST.functionDefinition(
      "is_ready",
      [],
      AST.blockExpression([AST.returnStatement(AST.implicitMemberExpression("ready"))]),
      AST.simpleTypeExpression("bool"),
      undefined,
      undefined,
      true,
    );

    const registerFn = AST.functionDefinition(
      "register",
      [AST.functionParameter("waker", AST.simpleTypeExpression("AwaitWaker"))],
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.implicitMemberExpression("waker"),
          AST.identifier("waker"),
        ),
        AST.assignmentExpression(
          "=",
          AST.identifier("last_waker"),
          AST.identifier("waker"),
        ),
        AST.returnStatement(
          AST.structLiteral(
            [AST.structFieldInitializer(AST.identifier("self"), "owner")],
            false,
            "ManualRegistration",
          ),
        ),
      ]),
      AST.simpleTypeExpression("ManualRegistration"),
      undefined,
      undefined,
      true,
    );

    const commitFn = AST.functionDefinition(
      "commit",
      [],
      AST.blockExpression([AST.returnStatement(AST.implicitMemberExpression("value"))]),
      AST.simpleTypeExpression("i32"),
      undefined,
      undefined,
      true,
    );

    const manualAwaitableMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("ManualAwaitable"),
      [isReadyFn, registerFn, commitFn],
    );
    I.evaluate(manualAwaitableMethods);

    const cancelFn = AST.functionDefinition(
      "cancel",
      [],
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.memberAccessExpression(AST.implicitMemberExpression("owner"), "waker"),
          AST.nilLiteral(),
        ),
        AST.returnStatement(AST.nilLiteral()),
      ]),
      AST.simpleTypeExpression("void"),
      undefined,
      undefined,
      true,
    );
    const manualRegistrationMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("ManualRegistration"),
      [cancelFn],
    );
    I.evaluate(manualRegistrationMethods);

    const manualArm = AST.structLiteral(
      [
        AST.structFieldInitializer(AST.booleanLiteral(false), "ready"),
        AST.structFieldInitializer(AST.integerLiteral(42), "value"),
        AST.structFieldInitializer(AST.nilLiteral(), "waker"),
      ],
      false,
      "ManualAwaitable",
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("arm"), manualArm));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.integerLiteral(0)));

    const awaitFuture = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([AST.identifier("arm")])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), awaitFuture));

    const handleVal = I.evaluate(AST.identifier("handle"));
    (I as any).runFuture(handleVal);

    const pendingResult = I.evaluate(AST.identifier("result"));
    expect(pendingResult).toEqual({ kind: "i32", value: 0n });

    const wakerValue = I.evaluate(AST.identifier("last_waker"));
    expect(wakerValue.kind).toBe("struct_instance");

    I.evaluate(
      AST.assignmentExpression(
        "=",
        AST.memberAccessExpression(AST.identifier("arm"), "ready"),
        AST.booleanLiteral(true),
      ),
    );

    I.evaluate(
      AST.functionCall(
        AST.memberAccessExpression(
          AST.memberAccessExpression(AST.identifier("arm"), "waker"),
          "wake",
        ),
        [],
      ),
    );

    (I as any).runFuture(handleVal);

    const finalResult = I.evaluate(AST.identifier("result"));
    expect(finalResult).toEqual({ kind: "i32", value: 42n });
  });

  test("await consumes iterable arms (iterator literal)", () => {
    const I = new Interpreter();

    const readyAwaitableStruct = AST.structDefinition(
      "ReadyAwaitable",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );
    I.evaluate(readyAwaitableStruct);

    const registrationStruct = AST.structDefinition("ReadyRegistration", [], "named");
    I.evaluate(registrationStruct);

    const readyMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("ReadyAwaitable"),
      [
        AST.functionDefinition(
          "is_ready",
          [],
          AST.blockExpression([AST.returnStatement(AST.booleanLiteral(true))]),
          AST.simpleTypeExpression("bool"),
          undefined,
          undefined,
          true,
        ),
        AST.functionDefinition(
          "register",
          [AST.functionParameter("waker", AST.simpleTypeExpression("AwaitWaker"))],
          AST.blockExpression([
            AST.returnStatement(AST.structLiteral([], false, "ReadyRegistration")),
          ]),
          AST.simpleTypeExpression("ReadyRegistration"),
          undefined,
          undefined,
          true,
        ),
        AST.functionDefinition(
          "commit",
          [],
          AST.blockExpression([AST.returnStatement(AST.implicitMemberExpression("value"))]),
          AST.simpleTypeExpression("i32"),
          undefined,
          undefined,
          true,
        ),
        AST.functionDefinition(
          "is_default",
          [],
          AST.blockExpression([AST.returnStatement(AST.booleanLiteral(false))]),
          AST.simpleTypeExpression("bool"),
          undefined,
          undefined,
          true,
        ),
      ],
    );
    I.evaluate(readyMethods);

    const registrationMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("ReadyRegistration"),
      [
        AST.functionDefinition(
          "cancel",
          [],
          AST.blockExpression([AST.returnStatement(AST.nilLiteral())]),
          AST.simpleTypeExpression("void"),
          undefined,
          undefined,
          true,
        ),
      ],
    );
    I.evaluate(registrationMethods);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("arm"),
        AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(7), "value")], false, "ReadyAwaitable"),
      ),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.integerLiteral(0)));

    const iterator = AST.iteratorLiteral([AST.yieldStatement(AST.identifier("arm"))]);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("arms"), iterator));

    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.identifier("arms")),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));

    const handleVal = I.evaluate(AST.identifier("handle"));
    while ((handleVal as any).state === "pending") {
      (I as any).runFuture(handleVal);
    }

    const final = I.evaluate(AST.identifier("result"));
    expect(final).toEqual({ kind: "i32", value: 7n });
  });

  test("await resolves Future handles", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.integerLiteral(0)));

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("fut"),
        AST.spawnExpression(
          AST.blockExpression([AST.integerLiteral(99)]),
        ),
      ),
    );

    const consumer = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([AST.identifier("fut")])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), consumer));

    const handleVal = I.evaluate(AST.identifier("handle")) as any;
    const futureVal = I.evaluate(AST.identifier("fut")) as any;
    (I as any).runFuture(handleVal);
    (I as any).runFuture(futureVal);
    (I as any).runFuture(handleVal);
    expect(futureVal.state).toBe("resolved");
    expect((handleVal as any).state).toBe("resolved");

    const final = I.evaluate(AST.identifier("result"));
    expect(final).toEqual({ kind: "i32", value: 99n });
  });

  test("await resolves Future handles", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.integerLiteral(0)));

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("worker"),
        AST.spawnExpression(
          AST.blockExpression([AST.integerLiteral(12)]),
        ),
      ),
    );

    const consumer = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([AST.identifier("worker")])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), consumer));

    const workerHandle = I.evaluate(AST.identifier("worker")) as any;
    const consumerHandle = I.evaluate(AST.identifier("handle")) as any;
    (I as any).runFuture(consumerHandle);
    (I as any).runFuture(workerHandle);
    (I as any).runFuture(consumerHandle);
    expect((workerHandle as any).state).toBe("resolved");
    expect((consumerHandle as any).state).toBe("resolved");

    const final = I.evaluate(AST.identifier("result"));
    expect(final).toEqual({ kind: "i32", value: 12n });
  });

  test("await rotates ready arms fairly", () => {
    const I = new Interpreter();

    const registrationDef = AST.structDefinition("ManualRegistration", [], "named");
    I.evaluate(registrationDef);
    const registrationMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("ManualRegistration"),
      [
        AST.functionDefinition(
          "cancel",
          [],
          AST.blockExpression([AST.returnStatement(AST.nilLiteral())]),
          AST.simpleTypeExpression("void"),
          undefined,
          undefined,
          true,
        ),
      ],
    );
    I.evaluate(registrationMethods);

    const readyAwaitableDef = AST.structDefinition(
      "ReadyAwaitable",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );
    I.evaluate(readyAwaitableDef);
    const readyAwaitableMethods = AST.methodsDefinition(
      AST.simpleTypeExpression("ReadyAwaitable"),
      [
        AST.functionDefinition(
          "is_ready",
          [],
          AST.blockExpression([AST.returnStatement(AST.booleanLiteral(true))]),
          AST.simpleTypeExpression("bool"),
          undefined,
          undefined,
          true,
        ),
        AST.functionDefinition(
          "register",
          [AST.functionParameter("waker", AST.simpleTypeExpression("AwaitWaker"))],
          AST.blockExpression([
            AST.returnStatement(AST.structLiteral([], false, "ManualRegistration")),
          ]),
          AST.simpleTypeExpression("ManualRegistration"),
          undefined,
          undefined,
          true,
        ),
        AST.functionDefinition(
          "commit",
          [],
          AST.blockExpression([AST.returnStatement(AST.implicitMemberExpression("value"))]),
          AST.simpleTypeExpression("i32"),
          undefined,
          undefined,
          true,
        ),
        AST.functionDefinition(
          "is_default",
          [],
          AST.blockExpression([AST.returnStatement(AST.booleanLiteral(false))]),
          AST.simpleTypeExpression("bool"),
          undefined,
          undefined,
          true,
        ),
      ],
    );
    I.evaluate(readyAwaitableMethods);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("arm1"),
        AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "value")], false, "ReadyAwaitable"),
      ),
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("arm2"),
        AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(2), "value")], false, "ReadyAwaitable"),
      ),
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("arms"),
        AST.arrayLiteral([AST.identifier("arm1"), AST.identifier("arm2")]),
      ),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("first"), AST.integerLiteral(0)));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("second"), AST.integerLiteral(0)));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("third"), AST.integerLiteral(0)));

    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("first"),
          AST.awaitExpression(AST.identifier("arms")),
        ),
        AST.assignmentExpression(
          "=",
          AST.identifier("second"),
          AST.awaitExpression(AST.identifier("arms")),
        ),
        AST.assignmentExpression(
          "=",
          AST.identifier("third"),
          AST.awaitExpression(AST.identifier("arms")),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));
    const handleVal = I.evaluate(AST.identifier("handle")) as any;
    while (handleVal.state === "pending") {
      (I as any).runFuture(handleVal);
    }

    const first = I.evaluate(AST.identifier("first"));
    const second = I.evaluate(AST.identifier("second"));
    const third = I.evaluate(AST.identifier("third"));
    expect(first).toEqual({ kind: "i32", value: 1n });
    expect(second).toEqual({ kind: "i32", value: 2n });
    expect(third).toEqual({ kind: "i32", value: 1n });
  });

  test("Await.default helper produces a default arm", async () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.stringLiteral("pending")));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("timer_hits"), AST.integerLiteral(0)));

    const timerArm = AST.functionCall(AST.identifier("__able_await_sleep_ms"), [
      AST.integerLiteral(10),
      AST.lambdaExpression([], AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("timer_hits"),
          AST.binaryExpression("+", AST.identifier("timer_hits"), AST.integerLiteral(1)),
        ),
        AST.stringLiteral("timer"),
      ])),
    ]);
    const defaultArm = AST.functionCall(AST.identifier("__able_await_default"), [
      AST.lambdaExpression([], AST.stringLiteral("fallback")),
    ]);
    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([timerArm, defaultArm])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));

    const handleVal = I.evaluate(AST.identifier("handle")) as any;
    (I as any).runFuture(handleVal);

    const result = I.evaluate(AST.identifier("result"));
    expect(result).toEqual({ kind: "String", value: "fallback" });

    await flushScheduler();
    await new Promise((resolve) => setTimeout(resolve, 15));
    I.evaluate(AST.functionCall(AST.identifier("future_flush"), []));

    const timerHits = I.evaluate(AST.identifier("timer_hits"));
    expect(timerHits).toEqual({ kind: "i32", value: 0n });
  });

  test("Await.sleep_ms awaitable wakes after the deadline", async () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.stringLiteral("pending")));

    const timerArm = AST.functionCall(AST.identifier("__able_await_sleep_ms"), [
      AST.integerLiteral(5),
      AST.lambdaExpression([], AST.stringLiteral("timer")),
    ]);
    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([timerArm])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));
    const handleVal = I.evaluate(AST.identifier("handle")) as any;

    (I as any).runFuture(handleVal);
    expect((handleVal as any).state).toBe("pending");

    await new Promise((resolve) => setTimeout(resolve, 15));
    I.evaluate(AST.functionCall(AST.identifier("future_flush"), []));

    expect((handleVal as any).state).toBe("resolved");
    const result = I.evaluate(AST.identifier("result"));
    expect(result).toEqual({ kind: "String", value: "timer" });
  });

  test("await cancellation cancels registrations and suppresses callbacks", async () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("hits"), AST.integerLiteral(0)));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.stringLiteral("pending")));

    const timerArm = AST.functionCall(AST.identifier("__able_await_sleep_ms"), [
      AST.integerLiteral(15),
      AST.lambdaExpression(
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("hits"),
            AST.binaryExpression("+", AST.identifier("hits"), AST.integerLiteral(1)),
          ),
          AST.stringLiteral("timer"),
        ]),
      ),
    ]);
    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([timerArm])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));

    const handleVal = I.evaluate(AST.identifier("handle")) as any;
    (I as any).runFuture(handleVal);
    expect(handleVal.state).toBe("pending");

    I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "cancel"), []));
    (I as any).runFuture(handleVal);

    await new Promise((resolve) => setTimeout(resolve, 30));
    I.evaluate(AST.functionCall(AST.identifier("future_flush"), []));

    const status = I.evaluate(
      AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "status"), []),
    ) as any;
    expect(status.kind).toBe("struct_instance");
    expect(status.def.id.name).toBe("Cancelled");

    const valueResult = I.evaluate(
      AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "value"), []),
    ) as any;
    expect(valueResult.kind).toBe("error");
    expect((valueResult.message as string).toLowerCase()).toContain("cancel");

    const hits = I.evaluate(AST.identifier("hits"));
    expect(hits).toEqual({ kind: "i32", value: 0n });
    const finalResult = I.evaluate(AST.identifier("result"));
    expect(finalResult).toEqual({ kind: "String", value: "pending" });
  });

  test("await channel awaitable registers + cancels on future cancellation", async () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("ch"), AST.functionCall(AST.identifier("__able_channel_new"), [AST.integerLiteral(1)])));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("hits"), AST.integerLiteral(0)));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.stringLiteral("pending")));

    const recvArm = AST.functionCall(AST.identifier("__able_channel_await_try_recv"), [
      AST.identifier("ch"),
      AST.lambdaExpression(
        [AST.functionParameter("v", null)],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("hits"),
            AST.binaryExpression("+", AST.identifier("hits"), AST.integerLiteral(1)),
          ),
          AST.identifier("v"),
        ]),
      ),
    ]);

    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([recvArm])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));
    const handleVal = I.evaluate(AST.identifier("handle")) as any;
    (I as any).runFuture(handleVal);
    expect(handleVal.state).toBe("pending");

    I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "cancel"), []));
    (I as any).runFuture(handleVal);

    I.evaluate(AST.functionCall(AST.identifier("__able_channel_send"), [AST.identifier("ch"), AST.stringLiteral("payload")]));
    await new Promise((resolve) => setTimeout(resolve, 10));
    I.evaluate(AST.functionCall(AST.identifier("future_flush"), []));

    const status = I.evaluate(
      AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "status"), []),
    ) as any;
    expect(status.def.id.name).toBe("Cancelled");

    const valueResult = I.evaluate(
      AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "value"), []),
    ) as any;
    expect(valueResult.kind).toBe("error");

    const hits = I.evaluate(AST.identifier("hits"));
    expect(hits).toEqual({ kind: "i32", value: 0n });
    const finalResult = I.evaluate(AST.identifier("result"));
    expect(finalResult).toEqual({ kind: "String", value: "pending" });
  });

  test("await channel send awaitable cancels pending send on cancellation", () => {
    const I = new Interpreter();
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("ch"), AST.functionCall(AST.identifier("__able_channel_new"), [AST.integerLiteral(0)])));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("hits"), AST.integerLiteral(0)));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("result"), AST.stringLiteral("pending")));

    const sendArm = AST.functionCall(AST.identifier("__able_channel_await_try_send"), [
      AST.identifier("ch"),
      AST.stringLiteral("payload"),
      AST.lambdaExpression(
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("hits"),
            AST.binaryExpression("+", AST.identifier("hits"), AST.integerLiteral(1)),
          ),
          AST.stringLiteral("sent"),
        ]),
      ),
    ]);

    const futureExpr = AST.spawnExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([sendArm])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), futureExpr));
    const handleVal = I.evaluate(AST.identifier("handle")) as any;
    (I as any).runFuture(handleVal);
    expect(handleVal.state).toBe("pending");

    I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "cancel"), []));
    (I as any).runFuture(handleVal);

    const tryRecv = I.evaluate(AST.functionCall(AST.identifier("__able_channel_try_receive"), [AST.identifier("ch")]));
    expect(tryRecv).toEqual({ kind: "nil", value: null });

    const status = I.evaluate(
      AST.functionCall(AST.memberAccessExpression(AST.identifier("handle"), "status"), []),
    ) as any;
    expect(status.def.id.name).toBe("Cancelled");

    const hits = I.evaluate(AST.identifier("hits"));
    expect(hits).toEqual({ kind: "i32", value: 0n });
    const finalResult = I.evaluate(AST.identifier("result"));
    expect(finalResult).toEqual({ kind: "String", value: "pending" });
  });
});
