import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeIntegerFromNumber, numericToNumber } from "./numeric";

type ArrayState = { values: RuntimeValue[]; capacity: number };

const NIL: Extract<RuntimeValue, { kind: "nil" }> = { kind: "nil", value: null };

declare module "./index" {
  interface Interpreter {
    arrayBuiltinsInitialized: boolean;
    nextArrayHandle: number;
    arrayStates: Map<number, ArrayState>;
    ensureArrayKernelBuiltins(): void;
    ensureArrayState(value: Extract<RuntimeValue, { kind: "array" }>, capacityHint?: number): ArrayState;
    makeArrayValue(elements?: RuntimeValue[], capacityHint?: number): Extract<RuntimeValue, { kind: "array" }>;
  }
}

const toHandleNumber = (value: RuntimeValue, label: string): number =>
  Math.trunc(numericToNumber(value, label, { requireSafeInteger: true }));

const arrayStateForHandle = (interp: Interpreter, handle: number): ArrayState => {
  const state = interp.arrayStates.get(handle);
  if (!state) {
    throw new Error(`array handle ${handle} is not defined`);
  }
  return state;
};

const defineHandle = (value: Extract<RuntimeValue, { kind: "array" }>, handle: number): void => {
  Object.defineProperty(value, "handle", {
    value: handle,
    writable: true,
    configurable: true,
    enumerable: false,
  });
};

const clampCapacity = (requested: number, fallback: number): number => {
  if (!Number.isFinite(requested) || requested < 0) return fallback;
  return requested;
};

const ensureCapacity = (state: ArrayState, minimum: number): number => {
  if (minimum <= state.capacity) {
    return state.capacity;
  }
  const next = Math.max(minimum, state.capacity > 0 ? state.capacity * 2 : 4);
  state.capacity = next;
  return state.capacity;
};

const setLength = (state: ArrayState, length: number): void => {
  if (length < 0) throw new Error("array length must be non-negative");
  if (length < state.values.length) {
    state.values.length = length;
    return;
  }
  while (state.values.length < length) {
    state.values.push(NIL);
  }
};

export function applyArrayKernelAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureArrayState = function ensureArrayState(
    this: Interpreter,
    value: Extract<RuntimeValue, { kind: "array" }>,
    capacityHint?: number,
  ): ArrayState {
    if (!value || value.kind !== "array") throw new Error("array state requires array value");
    if (value.handle && this.arrayStates.has(value.handle)) {
      const state = this.arrayStates.get(value.handle)!;
      if (value.elements !== state.values) {
        value.elements = state.values;
      }
      defineHandle(value, value.handle);
      return state;
    }
    const normalizedCapacity = Math.max(value.elements.length, clampCapacity(capacityHint ?? value.elements.length, value.elements.length));
    const state: ArrayState = { values: value.elements, capacity: normalizedCapacity };
    const handle = this.nextArrayHandle++;
    this.arrayStates.set(handle, state);
    defineHandle(value, handle);
    return state;
  };

  cls.prototype.makeArrayValue = function makeArrayValue(
    this: Interpreter,
    elements: RuntimeValue[] = [],
    capacityHint?: number,
  ): Extract<RuntimeValue, { kind: "array" }> {
    const backing = elements.slice();
    const arr: Extract<RuntimeValue, { kind: "array" }> = { kind: "array", elements: backing };
    this.ensureArrayState(arr, capacityHint);
    return arr;
  };

  cls.prototype.ensureArrayKernelBuiltins = function ensureArrayKernelBuiltins(this: Interpreter): void {
    if (this.arrayBuiltinsInitialized) return;
    this.arrayBuiltinsInitialized = true;

    const defineIfMissing = (name: string, factory: () => Extract<RuntimeValue, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        this.globals.define(name, factory());
      }
    };

    defineIfMissing("__able_array_new", () =>
      this.makeNativeFunction("__able_array_new", 0, (interp) => {
        const arr = interp.makeArrayValue();
        return makeIntegerFromNumber("i64", arr.handle ?? 0);
      }),
    );

    defineIfMissing("__able_array_with_capacity", () =>
      this.makeNativeFunction("__able_array_with_capacity", 1, (interp, args) => {
        const requested = args[0] ? Math.trunc(numericToNumber(args[0], "capacity", { requireSafeInteger: true })) : 0;
        const arr = interp.makeArrayValue([], requested);
        return makeIntegerFromNumber("i64", arr.handle ?? 0);
      }),
    );

    defineIfMissing("__able_array_size", () =>
      this.makeNativeFunction("__able_array_size", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const state = arrayStateForHandle(interp, handle);
        return makeIntegerFromNumber("i32", state.values.length);
      }),
    );

    defineIfMissing("__able_array_capacity", () =>
      this.makeNativeFunction("__able_array_capacity", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const state = arrayStateForHandle(interp, handle);
        return makeIntegerFromNumber("i32", state.capacity);
      }),
    );

    defineIfMissing("__able_array_set_len", () =>
      this.makeNativeFunction("__able_array_set_len", 2, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const length = Math.trunc(numericToNumber(args[1], "length", { requireSafeInteger: true }));
        const state = arrayStateForHandle(interp, handle);
        ensureCapacity(state, length);
        setLength(state, length);
        return NIL;
      }),
    );

    defineIfMissing("__able_array_read", () =>
      this.makeNativeFunction("__able_array_read", 2, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const idx = Math.trunc(numericToNumber(args[1], "index", { requireSafeInteger: true }));
        const state = arrayStateForHandle(interp, handle);
        if (idx < 0 || idx >= state.values.length) return NIL;
        const value = state.values[idx];
        return value ?? NIL;
      }),
    );

    defineIfMissing("__able_array_write", () =>
      this.makeNativeFunction("__able_array_write", 3, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const idx = Math.trunc(numericToNumber(args[1], "index", { requireSafeInteger: true }));
        const value = args[2] ?? NIL;
        const state = arrayStateForHandle(interp, handle);
        if (idx < 0) throw new Error("index must be non-negative");
        ensureCapacity(state, idx + 1);
        setLength(state, Math.max(state.values.length, idx + 1));
        state.values[idx] = value;
        return NIL;
      }),
    );

    defineIfMissing("__able_array_reserve", () =>
      this.makeNativeFunction("__able_array_reserve", 2, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const minCapacity = Math.trunc(numericToNumber(args[1], "capacity", { requireSafeInteger: true }));
        const state = arrayStateForHandle(interp, handle);
        const capacity = ensureCapacity(state, minCapacity);
        return makeIntegerFromNumber("i32", capacity);
      }),
    );

    defineIfMissing("__able_array_clone", () =>
      this.makeNativeFunction("__able_array_clone", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "handle");
        const source = arrayStateForHandle(interp, handle);
        const cloneValues = source.values.slice();
        const arr = interp.makeArrayValue(cloneValues, source.capacity);
        return makeIntegerFromNumber("i64", arr.handle ?? 0);
      }),
    );
  };
}

export type { ArrayState };
