import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";

export type ImplMethodEntry = {
  def: AST.ImplementationDefinition;
  methods: Map<string, Extract<V10Value, { kind: "function" }>>;
  targetArgTemplates: AST.TypeExpression[];
  genericParams: AST.GenericParameter[];
  whereClause?: AST.WhereClauseConstraint[];
  unionVariantSignatures?: string[];
};

export type V10Value =
  | { kind: "string"; value: string }
  | { kind: "bool"; value: boolean }
  | { kind: "char"; value: string }
  | { kind: "nil"; value: null }
  | { kind: "i32"; value: number }
  | { kind: "f64"; value: number }
  | { kind: "array"; elements: V10Value[] }
  | { kind: "range"; start: number; end: number; inclusive: boolean }
  | { kind: "function"; node: AST.FunctionDefinition | AST.LambdaExpression; closureEnv: Environment }
  | { kind: "struct_def"; def: AST.StructDefinition }
  | {
      kind: "struct_instance";
      def: AST.StructDefinition;
      values: V10Value[] | Map<string, V10Value>;
      typeArguments?: AST.TypeExpression[];
      typeArgMap?: Map<string, AST.TypeExpression>;
    }
  | { kind: "interface_def"; def: AST.InterfaceDefinition }
  | { kind: "union_def"; def: AST.UnionDefinition }
  | { kind: "package"; name: string; symbols: Map<string, V10Value> }
  | {
      kind: "impl_namespace";
      def: AST.ImplementationDefinition;
      symbols: Map<string, V10Value>;
      meta: {
        interfaceName: string;
        target: AST.TypeExpression;
        interfaceArgs?: AST.TypeExpression[];
      };
    }
  | { kind: "dyn_package"; name: string }
  | { kind: "dyn_ref"; pkg: string; name: string }
  | { kind: "error"; message: string; value?: V10Value }
  | { kind: "bound_method"; func: Extract<V10Value, { kind: "function" }>; self: V10Value }
  | {
      kind: "interface_value";
      interfaceName: string;
      value: V10Value;
      typeArguments?: AST.TypeExpression[];
      typeArgMap?: Map<string, AST.TypeExpression>;
    }
  | {
      kind: "proc_handle";
      state: "pending" | "resolved" | "failed" | "cancelled";
      expression: AST.FunctionCall | AST.BlockExpression;
      env: Environment;
      runner: (() => void) | null;
      result?: V10Value;
      error?: V10Value;
      failureInfo?: V10Value;
      isEvaluating?: boolean;
      cancelRequested?: boolean;
      hasStarted?: boolean;
    }
  | {
      kind: "future";
      state: "pending" | "resolved" | "failed";
      expression: AST.FunctionCall | AST.BlockExpression;
      env: Environment;
      runner: (() => void) | null;
      result?: V10Value;
      error?: V10Value;
      failureInfo?: V10Value;
      isEvaluating?: boolean;
    }
  | {
      kind: "native_function";
      name: string;
      arity: number;
      impl: (interpreter: InterpreterV10, args: V10Value[]) => V10Value;
    }
  | {
      kind: "native_bound_method";
      func: Extract<V10Value, { kind: "native_function" }>;
      self: V10Value;
    };

export type ConstraintSpec = {
  typeParam: string;
  ifaceType: AST.TypeExpression;
};
