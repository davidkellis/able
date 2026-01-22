import * as AST from "../../ast";
import { formatType, unknownType, type TypeInfo } from "../types";
import {
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
  matchImplementationTarget,
  typeImplementsInterface,
} from "./implementations";
import { expectsSelfType } from "./implementation-validation";
import { formatCalleeLabel } from "./function-call-errors";
import {
  arityMatches,
  buildEffectiveParams,
  dropOptionalParam,
  instantiateCallSignature,
} from "./function-call-parse";
import type { FunctionCallContext } from "./function-calls";
import type { FunctionInfo, ImplementationObligation } from "./types";

export type OverloadResolution =
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

export type MemberAccessResolution = { candidates: FunctionInfo[]; fieldError?: string };

type ReceiverLookup = {
  lookupType: TypeInfo;
  label: string;
  memberName: string;
  isTypeReference?: boolean;
  typeQualifier?: string | null;
  referenceName?: string | null;
};

export function selectBestOverload(
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
    isGeneric: boolean;
    specificity: number;
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
    const effective = buildEffectiveParams(info, call);
    const signatureParams = dropOptionalParam(effective.params, args.length, effective.optionalLast);
    const isGeneric = Array.isArray(info.genericParamNames) && info.genericParamNames.length > 0;
    const specificity = isGeneric ? signatureSpecificityScore(signatureParams) : 0;
    const priority = typeof info.methodResolutionPriority === "number" ? info.methodResolutionPriority : 0;
    compatible.push({
      info,
      params,
      optionalLast: instantiated.optionalLast,
      returnType: instantiated.returnType ?? unknownType,
      inferredTypeArgs: instantiated.inferredTypeArgs,
      score: score - (instantiated.optionalLast && params.length !== instantiated.params.length ? 0.5 : 0),
      priority,
      isGeneric,
      specificity,
    });
  }
  if (!compatible.length) {
    return { kind: "no-match" };
  }
  compatible.sort((a, b) => {
    const scoreDiff = b.score - a.score;
    if (Math.abs(scoreDiff) > 1e-9) return scoreDiff;
    if (a.isGeneric !== b.isGeneric) return a.isGeneric ? 1 : -1;
    if (a.isGeneric && b.isGeneric && a.specificity !== b.specificity) return b.specificity - a.specificity;
    if (a.priority !== b.priority) return b.priority - a.priority;
    return 0;
  });
  const best = compatible[0];
  const tied = compatible.filter(
    (entry) =>
      Math.abs(entry.score - best.score) <= 1e-9 &&
      entry.isGeneric === best.isGeneric &&
      (!best.isGeneric || entry.specificity === best.specificity) &&
      entry.priority === best.priority,
  );
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

