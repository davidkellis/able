import type * as AST from "../../ast";
import { formatType, primitiveType, unknownType, type TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import {
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
  let candidates = resolveFunctionInfos(ctx, call.callee);
  let callArgs = args;
  let callArgTypes = argTypes;
  if (!candidates.length) {
    const ufcs = resolveUfcsInherentCandidates(ctx, call, args, argTypes);
    if (ufcs) {
      candidates = ufcs.candidates;
      callArgs = ufcs.args;
      callArgTypes = ufcs.argTypes;
    }
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
    if (
      calleeType.kind !== "unknown" &&
      (call.callee?.type === "Identifier" || call.callee?.type === "MemberAccessExpression")
    ) {
      ctx.report(
        `typechecker: cannot call non-callable value ${formatType(calleeType)} (missing Apply implementation)`,
        call.callee ?? call,
      );
    }
    return;
  }
  const resolution = selectBestOverload(ctx, candidates, call, callArgs, callArgTypes);
  if (resolution.kind === "no-match") {
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
  let infos = resolveFunctionInfos(ctx, call.callee);
  let callArgs = args;
  let callArgTypes = argTypes;
  if (!infos.length) {
    const ufcs = resolveUfcsInherentCandidates(ctx, call, args, argTypes);
    if (ufcs) {
      infos = ufcs.candidates;
      callArgs = ufcs.args;
      callArgTypes = ufcs.argTypes;
    }
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
  const compatible: Array<{ info: FunctionInfo; params: TypeInfo[]; score: number }> = [];
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
    compatible.push({ info, params, score: score - (effective.optionalLast && params.length !== effective.params.length ? 0.5 : 0) });
  }
  if (!compatible.length) {
    return { kind: "no-match" };
  }
  compatible.sort((a, b) => b.score - a.score);
  const best = compatible[0];
  const tied = compatible.filter((entry) => entry.score === best.score);
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
  if (callee.type === "Identifier") {
    return ctx.functionInfos.get(callee.name) ?? [];
  }
  if (callee.type === "MemberAccessExpression") {
    if (ctx.handlePackageMemberAccess(callee)) {
      return [];
    }
    const memberName = ctx.getIdentifierName(callee.member);
    if (!memberName) return [];
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
    if (objectType.kind === "struct") {
      const structLabel = formatType(objectType);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey) ?? [];
      const genericMatches = lookupMethodSetsForCallHelper(
        ctx.implementationContext,
        structLabel,
        memberName,
        objectType,
      );
      if (genericMatches.length) {
        for (const match of genericMatches) {
          if (seen.has(match.fullName)) continue;
          infos.push(match);
          seen.add(match.fullName);
        }
      }
      if (!infos.length && info.length) {
        for (const entry of info) {
          if (seen.has(entry.fullName)) continue;
          infos.push(entry);
          seen.add(entry.fullName);
        }
      }
      if (!infos.length) {
        const ufcs = resolveUfcsFreeFunctionCandidates(ctx, objectType, memberName);
        for (const entry of ufcs) {
          if (seen.has(entry.fullName)) continue;
          infos.push(entry);
          seen.add(entry.fullName);
        }
      }
      return infos;
    }
    if (objectType.kind === "array") {
      const arrayStruct: TypeInfo = {
        kind: "struct",
        name: "Array",
        typeArguments: [objectType.element ?? unknownType],
      };
      const structLabel = formatType(arrayStruct);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey) ?? [];
      const genericMatches = lookupMethodSetsForCallHelper(
        ctx.implementationContext,
        structLabel,
        memberName,
        arrayStruct,
      );
      if (genericMatches.length) {
        for (const match of genericMatches) {
          if (seen.has(match.fullName)) continue;
          infos.push(match);
          seen.add(match.fullName);
        }
      }
      if (!infos.length && info.length) {
        for (const entry of info) {
          if (seen.has(entry.fullName)) continue;
          infos.push(entry);
          seen.add(entry.fullName);
        }
      }
      if (!infos.length) {
        const ufcs = resolveUfcsFreeFunctionCandidates(ctx, arrayStruct, memberName);
        for (const entry of ufcs) {
          if (seen.has(entry.fullName)) continue;
          infos.push(entry);
          seen.add(entry.fullName);
        }
      }
      return infos;
    }
    if (objectType.kind === "primitive" && objectType.name === "string") {
      const structLabel = formatType(objectType);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey) ?? [];
      const genericMatches = lookupMethodSetsForCallHelper(
        ctx.implementationContext,
        structLabel,
        memberName,
        objectType,
      );
      if (genericMatches.length) {
        for (const match of genericMatches) {
          if (seen.has(match.fullName)) continue;
          infos.push(match);
          seen.add(match.fullName);
        }
      }
      if (!infos.length && info.length) {
        for (const entry of info) {
          if (seen.has(entry.fullName)) continue;
          infos.push(entry);
          seen.add(entry.fullName);
        }
      }
      if (!infos.length) {
        const ufcs = resolveUfcsFreeFunctionCandidates(ctx, objectType, memberName);
        for (const entry of ufcs) {
          if (seen.has(entry.fullName)) continue;
          infos.push(entry);
          seen.add(entry.fullName);
        }
      }
      return infos;
    }
    if (objectType.kind !== "unknown") {
      const ufcs = resolveUfcsFreeFunctionCandidates(ctx, objectType, memberName);
      if (ufcs.length) {
        return ufcs;
      }
    }
  }
  return [];
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

