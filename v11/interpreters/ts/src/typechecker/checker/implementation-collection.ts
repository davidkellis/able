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

function typeInfoToTypeExpression(type: TypeInfo | undefined): AST.TypeExpression | null {
  if (!type) return null;
  switch (type.kind) {
    case "primitive":
      return AST.simpleTypeExpression(type.name);
    case "struct": {
      const base = AST.simpleTypeExpression(type.name);
      const args =
        Array.isArray(type.typeArguments) && type.typeArguments.length > 0
          ? type.typeArguments.map((arg) => typeInfoToTypeExpression(arg) ?? AST.wildcardTypeExpression())
          : undefined;
      return args && args.length > 0 ? AST.genericTypeExpression(base, args) : base;
    }
    case "interface": {
      const base = AST.simpleTypeExpression(type.name);
      const args =
        Array.isArray(type.typeArguments) && type.typeArguments.length > 0
          ? type.typeArguments.map((arg) => typeInfoToTypeExpression(arg) ?? AST.wildcardTypeExpression())
          : undefined;
      return args && args.length > 0 ? AST.genericTypeExpression(base, args) : base;
    }
    case "array":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Array"),
        [typeInfoToTypeExpression(type.element) ?? AST.wildcardTypeExpression()],
      );
    case "map":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Map"),
        [
          typeInfoToTypeExpression(type.key) ?? AST.wildcardTypeExpression(),
          typeInfoToTypeExpression(type.value) ?? AST.wildcardTypeExpression(),
        ],
      );
    case "range":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Range"),
        [typeInfoToTypeExpression(type.element) ?? AST.wildcardTypeExpression()],
      );
    case "iterator":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Iterator"),
        [typeInfoToTypeExpression(type.element) ?? AST.wildcardTypeExpression()],
      );
    case "proc":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Proc"),
        [typeInfoToTypeExpression(type.result) ?? AST.wildcardTypeExpression()],
      );
    case "future":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Future"),
        [typeInfoToTypeExpression(type.result) ?? AST.wildcardTypeExpression()],
      );
    case "nullable":
      return AST.nullableTypeExpression(typeInfoToTypeExpression(type.inner) ?? AST.wildcardTypeExpression());
    case "result":
      return AST.resultTypeExpression(typeInfoToTypeExpression(type.inner) ?? AST.wildcardTypeExpression());
    case "union":
      return AST.unionTypeExpression(
        type.members.map((member) => typeInfoToTypeExpression(member) ?? AST.wildcardTypeExpression()),
      );
    case "function": {
      const params = (type.parameters ?? []).map((param) => typeInfoToTypeExpression(param) ?? AST.wildcardTypeExpression());
      const returnType = typeInfoToTypeExpression(type.returnType) ?? AST.wildcardTypeExpression();
      return AST.functionTypeExpression(params, returnType);
    }
    default:
      return AST.wildcardTypeExpression();
  }
}

function canonicalizeTargetType(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
): AST.TypeExpression | null | undefined {
  const expanded = expandTypeAliases(ctx, expr);
  switch (expanded?.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expanded.name);
      if (name) {
        const binding = ctx.lookupIdentifier?.(name);
        const canonical = binding ? typeInfoToTypeExpression(binding) : null;
        if (canonical) {
          return canonical;
        }
      }
      return expanded;
    }
    case "GenericTypeExpression":
      return {
        ...expanded,
        base: canonicalizeTargetType(ctx, expanded.base) ?? expanded.base,
        arguments: (expanded.arguments ?? []).map((arg) => canonicalizeTargetType(ctx, arg) ?? arg),
      };
    case "NullableTypeExpression":
      return { ...expanded, innerType: canonicalizeTargetType(ctx, expanded.innerType) ?? expanded.innerType };
    case "ResultTypeExpression":
      return { ...expanded, innerType: canonicalizeTargetType(ctx, expanded.innerType) ?? expanded.innerType };
    case "UnionTypeExpression":
      return { ...expanded, members: (expanded.members ?? []).map((member) => canonicalizeTargetType(ctx, member) ?? member) };
    case "FunctionTypeExpression":
      return {
        ...expanded,
        paramTypes: (expanded.paramTypes ?? []).map((param) => canonicalizeTargetType(ctx, param) ?? param),
        returnType: canonicalizeTargetType(ctx, expanded.returnType) ?? expanded.returnType,
      };
    default:
      return expanded;
  }
}

function substituteTypeExpression(
  ctx: ImplementationContext,
  expr: AST.TypeExpression,
  substitutions: Map<string, AST.TypeExpression>,
  seen: Set<string>,
): AST.TypeExpression {
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expr.name);
      if (name && substitutions.has(name)) {
        return expandTypeAliases(ctx, substitutions.get(name), seen) ?? substitutions.get(name)!;
      }
      return expr;
    }
    case "GenericTypeExpression": {
      const base = expandTypeAliases(ctx, expr.base, seen) ?? expr.base;
      const args = (expr.arguments ?? []).map((arg) => {
        const next = arg ? substituteTypeExpression(ctx, arg, substitutions, seen) : undefined;
        return expandTypeAliases(ctx, next, seen) ?? next ?? arg;
      });
      return { ...expr, base, arguments: args };
    }
    case "NullableTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, substituteTypeExpression(ctx, expr.innerType, substitutions, seen), seen) };
    case "ResultTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, substituteTypeExpression(ctx, expr.innerType, substitutions, seen), seen) };
    case "UnionTypeExpression":
      return {
        ...expr,
        members: (expr.members ?? []).map((member) =>
          expandTypeAliases(ctx, substituteTypeExpression(ctx, member, substitutions, seen), seen),
        ),
      };
    case "FunctionTypeExpression":
      return {
        ...expr,
        paramTypes: (expr.paramTypes ?? []).map((param) =>
          expandTypeAliases(ctx, substituteTypeExpression(ctx, param, substitutions, seen), seen),
        ),
        returnType: expandTypeAliases(ctx, substituteTypeExpression(ctx, expr.returnType, substitutions, seen), seen) ?? expr.returnType,
      };
    default:
      return expr;
  }
}

