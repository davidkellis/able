import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { Executor } from "./executor";
import { ExitSignal, ProcYieldSignal, RaiseSignal } from "./signals";
import type { RuntimeValue } from "./values";
import { ProcContinuationContext } from "./proc_continuations";
import { isFloatValue, isIntegerValue, makeIntegerValue } from "./numeric";
import { callCallableValue } from "./functions";
import { memberAccessOnValue } from "./structs";

type AwaitWakerPayload = {
  wake: () => void;
};

type AsyncHandle = Extract<RuntimeValue, { kind: "proc_handle" | "future" }>;

type ProcValueWaitState = {
  target: Extract<RuntimeValue, { kind: "proc_handle" }>;
  registration?: RuntimeValue;
  wakePending?: boolean;
};

type FutureValueWaitState = {
  target: Extract<RuntimeValue, { kind: "future" }>;
  registration?: RuntimeValue;
  wakePending?: boolean;
};

type TimerAwaitableState = {
  deadlineMs: number;
  ready: boolean;
  cancelled: boolean;
  timerId?: ReturnType<typeof setTimeout>;
  callback?: RuntimeValue;
};

const MAX_SLEEP_DELAY_MS = 2_147_483_647;

function toMilliseconds(value: RuntimeValue): number {
  if (isIntegerValue(value)) {
    const limit = BigInt(MAX_SLEEP_DELAY_MS);
    const clamped = value.value < 0n ? 0n : value.value > limit ? limit : value.value;
    return Number(clamped);
  }
  if (isFloatValue(value)) {
    if (!Number.isFinite(value.value)) {
      throw new Error("sleep_ms expects a finite duration");
    }
    const truncated = Math.trunc(value.value);
    return Math.max(0, Math.min(truncated, MAX_SLEEP_DELAY_MS));
  }
  throw new Error("sleep_ms expects a numeric duration");
}

function cancelAwaitRegistration(interp: Interpreter, registration: RuntimeValue | undefined): void {
  if (!registration) return;
  try {
    const cancel = memberAccessOnValue(interp, registration, AST.identifier("cancel"), interp.globals);
    callCallableValue(interp, cancel, [], interp.globals);
  } catch {
    // ignore cancellation errors
  }
}

