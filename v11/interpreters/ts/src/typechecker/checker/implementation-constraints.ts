import * as AST from "../../ast";
import { formatType, unknownType } from "../types";
import type { TypeInfo } from "../types";
import {
  buildConstraintKeySet,
  computeImplementationSpecificity,
  selectMostSpecificImplementationMatch,
  type ImplementationMatch,
} from "./impl_matches";
import type { ImplementationContext } from "./implementation-context";
import { methodSetProvidesInterface } from "./method_sets";
import {
  expectsSelfType,
  isImplicitSelfParameter,
  typeExpressionsEquivalent,
} from "./implementation-collection";
import type {
  FunctionInfo,
  ImplementationObligation,
  ImplementationRecord,
  InterfaceCheckResult,
  MethodSetRecord,
} from "./types";

function methodDefinitionHasImplicitSelf(ctx: ImplementationContext, method: AST.FunctionDefinition): boolean {
  if (Array.isArray(method.params) && method.params.length > 0) {
    const first = method.params[0];
    const name = ctx.getIdentifierName(first?.name)?.toLowerCase();
    if (name === "self") {
      if (!first.paramType) {
        first.paramType = AST.simpleTypeExpression("Self");
      }
      return true;
    }
    if (
      first?.paramType?.type === "SimpleTypeExpression" &&
      ctx.getIdentifierName(first.paramType.name) === "Self"
    ) {
      return true;
    }
  }
  return Boolean(method.isMethodShorthand);
}

export function lookupMethodSetsForCall(
  ctx: ImplementationContext,
  structLabel: string,
  methodName: string,
  objectType: TypeInfo,
): FunctionInfo[] {
  const results: FunctionInfo[] = [];
  for (const record of ctx.getMethodSets()) {
    const paramNames = new Set(record.genericParams);
    const substitutions = new Map<string, TypeInfo>();
    substitutions.set("Self", objectType);
    if (!matchImplementationTarget(ctx, objectType, record.target, paramNames, substitutions)) {
      continue;
    }
    const method = record.definition.definitions?.find(
      (fn): fn is AST.FunctionDefinition => fn?.type === "FunctionDefinition" && fn.id?.name === methodName,
    );
    if (!method) {
      continue;
    }
    const hasImplicitSelf = methodDefinitionHasImplicitSelf(ctx, method);
    const methodGenericNames = Array.isArray(method.genericParams)
      ? method.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((name): name is string => Boolean(name))
      : [];
    const signatureSubstitutions = new Map(substitutions);
    for (const name of methodGenericNames) {
      signatureSubstitutions.set(name, unknownType);
    }
    const parameterTypes = Array.isArray(method.params)
      ? method.params.map((param) => ctx.resolveTypeExpression(param?.paramType, signatureSubstitutions))
      : [];
    const info: FunctionInfo = {
      name: methodName,
      fullName: `${record.label}::${methodName}`,
      structName: structLabel,
      hasImplicitSelf,
      parameters: parameterTypes,
      genericConstraints: [],
      genericParamNames: methodGenericNames,
      whereClause: record.obligations,
      methodSetSubstitutions: Array.from(substitutions.entries()),
      returnType: ctx.resolveTypeExpression(method.returnType, signatureSubstitutions),
    };
    if (Array.isArray(method.genericParams)) {
      for (const param of method.genericParams) {
        const paramName = ctx.getIdentifierName(param?.name);
        if (!paramName || !Array.isArray(param?.constraints)) continue;
        for (const constraint of param.constraints) {
          const interfaceName = ctx.getInterfaceNameFromConstraint(constraint);
          info.genericConstraints.push({
            paramName,
            interfaceName: interfaceName ?? "<unknown>",
            interfaceDefined: !!interfaceName,
            interfaceType: constraint?.interfaceType,
          });
        }
      }
    }
    results.push(info);
  }
  return results;
}

