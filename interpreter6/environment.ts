import type { AbleValue } from "./runtime";

export class Environment {
  private values: Map<string, AbleValue> = new Map();
  constructor(private enclosing: Environment | null = null) {}

  define(name: string, value: AbleValue): void {
    if (this.values.has(name)) {
      console.warn(`Warning: Redefining variable "${name}" in the same scope.`);
    }
    this.values.set(name, value);
  }

  assign(name: string, value: AbleValue): void {
    if (this.values.has(name)) {
      this.values.set(name, value);
      return;
    }
    if (this.enclosing !== null) {
      this.enclosing.assign(name, value);
      return;
    }
    throw new Error(`Interpreter Error: Undefined variable '${name}' for assignment.`);
  }

  get(name: string): AbleValue {
    if (this.values.has(name)) {
      return this.values.get(name)!;
    }
    if (this.enclosing !== null) {
      return this.enclosing.get(name);
    }
    throw new Error(`Interpreter Error: Undefined variable '${name}'.`);
  }

  hasInCurrentScope(name: string): boolean {
    return this.values.has(name);
  }
}
