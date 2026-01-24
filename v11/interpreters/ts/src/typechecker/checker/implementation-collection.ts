import * as AST from "../../ast";
import { collectFunctionDefinition } from "./declarations";
import { collectUnionVariantLabels } from "./impl_matches";
import type { ImplementationContext } from "./implementation-context";
import {
  canonicalizeTargetType,
  collectTargetTypeParams,
} from "./implementation-collection-helpers";
import {
  collectInterfaceGenericParamNames,
  containsWildcardTypeExpression,
  ensureImplementationMethods,
  validateImplementationSelfTypePattern,
} from "./implementation-validation";
import type { ImplementationObligation, ImplementationRecord, MethodSetRecord } from "./types";
import { unknownType, type TypeInfo } from "../types";
import { typeInfoToTypeExpression } from "./type-expression-utils";

export function collectMethodsDefinition(ctx: ImplementationContext, definition: AST.MethodsDefinition): void {
  const targetType = canonicalizeTargetType(ctx, definition.targetType);
  const explicitParams = Array.isArray(definition.genericParams)
    ? definition.genericParams
        .map((param) => ctx.getIdentifierName(param?.name))
        .filter((name): name is string => Boolean(name))
    : [];
  const targetParams = collectTargetTypeParams(ctx, targetType);
  const genericParams = [...new Set([...targetParams, ...explicitParams])];
  const substitutionMap = new Map<string, TypeInfo>();
  genericParams.forEach((name) => substitutionMap.set(name, unknownType));
  const selfType = ctx.resolveTypeExpression(targetType, substitutionMap, { allowTypeConstructors: true });
  const canonicalTarget = targetType ?? definition.targetType;
  const structLabel =
    ctx.describeTypeArgument(selfType ?? unknownType) ??
    ctx.formatImplementationTarget(canonicalTarget) ??
    ctx.getIdentifierNameFromTypeExpression(canonicalTarget);
  const structBaseName =
    ctx.getIdentifierNameFromTypeExpression?.(canonicalTarget) ??
    ctx.getIdentifierNameFromTypeExpression?.(definition.targetType);
  const structName = structBaseName ?? structLabel;
  if (!structLabel) return;
  const record: MethodSetRecord = {
    label: `methods for ${structLabel}`,
    target: canonicalTarget,
    genericParams,
    obligations: extractMethodSetObligations(ctx, definition),
    definition,
    resolvedTarget: selfType ?? unknownType,
    packageName: ctx.getCurrentPackageName?.(),
  };
  ctx.registerMethodSet(record);
  const methodObligations = Array.isArray(record.obligations) ? record.obligations : [];
  if (Array.isArray(definition.definitions)) {
    for (const entry of definition.definitions) {
      if (entry?.type === "FunctionDefinition") {
        collectFunctionDefinition(ctx, entry, {
          structName,
          structBaseName,
          typeParamNames: record.genericParams,
          fromMethodSet: true,
        });
        if (entry.id?.name && methodObligations.length > 0) {
          const fullName = `${structName ?? structLabel}::${entry.id.name}`;
          appendMethodSetObligations(ctx, fullName, methodObligations, selfType ?? unknownType);
        }
      }
    }
  }
}

