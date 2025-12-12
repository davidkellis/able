import type * as AST from "../ast";
import { ReturnSignal } from "./signals";
import { Environment } from "./environment";
import type { V10Value } from "./values";
import type { InterpreterV10 } from "./index";
import { isFloatValue, isIntegerValue } from "./numeric";

declare module "./index" {
  interface InterpreterV10 {
    valueToString(v: V10Value): string;
    valueToStringWithEnv(v: V10Value, env: Environment): string;
  }
}

export function applyStringifyAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.valueToString = function valueToString(this: InterpreterV10, v: V10Value): string {
    return this.valueToStringWithEnv(v, this.globals);
  };

  cls.prototype.valueToStringWithEnv = function valueToStringWithEnv(this: InterpreterV10, v: V10Value, env: Environment): string {
    if (isIntegerValue(v) || isFloatValue(v)) {
      return String(v.value);
    }
    switch (v.kind) {
      case "String": return v.value;
      case "bool": return String(v.value);
      case "char": return v.value;
      case "nil": return "nil";
      case "array": return `[${v.elements.map(e => this.valueToString(e)).join(", ")}]`;
      case "function": return "<function>";
      case "struct_def": return `<struct ${v.def.id.name}>`;
      case "interface_def": return `<interface ${v.def.id.name}>`;
      case "union_def": return `<union ${v.def.id.name}>`;
      case "struct_instance": {
        const toStr =
          this.findMethod(v.def.id.name, "to_String", { typeArgs: v.typeArguments, typeArgMap: v.typeArgMap }) ??
          this.findMethod(v.def.id.name, "to_string", { typeArgs: v.typeArguments, typeArgMap: v.typeArgMap });
        if (toStr) {
          try {
            const funcNode = toStr.node;
            const funcEnv = new Environment(toStr.closureEnv);
            const firstParam = funcNode.params[0];
            if (firstParam) {
              if (firstParam.name.type === "Identifier") funcEnv.define(firstParam.name.name, v);
              else this.assignByPattern(firstParam.name as AST.Pattern, v, funcEnv, true);
            }
            let rv: V10Value;
            try {
              rv = this.evaluate(funcNode.body, funcEnv);
            } catch (e) {
              if (e instanceof ReturnSignal) rv = e.value; else throw e;
            }
            if (rv.kind === "String") return rv.value;
          } catch {}
        }
        const renderEntries = (entries: string[]): string => {
          const inner = entries.join(", ");
          return inner ? `${v.def.id.name} { ${inner} }` : `${v.def.id.name} { }`;
        };
        if (Array.isArray(v.values)) {
          return renderEntries(v.values.map(e => this.valueToString(e)));
        }
        return renderEntries(Array.from(v.values.entries()).map(([k, val]) => `${k}: ${this.valueToString(val)}`));
      }
      case "package": return `<package ${v.name}>`;
      case "impl_namespace": return `<impl ${v.def.interfaceName.name} for ${v.meta.target.type === "SimpleTypeExpression" ? v.meta.target.name.name : "target"}>`;
      case "interface_value": return `<interface ${v.interfaceName}>`;
      case "proc_handle": return `<proc ${v.state}>`;
      case "future": return `<future ${v.state}>`;
      case "iterator": return "<iterator>";
      case "iterator_end": return "IteratorEnd";
      case "native_function": return `<native ${v.name}>`;
      case "native_bound_method": return `<native bound ${v.func.name}>`;
      case "partial_function": return "<partial>";
      case "error": return `<error ${v.message}>`;
      case "dyn_package": return `<dyn package ${v.name}>`;
      case "dyn_ref": return `<dyn ref ${v.pkg}.${v.name}>`;
      case "bound_method": return `<bound method ${this.valueToString(v.func)}>`;
      default:
        return "<unknown>";
    }
  };
}
