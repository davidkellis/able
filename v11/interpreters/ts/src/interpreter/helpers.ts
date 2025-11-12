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
    this.packageRegistry.get(this.currentPackage)!.set(name, value);
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
