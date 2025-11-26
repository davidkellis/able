import type { V10Value } from "./values";
import { isNumericValue, numericEquals } from "./numeric";

const structName = (value: Extract<V10Value, { kind: "struct_def" | "struct_instance" }>): string | null => {
  if (value.kind === "struct_def") {
    return value.def?.id?.name ?? null;
  }
  return value.def?.id?.name ?? null;
};

const isStructInstanceEmpty = (value: Extract<V10Value, { kind: "struct_instance" }>): boolean => {
  if (Array.isArray(value.values)) {
    return value.values.length === 0;
  }
  if (value.values instanceof Map) {
    return value.values.size === 0;
  }
  return true;
};

const structInstancesEqual = (
  left: Extract<V10Value, { kind: "struct_instance" }>,
  right: Extract<V10Value, { kind: "struct_instance" }>,
): boolean => {
  const leftName = structName(left);
  const rightName = structName(right);
  if (!leftName || leftName !== rightName) {
    return false;
  }
  if (Array.isArray(left.values) || Array.isArray(right.values)) {
    if (!Array.isArray(left.values) || !Array.isArray(right.values)) return false;
    if (left.values.length !== right.values.length) return false;
    for (let i = 0; i < left.values.length; i++) {
      const lv = left.values[i];
      const rv = right.values[i];
      if (lv === undefined || rv === undefined) return false;
      if (!valuesEqual(lv, rv)) return false;
    }
    return true;
  }
  if (!(left.values instanceof Map) || !(right.values instanceof Map)) return false;
  if (left.values.size !== right.values.size) return false;
  for (const [field, lval] of left.values.entries()) {
    if (!right.values.has(field)) return false;
    const rval = right.values.get(field);
    if (rval === undefined) return false;
    if (!valuesEqual(lval, rval)) return false;
  }
  return true;
};

/**
 * Determines structural equality between two Able values for the limited
 * set of types currently used by the interpreters (strings, bools, chars,
 * nil, and numeric primitives). This mirrors the Go interpreter logic so
 * match literals and runtime equality behave consistently without relying
 * on JSON.stringify (which cannot handle BigInt payloads).
 */
export function valuesEqual(left: V10Value, right: V10Value): boolean {
  if (left.kind === "interface_value") {
    return valuesEqual(left.value, right);
  }
  if (right.kind === "interface_value") {
    return valuesEqual(left, right.value);
  }
  if (isNumericValue(left) && isNumericValue(right)) {
    return numericEquals(left, right);
  }
  switch (left.kind) {
    case "string":
      return right.kind === "string" && left.value === right.value;
    case "bool":
      return right.kind === "bool" && left.value === right.value;
    case "char":
      return right.kind === "char" && left.value === right.value;
    case "nil":
      return right.kind === "nil";
    case "struct_def": {
      if (right.kind === "struct_def") {
        const leftName = structName(left);
        const rightName = structName(right);
        return !!leftName && leftName === rightName;
      }
      if (right.kind === "struct_instance") {
        const leftName = structName(left);
        const rightName = structName(right);
        return !!leftName && leftName === rightName && isStructInstanceEmpty(right);
      }
      return false;
    }
    case "struct_instance": {
      if (right.kind === "struct_def") {
        const leftName = structName(left);
        const rightName = structName(right);
        return !!leftName && leftName === rightName && isStructInstanceEmpty(left);
      }
      if (right.kind !== "struct_instance") return false;
      return structInstancesEqual(left, right);
    }
    default:
      return false;
  }
}
