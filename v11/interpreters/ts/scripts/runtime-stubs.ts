import type { InterpreterV10 } from "../src/interpreter";
import type { V10Value } from "../src/interpreter/values";

export function ensureConsolePrint(interpreter: InterpreterV10): void {
  if ((interpreter.globals as any).lookup && interpreter.globals.lookup("print")) {
    return;
  }
  const printFn = (interpreter as any).makeNativeFunction?.("print", 1, (_ctx: InterpreterV10, [value]: V10Value[]) => {
    if (!value) {
      console.log("nil");
      return { kind: "nil", value: null };
    }
    switch (value.kind) {
      case "string":
      case "char":
        console.log(String(value.value));
        break;
      case "bool":
        console.log(value.value ? "true" : "false");
        break;
      case "i32":
      case "f64":
        console.log(String(value.value));
        break;
      default:
        console.log(`[${value.kind}]`);
    }
    return { kind: "nil", value: null };
  });
  if (printFn) {
    interpreter.globals.define("print", printFn);
  }
}

export function installRuntimeStubs(interpreter: InterpreterV10): void {
  const channels = new Map<number, { queue: V10Value[]; capacity: number; closed: boolean }>();
  const mutexes = new Map<number, { locked: boolean }>();
  const hashers = new Map<number, number>();
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
  const defineIfMissing = (name: string, factory: () => V10Value) => {
    if (!hasGlobal(name)) {
      interpreter.globals.define(name, factory());
    }
  };

  const makeHandle = () => ({ kind: "i32", value: handleCounter++ } as V10Value);
  const toHandle = (value: V10Value | undefined): number => {
    if (!value || value.kind !== "i32") throw new Error("expected handle");
    return value.value;
  };

  const checkCancelled = (interp: InterpreterV10): boolean => {
    try {
      const result = (interp as any).procCancelled?.();
      return Boolean(result);
    } catch {
      return false;
    }
  };

  const blockOnNilChannel = (interp: InterpreterV10): V10Value | null => {
    if (checkCancelled(interp)) {
      return { kind: "nil", value: null };
    }
    interp.procYield();
    return null;
  };

  defineIfMissing("__able_channel_new", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_new", 1, (_ctx: InterpreterV10, [capacity]: V10Value[]) => {
      if (!capacity || capacity.kind !== "i32") throw new Error("capacity must be i32");
      const handle = makeHandle();
      channels.set(handle.value, { queue: [], capacity: capacity.value, closed: false });
      return handle;
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_send", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_send", 2, (ctx: InterpreterV10, [handleArg, value]: V10Value[]) => {
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
    (interpreter as any).makeNativeFunction?.("__able_channel_receive", 1, (ctx: InterpreterV10, [handleArg]: V10Value[]) => {
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
    (interpreter as any).makeNativeFunction?.("__able_channel_try_send", 2, (_ctx: InterpreterV10, [handleArg, value]: V10Value[]) => {
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
    (interpreter as any).makeNativeFunction?.("__able_channel_try_receive", 1, (_ctx: InterpreterV10, [handleArg]: V10Value[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (!channel) return { kind: "nil", value: null };
      if (channel.queue.length > 0) {
        return channel.queue.shift()!;
      }
      return channel.closed ? { kind: "nil", value: null } : { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_close", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_close", 1, (_ctx: InterpreterV10, [handleArg]: V10Value[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      if (channel) channel.closed = true;
      return { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_channel_is_closed", () =>
    (interpreter as any).makeNativeFunction?.("__able_channel_is_closed", 1, (_ctx: InterpreterV10, [handleArg]: V10Value[]) => {
      const handle = toHandle(handleArg);
      const channel = channels.get(handle);
      return { kind: "bool", value: channel ? channel.closed : false };
    }) ?? { kind: "bool", value: false },
  );

  defineIfMissing("__able_mutex_new", () => makeHandle());
  defineIfMissing("__able_mutex_lock", () =>
    (interpreter as any).makeNativeFunction?.("__able_mutex_lock", 1, (ctx: InterpreterV10, [handleArg]: V10Value[]) => {
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
    (interpreter as any).makeNativeFunction?.("__able_mutex_unlock", 1, (_ctx: InterpreterV10, [handleArg]: V10Value[]) => {
      const handle = toHandle(handleArg);
      let state = mutexes.get(handle);
      if (!state) {
        state = { locked: false };
        mutexes.set(handle, state);
      }
      state.locked = false;
      return { kind: "nil", value: null };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_string_from_builtin", () =>
    (interpreter as any).makeNativeFunction?.("__able_string_from_builtin", 1, (_ctx: InterpreterV10, [value]: V10Value[]) => {
      if (!value || value.kind !== "string") throw new Error("argument must be string");
      const bytes = textEncoder.encode(value.value);
      return { kind: "array", elements: Array.from(bytes, (b) => ({ kind: "i32", value: b })) };
    }) ?? { kind: "nil", value: null },
  );

  defineIfMissing("__able_string_to_builtin", () =>
    (interpreter as any).makeNativeFunction?.("__able_string_to_builtin", 1, (_ctx: InterpreterV10, [arr]: V10Value[]) => {
      if (!arr || arr.kind !== "array") throw new Error("argument must be array");
      const bytes = Uint8Array.from(arr.elements.map((el, idx) => {
        if (!el || (el.kind !== "i32" && el.kind !== "f64")) throw new Error(`array element ${idx} must be numeric`);
        const n = Math.trunc(el.value);
        if (n < 0 || n > 0xff) throw new Error(`array element ${idx} must be in range 0..255`);
        return n;
      }));
      return { kind: "string", value: textDecoder.decode(bytes) };
    }) ?? { kind: "string", value: "" },
  );

  defineIfMissing("__able_char_from_codepoint", () =>
    (interpreter as any).makeNativeFunction?.("__able_char_from_codepoint", 1, (_ctx: InterpreterV10, [code]: V10Value[]) => {
      if (!code || (code.kind !== "i32" && code.kind !== "f64")) throw new Error("codepoint must be numeric");
      const cp = Math.trunc(code.value);
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
      return { kind: "i32", value: handle };
    }) ?? { kind: "i32", value: 0 },
  );

  defineIfMissing("__able_hasher_write", () =>
    (interpreter as any).makeNativeFunction?.("__able_hasher_write", 2, (_ctx: InterpreterV10, [handleArg, bytesArg]: V10Value[]) => {
      const handle = toHandle(handleArg);
      if (!bytesArg || bytesArg.kind !== "string") throw new Error("bytes must be string");
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
    (interpreter as any).makeNativeFunction?.("__able_hasher_finish", 1, (_ctx: InterpreterV10, [handleArg]: V10Value[]) => {
      const handle = toHandle(handleArg);
      const state = hashers.get(handle);
      if (state === undefined) throw new Error("unknown hasher handle");
      hashers.delete(handle);
      return { kind: "i32", value: state >>> 0 };
    }) ?? { kind: "i32", value: 0 },
  );
}
