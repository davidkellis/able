import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { Executor } from "./executor";
import { ProcYieldSignal, RaiseSignal } from "./signals";
import type { V10Value } from "./values";
import { ProcContinuationContext } from "./proc_continuations";

declare module "./index" {
  interface InterpreterV10 {
    initConcurrencyBuiltins(): void;
    scheduleAsync(fn: () => void): void;
    ensureSchedulerTick(): void;
    currentAsyncContext(): { kind: "proc"; handle: Extract<V10Value, { kind: "proc_handle" }> } | { kind: "future"; handle: Extract<V10Value, { kind: "future" }> } | null;
    procYield(): V10Value;
    procCancelled(): V10Value;
    procFlush(): V10Value;
    processScheduler(limit?: number): void;
    makeNamedStructInstance(def: AST.StructDefinition, entries: Array<[string, V10Value]>): V10Value;
    makeProcError(details: string): V10Value;
    getProcErrorDetails(procError: V10Value): string;
    makeProcStatusFailed(procError: V10Value): V10Value;
    markProcCancelled(handle: Extract<V10Value, { kind: "proc_handle" }>, message?: string): void;
    procHandleStatus(handle: Extract<V10Value, { kind: "proc_handle" }>): V10Value;
    futureStatus(future: Extract<V10Value, { kind: "future" }>): V10Value;
    toProcError(value: V10Value | undefined, fallback: string): V10Value;
    makeNativeFunction(name: string, arity: number, impl: (interpreter: InterpreterV10, args: V10Value[]) => V10Value): Extract<V10Value, { kind: "native_function" }>;
    bindNativeMethod(func: Extract<V10Value, { kind: "native_function" }>, self: V10Value): Extract<V10Value, { kind: "native_bound_method" }>;
    procHandleValue(handle: Extract<V10Value, { kind: "proc_handle" }>): V10Value;
    procHandleCancel(handle: Extract<V10Value, { kind: "proc_handle" }>): void;
    futureValue(future: Extract<V10Value, { kind: "future" }>): V10Value;
    runProcHandle(handle: Extract<V10Value, { kind: "proc_handle" }>): void;
    runFuture(future: Extract<V10Value, { kind: "future" }>): void;
    makeRuntimeError(message: string, value?: V10Value, cause?: V10Value): V10Value;
    executor: Executor;
    pushProcContext(ctx: ProcContinuationContext): void;
    popProcContext(ctx: ProcContinuationContext): void;
    currentProcContext(): ProcContinuationContext | null;
    procContextStack: ProcContinuationContext[];
  }
}