export function collectImplementationDefinition(
  ctx: ImplementationContext,
  definition: AST.ImplementationDefinition,
): void {
  const targetType = canonicalizeTargetType(ctx, definition.targetType);
  const interfaceName = ctx.getIdentifierName(definition.interfaceName);
  if (!interfaceName) {
    return;
  }
  if (definition.implName?.name) {
    ctx.defineValue(definition.implName.name, unknownType);
  }
  const targetLabel = ctx.formatImplementationTarget(targetType);
  const fallbackName = ctx.getIdentifierNameFromTypeExpression(targetType);
  const initialContextName = targetLabel ?? fallbackName ?? "<unknown>";
  let contextName = initialContextName;
  const interfaceDefinition = ctx.getInterfaceDefinition(interfaceName);
  if (!interfaceDefinition) {
    const fallback = ctx.getIdentifierNameFromTypeExpression(targetType);
    ctx.report(
      `typechecker: impl for ${fallback ?? "<unknown>"} references unknown interface '${interfaceName}'`,
      definition,
    );
    return;
  }
  const interfaceArgs = resolveImplementationInterfaceArguments(
    ctx,
    definition,
    interfaceDefinition,
    contextName,
    interfaceName,
  );
  const interfaceGenericNames = collectInterfaceGenericParamNames(ctx, interfaceDefinition);
  const explicitImplementationGenerics = collectImplementationGenericParamNames(ctx, definition);
  const targetParamNames = collectTargetTypeParams(ctx, targetType);
  const implementationGenericNames = [...new Set([...explicitImplementationGenerics, ...targetParamNames])];
  const implementationGenericNameSet = new Set(implementationGenericNames);
  const substitutionMap = new Map<string, TypeInfo>();
  implementationGenericNames.forEach((name) => substitutionMap.set(name, unknownType));
  const resolvedTarget =
    targetType ?? definition.targetType
      ? ctx.resolveTypeExpression(targetType ?? definition.targetType, substitutionMap, { allowTypeConstructors: true })
      : unknownType;
  const resolvedTargetExpr = typeInfoToTypeExpression(resolvedTarget ?? unknownType);
  const canonicalTarget =
    resolvedTargetExpr &&
    resolvedTargetExpr.type !== "WildcardTypeExpression" &&
    !containsWildcardTypeExpression(resolvedTargetExpr)
      ? resolvedTargetExpr
      : targetType ?? definition.targetType;
  const resolvedLabel = resolvedTarget && resolvedTarget.kind !== "unknown"
    ? ctx.describeTypeArgument(resolvedTarget)
    : null;
  const canonicalTargetLabel = ctx.formatImplementationTarget(canonicalTarget) ?? resolvedLabel ?? initialContextName;
  contextName = canonicalTargetLabel ?? initialContextName;
  const targetValid = validateImplementationSelfTypePattern(
    ctx,
    canonicalTarget ?? targetType ?? definition.targetType,
    definition,
    interfaceDefinition,
    canonicalTargetLabel ?? contextName,
    interfaceName,
    interfaceGenericNames,
    implementationGenericNameSet,
  );
  const hasRequiredMethods =
    targetValid &&
    ensureImplementationMethods(ctx, definition, interfaceDefinition, contextName, interfaceName);
  if (targetValid && hasRequiredMethods) {
    const record = createImplementationRecord(
      ctx,
      definition,
      interfaceName,
      contextName,
      contextName,
      implementationGenericNames,
      interfaceArgs,
      canonicalTarget ?? targetType ?? definition.targetType,
      resolvedTarget ?? unknownType,
    );
    if (record) {
      ctx.registerImplementationRecord(record);
    }
  }

  if (Array.isArray(definition.definitions)) {
    for (const entry of definition.definitions) {
      if (entry?.type === "FunctionDefinition") {
        collectFunctionDefinition(ctx, entry, {
          structName: contextName,
          typeParamNames: implementationGenericNames,
        });
      }
    }
  }
}

function extractMethodSetObligations(
  ctx: ImplementationContext,
  definition: AST.MethodsDefinition,
): ImplementationObligation[] {
  const obligations: ImplementationObligation[] = [];
  const appendObligation = (
    typeParam: string | null,
    interfaceType: AST.TypeExpression | null | undefined,
    context: string,
    subjectExpr?: AST.TypeExpression,
  ) => {
    const interfaceName = ctx.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context, subjectExpr });
  };

  if (Array.isArray(definition.genericParams)) {
    for (const param of definition.genericParams) {
      const paramName = ctx.getIdentifierName(param?.name);
      if (!paramName || !Array.isArray(param?.constraints)) continue;
      for (const constraint of param.constraints) {
        appendObligation(paramName, constraint?.interfaceType, "generic constraint");
      }
    }
  }

  if (Array.isArray(definition.whereClause)) {
    for (const clause of definition.whereClause) {
      const subjectExpr = clause?.typeParam;
      const typeParamName = subjectExpr ? ctx.formatTypeExpression(subjectExpr) : null;
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause", subjectExpr);
      }
    }
  }

  return obligations;
}

