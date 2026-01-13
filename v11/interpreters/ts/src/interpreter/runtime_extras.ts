import * as AST from "../ast";
import { Environment } from "./environment";
import { Interpreter } from "./index";
import { BreakLabelSignal, GeneratorYieldSignal, ProcYieldSignal, RaiseSignal } from "./signals";
import type { RuntimeValue } from "./values";
import type { ContinuationContext } from "./continuations";

export function evaluateStringInterpolation(ctx: Interpreter, node: AST.StringInterpolation, env: Environment): RuntimeValue {
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateStringInterpolationWithContinuation(ctx, node, env, generator);
  }
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateStringInterpolationWithContinuation(ctx, node, env, procContext);
  }
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

function evaluateStringInterpolationWithContinuation(
  ctx: Interpreter,
  node: AST.StringInterpolation,
  env: Environment,
  continuation: ContinuationContext,
): RuntimeValue {
  let state = continuation.getStringInterpolationState(node);
  if (!state) {
    state = { index: 0, output: "" };
    continuation.setStringInterpolationState(node, state);
  }

  while (state.index < node.parts.length) {
    const part = node.parts[state.index]!;
    if (part.type === "StringLiteral") {
      state.output += part.value;
      state.index += 1;
      continue;
    }
    try {
      const val = ctx.evaluate(part, env);
      state.output += ctx.valueToStringWithEnv(val, env);
      state.index += 1;
    } catch (err) {
      if (isContinuationYield(continuation, err)) {
        continuation.markStatementIncomplete();
      } else {
        continuation.clearStringInterpolationState(node);
      }
      throw err;
    }
  }

  continuation.clearStringInterpolationState(node);
  return { kind: "String", value: state.output };
}

function isContinuationYield(context: ContinuationContext, err: unknown): boolean {
  if (context.kind === "generator") {
    return err instanceof GeneratorYieldSignal || err instanceof ProcYieldSignal;
  }
  return err instanceof ProcYieldSignal;
}

export function evaluateBreakpointExpression(ctx: Interpreter, node: AST.BreakpointExpression, env: Environment): RuntimeValue {
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

export function evaluateProcExpression(ctx: Interpreter, node: AST.ProcExpression, env: Environment): RuntimeValue {
  const capturedEnv = new Environment(env);
  const handle: Extract<RuntimeValue, { kind: "proc_handle" }> = {
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

export function evaluateSpawnExpression(ctx: Interpreter, node: AST.SpawnExpression, env: Environment): RuntimeValue {
  const capturedEnv = new Environment(env);
  const future: Extract<RuntimeValue, { kind: "future" }> = {
    kind: "future",
    state: "pending",
    expression: node.expression,
    env: capturedEnv,
    runner: null,
    cancelRequested: false,
    hasStarted: false,
    awaitBlocked: false,
  };
  future.runner = () => ctx.runFuture(future);
  ctx.scheduleAsync(future.runner);
  return future;
}

export function evaluateExternFunctionBody(ctx: Interpreter, node: AST.ExternFunctionBody): RuntimeValue {
  const name = node.signature?.id?.name;
  if (!name) {
    return { kind: "nil", value: null };
  }
  if (node.target !== "typescript") {
    return { kind: "nil", value: null };
  }
  if (!node.body.trim() && !ctx.isKernelExtern(name)) {
    throw new RaiseSignal(ctx.makeRuntimeError(`extern function ${name} for ${node.target} must provide a host body`));
  }
  const arity = Array.isArray(node.signature.params) ? node.signature.params.length : 0;
  let binding: RuntimeValue | null = null;
  try {
    binding = ctx.globals.get(name);
  } catch {
    // ignore; we'll install a handler or stub below
  }

  if (!binding) {
    const pkgName = ctx.currentPackage ?? "<root>";
    binding = ctx.makeNativeFunction(name, arity, (interp, args) =>
      interp.invokeExternHostFunction(pkgName, node, args),
    );
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
