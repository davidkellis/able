import * as AST from "../ast";
import { applyHelperAugmentations } from "./helpers";
import { applyOperationsAugmentations } from "./operations";
import { applyStringifyAugmentations } from "./stringify";
import { applyPatternAugmentations } from "./patterns";
import { applyTypesAugmentations } from "./types";
import { applyMemberAugmentations } from "./members";
import { applyImplResolutionAugmentations } from "./impl_resolution";
import { applyRangeAugmentations, type RangeImplementationRecord } from "./range";
import { applyPlaceholderAugmentations } from "./placeholders";
import { applyEvaluationAugmentations } from "./eval_expressions";
import { applyConcurrencyAugmentations } from "./concurrency";
import { applyIteratorAugmentations } from "./iterators";
import { applyChannelMutexAugmentations } from "./channels_mutex";
import { applyStringHostAugmentations } from "./string_host";
import { applyNumericHostAugmentations } from "./numeric_host";
import { applyOsHostAugmentations } from "./os_host";
import { applyDynamicAugmentations } from "./dynamic";
import { applyExternHostAugmentations } from "./extern_host";
import { buildStandardInterfaceBuiltins } from "../builtins/interfaces";
import { applyArrayKernelAugmentations, type ArrayState } from "./array_kernel";
import { applyHashMapKernelAugmentations, type HashMapState } from "./hash_map_kernel";
import { evaluateImplementationDefinition, evaluateInterfaceDefinition } from "./definitions";
import "./imports";

import { Environment } from "./environment";
import type { ImplMethodEntry, RuntimeValue, ConstraintSpec } from "./values";
import { CooperativeExecutor, type Executor } from "./executor";
import { ProcYieldSignal } from "./signals";
import type { ProcContinuationContext } from "./proc_continuations";
import type { RuntimeCallFrame } from "./runtime_diagnostics";

// =============================================================================
// Interpreter (modular layout)
// =============================================================================

export type InterpreterOptions = {
  executor?: Executor;
  schedulerMaxSteps?: number;
  args?: string[];
};

export class Interpreter {
  readonly globals = new Environment();

  structs: Map<string, AST.StructDefinition> = new Map();
  interfaces: Map<string, AST.InterfaceDefinition> = new Map();
  unions: Map<string, AST.UnionDefinition> = new Map();
  typeAliases: Map<string, AST.TypeAliasDefinition> = new Map();
  interfaceEnvs: Map<string, Environment> = new Map();
  inherentMethods: Map<string, Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>> = new Map();
  implMethods: Map<string, ImplMethodEntry[]> = new Map();
  genericImplMethods: ImplMethodEntry[] = [];
  rangeImplementations: RangeImplementationRecord[] = [];
  unnamedImplsSeen: Map<string, Map<string, Set<string>>> = new Map();
  implDuplicateAllowlist: Set<string> = new Set(["Error::FutureError", "Clone::String", "Clone::Grapheme"]);
  raiseStack: RuntimeValue[] = [];
  packageRegistry: Map<string, Map<string, RuntimeValue>> = new Map();
  currentPackage: string | null = null;
  breakpointStack: string[] = [];
  implicitReceiverStack: RuntimeValue[] = [];
  placeholderFrames: PlaceholderFrame[] = [];
  runtimeCallStack: RuntimeCallFrame[] = [];

  futureNativeMethods!: {
    status: Extract<RuntimeValue, { kind: "native_function" }>;
    value: Extract<RuntimeValue, { kind: "native_function" }>;
    cancel: Extract<RuntimeValue, { kind: "native_function" }>;
    is_ready: Extract<RuntimeValue, { kind: "native_function" }>;
    register: Extract<RuntimeValue, { kind: "native_function" }>;
    commit: Extract<RuntimeValue, { kind: "native_function" }>;
    is_default: Extract<RuntimeValue, { kind: "native_function" }>;
  };

  errorNativeMethods!: {
    message: Extract<RuntimeValue, { kind: "native_function" }>;
    cause: Extract<RuntimeValue, { kind: "native_function" }>;
  };

