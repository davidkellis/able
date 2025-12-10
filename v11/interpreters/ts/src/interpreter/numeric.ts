import type { FloatKind, IntegerKind, V10Value } from "./values";

type IntegerInfo = {
  kind: IntegerKind;
  bits: number;
  signed: boolean;
  min: bigint;
  max: bigint;
  mask: bigint;
};

const signedInfo = (kind: IntegerKind, bits: number): IntegerInfo => {
  const shift = BigInt(bits - 1);
  const max = (1n << shift) - 1n;
  const min = -(1n << shift);
  const mask = (1n << BigInt(bits)) - 1n;
  return { kind, bits, signed: true, min, max, mask };
};

const unsignedInfo = (kind: IntegerKind, bits: number): IntegerInfo => {
  const max = (1n << BigInt(bits)) - 1n;
  const mask = max;
  return { kind, bits, signed: false, min: 0n, max, mask };
};

const INTEGER_INFO: Record<IntegerKind, IntegerInfo> = {
  i8: signedInfo("i8", 8),
  i16: signedInfo("i16", 16),
  i32: signedInfo("i32", 32),
  i64: signedInfo("i64", 64),
  i128: signedInfo("i128", 128),
  u8: unsignedInfo("u8", 8),
  u16: unsignedInfo("u16", 16),
  u32: unsignedInfo("u32", 32),
  u64: unsignedInfo("u64", 64),
  u128: unsignedInfo("u128", 128),
};

const SIGNED_SEQUENCE: IntegerKind[] = ["i8", "i16", "i32", "i64", "i128"];
const UNSIGNED_SEQUENCE: IntegerKind[] = ["u8", "u16", "u32", "u64", "u128"];

type FloatInfo = {
  kind: FloatKind;
  apply: (value: number) => number;
};

const FLOAT_INFO: Record<FloatKind, FloatInfo> = {
  f32: { kind: "f32", apply: (value: number) => Math.fround(value) },
  f64: { kind: "f64", apply: (value: number) => value },
};

type IntegerValue = Extract<V10Value, { kind: IntegerKind }>;

export function integerKinds(): IntegerKind[] {
  return [...SIGNED_SEQUENCE, ...UNSIGNED_SEQUENCE];
}

export function getIntegerInfo(kind: IntegerKind): IntegerInfo {
  const info = INTEGER_INFO[kind];
  if (!info) throw new Error(`Unknown integer kind ${kind}`);
  return info;
}

export function makeIntegerValue(kind: IntegerKind, raw: bigint): Extract<V10Value, { kind: IntegerKind }> {
  const info = getIntegerInfo(kind);
  ensureIntegerInRange(raw, info);
  return { kind, value: raw };
}

export function makeIntegerFromNumber(kind: IntegerKind, raw: number): Extract<V10Value, { kind: IntegerKind }> {
  if (!Number.isFinite(raw)) {
    throw new Error("integer requires finite numeric value");
  }
  return makeIntegerValue(kind, BigInt(Math.trunc(raw)));
}

export function makeFloatValue(kind: FloatKind, raw: number): Extract<V10Value, { kind: FloatKind }> {
  const info = FLOAT_INFO[kind];
  if (!info) throw new Error(`Unknown float kind ${kind}`);
  return { kind, value: info.apply(raw) };
}

export function isIntegerValue(value: V10Value): value is Extract<V10Value, { kind: IntegerKind }> {
  return Object.prototype.hasOwnProperty.call(INTEGER_INFO, value.kind);
}

export function isFloatValue(value: V10Value): value is Extract<V10Value, { kind: FloatKind }> {
  return Object.prototype.hasOwnProperty.call(FLOAT_INFO, value.kind);
}

export function isNumericValue(value: V10Value): value is Extract<V10Value, { kind: IntegerKind | FloatKind }> {
  return isIntegerValue(value) || isFloatValue(value);
}

type NumericClassification =
  | { tag: "integer"; info: IntegerInfo; value: bigint }
  | { tag: "float"; kind: FloatKind; value: number };

