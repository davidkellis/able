import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { BreakLabelSignal, BreakSignal, ContinueSignal, GeneratorYieldSignal, ProcYieldSignal, ReturnSignal } from "./signals";
import type { IteratorValue, V10Value } from "./values";
import * as AST from "../ast";
import type { ContinuationContext } from "./continuations";
import { memberAccessOnValue } from "./structs";
import { callCallableValue } from "./functions";

function isContinuationYield(context: ContinuationContext, err: unknown): boolean {
  if (context.kind === "generator") {
    return err instanceof GeneratorYieldSignal;
  }
  return err instanceof ProcYieldSignal;
}

export function evaluateBlockExpression(ctx: InterpreterV10, node: AST.BlockExpression, env: Environment): V10Value {
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateBlockExpressionWithContinuation(ctx, node, env, procContext);
  }
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateBlockExpressionWithContinuation(ctx, node, env, generator);
  }
  const blockEnv = new Environment(env);
  let result: V10Value = { kind: "nil", value: null };
  for (const statement of node.body) {
    ctx.checkTimeSlice();
    result = ctx.evaluate(statement, blockEnv);
  }
  return result;
}

function evaluateBlockExpressionWithContinuation(
  ctx: InterpreterV10,
  node: AST.BlockExpression,
  env: Environment,
  continuation: ContinuationContext,
): V10Value {
  let state = continuation.getBlockState(node);
  if (!state) {
    state = {
      env: new Environment(env),
      index: 0,
      result: { kind: "nil", value: null },
    };
    continuation.setBlockState(node, state);
  }

  const blockEnv = state.env;
  let result = state.result ?? { kind: "nil", value: null };
  let index = state.index;

  while (index < node.body.length) {
    const statement = node.body[index]!;
    try {
      ctx.checkTimeSlice();
      result = ctx.evaluate(statement, blockEnv);
    } catch (err) {
      if (isContinuationYield(continuation, err)) {
        let advanceIndex = false;
        let repeatStatement = false;
        let awaitBlocked = false;
        if (continuation.kind === "generator" && (statement.type === "YieldStatement" || isGenYieldCall(statement))) {
          advanceIndex = true;
        } else if (continuation.kind === "proc") {
          if (typeof (continuation as any).consumeRepeatCurrentStatement === "function") {
            repeatStatement = (continuation as any).consumeRepeatCurrentStatement();
          }
          const asyncCtx = ctx.currentAsyncContext ? ctx.currentAsyncContext() : null;
          awaitBlocked = asyncCtx && asyncCtx.kind === "proc" ? asyncCtx.handle.awaitBlocked : false;
          if (!repeatStatement && ctx.manualYieldRequested && !awaitBlocked) {
            advanceIndex = true;
          }
        }
        if (continuation.kind === "proc" && repeatStatement && ctx.manualYieldRequested && !awaitBlocked) {
          index = 0;
        } else if (advanceIndex) {
          index += 1;
        }
        state.index = index;
        state.result = result;
        throw err;
      }
      continuation.clearBlockState(node);
      throw err;
    }
    index += 1;
    state.index = index;
    state.result = result;
  }

  continuation.clearBlockState(node);
  return result;
}

function isGenYieldCall(statement: AST.Statement): boolean {
  if (statement.type !== "FunctionCall") return false;
  const callee = statement.callee;
  if (callee.type !== "MemberAccessExpression") return false;
  const object = callee.object;
  const member = callee.member;
  if (object.type !== "Identifier") return false;
  if (member.type !== "Identifier") return false;
  return object.name === "gen" && member.name === "yield";
}

export function evaluateIfExpression(ctx: InterpreterV10, node: AST.IfExpression, env: Environment): V10Value {
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateIfExpressionWithContinuation(ctx, node, env, procContext);
  }
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateIfExpressionWithContinuation(ctx, node, env, generator);
  }
  const cond = ctx.evaluate(node.ifCondition, env);
  if (ctx.isTruthy(cond)) return ctx.evaluate(node.ifBody, env);
  for (const clause of node.elseIfClauses) {
    const c = ctx.evaluate(clause.condition, env);
    if (ctx.isTruthy(c)) return ctx.evaluate(clause.body, env);
  }
  if (node.elseBody) return ctx.evaluate(node.elseBody, env);
  return { kind: "nil", value: null };
}

