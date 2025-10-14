import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { BreakLabelSignal, BreakSignal, ContinueSignal, ReturnSignal } from "./signals";
import type { V10Value } from "./values";

export function evaluateBlockExpression(ctx: InterpreterV10, node: AST.BlockExpression, env: Environment): V10Value {
  const blockEnv = new Environment(env);
  let result: V10Value = { kind: "nil", value: null };
  for (const statement of node.body) {
    result = ctx.evaluate(statement, blockEnv);
  }
  return result;
}

export function evaluateIfExpression(ctx: InterpreterV10, node: AST.IfExpression, env: Environment): V10Value {
  const cond = ctx.evaluate(node.ifCondition, env);
  if (ctx.isTruthy(cond)) return ctx.evaluate(node.ifBody, env);
  for (const clause of node.orClauses) {
    if (clause.condition) {
      const c = ctx.evaluate(clause.condition, env);
      if (ctx.isTruthy(c)) return ctx.evaluate(clause.body, env);
    } else {
      return ctx.evaluate(clause.body, env);
    }
  }
  return { kind: "nil", value: null };
}

export function evaluateWhileLoop(ctx: InterpreterV10, node: AST.WhileLoop, env: Environment): V10Value {
  let result: V10Value = { kind: "nil", value: null };
  while (true) {
    const condition = ctx.evaluate(node.condition, env);
    if (!ctx.isTruthy(condition)) {
      return result;
    }
    const bodyEnv = new Environment(env);
    try {
      result = ctx.evaluate(node.body, bodyEnv);
    } catch (e) {
      if (e instanceof BreakSignal) {
        if (e.label) throw new Error("Labeled break not supported");
        return e.value;
      }
      if (e instanceof BreakLabelSignal) throw e;
      if (e instanceof ContinueSignal) {
        if (e.label) throw new Error("Labeled continue not supported");
        continue;
      }
      throw e;
    }
  }
}

export function evaluateBreakStatement(ctx: InterpreterV10, node: AST.BreakStatement, env: Environment): never {
  const labelName = node.label ? node.label.name : null;
  const value = node.value ? ctx.evaluate(node.value, env) : { kind: "nil", value: null };
  if (labelName && ctx.breakpointStack.includes(labelName)) {
    throw new BreakLabelSignal(labelName, value);
  }
  throw new BreakSignal(labelName, value);
}

export function evaluateContinueStatement(ctx: InterpreterV10, node: AST.ContinueStatement): never {
  const labelName = node.label ? node.label.name : null;
  if (labelName) throw new Error("Labeled continue not supported");
  throw new ContinueSignal(null);
}

export function evaluateForLoop(ctx: InterpreterV10, node: AST.ForLoop, env: Environment): V10Value {
  const iterableValue = ctx.evaluate(node.iterable, env);
  const baseEnv = new Environment(env);
  const bindPattern = (value: V10Value, targetEnv: Environment) => {
    if (node.pattern.type === "Identifier") {
      targetEnv.define(node.pattern.name, value);
      return;
    }
    if (node.pattern.type === "WildcardPattern") {
      return;
    }
    ctx.assignByPattern(node.pattern as AST.Pattern, value, targetEnv, true);
  };

  const values: V10Value[] = [];
  if (iterableValue.kind === "array") {
    values.push(...iterableValue.elements);
  } else if (iterableValue.kind === "range") {
    const toEndpoint = (value: number): number => {
      if (!Number.isFinite(value)) throw new Error("Range endpoint must be finite");
      return Math.trunc(value);
    };
    const start = toEndpoint(iterableValue.start);
    const end = toEndpoint(iterableValue.end);
    const step = start <= end ? 1 : -1;
    for (let current = start; ; current += step) {
      if (step > 0) {
        if (iterableValue.inclusive) {
          if (current > end) break;
        } else if (current >= end) {
          break;
        }
      } else {
        if (iterableValue.inclusive) {
          if (current < end) break;
        } else if (current <= end) {
          break;
        }
      }
      values.push({ kind: "i32", value: current });
    }
  } else {
    throw new Error("For loop iterable must be array or range");
  }

  let last: V10Value = { kind: "nil", value: null };
  for (const value of values) {
    const loopEnv = new Environment(baseEnv);
    bindPattern(value, loopEnv);
    try {
      last = ctx.evaluate(node.body, loopEnv);
    } catch (e) {
      if (e instanceof BreakSignal) {
        if (e.label) throw new Error("Labeled break not supported");
        return e.value;
      }
      if (e instanceof BreakLabelSignal) throw e;
      if (e instanceof ContinueSignal) {
        if (e.label) throw new Error("Labeled continue not supported");
        continue;
      }
      throw e;
    }
  }
  return last;
}

export function evaluateReturnStatement(ctx: InterpreterV10, node: AST.ReturnStatement, env: Environment): never {
  const value = node.argument ? ctx.evaluate(node.argument, env) : { kind: "nil", value: null };
  throw new ReturnSignal(value);
}
