import * as AST from "./ast";
import type { Environment } from "./environment";

export type AblePrimitive =
  | { kind: "i8"; value: number }
  | { kind: "i16"; value: number }
  | { kind: "i32"; value: number }
  | { kind: "i64"; value: bigint }
  | { kind: "i128"; value: bigint }
  | { kind: "u8"; value: number }
  | { kind: "u16"; value: number }
  | { kind: "u32"; value: number }
  | { kind: "u64"; value: bigint }
  | { kind: "u128"; value: bigint }
  | { kind: "f32"; value: number }
  | { kind: "f64"; value: number }
  | { kind: "string"; value: string }
  | { kind: "bool"; value: boolean }
  | { kind: "char"; value: string }
  | { kind: "nil"; value: null }
  | { kind: "void"; value: undefined };

export interface AbleFunction {
  kind: "function";
  node: AST.FunctionDefinition | AST.LambdaExpression;
  closureEnv: Environment;
  isBoundMethod?: boolean;
}

export interface AbleStructDefinition {
  kind: "struct_definition";
  name: string;
  definitionNode: AST.StructDefinition;
}

export interface AbleStructInstance {
  kind: "struct_instance";
  definition: AbleStructDefinition;
  values: AbleValue[] | Map<string, AbleValue>;
}

export interface AbleUnionDefinition {
  kind: "union_definition";
  name: string;
  definitionNode: AST.UnionDefinition;
}

export interface AbleInterfaceDefinition {
  kind: "interface_definition";
  name: string;
  definitionNode: AST.InterfaceDefinition;
}

export interface AbleImplementationDefinition {
  kind: "implementation_definition";
  implNode: AST.ImplementationDefinition;
  interfaceDef: AbleInterfaceDefinition;
  methods: Map<string, AbleFunction>;
  closureEnv: Environment;
}

export interface AbleMethodsCollection {
  kind: "methods_collection";
  methodsNode: AST.MethodsDefinition;
  methods: Map<string, AbleFunction>;
  closureEnv: Environment;
}

export interface AbleError {
  kind: "error";
  message: string;
  originalValue?: any;
}

export interface AbleProcHandle {
  kind: "proc_handle";
  id: number;
}

export interface AbleThunk {
  kind: "thunk";
  id: number;
}

export interface AbleArray {
  kind: "array";
  elements: AbleValue[];
}

export interface AbleRange {
  kind: "range";
  start: number | bigint;
  end: number | bigint;
  inclusive: boolean;
}

export interface AbleIterator {
  kind: "AbleIterator";
  next: () => AbleValue | typeof IteratorEnd;
}

export type AbleValue =
  | AblePrimitive
  | AbleFunction
  | AbleStructDefinition
  | AbleStructInstance
  | AbleUnionDefinition
  | AbleInterfaceDefinition
  | AbleImplementationDefinition
  | AbleMethodsCollection
  | AbleError
  | AbleProcHandle
  | AbleThunk
  | AbleArray
  | AbleRange
  | AbleIterator;

export function createError(message: string, originalValue?: any): AbleError {
  return { kind: "error", message, originalValue };
}

export function hasKind<K extends string>(value: any, kind: K): value is { kind: K } {
  return value !== null && typeof value === "object" && value.kind === kind;
}

export function isAblePrimitive(value: AbleValue): value is AblePrimitive {
  return (
    value !== null &&
    typeof value === "object" &&
    "value" in value &&
    (value.kind === "i8" ||
      value.kind === "i16" ||
      value.kind === "i32" ||
      value.kind === "i64" ||
      value.kind === "i128" ||
      value.kind === "u8" ||
      value.kind === "u16" ||
      value.kind === "u32" ||
      value.kind === "u64" ||
      value.kind === "u128" ||
      value.kind === "f32" ||
      value.kind === "f64" ||
      value.kind === "string" ||
      value.kind === "bool" ||
      value.kind === "char" ||
      value.kind === "nil" ||
      value.kind === "void")
  );
}

export function isAbleFunction(value: AbleValue): value is AbleFunction {
  return hasKind(value, "function");
}

export function isAbleStructInstance(value: AbleValue): value is AbleStructInstance {
  return hasKind(value, "struct_instance");
}

export function isAbleStructDefinition(value: AbleValue): value is AbleStructDefinition {
  return hasKind(value, "struct_definition");
}

export function isAbleArray(value: AbleValue): value is AbleArray {
  return hasKind(value, "array");
}

export function isAbleRange(value: AbleValue): value is AbleRange {
  return hasKind(value, "range");
}

export const IteratorEnd: AblePrimitive = { kind: "nil", value: null };

export function createArrayIterator(array: AbleArray): AbleIterator {
  let index = 0;
  return {
    kind: "AbleIterator",
    next: (): AbleValue | typeof IteratorEnd => {
      if (index < array.elements.length) {
        const value = array.elements[index];
        index++;
        if (value === undefined) {
          throw new Error("Internal Interpreter Error: Undefined element found in array during iteration.");
        }
        return value;
      } else {
        return IteratorEnd;
      }
    },
  };
}

export function createRangeIterator(range: AbleRange): AbleIterator {
  let current = range.start;
  const end = range.end;
  const inclusive = range.inclusive;
  const step = typeof current === "bigint" ? (current <= end ? 1n : -1n) : current <= end ? 1 : -1;

  return {
    kind: "AbleIterator",
    next: () => {
      let done: boolean;
      if (typeof current === "bigint" && typeof end === "bigint") {
        done = step > 0n ? (inclusive ? current > end : current >= end) : inclusive ? current < end : current <= end;
      } else if (typeof current === "number" && typeof end === "number") {
        done = step > 0 ? (inclusive ? current > end : current >= end) : inclusive ? current < end : current <= end;
      } else {
        throw new Error("Interpreter Error: Mismatched types in range iterator.");
      }

      if (done) {
        return IteratorEnd;
      } else {
        const valueToReturn = current;
        if (typeof current === "bigint" && typeof step === "bigint") {
          current = current + step;
        } else if (typeof current === "number" && typeof step === "number") {
          current = current + step;
        }
        let valueKind: AblePrimitive["kind"] = "i32";
        if (typeof valueToReturn === "bigint") {
          valueKind = "i64";
        } else if (typeof valueToReturn === "number") {
          valueKind = "i32";
        }

        return { kind: valueKind, value: valueToReturn } as AblePrimitive;
      }
    },
  };
}
