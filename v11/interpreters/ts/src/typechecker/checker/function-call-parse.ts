import * as AST from "../../ast";
import { unknownType, type TypeInfo } from "../types";
import { typeInfoToTypeExpression } from "./type-expression-utils";
import type { FunctionInfo } from "./types";
import type { FunctionCallContext } from "./function-calls";

type CallSignature = {
  params: TypeInfo[];
  optionalLast: boolean;
  returnType: TypeInfo;
  inferredTypeArgs?: AST.TypeExpression[];
};

export function instantiateCallSignature(
  ctx: FunctionCallContext,
  info: FunctionInfo,
  call: AST.FunctionCall,
  argTypes: TypeInfo[],
  expectedReturn?: TypeInfo,
): CallSignature {
  const definition = info.definition;
  if (!definition) {
    const effective = buildEffectiveParams(info, call);
    let inferredTypeArgs: AST.TypeExpression[] | undefined;
    let params = effective.params;
    let returnType = info.returnType ?? unknownType;
    const genericParamNames = Array.isArray(info.genericParamNames) ? info.genericParamNames : [];
    if (genericParamNames.length > 0) {
      const genericNames = new Set(genericParamNames);
      const bindings = new Map<string, AST.TypeExpression>();
      const paramCount = Math.min(effective.params.length, argTypes.length);
      for (let index = 0; index < paramCount; index += 1) {
        const template = typeInfoToTypeExpression(effective.params[index]);
        if (!template) continue;
        inferTypeArgumentsFromTypeExpression(ctx, template, argTypes[index], genericNames, bindings);
      }
      if (
        expectedReturn &&
        expectedReturn.kind !== "unknown" &&
        !(expectedReturn.kind === "primitive" && expectedReturn.name === "void")
      ) {
        const template = typeInfoToTypeExpression(info.returnType ?? unknownType);
        if (template) {
          inferTypeArgumentsFromTypeExpression(ctx, template, expectedReturn, genericNames, bindings);
        }
      }
      if (bindings.size > 0) {
        const substituteExpression = (expr: AST.TypeExpression): AST.TypeExpression => {
          switch (expr.type) {
            case "SimpleTypeExpression": {
              const name = ctx.getIdentifierName(expr.name);
              return name && bindings.has(name) ? bindings.get(name)! : expr;
            }
            case "GenericTypeExpression":
              return {
                ...expr,
                base: substituteExpression(expr.base),
                arguments: (expr.arguments ?? []).map((arg) => (arg ? substituteExpression(arg) : arg)),
              };
            case "NullableTypeExpression":
              return { ...expr, innerType: substituteExpression(expr.innerType) };
            case "ResultTypeExpression":
              return { ...expr, innerType: substituteExpression(expr.innerType) };
            case "UnionTypeExpression":
              return { ...expr, members: (expr.members ?? []).map((member) => substituteExpression(member)) };
            case "FunctionTypeExpression":
              return {
                ...expr,
                paramTypes: (expr.paramTypes ?? []).map((param) => substituteExpression(param)),
                returnType: substituteExpression(expr.returnType),
              };
            default:
              return expr;
          }
        };
        const substituteTypeInfo = (typeInfo: TypeInfo): TypeInfo => {
          const expr = typeInfoToTypeExpression(typeInfo);
          if (!expr) return typeInfo;
          const substituted = substituteExpression(expr);
          return ctx.resolveTypeExpression(substituted);
        };
        inferredTypeArgs = genericParamNames.map((name) => bindings.get(name) ?? AST.wildcardTypeExpression());
        params = effective.params.map((param) => substituteTypeInfo(param ?? unknownType));
        returnType = substituteTypeInfo(returnType ?? unknownType);
      }
    }
    return {
      params,
      optionalLast: effective.optionalLast,
      returnType,
      inferredTypeArgs,
    };
  }
  const genericParamNames = Array.isArray(info.genericParamNames) ? info.genericParamNames : [];
  const genericNames = new Set(genericParamNames);
  const substitutions = new Map<string, TypeInfo>();
  const bindings = new Map<string, AST.TypeExpression>();

  if (info.methodSetSubstitutions) {
    for (const [key, value] of info.methodSetSubstitutions) {
      substitutions.set(key, value ?? unknownType);
      if (genericNames.has(key)) {
        const expr = typeInfoToTypeExpression(value);
        if (expr) {
          bindings.set(key, expr);
        }
      }
    }
  }

  const selfSubstitutions = buildSelfArgSubstitutions(ctx, info, call, argTypes, { includeSelf: false });
  if (selfSubstitutions) {
    for (const [key, value] of selfSubstitutions) {
      substitutions.set(key, value ?? unknownType);
      if (genericNames.has(key)) {
        const expr = typeInfoToTypeExpression(value);
        if (expr) {
          bindings.set(key, expr);
        }
      }
    }
  }

  if (call.callee.type === "MemberAccessExpression" && !substitutions.has("Self")) {
    const receiverType = ctx.inferExpression(call.callee.object);
    if (receiverType.kind !== "unknown") {
      substitutions.set("Self", receiverType);
    }
  }

  const explicitTypeArgs = Array.isArray(call.typeArguments) ? call.typeArguments : [];
  if (explicitTypeArgs.length > 0) {
    for (let index = 0; index < genericParamNames.length; index += 1) {
      const name = genericParamNames[index];
      const argExpr = explicitTypeArgs[index];
      if (name && argExpr) {
        bindings.set(name, argExpr);
      }
    }
  } else if (genericParamNames.length > 0) {
    const params = Array.isArray(definition.params) ? definition.params : [];
    const skipSelf = info.hasImplicitSelf && call.callee.type === "MemberAccessExpression" ? 1 : 0;
    const paramCount = Math.min(Math.max(0, params.length - skipSelf), argTypes.length);
    for (let index = 0; index < paramCount; index += 1) {
      const param = params[index + skipSelf];
      if (!param?.paramType) continue;
      inferTypeArgumentsFromTypeExpression(ctx, param.paramType, argTypes[index], genericNames, bindings);
    }
    if (
      expectedReturn &&
      expectedReturn.kind !== "unknown" &&
      !(expectedReturn.kind === "primitive" && expectedReturn.name === "void") &&
      definition.returnType
    ) {
      const needsInference = genericParamNames.some((name) => !bindings.has(name));
      if (needsInference) {
        inferTypeArgumentsFromTypeExpression(ctx, definition.returnType, expectedReturn, genericNames, bindings);
      }
    }
  }

  for (const [name, expr] of bindings) {
    substitutions.set(name, ctx.resolveTypeExpression(expr));
  }

  let inferredTypeArgs: AST.TypeExpression[] | undefined;
  if (explicitTypeArgs.length === 0 && genericParamNames.length > 0) {
    const hasBinding = genericParamNames.some((name) => bindings.has(name));
    if (hasBinding) {
      inferredTypeArgs = genericParamNames.map((name) => bindings.get(name) ?? AST.wildcardTypeExpression());
    }
  }

  let params = Array.isArray(definition.params)
    ? definition.params.map((param) => ctx.resolveTypeExpression(param?.paramType, substitutions))
    : [];
  if (info.hasImplicitSelf && definition.isMethodShorthand) {
    const selfParam = substitutions.get("Self") ?? info.parameters?.[0] ?? unknownType;
    params = [selfParam, ...params];
  }
  if (info.hasImplicitSelf && call.callee.type === "MemberAccessExpression" && params.length > 0) {
    params = params.slice(1);
  }
  const optionalLast = params.length > 0 && params[params.length - 1]?.kind === "nullable";
  const returnType = definition.returnType
    ? ctx.resolveTypeExpression(definition.returnType, substitutions)
    : info.returnType ?? unknownType;

  return { params, optionalLast, returnType, inferredTypeArgs };
}

