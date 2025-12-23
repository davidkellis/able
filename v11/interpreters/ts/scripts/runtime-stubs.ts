import type { Interpreter } from "../src/interpreter";
import type { RuntimeValue } from "../src/interpreter/values";
import { isFloatValue, isIntegerValue, makeIntegerValue, numericToNumber } from "../src/interpreter/numeric";

export function ensureConsolePrint(interpreter: Interpreter): void {
  if ((interpreter.globals as any).lookup && interpreter.globals.lookup("print")) {
    return;
  }
  const printFn = (interpreter as any).makeNativeFunction?.(
    "print",
    1,
    (_ctx: Interpreter, [value]: RuntimeValue[]) => {
      if (!value) {
        console.log("nil");
        return { kind: "nil", value: null };
      }
      if (isIntegerValue(value) || isFloatValue(value)) {
        console.log(String(value.value));
      } else {
        switch (value.kind) {
          case "String":
          case "char":
            console.log(String(value.value));
            break;
          case "bool":
            console.log(value.value ? "true" : "false");
            break;
          default:
            console.log(`[${value.kind}]`);
        }
      }
      return { kind: "nil", value: null };
    },
  );
  if (printFn) {
    interpreter.globals.define("print", printFn);
  }
}

export function installRuntimeStubs(interpreter: Interpreter): void {
  const channels = new Map<number, { queue: RuntimeValue[]; capacity: number; closed: boolean }>();
  const mutexes = new Map<number, { locked: boolean }>();
  const hashers = new Map<number, number>();
  const arrays = new Map<number, { values: RuntimeValue[]; capacity: number }>();
  let handleCounter = 1;
  const textEncoder = new TextEncoder();
  const textDecoder = new TextDecoder();

  const hasGlobal = (name: string) => {
    try {
      interpreter.globals.get(name);
      return true;
    } catch {
      return false;
    }
  };
  const defineIfMissing = (name: string, factory: () => RuntimeValue) => {
    if (!hasGlobal(name)) {
      interpreter.globals.define(name, factory());
    }
  };

  const makeHandle = (): RuntimeValue => makeIntegerValue("i32", BigInt(handleCounter++));
  const toHandle = (value: RuntimeValue | undefined): number => {
    if (!value) throw new Error("expected handle");
    return Math.trunc(numericToNumber(value, "handle", { requireSafeInteger: true }));
  };
  const nilValue: RuntimeValue = { kind: "nil", value: null };

  const ensureArrayState = (handle: number): { values: RuntimeValue[]; capacity: number } => {
    const existing = arrays.get(handle);
    if (existing) return existing;
    const state = { values: [] as RuntimeValue[], capacity: 0 };
    arrays.set(handle, state);
    return state;
  };

  const ensureArrayCapacity = (state: { values: RuntimeValue[]; capacity: number }, minimum: number): number => {
    if (minimum <= state.capacity) return state.capacity;
    const next = Math.max(minimum, state.capacity > 0 ? state.capacity * 2 : 4);
    state.capacity = next;
    return state.capacity;
  };

  const setArrayLength = (state: { values: RuntimeValue[]; capacity: number }, length: number): void => {
    if (length < state.values.length) {
      state.values.length = length;
      return;
    }
    while (state.values.length < length) {
      state.values.push(nilValue);
    }
  };

  const checkCancelled = (interp: Interpreter): boolean => {
    try {
      const result = (interp as any).procCancelled?.();
      return Boolean(result);
    } catch {
      return false;
    }
  };

  const blockOnNilChannel = (interp: Interpreter): RuntimeValue | null => {
    if (checkCancelled(interp)) {
      return { kind: "nil", value: null };
    }
    interp.procYield();
    return null;
  };

  const awaitableDef = { id: { name: "ChannelAwaitable" } } as any;
  const awaitRegistrationDef = { id: { name: "AwaitRegistration" } } as any;

  const makeAwaitRegistration = (cancel?: () => void): RuntimeValue => {
    const inst: any = { kind: "struct_instance", def: awaitRegistrationDef, values: new Map<string, RuntimeValue>() };
    const cancelNative =
      (interpreter as any).makeNativeFunction?.("AwaitRegistration.cancel", 1, () => {
        if (cancel) cancel();
        return { kind: "nil", value: null };
      }) ?? { kind: "native_function", name: "AwaitRegistration.cancel", arity: 1, impl: () => ({ kind: "nil", value: null }) };
    const bound = (interpreter as any).bindNativeMethod?.(cancelNative, inst) ?? cancelNative;
    inst.values.set("cancel", bound);
    return inst;
  };

  const makeAwaitable = (handle: number, op: "send" | "receive", payload: RuntimeValue | null, callback?: RuntimeValue): RuntimeValue => {
    const inst: any = { kind: "struct_instance", def: awaitableDef, values: new Map<string, RuntimeValue>() };
    const isReady =
      (interpreter as any).makeNativeFunction?.("Awaitable.is_ready", 1, () => {
        const channel = channels.get(handle);
        if (!channel) return { kind: "bool", value: false };
        if (op === "receive") {
          const ready = channel.queue.length > 0 || (channel.capacity === 0 && channel.queue.length > 0) || channel.closed;
          return { kind: "bool", value: ready };
        }
        if (channel.closed) {
          throw new Error("send on closed channel");
        }
        const ready = channel.capacity === 0 ? true : channel.queue.length < channel.capacity;
        return { kind: "bool", value: ready };
      }) ?? { kind: "native_function", name: "Awaitable.is_ready", arity: 1, impl: () => ({ kind: "bool", value: true }) };
    const register =
      (interpreter as any).makeNativeFunction?.("Awaitable.register", 2, () => makeAwaitRegistration()) ??
      ({ kind: "native_function", name: "Awaitable.register", arity: 2, impl: () => makeAwaitRegistration() } as any);
    const commit =
      (interpreter as any).makeNativeFunction?.("Awaitable.commit", 1, (_ctx: Interpreter) => {
        const channel = channels.get(handle);
        if (!channel) return { kind: "nil", value: null };
        if (op === "receive") {
          const value = channel.queue.shift() ?? { kind: "nil", value: null };
          if (callback) {
            const fn = callback as any;
            if (fn.impl) {
              fn.impl(interpreter, [value]);
            }
          }
          return value;
        }
        if (channel.closed) {
          throw new Error("send on closed channel");
        }
        if (op === "send") {
          if (channel.capacity === 0 || channel.queue.length < channel.capacity) {
            channel.queue.push(payload ?? { kind: "nil", value: null });
          }
          if (callback) {
            const fn = callback as any;
            if (fn.impl) fn.impl(interpreter, []);
          }
        }
        return { kind: "nil", value: null };
      }) ?? { kind: "native_function", name: "Awaitable.commit", arity: 1, impl: () => ({ kind: "nil", value: null }) };
    const isDefault =
      (interpreter as any).makeNativeFunction?.("Awaitable.is_default", 1, () => ({ kind: "bool", value: false })) ??
      ({ kind: "native_function", name: "Awaitable.is_default", arity: 1, impl: () => ({ kind: "bool", value: false }) } as any);
    const values = inst.values as Map<string, RuntimeValue>;
    values.set("is_ready", (interpreter as any).bindNativeMethod?.(isReady, inst) ?? isReady);
    values.set("register", (interpreter as any).bindNativeMethod?.(register, inst) ?? register);
    values.set("commit", (interpreter as any).bindNativeMethod?.(commit, inst) ?? commit);
    values.set("is_default", (interpreter as any).bindNativeMethod?.(isDefault, inst) ?? isDefault);
    return inst;
  };

  defineIfMissing("__able_array_new", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_new", 0, () => {
      const handle = makeHandle();
      arrays.set(toHandle(handle), { values: [], capacity: 0 });
      return handle;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_with_capacity", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_with_capacity", 1, (_ctx: Interpreter, [capacity]: RuntimeValue[]) => {
      const cap = capacity ? Math.max(0, Math.trunc(numericToNumber(capacity, "capacity", { requireSafeInteger: true }))) : 0;
      const handle = makeHandle();
      arrays.set(toHandle(handle), { values: [], capacity: cap });
      return handle;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_size", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_size", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const state = ensureArrayState(toHandle(handleArg));
      return makeIntegerValue("u64", BigInt(state.values.length));
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_capacity", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_capacity", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const state = ensureArrayState(toHandle(handleArg));
      return makeIntegerValue("u64", BigInt(state.capacity));
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_set_len", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_set_len", 2, (_ctx: Interpreter, [handleArg, lenVal]: RuntimeValue[]) => {
      const length = Math.max(0, Math.trunc(numericToNumber(lenVal ?? { kind: "nil", value: null }, "length", { requireSafeInteger: true })));
      const state = ensureArrayState(toHandle(handleArg));
      ensureArrayCapacity(state, length);
      setArrayLength(state, length);
      return nilValue;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_read", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_read", 2, (_ctx: Interpreter, [handleArg, idxVal]: RuntimeValue[]) => {
      const idx = Math.trunc(numericToNumber(idxVal ?? { kind: "nil", value: null }, "index", { requireSafeInteger: true }));
      const state = ensureArrayState(toHandle(handleArg));
      if (idx < 0 || idx >= state.values.length) return nilValue;
      return state.values[idx] ?? nilValue;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_write", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_write", 3, (_ctx: Interpreter, [handleArg, idxVal, value]: RuntimeValue[]) => {
      const idx = Math.trunc(numericToNumber(idxVal ?? { kind: "nil", value: null }, "index", { requireSafeInteger: true }));
      const state = ensureArrayState(toHandle(handleArg));
      if (idx < 0) throw new Error("index must be non-negative");
      ensureArrayCapacity(state, idx + 1);
      setArrayLength(state, Math.max(state.values.length, idx + 1));
      state.values[idx] = value ?? nilValue;
      return nilValue;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_reserve", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_reserve", 2, (_ctx: Interpreter, [handleArg, capVal]: RuntimeValue[]) => {
      const capacity = Math.max(0, Math.trunc(numericToNumber(capVal ?? { kind: "nil", value: null }, "capacity", { requireSafeInteger: true })));
      const state = ensureArrayState(toHandle(handleArg));
      const cap = ensureArrayCapacity(state, capacity);
      return makeIntegerValue("u64", BigInt(cap));
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_array_clone", () =>
    (interpreter as any).makeNativeFunction?.("__able_array_clone", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const source = ensureArrayState(toHandle(handleArg));
      const handle = makeHandle();
      arrays.set(toHandle(handle), { values: source.values.slice(), capacity: source.capacity });
      return handle;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_new", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_new", 1, (_ctx: Interpreter, [capacity]: RuntimeValue[]) => {
      const capCount = capacity ? Math.max(0, Math.trunc(numericToNumber(capacity, "capacity", { requireSafeInteger: true }))) : 0;
      const handle = makeHandle();
      const handleId = Number(handle.value);
      channels.set(handleId, { queue: [], capacity: capCount, closed: false });
      return handle;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_send", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_send", 2, (ctx: Interpreter, [handleArg, value]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (!channel) return blockOnNilChannel(ctx);
      if (channel.closed) {
        throw new Error("send on closed channel");
      }
      if (channel.capacity === 0) {
        channel.queue = [value];
        ctx.procYield();
        return { kind: "nil", value: null };
      }
      if (channel.queue.length < channel.capacity) {
        channel.queue.push(value);
        return { kind: "nil", value: null };
      }
      while (channel.queue.length >= channel.capacity) {
        if (checkCancelled(ctx)) {
          return { kind: "nil", value: null };
        }
        ctx.procYield();
      }
      channel.queue.push(value);
      return { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_receive", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_receive", 1, (ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (!channel) return blockOnNilChannel(ctx);
      while (channel.queue.length === 0) {
        if (channel.closed) {
          return { kind: "nil", value: null };
        }
        if (checkCancelled(ctx)) {
          return { kind: "nil", value: null };
        }
        ctx.procYield();
      }
      return channel.queue.shift() ?? { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_try_send", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_try_send", 2, (_ctx: Interpreter, [handleArg, value]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (!channel) return { kind: "bool", value: false };
      if (channel.closed) throw new Error("send on closed channel");
      if (channel.capacity === 0) {
        channel.queue = [value];
        return { kind: "bool", value: true };
      }
      if (channel.queue.length < channel.capacity) {
        channel.queue.push(value);
        return { kind: "bool", value: true };
      }
      return { kind: "bool", value: false };
    }) ?? { kind: "bool", value: false },
  );

  defineIfMissing("__able_channel_try_receive", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_try_receive", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (!channel) return { kind: "nil", value: null };
      if (channel.queue.length > 0) {
        return channel.queue.shift()!;
      }
      return channel.closed ? { kind: "nil", value: null } : { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_await_try_recv", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_await_try_recv", 2, (_ctx: Interpreter, [handleArg, cb]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      return makeAwaitable(handle, "receive", null, cb);
    }) ?? makeAwaitable(0, "receive", null),
  );

  defineIfMissing("__able_channel_await_try_send", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_await_try_send", 3, (_ctx: Interpreter, [handleArg, value, cb]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      return makeAwaitable(handle, "send", value, cb);
    }) ?? makeAwaitable(0, "send", { kind: "nil", value: null }),
  );

  defineIfMissing("__able_channel_close", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_close", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (channel) channel.closed = true;
      return { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_is_closed", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_is_closed", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      return { kind: "bool", value: channel ? channel.closed : false };
    }) ?? { kind: "bool", value: false },
  );

  defineIfMissing("__able_mutex_new", () => makeHandle());
  defineIfMissing("__able_mutex_lock", () =>
    (interpreter as any).makeNativeFunction?.("__able_mutex_lock", 1, (ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      let state = mutexes.get(handle);
      if (!state) {
        state = { locked: false };
        mutexes.set(handle, state);
      }
      if (!state.locked) {
        state.locked = true;
        return { kind: "nil", value: null };
      }
      if (checkCancelled(ctx)) {
        return { kind: "nil", value: null };
      }
      ctx.procYield();
      return null;
    }) ?? { kind: "nil", value: null },
  );
  defineIfMissing("__able_mutex_unlock", () =>
    (interpreter as any).makeNativeFunction?.("__able_mutex_unlock", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      let state = mutexes.get(handle);
      if (!state) {
        state = { locked: false };
        mutexes.set(handle, state);
      }
      if (!state.locked) {
        throw new Error("unlock of unlocked mutex");
      }
      state.locked = false;
      return { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_String_from_builtin", () =>
    (interpreter as any).makeNativeFunction?.("__able_String_from_builtin", 1, (_ctx: Interpreter, [value]: RuntimeValue[]) => {
      if (!value || value.kind !== "String") throw new Error("argument must be String");
      const bytes = textEncoder.encode(value.value);
      return { kind: "array", elements: Array.from(bytes, (b) => makeIntegerValue("i32", BigInt(b))) };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_String_to_builtin", () =>
    (interpreter as any).makeNativeFunction?.("__able_String_to_builtin", 1, (_ctx: Interpreter, [arr]: RuntimeValue[]) => {
      if (!arr || arr.kind !== "array") throw new Error("argument must be array");
      const bytes = Uint8Array.from(arr.elements.map((el, idx) => {
        if (!el) throw new Error(`array element ${idx} must be numeric`);
        const n = Math.trunc(numericToNumber(el, `array element ${idx}`, { requireSafeInteger: true }));
        if (n < 0 || n > 0xff) throw new Error(`array element ${idx} must be in range 0..255`);
        return n;
      }));
      return { kind: "String", value: textDecoder.decode(bytes) };
    }) ?? { kind: "String", value: "" },
  );

  defineIfMissing("__able_char_from_codepoint", () =>
    (interpreter as any).makeNativeFunction?.("__able_char_from_codepoint", 1, (_ctx: Interpreter, [code]: RuntimeValue[]) => {
      if (!code) throw new Error("codepoint must be numeric");
      const cp = Math.trunc(numericToNumber(code, "codepoint", { requireSafeInteger: true }));
      if (cp < 0 || cp > 0x10ffff) throw new Error("codepoint out of range");
      return { kind: "char", value: String.fromCodePoint(cp) };
    }) ?? { kind: "char", value: "" },
  );

  const FNV_OFFSET = 0x811c9dc5;
  const FNV_PRIME = 0x01000193;

  defineIfMissing("__able_hasher_create", () =>
    (interpreter as any).makeNativeFunction?.("__able_hasher_create", 0, () => {
      const handle = handleCounter++;
      hashers.set(handle, FNV_OFFSET);
      return makeIntegerValue("i32", BigInt(handle));
    }) ?? makeIntegerValue("i32", 0n),
  );

  defineIfMissing("__able_hasher_write", () =>
    (interpreter as any).makeNativeFunction?.("__able_hasher_write", 2, (_ctx: Interpreter, [handleArg, bytesArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      if (!bytesArg || bytesArg.kind !== "String") throw new Error("bytes must be String");
      const state = hashers.get(handle);
      if (state === undefined) throw new Error("unknown hasher handle");
      let hash = state >>> 0;
      for (const b of textEncoder.encode(bytesArg.value)) {
        hash ^= b;
        hash = Math.imul(hash, FNV_PRIME) >>> 0;
      }
      hashers.set(handle, hash >>> 0);
      return { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_hasher_finish", () =>
    (interpreter as any).makeNativeFunction?.("__able_hasher_finish", 1, (_ctx: Interpreter, [handleArg]: RuntimeValue[]) => {
      const handle = toHandle(handleArg);
      const state = hashers.get(handle);
      if (state === undefined) throw new Error("unknown hasher handle");
      hashers.delete(handle);
      return makeIntegerValue("i32", BigInt(state >>> 0));
    }) ?? makeIntegerValue("i32", 0n),
  );
}
