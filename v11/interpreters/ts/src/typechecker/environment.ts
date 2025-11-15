import type { TypeInfo } from "./types";

type Scope = Map<string, TypeInfo>;

export class Environment {
  private readonly scopes: Scope[] = [new Map()];

  pushScope(): void {
    this.scopes.push(new Map());
  }

  popScope(): void {
    if (this.scopes.length === 1) {
      throw new Error("typechecker: cannot pop global scope");
    }
    this.scopes.pop();
  }

  define(name: string, type: TypeInfo): void {
    this.scopes[this.scopes.length - 1].set(name, type);
  }

  assign(name: string, type: TypeInfo): boolean {
    for (let i = this.scopes.length - 1; i >= 0; i -= 1) {
      const scope = this.scopes[i];
      if (scope.has(name)) {
        scope.set(name, type);
        return true;
      }
    }
    return false;
  }

  lookup(name: string): TypeInfo | undefined {
    for (let i = this.scopes.length - 1; i >= 0; i -= 1) {
      const scope = this.scopes[i];
      const type = scope.get(name);
      if (type) {
        return type;
      }
    }
    return undefined;
  }

  has(name: string): boolean {
    return this.lookup(name) !== undefined;
  }

  hasInCurrentScope(name: string): boolean {
    if (this.scopes.length === 0) {
      return false;
    }
    return this.scopes[this.scopes.length - 1].has(name);
  }

  fork(): Environment {
    const forked = new Environment();
    forked.scopes.length = 0;
    for (const scope of this.scopes) {
      forked.scopes.push(new Map(scope));
    }
    return forked;
  }
}