export function buildSelfArgSubstitutions(
  ctx: FunctionCallContext,
  info: FunctionInfo,
  call: AST.FunctionCall,
  argTypes: TypeInfo[],
  options?: { includeSelf?: boolean },
): Map<string, TypeInfo> | null {
  if (!info.hasImplicitSelf || call.callee?.type === "MemberAccessExpression") {
    return null;
  }
  const selfType = argTypes[0];
  if (!selfType || selfType.kind === "unknown") {
    return null;
  }
  const substitutions = new Map<string, TypeInfo>();
  if (options?.includeSelf) {
    substitutions.set("Self", selfType);
  }
  if (selfType.kind !== "struct" || !selfType.name) {
    return substitutions.size > 0 ? substitutions : null;
  }
  const def = ctx.structDefinitions.get(selfType.name);
  if (!def?.genericParams?.length) {
    return substitutions.size > 0 ? substitutions : null;
  }
  def.genericParams.forEach((param, idx) => {
    const paramName = ctx.getIdentifierName(param?.name);
    if (!paramName) return;
    const arg = Array.isArray(selfType.typeArguments) ? selfType.typeArguments[idx] : undefined;
    if (arg) {
      substitutions.set(paramName, arg);
    }
  });
  return substitutions.size > 0 ? substitutions : null;
}

function inferTypeArgumentsFromTypeExpression(
  ctx: FunctionCallContext,
  template: AST.TypeExpression | null | undefined,
  actualType: TypeInfo,
  genericNames: Set<string>,
  bindings: Map<string, AST.TypeExpression>,
): void {
  if (!template || !actualType || actualType.kind === "unknown") {
    return;
  }
  const actual = typeInfoToTypeExpression(actualType);
  if (!actual) {
    return;
  }
  const snapshot = new Map(bindings);
  if (!matchTypeExpressionTemplate(ctx, template, actual, genericNames, snapshot)) {
    return;
  }
  bindings.clear();
  for (const [key, value] of snapshot) {
    bindings.set(key, value);
  }
}