function classifyNumeric(value: V10Value): NumericClassification | null {
  if (isIntegerValue(value)) {
    return { tag: "integer", info: getIntegerInfo(value.kind), value: value.value };
  }
  if (isFloatValue(value)) {
    return { tag: "float", kind: value.kind, value: value.value };
  }
  return null;
}

function findSmallestSigned(bits: number): IntegerInfo | null {
  for (const kind of SIGNED_SEQUENCE) {
    const info = INTEGER_INFO[kind];
    if (info.bits >= bits) {
      return info;
    }
  }
  return null;
}

function findSmallestUnsigned(bits: number): IntegerInfo | null {
  for (const kind of UNSIGNED_SEQUENCE) {
    const info = INTEGER_INFO[kind];
    if (info.bits >= bits) {
      return info;
    }
  }
  return null;
}

function widestUnsignedInfo(candidates: IntegerInfo[]): IntegerInfo | null {
  let current: IntegerInfo | null = null;
  for (const info of candidates) {
    if (info.signed) continue;
    if (!current || info.bits > current.bits) {
      current = info;
    }
  }
  return current;
}

function promoteIntegerInfos(left: IntegerInfo, right: IntegerInfo): IntegerInfo | null {
  if (left.signed === right.signed) {
    const targetBits = Math.max(left.bits, right.bits);
    return left.signed ? findSmallestSigned(targetBits) : findSmallestUnsigned(targetBits);
  }
  const bitsNeeded = Math.max(left.bits + 1, right.bits + 1);
  const signedCandidate = findSmallestSigned(bitsNeeded);
  if (signedCandidate) {
    return signedCandidate;
  }
  const unsignedFallback = widestUnsignedInfo([left, right]);
  if (unsignedFallback && unsignedFallback.bits >= Math.max(left.bits, right.bits)) {
    return unsignedFallback;
  }
  // Fall back to the widest available operand when we are already at the maximum width.
  return left.bits >= right.bits ? left : right;
}

function promoteFloatKinds(left: FloatKind, right: FloatKind): FloatKind {
  if (left === "f64" || right === "f64") return "f64";
  return "f32";
}

type NumericPromotion =
  | { tag: "integer"; info: IntegerInfo }
  | { tag: "float"; kind: FloatKind };

function promoteNumericKinds(left: NumericClassification, right: NumericClassification): NumericPromotion {
  if (left.tag === "float" || right.tag === "float") {
    const leftKind = left.tag === "float" ? left.kind : "f32";
    const rightKind = right.tag === "float" ? right.kind : "f32";
    const target = promoteFloatKinds(leftKind, rightKind);
    return { tag: "float", kind: target };
  }
  const promoted = promoteIntegerInfos(left.info, right.info);
  if (!promoted) {
    throw new Error("Arithmetic operands exceed supported integer widths");
  }
  return { tag: "integer", info: promoted };
}

function convertToFloat(value: NumericClassification, target: FloatKind): number {
  const info = FLOAT_INFO[target];
  if (!info) throw new Error(`Unknown float kind ${target}`);
  const numeric = value.tag === "float" ? value.value : Number(value.value);
  return info.apply(numeric);
}

function ensureIntegerInRange(value: bigint, info: IntegerInfo): void {
  if (value < info.min || value > info.max) {
    throw new Error("integer overflow");
  }
}

function normalizeInteger(value: NumericClassification, info: IntegerInfo): bigint {
  if (value.tag !== "integer") {
    throw new Error("Expected integer value");
  }
  ensureIntegerInRange(value.value, value.info);
  return value.value;
}

function bitPattern(value: bigint, info: IntegerInfo): bigint {
  return value & info.mask;
}

function patternToInteger(pattern: bigint, info: IntegerInfo): bigint {
  const masked = pattern & info.mask;
  if (!info.signed) {
    return masked;
  }
  const signBit = 1n << BigInt(info.bits - 1);
  if (masked & signBit) {
    return masked - (1n << BigInt(info.bits));
  }
  return masked;
}

