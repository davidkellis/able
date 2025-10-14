import * as AST from "../ast";
import { applyHelperAugmentations } from "./helpers";
import { applyOperationsAugmentations } from "./operations";
import { applyStringifyAugmentations } from "./stringify";
import { applyPatternAugmentations } from "./patterns";
import { applyTypesAugmentations } from "./types";
import { applyMemberAugmentations } from "./members";
import { applyImplResolutionAugmentations } from "./impl_resolution";
import { applyEvaluationAugmentations } from "./eval_expressions";
import { applyConcurrencyAugmentations } from "./concurrency";
import "./definitions";
import "./imports";

import { Environment } from "./environment";
import type { ImplMethodEntry, V10Value, ConstraintSpec } from "./values";

// =============================================================================
// v10 Interpreter (modular layout)
// =============================================================================

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

  schedulerQueue: Array<() => void> = [];
  schedulerScheduled = false;
  schedulerActive = false;
  schedulerMaxSteps = 1024;
  asyncContextStack: Array<
    { kind: "proc"; handle: Extract<V10Value, { kind: "proc_handle" }> } |
    { kind: "future"; handle: Extract<V10Value, { kind: "future" }> }
  > = [];

  constructor() {
    this.initConcurrencyBuiltins();
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
}

applyHelperAugmentations(InterpreterV10);
applyOperationsAugmentations(InterpreterV10);
applyStringifyAugmentations(InterpreterV10);
applyPatternAugmentations(InterpreterV10);
applyTypesAugmentations(InterpreterV10);
applyMemberAugmentations(InterpreterV10);
applyImplResolutionAugmentations(InterpreterV10);
applyEvaluationAugmentations(InterpreterV10);
applyConcurrencyAugmentations(InterpreterV10);

export type { ConstraintSpec as InterpreterConstraintSpec } from "./values";

export { Environment } from "./environment";
export type { V10Value } from "./values";

// Side-effectful module imports attach feature-specific behaviour to InterpreterV10.

export function evaluate(node: AST.AstNode | null, env?: Environment): V10Value {
  return new InterpreterV10().evaluate(node, env);
}
