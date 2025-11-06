import { expect } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";
import type { V10Value } from "../../src/interpreter";

export const flushScheduler = () => new Promise<void>(resolve => setTimeout(resolve, 0));

export const expectErrorValue = (value: V10Value) => {
  expect(value.kind).toBe("error");
  if (value.kind !== "error") throw new Error("expected error value");
  return value;
};

export const expectStructInstance = (value: V10Value | undefined, structName: string) => {
  expect(value && value.kind === "struct_instance" && value.def.id.name === structName).toBe(true);
  if (!value || value.kind !== "struct_instance" || value.def.id.name !== structName) {
    throw new Error(`expected struct_instance ${structName}`);
  }
  return value;
};

export const appendToTrace = (literal: string) =>
  AST.assignmentExpression(
    "=",
    AST.identifier("trace"),
    AST.binaryExpression("+", AST.identifier("trace"), AST.stringLiteral(literal)),
  );

export const drainScheduler = (interp: InterpreterV10, maxTicks = 32) => {
  const runtime = interp as { executor?: { flush?: (limit?: number) => void; pendingTasks?: () => number } };
  const executor = runtime.executor;
  if (!executor || typeof executor.flush !== "function") {
    throw new Error("executor with flush() required for drainScheduler");
  }
  let ticks = 0;
  while (true) {
    executor.flush();
    const pending = executor.pendingTasks?.() ?? 0;
    if (pending === 0) {
      break;
    }
    ticks++;
    if (ticks > maxTicks) {
      throw new Error(`scheduler stuck; pending tasks: ${pending}`);
    }
  }
};