export function applyConcurrencyAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.initConcurrencyBuiltins = function initConcurrencyBuiltins(this: InterpreterV10): void {
    if (this.concurrencyBuiltinsInitialized) return;
    this.concurrencyBuiltinsInitialized = true;

    const procErrorDefAst = AST.structDefinition(
      "ProcError",
      [AST.structFieldDefinition(AST.simpleTypeExpression("string"), "details")],
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
  };

  cls.prototype.scheduleAsync = function scheduleAsync(this: InterpreterV10, fn: () => void): void {
    this.executor.schedule(fn);
  };

  cls.prototype.ensureSchedulerTick = function ensureSchedulerTick(this: InterpreterV10): void {
    this.executor.ensureTick();
  };

  cls.prototype.pushProcContext = function pushProcContext(this: InterpreterV10, ctx: ProcContinuationContext): void {
    if (!this.procContextStack) {
      this.procContextStack = [];
    }
    this.procContextStack.push(ctx);
  };

  cls.prototype.popProcContext = function popProcContext(this: InterpreterV10, ctx: ProcContinuationContext): void {
    if (!this.procContextStack || this.procContextStack.length === 0) return;
    const top = this.procContextStack[this.procContextStack.length - 1];
    if (top === ctx) {
      this.procContextStack.pop();
    }
  };

  cls.prototype.currentProcContext = function currentProcContext(this: InterpreterV10): ProcContinuationContext | null {
    if (!this.procContextStack || this.procContextStack.length === 0) return null;
    return this.procContextStack[this.procContextStack.length - 1]!;
  };

  cls.prototype.currentAsyncContext = function currentAsyncContext(this: InterpreterV10): { kind: "proc"; handle: Extract<V10Value, { kind: "proc_handle" }> } | { kind: "future"; handle: Extract<V10Value, { kind: "future" }> } | null {
    if (this.asyncContextStack.length === 0) return null;
    return this.asyncContextStack[this.asyncContextStack.length - 1];
  };

  cls.prototype.procYield = function procYield(this: InterpreterV10): V10Value {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_yield must be called inside an asynchronous task");
    this.manualYieldRequested = true;
    throw new ProcYieldSignal();
  };

  cls.prototype.procCancelled = function procCancelled(this: InterpreterV10): V10Value {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_cancelled must be called inside an asynchronous task");
    if (ctx.kind === "proc") {
      return { kind: "bool", value: !!ctx.handle.cancelRequested };
    }
    return { kind: "bool", value: false };
  };

  cls.prototype.procFlush = function procFlush(this: InterpreterV10): V10Value {
    this.processScheduler(this.schedulerMaxSteps);
    return { kind: "nil", value: null };
  };

  cls.prototype.processScheduler = function processScheduler(this: InterpreterV10, limit: number = this.schedulerMaxSteps): void {
    this.executor.flush(limit);
  };

  cls.prototype.makeNamedStructInstance = function makeNamedStructInstance(this: InterpreterV10, def: AST.StructDefinition, entries: Array<[string, V10Value]>): V10Value {
    const map = new Map<string, V10Value>();
    for (const [key, value] of entries) map.set(key, value);
    return { kind: "struct_instance", def, values: map };
  };

  cls.prototype.makeProcError = function makeProcError(this: InterpreterV10, details: string): V10Value {
    return this.makeNamedStructInstance(this.procErrorStruct, [["details", { kind: "string", value: details }]]);
  };

  cls.prototype.getProcErrorDetails = function getProcErrorDetails(this: InterpreterV10, procError: V10Value): string {
    if (procError.kind === "struct_instance" && procError.def.id.name === "ProcError") {
      const map = procError.values as Map<string, V10Value>;
      const detailsVal = map.get("details");
      if (detailsVal && detailsVal.kind === "string") return detailsVal.value;
    }
    return "unknown failure";
  };

  cls.prototype.makeProcStatusFailed = function makeProcStatusFailed(this: InterpreterV10, procError: V10Value): V10Value {
    return this.makeNamedStructInstance(this.procStatusStructs.Failed, [["error", procError]]);
  };

  cls.prototype.markProcCancelled = function markProcCancelled(this: InterpreterV10, handle: Extract<V10Value, { kind: "proc_handle" }>, message = "Proc cancelled"): void {
    const pendingSend = (handle as any).waitingChannelSend as
      | {
          state: any;
          value: V10Value;
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
  };

  cls.prototype.procHandleStatus = function procHandleStatus(this: InterpreterV10, handle: Extract<V10Value, { kind: "proc_handle" }>): V10Value {
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

  cls.prototype.futureStatus = function futureStatus(this: InterpreterV10, future: Extract<V10Value, { kind: "future" }>): V10Value {
    switch (future.state) {
      case "pending":
        return this.procStatusPendingValue;
      case "resolved":
        return this.procStatusResolvedValue;
      case "failed": {
        const procErr = future.failureInfo ?? this.makeProcError("unknown failure");
        return this.makeProcStatusFailed(procErr);
      }
    }
  };

  cls.prototype.toProcError = function toProcError(this: InterpreterV10, value: V10Value | undefined, fallback: string): V10Value {
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

  cls.prototype.makeNativeFunction = function makeNativeFunction(this: InterpreterV10, name: string, arity: number, impl: (interpreter: InterpreterV10, args: V10Value[]) => V10Value): Extract<V10Value, { kind: "native_function" }> {
    return { kind: "native_function", name, arity, impl };
  };

  cls.prototype.bindNativeMethod = function bindNativeMethod(this: InterpreterV10, func: Extract<V10Value, { kind: "native_function" }>, self: V10Value): Extract<V10Value, { kind: "native_bound_method" }> {
    return { kind: "native_bound_method", func, self };
  };

  cls.prototype.procHandleValue = function procHandleValue(this: InterpreterV10, handle: Extract<V10Value, { kind: "proc_handle" }>): V10Value {
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

  cls.prototype.procHandleCancel = function procHandleCancel(this: InterpreterV10, handle: Extract<V10Value, { kind: "proc_handle" }>): void {
    if (handle.state === "resolved" || handle.state === "failed" || handle.state === "cancelled") return;
    handle.cancelRequested = true;
    if (handle.state === "pending" && !handle.isEvaluating) {
      if (!handle.runner) handle.runner = () => this.runProcHandle(handle);
      this.scheduleAsync(handle.runner);
    }
  };

  cls.prototype.futureValue = function futureValue(this: InterpreterV10, future: Extract<V10Value, { kind: "future" }>): V10Value {
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
    switch (future.state) {
      case "failed": {
        if (future.error) return future.error;
        const procErr = this.makeProcError("Future failed");
        return this.makeRuntimeError("Future failed", procErr, procErr);
      }
      case "resolved":
        return future.result ?? { kind: "nil", value: null };
      case "pending": {
        const procErr = this.makeProcError("Future pending");
        return this.makeRuntimeError("Future pending", procErr, procErr);
      }
    }
  };

  cls.prototype.runProcHandle = function runProcHandle(this: InterpreterV10, handle: Extract<V10Value, { kind: "proc_handle" }>): void {
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
      if (e instanceof ProcYieldSignal) {
        const manualYield = this.manualYieldRequested;
        this.manualYieldRequested = false;
        if (manualYield) {
          procContext.reset();
        }
        if (handle.runner) {
          this.scheduleAsync(handle.runner);
        }
      } else if (e instanceof RaiseSignal) {
        const procErr = this.toProcError(e.value, "Proc task failed");
        const details = this.getProcErrorDetails(procErr);
        handle.failureInfo = procErr;
        handle.error = this.makeRuntimeError(`Proc failed: ${details}`, procErr, procErr);
        handle.state = "failed";
      } else if (handle.cancelRequested) {
        const msg = e instanceof Error ? e.message : "Proc cancelled";
        this.markProcCancelled(handle, msg || "Proc cancelled");
      } else {
        const msg = e instanceof Error ? e.message : "Proc execution error";
        const procErr = this.makeProcError(msg);
        handle.failureInfo = procErr;
        handle.error = this.makeRuntimeError(`Proc failed: ${msg}`, procErr, procErr);
        handle.state = "failed";
      }
    } finally {
      this.asyncContextStack.pop();
      this.popProcContext(procContext);
      handle.isEvaluating = false;
      this.manualYieldRequested = false;
      if (handle.state !== "pending") {
        procContext.reset();
        delete (handle as any).continuation;
        handle.runner = null;
      } else if (!handle.runner) {
        handle.runner = () => this.runProcHandle(handle);
      }
    }
  };

  cls.prototype.runFuture = function runFuture(this: InterpreterV10, future: Extract<V10Value, { kind: "future" }>): void {
    if (future.state !== "pending" || future.isEvaluating) return;
    if (!future.runner) {
      future.runner = () => this.runFuture(future);
    }
    let futureContext = future.continuation as ProcContinuationContext | undefined;
    if (!futureContext) {
      futureContext = new ProcContinuationContext();
      (future as any).continuation = futureContext;
    }
    this.resetTimeSlice();
    future.isEvaluating = true;
    this.pushProcContext(futureContext);
    this.asyncContextStack.push({ kind: "future", handle: future });
    try {
      const value = this.evaluate(future.expression, future.env);
      future.result = value;
      future.state = "resolved";
      future.error = undefined;
      future.failureInfo = undefined;
    } catch (e) {
      if (e instanceof ProcYieldSignal) {
        const manualYield = this.manualYieldRequested;
        this.manualYieldRequested = false;
        if (manualYield) {
          futureContext.reset();
        }
        if (future.runner) {
          this.scheduleAsync(future.runner);
        }
      } else if (e instanceof RaiseSignal) {
        const procErr = this.toProcError(e.value, "Future task failed");
        const details = this.getProcErrorDetails(procErr);
        future.failureInfo = procErr;
        future.error = this.makeRuntimeError(`Future failed: ${details}`, procErr, procErr);
        future.state = "failed";
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
        futureContext.reset();
        delete (future as any).continuation;
        future.runner = null;
      } else if (!future.runner) {
        future.runner = () => this.runFuture(future);
      }
    }
  };

  cls.prototype.makeRuntimeError = function makeRuntimeError(this: InterpreterV10, message: string, value?: V10Value, cause?: V10Value): V10Value {
    const err: Extract<V10Value, { kind: "error" }> = { kind: "error", message };
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
