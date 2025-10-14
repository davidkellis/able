import * as AST from "../ast";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    typeExpressionToString(t: AST.TypeExpression): string;
    parseTypeExpression(t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null;
    typeExpressionsEqual(a: AST.TypeExpression, b: AST.TypeExpression): boolean;
    cloneTypeExpression(t: AST.TypeExpression): AST.TypeExpression;
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

  cls.prototype.matchesType = function matchesType(this: InterpreterV10, t: AST.TypeExpression, v: V10Value): boolean {
    switch (t.type) {
      case "WildcardTypeExpression":
        return true;
      case "SimpleTypeExpression": {
        const name = t.name.name;
        if (name === "string") return v.kind === "string";
        if (name === "bool") return v.kind === "bool";
        if (name === "char") return v.kind === "char";
        if (name === "i32") return v.kind === "i32";
        if (name === "f64") return v.kind === "f64";
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
          if (v.kind !== "array") return false;
          if (!t.arguments || t.arguments.length === 0) return true;
          const elemT = t.arguments[0]!;
          return v.elements.every(el => this.matchesType(elemT, el));
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
    switch (value.kind) {
      case "struct_instance":
        return value.def.id.name;
      case "interface_value":
        return this.getTypeNameForValue(value.value);
      case "i32":
        return "i32";
      case "f64":
        return "f64";
      case "string":
        return "string";
      case "bool":
        return "bool";
      case "char":
        return "char";
      case "array":
        return "Array";
      case "range":
        return "Range";
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
