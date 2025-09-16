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
  | { kind: "range"; start: number; end: number; inclusive: boolean };

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
            case "<<": return { kind: "i32", value: left.value << right.value };
            case ">>": return { kind: "i32", value: left.value >> right.value };
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

      // --- Expressions scaffold we will fill later ---
      case "Identifier": return env.get((node as AST.Identifier).name);
      case "BlockExpression": {
        const block = node as AST.BlockExpression;
        const blockEnv = new Environment(env);
        let last: V10Value = { kind: "nil", value: null };
        for (const stmt of block.body) last = this.evaluate(stmt, blockEnv);
        return last;
      }
      case "AssignmentExpression": {
        const a = node as AST.AssignmentExpression;
        if (a.left.type !== "Identifier") throw new Error("Only simple identifier assignment supported in milestone 1");
        const val = this.evaluate(a.right, env);
        if (a.operator === ":=") env.define(a.left.name, val); else env.assign(a.left.name, val);
        return val;
      }

      default:
        throw new Error(`Not implemented in milestone: ${node.type}`);
    }
  }
}

export function evaluate(node: AST.AstNode | null, env?: Environment): V10Value {
  return new InterpreterV10().evaluate(node, env);
}


