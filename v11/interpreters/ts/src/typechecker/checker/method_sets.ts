import * as AST from "../../ast";
import { formatType } from "../types";
import type { TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import type { ImplementationObligation, InterfaceCheckResult, MethodSetRecord } from "./types";

export interface MethodSetHelpers {
  matchImplementationTarget: (
    ctx: ImplementationContext,
    actual: TypeInfo,
    target: AST.TypeExpression,
    paramNames: Set<string>,
    substitutions: Map<string, TypeInfo>,
  ) => boolean;
  buildStringSubstitutionMap: (substitutions: Map<string, TypeInfo>) => Map<string, string>;
  typeExpressionsEquivalent: (
    ctx: ImplementationContext,
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ) => boolean;
  isImplicitSelfParameter: (param: AST.FunctionParameter | null | undefined) => boolean;
  expectsSelfType: (expr: AST.TypeExpression | null | undefined) => boolean;
  lookupObligationSubject: (
    ctx: ImplementationContext,
    typeParam: string,
    substitutions: Map<string, TypeInfo>,
    selfType: TypeInfo,
  ) => TypeInfo | null;
  resolveInterfaceArgumentLabels: (
    ctx: ImplementationContext,
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
  ) => string[];
  typeImplementsInterface: (
    ctx: ImplementationContext,
    type: TypeInfo,
    interfaceName: string,
    expectedArgs: string[],
  ) => InterfaceCheckResult;
}

interface MethodSetMatchResult {
  ok: boolean;
  detail?: string;
  substitutions?: Map<string, TypeInfo>;
  obligations?: ImplementationObligation[];
}

export function methodSetProvidesInterface(
  ctx: ImplementationContext,
  type: TypeInfo,
  interfaceName: string,
  expectedArgs: string[],
  helpers: MethodSetHelpers,
): InterfaceCheckResult {
  const interfaceDefinition = ctx.getInterfaceDefinition(interfaceName);
  if (!interfaceDefinition) {
    return { ok: false };
  }
  let bestDetail: string | undefined;
  for (const record of ctx.getMethodSets()) {
    const evaluation = evaluateMethodSetAgainstInterface(ctx, record, type, interfaceDefinition, expectedArgs, helpers);
    if (!evaluation) {
      continue;
    }
    if (!evaluation.ok) {
      if (evaluation.detail && (!bestDetail || evaluation.detail.length > bestDetail.length)) {
        bestDetail = evaluation.detail;
      }
      continue;
    }
    const failure = enforceMethodSetObligations(
      ctx,
      record,
      type,
      evaluation.substitutions ?? new Map(),
      evaluation.obligations ?? [],
      helpers,
    );
    if (failure) {
      if (!bestDetail || failure.length > bestDetail.length) {
        bestDetail = failure;
      }
      continue;
    }
    return { ok: true };
  }
  if (bestDetail) {
    return { ok: false, detail: bestDetail };
  }
  return { ok: false };
}

function evaluateMethodSetAgainstInterface(
  ctx: ImplementationContext,
  record: MethodSetRecord,
  type: TypeInfo,
  interfaceDefinition: AST.InterfaceDefinition,
  expectedArgs: string[],
  helpers: MethodSetHelpers,
): MethodSetMatchResult | null {
  if (!Array.isArray(interfaceDefinition.signatures) || interfaceDefinition.signatures.length === 0) {
    return null;
  }
  const paramNames = new Set(record.genericParams);
  const substitutions = new Map<string, TypeInfo>();
  substitutions.set("Self", type);
  if (!helpers.matchImplementationTarget(ctx, type, record.target, paramNames, substitutions)) {
    return null;
  }
  const stringSubs = helpers.buildStringSubstitutionMap(substitutions);
  if (!stringSubs.has("Self")) {
    stringSubs.set("Self", ctx.describeTypeArgument(type));
  }
  const interfaceArgMap = buildInterfaceArgumentLabelMap(ctx, interfaceDefinition, expectedArgs);
  for (const [key, value] of interfaceArgMap.entries()) {
    stringSubs.set(key, value);
  }
  const provided = new Map<string, AST.FunctionDefinition>();
  if (Array.isArray(record.definition.definitions)) {
    for (const entry of record.definition.definitions) {
      if (entry?.type !== "FunctionDefinition") {
        continue;
      }
      const methodName = ctx.getIdentifierName(entry.id);
      if (!methodName || provided.has(methodName)) {
        continue;
      }
      provided.set(methodName, entry);
    }
  }
  const obligations: ImplementationObligation[] = [];
  for (const signature of interfaceDefinition.signatures) {
    if (!signature) {
      continue;
    }
    const methodName = ctx.getIdentifierName(signature.name);
    if (!methodName) {
      continue;
    }
    const method = provided.get(methodName);
    if (!method) {
      if (signature.defaultImpl) {
        continue;
      }
      return { ok: false, detail: `${record.label}: method '${methodName}' not provided` };
    }
    const comparison = compareMethodSetSignature(ctx, record, signature, method, stringSubs, helpers);
    if (!comparison.ok) {
      return comparison;
    }
    obligations.push(...extractFunctionObligations(ctx, method, `via method '${methodName}'`));
  }
  return { ok: true, substitutions, obligations: [...record.obligations, ...obligations] };
}

function buildInterfaceArgumentLabelMap(
  ctx: ImplementationContext,
  interfaceDefinition: AST.InterfaceDefinition,
  expectedArgs: string[],
): Map<string, string> {
  const map = new Map<string, string>();
  if (!Array.isArray(interfaceDefinition.genericParams)) {
    return map;
  }
  interfaceDefinition.genericParams.forEach((param, index) => {
    const name = ctx.getIdentifierName(param?.name);
    if (!name) {
      return;
    }
    map.set(name, expectedArgs[index] ?? "Unknown");
  });
  return map;
}

function compareMethodSetSignature(
  ctx: ImplementationContext,
  record: MethodSetRecord,
  signature: AST.FunctionSignature,
  method: AST.FunctionDefinition,
  substitutions: Map<string, string>,
  helpers: MethodSetHelpers,
): { ok: boolean; detail?: string } {
  const methodName = ctx.getIdentifierName(signature.name) ?? "<anonymous>";
  const interfaceGenerics = Array.isArray(signature.genericParams) ? signature.genericParams.length : 0;
  const implementationGenerics = Array.isArray(method.genericParams) ? method.genericParams.length : 0;
  if (interfaceGenerics !== implementationGenerics) {
    return {
      ok: false,
      detail: `${record.label}: method '${methodName}' expects ${interfaceGenerics} generic parameter(s), got ${implementationGenerics}`,
    };
  }
  const signatureParams = Array.isArray(signature.params) ? signature.params : [];
  const methodParams = Array.isArray(method.params) ? method.params : [];
  if (signatureParams.length !== methodParams.length) {
    return {
      ok: false,
      detail: `${record.label}: method '${methodName}' expects ${signatureParams.length} parameter(s), got ${methodParams.length}`,
    };
  }
  for (let index = 0; index < signatureParams.length; index += 1) {
    const expectedParam = signatureParams[index];
    const actualParam = methodParams[index];
    if (!expectedParam || !actualParam) {
      continue;
    }
    if (
      index === 0 &&
      helpers.isImplicitSelfParameter(actualParam) &&
      helpers.expectsSelfType(expectedParam.paramType)
    ) {
      actualParam.paramType = AST.simpleTypeExpression("Self");
    }
    const expectedDescription = ctx.describeTypeExpression(expectedParam.paramType, substitutions);
    const actualDescription = ctx.describeTypeExpression(actualParam.paramType, substitutions);
    if (!helpers.typeExpressionsEquivalent(ctx, expectedParam.paramType, actualParam.paramType, substitutions)) {
      return {
        ok: false,
        detail: `${record.label}: method '${methodName}' parameter ${index + 1} expected ${expectedDescription}, got ${actualDescription}`,
      };
    }
  }
  const expectedReturn = ctx.describeTypeExpression(signature.returnType, substitutions);
  const actualReturn = ctx.describeTypeExpression(method.returnType, substitutions);
  if (!helpers.typeExpressionsEquivalent(ctx, signature.returnType, method.returnType, substitutions)) {
    return {
      ok: false,
      detail: `${record.label}: method '${methodName}' return type expected ${expectedReturn}, got ${actualReturn}`,
    };
  }
  const signatureWhere = Array.isArray(signature.whereClause) ? signature.whereClause.length : 0;
  const methodWhere = Array.isArray(method.whereClause) ? method.whereClause.length : 0;
  if (signatureWhere !== methodWhere) {
    return {
      ok: false,
      detail: `${record.label}: method '${methodName}' expects ${signatureWhere} where-clause constraint(s), got ${methodWhere}`,
    };
  }
  if (method.isPrivate) {
    return {
      ok: false,
      detail: `${record.label}: method '${methodName}' must be public to satisfy interface`,
    };
  }
  return { ok: true };
}

function enforceMethodSetObligations(
  ctx: ImplementationContext,
  record: MethodSetRecord,
  type: TypeInfo,
  substitutions: Map<string, TypeInfo>,
  obligations: ImplementationObligation[],
  helpers: MethodSetHelpers,
): string | undefined {
  if (!obligations || obligations.length === 0) {
    return undefined;
  }
  for (const obligation of obligations) {
    const subject = helpers.lookupObligationSubject(ctx, obligation.typeParam, substitutions, type);
    if (!subject) {
      continue;
    }
    const obligationArgs = helpers.resolveInterfaceArgumentLabels(ctx, obligation.interfaceType, substitutions);
    const result = helpers.typeImplementsInterface(ctx, subject, obligation.interfaceName, obligationArgs);
    if (!result.ok) {
      return annotateMethodSetFailure(record, obligation, subject, result.detail, obligationArgs);
    }
  }
  return undefined;
}

function annotateMethodSetFailure(
  record: MethodSetRecord,
  obligation: ImplementationObligation,
  subject: TypeInfo,
  detail: string | undefined,
  expectedArgs: string[],
): string {
  const contextSuffix = obligation.context ? ` (${obligation.context})` : "";
  const subjectLabel = subject && subject.kind !== "unknown" ? ` (got ${formatType(subject)})` : "";
  const expectedSuffix = expectedArgs.length > 0 ? ` expects ${expectedArgs.join(" ")}` : "";
  const detailSuffix = detail ? `: ${detail}` : "";
  return `${record.label}: constraint on ${obligation.typeParam}${contextSuffix} requires ${obligation.interfaceName}${expectedSuffix}${subjectLabel}${detailSuffix}`;
}

function extractFunctionObligations(
  ctx: ImplementationContext,
  fn: AST.FunctionDefinition,
  context: string,
): ImplementationObligation[] {
  const obligations: ImplementationObligation[] = [];
  if (Array.isArray(fn.genericParams)) {
    for (const param of fn.genericParams) {
      const paramName = ctx.getIdentifierName(param?.name);
      if (!paramName || !Array.isArray(param?.constraints)) continue;
      for (const constraint of param.constraints) {
        const interfaceName = ctx.getInterfaceNameFromConstraint(constraint);
        if (!interfaceName) continue;
        obligations.push({
          typeParam: paramName,
          interfaceName,
          interfaceType: constraint?.interfaceType ?? undefined,
          context,
        });
      }
    }
  }
  if (Array.isArray(fn.whereClause)) {
    for (const clause of fn.whereClause) {
      const typeParamName = ctx.getIdentifierName(clause?.typeParam);
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        const interfaceName = ctx.getInterfaceNameFromConstraint(constraint);
        if (!interfaceName) continue;
        obligations.push({
          typeParam: typeParamName,
          interfaceName,
          interfaceType: constraint?.interfaceType ?? undefined,
          context,
        });
      }
    }
  }
  return obligations;
}
