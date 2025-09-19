import * as AST from "./ast";

// =============================================================================
// v10 Interpreter (initial scaffold)
// =============================================================================

type ConstraintSpec = { typeParam: string; ifaceType: AST.TypeExpression };
type ImplMethodEntry = {
  def: AST.ImplementationDefinition;
  methods: Map<string, Extract<V10Value, { kind: "function" }>>;
  targetArgTemplates: AST.TypeExpression[];
  genericParams: AST.GenericParameter[];
  whereClause?: AST.WhereClauseConstraint[];
  unionVariantSignatures?: string[];
};

// Runtime value union (start with primitives & array)
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
  | { kind: "struct_instance"; def: AST.StructDefinition; values: V10Value[] | Map<string, V10Value>; typeArguments?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> }
  | { kind: "interface_def"; def: AST.InterfaceDefinition }
  | { kind: "union_def"; def: AST.UnionDefinition }
  | { kind: "package"; name: string; symbols: Map<string, V10Value> }
  | { kind: "impl_namespace"; def: AST.ImplementationDefinition; symbols: Map<string, V10Value>; meta: { interfaceName: string; target: AST.TypeExpression; interfaceArgs?: AST.TypeExpression[] } }
  | { kind: "dyn_package"; name: string }
  | { kind: "dyn_ref"; pkg: string; name: string }
  | { kind: "error"; message: string; value?: V10Value }
  | { kind: "bound_method"; func: Extract<V10Value, { kind: "function" }>; self: V10Value }
  | { kind: "interface_value"; interfaceName: string; value: V10Value; typeArguments?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> }
  | { kind: "proc_handle"; state: "pending" | "resolved" | "failed" | "cancelled"; expression: AST.FunctionCall | AST.BlockExpression; env: Environment; runner: (() => void) | null; result?: V10Value; error?: V10Value; failureInfo?: V10Value; isEvaluating?: boolean; cancelRequested?: boolean; hasStarted?: boolean }
  | { kind: "future"; state: "pending" | "resolved" | "failed"; expression: AST.FunctionCall | AST.BlockExpression; env: Environment; runner: (() => void) | null; result?: V10Value; error?: V10Value; failureInfo?: V10Value; isEvaluating?: boolean }
  | { kind: "native_function"; name: string; arity: number; impl: (interpreter: InterpreterV10, args: V10Value[]) => V10Value }
  | { kind: "native_bound_method"; func: Extract<V10Value, { kind: "native_function" }>; self: V10Value };

class ReturnSignal extends Error {
  constructor(public value: V10Value) { super("ReturnSignal"); }
}
class RaiseSignal extends Error {
  constructor(public value: V10Value) { super("RaiseSignal"); }
}
class BreakSignalShim extends Error { constructor(){ super("BreakSignal"); } }
class BreakLabelSignal extends Error { constructor(public label: string, public value: V10Value){ super("BreakLabelSignal"); } }
class ProcYieldSignal extends Error { constructor(){ super("ProcYieldSignal"); } }

export class Environment {
  private values: Map<string, V10Value> = new Map();
  constructor(private enclosing: Environment | null = null) {}
  define(name: string, value: V10Value): void {
    if (this.values.has(name)) throw new Error(`Redefinition in current scope: ${name}`);
    this.values.set(name, value);
  }
  assign(name: string, value: V10Value): void {
    if (this.values.has(name)) { this.values.set(name, value); return; }
    if (this.enclosing) { this.enclosing.assign(name, value); return; }
    throw new Error(`Undefined variable '${name}'`);
  }
  get(name: string): V10Value {
    if (this.values.has(name)) return this.values.get(name)!;
    if (this.enclosing) return this.enclosing.get(name);
    throw new Error(`Undefined variable '${name}'`);
  }
}

