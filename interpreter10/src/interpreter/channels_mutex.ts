import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

type ProcHandleValue = Extract<V10Value, { kind: "proc_handle" }>;
type BoolValue = Extract<V10Value, { kind: "bool" }>;
type NilValue = Extract<V10Value, { kind: "nil" }>;

interface ChannelSendWaiter {
  handle: ProcHandleValue;
  value: V10Value;
}

interface ChannelReceiveWaiter {
  handle: ProcHandleValue;
}

interface ChannelState {
  id: number;
  capacity: number;
  queue: V10Value[];
  closed: boolean;
  sendWaiters: ChannelSendWaiter[];
  receiveWaiters: ChannelReceiveWaiter[];
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
    waitingChannelSend?: {
      state: ChannelState;
      value: V10Value;
      delivered?: boolean;
      error?: string;
    };
    waitingChannelReceive?: {
      state: ChannelState;
      ready?: boolean;
      value?: V10Value;
      closed?: boolean;
      error?: string;
    };
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

function scheduleProc(interp: InterpreterV10, handle: ProcHandleValue): void {
  if (!handle.runner) {
    handle.runner = () => interp.runProcHandle(handle);
  }
  interp.scheduleAsync(handle.runner);
}

function getChannelState(interp: InterpreterV10, handle: number): ChannelState | undefined {
  const state = interp.channelStates.get(handle);
  if (state) {
    if (!state.sendWaiters) {
      state.sendWaiters = [];
    }
    if (!state.receiveWaiters) {
      state.receiveWaiters = [];
    }
  }
  return state;
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
        interp.channelStates.set(handle, {
          id: handle,
          capacity,
          queue: [],
          closed: false,
          sendWaiters: [],
          receiveWaiters: [],
        });
        return { kind: "i32", value: handle };
      }),
    );

    defineIfMissing("__able_channel_send", () =>
      this.makeNativeFunction("__able_channel_send", 2, (interp, args) => {
        const handleValue = args[0];
        const incomingPayload = args[1];
        const handleNumber = toHandleNumber(handleValue, "channel handle");
        if (handleNumber === 0) {
          return blockOnNilChannel(interp);
        }
        const state = getChannelState(interp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }

        const ctx = interp.currentAsyncContext();
        const procHandle = ctx && ctx.kind === "proc" ? ctx.handle : null;
        let payload = incomingPayload;
        const pending = procHandle ? (procHandle as any).waitingChannelSend : undefined;

        if (pending && pending.state !== state) {
          delete (procHandle as any).waitingChannelSend;
        }

        if (pending && pending.state === state) {
          payload = pending.value;
          if (pending.error) {
            delete (procHandle as any).waitingChannelSend;
            throw new Error(pending.error);
          }
          if (pending.delivered) {
            delete (procHandle as any).waitingChannelSend;
            return NIL;
          }
        }

        if (procHandle && procHandle.cancelRequested) {
          state.sendWaiters = state.sendWaiters.filter((entry) => entry.handle !== procHandle);
          delete (procHandle as any).waitingChannelSend;
          throw new Error("Proc cancelled");
        }

        if (state.closed) {
          throw new Error("send on closed channel");
        }

        if (state.receiveWaiters.length > 0) {
          const receiver = state.receiveWaiters.shift()!;
          const receiverHandle = receiver.handle;
          const receiverPending = (receiverHandle as any).waitingChannelReceive;
          if (receiverPending && receiverPending.state === state) {
            receiverPending.ready = true;
            receiverPending.value = payload;
          }
          scheduleProc(interp, receiverHandle);
          if (procHandle) {
            delete (procHandle as any).waitingChannelSend;
          }
          return NIL;
        }

        if (state.capacity > 0 && state.queue.length < state.capacity) {
          state.queue.push(payload);
          if (procHandle) {
            delete (procHandle as any).waitingChannelSend;
          }
          return NIL;
        }

        if (!procHandle) {
          throw new Error("Channel send would block outside of proc context");
        }

        const existing = (procHandle as any).waitingChannelSend;
        if (!existing || existing.state !== state) {
          (procHandle as any).waitingChannelSend = { state, value: payload };
        } else {
          existing.value = payload;
          existing.delivered = false;
          existing.error = undefined;
        }
        if (!state.sendWaiters.some((entry) => entry.handle === procHandle)) {
          state.sendWaiters.push({ handle: procHandle, value: payload });
        } else {
          for (const entry of state.sendWaiters) {
            if (entry.handle === procHandle) {
              entry.value = payload;
              break;
            }
          }
        }
        interp.procYield();
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_receive", () =>
      this.makeNativeFunction("__able_channel_receive", 1, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        if (handleNumber === 0) {
          return blockOnNilChannel(interp);
        }
        const state = getChannelState(interp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }

        const ctx = interp.currentAsyncContext();
        const procHandle = ctx && ctx.kind === "proc" ? ctx.handle : null;
        const pending = procHandle ? (procHandle as any).waitingChannelReceive : undefined;

        if (pending && pending.state !== state) {
          delete (procHandle as any).waitingChannelReceive;
        }

        if (pending && pending.state === state) {
          if (pending.error) {
            delete (procHandle as any).waitingChannelReceive;
            throw new Error(pending.error);
          }
          if (pending.ready) {
            const result = pending.closed ? NIL : pending.value ?? NIL;
            delete (procHandle as any).waitingChannelReceive;
            return result;
          }
        }

        if (procHandle && procHandle.cancelRequested) {
          state.receiveWaiters = state.receiveWaiters.filter((entry) => entry.handle !== procHandle);
          delete (procHandle as any).waitingChannelReceive;
          throw new Error("Proc cancelled");
        }

        if (state.queue.length > 0) {
          const value = state.queue.shift()!;
          if (state.capacity > 0 && state.sendWaiters.length > 0) {
            const nextSender = state.sendWaiters.shift()!;
            const senderPending = (nextSender.handle as any).waitingChannelSend;
            state.queue.push(nextSender.value);
            if (senderPending && senderPending.state === state) {
              senderPending.delivered = true;
            }
            scheduleProc(interp, nextSender.handle);
          }
          return value ?? NIL;
        }

        if (state.sendWaiters.length > 0) {
          const sender = state.sendWaiters.shift()!;
          const senderPending = (sender.handle as any).waitingChannelSend;
          if (senderPending && senderPending.state === state) {
            senderPending.delivered = true;
          }
          scheduleProc(interp, sender.handle);
          return sender.value ?? NIL;
        }

        if (state.closed) {
          return NIL;
        }

        if (!procHandle) {
          throw new Error("Channel receive would block outside of proc context");
        }

        const existing = (procHandle as any).waitingChannelReceive;
        if (!existing || existing.state !== state) {
          (procHandle as any).waitingChannelReceive = { state };
        } else {
          existing.ready = false;
          existing.closed = false;
          existing.value = undefined;
        }
        if (!state.receiveWaiters.some((entry) => entry.handle === procHandle)) {
          state.receiveWaiters.push({ handle: procHandle });
        }
        interp.procYield();
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_try_send", () =>
      this.makeNativeFunction("__able_channel_try_send", 2, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        const payload = args[1];
        if (handleNumber === 0) {
          return { kind: "bool", value: false };
        }
        const state = getChannelState(interp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.closed) {
          throw new Error("send on closed channel");
        }
        if (state.receiveWaiters.length > 0) {
          const receiver = state.receiveWaiters.shift()!;
          const receiverPending = (receiver.handle as any).waitingChannelReceive;
          if (receiverPending && receiverPending.state === state) {
            receiverPending.ready = true;
            receiverPending.value = payload;
          }
          scheduleProc(interp, receiver.handle);
          return { kind: "bool", value: true };
        }
        if (state.capacity > 0 && state.queue.length < state.capacity) {
          state.queue.push(payload);
          return { kind: "bool", value: true };
        }
        return { kind: "bool", value: false };
      }),
    );

    defineIfMissing("__able_channel_try_receive", () =>
      this.makeNativeFunction("__able_channel_try_receive", 1, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        if (handleNumber === 0) {
          return NIL;
        }
        const state = getChannelState(interp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.queue.length > 0) {
          const value = state.queue.shift()!;
          if (state.capacity > 0 && state.sendWaiters.length > 0) {
            const nextSender = state.sendWaiters.shift()!;
            const senderPending = (nextSender.handle as any).waitingChannelSend;
            state.queue.push(nextSender.value);
            if (senderPending && senderPending.state === state) {
              senderPending.delivered = true;
            }
            scheduleProc(interp, nextSender.handle);
          }
          return value ?? NIL;
        }
        if (state.sendWaiters.length > 0) {
          const sender = state.sendWaiters.shift()!;
          const senderPending = (sender.handle as any).waitingChannelSend;
          if (senderPending && senderPending.state === state) {
            senderPending.delivered = true;
          }
          scheduleProc(interp, sender.handle);
          return sender.value ?? NIL;
        }
        if (state.closed) {
          return NIL;
        }
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_close", () =>
      this.makeNativeFunction("__able_channel_close", 1, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        if (handleNumber === 0) {
          throw new Error("close of nil channel");
        }
        const state = getChannelState(interp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.closed) {
          throw new Error("close of closed channel");
        }
        state.closed = true;

        while (state.receiveWaiters.length > 0) {
          const receiver = state.receiveWaiters.shift()!;
          const pending = (receiver.handle as any).waitingChannelReceive;
          if (pending && pending.state === state) {
            pending.ready = true;
            pending.closed = true;
            pending.value = undefined;
          }
          scheduleProc(interp, receiver.handle);
        }

        while (state.sendWaiters.length > 0) {
          const sender = state.sendWaiters.shift()!;
          const pending = (sender.handle as any).waitingChannelSend;
          if (pending && pending.state === state) {
            pending.error = "send on closed channel";
          }
          scheduleProc(interp, sender.handle);
        }

        return NIL;
      }),
    );

    defineIfMissing("__able_channel_is_closed", () =>
      this.makeNativeFunction("__able_channel_is_closed", 1, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        if (handleNumber === 0) {
          return { kind: "bool", value: false };
        }
        const state = getChannelState(interp, handleNumber);
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