type UfcsCandidateResult = { candidates: FunctionInfo[]; args: AST.Expression[]; argTypes: TypeInfo[] };

function resolveUfcsInherentCandidates(
  ctx: FunctionCallContext,
  call: AST.FunctionCall,
  args: AST.Expression[],
  argTypes: TypeInfo[],
): UfcsCandidateResult | null {
  if (call.callee?.type !== "Identifier" || args.length === 0) {
    return null;
  }
  const receiverType = argTypes[0];
  if (!receiverType || receiverType.kind === "unknown") {
    return null;
  }
  const lookup = normalizeUfcsReceiver(receiverType);
  if (!lookup) {
    return null;
  }
  const matches = lookupMethodSetsForCallHelper(ctx.implementationContext, lookup.label, call.callee.name, lookup.lookupType);
  if (!matches.length) {
    return null;
  }
  const ufcsInfos: FunctionInfo[] = [];
  for (const match of matches) {
    if (!match.hasImplicitSelf) continue;
    const params = Array.isArray(match.parameters) ? match.parameters.slice() : [];
    const dropSelf = params.length > 0 && typesMatchReceiver(lookup.lookupType, params[0]);
    const adjusted: FunctionInfo = {
      ...match,
      hasImplicitSelf: false,
      parameters: dropSelf ? params.slice(1) : params,
    };
    ufcsInfos.push(adjusted);
  }
  if (!ufcsInfos.length) {
    return null;
  }
  return { candidates: ufcsInfos, args: args.slice(1), argTypes: argTypes.slice(1) };
}

function normalizeUfcsReceiver(receiver: TypeInfo): { label: string; lookupType: TypeInfo } | null {
  if (receiver.kind === "struct") {
    return { label: formatType(receiver), lookupType: receiver };
  }
  if (receiver.kind === "array") {
    const arrayStruct: TypeInfo = {
      kind: "struct",
      name: "Array",
      typeArguments: [receiver.element ?? unknownType],
    };
    return { label: formatType(arrayStruct), lookupType: arrayStruct };
  }
  if (receiver.kind === "primitive" && receiver.name === "string") {
    return { label: formatType(receiver), lookupType: receiver };
  }
  return null;
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
): FunctionInfo[] {
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
      hasImplicitSelf: false,
    });
  }
  return candidates;
}

function typesMatchReceiverForFreeFunction(receiver: TypeInfo, param: TypeInfo | undefined): boolean {
  if (!param || param.kind === "unknown") return true;
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
