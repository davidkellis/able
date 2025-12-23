import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import { RaiseSignal } from "../../src/interpreter/signals";

const emptyBlock = AST.blockExpression([]);

describe("v11 interpreter - extern handling", () => {
  test("registers handled externs in globals and package registry", () => {
    const interpreter = new Interpreter();
    const signature = AST.functionDefinition("now_nanos", [], emptyBlock, AST.simpleTypeExpression("i64"));
    const mod = AST.module([AST.externFunctionBody("typescript", signature, "")], [], AST.packageStatement(["host"]));

    interpreter.evaluate(mod);

    const pkg = interpreter.packageRegistry.get("host");
    expect(pkg?.get("now_nanos")).toBeDefined();
    const qualified = interpreter.globals.get("host.now_nanos");
    expect(qualified.kind).toBe("native_function");
  });

  test("preserves existing bindings when an extern is already provided", () => {
    const interpreter = new Interpreter();
    const existing = interpreter.makeNativeFunction("now_nanos", 0, () => ({ kind: "i32", value: 1n }));
    interpreter.globals.define("now_nanos", existing);

    const signature = AST.functionDefinition("now_nanos", [], emptyBlock, AST.simpleTypeExpression("i64"));
    interpreter.evaluate(AST.externFunctionBody("typescript", signature, ""));

    expect(interpreter.globals.get("now_nanos")).toBe(existing);
  });

  test("installs a stub for unknown externs", () => {
    const interpreter = new Interpreter();
    const signature = AST.functionDefinition("not_impl", [], emptyBlock, AST.simpleTypeExpression("i64"));
    interpreter.evaluate(AST.externFunctionBody("typescript", signature, ""));
    const call = AST.functionCall(AST.identifier("not_impl"), []);

    try {
      interpreter.evaluate(call);
      throw new Error("expected stub call to throw");
    } catch (err) {
      expect(err).toBeInstanceOf(RaiseSignal);
      if (err instanceof RaiseSignal && err.value.kind === "error") {
        expect(err.value.message).toMatch(/not implemented/i);
      }
    }
  });
});
