import type * as AST from "../../ast";
import {
  arrayType,
  describe,
  formatType,
  futureType,
  iteratorType,
  isBoolean,
  isNumeric,
  primitiveType,
  procType,
  rangeType,
  unknownType,
  type PrimitiveName,
  type TypeInfo,
} from "../types";

export interface ExpressionContext {
  resolveStructDefinitionForPattern(
    pattern: AST.StructPattern,
    valueType: TypeInfo,
  ): AST.StructDefinition | undefined;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  report(message: string, node?: AST.Node | null | undefined): void;
  describeTypeExpression(expr: AST.TypeExpression | null | undefined): string | null;
  typeInfosEquivalent(a?: TypeInfo, b?: TypeInfo): boolean;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  resolveTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, TypeInfo>): TypeInfo;
  getStructDefinition(name: string): AST.StructDefinition | undefined;
  handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  checkStatement(node: AST.Statement | AST.Expression | undefined | null): void;
  pushAsyncContext(): void;
  popAsyncContext(): void;
  checkFunctionCall(call: AST.FunctionCall): void;
  inferFunctionCallReturnType(call: AST.FunctionCall): TypeInfo;
  checkFunctionDefinition(definition: AST.FunctionDefinition): void;
  checkReturnStatement(statement: AST.ReturnStatement): void;
  pushScope(): void;
  popScope(): void;
  withForkedEnv<T>(fn: () => T): T;
  lookupIdentifier(name: string): TypeInfo | undefined;
  defineValue(name: string, valueType: TypeInfo): void;
  assignValue(name: string, valueType: TypeInfo): boolean;
  hasBinding(name: string): boolean;
  hasBindingInCurrentScope(name: string): boolean;
  allowDynamicLookup(): boolean;
}

interface PatternBindingOptions {
  suppressMismatchReport?: boolean;
  declarationNames?: Set<string>;
  isDeclaration?: boolean;
  allowFallbackDeclaration?: boolean;
}

