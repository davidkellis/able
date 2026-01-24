import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { ProcContinuationContext } from "./proc_continuations";
import { ProcYieldSignal } from "./signals";
import { makeIntegerValue } from "./numeric";

export function applyConcurrencyScheduler(cls: typeof Interpreter): void {
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

  cls.prototype.currentAsyncContext = function currentAsyncContext(this: Interpreter):
    | { kind: "future"; handle: Extract<RuntimeValue, { kind: "future" }> }
    | null {
    if (this.asyncContextStack.length === 0) return null;
    return this.asyncContextStack[this.asyncContextStack.length - 1];
  };

  cls.prototype.procYield = function procYield(this: Interpreter, allowEntrypoint = false): RuntimeValue {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("future_yield must be called inside an asynchronous task");
    if (!allowEntrypoint && ctx.handle.entrypoint) {
      throw new Error("future_yield must be called inside an asynchronous task");
    }
    this.manualYieldRequested = true;
    throw new ProcYieldSignal();
  };

  cls.prototype.procCancelled = function procCancelled(this: Interpreter, allowEntrypoint = false): RuntimeValue {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("future_cancelled must be called inside an asynchronous task");
    if (!allowEntrypoint && ctx.handle.entrypoint) {
      throw new Error("future_cancelled must be called inside an asynchronous task");
    }
    return { kind: "bool", value: !!ctx.handle.cancelRequested };
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
}
