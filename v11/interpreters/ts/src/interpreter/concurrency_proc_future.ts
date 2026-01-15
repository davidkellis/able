import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import { ProcContinuationContext } from "./proc_continuations";
import { ExitSignal, ProcYieldSignal, RaiseSignal } from "./signals";
import { attachRuntimeDiagnosticContext, getRuntimeDiagnosticContext } from "./runtime_diagnostics";
import type { RuntimeValue } from "./values";
import {
  AsyncHandle,
  FutureValueWaitState,
  ProcValueWaitState,
  cancelAwaitRegistration,
} from "./concurrency_shared";

export function applyConcurrencyProcFuture(cls: typeof Interpreter): void {
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
      const signal = new RaiseSignal(handle.error ?? this.makeRuntimeError("Proc cancelled"));
      if (handle.errorContext) {
        attachRuntimeDiagnosticContext(signal, handle.errorContext);
      }
      throw signal;
    }
    const signal = new RaiseSignal(handle.error ?? this.makeRuntimeError("Proc failed"));
    if (handle.errorContext) {
      attachRuntimeDiagnosticContext(signal, handle.errorContext);
    }
    throw signal;
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
          handle.errorContext = getRuntimeDiagnosticContext(e);
        }
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
        if (process.env.ABLE_TRACE_ERRORS && e instanceof Error && e.stack) {
          console.error(e.stack);
        }
        const msg = e instanceof Error ? e.message : "Proc execution error";
        if (errorMode === "raw") {
          handle.failureInfo = this.makeProcError(msg);
          handle.error = this.makeRuntimeError(msg);
          handle.errorContext = getRuntimeDiagnosticContext(e);
          handle.state = "failed";
        } else {
          const procErr = this.makeProcError(msg);
          handle.failureInfo = procErr;
          handle.error = this.makeRuntimeError(`Proc failed: ${msg}`, procErr, procErr);
          handle.state = "failed";
        }
      }
    } finally {
      handle.isEvaluating = false;
      this.manualYieldRequested = false;
      this.asyncContextStack.pop();
      const popped = this.procContextStack.pop();
      if (popped && popped !== procContext) {
        this.procContextStack.push(popped);
      }
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
    this.resetTimeSlice();
    future.hasStarted = true;
    future.isEvaluating = true;
    let futureContext = future.continuation as ProcContinuationContext | undefined;
    if (!futureContext) {
      futureContext = new ProcContinuationContext();
      (future as any).continuation = futureContext;
    }
    this.pushProcContext(futureContext);
    this.asyncContextStack.push({ kind: "future", handle: future });
    const errorMode = future.errorMode ?? "future";
    try {
      const value = this.evaluate(future.expression, future.env);
      if (future.cancelRequested) {
        this.markFutureCancelled(future);
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
        if (errorMode === "raw") {
          future.errorContext = getRuntimeDiagnosticContext(e);
        }
        if (errorMode === "raw") {
          future.failureInfo = this.toProcError(e.value, "Future task failed");
          future.error = e.value;
          future.state = "failed";
        } else {
          const procErr = this.toProcError(e.value, "Future task failed");
          const details = this.getProcErrorDetails(procErr);
          future.failureInfo = procErr;
          future.error = this.makeRuntimeError(`Future failed: ${details}`, procErr, procErr);
          future.state = "failed";
        }
      } else if (future.cancelRequested) {
        const msg = e instanceof Error ? e.message : "Future cancelled";
        this.markFutureCancelled(future, msg || "Future cancelled");
      } else {
        if (process.env.ABLE_TRACE_ERRORS && e instanceof Error && e.stack) {
          console.error(e.stack);
        }
        const msg = e instanceof Error ? e.message : "Future execution error";
        if (errorMode === "raw") {
          future.failureInfo = this.makeProcError(msg);
          future.error = this.makeRuntimeError(msg);
          future.errorContext = getRuntimeDiagnosticContext(e);
          future.state = "failed";
        } else {
          const procErr = this.makeProcError(msg);
          future.failureInfo = procErr;
          future.error = this.makeRuntimeError(`Future failed: ${msg}`, procErr, procErr);
          future.state = "failed";
        }
      }
    } finally {
      future.isEvaluating = false;
      this.manualYieldRequested = false;
      this.asyncContextStack.pop();
      const popped = this.procContextStack.pop();
      if (popped && popped !== futureContext) {
        this.procContextStack.push(popped);
      }
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
}