export type StatementContext = ExpressionContext & {
  isExpression(node: AST.Node | undefined | null): node is AST.Expression;
};

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
    const nestedType = (fieldName && fieldTypes.get(fieldName)) || unknownType;
    if (field.pattern) {
      bindPatternToEnv(ctx, field.pattern as AST.Pattern, nestedType ?? unknownType, contextLabel, options);
    }
    if (field.binding?.name) {
      bindPatternToEnv(ctx, field.binding as AST.Pattern, nestedType ?? unknownType, contextLabel, options);
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
    bindPatternToEnv(ctx, rest, elementType ?? unknownType, contextLabel, options);
  }
}
export function inferExpression(ctx: ExpressionContext, expression: AST.Expression | undefined | null): TypeInfo {
  if (!expression) return unknownType;
  switch (expression.type) {
    case "StringLiteral":
      return primitiveType("string");
    case "BooleanLiteral":
      return primitiveType("bool");
    case "IntegerLiteral": {
      const literalType = (expression.integerType as PrimitiveName) ?? "i32";
      const result = primitiveType(literalType);
      const rawValue = expression.value;
      const literalValue = typeof rawValue === "bigint" ? rawValue : BigInt(Math.trunc(rawValue ?? 0));
      result.literal = {
        literalKind: "integer",
        value: literalValue,
        explicit: Boolean(expression.integerType),
      };
      return result;
    }
    case "FloatLiteral": {
      const literalType = (expression.floatType as PrimitiveName) ?? "f64";
      const result = primitiveType(literalType);
      result.literal = {
        literalKind: "float",
        value: expression.value,
        explicit: Boolean(expression.floatType),
      };
      return result;
    }
    case "NilLiteral":
      return primitiveType("nil");
    case "Identifier": {
      const name = expression.name;
      const existing = ctx.lookupIdentifier(name);
      if (existing) {
        return existing;
      }
      const structDefinition = ctx.getStructDefinition(name);
      if (structDefinition) {
        return { kind: "struct", name, typeArguments: [], definition: structDefinition };
      }
      if (ctx.allowDynamicLookup()) {
        return unknownType;
      }
      ctx.report(`typechecker: undefined identifier '${name}'`, expression);
      return unknownType;
    }
    case "BinaryExpression": {
      const left = ctx.inferExpression(expression.left);
      const right = ctx.inferExpression(expression.right);
      if (expression.operator === "&&" || expression.operator === "||") {
        if (!isBoolean(left)) {
          ctx.report(
            `typechecker: '${expression.operator}' left operand must be bool (got ${describe(left)})`,
            expression,
          );
        }
        if (!isBoolean(right)) {
          ctx.report(
            `typechecker: '${expression.operator}' right operand must be bool (got ${describe(right)})`,
            expression,
          );
        }
        return primitiveType("bool");
      }
      const comparisonOperators = ["==", "!=", "<", "<=", ">", ">="];
      if (comparisonOperators.includes(expression.operator)) {
        return primitiveType("bool");
      }
      return unknownType;
    }
    case "RangeExpression": {
      const start = ctx.inferExpression(expression.start);
      if (!isNumeric(start)) {
        ctx.report("typechecker: range start must be numeric", expression);
      }
      const end = ctx.inferExpression(expression.end);
      if (!isNumeric(end)) {
        ctx.report("typechecker: range end must be numeric", expression);
      }
      const elementType = resolveRangeElementType(ctx, start, end);
      const bounds: TypeInfo[] = [];
      if (start && start.kind !== "unknown") {
        bounds.push(start);
      }
      if (end && end.kind !== "unknown") {
        bounds.push(end);
      }
      return rangeType(elementType, bounds.length > 0 ? bounds : undefined);
    }
    case "ArrayLiteral": {
      const elementType = resolveArrayLiteralElementType(ctx, expression.elements);
      return arrayType(elementType);
    }
    case "MapLiteral": {
      let keyType: TypeInfo = unknownType;
      let valueType: TypeInfo = unknownType;
      for (const entry of expression.entries ?? []) {
        if (!entry) continue;
        if (entry.type === "MapLiteralEntry") {
          const inferredKey = ctx.inferExpression(entry.key);
          keyType = mergeMapComponent(ctx, keyType, inferredKey, "map key", entry.key);
          const inferredValue = ctx.inferExpression(entry.value);
          valueType = mergeMapComponent(ctx, valueType, inferredValue, "map value", entry.value);
        } else {
          const spreadType = ctx.inferExpression(entry.expression);
          if (spreadType.kind === "map") {
            keyType = mergeMapComponent(ctx, keyType, spreadType.key, "map key", entry.expression);
            valueType = mergeMapComponent(ctx, valueType, spreadType.value, "map value", entry.expression);
          } else if (spreadType.kind !== "unknown") {
            ctx.report(`typechecker: map spread expects Map, got ${formatType(spreadType)}`, entry.expression);
          }
        }
      }
      return { kind: "map", key: keyType ?? unknownType, value: valueType ?? unknownType };
    }
    case "MatchExpression":
      return evaluateMatchExpression(ctx, expression);
    case "RescueExpression":
      return evaluateRescueExpression(ctx, expression);
    case "IfExpression": {
      const branchTypes: TypeInfo[] = [];
      const condType = ctx.inferExpression(expression.ifCondition);
      if (!isBoolean(condType)) {
        ctx.report("typechecker: if condition must be bool", expression.ifCondition);
      }
      branchTypes.push(ctx.inferExpression(expression.ifBody));
      if (Array.isArray(expression.orClauses)) {
        for (const clause of expression.orClauses) {
          if (!clause) continue;
          if (clause.condition) {
            const clauseCond = ctx.inferExpression(clause.condition);
            if (!isBoolean(clauseCond)) {
              ctx.report("typechecker: if-or condition must be bool", clause.condition);
            }
          }
          branchTypes.push(ctx.inferExpression(clause.body));
        }
      }
      return mergeBranchTypes(ctx, branchTypes);
    }
    case "BlockExpression":
      return evaluateBlockExpression(ctx, expression);
    case "ProcExpression": {
      ctx.pushAsyncContext();
      const bodyType = ctx.inferExpression(expression.expression);
      ctx.popAsyncContext();
      return procType(bodyType);
    }
    case "SpawnExpression": {
      ctx.pushAsyncContext();
      const bodyType = ctx.inferExpression(expression.expression);
      ctx.popAsyncContext();
      return futureType(bodyType);
    }
    case "FunctionCall": {
      ctx.checkFunctionCall(expression);
      return ctx.inferFunctionCallReturnType(expression);
    }
    case "StructLiteral":
      return checkStructLiteral(ctx, expression);
    case "MemberAccessExpression":
      ctx.handlePackageMemberAccess(expression);
      return unknownType;
    case "IteratorLiteral":
      return checkIteratorLiteral(ctx, expression);
    default:
      return unknownType;
  }
}