function applyFloatOperation(op: string, left: number, right: number, kind: FloatKind): number {
  switch (op) {
    case "+":
      return FLOAT_INFO[kind].apply(left + right);
    case "-":
      return FLOAT_INFO[kind].apply(left - right);
    case "*":
      return FLOAT_INFO[kind].apply(left * right);
    case "/":
      if (right === 0) throw new Error("Division by zero");
      return FLOAT_INFO[kind].apply(left / right);
    case "^":
      return FLOAT_INFO[kind].apply(left ** right);
    default:
      throw new Error(`Unsupported arithmetic operator ${op}`);
  }
}

function applyIntegerOperation(op: string, left: bigint, right: bigint, info: IntegerInfo): bigint {
  switch (op) {
    case "+":
      return left + right;
    case "-":
      return left - right;
    case "*":
      return left * right;
    case "^": {
      if (right < 0n) {
        throw new Error("Negative integer exponent is not supported");
      }
      return left ** right;
    }
    default:
      throw new Error(`Unsupported arithmetic operator ${op}`);
  }
}

export function applyNumericUnaryMinus(value: V10Value): V10Value {
  const classified = classifyNumeric(value);
  if (!classified) throw new Error("Unary '-' requires numeric operand");
  if (classified.tag === "float") {
    return makeFloatValue(classified.kind, -classified.value);
  }
  const negated = -classified.value;
  ensureIntegerInRange(negated, classified.info);
  return { kind: classified.info.kind, value: negated };
}

export function applyBitwiseNot(value: V10Value): V10Value {
  const classified = classifyNumeric(value);
  if (!classified || classified.tag !== "integer") {
    throw new Error("Unary '.~' requires integer operand");
  }
  const pattern = bitPattern(classified.value, classified.info);
  const inverted = pattern ^ classified.info.mask;
  const result = patternToInteger(inverted, classified.info);
  ensureIntegerInRange(result, classified.info);
  return { kind: classified.info.kind, value: result };
}

export function applyArithmeticBinary(
  op: string,
  left: V10Value,
  right: V10Value,
  options?: {
    makeDivMod?: (
      kind: IntegerKind,
      parts: {
        quotient: Extract<V10Value, { kind: IntegerKind }>;
        remainder: Extract<V10Value, { kind: IntegerKind }>;
      },
    ) => V10Value;
  },
): V10Value {
  const leftClass = classifyNumeric(left);
  const rightClass = classifyNumeric(right);
  if (!leftClass || !rightClass) {
    const leftKind = (left as any)?.kind ?? "unknown";
    const rightKind = (right as any)?.kind ?? "unknown";
    throw new Error(`Arithmetic requires numeric operands (left: ${leftKind}, right: ${rightKind})`);
  }

  if (op === "/") {
    const targetKind = resolveDivisionFloatKind(leftClass, rightClass);
    const leftFloat = convertToFloat(leftClass, targetKind);
    const rightFloat = convertToFloat(rightClass, targetKind);
    const value = applyFloatOperation(op, leftFloat, rightFloat, targetKind);
    return makeFloatValue(targetKind, value);
  }

  if (op === "//" || op === "%" || op === "/%") {
    if (leftClass.tag !== "integer" || rightClass.tag !== "integer") {
      const leftKind = (left as any)?.kind ?? "unknown";
      const rightKind = (right as any)?.kind ?? "unknown";
      throw new Error(`'${op}' requires integer operands (left: ${leftKind}, right: ${rightKind})`);
    }
    const divMod = computeDivMod(leftClass, rightClass);
    if (op === "//") return divMod.quotient;
    if (op === "%") return divMod.remainder;
    if (!options?.makeDivMod) {
      throw new Error("DivMod factory not provided for '/%'");
    }
    return options.makeDivMod(divMod.kind, { quotient: divMod.quotient, remainder: divMod.remainder });
  }

  if (op === "^") {
    const promotion = promoteNumericKinds(leftClass, rightClass);
    if (promotion.tag === "float") {
      const leftFloat = convertToFloat(leftClass, promotion.kind);
      const rightFloat = convertToFloat(rightClass, promotion.kind);
      const value = applyFloatOperation(op, leftFloat, rightFloat, promotion.kind);
      return makeFloatValue(promotion.kind, value);
    }
    const base = leftClass.tag === "integer" ? leftClass.value : BigInt(Math.trunc(leftClass.value));
    const exp = rightClass.tag === "integer" ? rightClass.value : BigInt(Math.trunc(rightClass.value));
    if (exp < 0n) {
      throw new Error("Negative integer exponent is not supported");
    }
    const result = applyIntegerOperation(op, base, exp, promotion.info);
    ensureIntegerInRange(result, promotion.info);
    return { kind: promotion.info.kind, value: result };
  }

  if (op === "+" || op === "-" || op === "*") {
    const promotion = promoteNumericKinds(leftClass, rightClass);
    if (promotion.tag === "float") {
      const leftFloat = convertToFloat(leftClass, promotion.kind);
      const rightFloat = convertToFloat(rightClass, promotion.kind);
      const value = applyFloatOperation(op, leftFloat, rightFloat, promotion.kind);
      return makeFloatValue(promotion.kind, value);
    }
    const leftValue = leftClass.tag === "integer" ? leftClass.value : BigInt(Math.trunc(leftClass.value));
    const rightValue = rightClass.tag === "integer" ? rightClass.value : BigInt(Math.trunc(rightClass.value));
    const result = applyIntegerOperation(op, leftValue, rightValue, promotion.info);
    ensureIntegerInRange(result, promotion.info);
    return { kind: promotion.info.kind, value: result };
  }

  throw new Error(`Unsupported arithmetic operator ${op}`);
}