function expandTypeAliases(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
  seen: Set<string> = new Set(),
): AST.TypeExpression | null | undefined {
  if (!expr) return expr;
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = ctx.getIdentifierName(expr.name);
      if (!name || !ctx.getTypeAlias) return expr;
      if (seen.has(name)) return expr;
      const alias = ctx.getTypeAlias(name);
      if (!alias?.targetType) return expr;
      seen.add(name);
      const expanded = expandTypeAliases(ctx, alias.targetType, seen);
      seen.delete(name);
      return expanded ?? expr;
    }
    case "GenericTypeExpression": {
      const baseName = ctx.getIdentifierNameFromTypeExpression(expr.base);
      const expandedBase = expandTypeAliases(ctx, expr.base, seen) ?? expr.base;
      const expandedArgs = (expr.arguments ?? []).map((arg) => expandTypeAliases(ctx, arg, seen) ?? arg);
      if (!baseName || !ctx.getTypeAlias || seen.has(baseName)) {
        return { ...expr, base: expandedBase, arguments: expandedArgs };
      }
      const alias = ctx.getTypeAlias(baseName);
      if (!alias?.targetType) {
        return { ...expr, base: expandedBase, arguments: expandedArgs };
      }
      const substitutions = new Map<string, AST.TypeExpression>();
      (alias.genericParams ?? []).forEach((param, index) => {
        const paramName = ctx.getIdentifierName(param?.name);
        if (!paramName) return;
        substitutions.set(paramName, expandedArgs[index] ?? AST.wildcardTypeExpression());
      });
      seen.add(baseName);
      const substituted = substituteTypeExpression(ctx, alias.targetType, substitutions, seen);
      const expanded = expandTypeAliases(ctx, substituted, seen);
      seen.delete(baseName);
      return expanded ?? substituted;
    }
    case "NullableTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, expr.innerType, seen) ?? expr.innerType };
    case "ResultTypeExpression":
      return { ...expr, innerType: expandTypeAliases(ctx, expr.innerType, seen) ?? expr.innerType };
    case "UnionTypeExpression":
      return { ...expr, members: (expr.members ?? []).map((member) => expandTypeAliases(ctx, member, seen) ?? member) };
    case "FunctionTypeExpression":
      return {
        ...expr,
        paramTypes: (expr.paramTypes ?? []).map((param) => expandTypeAliases(ctx, param, seen) ?? param),
        returnType: expandTypeAliases(ctx, expr.returnType, seen) ?? expr.returnType,
      };
    default:
      return expr;
  }
}

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
  const selfType = ctx.resolveTypeExpression(targetType, substitutionMap);
  const canonicalTarget = typeInfoToTypeExpression(selfType) ?? targetType ?? definition.targetType;
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
  validateImplementationInterfaceArguments(ctx, definition, interfaceDefinition, contextName, interfaceName);
  const interfaceGenericNames = collectInterfaceGenericParamNames(ctx, interfaceDefinition);
  const implementationGenericNames = collectImplementationGenericParamNames(ctx, definition);
  const implementationGenericNameSet = new Set(implementationGenericNames);
  const substitutionMap = new Map<string, TypeInfo>();
  implementationGenericNames.forEach((name) => substitutionMap.set(name, unknownType));
  const resolvedTarget =
    targetType ?? definition.targetType
      ? ctx.resolveTypeExpression(targetType ?? definition.targetType, substitutionMap)
      : unknownType;
  const resolvedTargetExpr = typeInfoToTypeExpression(resolvedTarget ?? unknownType);
  const canonicalTarget =
    resolvedTargetExpr && resolvedTargetExpr.type !== "WildcardTypeExpression"
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
  const interfaceArgs = Array.isArray(definition.interfaceArgs)
    ? definition.interfaceArgs.filter((arg): arg is AST.TypeExpression => Boolean(arg))
    : [];
  const unionVariants = collectUnionVariantLabels(ctx, resolvedTarget);
  return {
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
  targetType: AST.TypeExpression | null | undefined,
  implementation: AST.ImplementationDefinition,
  interfaceDefinition: AST.InterfaceDefinition,
  targetLabel: string,
  interfaceName: string,
  interfaceGenericNames: Set<string>,
  implementationGenericNames: Set<string>,
): boolean {
  const subject = targetType ?? implementation.targetType;
  if (!subject) {
    return false;
  }
  const selfPattern = interfaceDefinition.selfTypePattern;
  if (selfPattern) {
    const matches = doesSelfPatternMatchTarget(ctx, selfPattern, subject, interfaceGenericNames);
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
  if (targetsBareTypeConstructor(ctx, subject, implementationGenericNames)) {
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
  const substitutions = buildImplementationSubstitutions(
    ctx,
    interfaceDefinition,
    implementation,
    targetLabel,
    implementation.targetType ?? null,
  );
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
  targetType: AST.TypeExpression | null | undefined,
): Map<string, string> {
  const substitutions = new Map<string, string>();
  const implGenerics = Array.isArray(implementation.genericParams) ? implementation.genericParams : [];
  implGenerics.forEach((param) => {
    const name = ctx.getIdentifierName(param?.name);
    if (name) {
      substitutions.set(name, name);
    }
  });
  const formattedSelf = targetType ? ctx.formatTypeExpression(targetType, substitutions) : null;
  substitutions.set("Self", formattedSelf ?? targetLabel);
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
