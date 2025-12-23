import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeIntegerFromNumber, numericToNumber, isFloatValue, isIntegerValue } from "./numeric";
import { callCallableValue } from "./functions";

export type HashMapEntry = { key: RuntimeValue; value: RuntimeValue };
export type HashMapState = { entries: Map<string, HashMapEntry>; order: string[] };

const NIL: Extract<RuntimeValue, { kind: "nil" }> = { kind: "nil", value: null };

declare module "./index" {
  interface Interpreter {
    hashMapBuiltinsInitialized: boolean;
    nextHashMapHandle: number;
    hashMapStates: Map<number, HashMapState>;
    ensureHashMapKernelBuiltins(): void;
    hashMapStateForHandle(handle: number): HashMapState;
    createHashMapHandle(): number;
  }
}

const toHandleNumber = (value: RuntimeValue, label: string): number =>
  Math.trunc(numericToNumber(value, label, { requireSafeInteger: true }));

const hashMapStateForHandle = (interp: Interpreter, handle: number): HashMapState => {
  const state = interp.hashMapStates.get(handle);
  if (!state) {
    throw new Error(`hash map handle ${handle} is not defined`);
  }
  return state;
};

export function hashMapEntries(state: HashMapState): HashMapEntry[] {
  return state.order.map((label) => {
    const entry = state.entries.get(label);
    if (!entry) {
      throw new Error("HashMap entry missing during serialization");
    }
    return entry;
  });
}

export function insertHashMapEntry(state: HashMapState, key: RuntimeValue, value: RuntimeValue): void {
  const label = keyLabel(key);
  if (!state.entries.has(label)) {
    state.order.push(label);
  }
  state.entries.set(label, { key, value });
}

function removeHashMapEntry(state: HashMapState, key: RuntimeValue): HashMapEntry | null {
  const label = keyLabel(key);
  const entry = state.entries.get(label) ?? null;
  if (!entry) return null;
  state.entries.delete(label);
  const idx = state.order.indexOf(label);
  if (idx >= 0) {
    state.order.splice(idx, 1);
  }
  return entry;
}

function keyLabel(value: RuntimeValue): string {
  switch (value.kind) {
    case "String":
      return `s:${value.value}`;
    case "bool":
      return `b:${value.value ? 1 : 0}`;
    case "char":
      return `c:${value.value}`;
    case "nil":
      return "n:";
    default:
      if (isIntegerValue(value)) {
        return `i:${value.value.toString()}`;
      }
      if (isFloatValue(value)) {
        if (!Number.isFinite(value.value)) {
          throw new Error("HashMap keys cannot be NaN or infinite numbers");
        }
        return `f:${value.value}`;
      }
      throw new Error("HashMap keys must be primitives (String, bool, char, nil, numeric)");
  }
}

export function applyHashMapKernelAugmentations(cls: typeof Interpreter): void {
  cls.prototype.hashMapStateForHandle = function hashMapStateForHandleMethod(
    this: Interpreter,
    handle: number,
  ): HashMapState {
    return hashMapStateForHandle(this, handle);
  };

  cls.prototype.createHashMapHandle = function createHashMapHandle(this: Interpreter): number {
    const state: HashMapState = { entries: new Map(), order: [] };
    const handle = this.nextHashMapHandle++;
    this.hashMapStates.set(handle, state);
    return handle;
  };

  cls.prototype.ensureHashMapKernelBuiltins = function ensureHashMapKernelBuiltins(this: Interpreter): void {
    if (this.hashMapBuiltinsInitialized) return;
    this.hashMapBuiltinsInitialized = true;

    const defineIfMissing = (name: string, factory: () => Extract<RuntimeValue, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        this.globals.define(name, factory());
      }
    };

    defineIfMissing("__able_hash_map_new", () =>
      this.makeNativeFunction("__able_hash_map_new", 0, (interp, args) => {
        if (args.length !== 0) throw new Error("__able_hash_map_new expects no arguments");
        const handle = interp.createHashMapHandle();
        return makeIntegerFromNumber("i64", handle);
      }),
    );

    defineIfMissing("__able_hash_map_with_capacity", () =>
      this.makeNativeFunction("__able_hash_map_with_capacity", 1, (interp, args) => {
        const _capacity = args[0]
          ? Math.trunc(numericToNumber(args[0], "capacity", { requireSafeInteger: true }))
          : 0;
        const handle = interp.createHashMapHandle();
        return makeIntegerFromNumber("i64", handle);
      }),
    );

    defineIfMissing("__able_hash_map_get", () =>
      this.makeNativeFunction("__able_hash_map_get", 2, (interp, args) => {
        if (args.length !== 2) throw new Error("__able_hash_map_get expects handle and key");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        const entry = state.entries.get(keyLabel(args[1]));
        return entry ? entry.value : NIL;
      }),
    );

    defineIfMissing("__able_hash_map_set", () =>
      this.makeNativeFunction("__able_hash_map_set", 3, (interp, args) => {
        if (args.length !== 3) throw new Error("__able_hash_map_set expects handle, key, and value");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        insertHashMapEntry(state, args[1], args[2] ?? NIL);
        return NIL;
      }),
    );

    defineIfMissing("__able_hash_map_remove", () =>
      this.makeNativeFunction("__able_hash_map_remove", 2, (interp, args) => {
        if (args.length !== 2) throw new Error("__able_hash_map_remove expects handle and key");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        const entry = removeHashMapEntry(state, args[1]);
        return entry ? entry.value : NIL;
      }),
    );

    defineIfMissing("__able_hash_map_contains", () =>
      this.makeNativeFunction("__able_hash_map_contains", 2, (interp, args) => {
        if (args.length !== 2) throw new Error("__able_hash_map_contains expects handle and key");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        return { kind: "bool", value: state.entries.has(keyLabel(args[1])) };
      }),
    );

    defineIfMissing("__able_hash_map_size", () =>
      this.makeNativeFunction("__able_hash_map_size", 1, (interp, args) => {
        if (args.length !== 1) throw new Error("__able_hash_map_size expects handle");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        return makeIntegerFromNumber("i32", state.entries.size);
      }),
    );

    defineIfMissing("__able_hash_map_clear", () =>
      this.makeNativeFunction("__able_hash_map_clear", 1, (interp, args) => {
        if (args.length !== 1) throw new Error("__able_hash_map_clear expects handle");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        state.entries.clear();
        state.order.length = 0;
        return NIL;
      }),
    );

    defineIfMissing("__able_hash_map_for_each", () =>
      this.makeNativeFunction("__able_hash_map_for_each", 2, (interp, args) => {
        if (args.length !== 2) throw new Error("__able_hash_map_for_each expects handle and callback");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        const callback = args[1];
        for (const entry of hashMapEntries(state)) {
          callCallableValue(interp, callback, [entry.key, entry.value], interp.globals);
        }
        return NIL;
      }),
    );

    defineIfMissing("__able_hash_map_clone", () =>
      this.makeNativeFunction("__able_hash_map_clone", 1, (interp, args) => {
        if (args.length !== 1) throw new Error("__able_hash_map_clone expects handle");
        const handle = toHandleNumber(args[0], "handle");
        const source = hashMapStateForHandle(interp, handle);
        const next = interp.createHashMapHandle();
        const target = hashMapStateForHandle(interp, next);
        for (const entry of hashMapEntries(source)) {
          insertHashMapEntry(target, entry.key, entry.value);
        }
        return makeIntegerFromNumber("i64", next);
      }),
    );
  };
}
