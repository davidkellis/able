import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeIntegerFromNumber, numericToNumber, isIntegerValue } from "./numeric";
import { callCallableValue } from "./functions";
import { collectTypeDispatches } from "./type-dispatch";
import { RaiseSignal } from "./signals";

export type HashMapEntry = { key: RuntimeValue; value: RuntimeValue };
type HashMapStoredEntry = HashMapEntry & { hash: bigint };
export type HashMapState = { entries: HashMapStoredEntry[] };

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
    throw new RaiseSignal(interp.makeRuntimeError(`hash map handle ${handle} is not defined`));
  }
  return state;
};

export function hashMapEntries(state: HashMapState): HashMapEntry[] {
  return state.entries;
}

export function insertHashMapEntry(
  interp: Interpreter,
  state: HashMapState,
  key: RuntimeValue,
  value: RuntimeValue,
): void {
  const hash = hashKey(interp, key);
  const index = findEntryIndex(interp, state, key, hash);
  if (index >= 0) {
    state.entries[index] = { key, value, hash };
    return;
  }
  state.entries.push({ key, value, hash });
}

function removeHashMapEntry(interp: Interpreter, state: HashMapState, key: RuntimeValue): HashMapEntry | null {
  const hash = hashKey(interp, key);
  const index = findEntryIndex(interp, state, key, hash);
  if (index < 0) return null;
  const [entry] = state.entries.splice(index, 1);
  return entry ?? null;
}

function unwrapInterface(value: RuntimeValue): RuntimeValue {
  return value.kind === "interface_value" ? value.value : value;
}

function typeNameForValue(interp: Interpreter, value: RuntimeValue): string {
  return interp.getTypeNameForValue(value) ?? value.kind;
}

function resolveInterfaceMethod(
  interp: Interpreter,
  receiver: RuntimeValue,
  interfaceName: string,
  methodName: string,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  const dispatches = collectTypeDispatches(interp, receiver);
  for (const dispatch of dispatches) {
    const method = interp.findMethod(dispatch.typeName, methodName, {
      typeArgs: dispatch.typeArgs,
      interfaceName,
      includeInherent: false,
    });
    if (method) return method;
  }
  return null;
}

function raiseKernelError(interp: Interpreter, message: string): never {
  throw new RaiseSignal(interp.makeRuntimeError(message));
}

function createKernelHasher(interp: Interpreter): RuntimeValue {
  const method = interp.findMethod("KernelHasher", "new", { includeInherent: true });
  if (!method) {
    raiseKernelError(interp, "KernelHasher.new is not available");
  }
  const result = callCallableValue(interp, method, [], interp.globals);
  if (result.kind !== "struct_instance" || result.def.id.name !== "KernelHasher") {
    raiseKernelError(interp, "KernelHasher.new returned unexpected value");
  }
  return result;
}

function finishKernelHasher(interp: Interpreter, hasher: RuntimeValue): bigint {
  const method = resolveInterfaceMethod(interp, hasher, "Hasher", "finish");
  if (!method) {
    raiseKernelError(interp, "Hasher.finish is not available for KernelHasher");
  }
  const result = callCallableValue(interp, method, [hasher], interp.globals);
  if (!isIntegerValue(result) || result.value < 0n) {
    raiseKernelError(interp, "Hasher.finish must return u64");
  }
  return result.value;
}

function hashKey(interp: Interpreter, key: RuntimeValue): bigint {
  const receiver = unwrapInterface(key);
  const method = resolveInterfaceMethod(interp, receiver, "Hash", "hash");
  if (!method) {
    const typeName = typeNameForValue(interp, receiver);
    raiseKernelError(interp, `HashMap key type ${typeName} does not implement Hash.hash`);
  }
  const hasher = createKernelHasher(interp);
  const result = callCallableValue(interp, method, [receiver, hasher], interp.globals);
  if (result.kind !== "void" && result.kind !== "nil") {
    raiseKernelError(interp, "Hash.hash must return void");
  }
  return finishKernelHasher(interp, hasher);
}

function keysEqual(interp: Interpreter, left: RuntimeValue, right: RuntimeValue): boolean {
  const receiver = unwrapInterface(left);
  const other = unwrapInterface(right);
  const method = resolveInterfaceMethod(interp, receiver, "Eq", "eq");
  if (!method) {
    const typeName = typeNameForValue(interp, receiver);
    raiseKernelError(interp, `HashMap key type ${typeName} does not implement Eq.eq`);
  }
  const result = callCallableValue(interp, method, [receiver, other], interp.globals);
  if (result.kind !== "bool") {
    raiseKernelError(interp, "Eq.eq must return bool");
  }
  return result.value;
}

function findEntryIndex(interp: Interpreter, state: HashMapState, key: RuntimeValue, hash: bigint): number {
  for (let idx = 0; idx < state.entries.length; idx++) {
    const entry = state.entries[idx];
    if (!entry || entry.hash !== hash) {
      continue;
    }
    if (keysEqual(interp, entry.key, key)) {
      return idx;
    }
  }
  return -1;
}

export function applyHashMapKernelAugmentations(cls: typeof Interpreter): void {
  cls.prototype.hashMapStateForHandle = function hashMapStateForHandleMethod(
    this: Interpreter,
    handle: number,
  ): HashMapState {
    return hashMapStateForHandle(this, handle);
  };

  cls.prototype.createHashMapHandle = function createHashMapHandle(this: Interpreter): number {
    const state: HashMapState = { entries: [] };
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
        const hash = hashKey(interp, args[1]);
        const index = findEntryIndex(interp, state, args[1], hash);
        const entry = index >= 0 ? state.entries[index] : null;
        return entry ? entry.value : NIL;
      }),
    );

    defineIfMissing("__able_hash_map_set", () =>
      this.makeNativeFunction("__able_hash_map_set", 3, (interp, args) => {
        if (args.length !== 3) throw new Error("__able_hash_map_set expects handle, key, and value");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        insertHashMapEntry(interp, state, args[1], args[2] ?? NIL);
        return NIL;
      }),
    );

    defineIfMissing("__able_hash_map_remove", () =>
      this.makeNativeFunction("__able_hash_map_remove", 2, (interp, args) => {
        if (args.length !== 2) throw new Error("__able_hash_map_remove expects handle and key");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        const entry = removeHashMapEntry(interp, state, args[1]);
        return entry ? entry.value : NIL;
      }),
    );

    defineIfMissing("__able_hash_map_contains", () =>
      this.makeNativeFunction("__able_hash_map_contains", 2, (interp, args) => {
        if (args.length !== 2) throw new Error("__able_hash_map_contains expects handle and key");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        const hash = hashKey(interp, args[1]);
        return { kind: "bool", value: findEntryIndex(interp, state, args[1], hash) >= 0 };
      }),
    );

    defineIfMissing("__able_hash_map_size", () =>
      this.makeNativeFunction("__able_hash_map_size", 1, (interp, args) => {
        if (args.length !== 1) throw new Error("__able_hash_map_size expects handle");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        return makeIntegerFromNumber("i32", state.entries.length);
      }),
    );

    defineIfMissing("__able_hash_map_clear", () =>
      this.makeNativeFunction("__able_hash_map_clear", 1, (interp, args) => {
        if (args.length !== 1) throw new Error("__able_hash_map_clear expects handle");
        const handle = toHandleNumber(args[0], "handle");
        const state = hashMapStateForHandle(interp, handle);
        state.entries.length = 0;
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
        for (const entry of source.entries) {
          target.entries.push({ ...entry });
        }
        return makeIntegerFromNumber("i64", next);
      }),
    );
  };
}