export class InterpreterV10 {
  readonly globals = new Environment();
  private interfaces: Map<string, AST.InterfaceDefinition> = new Map();
  private interfaceEnvs: Map<string, Environment> = new Map();
  private inherentMethods: Map<string, Map<string, Extract<V10Value, { kind: "function" }>>> = new Map();
  private implMethods: Map<string, ImplMethodEntry[]> = new Map();
  private unnamedImplsSeen: Map<string, Map<string, Set<string>>> = new Map();
  private raiseStack: V10Value[] = [];
  private packageRegistry: Map<string, Map<string, V10Value>> = new Map();
  private currentPackage: string | null = null;
  private breakpointStack: string[] = [];
  private procNativeMethods: {
    status: Extract<V10Value, { kind: "native_function" }>;
    value: Extract<V10Value, { kind: "native_function" }>;
    cancel: Extract<V10Value, { kind: "native_function" }>;
  };
  private futureNativeMethods: {
    status: Extract<V10Value, { kind: "native_function" }>;
    value: Extract<V10Value, { kind: "native_function" }>;
  };
  private concurrencyBuiltinsInitialized = false;
  private procErrorStruct!: AST.StructDefinition;
  private procStatusStructs!: {
    Pending: AST.StructDefinition;
    Resolved: AST.StructDefinition;
    Cancelled: AST.StructDefinition;
    Failed: AST.StructDefinition;
  };
  private procStatusPendingValue!: V10Value;
  private procStatusResolvedValue!: V10Value;
  private procStatusCancelledValue!: V10Value;
  private schedulerQueue: Array<() => void> = [];
  private schedulerScheduled = false;
  private schedulerActive = false;
  private schedulerMaxSteps = 1024;
  private asyncContextStack: Array<
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
    this.globals.define("proc_yield", procYieldFn);
    this.globals.define("proc_cancelled", procCancelledFn);
  }

  private registerSymbol(name: string, value: V10Value): void {
    if (!this.currentPackage) return;
    if (!this.packageRegistry.has(this.currentPackage)) this.packageRegistry.set(this.currentPackage, new Map());
    this.packageRegistry.get(this.currentPackage)!.set(name, value);
  }

  private qualifiedName(name: string): string | null {
    return this.currentPackage ? `${this.currentPackage}.${name}` : null;
  }

  evaluate(node: AST.AstNode | null, env: Environment = this.globals): V10Value {
    if (!node) return { kind: "nil", value: null };
    switch (node.type) {
      // --- Literals ---
      case "StringLiteral": return { kind: "string", value: (node as AST.StringLiteral).value };
      case "BooleanLiteral": return { kind: "bool", value: (node as AST.BooleanLiteral).value };
      case "CharLiteral": return { kind: "char", value: (node as AST.CharLiteral).value };
      case "NilLiteral": return { kind: "nil", value: null };
      case "FloatLiteral": {
        const n = (node as AST.FloatLiteral).value;
        return { kind: "f64", value: n };
      }
      case "IntegerLiteral": {
        const intNode = node as AST.IntegerLiteral;
        const kind = intNode.integerType ?? "i32";
        // For now, treat all number-like integer types as JS number when no bigint required
        if (kind === "i64" || kind === "i128" || kind === "u64" || kind === "u128") {
          // Keep it simple initially: coerce via Number (lossy) until bigint support is added
          return { kind: "i32", value: Number(intNode.value) };
        }
        return { kind: "i32", value: Number(intNode.value) };
      }
      case "ArrayLiteral": {
        const arr = (node as AST.ArrayLiteral).elements.map(e => this.evaluate(e, env));
        return { kind: "array", elements: arr };
      }

      // --- Unary ---
      case "UnaryExpression": {
        const u = node as AST.UnaryExpression;
        const v = this.evaluate(u.operand, env);
        if (u.operator === "-") {
          if (v.kind === "i32") return { kind: "i32", value: -v.value };
          if (v.kind === "f64") return { kind: "f64", value: -v.value };
          throw new Error("Unary '-' requires numeric operand");
        }
        if (u.operator === "!") {
          if (v.kind === "bool") return { kind: "bool", value: !v.value };
          throw new Error("Unary '!' requires boolean operand");
        }
        if (u.operator === "~") {
          if (v.kind === "i32") return { kind: "i32", value: ~v.value };
          throw new Error("Unary '~' requires i32 operand");
        }
        throw new Error(`Unknown unary operator ${u.operator}`);
      }

      // --- Binary ---
      case "BinaryExpression": {
        const b = node as AST.BinaryExpression;
        // Logical short-circuit
        if (b.operator === "&&" || b.operator === "||") {
          const lv = this.evaluate(b.left, env);
          if (lv.kind !== "bool") throw new Error("Logical operands must be bool");
          if (b.operator === "&&") {
            if (!lv.value) return { kind: "bool", value: false };
            const rv = this.evaluate(b.right, env);
            if (rv.kind !== "bool") throw new Error("Logical operands must be bool");
            return { kind: "bool", value: lv.value && rv.value };
          } else {
            if (lv.value) return { kind: "bool", value: true };
            const rv = this.evaluate(b.right, env);
            if (rv.kind !== "bool") throw new Error("Logical operands must be bool");
            return { kind: "bool", value: lv.value || rv.value };
          }
        }

        const left = this.evaluate(b.left, env);
        const right = this.evaluate(b.right, env);

        // String concatenation
        if (b.operator === "+" && left.kind === "string" && right.kind === "string") {
          return { kind: "string", value: left.value + right.value };
        }

        // Numeric helpers
        const isNum = (v: V10Value) => v.kind === "i32" || v.kind === "f64";
        const asNumber = (v: V10Value): number => v.kind === "i32" || v.kind === "f64" ? v.value : NaN;
        const resultKind = (a: V10Value, c: V10Value): "i32" | "f64" => (a.kind === "f64" || c.kind === "f64") ? "f64" : "i32";

        // Arithmetic
        if (["+","-","*","/","%"].includes(b.operator)) {
          if (!isNum(left) || !isNum(right)) throw new Error("Arithmetic requires numeric operands");
          const kind = resultKind(left, right);
          const l = asNumber(left);
          const r = asNumber(right);
          switch (b.operator) {
            case "+": return kind === "i32" ? { kind, value: (l + r) | 0 } : { kind, value: l + r };
            case "-": return kind === "i32" ? { kind, value: (l - r) | 0 } : { kind, value: l - r };
            case "*": return kind === "i32" ? { kind, value: (l * r) | 0 } : { kind, value: l * r };
            case "/": {
              if (r === 0) throw new Error("Division by zero");
              return kind === "i32" ? { kind, value: (l / r) | 0 } : { kind, value: l / r };
            }
            case "%": {
              if (r === 0) throw new Error("Division by zero");
              return { kind, value: kind === "i32" ? (l % r) | 0 : l % r };
            }
          }
        }

        // Comparisons
        if ([">","<",">=","<=","==","!="].includes(b.operator)) {
          if (isNum(left) && isNum(right)) {
            const l = asNumber(left); const r = asNumber(right);
            switch (b.operator) {
              case ">": return { kind: "bool", value: l > r };
              case "<": return { kind: "bool", value: l < r };
              case ">=": return { kind: "bool", value: l >= r };
              case "<=": return { kind: "bool", value: l <= r };
              case "==": return { kind: "bool", value: l === r };
              case "!=": return { kind: "bool", value: l !== r };
            }
          }
          if (left.kind === "string" && right.kind === "string") {
            switch (b.operator) {
              case ">": return { kind: "bool", value: left.value > right.value };
              case "<": return { kind: "bool", value: left.value < right.value };
              case ">=": return { kind: "bool", value: left.value >= right.value };
              case "<=": return { kind: "bool", value: left.value <= right.value };
              case "==": return { kind: "bool", value: left.value === right.value };
              case "!=": return { kind: "bool", value: left.value !== right.value };
            }
          }
          // Fallback: only equal if same kind and deep-equal of value for simple cases
          if (b.operator === "==") return { kind: "bool", value: JSON.stringify(left) === JSON.stringify(right) };
          if (b.operator === "!=") return { kind: "bool", value: JSON.stringify(left) !== JSON.stringify(right) };
          throw new Error("Unsupported comparison operands");
        }

        // Bitwise on i32
        if (["&","|","^","<<",">>"] .includes(b.operator)) {
          if (left.kind !== "i32" || right.kind !== "i32") throw new Error("Bitwise requires i32 operands");
          switch (b.operator) {
            case "&": return { kind: "i32", value: left.value & right.value };
            case "|": return { kind: "i32", value: left.value | right.value };
            case "^": return { kind: "i32", value: left.value ^ right.value };
            case "<<": {
              const count = right.value;
              if (count < 0 || count >= 32) throw new Error("shift out of range");
              return { kind: "i32", value: left.value << count };
            }
            case ">>": {
              const count = right.value;
              if (count < 0 || count >= 32) throw new Error("shift out of range");
              return { kind: "i32", value: left.value >> count };
            }
          }
        }

        throw new Error(`Unknown binary operator ${b.operator}`);
      }

      // --- Range ---
      case "RangeExpression": {
        const r = node as AST.RangeExpression;
        const s = this.evaluate(r.start, env);
        const e = this.evaluate(r.end, env);
        const sNum = (s.kind === "i32" || s.kind === "f64") ? s.value : NaN;
        const eNum = (e.kind === "i32" || e.kind === "f64") ? e.value : NaN;
        if (Number.isNaN(sNum) || Number.isNaN(eNum)) throw new Error("Range boundaries must be numeric");
        return { kind: "range", start: sNum, end: eNum, inclusive: r.inclusive };
      }

      case "IndexExpression": {
        const ix = node as AST.IndexExpression;
        const obj = this.evaluate(ix.object, env);
        const idxVal = this.evaluate(ix.index, env);
        if (obj.kind !== "array") throw new Error("Indexing is only supported on arrays in this milestone");
        const idx = (idxVal.kind === "i32" || idxVal.kind === "f64") ? Math.trunc(idxVal.value) : NaN;
        if (!Number.isFinite(idx)) throw new Error("Array index must be a number");
        if (idx < 0 || idx >= obj.elements.length) throw new Error("Array index out of bounds");
        const el = obj.elements[idx];
        if (el === undefined) throw new Error("Internal error: array element undefined");
        return el;
      }

      // --- If / Or ---
      case "IfExpression": {
        const ife = node as AST.IfExpression;
        const cond = this.evaluate(ife.ifCondition, env);
        if (this.isTruthy(cond)) return this.evaluate(ife.ifBody, env);
        for (const orc of ife.orClauses) {
          if (orc.condition) {
            const c = this.evaluate(orc.condition, env);
            if (this.isTruthy(c)) return this.evaluate(orc.body, env);
          } else {
            return this.evaluate(orc.body, env);
          }
        }
        return { kind: "nil", value: null };
      }

      // --- While ---
      case "WhileLoop": {
        const wl = node as AST.WhileLoop;
        while (true) {
          const c = this.evaluate(wl.condition, env);
          if (!this.isTruthy(c)) break;
          const bodyEnv = new Environment(env);
          try {
            this.evaluate(wl.body, bodyEnv);
          } catch (e) {
            if ((e as any) instanceof BreakSignalShim) break;
            if (e instanceof BreakLabelSignal) throw e;
            throw e;
          }
        }
        return { kind: "nil", value: null };
      }

      case "BreakStatement": {
        const br = node as AST.BreakStatement;
        const labelName = br.label.name;
        // If targeting an active breakpoint label, throw labeled break with value
        if (this.breakpointStack.includes(labelName)) {
          const val = this.evaluate(br.value, env);
          throw new BreakLabelSignal(labelName, val);
        }
        // Otherwise treat as loop break
        throw new BreakSignalShim();
      }

      // --- For --- (arrays & ranges)
      case "ForLoop": {
        const fl = node as AST.ForLoop;
        const iterable = this.evaluate(fl.iterable, env);
        const bodyEnvBase = new Environment(env);
        const bindPattern = (value: V10Value, targetEnv: Environment) => {
          if (fl.pattern.type === "Identifier") {
            targetEnv.define(fl.pattern.name, value);
            return;
          }
          if (fl.pattern.type === "WildcardPattern") {
            return;
          }
          // Use full destructuring semantics for struct/array/typed/literal patterns
          this.assignByPattern(fl.pattern as AST.Pattern, value, targetEnv, true);
        };

        if (iterable.kind === "array") {
          for (const el of iterable.elements) {
            const iterEnv = new Environment(bodyEnvBase);
            bindPattern(el, iterEnv);
            this.evaluate(fl.body, iterEnv);
          }
          return { kind: "nil", value: null };
        }
        if (iterable.kind === "range") {
          const step = iterable.start <= iterable.end ? 1 : -1;
          if (step > 0) {
            for (let i = iterable.start; i < (iterable.inclusive ? iterable.end + 1 : iterable.end); i += 1) {
              const iterEnv = new Environment(bodyEnvBase);
              bindPattern({ kind: "i32", value: i }, iterEnv);
              this.evaluate(fl.body, iterEnv);
            }
          } else {
            for (let i = iterable.start; i > (iterable.inclusive ? iterable.end - 1 : iterable.end); i -= 1) {
              const iterEnv = new Environment(bodyEnvBase);
              bindPattern({ kind: "i32", value: i }, iterEnv);
              this.evaluate(fl.body, iterEnv);
            }
          }
          return { kind: "nil", value: null };
        }
        throw new Error("ForLoop iterable must be array or range");
      }

      // --- Match ---
      case "MatchExpression": {
        const me = node as AST.MatchExpression;
        const value = this.evaluate(me.subject, env);
        for (const clause of me.clauses) {
          const matchEnv = this.tryMatchPattern(clause.pattern, value, env);
          if (matchEnv) {
            if (clause.guard) {
              const g = this.evaluate(clause.guard, matchEnv);
              if (!this.isTruthy(g)) continue;
            }
            return this.evaluate(clause.body, matchEnv);
          }
        }
        throw new Error("Non-exhaustive match");
      }

      // --- Expressions scaffold we will fill later ---
      case "Identifier": return env.get((node as AST.Identifier).name);
      case "BlockExpression": {
        const block = node as AST.BlockExpression;
        const blockEnv = new Environment(env);
        let last: V10Value = { kind: "nil", value: null };
        for (const stmt of block.body) last = this.evaluate(stmt, blockEnv);
        return last;
      }
      case "ReturnStatement": {
        const r = node as AST.ReturnStatement;
        const val = r.argument ? this.evaluate(r.argument, env) : ({ kind: "nil", value: null } as V10Value);
        throw new ReturnSignal(val);
      }
      case "AssignmentExpression": {
        const a = node as AST.AssignmentExpression;
        const val = this.evaluate(a.right, env);
        const isCompound = ["+=","-=","*=","/=","%=","&=","|=","^=","<<=",">>="].includes(a.operator);
        // Identifier assignment / declaration
        if (a.left.type === "Identifier") {
          if (a.operator === ":=") { env.define(a.left.name, val); return val; }
          if (isCompound) {
            const current = env.get(a.left.name);
            const op = a.operator.slice(0, -1);
            const computed = this.computeBinaryForCompound(op, current, val);
            env.assign(a.left.name, computed);
            return computed;
          }
          env.assign(a.left.name, val);
          return val;
        }
        // Destructuring assignment
        if (a.left.type === "StructPattern" || a.left.type === "ArrayPattern" || a.left.type === "WildcardPattern" || a.left.type === "LiteralPattern" || (a.left as any).type === "TypedPattern") {
          if (isCompound) throw new Error("Compound assignment not supported with destructuring");
          const isDecl = a.operator === ":=";
          this.assignByPattern(a.left as AST.Pattern, val, env, isDecl);
          return val;
        }
        // Member assignment on struct instances
        if (a.left.type === "MemberAccessExpression") {
          if (a.operator === ":=") throw new Error("Cannot use := on member access");
          const targetObj = this.evaluate(a.left.object, env);
          if (targetObj.kind === "struct_instance") {
            if (a.left.member.type === "Identifier") {
              if (!(targetObj.values instanceof Map)) throw new Error("Expected named struct instance");
              if (!targetObj.values.has(a.left.member.name)) throw new Error(`No field named '${a.left.member.name}'`);
              if (isCompound) {
                const current = targetObj.values.get(a.left.member.name)!;
                const op = a.operator.slice(0, -1);
                const computed = this.computeBinaryForCompound(op, current, val);
                targetObj.values.set(a.left.member.name, computed);
                return computed;
              }
              targetObj.values.set(a.left.member.name, val);
              return val;
            } else {
              if (!Array.isArray(targetObj.values)) throw new Error("Expected positional struct instance");
              const idx = Number(a.left.member.value);
              if (idx < 0 || idx >= targetObj.values.length) throw new Error("Struct field index out of bounds");
              if (isCompound) {
                const current = targetObj.values[idx] as V10Value;
                const op = a.operator.slice(0, -1);
                const computed = this.computeBinaryForCompound(op, current, val);
                targetObj.values[idx] = computed;
                return computed;
              }
              targetObj.values[idx] = val;
              return val;
            }
          } else if (targetObj.kind === "array") {
            if (a.left.member.type !== "IntegerLiteral") throw new Error("Array member assignment requires integer member");
            const idx = Number(a.left.member.value);
            if (idx < 0 || idx >= targetObj.elements.length) throw new Error("Array index out of bounds");
            if (isCompound) {
              const currEl = targetObj.elements[idx]!;
              const op = a.operator.slice(0, -1);
              const computed = this.computeBinaryForCompound(op, currEl, val);
              targetObj.elements[idx] = computed;
              return computed;
            }
            targetObj.elements[idx] = val;
            return val;
          } else {
            throw new Error("Member assignment requires struct or array");
          }
        }
        // Index assignment on arrays
        if (a.left.type === "IndexExpression") {
          if (a.operator === ":=") throw new Error("Cannot use := on index assignment");
          const obj = this.evaluate(a.left.object, env);
          const idxVal = this.evaluate(a.left.index, env);
          if (obj.kind !== "array") throw new Error("Index assignment requires array");
          const idx = (idxVal.kind === "i32" || idxVal.kind === "f64") ? Math.trunc(idxVal.value) : NaN;
          if (!Number.isFinite(idx)) throw new Error("Array index must be a number");
          if (idx < 0 || idx >= obj.elements.length) throw new Error("Array index out of bounds");
          if (isCompound) {
            const currEl2 = obj.elements[idx]!;
            const op = a.operator.slice(0, -1);
            const computed = this.computeBinaryForCompound(op, currEl2, val);
            obj.elements[idx] = computed;
            return computed;
          }
          obj.elements[idx] = val;
          return val;
        }
        throw new Error("Unsupported assignment target");
      }

      // --- Functions & Lambdas ---
      case "FunctionDefinition": {
        const fn = node as AST.FunctionDefinition;
        const value: V10Value = { kind: "function", node: fn, closureEnv: env };
        env.define(fn.id.name, value);
        this.registerSymbol(fn.id.name, value);
        // Also bind package-qualified name in globals for import lookups
        const qn = this.qualifiedName(fn.id.name);
        if (qn) {
          try { this.globals.define(qn, value); } catch {}
        }
        return { kind: "nil", value: null };
      }
      case "LambdaExpression": {
        const lam = node as AST.LambdaExpression;
        return { kind: "function", node: lam, closureEnv: env };
      }
      case "FunctionCall": {
        const call = node as AST.FunctionCall;
        const calleeEvaluated = this.evaluate(call.callee, env);
        let funcValue: Extract<V10Value, { kind: "function" }> | null = null;
        let nativeFunc: Extract<V10Value, { kind: "native_function" }> | null = null;
        let injectedArgs: V10Value[] = [];
        if (calleeEvaluated.kind === "bound_method") {
          funcValue = calleeEvaluated.func;
          injectedArgs = [calleeEvaluated.self];
        } else if (calleeEvaluated.kind === "function") {
          funcValue = calleeEvaluated;
        } else if (calleeEvaluated.kind === "dyn_ref") {
          // Resolve the actual function/value now
          const bucket = this.packageRegistry.get(calleeEvaluated.pkg);
          const sym = bucket?.get(calleeEvaluated.name);
          if (!sym || sym.kind !== "function") throw new Error(`dyn ref '${calleeEvaluated.pkg}.${calleeEvaluated.name}' is not callable`);
          funcValue = sym;
        } else if (calleeEvaluated.kind === "native_bound_method") {
          nativeFunc = calleeEvaluated.func;
          injectedArgs = [calleeEvaluated.self];
        } else if (calleeEvaluated.kind === "native_function") {
          nativeFunc = calleeEvaluated;
        } else {
          throw new Error("Cannot call non-function");
        }
        const callArgs = call.arguments.map(a => this.evaluate(a, env));
        if (nativeFunc) {
          const evalArgs = [...injectedArgs, ...callArgs];
          if (evalArgs.length !== nativeFunc.arity) {
            throw new Error(`Arity mismatch calling ${nativeFunc.name}: expected ${nativeFunc.arity}, got ${evalArgs.length}`);
          }
          return nativeFunc.impl(this, evalArgs);
        }
        if (!funcValue) throw new Error("Callable target missing function value");
        const funcNode = funcValue.node;
        // Enforce minimal generic/interface constraints if present
        this.enforceGenericConstraintsIfAny(funcNode, call);
        const funcEnv = new Environment(funcValue.closureEnv);
        // Bind generic type arguments into function environment for introspection (as `${T}_type` strings)
        this.bindTypeArgumentsIfAny(funcNode, call, funcEnv);
        // Bind params (identifier-only for now)
        const params = funcNode.type === "FunctionDefinition" ? funcNode.params : funcNode.params;
        const evalArgs: V10Value[] = [...injectedArgs, ...callArgs];
        if (evalArgs.length !== params.length) {
          const name = (funcNode as any).id?.name ?? "(lambda)";
          throw new Error(`Arity mismatch calling ${name}: expected ${params.length}, got ${evalArgs.length}`);
        }
        for (let i = 0; i < params.length; i++) {
          const p = params[i];
          const argVal = evalArgs[i];
          if (!p) throw new Error(`Parameter missing at index ${i}`);
          if (argVal === undefined) throw new Error(`Argument missing at index ${i}`);
          // Enforce minimal runtime type checks when a parameter type annotation is present
          if (p.paramType) {
            if (!this.matchesType(p.paramType, argVal)) {
              const pname = (p.name as any).name ?? `param_${i}`;
              throw new Error(`Parameter type mismatch for '${pname}'`);
            }
          }
          const coercedArg = p.paramType ? this.coerceValueToType(p.paramType, argVal) : argVal;
          evalArgs[i] = coercedArg;
          if (p.name.type === "Identifier") {
            funcEnv.define(p.name.name, coercedArg);
          } else if (p.name.type === "WildcardPattern") {
            // ignore
          } else if (p.name.type === "StructPattern" || p.name.type === "ArrayPattern" || p.name.type === "LiteralPattern") {
            // destructuring param
            this.assignByPattern(p.name as any, coercedArg, funcEnv, true);
          } else {
            throw new Error("Only simple identifier and destructuring params supported for now");
          }
        }
        // Execute body
        try {
          if (funcNode.type === "FunctionDefinition") {
            // body is BlockExpression
            return this.evaluate(funcNode.body, funcEnv);
          } else {
            // Lambda: expression or block
            const b = funcNode.body;
            return this.evaluate(b as AST.AstNode, funcEnv);
          }
        } catch (e) {
          if (e instanceof ReturnSignal) return e.value;
          throw e;
        }
      }

      // --- Structs ---
      case "StructDefinition": {
        const sd = node as AST.StructDefinition;
        env.define(sd.id.name, { kind: "struct_def", def: sd });
        this.registerSymbol(sd.id.name, { kind: "struct_def", def: sd });
        const qn = this.qualifiedName(sd.id.name);
        if (qn) {
          try { this.globals.define(qn, { kind: "struct_def", def: sd }); } catch {}
        }
        return { kind: "nil", value: null };
      }
      case "StructLiteral": {
        const sl = node as AST.StructLiteral;
        if (!sl.structType) throw new Error("Struct literal requires explicit struct type in this milestone");
        const defVal = env.get(sl.structType.name);
        if (defVal.kind !== "struct_def") throw new Error(`'${sl.structType.name}' is not a struct type`);
        const structDef = defVal.def;
        const generics = structDef.genericParams;
        const constraints = this.collectConstraintSpecs(generics, structDef.whereClause);
        let typeArguments: AST.TypeExpression[] | undefined = sl.typeArguments;
        let typeArgMap: Map<string, AST.TypeExpression> | undefined;
        if (generics && generics.length > 0) {
          typeArgMap = this.mapTypeArguments(generics, typeArguments, `instantiating ${structDef.id.name}`);
          if (constraints.length > 0) {
            this.enforceConstraintSpecs(constraints, typeArgMap, `struct ${structDef.id.name}`);
          }
        } else if (sl.typeArguments && sl.typeArguments.length > 0) {
          throw new Error(`Type '${structDef.id.name}' does not accept type arguments`);
        }
        if (sl.isPositional) {
          // Positional: build array values in order
          const vals: V10Value[] = sl.fields.map(f => this.evaluate(f.value, env));
          return { kind: "struct_instance", def: structDef, values: vals, typeArguments, typeArgMap: typeArgMap ? new Map(typeArgMap) : undefined };
        } else {
          // Named: build map with optional functional update source
          const map = new Map<string, V10Value>();
          if (sl.functionalUpdateSource) {
            const baseVal = this.evaluate(sl.functionalUpdateSource, env);
            if (baseVal.kind !== "struct_instance") throw new Error("Functional update source must be a struct instance");
            if (baseVal.def.id.name !== structDef.id.name) throw new Error("Functional update source must be same struct type");
            if (!(baseVal.values instanceof Map)) throw new Error("Functional update only supported for named structs");
            if (typeArguments && baseVal.typeArguments) {
              if (typeArguments.length !== baseVal.typeArguments.length) {
                throw new Error("Functional update must use same type arguments as source");
              }
              for (let i = 0; i < typeArguments.length; i++) {
                const ta = typeArguments[i]!;
                const baseTa = baseVal.typeArguments[i]!;
                if (!this.typeExpressionsEqual(ta, baseTa)) {
                  throw new Error("Functional update must use same type arguments as source");
                }
              }
            }
            for (const [k, v] of (baseVal.values as Map<string, V10Value>).entries()) {
              map.set(k, v);
            }
          }
          for (const f of sl.fields) {
            let fname: string | undefined = f.name?.name;
            if (!fname) {
              if (f.isShorthand && f.value.type === "Identifier") fname = (f.value as AST.Identifier).name;
            }
            if (!fname) throw new Error("Named struct field initializer must have a field name");
            const val = this.evaluate(f.value, env);
            map.set(fname, val);
          }
          return { kind: "struct_instance", def: structDef, values: map, typeArguments, typeArgMap: typeArgMap ? new Map(typeArgMap) : undefined };
        }
      }
      case "MemberAccessExpression": {
        const ma = node as AST.MemberAccessExpression;
        const obj = this.evaluate(ma.object, env);
        // Package alias member access
        if (obj.kind === "package") {
          if (ma.member.type !== "Identifier") throw new Error("Package member access expects identifier");
          const sym = obj.symbols.get(ma.member.name);
          if (!sym) throw new Error(`No public member '${ma.member.name}' on package ${obj.name}`);
          return sym;
        }
        if (obj.kind === "dyn_package") {
          if (ma.member.type !== "Identifier") throw new Error("Dyn package member access expects identifier");
          // Resolve late from registry
          const bucket = this.packageRegistry.get(obj.name);
          const sym = bucket?.get(ma.member.name);
          if (!sym) throw new Error(`dyn package '${obj.name}' has no member '${ma.member.name}'`);
          if (sym.kind === "function" && sym.node.type === "FunctionDefinition" && sym.node.isPrivate) throw new Error(`dyn package '${obj.name}' member '${ma.member.name}' is private`);
          if (sym.kind === "struct_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${ma.member.name}' is private`);
          if (sym.kind === "interface_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${ma.member.name}' is private`);
          if (sym.kind === "union_def" && sym.def.isPrivate) throw new Error(`dyn package '${obj.name}' member '${ma.member.name}' is private`);
          return { kind: "dyn_ref", pkg: obj.name, name: ma.member.name };
        }
        if (obj.kind === "interface_value") {
          if (ma.member.type !== "Identifier") throw new Error("Interface member access expects identifier");
          const underlying = obj.value;
          const typeName = this.getTypeNameForValue(underlying);
          if (!typeName) throw new Error(`No method '${ma.member.name}' for interface ${obj.interfaceName}`);
          const typeArgs = underlying.kind === "struct_instance" ? underlying.typeArguments : undefined;
          const typeArgMap = underlying.kind === "struct_instance" ? underlying.typeArgMap : undefined;
          const method = this.findMethod(typeName, ma.member.name, {
            typeArgs,
            typeArgMap,
            interfaceName: obj.interfaceName,
          });
          if (!method) throw new Error(`No method '${ma.member.name}' for interface ${obj.interfaceName}`);
          if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
            throw new Error(`Method '${ma.member.name}' on ${typeName} is private`);
          }
          return { kind: "bound_method", func: method, self: underlying };
        }
        if (obj.kind === "proc_handle") {
          if (ma.member.type !== "Identifier") throw new Error("Proc handle member access expects identifier");
          const fn = (this.procNativeMethods as Record<string, Extract<V10Value, { kind: "native_function" }>>)[ma.member.name];
          if (!fn) throw new Error(`Unknown proc handle method '${ma.member.name}'`);
          return this.bindNativeMethod(fn, obj);
        }
        if (obj.kind === "future") {
          if (ma.member.type !== "Identifier") throw new Error("Future member access expects identifier");
          const fn = (this.futureNativeMethods as Record<string, Extract<V10Value, { kind: "native_function" }>>)[ma.member.name];
          if (!fn) throw new Error(`Unknown future method '${ma.member.name}'`);
          return this.bindNativeMethod(fn, obj);
        }
        if (obj.kind === "impl_namespace") {
          if (ma.member.type !== "Identifier") throw new Error("Impl namespace member access expects identifier");
          if (ma.member.name === "interface") {
            return { kind: "string", value: obj.meta.interfaceName };
          }
          if (ma.member.name === "target") {
            return { kind: "string", value: this.typeExpressionToString(obj.meta.target) };
          }
          if (ma.member.name === "interface_args") {
            const args = obj.meta.interfaceArgs ?? [];
            return {
              kind: "array",
              elements: args.map(a => ({ kind: "string", value: this.typeExpressionToString(a) } as V10Value)),
            };
          }
          const sym = obj.symbols.get(ma.member.name);
          if (!sym) throw new Error(`No method '${ma.member.name}' on impl ${obj.def.implName?.name ?? "<unnamed>"}`);
          return sym;
        }
        // Static method access: struct_def . identifier
        if (obj.kind === "struct_def") {
          if (ma.member.type !== "Identifier") throw new Error("Static access expects identifier member");
          const typeName = obj.def.id.name;
          const method = this.findMethod(typeName, ma.member.name);
          if (!method) throw new Error(`No static method '${ma.member.name}' for ${typeName}`);
          // Enforce static method privacy: if the underlying FunctionDefinition is private, deny access
          if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
            throw new Error(`Method '${ma.member.name}' on ${typeName} is private`);
          }
          // Static methods are plain functions (no self injected)
          return method;
        }
        if (obj.kind !== "struct_instance" && obj.kind !== "array") {
          if (ma.member.type === "Identifier") {
            const ufcs0 = this.tryUfcs(env, ma.member.name, obj);
            if (ufcs0) return ufcs0;
          }
          throw new Error("Member access only supported on structs/arrays in this milestone");
        }
        if (ma.member.type === "Identifier") {
          if (obj.kind !== "struct_instance") throw new Error("Named member access only valid on structs");
          if (!(obj.values instanceof Map)) throw new Error("Expected named struct instance");
          if (obj.values.has(ma.member.name)) {
            return obj.values.get(ma.member.name)!;
          }
          // Not a field; try method lookup
          const typeName = obj.def.id.name;
          const method = this.findMethod(typeName, ma.member.name, { typeArgs: obj.typeArguments, typeArgMap: obj.typeArgMap });
          if (method) {
            // Enforce instance method privacy before binding
            if (method.node.type === "FunctionDefinition" && method.node.isPrivate) {
              throw new Error(`Method '${ma.member.name}' on ${typeName} is private`);
            }
            return { kind: "bound_method", func: method, self: obj };
          }
          // UFCS fallback: free function with name taking receiver as first param
          const ufcs = this.tryUfcs(env, ma.member.name, obj);
          if (ufcs) return ufcs;
          throw new Error(`No field or method named '${ma.member.name}'`);
        } else {
          const idx = Number(ma.member.value);
          if (obj.kind === "struct_instance") {
            if (!Array.isArray(obj.values)) throw new Error("Expected positional struct instance");
            if (idx < 0 || idx >= obj.values.length) throw new Error("Struct field index out of bounds");
            const val = obj.values[idx];
            if (val === undefined) throw new Error("Internal error: positional field is undefined");
            return val;
          } else {
            // array member access by integer like a.0
            if (idx < 0 || idx >= obj.elements.length) throw new Error("Array index out of bounds");
            const el = obj.elements[idx];
            if (el === undefined) throw new Error("Internal error: array element undefined");
            return el;
          }
        }
      }

      // --- String Interpolation ---
      case "StringInterpolation": {
        const si = node as AST.StringInterpolation;
        let out = "";
        for (const part of si.parts) {
          if (part.type === "StringLiteral") out += part.value;
          else {
            const val = this.evaluate(part, env);
            out += this.valueToStringWithEnv(val, env);
          }
        }
        return { kind: "string", value: out };
      }
      case "BreakpointExpression": {
        const bp = node as AST.BreakpointExpression;
        this.breakpointStack.push(bp.label.name);
        try {
          return this.evaluate(bp.body, env);
        } catch (e) {
          if (e instanceof BreakLabelSignal) {
            if (e.label === bp.label.name) return e.value;
            throw e;
          }
          throw e;
        } finally {
          this.breakpointStack.pop();
        }
      }

      // --- Error Handling ---
      case "RaiseStatement": {
        const rs = node as AST.RaiseStatement;
        const val = this.evaluate(rs.expression, env);
        const err: V10Value = val.kind === "error" ? val : { kind: "error", message: this.valueToString(val), value: val };
        this.raiseStack.push(err);
        try {
          throw new RaiseSignal(err as V10Value);
        } finally {
          this.raiseStack.pop();
        }
      }
      case "RescueExpression": {
        const re = node as AST.RescueExpression;
        try {
          return this.evaluate(re.monitoredExpression, env);
        } catch (e) {
          if (e instanceof RaiseSignal) {
            for (const clause of re.clauses) {
              const matchEnv = this.tryMatchPattern(clause.pattern, e.value, env);
              if (matchEnv) {
                if (clause.guard) {
                  const g = this.evaluate(clause.guard, matchEnv);
                  if (!this.isTruthy(g)) continue;
                }
                return this.evaluate(clause.body, matchEnv);
              }
            }
            throw e;
          }
          throw e;
        }
      }
      case "OrElseExpression": {
        const oe = node as AST.OrElseExpression;
        try {
          return this.evaluate(oe.expression, env);
        } catch (e) {
          if (e instanceof RaiseSignal) {
            const hEnv = new Environment(env);
            if (oe.errorBinding) hEnv.define(oe.errorBinding.name, e.value);
            return this.evaluate(oe.handler, hEnv);
          }
          throw e;
        }
      }
      case "PropagationExpression": {
        const pe = node as AST.PropagationExpression;
        try {
          const val = this.evaluate(pe.expression, env);
          if (val.kind === "error") throw new RaiseSignal(val);
          return val;
        } catch (e) {
          if (e instanceof RaiseSignal) throw e;
          throw e;
        }
      }
      case "EnsureExpression": {
        const ee = node as AST.EnsureExpression;
        let result: V10Value | null = null;
        let caught: RaiseSignal | null = null;
        try {
          result = this.evaluate(ee.tryExpression, env);
        } catch (e) {
          if (e instanceof RaiseSignal) caught = e; else throw e;
        } finally {
          this.evaluate(ee.ensureBlock, env);
        }
        if (caught) throw caught;
        return result ?? { kind: "nil", value: null };
      }

      case "RethrowStatement": {
        const err = this.raiseStack[this.raiseStack.length - 1] || { kind: "error", message: "Unknown rethrow" } as V10Value;
        throw new RaiseSignal(err);
      }

      // --- Concurrency Handles ---
      case "ProcExpression": {
        const pr = node as AST.ProcExpression;
        const capturedEnv = new Environment(env);
        const handle: Extract<V10Value, { kind: "proc_handle" }> = {
          kind: "proc_handle",
          state: "pending",
          expression: pr.expression,
          env: capturedEnv,
          runner: null,
          cancelRequested: false,
        };
        handle.runner = () => this.runProcHandle(handle);
        this.scheduleAsync(handle.runner);
        return handle;
      }
      case "SpawnExpression": {
        const sp = node as AST.SpawnExpression;
        const capturedEnv = new Environment(env);
        const future: Extract<V10Value, { kind: "future" }> = {
          kind: "future",
          state: "pending",
          expression: sp.expression,
          env: capturedEnv,
          runner: null,
        };
        future.runner = () => this.runFuture(future);
        this.scheduleAsync(future.runner);
        return future;
      }

      // --- Module & Imports (minimal) ---
      case "Module": {
        const mod = node as AST.Module;
        // If this module declares a package, evaluate in a child env to avoid leaking
        // unqualified names into globals; otherwise, use globals directly.
        const moduleEnv = mod.package ? new Environment(this.globals) : this.globals;
        // Track current package for registry
        const prevPkg = this.currentPackage;
        if (mod.package) {
          this.currentPackage = mod.package.namePath.map(p => p.name).join(".");
          if (!this.packageRegistry.has(this.currentPackage)) this.packageRegistry.set(this.currentPackage, new Map());
        } else {
          this.currentPackage = null;
        }
        for (const imp of mod.imports) {
          this.evaluate(imp, moduleEnv);
        }
        let last: V10Value = { kind: "nil", value: null };
        for (const stmt of mod.body) {
          last = this.evaluate(stmt, moduleEnv);
        }
        this.currentPackage = prevPkg;
        return last;
      }
      case "PackageStatement": {
        return { kind: "nil", value: null };
      }
      case "ImportStatement": {
        const imp = node as AST.ImportStatement;
        // Minimal selector-based aliasing: pull from globals by original name and define alias in current env.
        // Package alias import: bind a package value exposing public symbols only
        if (!imp.isWildcard && (!imp.selectors || imp.selectors.length === 0) && imp.alias) {
          const pkg = imp.packagePath.map(p => p.name).join(".");
          const bucket = this.packageRegistry.get(pkg);
          if (!bucket) throw new Error(`Import error: package '${pkg}' not found`);
          const filtered = new Map<string, V10Value>();
          for (const [name, val] of bucket.entries()) {
            if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
            if (val.kind === "struct_def" && val.def.isPrivate) continue;
            if (val.kind === "interface_def" && val.def.isPrivate) continue;
            if (val.kind === "union_def" && val.def.isPrivate) continue;
            filtered.set(name, val);
          }
          env.define(imp.alias.name, { kind: "package", name: pkg, symbols: filtered });
        } else if (imp.isWildcard) {
          const pkg = imp.packagePath.map(p => p.name).join(".");
          const bucket = this.packageRegistry.get(pkg);
          if (!bucket) throw new Error(`Import error: package '${pkg}' not found`);
          for (const [name, val] of bucket.entries()) {
            if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
            if (val.kind === "struct_def" && val.def.isPrivate) continue;
            if (val.kind === "interface_def" && val.def.isPrivate) continue;
            if (val.kind === "union_def" && val.def.isPrivate) continue;
            // define into current env
            try { (env as Environment).define(name, val); } catch {}
          }
        } else if (imp.selectors && imp.selectors.length > 0) {
          const pkg = imp.packagePath.map(p => p.name).join(".");
          for (const sel of imp.selectors) {
            const original = sel.name.name;
            const alias = sel.alias ? sel.alias.name : original;
            let val: V10Value | null = null;
            // Prefer package-qualified lookup when a package path is given
            if (pkg) {
              try { val = this.globals.get(`${pkg}.${original}`); } catch {}
            }
            if (!val) {
              try { val = this.globals.get(original); } catch {}
            }
            if (!val && pkg) {
              try { val = this.globals.get(`${pkg}.${original}`); } catch {}
            }
            if (!val) throw new Error(`Import error: symbol '${original}'${pkg ? ` from '${pkg}'` : ''} not found in globals`);
            // Enforce privacy for functions and types tagged as private in their AST
            if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) {
              throw new Error(`Import error: function '${original}' is private`);
            }
            if (val.kind === "struct_def" && val.def.isPrivate) {
              throw new Error(`Import error: struct '${original}' is private`);
            }
            if (val.kind === "interface_def" && val.def.isPrivate) {
              throw new Error(`Import error: interface '${original}' is private`);
            }
            if (val.kind === "union_def" && val.def.isPrivate) {
              throw new Error(`Import error: union '${original}' is private`);
            }
            env.define(alias, val);
          }
        }
        return { kind: "nil", value: null };
      }
      case "DynImportStatement": {
        const dimp = node as AST.DynImportStatement;
        const pkg = dimp.packagePath.map(p => p.name).join(".");
        const bucket = this.packageRegistry.get(pkg);
        if (!bucket) throw new Error(`dynimport error: package '${pkg}' not found`);
        if (dimp.isWildcard) {
          for (const [name, val] of bucket.entries()) {
            // dynamic world would not enforce visibility; but spec says dynimport resolves late and can fail at use time.
            // However, to be safe, we still skip private here to mirror import behavior for now.
            if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) continue;
            if (val.kind === "struct_def" && val.def.isPrivate) continue;
            if (val.kind === "interface_def" && val.def.isPrivate) continue;
            if (val.kind === "union_def" && val.def.isPrivate) continue;
            try { env.define(name, { kind: "dyn_ref", pkg, name }); } catch {}
          }
        } else if (dimp.selectors && dimp.selectors.length > 0) {
          for (const sel of dimp.selectors) {
            const original = sel.name.name;
            const alias = sel.alias ? sel.alias.name : original;
            const val = bucket.get(original);
            if (!val) throw new Error(`dynimport error: '${original}' not found in '${pkg}'`);
            if (val.kind === "function" && val.node.type === "FunctionDefinition" && val.node.isPrivate) throw new Error(`dynimport error: function '${original}' is private`);
            if (val.kind === "struct_def" && val.def.isPrivate) throw new Error(`dynimport error: struct '${original}' is private`);
            if (val.kind === "interface_def" && val.def.isPrivate) throw new Error(`dynimport error: interface '${original}' is private`);
            if (val.kind === "union_def" && val.def.isPrivate) throw new Error(`dynimport error: union '${original}' is private`);
            env.define(alias, { kind: "dyn_ref", pkg, name: original });
          }
        } else if (dimp.alias) {
          // Bind a dyn_package alias (late-resolving container); member access will resolve dynamically
          env.define(dimp.alias.name, { kind: "dyn_package", name: pkg });
        }
        return { kind: "nil", value: null };
      }

      // --- Interfaces & Implementations & Methods ---
      case "InterfaceDefinition": {
        const idef = node as AST.InterfaceDefinition;
        this.interfaces.set(idef.id.name, idef);
        this.interfaceEnvs.set(idef.id.name, env);
        // Expose interface symbol in env for imports / visibility tests
        env.define(idef.id.name, { kind: "interface_def", def: idef });
        this.registerSymbol(idef.id.name, { kind: "interface_def", def: idef });
        const qn = this.qualifiedName(idef.id.name);
        if (qn) {
          try { this.globals.define(qn, { kind: "interface_def", def: idef }); } catch {}
        }
        return { kind: "nil", value: null };
      }
      case "UnionDefinition": {
        const udef = node as AST.UnionDefinition;
        // Expose union symbol in env for imports / visibility tests
        env.define(udef.id.name, { kind: "union_def", def: udef });
        this.registerSymbol(udef.id.name, { kind: "union_def", def: udef });
        const qn = this.qualifiedName(udef.id.name);
        if (qn) {
          try { this.globals.define(qn, { kind: "union_def", def: udef }); } catch {}
        }
        return { kind: "nil", value: null };
      }
      case "MethodsDefinition": {
        const md = node as AST.MethodsDefinition;
        // Only SimpleTypeExpression supported for now
        if (md.targetType.type !== "SimpleTypeExpression") throw new Error("Only simple target types supported in methods");
        const typeName = md.targetType.name.name;
        if (!this.inherentMethods.has(typeName)) this.inherentMethods.set(typeName, new Map());
        const bucket = this.inherentMethods.get(typeName)!;
        for (const def of md.definitions) {
          bucket.set(def.id.name, { kind: "function", node: def, closureEnv: env });
        }
        return { kind: "nil", value: null };
      }
      case "ImplementationDefinition": {
        const imp = node as AST.ImplementationDefinition;
        const variants = this.expandImplementationTargetVariants(imp.targetType);
        const unionVariantSignatures = imp.targetType.type === "UnionTypeExpression"
          ? [...new Set(variants.map(v => v.signature))].sort()
          : undefined;
        const unionSignatureKey = unionVariantSignatures ? unionVariantSignatures.join("|") : null;
        const funcs = new Map<string, Extract<V10Value, { kind: "function" }>>();
        for (const def of imp.definitions) {
          funcs.set(def.id.name, { kind: "function", node: def, closureEnv: env });
        }
        this.attachDefaultInterfaceMethods(imp, funcs);
        if (imp.implName) {
          // Named impl: expose as its own impl_namespace value; do not register for implicit resolution
          const name = imp.implName.name;
          const symMap = new Map<string, V10Value>();
          for (const [k, v] of funcs.entries()) symMap.set(k, v);
          const implVal: V10Value = {
            kind: "impl_namespace",
            def: imp,
            symbols: symMap,
            meta: { interfaceName: imp.interfaceName.name, target: imp.targetType, interfaceArgs: imp.interfaceArgs },
          };
          env.define(name, implVal);
          this.registerSymbol(name, implVal);
          const qn = this.qualifiedName(name);
          if (qn) { try { this.globals.define(qn, implVal); } catch {} }
        } else {
          // Unnamed impl: participates in implicit method resolution
          const constraintSpecs = this.collectConstraintSpecs(imp.genericParams, imp.whereClause);
          const baseConstraintSig = constraintSpecs
            .map(c => `${c.typeParam}->${this.typeExpressionToString(c.ifaceType)}`)
            .sort()
            .join("&") || "<none>";
          for (const variant of variants) {
            const typeName = variant.typeName;
            const targetArgTemplates = variant.argTemplates;
            const key = `${imp.interfaceName.name}::${typeName}`;
            if (!this.unnamedImplsSeen.has(key)) this.unnamedImplsSeen.set(key, new Map());
            const templateKeyBase = targetArgTemplates.length === 0
              ? "<none>"
              : targetArgTemplates.map(t => this.typeExpressionToString(t)).join("|");
            const templateKey = unionSignatureKey ? `${unionSignatureKey}::${templateKeyBase}` : templateKeyBase;
            const templateBucket = this.unnamedImplsSeen.get(key)!;
            if (!templateBucket.has(templateKey)) templateBucket.set(templateKey, new Set());
            const constraintKey = unionSignatureKey ? `${unionSignatureKey}::${baseConstraintSig}` : baseConstraintSig;
            const constraintSet = templateBucket.get(templateKey)!;
            if (constraintSet.has(constraintKey)) {
              throw new Error(`Unnamed impl for (${imp.interfaceName.name}, ${this.typeExpressionToString(imp.targetType)}) already exists`);
            }
            constraintSet.add(constraintKey);
            if (!this.implMethods.has(typeName)) this.implMethods.set(typeName, []);
            this.implMethods.get(typeName)!.push({
              def: imp,
              methods: funcs,
              targetArgTemplates,
              genericParams: imp.genericParams ?? [],
              whereClause: imp.whereClause,
              unionVariantSignatures,
            });
          }
        }
        return { kind: "nil", value: null };
      }

      default:
        throw new Error(`Not implemented in milestone: ${node.type}`);
    }
  }

  private enforceGenericConstraintsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall): void {
    const generics = (funcNode as any).genericParams as AST.GenericParameter[] | undefined;
    const where = (funcNode as any).whereClause as AST.WhereClauseConstraint[] | undefined;
    const typeArgs = call.typeArguments ?? [];
    const genericCount = generics ? generics.length : 0;
    if (genericCount > 0 && typeArgs.length !== genericCount) {
      const name = (funcNode as any).id?.name ?? "(lambda)";
      throw new Error(`Type arguments count mismatch calling ${name}: expected ${genericCount}, got ${typeArgs.length}`);
    }
    const constraints = this.collectConstraintSpecs(generics, where);
    if (constraints.length === 0) return;
    const name = (funcNode as any).id?.name ?? "(lambda)";
    const typeArgMap = this.mapTypeArguments(generics, typeArgs, `calling ${name}`);
    this.enforceConstraintSpecs(constraints, typeArgMap, `function ${name}`);
  }

  private collectConstraintSpecs(generics?: AST.GenericParameter[], where?: AST.WhereClauseConstraint[]): ConstraintSpec[] {
    const all: ConstraintSpec[] = [];
    if (generics) {
      for (const gp of generics) {
        if (!gp.constraints) continue;
        for (const c of gp.constraints) {
          all.push({ typeParam: gp.name.name, ifaceType: c.interfaceType });
        }
      }
    }
    if (where) {
      for (const clause of where) {
        for (const c of clause.constraints) {
          all.push({ typeParam: clause.typeParam.name, ifaceType: c.interfaceType });
        }
      }
    }
    return all;
  }

  private mapTypeArguments(
    generics: AST.GenericParameter[] | undefined,
    provided: AST.TypeExpression[] | undefined,
    context: string,
  ): Map<string, AST.TypeExpression> {
    const map = new Map<string, AST.TypeExpression>();
    if (!generics || generics.length === 0) return map;
    const actual = provided ?? [];
    if (actual.length !== generics.length) {
      throw new Error(`Type arguments count mismatch ${context}: expected ${generics.length}, got ${actual.length}`);
    }
    for (let i = 0; i < generics.length; i++) {
      const gp = generics[i]!;
      const ta = actual[i];
      if (!ta) {
        throw new Error(`Missing type argument for '${gp.name.name}' required by ${context}`);
      }
      map.set(gp.name.name, ta);
    }
    return map;
  }

  private enforceConstraintSpecs(constraints: ConstraintSpec[], typeArgMap: Map<string, AST.TypeExpression>, context: string): void {
    for (const c of constraints) {
      const actual = typeArgMap.get(c.typeParam);
      if (!actual) {
        throw new Error(`Missing type argument for '${c.typeParam}' required by constraints`);
      }
      const typeInfo = this.parseTypeExpression(actual);
      if (!typeInfo) continue;
      this.ensureTypeSatisfiesInterface(typeInfo, c.ifaceType, c.typeParam, new Set());
    }
  }

  private ensureTypeSatisfiesInterface(
    typeInfo: { name: string; typeArgs: AST.TypeExpression[] },
    interfaceType: AST.TypeExpression,
    context: string,
    visited: Set<string>,
  ): void {
    const ifaceInfo = this.parseTypeExpression(interfaceType);
    if (!ifaceInfo) return;
    if (visited.has(ifaceInfo.name)) return;
    visited.add(ifaceInfo.name);
    const iface = this.interfaces.get(ifaceInfo.name);
    if (!iface) throw new Error(`Unknown interface '${ifaceInfo.name}' in constraint on '${context}'`);
    for (const base of iface.baseInterfaces ?? []) {
      this.ensureTypeSatisfiesInterface(typeInfo, base, context, visited);
    }
    for (const sig of iface.signatures) {
      const methodName = sig.name.name;
      const method = this.findMethod(typeInfo.name, methodName, { typeArgs: typeInfo.typeArgs, interfaceName: ifaceInfo.name });
      if (!method) {
        throw new Error(`Type '${typeInfo.name}' does not satisfy interface '${ifaceInfo.name}': missing method '${methodName}'`);
      }
    }
  }

  private parseTypeExpression(t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null {
    if (t.type === "SimpleTypeExpression") {
      return { name: t.name.name, typeArgs: [] };
    }
    if (t.type === "GenericTypeExpression") {
      if (t.base.type === "SimpleTypeExpression") {
        return { name: t.base.name.name, typeArgs: t.arguments ?? [] };
      }
      return null;
    }
    return null;
  }

  private typeExpressionsEqual(a: AST.TypeExpression, b: AST.TypeExpression): boolean {
    return this.typeExpressionToString(a) === this.typeExpressionToString(b);
  }

  private bindTypeArgumentsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, env: Environment): void {
    const generics = (funcNode as any).genericParams as AST.GenericParameter[] | undefined;
    if (!generics || generics.length === 0) return;
    const args = call.typeArguments ?? [];
    // Do not throw for count mismatch here; constraints enforcement handles strictness when needed
    const count = Math.min(generics.length, args.length);
    for (let i = 0; i < count; i++) {
      const gp = generics[i]!;
      const ta = args[i]!;
      const name = `${gp.name.name}_type`;
      const s = this.typeExpressionToString(ta);
      try { env.define(name, { kind: "string", value: s }); } catch {}
    }
  }

  private typeExpressionToString(t: AST.TypeExpression): string {
    switch (t.type) {
      case "SimpleTypeExpression":
        return t.name.name;
      case "GenericTypeExpression": {
        const base = this.typeExpressionToString(t.base);
        const args = (t.arguments ?? []).map(a => this.typeExpressionToString(a)).join(", ");
        return `${base}<${args}>`;
      }
      case "NullableTypeExpression":
        return `${this.typeExpressionToString(t.innerType)}?`;
      case "FunctionTypeExpression":
        return `fn(${t.paramTypes.map(p => this.typeExpressionToString(p)).join(", ")}) -> ${this.typeExpressionToString(t.returnType)}`;
      case "UnionTypeExpression":
        return t.members.map(m => this.typeExpressionToString(m)).join(" | ");
      default:
        return "<?>";
    }
  }

  private isTruthy(v: V10Value): boolean {
    if (v.kind === "nil") return false;
    if (v.kind === "bool") return v.value;
    if (v.kind === "error") return false;
    if (v.kind === "i32" || v.kind === "f64") return v.value !== 0;
    if (v.kind === "string") return v.value.length > 0;
    return true;
  }

  private valueToString(v: V10Value): string {
    return this.valueToStringWithEnv(v, this.globals);
  }
  private valueToStringWithEnv(v: V10Value, env: Environment): string {
    switch (v.kind) {
      case "string": return v.value;
      case "bool": return String(v.value);
      case "char": return v.value;
      case "nil": return "nil";
      case "i32": return String(v.value);
      case "f64": return String(v.value);
      case "array": return `[${v.elements.map(e => this.valueToString(e)).join(", ")}]`;
      case "range": return `${v.start}${v.inclusive ? ".." : "..."}${v.end}`;
      case "function": return `<function>`;
      case "struct_def": return `<struct ${v.def.id.name}>`;
      case "interface_def": return `<interface ${v.def.id.name}>`;
      case "union_def": return `<union ${v.def.id.name}>`;
      case "struct_instance": {
        // Prefer to_string if available
        const toStr = this.findMethod(v.def.id.name, 'to_string', { typeArgs: v.typeArguments, typeArgMap: v.typeArgMap });
        if (toStr) {
          try {
            const funcNode = toStr.node;
            const funcEnv = new Environment(toStr.closureEnv);
            const firstParam = funcNode.params[0];
            if (firstParam) {
              if (firstParam.name.type === 'Identifier') funcEnv.define(firstParam.name.name, v);
              else this.assignByPattern(firstParam.name as any, v, funcEnv, true);
            }
            let rv: V10Value;
            try {
              rv = this.evaluate(funcNode.body, funcEnv);
            } catch (e) {
              if (e instanceof ReturnSignal) rv = e.value; else throw e;
            }
            if (rv.kind === 'string') return rv.value;
          } catch {}
        }
        if (Array.isArray(v.values)) {
          return `${v.def.id.name} { ${v.values.map(e => this.valueToString(e)).join(", ")} }`;
        } else {
          return `${v.def.id.name} { ${Array.from(v.values.entries()).map(([k, val]) => `${k}: ${this.valueToString(val)}`).join(", ")} }`;
        }
      }
      case "interface_value":
        return `<interface ${v.interfaceName}>`;
      case "proc_handle": return `<proc ${v.state}>`;
      case "future": return `<future ${v.state}>`;
      case "native_function": return `<native ${v.name}>`;
      case "native_bound_method": return `<native bound ${v.func.name}>`;
      case "error": return `<error ${v.message}>`;
    }
    return "<?>";
  }

  private tryMatchPattern(pattern: AST.Pattern, value: V10Value, baseEnv: Environment): Environment | null {
    // Identifier: bind value
    if (pattern.type === "Identifier") {
      const e = new Environment(baseEnv);
      e.define(pattern.name, value);
      return e;
    }
    // Wildcard: always matches, no bindings
    if (pattern.type === "WildcardPattern") {
      return new Environment(baseEnv);
    }
    // LiteralPattern: compare by evaluated literal
    if (pattern.type === "LiteralPattern") {
      const litVal = this.evaluate(pattern.literal, baseEnv);
      const equals = JSON.stringify(litVal) === JSON.stringify(value);
      return equals ? new Environment(baseEnv) : null;
    }
    // StructPattern
    if (pattern.type === "StructPattern") {
      if (value.kind !== "struct_instance") return null;
      if (pattern.structType && value.def.id.name !== pattern.structType.name) return null;
      let env = new Environment(baseEnv);
      if (pattern.isPositional) {
        if (!Array.isArray(value.values)) return null;
        if (pattern.fields.length !== value.values.length) return null;
        for (let i = 0; i < pattern.fields.length; i++) {
          const field = pattern.fields[i];
          const val = value.values[i];
          if (!field || val === undefined) return null;
          const sub = this.tryMatchPattern(field.pattern, val as V10Value, env);
          if (!sub) return null;
          env = sub;
        }
        return env;
      } else {
        if (!(value.values instanceof Map)) return null;
        for (const f of pattern.fields) {
          if (!f.fieldName) return null;
          const name = f.fieldName.name;
          if (!value.values.has(name)) return null;
          const sub = this.tryMatchPattern(f.pattern, value.values.get(name) as V10Value, env);
          if (!sub) return null;
          env = sub;
        }
        return env;
      }
    }
    // ArrayPattern
    if (pattern.type === "ArrayPattern") {
      if (value.kind !== "array") return null;
      const arr = value.elements;
      const minLen = pattern.elements.length;
      const hasRest = !!pattern.restPattern;
      if (!hasRest && arr.length !== minLen) return null;
      if (arr.length < minLen) return null;
      let env = new Environment(baseEnv);
      for (let i = 0; i < minLen; i++) {
        const pe = pattern.elements[i];
        const av = arr[i];
        if (!pe || av === undefined) return null;
        const sub = this.tryMatchPattern(pe, av, env);
        if (!sub) return null;
        env = sub;
      }
      if (hasRest && pattern.restPattern) {
        if (pattern.restPattern.type === "Identifier") {
          env.define(pattern.restPattern.name, { kind: "array", elements: arr.slice(minLen) });
        }
      }
      return env;
    }
    // TypedPattern: ignore type annotation, just match inner
    if ((pattern as any).type === "TypedPattern") {
      const tp = pattern as AST.TypedPattern;
      if (!this.matchesType(tp.typeAnnotation, value)) return null;
      const coerced = this.coerceValueToType(tp.typeAnnotation, value);
      return this.tryMatchPattern(tp.pattern, coerced, baseEnv);
    }
    return null;
  }

  private assignByPattern(pattern: AST.Pattern, value: V10Value, env: Environment, isDeclaration: boolean): void {
    if (pattern.type === "Identifier") {
      if (isDeclaration) env.define(pattern.name, value); else env.assign(pattern.name, value);
      return;
    }
    if (pattern.type === "WildcardPattern") return;
    if (pattern.type === "LiteralPattern") {
      const lit = this.evaluate(pattern.literal, env);
      if (JSON.stringify(lit) !== JSON.stringify(value)) throw new Error("Pattern literal mismatch in assignment");
      return;
    }
    if (pattern.type === "StructPattern") {
      if (value.kind !== "struct_instance") throw new Error("Cannot destructure non-struct value");
      if (pattern.structType && value.def.id.name !== pattern.structType.name) throw new Error("Struct type mismatch in destructuring");
      if (pattern.isPositional) {
        if (!Array.isArray(value.values)) throw new Error("Expected positional struct");
        if (pattern.fields.length !== value.values.length) throw new Error("Struct field count mismatch");
        for (let i = 0; i < pattern.fields.length; i++) {
          const fieldPat = pattern.fields[i];
          const fieldVal = value.values[i];
          if (!fieldPat || fieldVal === undefined) throw new Error("Invalid positional field during destructuring");
          this.assignByPattern(fieldPat.pattern, fieldVal, env, isDeclaration);
        }
        return;
      } else {
        if (!(value.values instanceof Map)) throw new Error("Expected named struct");
        for (const f of pattern.fields) {
          if (!f.fieldName) throw new Error("Named struct pattern missing field name");
          const name = f.fieldName.name;
          if (!value.values.has(name)) throw new Error(`Missing field '${name}' during destructuring`);
          const fieldVal = value.values.get(name)!;
          this.assignByPattern(f.pattern, fieldVal, env, isDeclaration);
        }
        return;
      }
    }
    if (pattern.type === "ArrayPattern") {
      if (value.kind !== "array") throw new Error("Cannot destructure non-array value");
      const arr = value.elements;
      const minLen = pattern.elements.length;
      const hasRest = !!pattern.restPattern;
      if (!hasRest && arr.length !== minLen) throw new Error("Array length mismatch in destructuring");
      if (arr.length < minLen) throw new Error("Array too short for destructuring");
      for (let i = 0; i < minLen; i++) {
        const pe = pattern.elements[i];
        const av = arr[i];
        if (!pe || av === undefined) throw new Error("Invalid array element during destructuring");
        this.assignByPattern(pe, av, env, isDeclaration);
      }
      if (hasRest && pattern.restPattern && pattern.restPattern.type === "Identifier") {
        const rest = { kind: "array", elements: arr.slice(minLen) } as V10Value;
        if (isDeclaration) env.define(pattern.restPattern.name, rest); else env.assign(pattern.restPattern.name, rest);
      }
      return;
    }
    if ((pattern as any).type === "TypedPattern") {
      const tp = pattern as AST.TypedPattern;
      if (!this.matchesType(tp.typeAnnotation, value)) throw new Error("Typed pattern mismatch in assignment");
      const coerced = this.coerceValueToType(tp.typeAnnotation, value);
      this.assignByPattern(tp.pattern, coerced, env, isDeclaration);
      return;
    }
    throw new Error(`Unsupported pattern in assignment: ${(pattern as any).type}`);
  }

  private matchesType(t: AST.TypeExpression, v: V10Value): boolean {
    switch (t.type) {
      case "WildcardTypeExpression":
        return true;
      case "SimpleTypeExpression": {
        const name = t.name.name;
        if (name === "string") return v.kind === "string";
        if (name === "bool") return v.kind === "bool";
        if (name === "char") return v.kind === "char";
        if (name === "i32") return v.kind === "i32";
        if (name === "f64") return v.kind === "f64";
        if (name === "Error") return v.kind === "error";
        if (this.interfaces.has(name)) {
          if (v.kind === "interface_value") return v.interfaceName === name;
          const typeName = this.getTypeNameForValue(v);
          if (!typeName) return false;
          return this.typeImplementsInterface(typeName, name);
        }
        // Treat other simple names as struct types
        return v.kind === "struct_instance" && v.def.id.name === name;
      }
      case "GenericTypeExpression": {
        // Support Array T
        if (t.base.type === "SimpleTypeExpression" && t.base.name.name === "Array") {
          if (v.kind !== "array") return false;
          if (!t.arguments || t.arguments.length === 0) return true;
          const elemT = t.arguments[0]!;
          return v.elements.every(el => this.matchesType(elemT, el));
        }
        return true;
      }
      case "FunctionTypeExpression":
        return v.kind === "function";
      case "NullableTypeExpression":
        if (v.kind === "nil") return true;
        return this.matchesType(t.innerType, v);
      case "ResultTypeExpression":
        // Minimal: treat like inner type
        return this.matchesType(t.innerType, v);
      case "UnionTypeExpression":
        return t.members.some(member => this.matchesType(member, v));
      default:
        return true;
    }
  }
  private computeBinaryForCompound(op: string, left: V10Value, right: V10Value): V10Value {
    const be: AST.BinaryExpression = { type: 'BinaryExpression', operator: op, left: { type: 'NilLiteral', value: null } as any, right: { type: 'NilLiteral', value: null } as any };
    // Reuse numeric/bitwise logic by switching on op and kinds
    // Implement minimal inline logic mirroring BinaryExpression cases for i32/f64 and bitwise i32.
    const isNum = (v: V10Value) => v.kind === 'i32' || v.kind === 'f64';
    const asNumber = (v: V10Value) => (v.kind === 'i32' || v.kind === 'f64') ? v.value : NaN;
    const resultKind = (a: V10Value, c: V10Value): 'i32' | 'f64' => (a.kind === 'f64' || c.kind === 'f64') ? 'f64' : 'i32';
    if (["+","-","*","/","%"].includes(op)) {
      if (!isNum(left) || !isNum(right)) throw new Error("Arithmetic requires numeric operands");
      const kind = resultKind(left, right);
      const l = asNumber(left); const r = asNumber(right);
      switch (op) {
        case '+': return kind === 'i32' ? { kind, value: (l + r) | 0 } : { kind, value: l + r } as any;
        case '-': return kind === 'i32' ? { kind, value: (l - r) | 0 } : { kind, value: l - r } as any;
        case '*': return kind === 'i32' ? { kind, value: (l * r) | 0 } : { kind, value: l * r } as any;
        case '/': if (r === 0) throw new Error("Division by zero"); return kind === 'i32' ? { kind, value: (l / r) | 0 } : { kind, value: l / r } as any;
        case '%': if (r === 0) throw new Error("Division by zero"); return { kind, value: kind === 'i32' ? (l % r) | 0 : l % r } as any;
      }
    }
    if (["&","|","^","<<",">>"].includes(op)) {
      if (left.kind !== 'i32' || right.kind !== 'i32') throw new Error("Bitwise requires i32 operands");
      switch (op) {
        case '&': return { kind: 'i32', value: left.value & right.value };
        case '|': return { kind: 'i32', value: left.value | right.value };
        case '^': return { kind: 'i32', value: left.value ^ right.value };
        case '<<': {
          const count = right.value;
          if (count < 0 || count >= 32) throw new Error("shift out of range");
          return { kind: 'i32', value: left.value << count };
        }
        case '>>': {
          const count = right.value;
          if (count < 0 || count >= 32) throw new Error("shift out of range");
          return { kind: 'i32', value: left.value >> count };
        }
      }
    }
    throw new Error(`Unsupported compound operator ${op}`);
  }
  private findMethod(
    typeName: string,
    methodName: string,
    opts?: { typeArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; interfaceName?: string },
  ): Extract<V10Value, { kind: "function" }> | null {
    const inherent = this.inherentMethods.get(typeName);
    if (inherent && inherent.has(methodName)) return inherent.get(methodName)!;
    const entries = this.implMethods.get(typeName);
    let constraintError: Error | null = null;
    const matches: Array<{
      method: Extract<V10Value, { kind: "function" }>;
      score: number;
      entry: ImplMethodEntry;
      constraints: ConstraintSpec[];
    }> = [];
    if (entries) {
      for (const entry of entries) {
        if (opts?.interfaceName && entry.def.interfaceName.name !== opts.interfaceName) continue;
        const bindings = this.matchImplEntry(entry, opts);
        if (!bindings) continue;
        const constraints = this.collectConstraintSpecs(entry.genericParams, entry.whereClause);
        if (constraints.length > 0) {
          try {
            this.enforceConstraintSpecs(constraints, bindings, `impl ${entry.def.interfaceName.name} for ${typeName}`);
          } catch (err) {
            if (!constraintError && err instanceof Error) constraintError = err;
            continue;
          }
        }
        const method = entry.methods.get(methodName);
        if (!method) continue;
        const score = this.computeImplSpecificity(entry, bindings, constraints);
        matches.push({ method, score, entry, constraints });
      }
    }
    if (matches.length === 0) {
      if (constraintError) throw constraintError;
      return null;
    }
    const [firstMatch, ...remainingMatches] = matches;
    let best = firstMatch!;
    let contenders: typeof matches = [best];
    for (const candidate of remainingMatches) {
      const cmp = this.compareMethodMatches(candidate, best);
      if (cmp > 0) {
        best = candidate;
        contenders = [candidate];
        continue;
      }
      if (cmp === 0) {
        const reverse = this.compareMethodMatches(best, candidate);
        if (reverse < 0) {
          best = candidate;
          contenders = [candidate];
        } else if (reverse === 0) {
          contenders.push(candidate);
        }
      }
    }
    if (contenders.length > 1) {
      const detail = Array.from(new Set(contenders.map(c => this.typeExpressionToString(c.entry.def.targetType)))).join(", ");
      throw new Error(`Ambiguous method '${methodName}' for type '${typeName}' (candidates: ${detail})`);
    }
    return best.method;
  }

  private compareMethodMatches(
    a: { score: number; entry: ImplMethodEntry; constraints: ConstraintSpec[] },
    b: { score: number; entry: ImplMethodEntry; constraints: ConstraintSpec[] },
  ): number {
    if (a.score > b.score) return 1;
    if (a.score < b.score) return -1;
    const aUnion = a.entry.unionVariantSignatures;
    const bUnion = b.entry.unionVariantSignatures;
    if (aUnion && !bUnion) return -1;
    if (!aUnion && bUnion) return 1;
    if (aUnion && bUnion) {
      if (this.isProperSubset(aUnion, bUnion)) return 1;
      if (this.isProperSubset(bUnion, aUnion)) return -1;
      if (aUnion.length !== bUnion.length) {
        return aUnion.length < bUnion.length ? 1 : -1;
      }
    }
    const aConstraints = this.buildConstraintKeySet(a.constraints);
    const bConstraints = this.buildConstraintKeySet(b.constraints);
    if (this.isConstraintSuperset(aConstraints, bConstraints)) return 1;
    if (this.isConstraintSuperset(bConstraints, aConstraints)) return -1;
    return 0;
  }

  private buildConstraintKeySet(constraints: ConstraintSpec[]): Set<string> {
    const set = new Set<string>();
    for (const c of constraints) {
      const expanded = this.collectInterfaceConstraintExpressions(c.ifaceType);
      for (const expr of expanded) {
        set.add(`${c.typeParam}->${this.typeExpressionToString(expr)}`);
      }
    }
    return set;
  }

  private isConstraintSuperset(a: Set<string>, b: Set<string>): boolean {
    if (a.size <= b.size) return false;
    for (const key of b) {
      if (!a.has(key)) return false;
    }
    return true;
  }

  private isProperSubset(a: string[], b: string[]): boolean {
    const aSet = new Set(a);
    const bSet = new Set(b);
    if (aSet.size >= bSet.size) return false;
    for (const val of aSet) {
      if (!bSet.has(val)) return false;
    }
    return true;
  }

  private collectInterfaceConstraintExpressions(typeExpr: AST.TypeExpression, memo: Set<string> = new Set()): AST.TypeExpression[] {
    const key = this.typeExpressionToString(typeExpr);
    if (memo.has(key)) return [];
    memo.add(key);
    const expressions: AST.TypeExpression[] = [typeExpr];
    if (typeExpr.type === "SimpleTypeExpression") {
      const iface = this.interfaces.get(typeExpr.name.name);
      if (iface && iface.baseInterfaces) {
        for (const base of iface.baseInterfaces) {
          const cloned = this.cloneTypeExpression(base);
          expressions.push(...this.collectInterfaceConstraintExpressions(cloned, memo));
        }
      }
    }
    return expressions;
  }

  private matchImplEntry(
    entry: ImplMethodEntry,
    opts?: { typeArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> },
  ): Map<string, AST.TypeExpression> | null {
    const bindings = new Map<string, AST.TypeExpression>();
    const genericNames = new Set(entry.genericParams.map(g => g.name.name));
    const expectedArgs = entry.targetArgTemplates;
    const actualArgs = opts?.typeArgs;
    if (expectedArgs.length > 0) {
      if (!actualArgs || actualArgs.length !== expectedArgs.length) return null;
      for (let i = 0; i < expectedArgs.length; i++) {
        const template = expectedArgs[i]!;
        const actual = actualArgs[i]!;
        if (!this.matchTypeExpressionTemplate(template, actual, genericNames, bindings)) return null;
      }
    }
    if (opts?.typeArgMap) {
      for (const [k, v] of opts.typeArgMap.entries()) {
        if (!bindings.has(k)) bindings.set(k, v);
      }
    }
    for (const gp of entry.genericParams) {
      if (!bindings.has(gp.name.name)) return null;
    }
    return bindings;
  }

  private matchTypeExpressionTemplate(
    template: AST.TypeExpression,
    actual: AST.TypeExpression,
    genericNames: Set<string>,
    bindings: Map<string, AST.TypeExpression>,
  ): boolean {
    if (template.type === "SimpleTypeExpression") {
      const name = template.name.name;
      if (genericNames.has(name)) {
        const existing = bindings.get(name);
        if (existing) return this.typeExpressionsEqual(existing, actual);
        bindings.set(name, actual);
        return true;
      }
      return this.typeExpressionsEqual(template, actual);
    }
    if (template.type === "GenericTypeExpression") {
      if (actual.type !== "GenericTypeExpression") return false;
      if (!this.matchTypeExpressionTemplate(template.base, actual.base, genericNames, bindings)) return false;
      const templateArgs = template.arguments ?? [];
      const actualArgs = actual.arguments ?? [];
      if (templateArgs.length !== actualArgs.length) return false;
      for (let i = 0; i < templateArgs.length; i++) {
        if (!this.matchTypeExpressionTemplate(templateArgs[i]!, actualArgs[i]!, genericNames, bindings)) return false;
      }
      return true;
    }
    return this.typeExpressionsEqual(template, actual);
  }

  private expandImplementationTargetVariants(
    target: AST.TypeExpression,
  ): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> {
    if (target.type === "UnionTypeExpression") {
      const expanded: Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> = [];
      for (const member of target.members) {
        const memberVariants = this.expandImplementationTargetVariants(member);
        for (const variant of memberVariants) expanded.push(variant);
      }
      const seen = new Set<string>();
      const unique: Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> = [];
      for (const variant of expanded) {
        if (seen.has(variant.signature)) continue;
        seen.add(variant.signature);
        unique.push(variant);
      }
      if (unique.length === 0) {
        throw new Error("Union target must contain at least one concrete type");
      }
      return unique;
    }
    if (target.type === "SimpleTypeExpression") {
      const signature = this.typeExpressionToString(target);
      return [{ typeName: target.name.name, argTemplates: [], signature }];
    }
    if (target.type === "GenericTypeExpression" && target.base.type === "SimpleTypeExpression") {
      const signature = this.typeExpressionToString(target);
      return [{ typeName: target.base.name.name, argTemplates: target.arguments ?? [], signature }];
    }
    throw new Error("Only simple, generic, or union target types supported in impl");
  }

  private computeImplSpecificity(
    entry: ImplMethodEntry,
    bindings: Map<string, AST.TypeExpression>,
    constraints: ConstraintSpec[],
  ): number {
    const genericNames = new Set(entry.genericParams.map(g => g.name.name));
    let concreteScore = 0;
    for (const template of entry.targetArgTemplates) {
      concreteScore += this.measureTemplateSpecificity(template, genericNames);
    }
    const constraintScore = constraints.length;
    const bindingScore = bindings.size;
    const unionPenalty = entry.unionVariantSignatures ? entry.unionVariantSignatures.length : 0;
    return concreteScore * 100 + constraintScore * 10 + bindingScore - unionPenalty;
  }

  private measureTemplateSpecificity(t: AST.TypeExpression, genericNames: Set<string>): number {
    switch (t.type) {
      case "SimpleTypeExpression":
        return genericNames.has(t.name.name) ? 0 : 1;
      case "GenericTypeExpression": {
        let score = this.measureTemplateSpecificity(t.base, genericNames);
        for (const arg of t.arguments ?? []) {
          score += this.measureTemplateSpecificity(arg, genericNames);
        }
        return score;
      }
      case "NullableTypeExpression":
      case "ResultTypeExpression":
        return this.measureTemplateSpecificity(t.innerType, genericNames);
      case "UnionTypeExpression":
        return t.members.reduce((acc, member) => acc + this.measureTemplateSpecificity(member, genericNames), 0);
      default:
        return 0;
    }
  }

  private attachDefaultInterfaceMethods(
    imp: AST.ImplementationDefinition,
    funcs: Map<string, Extract<V10Value, { kind: "function" }>>,
  ): void {
    const interfaceName = imp.interfaceName.name;
    const iface = this.interfaces.get(interfaceName);
    if (!iface) return;
    const ifaceEnv = this.interfaceEnvs.get(interfaceName) ?? this.globals;
    const targetType = imp.targetType;
    for (const sig of iface.signatures) {
      if (!sig.defaultImpl) continue;
      const methodName = sig.name.name;
      if (funcs.has(methodName)) continue;
      const defaultFunc = this.createDefaultMethodFunction(sig, ifaceEnv, targetType);
      if (defaultFunc) funcs.set(methodName, defaultFunc);
    }
  }

  private createDefaultMethodFunction(
    sig: AST.FunctionSignature,
    env: Environment,
    targetType: AST.TypeExpression,
  ): Extract<V10Value, { kind: "function" }> | null {
    if (!sig.defaultImpl) return null;
    const params = sig.params.map(param => {
      const substitutedPattern = this.substituteSelfInPattern(param.name as AST.Pattern, targetType);
      const substitutedType = this.substituteSelfTypeExpression(param.paramType, targetType);
      if (substitutedPattern === param.name && substitutedType === param.paramType) return param;
      return { type: "FunctionParameter", name: substitutedPattern, paramType: substitutedType } as AST.FunctionParameter;
    });
    const returnType = this.substituteSelfTypeExpression(sig.returnType, targetType) ?? sig.returnType;
    const fnDef: AST.FunctionDefinition = {
      type: "FunctionDefinition",
      id: sig.name,
      params,
      returnType,
      genericParams: sig.genericParams,
      whereClause: sig.whereClause,
      body: sig.defaultImpl,
      isMethodShorthand: false,
      isPrivate: false,
    };
    return { kind: "function", node: fnDef, closureEnv: env };
  }

  private substituteSelfTypeExpression(
    t: AST.TypeExpression | undefined,
    target: AST.TypeExpression,
  ): AST.TypeExpression | undefined {
    if (!t) return t;
    switch (t.type) {
      case "SimpleTypeExpression":
        if (t.name.name === "Self") return this.cloneTypeExpression(target);
        return t;
      case "GenericTypeExpression": {
        const base = this.substituteSelfTypeExpression(t.base, target) ?? t.base;
        const args = t.arguments?.map(arg => this.substituteSelfTypeExpression(arg, target) ?? arg) ?? [];
        if (base === t.base && args.every((arg, idx) => arg === (t.arguments ?? [])[idx])) return t;
        return { type: "GenericTypeExpression", base, arguments: args };
      }
      case "FunctionTypeExpression": {
        const paramTypes = t.paramTypes.map(pt => this.substituteSelfTypeExpression(pt, target) ?? pt);
        const returnType = this.substituteSelfTypeExpression(t.returnType, target) ?? t.returnType;
        if (paramTypes.every((pt, idx) => pt === t.paramTypes[idx]) && returnType === t.returnType) return t;
        return { type: "FunctionTypeExpression", paramTypes, returnType };
      }
      case "NullableTypeExpression": {
        const inner = this.substituteSelfTypeExpression(t.innerType, target) ?? t.innerType;
        if (inner === t.innerType) return t;
        return { type: "NullableTypeExpression", innerType: inner };
      }
      case "ResultTypeExpression": {
        const inner = this.substituteSelfTypeExpression(t.innerType, target) ?? t.innerType;
        if (inner === t.innerType) return t;
        return { type: "ResultTypeExpression", innerType: inner };
      }
      case "UnionTypeExpression": {
        let changed = false;
        const members = t.members.map(member => {
          const next = this.substituteSelfTypeExpression(member, target) ?? member;
          if (next !== member) changed = true;
          return next;
        });
        if (!changed) return t;
        return { type: "UnionTypeExpression", members };
      }
      case "WildcardTypeExpression":
      default:
        return t;
    }
  }

  private substituteSelfInPattern(pattern: AST.Pattern, target: AST.TypeExpression): AST.Pattern {
    if ((pattern as any).type === "TypedPattern") {
      const tp = pattern as AST.TypedPattern;
      const inner = this.substituteSelfInPattern(tp.pattern, target);
      const typeAnnotation = this.substituteSelfTypeExpression(tp.typeAnnotation, target) ?? tp.typeAnnotation;
      if (inner === tp.pattern && typeAnnotation === tp.typeAnnotation) return tp;
      return { type: "TypedPattern", pattern: inner, typeAnnotation };
    }
    if (pattern.type === "StructPattern") {
      let changed = false;
      const fields = pattern.fields.map(field => {
        const newPattern = this.substituteSelfInPattern(field.pattern, target);
        if (newPattern !== field.pattern) {
          changed = true;
          return { ...field, pattern: newPattern };
        }
        return field;
      });
      let structType = pattern.structType;
      if (structType && structType.name === "Self" && target.type === "SimpleTypeExpression") {
        structType = AST.identifier(target.name.name);
        changed = true;
      }
      if (!changed) return pattern;
      return { ...pattern, fields, structType };
    }
    if (pattern.type === "ArrayPattern") {
      let changed = false;
      const elements = pattern.elements.map(el => {
        if (!el) return el;
        const newEl = this.substituteSelfInPattern(el, target);
        if (newEl !== el) changed = true;
        return newEl ?? el;
      });
      const restPattern = pattern.restPattern
        ? (this.substituteSelfInPattern(pattern.restPattern, target) as AST.Identifier | AST.WildcardPattern)
        : undefined;
      if (restPattern !== pattern.restPattern) changed = true;
      if (!changed) return pattern;
      return { ...pattern, elements, restPattern };
    }
    return pattern;
  }

  private cloneTypeExpression(t: AST.TypeExpression): AST.TypeExpression {
    switch (t.type) {
      case "SimpleTypeExpression":
        return { type: "SimpleTypeExpression", name: AST.identifier(t.name.name) };
      case "GenericTypeExpression":
        return {
          type: "GenericTypeExpression",
          base: this.cloneTypeExpression(t.base),
          arguments: (t.arguments ?? []).map(arg => this.cloneTypeExpression(arg)),
        };
      case "FunctionTypeExpression":
        return {
          type: "FunctionTypeExpression",
          paramTypes: t.paramTypes.map(pt => this.cloneTypeExpression(pt)),
          returnType: this.cloneTypeExpression(t.returnType),
        };
      case "NullableTypeExpression":
        return { type: "NullableTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "ResultTypeExpression":
        return { type: "ResultTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "UnionTypeExpression":
        return { type: "UnionTypeExpression", members: t.members.map(member => this.cloneTypeExpression(member)) };
      case "WildcardTypeExpression":
      default:
        return { type: "WildcardTypeExpression" };
    }
  }

  private typeImplementsInterface(typeName: string, interfaceName: string): boolean {
    const entries = this.implMethods.get(typeName);
    if (!entries) return false;
    for (const entry of entries) {
      if (entry.def.interfaceName.name === interfaceName) return true;
    }
    return false;
  }

  private getTypeNameForValue(value: V10Value): string | null {
    switch (value.kind) {
      case "struct_instance":
        return value.def.id.name;
      case "interface_value":
        return this.getTypeNameForValue(value.value);
      case "i32":
        return "i32";
      case "f64":
        return "f64";
      case "string":
        return "string";
      case "bool":
        return "bool";
      case "char":
        return "char";
      case "array":
        return "Array";
      case "range":
        return "Range";
      default:
        return null;
    }
  }

  private coerceValueToType(typeExpr: AST.TypeExpression | undefined, value: V10Value): V10Value {
    if (!typeExpr) return value;
    if (typeExpr.type === "SimpleTypeExpression") {
      const name = typeExpr.name.name;
      if (this.interfaces.has(name)) {
        return this.toInterfaceValue(name, value);
      }
    }
    return value;
  }

  private toInterfaceValue(interfaceName: string, rawValue: V10Value): V10Value {
    if (!this.interfaces.has(interfaceName)) {
      throw new Error(`Unknown interface '${interfaceName}'`);
    }
    if (rawValue.kind === "interface_value") {
      if (rawValue.interfaceName === interfaceName) return rawValue;
      return this.toInterfaceValue(interfaceName, rawValue.value);
    }
    const typeName = this.getTypeNameForValue(rawValue);
    if (!typeName || !this.typeImplementsInterface(typeName, interfaceName)) {
      throw new Error(`Type '${typeName ?? "<unknown>"}' does not implement interface '${interfaceName}'`);
    }
    let typeArguments: AST.TypeExpression[] | undefined;
    let typeArgMap: Map<string, AST.TypeExpression> | undefined;
    if (rawValue.kind === "struct_instance") {
      typeArguments = rawValue.typeArguments;
      typeArgMap = rawValue.typeArgMap;
    }
    return { kind: "interface_value", interfaceName, value: rawValue, typeArguments, typeArgMap };
  }

  private initConcurrencyBuiltins(): void {
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
  }

  private scheduleAsync(fn: () => void): void {
    this.schedulerQueue.push(fn);
    this.ensureSchedulerTick();
  }

  private ensureSchedulerTick(): void {
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
  }

  private currentAsyncContext():
    | { kind: "proc"; handle: Extract<V10Value, { kind: "proc_handle" }> }
    | { kind: "future"; handle: Extract<V10Value, { kind: "future" }> }
    | null {
    if (this.asyncContextStack.length === 0) return null;
    return this.asyncContextStack[this.asyncContextStack.length - 1];
  }

  private procYield(): V10Value {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_yield must be called inside an asynchronous task");
    throw new ProcYieldSignal();
  }

  private procCancelled(): V10Value {
    const ctx = this.currentAsyncContext();
    if (!ctx) throw new Error("proc_cancelled must be called inside an asynchronous task");
    if (ctx.kind === "proc") {
      return { kind: "bool", value: !!ctx.handle.cancelRequested };
    }
    return { kind: "bool", value: false };
  }

  private processScheduler(limit: number = this.schedulerMaxSteps): void {
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
  }

  private makeNamedStructInstance(def: AST.StructDefinition, entries: Array<[string, V10Value]>): V10Value {
    const map = new Map<string, V10Value>();
    for (const [key, value] of entries) map.set(key, value);
    return { kind: "struct_instance", def, values: map };
  }

  private makeProcError(details: string): V10Value {
    return this.makeNamedStructInstance(this.procErrorStruct, [["details", { kind: "string", value: details }]]);
  }

  private getProcErrorDetails(procError: V10Value): string {
    if (procError.kind === "struct_instance" && procError.def.id.name === "ProcError") {
      const map = procError.values as Map<string, V10Value>;
      const detailsVal = map.get("details");
      if (detailsVal && detailsVal.kind === "string") return detailsVal.value;
    }
    return "unknown failure";
  }

  private makeProcStatusFailed(procError: V10Value): V10Value {
    return this.makeNamedStructInstance(this.procStatusStructs.Failed, [["error", procError]]);
  }

  private markProcCancelled(handle: Extract<V10Value, { kind: "proc_handle" }>, message = "Proc cancelled"): void {
    const procErr = this.makeProcError(message);
    handle.state = "cancelled";
    handle.result = undefined;
    handle.failureInfo = procErr;
    handle.error = this.makeRuntimeError(message, procErr);
    handle.runner = null;
  }

  private procHandleStatus(handle: Extract<V10Value, { kind: "proc_handle" }>): V10Value {
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
  }

  private futureStatus(future: Extract<V10Value, { kind: "future" }>): V10Value {
    switch (future.state) {
      case "pending":
        return this.procStatusPendingValue;
      case "resolved":
        return this.procStatusResolvedValue;
      case "failed": {
        const procErr = future.failureInfo ?? this.makeProcError("unknown failure");
        return this.makeProcStatusFailed(procErr);
      }
      default:
        return this.procStatusPendingValue;
    }
  }

  private toProcError(value: V10Value | undefined, fallback: string): V10Value {
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
  }

  private makeNativeFunction(
    name: string,
    arity: number,
    impl: (interpreter: InterpreterV10, args: V10Value[]) => V10Value,
  ): Extract<V10Value, { kind: "native_function" }> {
    return { kind: "native_function", name, arity, impl };
  }

  private bindNativeMethod(
    func: Extract<V10Value, { kind: "native_function" }>,
    self: V10Value,
  ): Extract<V10Value, { kind: "native_bound_method" }> {
    return { kind: "native_bound_method", func, self };
  }

  private procHandleValue(handle: Extract<V10Value, { kind: "proc_handle" }>): V10Value {
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
  }

  private procHandleCancel(handle: Extract<V10Value, { kind: "proc_handle" }>): void {
    if (handle.state === "resolved" || handle.state === "failed" || handle.state === "cancelled") return;
    handle.cancelRequested = true;
    if (handle.state === "pending" && !handle.isEvaluating) {
      if (!handle.runner) handle.runner = () => this.runProcHandle(handle);
      this.scheduleAsync(handle.runner);
    }
  }

  private futureValue(future: Extract<V10Value, { kind: "future" }>): V10Value {
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
  }

  private runProcHandle(handle: Extract<V10Value, { kind: "proc_handle" }>): void {
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
  }

  private runFuture(future: Extract<V10Value, { kind: "future" }>): void {
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
  }

  private makeRuntimeError(message: string, value?: V10Value): V10Value {
    return { kind: "error", message, value };
  }

  private tryUfcs(env: Environment, funcName: string, receiver: V10Value): Extract<V10Value, { kind: "bound_method" }> | null {
    try {
      const candidate = env.get(funcName);
      if (candidate && candidate.kind === 'function') {
        return { kind: 'bound_method', func: candidate, self: receiver };
      }
    } catch {}
    return null;
  }
}

export function evaluate(node: AST.AstNode | null, env?: Environment): V10Value {
  return new InterpreterV10().evaluate(node, env);
}