  arrayNativeMethods: Record<string, Extract<RuntimeValue, { kind: "native_function" }>> = {};

  arrayBuiltinsInitialized = false;
  nextArrayHandle = 1;
  arrayStates: Map<number, ArrayState> = new Map();
  hashMapBuiltinsInitialized = false;
  nextHashMapHandle = 1;
  hashMapStates: Map<number, HashMapState> = new Map();

  concurrencyBuiltinsInitialized = false;
  futureErrorStruct!: AST.StructDefinition;
  futureStatusStructs!: {
    Pending: AST.StructDefinition;
    Resolved: AST.StructDefinition;
    Cancelled: AST.StructDefinition;
    Failed: AST.StructDefinition;
  };
  futureStatusPendingValue!: RuntimeValue;
  futureStatusResolvedValue!: RuntimeValue;
  futureStatusCancelledValue!: RuntimeValue;
  awaitHelpersBuiltinsInitialized = false;

  channelMutexBuiltinsInitialized = false;
  stringHostBuiltinsInitialized = false;
  numericBuiltinsInitialized = false;
  osBuiltinsInitialized = false;
  nextChannelHandle = 1;
  channelStates: Map<number, any> = new Map();
  channelErrorStructs: Map<string, AST.StructDefinition> = new Map();
  standardErrorStructs: Map<string, AST.StructDefinition> = new Map();
  nextMutexHandle = 1;
  mutexStates: Map<number, any> = new Map();
  osArgs: string[] = [];

  externHostPackages: Map<string, any> = new Map();

  dynamicBuiltinsInitialized = false;
  dynPackageDefMethod!: Extract<RuntimeValue, { kind: "native_function" }>;
  dynPackageEvalMethod!: Extract<RuntimeValue, { kind: "native_function" }>;
  dynamicDefinitionMode = false;
  dynamicPackageEnvs: Map<string, Environment> = new Map();

  schedulerMaxSteps = 1024;
  executor: Executor;
  timeSliceCounter = 0;
  manualYieldRequested = false;
  asyncContextStack: Array<{ kind: "future"; handle: Extract<RuntimeValue, { kind: "future" }> }> = [];
  procContextStack: ProcContinuationContext[] = [];
  awaitRoundRobinIndex = 0;

