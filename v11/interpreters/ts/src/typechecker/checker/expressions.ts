import type * as AST from "../../ast";
import {
  arrayType,
  describe,
  formatType,
  futureType,
  iteratorType,
  isBoolean,
  isNumeric,
  isFloatPrimitiveType,
  isIntegerPrimitiveType,
  isUnknown,
  primitiveType,
  procType,
  rangeType,
  unknownType,
  type FloatPrimitive,
  type PrimitiveName,
  type TypeInfo,
} from "../types";
import {
  findSmallestSigned,
  findSmallestUnsigned,
  getIntegerTypeInfo,
  widestUnsignedInfo,
  type IntegerTypeInfo,
} from "../numeric";
import type { ExpressionContext, StatementContext } from "./expression-context";
import { bindPatternToEnv } from "./patterns";
import { checkAssignment } from "./statements";
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
    case "BinaryExpression":
      return inferBinaryExpression(ctx, expression);
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
    return objectType.element ?? unknownType;
  }
  if (objectType.kind === "struct" && objectType.name === "Array") {
    requiresIntegerIndex();
    return (objectType.typeArguments ?? [])[0] ?? unknownType;
  }
  if (objectType.kind === "map") {
    return objectType.value ?? unknownType;
  }
  if (objectType.kind === "struct" && objectType.name === "HashMap") {
    const args = objectType.typeArguments ?? [];
    if (args.length >= 2) {
      return args[1] ?? unknownType;
    }
    return unknownType;
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
    return valueType ?? unknownType;
  }
  if (ctx.typeImplementsInterface?.(objectType, "Index", ["Unknown", "Unknown"])?.ok) {
    return unknownType;
  }
  if (ctx.typeImplementsInterface?.(objectType, "IndexMut", ["Unknown", "Unknown"])?.ok) {
    return unknownType;
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
  } else if (elementType.kind === "proc") {
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
  if (type.kind === "primitive" && type.name === "string") {
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
      if (!isBoolean(operandType) && operandType.kind !== "unknown") {
        ctx.report("typechecker: unary '!' requires boolean operand", expression);
      }
      return primitiveType("bool");
    case "~":
      if (operandType.kind === "unknown") {
        return unknownType;
      }
      if (!isIntegerPrimitiveType(operandType)) {
        ctx.report(`typechecker: unary '~' requires integer operand (got ${describe(operandType)})`, expression);
        return unknownType;
      }
      return operandType;
    default:
      ctx.report(`typechecker: unsupported unary operator '${expression.operator}'`, expression);
      return unknownType;
  }
}

