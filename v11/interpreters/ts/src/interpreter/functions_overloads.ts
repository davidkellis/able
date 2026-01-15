import * as AST from "../ast";
import type { Interpreter } from "./index";
import type { ConstraintSpec, RuntimeValue } from "./values";

export function makePartialFunction(
  target: RuntimeValue,
  boundArgs: RuntimeValue[],
  callNode?: AST.FunctionCall,
): Extract<RuntimeValue, { kind: "partial_function" }> {
  return { kind: "partial_function", target, boundArgs, callNode };
}

export function functionArityRange(
  funcNode: AST.FunctionDefinition | AST.LambdaExpression,
): { minArgs: number; maxArgs: number; optionalLast: boolean } {
  const params = funcNode.params ?? [];
  const paramCount = params.length;
  const optionalLast = paramCount > 0 && params[paramCount - 1]?.paramType?.type === "NullableTypeExpression";
  const maxArgs = funcNode.type === "FunctionDefinition" && funcNode.isMethodShorthand ? paramCount + 1 : paramCount;
  const minArgs = optionalLast ? Math.max(0, maxArgs - 1) : maxArgs;
  return { minArgs, maxArgs, optionalLast };
}

export function overloadArityRange(
  overloads: Array<Extract<RuntimeValue, { kind: "function" }>>,
): { minArgs: number; maxArgs: number } {
  let minArgs = Number.POSITIVE_INFINITY;
  let maxArgs = 0;
  for (const fn of overloads) {
    const arity = functionArityRange(fn.node);
    if (arity.minArgs < minArgs) minArgs = arity.minArgs;
    if (arity.maxArgs > maxArgs) maxArgs = arity.maxArgs;
  }
  if (!Number.isFinite(minArgs)) minArgs = 0;
  return { minArgs, maxArgs };
}

export function selectRuntimeOverload(
  ctx: Interpreter,
  overloads: Array<Extract<RuntimeValue, { kind: "function" }>>,
  evalArgs: RuntimeValue[],
  callNode?: AST.FunctionCall,
): Extract<RuntimeValue, { kind: "function" }> | null {
  const candidates: Array<{
    fn: Extract<RuntimeValue, { kind: "function" }>;
    params: AST.FunctionParameter[];
    optionalLast: boolean;
    score: number;
    priority: number;
  }> = [];
  for (const fn of overloads) {
    const funcNode = fn.node;
    if (funcNode.type !== "FunctionDefinition") {
      candidates.push({ fn, params: funcNode.params ?? [], optionalLast: false, score: 0, priority: 0 });
      continue;
    }
    const params = funcNode.params ?? [];
    const arity = functionArityRange(funcNode);
    if (evalArgs.length < arity.minArgs || evalArgs.length > arity.maxArgs) {
      continue;
    }
    const paramsForCheck =
      arity.optionalLast && evalArgs.length === arity.maxArgs - 1
        ? params.slice(0, Math.max(0, params.length - 1))
        : params;
    const argsForCheck = funcNode.isMethodShorthand ? evalArgs.slice(1) : evalArgs;
    if (argsForCheck.length !== paramsForCheck.length) {
      continue;
    }
    const generics = new Set((funcNode.genericParams ?? []).map((gp) => gp.name.name));
    let score = arity.optionalLast && evalArgs.length === arity.maxArgs - 1 ? -0.5 : 0;
    let compatible = true;
    for (let i = 0; i < paramsForCheck.length; i += 1) {
      const param = paramsForCheck[i];
      const arg = argsForCheck[i];
      if (!param || arg === undefined) {
        compatible = false;
        break;
      }
      if (param.paramType) {
        const paramType = substituteSelfType(param.paramType, fn);
        if (!ctx.matchesType(paramType, arg)) {
          compatible = false;
          break;
        }
        score += parameterSpecificity(paramType, generics);
      }
    }
    if (compatible) {
      const priority = typeof (fn as any).methodResolutionPriority === "number" ? (fn as any).methodResolutionPriority : 0;
      candidates.push({ fn, params: paramsForCheck, optionalLast: arity.optionalLast, score, priority });
    }
  }
  if (!candidates.length) return null;
  candidates.sort((a, b) => {
    if (a.score !== b.score) return b.score - a.score;
    return b.priority - a.priority;
  });
  const best = candidates[0]!;
  const ties = candidates.filter((c) => c.score === best.score && c.priority === best.priority);
  if (ties.length > 1) {
    const name =
      callNode?.callee?.type === "Identifier"
        ? callNode.callee.name
        : callNode?.callee?.type === "MemberAccessExpression"
          ? callNode.callee.member?.type === "Identifier"
            ? callNode.callee.member.name
            : "(function)"
          : "(function)";
    throw new Error(`Ambiguous overload for ${name}`);
  }
  return best.fn;
}

export function substituteSelfType(
  typeExpr: AST.TypeExpression,
  fn: Extract<RuntimeValue, { kind: "function" }>,
): AST.TypeExpression {
  const target = (fn as any).methodSetTargetType as AST.TypeExpression | undefined;
  if (!target) return typeExpr;
  return replaceSelfTypeExpr(typeExpr, target);
}