export function checkIteratorLiteral(ctx: ExpressionContext, literal: AST.IteratorLiteral): TypeInfo {
  const explicitType = literal?.elementType ? ctx.resolveTypeExpression(literal.elementType) : null;
  const expectedType = explicitType && explicitType.kind !== "unknown" ? explicitType : unknownType;
  const inferredType = analyzeIteratorBody(ctx, literal, expectedType);
  const elementType = explicitType && explicitType.kind !== "unknown" ? explicitType : inferredType;
  return iteratorType(elementType ?? unknownType);
}

export function analyzeIteratorBody(
  ctx: ExpressionContext,
  literal: AST.IteratorLiteral | undefined | null,
  expectedType: TypeInfo,
): TypeInfo {
  if (!literal || !Array.isArray(literal.body) || literal.body.length === 0) {
    return expectedType ?? unknownType;
  }
  const expectedLabel =
    literal.elementType && ctx.describeTypeExpression(literal.elementType)
      ? ctx.describeTypeExpression(literal.elementType)
      : formatType(expectedType);
  let inferred: TypeInfo = expectedType ?? unknownType;
  ctx.withForkedEnv(() => {
    ctx.defineValue("gen", unknownType);
    if (literal?.binding?.name) {
      ctx.defineValue(literal.binding.name, unknownType);
    }
    for (const statement of literal.body ?? []) {
      if (!statement) continue;
      if (statement.type === "YieldStatement") {
        const yieldType = checkIteratorYield(ctx, statement, expectedType, expectedLabel);
        if (yieldType && yieldType.kind !== "unknown") {
          if (inferred.kind === "unknown") {
            inferred = yieldType;
          } else if (!ctx.typeInfosEquivalent(inferred, yieldType)) {
            inferred = unknownType;
          }
        }
        continue;
      }
      ctx.checkStatement(statement);
    }
  });
  return inferred;
}

export function checkIteratorYield(
  ctx: ExpressionContext,
  statement: AST.YieldStatement,
  expectedType: TypeInfo,
  expectedLabel: string,
): TypeInfo {
  const valueType = statement.expression ? ctx.inferExpression(statement.expression) : primitiveType("nil");
  if (!ctx.typeInfosEquivalent(valueType, expectedType) && expectedType.kind !== "unknown") {
    const actualLabel = formatType(valueType);
    const literalMessage = ctx.describeLiteralMismatch(valueType, expectedType);
    ctx.report(
      literalMessage ??
        `typechecker: iterator annotation expects elements of type ${expectedLabel}, got ${actualLabel}`,
      statement,
    );
  }
  return valueType ?? unknownType;
}

export function evaluateMatchExpression(ctx: ExpressionContext, expression: AST.MatchExpression): TypeInfo {
  if (!expression) {
    return unknownType;
  }
  const subjectType = ctx.inferExpression(expression.subject);
  if (!Array.isArray(expression.clauses) || expression.clauses.length === 0) {
    return unknownType;
  }
  const branchTypes: TypeInfo[] = [];
  for (const clause of expression.clauses) {
    if (!clause) continue;
    ctx.pushScope();
    try {
      bindPatternToEnv(ctx, clause.pattern as AST.Pattern, subjectType, "match pattern");
      if (clause.guard) {
        const guardType = ctx.inferExpression(clause.guard);
        if (guardType && guardType.kind !== "unknown" && !isBoolean(guardType)) {
          ctx.report("typechecker: match guard must be bool", clause.guard);
        }
      }
      branchTypes.push(ctx.inferExpression(clause.body));
    } finally {
      ctx.popScope();
    }
  }
  return mergeBranchTypes(ctx, branchTypes);
}

export function evaluateRescueExpression(ctx: ExpressionContext, expression: AST.RescueExpression): TypeInfo {
  if (!expression) {
    return unknownType;
  }
  const monitoredType = ctx.inferExpression(expression.monitoredExpression);
  const branchTypes: TypeInfo[] = [monitoredType];
  const errorType = lookupErrorType(ctx);
  if (Array.isArray(expression.clauses)) {
    for (const clause of expression.clauses) {
      if (!clause) continue;
      ctx.pushScope();
      try {
        bindPatternToEnv(ctx, clause.pattern as AST.Pattern, errorType, "rescue pattern");
        if (clause.guard) {
          const guardType = ctx.inferExpression(clause.guard);
          if (guardType && guardType.kind !== "unknown" && !isBoolean(guardType)) {
            ctx.report("typechecker: rescue guard must be bool", clause.guard);
          }
        }
        branchTypes.push(ctx.inferExpression(clause.body));
      } finally {
        ctx.popScope();
      }
    }
  }
  return mergeBranchTypes(ctx, branchTypes);
}