export function enforceFunctionConstraints(
  ctx: ImplementationContext,
  info: FunctionInfo,
  call: AST.FunctionCall,
): void {
  const typeArgs = Array.isArray(call.typeArguments) ? call.typeArguments : [];
  const substitutions = new Map<string, TypeInfo>();
  if (info.methodSetSubstitutions) {
    for (const [key, value] of info.methodSetSubstitutions) {
      substitutions.set(key, value);
    }
  } else if (call.callee?.type === "MemberAccessExpression") {
    const selfType = ctx.inferExpression(call.callee.object);
    if (selfType.kind !== "unknown") {
      substitutions.set("Self", selfType);
    }
  }
  info.genericParamNames.forEach((paramName, idx) => {
    const argExpr = typeArgs[idx];
    if (!paramName || !argExpr) return;
    substitutions.set(paramName, ctx.resolveTypeExpression(argExpr));
  });

  if (info.genericConstraints.length > 0) {
    info.genericConstraints.forEach((constraint, index) => {
      const typeArgExpr = typeArgs[index];
      const typeArg = ctx.resolveTypeExpression(typeArgExpr);
      if (!constraint.interfaceDefined) {
        const message = info.structName
          ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${constraint.paramName} references unknown interface '${constraint.interfaceName}'`
          : `typechecker: fn ${info.name} constraint on ${constraint.paramName} references unknown interface '${constraint.interfaceName}'`;
        ctx.report(message, typeArgExpr ?? call);
        return;
      }
      const expectedArgs = resolveInterfaceArgumentLabels(ctx, constraint.interfaceType, substitutions);
      const result = typeImplementsInterface(ctx, typeArg, constraint.interfaceName, expectedArgs);
      if (!result.ok) {
        const typeName = ctx.describeTypeArgument(typeArg);
        const detailSuffix = result.detail ? `: ${result.detail}` : "";
        const message = info.structName
          ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${constraint.paramName} is not satisfied: ${typeName} does not implement ${constraint.interfaceName}${detailSuffix}`
          : `typechecker: fn ${info.name} constraint on ${constraint.paramName} is not satisfied: ${typeName} does not implement ${constraint.interfaceName}${detailSuffix}`;
        ctx.report(message, typeArgExpr ?? call);
      }
    });
  }

  if (info.whereClause.length > 0) {
    for (const obligation of info.whereClause) {
      const subject = lookupObligationSubject(ctx, obligation.typeParam, substitutions, substitutions.get("Self") ?? unknownType);
      if (!subject) {
        continue;
      }
      const subjectLabel = ctx.describeTypeArgument(subject);
      const obligationArgs = resolveInterfaceArgumentLabels(ctx, obligation.interfaceType, substitutions);
      const result = typeImplementsInterface(ctx, subject, obligation.interfaceName, obligationArgs);
      if (!result.ok) {
        const detailSuffix = result.detail ? `: ${result.detail}` : "";
        const message = info.structName
          ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${obligation.typeParam} is not satisfied: ${subjectLabel} does not implement ${obligation.interfaceName}${detailSuffix}`
          : `typechecker: fn ${info.name} constraint on ${obligation.typeParam} is not satisfied: ${subjectLabel} does not implement ${obligation.interfaceName}${detailSuffix}`;
        ctx.report(message, call);
      }
    }
  }
}

export function typeImplementsInterface(
  ctx: ImplementationContext,
  type: TypeInfo,
  interfaceName: string,
  expectedArgs: string[] = [],
): InterfaceCheckResult {
  if (!type || type.kind === "unknown") {
    return { ok: true };
  }
  if (type.kind === "nullable") {
    const impl = implementationProvidesInterface(ctx, type, interfaceName, expectedArgs);
    if (impl.ok) {
      return impl;
    }
    const inner = typeImplementsInterface(ctx, type.inner, interfaceName, expectedArgs);
    if (!inner.ok) {
      return inner.detail ? inner : impl.detail ? { ok: false, detail: impl.detail } : inner;
    }
    return impl.detail ? { ok: false, detail: impl.detail } : { ok: true };
  }
  if (type.kind === "result") {
    const impl = implementationProvidesInterface(ctx, type, interfaceName, expectedArgs);
    if (impl.ok) {
      return impl;
    }
    const inner = typeImplementsInterface(ctx, type.inner, interfaceName, expectedArgs);
    if (!inner.ok) {
      return inner.detail ? inner : impl.detail ? { ok: false, detail: impl.detail } : inner;
    }
    return impl.detail ? { ok: false, detail: impl.detail } : { ok: true };
  }
  if (type.kind === "union") {
    const impl = implementationProvidesInterface(ctx, type, interfaceName, expectedArgs);
    if (impl.ok) {
      return impl;
    }
    for (const member of type.members) {
      const result = typeImplementsInterface(ctx, member, interfaceName, expectedArgs);
      if (!result.ok) {
        return result.detail ? result : impl.detail ? { ok: false, detail: impl.detail } : result;
      }
    }
    return impl.detail ? { ok: false, detail: impl.detail } : { ok: true };
  }
  if (type.kind === "interface" && type.name === interfaceName) {
    return { ok: true };
  }
  const impl = implementationProvidesInterface(ctx, type, interfaceName, expectedArgs);
  if (impl.ok) {
    return impl;
  }
  if (impl.detail) {
    return { ok: false, detail: impl.detail };
  }
  const methodSetResult = methodSetProvidesInterface(ctx, type, interfaceName, expectedArgs, {
    matchImplementationTarget,
    buildStringSubstitutionMap,
    typeExpressionsEquivalent,
    isImplicitSelfParameter,
    expectsSelfType,
    lookupObligationSubject,
    resolveInterfaceArgumentLabels,
    typeImplementsInterface,
  });
  if (methodSetResult.ok) {
    return methodSetResult;
  }
  if (methodSetResult.detail) {
    return { ok: false, detail: methodSetResult.detail };
  }
  return { ok: false };
}

function resolveInterfaceArgumentLabels(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
  substitutions?: Map<string, TypeInfo>,
): string[] {
  if (!expr || expr.type !== "GenericTypeExpression") {
    return [];
  }
  return resolveInterfaceArgumentLabelsFromArray(ctx, expr.arguments ?? [], substitutions);
}

function resolveInterfaceArgumentLabelsFromArray(
  ctx: ImplementationContext,
  args: Array<AST.TypeExpression | null | undefined>,
  substitutions?: Map<string, TypeInfo>,
): string[] {
  if (!args || args.length === 0) {
    return [];
  }
  const stringSubs = substitutions ? buildStringSubstitutionMap(substitutions) : undefined;
  return args.map((arg) => (arg ? ctx.formatTypeExpression(arg, stringSubs) : "Unknown"));
}

function buildStringSubstitutionMap(substitutions: Map<string, TypeInfo>): Map<string, string> {
  const result = new Map<string, string>();
  substitutions.forEach((value, key) => {
    result.set(key, formatType(value));
  });
  return result;
}

function implementationProvidesInterface(
  ctx: ImplementationContext,
  type: TypeInfo,
  interfaceName: string,
  expectedArgs: string[] = [],
): InterfaceCheckResult {
  const candidates = lookupImplementationCandidates(ctx, type);
  const matches: ImplementationMatch[] = [];
  let bestDetail: string | undefined;
  for (const record of candidates) {
    if (record.interfaceName !== interfaceName) {
      continue;
    }
    const paramNames = new Set(record.genericParams);
    const substitutions = new Map<string, TypeInfo>();
    substitutions.set("Self", type);
    if (!matchImplementationTarget(ctx, type, record.target, paramNames, substitutions)) {
      continue;
    }
    const actualArgs = record.interfaceArgs.length
      ? resolveInterfaceArgumentLabelsFromArray(ctx, record.interfaceArgs, substitutions)
      : [];
    if (!interfaceArgsCompatible(actualArgs, expectedArgs)) {
      const expectedLabel = expectedArgs.length > 0 ? expectedArgs.join(" ") : "(none)";
      const detail = `${ctx.appendInterfaceArgsToLabel(record.label, actualArgs)}: interface arguments do not match expected ${expectedLabel}`;
      if (!bestDetail || detail.length > bestDetail.length) {
        bestDetail = detail;
      }
      continue;
    }
    let failedDetail: string | undefined;
    for (const obligation of record.obligations) {
      const subject = lookupObligationSubject(ctx, obligation.typeParam, substitutions, type);
      if (!subject) {
        continue;
      }
      const obligationArgs = resolveInterfaceArgumentLabels(ctx, obligation.interfaceType, substitutions);
      const result = typeImplementsInterface(ctx, subject, obligation.interfaceName, obligationArgs);
      if (!result.ok) {
        const detail = annotateImplementationFailure(
          ctx,
          record,
          obligation,
          subject,
          result.detail,
          actualArgs,
          obligationArgs,
        );
        if (!bestDetail || detail.length > bestDetail.length) {
          bestDetail = detail;
        }
        failedDetail = detail;
        break;
      }
    }
    if (failedDetail) {
      continue;
    }
    matches.push({
      record,
      substitutions,
      interfaceArgs: actualArgs,
      score: computeImplementationSpecificity(ctx, record, substitutions),
      constraintKeys: buildConstraintKeySet(ctx, record.obligations),
    });
  }
  if (matches.length === 0) {
    return bestDetail ? { ok: false, detail: bestDetail } : { ok: false };
  }
  if (matches.length === 1) {
    return { ok: true };
  }
  const resolution = selectMostSpecificImplementationMatch(ctx, matches, interfaceName, type);
  if (resolution.ok) {
    return { ok: true };
  }
  return { ok: false, detail: resolution.detail ?? bestDetail };
}

function lookupImplementationCandidates(ctx: ImplementationContext, type: TypeInfo): ImplementationRecord[] {
  const key = formatType(type);
  const seen = new Set<ImplementationRecord>();
  const direct = ctx.getImplementationBucket(key);
  if (direct) {
    for (const record of direct) {
      seen.add(record);
    }
  }
  for (const record of ctx.getImplementationRecords()) {
    seen.add(record);
  }
  return Array.from(seen);
}

function matchImplementationTarget(
  ctx: ImplementationContext,
  actual: TypeInfo,
  target: AST.TypeExpression,
  paramNames: Set<string>,
  substitutions: Map<string, TypeInfo>,
): boolean {
  if (!target) {
    return false;
  }
  if (!actual || actual.kind === "unknown") {
    return true;
  }
  switch (target.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(target.name);
      if (!name) {
        return false;
      }
      if (name === "Self") {
        const existing = substitutions.get("Self");
        if (existing) {
          return ctx.typeInfosEquivalent(existing, actual);
        }
        substitutions.set("Self", actual);
        return true;
      }
      if (paramNames.has(name)) {
        const existing = substitutions.get(name);
        if (existing) {
          return ctx.typeInfosEquivalent(existing, actual);
        }
        substitutions.set(name, actual);
        return true;
      }
      if (actual.kind === "primitive") {
        return actual.name === name;
      }
      if (actual.kind === "struct") {
        return actual.name === name && (actual.typeArguments?.length ?? 0) === 0;
      }
      if (actual.kind === "interface") {
        return actual.name === name && (actual.typeArguments?.length ?? 0) === 0;
      }
      return formatType(actual) === name;
    }
    case "GenericTypeExpression": {
      const baseName = ctx.getIdentifierNameFromTypeExpression(target.base);
      if (!baseName) {
        return false;
      }
      if (paramNames.has(baseName)) {
        const existing = substitutions.get(baseName);
        if (existing) {
          return ctx.typeInfosEquivalent(existing, actual);
        }
        substitutions.set(baseName, actual);
        return true;
      }
      if (actual.kind !== "struct" && actual.kind !== "interface") {
        return false;
      }
      if (actual.name !== baseName) {
        return false;
      }
      const expectedArgs = Array.isArray(target.arguments) ? target.arguments : [];
      const actualArgs = actual.typeArguments ?? [];
      if (expectedArgs.length !== actualArgs.length) {
        return false;
      }
      for (let index = 0; index < expectedArgs.length; index += 1) {
        const expectedArg = expectedArgs[index];
        const actualArg = actualArgs[index] ?? unknownType;
        if (!expectedArg) {
          return false;
        }
        if (!matchImplementationTarget(ctx, actualArg, expectedArg, paramNames, substitutions)) {
          return false;
        }
      }
      return true;
    }
    case "NullableTypeExpression":
      if (actual.kind !== "nullable") {
        return false;
      }
      return matchImplementationTarget(ctx, actual.inner, target.innerType, paramNames, substitutions);
    case "ResultTypeExpression":
      if (actual.kind !== "result") {
        return false;
      }
      return matchImplementationTarget(ctx, actual.inner, target.innerType, paramNames, substitutions);
    case "UnionTypeExpression": {
      const expectedMembers = Array.isArray(target.members) ? target.members : [];
      if (actual.kind === "union") {
        if (expectedMembers.length !== actual.members.length) {
          return false;
        }
        for (let index = 0; index < expectedMembers.length; index += 1) {
          const expectedMember = expectedMembers[index];
          const actualMember = actual.members[index];
          if (!expectedMember) {
            return false;
          }
          if (!matchImplementationTarget(ctx, actualMember, expectedMember, paramNames, substitutions)) {
            return false;
          }
        }
        return true;
      }
      for (const member of expectedMembers) {
        if (!member) {
          continue;
        }
        const snapshot = new Map(substitutions);
        if (matchImplementationTarget(ctx, actual, member, paramNames, snapshot)) {
          snapshot.forEach((value, key) => substitutions.set(key, value));
          return true;
        }
      }
      return false;
    }
    case "FunctionTypeExpression":
      return actual.kind === "function";
    default:
      return formatType(actual) === ctx.formatTypeExpression(target);
  }
}

function lookupObligationSubject(
  ctx: ImplementationContext,
  typeParam: string,
  substitutions: Map<string, TypeInfo>,
  selfType: TypeInfo,
): TypeInfo | null {
  if (typeParam === "Self") {
    return selfType;
  }
  if (substitutions.has(typeParam)) {
    return substitutions.get(typeParam) ?? unknownType;
  }
  return unknownType;
}

function annotateImplementationFailure(
  ctx: ImplementationContext,
  record: ImplementationRecord,
  obligation: ImplementationObligation,
  subject: TypeInfo,
  detail: string | undefined,
  actualArgs: string[],
  expectedArgs: string[],
): string {
  const label = ctx.appendInterfaceArgsToLabel(record.label, actualArgs);
  const contextSuffix = obligation.context ? ` (${obligation.context})` : "";
  const subjectLabel = subject && subject.kind !== "unknown" ? ` (got ${formatType(subject)})` : "";
  const expectedSuffix = expectedArgs.length ? ` expects ${expectedArgs.join(" ")}` : "";
  const detailSuffix = detail ? `: ${detail}` : "";
  return `${label}: constraint on ${obligation.typeParam}${contextSuffix} requires ${obligation.interfaceName}${expectedSuffix}${subjectLabel}${detailSuffix}`;
}

function interfaceArgsCompatible(actual: string[], expected: string[]): boolean {
  if (actual.length !== expected.length) {
    return false;
  }
  for (let index = 0; index < expected.length; index += 1) {
    const exp = expected[index];
    const act = actual[index];
    if (exp === act) {
      continue;
    }
    if (exp === "Unknown" || act === "Unknown") {
      continue;
    }
    return false;
  }
  return true;
}
