import * as AST from "../../ast";
import { formatType, unknownType } from "../types";
import type { TypeInfo } from "../types";
import type { DeclarationsContext } from "./declarations";
import { collectFunctionDefinition } from "./declarations";
import {
  buildConstraintKeySet,
  collectUnionVariantLabels,
  computeImplementationSpecificity,
  selectMostSpecificImplementationMatch,
  type ImplementationMatch,
} from "./impl_matches";
import { methodSetProvidesInterface } from "./method_sets";
import type {
  FunctionContext,
  FunctionInfo,
  ImplementationObligation,
  ImplementationRecord,
  InterfaceCheckResult,
  MethodSetRecord,
} from "./types";

export interface ImplementationContext extends DeclarationsContext {
  formatImplementationTarget(expr: AST.TypeExpression | null | undefined): string | null;
  formatImplementationLabel(interfaceName: string, targetLabel: string): string;
  registerMethodSet(record: MethodSetRecord): void;
  getMethodSets(): Iterable<MethodSetRecord>;
  registerImplementationRecord(record: ImplementationRecord): void;
  getImplementationRecords(): Iterable<ImplementationRecord>;
  getImplementationBucket(key: string): ImplementationRecord[] | undefined;
  describeTypeArgument(type: TypeInfo): string;
  appendInterfaceArgsToLabel(label: string, args: string[]): string;
  formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string;
}

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
  "string",
  "char",
  "nil",
  "void",
]);


export function collectMethodsDefinition(ctx: ImplementationContext, definition: AST.MethodsDefinition): void {
  const structLabel =
    ctx.formatImplementationTarget(definition.targetType) ?? ctx.getIdentifierNameFromTypeExpression(definition.targetType);
  if (!structLabel) return;
  const record: MethodSetRecord = {
    label: `methods for ${structLabel}`,
    target: definition.targetType,
    genericParams: Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((name): name is string => Boolean(name))
      : [],
    obligations: extractMethodSetObligations(ctx, definition),
    definition,
  };
  ctx.registerMethodSet(record);
  if (Array.isArray(definition.definitions)) {
    for (const entry of definition.definitions) {
      if (entry?.type === "FunctionDefinition") {
        collectFunctionDefinition(ctx, entry, { structName: structLabel });
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
        collectFunctionDefinition(ctx, entry, { structName: contextName });
      }
    }
  }
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
    const methodGenericNames = Array.isArray(method.genericParams)
      ? method.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((name): name is string => Boolean(name))
      : [];
    const parameterTypes = Array.isArray(method.params)
      ? method.params.map((param) => ctx.resolveTypeExpression(param?.paramType))
      : [];
    const info: FunctionInfo = {
      name: methodName,
      fullName: `${record.label}::${methodName}`,
      structName: structLabel,
      parameters: parameterTypes,
      genericConstraints: [],
      genericParamNames: methodGenericNames,
      whereClause: record.obligations,
      methodSetSubstitutions: Array.from(substitutions.entries()),
      returnType: ctx.resolveTypeExpression(method.returnType),
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
  if (!name || name === "Self") {
    return false;
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

  for (const methodName of provided.keys()) {
    const extraMethod = provided.get(methodName);
    ctx.report(
      `typechecker: ${label} defines method '${methodName}' not declared in interface ${interfaceName}`,
      extraMethod ?? implementation,
    );
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
  const interfaceGenerics = Array.isArray(signature.genericParams) ? signature.genericParams.length : 0;
  const implementationGenerics = Array.isArray(method.genericParams) ? method.genericParams.length : 0;
  if (interfaceGenerics !== implementationGenerics) {
    ctx.report(
      `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceGenerics} generic parameter(s), got ${implementationGenerics}`,
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
    const substitutions = new Map<string, string>();
    substitutions.set("Self", targetLabel);
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

  const returnExpected = ctx.describeTypeExpression(signature.returnType, new Map([["Self", targetLabel]]));
  const returnActual = ctx.describeTypeExpression(method.returnType, new Map([["Self", targetLabel]]));
  if (!typeExpressionsEquivalent(ctx, signature.returnType, method.returnType, new Map([["Self", targetLabel]]))) {
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

function typeExpressionsEquivalent(
  ctx: ImplementationContext,
  a: AST.TypeExpression | null | undefined,
  b: AST.TypeExpression | null | undefined,
  substitutions?: Map<string, string>,
): boolean {
  if (!a && !b) return true;
  if (!a || !b) return false;
  return ctx.formatTypeExpression(a, substitutions) === ctx.formatTypeExpression(b, substitutions);
}

function isImplicitSelfParameter(param: AST.FunctionParameter | null | undefined): boolean {
  if (!param || param.paramType) return false;
  if (param.name?.type !== "Identifier") return false;
  return param.name.name?.toLowerCase() === "self";
}

function expectsSelfType(expr: AST.TypeExpression | null | undefined): boolean {
  return expr?.type === "SimpleTypeExpression" && expr.name?.name === "Self";
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
