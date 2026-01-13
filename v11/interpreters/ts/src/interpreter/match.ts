import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { GeneratorYieldSignal, ProcYieldSignal } from "./signals";
import type { ContinuationContext } from "./continuations";

function isContinuationYield(context: ContinuationContext, err: unknown): boolean {
  if (context.kind === "generator") {
    return err instanceof GeneratorYieldSignal || err instanceof ProcYieldSignal;
  }
  return err instanceof ProcYieldSignal;
}

export function evaluateMatchExpression(ctx: Interpreter, node: AST.MatchExpression, env: Environment): RuntimeValue {
  const generator = ctx.currentGeneratorContext();
  if (generator) {
    return evaluateMatchExpressionWithContinuation(ctx, node, env, generator);
  }
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    return evaluateMatchExpressionWithContinuation(ctx, node, env, procContext);
  }
  const value = ctx.evaluate(node.subject, env);
  for (const clause of node.clauses) {
    ctx.checkTimeSlice();
    const matchEnv = ctx.tryMatchPattern(clause.pattern, value, env);
    if (matchEnv) {
      if (clause.guard) {
        const guard = ctx.evaluate(clause.guard, matchEnv);
        if (!ctx.isTruthy(guard)) continue;
      }
      return ctx.evaluate(clause.body, matchEnv);
    }
  }
  throw new Error("Non-exhaustive match");
}

function evaluateMatchExpressionWithContinuation(
  ctx: Interpreter,
  node: AST.MatchExpression,
  env: Environment,
  continuation: ContinuationContext,
): RuntimeValue {
  let state = continuation.getMatchState(node);
  if (!state) {
    state = {
      stage: "subject",
      clauseIndex: 0,
    };
    continuation.setMatchState(node, state);
  }

  while (true) {
    ctx.checkTimeSlice();
    switch (state.stage) {
      case "subject": {
        try {
          const value = ctx.evaluate(node.subject, env);
          state.subject = value;
          state.stage = "clause";
          continue;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
          } else {
            continuation.clearMatchState(node);
          }
          throw err;
        }
      }
      case "clause": {
        if (state.clauseIndex >= node.clauses.length) {
          continuation.clearMatchState(node);
          throw new Error("Non-exhaustive match");
        }
        const clause = node.clauses[state.clauseIndex]!;
        const subject = state.subject!;
        const matchEnv = ctx.tryMatchPattern(clause.pattern, subject, env);
        if (!matchEnv) {
          state.clauseIndex += 1;
          continue;
        }
        state.matchEnv = matchEnv;
        if (clause.guard) {
          state.stage = "guard";
          continue;
        }
        state.stage = "body";
        continue;
      }
      case "guard": {
        const clause = node.clauses[state.clauseIndex]!;
        const guardExpr = clause.guard;
        if (!guardExpr) {
          state.stage = "body";
          continue;
        }
        const matchEnv = state.matchEnv;
        if (!matchEnv) {
          state.stage = "clause";
          continue;
        }
        try {
          const guardVal = ctx.evaluate(guardExpr, matchEnv);
          if (ctx.isTruthy(guardVal)) {
            state.stage = "body";
            continue;
          }
          state.clauseIndex += 1;
          state.matchEnv = undefined;
          state.stage = "clause";
          continue;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
          } else {
            continuation.clearMatchState(node);
          }
          throw err;
        }
      }
      case "body": {
        const clause = node.clauses[state.clauseIndex]!;
        const matchEnv = state.matchEnv;
        if (!matchEnv) {
          state.clauseIndex += 1;
          state.stage = "clause";
          continue;
        }
        try {
          const result = ctx.evaluate(clause.body, matchEnv);
          continuation.clearMatchState(node);
          return result;
        } catch (err) {
          if (isContinuationYield(continuation, err)) {
            continuation.markStatementIncomplete();
          } else {
            continuation.clearMatchState(node);
          }
          throw err;
        }
      }
      default:
        continuation.clearMatchState(node);
        throw new Error("Non-exhaustive match");
    }
  }
}
