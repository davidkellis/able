import * as AST from "../ast";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { RaiseSignal } from "./signals";
import { makeIntegerValue, numericToNumber } from "./numeric";
import { callCallableValue } from "./functions";
import { memberAccessOnValue } from "./structs";

type ProcHandleValue = Extract<RuntimeValue, { kind: "proc_handle" }>;
type BoolValue = Extract<RuntimeValue, { kind: "bool" }>;
type NilValue = Extract<RuntimeValue, { kind: "nil" }>;
type ErrorValue = Extract<RuntimeValue, { kind: "error" }>;

interface ChannelSendWaiter {
  handle: ProcHandleValue;
  value: RuntimeValue;
  error?: ErrorValue;
}

interface ChannelReceiveWaiter {
  handle: ProcHandleValue;
}

interface ChannelState {
  id: number;
  capacity: number;
  queue: RuntimeValue[];
  closed: boolean;
  sendWaiters: ChannelSendWaiter[];
  receiveWaiters: ChannelReceiveWaiter[];
  awaitSendRegistrations?: Set<ChannelAwaitRegistration>;
  awaitReceiveRegistrations?: Set<ChannelAwaitRegistration>;
}

interface MutexState {
  id: number;
  locked: boolean;
  owner: ProcHandleValue | null;
  waiters: ProcHandleValue[];
  awaitRegistrations?: Set<MutexAwaitRegistration>;
}

interface MutexAwaitRegistration {
  waker: Extract<RuntimeValue, { kind: "struct_instance" }>;
  cancelled?: boolean;
}

interface ChannelAwaitRegistration {
  kind: "send" | "receive";
  waker: Extract<RuntimeValue, { kind: "struct_instance" }>;
  cancelled?: boolean;
}

const NIL: NilValue = { kind: "nil", value: null };
const AWAITABLE_STRUCT = AST.structDefinition("ChannelAwaitable", [], "named");

declare module "./index" {
  interface Interpreter {
    ensureChannelMutexBuiltins(): void;
    channelMutexBuiltinsInitialized: boolean;
    nextChannelHandle: number;
    channelStates: Map<number, ChannelState>;
    channelErrorStructs: Map<string, AST.StructDefinition>;
    nextMutexHandle: number;
    mutexStates: Map<number, MutexState>;
  }
}

declare module "./values" {
  interface ProcHandleValue {
    waitingMutex?: MutexState;
    waitingChannelSend?: {
      state: ChannelState;
      value: RuntimeValue;
      delivered?: boolean;
      error?: ErrorValue;
    };
    waitingChannelReceive?: {
      state: ChannelState;
      ready?: boolean;
      value?: RuntimeValue;
      closed?: boolean;
    };
  }
}

function toHandleNumber(value: RuntimeValue, label: string): number {
  return Math.trunc(numericToNumber(value, label, { requireSafeInteger: true }));
}

function blockOnNilChannel(interp: Interpreter): RuntimeValue {
  const ctx = interp.currentAsyncContext();
  if (!ctx || ctx.kind !== "proc") {
    throw new Error("Nil channel operations must occur inside a proc");
  }
  ctx.handle.awaitBlocked = true;
  const cancelled = interp.procCancelled(true) as BoolValue;
  if (cancelled.value) {
    return NIL;
  }
  interp.procYield(true);
  return NIL;
}

function requireProcContext(interp: Interpreter, action: string): ProcHandleValue {
  const ctx = interp.currentAsyncContext();
  if (!ctx || ctx.kind !== "proc") {
    throw new Error(`${action} must occur inside a proc`);
  }
  return ctx.handle;
}

function scheduleProc(interp: Interpreter, handle: ProcHandleValue): void {
  handle.awaitBlocked = false;
  if (!handle.runner) {
    handle.runner = () => interp.runProcHandle(handle);
  }
  interp.scheduleAsync(handle.runner);
}