function resolveDivisionFloatKind(left: NumericClassification, right: NumericClassification): FloatKind {
  if (left.tag === "float" || right.tag === "float") {
    const leftKind = left.tag === "float" ? left.kind : "f32";
    const rightKind = right.tag === "float" ? right.kind : "f32";
    return promoteFloatKinds(leftKind, rightKind);
  }
  return "f64";
}

function computeDivMod(
  left: Extract<NumericClassification, { tag: "integer" }>,
  right: Extract<NumericClassification, { tag: "integer" }>,
): { kind: IntegerKind; quotient: IntegerValue; remainder: IntegerValue } {
  const promotion = promoteIntegerInfos(left.info, right.info);
  if (!promotion) {
    throw new Error("Arithmetic operands exceed supported integer widths");
  }
  if (right.value === 0n) {
    throw new Error("Division by zero");
  }
  const { quotient, remainder } = euclideanDivMod(left.value, right.value);
  ensureIntegerInRange(quotient, promotion);
  ensureIntegerInRange(remainder, promotion);
  return {
    kind: promotion.kind,
    quotient: makeIntegerValue(promotion.kind, quotient),
    remainder: makeIntegerValue(promotion.kind, remainder),
  };
}

function euclideanDivMod(dividend: bigint, divisor: bigint): { quotient: bigint; remainder: bigint } {
  if (divisor === 0n) {
    throw new Error("Division by zero");
  }
  let quotient = dividend / divisor;
  let remainder = dividend % divisor;
  if (remainder < 0n) {
    if (divisor > 0n) {
      quotient -= 1n;
      remainder += divisor;
    } else {
      quotient += 1n;
      remainder -= divisor;
    }
  }
  return { quotient, remainder };
}

export function applyComparisonBinary(op: string, left: V10Value, right: V10Value): V10Value {
  const leftClass = classifyNumeric(left);
  const rightClass = classifyNumeric(right);
  if (!leftClass || !rightClass) {
    throw new Error("Arithmetic requires numeric operands");
  }
  if (op === "==" || op === "!=") {
    const equal = numericEquals(left, right);
    return { kind: "bool", value: op === "==" ? equal : !equal };
  }
  const promotion = promoteNumericKinds(leftClass, rightClass);
  let comparison = 0;
  if (promotion.tag === "float") {
    const leftFloat = convertToFloat(leftClass, promotion.kind);
    const rightFloat = convertToFloat(rightClass, promotion.kind);
    if (leftFloat < rightFloat) comparison = -1;
    else if (leftFloat > rightFloat) comparison = 1;
  } else {
    const leftValue = leftClass.tag === "integer" ? leftClass.value : BigInt(Math.trunc(leftClass.value));
    const rightValue = rightClass.tag === "integer" ? rightClass.value : BigInt(Math.trunc(rightClass.value));
    if (leftValue < rightValue) comparison = -1;
    else if (leftValue > rightValue) comparison = 1;
  }
  let result = false;
  switch (op) {
    case "<":
      result = comparison < 0;
      break;
    case "<=":
      result = comparison <= 0;
      break;
    case ">":
      result = comparison > 0;
      break;
    case ">=":
      result = comparison >= 0;
      break;
    default:
      throw new Error(`Unsupported comparison operator ${op}`);
  }
  return { kind: "bool", value: result };
}

