import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { ProcContinuationContext } from "./proc_continuations";
import type { RuntimeDiagnosticContext } from "./runtime_diagnostics";

export type IntegerKind = "i8" | "i16" | "i32" | "i64" | "i128" | "u8" | "u16" | "u32" | "u64" | "u128";
export type FloatKind = "f32" | "f64";

export type ImplMethodEntry = {
  def: AST.ImplementationDefinition;
  methods: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>;
  targetArgTemplates: AST.TypeExpression[];
  genericParams: AST.GenericParameter[];
  whereClause?: AST.WhereClauseConstraint[];
  unionVariantSignatures?: string[];
};

export type RuntimeValue =
  | { kind: "String"; value: string }
  | { kind: "bool"; value: boolean }
  | { kind: "char"; value: string }
  | { kind: "nil"; value: null }
  | { kind: "void" }
  | { kind: IntegerKind; value: bigint }
  | { kind: FloatKind; value: number }
  | { kind: "array"; elements: RuntimeValue[]; handle?: number }
  | {
      kind: "hash_map";
      entries: Map<string, { key: RuntimeValue; value: RuntimeValue }>;
      order: string[];
    }
  | IteratorValue
  | IteratorEndValue
  | { kind: "function"; node: AST.FunctionDefinition | AST.LambdaExpression; closureEnv: Environment }
  | { kind: "struct_def"; def: AST.StructDefinition }
  | { kind: "type_ref"; typeName: string; typeArgs?: AST.TypeExpression[] }
  | {
      kind: "struct_instance";
      def: AST.StructDefinition;
      values: RuntimeValue[] | Map<string, RuntimeValue>;
      typeArguments?: AST.TypeExpression[];
      typeArgMap?: Map<string, AST.TypeExpression>;
    }
  | { kind: "interface_def"; def: AST.InterfaceDefinition }
  | { kind: "union_def"; def: AST.UnionDefinition }
  | { kind: "package"; name: string; symbols: Map<string, RuntimeValue> }
  | {
      kind: "impl_namespace";
      def: AST.ImplementationDefinition;
      symbols: Map<string, RuntimeValue>;
      meta: {
        interfaceName: string;
        target: AST.TypeExpression;
        interfaceArgs?: AST.TypeExpression[];
      };
    }
  | {
      kind: "function_overload";
      overloads: Array<Extract<RuntimeValue, { kind: "function" }>>;
    }
  | { kind: "dyn_package"; name: string }
  | { kind: "dyn_ref"; pkg: string; name: string }
  | { kind: "error"; message: string; value?: RuntimeValue; cause?: RuntimeValue }
  | { kind: "host_handle"; handleType: "IoHandle" | "ProcHandle"; value: unknown }
  | { kind: "bound_method"; func: Extract<RuntimeValue, { kind: "function" | "function_overload" }>; self: RuntimeValue }
  | {
      kind: "interface_value";
      interfaceName: string;
      value: RuntimeValue;
      interfaceArgs?: AST.TypeExpression[];
      typeArguments?: AST.TypeExpression[];
      typeArgMap?: Map<string, AST.TypeExpression>;
      methods?: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" | "native_function" }>>;
    }
  | {
      kind: "proc_handle";
      state: "pending" | "resolved" | "failed" | "cancelled";
      expression: AST.AstNode;
      env: Environment;
      runner: (() => void) | null;
      result?: RuntimeValue;
      error?: RuntimeValue;
      failureInfo?: RuntimeValue;
      errorContext?: RuntimeDiagnosticContext;
      isEvaluating?: boolean;
      errorMode?: "raw" | "proc";
      cancelRequested?: boolean;
      hasStarted?: boolean;
      entrypoint?: boolean;
      waitingMutex?: unknown;
      continuation?: ProcContinuationContext;
      awaitBlocked?: boolean;
    }
  | {
      kind: "future";
      state: "pending" | "resolved" | "failed" | "cancelled";
      expression: AST.AstNode;
      env: Environment;
      runner: (() => void) | null;
      result?: RuntimeValue;
      error?: RuntimeValue;
      failureInfo?: RuntimeValue;
      isEvaluating?: boolean;
      continuation?: ProcContinuationContext;
      cancelRequested?: boolean;
      hasStarted?: boolean;
      awaitBlocked?: boolean;
      waitingChannelSend?: {
        state: any;
        value: RuntimeValue;
      };
      waitingChannelReceive?: {
        state: any;
      };
      waitingMutex?: unknown;
    }
  | {
      kind: "native_function";
      name: string;
      arity: number;
      impl: (interpreter: Interpreter, args: RuntimeValue[]) => RuntimeValue | Promise<RuntimeValue>;
    }
  | {
      kind: "native_bound_method";
      func: Extract<RuntimeValue, { kind: "native_function" }>;
      self: RuntimeValue;
    }
  | {
      kind: "partial_function";
      target: RuntimeValue;
      boundArgs: RuntimeValue[];
      callNode?: AST.FunctionCall;
    };

export type HashMapValue = Extract<RuntimeValue, { kind: "hash_map" }>;
export type HashMapEntry = { key: RuntimeValue; value: RuntimeValue };

export type ConstraintSpec = {
  subjectExpr: AST.TypeExpression;
  ifaceType: AST.TypeExpression;
};

export type IteratorStep = {
  value: RuntimeValue;
  done: boolean;
};

export interface IteratorValue {
  kind: "iterator";
  iterator: {
    next: () => IteratorStep;
    close: () => void;
  };
}

export interface IteratorEndValue {
  kind: "iterator_end";
}
