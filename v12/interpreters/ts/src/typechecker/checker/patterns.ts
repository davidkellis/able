import * as AST from "../../ast";
import { arrayType, formatType, unknownType, type TypeInfo } from "../types";
import type { ExpressionContext } from "./expression-context";

export interface PatternBindingOptions {
  suppressMismatchReport?: boolean;
  declarationNames?: Set<string>;
  isDeclaration?: boolean;
  allowFallbackDeclaration?: boolean;
}

export function checkStructPattern(
  ctx: ExpressionContext,
  pattern: AST.StructPattern,
  valueType: TypeInfo,
): void {
  const definition = ctx.resolveStructDefinitionForPattern(pattern, valueType);
  if (!definition) return;

  const knownFields = new Set<string>();
  if (Array.isArray(definition.fields)) {
    for (const field of definition.fields) {
      const fieldName = ctx.getIdentifierName(field?.name);
      if (fieldName) knownFields.add(fieldName);
    }
  }

  if (!Array.isArray(pattern.fields)) {
    return;
  }

  for (const field of pattern.fields) {
    if (!field) continue;
    const fieldName = ctx.getIdentifierName(field.fieldName);
    if (fieldName && !knownFields.has(fieldName)) {
      ctx.report(`typechecker: struct pattern field '${fieldName}' not found`, field ?? pattern);
    }
  }
}

export function bindPatternToEnv(
  ctx: ExpressionContext,
  pattern: AST.Pattern | undefined | null,
  valueType: TypeInfo,
  contextLabel: string,
  options?: PatternBindingOptions,
): void {
  if (!pattern) return;
  const declarationNames = options?.declarationNames;
  const isDeclaration = options?.isDeclaration !== false;
  const allowFallback = !!options?.allowFallbackDeclaration;
  const bindIdentifier = (name: string | undefined | null, typeInfo: TypeInfo): void => {
    if (!name) return;
    const resolved = typeInfo ?? unknownType;
    if (isDeclaration) {
      if (shouldDeclareIdentifier(name, declarationNames)) {
        ctx.defineValue(name, resolved);
      } else {
        ctx.assignValue(name, resolved);
      }
      return;
    }
    const assigned = ctx.assignValue(name, resolved);
    if (!assigned && allowFallback) {
      ctx.defineValue(name, resolved);
    }
  };
  switch (pattern.type) {
    case "Identifier":
      bindIdentifier(pattern.name, valueType ?? unknownType);
      return;
    case "WildcardPattern":
      return;
    case "TypedPattern": {
      const annotationType = ctx.resolveTypeExpression(pattern.typeAnnotation);
      const resolvedType =
        annotationType && annotationType.kind !== "unknown" ? annotationType : valueType ?? unknownType;
      const literalMessage = ctx.describeLiteralMismatch(valueType, annotationType);
      if (literalMessage) {
        ctx.report(literalMessage, pattern);
      } else if (
        !options?.suppressMismatchReport &&
        annotationType &&
        annotationType.kind !== "unknown" &&
        valueType &&
        valueType.kind !== "unknown" &&
        !patternAcceptsType(ctx, annotationType, valueType)
      ) {
        const expectedLabel = ctx.describeTypeExpression(pattern.typeAnnotation);
        const actualLabel = formatType(valueType);
        ctx.report(`typechecker: ${contextLabel} expects type ${expectedLabel}, got ${actualLabel}`, pattern);
      }
      if (pattern.pattern) {
        bindPatternToEnv(ctx, pattern.pattern as AST.Pattern, resolvedType ?? valueType, contextLabel, options);
      }
      return;
    }
    case "StructPattern":
      checkStructPattern(ctx, pattern, valueType);
      bindStructPatternFields(ctx, pattern, valueType, contextLabel, options);
      return;
    case "ArrayPattern":
      bindArrayPatternElements(ctx, pattern, valueType, contextLabel, options);
      return;
    case "LiteralPattern":
    default:
      return;
  }
}

function bindStructPatternFields(
  ctx: ExpressionContext,
  pattern: AST.StructPattern,
  valueType: TypeInfo,
  contextLabel: string,
  options?: PatternBindingOptions,
): void {
  if (!Array.isArray(pattern.fields) || pattern.fields.length === 0) {
    return;
  }
  const fieldTypes = new Map<string, TypeInfo>();
  const definition = ctx.resolveStructDefinitionForPattern(pattern, valueType);
  if (definition && Array.isArray(definition.fields)) {
    for (const fieldDef of definition.fields) {
      if (!fieldDef) continue;
      const name = ctx.getIdentifierName(fieldDef.name);
      if (!name) continue;
      const resolved = fieldDef.fieldType ? ctx.resolveTypeExpression(fieldDef.fieldType) : undefined;
      if (resolved) {
        fieldTypes.set(name, resolved);
      }
    }
  }
  for (const field of pattern.fields) {
    if (!field) continue;
    const fieldName = ctx.getIdentifierName(field.fieldName);
    const annotatedType = field.typeAnnotation ? ctx.resolveTypeExpression(field.typeAnnotation) : undefined;
    const nestedType = annotatedType ?? ((fieldName && fieldTypes.get(fieldName)) ?? unknownType);
    if (field.pattern) {
      bindPatternToEnv(ctx, field.pattern as AST.Pattern, nestedType ?? unknownType, contextLabel, options);
    }
    if (field.binding?.name) {
      const bindingPattern = annotatedType ? AST.typedPattern(field.binding as AST.Pattern, annotatedType) : (field.binding as AST.Pattern);
      bindPatternToEnv(ctx, bindingPattern, nestedType ?? unknownType, contextLabel, options);
    }
  }
}

function bindArrayPatternElements(
  ctx: ExpressionContext,
  pattern: AST.ArrayPattern,
  valueType: TypeInfo,
  contextLabel: string,
  options?: PatternBindingOptions,
): void {
  const elementType =
    valueType && valueType.kind === "array"
      ? valueType.element ?? unknownType
      : unknownType;
  if (Array.isArray(pattern.elements)) {
    for (const element of pattern.elements) {
      if (!element) continue;
      bindPatternToEnv(ctx, element as AST.Pattern, elementType ?? unknownType, contextLabel, options);
    }
  }
  const rest = pattern.restPattern;
  if (rest && rest.type === "Identifier" && rest.name) {
    bindPatternToEnv(ctx, rest, arrayType(elementType ?? unknownType), contextLabel, options);
  }
}

function shouldDeclareIdentifier(
  name: string | null | undefined,
  declarationNames: Set<string> | undefined,
): boolean {
  if (!name) return false;
  if (!declarationNames) return true;
  return declarationNames.has(name);
}

function patternAcceptsType(ctx: ExpressionContext, expected: TypeInfo, actual: TypeInfo): boolean {
  if (ctx.isTypeAssignable(actual, expected)) {
    return true;
  }
  if (actual.kind === "union") {
    return actual.members.some((member) => patternAcceptsType(ctx, expected, member));
  }
  if (actual.kind === "result") {
    if (expected.kind === "interface" && expected.name === "Error") {
      return true;
    }
    if (ctx.typeImplementsInterface?.(expected, "Error")?.ok) {
      return true;
    }
    const members: TypeInfo[] = [];
    members.push({
      kind: "interface",
      name: "Error",
      typeArguments: [],
      definition: ctx.getInterfaceDefinition("Error"),
    });
    if (actual.inner) {
      members.push(actual.inner);
    }
    return members.some((member) => patternAcceptsType(ctx, expected, member));
  }
  return false;
}
