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
  | { kind: "array"; elements: V10Value[] };

export class Environment {
  private values: Map<string, V10Value> = new Map();
  constructor(private enclosing: Environment | null = null) {}
  define(name: string, value: V10Value): void { this.values.set(name, value); }
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


