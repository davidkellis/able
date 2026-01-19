import * as AST from "../../ast";
import { formatType, unknownType, type TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import {
  ambiguousImplementationDetail,
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
  matchImplementationTarget,
  typeImplementsInterface,
} from "./implementations";
import { mergeBranchTypes as mergeBranchTypesHelper } from "./expressions";
import type { StatementContext } from "./expression-context";
import { typeInfoToTypeExpression } from "./type-expression-utils";
import type { FunctionInfo, ImplementationObligation, InterfaceCheckResult } from "./types";
import { expectsSelfType } from "./implementation-collection";

type FunctionCallContext = {
  implementationContext: ImplementationContext;
  functionInfos: Map<string, FunctionInfo[]>;
  structDefinitions: Map<string, AST.StructDefinition>;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  resolveTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, TypeInfo>): TypeInfo;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean;
  typeExpressionsEquivalent(
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
  ): boolean;
  report(message: string, node?: AST.Node | null): void;
  handlePackageMemberAccess(expression: AST.MemberAccessExpression): TypeInfo | null;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  checkBuiltinCallContext(name: string | undefined, call: AST.FunctionCall): void;
  getBuiltinCallName(callee: AST.Expression | undefined | null): string | undefined;
  getTypeParamConstraints(name: string): AST.TypeExpression[];
  typeImplementsInterface?: (
    type: TypeInfo,
    interfaceName: string,
    expectedArgs?: string[],
  ) => InterfaceCheckResult;
  statementContext: StatementContext;
};

function reportAmbiguousInterfaceMethod(
  ctx: FunctionCallContext,
  receiverType: TypeInfo,
  methodName: string,
  node: AST.Node,
): boolean {
  if (!receiverType || receiverType.kind === "unknown") {
    return false;
  }
  const implementations = ctx.implementationContext.getImplementationRecords?.();
  if (!implementations) {
    return false;
  }
  const interfaces = new Set<string>();
  for (const record of implementations) {
    if (!record?.definition || !Array.isArray(record.definition.definitions)) continue;
    const hasMethod = record.definition.definitions.some(
      (fn) => fn?.type === "FunctionDefinition" && fn.id?.name === methodName,
    );
    if (hasMethod) {
      interfaces.add(record.interfaceName);
    }
  }
  for (const interfaceName of interfaces) {
    const detail = ambiguousImplementationDetail(ctx.implementationContext, receiverType, interfaceName);
    if (detail) {
      ctx.report(`typechecker: ${detail}`, node);
      return true;
    }
    const result = typeImplementsInterface(ctx.implementationContext, receiverType, interfaceName);
    if (!result.ok && result.detail && result.detail.includes("ambiguous implementations")) {
      ctx.report(`typechecker: ${result.detail}`, node);
      return true;
    }
  }
  return false;
}

