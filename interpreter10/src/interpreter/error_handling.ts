import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { RaiseSignal } from "./signals";
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
      for (const clause of node.clauses) {
        const matchEnv = ctx.tryMatchPattern(clause.pattern, e.value, env);
        if (matchEnv) {
          if (clause.guard) {
            const g = ctx.evaluate(clause.guard, matchEnv);
            if (!ctx.isTruthy(g)) continue;
          }
          return ctx.evaluate(clause.body, matchEnv);
        }
      }
      throw e;
    }
    throw e;
  }
}

export function evaluateOrElseExpression(ctx: InterpreterV10, node: AST.OrElseExpression, env: Environment): V10Value {
  try {
    return ctx.evaluate(node.expression, env);
  } catch (e) {
    if (e instanceof RaiseSignal) {
      const handlerEnv = new Environment(env);
      if (node.errorBinding) handlerEnv.define(node.errorBinding.name, e.value);
      return ctx.evaluate(node.handler, handlerEnv);
    }
    throw e;
  }
}

export function evaluatePropagationExpression(ctx: InterpreterV10, node: AST.PropagationExpression, env: Environment): V10Value {
  try {
    const val = ctx.evaluate(node.expression, env);
    if (val.kind === "error") throw new RaiseSignal(val);
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
