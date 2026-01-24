import * as AST from "../../ast";
import { applyGenericTypeExpression } from "./type-expression-utils";
import { collectTargetTypeParams } from "./implementation-collection-helpers";
import type { ImplementationContext } from "./implementation-context";

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
  "IoHandle",
  "ProcHandle",
  "char",
  "nil",
  "void",
]);

const BUILTIN_TYPE_ARITY = new Map<string, number>([
  ["Array", 1],
  ["Iterator", 1],
  ["Range", 1],
  ["Future", 1],
  ["Map", 2],
  ["HashMap", 2],
  ["Channel", 1],
  ["Mutex", 0],
]);

export function validateImplementationSelfTypePattern(
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
    if (
      selfPattern.type === "GenericTypeExpression" &&
      patternAllowsBareConstructor(selfPattern) &&
      !isTypeConstructorTarget(ctx, subject, implementationGenericNames)
    ) {
      const expected = ctx.formatTypeExpression(selfPattern);
      ctx.report(
        `typechecker: impl ${interfaceName} for ${targetLabel} must match interface self type '${expected}'`,
        implementation,
      );
      return false;
    }
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

export function ensureImplementationMethods(
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

  const signatures = collectInterfaceSignatures(ctx, interfaceDefinition);
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

export function collectInterfaceGenericParamNames(
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

export function containsWildcardTypeExpression(expr: AST.TypeExpression | null | undefined): boolean {
  if (!expr) return false;
  if (isWildcardTypeExpression(expr)) return true;
  switch (expr.type) {
    case "GenericTypeExpression":
      return (
        containsWildcardTypeExpression(expr.base) ||
        (expr.arguments ?? []).some((arg) => containsWildcardTypeExpression(arg))
      );
    case "NullableTypeExpression":
    case "ResultTypeExpression":
      return containsWildcardTypeExpression(expr.innerType);
    case "UnionTypeExpression":
      return (expr.members ?? []).some((member) => containsWildcardTypeExpression(member));
    case "FunctionTypeExpression":
      return (
        (expr.paramTypes ?? []).some((param) => containsWildcardTypeExpression(param)) ||
        containsWildcardTypeExpression(expr.returnType)
      );
    default:
      return false;
  }
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
      if (patternName === "_") {
        return true;
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
      const allowsBare = patternAllowsBareConstructor(pattern);
      if (allowsBare && target.type === "SimpleTypeExpression") {
        return matchSelfTypePattern(ctx, pattern.base, target, interfaceGenericNames, bindings);
      }
      if (target.type !== "GenericTypeExpression") {
        return false;
      }
      if (allowsBare) {
        if (pattern.base.type === "SimpleTypeExpression") {
          const baseName = ctx.getIdentifierName(pattern.base.name);
          if (baseName && isPatternPlaceholderName(ctx, baseName, interfaceGenericNames)) {
            if (!bindPlaceholder(ctx, baseName, target, bindings)) {
              return false;
            }
          } else if (!matchSelfTypePattern(ctx, pattern.base, target.base, interfaceGenericNames, bindings)) {
            return false;
          }
        } else if (!matchSelfTypePattern(ctx, pattern.base, target.base, interfaceGenericNames, bindings)) {
          return false;
        }
      } else if (!matchSelfTypePattern(ctx, pattern.base, target.base, interfaceGenericNames, bindings)) {
        return false;
      }
      const patternArgs = Array.isArray(pattern.arguments) ? pattern.arguments : [];
      const targetArgs = Array.isArray(target.arguments) ? target.arguments : [];
      if (!selfPatternArgsCompatible(patternArgs, targetArgs)) {
        return false;
      }
      const sharedLength = Math.min(patternArgs.length, targetArgs.length);
      for (let index = 0; index < sharedLength; index += 1) {
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
  if (!expr) return false;
  if (expr.type === "WildcardTypeExpression") return true;
  return expr.type === "SimpleTypeExpression" && expr.name?.name === "_";
}

function selfPatternArgsCompatible(
  patternArgs: AST.TypeExpression[],
  targetArgs: AST.TypeExpression[],
): boolean {
  if (patternArgs.length === targetArgs.length) {
    return true;
  }
  if (patternArgs.length > targetArgs.length) {
    return patternArgs.slice(targetArgs.length).every((arg) => isWildcardTypeExpression(arg));
  }
  return targetArgs.slice(patternArgs.length).every((arg) => isWildcardTypeExpression(arg));
}

function isPatternPlaceholderName(
  ctx: ImplementationContext,
  name: string,
  interfaceGenericNames: Set<string>,
): boolean {
  if (!name) {
    return false;
  }
  if (name === "_") {
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
  return isTypeConstructorTarget(ctx, target, implementationGenericNames);
}

function isTypeConstructorTarget(
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
      const expected = expectedTypeArgumentCount(ctx, name);
      return expected !== null && expected > 0;
    }
    case "GenericTypeExpression": {
      if (containsWildcardTypeExpression(target)) {
        return true;
      }
      const baseName = ctx.getIdentifierNameFromTypeExpression(target.base);
      if (!baseName || implementationGenericNames.has(baseName)) {
        return false;
      }
      const expected = expectedTypeArgumentCount(ctx, baseName);
      if (expected === null) {
        return false;
      }
      const provided = Array.isArray(target.arguments) ? target.arguments.length : 0;
      return provided < expected;
    }
    default:
      return false;
  }
}

function expectedTypeArgumentCount(ctx: ImplementationContext, name: string): number | null {
  const alias = ctx.getTypeAlias?.(name);
  if (alias) {
    return Array.isArray(alias.genericParams) ? alias.genericParams.length : 0;
  }
  const structDef = ctx.getStructDefinition(name);
  if (structDef) {
    return Array.isArray(structDef.genericParams) ? structDef.genericParams.length : 0;
  }
  const unionDef = ctx.getUnionDefinition?.(name);
  if (unionDef) {
    return Array.isArray(unionDef.genericParams) ? unionDef.genericParams.length : 0;
  }
  const ifaceDef = ctx.getInterfaceDefinition(name);
  if (ifaceDef) {
    return Array.isArray(ifaceDef.genericParams) ? ifaceDef.genericParams.length : 0;
  }
  const builtinArity = BUILTIN_TYPE_ARITY.get(name);
  if (builtinArity !== undefined) {
    return builtinArity;
  }
  if (PRIMITIVE_TYPE_NAMES.has(name)) {
    return 0;
  }
  return null;
}

function collectInterfaceSignatures(
  ctx: ImplementationContext,
  interfaceDefinition: AST.InterfaceDefinition,
): AST.FunctionSignature[] {
  const signatures: AST.FunctionSignature[] = [];
  const seenMethods = new Set<string>();
  const seenInterfaces = new Set<string>();

  const addSignature = (sig: AST.FunctionSignature | null | undefined): void => {
    if (!sig) return;
    const name = ctx.getIdentifierName(sig.name);
    if (!name || seenMethods.has(name)) return;
    seenMethods.add(name);
    signatures.push(sig);
  };

  const substituteTypeExpression = (
    expr: AST.TypeExpression | null | undefined,
    substitutions: Map<string, AST.TypeExpression>,
  ): AST.TypeExpression | null | undefined => {
    if (!expr) return expr;
    switch (expr.type) {
      case "SimpleTypeExpression": {
        const name = ctx.getIdentifierName(expr.name);
        if (name && substitutions.has(name)) {
          return substitutions.get(name);
        }
        return expr;
      }
      case "GenericTypeExpression":
        return {
          ...expr,
          base: substituteTypeExpression(expr.base, substitutions) ?? expr.base,
          arguments: (expr.arguments ?? []).map((arg) => substituteTypeExpression(arg, substitutions) ?? arg),
        };
      case "NullableTypeExpression":
        return { ...expr, innerType: substituteTypeExpression(expr.innerType, substitutions) ?? expr.innerType };
      case "ResultTypeExpression":
        return { ...expr, innerType: substituteTypeExpression(expr.innerType, substitutions) ?? expr.innerType };
      case "UnionTypeExpression":
        return {
          ...expr,
          members: (expr.members ?? []).map((member) => substituteTypeExpression(member, substitutions) ?? member),
        };
      case "FunctionTypeExpression":
        return {
          ...expr,
          paramTypes: (expr.paramTypes ?? []).map((param) => substituteTypeExpression(param, substitutions) ?? param),
          returnType: substituteTypeExpression(expr.returnType, substitutions) ?? expr.returnType,
        };
      default:
        return expr;
    }
  };

  const substituteSignature = (
    sig: AST.FunctionSignature,
    substitutions: Map<string, AST.TypeExpression>,
  ): AST.FunctionSignature => {
    const params = Array.isArray(sig.params)
      ? sig.params.map((param) =>
          param && param.paramType
            ? { ...param, paramType: substituteTypeExpression(param.paramType, substitutions) ?? param.paramType }
            : param,
        )
      : sig.params;
    const returnType = substituteTypeExpression(sig.returnType, substitutions) ?? sig.returnType;
    const whereClause = Array.isArray(sig.whereClause)
      ? sig.whereClause.map((clause) => {
          if (!clause) return clause;
          return {
            ...clause,
            typeParam: substituteTypeExpression(clause.typeParam, substitutions) ?? clause.typeParam,
            constraints: Array.isArray(clause.constraints)
              ? clause.constraints.map((constraint) =>
                  constraint
                    ? {
                        ...constraint,
                        interfaceType: substituteTypeExpression(constraint.interfaceType, substitutions) ?? constraint.interfaceType,
                      }
                    : constraint,
                )
              : clause.constraints,
          };
        })
      : sig.whereClause;
    return { ...sig, params, returnType, whereClause };
  };

  const walkInterface = (
    def: AST.InterfaceDefinition | null | undefined,
    substitutions: Map<string, AST.TypeExpression>,
  ): void => {
    if (!def) return;
    const defName = ctx.getIdentifierName(def.id);
    if (defName) {
      if (seenInterfaces.has(defName)) return;
      seenInterfaces.add(defName);
    }
    const defSignatures = Array.isArray(def.signatures) ? def.signatures : [];
    for (const sig of defSignatures) {
      addSignature(substituteSignature(sig, substitutions));
    }
    const bases = Array.isArray(def.baseInterfaces) ? def.baseInterfaces : [];
    for (const baseExpr of bases) {
      if (!baseExpr) continue;
      const substitutedBase = substituteTypeExpression(baseExpr, substitutions) ?? baseExpr;
      const baseName = ctx.getInterfaceNameFromTypeExpression(substitutedBase);
      if (!baseName) continue;
      const baseDef = ctx.getInterfaceDefinition(baseName);
      if (!baseDef) continue;
      const baseArgs = substitutedBase.type === "GenericTypeExpression" ? substitutedBase.arguments ?? [] : [];
      const baseSubstitutions = new Map<string, AST.TypeExpression>();
      const baseParams = Array.isArray(baseDef.genericParams) ? baseDef.genericParams : [];
      baseParams.forEach((param, index) => {
        const paramName = ctx.getIdentifierName(param?.name);
        if (!paramName) return;
        baseSubstitutions.set(paramName, baseArgs[index] ?? AST.wildcardTypeExpression());
      });
      walkInterface(baseDef, baseSubstitutions);
    }
  };

  walkInterface(interfaceDefinition, new Map());
  return signatures;
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
  const implementationWhere = Array.isArray(method.whereClause) ? method.whereClause.length : 0;
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
  const explicitGenerics = Array.isArray(implementation.genericParams) ? implementation.genericParams : [];
  const genericNames = new Set<string>();
  for (const param of explicitGenerics) {
    const name = ctx.getIdentifierName(param?.name);
    if (name) {
      genericNames.add(name);
    }
  }
  if (targetType) {
    collectTargetTypeParams(ctx, targetType).forEach((name) => {
      if (name) {
        genericNames.add(name);
      }
    });
  }
  for (const name of genericNames) {
    substitutions.set(name, name);
  }
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
  let selfExpr = targetType ?? null;
  if (
    selfExpr &&
    interfaceArgs.length > 0 &&
    targetsBareTypeConstructor(ctx, selfExpr, new Set(genericNames))
  ) {
    selfExpr = applyGenericTypeExpression(selfExpr, interfaceArgs);
  }
  const formattedSelf = selfExpr ? ctx.formatTypeExpression(selfExpr, substitutions) : null;
  substitutions.set("Self", formattedSelf ?? targetLabel);
  if (interfaceDefinition.selfTypePattern) {
    const interfaceGenericNames = collectInterfaceGenericParamNames(ctx, interfaceDefinition);
    const bindings = new Map<string, AST.TypeExpression>();
    if (
      targetType &&
      matchSelfTypePattern(ctx, interfaceDefinition.selfTypePattern, targetType, interfaceGenericNames, bindings)
    ) {
      for (const [name, expr] of bindings) {
        if (!name) {
          continue;
        }
        if (genericNames.has(name)) {
          continue;
        }
        const existing = substitutions.get(name);
        if (existing && existing !== name) {
          continue;
        }
        substitutions.set(name, ctx.formatTypeExpression(expr, substitutions));
      }
    }
  }
  return substitutions;
}
