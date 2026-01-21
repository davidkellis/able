import * as AST from "../ast";
import { callCallableValue } from "./functions";
import type { Interpreter } from "./index";
import { memberAccessOnValue } from "./structs";
import type { RuntimeValue } from "./values";
import {
  AsyncHandle,
  AwaitWakerPayload,
  MAX_SLEEP_DELAY_MS,
  TimerAwaitableState,
  toMilliseconds,
} from "./concurrency_shared";

export function applyConcurrencyAwait(cls: typeof Interpreter): void {
  cls.prototype.makeAwaitRegistration = function makeAwaitRegistration(
    this: Interpreter,
    cancelFn?: () => void,
  ): Extract<RuntimeValue, { kind: "struct_instance" }> {
    if (!this.awaitRegistrationStruct) {
      throw new Error("Await registration builtins are not initialized");
    }
    const registration: Extract<RuntimeValue, { kind: "struct_instance" }> = {
      kind: "struct_instance",
      def: this.awaitRegistrationStruct,
      values: new Map(),
    };
    const cancelNative = this.makeNativeFunction("AwaitRegistration.cancel", 1, () => {
      if (cancelFn) cancelFn();
      return { kind: "nil", value: null };
    });
    registration.values.set("cancel", this.bindNativeMethod(cancelNative, registration));
    return registration;
  };

  cls.prototype.invokeAwaitWaker = function invokeAwaitWaker(this: Interpreter, waker: RuntimeValue): void {
    try {
      const member = memberAccessOnValue(this, waker, AST.identifier("wake"), this.globals);
      callCallableValue(this, member, [], this.globals);
    } catch {
      // ignore wake errors
    }
  };

  cls.prototype.registerProcAwaiter = function registerProcAwaiter(
    this: Interpreter,
    handle: Extract<RuntimeValue, { kind: "proc_handle" }>,
    waker: Extract<RuntimeValue, { kind: "struct_instance" }>,
  ): Extract<RuntimeValue, { kind: "struct_instance" }> {
    if (!handle || handle.state !== "pending") {
      this.invokeAwaitWaker(waker);
      return this.makeAwaitRegistration();
    }
    const bucket: Set<{ waker: typeof waker; cancelled: boolean }> =
      (handle as any).awaitRegistrations ?? new Set();
    (handle as any).awaitRegistrations = bucket;
    const entry = { waker, cancelled: false };
    bucket.add(entry);
    return this.makeAwaitRegistration(() => {
      entry.cancelled = true;
      bucket.delete(entry);
      if (bucket.size === 0) {
        delete (handle as any).awaitRegistrations;
      }
    });
  };

  cls.prototype.registerFutureAwaiter = function registerFutureAwaiter(
    this: Interpreter,
    future: Extract<RuntimeValue, { kind: "future" }>,
    waker: Extract<RuntimeValue, { kind: "struct_instance" }>,
  ): Extract<RuntimeValue, { kind: "struct_instance" }> {
    if (!future || future.state !== "pending") {
      this.invokeAwaitWaker(waker);
      return this.makeAwaitRegistration();
    }
    const bucket: Set<{ waker: typeof waker; cancelled: boolean }> =
      (future as any).awaitRegistrations ?? new Set();
    (future as any).awaitRegistrations = bucket;
    const entry = { waker, cancelled: false };
    bucket.add(entry);
    return this.makeAwaitRegistration(() => {
      entry.cancelled = true;
      bucket.delete(entry);
      if (bucket.size === 0) {
        delete (future as any).awaitRegistrations;
      }
    });
  };

  cls.prototype.triggerProcAwaiters = function triggerProcAwaiters(
    this: Interpreter,
    handle: Extract<RuntimeValue, { kind: "proc_handle" }>,
  ): void {
    const bucket: Set<{ waker: Extract<RuntimeValue, { kind: "struct_instance" }>; cancelled: boolean }> | undefined = (handle as any)
      .awaitRegistrations;
    if (!bucket || bucket.size === 0) return;
    delete (handle as any).awaitRegistrations;
    for (const entry of bucket) {
      if (!entry.cancelled) {
        this.invokeAwaitWaker(entry.waker);
      }
    }
  };

  cls.prototype.triggerFutureAwaiters = function triggerFutureAwaiters(
    this: Interpreter,
    future: Extract<RuntimeValue, { kind: "future" }>,
  ): void {
    const bucket: Set<{ waker: Extract<RuntimeValue, { kind: "struct_instance" }>; cancelled: boolean }> | undefined = (future as any)
      .awaitRegistrations;
    if (!bucket || bucket.size === 0) return;
    delete (future as any).awaitRegistrations;
    for (const entry of bucket) {
      if (!entry.cancelled) {
        this.invokeAwaitWaker(entry.waker);
      }
    }
  };

  cls.prototype.createAwaitWaker = function createAwaitWaker(
    this: Interpreter,
    handle: AsyncHandle,
    state: unknown,
  ): Extract<RuntimeValue, { kind: "struct_instance" }> {
    if (!this.awaitWakerStruct || !this.awaitWakerNativeMethods) {
      throw new Error("Await waker builtins are not initialized");
    }
    const instance = this.makeNamedStructInstance(this.awaitWakerStruct, []) as Extract<RuntimeValue, { kind: "struct_instance" }>;
    const payload: AwaitWakerPayload = {
      wake: () => {
        if ((instance as any).__awaitTriggered) return;
        (instance as any).__awaitTriggered = true;
        if (state && typeof state === "object") {
          (state as any).wakePending = true;
        }
        handle.awaitBlocked = false;
        if (!handle.runner) {
          handle.runner = () => {
            if (handle.kind === "proc_handle") {
              this.runProcHandle(handle);
            } else {
              this.runFuture(handle);
            }
          };
        }
        this.scheduleAsync(handle.runner);
      },
    };
    (instance as any).__awaitPayload = payload;
    const wakeMethod = this.bindNativeMethod(this.awaitWakerNativeMethods.wake, instance);
    const values = instance.values as Map<string, RuntimeValue>;
    values.set("wake", wakeMethod);
    return instance;
  };

  cls.prototype.ensureAwaitHelperBuiltins = function ensureAwaitHelperBuiltins(this: Interpreter): void {
    if (this.awaitHelpersBuiltinsInitialized) return;
    this.awaitHelpersBuiltinsInitialized = true;

    const awaitableStruct = AST.structDefinition("AwaitHelper", [], "named");

    const wrapAwaitable = (
      inst: Extract<RuntimeValue, { kind: "struct_instance" }>,
      methods: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" | "native_function" }>>,
    ): RuntimeValue => ({
      kind: "interface_value",
      interfaceName: "Awaitable",
      value: inst,
      methods,
    });

    const makeDefaultAwaitable = (callback?: RuntimeValue): RuntimeValue => {
      const inst: Extract<RuntimeValue, { kind: "struct_instance" }> = {
        kind: "struct_instance",
        def: awaitableStruct,
        values: new Map(),
      };
      const isReady = this.makeNativeFunction("Awaitable.is_ready", 1, () => ({ kind: "bool", value: true }));
      const register = this.makeNativeFunction("Awaitable.register", 2, (nativeInterp) => nativeInterp.makeAwaitRegistration());
      const commit = this.makeNativeFunction("Awaitable.commit", 1, (nativeInterp) => {
        if (!callback || callback.kind === "nil") {
          return { kind: "nil", value: null };
        }
        return callCallableValue(nativeInterp, callback, [], nativeInterp.globals);
      });
      const isDefault = this.makeNativeFunction("Awaitable.is_default", 1, () => ({ kind: "bool", value: true }));
      const values = inst.values as Map<string, RuntimeValue>;
      values.set("is_ready", this.bindNativeMethod(isReady, inst));
      values.set("register", this.bindNativeMethod(register, inst));
      values.set("commit", this.bindNativeMethod(commit, inst));
      values.set("is_default", this.bindNativeMethod(isDefault, inst));
      const methods = new Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" | "native_function" }>>();
      methods.set("is_ready", isReady);
      methods.set("register", register);
      methods.set("commit", commit);
      methods.set("is_default", isDefault);
      return wrapAwaitable(inst, methods);
    };

    const makeTimerAwaitable = (durationMs: number, callback?: RuntimeValue): RuntimeValue => {
      const state: TimerAwaitableState = {
        deadlineMs: Date.now() + durationMs,
        ready: durationMs <= 0,
        cancelled: false,
        callback,
      };
      const inst: Extract<RuntimeValue, { kind: "struct_instance" }> = {
        kind: "struct_instance",
        def: awaitableStruct,
        values: new Map(),
      };

      const refreshReady = () => {
        if (!state.ready && Date.now() >= state.deadlineMs) {
          state.ready = true;
        }
      };

      const isReady = this.makeNativeFunction("Awaitable.is_ready", 1, () => {
        refreshReady();
        return { kind: "bool", value: state.ready };
      });

      const register = this.makeNativeFunction("Awaitable.register", 2, (nativeInterp, args) => {
        const waker = args[1];
        if (!waker || waker.kind !== "struct_instance") {
          throw new Error("register expects waker instance");
        }
        refreshReady();
        if (state.ready) {
          nativeInterp.invokeAwaitWaker(waker);
          return nativeInterp.makeAwaitRegistration();
        }
        if (state.timerId !== undefined) {
          clearTimeout(state.timerId);
          state.timerId = undefined;
        }
        state.cancelled = false;
        const remaining = Math.max(0, Math.min(state.deadlineMs - Date.now(), MAX_SLEEP_DELAY_MS));
        state.timerId = setTimeout(() => {
          state.timerId = undefined;
          if (state.cancelled) return;
          state.ready = true;
          nativeInterp.invokeAwaitWaker(waker);
        }, remaining);
        return nativeInterp.makeAwaitRegistration(() => {
          state.cancelled = true;
          if (state.timerId !== undefined) {
            clearTimeout(state.timerId);
            state.timerId = undefined;
          }
        });
      });

      const commit = this.makeNativeFunction("Awaitable.commit", 1, (nativeInterp) => {
        state.ready = true;
        if (state.timerId !== undefined) {
          clearTimeout(state.timerId);
          state.timerId = undefined;
        }
        state.cancelled = false;
        if (!state.callback || state.callback.kind === "nil") {
          return { kind: "nil", value: null };
        }
        return callCallableValue(nativeInterp, state.callback, [], nativeInterp.globals);
      });

      const isDefault = this.makeNativeFunction("Awaitable.is_default", 1, () => ({ kind: "bool", value: false }));
      const values = inst.values as Map<string, RuntimeValue>;
      values.set("is_ready", this.bindNativeMethod(isReady, inst));
      values.set("register", this.bindNativeMethod(register, inst));
      values.set("commit", this.bindNativeMethod(commit, inst));
      values.set("is_default", this.bindNativeMethod(isDefault, inst));
      const methods = new Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" | "native_function" }>>();
      methods.set("is_ready", isReady);
      methods.set("register", register);
      methods.set("commit", commit);
      methods.set("is_default", isDefault);
      return wrapAwaitable(inst, methods);
    };

    const awaitDefault = this.makeNativeFunction("__able_await_default", 1, (_nativeInterp, args) => {
      const callback = args[0];
      return makeDefaultAwaitable(callback && callback.kind !== "nil" ? callback : undefined);
    });
    this.globals.define("__able_await_default", awaitDefault);

    const awaitSleepMs = this.makeNativeFunction("__able_await_sleep_ms", 2, (_nativeInterp, args) => {
      const duration = toMilliseconds(args[0]);
      const callback = args[1];
      return makeTimerAwaitable(duration, callback && callback.kind !== "nil" ? callback : undefined);
    });
    this.globals.define("__able_await_sleep_ms", awaitSleepMs);
  };
}
