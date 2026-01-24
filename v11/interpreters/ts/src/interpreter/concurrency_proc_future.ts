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
  cancelAwaitRegistration,
} from "./concurrency_shared";

export function applyConcurrencyProcFuture(cls: typeof Interpreter): void {
  cls.prototype.makeNamedStructInstance = function makeNamedStructInstance(this: Interpreter, def: AST.StructDefinition, entries: Array<[string, RuntimeValue]>): RuntimeValue {
    const map = new Map<string, RuntimeValue>();
    for (const [key, value] of entries) map.set(key, value);
    return { kind: "struct_instance", def, values: map };
  };

  cls.prototype.makeFutureError = function makeFutureError(this: Interpreter, details: string): RuntimeValue {
    return this.makeNamedStructInstance(this.futureErrorStruct, [["details", { kind: "String", value: details }]]);
  };

  cls.prototype.getFutureErrorDetails = function getFutureErrorDetails(this: Interpreter, procError: RuntimeValue): string {
    if (procError.kind === "struct_instance" && procError.def.id.name === "FutureError") {
      const map = procError.values as Map<string, RuntimeValue>;
      const detailsVal = map.get("details");
      if (detailsVal && detailsVal.kind === "String") return detailsVal.value;
    }
    return "unknown failure";
  };

  cls.prototype.makeFutureStatusFailed = function makeFutureStatusFailed(this: Interpreter, procError: RuntimeValue): RuntimeValue {
    return this.makeNamedStructInstance(this.futureStatusStructs.Failed, [["error", procError]]);
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
    const procErr = this.makeFutureError(message);
    future.state = "cancelled";
    future.result = undefined;
    future.failureInfo = procErr;
    future.error = this.makeRuntimeError(message, procErr, procErr);
    future.runner = null;
    future.awaitBlocked = false;
    this.triggerFutureAwaiters(future);
  };

  cls.prototype.futureStatus = function futureStatus(this: Interpreter, future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue {
    switch (future.state) {
      case "pending":
        return this.futureStatusPendingValue;
      case "resolved":
        return this.futureStatusResolvedValue;
      case "cancelled":
        return this.futureStatusCancelledValue;
      case "failed": {
        const procErr = future.failureInfo ?? this.makeFutureError("unknown failure");
        return this.makeFutureStatusFailed(procErr);
      }
    }
  };

  cls.prototype.toFutureError = function toFutureError(this: Interpreter, value: RuntimeValue | undefined, fallback: string): RuntimeValue {
    if (value) {
      if (value.kind === "struct_instance" && value.def.id.name === "FutureError") {
        return value;
      }
      if (value.kind === "error") {
        if (value.cause && value.cause.kind === "struct_instance" && value.cause.def.id.name === "FutureError") {
          return value.cause;
        }
        if (value.value && value.value.kind === "struct_instance" && value.value.def.id.name === "FutureError") {
          return value.value;
        }
        return this.makeFutureError(value.message ?? fallback);
      }
      return this.makeFutureError(this.valueToString(value));
    }
    return this.makeFutureError(fallback);
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
          const procErr = this.makeFutureError("Future failed");
          return this.makeRuntimeError("Future failed", procErr, procErr);
        }
        case "cancelled": {
          if (future.error) return future.error;
          const procErr = this.makeFutureError("Future cancelled");
          return this.makeRuntimeError("Future cancelled", procErr, procErr);
        }
        case "resolved":
          return future.result ?? { kind: "nil", value: null };
        case "pending": {
          const procErr = this.makeFutureError("Future pending");
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
        throw new Error("Future cancelled");
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
    const future: Extract<RuntimeValue, { kind: "future" }> = {
      kind: "future",
      state: "pending",
      expression: node,
      env: capturedEnv,
      runner: null,
      errorMode: "raw",
      cancelRequested: false,
      awaitBlocked: false,
      entrypoint: true,
    };
    future.runner = () => this.runFuture(future);
    this.scheduleAsync(future.runner);
    while (future.state === "pending") {
      this.processScheduler(this.schedulerMaxSteps);
      if (future.state !== "pending") break;
      const pending = typeof this.executor.pendingTasks === "function" ? this.executor.pendingTasks() : 0;
      if (pending === 0) {
        await new Promise((resolve) => setTimeout(resolve, 0));
      }
    }
    if (future.state === "resolved") {
      return future.result ?? { kind: "nil", value: null };
    }
    if (future.state === "cancelled") {
      const signal = new RaiseSignal(future.error ?? this.makeRuntimeError("Future cancelled"));
      if (future.errorContext) {
        attachRuntimeDiagnosticContext(signal, future.errorContext);
      }
      throw signal;
    }
    const signal = new RaiseSignal(future.error ?? this.makeRuntimeError("Future failed"));
    if (future.errorContext) {
      attachRuntimeDiagnosticContext(signal, future.errorContext);
    }
    throw signal;
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
          future.failureInfo = this.toFutureError(e.value, "Future task failed");
          future.error = e.value;
          future.state = "failed";
        } else {
          const procErr = this.toFutureError(e.value, "Future task failed");
          const details = this.getFutureErrorDetails(procErr);
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
          future.failureInfo = this.makeFutureError(msg);
          future.error = this.makeRuntimeError(msg);
          future.errorContext = getRuntimeDiagnosticContext(e);
          future.state = "failed";
        } else {
          const procErr = this.makeFutureError(msg);
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
