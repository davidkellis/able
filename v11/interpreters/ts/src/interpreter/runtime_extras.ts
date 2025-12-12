import fs from "node:fs";

import * as AST from "../ast";
import { Environment } from "./environment";
import { InterpreterV10 } from "./index";
import { BreakLabelSignal, RaiseSignal } from "./signals";
import { makeIntegerValue } from "./numeric";
import type { V10Value } from "./values";

type ExternHandler = (
  ctx: InterpreterV10,
  extern: AST.ExternFunctionBody,
  arity: number,
) => Extract<V10Value, { kind: "native_function" }>;

const externHandlers: Partial<Record<AST.HostTarget, Record<string, ExternHandler>>> = {
  typescript: {
    now_nanos: (ctx, _extern, arity) =>
      ctx.makeNativeFunction("now_nanos", arity, () => {
        return makeIntegerValue("i64", 1_234_567_890_123_456n);
      }),
    read_text: (ctx, _extern, arity) =>
      ctx.makeNativeFunction("read_text", arity, (interp, args) => {
        const pathVal = args[0];
        if (!pathVal || pathVal.kind !== "String") {
          throw new Error("read_text expects a String path");
        }
        try {
          const data = fs.readFileSync(pathVal.value, "utf8");
          return { kind: "String", value: data };
        } catch (error) {
          const message = error instanceof Error ? error.message : String(error);
          throw new RaiseSignal(interp.makeRuntimeError(message));
        }
      }),
  },
};

export function registerExternHandler(target: AST.HostTarget, name: string, handler: ExternHandler): void {
  if (!externHandlers[target]) {
    externHandlers[target] = {};
  }
  externHandlers[target]![name] = handler;
}

export function evaluateStringInterpolation(ctx: InterpreterV10, node: AST.StringInterpolation, env: Environment): V10Value {
  let out = "";
  for (const part of node.parts) {
    if (part.type === "StringLiteral") out += part.value;
    else {
      const val = ctx.evaluate(part, env);
      out += ctx.valueToStringWithEnv(val, env);
    }
  }
  return { kind: "String", value: out };
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
    awaitBlocked: false,
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

export function evaluateExternFunctionBody(ctx: InterpreterV10, node: AST.ExternFunctionBody): V10Value {
  const name = node.signature?.id?.name;
  if (!name) {
    return { kind: "nil", value: null };
  }
  if (node.target !== "typescript") {
    return { kind: "nil", value: null };
  }
  const arity = Array.isArray(node.signature.params) ? node.signature.params.length : 0;
  let binding: V10Value | null = null;
  try {
    binding = ctx.globals.get(name);
  } catch {
    // ignore; we'll install a handler or stub below
  }

  if (!binding) {
    const handler = externHandlers[node.target]?.[name];
    if (handler) {
      binding = handler(ctx, node, arity);
    }
  }

  if (!binding) {
    binding = ctx.makeNativeFunction(name, arity, () => {
      throw new RaiseSignal(ctx.makeRuntimeError(`extern function ${name} for ${node.target} is not implemented`));
    });
  }

  if (!ctx.globals.has(name)) {
    ctx.globals.define(name, binding);
  }
  const qualified = ctx.qualifiedName(name);
  if (qualified && !ctx.globals.has(qualified)) {
    ctx.globals.define(qualified, binding);
  }
  ctx.registerSymbol(name, binding);
  return binding;
}
