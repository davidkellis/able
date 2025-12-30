import type * as AST from "../../ast";
import { formatType, primitiveType, unknownType, type TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import {
  ambiguousImplementationDetail,
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
  typeImplementsInterface,
} from "./implementations";
import { mergeBranchTypes as mergeBranchTypesHelper } from "./expressions";
import type { StatementContext } from "./expression-context";
import type { FunctionInfo, InterfaceCheckResult } from "./types";

type FunctionCallContext = {
  implementationContext: ImplementationContext;
  functionInfos: Map<string, FunctionInfo[]>;
  structDefinitions: Map<string, AST.StructDefinition>;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  resolveTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, TypeInfo>): TypeInfo;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean;
  report(message: string, node?: AST.Node | null): void;
  handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  checkBuiltinCallContext(name: string | undefined, call: AST.FunctionCall): void;
  getBuiltinCallName(callee: AST.Expression | undefined | null): string | undefined;
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

export function checkFunctionCall(ctx: FunctionCallContext, call: AST.FunctionCall): void {
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
  const resolution = selectBestOverload(ctx, candidates, call, callArgs, callArgTypes);
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
  const { info, params, optionalLast } = resolution;
  const ok = reportArgumentDiagnostics(ctx, info, params, optionalLast, call, callArgs, callArgTypes);
  if (ok) {
    enforceFunctionConstraintsHelper(ctx.implementationContext, info, call);
  }
}

export function inferFunctionCallReturnType(ctx: FunctionCallContext, call: AST.FunctionCall): TypeInfo {
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
  const resolution = selectBestOverload(ctx, infos, call, callArgs, callArgTypes);
  if (resolution.kind === "match") {
    return resolution.info.returnType ?? unknownType;
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
  | { kind: "match"; info: FunctionInfo; params: TypeInfo[]; optionalLast: boolean }
  | { kind: "ambiguous"; infos: FunctionInfo[] }
  | { kind: "no-match" };

function selectBestOverload(
  ctx: FunctionCallContext,
  infos: FunctionInfo[],
  call: AST.FunctionCall,
  args: AST.Expression[],
  argTypes: TypeInfo[],
): OverloadResolution {
  const compatible: Array<{ info: FunctionInfo; params: TypeInfo[]; score: number; priority: number }> = [];
  for (const info of infos) {
    const effective = buildEffectiveParams(info, call);
    if (!arityMatches(effective.params, args.length, effective.optionalLast)) {
      continue;
    }
    const params = dropOptionalParam(effective.params, args.length, effective.optionalLast);
    const score = scoreCompatibility(ctx, params, argTypes);
    if (score < 0) {
      continue;
    }
    const priority = typeof info.methodResolutionPriority === "number" ? info.methodResolutionPriority : 0;
    compatible.push({
      info,
      params,
      score: score - (effective.optionalLast && params.length !== effective.params.length ? 0.5 : 0),
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
  const effective = buildEffectiveParams(best.info, call);
  return { kind: "match", info: best.info, params: best.params, optionalLast: effective.optionalLast };
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
  if (fieldResolution?.nonCallable) {
    return { candidates: [], fieldError: `field '${memberName}' is not callable` };
  }
  return { candidates: collectUnifiedMemberCandidates(ctx, receiver, memberName) };
}

function buildReceiverLookup(
  ctx: FunctionCallContext,
  callee: AST.MemberAccessExpression,
  memberName: string,
): ReceiverLookup | null {
  const objectName = callee.object?.type === "Identifier" ? callee.object.name : null;
  const isTypeName = objectName ? ctx.structDefinitions.has(objectName) : false;
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
    objectType.kind === "struct" &&
    (!hasObjectBinding || isTypeName);
  if (objectType.kind === "array") {
    const lookupType: TypeInfo = {
      kind: "struct",
      name: "Array",
      typeArguments: [objectType.element ?? unknownType],
    };
    return { lookupType, label: formatType(lookupType), memberName, typeQualifier: "Array" };
  }
  const typeQualifier = objectType.kind === "struct" ? objectType.name : null;
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

function collectUnifiedMemberCandidates(
  ctx: FunctionCallContext,
  receiver: ReceiverLookup,
  memberName: string,
): FunctionInfo[] {
  if (!receiver.lookupType || receiver.lookupType.kind === "unknown") return [];
  const unqualifiedInScope = ctx.statementContext.hasBinding?.(memberName) ?? false;
  const qualifier = receiver.typeQualifier ?? (receiver.lookupType.kind === "struct" ? receiver.lookupType.name : null);
  const typeQualifiedLabel = qualifier ? `${qualifier}.${memberName}` : null;
  const typeQualifiedInScope = typeQualifiedLabel ? ctx.statementContext.hasBinding?.(typeQualifiedLabel) ?? false : false;
  const allowTypeQualified = Boolean(receiver.isTypeReference);
  const isPrimitiveReceiver =
    receiver.lookupType.kind === "primitive" ||
    (receiver.lookupType.kind === "struct" && receiver.lookupType.name === "Array");
  const typeNameInScope =
    receiver.lookupType.kind === "struct" && receiver.lookupType.name
      ? ctx.statementContext.hasBinding?.(receiver.lookupType.name) ?? false
      : false;
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
