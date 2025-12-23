import type { RuntimeValue } from "./values";

export class Environment {
  private values: Map<string, RuntimeValue> = new Map();

  constructor(private enclosing: Environment | null = null) {}

  define(name: string, value: RuntimeValue): void {
    if (this.values.has(name)) {
      const existing = this.values.get(name)!;
      const merged = mergeFunctionValue(existing, value);
      if (!merged) {
        throw new Error(`Redefinition in current scope: ${name}`);
      }
      this.values.set(name, merged);
      return;
    }
    this.values.set(name, value);
  }

  assign(name: string, value: RuntimeValue): void {
    if (this.values.has(name)) {
      this.values.set(name, value);
      return;
    }
    if (this.enclosing) {
      this.enclosing.assign(name, value);
      return;
    }
    throw new Error(`Undefined variable '${name}'`);
  }

  get(name: string): RuntimeValue {
    if (this.values.has(name)) {
      return this.values.get(name)!;
    }
    if (this.enclosing) {
      return this.enclosing.get(name);
    }
    throw new Error(`Undefined variable '${name}'`);
  }

  has(name: string): boolean {
    if (this.values.has(name)) {
      return true;
    }
    return this.enclosing ? this.enclosing.has(name) : false;
  }

  hasInCurrentScope(name: string): boolean {
    return this.values.has(name);
  }

  assignExisting(name: string, value: RuntimeValue): boolean {
    if (this.values.has(name)) {
      this.values.set(name, value);
      return true;
    }
    if (this.enclosing) {
      return this.enclosing.assignExisting(name, value);
    }
    return false;
  }
}

function mergeFunctionValue(existing: RuntimeValue, incoming: RuntimeValue): RuntimeValue | null {
  const isFunctionLike = (v: RuntimeValue) => v.kind === "function" || v.kind === "function_overload";
  if (isFunctionLike(existing) && isFunctionLike(incoming)) {
    const overloads: Array<Extract<RuntimeValue, { kind: "function" }>> = [];
    if (existing.kind === "function") overloads.push(existing);
    if (existing.kind === "function_overload") overloads.push(...existing.overloads);
    if (incoming.kind === "function") overloads.push(incoming);
    if (incoming.kind === "function_overload") overloads.push(...incoming.overloads);
    return { kind: "function_overload", overloads };
  }
  return null;
}