  constructor(options: InterpreterOptions = {}) {
    if (options.schedulerMaxSteps !== undefined) {
      this.schedulerMaxSteps = options.schedulerMaxSteps;
    }
    this.osArgs = options.args ? [...options.args] : [];
    this.executor = options.executor ?? new CooperativeExecutor({ maxSteps: this.schedulerMaxSteps });
    this.initConcurrencyBuiltins();
    this.ensureChannelMutexBuiltins();
    this.ensureStringHostBuiltins();
    this.ensureNumericBuiltins();
    this.ensureOsBuiltins();
    this.ensureDynamicBuiltins();
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
      cancel: this.makeNativeFunction("Future.cancel", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "future") throw new Error("Future.cancel called on non-future");
        interp.futureCancel(self);
        return { kind: "nil", value: null };
      }),
      is_ready: this.makeNativeFunction("Future.is_ready", 1, (_interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "future") throw new Error("Future.is_ready called on non-future");
        return { kind: "bool", value: self.state !== "pending" };
      }),
      register: this.makeNativeFunction("Future.register", 2, (interp, args) => {
        const self = args[0];
        const waker = args[1];
        if (!self || self.kind !== "future") throw new Error("Future.register called on non-future");
        if (!waker || waker.kind !== "struct_instance") throw new Error("Future.register expects AwaitWaker");
        return interp.registerFutureAwaiter(self, waker);
      }),
      commit: this.makeNativeFunction("Future.commit", 1, (interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "future") throw new Error("Future.commit called on non-future");
        return interp.futureValue(self);
      }),
      is_default: this.makeNativeFunction("Future.is_default", 1, () => ({ kind: "bool", value: false })),
    };

    this.errorNativeMethods = {
      message: this.makeNativeFunction("Error.message", 1, (_interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "error") throw new Error("Error.message called on non-error");
        return { kind: "String", value: self.message ?? "" };
      }),
      cause: this.makeNativeFunction("Error.cause", 1, (_interp, args) => {
        const self = args[0];
        if (!self || self.kind !== "error") throw new Error("Error.cause called on non-error");
        if (self.cause) {
          return self.cause;
        }
        if (self.value && self.value.kind === "error") {
          return self.value;
        }
        return { kind: "nil", value: null };
      }),
    };

    const procYieldFn = this.makeNativeFunction("future_yield", 0, (interp) => interp.procYield());
    const procCancelledFn = this.makeNativeFunction("future_cancelled", 0, (interp) => interp.procCancelled());
    const procFlushFn = this.makeNativeFunction("future_flush", 0, (interp) => interp.procFlush());
    const procPendingTasksFn = this.makeNativeFunction("future_pending_tasks", 0, (interp) => interp.procPendingTasks());
    this.globals.define("future_yield", procYieldFn);
    this.globals.define("future_cancelled", procCancelledFn);
    this.globals.define("future_flush", procFlushFn);
    this.globals.define("future_pending_tasks", procPendingTasksFn);
    this.ensureArrayKernelBuiltins();
    this.ensureHashMapKernelBuiltins();
    this.installBuiltinInterfaces();
  }

  resetTimeSlice(): void {
    this.timeSliceCounter = 0;
  }

  checkTimeSlice(): void {
    if (this.asyncContextStack.length === 0) return;
    const asyncCtx = this.currentAsyncContext ? this.currentAsyncContext() : null;
    if (asyncCtx?.handle.entrypoint) {
      const pending = typeof this.executor.pendingTasks === "function" ? this.executor.pendingTasks() : 0;
      if (pending === 0) return;
    }
    this.timeSliceCounter += 1;
    if (this.timeSliceCounter >= this.schedulerMaxSteps) {
      this.timeSliceCounter = 0;
      throw new ProcYieldSignal();
    }
  }

  private installBuiltinInterfaces(): void {
    const { interfaces, implementations } = buildStandardInterfaceBuiltins();
    for (const iface of interfaces) {
      evaluateInterfaceDefinition(this, iface, this.globals);
    }
    for (const impl of implementations) {
      evaluateImplementationDefinition(this, impl, this.globals);
    }
  }
}

applyHelperAugmentations(Interpreter);
applyOperationsAugmentations(Interpreter);
applyStringifyAugmentations(Interpreter);
applyPatternAugmentations(Interpreter);
applyTypesAugmentations(Interpreter);
applyMemberAugmentations(Interpreter);
applyImplResolutionAugmentations(Interpreter);
applyRangeAugmentations(Interpreter);
applyPlaceholderAugmentations(Interpreter);
applyIteratorAugmentations(Interpreter);
applyArrayKernelAugmentations(Interpreter);
applyHashMapKernelAugmentations(Interpreter);
applyChannelMutexAugmentations(Interpreter);
applyStringHostAugmentations(Interpreter);
applyNumericHostAugmentations(Interpreter);
applyOsHostAugmentations(Interpreter);
applyExternHostAugmentations(Interpreter);
applyDynamicAugmentations(Interpreter);
applyEvaluationAugmentations(Interpreter);
applyConcurrencyAugmentations(Interpreter);

export type { ConstraintSpec as InterpreterConstraintSpec } from "./values";

export { Environment } from "./environment";
export type { RuntimeValue } from "./values";
export type { Executor } from "./executor";
export { CooperativeExecutor } from "./executor";

export type { PlaceholderFrame } from "./placeholders";
// Side-effectful module imports attach feature-specific behaviour to Interpreter.

export function evaluate(node: AST.AstNode | null, env?: Environment): RuntimeValue {
  return new Interpreter().evaluate(node, env);
}