function matchTypeExpressionTemplate(
  ctx: FunctionCallContext,
  template: AST.TypeExpression,
  actual: AST.TypeExpression,
  genericNames: Set<string>,
  bindings: Map<string, AST.TypeExpression>,
): boolean {
  if (template.type === "WildcardTypeExpression" || actual.type === "WildcardTypeExpression") {
    return true;
  }
  if (template.type === "SimpleTypeExpression") {
    const name = ctx.getIdentifierName(template.name);
    if (name && genericNames.has(name)) {
      const existing = bindings.get(name);
      if (existing) {
        return ctx.typeExpressionsEquivalent(existing, actual);
      }
      bindings.set(name, actual);
      return true;
    }
    return ctx.typeExpressionsEquivalent(template, actual);
  }
  if (template.type === "GenericTypeExpression") {
    if (actual.type !== "GenericTypeExpression") {
      return false;
    }
    if (!matchTypeExpressionTemplate(ctx, template.base, actual.base, genericNames, bindings)) {
      return false;
    }
    const templateArgs = template.arguments ?? [];
    const actualArgs = actual.arguments ?? [];
    if (templateArgs.length !== actualArgs.length) {
      return false;
    }
    for (let index = 0; index < templateArgs.length; index += 1) {
      if (!matchTypeExpressionTemplate(ctx, templateArgs[index]!, actualArgs[index]!, genericNames, bindings)) {
        return false;
      }
    }
    return true;
  }
  if (template.type === "NullableTypeExpression") {
    if (actual.type !== "NullableTypeExpression") {
      return false;
    }
    return matchTypeExpressionTemplate(ctx, template.innerType, actual.innerType, genericNames, bindings);
  }
  if (template.type === "ResultTypeExpression") {
    if (actual.type !== "ResultTypeExpression") {
      return false;
    }
    return matchTypeExpressionTemplate(ctx, template.innerType, actual.innerType, genericNames, bindings);
  }
  if (template.type === "FunctionTypeExpression") {
    if (actual.type !== "FunctionTypeExpression") {
      return false;
    }
    const templateParams = template.paramTypes ?? [];
    const actualParams = actual.paramTypes ?? [];
    if (templateParams.length !== actualParams.length) {
      return false;
    }
    for (let index = 0; index < templateParams.length; index += 1) {
      if (!matchTypeExpressionTemplate(ctx, templateParams[index]!, actualParams[index]!, genericNames, bindings)) {
        return false;
      }
    }
    return matchTypeExpressionTemplate(ctx, template.returnType, actual.returnType, genericNames, bindings);
  }
  if (template.type === "UnionTypeExpression") {
    const templateMembers = template.members ?? [];
    if (actual.type === "UnionTypeExpression") {
      const actualMembers = actual.members ?? [];
      if (templateMembers.length !== actualMembers.length) {
        return false;
      }
      for (let index = 0; index < templateMembers.length; index += 1) {
        if (!matchTypeExpressionTemplate(ctx, templateMembers[index]!, actualMembers[index]!, genericNames, bindings)) {
          return false;
        }
      }
      return true;
    }
    for (const member of templateMembers) {
      const snapshot = new Map(bindings);
      if (matchTypeExpressionTemplate(ctx, member!, actual, genericNames, snapshot)) {
        bindings.clear();
        for (const [key, value] of snapshot) {
          bindings.set(key, value);
        }
        return true;
      }
    }
    return false;
  }
  return ctx.typeExpressionsEquivalent(template, actual);
}

export function buildEffectiveParams(info: FunctionInfo, call: AST.FunctionCall): {
  params: TypeInfo[];
  optionalLast: boolean;
} {
  const rawParams = Array.isArray(info.parameters) ? info.parameters : [];
  const implicitSelf =
    Boolean(info.structName && info.hasImplicitSelf) &&
    call.callee?.type === "MemberAccessExpression" &&
    rawParams.length > 0;
  const params = implicitSelf ? rawParams.slice(1) : rawParams;
  const optionalLast = params.length > 0 && params[params.length - 1]?.kind === "nullable";
  return { params, optionalLast };
}

export function arityMatches(params: TypeInfo[], argCount: number, optionalLast: boolean): boolean {
  return params.length === argCount || (optionalLast && argCount === params.length - 1);
}

export function dropOptionalParam(params: TypeInfo[], argCount: number, optionalLast: boolean): TypeInfo[] {
  if (optionalLast && argCount === params.length - 1) {
    return params.slice(0, params.length - 1);
  }
  return params;
}