export function checkFunctionCall(ctx: FunctionCallContext, call: AST.FunctionCall, expectedReturn?: TypeInfo): void {
  if (call.callee.type === "MemberAccessExpression" && call.callee.member?.type === "Identifier") {
    const receiverType = ctx.inferExpression(call.callee.object);
    if (reportAmbiguousInterfaceMethod(ctx, receiverType, call.callee.member.name, call)) {
      return;
    }
  }
  const builtinName = ctx.getBuiltinCallName(call.callee);
  const args = Array.isArray(call.arguments) ? call.arguments : [];
  const argTypes = args.map((arg) => ctx.inferExpression(arg));
  ctx.checkBuiltinCallContext(builtinName, call);
  let candidates: FunctionInfo[] = [];
  let callArgs = args;
  let callArgTypes = argTypes;
  if (call.callee.type === "MemberAccessExpression") {
    const memberResolution = resolveMemberAccessCandidates(ctx, call.callee, args.length);
    if (memberResolution.fieldError) {
      ctx.report(`typechecker: ${memberResolution.fieldError}`, call.callee);
      return;
    }
    candidates = memberResolution.candidates;
  } else {
    candidates = resolveFunctionInfos(ctx, call.callee);
  }
  if (!candidates.length) {
    const fnCandidate = resolveFunctionTypeCandidate(ctx, call.callee);
    if (fnCandidate) {
      candidates = [fnCandidate];
    }
  }
  if (!candidates.length) {
    const applyMatch = resolveApplyInterface(ctx, call.callee);
    if (applyMatch) {
      let params = applyMatch.paramTypes ?? [];
      const optionalLast = params.length > 0 && params[params.length - 1]?.kind === "nullable";
      if (params.length !== callArgs.length && !(optionalLast && callArgs.length === params.length - 1)) {
        ctx.report(`typechecker: Apply.apply expects ${params.length} arguments, got ${callArgs.length}`, call);
      }
      if (optionalLast && callArgs.length === params.length - 1) {
        params = params.slice(0, params.length - 1);
      }
      const compareCount = Math.min(params.length, callArgTypes.length);
      for (let index = 0; index < compareCount; index += 1) {
        const expected = params[index];
        const actual = callArgTypes[index];
        if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
          continue;
        }
        const literalMessage = ctx.describeLiteralMismatch(actual, expected);
        if (literalMessage) {
          ctx.report(literalMessage, callArgs[index] ?? call);
          continue;
        }
        if (!ctx.isTypeAssignable(actual, expected)) {
          ctx.report(
            `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
            callArgs[index] ?? call,
          );
        }
      }
      return;
    }
    const calleeType = ctx.inferExpression(call.callee);
    if (calleeType.kind !== "unknown") {
      ctx.report(
        `typechecker: cannot call non-callable value ${formatType(calleeType)} (missing Apply implementation)`,
        call.callee ?? call,
      );
    }
    return;
  }
  const resolution = selectBestOverload(ctx, candidates, call, callArgs, callArgTypes, expectedReturn);
  if (resolution.kind === "no-match") {
    const partial = findPartialApplication(candidates, callArgs.length, call);
    if (partial) {
      return;
    }
    if (candidates.length === 1) {
      const effective = buildEffectiveParams(candidates[0], call);
      reportArgumentDiagnostics(ctx, candidates[0], effective.params, effective.optionalLast, call, callArgs, callArgTypes);
    } else {
      const name = formatCalleeLabel(call.callee);
      ctx.report(`typechecker: no overloads of ${name ?? "function"} match provided arguments`, call);
    }
    return;
  }
  if (resolution.kind === "ambiguous") {
    const name = formatCalleeLabel(call.callee);
    ctx.report(`typechecker: ambiguous overload for ${name ?? "function"}`, call);
    return;
  }
  const { info, params, optionalLast, inferredTypeArgs } = resolution;
  const ok = reportArgumentDiagnostics(ctx, info, params, optionalLast, call, callArgs, callArgTypes);
  if (ok) {
    if ((!call.typeArguments || call.typeArguments.length === 0) && inferredTypeArgs && inferredTypeArgs.length > 0) {
      call.typeArguments = inferredTypeArgs;
    }
    const selfSubstitutions = buildSelfArgSubstitutions(ctx, info, call, callArgTypes, { includeSelf: true });
    enforceFunctionConstraintsHelper(ctx.implementationContext, info, call, selfSubstitutions ?? undefined);
  }
}

export function inferFunctionCallReturnType(
  ctx: FunctionCallContext,
  call: AST.FunctionCall,
  expectedReturn?: TypeInfo,
): TypeInfo {
  const args = Array.isArray(call.arguments) ? call.arguments : [];
  const argTypes = args.map((arg) => ctx.inferExpression(arg));
  let infos: FunctionInfo[] = [];
  let callArgs = args;
  let callArgTypes = argTypes;
  if (call.callee.type === "MemberAccessExpression") {
    const memberResolution = resolveMemberAccessCandidates(ctx, call.callee, args.length);
    if (memberResolution.fieldError) {
      return unknownType;
    }
    infos = memberResolution.candidates;
  } else {
    infos = resolveFunctionInfos(ctx, call.callee);
  }
  if (!infos.length) {
    const fnCandidate = resolveFunctionTypeCandidate(ctx, call.callee);
    if (fnCandidate) {
      infos = [fnCandidate];
    }
  }
  if (!infos.length) {
    const applyMatch = resolveApplyInterface(ctx, call.callee);
    return applyMatch?.returnType ?? unknownType;
  }
  const resolution = selectBestOverload(ctx, infos, call, callArgs, callArgTypes, expectedReturn);
  if (resolution.kind === "match") {
    return resolution.returnType ?? unknownType;
  }
  if (resolution.kind === "no-match") {
    const partial = findPartialApplication(infos, callArgs.length, call);
    if (partial) {
      return {
        kind: "function",
        parameters: partial.remaining,
        returnType: partial.returnType ?? unknownType,
      };
    }
  }
  const returnTypes = infos
    .map((info) => info.returnType ?? unknownType)
    .filter((type) => type && type.kind !== "unknown");
  if (!returnTypes.length) {
    return unknownType;
  }
  return mergeBranchTypesHelper(ctx.statementContext, returnTypes);
}

type OverloadResolution =
  | {
      kind: "match";
      info: FunctionInfo;
      params: TypeInfo[];
      optionalLast: boolean;
      returnType: TypeInfo;
      inferredTypeArgs?: AST.TypeExpression[];
    }
  | { kind: "ambiguous"; infos: FunctionInfo[] }
  | { kind: "no-match" };

type CallSignature = {
  params: TypeInfo[];
  optionalLast: boolean;
  returnType: TypeInfo;
  inferredTypeArgs?: AST.TypeExpression[];
};

function instantiateCallSignature(
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

function buildSelfArgSubstitutions(
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

function selectBestOverload(
  ctx: FunctionCallContext,
  infos: FunctionInfo[],
  call: AST.FunctionCall,
  args: AST.Expression[],
  argTypes: TypeInfo[],
  expectedReturn?: TypeInfo,
): OverloadResolution {
  const compatible: Array<{
    info: FunctionInfo;
    params: TypeInfo[];
    optionalLast: boolean;
    returnType: TypeInfo;
    inferredTypeArgs?: AST.TypeExpression[];
    score: number;
    priority: number;
  }> = [];
  for (const info of infos) {
    const instantiated = instantiateCallSignature(ctx, info, call, argTypes, expectedReturn);
    if (!arityMatches(instantiated.params, args.length, instantiated.optionalLast)) {
      continue;
    }
    const params = dropOptionalParam(instantiated.params, args.length, instantiated.optionalLast);
    const score = scoreCompatibility(ctx, params, argTypes);
    if (score < 0) {
      continue;
    }
    const priority = typeof info.methodResolutionPriority === "number" ? info.methodResolutionPriority : 0;
    compatible.push({
      info,
      params,
      optionalLast: instantiated.optionalLast,
      returnType: instantiated.returnType ?? unknownType,
      inferredTypeArgs: instantiated.inferredTypeArgs,
      score: score - (instantiated.optionalLast && params.length !== instantiated.params.length ? 0.5 : 0),
      priority,
    });
  }
  if (!compatible.length) {
    return { kind: "no-match" };
  }
  compatible.sort((a, b) => {
    if (a.score !== b.score) return b.score - a.score;
    return b.priority - a.priority;
  });
  const best = compatible[0];
  const tied = compatible.filter((entry) => entry.score === best.score && entry.priority === best.priority);
  if (tied.length > 1) {
    return { kind: "ambiguous", infos: tied.map((entry) => entry.info) };
  }
  return {
    kind: "match",
    info: best.info,
    params: best.params,
    optionalLast: best.optionalLast,
    returnType: best.returnType ?? unknownType,
    inferredTypeArgs: best.inferredTypeArgs,
  };
}

function buildEffectiveParams(info: FunctionInfo, call: AST.FunctionCall): {
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

function arityMatches(params: TypeInfo[], argCount: number, optionalLast: boolean): boolean {
  return params.length === argCount || (optionalLast && argCount === params.length - 1);
}

function dropOptionalParam(params: TypeInfo[], argCount: number, optionalLast: boolean): TypeInfo[] {
  if (optionalLast && argCount === params.length - 1) {
    return params.slice(0, params.length - 1);
  }
  return params;
}

function scoreCompatibility(ctx: FunctionCallContext, params: TypeInfo[], argTypes: TypeInfo[]): number {
  const compareCount = Math.min(params.length, argTypes.length);
  let score = 0;
  for (let index = 0; index < compareCount; index += 1) {
    const expected = params[index];
    const actual = argTypes[index];
    if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
      continue;
    }
    const literalMessage = ctx.describeLiteralMismatch(actual, expected);
    if (literalMessage) {
      return -1;
    }
    if (!ctx.isTypeAssignable(actual, expected)) {
      const interfaceArgs =
        expected.kind === "interface" && Array.isArray(expected.typeArguments)
          ? expected.typeArguments.map((arg) => (arg?.kind === "unknown" ? "Unknown" : formatType(arg)))
          : [];
      if (
        expected.kind === "interface" &&
        typeImplementsInterface(ctx.implementationContext, actual, expected.name, interfaceArgs).ok
      ) {
        score += 1;
        continue;
      }
      return -1;
    }
    score += expected.kind === "nullable" ? 1 : 2;
  }
  return score;
}

function reportArgumentDiagnostics(
  ctx: FunctionCallContext,
  info: FunctionInfo,
  params: TypeInfo[],
  optionalLast: boolean,
  call: AST.FunctionCall,
  args: AST.Expression[],
  argTypes: TypeInfo[],
): boolean {
  if (params.length !== args.length && !(optionalLast && args.length === params.length - 1)) {
    ctx.report(`typechecker: function expects ${params.length} arguments, got ${args.length}`, call);
    return false;
  }
  const compareParams = dropOptionalParam(params, args.length, optionalLast);
  const compareCount = Math.min(compareParams.length, argTypes.length);
  for (let index = 0; index < compareCount; index += 1) {
    const expected = compareParams[index];
    const actual = argTypes[index];
    if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
      continue;
    }
    const literalMessage = ctx.describeLiteralMismatch(actual, expected);
    if (literalMessage) {
      ctx.report(literalMessage, args[index] ?? call);
      return false;
    }
    if (!ctx.isTypeAssignable(actual, expected)) {
      const interfaceArgs =
        expected.kind === "interface" && Array.isArray(expected.typeArguments)
          ? expected.typeArguments.map((arg) => (arg?.kind === "unknown" ? "Unknown" : formatType(arg)))
          : [];
      if (
        expected.kind === "interface" &&
        typeImplementsInterface(ctx.implementationContext, actual, expected.name, interfaceArgs).ok
      ) {
        continue;
      }
      if (
        expected.kind === "interface" &&
        ctx.implementationContext.getImplementationBucket?.(formatType(actual))?.some(
          (record) => record.interfaceName === expected.name,
        )
      ) {
        continue;
      }
      ctx.report(
        `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
        args[index] ?? call,
      );
      return false;
    }
  }
  return true;
}

function findPartialApplication(
  infos: FunctionInfo[],
  argCount: number,
  call: AST.FunctionCall,
): { remaining: TypeInfo[]; returnType: TypeInfo } | null {
  let best: { remaining: TypeInfo[]; returnType: TypeInfo } | null = null;
  for (const info of infos) {
    const effective = buildEffectiveParams(info, call);
    const requiredCount = effective.optionalLast ? Math.max(0, effective.params.length - 1) : effective.params.length;
    if (argCount < requiredCount) {
      const remaining = effective.params.slice(argCount);
      if (!best || remaining.length < best.remaining.length) {
        best = { remaining, returnType: info.returnType ?? unknownType };
      }
    }
  }
  return best;
}

function formatCalleeLabel(callee: AST.Expression | undefined | null): string | null {
  if (!callee) return null;
  if (callee.type === "Identifier") {
    return callee.name;
  }
  if (callee.type === "MemberAccessExpression") {
    const member = callee.member;
    if (member?.type === "Identifier") {
      return member.name;
    }
  }
  return null;
}

function resolveFunctionInfos(ctx: FunctionCallContext, callee: AST.Expression | undefined | null): FunctionInfo[] {
  if (!callee) return [];
  if (callee.type === "MemberAccessExpression") {
    const resolution = resolveMemberAccessCandidates(ctx, callee);
    if (resolution.fieldError) return [];
    return resolution.candidates;
  }
  if (callee.type === "Identifier") {
    const infos = ctx.functionInfos.get(callee.name) ?? [];
    const nonMethodInfos = infos.filter((info) => !info.structName);
    return nonMethodInfos.length > 0 ? nonMethodInfos : infos;
  }
  return [];
}

type MemberAccessResolution = { candidates: FunctionInfo[]; fieldError?: string };

type ReceiverLookup = {
  lookupType: TypeInfo;
  label: string;
  memberName: string;
  isTypeReference?: boolean;
  typeQualifier?: string | null;
};

function resolveMemberAccessCandidates(
  ctx: FunctionCallContext,
  callee: AST.MemberAccessExpression,
  argCount?: number,
): MemberAccessResolution {
  if (ctx.handlePackageMemberAccess(callee)) {
    return { candidates: [] };
  }
  const memberName = ctx.getIdentifierName(callee.member);
  if (!memberName) return { candidates: [] };
  const receiver = buildReceiverLookup(ctx, callee, memberName);
  if (!receiver) return { candidates: [] };
  const fieldResolution = resolveCallableFieldCandidate(ctx, receiver, memberName, argCount);
  if (fieldResolution?.callable) {
    return { candidates: [fieldResolution.callable] };
  }
  const methodCandidates = collectUnifiedMemberCandidates(ctx, receiver, memberName);
  if (methodCandidates.length > 0) {
    return { candidates: methodCandidates };
  }
  if (fieldResolution?.nonCallable) {
    return { candidates: [], fieldError: `field '${memberName}' is not callable` };
  }
  return { candidates: [] };
}

function buildReceiverLookup(
  ctx: FunctionCallContext,
  callee: AST.MemberAccessExpression,
  memberName: string,
): ReceiverLookup | null {
  const objectName = callee.object?.type === "Identifier" ? callee.object.name : null;
  const isTypeName =
    objectName ? ctx.structDefinitions.has(objectName) || ctx.implementationContext.hasInterfaceDefinition?.(objectName) : false;
  const hasObjectBinding = objectName ? ctx.statementContext.hasBinding?.(objectName) ?? false : true;
  let objectType = ctx.inferExpression(callee.object);
  if (
    objectType.kind !== "struct" &&
    callee.object?.type === "Identifier" &&
    callee.object.name &&
    ctx.structDefinitions.has(callee.object.name)
  ) {
    objectType = {
      kind: "struct",
      name: callee.object.name,
      typeArguments: [],
      definition: ctx.structDefinitions.get(callee.object.name),
    };
  }
  const isTypeReference =
    callee.object?.type === "Identifier" &&
    (objectType.kind === "struct" || objectType.kind === "interface" || objectType.kind === "type_parameter") &&
    (!hasObjectBinding || isTypeName || objectType.kind === "type_parameter");
  if (objectType.kind === "array") {
    const lookupType: TypeInfo = {
      kind: "struct",
      name: "Array",
      typeArguments: [objectType.element ?? unknownType],
    };
    return { lookupType, label: formatType(lookupType), memberName, typeQualifier: "Array" };
  }
  const typeQualifier =
    objectType.kind === "struct" || objectType.kind === "interface" || objectType.kind === "type_parameter"
      ? objectType.name
      : null;
  return { lookupType: objectType, label: formatType(objectType), memberName, isTypeReference, typeQualifier };
}

function resolveCallableFieldCandidate(
  ctx: FunctionCallContext,
  receiver: ReceiverLookup,
  memberName: string,
  argCount?: number,
): { callable?: FunctionInfo; nonCallable?: TypeInfo } | null {
  const fieldType = resolveStructFieldType(ctx, receiver, memberName);
  if (!fieldType) return null;
  if (isCallableFieldType(ctx, fieldType)) {
    const info = functionInfoFromFieldType(fieldType, `${receiver.label}.${memberName}`, argCount);
    if (info) {
      return { callable: info };
    }
  }
  return { nonCallable: fieldType };
}

function resolveStructFieldType(ctx: FunctionCallContext, receiver: ReceiverLookup, memberName: string): TypeInfo | null {
  if (receiver.lookupType.kind !== "struct") return null;
  const definition =
    receiver.lookupType.definition ?? (receiver.lookupType.name ? ctx.structDefinitions.get(receiver.lookupType.name) : undefined);
  if (!definition || !Array.isArray(definition.fields)) return null;
  const substitutions = new Map<string, TypeInfo>();
  const generics = definition.genericParams ?? [];
  for (let i = 0; i < generics.length; i += 1) {
    const gpName = ctx.getIdentifierName(generics[i]?.name);
    if (!gpName) continue;
    const arg = Array.isArray(receiver.lookupType.typeArguments) ? receiver.lookupType.typeArguments[i] : undefined;
    substitutions.set(gpName, arg ?? unknownType);
  }
  for (const field of definition.fields) {
    if (!field) continue;
    const name = ctx.getIdentifierName(field.name);
    if (!name || name !== memberName) continue;
    return ctx.resolveTypeExpression(field.fieldType, substitutions);
  }
  return null;
}

function isCallableFieldType(ctx: FunctionCallContext, type: TypeInfo): boolean {
  if (!type || type.kind === "unknown") return false;
  if (type.kind === "function") return true;
  if (type.kind === "interface" && type.name === "Apply") return true;
  return Boolean(ctx.typeImplementsInterface?.(type, "Apply")?.ok);
}

function functionInfoFromFieldType(type: TypeInfo, label: string, argCount?: number): FunctionInfo | null {
  if (type.kind !== "function") {
    const params = argCount && argCount > 0 ? new Array<unknown>(argCount).fill(unknownType) as TypeInfo[] : [];
    return {
      name: label,
      fullName: label,
      parameters: params,
      returnType: unknownType,
      genericConstraints: [],
      genericParamNames: [],
      whereClause: [],
    };
  }
  const params = Array.isArray(type.parameters) ? type.parameters : [];
  return {
    name: label,
    fullName: label,
    parameters: params,
    genericConstraints: [],
    genericParamNames: [],
    whereClause: [],
    returnType: type.returnType ?? unknownType,
  };
}

function signatureHasImplicitSelf(signature: AST.FunctionSignature): boolean {
  if (!Array.isArray(signature.params) || signature.params.length === 0) {
    return false;
  }
  const first = signature.params[0];
  if (!first) return false;
  if (!first.paramType && first.name?.type === "Identifier") {
    return first.name.name?.toLowerCase() === "self";
  }
  return expectsSelfType(first.paramType);
}

function collectSignatureWhereObligations(
  ctx: FunctionCallContext,
  signature: AST.FunctionSignature,
): ImplementationObligation[] {
  const obligations: ImplementationObligation[] = [];
  const appendObligation = (
    typeParam: string | null,
    interfaceType: AST.TypeExpression | null | undefined,
    context: string,
  ) => {
    const interfaceName = ctx.implementationContext.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
  };
  if (Array.isArray(signature.genericParams)) {
    for (const param of signature.genericParams) {
      const paramName = ctx.getIdentifierName(param?.name);
      if (!paramName || !Array.isArray(param?.constraints)) continue;
      for (const constraint of param.constraints) {
        appendObligation(paramName, constraint?.interfaceType, "generic constraint");
      }
    }
  }
  if (Array.isArray(signature.whereClause)) {
    for (const clause of signature.whereClause) {
      const typeParamName = ctx.getIdentifierName(clause?.typeParam);
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause");
      }
    }
  }
  return obligations;
}

