import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import type { RuntimeValue } from "../../src/interpreter";

import { appendToTrace, drainScheduler, expectErrorValue, expectStructInstance, flushScheduler } from "./proc_spawn.helpers";

describe("v11 interpreter - proc & spawn handles", () => {
  test("future resolves after scheduler tick", async () => {
    const I = new Interpreter();

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
    expect(I.evaluate(valueCall)).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 1n });
  });

  test("proc yield allows interleaving tasks", async () => {
    const I = new Interpreter();

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
        AST.spawnExpression(
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
                AST.functionCall(AST.identifier("future_yield"), []),
              ]),
              [],
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
            AST.integerLiteral(1),
          ])
        )
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("fast"),
        AST.spawnExpression(
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
    expect(slowValue).toEqual({ kind: "i32", value: 1n });

    const traceVal = I.evaluate(AST.identifier("trace"));
    expect(traceVal).toEqual({ kind: "String", value: "ABC" });

    const slowStatus = I.evaluate(slowStatusCall) as any;
    expect(slowStatus.kind).toBe("struct_instance");
    expect(slowStatus.def.id.name).toBe("Resolved");

    const fastStatus = I.evaluate(fastStatusCall) as any;
    expect(fastStatus.kind).toBe("struct_instance");
    expect(fastStatus.def.id.name).toBe("Resolved");
  });

  test("proc and future interleave across multiple yields", () => {
    const I = new Interpreter();

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
        AST.identifier("worker_stage"),
        AST.integerLiteral(0)
      )
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("future_stage"),
        AST.integerLiteral(0)
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("worker"),
        AST.spawnExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("worker_stage"),
                AST.integerLiteral(0)
              ),
              AST.blockExpression([
                appendToTrace("A1"),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("worker_stage"),
                  AST.integerLiteral(1)
                ),
                AST.functionCall(AST.identifier("future_yield"), []),
                AST.integerLiteral(0),
              ]),
              [
                AST.elseIfClause(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("worker_stage"),
                    AST.integerLiteral(1)
                  ),
                  AST.blockExpression([
                    appendToTrace("A2"),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("worker_stage"),
                      AST.integerLiteral(2)
                    ),
                    AST.functionCall(AST.identifier("future_yield"), []),
                    AST.integerLiteral(0),
                  ]),
                ),
                AST.elseIfClause(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("worker_stage"),
                    AST.integerLiteral(2)
                  ),
                  AST.blockExpression([
                    appendToTrace("A3"),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("worker_stage"),
                      AST.integerLiteral(3)
                    ),
                    AST.integerLiteral(0),
                  ]),
                ),
              ]
            ),
            AST.integerLiteral(0),
          ])
        )
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("future"),
        AST.spawnExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("future_stage"),
                AST.integerLiteral(0)
              ),
              AST.blockExpression([
                appendToTrace("B1"),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("future_stage"),
                  AST.integerLiteral(1)
                ),
                AST.functionCall(AST.identifier("future_yield"), []),
                AST.integerLiteral(0),
              ]),
              [
                AST.elseIfClause(
                  AST.binaryExpression(
                    "==",
                    AST.identifier("future_stage"),
                    AST.integerLiteral(1)
                  ),
                  AST.blockExpression([
                    appendToTrace("B2"),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("future_stage"),
                      AST.integerLiteral(2)
                    ),
                    AST.integerLiteral(0),
                  ]),
                ),
              ]
            ),
            AST.integerLiteral(0),
          ])
        )
      )
    );

    drainScheduler(I);

    const traceVal = I.evaluate(AST.identifier("trace"));
    expect(traceVal).toEqual({ kind: "String", value: "A1B1A2B2A3" });

    const workerStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("worker"), "status"),
      []
    );
    const futureStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future"), "status"),
      []
    );
    const workerValueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("worker"), "value"),
      []
    );
    const futureValueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future"), "value"),
      []
    );

    const workerStatus = I.evaluate(workerStatusCall) as any;
    expect(workerStatus.kind).toBe("struct_instance");
    expect(workerStatus.def.id.name).toBe("Resolved");

    const futureStatus = I.evaluate(futureStatusCall) as any;
    expect(futureStatus.kind).toBe("struct_instance");
    expect(futureStatus.def.id.name).toBe("Resolved");

    expect(I.evaluate(workerValueCall)).toEqual({ kind: "i32", value: 0n });
    expect(I.evaluate(futureValueCall)).toEqual({ kind: "i32", value: 0n });
  });

  test("future_pending_tasks reports queued cooperative work", () => {
    const I = new Interpreter();
    const pendingCall = () => AST.functionCall(AST.identifier("future_pending_tasks"), []);

    const initial = I.evaluate(pendingCall()) as RuntimeValue;
    expect(initial).toEqual({ kind: "i32", value: 0n });

    I.evaluate(
      AST.spawnExpression(
        AST.blockExpression([
          AST.integerLiteral(1),
        ]),
      ),
    );

    const pendingAfterSpawn = I.evaluate(pendingCall()) as RuntimeValue;
    expect(pendingAfterSpawn.kind).toBe("i32");
    if (pendingAfterSpawn.kind !== "i32") throw new Error("expected integer result");
    expect(pendingAfterSpawn.value).toBeGreaterThan(0);

    let drained = false;
    for (let attempt = 0; attempt < 16; attempt += 1) {
      I.evaluate(AST.functionCall(AST.identifier("future_flush"), []));
      const pendingAfterFlush = I.evaluate(pendingCall()) as RuntimeValue;
      expect(pendingAfterFlush.kind).toBe("i32");
      if (pendingAfterFlush.kind !== "i32") throw new Error("expected integer result");
      if (pendingAfterFlush.value === 0) {
        drained = true;
        break;
      }
    }
    if (!drained) {
      drainScheduler(I);
    }
    const finalPending = I.evaluate(pendingCall()) as RuntimeValue;
    expect(finalPending.kind).toBe("i32");
    if (finalPending.kind !== "i32") throw new Error("expected integer result");
    expect(finalPending.value).toBe(0n);
  });

  test("proc awaiting future with nested yields resolves cleanly", () => {
    const I = new Interpreter();

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
        AST.identifier("future_handle"),
        AST.nilLiteral()
      )
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("outer_stage"),
        AST.integerLiteral(0)
      )
    );
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("inner_stage"),
        AST.integerLiteral(0)
      )
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("outer"),
        AST.spawnExpression(
          AST.blockExpression([
            AST.ifExpression(
              AST.binaryExpression(
                "==",
                AST.identifier("outer_stage"),
                AST.integerLiteral(0)
              ),
              AST.blockExpression([
                appendToTrace("A"),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("future_handle"),
                  AST.spawnExpression(
                    AST.blockExpression([
                      AST.ifExpression(
                        AST.binaryExpression(
                          "==",
                          AST.identifier("inner_stage"),
                          AST.integerLiteral(0)
                        ),
                        AST.blockExpression([
                          appendToTrace("B"),
                          AST.assignmentExpression(
                            "=",
                            AST.identifier("inner_stage"),
                            AST.integerLiteral(1)
                          ),
                          AST.functionCall(AST.identifier("future_yield"), []),
                          AST.integerLiteral(7),
                        ]),
                        [
                          AST.elseIfClause(
                            AST.binaryExpression(
                              "==",
                              AST.identifier("inner_stage"),
                              AST.integerLiteral(1)
                            ),
                            AST.blockExpression([
                              appendToTrace("D"),
                              AST.assignmentExpression(
                                "=",
                                AST.identifier("inner_stage"),
                                AST.integerLiteral(2)
                              ),
                              AST.integerLiteral(7),
                            ]),
                          ),
                        ]
                      ),
                      AST.integerLiteral(7),
                    ])
                  )
                ),
              AST.assignmentExpression(
                "=",
                AST.identifier("outer_stage"),
                AST.integerLiteral(1)
              ),
              AST.functionCall(AST.identifier("future_yield"), []),
              AST.integerLiteral(0),
            ]),
            [
              AST.elseIfClause(
                AST.binaryExpression(
                  "==",
                  AST.identifier("outer_stage"),
                  AST.integerLiteral(1)
                ),
                AST.blockExpression([
                  appendToTrace("C"),
                  AST.assignmentExpression(
                    ":=",
                    AST.identifier("result"),
                    AST.functionCall(
                      AST.memberAccessExpression(AST.identifier("future_handle"), "value"),
                      []
                    )
                  ),
                  AST.assignmentExpression(
                    "=",
                    AST.identifier("outer_stage"),
                    AST.integerLiteral(2)
                  ),
                  AST.identifier("result"),
                ]),
              ),
            ]
          ),
        ])
      )
    )
  );

    drainScheduler(I);

    const traceVal = I.evaluate(AST.identifier("trace"));
    expect(traceVal).toEqual({ kind: "String", value: "ABCD" });

    const futureStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("future_handle"), "status"),
      []
    );
    const outerStatusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("outer"), "status"),
      []
    );
    const outerValueCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("outer"), "value"),
      []
    );

    const futureStatus = I.evaluate(futureStatusCall) as any;
    expect(futureStatus.kind).toBe("struct_instance");
    expect(futureStatus.def.id.name).toBe("Resolved");

    const outerStatus = I.evaluate(outerStatusCall) as any;
    expect(outerStatus.kind).toBe("struct_instance");
    expect(outerStatus.def.id.name).toBe("Resolved");

    const outerValue = I.evaluate(outerValueCall);
    expect(outerValue).toEqual({ kind: "i32", value: 7n });
  });

});
