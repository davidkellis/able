import * as AST from "../ast";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

export type StandardRuntimeErrorKind = "DivisionByZeroError" | "OverflowError" | "ShiftOutOfRangeError";

type StandardRuntimeErrorDetails = {
  operation?: string;
  shift?: bigint;
};

export class StandardRuntimeError extends Error {
  kind: StandardRuntimeErrorKind;
  operation?: string;
  shift?: bigint;

  constructor(kind: StandardRuntimeErrorKind, message: string, details: StandardRuntimeErrorDetails = {}) {
    super(message);
    this.kind = kind;
    this.operation = details.operation;
    this.shift = details.shift;
  }
}

const I32_MIN = -(1n << 31n);
const I32_MAX = (1n << 31n) - 1n;

function clampI32(value: bigint): bigint {
  if (value < I32_MIN || value > I32_MAX) return 0n;
  return value;
}

function resolveStandardErrorStruct(interp: Interpreter, name: string): AST.StructDefinition {
  const cached = interp.standardErrorStructs.get(name);
  if (cached) return cached;
  const candidates = [
    name,
    `able.core.errors.${name}`,
    `core.errors.${name}`,
    `errors.${name}`,
  ];
  for (const candidate of candidates) {
    try {
      const val = interp.globals.get(candidate);
      if (val && val.kind === "struct_def") {
        interp.standardErrorStructs.set(name, val.def);
        return val.def;
      }
    } catch {
      // ignore lookup errors
    }
  }
  for (const bucket of interp.packageRegistry.values()) {
    const val = bucket.get(name);
    if (val && val.kind === "struct_def") {
      interp.standardErrorStructs.set(name, val.def);
      return val.def;
    }
  }
  const placeholder = AST.structDefinition(name, [], "named");
  interp.standardErrorStructs.set(name, placeholder);
  return placeholder;
}

export function makeStandardErrorValue(
  interp: Interpreter,
  err: StandardRuntimeError,
): Extract<RuntimeValue, { kind: "error" }> {
  const def = resolveStandardErrorStruct(interp, err.kind);
  const entries: Array<[string, RuntimeValue]> = [];
  if (err.kind === "OverflowError") {
    const operation = err.operation ?? err.message;
    entries.push(["operation", { kind: "String", value: operation }]);
  }
  if (err.kind === "ShiftOutOfRangeError") {
    const shift = clampI32(err.shift ?? 0n);
    entries.push(["shift", { kind: "i32", value: shift }]);
  }
  const instance = interp.makeNamedStructInstance(def, entries) as Extract<RuntimeValue, { kind: "struct_instance" }>;
  return interp.makeRuntimeError(err.message, instance) as Extract<RuntimeValue, { kind: "error" }>;
}
