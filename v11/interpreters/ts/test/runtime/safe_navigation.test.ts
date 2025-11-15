import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("runtime safe navigation", () => {
  test("returns nil when receiver is nil", () => {
    const interpreter = new InterpreterV10();
    interpreter.evaluate(AST.assign("user", AST.nilLiteral()));
    const result = interpreter.evaluate(
      AST.memberAccessExpression(AST.identifier("user"), "profile", { isSafe: true }),
    );
    expect(result.kind).toBe("nil");
  });

  test("short-circuits method call arguments when receiver is nil", () => {
    const interpreter = new InterpreterV10();
    interpreter.evaluate(AST.assign("calls", AST.int(0)));
    interpreter.evaluate(
      AST.functionDefinition(
        "trigger",
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("calls"),
            AST.binaryExpression("+", AST.identifier("calls"), AST.int(1)),
          ),
          AST.identifier("calls"),
        ]),
        AST.simpleTypeExpression("i32"),
      ),
    );
    interpreter.evaluate(AST.assign("wrapper", AST.nilLiteral()));
    const safeCall = AST.functionCall(
      AST.memberAccessExpression(AST.identifier("wrapper"), "call", { isSafe: true }),
      [AST.functionCall(AST.identifier("trigger"), [])],
    );
    const result = interpreter.evaluate(safeCall);
    expect(result.kind).toBe("nil");
    const callsValue = interpreter.evaluate(AST.identifier("calls"));
    expect(callsValue.kind).toBe("i32");
    expect(callsValue.value).toBe(0n);
  });
});
