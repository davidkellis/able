import * as AST from "../../ast";
import { formatType, unknownType, type TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import {
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
} from "./implementations";
import { mergeBranchTypes as mergeBranchTypesHelper } from "./expressions";
import type { StatementContext } from "./expression-context";
import type { FunctionInfo, InterfaceCheckResult } from "./types";
import {
  buildEffectiveParams,
  buildSelfArgSubstitutions,
} from "./function-call-parse";
import {
  findPartialApplication,
  resolveApplyInterface,
  resolveFunctionInfos,
  resolveFunctionTypeCandidate,
  resolveMemberAccessCandidates,
  selectBestOverload,
} from "./function-call-resolve";
import {
  formatCalleeLabel,
  reportAmbiguousInterfaceMethod,
  reportArgumentDiagnostics,
} from "./function-call-errors";

export type FunctionCallContext = {
  implementationContext: ImplementationContext;
  functionInfos: Map<string, FunctionInfo[]>;
  structDefinitions: Map<string, AST.StructDefinition>;
  currentPackageName?: string;
  symbolOrigins?: Map<string, string>;
  getStructDefinition(name: string): AST.StructDefinition | undefined;
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