export function findPartialApplication(
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

function signatureSpecificityScore(params: TypeInfo[]): number {
  let total = 0;
  for (const param of params) {
    total += typeSpecificityScore(param);
  }
  return total;
}

function typeSpecificityScore(type: TypeInfo): number {
  switch (type.kind) {
    case "unknown":
    case "type_parameter":
      return 0;
    case "primitive":
      return 1;
    case "array":
      return 1 + typeSpecificityScore(type.element);
    case "map":
      return 1 + typeSpecificityScore(type.key) + typeSpecificityScore(type.value);
    case "range":
      return 1 + typeSpecificityScore(type.element) + (type.bounds ?? []).reduce((acc, bound) => acc + typeSpecificityScore(bound), 0);
    case "iterator":
      return 1 + typeSpecificityScore(type.element);
    case "proc":
      return 1 + typeSpecificityScore(type.result);
    case "future":
      return 1 + typeSpecificityScore(type.result);
    case "struct": {
      const args = Array.isArray(type.typeArguments) ? type.typeArguments : [];
      return 1 + args.reduce((acc, arg) => acc + typeSpecificityScore(arg), 0);
    }
    case "interface": {
      const args = Array.isArray(type.typeArguments) ? type.typeArguments : [];
      return 1 + args.reduce((acc, arg) => acc + typeSpecificityScore(arg), 0);
    }
    case "function": {
      const params = Array.isArray(type.parameters) ? type.parameters : [];
      let score = 1 + typeSpecificityScore(type.returnType);
      for (const param of params) {
        score += typeSpecificityScore(param);
      }
      return score;
    }
    case "nullable":
      return 1 + typeSpecificityScore(type.inner);
    case "result":
      return 1 + typeSpecificityScore(type.inner);
    case "union": {
      const members = Array.isArray(type.members) ? type.members : [];
      return 1 + members.reduce((acc, member) => acc + typeSpecificityScore(member), 0);
    }
    default:
      return 0;
  }
}

export function resolveFunctionInfos(ctx: FunctionCallContext, callee: AST.Expression | undefined | null): FunctionInfo[] {
  if (!callee) return [];
  if (callee.type === "MemberAccessExpression") {
    const resolution = resolveMemberAccessCandidates(ctx, callee);
    if (resolution.fieldError) return [];
    return resolution.candidates;
  }
  if (callee.type === "Identifier") {
    if (!ctx.statementContext.lookupIdentifier?.(callee.name)) {
      return [];
    }
    const infos = ctx.functionInfos.get(callee.name) ?? [];
    const inScope = filterBySymbolScope(ctx, infos, callee.name);
    const nonMethodInfos = inScope.filter((info) => !info.structName);
    return nonMethodInfos.length > 0 ? nonMethodInfos : inScope;
  }
  return [];
}

function filterBySymbolScope(
  ctx: FunctionCallContext,
  infos: FunctionInfo[],
  symbolName: string,
): FunctionInfo[] {
  if (!infos.length) return infos;
  const isBuiltin = (info: FunctionInfo) => !info.packageName || info.packageName === "<builtin>";
  const origin = ctx.symbolOrigins?.get(symbolName);
  if (origin && origin.size > 0) {
    if (origin.has("<builtin>")) {
      return infos;
    }
    return infos.filter((info) => !info.packageName || origin.has(info.packageName));
  }
  return infos.filter((info) => isBuiltin(info));
}

export function resolveMemberAccessCandidates(
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
  if (receiver.lookupType.kind === "future" && memberName === "cancel") {
    return { candidates: [], fieldError: "future handles do not support cancel()" };
  }
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

export function resolveFunctionTypeCandidate(
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

export function resolveApplyInterface(
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

function buildReceiverLookup(
  ctx: FunctionCallContext,
  callee: AST.MemberAccessExpression,
  memberName: string,
): ReceiverLookup | null {
  const objectName = callee.object?.type === "Identifier" ? callee.object.name : null;
  const referenceName = objectName;
  const typeNameInScope = referenceName ? ctx.statementContext.isTypeNameInScope?.(referenceName) ?? false : false;
  const hasObjectBinding = objectName ? ctx.statementContext.hasBinding?.(objectName) ?? false : true;
  let objectType = ctx.inferExpression(callee.object);
  if (
    objectType.kind !== "struct" &&
    callee.object?.type === "Identifier" &&
    callee.object.name &&
    typeNameInScope &&
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
    (objectType.kind === "type_parameter" || typeNameInScope || !hasObjectBinding);
  if (objectType.kind === "array") {
    const lookupType: TypeInfo = {
      kind: "struct",
      name: "Array",
      typeArguments: [objectType.element ?? unknownType],
    };
    return { lookupType, label: formatType(lookupType), memberName, typeQualifier: "Array", referenceName };
  }
  const typeQualifier =
    objectType.kind === "struct" || objectType.kind === "interface" || objectType.kind === "type_parameter"
      ? objectType.name
      : null;
  return {
    lookupType: objectType,
    label: formatType(objectType),
    memberName,
    isTypeReference,
    typeQualifier,
    referenceName,
  };
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
      genericConstraints: [],
      genericParamNames: [],
      whereClause: [],
      returnType: unknownType,
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
    subjectExpr?: AST.TypeExpression,
  ) => {
    const interfaceName = ctx.implementationContext.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context, subjectExpr });
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
      const subjectExpr = clause?.typeParam;
      const typeParamName = subjectExpr ? ctx.implementationContext.formatTypeExpression(subjectExpr) : null;
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause", subjectExpr);
      }
    }
  }
  return obligations;
}

function isSelfPatternPlaceholderName(
  ctx: FunctionCallContext,
  name: string,
  interfaceGenericNames: Set<string>,
): boolean {
  if (!name || name === "Self") {
    return false;
  }
  if (interfaceGenericNames.has(name)) {
    return false;
  }
  if (ctx.implementationContext.isKnownTypeName(name)) {
    return false;
  }
  return true;
}

function collectSelfPatternPlaceholderNames(
  ctx: FunctionCallContext,
  expr: AST.TypeExpression | null | undefined,
  interfaceGenericNames: Set<string>,
  out: Set<string> = new Set(),
): Set<string> {
  if (!expr) return out;
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expr.name);
      if (name && isSelfPatternPlaceholderName(ctx, name, interfaceGenericNames)) {
        out.add(name);
      }
      return out;
    }
    case "GenericTypeExpression": {
      collectSelfPatternPlaceholderNames(ctx, expr.base, interfaceGenericNames, out);
      (expr.arguments ?? []).forEach((arg) =>
        collectSelfPatternPlaceholderNames(ctx, arg, interfaceGenericNames, out),
      );
      return out;
    }
    case "FunctionTypeExpression": {
      (expr.paramTypes ?? []).forEach((param) =>
        collectSelfPatternPlaceholderNames(ctx, param, interfaceGenericNames, out),
      );
      collectSelfPatternPlaceholderNames(ctx, expr.returnType, interfaceGenericNames, out);
      return out;
    }
    case "NullableTypeExpression":
    case "ResultTypeExpression":
      return collectSelfPatternPlaceholderNames(ctx, expr.innerType, interfaceGenericNames, out);
    case "UnionTypeExpression":
      (expr.members ?? []).forEach((member) =>
        collectSelfPatternPlaceholderNames(ctx, member, interfaceGenericNames, out),
      );
      return out;
    default:
      return out;
  }
}

