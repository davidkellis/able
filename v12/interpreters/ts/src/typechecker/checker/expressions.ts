import type * as AST from "../../ast";
import {
  arrayType,
  describe,
  formatType,
  futureType,
  iteratorType,
  isBoolean,
  isFloatPrimitiveType,
  isIntegerPrimitiveType,
  isNumeric,
  isRatioType,
  primitiveType,
  rangeType,
  unknownType,
  type PrimitiveName,
  type TypeInfo,
} from "../types";
import type { ExpressionContext, StatementContext } from "./expression-context";
import { bindPatternToEnv } from "./patterns";
import { checkAssignment } from "./statements";
import {
  applyNumericResolution,
  checkStructLiteral,
  evaluateBlockExpression,
  isStringType,
  mergeMapComponent,
  resolveArrayLiteralElementType,
  resolveDivisionType,
  resolveIntegerBinaryType,
  resolveNumericBinaryType,
  resolveRangeElementType,
  type NumericResolution,
} from "./expression-helpers";
import { placeholderFunctionPlan } from "./placeholders";

function buildPipeCall(expression: AST.BinaryExpression): AST.FunctionCall {
  const rhs = expression.right;
  const placeholderPlan = placeholderFunctionPlan(rhs);
  if (placeholderPlan) {
    return {
      type: "FunctionCall",
      callee: rhs,
      arguments: [expression.left],
      isTrailingLambda: false,
    };
  }
  if (rhs.type === "FunctionCall") {
    return {
      ...rhs,
      arguments: [expression.left, ...rhs.arguments],
      isTrailingLambda: rhs.isTrailingLambda,
      type: "FunctionCall",
    };
  }
  return {
    type: "FunctionCall",
    callee: rhs,
    arguments: [expression.left],
    isTrailingLambda: false,
  };
}
export function inferExpression(ctx: ExpressionContext, expression: AST.Expression | undefined | null): TypeInfo {
  if (!expression) return unknownType;
  const placeholderPlan = placeholderFunctionPlan(expression);
  if (placeholderPlan) {
    return {
      kind: "function",
      parameters: Array.from({ length: placeholderPlan.paramCount }, () => unknownType),
      returnType: unknownType,
    };
  }
  switch (expression.type) {
    case "StringLiteral":
      return primitiveType("String");
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
      if (ctx.isTypeParamInScope(name)) {
        return { kind: "type_parameter", name };
      }
      const structDefinition = ctx.getStructDefinition(name);
      if (structDefinition && ctx.isTypeNameInScope(name)) {
        const paramCount = Array.isArray(structDefinition.genericParams) ? structDefinition.genericParams.length : 0;
        const typeArguments = paramCount > 0 ? Array.from({ length: paramCount }, () => unknownType) : [];
        return { kind: "struct", name, typeArguments, definition: structDefinition };
      }
      if (ctx.allowDynamicLookup()) {
        return unknownType;
      }
      ctx.report(`typechecker: undefined identifier '${name}'`, expression);
      return unknownType;
    }
    case "UnaryExpression":
      return inferUnaryExpression(ctx, expression);
    case "TypeCastExpression":
      return inferTypeCastExpression(ctx, expression);
    case "BinaryExpression":
      return inferBinaryExpression(ctx, expression);
    case "RangeExpression": {
      const start = ctx.inferExpression(expression.start);
      if (start.kind !== "unknown" && !isIntegerPrimitiveType(start)) {
        ctx.report("typechecker: range start must be numeric", expression);
      }
      const end = ctx.inferExpression(expression.end);
      if (end.kind !== "unknown" && !isIntegerPrimitiveType(end)) {
        ctx.report("typechecker: range end must be numeric", expression);
      }
      const elementType = resolveRangeElementType(ctx, start, end);
      const bounds: TypeInfo[] = [];
      if (start.kind !== "unknown" && isIntegerPrimitiveType(start)) {
        bounds.push(start);
      }
      if (end.kind !== "unknown" && isIntegerPrimitiveType(end)) {
        bounds.push(end);
      }
      return rangeType(elementType, bounds.length > 0 ? bounds : undefined);
    }
    case "ArrayLiteral": {
      const elementType = resolveArrayLiteralElementType(ctx, expression.elements);
      return arrayType(elementType);
    }
    case "IndexExpression":
      return inferIndexExpression(ctx, expression);
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
            continue;
          }
          if (spreadType.kind === "struct" && spreadType.name === "HashMap") {
            const args = spreadType.typeArguments ?? [];
            keyType = mergeMapComponent(ctx, keyType, args[0] ?? unknownType, "map key", entry.expression);
            valueType = mergeMapComponent(ctx, valueType, args[1] ?? unknownType, "map value", entry.expression);
            continue;
          }
          if (spreadType.kind !== "unknown") {
            ctx.report(`typechecker: map spread expects Map or HashMap, got ${formatType(spreadType)}`, entry.expression);
          }
        }
      }
      const structDef = ctx.getStructDefinition("HashMap");
      return {
        kind: "struct",
        name: "HashMap",
        typeArguments: [keyType ?? unknownType, valueType ?? unknownType],
        definition: structDef,
      };
    }
    case "MatchExpression":
      return evaluateMatchExpression(ctx, expression);
    case "RescueExpression":
      return evaluateRescueExpression(ctx, expression);
    case "OrElseExpression":
      return inferOrElseExpression(ctx, expression);
    case "IfExpression": {
      const branchTypes: TypeInfo[] = [];
      ctx.inferExpression(expression.ifCondition);
      branchTypes.push(ctx.inferExpression(expression.ifBody));
      for (const clause of expression.elseIfClauses ?? []) {
        if (!clause) continue;
        ctx.inferExpression(clause.condition);
        branchTypes.push(ctx.inferExpression(clause.body));
      }
      if (expression.elseBody) {
        branchTypes.push(ctx.inferExpression(expression.elseBody));
      }
      return mergeBranchTypes(ctx, branchTypes);
    }
    case "BlockExpression":
      return evaluateBlockExpression(ctx, expression);
    case "LoopExpression": {
      ctx.pushLoopContext();
      let loopResult: TypeInfo;
      try {
        ctx.inferExpression(expression.body);
      } finally {
        loopResult = ctx.popLoopContext();
      }
      return loopResult ?? unknownType;
    }
    case "BreakpointExpression": {
      const labelName = ctx.getIdentifierName(expression.label);
      if (!labelName) {
        ctx.report("typechecker: breakpoint requires a label", expression.label ?? expression);
        return ctx.inferExpression(expression.body);
      }
      ctx.pushBreakpointLabel(labelName);
      let bodyType: TypeInfo;
      try {
        bodyType = ctx.inferExpression(expression.body);
      } finally {
        ctx.popBreakpointLabel();
      }
      return bodyType ?? unknownType;
    }
    case "SpawnExpression": {
      ctx.pushAsyncContext();
      const bodyType = ctx.inferExpression(expression.expression);
      ctx.popAsyncContext();
      return futureType(bodyType);
    }
    case "AwaitExpression":
      return inferAwaitExpression(ctx, expression);
    case "AssignmentExpression":
      // Allow assignments to appear in expression position (e.g., inside a pipe).
      // Reuse the statement checker so bindings are recorded, then return the RHS type.
      return checkAssignment(ctx as StatementContext, expression);
    case "FunctionCall": {
      ctx.checkFunctionCall(expression);
      return ctx.inferFunctionCallReturnType(expression);
    }
    case "StructLiteral":
      return checkStructLiteral(ctx, expression);
    case "MemberAccessExpression":
      return ctx.handlePackageMemberAccess(expression) ?? unknownType;
    case "IteratorLiteral":
      return checkIteratorLiteral(ctx, expression);
    default:
      return unknownType;
  }
}

