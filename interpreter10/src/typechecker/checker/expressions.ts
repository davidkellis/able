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
  resolveTypeExpression(expr: AST.TypeExpression | null | undefined): TypeInfo;
  getStructDefinition(name: string): AST.StructDefinition | undefined;
  handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  checkStatement(node: AST.Statement | AST.Expression | undefined | null): void;
  pushAsyncContext(): void;
  popAsyncContext(): void;
  checkFunctionCall(call: AST.FunctionCall): void;
  inferFunctionCallReturnType(call: AST.FunctionCall): TypeInfo;
  pushScope(): void;
  popScope(): void;
  withForkedEnv<T>(fn: () => T): T;
  lookupIdentifier(name: string): TypeInfo | undefined;
  defineValue(name: string, valueType: TypeInfo): void;
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

export function inferExpression(ctx: ExpressionContext, expression: AST.Expression | undefined | null): TypeInfo {
  if (!expression) return unknownType;
  switch (expression.type) {
    case "StringLiteral":
      return primitiveType("string");
    case "BooleanLiteral":
      return primitiveType("bool");
    case "IntegerLiteral":
      return primitiveType("i32");
    case "FloatLiteral":
      return primitiveType("f64");
    case "NilLiteral":
      return primitiveType("nil");
    case "Identifier": {
      const existing = ctx.lookupIdentifier(expression.name);
      return existing ?? unknownType;
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
      return rangeType(elementType);
    }
    case "ArrayLiteral": {
      const elementType = resolveArrayLiteralElementType(ctx, expression.elements);
      return arrayType(elementType);
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
    case "StructLiteral": {
      const structName = ctx.getIdentifierName(expression.structType);
      if (structName) {
        const typeArguments = Array.isArray(expression.typeArguments)
          ? expression.typeArguments.map((arg) => ctx.resolveTypeExpression(arg))
          : [];
        return {
          kind: "struct",
          name: structName,
          typeArguments,
          definition: ctx.getStructDefinition(structName),
        };
      }
      return unknownType;
    }
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
    ctx.report(
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

export function bindPatternToEnv(
  ctx: ExpressionContext,
  pattern: AST.Pattern | undefined | null,
  valueType: TypeInfo,
  contextLabel: string,
): void {
  if (!pattern) return;
  switch (pattern.type) {
    case "Identifier":
      if (pattern.name) {
        ctx.defineValue(pattern.name, valueType ?? unknownType);
      }
      return;
    case "WildcardPattern":
      return;
    case "TypedPattern": {
      const annotationType = ctx.resolveTypeExpression(pattern.typeAnnotation);
      const resolvedType =
        annotationType && annotationType.kind !== "unknown" ? annotationType : valueType ?? unknownType;
      if (
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
        bindPatternToEnv(ctx, pattern.pattern as AST.Pattern, resolvedType ?? valueType, contextLabel);
      }
      return;
    }
    case "StructPattern":
      checkStructPattern(ctx, pattern, valueType);
      return;
    case "ArrayPattern":
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
      if (isLast && (statement as AST.Expression).type) {
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

function resolveRangeElementType(ctx: ExpressionContext, start: TypeInfo, end: TypeInfo): TypeInfo {
  if (start && start.kind !== "unknown") {
    return start;
  }
  if (end && end.kind !== "unknown") {
    return end;
  }
  return primitiveType("i32");
}
