import * as AST from "./ast";

// =============================================================================
// v10 Interpreter (initial scaffold)
// =============================================================================

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
  | { kind: "struct_instance"; def: AST.StructDefinition; values: V10Value[] | Map<string, V10Value> }
  | { kind: "interface_def"; def: AST.InterfaceDefinition }
  | { kind: "union_def"; def: AST.UnionDefinition }
  | { kind: "package"; name: string; symbols: Map<string, V10Value> }
  | { kind: "impl_namespace"; def: AST.ImplementationDefinition; symbols: Map<string, V10Value> }
  | { kind: "dyn_package"; name: string }
  | { kind: "dyn_ref"; pkg: string; name: string }
  | { kind: "error"; message: string; value?: V10Value }
  | { kind: "bound_method"; func: Extract<V10Value, { kind: "function" }>; self: V10Value };

class ReturnSignal extends Error {
  constructor(public value: V10Value) { super("ReturnSignal"); }
}
class RaiseSignal extends Error {
  constructor(public value: V10Value) { super("RaiseSignal"); }
}
class BreakSignalShim extends Error { constructor(){ super("BreakSignal"); } }
class BreakLabelSignal extends Error { constructor(public label: string, public value: V10Value){ super("BreakLabelSignal"); } }

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
  private inherentMethods: Map<string, Map<string, Extract<V10Value, { kind: "function" }>>> = new Map();
  private implMethods: Map<string, Map<string, Extract<V10Value, { kind: "function" }>>> = new Map();
  private unnamedImplsSeen: Set<string> = new Set();
  private raiseStack: V10Value[] = [];
  private packageRegistry: Map<string, Map<string, V10Value>> = new Map();
  private currentPackage: string | null = null;
  private breakpointStack: string[] = [];

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
          // Only Identifier and WildcardPattern for now
          if (fl.pattern.type === "Identifier") {
            targetEnv.define(fl.pattern.name, value);
          } else if (fl.pattern.type === "WildcardPattern") {
            // ignore
          } else {
            throw new Error("Only Identifier and Wildcard patterns supported in for-loop for now");
          }
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
        let funcValue: Extract<V10Value, { kind: "function" }>;
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
        } else {
          throw new Error("Cannot call non-function");
        }
        const funcNode = funcValue.node;
        const funcEnv = new Environment(funcValue.closureEnv);
        // Bind params (identifier-only for now)
        const params = funcNode.type === "FunctionDefinition" ? funcNode.params : funcNode.params;
        const evalArgs: V10Value[] = [...injectedArgs, ...call.arguments.map(a => this.evaluate(a, env))];
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
          if (p.name.type === "Identifier") {
            funcEnv.define(p.name.name, argVal);
          } else if (p.name.type === "WildcardPattern") {
            // ignore
          } else if (p.name.type === "StructPattern" || p.name.type === "ArrayPattern" || p.name.type === "LiteralPattern") {
            // destructuring param
            this.assignByPattern(p.name as any, argVal, funcEnv, true);
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
        if (sl.isPositional) {
          // Positional: build array values in order
          const vals: V10Value[] = sl.fields.map(f => this.evaluate(f.value, env));
          return { kind: "struct_instance", def: structDef, values: vals };
        } else {
          // Named: build map with optional functional update source
          const map = new Map<string, V10Value>();
          if (sl.functionalUpdateSource) {
            const baseVal = this.evaluate(sl.functionalUpdateSource, env);
            if (baseVal.kind !== "struct_instance") throw new Error("Functional update source must be a struct instance");
            if (baseVal.def.id.name !== structDef.id.name) throw new Error("Functional update source must be same struct type");
            if (!(baseVal.values instanceof Map)) throw new Error("Functional update only supported for named structs");
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
          return { kind: "struct_instance", def: structDef, values: map };
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
        if (obj.kind === "impl_namespace") {
          if (ma.member.type !== "Identifier") throw new Error("Impl namespace member access expects identifier");
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
          const method = this.findMethod(typeName, ma.member.name);
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

      // --- Concurrency (sync placeholders) ---
      case "ProcExpression": {
        const pr = node as AST.ProcExpression;
        return this.evaluate(pr.expression, env);
      }
      case "SpawnExpression": {
        const sp = node as AST.SpawnExpression;
        return this.evaluate(sp.expression, env);
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
        // Only SimpleTypeExpression target supported for now
        if (imp.targetType.type !== "SimpleTypeExpression") throw new Error("Only simple target types supported in impl");
        const typeName = imp.targetType.name.name;
        const funcs = new Map<string, Extract<V10Value, { kind: "function" }>>();
        for (const def of imp.definitions) {
          funcs.set(def.id.name, { kind: "function", node: def, closureEnv: env });
        }
        if (imp.implName) {
          // Named impl: expose as its own impl_namespace value; do not register for implicit resolution
          const name = imp.implName.name;
          const symMap = new Map<string, V10Value>();
          for (const [k, v] of funcs.entries()) symMap.set(k, v);
          const implVal: V10Value = { kind: "impl_namespace", def: imp, symbols: symMap };
          env.define(name, implVal);
          this.registerSymbol(name, implVal);
          const qn = this.qualifiedName(name);
          if (qn) { try { this.globals.define(qn, implVal); } catch {} }
        } else {
          // Unnamed impl: participates in implicit method resolution
          const key = `${imp.interfaceName.name}::${typeName}`;
          if (this.unnamedImplsSeen.has(key)) {
            throw new Error(`Unnamed impl for (${imp.interfaceName.name}, ${typeName}) already exists`);
          }
          this.unnamedImplsSeen.add(key);
          if (!this.implMethods.has(typeName)) this.implMethods.set(typeName, new Map());
          const bucket = this.implMethods.get(typeName)!;
          for (const [k, v] of funcs.entries()) bucket.set(k, v);
        }
        return { kind: "nil", value: null };
      }

      default:
        throw new Error(`Not implemented in milestone: ${node.type}`);
    }
  }

  private isTruthy(v: V10Value): boolean {
    if (v.kind === "nil") return false;
    if (v.kind === "bool") return v.value;
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
        const toStr = this.findMethod(v.def.id.name, 'to_string');
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
      return this.tryMatchPattern(tp.pattern, value, baseEnv);
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
      this.assignByPattern(tp.pattern, value, env, isDeclaration);
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
  private findMethod(typeName: string, methodName: string): Extract<V10Value, { kind: "function" }> | null {
    const inherent = this.inherentMethods.get(typeName);
    if (inherent && inherent.has(methodName)) return inherent.get(methodName)!;
    const impls = this.implMethods.get(typeName);
    if (impls && impls.has(methodName)) return impls.get(methodName)!;
    return null;
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


