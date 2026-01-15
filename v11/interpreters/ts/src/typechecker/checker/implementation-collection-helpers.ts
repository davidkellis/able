import * as AST from "../../ast";
import type { ImplementationContext } from "./implementation-context";
import { typeInfoToTypeExpression } from "./type-expression-utils";

export function canonicalizeTargetType(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
): AST.TypeExpression | null | undefined {
  const expanded = expandTypeAliases(ctx, expr);
  switch (expanded?.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expanded.name);
      if (name) {
        const binding = ctx.lookupIdentifier?.(name);
        const canonical = binding ? typeInfoToTypeExpression(binding) : null;
        if (canonical) {
          return canonical;
        }
      }
      return expanded;
    }
    case "GenericTypeExpression":
      return {
        ...expanded,
        base: canonicalizeTargetType(ctx, expanded.base) ?? expanded.base,
        arguments: (expanded.arguments ?? []).map((arg) => canonicalizeTargetType(ctx, arg) ?? arg),
      };
    case "NullableTypeExpression":
      return { ...expanded, innerType: canonicalizeTargetType(ctx, expanded.innerType) ?? expanded.innerType };
    case "ResultTypeExpression":
      return { ...expanded, innerType: canonicalizeTargetType(ctx, expanded.innerType) ?? expanded.innerType };
    case "UnionTypeExpression":
      return { ...expanded, members: (expanded.members ?? []).map((member) => canonicalizeTargetType(ctx, member) ?? member) };
    case "FunctionTypeExpression":
      return {
        ...expanded,
        paramTypes: (expanded.paramTypes ?? []).map((param) => canonicalizeTargetType(ctx, param) ?? param),
        returnType: canonicalizeTargetType(ctx, expanded.returnType) ?? expanded.returnType,
      };
    default:
      return expanded;
  }
}

export function expandTypeAliases(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
  seen: Set<string> = new Set(),
): AST.TypeExpression | null | undefined {
  if (!expr) return expr;
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expr.name);
      if (!name || !ctx.getTypeAlias) return expr;
      if (seen.has(name)) return expr;
      const alias = ctx.getTypeAlias(name);
      if (!alias?.targetType) return expr;
      seen.add(name);
      const expanded = expandTypeAliases(ctx, alias.targetType, seen);
      seen.delete(name);
      return expanded ?? expr;
    }
    case "GenericTypeExpression": {
      const baseName = ctx.getIdentifierNameFromTypeExpression(expr.base);
      const expandedBase = expandTypeAliases(ctx, expr.base, seen) ?? expr.base;
      const expandedArgs = (expr.arguments ?? []).map((arg) => expandTypeAliases(ctx, arg, seen) ?? arg);
      if (!baseName || !ctx.getTypeAlias || seen.has(baseName)) {
        return { ...expr, base: expandedBase, arguments: expandedArgs };
      }
      const alias = ctx.getTypeAlias(baseName);
      if (!alias?.targetType) {
        return { ...expr, base: expandedBase, arguments: expandedArgs };
      }
      const substitutions = new Map<string, AST.TypeExpression>();
      (alias.genericParams ?? []).forEach((param, index) => {
        const paramName = ctx.getIdentifierName(param?.name);
        if (!paramName) return;
        substitutions.set(paramName, expandedArgs[index] ?? AST.wildcardTypeExpression());
      });
      seen.add(baseName);
      const substituted = substituteTypeExpression(ctx, alias.targetType, substitutions, seen);
      const expanded = expandTypeAliases(ctx, substituted, seen);
      seen.delete(baseName);
      return expanded ?? substituted;
    }
    case "NullableTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, expr.innerType, seen) ?? expr.innerType };
    case "ResultTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, expr.innerType, seen) ?? expr.innerType };
    case "UnionTypeExpression":
      return { ...expr, members: (expr.members ?? []).map((member) => expandTypeAliases(ctx, member, seen) ?? member) };
    case "FunctionTypeExpression":
      return {
        ...expr,
        paramTypes: (expr.paramTypes ?? []).map((param) => expandTypeAliases(ctx, param, seen) ?? param),
        returnType: expandTypeAliases(ctx, expr.returnType, seen) ?? expr.returnType,
      };
    default:
      return expr;
  }
}

export function collectTargetTypeParams(ctx: ImplementationContext, targetType: AST.TypeExpression | null | undefined): string[] {
  if (!targetType) return [];
  if (targetType.type === "GenericTypeExpression" && Array.isArray(targetType.arguments)) {
    return targetType.arguments
      .map((arg) => ctx.getIdentifierNameFromTypeExpression(arg))
      .filter((name): name is string => Boolean(name) && !ctx.isKnownTypeName(name));
  }
  return [];
}

function substituteTypeExpression(
  ctx: ImplementationContext,
  expr: AST.TypeExpression,
  substitutions: Map<string, AST.TypeExpression>,
  seen: Set<string>,
): AST.TypeExpression {
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expr.name);
      if (name && substitutions.has(name)) {
        return expandTypeAliases(ctx, substitutions.get(name), seen) ?? substitutions.get(name)!;
      }
      return expr;
    }
    case "GenericTypeExpression": {
      const base = expandTypeAliases(ctx, expr.base, seen) ?? expr.base;
      const args = (expr.arguments ?? []).map((arg) => {
        const next = arg ? substituteTypeExpression(ctx, arg, substitutions, seen) : undefined;
        return expandTypeAliases(ctx, next, seen) ?? next ?? arg;
      });
      return { ...expr, base, arguments: args };
    }
    case "NullableTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, substituteTypeExpression(ctx, expr.innerType, substitutions, seen), seen) };
    case "ResultTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, substituteTypeExpression(ctx, expr.innerType, substitutions, seen), seen) };
    case "UnionTypeExpression":
      return {
        ...expr,
        members: (expr.members ?? []).map((member) =>
          expandTypeAliases(ctx, substituteTypeExpression(ctx, member, substitutions, seen), seen),
        ),
      };
    case "FunctionTypeExpression":
      return {
        ...expr,
        paramTypes: (expr.paramTypes ?? []).map((param) =>
          expandTypeAliases(ctx, substituteTypeExpression(ctx, param, substitutions, seen), seen),
        ),
        returnType: expandTypeAliases(ctx, substituteTypeExpression(ctx, expr.returnType, substitutions, seen), seen) ?? expr.returnType,
      };
    default:
      return expr;
  }
}
