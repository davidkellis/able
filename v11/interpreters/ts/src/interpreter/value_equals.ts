import type { V10Value } from "./values";
import { isNumericValue, numericEquals } from "./numeric";

/**
 * Determines structural equality between two Able values for the limited
 * set of types currently used by the interpreters (strings, bools, chars,
 * nil, and numeric primitives). This mirrors the Go interpreter logic so
 * match literals and runtime equality behave consistently without relying
 * on JSON.stringify (which cannot handle BigInt payloads).
 */
export function valuesEqual(left: V10Value, right: V10Value): boolean {
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
    default:
      return false;
  }
}
