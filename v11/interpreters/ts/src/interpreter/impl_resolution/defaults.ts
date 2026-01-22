import * as AST from "../../ast";
import { Environment } from "../environment";
import type { Interpreter } from "../index";
import type { RuntimeValue } from "../values";
import { collectSelfPatternPlaceholderNames } from "./helpers";

export function applyImplDefaultAugmentations(cls: typeof Interpreter): void {
  cls.prototype.buildSelfTypePatternBindings = function buildSelfTypePatternBindings(
    this: Interpreter,
    iface: AST.InterfaceDefinition,
    targetType: AST.TypeExpression,
  ): Map<string, AST.TypeExpression> {
    const bindings = new Map<string, AST.TypeExpression>();
    const pattern = iface.selfTypePattern;
    if (!pattern) return bindings;
    const interfaceGenericNames = new Set(
      (iface.genericParams ?? [])
        .map((param) => param?.name?.name)
        .filter((name): name is string => Boolean(name)),
    );
    const targetBase = targetType.type === "GenericTypeExpression" ? targetType.base : targetType;
    if (pattern.type === "SimpleTypeExpression") {
      const baseName = pattern.name.name;
      if (baseName && baseName !== "Self" && !interfaceGenericNames.has(baseName)) {
        bindings.set(baseName, this.cloneTypeExpression(targetBase));
        return bindings;
      }
    }
    if (pattern.type === "GenericTypeExpression" && pattern.base.type === "SimpleTypeExpression") {
      const baseName = pattern.base.name.name;
      if (baseName && baseName !== "Self" && !interfaceGenericNames.has(baseName)) {
        bindings.set(baseName, this.cloneTypeExpression(targetBase));
        return bindings;
      }
    }
    const placeholders = collectSelfPatternPlaceholderNames(this, pattern, interfaceGenericNames);
    if (placeholders.size === 0) return bindings;
    for (const name of placeholders) {
      bindings.set(name, this.cloneTypeExpression(targetBase));
    }
    return bindings;
  };

  cls.prototype.attachDefaultInterfaceMethods = function attachDefaultInterfaceMethods(
    this: Interpreter,
    imp: AST.ImplementationDefinition,
    funcs: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>,
  ): void {
    const interfaceName = imp.interfaceName.name;
    const iface = this.interfaces.get(interfaceName);
    if (!iface) return;
    const ifaceEnv = this.interfaceEnvs.get(interfaceName) ?? this.globals;
    const targetType = imp.targetType;
    const typeBindings = this.buildSelfTypePatternBindings(iface, targetType);
    for (const sig of iface.signatures) {
      if (!sig.defaultImpl) continue;
      const methodName = sig.name.name;
      if (funcs.has(methodName)) continue;
      const defaultFunc = this.createDefaultMethodFunction(sig, ifaceEnv, targetType, typeBindings);
      if (defaultFunc) funcs.set(methodName, defaultFunc);
    }
  };

  cls.prototype.createDefaultMethodFunction = function createDefaultMethodFunction(
    this: Interpreter,
    sig: AST.FunctionSignature,
    env: Environment,
    targetType: AST.TypeExpression,
    typeBindings?: Map<string, AST.TypeExpression>,
  ): Extract<RuntimeValue, { kind: "function" }> | null {
    if (!sig.defaultImpl) return null;
    const combinedBindings = new Map<string, AST.TypeExpression>();
    if (typeBindings) {
      for (const [name, expr] of typeBindings.entries()) {
        combinedBindings.set(name, expr);
      }
    }
    combinedBindings.set("Self", targetType);
    const substituteTypes = (expr: AST.TypeExpression | undefined): AST.TypeExpression | undefined => {
      if (!expr) return expr;
      switch (expr.type) {
        case "SimpleTypeExpression": {
          const replacement = combinedBindings.get(expr.name.name);
          return replacement ? this.cloneTypeExpression(replacement) : expr;
        }
        case "GenericTypeExpression":
          return {
            type: "GenericTypeExpression",
            base: substituteTypes(expr.base) ?? expr.base,
            arguments: (expr.arguments ?? []).map((arg) => substituteTypes(arg)),
          };
        case "FunctionTypeExpression":
          return {
            type: "FunctionTypeExpression",
            paramTypes: expr.paramTypes.map((param) => substituteTypes(param) ?? param),
            returnType: substituteTypes(expr.returnType) ?? expr.returnType,
          };
        case "NullableTypeExpression":
          return { type: "NullableTypeExpression", innerType: substituteTypes(expr.innerType) ?? expr.innerType };
        case "ResultTypeExpression":
          return { type: "ResultTypeExpression", innerType: substituteTypes(expr.innerType) ?? expr.innerType };
        case "UnionTypeExpression":
          return { type: "UnionTypeExpression", members: expr.members.map((member) => substituteTypes(member) ?? member) };
        case "WildcardTypeExpression":
        default:
          return expr;
      }
    };
    const params = sig.params.map(param => {
      const substitutedPattern = this.substituteSelfInPattern(param.name as AST.Pattern, targetType);
      const substitutedType = substituteTypes(param.paramType) ?? this.substituteSelfTypeExpression(param.paramType, targetType);
      if (substitutedPattern === param.name && substitutedType === param.paramType) return param;
      return { type: "FunctionParameter", name: substitutedPattern, paramType: substitutedType } as AST.FunctionParameter;
    });
    const returnType = substituteTypes(sig.returnType) ?? this.substituteSelfTypeExpression(sig.returnType, targetType) ?? sig.returnType;
    let closureEnv = env;
    if (combinedBindings.size > 0) {
      closureEnv = new Environment(env);
      for (const [name, expr] of combinedBindings.entries()) {
        const parsed = this.parseTypeExpression(expr);
        if (parsed) {
          try {
            closureEnv.define(name, { kind: "type_ref", typeName: parsed.name, typeArgs: parsed.typeArgs });
          } catch {}
        }
      }
    }
    const fnDef: AST.FunctionDefinition = {
      type: "FunctionDefinition",
      id: sig.name,
      params,
      returnType,
      genericParams: sig.genericParams,
      whereClause: sig.whereClause,
      body: sig.defaultImpl,
      isMethodShorthand: false,
      isPrivate: false,
    };
    const func: Extract<RuntimeValue, { kind: "function" }> = { kind: "function", node: fnDef, closureEnv };
    (func as any).methodResolutionPriority = -2;
    return func;
  };

  cls.prototype.substituteSelfTypeExpression = function substituteSelfTypeExpression(
    this: Interpreter,
    t: AST.TypeExpression | undefined,
    target: AST.TypeExpression,
  ): AST.TypeExpression | undefined {
    if (!t) return t;
    switch (t.type) {
      case "SimpleTypeExpression":
        if (t.name.name === "Self") return this.cloneTypeExpression(target);
        return t;
      case "GenericTypeExpression": {
        const base = this.substituteSelfTypeExpression(t.base, target) ?? t.base;
        const args = t.arguments?.map(arg => this.substituteSelfTypeExpression(arg, target) ?? arg) ?? [];
        if (base === t.base && args.every((arg, idx) => arg === (t.arguments ?? [])[idx])) return t;
        return { type: "GenericTypeExpression", base, arguments: args };
      }
      case "FunctionTypeExpression": {
        const paramTypes = t.paramTypes.map(pt => this.substituteSelfTypeExpression(pt, target) ?? pt);
        const returnType = this.substituteSelfTypeExpression(t.returnType, target) ?? t.returnType;
        if (paramTypes.every((pt, idx) => pt === t.paramTypes[idx]) && returnType === t.returnType) return t;
        return { type: "FunctionTypeExpression", paramTypes, returnType };
      }
      case "NullableTypeExpression": {
        const inner = this.substituteSelfTypeExpression(t.innerType, target) ?? t.innerType;
        if (inner === t.innerType) return t;
        return { type: "NullableTypeExpression", innerType: inner };
      }
      case "ResultTypeExpression": {
        const inner = this.substituteSelfTypeExpression(t.innerType, target) ?? t.innerType;
        if (inner === t.innerType) return t;
        return { type: "ResultTypeExpression", innerType: inner };
      }
      case "UnionTypeExpression": {
        let changed = false;
        const members = t.members.map(member => {
          const next = this.substituteSelfTypeExpression(member, target) ?? member;
          if (next !== member) changed = true;
          return next;
        });
        if (!changed) return t;
        return { type: "UnionTypeExpression", members };
      }
      case "WildcardTypeExpression":
      default:
        return t;
    }
  };

  cls.prototype.substituteSelfInPattern = function substituteSelfInPattern(
    this: Interpreter,
    pattern: AST.Pattern,
    target: AST.TypeExpression,
  ): AST.Pattern {
    if ((pattern as any).type === "TypedPattern") {
      const tp = pattern as AST.TypedPattern;
      const inner = this.substituteSelfInPattern(tp.pattern, target);
      const typeAnnotation = this.substituteSelfTypeExpression(tp.typeAnnotation, target) ?? tp.typeAnnotation;
      if (inner === tp.pattern && typeAnnotation === tp.typeAnnotation) return tp;
      return { type: "TypedPattern", pattern: inner, typeAnnotation };
    }
    if (pattern.type === "StructPattern") {
      let changed = false;
      const fields = pattern.fields.map(field => {
        const newPattern = this.substituteSelfInPattern(field.pattern, target);
        if (newPattern !== field.pattern) {
          changed = true;
          return { ...field, pattern: newPattern };
        }
        return field;
      });
      let structType = pattern.structType;
      if (structType && structType.name === "Self" && target.type === "SimpleTypeExpression") {
        structType = AST.identifier(target.name.name);
        changed = true;
      }
      if (!changed) return pattern;
      return { ...pattern, fields, structType };
    }
    if (pattern.type === "ArrayPattern") {
      let changed = false;
      const elements = pattern.elements.map(el => {
        if (!el) return el;
        const newEl = this.substituteSelfInPattern(el, target);
        if (newEl !== el) changed = true;
        return newEl ?? el;
      });
      const restPattern = pattern.restPattern
        ? (this.substituteSelfInPattern(pattern.restPattern, target) as AST.Identifier | AST.WildcardPattern)
        : undefined;
      if (restPattern !== pattern.restPattern) changed = true;
      if (!changed) return pattern;
      return { ...pattern, elements, restPattern };
    }
    return pattern;
  };
}
