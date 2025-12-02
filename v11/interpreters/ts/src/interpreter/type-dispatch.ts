import * as AST from "../ast";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

export type TypeDispatch = { typeName: string; typeArgs: AST.TypeExpression[] };

function parseTypeDispatch(expr: AST.TypeExpression | null | undefined): TypeDispatch | null {
  if (!expr) return null;
  let base: AST.TypeExpression = expr;
  let args: AST.TypeExpression[] = [];
  while (base.type === "GenericTypeExpression") {
    args = base.arguments ?? [];
    base = base.base;
  }
  if (base.type !== "SimpleTypeExpression") return null;
  return { typeName: base.name.name, typeArgs: args };
}

export function collectTypeDispatches(ctx: InterpreterV10, value: V10Value): TypeDispatch[] {
  const dispatches: TypeDispatch[] = [];
  const primary = parseTypeDispatch(ctx.typeExpressionForValue(value));
  if (primary) dispatches.push(primary);
  if (value.kind === "interface_value") {
    const underlying = parseTypeDispatch(ctx.typeExpressionForValue(value.value));
    if (underlying) dispatches.push(underlying);
  }
  return dispatches;
}