export function lookupErrorType(ctx: ExpressionContext): TypeInfo {
  const structDefinition = ctx.getStructDefinition("Error");
  return {
    kind: "struct",
    name: "Error",
    typeArguments: [],
    definition: structDefinition,
  };
}

export function mergeBranchTypes(ctx: ExpressionContext, types: TypeInfo[]): TypeInfo {
  if (!types || types.length === 0) {
    return unknownType;
  }
  let current: TypeInfo = unknownType;
  for (const type of types) {
    if (!type || type.kind === "unknown") {
      continue;
    }
    if (current.kind === "unknown") {
      current = type;
      continue;
    }
    if (!ctx.typeInfosEquivalent(current, type)) {
      return unknownType;
    }
  }
  return current;
}

function mergeMapComponent(
  ctx: ExpressionContext,
  current: TypeInfo,
  candidate: TypeInfo,
  label: string,
  node: AST.Node,
): TypeInfo {
  if (!current || current.kind === "unknown") {
    return candidate ?? unknownType;
  }
  if (!candidate || candidate.kind === "unknown") {
    return current;
  }
  if (!ctx.typeInfosEquivalent(current, candidate)) {
    const expectedLabel = formatType(current);
    const actualLabel = formatType(candidate);
    ctx.report(`typechecker: ${label} expects type ${expectedLabel}, got ${actualLabel}`, node);
  }
  return current;
}

