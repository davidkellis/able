import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { ProcYieldSignal, RaiseSignal } from "./signals";
import type { V10Value } from "./values";

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
    makeRuntimeError(message: string, value?: V10Value): V10Value;
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
    this.schedulerQueue.push(fn);
    this.ensureSchedulerTick();
  };

  cls.prototype.ensureSchedulerTick = function ensureSchedulerTick(this: InterpreterV10): void {
    if (this.schedulerScheduled || this.schedulerActive) return;
    this.schedulerScheduled = true;
    const runner = () => this.processScheduler();
    if (typeof queueMicrotask === "function") {
      queueMicrotask(runner);
    } else if (typeof setTimeout === "function") {
      setTimeout(runner, 0);
    } else {
      runner();
    }
  };

  cls.prototype.currentAsyncContext = function currentAsyncContext(this: InterpreterV10): { kind: "proc"; handle: Extract<V10Value, { kind: "proc_handle" }> } | { kind: "future"; handle: Extract<V10Value, { kind: "future" }> } | null {
    if (this.asyncContextStack.length === 0) return null;
    return this.asyncContextStack[this.asyncContextStack.length - 1];
  };

  cls.prototype.procYield = function procYield(this: InterpreterV10): V10Value {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_yield must be called inside an asynchronous task");
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
    if (this.schedulerActive) return;
    this.schedulerActive = true;
    this.schedulerScheduled = false;
    let steps = 0;
    while (this.schedulerQueue.length > 0 && steps < limit) {
      const task = this.schedulerQueue.shift()!;
      task();
      steps += 1;
    }
    this.schedulerActive = false;
    if (this.schedulerQueue.length > 0) this.ensureSchedulerTick();
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
    const procErr = this.makeProcError(message);
    handle.state = "cancelled";
    handle.result = undefined;
    handle.failureInfo = procErr;
    handle.error = this.makeRuntimeError(message, procErr);
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
      case "failed":
        return handle.error ?? this.makeRuntimeError("Proc failed", this.makeProcError("Proc failed"));
      case "cancelled":
        return handle.error ?? this.makeRuntimeError("Proc cancelled", this.makeProcError("Proc cancelled"));
      default:
        return this.makeRuntimeError("Proc pending", this.makeProcError("Proc pending"));
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
      case "failed":
        return future.error ?? this.makeRuntimeError("Future failed", this.makeProcError("Future failed"));
      case "resolved":
        return future.result ?? { kind: "nil", value: null };
      case "pending":
        return this.makeRuntimeError("Future pending", this.makeProcError("Future pending"));
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
    handle.hasStarted = true;
    handle.isEvaluating = true;
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
        if (handle.runner) {
          this.scheduleAsync(handle.runner);
        }
      } else if (e instanceof RaiseSignal) {
        const procErr = this.toProcError(e.value, "Proc task failed");
        const details = this.getProcErrorDetails(procErr);
        handle.failureInfo = procErr;
        handle.error = this.makeRuntimeError(`Proc failed: ${details}`, procErr);
        handle.state = "failed";
      } else {
        const msg = e instanceof Error ? e.message : "Proc execution error";
        const procErr = this.makeProcError(msg);
        handle.failureInfo = procErr;
        handle.error = this.makeRuntimeError(`Proc failed: ${msg}`, procErr);
        handle.state = "failed";
      }
    } finally {
      this.asyncContextStack.pop();
      handle.isEvaluating = false;
      if (handle.state !== "pending") {
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
    future.isEvaluating = true;
    this.asyncContextStack.push({ kind: "future", handle: future });
    try {
      const value = this.evaluate(future.expression, future.env);
      future.result = value;
      future.state = "resolved";
      future.error = undefined;
      future.failureInfo = undefined;
    } catch (e) {
      if (e instanceof ProcYieldSignal) {
        if (future.runner) {
          this.scheduleAsync(future.runner);
        }
      } else if (e instanceof RaiseSignal) {
        const procErr = this.toProcError(e.value, "Future task failed");
        const details = this.getProcErrorDetails(procErr);
        future.failureInfo = procErr;
        future.error = this.makeRuntimeError(`Future failed: ${details}`, procErr);
        future.state = "failed";
      } else {
        const msg = e instanceof Error ? e.message : "Future execution error";
        const procErr = this.makeProcError(msg);
        future.failureInfo = procErr;
        future.error = this.makeRuntimeError(`Future failed: ${msg}`, procErr);
        future.state = "failed";
      }
    } finally {
      this.asyncContextStack.pop();
      future.isEvaluating = false;
      if (future.state !== "pending") {
        future.runner = null;
      } else if (!future.runner) {
        future.runner = () => this.runFuture(future);
      }
    }
  };

  cls.prototype.makeRuntimeError = function makeRuntimeError(this: InterpreterV10, message: string, value?: V10Value): V10Value {
    return { kind: "error", message, value };
  };
}
