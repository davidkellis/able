import type { V10Value } from "./values";

export class Environment {
  private values: Map<string, V10Value> = new Map();

  constructor(private enclosing: Environment | null = null) {}

  define(name: string, value: V10Value): void {
    if (this.values.has(name)) {
      throw new Error(`Redefinition in current scope: ${name}`);
    }
    this.values.set(name, value);
  }

  assign(name: string, value: V10Value): void {
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

  get(name: string): V10Value {
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

  assignExisting(name: string, value: V10Value): boolean {
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
