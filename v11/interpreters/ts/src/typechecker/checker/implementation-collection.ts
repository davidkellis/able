import * as AST from "../../ast";
import { collectFunctionDefinition } from "./declarations";
import { collectUnionVariantLabels } from "./impl_matches";
import type { ImplementationContext } from "./implementation-context";
import type { ImplementationObligation, ImplementationRecord, MethodSetRecord } from "./types";
import { unknownType, type TypeInfo } from "../types";

const PRIMITIVE_TYPE_NAMES = new Set([
  "i8",
  "i16",
  "i32",
  "i64",
  "i128",
  "u8",
  "u16",
  "u32",
  "u64",
  "u128",
  "f32",
  "f64",
  "bool",
  "String",
  "char",
  "nil",
  "void",
]);

function collectTargetTypeParams(ctx: ImplementationContext, targetType: AST.TypeExpression | null | undefined): string[] {
  if (!targetType) return [];
  if (targetType.type === "GenericTypeExpression" && Array.isArray(targetType.arguments)) {
    return targetType.arguments
      .map((arg) => ctx.getIdentifierNameFromTypeExpression(arg))
      .filter((name): name is string => Boolean(name) && !ctx.isKnownTypeName(name));
  }
  return [];
}

export function collectMethodsDefinition(ctx: ImplementationContext, definition: AST.MethodsDefinition): void {
  const structLabel =
    ctx.formatImplementationTarget(definition.targetType) ?? ctx.getIdentifierNameFromTypeExpression(definition.targetType);
  if (!structLabel) return;
  const explicitParams = Array.isArray(definition.genericParams)
    ? definition.genericParams
        .map((param) => ctx.getIdentifierName(param?.name))
        .filter((name): name is string => Boolean(name))
    : [];
  const targetParams = collectTargetTypeParams(ctx, definition.targetType);
  const genericParams = [...new Set([...targetParams, ...explicitParams])];
  const substitutionMap = new Map<string, TypeInfo>();
  genericParams.forEach((name) => substitutionMap.set(name, unknownType));
  const selfType = ctx.resolveTypeExpression(definition.targetType, substitutionMap);
  const record: MethodSetRecord = {
    label: `methods for ${structLabel}`,
    target: definition.targetType,
    genericParams,
    obligations: extractMethodSetObligations(ctx, definition),
    definition,
  };
  ctx.registerMethodSet(record);
  const methodObligations = Array.isArray(record.obligations) ? record.obligations : [];
  if (Array.isArray(definition.definitions)) {
    for (const entry of definition.definitions) {
      if (entry?.type === "FunctionDefinition") {
        collectFunctionDefinition(ctx, entry, { structName: structLabel, typeParamNames: record.genericParams });
        if (entry.id?.name && methodObligations.length > 0) {
          const fullName = `${structLabel}::${entry.id.name}`;
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
  const interfaceName = ctx.getIdentifierName(definition.interfaceName);
  if (!interfaceName) {
    return;
  }
  if (definition.implName?.name) {
    ctx.defineValue(definition.implName.name, unknownType);
  }
  const targetLabel = ctx.formatImplementationTarget(definition.targetType);
  const fallbackName = ctx.getIdentifierNameFromTypeExpression(definition.targetType);
  const contextName = targetLabel ?? fallbackName ?? "<unknown>";
  const targetKey = contextName;
  const interfaceDefinition = ctx.getInterfaceDefinition(interfaceName);
  if (!interfaceDefinition) {
    const fallback = ctx.getIdentifierNameFromTypeExpression(definition.targetType);
    ctx.report(
      `typechecker: impl for ${fallback ?? "<unknown>"} references unknown interface '${interfaceName}'`,
      definition,
    );
    return;
  }
  validateImplementationInterfaceArguments(ctx, definition, interfaceDefinition, contextName, interfaceName);
  const interfaceGenericNames = collectInterfaceGenericParamNames(ctx, interfaceDefinition);
  const implementationGenericNames = collectImplementationGenericParamNames(ctx, definition);
  const implementationGenericNameSet = new Set(implementationGenericNames);
  const targetValid = validateImplementationSelfTypePattern(
    ctx,
    definition,
    interfaceDefinition,
    contextName,
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
      targetKey,
      implementationGenericNames,
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
  ) => {
    const interfaceName = ctx.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
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
      const typeParamName = ctx.getIdentifierName(clause?.typeParam);
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause");
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
  ) => {
    const interfaceName = ctx.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
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
      const typeParamName = ctx.getIdentifierName(clause?.typeParam);
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause");
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
): ImplementationRecord | null {
  if (!definition.targetType) {
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
  const interfaceArgs = Array.isArray(definition.interfaceArgs)
    ? definition.interfaceArgs.filter((arg): arg is AST.TypeExpression => Boolean(arg))
    : [];
  const unionVariants = collectUnionVariantLabels(ctx, definition.targetType);
  return {
    interfaceName,
    label: ctx.formatImplementationLabel(interfaceName, targetLabel),
    target: definition.targetType,
    targetKey,
    genericParams,
    obligations,
    interfaceArgs,
    unionVariants,
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

function collectInterfaceGenericParamNames(
  ctx: ImplementationContext,
  definition: AST.InterfaceDefinition,
): Set<string> {
  const names = new Set<string>();
  if (!Array.isArray(definition.genericParams)) {
    return names;
  }
  for (const param of definition.genericParams) {
    const name = ctx.getIdentifierName(param?.name);
    if (name) {
      names.add(name);
    }
  }
  return names;
}

function validateImplementationInterfaceArguments(
  ctx: ImplementationContext,
  implementation: AST.ImplementationDefinition,
  interfaceDefinition: AST.InterfaceDefinition,
  targetLabel: string,
  interfaceName: string,
): void {
  const expected = Array.isArray(interfaceDefinition.genericParams) ? interfaceDefinition.genericParams.length : 0;
  const provided = Array.isArray(implementation.interfaceArgs) ? implementation.interfaceArgs.length : 0;
  if (expected === 0 && provided > 0) {
    ctx.report(`typechecker: impl ${interfaceName} does not accept type arguments`, implementation);
    return;
  }
  if (expected > 0) {
    const targetDescription = targetLabel;
    if (provided === 0) {
      ctx.report(
        `typechecker: impl ${interfaceName} for ${targetDescription} requires ${expected} interface type argument(s)`,
        implementation,
      );
      return;
    }
    if (provided !== expected) {
      ctx.report(
        `typechecker: impl ${interfaceName} for ${targetDescription} expected ${expected} interface type argument(s), got ${provided}`,
        implementation,
      );
    }
  }
}

function validateImplementationSelfTypePattern(
  ctx: ImplementationContext,
  implementation: AST.ImplementationDefinition,
  interfaceDefinition: AST.InterfaceDefinition,
  targetLabel: string,
  interfaceName: string,
  interfaceGenericNames: Set<string>,
  implementationGenericNames: Set<string>,
): boolean {
  if (!implementation.targetType) {
    return false;
  }
  const selfPattern = interfaceDefinition.selfTypePattern;
  if (selfPattern) {
    const matches = doesSelfPatternMatchTarget(ctx, selfPattern, implementation.targetType, interfaceGenericNames);
    if (!matches) {
      const expected = ctx.formatTypeExpression(selfPattern);
      ctx.report(
        `typechecker: impl ${interfaceName} for ${targetLabel} must match interface self type '${expected}'`,
        implementation,
      );
      return false;
    }
    return true;
  }
  if (targetsBareTypeConstructor(ctx, implementation.targetType, implementationGenericNames)) {
    ctx.report(
      `typechecker: impl ${interfaceName} for ${targetLabel} cannot target a type constructor because the interface does not declare a self type (use 'for ...' to enable constructor implementations)`,
      implementation,
    );
    return false;
  }
  return true;
}

function doesSelfPatternMatchTarget(
  ctx: ImplementationContext,
  pattern: AST.TypeExpression,
  target: AST.TypeExpression,
  interfaceGenericNames: Set<string>,
): boolean {
  if (!pattern || !target) {
    return false;
  }
  return matchSelfTypePattern(ctx, pattern, target, interfaceGenericNames, new Map());
}

function matchSelfTypePattern(
  ctx: ImplementationContext,
  pattern: AST.TypeExpression,
  target: AST.TypeExpression,
  interfaceGenericNames: Set<string>,
  bindings: Map<string, AST.TypeExpression>,
): boolean {
  switch (pattern.type) {
    case "WildcardTypeExpression":
      return true;
    case "SimpleTypeExpression": {
      const patternName = ctx.getIdentifierName(pattern.name);
      if (!patternName) {
        return typeExpressionsEquivalent(ctx, pattern, target);
      }
      if (isPatternPlaceholderName(ctx, patternName, interfaceGenericNames)) {
        return bindPlaceholder(ctx, patternName, target, bindings);
      }
      if (target.type !== "SimpleTypeExpression") {
        return false;
      }
      const targetName = ctx.getIdentifierName(target.name);
      return !!targetName && targetName === patternName;
    }
    case "GenericTypeExpression": {
      if (patternAllowsBareConstructor(pattern)) {
        if (target.type !== "SimpleTypeExpression") {
          return false;
        }
        return matchSelfTypePattern(ctx, pattern.base, target, interfaceGenericNames, bindings);
      }
      if (target.type !== "GenericTypeExpression") {
        return false;
      }
      if (!matchSelfTypePattern(ctx, pattern.base, target.base, interfaceGenericNames, bindings)) {
        return false;
      }
      const patternArgs = Array.isArray(pattern.arguments) ? pattern.arguments : [];
      const targetArgs = Array.isArray(target.arguments) ? target.arguments : [];
      if (patternArgs.length !== targetArgs.length) {
        return false;
      }
      for (let index = 0; index < patternArgs.length; index += 1) {
        const expectedArg = patternArgs[index];
        const actualArg = targetArgs[index];
        if (!expectedArg || !actualArg) {
          if (!expectedArg && !actualArg) {
            continue;
          }
          return false;
        }
        if (isWildcardTypeExpression(expectedArg)) {
          continue;
        }
        if (!matchSelfTypePattern(ctx, expectedArg, actualArg, interfaceGenericNames, bindings)) {
          return false;
        }
      }
      return true;
    }
    default:
      return typeExpressionsEquivalent(ctx, pattern, target);
  }
}

function bindPlaceholder(
  ctx: ImplementationContext,
  name: string,
  target: AST.TypeExpression,
  bindings: Map<string, AST.TypeExpression>,
): boolean {
  if (!bindings.has(name)) {
    bindings.set(name, target);
    return true;
  }
  const existing = bindings.get(name);
  if (!existing) {
    bindings.set(name, target);
    return true;
  }
  return typeExpressionsEquivalent(ctx, existing, target);
}

function patternAllowsBareConstructor(pattern: AST.GenericTypeExpression): boolean {
  if (!Array.isArray(pattern.arguments)) {
    return false;
  }
  return pattern.arguments.some((arg) => isWildcardTypeExpression(arg));
}

function isWildcardTypeExpression(expr: AST.TypeExpression | null | undefined): boolean {
  return expr?.type === "WildcardTypeExpression";
}

function isPatternPlaceholderName(
  ctx: ImplementationContext,
  name: string,
  interfaceGenericNames: Set<string>,
): boolean {
  if (!name) {
    return false;
  }
  if (name === "Self") {
    return true;
  }
  if (interfaceGenericNames.has(name)) {
    return true;
  }
  if (PRIMITIVE_TYPE_NAMES.has(name)) {
    return false;
  }
  if (ctx.getStructDefinition(name)) {
    return false;
  }
  if (ctx.hasInterfaceDefinition(name)) {
    return false;
  }
  return true;
}

function targetsBareTypeConstructor(
  ctx: ImplementationContext,
  target: AST.TypeExpression,
  implementationGenericNames: Set<string>,
): boolean {
  switch (target.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(target.name);
      if (!name) {
        return false;
      }
      if (implementationGenericNames.has(name)) {
        return false;
      }
      if (PRIMITIVE_TYPE_NAMES.has(name)) {
        return false;
      }
      const structDefinition = ctx.getStructDefinition(name);
      return !!structDefinition && Array.isArray(structDefinition.genericParams) && structDefinition.genericParams.length > 0;
    }
    case "GenericTypeExpression": {
      if (!Array.isArray(target.arguments)) {
        return false;
      }
      return target.arguments.some((arg) => isWildcardTypeExpression(arg));
    }
    default:
      return false;
  }
}

function ensureImplementationMethods(
  ctx: ImplementationContext,
  implementation: AST.ImplementationDefinition,
  interfaceDefinition: AST.InterfaceDefinition,
  targetLabel: string,
  interfaceName: string,
): boolean {
  const provided = new Map<string, AST.FunctionDefinition>();
  if (Array.isArray(implementation.definitions)) {
    for (const fn of implementation.definitions) {
      if (!fn || fn.type !== "FunctionDefinition") continue;
      const methodName = fn.id?.name;
      if (!methodName) continue;
      if (provided.has(methodName)) {
        const label = ctx.formatImplementationLabel(interfaceName, targetLabel);
        ctx.report(`typechecker: ${label} defines duplicate method '${methodName}'`, fn);
        continue;
      }
      provided.set(methodName, fn);
    }
  }

  const signatures = Array.isArray(interfaceDefinition.signatures) ? interfaceDefinition.signatures : [];
  if (signatures.length === 0) {
    return true;
  }

  const label = ctx.formatImplementationLabel(interfaceName, targetLabel);
  let allRequiredPresent = true;

  for (const signature of signatures) {
    if (!signature) continue;
    const methodName = ctx.getIdentifierName(signature.name);
    if (!methodName) continue;
    if (!provided.has(methodName)) {
      if (signature.defaultImpl) {
        continue;
      }
      ctx.report(`typechecker: ${label} missing method '${methodName}'`, implementation);
      allRequiredPresent = false;
      continue;
    }
    const method = provided.get(methodName);
    if (method) {
      const methodValid = validateImplementationMethod(
        ctx,
        interfaceDefinition,
        implementation,
        signature,
        method,
        label,
        targetLabel,
      );
      if (!methodValid) {
        allRequiredPresent = false;
      }
      provided.delete(methodName);
    }
  }

  return allRequiredPresent;
}

function validateImplementationMethod(
  ctx: ImplementationContext,
  interfaceDefinition: AST.InterfaceDefinition,
  implementation: AST.ImplementationDefinition,
  signature: AST.FunctionSignature,
  method: AST.FunctionDefinition,
  label: string,
  targetLabel: string,
): boolean {
  let valid = true;
  const interfaceDefinitionGenerics = Array.isArray(interfaceDefinition.genericParams)
    ? interfaceDefinition.genericParams.length
    : 0;
  const interfaceGenerics = Array.isArray(signature.genericParams) ? signature.genericParams.length : 0;
  const implementationGenerics = Array.isArray(method.genericParams) ? method.genericParams.length : 0;
  const substitutions = buildImplementationSubstitutions(ctx, interfaceDefinition, implementation, targetLabel);
  const expectedGenerics = interfaceGenerics;
  if (expectedGenerics !== implementationGenerics) {
    ctx.report(
      `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${expectedGenerics} generic parameter(s), got ${implementationGenerics}`,
      method,
    );
    valid = false;
  }
  const interfaceParams = Array.isArray(signature.params) ? signature.params : [];
  const implementationParams = Array.isArray(method.params) ? method.params : [];
  if (interfaceParams.length !== implementationParams.length) {
    ctx.report(
      `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceParams.length} parameter(s), got ${implementationParams.length}`,
      method,
    );
    valid = false;
  } else {
    for (let index = 0; index < interfaceParams.length; index += 1) {
      const interfaceParam = interfaceParams[index];
      const implementationParam = implementationParams[index];
      if (!interfaceParam || !implementationParam) continue;
      if (
        index === 0 &&
        isImplicitSelfParameter(implementationParam) &&
        expectsSelfType(interfaceParam.paramType)
      ) {
        implementationParam.paramType = AST.simpleTypeExpression("Self");
      }
      const expectedDescription = ctx.describeTypeExpression(interfaceParam.paramType, substitutions);
      const actualDescription = ctx.describeTypeExpression(implementationParam.paramType, substitutions);
      if (!typeExpressionsEquivalent(ctx, interfaceParam.paramType, implementationParam.paramType, substitutions)) {
        ctx.report(
          `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' parameter ${index + 1} expected ${expectedDescription}, got ${actualDescription}`,
          implementation,
        );
        valid = false;
      }
    }
  }

  const returnExpected = ctx.describeTypeExpression(signature.returnType, substitutions);
  const returnActual = ctx.describeTypeExpression(method.returnType, substitutions);
  if (!typeExpressionsEquivalent(ctx, signature.returnType, method.returnType, substitutions)) {
    ctx.report(
      `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' return type expected ${returnExpected}, got ${returnActual}`,
      implementation,
    );
    valid = false;
  }

  const interfaceWhere = Array.isArray(signature.whereClause) ? signature.whereClause.length : 0;
  const implementationWhere = Array.isArray(implementation.whereClause) ? implementation.whereClause.length : 0;
  if (interfaceWhere !== implementationWhere) {
    ctx.report(
      `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceWhere} where-clause constraint(s), got ${implementationWhere}`,
      implementation,
    );
    valid = false;
  }

  if (implementation.isPrivate) {
    ctx.report(
      `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' must be public to satisfy interface`,
      implementation,
    );
    valid = false;
  }

  return valid;
}

function buildImplementationSubstitutions(
  ctx: ImplementationContext,
  interfaceDefinition: AST.InterfaceDefinition,
  implementation: AST.ImplementationDefinition,
  targetLabel: string,
): Map<string, string> {
  const substitutions = new Map<string, string>();
  substitutions.set("Self", targetLabel);
  const interfaceArgs = Array.isArray(implementation.interfaceArgs) ? implementation.interfaceArgs : [];
  const interfaceParams = Array.isArray(interfaceDefinition.genericParams) ? interfaceDefinition.genericParams : [];
  interfaceParams.forEach((param, index) => {
    const paramName = ctx.getIdentifierName(param?.name);
    if (!paramName) {
      return;
    }
    const argument = interfaceArgs[index];
    if (!argument) {
      return;
    }
    substitutions.set(paramName, ctx.formatTypeExpression(argument, substitutions));
  });
  return substitutions;
}

export function typeExpressionsEquivalent(
  ctx: ImplementationContext,
  a: AST.TypeExpression | null | undefined,
  b: AST.TypeExpression | null | undefined,
  substitutions?: Map<string, string>,
): boolean {
  if (!a && !b) return true;
  if (!a || !b) return false;
  return ctx.formatTypeExpression(a, substitutions) === ctx.formatTypeExpression(b, substitutions);
}

export function isImplicitSelfParameter(param: AST.FunctionParameter | null | undefined): boolean {
  if (!param || param.paramType) return false;
  if (param.name?.type !== "Identifier") return false;
  return param.name.name?.toLowerCase() === "self";
}

export function expectsSelfType(expr: AST.TypeExpression | null | undefined): boolean {
  return expr?.type === "SimpleTypeExpression" && expr.name?.name === "Self";
}
