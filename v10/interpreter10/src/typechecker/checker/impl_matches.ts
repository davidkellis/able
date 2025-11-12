import type * as AST from "../../ast";
import { formatType } from "../types";
import type { TypeInfo } from "../types";
import type { ImplementationObligation, ImplementationRecord } from "./types";
import type { ImplementationContext } from "./implementations";

export interface ImplementationMatch {
  record: ImplementationRecord;
  substitutions: Map<string, TypeInfo>;
  interfaceArgs: string[];
  score: number;
  constraintKeys: Set<string>;
}

export function collectUnionVariantLabels(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
): string[] | undefined {
  if (!expr || expr.type !== "UnionTypeExpression") {
    return undefined;
  }
  const variants = new Set<string>();
  const visit = (node: AST.TypeExpression | null | undefined): void => {
    if (!node) return;
    if (node.type === "UnionTypeExpression") {
      for (const member of node.members ?? []) {
        visit(member);
      }
      return;
    }
    const label = ctx.formatTypeExpression(node);
    if (label) {
      variants.add(label);
    }
  };
  for (const member of expr.members ?? []) {
    visit(member);
  }
  if (variants.size === 0) {
    return undefined;
  }
  return Array.from(variants).sort();
}

export function computeImplementationSpecificity(
  ctx: ImplementationContext,
  record: ImplementationRecord,
  substitutions: Map<string, TypeInfo>,
): number {
  const genericNames = new Set(record.genericParams);
  let bindingScore = 0;
  for (const name of genericNames) {
    if (substitutions.has(name)) {
      bindingScore += 1;
    }
  }
  const concreteScore = measureTemplateSpecificity(record.target, genericNames);
  const constraintScore = record.obligations.length;
  const unionPenalty = record.unionVariants ? record.unionVariants.length : 0;
  return concreteScore * 100 + constraintScore * 10 + bindingScore - unionPenalty;
}

export function buildConstraintKeySet(
  ctx: ImplementationContext,
  obligations: ImplementationObligation[],
): Set<string> {
  const result = new Set<string>();
  for (const obligation of obligations) {
    const expressions = collectInterfaceConstraintExpressions(ctx, obligation.interfaceType);
    for (const expr of expressions) {
      result.add(`${obligation.typeParam}->${expr}`);
    }
  }
  return result;
}

export function selectMostSpecificImplementationMatch(
  ctx: ImplementationContext,
  matches: ImplementationMatch[],
  interfaceName: string,
  type: TypeInfo,
): { ok: boolean; detail?: string } {
  let best = matches[0]!;
  let contenders: ImplementationMatch[] = [best];
  for (const candidate of matches.slice(1)) {
    const cmp = compareImplementationMatches(candidate, best);
    if (cmp > 0) {
      best = candidate;
      contenders = [candidate];
      continue;
    }
    if (cmp === 0) {
      const reverse = compareImplementationMatches(best, candidate);
      if (reverse < 0) {
        best = candidate;
        contenders = [candidate];
      } else if (reverse === 0) {
        contenders.push(candidate);
      }
    }
  }
  if (contenders.length === 1) {
    return { ok: true };
  }
  return {
    ok: false,
    detail: formatAmbiguousImplementationDetail(ctx, interfaceName, type, contenders),
  };
}

function compareImplementationMatches(a: ImplementationMatch, b: ImplementationMatch): number {
  if (a.score > b.score) return 1;
  if (a.score < b.score) return -1;
  const aUnion = a.record.unionVariants;
  const bUnion = b.record.unionVariants;
  if (aUnion && !bUnion) return -1;
  if (!aUnion && bUnion) return 1;
  if (aUnion && bUnion) {
    if (isProperSubset(aUnion, bUnion)) return 1;
    if (isProperSubset(bUnion, aUnion)) return -1;
    if (aUnion.length !== bUnion.length) {
      return aUnion.length < bUnion.length ? 1 : -1;
    }
  }
  if (isConstraintSuperset(a.constraintKeys, b.constraintKeys)) return 1;
  if (isConstraintSuperset(b.constraintKeys, a.constraintKeys)) return -1;
  return 0;
}

function formatAmbiguousImplementationDetail(
  ctx: ImplementationContext,
  interfaceName: string,
  type: TypeInfo,
  matches: ImplementationMatch[],
): string {
  const typeLabel = formatType(type);
  const labels = matches.map((match) => ctx.appendInterfaceArgsToLabel(match.record.label, match.interfaceArgs));
  const unique = Array.from(new Set(labels));
  return `ambiguous implementations of ${interfaceName} for ${typeLabel}: ${unique.join(", ")}`;
}

function measureTemplateSpecificity(expr: AST.TypeExpression | null | undefined, generics: Set<string>): number {
  if (!expr) {
    return 0;
  }
  switch (expr.type) {
    case "SimpleTypeExpression":
      return generics.has(expr.name.name) ? 0 : 1;
    case "GenericTypeExpression": {
      let score = measureTemplateSpecificity(expr.base, generics);
      for (const arg of expr.arguments ?? []) {
        score += measureTemplateSpecificity(arg, generics);
      }
      return score;
    }
    case "NullableTypeExpression":
    case "ResultTypeExpression":
      return measureTemplateSpecificity(expr.innerType, generics);
    case "UnionTypeExpression":
      return expr.members?.reduce((sum, member) => sum + measureTemplateSpecificity(member, generics), 0) ?? 0;
    case "FunctionTypeExpression": {
      let total = measureTemplateSpecificity(expr.returnType, generics);
      for (const param of expr.paramTypes ?? []) {
        total += measureTemplateSpecificity(param, generics);
      }
      return total;
    }
    default:
      return 0;
  }
}

function collectInterfaceConstraintExpressions(
  ctx: ImplementationContext,
  typeExpr: AST.TypeExpression | null | undefined,
  memo: Set<string> = new Set(),
): string[] {
  if (!typeExpr) {
    return [];
  }
  const label = ctx.formatTypeExpression(typeExpr);
  if (!label || memo.has(label)) {
    return [];
  }
  memo.add(label);
  const results = [label];
  const interfaceName = ctx.getInterfaceNameFromTypeExpression(typeExpr);
  if (interfaceName) {
    const iface = ctx.getInterfaceDefinition(interfaceName);
    if (iface?.baseInterfaces) {
      for (const base of iface.baseInterfaces) {
        results.push(...collectInterfaceConstraintExpressions(ctx, base, memo));
      }
    }
  }
  return results;
}

function isConstraintSuperset(a: Set<string>, b: Set<string>): boolean {
  if (a.size <= b.size) {
    return false;
  }
  for (const key of b) {
    if (!a.has(key)) {
      return false;
    }
  }
  return true;
}

function isProperSubset(a: string[], b: string[]): boolean {
  if (a.length >= b.length) {
    return false;
  }
  const bSet = new Set(b);
  for (const value of a) {
    if (!bSet.has(value)) {
      return false;
    }
  }
  return true;
}
