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
  functionInfos: Map<string, FunctionInfo>;
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

export function checkFunctionCall(ctx: FunctionCallContext, call: AST.FunctionCall): void {
  const builtinName = ctx.getBuiltinCallName(call.callee);
  const args = Array.isArray(call.arguments) ? call.arguments : [];
  const argTypes = args.map((arg) => ctx.inferExpression(arg));
  ctx.checkBuiltinCallContext(builtinName, call);
  const infos = resolveFunctionInfos(ctx, call.callee);
  if (!infos.length) {
    const applyMatch = resolveApplyInterface(ctx, call.callee);
    if (applyMatch) {
      let params = applyMatch.paramTypes ?? [];
      const optionalLast = params.length > 0 && params[params.length - 1]?.kind === "nullable";
      if (params.length !== args.length && !(optionalLast && args.length === params.length - 1)) {
        ctx.report(`typechecker: Apply.apply expects ${params.length} arguments, got ${args.length}`, call);
      }
      if (optionalLast && args.length === params.length - 1) {
        params = params.slice(0, params.length - 1);
      }
      const compareCount = Math.min(params.length, argTypes.length);
      for (let index = 0; index < compareCount; index += 1) {
        const expected = params[index];
        const actual = argTypes[index];
        if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
          continue;
        }
        const literalMessage = ctx.describeLiteralMismatch(actual, expected);
        if (literalMessage) {
          ctx.report(literalMessage, args[index] ?? call);
          continue;
        }
        if (!ctx.isTypeAssignable(actual, expected)) {
          ctx.report(
            `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
            args[index] ?? call,
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
  const info = infos[0];
  if (info) {
    const rawParams = Array.isArray(info.parameters) ? info.parameters : [];
    const implicitSelf =
      Boolean(info.structName && info.hasImplicitSelf) &&
      call.callee?.type === "MemberAccessExpression" &&
      rawParams.length > 0;
    let params = implicitSelf ? rawParams.slice(1) : rawParams;
    const optionalLast = params.length > 0 && params[params.length - 1]?.kind === "nullable";
    if (params.length !== args.length && !(optionalLast && args.length === params.length - 1)) {
      ctx.report(`typechecker: function expects ${params.length} arguments, got ${args.length}`, call);
    }
    if (optionalLast && args.length === params.length - 1) {
      params = params.slice(0, params.length - 1);
    }
    const compareCount = Math.min(params.length, argTypes.length);
    for (let index = 0; index < compareCount; index += 1) {
      const expected = params[index];
      const actual = argTypes[index];
      if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
        continue;
      }
      const literalMessage = ctx.describeLiteralMismatch(actual, expected);
      if (literalMessage) {
        ctx.report(literalMessage, args[index] ?? call);
        continue;
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
        ctx.report(
          `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
          args[index] ?? call,
        );
      }
    }
  }
  for (const info of infos) {
    enforceFunctionConstraintsHelper(ctx.implementationContext, info, call);
  }
}

export function inferFunctionCallReturnType(ctx: FunctionCallContext, call: AST.FunctionCall): TypeInfo {
  const infos = resolveFunctionInfos(ctx, call.callee);
  if (!infos.length) {
    const applyMatch = resolveApplyInterface(ctx, call.callee);
    return applyMatch?.returnType ?? unknownType;
  }
  const returnTypes = infos
    .map((info) => info.returnType ?? unknownType)
    .filter((type) => type && type.kind !== "unknown");
  if (!returnTypes.length) {
    return unknownType;
  }
  return mergeBranchTypesHelper(ctx.statementContext, returnTypes);
}

function resolveFunctionInfos(ctx: FunctionCallContext, callee: AST.Expression | undefined | null): FunctionInfo[] {
  if (!callee) return [];
  if (callee.type === "Identifier") {
    const info = ctx.functionInfos.get(callee.name);
    return info ? [info] : [];
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
      const info = ctx.functionInfos.get(memberKey);
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
      if (!infos.length && info) {
        infos.push(info);
        seen.add(info.fullName);
      }
      if (!infos.length) {
        const fallback = ctx.functionInfos.get(memberName);
        if (fallback && !seen.has(fallback.fullName)) {
          infos.push(fallback);
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
      const info = ctx.functionInfos.get(memberKey);
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
      if (!infos.length && info) {
        infos.push(info);
        seen.add(info.fullName);
      }
      if (!infos.length) {
        const fallback = ctx.functionInfos.get(memberName);
        if (fallback && !seen.has(fallback.fullName)) {
          infos.push(fallback);
        }
      }
      return infos;
    }
    if (objectType.kind === "primitive" && objectType.name === "string") {
      const structLabel = formatType(objectType);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey);
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
      if (!infos.length && info) {
        infos.push(info);
        seen.add(info.fullName);
      }
      if (!infos.length) {
        const fallback = ctx.functionInfos.get(memberName);
        if (fallback && !seen.has(fallback.fullName)) {
          infos.push(fallback);
        }
      }
      return infos;
    }
  }
  return [];
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
  const direct = ctx.functionInfos.get(`${label}::apply`);
  if (direct && !candidates.some((info) => info.fullName === direct.fullName)) {
    candidates.push(direct);
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
