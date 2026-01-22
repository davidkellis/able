import * as AST from "../../ast";
import type { Interpreter } from "../index";
import type { RuntimeValue } from "../values";
import {
  TypeImplementsInterfaceOptions,
  buildInterfaceMethodDictionary,
  normalizeTypeImplementsInterfaceOptions,
} from "./helpers";

export function applyTypeStructAugmentations(cls: typeof Interpreter): void {
  cls.prototype.typeImplementsInterface = function typeImplementsInterface(
    this: Interpreter,
    typeName: string,
    interfaceName: string,
    opts?: AST.TypeExpression[] | TypeImplementsInterfaceOptions,
  ): boolean {
    const options = normalizeTypeImplementsInterfaceOptions(opts);
    const subjectTypeFor = (name: string): AST.TypeExpression => {
      if (options.subjectType && name === typeName) {
        return options.subjectType;
      }
      if (options.subjectTypeArgs && options.subjectTypeArgs.length > 0) {
        return AST.genericTypeExpression(AST.simpleTypeExpression(name), options.subjectTypeArgs);
      }
      return AST.simpleTypeExpression(name);
    };
    const interfaceInfoFromExpr = (expr: AST.TypeExpression | undefined): { name: string; args?: AST.TypeExpression[] } | null => {
      if (!expr) return null;
      if (expr.type === "SimpleTypeExpression") {
        return { name: expr.name.name };
      }
      if (expr.type === "GenericTypeExpression" && expr.base.type === "SimpleTypeExpression") {
        return { name: expr.base.name.name, args: expr.arguments ?? [] };
      }
      return null;
    };

    const directImplements = (name: string, ifaceName: string, args?: AST.TypeExpression[]): boolean => {
      if (ifaceName === "Error" && name === "Error") {
        return true;
      }
      const subjectType = subjectTypeFor(name);
      const entries = [
        ...(this.implMethods.get(name) ?? []),
        ...this.genericImplMethods,
      ];
      if (entries.length === 0) return false;
      const interfaceArgs = args && args.length > 0 ? args : undefined;
      for (const entry of entries) {
        if (entry.def.interfaceName.name !== ifaceName) continue;
        if (this.matchImplEntry(entry, { subjectType, typeArgs: options.subjectTypeArgs, interfaceArgs })) return true;
      }
      return false;
    };

    const interfaceExtends = (candidate: string, target: string, visited: Set<string>): boolean => {
      if (!candidate || !target) return false;
      if (candidate === target) return true;
      if (visited.has(candidate)) return false;
      visited.add(candidate);
      const ifaceDef = this.interfaces.get(candidate);
      if (!ifaceDef?.baseInterfaces || ifaceDef.baseInterfaces.length === 0) {
        return false;
      }
      for (const base of ifaceDef.baseInterfaces) {
        const info = interfaceInfoFromExpr(base);
        if (!info) continue;
        if (info.name === target) return true;
        if (interfaceExtends(info.name, target, visited)) return true;
      }
      return false;
    };

    const check = (name: string, ifaceName: string, args: AST.TypeExpression[] | undefined, visited: Set<string>): boolean => {
      const argKey = (args ?? []).map((arg) => this.typeExpressionToString(arg)).join("|");
      const key = `${name}::${ifaceName}::${argKey}`;
      if (visited.has(key)) return true;
      visited.add(key);
      const ifaceDef = this.interfaces.get(ifaceName);
      if (ifaceDef?.baseInterfaces && ifaceDef.baseInterfaces.length > 0) {
        for (const base of ifaceDef.baseInterfaces) {
          const info = interfaceInfoFromExpr(base);
          if (!info) return false;
          if (!check(name, info.name, info.args, visited)) return false;
        }
        if (!ifaceDef.signatures || ifaceDef.signatures.length === 0) {
          return true;
        }
      }
      if (directImplements(name, ifaceName, args)) {
        return true;
      }
      for (const candidate of this.interfaces.keys()) {
        if (candidate === ifaceName) continue;
        if (!interfaceExtends(candidate, ifaceName, new Set())) continue;
        if (directImplements(name, candidate, args)) {
          return true;
        }
      }
      return false;
    };
    const interfaceArgs = options.interfaceArgs;
    return check(typeName, interfaceName, interfaceArgs, new Set());
  };

  cls.prototype.toInterfaceValue = function toInterfaceValue(
    this: Interpreter,
    interfaceName: string,
    rawValue: RuntimeValue,
    interfaceArgs?: AST.TypeExpression[],
  ): RuntimeValue {
    if (!this.interfaces.has(interfaceName)) {
      throw new Error(`Unknown interface '${interfaceName}'`);
    }
    if (rawValue.kind === "interface_value") {
      if (rawValue.interfaceName === interfaceName) return rawValue;
      return this.toInterfaceValue(interfaceName, rawValue.value, interfaceArgs);
    }
    if (interfaceName === "Error" && rawValue.kind === "error") {
      return {
        kind: "interface_value",
        interfaceName,
        interfaceArgs,
        value: rawValue,
        methods: new Map(),
      };
    }
    if (interfaceName === "Iterator" && rawValue.kind === "iterator") {
      const methods = buildInterfaceMethodDictionary(this, interfaceName, interfaceArgs, rawValue, "Iterator");
      return {
        kind: "interface_value",
        interfaceName,
        interfaceArgs,
        value: rawValue,
        methods,
      };
    }
    const typeName = this.getTypeNameForValue(rawValue);
    const subjectTypeArgs = rawValue.kind === "struct_instance" ? rawValue.typeArguments : undefined;
    if (!typeName || !this.typeImplementsInterface(typeName, interfaceName, { subjectTypeArgs, interfaceArgs })) {
      throw new Error(`Type '${typeName ?? "<unknown>"}' does not implement interface '${interfaceName}'`);
    }
    let typeArguments: AST.TypeExpression[] | undefined;
    let typeArgMap: Map<string, AST.TypeExpression> | undefined;
    if (rawValue.kind === "struct_instance") {
      typeArguments = rawValue.typeArguments;
      typeArgMap = rawValue.typeArgMap;
    }
    const resolution = this.resolveInterfaceImplementation(typeName, interfaceName, {
      typeArgs: typeArguments,
      interfaceArgs,
      typeArgMap,
    });
    if (!resolution.ok) {
      if (resolution.error) {
        throw resolution.error;
      }
      throw new Error(`Type '${typeName ?? "<unknown>"}' does not implement interface '${interfaceName}'`);
    }
    const methods = buildInterfaceMethodDictionary(
      this,
      interfaceName,
      interfaceArgs,
      rawValue,
      typeName,
      typeArguments,
      typeArgMap,
    );
    return {
      kind: "interface_value",
      interfaceName,
      interfaceArgs,
      value: rawValue,
      typeArguments,
      typeArgMap,
      methods,
    };
  };
}