function shouldDeclareIdentifier(
  name: string | null | undefined,
  declarationNames: Set<string> | undefined,
): boolean {
  if (!name) return false;
  if (!declarationNames) return true;
  return declarationNames.has(name);
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
        !ctx.typeInfosEquivalent(annotationType, valueType)
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

export function evaluateBlockExpression(ctx: ExpressionContext, block: AST.BlockExpression): TypeInfo {
  const statements = Array.isArray(block?.body) ? block.body : [];
  if (!statements.length) {
    return unknownType;
  }
  ctx.pushScope();
  let resultType: TypeInfo = unknownType;
  try {
    for (let index = 0; index < statements.length; index += 1) {
      const statement = statements[index];
      if (!statement) continue;
      const isLast = index === statements.length - 1;
      const isExpr = ctx.isExpression(statement as AST.Node);
      if (isLast && isExpr) {
        resultType = ctx.inferExpression(statement as AST.Expression);
      } else {
        ctx.checkStatement(statement as AST.Statement);
      }
    }
  } finally {
    ctx.popScope();
  }
  return resultType;
}

function resolveArrayLiteralElementType(
  ctx: ExpressionContext,
  elements: Array<AST.Expression | null | undefined>,
): TypeInfo {
  if (!Array.isArray(elements) || elements.length === 0) {
    return unknownType;
  }
  let current = unknownType;
  for (const element of elements) {
    if (!element) continue;
    const inferred = ctx.inferExpression(element);
    if (!inferred || inferred.kind === "unknown") {
      continue;
    }
    if (current.kind === "unknown") {
      current = inferred;
      continue;
    }
    if (!ctx.typeInfosEquivalent(current, inferred)) {
      return unknownType;
    }
  }
  return current;
}

function checkStructLiteral(ctx: ExpressionContext, literal: AST.StructLiteral): TypeInfo {
  if (!literal) {
    return unknownType;
  }
  const structName = ctx.getIdentifierName(literal.structType);
  const typeArguments = Array.isArray(literal.typeArguments)
    ? literal.typeArguments.map((arg) => ctx.resolveTypeExpression(arg))
    : [];
  const definition = structName ? ctx.getStructDefinition(structName) : undefined;
  if (structName && !definition) {
    ctx.report(`typechecker: unknown struct '${structName}'`, literal);
  }
  const substitution = buildStructTypeSubstitution(ctx, definition, typeArguments, literal, structName);
  if (Array.isArray(literal.functionalUpdateSources)) {
    for (const source of literal.functionalUpdateSources) {
      if (!source) continue;
      const sourceType = ctx.inferExpression(source);
      if (
        structName &&
        sourceType &&
        sourceType.kind !== "unknown" &&
        (sourceType.kind !== "struct" || sourceType.name !== structName)
      ) {
        ctx.report(
          `typechecker: functional update expects struct ${structName}, got ${formatType(sourceType)}`,
          source,
        );
      }
    }
  }
  const { namedFields, positionalFields } = resolveStructFieldTypes(ctx, definition, substitution);
  const fields = Array.isArray(literal.fields) ? literal.fields : [];
  const seenFields = new Set<string>();
  fields.forEach((field, index) => {
    if (!field) return;
    const fieldName = ctx.getIdentifierName(field.name);
    const valueType = ctx.inferExpression(field.value);
    if (!literal.isPositional && !fieldName) {
      ctx.report("typechecker: struct field requires a name", field);
      return;
    }
    if (fieldName) {
      if (seenFields.has(fieldName)) {
        ctx.report(`typechecker: duplicate struct field '${fieldName}'`, field);
        return;
      }
      seenFields.add(fieldName);
    }
    let expected: TypeInfo | undefined;
    if (definition) {
      if (!literal.isPositional && fieldName) {
        expected = namedFields.get(fieldName);
        if (!expected) {
          ctx.report(`typechecker: struct '${structName}' has no field '${fieldName}'`, field);
          return;
        }
      } else if (literal.isPositional) {
        expected = positionalFields[index];
        if (!expected) {
          ctx.report(
            `typechecker: positional field ${index} out of range for struct '${structName}'`,
            field,
          );
          return;
        }
      }
    }
    if (
      expected &&
      expected.kind !== "unknown" &&
      valueType &&
      valueType.kind !== "unknown"
    ) {
      const literalMessage = ctx.describeLiteralMismatch(valueType, expected);
      if (literalMessage) {
        ctx.report(literalMessage, field.value);
        return;
      }
      if (!ctx.typeInfosEquivalent(valueType, expected)) {
        const label = fieldName ?? `#${index}`;
        ctx.report(
          `typechecker: struct field '${label}' expects type ${formatType(expected)}, got ${formatType(valueType)}`,
          field.value,
        );
      }
    }
  });
  if (structName) {
    return {
      kind: "struct",
      name: structName,
      typeArguments,
      definition,
    };
  }
  return unknownType;
}

function buildStructTypeSubstitution(
  ctx: ExpressionContext,
  definition: AST.StructDefinition | undefined,
  typeArguments: TypeInfo[],
  literal: AST.StructLiteral,
  structName: string | null,
): Map<string, TypeInfo> {
  const substitution = new Map<string, TypeInfo>();
  if (!definition || !Array.isArray(definition.genericParams)) {
    return substitution;
  }
  const expectedCount = definition.genericParams.length;
  if (structName && typeArguments.length > 0 && typeArguments.length !== expectedCount) {
    ctx.report(
      `typechecker: struct '${structName}' expects ${expectedCount} type argument(s), got ${typeArguments.length}`,
      literal,
    );
  }
  definition.genericParams.forEach((param, index) => {
    const paramName = ctx.getIdentifierName(param?.name);
    if (!paramName) {
      return;
    }
    substitution.set(paramName, typeArguments[index] ?? unknownType);
  });
  return substitution;
}

function resolveStructFieldTypes(
  ctx: ExpressionContext,
  definition: AST.StructDefinition | undefined,
  substitution: Map<string, TypeInfo>,
): { namedFields: Map<string, TypeInfo>; positionalFields: TypeInfo[] } {
  const namedFields = new Map<string, TypeInfo>();
  const positionalFields: TypeInfo[] = [];
  if (!definition || !Array.isArray(definition.fields)) {
    return { namedFields, positionalFields };
  }
  definition.fields.forEach((field, index) => {
    if (!field) {
      positionalFields[index] = unknownType;
      return;
    }
    const resolvedType = ctx.resolveTypeExpression(field.fieldType, substitution);
    const fieldName = ctx.getIdentifierName(field.name);
    if (fieldName) {
      namedFields.set(fieldName, resolvedType);
    }
    positionalFields[index] = resolvedType;
  });
  return { namedFields, positionalFields };
}

function resolveRangeElementType(ctx: ExpressionContext, start: TypeInfo, end: TypeInfo): TypeInfo {
  if (start && start.kind !== "unknown") {
    return start;
  }
  if (end && end.kind !== "unknown") {
    return end;
  }
  return primitiveType("i32");
}
