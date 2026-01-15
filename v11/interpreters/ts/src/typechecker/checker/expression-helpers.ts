import type * as AST from "../../ast";
import {
  formatType,
  primitiveType,
  unknownType,
  isFloatPrimitiveType,
  isIntegerPrimitiveType,
  isNumeric,
  isRatioType,
  type TypeInfo,
  type FloatPrimitive,
} from "../types";
import {
  findSmallestSigned,
  findSmallestUnsigned,
  getIntegerTypeInfo,
  widestUnsignedInfo,
  type IntegerTypeInfo,
} from "../numeric";
import type { ExpressionContext } from "./expression-context";

export function mergeMapComponent(
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

export type NumericResolution =
  | { kind: "ok"; type: TypeInfo }
  | { kind: "unknown" }
  | { kind: "error"; message: string };

export function applyNumericResolution(
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

export function resolveNumericBinaryType(left: TypeInfo, right: TypeInfo): NumericResolution {
  if (!left || left.kind === "unknown" || !right || right.kind === "unknown") {
    return { kind: "unknown" };
  }
  if (isRatioType(left) || isRatioType(right)) {
    if (!isNumeric(left) || !isNumeric(right)) {
      return {
        kind: "error",
        message: `requires numeric operands (got ${formatType(left)} and ${formatType(right)})`,
      };
    }
    return { kind: "ok", type: { kind: "struct", name: "Ratio", typeArguments: [] } };
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

export function resolveDivisionType(left: TypeInfo, right: TypeInfo): NumericResolution {
  if (!left || left.kind === "unknown" || !right || right.kind === "unknown") {
    return { kind: "unknown" };
  }
  if (isRatioType(left) || isRatioType(right)) {
    if (!isNumeric(left) || !isNumeric(right)) {
      return {
        kind: "error",
        message: `requires numeric operands (got ${formatType(left)} and ${formatType(right)})`,
      };
    }
    return { kind: "ok", type: { kind: "struct", name: "Ratio", typeArguments: [] } };
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
  return { kind: "ok", type: primitiveType("f64") };
}

export function resolveIntegerBinaryType(left: TypeInfo, right: TypeInfo): NumericResolution {
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

export function resolveIntegerBinaryFromInfos(leftInfo: IntegerTypeInfo, rightInfo: IntegerTypeInfo): NumericResolution {
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

export function promoteIntegerInfos(leftInfo: IntegerTypeInfo, rightInfo: IntegerTypeInfo): IntegerTypeInfo | null {
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

export function extractIntegerInfo(type: TypeInfo): IntegerTypeInfo | null {
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

export function classifyNumericPrimitive(type: TypeInfo): PrimitiveNumericClassification | null {
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

export function isStringType(type: TypeInfo): boolean {
  return type.kind === "primitive" && type.name === "String";
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
      if (statement.type === "ReturnStatement") {
        ctx.checkStatement(statement as AST.Statement);
        resultType = unknownType;
        break;
      }
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

export function resolveArrayLiteralElementType(
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

export function checkStructLiteral(ctx: ExpressionContext, literal: AST.StructLiteral): TypeInfo {
  if (!literal) {
    return unknownType;
  }
  const structName = ctx.getIdentifierName(literal.structType);
  let typeArguments = Array.isArray(literal.typeArguments)
    ? literal.typeArguments.map((arg) => ctx.resolveTypeExpression(arg))
    : [];
  const structBinding = structName ? ctx.lookupIdentifier(structName) : undefined;
  const definition = structName ? ctx.getStructDefinition(structName) : undefined;
  if (definition?.genericParams?.length && typeArguments.length === 0) {
    typeArguments = definition.genericParams.map(() => unknownType);
  }
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

export function buildStructTypeSubstitution(
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

export function resolveStructFieldTypes(
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

export function resolveRangeElementType(ctx: ExpressionContext, start: TypeInfo, end: TypeInfo): TypeInfo {
  if (start && start.kind !== "unknown" && isIntegerPrimitiveType(start)) {
    return start;
  }
  if (end && end.kind !== "unknown" && isIntegerPrimitiveType(end)) {
    return end;
  }
  return primitiveType("i32");
}