export function inferExpressionWithExpected(
  ctx: ExpressionContext,
  expression: AST.Expression | undefined | null,
  expectedType: TypeInfo,
): TypeInfo {
  if (!expression) return unknownType;
  if (!expectedType || expectedType.kind === "unknown") {
    return inferExpression(ctx, expression);
  }
  switch (expression.type) {
    case "FunctionCall":
      ctx.checkFunctionCall(expression, expectedType);
      return ctx.inferFunctionCallReturnType(expression, expectedType);
    case "BlockExpression":
      return evaluateBlockExpression(ctx, expression, expectedType);
    case "BinaryExpression":
      if (expression.operator === "|>" || expression.operator === "|>>") {
        const call = buildPipeCall(expression);
        ctx.checkFunctionCall(call, expectedType);
        return ctx.inferFunctionCallReturnType(call, expectedType);
      }
      return inferBinaryExpression(ctx, expression);
    default:
      return inferExpression(ctx, expression);
  }
}

export function refineTypeWithExpected(actual: TypeInfo | undefined | null, expected: TypeInfo | undefined | null): TypeInfo {
  if (!expected || expected.kind === "unknown") {
    return actual ?? unknownType;
  }
  if (!actual || actual.kind === "unknown") {
    return expected;
  }
  if (actual.kind !== expected.kind) {
    return actual;
  }
  switch (actual.kind) {
    case "array":
      return { kind: "array", element: refineTypeWithExpected(actual.element, expected.element) };
    case "map":
      return {
        kind: "map",
        key: refineTypeWithExpected(actual.key, expected.key),
        value: refineTypeWithExpected(actual.value, expected.value),
      };
    case "range":
      return { kind: "range", element: refineTypeWithExpected(actual.element, expected.element), bounds: actual.bounds };
    case "iterator":
      return { kind: "iterator", element: refineTypeWithExpected(actual.element, expected.element) };
    case "future":
      return { kind: "future", result: refineTypeWithExpected(actual.result, expected.result) };
    case "nullable":
      return { kind: "nullable", inner: refineTypeWithExpected(actual.inner, expected.inner) };
    case "result":
      return { kind: "result", inner: refineTypeWithExpected(actual.inner, expected.inner) };
    case "struct": {
      if (actual.name !== expected.name) {
        return actual;
      }
      const actualArgs = Array.isArray(actual.typeArguments) ? actual.typeArguments : [];
      const expectedArgs = Array.isArray(expected.typeArguments) ? expected.typeArguments : [];
      if (actualArgs.length === 0) {
        return actual;
      }
      const typeArguments = actualArgs.map((arg, index) => refineTypeWithExpected(arg, expectedArgs[index]));
      return { ...actual, typeArguments };
    }
    case "interface": {
      if (actual.name !== expected.name) {
        return actual;
      }
      const actualArgs = Array.isArray(actual.typeArguments) ? actual.typeArguments : [];
      const expectedArgs = Array.isArray(expected.typeArguments) ? expected.typeArguments : [];
      if (actualArgs.length === 0) {
        return actual;
      }
      const typeArguments = actualArgs.map((arg, index) => refineTypeWithExpected(arg, expectedArgs[index]));
      return { ...actual, typeArguments };
    }
    case "union": {
      if (!Array.isArray(actual.members) || !Array.isArray(expected.members)) {
        return actual;
      }
      if (actual.members.length !== expected.members.length) {
        return actual;
      }
      return {
        kind: "union",
        members: actual.members.map((member, index) => refineTypeWithExpected(member, expected.members[index])),
      };
    }
    default:
      return actual;
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
  if (!ctx.isTypeAssignable(valueType, expectedType) && expectedType.kind !== "unknown") {
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
        ctx.inferExpression(clause.guard);
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
          ctx.inferExpression(clause.guard);
        }
        branchTypes.push(ctx.inferExpression(clause.body));
      } finally {
        ctx.popScope();
      }
    }
  }
  return mergeBranchTypes(ctx, branchTypes);
}

