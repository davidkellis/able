import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import { RaiseSignal } from "../../src/interpreter/signals";

const emptyBlock = AST.blockExpression([]);

describe("v11 interpreter - extern handling", () => {
  test("executes extern host bodies via the host module", async () => {
    const interpreter = new Interpreter();
    const signature = AST.functionDefinition("now_nanos", [], emptyBlock, AST.simpleTypeExpression("i64"));
    const mod = AST.module(
      [AST.externFunctionBody("typescript", signature, "return 42")],
      [],
      AST.packageStatement(["host"]),
    );

    interpreter.evaluate(mod);

    const pkg = interpreter.packageRegistry.get("host");
    expect(pkg?.get("now_nanos")).toBeDefined();
    const qualified = interpreter.globals.get("host.now_nanos");
    expect(qualified.kind).toBe("native_function");

    const call = AST.functionCall(AST.identifier("now_nanos"), []);
    const result = await interpreter.evaluateAsTask(call);
    expect(result).toEqual({ kind: "i64", value: 42n });
  });

  test("preserves existing bindings when an extern is already provided", () => {
    const interpreter = new Interpreter();
    const existing = interpreter.makeNativeFunction("now_nanos", 0, () => ({ kind: "i32", value: 1n }));
    interpreter.globals.define("now_nanos", existing);

    const signature = AST.functionDefinition("now_nanos", [], emptyBlock, AST.simpleTypeExpression("i64"));
    interpreter.evaluate(AST.externFunctionBody("typescript", signature, "return 0"));

    expect(interpreter.globals.get("now_nanos")).toBe(existing);
  });

  test("rejects empty extern bodies for non-kernel symbols", () => {
    const interpreter = new Interpreter();
    const signature = AST.functionDefinition("not_impl", [], emptyBlock, AST.simpleTypeExpression("i64"));
    try {
      interpreter.evaluate(AST.externFunctionBody("typescript", signature, ""));
      throw new Error("expected extern evaluation to throw");
    } catch (err) {
      expect(err).toBeInstanceOf(RaiseSignal);
      if (err instanceof RaiseSignal && err.value.kind === "error") {
        expect(err.value.message).toMatch(/must provide a host body/i);
      }
    }
  });
});
