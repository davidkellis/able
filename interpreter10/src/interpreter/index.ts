import * as AST from "../ast";
import { applyHelperAugmentations } from "./helpers";
import { applyOperationsAugmentations } from "./operations";
import { applyStringifyAugmentations } from "./stringify";
import { applyPatternAugmentations } from "./patterns";
import { applyTypesAugmentations } from "./types";
import { applyMemberAugmentations } from "./members";
import { applyImplResolutionAugmentations } from "./impl_resolution";
import { applyPlaceholderAugmentations } from "./placeholders";
import { applyEvaluationAugmentations } from "./eval_expressions";
import { applyConcurrencyAugmentations } from "./concurrency";
import { applyIteratorAugmentations } from "./iterators";
import { applyChannelMutexAugmentations } from "./channels_mutex";
import "./definitions";
import "./imports";

import { Environment } from "./environment";
import type { ImplMethodEntry, V10Value, ConstraintSpec } from "./values";
import { CooperativeExecutor, type Executor } from "./executor";
import { ProcYieldSignal } from "./signals";
import type { ProcContinuationContext } from "./proc_continuations";

// =============================================================================
// v10 Interpreter (modular layout)
// =============================================================================

export type InterpreterOptions = {
  executor?: Executor;
  schedulerMaxSteps?: number;
};

export class InterpreterV10 {
  readonly globals = new Environment();

  interfaces: Map<string, AST.InterfaceDefinition> = new Map();
  interfaceEnvs: Map<string, Environment> = new Map();
  inherentMethods: Map<string, Map<string, Extract<V10Value, { kind: "function" }>>> = new Map();
  implMethods: Map<string, ImplMethodEntry[]> = new Map();
  unnamedImplsSeen: Map<string, Map<string, Set<string>>> = new Map();
  raiseStack: V10Value[] = [];
  packageRegistry: Map<string, Map<string, V10Value>> = new Map();
  currentPackage: string | null = null;
  breakpointStack: string[] = [];
  implicitReceiverStack: V10Value[] = [];
  topicStack: V10Value[] = [];
  topicUsageStack: boolean[] = [];
  placeholderFrames: PlaceholderFrame[] = [];

  procNativeMethods!: {
    status: Extract<V10Value, { kind: "native_function" }>;
    value: Extract<V10Value, { kind: "native_function" }>;
    cancel: Extract<V10Value, { kind: "native_function" }>;
  };

  futureNativeMethods!: {
    status: Extract<V10Value, { kind: "native_function" }>;
    value: Extract<V10Value, { kind: "native_function" }>;
  };

  concurrencyBuiltinsInitialized = false;
  procErrorStruct!: AST.StructDefinition;
  procStatusStructs!: {
    Pending: AST.StructDefinition;
    Resolved: AST.StructDefinition;
    Cancelled: AST.StructDefinition;
    Failed: AST.StructDefinition;
  };
  procStatusPendingValue!: V10Value;
  procStatusResolvedValue!: V10Value;
  procStatusCancelledValue!: V10Value;

  channelMutexBuiltinsInitialized = false;
  nextChannelHandle = 1;
  channelStates: Map<number, any> = new Map();
  nextMutexHandle = 1;
  mutexStates: Map<number, any> = new Map();

  schedulerMaxSteps = 1024;
  executor: Executor;
  timeSliceCounter = 0;
  manualYieldRequested = false;
  asyncContextStack: Array<
    { kind: "proc"; handle: Extract<V10Value, { kind: "proc_handle" }> } |
    { kind: "future"; handle: Extract<V10Value, { kind: "future" }> }
  > = [];
  procContextStack: ProcContinuationContext[] = [];

  constructor(options: InterpreterOptions = {}) {
    if (options.schedulerMaxSteps !== undefined) {
      this.schedulerMaxSteps = options.schedulerMaxSteps;
    }
    this.executor = options.executor ?? new CooperativeExecutor({ maxSteps: this.schedulerMaxSteps });
    this.initConcurrencyBuiltins();
    this.ensureChannelMutexBuiltins();
    this.procNativeMethods = {
      status: this.makeNativeFunction("Proc.status", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "proc_handle") throw new Error("Proc.status called on non-proc handle");
        return interp.procHandleStatus(self);
      }),
      value: this.makeNativeFunction("Proc.value", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "proc_handle") throw new Error("Proc.value called on non-proc handle");
        return interp.procHandleValue(self);
      }),
      cancel: this.makeNativeFunction("Proc.cancel", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "proc_handle") throw new Error("Proc.cancel called on non-proc handle");
        interp.procHandleCancel(self);
        return { kind: "nil", value: null };
      }),
    };

    this.futureNativeMethods = {
      status: this.makeNativeFunction("Future.status", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "future") throw new Error("Future.status called on non-future");
        return interp.futureStatus(self);
      }),
      value: this.makeNativeFunction("Future.value", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "future") throw new Error("Future.value called on non-future");
        return interp.futureValue(self);
      }),
    };

    const procYieldFn = this.makeNativeFunction("proc_yield", 0, (interp) => interp.procYield());
    const procCancelledFn = this.makeNativeFunction("proc_cancelled", 0, (interp) => interp.procCancelled());
    const procFlushFn = this.makeNativeFunction("proc_flush", 0, (interp) => interp.procFlush());
    this.globals.define("proc_yield", procYieldFn);
    this.globals.define("proc_cancelled", procCancelledFn);
    this.globals.define("proc_flush", procFlushFn);
  }

  resetTimeSlice(): void {
    this.timeSliceCounter = 0;
  }

  checkTimeSlice(): void {
    if (this.asyncContextStack.length === 0) return;
    this.timeSliceCounter += 1;
    if (this.timeSliceCounter >= this.schedulerMaxSteps) {
      this.timeSliceCounter = 0;
      throw new ProcYieldSignal();
    }
  }
}

applyHelperAugmentations(InterpreterV10);
applyOperationsAugmentations(InterpreterV10);
applyStringifyAugmentations(InterpreterV10);
applyPatternAugmentations(InterpreterV10);
applyTypesAugmentations(InterpreterV10);
applyMemberAugmentations(InterpreterV10);
applyImplResolutionAugmentations(InterpreterV10);
applyPlaceholderAugmentations(InterpreterV10);
applyIteratorAugmentations(InterpreterV10);
applyChannelMutexAugmentations(InterpreterV10);
applyEvaluationAugmentations(InterpreterV10);
applyConcurrencyAugmentations(InterpreterV10);

export type { ConstraintSpec as InterpreterConstraintSpec } from "./values";

export { Environment } from "./environment";
export type { V10Value } from "./values";
export type { Executor } from "./executor";
export { CooperativeExecutor } from "./executor";

export type { PlaceholderFrame } from "./placeholders";
// Side-effectful module imports attach feature-specific behaviour to InterpreterV10.

export function evaluate(node: AST.AstNode | null, env?: Environment): V10Value {
  return new InterpreterV10().evaluate(node, env);
}
