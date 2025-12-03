import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    registerSymbol(name: string, value: V10Value): void;
    qualifiedName(name: string): string | null;
    isTruthy(v: V10Value): boolean;
  }
}

export function applyHelperAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.registerSymbol = function registerSymbol(this: InterpreterV10, name: string, value: V10Value): void {
    if (!this.currentPackage) return;
    if (!this.packageRegistry.has(this.currentPackage)) this.packageRegistry.set(this.currentPackage, new Map());
    const bucket = this.packageRegistry.get(this.currentPackage)!;
    const existing = bucket.get(name);
    if (existing && isFunctionLike(existing) && isFunctionLike(value)) {
      const overloads = [];
      if (existing.kind === "function") overloads.push(existing);
      if (existing.kind === "function_overload") overloads.push(...existing.overloads);
      if (value.kind === "function") overloads.push(value);
      if (value.kind === "function_overload") overloads.push(...value.overloads);
      bucket.set(name, { kind: "function_overload", overloads });
    } else {
      bucket.set(name, value);
    }
  };

  cls.prototype.qualifiedName = function qualifiedName(this: InterpreterV10, name: string): string | null {
    return this.currentPackage ? `${this.currentPackage}.${name}` : null;
  };

  cls.prototype.isTruthy = function isTruthy(this: InterpreterV10, v: V10Value): boolean {
    switch (v.kind) {
      case "bool":
        return v.value;
      case "nil":
        return false;
      default:
        return true;
    }
  };
}

function isFunctionLike(v: V10Value): v is Extract<V10Value, { kind: "function" | "function_overload" }> {
  return v.kind === "function" || v.kind === "function_overload";
}
