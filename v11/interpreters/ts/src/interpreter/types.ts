import * as AST from "../ast";
import type { InterpreterV10 } from "./index";
import { getIntegerInfo, isIntegerValue, makeIntegerValue } from "./numeric";
import type { FloatKind, IntegerKind, V10Value } from "./values";

const INTEGER_KINDS: IntegerKind[] = ["i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128"];
const FLOAT_KINDS: FloatKind[] = ["f32", "f64"];
const INTEGER_KIND_SET: Set<IntegerKind> = new Set(INTEGER_KINDS);

function integerRangeWithin(source: IntegerKind, target: IntegerKind): boolean {
  const sourceInfo = getIntegerInfo(source);
  const targetInfo = getIntegerInfo(target);
  return sourceInfo.min >= targetInfo.min && sourceInfo.max <= targetInfo.max;
}

declare module "./index" {
  interface InterpreterV10 {
    typeExpressionToString(t: AST.TypeExpression): string;
    parseTypeExpression(t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null;
    typeExpressionsEqual(a: AST.TypeExpression, b: AST.TypeExpression): boolean;
    cloneTypeExpression(t: AST.TypeExpression): AST.TypeExpression;
    typeExpressionFromValue(value: V10Value): AST.TypeExpression | null;
    matchesType(t: AST.TypeExpression, v: V10Value): boolean;
    getTypeNameForValue(value: V10Value): string | null;
    typeImplementsInterface(typeName: string, interfaceName: string): boolean;
    coerceValueToType(typeExpr: AST.TypeExpression | undefined, value: V10Value): V10Value;
    toInterfaceValue(interfaceName: string, rawValue: V10Value): V10Value;
  }
}

export function applyTypesAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.typeExpressionToString = function typeExpressionToString(this: InterpreterV10, t: AST.TypeExpression): string {
    switch (t.type) {
      case "SimpleTypeExpression":
        return t.name.name;
      case "GenericTypeExpression":
        return `${this.typeExpressionToString(t.base)}<${(t.arguments ?? []).map(arg => this.typeExpressionToString(arg)).join(",")}>`;
      case "NullableTypeExpression":
        return `${this.typeExpressionToString(t.innerType)}?`;
      case "ResultTypeExpression":
        return `Result<${this.typeExpressionToString(t.innerType)}>`;
      case "FunctionTypeExpression":
        return `(${t.paramTypes.map(pt => this.typeExpressionToString(pt)).join(", ")}) -> ${this.typeExpressionToString(t.returnType)}`;
      case "UnionTypeExpression":
        return t.members.map(member => this.typeExpressionToString(member)).join(" | ");
      case "WildcardTypeExpression":
        return "_";
      default:
        return "<type>";
    }
  };