export function numericEquals(left: V10Value, right: V10Value): boolean {
  const leftClass = classifyNumeric(left);
  const rightClass = classifyNumeric(right);
  if (!leftClass || !rightClass) {
    return false;
  }
  if (leftClass.tag === "integer" && rightClass.tag === "integer") {
    return leftClass.value === rightClass.value;
  }
  const promotion = promoteNumericKinds(leftClass, rightClass);
  if (promotion.tag !== "float") {
    return leftClass.value === (rightClass.tag === "integer" ? rightClass.value : BigInt(Math.trunc(rightClass.value)));
  }
  const leftFloat = convertToFloat(leftClass, promotion.kind);
  const rightFloat = convertToFloat(rightClass, promotion.kind);
  return Object.is(leftFloat, rightFloat);
}

export function applyBitwiseBinary(op: string, left: V10Value, right: V10Value): V10Value {
  const normalized = op.startsWith(".") ? op.slice(1) : op;
  const leftClass = classifyNumeric(left);
  const rightClass = classifyNumeric(right);
  if (!leftClass || !rightClass || leftClass.tag !== "integer" || rightClass.tag !== "integer") {
    const leftLabel = `${left.kind}:${"value" in left ? String((left as any).value) : ""}`;
    const rightLabel = `${right.kind}:${"value" in right ? String((right as any).value) : ""}`;
    throw new Error(`Bitwise requires integer operands for '${op}' (left: ${leftLabel}, right: ${rightLabel})`);
  }
  const promotion = promoteIntegerInfos(leftClass.info, rightClass.info);
  if (!promotion) {
    throw new Error("Bitwise operands exceed supported widths");
  }
  if (normalized === "<<" || normalized === ">>") {
    return applyShift(normalized, leftClass.value, rightClass.value, promotion);
  }
  const leftPattern = bitPattern(leftClass.value, promotion);
  const rightPattern = bitPattern(rightClass.value, promotion);
  let resultPattern: bigint;
  switch (normalized) {
    case "&":
      resultPattern = leftPattern & rightPattern;
      break;
    case "|":
      resultPattern = leftPattern | rightPattern;
      break;
    case "^":
      resultPattern = leftPattern ^ rightPattern;
      break;
    default:
      throw new Error(`Unsupported bitwise operator ${op}`);
  }
  const result = patternToInteger(resultPattern, promotion);
  ensureIntegerInRange(result, promotion);
  return { kind: promotion.kind, value: result };
}

function applyShift(op: string, left: bigint, right: bigint, info: IntegerInfo): Extract<V10Value, { kind: IntegerKind }> {
  if (right < 0n || right >= BigInt(info.bits)) {
    throw new Error("shift out of range");
  }
  let result: bigint;
  if (op === "<<") {
    result = left << right;
  } else {
    if (info.signed) {
      result = left >> right;
    } else {
      const normalized = bitPattern(left, info);
      result = normalized >> right;
    }
  }
  ensureIntegerInRange(result, info);
  return { kind: info.kind, value: result };
}

export function numericToNumber(value: V10Value, label: string, options?: { requireSafeInteger?: boolean }): number {
  if (isFloatValue(value)) {
    if (!Number.isFinite(value.value)) {
      throw new Error(`${label} must be finite`);
    }
    return value.value;
  }
  if (isIntegerValue(value)) {
    const num = Number(value.value);
    if (!Number.isFinite(num)) {
      throw new Error(`${label} is too large for numeric conversion`);
    }
    if (options?.requireSafeInteger && !Number.isSafeInteger(num)) {
      throw new Error(`${label} exceeds supported numeric range`);
    }
    return num;
  }
  throw new Error(`${label} must be numeric`);
}
