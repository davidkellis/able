import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";
import type { V10Value } from "../../src/interpreter";

import { appendToTrace, drainScheduler, expectErrorValue, expectStructInstance, flushScheduler } from "./proc_spawn.helpers";

describe("v11 interpreter - proc & spawn handles", () => {
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
              [],
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
            AST.integerLiteral(0),
          ])
        )
      )
    );

    const handleVal = I.evaluate(AST.identifier("handle"));
    (I as any).runProcHandle(handleVal);
    expect(I.evaluate(AST.identifier("trace"))).toEqual({ kind: "String", value: "w" });

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
    expect(trace.kind).toBe("String");
    expect(trace.value).toBe("wx");
  });

  test("automatic time slicing yields progress without explicit proc_yield", () => {
    const I = new InterpreterV10({ schedulerMaxSteps: 4 });

    I.evaluate(AST.assignmentExpression(":=", AST.identifier("counter"), AST.integerLiteral(0)));
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("done"), AST.booleanLiteral(false)));

    const handleProc = AST.procExpression(
      AST.blockExpression([
        AST.assignmentExpression(":=", AST.identifier("i"), AST.integerLiteral(0)),
        AST.whileLoop(
          AST.binaryExpression("<", AST.identifier("i"), AST.integerLiteral(10)),
          AST.blockExpression([
            AST.assignmentExpression("=", AST.identifier("counter"), AST.identifier("i")),
            AST.assignmentExpression(
              "=",
              AST.identifier("i"),
              AST.binaryExpression("+", AST.identifier("i"), AST.integerLiteral(1))
            ),
          ])
        ),
        AST.assignmentExpression("=", AST.identifier("done"), AST.booleanLiteral(true)),
      ])
    );

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("slice_handle"),
        handleProc,
      ),
    );

    const flushCall = AST.functionCall(AST.identifier("proc_flush"), []);
    I.evaluate(flushCall);

    const doneValue = I.evaluate(AST.identifier("done")) as any;
    expect(doneValue).toEqual({ kind: "bool", value: false });

    const counterValue = I.evaluate(AST.identifier("counter")) as any;
    expect(counterValue.kind).toBe("i32");
    expect(counterValue.value).toBeGreaterThan(0);
    expect(counterValue.value).toBeLessThan(10);

    const statusCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("slice_handle"), "status"),
      [],
    );

    for (let n = 0; n < 5; n += 1) {
      I.evaluate(flushCall);
      const stillPending = I.evaluate(statusCall) as any;
      if (stillPending.def.id.name === "Resolved") {
        break;
      }
    }

    const finalStatus = I.evaluate(statusCall) as any;
    expect(finalStatus.def.id.name).toBe("Resolved");

    const finalDone = I.evaluate(AST.identifier("done")) as any;
    expect(finalDone).toEqual({ kind: "bool", value: true });

    const finalCounter = I.evaluate(AST.identifier("counter")) as any;
    expect(finalCounter.kind).toBe("i32");
    expect(finalCounter.value).toBe(9n);
  });
});