function evaluateIfExpressionWithContinuation(
  ctx: InterpreterV10,
  node: AST.IfExpression,
  env: Environment,
  continuation: ContinuationContext,
): V10Value {
  let state = continuation.getIfState(node);
  if (!state) {
    state = {
      stage: "if_condition",
      elseIfIndex: 0,
      result: { kind: "nil", value: null },
    };
    continuation.setIfState(node, state);
  }

  while (true) {
    switch (state.stage) {
      case "if_condition": {
        try {
          const cond = ctx.evaluate(node.ifCondition, env);
          if (ctx.isTruthy(cond)) {
            state.stage = "if_body";
            continue;
          }
          state.stage = "elseif_condition";
          state.elseIfIndex = 0;
          continue;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
            if (continuation.kind === "proc") {
              continuation.clearIfState(node);
            }
          } else {
            continuation.clearIfState(node);
          }
          throw err;
        }
      }
      case "if_body": {
        try {
          const result = ctx.evaluate(node.ifBody, env);
          continuation.clearIfState(node);
          return result;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
            if (continuation.kind === "proc") {
              continuation.clearIfState(node);
            }
          } else {
            continuation.clearIfState(node);
          }
          throw err;
        }
      }
      case "elseif_condition": {
        if (state.elseIfIndex >= node.elseIfClauses.length) {
          state.stage = "else_body";
          continue;
        }
        const clause = node.elseIfClauses[state.elseIfIndex]!;
        try {
          const cond = ctx.evaluate(clause.condition, env);
          if (ctx.isTruthy(cond)) {
            state.stage = "elseif_body";
            continue;
          }
          state.elseIfIndex += 1;
          continue;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
            if (continuation.kind === "proc") {
              continuation.clearIfState(node);
            }
          } else {
            continuation.clearIfState(node);
          }
          throw err;
        }
      }
      case "elseif_body": {
        const clause = node.elseIfClauses[state.elseIfIndex]!;
        try {
          const result = ctx.evaluate(clause.body, env);
          continuation.clearIfState(node);
          return result;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
            if (continuation.kind === "proc") {
              continuation.clearIfState(node);
            }
          } else {
            continuation.clearIfState(node);
          }
          throw err;
        }
      }
      case "else_body": {
        try {
          const result = node.elseBody ? ctx.evaluate(node.elseBody, env) : { kind: "nil", value: null };
          continuation.clearIfState(node);
          return result;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
            if (continuation.kind === "proc") {
              continuation.clearIfState(node);
            }
          } else {
            continuation.clearIfState(node);
          }
          throw err;
        }
      }
      default: {
        continuation.clearIfState(node);
        return { kind: "nil", value: null };
      }
    }
  }
}

export function evaluateWhileLoop(ctx: InterpreterV10, node: AST.WhileLoop, env: Environment): V10Value {
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateWhileLoopWithContinuation(ctx, node, env, procContext);
  }
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateWhileLoopWithContinuation(ctx, node, env, generator);
  }
  while (true) {
    ctx.checkTimeSlice();
    const condition = ctx.evaluate(node.condition, env);
    if (!ctx.isTruthy(condition)) {
      return { kind: "nil", value: null };
    }
    const bodyEnv = new Environment(env);
    try {
      ctx.evaluate(node.body, bodyEnv);
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

export function evaluateLoopExpression(ctx: InterpreterV10, node: AST.LoopExpression, env: Environment): V10Value {
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateLoopExpressionWithContinuation(ctx, node, env, procContext);
  }
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateLoopExpressionWithContinuation(ctx, node, env, generator);
  }
  let result: V10Value = { kind: "nil", value: null };
  while (true) {
    ctx.checkTimeSlice();
    const loopEnv = new Environment(env);
    try {
      result = ctx.evaluate(node.body, loopEnv);
    } catch (err) {
      if (err instanceof BreakSignal) {
        if (err.label) throw new Error("Labeled break not supported");
        return err.value;
      }
      if (err instanceof BreakLabelSignal) throw err;
      if (err instanceof ContinueSignal) {
        if (err.label) throw new Error("Labeled continue not supported");
        continue;
      }
      throw err;
    }
  }
}