function normalizeSelfTypeConstructor(ctx: FunctionCallContext, receiverType: TypeInfo): TypeInfo {
  if (receiverType.kind === "struct") {
    const def = receiverType.definition ?? ctx.getStructDefinition(receiverType.name);
    return { kind: "struct", name: receiverType.name, typeArguments: [], definition: def };
  }
  if (receiverType.kind === "interface") {
    const def =
      receiverType.definition ?? ctx.implementationContext.getInterfaceDefinition?.(receiverType.name) ?? undefined;
    return { kind: "interface", name: receiverType.name, typeArguments: [], definition: def };
  }
  if (receiverType.kind === "array") {
    const def = ctx.getStructDefinition("Array");
    return { kind: "struct", name: "Array", typeArguments: [], definition: def ?? undefined };
  }
  if (receiverType.kind === "iterator") {
    const def = ctx.implementationContext.getInterfaceDefinition?.("Iterator");
    return { kind: "interface", name: "Iterator", typeArguments: [], definition: def ?? undefined };
  }
  return receiverType;
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
  const interfaceGenericNames = new Set(
    (interfaceDef.genericParams ?? [])
      .map((param) => ctx.getIdentifierName(param?.name))
      .filter((name): name is string => Boolean(name)),
  );
  if (Array.isArray(interfaceDef.genericParams)) {
    interfaceDef.genericParams.forEach((param, idx) => {
      const name = ctx.getIdentifierName(param?.name);
      if (!name) return;
      const arg = interfaceArgs[idx] ?? unknownType;
      substitutions.set(name, arg);
    });
  }
  const placeholderNames = collectSelfPatternPlaceholderNames(ctx, interfaceDef.selfTypePattern, interfaceGenericNames);
  if (placeholderNames.size > 0) {
    const constructorType = normalizeSelfTypeConstructor(ctx, receiverType);
    for (const name of placeholderNames) {
      if (!substitutions.has(name)) {
        substitutions.set(name, constructorType);
      }
    }
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
  const packageName = (interfaceDef as unknown as { _package?: string })._package;
  return {
    name: signature.name?.name ?? "<anonymous>",
    fullName: `${interfaceName}::${signature.name?.name ?? "<anonymous>"}`,
    packageName,
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
  const referenceName = receiver.referenceName ?? null;
  const symbolPackages = referenceName ? ctx.getTypeOriginsForSymbol?.(referenceName) ?? null : null;
  const hasSymbolScope = Boolean(symbolPackages && symbolPackages.size > 0);
  const canonicalName =
    receiver.typeQualifier ??
    (receiver.lookupType.kind === "struct" || receiver.lookupType.kind === "interface" ? receiver.lookupType.name : null);
  const canonicalPackages = canonicalName ? ctx.getTypeOriginsForCanonical?.(canonicalName) ?? null : null;
  const hasCanonicalScope = Boolean(canonicalPackages && canonicalPackages.size > 0);
  const allowTypeQualified = Boolean(receiver.isTypeReference) && hasSymbolScope;
  const isPrimitiveReceiver =
    receiver.lookupType.kind === "primitive" ||
    (receiver.lookupType.kind === "struct" && receiver.lookupType.name === "Array");
  const allowInherent = hasCanonicalScope || isPrimitiveReceiver;
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
  const matchesPackage = (entry: FunctionInfo, allowed: Set<string> | null): boolean => {
    if (!entry?.packageName || !allowed || allowed.size === 0) {
      return true;
    }
    if (allowed.has("<builtin>")) {
      return true;
    }
    return allowed.has(entry.packageName);
  };
  const candidateAllowed = (entry: FunctionInfo): boolean => {
    if (!entry) return false;
    if (entry.isTypeQualified) {
      if (!allowTypeQualified) return false;
      if (entry.typeQualifier && qualifier && entry.typeQualifier !== qualifier) return false;
      if (!matchesPackage(entry, symbolPackages)) return false;
      return true;
    }
    if (entry.structName && !entry.isTypeQualified) {
      if (!allowInherent) return false;
      return true;
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
    append(
      resolveUfcsFreeFunctionCandidates(
        ctx,
        receiver.lookupType,
        memberName,
        unqualifiedInScope,
        hasCanonicalScope || isPrimitiveReceiver || receiver.lookupType.kind === "type_parameter",
      ),
    );
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
  receiverInScope: boolean,
): FunctionInfo[] {
  if (!inScope || !receiverInScope) return [];
  const candidates: FunctionInfo[] = [];
  const entries = filterBySymbolScope(ctx, ctx.functionInfos.get(memberName) ?? [], memberName);
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

function hasApplyImplementationRecord(ctx: FunctionCallContext, label: string): boolean {
  for (const record of ctx.implementationContext.getImplementationRecords()) {
    if (record.interfaceName === "Apply" && record.targetKey === label) {
      return true;
    }
  }
  return false;
}