function buildInterfaceMethodInfo(
  ctx: FunctionCallContext,
  receiverType: TypeInfo,
  interfaceName: string,
  interfaceDef: AST.InterfaceDefinition,
  interfaceArgs: TypeInfo[],
  signature: AST.FunctionSignature,
  priority: number,
): FunctionInfo {
  const substitutions = new Map<string, TypeInfo>();
  substitutions.set("Self", receiverType);
  if (Array.isArray(interfaceDef.genericParams)) {
    interfaceDef.genericParams.forEach((param, idx) => {
      const name = ctx.getIdentifierName(param?.name);
      if (!name) return;
      const arg = interfaceArgs[idx] ?? unknownType;
      substitutions.set(name, arg);
    });
  }
  const methodGenericNames = Array.isArray(signature.genericParams)
    ? signature.genericParams
        .map((param) => ctx.getIdentifierName(param?.name))
        .filter((name): name is string => Boolean(name))
    : [];
  const signatureSubstitutions = new Map(substitutions);
  methodGenericNames.forEach((name) => {
    signatureSubstitutions.set(name, { kind: "type_parameter", name });
  });
  const parameterTypes = Array.isArray(signature.params)
    ? signature.params.map((param) => ctx.resolveTypeExpression(param?.paramType, signatureSubstitutions))
    : [];
  const returnType = ctx.resolveTypeExpression(signature.returnType, signatureSubstitutions);
  const genericConstraints: FunctionInfo["genericConstraints"] = [];
  if (Array.isArray(signature.genericParams)) {
    for (const param of signature.genericParams) {
      const paramName = param.name?.name ?? "T";
      if (!Array.isArray(param.constraints)) continue;
      for (const constraint of param.constraints) {
        const ifaceName = ctx.implementationContext.getInterfaceNameFromConstraint(constraint);
        if (!ifaceName) continue;
        genericConstraints.push({
          paramName,
          interfaceName: ifaceName,
          interfaceDefined: ctx.implementationContext.hasInterfaceDefinition(ifaceName),
          interfaceType: constraint.interfaceType,
        });
      }
    }
  }
  return {
    name: signature.name?.name ?? "<anonymous>",
    fullName: `${interfaceName}::${signature.name?.name ?? "<anonymous>"}`,
    definition: undefined,
    structName: receiverType.kind === "interface" ? receiverType.name : undefined,
    hasImplicitSelf: signatureHasImplicitSelf(signature),
    isTypeQualified: !signatureHasImplicitSelf(signature),
    typeQualifier: !signatureHasImplicitSelf(signature) ? interfaceName : undefined,
    methodResolutionPriority: priority,
    parameters: parameterTypes,
    genericConstraints,
    genericParamNames: methodGenericNames,
    whereClause: collectSignatureWhereObligations(ctx, signature),
    methodSetSubstitutions: Array.from(substitutions.entries()),
    returnType: returnType ?? unknownType,
  };
}

