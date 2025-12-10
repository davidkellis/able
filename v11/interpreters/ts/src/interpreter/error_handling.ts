import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { callCallableValue } from "./functions";
import { RaiseSignal } from "./signals";
import { memberAccessOnValue } from "./structs";
import type { V10Value } from "./values";

export function evaluateRaiseStatement(ctx: InterpreterV10, node: AST.RaiseStatement, env: Environment): never {
  const val = ctx.evaluate(node.expression, env);
  const err: V10Value = val.kind === "error" ? val : { kind: "error", message: ctx.valueToString(val), value: val };
  ctx.raiseStack.push(err);
  try {
    throw new RaiseSignal(err);
  } finally {
    ctx.raiseStack.pop();
  }
}

export function evaluateRescueExpression(ctx: InterpreterV10, node: AST.RescueExpression, env: Environment): V10Value {
  try {
    return ctx.evaluate(node.monitoredExpression, env);
  } catch (e) {
    if (e instanceof RaiseSignal) {
      ctx.raiseStack.push(e.value);
      for (const clause of node.clauses) {
        const matchEnv = ctx.tryMatchPattern(clause.pattern, e.value, env);
        if (matchEnv) {
          if (clause.guard) {
            const g = ctx.evaluate(clause.guard, matchEnv);
            if (!ctx.isTruthy(g)) continue;
          }
          try {
            return ctx.evaluate(clause.body, matchEnv);
          } finally {
            ctx.raiseStack.pop();
          }
        }
      }
      ctx.raiseStack.pop();
      throw e;
    }
    throw e;
  }
}

export function evaluateOrElseExpression(ctx: InterpreterV10, node: AST.OrElseExpression, env: Environment): V10Value {
  try {
    const value = ctx.evaluate(node.expression, env);
    const failure = classifyOptionOrResultFailure(ctx, value);
    if (!failure) {
      return value;
    }
    const handlerEnv = new Environment(env);
    if (failure.kind === "error" && node.errorBinding) {
      handlerEnv.define(node.errorBinding.name, failure.value);
    }
    return ctx.evaluate(node.handler, handlerEnv);
  } catch (e) {
    if (e instanceof RaiseSignal) {
      const handlerEnv = new Environment(env);
      if (node.errorBinding) handlerEnv.define(node.errorBinding.name, e.value);
      ctx.raiseStack.push(e.value);
      try {
        return ctx.evaluate(node.handler, handlerEnv);
      } finally {
        ctx.raiseStack.pop();
      }
    }
    throw e;
  }
}

type FailureKind = { kind: "nil" } | { kind: "error"; value: V10Value };

function classifyOptionOrResultFailure(ctx: InterpreterV10, value: V10Value): FailureKind | null {
  if (value.kind === "nil") return { kind: "nil" };
  if (value.kind === "error") return { kind: "error", value };
  if (value.kind === "interface_value" && value.interfaceName === "Error") {
    return { kind: "error", value: value.value };
  }
  const typeName = ctx.getTypeNameForValue(value);
  if (!typeName) return null;
  const typeArgs = value.kind === "struct_instance" ? value.typeArguments : undefined;
  if (ctx.typeImplementsInterface(typeName, "Error", typeArgs)) {
    return { kind: "error", value };
  }
  return null;
}

export function evaluatePropagationExpression(ctx: InterpreterV10, node: AST.PropagationExpression, env: Environment): V10Value {
  try {
    const val = ctx.evaluate(node.expression, env);
    const errVal = coerceToErrorValue(ctx, val, env);
    if (errVal) throw new RaiseSignal(errVal);
    return val;
  } catch (e) {
    if (e instanceof RaiseSignal) throw e;
    throw e;
  }
}

export function evaluateEnsureExpression(ctx: InterpreterV10, node: AST.EnsureExpression, env: Environment): V10Value {
  let result: V10Value | null = null;
  let caught: RaiseSignal | null = null;
  try {
    result = ctx.evaluate(node.tryExpression, env);
  } catch (e) {
    if (e instanceof RaiseSignal) caught = e; else throw e;
  } finally {
    ctx.evaluate(node.ensureBlock, env);
  }
  if (caught) throw caught;
  return result ?? { kind: "nil", value: null };
}

export function evaluateRethrowStatement(ctx: InterpreterV10, _node: AST.RethrowStatement): never {
  const err = ctx.raiseStack[ctx.raiseStack.length - 1] || { kind: "error", message: "Unknown rethrow" } as V10Value;
  throw new RaiseSignal(err);
}

function coerceToErrorValue(ctx: InterpreterV10, val: V10Value, env: Environment): Extract<V10Value, { kind: "error" }> | null {
  if (val.kind === "error") return val;
  if (val.kind === "interface_value" && val.interfaceName === "Error" && val.value.kind === "error") {
    return val.value;
  }
  const typeName = ctx.getTypeNameForValue(val);
  const implementsError = typeName
    ? ctx.typeImplementsInterface(typeName, "Error", val.kind === "struct_instance" ? val.typeArguments : undefined)
    : val.kind === "interface_value" && val.interfaceName === "Error";
  if (!implementsError) return null;

  let errorIface: V10Value = val;
  if (val.kind !== "interface_value" || val.interfaceName !== "Error") {
    errorIface = ctx.toInterfaceValue("Error", val);
  }

  let message = ctx.valueToString(val);
  try {
    const msgMember = memberAccessOnValue(ctx, errorIface, AST.identifier("message"), env);
    const msgVal = callCallableValue(ctx, msgMember, [], env);
    if (msgVal.kind === "String") {
      message = msgVal.value;
    }
  } catch {
    // fall back to valueToString
  }

  let cause: V10Value | undefined;
  try {
    const causeMember = memberAccessOnValue(ctx, errorIface, AST.identifier("cause"), env);
    const causeVal = callCallableValue(ctx, causeMember, [], env);
    if (causeVal.kind !== "nil") {
      cause = causeVal;
    }
  } catch {
    // ignore cause lookup failures
  }

  const underlying = val.kind === "interface_value" ? val.value : val;
  return ctx.makeRuntimeError(message, underlying, cause);
}