function evaluateLoopExpressionWithContinuation(
  ctx: InterpreterV10,
  node: AST.LoopExpression,
  env: Environment,
  continuation: ContinuationContext,
): V10Value {
  if (!continuation) throw new Error("Continuation context missing");
  let state = continuation.getLoopExpressionState(node);
  if (!state) {
    state = {
      baseEnv: env,
      result: { kind: "nil", value: null },
      inBody: false,
      loopEnv: undefined,
    };
    continuation.setLoopExpressionState(node, state);
  }

  let result = state.result ?? { kind: "nil", value: null };

  const resetBody = () => {
    state.inBody = false;
    state.loopEnv = undefined;
  };

  while (true) {
    ctx.checkTimeSlice();
    if (!state.inBody) {
      state.loopEnv = new Environment(state.baseEnv);
      state.inBody = true;
    }
    const loopEnv = state.loopEnv!;
    try {
      const bodyResult = ctx.evaluate(node.body, loopEnv);
      result = bodyResult;
      state.result = result;
      resetBody();
      continue;
    } catch (err) {
      if (isContinuationYield(continuation, err)) {
        state.result = result;
        continuation.markStatementIncomplete();
        throw err;
      }
      if (err instanceof BreakSignal) {
        if (err.label) {
          continuation.clearLoopExpressionState(node);
          throw err;
        }
        continuation.clearLoopExpressionState(node);
        return err.value;
      }
      if (err instanceof BreakLabelSignal) {
        continuation.clearLoopExpressionState(node);
        throw err;
      }
      if (err instanceof ContinueSignal) {
        if (err.label) {
          continuation.clearLoopExpressionState(node);
          throw new Error("Labeled continue not supported");
        }
        resetBody();
        continue;
      }
      continuation.clearLoopExpressionState(node);
      throw err;
    }
  }
}

