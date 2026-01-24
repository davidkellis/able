import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Executor } from "./executor";
import type { Interpreter } from "./index";
import { isFloatValue, isIntegerValue } from "./numeric";
import { ProcContinuationContext } from "./proc_continuations";
import { callCallableValue } from "./functions";
import { memberAccessOnValue } from "./structs";
import type { RuntimeValue } from "./values";

export type AwaitWakerPayload = {
  wake: () => void;
};

export type AsyncHandle = Extract<RuntimeValue, { kind: "future" }>;

export type FutureValueWaitState = {
  target: Extract<RuntimeValue, { kind: "future" }>;
  registration?: RuntimeValue;
  wakePending?: boolean;
};

export type TimerAwaitableState = {
  deadlineMs: number;
  ready: boolean;
  cancelled: boolean;
  timerId?: ReturnType<typeof setTimeout>;
  callback?: RuntimeValue;
};

export const MAX_SLEEP_DELAY_MS = 2_147_483_647;

export function toMilliseconds(value: RuntimeValue): number {
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

export function cancelAwaitRegistration(interp: Interpreter, registration: RuntimeValue | undefined): void {
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
    currentAsyncContext(): { kind: "future"; handle: Extract<RuntimeValue, { kind: "future" }> } | null;
    createAwaitWaker(handle: AsyncHandle, state: unknown): Extract<RuntimeValue, { kind: "struct_instance" }>;
    makeAwaitRegistration(cancelFn?: () => void): Extract<RuntimeValue, { kind: "struct_instance" }>;
    invokeAwaitWaker(waker: RuntimeValue): void;
    registerFutureAwaiter(
      future: Extract<RuntimeValue, { kind: "future" }>,
      waker: Extract<RuntimeValue, { kind: "struct_instance" }>,
    ): Extract<RuntimeValue, { kind: "struct_instance" }>;
    triggerFutureAwaiters(future: Extract<RuntimeValue, { kind: "future" }>): void;
    procYield(allowEntrypoint?: boolean): RuntimeValue;
    procCancelled(allowEntrypoint?: boolean): RuntimeValue;
    procFlush(): RuntimeValue;
    procPendingTasks(): RuntimeValue;
    processScheduler(limit?: number): void;
    makeNamedStructInstance(def: AST.StructDefinition, entries: Array<[string, RuntimeValue]>): RuntimeValue;
    makeFutureError(details: string): RuntimeValue;
    getFutureErrorDetails(procError: RuntimeValue): string;
    makeFutureStatusFailed(procError: RuntimeValue): RuntimeValue;
    markFutureCancelled(future: Extract<RuntimeValue, { kind: "future" }>, message?: string): void;
    futureStatus(future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue;
    toFutureError(value: RuntimeValue | undefined, fallback: string): RuntimeValue;
    makeNativeFunction(
      name: string,
      arity: number,
      impl: (interpreter: Interpreter, args: RuntimeValue[]) => RuntimeValue | Promise<RuntimeValue>,
    ): Extract<RuntimeValue, { kind: "native_function" }>;
    bindNativeMethod(
      func: Extract<RuntimeValue, { kind: "native_function" }>,
      self: RuntimeValue,
    ): Extract<RuntimeValue, { kind: "native_bound_method" }>;
    futureCancel(future: Extract<RuntimeValue, { kind: "future" }>): void;
    futureValue(future: Extract<RuntimeValue, { kind: "future" }>): RuntimeValue;
    evaluateAsTask(node: AST.AstNode, env?: Environment): Promise<RuntimeValue>;
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