export function replaceSelfTypeExpr(typeExpr: AST.TypeExpression, replacement: AST.TypeExpression): AST.TypeExpression {
  switch (typeExpr.type) {
    case "SimpleTypeExpression":
      if (typeExpr.name.name === "Self") return replacement;
      return typeExpr;
    case "GenericTypeExpression": {
      const base = replaceSelfTypeExpr(typeExpr.base, replacement);
      const args = (typeExpr.arguments ?? []).map((arg) => (arg ? replaceSelfTypeExpr(arg, replacement) : arg));
      if (base === typeExpr.base && args.every((arg, idx) => arg === (typeExpr.arguments ?? [])[idx])) {
        return typeExpr;
      }
      return AST.genericTypeExpression(base, args);
    }
    case "NullableTypeExpression": {
      const inner = replaceSelfTypeExpr(typeExpr.innerType, replacement);
      if (inner === typeExpr.innerType) return typeExpr;
      return AST.nullableTypeExpression(inner);
    }
    case "ResultTypeExpression": {
      const inner = replaceSelfTypeExpr(typeExpr.innerType, replacement);
      if (inner === typeExpr.innerType) return typeExpr;
      return AST.resultTypeExpression(inner);
    }
    case "UnionTypeExpression": {
      const members = typeExpr.members.map((member) => replaceSelfTypeExpr(member, replacement));
      if (members.every((member, idx) => member === typeExpr.members[idx])) return typeExpr;
      return AST.unionTypeExpression(members);
    }
    case "FunctionTypeExpression": {
      const params = (typeExpr.paramTypes ?? []).map((param) => replaceSelfTypeExpr(param, replacement));
      const ret = replaceSelfTypeExpr(typeExpr.returnType, replacement);
      if (ret === typeExpr.returnType && params.every((param, idx) => param === (typeExpr.paramTypes ?? [])[idx])) {
        return typeExpr;
      }
      return AST.functionTypeExpression(params, ret);
    }
    default:
      return typeExpr;
  }
}

export function resolveMethodSetReceiver(
  funcNode: AST.FunctionDefinition | AST.LambdaExpression,
  evalArgs: RuntimeValue[],
): RuntimeValue | null {
  if (funcNode.type !== "FunctionDefinition") return null;
  if (funcNode.isMethodShorthand) {
    return evalArgs[0] ?? null;
  }
  const params = funcNode.params ?? [];
  if (params.length === 0) return null;
  const first = params[0];
  if (!first) return null;
  if (first.name?.type === "Identifier" && first.name.name.toLowerCase() === "self") {
    return evalArgs[0] ?? null;
  }
  if (first.paramType?.type === "SimpleTypeExpression" && first.paramType.name.name === "Self") {
    return evalArgs[0] ?? null;
  }
  return null;
}

export function enforceMethodSetConstraints(
  ctx: Interpreter,
  funcValue: Extract<RuntimeValue, { kind: "function" }>,
  receiver: RuntimeValue,
): void {
  const constraints = (funcValue as any).methodSetConstraints as ConstraintSpec[] | undefined;
  if (!constraints || constraints.length === 0) return;
  const targetType = (funcValue as any).methodSetTargetType as AST.TypeExpression | undefined;
  const genericParams = (funcValue as any).methodSetGenericParams as AST.GenericParameter[] | undefined;
  const actualTypeExpr = ctx.typeExpressionFromValue(receiver);
  if (!actualTypeExpr) return;
  const bindings = new Map<string, AST.TypeExpression>();
  const genericNames = new Set((genericParams ?? []).map((gp) => gp.name.name));
  if (targetType && genericNames.size > 0) {
    const canonicalTarget = ctx.expandTypeAliases(targetType);
    const canonicalActual = ctx.expandTypeAliases(actualTypeExpr);
    ctx.matchTypeExpressionTemplate(canonicalTarget, canonicalActual, genericNames, bindings);
    bindings.set("Self", canonicalActual);
  } else {
    bindings.set("Self", ctx.expandTypeAliases(actualTypeExpr));
  }
  if (receiver.kind === "struct_instance" && receiver.typeArgMap) {
    for (const [key, value] of receiver.typeArgMap.entries()) {
      if (!bindings.has(key)) bindings.set(key, value);
    }
  }
  ctx.enforceConstraintSpecs(constraints, bindings, "method set");
}

export function parameterSpecificity(typeExpr: AST.TypeExpression, generics: Set<string>): number {
  if (!typeExpr) return 0;
  switch (typeExpr.type) {
    case "WildcardTypeExpression":
      return 0;
    case "SimpleTypeExpression": {
      const name = typeExpr.name.name;
      if (generics.has(name) || /^[A-Z]$/.test(name)) {
        return 1;
      }
      return 3;
    }
    case "NullableTypeExpression":
      return 1 + parameterSpecificity(typeExpr.innerType, generics);
    case "GenericTypeExpression":
      return 2 + (typeExpr.arguments ?? []).reduce((acc, arg) => acc + parameterSpecificity(arg, generics), 0);
    case "FunctionTypeExpression":
    case "UnionTypeExpression":
      return 2;
    default:
      return 1;
  }
}