declare module "./index" {
  interface Interpreter {
    initConcurrencyBuiltins(): void;
    scheduleAsync(fn: () => void): void;
    ensureSchedulerTick(): void;
    currentAsyncContext(): { kind: "proc"; handle: Extract<RuntimeValue, { kind: "proc_handle" }> } | { kind: "future"; handle: Extract<RuntimeValue, { kind: "future" }> } | null;
    createAwaitWaker(handle: AsyncHandle, state: unknown): Extract<RuntimeValue, { kind: "struct_instance" }>;
    makeAwaitRegistration(cancelFn?: () => void): Extract<RuntimeValue, { kind: "struct_instance" }>;
    invokeAwaitWaker(waker: RuntimeValue): void;
    registerProcAwaiter(
      handle: Extract<RuntimeValue, { kind: "proc_handle" }>,
      waker: Extract<RuntimeValue, { kind: "struct_instance" }>,
    ): Extract<RuntimeValue, { kind: "struct_instance" }>;
    registerFutureAwaiter(
      future: Extract<RuntimeValue, { kind: "future" }>,
      waker: Extract<RuntimeValue, { kind: "struct_instance" }>,
    ): Extract<RuntimeValue, { kind: "struct_instance" }>;
    triggerProcAwaiters(handle: Extract<RuntimeValue, { kind: "proc_handle" }>): void;
    triggerFutureAwaiters(future: Extract<RuntimeValue, { kind: "future" }>): void;
    procYield(allowEntrypoint?: boolean): RuntimeValue;
    procCancelled(allowEntrypoint?: boolean): RuntimeValue;
    procFlush(): RuntimeValue;
    procPendingTasks(): RuntimeValue;
    processScheduler(limit?: number): void;
    makeNamedStructInstance(def: AST.StructDefinition, entries: Array<[string, RuntimeValue]>): RuntimeValue;
    makeProcError(details: string): RuntimeValue;
    getProcErrorDetails(procError: RuntimeValue): string;
    makeProcStatusFailed(procError: RuntimeValue): RuntimeValue;
    markProcCancelled(handle: Extract<RuntimeValue, { kind: "proc_handle" }>, message?: string): void;
    markFutureCancelled(future: Extract<RuntimeValue, { kind: "future" }>, message?: string): void;
    procHandleStatus(handle: Extract<RuntimeValue, { kind: "proc_handle" }>): RuntimeValue;
    futureStatus(future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue;
    toProcError(value: RuntimeValue | undefined, fallback: string): RuntimeValue;
    makeNativeFunction(name: string, arity: number, impl: (interpreter: Interpreter, args: RuntimeValue[]) => RuntimeValue | Promise<RuntimeValue>): Extract<RuntimeValue, { kind: "native_function" }>;
    bindNativeMethod(func: Extract<RuntimeValue, { kind: "native_function" }>, self: RuntimeValue): Extract<RuntimeValue, { kind: "native_bound_method" }>;
    procHandleValue(handle: Extract<RuntimeValue, { kind: "proc_handle" }>): RuntimeValue;
    procHandleCancel(handle: Extract<RuntimeValue, { kind: "proc_handle" }>): void;
    futureCancel(future: Extract<RuntimeValue, { kind: "future" }>): void;
    futureValue(future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue;
    evaluateAsTask(node: AST.AstNode, env?: Environment): Promise<RuntimeValue>;
    runProcHandle(handle: Extract<RuntimeValue, { kind: "proc_handle" }>): void;
    runFuture(future: Extract<RuntimeValue, { kind: "future" }>): void;
    makeRuntimeError(message: string, value?: RuntimeValue, cause?: RuntimeValue): RuntimeValue;
    executor: Executor;
    pushProcContext(ctx: ProcContinuationContext): void;
    popProcContext(ctx: ProcContinuationContext): void;
    currentProcContext(): ProcContinuationContext | null;
    procContextStack: ProcContinuationContext[];
    awaitWakerStruct?: AST.StructDefinition;
    awaitRegistrationStruct?: AST.StructDefinition;
    awaitWakerNativeMethods?: {
      wake: Extract<RuntimeValue, { kind: "native_function" }>;
    };
    awaitHelpersBuiltinsInitialized?: boolean;
    ensureAwaitHelperBuiltins(): void;
  }
}

export function applyConcurrencyAugmentations(cls: typeof Interpreter): void {
  cls.prototype.initConcurrencyBuiltins = function initConcurrencyBuiltins(this: Interpreter): void {
    if (this.concurrencyBuiltinsInitialized) return;
    this.concurrencyBuiltinsInitialized = true;

    const procErrorDefAst = AST.structDefinition(
      "ProcError",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "details")],
      "named"
    );
    const pendingDefAst = AST.structDefinition("Pending", [], "named");
    const resolvedDefAst = AST.structDefinition("Resolved", [], "named");
    const cancelledDefAst = AST.structDefinition("Cancelled", [], "named");
    const failedDefAst = AST.structDefinition(
      "Failed",
      [AST.structFieldDefinition(AST.simpleTypeExpression("ProcError"), "error")],
      "named"
    );

    this.evaluate(procErrorDefAst, this.globals);
    this.evaluate(pendingDefAst, this.globals);
    this.evaluate(resolvedDefAst, this.globals);
    this.evaluate(cancelledDefAst, this.globals);
    this.evaluate(failedDefAst, this.globals);
    this.evaluate(
      AST.unionDefinition(
        "ProcStatus",
        [
          AST.simpleTypeExpression("Pending"),
          AST.simpleTypeExpression("Resolved"),
          AST.simpleTypeExpression("Cancelled"),
          AST.simpleTypeExpression("Failed"),
        ],
        undefined,
        undefined,
        false
      ),
      this.globals,
    );

    const getStructDef = (name: string): AST.StructDefinition => {
      const val = this.globals.get(name);
      if (val.kind !== "struct_def") throw new Error(`Failed to initialize struct '${name}'`);
      return val.def;
    };

    this.procErrorStruct = getStructDef("ProcError");
    this.procStatusStructs = {
      Pending: getStructDef("Pending"),
      Resolved: getStructDef("Resolved"),
      Cancelled: getStructDef("Cancelled"),
      Failed: getStructDef("Failed"),
    };

