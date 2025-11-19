import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - await expression", () => {
  test("await resolves manual awaitable once waker fires", () => {
    const I = new InterpreterV10();

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

    const awaitProc = AST.procExpression(
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.identifier("result"),
          AST.awaitExpression(AST.arrayLiteral([AST.identifier("arm")])),
        ),
      ]),
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("handle"), awaitProc));

    const handleVal = I.evaluate(AST.identifier("handle"));
    (I as any).runProcHandle(handleVal);

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

    (I as any).runProcHandle(handleVal);

    const finalResult = I.evaluate(AST.identifier("result"));
    expect(finalResult).toEqual({ kind: "i32", value: 42n });
  });
});