function collectInterfaceMethodCandidates(
  ctx: FunctionCallContext,
  receiver: ReceiverLookup,
  memberName: string,
): FunctionInfo[] {
  const receiverType = receiver.lookupType;
  const candidates: FunctionInfo[] = [];
  const seen = new Set<string>();
  const addCandidate = (
    interfaceName: string,
    interfaceDef: AST.InterfaceDefinition,
    interfaceArgs: TypeInfo[],
    priority: number,
  ) => {
    const signature = interfaceDef.signatures?.find(
      (sig) => sig?.name?.name === memberName,
    );
    if (!signature) return;
    if (!signatureHasImplicitSelf(signature)) return;
    const key = `${interfaceName}::${memberName}::${interfaceArgs.map(formatType).join("|")}`;
    if (seen.has(key)) return;
    seen.add(key);
    candidates.push(
      buildInterfaceMethodInfo(ctx, receiverType, interfaceName, interfaceDef, interfaceArgs, signature, priority),
    );
  };

  if (receiverType.kind === "interface") {
    const ifaceDef = ctx.implementationContext.getInterfaceDefinition(receiverType.name);
    if (ifaceDef) {
      addCandidate(receiverType.name, ifaceDef, receiverType.typeArguments ?? [], -1);
      const appendBases = (def: AST.InterfaceDefinition, args: TypeInfo[]): void => {
        if (!def?.baseInterfaces || def.baseInterfaces.length === 0) return;
        const substitutions = new Map<string, TypeInfo>();
        substitutions.set("Self", receiverType);
        if (Array.isArray(def.genericParams)) {
          def.genericParams.forEach((param, idx) => {
            const name = ctx.getIdentifierName(param?.name);
            if (!name) return;
            substitutions.set(name, args[idx] ?? unknownType);
          });
        }
        for (const base of def.baseInterfaces) {
          const baseInfo = ctx.resolveTypeExpression(base, substitutions);
          if (!baseInfo || baseInfo.kind !== "interface") continue;
          const baseDef = ctx.implementationContext.getInterfaceDefinition(baseInfo.name);
          if (!baseDef) continue;
          addCandidate(baseInfo.name, baseDef, baseInfo.typeArguments ?? [], -2);
          appendBases(baseDef, baseInfo.typeArguments ?? []);
        }
      };
      appendBases(ifaceDef, receiverType.typeArguments ?? []);
    }
  }

  for (const record of ctx.implementationContext.getImplementationRecords()) {
    if (record.definition?.implName?.name) {
      continue;
    }
    const paramNames = new Set(record.genericParams);
    const substitutions = new Map<string, TypeInfo>();
    substitutions.set("Self", receiverType);
    if (!matchImplementationTarget(ctx.implementationContext, receiverType, record.target, paramNames, substitutions)) {
      continue;
    }
    for (const name of paramNames) {
      if (!substitutions.has(name)) {
        substitutions.set(name, unknownType);
      }
    }
    const ifaceDef = ctx.implementationContext.getInterfaceDefinition(record.interfaceName);
    if (!ifaceDef) continue;
    const interfaceArgs = record.interfaceArgs.map((arg) => ctx.resolveTypeExpression(arg, substitutions));
    addCandidate(record.interfaceName, ifaceDef, interfaceArgs, -1);
  }

  return candidates;
}