    this.procStatusPendingValue = this.makeNamedStructInstance(this.procStatusStructs.Pending, []);
    this.procStatusResolvedValue = this.makeNamedStructInstance(this.procStatusStructs.Resolved, []);
    this.procStatusCancelledValue = this.makeNamedStructInstance(this.procStatusStructs.Cancelled, []);

    const awaitWakerDefAst = AST.structDefinition("AwaitWaker", [], "named");
    const awaitRegistrationDefAst = AST.structDefinition("AwaitRegistration", [], "named");
    this.evaluate(awaitWakerDefAst, this.globals);
    this.evaluate(awaitRegistrationDefAst, this.globals);
    const awaitWakerVal = this.globals.get("AwaitWaker");
    if (!awaitWakerVal || awaitWakerVal.kind !== "struct_def") {
      throw new Error("Failed to initialize struct 'AwaitWaker'");
    }
    this.awaitWakerStruct = awaitWakerVal.def;
    const awaitRegistrationVal = this.globals.get("AwaitRegistration");
    if (!awaitRegistrationVal || awaitRegistrationVal.kind !== "struct_def") {
      throw new Error("Failed to initialize struct 'AwaitRegistration'");
    }
    this.awaitRegistrationStruct = awaitRegistrationVal.def;
    const wakeNative = this.makeNativeFunction("AwaitWaker.wake", 1, (_interp, args) => {
      const self = args[0];
      if (!self || self.kind !== "struct_instance") {
        return { kind: "nil", value: null };
      }
      const payload = (self as any).__awaitPayload as AwaitWakerPayload | undefined;
      if (payload) {
        payload.wake();
      }
      return { kind: "nil", value: null };
    });
    this.awaitWakerNativeMethods = { wake: wakeNative };
    this.ensureAwaitHelperBuiltins();
  };

  cls.prototype.scheduleAsync = function scheduleAsync(this: Interpreter, fn: () => void): void {
    this.executor.schedule(fn);
  };

  cls.prototype.ensureSchedulerTick = function ensureSchedulerTick(this: Interpreter): void {
    this.executor.ensureTick();
  };

  cls.prototype.pushProcContext = function pushProcContext(this: Interpreter, ctx: ProcContinuationContext): void {
    if (!this.procContextStack) {
      this.procContextStack = [];
    }
    this.procContextStack.push(ctx);
  };

  cls.prototype.popProcContext = function popProcContext(this: Interpreter, ctx: ProcContinuationContext): void {
    if (!this.procContextStack || this.procContextStack.length === 0) return;
    const top = this.procContextStack[this.procContextStack.length - 1];
    if (top === ctx) {
      this.procContextStack.pop();
    }
  };

  cls.prototype.currentProcContext = function currentProcContext(this: Interpreter): ProcContinuationContext | null {
    if (!this.procContextStack || this.procContextStack.length === 0) return null;
    return this.procContextStack[this.procContextStack.length - 1]!;
  };

  cls.prototype.currentAsyncContext = function currentAsyncContext(this: Interpreter): { kind: "proc"; handle: Extract<RuntimeValue, { kind: "proc_handle" }> } | { kind: "future"; handle: Extract<RuntimeValue, { kind: "future" }> } | null {
    if (this.asyncContextStack.length === 0) return null;
    return this.asyncContextStack[this.asyncContextStack.length - 1];
  };

  cls.prototype.procYield = function procYield(this: Interpreter, allowEntrypoint = false): RuntimeValue {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_yield must be called inside an asynchronous task");
    if (!allowEntrypoint && ctx.kind === "proc" && ctx.handle.entrypoint) {
      throw new Error("proc_yield must be called inside an asynchronous task");
    }
    this.manualYieldRequested = true;
    throw new ProcYieldSignal();
  };

  cls.prototype.procCancelled = function procCancelled(this: Interpreter, allowEntrypoint = false): RuntimeValue {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_cancelled must be called inside an asynchronous task");
    if (!allowEntrypoint && ctx.kind === "proc" && ctx.handle.entrypoint) {
      throw new Error("proc_cancelled must be called inside an asynchronous task");
    }
    if (ctx.kind === "proc") {
      return { kind: "bool", value: !!ctx.handle.cancelRequested };
    }
    return { kind: "bool", value: false };
  };