  cls.prototype.parseTypeExpression = function parseTypeExpression(this: InterpreterV10, t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null {
    if (t.type === "SimpleTypeExpression") {
      return { name: t.name.name, typeArgs: [] };
    }
    if (t.type === "GenericTypeExpression" && t.base.type === "SimpleTypeExpression") {
      return { name: t.base.name.name, typeArgs: t.arguments ?? [] };
    }
    return null;
  };

  cls.prototype.cloneTypeExpression = function cloneTypeExpression(this: InterpreterV10, t: AST.TypeExpression): AST.TypeExpression {
    switch (t.type) {
      case "SimpleTypeExpression":
        return { type: "SimpleTypeExpression", name: AST.identifier(t.name.name) };
      case "GenericTypeExpression":
        return {
          type: "GenericTypeExpression",
          base: this.cloneTypeExpression(t.base),
          arguments: (t.arguments ?? []).map(arg => this.cloneTypeExpression(arg)),
        };
      case "FunctionTypeExpression":
        return {
          type: "FunctionTypeExpression",
          paramTypes: t.paramTypes.map(pt => this.cloneTypeExpression(pt)),
          returnType: this.cloneTypeExpression(t.returnType),
        };
      case "NullableTypeExpression":
        return { type: "NullableTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "ResultTypeExpression":
        return { type: "ResultTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "UnionTypeExpression":
        return { type: "UnionTypeExpression", members: t.members.map(member => this.cloneTypeExpression(member)) };
      case "WildcardTypeExpression":
      default:
        return { type: "WildcardTypeExpression" };
    }
  };

  cls.prototype.typeExpressionsEqual = function typeExpressionsEqual(this: InterpreterV10, a: AST.TypeExpression, b: AST.TypeExpression): boolean {
    return this.typeExpressionToString(a) === this.typeExpressionToString(b);
  };

  cls.prototype.typeExpressionFromValue = function typeExpressionFromValue(this: InterpreterV10, value: V10Value): AST.TypeExpression | null {
    switch (value.kind) {
      case "string":
        return AST.simpleTypeExpression("string");
      case "bool":
        return AST.simpleTypeExpression("bool");
      case "char":
        return AST.simpleTypeExpression("char");
      case "nil":
        return AST.simpleTypeExpression("nil");
      case "error":
        return AST.simpleTypeExpression("Error");
      case "struct_instance": {
        const base = AST.simpleTypeExpression(value.def.id.name);
        if (value.typeArguments && value.typeArguments.length > 0) {
          return AST.genericTypeExpression(base, value.typeArguments);
        }
        return base;
      }
      case "interface_value":
        return AST.simpleTypeExpression(value.interfaceName);
      case "array": {
        const state = this.ensureArrayState ? this.ensureArrayState(value) : { values: value.elements };
        const elems = state.values ?? value.elements ?? [];
        let elemType: AST.TypeExpression | null = null;
        for (const el of elems) {
          const inferred = this.typeExpressionFromValue(el);
          if (!inferred) continue;
          if (!elemType) {
            elemType = inferred;
            continue;
          }
          if (!this.typeExpressionsEqual(elemType, inferred)) {
            elemType = AST.wildcardTypeExpression();
            break;
          }
        }
        if (!elemType) elemType = AST.wildcardTypeExpression();
        return AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [elemType]);
      }
      case "hash_map":
        return AST.simpleTypeExpression("Map");
      default: {
        if (value.kind === "interface_def" && (value as any).def?.id?.name) {
          return AST.simpleTypeExpression((value as any).def.id.name);
        }
        if ((value as any).kind === "i8" || (value as any).kind === "i16" || (value as any).kind === "i32" || (value as any).kind === "i64" || (value as any).kind === "i128" || (value as any).kind === "u8" || (value as any).kind === "u16" || (value as any).kind === "u32" || (value as any).kind === "u64" || (value as any).kind === "u128") {
          return AST.simpleTypeExpression((value as any).kind);
        }
        if ((value as any).kind === "f32" || (value as any).kind === "f64") {
          return AST.simpleTypeExpression((value as any).kind);
        }
        return null;
      }
    }
  };

  cls.prototype.matchesType = function matchesType(this: InterpreterV10, t: AST.TypeExpression, v: V10Value): boolean {
    switch (t.type) {
      case "WildcardTypeExpression":
        return true;
      case "SimpleTypeExpression": {
        const name = t.name.name;
        if (name === "Self") return true;
        if (/^[A-Z]$/.test(name)) return true;
        if (name === "string") return v.kind === "string";
        if (name === "bool") return v.kind === "bool";
        if (name === "char") return v.kind === "char";
        if (name === "nil") return v.kind === "nil";
        if (INTEGER_KINDS.includes(name as IntegerKind)) {
          if (!isIntegerValue(v)) {
            return false;
          }
          const actualKind = v.kind as IntegerKind;
          const expectedKind = name as IntegerKind;
          return actualKind === expectedKind || integerRangeWithin(actualKind, expectedKind);
        }
        if (FLOAT_KINDS.includes(name as FloatKind)) return v.kind === name;
        if (name === "Error") return v.kind === "error";
        if (this.interfaces.has(name)) {
          if (v.kind === "interface_value") return v.interfaceName === name;
          const typeName = this.getTypeNameForValue(v);
          if (!typeName) return false;
          return this.typeImplementsInterface(typeName, name);
        }
        return v.kind === "struct_instance" && v.def.id.name === name;
      }
      case "GenericTypeExpression": {
        if (t.base.type === "SimpleTypeExpression" && t.base.name.name === "Array") {
          const isArrayValue = v.kind === "array" || (v.kind === "struct_instance" && v.def.id.name === "Array");
          if (!isArrayValue) return false;
          if (!t.arguments || t.arguments.length === 0) return true;
          const elemT = t.arguments[0]!;
          if (v.kind === "array") {
            return v.elements.every(el => this.matchesType(elemT, el));
          }
          return true;
        }
        return true;
      }
      case "FunctionTypeExpression":
        return v.kind === "function";
      case "NullableTypeExpression":
        if (v.kind === "nil") return true;
        return this.matchesType(t.innerType, v);
      case "ResultTypeExpression":
        return this.matchesType(t.innerType, v);
      case "UnionTypeExpression":
        return t.members.some(member => this.matchesType(member, v));
      default:
        return true;
    }
  };

  cls.prototype.getTypeNameForValue = function getTypeNameForValue(this: InterpreterV10, value: V10Value): string | null {
    if (INTEGER_KINDS.includes(value.kind as IntegerKind)) {
      return value.kind;
    }
    if (FLOAT_KINDS.includes(value.kind as FloatKind)) {
      return value.kind;
    }
    switch (value.kind) {
      case "struct_instance":
        return value.def.id.name;
      case "interface_value":
        return this.getTypeNameForValue(value.value);
      case "string":
        return "string";
      case "bool":
        return "bool";
      case "char":
        return "char";
      case "array":
        return "Array";
      default:
        return null;
    }
  };

  cls.prototype.typeImplementsInterface = function typeImplementsInterface(this: InterpreterV10, typeName: string, interfaceName: string): boolean {
    const entries = this.implMethods.get(typeName);
    if (!entries) return false;
    for (const entry of entries) {
      if (entry.def.interfaceName.name === interfaceName) return true;
    }
    return false;
  };

  cls.prototype.coerceValueToType = function coerceValueToType(this: InterpreterV10, typeExpr: AST.TypeExpression | undefined, value: V10Value): V10Value {
    if (!typeExpr) return value;
    if (typeExpr.type === "SimpleTypeExpression") {
      const name = typeExpr.name.name;
      if (INTEGER_KIND_SET.has(name as IntegerKind) && isIntegerValue(value)) {
        const targetKind = name as IntegerKind;
        const actualKind = value.kind as IntegerKind;
        if (actualKind !== targetKind && integerRangeWithin(actualKind, targetKind)) {
          return makeIntegerValue(targetKind, value.value);
        }
      }
      if (this.interfaces.has(name)) {
        return this.toInterfaceValue(name, value);
      }
    }
    return value;
  };

  cls.prototype.toInterfaceValue = function toInterfaceValue(this: InterpreterV10, interfaceName: string, rawValue: V10Value): V10Value {
    if (!this.interfaces.has(interfaceName)) {
      throw new Error(`Unknown interface '${interfaceName}'`);
    }
    if (rawValue.kind === "interface_value") {
      if (rawValue.interfaceName === interfaceName) return rawValue;
      return this.toInterfaceValue(interfaceName, rawValue.value);
    }
    const typeName = this.getTypeNameForValue(rawValue);
    if (!typeName || !this.typeImplementsInterface(typeName, interfaceName)) {
      throw new Error(`Type '${typeName ?? "<unknown>"}' does not implement interface '${interfaceName}'`);
    }
    let typeArguments: AST.TypeExpression[] | undefined;
    let typeArgMap: Map<string, AST.TypeExpression> | undefined;
    if (rawValue.kind === "struct_instance") {
      typeArguments = rawValue.typeArguments;
      typeArgMap = rawValue.typeArgMap;
    }
    return { kind: "interface_value", interfaceName, value: rawValue, typeArguments, typeArgMap };
  };
}