function collectTypeParamMethodCandidates(
  ctx: FunctionCallContext,
  receiver: ReceiverLookup,
  memberName: string,
): FunctionInfo[] {
  const receiverType = receiver.lookupType;
  if (receiverType.kind !== "type_parameter") {
    return [];
  }
  const constraints = ctx.getTypeParamConstraints(receiverType.name) ?? [];
  if (!constraints.length) {
    return [];
  }
  const candidates: FunctionInfo[] = [];
  for (const constraint of constraints) {
    const interfaceName = ctx.implementationContext.getInterfaceNameFromTypeExpression(constraint);
    if (!interfaceName) continue;
    const ifaceDef = ctx.implementationContext.getInterfaceDefinition(interfaceName);
    if (!ifaceDef) continue;
    const signature = ifaceDef.signatures?.find((sig) => sig?.name?.name === memberName);
    if (!signature) continue;
    if (signatureHasImplicitSelf(signature)) {
      continue;
    }
    const argSubstitutions = new Map<string, TypeInfo>();
    argSubstitutions.set("Self", receiverType);
    argSubstitutions.set(receiverType.name, receiverType);
    const interfaceArgs =
      constraint?.type === "GenericTypeExpression"
        ? (constraint.arguments ?? []).map((arg) => ctx.resolveTypeExpression(arg, argSubstitutions))
        : [];
    const info = buildInterfaceMethodInfo(ctx, receiverType, interfaceName, ifaceDef, interfaceArgs, signature, -1);
    info.isTypeQualified = true;
    info.typeQualifier = receiverType.name;
    info.structName = receiverType.name;
    candidates.push(info);
  }
  return candidates;
}

