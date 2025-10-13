import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

export function evaluateMatchExpression(ctx: InterpreterV10, node: AST.MatchExpression, env: Environment): V10Value {
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