function inferBinaryExpression(ctx: ExpressionContext, expression: AST.BinaryExpression): TypeInfo {
  if (!expression) {
    return unknownType;
  }
  const operator = expression.operator;
  if (operator === "|>" || operator === "|>>") {
    ctx.inferExpression(expression.left);
    return unknownType;
  }
  const left = ctx.inferExpression(expression.left);
  const right = ctx.inferExpression(expression.right);
  if (operator === "&&" || operator === "||") {
    if (!isBoolean(left)) {
      ctx.report(`typechecker: '${operator}' left operand must be bool (got ${describe(left)})`, expression);
    }
    if (!isBoolean(right)) {
      ctx.report(`typechecker: '${operator}' right operand must be bool (got ${describe(right)})`, expression);
    }
    return primitiveType("bool");
  }
  if (operator === "+") {
    if (isStringType(left) && isStringType(right)) {
      return primitiveType("string");
    }
  }
  if (["+", "-", "*", "/", "%"].includes(operator)) {
    const result = resolveNumericBinaryType(left, right);
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
  if (["&", "|", "\\xor"].includes(operator)) {
    const result = resolveIntegerBinaryType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  if (["<<", ">>"].includes(operator)) {
    const result = resolveIntegerBinaryType(left, right);
    return applyNumericResolution(ctx, expression, operator, result);
  }
  ctx.report(`typechecker: unsupported binary operator '${operator}'`, expression);
  return unknownType;
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

type NumericResolution =
  | { kind: "ok"; type: TypeInfo }
  | { kind: "unknown" }
  | { kind: "error"; message: string };

function applyNumericResolution(
  ctx: ExpressionContext,
  node: AST.Node,
  operator: string,
  resolution: NumericResolution,
): TypeInfo {
  if (resolution.kind === "ok") {
    return resolution.type ?? unknownType;
  }
  if (resolution.kind === "unknown") {
    return unknownType;
  }
  ctx.report(`typechecker: '${operator}' ${resolution.message}`, node);
  return unknownType;
}

function resolveNumericBinaryType(left: TypeInfo, right: TypeInfo): NumericResolution {
  if (!left || left.kind === "unknown" || !right || right.kind === "unknown") {
    return { kind: "unknown" };
  }
  const leftClass = classifyNumericPrimitive(left);
  const rightClass = classifyNumericPrimitive(right);
  if (!leftClass || !rightClass) {
    return {
      kind: "error",
      message: `requires numeric operands (got ${formatType(left)} and ${formatType(right)})`,
    };
  }
  if (leftClass.kind === "float" || rightClass.kind === "float") {
    const resultName = leftClass.kind === "float" && leftClass.name === "f64"
      ? "f64"
      : rightClass.kind === "float" && rightClass.name === "f64"
        ? "f64"
        : "f32";
    return { kind: "ok", type: primitiveType(resultName) };
  }
  return resolveIntegerBinaryFromInfos(leftClass.info, rightClass.info);
}

function resolveIntegerBinaryType(left: TypeInfo, right: TypeInfo): NumericResolution {
  if (!left || left.kind === "unknown" || !right || right.kind === "unknown") {
    return { kind: "unknown" };
  }
  const leftInfo = extractIntegerInfo(left);
  const rightInfo = extractIntegerInfo(right);
  if (!leftInfo || !rightInfo) {
    return {
      kind: "error",
      message: `requires integer operands (got ${formatType(left)} and ${formatType(right)})`,
    };
  }
  return resolveIntegerBinaryFromInfos(leftInfo, rightInfo);
}

function resolveIntegerBinaryFromInfos(leftInfo: IntegerTypeInfo, rightInfo: IntegerTypeInfo): NumericResolution {
  const promoted = promoteIntegerInfos(leftInfo, rightInfo);
  if (promoted) {
    return { kind: "ok", type: primitiveType(promoted.name) };
  }
  const bitsNeeded =
    leftInfo.signed === rightInfo.signed
      ? Math.max(leftInfo.bits, rightInfo.bits)
      : Math.max(leftInfo.bits + 1, rightInfo.bits + 1);
  return {
    kind: "error",
    message: `operands ${leftInfo.name} and ${rightInfo.name} require ${bitsNeeded} bits, exceeding available integer widths`,
  };
}

function promoteIntegerInfos(leftInfo: IntegerTypeInfo, rightInfo: IntegerTypeInfo): IntegerTypeInfo | null {
  if (leftInfo.signed === rightInfo.signed) {
    const targetBits = Math.max(leftInfo.bits, rightInfo.bits);
    return leftInfo.signed ? findSmallestSigned(targetBits) : findSmallestUnsigned(targetBits);
  }
  const signed = leftInfo.signed ? leftInfo : rightInfo;
  const unsigned = leftInfo.signed ? rightInfo : leftInfo;
  if (signed.bits > unsigned.bits) {
    return findSmallestSigned(signed.bits) ?? null;
  }
  const bitsNeeded = Math.max(leftInfo.bits + 1, rightInfo.bits + 1);
  const signedCandidate = findSmallestSigned(bitsNeeded);
  if (signedCandidate) {
    return signedCandidate;
  }
  const unsignedFallback = widestUnsignedInfo([leftInfo, rightInfo]);
  if (unsignedFallback && unsignedFallback.bits >= Math.max(leftInfo.bits, rightInfo.bits)) {
    return unsignedFallback;
  }
  return null;
}

function extractIntegerInfo(type: TypeInfo): IntegerTypeInfo | null {
  if (type.kind === "primitive" && type.name === "char") {
    return getIntegerTypeInfo("i32") ?? null;
  }
  if (type.kind !== "primitive" || !isIntegerPrimitiveType(type)) {
    return null;
  }
  return getIntegerTypeInfo(type.name) ?? null;
}

type PrimitiveNumericClassification =
  | { kind: "float"; name: FloatPrimitive }
  | { kind: "integer"; info: IntegerTypeInfo };

function classifyNumericPrimitive(type: TypeInfo): PrimitiveNumericClassification | null {
  if (type.kind !== "primitive") {
    return null;
  }
  if (isFloatPrimitiveType(type)) {
    return { kind: "float", name: type.name };
  }
  if (type.name === "char") {
    const info = getIntegerTypeInfo("i32");
    return info ? { kind: "integer", info } : null;
  }
  if (isIntegerPrimitiveType(type)) {
    const info = getIntegerTypeInfo(type.name);
    return info ? { kind: "integer", info } : null;
  }
  return null;
}

function isStringType(type: TypeInfo): boolean {
  return type.kind === "primitive" && type.name === "string";
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
  const structBinding = structName ? ctx.lookupIdentifier(structName) : undefined;
  const definition = structName ? ctx.getStructDefinition(structName) : undefined;
  if (structName && !definition && structBinding?.kind !== "struct") {
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
      if (!ctx.isTypeAssignable(valueType, expected)) {
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
      definition: definition ?? (structBinding?.kind === "struct" ? structBinding.definition : undefined),
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