function collectUnifiedMemberCandidates(
  ctx: FunctionCallContext,
  receiver: ReceiverLookup,
  memberName: string,
): FunctionInfo[] {
  if (!receiver.lookupType || receiver.lookupType.kind === "unknown") return [];
  if (receiver.lookupType.kind === "type_parameter") {
    return collectTypeParamMethodCandidates(ctx, receiver, memberName);
  }
  const unqualifiedInScope = ctx.statementContext.hasBinding?.(memberName) ?? false;
  const qualifier = receiver.typeQualifier ?? (receiver.lookupType.kind === "struct" ? receiver.lookupType.name : null);
  const typeQualifiedLabel = qualifier ? `${qualifier}.${memberName}` : null;
  const typeQualifiedInScope = typeQualifiedLabel ? ctx.statementContext.hasBinding?.(typeQualifiedLabel) ?? false : false;
  const allowTypeQualified = Boolean(receiver.isTypeReference);
  const isPrimitiveReceiver =
    receiver.lookupType.kind === "primitive" ||
    (receiver.lookupType.kind === "struct" && receiver.lookupType.name === "Array");
  const typeNameInScope = (() => {
    if (!receiver.lookupType.name) return false;
    if (receiver.lookupType.kind === "struct") {
      return ctx.statementContext.hasBinding?.(receiver.lookupType.name) ?? false;
    }
    if (receiver.lookupType.kind === "interface") {
      return ctx.implementationContext.hasInterfaceDefinition?.(receiver.lookupType.name) ?? false;
    }
    return false;
  })();
  const allowInherent = unqualifiedInScope || typeNameInScope || isPrimitiveReceiver;
  const bySignature = new Map<string, FunctionInfo>();
  const methodSetDuplicates = new Set<FunctionInfo>();
  const signatureKey = (entry: FunctionInfo): string => {
    const paramSig = (entry.parameters ?? []).map((param) => formatType(param ?? unknownType)).join("|");
    const receiverLabel = entry.structName ?? "";
    const methodFlag = entry.hasImplicitSelf ? "self" : "free";
    const baseName = entry.name ?? entry.fullName;
    return `${receiverLabel}::${baseName}::${methodFlag}::${paramSig}`;
  };
  const methodSignatures = new Set<string>();
  const candidateAllowed = (entry: FunctionInfo): boolean => {
    if (!entry) return false;
    if (entry.isTypeQualified) {
      if (!allowTypeQualified) return false;
      if (entry.typeQualifier && qualifier && entry.typeQualifier !== qualifier) return false;
      return typeQualifiedInScope || allowTypeQualified;
    }
    if (entry.structName && !entry.isTypeQualified) {
      return allowInherent;
    }
    return unqualifiedInScope;
  };
  const enrichMethodSubstitutions = (entry: FunctionInfo): FunctionInfo => {
    if (!entry || receiver.lookupType.kind !== "struct") {
      return entry;
    }
    const def = receiver.lookupType.name ? ctx.structDefinitions.get(receiver.lookupType.name) : undefined;
    if (!def?.genericParams?.length) {
      return entry;
    }
    const mapped = new Map(entry.methodSetSubstitutions ?? []);
    if (!mapped.has("Self")) {
      mapped.set("Self", receiver.lookupType);
    }
    def.genericParams.forEach((param, idx) => {
      const paramName = ctx.getIdentifierName(param?.name);
      if (!paramName) return;
      if (mapped.has(paramName)) return;
      const arg = Array.isArray(receiver.lookupType.typeArguments) ? receiver.lookupType.typeArguments[idx] : undefined;
      if (arg) {
        mapped.set(paramName, arg);
      }
    });
    if (!mapped.size) {
      return entry;
    }
    return { ...entry, methodSetSubstitutions: Array.from(mapped.entries()) };
  };
  const append = (entries: FunctionInfo[], options?: { skipMethodSets?: boolean }) => {
    for (const entry of entries) {
      if (!entry) continue;
      if (options?.skipMethodSets && entry.fromMethodSet) continue;
      if (!candidateAllowed(entry)) continue;
      const enriched = enrichMethodSubstitutions(entry);
      const key = signatureKey(enriched);
      const incomingPriority =
        typeof enriched.methodResolutionPriority === "number" ? enriched.methodResolutionPriority : 0;
      const existing = bySignature.get(key);
      const existingPriority =
        typeof existing?.methodResolutionPriority === "number" ? existing.methodResolutionPriority : Number.NEGATIVE_INFINITY;
      if (existing && existing.fromMethodSet && enriched.fromMethodSet) {
        methodSetDuplicates.add(existing);
        methodSetDuplicates.add(enriched);
      }
      if (!existing || incomingPriority > existingPriority) {
        bySignature.set(key, enriched);
      }
      if (enriched.structName && !enriched.isTypeQualified) {
        methodSignatures.add(key);
      }
    }
  };
  append(ctx.functionInfos.get(`${receiver.label}::${memberName}`) ?? [], { skipMethodSets: true });
  if (receiver.lookupType.kind === "struct" && receiver.lookupType.name) {
    append(ctx.functionInfos.get(`${receiver.lookupType.name}::${memberName}`) ?? [], { skipMethodSets: true });
  }
  const methodSetLabel =
    receiver.typeQualifier ??
    (receiver.lookupType.kind === "struct" ? receiver.lookupType.name : null) ??
    receiver.label;
  const genericMatches = lookupMethodSetsForCallHelper(
    ctx.implementationContext,
    methodSetLabel,
    memberName,
    receiver.lookupType,
  );
  append(genericMatches);
  append(collectInterfaceMethodCandidates(ctx, receiver, memberName));
  if (!methodSignatures.size) {
    append(resolveUfcsFreeFunctionCandidates(ctx, receiver.lookupType, memberName, unqualifiedInScope));
  }
  const results = Array.from(bySignature.values());
  if (methodSetDuplicates.size > 0) {
    for (const entry of methodSetDuplicates) {
      if (!results.includes(entry)) {
        results.push(entry);
      }
    }
  }
  return results;
}

