import type { Interpreter } from "./interpreter";
import type { AbleValue } from "./runtime";
import { hasKind, isAbleArray } from "./runtime";

export function isTruthy(this: Interpreter, value: AbleValue): boolean {
  if (hasKind(value, "bool")) return value.value;
  if (hasKind(value, "nil") || hasKind(value, "void")) return false;
  if (
    ((hasKind(value, "i32") ||
      hasKind(value, "f64") ||
      hasKind(value, "i8") ||
      hasKind(value, "i16") ||
      hasKind(value, "u8") ||
      hasKind(value, "u16") ||
      hasKind(value, "u32") ||
      hasKind(value, "f32")) &&
      value.value === 0)
  ) {
    return false;
  }
  if ((hasKind(value, "i64") || hasKind(value, "i128") || hasKind(value, "u64") || hasKind(value, "u128")) && value.value === 0n) {
    return false;
  }
  if (hasKind(value, "string") && value.value === "") return false;
  if (isAbleArray(value) && value.elements.length === 0) return false;
  return true;
}

export function valueToString(this: Interpreter, value: AbleValue): string {
  const self = this as any;
  if (value === null || value === undefined) return "<?>";
  switch (value.kind) {
    case "i8":
    case "i16":
    case "i32":
    case "u8":
    case "u16":
    case "u32":
    case "f32":
    case "f64":
      return value.value.toString();
    case "i64":
    case "i128":
    case "u64":
    case "u128":
      return value.value.toString();
    case "string":
      return value.value;
    case "bool":
      return value.value.toString();
    case "char":
      return `'${value.value}'`;
    case "nil":
      return "nil";
    case "void":
      return "void";
    case "function": {
      if (typeof (value as any).apply === "function" && value.node) {
        const funcName = value.node?.type === "FunctionDefinition" && value.node.id ? value.node.id.name : "(bound method)";
        return `<function ${funcName}>`;
      }
      const funcName = value.node?.type === "FunctionDefinition" && value.node.id ? value.node.id.name : "(anonymous)";
      return `<function ${funcName}>`;
    }
    case "struct_definition":
      return `<struct ${value.name}>`;
    case "struct_instance":
      if (Array.isArray(value.values)) {
        return `${value.definition.name} { ${value.values.map((v) => self.valueToString(v)).join(", ")} }`;
      }
      if (value.values instanceof Map) {
        const fields = Array.from(value.values.entries())
          .map(([key, val]) => `${key}: ${self.valueToString(val)}`)
          .join(", ");
        return `${value.definition.name} { ${fields} }`;
      }
      return `${value.definition.name} { ... }`;
    case "array":
      return `[${value.elements.map((v) => self.valueToString(v)).join(", ")}]`;
    case "error":
      return `<error: ${value.message}>`;
    case "union_definition":
      return `<union ${value.name}>`;
    case "interface_definition":
      return `<interface ${value.name}>`;
    case "implementation_definition": {
      const ifaceName = value.implNode.interfaceName?.name ?? "?";
      const targetTypeName = (value.implNode.targetType as any)?.name?.name ?? "?";
      return `<impl ${ifaceName} for ${targetTypeName}>`;
    }
    case "methods_collection": {
      const methodsTargetTypeName = (value.methodsNode.targetType as any)?.name?.name ?? "?";
      return `<methods for ${methodsTargetTypeName}>`;
    }
    case "proc_handle":
      return `<proc ${value.id}>`;
    case "thunk":
      return `<thunk ${value.id}>`;
    case "range":
      return `${value.start}${value.inclusive ? ".." : "..."}${value.end}`;
    case "AbleIterator":
      return `<iterator>`;
    default: {
      const _exhaustiveCheck: never = value;
      return `<${(_exhaustiveCheck as any).kind}>`;
    }
  }
}
