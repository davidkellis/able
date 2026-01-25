import * as AST from "../../ast";
import { Environment } from "../environment";
import type { Interpreter } from "../index";
import type { ConstraintSpec, RuntimeValue } from "../values";
import { isKnownTypeName, primitiveImplementsInterfaceMethod } from "./helpers";

export function applyImplConstraintAugmentations(cls: typeof Interpreter): void {
  cls.prototype.enforceGenericConstraintsIfAny = function enforceGenericConstraintsIfAny(
    this: Interpreter,
    funcNode: AST.FunctionDefinition | AST.LambdaExpression,
    call: AST.FunctionCall,
  ): void {
    const generics = (funcNode as any).genericParams as AST.GenericParameter[] | undefined;
    const where = (funcNode as any).whereClause as AST.WhereClauseConstraint[] | undefined;
    const typeArgs = call.typeArguments ?? [];
    const genericCount = generics ? generics.length : 0;
    if (genericCount > 0 && typeArgs.length !== genericCount) {
      const name = (funcNode as any).id?.name ?? "(lambda)";
      throw new Error(`Type arguments count mismatch calling ${name}: expected ${genericCount}, got ${typeArgs.length}`);
    }
    const constraints = this.collectConstraintSpecs(generics, where);
    if (constraints.length === 0) return;
    const name = (funcNode as any).id?.name ?? "(lambda)";
    const typeArgMap = this.mapTypeArguments(generics, typeArgs, `calling ${name}`);
    this.enforceConstraintSpecs(constraints, typeArgMap, `function ${name}`);
  };

  cls.prototype.collectConstraintSpecs = function collectConstraintSpecs(
    this: Interpreter,
    generics?: AST.GenericParameter[],
    where?: AST.WhereClauseConstraint[],
  ): ConstraintSpec[] {
    const all: ConstraintSpec[] = [];
    if (generics) {
      for (const gp of generics) {
        if (!gp.constraints) continue;
        for (const c of gp.constraints) {
          all.push({ subjectExpr: AST.simpleTypeExpression(gp.name), ifaceType: c.interfaceType });
        }
      }
    }
    if (where) {
      for (const clause of where) {
        for (const c of clause.constraints) {
          all.push({ subjectExpr: clause.typeParam, ifaceType: c.interfaceType });
        }
      }
    }
    return all;
  };

  cls.prototype.mapTypeArguments = function mapTypeArguments(
    this: Interpreter,
    generics: AST.GenericParameter[] | undefined,
    provided: AST.TypeExpression[] | undefined,
    context: string,
  ): Map<string, AST.TypeExpression> {
    const map = new Map<string, AST.TypeExpression>();
    if (!generics || generics.length === 0) return map;
    const actual = provided ?? [];
    if (actual.length !== generics.length) {
      throw new Error(`Type arguments count mismatch ${context}: expected ${generics.length}, got ${actual.length}`);
    }
    for (let i = 0; i < generics.length; i++) {
      const gp = generics[i]!;
      const ta = actual[i];
      if (!ta) {
        throw new Error(`Missing type argument for '${gp.name.name}' required by ${context}`);
      }
      map.set(gp.name.name, ta);
    }
    return map;
  };

  cls.prototype.enforceConstraintSpecs = function enforceConstraintSpecs(
    this: Interpreter,
    constraints: ConstraintSpec[],
    typeArgMap: Map<string, AST.TypeExpression>,
    context: string,
  ): void {
    const substituteTypeParams = (expr: AST.TypeExpression): AST.TypeExpression => {
      switch (expr.type) {
        case "SimpleTypeExpression": {
          const name = expr.name.name;
          const replacement = typeArgMap.get(name);
          return replacement ? this.cloneTypeExpression(replacement) : expr;
        }
        case "GenericTypeExpression":
          return {
            type: "GenericTypeExpression",
            base: substituteTypeParams(expr.base),
            arguments: (expr.arguments ?? []).map(arg => substituteTypeParams(arg)),
          };
        case "FunctionTypeExpression":
          return {
            type: "FunctionTypeExpression",
            paramTypes: expr.paramTypes.map(pt => substituteTypeParams(pt)),
            returnType: substituteTypeParams(expr.returnType),
          };
        case "NullableTypeExpression":
          return { type: "NullableTypeExpression", innerType: substituteTypeParams(expr.innerType) };
        case "ResultTypeExpression":
          return { type: "ResultTypeExpression", innerType: substituteTypeParams(expr.innerType) };
        case "UnionTypeExpression":
          return { type: "UnionTypeExpression", members: expr.members.map(member => substituteTypeParams(member)) };
        case "WildcardTypeExpression":
        default:
          return expr;
      }
    };
    const hasUnknownTypeNames = (expr: AST.TypeExpression): boolean => {
      switch (expr.type) {
        case "SimpleTypeExpression":
          return !isKnownTypeName(this, expr.name.name);
        case "GenericTypeExpression":
          if (hasUnknownTypeNames(expr.base)) return true;
          return (expr.arguments ?? []).some(arg => hasUnknownTypeNames(arg));
        case "FunctionTypeExpression":
          if (hasUnknownTypeNames(expr.returnType)) return true;
          return expr.paramTypes.some(pt => hasUnknownTypeNames(pt));
        case "NullableTypeExpression":
        case "ResultTypeExpression":
          return hasUnknownTypeNames(expr.innerType);
        case "UnionTypeExpression":
          return expr.members.some(member => hasUnknownTypeNames(member));
        case "WildcardTypeExpression":
        default:
          return true;
      }
    };
    for (const c of constraints) {
      const subject = substituteTypeParams(c.subjectExpr);
      if (hasUnknownTypeNames(subject)) {
        continue;
      }
      const typeInfo = this.parseTypeExpression(subject);
      if (!typeInfo) continue;
      this.ensureTypeSatisfiesInterface(typeInfo, c.ifaceType, this.typeExpressionToString(subject), new Set());
    }
  };

  cls.prototype.ensureTypeSatisfiesInterface = function ensureTypeSatisfiesInterface(
    this: Interpreter,
    typeInfo: { name: string; typeArgs: AST.TypeExpression[] },
    interfaceType: AST.TypeExpression,
    context: string,
    visited: Set<string>,
  ): void {
    const ifaceInfo = this.parseTypeExpression(interfaceType);
    if (!ifaceInfo) return;
    if (visited.has(ifaceInfo.name)) return;
    visited.add(ifaceInfo.name);
    const iface = this.interfaces.get(ifaceInfo.name);
    if (!iface) throw new Error(`Unknown interface '${ifaceInfo.name}' in constraint on '${context}'`);
    const interfaceExtends = (candidate: string, target: string, seen: Set<string>): boolean => {
      if (!candidate || !target) return false;
      if (candidate === target) return true;
      if (seen.has(candidate)) return false;
      seen.add(candidate);
      const def = this.interfaces.get(candidate);
      if (!def?.baseInterfaces || def.baseInterfaces.length === 0) {
        return false;
      }
      for (const base of def.baseInterfaces) {
        const info = this.parseTypeExpression(base);
        if (!info) continue;
        if (info.name === target) return true;
        if (interfaceExtends(info.name, target, seen)) return true;
      }
      return false;
    };
    const interfaceArgs = ifaceInfo.typeArgs.length > 0 ? ifaceInfo.typeArgs : undefined;
    const hasMethodViaDerivedInterface = (methodName: string): boolean => {
      for (const candidate of this.interfaces.keys()) {
        if (candidate === ifaceInfo.name) continue;
        if (!interfaceExtends(candidate, ifaceInfo.name, new Set())) continue;
        const method = this.findMethod(typeInfo.name, methodName, {
          typeArgs: typeInfo.typeArgs,
          interfaceName: candidate,
        });
        if (method) return true;
      }
      return false;
    };
    for (const base of iface.baseInterfaces ?? []) {
      this.ensureTypeSatisfiesInterface(typeInfo, base, context, visited);
    }
    for (const sig of iface.signatures) {
      const methodName = sig.name.name;
      if (primitiveImplementsInterfaceMethod(typeInfo.name, ifaceInfo.name, methodName)) {
        continue;
      }
      const method = this.findMethod(typeInfo.name, methodName, {
        typeArgs: typeInfo.typeArgs,
        interfaceArgs,
        interfaceName: ifaceInfo.name,
      });
      if (!method && !hasMethodViaDerivedInterface(methodName)) {
        throw new Error(`Type '${typeInfo.name}' does not satisfy interface '${ifaceInfo.name}': missing method '${methodName}'`);
      }
    }
  };

  cls.prototype.inferTypeArgumentsFromCall = function inferTypeArgumentsFromCall(
    this: Interpreter,
    funcNode: AST.FunctionDefinition | AST.LambdaExpression,
    call: AST.FunctionCall,
    args: RuntimeValue[],
  ): void {
    const generics = (funcNode as any).genericParams as AST.GenericParameter[] | undefined;
    if (!generics || generics.length === 0) return;
    if (call.typeArguments && call.typeArguments.length > 0) {
      if (call.typeArguments.length !== generics.length) {
        const name = (funcNode as any).id?.name ?? "(lambda)";
        throw new Error(`Type arguments count mismatch calling ${name}: expected ${generics.length}, got ${call.typeArguments.length}`);
      }
      return;
    }
    const bindings = new Map<string, AST.TypeExpression>();
    const genericNames = new Set(generics.map(g => g.name.name));
    const params = (funcNode as any).params as AST.FunctionParameter[] | undefined;
    let bindArgs = args;
    if ((funcNode as any).isMethodShorthand && bindArgs.length > 0) {
      bindArgs = bindArgs.slice(1);
    }
    if (params && params.length > 0 && bindArgs.length > 0) {
      const count = Math.min(params.length, bindArgs.length);
      for (let i = 0; i < count; i++) {
        const param = params[i];
        const actual = bindArgs[i];
        if (!param || !param.paramType || !actual) continue;
        const inferred = this.typeExpressionFromValue(actual);
        if (!inferred) continue;
        this.matchTypeExpressionTemplate(param.paramType, inferred, genericNames, bindings);
      }
    }
    call.typeArguments = generics.map(gp => {
      const binding = bindings.get(gp.name.name);
      return binding ?? AST.wildcardTypeExpression();
    });
  };

  cls.prototype.bindTypeArgumentsIfAny = function bindTypeArgumentsIfAny(
    this: Interpreter,
    funcNode: AST.FunctionDefinition | AST.LambdaExpression,
    call: AST.FunctionCall,
    env: Environment,
  ): void {
    const generics = (funcNode as any).genericParams as AST.GenericParameter[] | undefined;
    if (!generics || generics.length === 0) return;
    const args = call.typeArguments ?? [];
    const count = Math.min(generics.length, args.length);
    const stringifyTypeExpr = (expr: AST.TypeExpression, depth = 0): string => {
      if (depth > 8) return "<type>";
      switch (expr.type) {
        case "SimpleTypeExpression":
          return expr.name.name;
        case "GenericTypeExpression":
          return `${stringifyTypeExpr(expr.base, depth + 1)}<${(expr.arguments ?? []).map((arg) => (arg ? stringifyTypeExpr(arg, depth + 1) : "_")).join(",")}>`;
        case "NullableTypeExpression":
          return `${stringifyTypeExpr(expr.innerType, depth + 1)}?`;
        case "ResultTypeExpression":
          return `Result<${stringifyTypeExpr(expr.innerType, depth + 1)}>`;
        case "UnionTypeExpression":
          return (expr.members ?? []).map((member) => stringifyTypeExpr(member, depth + 1)).join(" | ");
        case "FunctionTypeExpression":
          return "fn(...)";
        case "WildcardTypeExpression":
          return "_";
        default:
          return "<type>";
      }
    };
    for (let i = 0; i < count; i++) {
      const gp = generics[i]!;
      const ta = args[i]!;
      const name = `${gp.name.name}_type`;
      const s = stringifyTypeExpr(ta);
      try { env.define(name, { kind: "String", value: s }); } catch {}
      const parsed = this.parseTypeExpression(ta);
      if (parsed) {
        try {
          env.define(gp.name.name, { kind: "type_ref", typeName: parsed.name, typeArgs: parsed.typeArgs });
        } catch {}
      }
    }
  };

  cls.prototype.collectInterfaceConstraintExpressions = function collectInterfaceConstraintExpressions(
    this: Interpreter,
    typeExpr: AST.TypeExpression,
    memo: Set<string> = new Set(),
  ): AST.TypeExpression[] {
    const key = this.typeExpressionToString(typeExpr);
    if (memo.has(key)) return [];
    memo.add(key);
    const expressions: AST.TypeExpression[] = [typeExpr];
    if (typeExpr.type === "SimpleTypeExpression") {
      const iface = this.interfaces.get(typeExpr.name.name);
      if (iface && iface.baseInterfaces) {
        for (const base of iface.baseInterfaces) {
          const cloned = this.cloneTypeExpression(base);
          expressions.push(...this.collectInterfaceConstraintExpressions(cloned, memo));
        }
      }
    }
    return expressions;
  };
}