function evaluateWhileLoopWithContinuation(
  ctx: InterpreterV10,
  node: AST.WhileLoop,
  env: Environment,
  continuation: ContinuationContext,
): V10Value {
  if (!continuation) throw new Error("Continuation context missing");

  let state = continuation.getWhileLoopState(node);
  if (!state) {
    state = {
      baseEnv: env,
      result: { kind: "nil", value: null },
      inBody: false,
      loopEnv: undefined,
      conditionInProgress: false,
    };
    continuation.setWhileLoopState(node, state);
  }

  let result = state.result ?? { kind: "nil", value: null };

  const resetBody = () => {
    state.inBody = false;
    state.loopEnv = undefined;
  };

  while (true) {
    ctx.checkTimeSlice();
    if (!state.inBody) {
      state.conditionInProgress = true;
      let condition: V10Value;
      try {
        condition = ctx.evaluate(node.condition, env);
      } catch (err) {
        if (isContinuationYield(continuation, err)) {
          state.result = result;
          continuation.markStatementIncomplete();
          throw err;
        }
        continuation.clearWhileLoopState(node);
        throw err;
      } finally {
        state.conditionInProgress = false;
      }
      if (!ctx.isTruthy(condition)) {
        continuation.clearWhileLoopState(node);
        return { kind: "nil", value: null };
      }
      state.inBody = true;
      state.loopEnv = new Environment(state.baseEnv);
    }

    const loopEnv = state.loopEnv!;
    try {
      const bodyResult = ctx.evaluate(node.body, loopEnv);
      result = bodyResult;
      state.result = result;
      resetBody();
      continue;
    } catch (err) {
      if (isContinuationYield(continuation, err)) {
        state.result = result;
        continuation.markStatementIncomplete();
        throw err;
      }
      if (err instanceof BreakSignal) {
        if (err.label) {
          continuation.clearWhileLoopState(node);
          throw err;
        }
        continuation.clearWhileLoopState(node);
        return err.value;
      }
      if (err instanceof BreakLabelSignal) {
        continuation.clearWhileLoopState(node);
        throw err;
      }
      if (err instanceof ContinueSignal) {
        if (err.label) {
          continuation.clearWhileLoopState(node);
          throw new Error("Labeled continue not supported");
        }
        resetBody();
        continue;
      }
      continuation.clearWhileLoopState(node);
      throw err;
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
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateForLoopWithContinuation(ctx, node, env, procContext, iterableValue);
  }
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateForLoopWithContinuation(ctx, node, env, generator, iterableValue);
  }
  const baseEnv = new Environment(env);
  const values: V10Value[] = [];
  if (iterableValue.kind === "array") {
    values.push(...iterableValue.elements);
  } else if (iterableValue.kind === "iterator") {
    return iterateDynamicIterator(ctx, node, baseEnv, iterableValue);
  } else {
    const iterator = resolveIteratorValue(ctx, iterableValue, env);
    return iterateDynamicIterator(ctx, node, baseEnv, iterator);
  }

  let last: V10Value = { kind: "nil", value: null };
  for (const value of values) {
    ctx.checkTimeSlice();
    const loopEnv = new Environment(baseEnv);
    bindPattern(ctx, node.pattern, value, loopEnv);
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
  return { kind: "nil", value: null };
}

function iterateDynamicIterator(ctx: InterpreterV10, loop: AST.ForLoop, baseEnv: Environment, iterator: IteratorValue): V10Value {
  let result: V10Value = { kind: "nil", value: null };
  try {
    while (true) {
      ctx.checkTimeSlice();
      let step;
      try {
        step = iterator.iterator.next();
      } catch (err) {
        throw err;
      }
      if (step.done) {
        return { kind: "nil", value: null };
      }
      const loopEnv = new Environment(baseEnv);
      bindPattern(ctx, loop.pattern, step.value, loopEnv);
      try {
        const val = ctx.evaluate(loop.body, loopEnv);
        result = val;
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
  } finally {
    iterator.iterator.close();
  }
}

function evaluateForLoopWithContinuation(
  ctx: InterpreterV10,
  loop: AST.ForLoop,
  env: Environment,
  continuation: ContinuationContext,
  iterableValue: V10Value,
): V10Value {
  if (!continuation) {
    throw new Error("Continuation context missing");
  }

  let state = continuation.getForLoopState(loop);
  if (!state) {
    const baseEnv = new Environment(env);
    const initialResult: V10Value = { kind: "nil", value: null };
    if (iterableValue.kind === "array") {
      state = {
        mode: "static",
        values: [...iterableValue.elements],
        baseEnv,
        index: 0,
        result: initialResult,
        awaitingBody: false,
      };
    } else if (iterableValue.kind === "iterator") {
      state = {
        mode: "iterator",
        iterator: iterableValue,
        baseEnv,
        index: 0,
        result: initialResult,
        awaitingBody: false,
      };
    } else {
      const iterator = resolveIteratorValue(ctx, iterableValue, env);
      state = {
        mode: "iterator",
        iterator,
        baseEnv,
        index: 0,
        result: initialResult,
        awaitingBody: false,
      };
    }
    continuation.setForLoopState(loop, state);
  }

  const baseEnv = state.baseEnv;
  let result = state.result ?? { kind: "nil", value: null };

  const cleanup = () => {
    if (state?.mode === "iterator" && state.iterator && !state.iteratorClosed) {
      try {
        state.iterator.iterator.close();
      } catch {}
      state.iteratorClosed = true;
    }
    continuation.clearForLoopState(loop);
  };

  while (true) {
    ctx.checkTimeSlice();
    let iterationEnv = state.iterationEnv;
    let value: V10Value | undefined;
    if (state.awaitingBody && iterationEnv) {
      value = state.pendingValue;
    } else {
      if (state.mode === "static") {
        const values = state.values ?? [];
        if (state.index >= values.length) {
          cleanup();
          return { kind: "nil", value: null };
        }
        value = values[state.index]!;
      } else {
        const iterator = state.iterator;
        if (!iterator) {
          cleanup();
          throw new Error("iterator() did not return Iterator value");
        }
        let step;
        try {
          step = iterator.iterator.next();
        } catch (err) {
          cleanup();
          throw err;
        }
        if (step.done) {
          cleanup();
          return { kind: "nil", value: null };
        }
        value = step.value;
      }
      iterationEnv = new Environment(baseEnv);
      state.iterationEnv = iterationEnv;
      state.pendingValue = value;
      state.awaitingBody = true;
      try {
        bindPattern(ctx, loop.pattern, value!, iterationEnv);
      } catch (err) {
        cleanup();
        throw err;
      }
    }

    try {
      const bodyResult = ctx.evaluate(loop.body, iterationEnv);
      result = bodyResult;
      state.result = result;
      state.awaitingBody = false;
      state.iterationEnv = undefined;
      state.pendingValue = undefined;
      state.index += 1;
      continue;
    } catch (err) {
      if (isContinuationYield(continuation, err)) {
        state.result = result;
        continuation.markStatementIncomplete();
        throw err;
      }
      if (err instanceof BreakSignal) {
        if (err.label) {
          cleanup();
          throw new Error("Labeled break not supported");
        }
        cleanup();
        return err.value;
      }
      if (err instanceof BreakLabelSignal) {
        cleanup();
        throw err;
      }
      if (err instanceof ContinueSignal) {
        if (err.label) {
          cleanup();
          throw new Error("Labeled continue not supported");
        }
        state.awaitingBody = false;
        state.iterationEnv = undefined;
        state.pendingValue = undefined;
        state.index += 1;
        continue;
      }
      cleanup();
      throw err;
    }
  }
}

function bindPattern(ctx: InterpreterV10, pattern: AST.Pattern, value: V10Value, env: Environment): void {
  if (pattern.type === "Identifier") {
    env.define(pattern.name, value);
    return;
  }
  if (pattern.type === "WildcardPattern") {
    return;
  }
  ctx.assignByPattern(pattern as AST.Pattern, value, env, true);
}

export function resolveIteratorValue(ctx: InterpreterV10, iterable: V10Value, env: Environment): IteratorValue {
  ctx.ensureIteratorBuiltins();
  if (iterable.kind === "iterator") {
    return iterable;
  }
  const direct = adaptIteratorValue(ctx, iterable, env);
  if (direct) return direct;
  const tempEnv = new Environment(env);
  const tempIdent = "__able_iter_target";
  tempEnv.define(tempIdent, iterable);
  const call = AST.functionCall(
    AST.memberAccessExpression(AST.identifier(tempIdent), "iterator"),
    [],
  );
  const result = ctx.evaluate(call, tempEnv);
  if (result && result.kind !== "iterator") {
    const adapted = adaptIteratorValue(ctx, result, env);
    if (adapted) return adapted;
  }
  if (!result || result.kind !== "iterator") {
    throw new Error("iterator() did not return Iterator value");
  }
  return result;
}

function adaptIteratorValue(ctx: InterpreterV10, candidate: V10Value, env: Environment): IteratorValue | null {
  const receiver = candidate.kind === "interface_value" ? candidate.value : candidate;
  if (!receiver || receiver.kind !== "struct_instance") {
    return null;
  }
  const nextMethod = bindIteratorMethod(ctx, receiver, "next", env);
  if (!nextMethod) return null;
  const closeMethod = bindIteratorMethod(ctx, receiver, "close", env);
  return {
    kind: "iterator",
      iterator: {
        next: () => {
          const step = callCallableValue(ctx as any, nextMethod as any, [], env);
          if (isIteratorEnd(ctx, step)) {
            return { done: true, value: ctx.iteratorEndValue };
          }
          return { done: false, value: step ?? ctx.iteratorEndValue };
        },
        close: () => {
          if (closeMethod) {
            callCallableValue(ctx as any, closeMethod as any, [], env);
          }
        },
      },
    };
}

function bindIteratorMethod(ctx: InterpreterV10, receiver: V10Value, name: string, env: Environment): V10Value | null {
  try {
    const access = memberAccessOnValue(ctx, receiver, AST.identifier(name), env, { preferMethods: true });
    if (access && (access.kind === "bound_method" || access.kind === "native_bound_method" || access.kind === "function")) {
      return access;
    }
  } catch {}
  return null;
}

function isIteratorEnd(ctx: InterpreterV10, value: V10Value): boolean {
  if (!value) return false;
  if (value.kind === "iterator_end") return true;
  if (value.kind === "interface_value") return isIteratorEnd(ctx, value.value);
  if (value.kind === "struct_instance" && value.def?.id?.name === "IteratorEnd") return true;
  return false;
}

export function evaluateReturnStatement(ctx: InterpreterV10, node: AST.ReturnStatement, env: Environment): never {
  const value = node.argument ? ctx.evaluate(node.argument, env) : { kind: "nil", value: null };
  throw new ReturnSignal(value);
}
