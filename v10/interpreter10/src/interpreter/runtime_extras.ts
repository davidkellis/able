import * as AST from "../ast";
import { Environment } from "./environment";
import { InterpreterV10 } from "./index";
import { BreakLabelSignal } from "./signals";
import type { V10Value } from "./values";

export function evaluateStringInterpolation(ctx: InterpreterV10, node: AST.StringInterpolation, env: Environment): V10Value {
  let out = "";
  for (const part of node.parts) {
    if (part.type === "StringLiteral") out += part.value;
    else {
      const val = ctx.evaluate(part, env);
      out += ctx.valueToStringWithEnv(val, env);
    }
  }
  return { kind: "string", value: out };
}

export function evaluateBreakpointExpression(ctx: InterpreterV10, node: AST.BreakpointExpression, env: Environment): V10Value {
  ctx.breakpointStack.push(node.label.name);
  try {
    return ctx.evaluate(node.body, env);
  } catch (e) {
    if (e instanceof BreakLabelSignal) {
      if (e.label === node.label.name) return e.value;
      throw e;
    }
    throw e;
  } finally {
    ctx.breakpointStack.pop();
  }
}

export function evaluateProcExpression(ctx: InterpreterV10, node: AST.ProcExpression, env: Environment): V10Value {
  const capturedEnv = new Environment(env);
  const handle: Extract<V10Value, { kind: "proc_handle" }> = {
    kind: "proc_handle",
    state: "pending",
    expression: node.expression,
    env: capturedEnv,
    runner: null,
    cancelRequested: false,
  };
  handle.runner = () => ctx.runProcHandle(handle);
  ctx.scheduleAsync(handle.runner);
  return handle;
}

export function evaluateSpawnExpression(ctx: InterpreterV10, node: AST.SpawnExpression, env: Environment): V10Value {
  const capturedEnv = new Environment(env);
  const future: Extract<V10Value, { kind: "future" }> = {
    kind: "future",
    state: "pending",
    expression: node.expression,
    env: capturedEnv,
    runner: null,
    cancelRequested: false,
    hasStarted: false,
  };
  future.runner = () => ctx.runFuture(future);
  ctx.scheduleAsync(future.runner);
  return future;
}