  cls.prototype.procFlush = function procFlush(this: Interpreter): RuntimeValue {
    this.processScheduler(this.schedulerMaxSteps);
    return { kind: "nil", value: null };
  };

  cls.prototype.procPendingTasks = function procPendingTasks(this: Interpreter): RuntimeValue {
    const pending = typeof this.executor.pendingTasks === "function" ? this.executor.pendingTasks() : 0;
    return makeIntegerValue("i32", BigInt(pending));
  };

  cls.prototype.processScheduler = function processScheduler(this: Interpreter, limit: number = this.schedulerMaxSteps): void {
    this.executor.flush(limit);
  };

  cls.prototype.makeNamedStructInstance = function makeNamedStructInstance(this: Interpreter, def: AST.StructDefinition, entries: Array<[string, RuntimeValue]>): RuntimeValue {
    const map = new Map<string, RuntimeValue>();
    for (const [key, value] of entries) map.set(key, value);
    return { kind: "struct_instance", def, values: map };
  };

  cls.prototype.makeProcError = function makeProcError(this: Interpreter, details: string): RuntimeValue {
    return this.makeNamedStructInstance(this.procErrorStruct, [["details", { kind: "String", value: details }]]);
  };

  cls.prototype.getProcErrorDetails = function getProcErrorDetails(this: Interpreter, procError: RuntimeValue): string {
    if (procError.kind === "struct_instance" && procError.def.id.name === "ProcError") {
      const map = procError.values as Map<string, RuntimeValue>;
      const detailsVal = map.get("details");
      if (detailsVal && detailsVal.kind === "String") return detailsVal.value;
    }
    return "unknown failure";
  };

  cls.prototype.makeProcStatusFailed = function makeProcStatusFailed(this: Interpreter, procError: RuntimeValue): RuntimeValue {
    return this.makeNamedStructInstance(this.procStatusStructs.Failed, [["error", procError]]);
  };

  cls.prototype.markProcCancelled = function markProcCancelled(this: Interpreter, handle: Extract<RuntimeValue, { kind: "proc_handle" }>, message = "Proc cancelled"): void {
    const pendingSend = (handle as any).waitingChannelSend as
      | {
          state: any;
          value: RuntimeValue;
        }
      | undefined;
    if (pendingSend) {
      pendingSend.state.sendWaiters = pendingSend.state.sendWaiters.filter((entry: any) => entry.handle !== handle);
      delete (handle as any).waitingChannelSend;
    }
    const pendingReceive = (handle as any).waitingChannelReceive as
      | {
          state: any;
        }
      | undefined;
    if (pendingReceive) {
      pendingReceive.state.receiveWaiters = pendingReceive.state.receiveWaiters.filter((entry: any) => entry.handle !== handle);
      delete (handle as any).waitingChannelReceive;
    }
    if ((handle as any).waitingMutex) {
      const state = (handle as any).waitingMutex as any;
      state.waiters = state.waiters.filter((entry: any) => entry !== handle);
      delete (handle as any).waitingMutex;
    }
    const procErr = this.makeProcError(message);
    handle.state = "cancelled";
    handle.result = undefined;
    handle.failureInfo = procErr;
    handle.error = this.makeRuntimeError(message, procErr, procErr);
    handle.runner = null;
    handle.awaitBlocked = false;
    this.triggerProcAwaiters(handle);
  };

  cls.prototype.markFutureCancelled = function markFutureCancelled(this: Interpreter, future: Extract<RuntimeValue, { kind: "future" }>, message = "Future cancelled"): void {
    const pendingSend = (future as any).waitingChannelSend as
      | {
          state: any;
          value: RuntimeValue;
        }
      | undefined;
    if (pendingSend) {
      pendingSend.state.sendWaiters = pendingSend.state.sendWaiters.filter((entry: any) => entry.handle !== future);
      delete (future as any).waitingChannelSend;
    }
    const pendingReceive = (future as any).waitingChannelReceive as
      | {
          state: any;
        }
      | undefined;
    if (pendingReceive) {
      pendingReceive.state.receiveWaiters = pendingReceive.state.receiveWaiters.filter((entry: any) => entry.handle !== future);
      delete (future as any).waitingChannelReceive;
    }
    if ((future as any).waitingMutex) {
      const state = (future as any).waitingMutex as any;
      state.waiters = state.waiters.filter((entry: any) => entry !== future);
      delete (future as any).waitingMutex;
    }
    const procErr = this.makeProcError(message);
    future.state = "cancelled";
    future.result = undefined;
    future.failureInfo = procErr;
    future.error = this.makeRuntimeError(message, procErr, procErr);
    future.runner = null;
    future.awaitBlocked = false;
    this.triggerFutureAwaiters(future);
  };