function resolveFunctionTypeCandidate(
  ctx: FunctionCallContext,
  callee: AST.Expression | undefined | null,
): FunctionInfo | null {
  if (!callee) return null;
  const calleeType = ctx.inferExpression(callee);
  if (!calleeType || calleeType.kind !== "function") {
    return null;
  }
  const params = Array.isArray(calleeType.parameters) ? calleeType.parameters : [];
  const label = formatCalleeLabel(callee) ?? "<function>";
  return {
    name: label,
    fullName: label,
    parameters: params,
    genericConstraints: [],
    genericParamNames: [],
    whereClause: [],
    returnType: calleeType.returnType ?? unknownType,
  };
}

function typesMatchReceiver(receiver: TypeInfo, param: TypeInfo | undefined): boolean {
  if (!param || param.kind === "unknown") {
    return false;
  }
  if (param.kind === "type_parameter" && param.name === "Self") {
    return true;
  }
  return formatType(param) === formatType(receiver);
}

function resolveUfcsFreeFunctionCandidates(
  ctx: FunctionCallContext,
  receiverType: TypeInfo,
  memberName: string,
  inScope: boolean,
): FunctionInfo[] {
  if (!inScope) return [];
  const candidates: FunctionInfo[] = [];
  const entries = ctx.functionInfos.get(memberName) ?? [];
  for (const entry of entries) {
    const params = Array.isArray(entry.parameters) ? entry.parameters : [];
    if (!params.length) continue;
    const first = params[0];
    if (!typesMatchReceiverForFreeFunction(receiverType, first)) {
      continue;
    }
    const adjustedParams = params.slice(1);
    candidates.push({
      ...entry,
      parameters: adjustedParams,
      methodResolutionPriority: entry.methodResolutionPriority,
      hasImplicitSelf: false,
    });
  }
  return candidates;
}