function getChannelState(interp: Interpreter, handle: number): ChannelState | undefined {
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

function resolveChannelErrorStruct(interp: Interpreter, structName: string): AST.StructDefinition | null {
  if (interp.channelErrorStructs.has(structName)) {
    return interp.channelErrorStructs.get(structName)!;
  }
  const candidateNames = [
    structName,
    `concurrency.${structName}`,
    `able.${structName}`,
    `able.concurrency.${structName}`,
    `able.concurrency.channel.${structName}`,
    `able.concurrency.mutex.${structName}`,
  ];
  for (const name of candidateNames) {
    try {
      const val = interp.globals.get(name);
      if (val && val.kind === "struct_def") {
        interp.channelErrorStructs.set(structName, val.def);
        return val.def;
      }
    } catch {
      // ignore lookup errors
    }
  }
  for (const bucket of interp.packageRegistry.values()) {
    const val = bucket.get(structName);
    if (val && val.kind === "struct_def") {
      interp.channelErrorStructs.set(structName, val.def);
      return val.def;
    }
  }
  return null;
}

function makeChannelErrorValue(interp: Interpreter, structName: string, fallbackMessage: string): ErrorValue {
  let structDef = resolveChannelErrorStruct(interp, structName);
  if (!structDef) {
    structDef = AST.structDefinition(structName, [], "named");
    interp.channelErrorStructs.set(structName, structDef);
  }
  const instance = interp.makeNamedStructInstance(structDef, []) as Extract<RuntimeValue, { kind: "struct_instance" }>;
  return { kind: "error", message: fallbackMessage, value: instance };
}

function raiseChannelError(interp: Interpreter, structName: string, fallbackMessage: string): never {
  const err = makeChannelErrorValue(interp, structName, fallbackMessage);
  throw new RaiseSignal(err);
}

function addChannelAwaiter(
  state: ChannelState,
  kind: ChannelAwaitRegistration["kind"],
  waker: Extract<RuntimeValue, { kind: "struct_instance" }>,
): ChannelAwaitRegistration {
  const registration: ChannelAwaitRegistration = { kind, waker };
  const bucketKey = kind === "send" ? "awaitSendRegistrations" : "awaitReceiveRegistrations";
  if (!state[bucketKey]) {
    state[bucketKey] = new Set();
  }
  state[bucketKey]!.add(registration);
  return registration;
}

function cancelChannelAwaiter(state: ChannelState, registration: ChannelAwaitRegistration): void {
  if (!registration || registration.cancelled) return;
  registration.cancelled = true;
  const bucket = registration.kind === "send" ? state.awaitSendRegistrations : state.awaitReceiveRegistrations;
  bucket?.delete(registration);
}

function triggerChannelAwaiter(interp: Interpreter, registration: ChannelAwaitRegistration): void {
  if (!registration || registration.cancelled) return;
  const wakeMember = AST.identifier("wake");
  try {
    const member = memberAccessOnValue(interp, registration.waker, wakeMember, interp.globals);
    callCallableValue(interp, member, [], interp.globals);
  } catch {
    // ignore wake failures
  }
}

function notifyChannelAwaiters(
  interp: Interpreter,
  state: ChannelState,
  kind: ChannelAwaitRegistration["kind"],
): void {
  const bucket = kind === "send" ? state.awaitSendRegistrations : state.awaitReceiveRegistrations;
  if (!bucket || bucket.size === 0) return;
  for (const reg of Array.from(bucket)) {
    triggerChannelAwaiter(interp, reg);
  }
}

function addMutexAwaiter(state: MutexState, waker: Extract<RuntimeValue, { kind: "struct_instance" }>): MutexAwaitRegistration {
  const registration: MutexAwaitRegistration = { waker };
  if (!state.awaitRegistrations) {
    state.awaitRegistrations = new Set();
  }
  state.awaitRegistrations.add(registration);
  return registration;
}

function cancelMutexAwaiter(state: MutexState, registration: MutexAwaitRegistration): void {
  if (!registration || registration.cancelled) return;
  registration.cancelled = true;
  state.awaitRegistrations?.delete(registration);
}

function notifyMutexAwaiters(interp: Interpreter, state: MutexState): void {
  const regs = state.awaitRegistrations ? Array.from(state.awaitRegistrations) : [];
  state.awaitRegistrations?.clear();
  for (const reg of regs) {
    if (reg.cancelled) continue;
    try {
      const wakeMember = memberAccessOnValue(interp, reg.waker, AST.identifier("wake"), interp.globals);
      callCallableValue(interp, wakeMember, [], interp.globals);
    } catch {
      // ignore
    }
  }
}

export function applyChannelMutexAugmentations(cls: typeof Interpreter): void {
  cls.prototype.ensureChannelMutexBuiltins = function ensureChannelMutexBuiltins(this: Interpreter): void {
    if (this.channelMutexBuiltinsInitialized) return;
    this.channelMutexBuiltinsInitialized = true;

    if (!this.channelStates) this.channelStates = new Map();
    if (!this.mutexStates) this.mutexStates = new Map();

    const defineIfMissing = (name: string, factory: () => Extract<RuntimeValue, { kind: "native_function" }>) => {
      try {
        this.globals.get(name);
        return;
      } catch {
        // not defined, continue
      }
      this.globals.define(name, factory());
    };

    const makeChannelAwaitable = (
      handleNumber: number,
      op: "send" | "receive",
      payload: RuntimeValue | null,
      callback?: RuntimeValue,
    ): Extract<RuntimeValue, { kind: "struct_instance" }> => {
      const inst: Extract<RuntimeValue, { kind: "struct_instance" }> = {
        kind: "struct_instance",
        def: AWAITABLE_STRUCT,
        values: new Map(),
      };

      const isReady = this.makeNativeFunction("Awaitable.is_ready", 1, (nativeInterp) => {
        if (handleNumber === 0) {
          return { kind: "bool", value: false };
        }
        const state = getChannelState(nativeInterp, handleNumber);
        if (!state) throw new Error("Invalid channel handle");
        const closed = state.closed;
        const length = state.queue.length;
        const awaitingSend = state.awaitSendRegistrations?.size ?? 0;
        const awaitingReceive = state.awaitReceiveRegistrations?.size ?? 0;
        if (op === "receive") {
          const ready =
            length > 0 || (state.capacity === 0 && (state.sendWaiters.length > 0 || awaitingSend > 0)) || closed;
          return { kind: "bool", value: ready };
        }
        if (closed) {
          raiseChannelError(nativeInterp, "ChannelSendOnClosed", "send on closed channel");
        }
        const ready = state.capacity === 0 ? state.receiveWaiters.length > 0 || awaitingReceive > 0 : length < state.capacity;
        return { kind: "bool", value: ready };
      });

      const register = this.makeNativeFunction("Awaitable.register", 2, (nativeInterp, args) => {
        const waker = args[1];
        if (!waker || waker.kind !== "struct_instance") {
          throw new Error("register expects waker instance");
        }
        if (handleNumber === 0) {
          return nativeInterp.makeAwaitRegistration();
        }
        const state = getChannelState(nativeInterp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        const registration = addChannelAwaiter(state, op, waker);
        const cancelFn = () => cancelChannelAwaiter(state, registration);
        return nativeInterp.makeAwaitRegistration(cancelFn);
      });

      const commit = this.makeNativeFunction("Awaitable.commit", 1, (nativeInterp) => {
        const handleValue = makeIntegerValue("i32", BigInt(handleNumber));
        if (op === "receive") {
          const recvFn = nativeInterp.globals.get("__able_channel_receive");
          if (!recvFn || recvFn.kind !== "native_function") {
            throw new Error("__able_channel_receive is not available");
          }
          const value = recvFn.impl(nativeInterp, [handleValue]);
          if (!callback) {
            return value;
          }
          return callCallableValue(nativeInterp, callback, [value], nativeInterp.globals);
        }

        const sendFn = nativeInterp.globals.get("__able_channel_send");
        if (!sendFn || sendFn.kind !== "native_function") {
          throw new Error("__able_channel_send is not available");
        }
        sendFn.impl(nativeInterp, [handleValue, payload ?? NIL]);
        if (!callback) {
          return NIL;
        }
        return callCallableValue(nativeInterp, callback, [], nativeInterp.globals);
      });

      const isDefault = this.makeNativeFunction("Awaitable.is_default", 1, () => ({ kind: "bool", value: false }));

      const values = inst.values as Map<string, RuntimeValue>;
      values.set("is_ready", this.bindNativeMethod(isReady, inst));
      values.set("register", this.bindNativeMethod(register, inst));
      values.set("commit", this.bindNativeMethod(commit, inst));
      values.set("is_default", this.bindNativeMethod(isDefault, inst));
      return inst;
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
        return makeIntegerValue("i32", BigInt(handle));
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
            throw new RaiseSignal(pending.error);
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
          raiseChannelError(interp, "ChannelSendOnClosed", "send on closed channel");
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
          notifyChannelAwaiters(interp, state, "receive");
          return NIL;
        }

        if (!procHandle) {
          throw new Error("Channel send would block outside of proc context");
        }

        const hadSendWaiters = state.sendWaiters.length;
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
        if (state.capacity === 0 && hadSendWaiters === 0 && state.sendWaiters.length > 0) {
          notifyChannelAwaiters(interp, state, "receive");
        }
        procHandle.awaitBlocked = true;
        interp.procYield(true);
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
          notifyChannelAwaiters(interp, state, "send");
          if (state.queue.length > 0) {
            notifyChannelAwaiters(interp, state, "receive");
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
          notifyChannelAwaiters(interp, state, "send");
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
        const hadWaiters = state.receiveWaiters.length;
        if (!state.receiveWaiters.some((entry) => entry.handle === procHandle)) {
          state.receiveWaiters.push({ handle: procHandle });
        }
        if (state.capacity === 0 && hadWaiters === 0 && state.receiveWaiters.length > 0) {
          notifyChannelAwaiters(interp, state, "send");
        }
        procHandle.awaitBlocked = true;
        interp.procYield(true);
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
          raiseChannelError(interp, "ChannelSendOnClosed", "send on closed channel");
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
          notifyChannelAwaiters(interp, state, "receive");
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
          notifyChannelAwaiters(interp, state, "send");
          if (state.queue.length > 0) {
            notifyChannelAwaiters(interp, state, "receive");
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
          notifyChannelAwaiters(interp, state, "send");
          return sender.value ?? NIL;
        }
        if (state.closed) {
          return NIL;
        }
        return NIL;
      }),
    );

    defineIfMissing("__able_channel_await_try_recv", () =>
      this.makeNativeFunction("__able_channel_await_try_recv", 2, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        const callback = args[1];
        return makeChannelAwaitable(handleNumber, "receive", null, callback);
      }),
    );

    defineIfMissing("__able_channel_await_try_send", () =>
      this.makeNativeFunction("__able_channel_await_try_send", 3, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        const payload = args[1];
        const callback = args[2];
        return makeChannelAwaitable(handleNumber, "send", payload, callback);
      }),
    );

    defineIfMissing("__able_channel_close", () =>
      this.makeNativeFunction("__able_channel_close", 1, (interp, args) => {
        const handleNumber = toHandleNumber(args[0], "channel handle");
        if (handleNumber === 0) {
          raiseChannelError(interp, "ChannelNil", "close of nil channel");
        }
        const state = getChannelState(interp, handleNumber);
        if (!state) {
          throw new Error("Invalid channel handle");
        }
        if (state.closed) {
          raiseChannelError(interp, "ChannelClosed", "close of closed channel");
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
            pending.error = makeChannelErrorValue(interp, "ChannelSendOnClosed", "send on closed channel");
          }
          scheduleProc(interp, sender.handle);
        }

        notifyChannelAwaiters(interp, state, "receive");
        notifyChannelAwaiters(interp, state, "send");

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
        return makeIntegerValue("i32", BigInt(handle));
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
          procHandle.awaitBlocked = true;
          interp.procYield(true);
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
          raiseChannelError(interp, "MutexUnlocked", "unlock of unlocked mutex");
        }

        state.locked = false;
        state.owner = null;
        notifyMutexAwaiters(interp, state);
        if (state.waiters.length > 0) {
          const next = state.waiters.shift()!;
          if ((next as any).waitingMutex === state) {
            delete (next as any).waitingMutex;
          }
          scheduleProc(interp, next);
        }

        return NIL;
      }),
    );

    defineIfMissing("__able_mutex_await_lock", () =>
      this.makeNativeFunction("__able_mutex_await_lock", 2, (interp, args) => {
        const handle = toHandleNumber(args[0], "mutex handle");
        const callback = args[1];
        const inst: Extract<RuntimeValue, { kind: "struct_instance" }> = {
          kind: "struct_instance",
          def: AWAITABLE_STRUCT,
          values: new Map(),
        };
        const isReady = interp.makeNativeFunction("Awaitable.is_ready", 1, (nativeInterp) => {
          const state = nativeInterp.mutexStates.get(handle);
          if (!state) {
            throw new Error("Invalid mutex handle");
          }
          return { kind: "bool", value: state.locked === false };
        });
        const register = interp.makeNativeFunction("Awaitable.register", 2, (nativeInterp, regArgs) => {
          const waker = regArgs[1];
          if (!waker || waker.kind !== "struct_instance") {
            throw new Error("register expects waker instance");
          }
          const state = nativeInterp.mutexStates.get(handle);
          if (!state) {
            throw new Error("Invalid mutex handle");
          }
          if (!state.locked) {
            nativeInterp.invokeAwaitWaker(waker);
            return nativeInterp.makeAwaitRegistration();
          }
          const registration = addMutexAwaiter(state, waker);
          return nativeInterp.makeAwaitRegistration(() => cancelMutexAwaiter(state, registration));
        });
        const commit = interp.makeNativeFunction("Awaitable.commit", 1, (nativeInterp) => {
          const lockFn = nativeInterp.globals.get("__able_mutex_lock");
          if (!lockFn || lockFn.kind !== "native_function") {
            throw new Error("__able_mutex_lock is not available");
          }
          lockFn.impl(nativeInterp, [makeIntegerValue("i32", BigInt(handle))]);
          if (!callback || callback.kind === "nil") {
            return NIL;
          }
          return callCallableValue(nativeInterp, callback, [], nativeInterp.globals);
        });
        const isDefault = interp.makeNativeFunction("Awaitable.is_default", 1, () => ({ kind: "bool", value: false }));
        const values = inst.values as Map<string, RuntimeValue>;
        values.set("is_ready", interp.bindNativeMethod(isReady, inst));
        values.set("register", interp.bindNativeMethod(register, inst));
        values.set("commit", interp.bindNativeMethod(commit, inst));
        values.set("is_default", interp.bindNativeMethod(isDefault, inst));
        return inst;
      }),
    );
  };
}