  cls.prototype.procHandleStatus = function procHandleStatus(this: Interpreter, handle: Extract<RuntimeValue, { kind: "proc_handle" }>): RuntimeValue {
    switch (handle.state) {
      case "pending":
        return this.procStatusPendingValue;
      case "resolved":
        return this.procStatusResolvedValue;
      case "cancelled":
        return this.procStatusCancelledValue;
      case "failed": {
        const procErr = handle.failureInfo ?? this.makeProcError("unknown failure");
        return this.makeProcStatusFailed(procErr);
      }
      default:
        return this.procStatusPendingValue;
    }
  };

  cls.prototype.futureStatus = function futureStatus(this: Interpreter, future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue {
    switch (future.state) {
      case "pending":
        return this.procStatusPendingValue;
      case "resolved":
        return this.procStatusResolvedValue;
      case "cancelled":
        return this.procStatusCancelledValue;
      case "failed": {
        const procErr = future.failureInfo ?? this.makeProcError("unknown failure");
        return this.makeProcStatusFailed(procErr);
      }
    }
  };

  cls.prototype.toProcError = function toProcError(this: Interpreter, value: RuntimeValue | undefined, fallback: string): RuntimeValue {
    if (value) {
      if (value.kind === "struct_instance" && value.def.id.name === "ProcError") {
        return value;
      }
      if (value.kind === "error") {
        if (value.cause && value.cause.kind === "struct_instance" && value.cause.def.id.name === "ProcError") {
          return value.cause;
        }
        if (value.value && value.value.kind === "struct_instance" && value.value.def.id.name === "ProcError") {
          return value.value;
        }
        return this.makeProcError(value.message ?? fallback);
      }
      return this.makeProcError(this.valueToString(value));
    }
    return this.makeProcError(fallback);
  };

  cls.prototype.makeNativeFunction = function makeNativeFunction(
    this: Interpreter,
    name: string,
    arity: number,
    impl: (interpreter: Interpreter, args: RuntimeValue[]) => RuntimeValue | Promise<RuntimeValue>,
  ): Extract<RuntimeValue, { kind: "native_function" }> {
    return { kind: "native_function", name, arity, impl };
  };

  cls.prototype.bindNativeMethod = function bindNativeMethod(this: Interpreter, func: Extract<RuntimeValue, { kind: "native_function" }>, self: RuntimeValue): Extract<RuntimeValue, { kind: "native_bound_method" }> {
    return { kind: "native_bound_method", func, self };
  };

  cls.prototype.procHandleValue = function procHandleValue(this: Interpreter, handle: Extract<RuntimeValue, { kind: "proc_handle" }>): RuntimeValue {
    const finalize = (): RuntimeValue => {
      switch (handle.state) {
        case "resolved":
          return handle.result ?? { kind: "nil", value: null };
        case "failed": {
          if (handle.error) return handle.error;
          const procErr = this.makeProcError("Proc failed");
          return this.makeRuntimeError("Proc failed", procErr, procErr);
        }
        case "cancelled": {
          if (handle.error) return handle.error;
          const procErr = this.makeProcError("Proc cancelled");
          return this.makeRuntimeError("Proc cancelled", procErr, procErr);
        }
        default: {
          const procErr = this.makeProcError("Proc pending");
          return this.makeRuntimeError("Proc pending", procErr, procErr);
        }
      }
    };

    const asyncCtx = this.currentAsyncContext();
    if (asyncCtx) {
      const waiter = asyncCtx.handle as AsyncHandle;
      let waitState = (waiter as any).waitingProcValue as ProcValueWaitState | undefined;
      if (waitState && waitState.target !== handle) {
        cancelAwaitRegistration(this, waitState.registration);
        delete (waiter as any).waitingProcValue;
        waitState = undefined;
      }

      if (handle.state !== "pending") {
        if (waitState) {
          cancelAwaitRegistration(this, waitState.registration);
          delete (waiter as any).waitingProcValue;
        }
        return finalize();
      }

      if (waiter.cancelRequested) {
        if (waitState) {
          cancelAwaitRegistration(this, waitState.registration);
          delete (waiter as any).waitingProcValue;
        }
        throw new Error("Proc cancelled");
      }

      if (!waitState) {
        waitState = { target: handle };
        (waiter as any).waitingProcValue = waitState;
        const waker = this.createAwaitWaker(waiter, waitState);
        waitState.registration = this.registerProcAwaiter(handle, waker);
      }
      if (handle.state !== "pending" || waitState.wakePending) {
        cancelAwaitRegistration(this, waitState.registration);
        delete (waiter as any).waitingProcValue;
        return finalize();
      }
      waiter.awaitBlocked = true;
      return this.procYield(true);
    }

    if (handle.state === "pending") {
      if (handle.runner) {
        const runner = handle.runner;
        handle.runner = null;
        runner();
      } else {
        this.runProcHandle(handle);
      }
    }
    if (handle.state === "pending") {
      this.runProcHandle(handle);
    }
    return finalize();
  };

  cls.prototype.procHandleCancel = function procHandleCancel(this: Interpreter, handle: Extract<RuntimeValue, { kind: "proc_handle" }>): void {
    if (handle.state === "resolved" || handle.state === "failed" || handle.state === "cancelled") return;
    handle.cancelRequested = true;
    if (handle.state === "pending" && !handle.isEvaluating) {
      if (!handle.runner) handle.runner = () => this.runProcHandle(handle);
      this.scheduleAsync(handle.runner);
    }
  };

  cls.prototype.futureCancel = function futureCancel(this: Interpreter, future: Extract<RuntimeValue, { kind: "future" }>): void {
    if (future.state === "resolved" || future.state === "failed" || future.state === "cancelled") return;
    future.cancelRequested = true;
    if (future.state === "pending" && !future.isEvaluating) {
      if (!future.runner) future.runner = () => this.runFuture(future);
      this.scheduleAsync(future.runner);
    }
  };

  cls.prototype.futureValue = function futureValue(this: Interpreter, future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue {
    const finalize = (): RuntimeValue => {
      switch (future.state) {
        case "failed": {
          if (future.error) return future.error;
          const procErr = this.makeProcError("Future failed");
          return this.makeRuntimeError("Future failed", procErr, procErr);
        }
        case "cancelled": {
          if (future.error) return future.error;
          const procErr = this.makeProcError("Future cancelled");
          return this.makeRuntimeError("Future cancelled", procErr, procErr);
        }
        case "resolved":
          return future.result ?? { kind: "nil", value: null };
        case "pending": {
          const procErr = this.makeProcError("Future pending");
          return this.makeRuntimeError("Future pending", procErr, procErr);
        }
      }
    };

    const asyncCtx = this.currentAsyncContext();
    if (asyncCtx) {
      const waiter = asyncCtx.handle as AsyncHandle;
      let waitState = (waiter as any).waitingFutureValue as FutureValueWaitState | undefined;
      if (waitState && waitState.target !== future) {
        cancelAwaitRegistration(this, waitState.registration);
        delete (waiter as any).waitingFutureValue;
        waitState = undefined;
      }

      if (future.state !== "pending") {
        if (waitState) {
          cancelAwaitRegistration(this, waitState.registration);
          delete (waiter as any).waitingFutureValue;
        }
        return finalize();
      }

      if (waiter.cancelRequested) {
        if (waitState) {
          cancelAwaitRegistration(this, waitState.registration);
          delete (waiter as any).waitingFutureValue;
        }
        throw new Error("Proc cancelled");
      }

      if (!waitState) {
        waitState = { target: future };
        (waiter as any).waitingFutureValue = waitState;
        const waker = this.createAwaitWaker(waiter, waitState);
        waitState.registration = this.registerFutureAwaiter(future, waker);
      }
      if (future.state !== "pending" || waitState.wakePending) {
        cancelAwaitRegistration(this, waitState.registration);
        delete (waiter as any).waitingFutureValue;
        return finalize();
      }
      waiter.awaitBlocked = true;
      return this.procYield(true);
    }

    if (future.state === "pending") {
      if (future.runner) {
        const runner = future.runner;
        future.runner = null;
        runner();
      } else {
        this.runFuture(future);
      }
    }
    if (future.state === "pending") {
      this.runFuture(future);
    }
    return finalize();
  };

  cls.prototype.evaluateAsTask = async function evaluateAsTask(this: Interpreter, node: AST.AstNode, env: Environment = this.globals): Promise<RuntimeValue> {
    const capturedEnv = new Environment(env);
    const handle: Extract<RuntimeValue, { kind: "proc_handle" }> = {
      kind: "proc_handle",
      state: "pending",
      expression: node,
      env: capturedEnv,
      runner: null,
      errorMode: "raw",
      cancelRequested: false,
      awaitBlocked: false,
      entrypoint: true,
    };
    handle.runner = () => this.runProcHandle(handle);
    this.scheduleAsync(handle.runner);
    while (handle.state === "pending") {
      this.processScheduler(this.schedulerMaxSteps);
      if (handle.state !== "pending") break;
      const pending = typeof this.executor.pendingTasks === "function" ? this.executor.pendingTasks() : 0;
      if (pending === 0) {
        await new Promise((resolve) => setTimeout(resolve, 0));
      }
    }
    if (handle.state === "resolved") {
      return handle.result ?? { kind: "nil", value: null };
    }
    if (handle.state === "cancelled") {
      throw new RaiseSignal(handle.error ?? this.makeRuntimeError("Proc cancelled"));
    }
    throw new RaiseSignal(handle.error ?? this.makeRuntimeError("Proc failed"));
  };

  cls.prototype.runProcHandle = function runProcHandle(this: Interpreter, handle: Extract<RuntimeValue, { kind: "proc_handle" }>): void {
    if (handle.state !== "pending" || handle.isEvaluating) return;
    if (!handle.runner) {
      handle.runner = () => this.runProcHandle(handle);
    }
    if (handle.cancelRequested && !handle.hasStarted) {
      this.markProcCancelled(handle);
      return;
    }
    this.resetTimeSlice();
    handle.hasStarted = true;
    handle.isEvaluating = true;
    let procContext = handle.continuation as ProcContinuationContext | undefined;
    if (!procContext) {
      procContext = new ProcContinuationContext();
      (handle as any).continuation = procContext;
    }
    this.pushProcContext(procContext);
    this.asyncContextStack.push({ kind: "proc", handle });
    const errorMode = handle.errorMode ?? "proc";
    try {
      const value = this.evaluate(handle.expression, handle.env);
      if (handle.cancelRequested) {
        this.markProcCancelled(handle);
      } else {
        handle.result = value;
        handle.state = "resolved";
        handle.error = undefined;
        handle.failureInfo = undefined;
      }
    } catch (e) {
      if (e instanceof ExitSignal) {
        throw e;
      }
      if (e instanceof ProcYieldSignal) {
        this.manualYieldRequested = false;
        if (!handle.awaitBlocked && handle.runner) {
          this.scheduleAsync(handle.runner);
        }
      } else if (e instanceof RaiseSignal) {
        if (errorMode === "raw") {
          handle.failureInfo = this.toProcError(e.value, "Proc task failed");
          handle.error = e.value;
          handle.state = "failed";
        } else {
          const procErr = this.toProcError(e.value, "Proc task failed");
          const details = this.getProcErrorDetails(procErr);
          handle.failureInfo = procErr;
          handle.error = this.makeRuntimeError(`Proc failed: ${details}`, procErr, procErr);
          handle.state = "failed";
        }
      } else if (handle.cancelRequested) {
        const msg = e instanceof Error ? e.message : "Proc cancelled";
        this.markProcCancelled(handle, msg || "Proc cancelled");
      } else {
        const msg = e instanceof Error ? e.message : "Proc execution error";
        if (errorMode === "raw") {
          handle.failureInfo = this.makeProcError(msg);
          handle.error = this.makeRuntimeError(msg);
          handle.state = "failed";
        } else {
          const procErr = this.makeProcError(msg);
          handle.failureInfo = procErr;
          handle.error = this.makeRuntimeError(`Proc failed: ${msg}`, procErr, procErr);
          handle.state = "failed";
        }
      }
    } finally {
      this.asyncContextStack.pop();
      this.popProcContext(procContext);
      handle.isEvaluating = false;
      this.manualYieldRequested = false;
      if (handle.state !== "pending") {
        this.triggerProcAwaiters(handle);
        procContext.reset();
        delete (handle as any).continuation;
        handle.runner = null;
      } else if (!handle.runner) {
        handle.runner = () => this.runProcHandle(handle);
      }
    }
  };

  cls.prototype.runFuture = function runFuture(this: Interpreter, future: Extract<RuntimeValue, { kind: "future" }>): void {
    if (future.state !== "pending" || future.isEvaluating) return;
    if (!future.runner) {
      future.runner = () => this.runFuture(future);
    }
    if (future.cancelRequested && !future.hasStarted) {
      this.markFutureCancelled(future);
      return;
    }
    let futureContext = future.continuation as ProcContinuationContext | undefined;
    if (!futureContext) {
      futureContext = new ProcContinuationContext();
      (future as any).continuation = futureContext;
    }
    this.resetTimeSlice();
    future.hasStarted = true;
    future.isEvaluating = true;
    this.pushProcContext(futureContext);
    this.asyncContextStack.push({ kind: "future", handle: future });
    try {
      const value = this.evaluate(future.expression, future.env);
      if (future.cancelRequested) {
        const msg = "Future cancelled";
        this.markFutureCancelled(future, msg);
      } else {
        future.result = value;
        future.state = "resolved";
        future.error = undefined;
        future.failureInfo = undefined;
      }
    } catch (e) {
      if (e instanceof ExitSignal) {
        throw e;
      }
      if (e instanceof ProcYieldSignal) {
        this.manualYieldRequested = false;
        if (!future.awaitBlocked && future.runner) {
          this.scheduleAsync(future.runner);
        }
      } else if (e instanceof RaiseSignal) {
        const procErr = this.toProcError(e.value, "Future task failed");
        const details = this.getProcErrorDetails(procErr);
        future.failureInfo = procErr;
        future.error = this.makeRuntimeError(`Future failed: ${details}`, procErr, procErr);
        future.state = "failed";
      } else if (future.cancelRequested) {
        const msg = e instanceof Error ? e.message : "Future cancelled";
        this.markFutureCancelled(future, msg || "Future cancelled");
      } else {
        const msg = e instanceof Error ? e.message : "Future execution error";
        const procErr = this.makeProcError(msg);
        future.failureInfo = procErr;
        future.error = this.makeRuntimeError(`Future failed: ${msg}`, procErr, procErr);
        future.state = "failed";
      }
    } finally {
      this.asyncContextStack.pop();
      this.popProcContext(futureContext);
      future.isEvaluating = false;
      this.manualYieldRequested = false;
      if (future.state !== "pending") {
        this.triggerFutureAwaiters(future);
        futureContext.reset();
        delete (future as any).continuation;
        future.runner = null;
      } else if (!future.runner) {
        future.runner = () => this.runFuture(future);
      }
    }
  };

  cls.prototype.makeRuntimeError = function makeRuntimeError(this: Interpreter, message: string, value?: RuntimeValue, cause?: RuntimeValue): RuntimeValue {
    const err: Extract<RuntimeValue, { kind: "error" }> = { kind: "error", message };
    if (value !== undefined) {
      err.value = value;
    }
    if (cause !== undefined) {
      err.cause = cause;
    } else if (value && value.kind === "error") {
      err.cause = value;
    }
    return err;
  };

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

    const makeDefaultAwaitable = (callback?: RuntimeValue): Extract<RuntimeValue, { kind: "struct_instance" }> => {
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
      return inst;
    };

    const makeTimerAwaitable = (durationMs: number, callback?: RuntimeValue): Extract<RuntimeValue, { kind: "struct_instance" }> => {
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
      return inst;
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
