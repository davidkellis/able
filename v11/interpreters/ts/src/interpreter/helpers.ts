import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

declare module "./index" {
  interface Interpreter {
    registerSymbol(name: string, value: RuntimeValue): void;
    qualifiedName(name: string): string | null;
    isTruthy(v: RuntimeValue): boolean;
  }
}

export function applyHelperAugmentations(cls: typeof Interpreter): void {
  cls.prototype.registerSymbol = function registerSymbol(this: Interpreter, name: string, value: RuntimeValue): void {
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

  cls.prototype.qualifiedName = function qualifiedName(this: Interpreter, name: string): string | null {
    return this.currentPackage ? `${this.currentPackage}.${name}` : null;
  };

  cls.prototype.isTruthy = function isTruthy(this: Interpreter, v: RuntimeValue): boolean {
    switch (v.kind) {
      case "bool":
        return v.value;
      case "nil":
        return false;
      case "error":
        return false;
      case "interface_value":
        if (v.interfaceName === "Error") return false;
        break;
      default:
        break;
    }
    const typeName = this.getTypeNameForValue(v);
    if (!typeName) return true;
    const typeArgs = v.kind === "struct_instance" ? v.typeArguments : undefined;
    return !this.typeImplementsInterface(typeName, "Error", typeArgs);
  };
}

function isFunctionLike(v: RuntimeValue): v is Extract<RuntimeValue, { kind: "function" | "function_overload" }> {
  return v.kind === "function" || v.kind === "function_overload";
}