export function inferOrElseExpression(ctx: ExpressionContext, expression: AST.OrElseExpression): TypeInfo {
  if (!expression) {
    return unknownType;
  }
  const expressionType = ctx.inferExpression(expression.expression);
  const successType = stripOptionOrResult(ctx, expressionType);

  let handlerType: TypeInfo = primitiveType("nil");
  let handlerReturns = false;
  if (expression.handler) {
    ctx.pushScope();
    try {
      if (expression.errorBinding?.name) {
        ctx.defineValue(expression.errorBinding.name, lookupErrorType(ctx));
      }
      handlerType = ctx.inferExpression(expression.handler);
      handlerReturns = blockHasReturn(expression.handler);
    } finally {
      ctx.popScope();
    }
  }

  if (handlerReturns) {
    return successType ?? unknownType;
  }
  return mergeElseTypes(ctx, successType, handlerType);
}

export function lookupErrorType(ctx: ExpressionContext): TypeInfo {
  const interfaceDefinition = ctx.getInterfaceDefinition("Error");
  return {
    kind: "interface",
    name: "Error",
    typeArguments: [],
    definition: interfaceDefinition,
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

function stripOptionOrResult(ctx: ExpressionContext, type: TypeInfo): TypeInfo {
  if (!type || type.kind === "unknown") {
    return unknownType;
  }
  switch (type.kind) {
    case "nullable":
      return stripOptionOrResult(ctx, type.inner);
    case "result":
      return stripOptionOrResult(ctx, type.inner);
    case "union": {
      const members = (type.members ?? []).filter((member) => !isFailureType(ctx, member));
      if (members.length === 0) {
        return unknownType;
      }
      if (members.length === 1) {
        return stripOptionOrResult(ctx, members[0] ?? unknownType);
      }
      const stripped = members.map((member) => stripOptionOrResult(ctx, member));
      return ctx.normalizeUnionType(stripped);
    }
    default:
      return type;
  }
}

function isFailureType(ctx: ExpressionContext, type: TypeInfo | undefined): boolean {
  if (!type || type.kind === "unknown") return false;
  if (type.kind === "primitive" && type.name === "nil") return true;
  if (type.kind === "interface" && type.name === "Error") return true;
  if (type.kind === "struct" && type.name === "Error") return true;
  if (type.kind === "union") {
    return (type.members ?? []).some((member) => isFailureType(ctx, member));
  }
  if (ctx.typeImplementsInterface && ctx.typeImplementsInterface(type, "Error").ok) {
    return true;
  }
  return false;
}

function mergeElseTypes(ctx: ExpressionContext, success: TypeInfo, handler: TypeInfo): TypeInfo {
  const members: TypeInfo[] = [];
  for (const candidate of [success, handler]) {
    if (!candidate || candidate.kind === "unknown") continue;
    const exists = members.some((existing) => ctx.typeInfosEquivalent(existing, candidate));
    if (!exists) members.push(candidate);
  }
  if (members.length === 0) return unknownType;
  if (members.length === 1) return members[0]!;
  return ctx.normalizeUnionType(members);
}

function blockHasReturn(block: AST.BlockExpression | null | undefined): boolean {
  if (!block || !Array.isArray(block.body)) return false;
  return block.body.some((stmt) => stmt?.type === "ReturnStatement");
}

function inferIndexExpression(ctx: ExpressionContext, expression: AST.IndexExpression): TypeInfo {
  const objectType = ctx.inferExpression(expression.object);
  const indexType = ctx.inferExpression(expression.index);
  return resolveIndexResultType(ctx, objectType, indexType, expression);
}

function resolveIndexResultType(
  ctx: ExpressionContext,
  objectType: TypeInfo,
  indexType: TypeInfo,
  node: AST.Node,
): TypeInfo {
  const wrapResult = (inner: TypeInfo): TypeInfo => ({ kind: "result", inner: inner ?? unknownType });
  const requiresIntegerIndex = (): void => {
    if (!isIntegerPrimitiveType(indexType) && indexType.kind !== "unknown") {
      ctx.report("typechecker: index must be an integer", node);
    }
  };
  if (!objectType || objectType.kind === "unknown") {
    return unknownType;
  }
  if (objectType.kind === "array") {
    requiresIntegerIndex();
    return wrapResult(objectType.element ?? unknownType);
  }
  if (objectType.kind === "struct" && objectType.name === "Array") {
    requiresIntegerIndex();
    return wrapResult((objectType.typeArguments ?? [])[0] ?? unknownType);
  }
  if (objectType.kind === "map") {
    return wrapResult(objectType.value ?? unknownType);
  }
  if (objectType.kind === "struct" && objectType.name === "HashMap") {
    const args = objectType.typeArguments ?? [];
    if (args.length >= 2) {
      return wrapResult(args[1] ?? unknownType);
    }
    return wrapResult(unknownType);
  }
  if (objectType.kind === "interface" && objectType.name === "Index") {
    const keyType = (objectType.typeArguments ?? [])[0];
    const valueType = (objectType.typeArguments ?? [])[1];
    if (keyType && !ctx.isTypeAssignable(indexType, keyType) && indexType.kind !== "unknown") {
      ctx.report(
        `typechecker: index expects type ${formatType(keyType)}, got ${formatType(indexType)}`,
        node,
      );
    }
    return wrapResult(valueType ?? unknownType);
  }
  if (ctx.typeImplementsInterface?.(objectType, "Index", ["Unknown", "Unknown"])?.ok) {
    return wrapResult(unknownType);
  }
  if (ctx.typeImplementsInterface?.(objectType, "IndexMut", ["Unknown", "Unknown"])?.ok) {
    return wrapResult(unknownType);
  }
  ctx.report(`typechecker: cannot index into type ${formatType(objectType)}`, node);
  return unknownType;
}

function inferAwaitExpression(ctx: ExpressionContext, expression: AST.AwaitExpression): TypeInfo {
  const iterableType = ctx.inferExpression(expression.expression);
  const { elementType, recognized } = resolveIterableElementTypeLite(iterableType);
  if (!recognized && iterableType.kind !== "unknown") {
    ctx.report(
      `typechecker: await expects an Iterable of Awaitable values (got ${formatType(iterableType)})`,
      expression.expression,
    );
  }
  let matched = false;
  let resultType: TypeInfo = unknownType;
  if (elementType.kind === "interface" && elementType.name === "Awaitable") {
    matched = true;
    resultType = (elementType.typeArguments ?? [])[0] ?? unknownType;
  } else if (elementType.kind === "future") {
    matched = true;
    resultType = elementType.result ?? unknownType;
  }

  if (!matched && elementType.kind !== "unknown") {
    ctx.report(
      `typechecker: await expects Awaitable values (got ${formatType(elementType)})`,
      expression.expression,
    );
  }
  return resultType;
}

function resolveIterableElementTypeLite(type: TypeInfo): { elementType: TypeInfo; recognized: boolean } {
  if (!type || type.kind === "unknown") {
    return { elementType: unknownType, recognized: true };
  }
  if (type.kind === "primitive" && type.name === "String") {
    return { elementType: primitiveType("u8"), recognized: true };
  }
  if (type.kind === "iterator") {
    return { elementType: type.element ?? unknownType, recognized: true };
  }
  if (type.kind === "array") {
    return { elementType: type.element ?? unknownType, recognized: true };
  }
  if (type.kind === "range") {
    return { elementType: type.element ?? unknownType, recognized: true };
  }
  if (type.kind === "struct") {
    const arg0 = (type.typeArguments ?? [])[0];
    if (
      type.name === "Array" ||
      type.name === "Iterator" ||
      type.name === "List" ||
      type.name === "LinkedList" ||
      type.name === "LazySeq" ||
      type.name === "Vector" ||
      type.name === "HashSet" ||
      type.name === "Deque" ||
      type.name === "Queue"
    ) {
      return { elementType: arg0 ?? unknownType, recognized: true };
    }
    if (type.name === "BitSet") {
      return { elementType: primitiveType("i32"), recognized: true };
    }
    if (type.name === "Range") {
      return { elementType: unknownType, recognized: true };
    }
  }
  if (type.kind === "interface" && type.name === "Iterable") {
    const arg0 = (type.typeArguments ?? [])[0];
    return { elementType: arg0 ?? unknownType, recognized: true };
  }
  return { elementType: unknownType, recognized: false };
}

function inferUnaryExpression(ctx: ExpressionContext, expression: AST.UnaryExpression): TypeInfo {
  if (!expression) {
    return unknownType;
  }
  const operandType = ctx.inferExpression(expression.operand);
  switch (expression.operator) {
    case "-":
      if (operandType.kind === "unknown") {
        return unknownType;
      }
      if (!isNumeric(operandType)) {
        ctx.report(`typechecker: unary '-' requires numeric operand (got ${describe(operandType)})`, expression);
        return unknownType;
      }
      return operandType;
    case "!":
      return primitiveType("bool");
    case ".~":
      if (operandType.kind === "unknown") {
        return unknownType;
      }
      if (!isIntegerPrimitiveType(operandType)) {
        ctx.report(`typechecker: unary '.~' requires integer operand (got ${describe(operandType)})`, expression);
        return unknownType;
      }
      return operandType;
    default:
      ctx.report(`typechecker: unsupported unary operator '${expression.operator}'`, expression);
      return unknownType;
  }
}

function isPrimitiveNumericType(type: TypeInfo): boolean {
  return isIntegerPrimitiveType(type) || isFloatPrimitiveType(type);
}

function inferTypeCastExpression(ctx: ExpressionContext, expression: AST.TypeCastExpression): TypeInfo {
  const valueType = ctx.inferExpression(expression.expression);
  const targetType = ctx.resolveTypeExpression(expression.targetType);
  if (valueType.kind === "unknown" || targetType.kind === "unknown") {
    return targetType;
  }
  if (ctx.isTypeAssignable(valueType, targetType)) {
    return targetType;
  }
  if (isPrimitiveNumericType(valueType) && isPrimitiveNumericType(targetType)) {
    return targetType;
  }
  ctx.report(`typechecker: cannot cast ${describe(valueType)} to ${describe(targetType)}`, expression);
  return targetType;
}

function inferBinaryExpression(ctx: ExpressionContext, expression: AST.BinaryExpression): TypeInfo {
  if (!expression) {
    return unknownType;
  }
  const operator = expression.operator;
  if (operator === "|>" || operator === "|>>") {
    const call = buildPipeCall(expression);
    ctx.checkFunctionCall(call);
    return ctx.inferFunctionCallReturnType(call);
  }
  const left = ctx.inferExpression(expression.left);
  const right = ctx.inferExpression(expression.right);
  if (operator === "&&" || operator === "||") {
    return mergeBranchTypes(ctx, [left, right]);
  }
  if (operator === "+") {
    if (isStringType(left) && isStringType(right)) {
      return primitiveType("String");
    }
  }
  if (["+", "-", "*", "^"].includes(operator)) {
    if (operator === "^" && (isRatioType(left) || isRatioType(right))) {
      ctx.report("typechecker: '^' does not support Ratio operands", expression);
      return unknownType;
    }
    const result = resolveNumericBinaryType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  if (operator === "/") {
    const result = resolveDivisionType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  if (operator === "//" || operator === "%") {
    const result = resolveIntegerBinaryType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  if (operator === "/%") {
    const result = resolveIntegerBinaryType(left, right);
    if (result.kind === "ok") {
      const base = ctx.getStructDefinition("DivMod");
      return {
        kind: "struct",
        name: "DivMod",
        typeArguments: [result.type ?? unknownType],
        definition: base,
      };
    }
    return applyNumericResolution(ctx, expression, operator, result);
  }
  if (operator === "==" || operator === "!=") {
    return primitiveType("bool");
  }
  if ([">", "<", ">=", "<="].includes(operator)) {
    if (isStringType(left) && isStringType(right)) {
      return primitiveType("bool");
    }
    const resolution = resolveNumericBinaryType(left, right);
    if (resolution.kind === "error") {
      ctx.report(`typechecker: '${operator}' ${resolution.message}`, expression);
    }
    return primitiveType("bool");
  }
  if ([".&", ".|", ".^"].includes(operator)) {
    const result = resolveIntegerBinaryType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  if ([".<<", ".>>"].includes(operator)) {
    const result = resolveIntegerBinaryType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  ctx.report(`typechecker: unsupported binary operator '${operator}'`, expression);
  return unknownType;
}
