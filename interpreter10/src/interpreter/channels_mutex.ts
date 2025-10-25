import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

type ProcHandleValue = Extract<V10Value, { kind: "proc_handle" }>;
type BoolValue = Extract<V10Value, { kind: "bool" }>;
type NilValue = Extract<V10Value, { kind: "nil" }>;

interface ChannelState {
  id: number;
  capacity: number;
  queue: V10Value[];
  closed: boolean;
}

interface MutexState {
  id: number;
  locked: boolean;
  owner: ProcHandleValue | null;
  waiters: ProcHandleValue[];
}

const NIL: NilValue = { kind: "nil", value: null };

declare module "./index" {
  interface InterpreterV10 {
    ensureChannelMutexBuiltins(): void;
    channelMutexBuiltinsInitialized: boolean;
    nextChannelHandle: number;
    channelStates: Map<number, ChannelState>;
    nextMutexHandle: number;
    mutexStates: Map<number, MutexState>;
  }
}

declare module "./values" {
  interface ProcHandleValue {
    waitingMutex?: MutexState;
  }
}

function isNumeric(value: V10Value): value is Extract<V10Value, { kind: "i32" | "f64" }> {
  return value.kind === "i32" || value.kind === "f64";
}

function toHandleNumber(value: V10Value, label: string): number {
  if (!isNumeric(value)) {
    throw new Error(`${label} must be numeric`);
  }
  return Math.trunc(value.value);
}

function blockOnNilChannel(interp: InterpreterV10): V10Value {
  const ctx = interp.currentAsyncContext();
  if (!ctx || ctx.kind !== "proc") {
    throw new Error("Nil channel operations must occur inside a proc");
  }
  const cancelled = interp.procCancelled() as BoolValue;
  if (cancelled.value) {
    return NIL;
  }
  interp.procYield();
  return NIL;
}

function requireProcContext(interp: InterpreterV10, action: string): ProcHandleValue {
  const ctx = interp.currentAsyncContext();
  if (!ctx || ctx.kind !== "proc") {
    throw new Error(`${action} must occur inside a proc`);
  }
  return ctx.handle;
}

export function applyChannelMutexAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.ensureChannelMutexBuiltins = function ensureChannelMutexBuiltins(this: InterpreterV10): void {
    if (this.channelMutexBuiltinsInitialized) return;
    this.channelMutexBuiltinsInitialized = true;

    if (!this.channelStates) this.channelStates = new Map();
    if (!this.mutexStates) this.mutexStates = new Map();

    const defineIfMissing = (name: string, factory: () => Extract<V10Value, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        // not defined, continue
      }
      this.globals.define(name, factory());
    };

    defineIfMissing("__able_channel_new", () =>
      this.makeNativeFunction("__able_channel_new", 1, (interp, args) => {
        const capacity = args[0] ? Math.max(0, Math.trunc(toHandleNumber(args[0], "capacity"))) : 0;
        const handle = interp.nextChannelHandle++;
        interp.channelStates.set(handle, { id: handle, capacity, queue: [], closed: false });
        return { kind: "i32", value: handle };
      }),
    );

    defineIfMissing("__able_channel_send", () =>
      this.makeNativeFunction("__able_channel_send", 2, (interp, args) => {
        const handleValue = args[0];
        const payload = args[1];
        const handle = toHandleNumber(handleValue, "channel handle");
        if (handle === 0) {
          return blockOnNilChannel(interp);
        }
        const state = interp.channelStates.get(handle);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.closed) {
          throw new Error("send on closed channel");
        }
        if (state.capacity > 0 && state.queue.length >= state.capacity) {
          // Drop oldest item to keep behaviour deterministic but non-blocking.
          state.queue.shift();
        }
        state.queue.push(payload);
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_receive", () =>
      this.makeNativeFunction("__able_channel_receive", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "channel handle");
        if (handle === 0) {
          return blockOnNilChannel(interp);
        }
        const state = interp.channelStates.get(handle);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.queue.length > 0) {
          return state.queue.shift()!;
        }
        if (state.closed) {
          return NIL;
        }
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_try_send", () =>
      this.makeNativeFunction("__able_channel_try_send", 2, (interp, args) => {
        const handle = toHandleNumber(args[0], "channel handle");
        const payload = args[1];
        if (handle === 0) {
          return { kind: "bool", value: false };
        }
        const state = interp.channelStates.get(handle);
        if (!state || state.closed) {
          return { kind: "bool", value: false };
        }
        if (state.capacity > 0 && state.queue.length >= state.capacity) {
          return { kind: "bool", value: false };
        }
        state.queue.push(payload);
        return { kind: "bool", value: true };
      }),
    );

    defineIfMissing("__able_channel_try_receive", () =>
      this.makeNativeFunction("__able_channel_try_receive", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "channel handle");
        if (handle === 0) {
          return NIL;
        }
        const state = interp.channelStates.get(handle);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.queue.length > 0) {
          return state.queue.shift()!;
        }
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_close", () =>
      this.makeNativeFunction("__able_channel_close", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "channel handle");
        if (handle === 0) {
          throw new Error("close of nil channel");
        }
        const state = interp.channelStates.get(handle);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.closed) {
          throw new Error("close of closed channel");
        }
        state.closed = true;
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_is_closed", () =>
      this.makeNativeFunction("__able_channel_is_closed", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "channel handle");
        if (handle === 0) {
          return { kind: "bool", value: false };
        }
        const state = interp.channelStates.get(handle);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        return { kind: "bool", value: state.closed };
      }),
    );

    defineIfMissing("__able_mutex_new", () =>
      this.makeNativeFunction("__able_mutex_new", 0, (interp) => {
        const handle = interp.nextMutexHandle++;
        interp.mutexStates.set(handle, { id: handle, locked: false, owner: null, waiters: [] });
        return { kind: "i32", value: handle };
      }),
    );

    defineIfMissing("__able_mutex_lock", () =>
      this.makeNativeFunction("__able_mutex_lock", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "mutex handle");
        const state = interp.mutexStates.get(handle);
        if (!state) {
          throw new Error("Invalid mutex handle");
        }

        const procHandle = (() => {
          const ctx = interp.currentAsyncContext();
          return ctx && ctx.kind === "proc" ? ctx.handle : null;
        })();

        if (state.locked) {
          if (!procHandle) {
            throw new Error("Mutex already locked");
          }
          if (!state.waiters.includes(procHandle)) {
            state.waiters.push(procHandle);
          }
          (procHandle as any).waitingMutex = state;
          interp.procYield();
          return NIL;
        }

        state.locked = true;
        state.owner = procHandle ?? null;
        return NIL;
      }),
    );

    defineIfMissing("__able_mutex_unlock", () =>
      this.makeNativeFunction("__able_mutex_unlock", 1, (interp, args) => {
        const handle = toHandleNumber(args[0], "mutex handle");
        const state = interp.mutexStates.get(handle);
        if (!state) {
          throw new Error("Invalid mutex handle");
        }
        if (!state.locked) {
          return NIL;
        }

        state.locked = false;
        state.owner = null;
        if (state.waiters.length > 0) {
          const next = state.waiters.shift()!;
          if ((next as any).waitingMutex === state) {
            delete (next as any).waitingMutex;
          }
          if (!next.runner) {
            next.runner = () => interp.runProcHandle(next);
          }
          interp.scheduleAsync(next.runner);
        }

        return NIL;
      }),
    );
  };
}