function appendMethodSetObligations(
  ctx: ImplementationContext,
  key: string,
  obligations: ImplementationObligation[],
  selfType: TypeInfo,
): void {
  if (!key || !Array.isArray(obligations) || obligations.length === 0) {
    return;
  }
  const infos = ctx.getFunctionInfos(key);
  if (!Array.isArray(infos) || infos.length === 0) {
    return;
  }
  for (const info of infos) {
    if (!info) {
      continue;
    }
    const existing = Array.isArray(info.whereClause) ? info.whereClause : [];
    info.whereClause = [...existing, ...obligations];
    if (!info.methodSetSubstitutions || !info.methodSetSubstitutions.length) {
      info.methodSetSubstitutions = [["Self", selfType]];
    }
  }
}

function extractImplementationObligations(
  ctx: ImplementationContext,
  definition: AST.ImplementationDefinition,
): ImplementationObligation[] {
  const obligations: ImplementationObligation[] = [];
  const appendObligation = (
    typeParam: string | null,
    interfaceType: AST.TypeExpression | null | undefined,
    context: string,
    subjectExpr?: AST.TypeExpression,
  ) => {
    const interfaceName = ctx.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context, subjectExpr });
  };

  if (Array.isArray(definition.genericParams)) {
    for (const param of definition.genericParams) {
      const paramName = ctx.getIdentifierName(param?.name);
      if (!paramName || !Array.isArray(param?.constraints)) continue;
      for (const constraint of param.constraints) {
        appendObligation(paramName, constraint?.interfaceType, "generic constraint");
      }
    }
  }

  if (Array.isArray(definition.whereClause)) {
    for (const clause of definition.whereClause) {
      const subjectExpr = clause?.typeParam;
      const typeParamName = subjectExpr ? ctx.formatTypeExpression(subjectExpr) : null;
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause", subjectExpr);
      }
    }
  }

  return obligations;
}

function createImplementationRecord(
  ctx: ImplementationContext,
  definition: AST.ImplementationDefinition,
  interfaceName: string,
  targetLabel: string,
  targetKey: string,
  implementationGenericNames?: string[],
  interfaceArgs: AST.TypeExpression[] = [],
  targetType?: AST.TypeExpression,
  resolvedTarget?: TypeInfo,
): ImplementationRecord | null {
  const resolvedTargetExpr = targetType ?? definition.targetType;
  if (!resolvedTargetExpr) {
    return null;
  }
  const genericParams =
    implementationGenericNames ??
    (Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((name): name is string => Boolean(name))
      : []);
  const obligations = extractImplementationObligations(ctx, definition);
  const unionVariants = collectUnionVariantLabels(ctx, resolvedTarget);
  return {
    packageName: ctx.getCurrentPackageName?.(),
    interfaceName,
    label: ctx.formatImplementationLabel(interfaceName, targetLabel),
    target: resolvedTargetExpr,
    targetKey,
    genericParams,
    obligations,
    interfaceArgs,
    unionVariants,
    resolvedTarget,
    definition,
  };
}

function collectImplementationGenericParamNames(
  ctx: ImplementationContext,
  definition: AST.ImplementationDefinition,
): string[] {
  if (!Array.isArray(definition.genericParams)) {
    return [];
  }
  return definition.genericParams
    .map((param) => ctx.getIdentifierName(param?.name))
    .filter((name): name is string => Boolean(name));
}

function resolveImplementationInterfaceArguments(
  ctx: ImplementationContext,
  implementation: AST.ImplementationDefinition,
  interfaceDefinition: AST.InterfaceDefinition,
  targetLabel: string,
  interfaceName: string,
): AST.TypeExpression[] {
  const expected = Array.isArray(interfaceDefinition.genericParams) ? interfaceDefinition.genericParams.length : 0;
  const rawArgs = Array.isArray(implementation.interfaceArgs)
    ? implementation.interfaceArgs.filter((arg): arg is AST.TypeExpression => Boolean(arg))
    : [];
  const provided = rawArgs.length;
  if (expected === 0 && provided > 0) {
    ctx.report(`typechecker: impl ${interfaceName} does not accept type arguments`, implementation);
    return rawArgs;
  }
  if (expected > 0) {
    const targetDescription = targetLabel;
    if (provided === 0) {
      ctx.report(
        `typechecker: impl ${interfaceName} for ${targetDescription} requires ${expected} interface type argument(s)`,
        implementation,
      );
      return rawArgs;
    }
    if (provided !== expected) {
      ctx.report(
        `typechecker: impl ${interfaceName} for ${targetDescription} expected ${expected} interface type argument(s), got ${provided}`,
        implementation,
      );
    }
  }
  return rawArgs;
}