function typesMatchReceiverForFreeFunction(receiver: TypeInfo, param: TypeInfo | undefined): boolean {
  if (!param) return true;
  if (param.kind === "unknown") return false;
  if (param.kind === "type_parameter") return true;
  return typesMatchReceiver(receiver, param);
}

function resolveApplyInterface(
  ctx: FunctionCallContext,
  callee: AST.Expression | undefined | null,
): { returnType: TypeInfo; paramTypes: TypeInfo[] } | null {
  if (!callee) return null;
  const calleeType = ctx.inferExpression(callee);
  if (!calleeType || calleeType.kind === "unknown") {
    return null;
  }
  if (calleeType.kind === "interface" && calleeType.name === "Apply") {
    const args = calleeType.typeArguments ?? [];
    const expectedArg = args[0] ?? unknownType;
    const returnType = args[1] ?? unknownType;
    return { returnType, paramTypes: [expectedArg] };
  }
  const label = formatType(calleeType);
  const applyResult = ctx.typeImplementsInterface?.(calleeType, "Apply");
  const implementationMatch = applyResult?.ok || hasApplyImplementationRecord(ctx, label);
  if (!implementationMatch) {
    return null;
  }
  const candidates: FunctionInfo[] = [];
  const methodInfos = lookupMethodSetsForCallHelper(ctx.implementationContext, label, "apply", calleeType);
  if (methodInfos.length) {
    candidates.push(...methodInfos);
  }
  const direct = ctx.functionInfos.get(`${label}::apply`) ?? [];
  for (const entry of direct) {
    if (candidates.some((info) => info.fullName === entry.fullName)) {
      continue;
    }
    candidates.push(entry);
  }
  if (!candidates.length) {
    return null;
  }
  const info = candidates[0];
  let paramTypes = Array.isArray(info.parameters) ? info.parameters.slice() : [];
  if (info.hasImplicitSelf && paramTypes.length > 0) {
    paramTypes = paramTypes.slice(1);
  }
  return { returnType: info.returnType ?? unknownType, paramTypes };
}

function hasApplyImplementationRecord(ctx: FunctionCallContext, label: string): boolean {
  for (const record of ctx.implementationContext.getImplementationRecords()) {
    if (record.interfaceName === "Apply" && record.targetKey === label) {
      return true;
    }
  }
  return false;
}
