import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { GeneratorYieldSignal } from "./signals";

export function evaluateMatchExpression(ctx: InterpreterV10, node: AST.MatchExpression, env: Environment): V10Value {
  const generator = ctx.currentGeneratorContext();
  if (!generator) {
    const value = ctx.evaluate(node.subject, env);
    for (const clause of node.clauses) {
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

  let state = generator.getMatchState(node);
  if (!state) {
    state = {
      stage: "subject",
      clauseIndex: 0,
    };
    generator.setMatchState(node, state);
  }

  while (true) {
    switch (state.stage) {
      case "subject": {
        try {
          const value = ctx.evaluate(node.subject, env);
          state.subject = value;
          state.stage = "clause";
          continue;
        } catch (err) {
          if (err instanceof GeneratorYieldSignal) {
            generator.markStatementIncomplete();
          } else {
            generator.clearMatchState(node);
          }
          throw err;
        }
      }
      case "clause": {
        if (state.clauseIndex >= node.clauses.length) {
          generator.clearMatchState(node);
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
          if (err instanceof GeneratorYieldSignal) {
            generator.markStatementIncomplete();
          } else {
            generator.clearMatchState(node);
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
          generator.clearMatchState(node);
          return result;
        } catch (err) {
          if (err instanceof GeneratorYieldSignal) {
            generator.markStatementIncomplete();
          } else {
            generator.clearMatchState(node);
          }
          throw err;
        }
      }
      default:
        generator.clearMatchState(node);
        throw new Error("Non-exhaustive match");
    }
  }
}
