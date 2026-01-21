import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { ConstraintSpec, ImplMethodEntry, RuntimeValue } from "./values";

const INTEGER_TYPES = new Set([
  "i8", "i16", "i32", "i64", "i128",
  "u8", "u16", "u32", "u64", "u128",
]);

const FLOAT_TYPES = new Set(["f32", "f64"]);

function isPrimitiveTypeName(name: string): boolean {
  if (name === "bool" || name === "String" || name === "IoHandle" || name === "ProcHandle" || name === "char" || name === "nil" || name === "void") {
    return true;
  }
  return INTEGER_TYPES.has(name) || FLOAT_TYPES.has(name);
}

function isKnownTypeName(interp: Interpreter, name: string): boolean {
  if (!name) return false;
  if (name === "Self" || name === "_") return false;
  if (isPrimitiveTypeName(name)) return true;
  if (interp.structs.has(name) || interp.interfaces.has(name) || interp.unions.has(name) || interp.typeAliases.has(name)) {
    return true;
  }
  return false;
}

function isSelfPatternPlaceholderName(
  interp: Interpreter,
  name: string,
  interfaceGenericNames: Set<string>,
): boolean {
  if (!name || name === "Self") {
    return false;
  }
  if (interfaceGenericNames.has(name)) {
    return false;
  }
  return !isKnownTypeName(interp, name);
}

function collectSelfPatternPlaceholderNames(
  interp: Interpreter,
  expr: AST.TypeExpression | undefined,
  interfaceGenericNames: Set<string>,
  out: Set<string> = new Set(),
): Set<string> {
  if (!expr) return out;
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = expr.name.name;
      if (isSelfPatternPlaceholderName(interp, name, interfaceGenericNames)) {
        out.add(name);
      }
      return out;
    }
    case "GenericTypeExpression": {
      collectSelfPatternPlaceholderNames(interp, expr.base, interfaceGenericNames, out);
      (expr.arguments ?? []).forEach((arg) =>
        collectSelfPatternPlaceholderNames(interp, arg, interfaceGenericNames, out),
      );
      return out;
    }
    case "FunctionTypeExpression": {
      expr.paramTypes.forEach((param) =>
        collectSelfPatternPlaceholderNames(interp, param, interfaceGenericNames, out),
      );
      collectSelfPatternPlaceholderNames(interp, expr.returnType, interfaceGenericNames, out);
      return out;
    }
    case "NullableTypeExpression":
    case "ResultTypeExpression":
      return collectSelfPatternPlaceholderNames(interp, expr.innerType, interfaceGenericNames, out);
    case "UnionTypeExpression":
      expr.members.forEach((member) =>
        collectSelfPatternPlaceholderNames(interp, member, interfaceGenericNames, out),
      );
      return out;
    default:
      return out;
  }
}

function primitiveImplementsInterfaceMethod(typeName: string, ifaceName: string, methodName: string): boolean {
  if (!typeName || typeName === "nil" || typeName === "void") {
    return false;
  }
  if (!isPrimitiveTypeName(typeName)) {
    return false;
  }
  switch (ifaceName) {
    case "Hash":
      return methodName === "hash";
    case "Eq":
      return methodName === "eq" || methodName === "ne";
    default:
      return false;
  }
}

declare module "./index" {
  interface Interpreter {
    enforceGenericConstraintsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall): void;
    collectConstraintSpecs(generics?: AST.GenericParameter[], where?: AST.WhereClauseConstraint[]): ConstraintSpec[];
    mapTypeArguments(generics: AST.GenericParameter[] | undefined, provided: AST.TypeExpression[] | undefined, context: string): Map<string, AST.TypeExpression>;
    enforceConstraintSpecs(constraints: ConstraintSpec[], typeArgMap: Map<string, AST.TypeExpression>, context: string): void;
    ensureTypeSatisfiesInterface(typeInfo: { name: string; typeArgs: AST.TypeExpression[] }, interfaceType: AST.TypeExpression, context: string, visited: Set<string>): void;
    inferTypeArgumentsFromCall(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, args: RuntimeValue[]): void;
    bindTypeArgumentsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, env: Environment): void;
    collectInterfaceConstraintExpressions(typeExpr: AST.TypeExpression, memo?: Set<string>): AST.TypeExpression[];
    findMethod(
      typeName: string,
      methodName: string,
      opts?: {
        typeArgs?: AST.TypeExpression[];
        interfaceArgs?: AST.TypeExpression[];
        typeArgMap?: Map<string, AST.TypeExpression>;
        interfaceName?: string;
        includeInherent?: boolean;
      },
    ): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null;
    resolveInterfaceImplementation(
      typeName: string,
      interfaceName: string,
      opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> },
    ): { ok: boolean; error?: Error };
    compareMethodMatches(
      a: { entry: ImplMethodEntry; bindings: Map<string, AST.TypeExpression>; constraints: ConstraintSpec[]; isConcreteTarget: boolean; score: number; method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }> },
      b: { entry: ImplMethodEntry; bindings: Map<string, AST.TypeExpression>; constraints: ConstraintSpec[]; isConcreteTarget: boolean; score: number; method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }> },
    ): number;
    buildConstraintKeySet(constraints: ConstraintSpec[]): Set<string>;
    isConstraintSuperset(a: Set<string>, b: Set<string>): boolean;
    isProperSubset(a: string[], b: string[]): boolean;
    matchImplEntry(entry: ImplMethodEntry, opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; subjectType?: AST.TypeExpression }): Map<string, AST.TypeExpression> | null;
    matchTypeExpressionTemplate(template: AST.TypeExpression, actual: AST.TypeExpression, genericNames: Set<string>, bindings: Map<string, AST.TypeExpression>): boolean;
    expandImplementationTargetVariants(target: AST.TypeExpression): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }>;
    measureTemplateSpecificity(t: AST.TypeExpression, genericNames: Set<string>): number;
    attachDefaultInterfaceMethods(
      imp: AST.ImplementationDefinition,
      funcs: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>,
    ): void;
    buildSelfTypePatternBindings(
      iface: AST.InterfaceDefinition,
      targetType: AST.TypeExpression,
    ): Map<string, AST.TypeExpression>;
    createDefaultMethodFunction(
      sig: AST.FunctionSignature,
      env: Environment,
      targetType: AST.TypeExpression,
      typeBindings?: Map<string, AST.TypeExpression>,
    ): Extract<RuntimeValue, { kind: "function" }> | null;
    substituteSelfTypeExpression(t: AST.TypeExpression | undefined, target: AST.TypeExpression): AST.TypeExpression | undefined;
    substituteSelfInPattern(pattern: AST.Pattern, target: AST.TypeExpression): AST.Pattern;
  }
}

export function applyImplResolutionAugmentations(cls: typeof Interpreter): void {
  cls.prototype.enforceGenericConstraintsIfAny = function enforceGenericConstraintsIfAny(this: Interpreter, funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall): void {
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

  cls.prototype.collectConstraintSpecs = function collectConstraintSpecs(this: Interpreter, generics?: AST.GenericParameter[], where?: AST.WhereClauseConstraint[]): ConstraintSpec[] {
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

  cls.prototype.mapTypeArguments = function mapTypeArguments(this: Interpreter, generics: AST.GenericParameter[] | undefined, provided: AST.TypeExpression[] | undefined, context: string): Map<string, AST.TypeExpression> {
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

  cls.prototype.enforceConstraintSpecs = function enforceConstraintSpecs(this: Interpreter, constraints: ConstraintSpec[], typeArgMap: Map<string, AST.TypeExpression>, context: string): void {
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
  const isKnownTypeName = (name: string): boolean => {
    if (!name) return false;
    if (name === "Self") return false;
    if (name === "_") return false;
    if (isPrimitiveTypeName(name)) return true;
    if (this.structs.has(name) || this.interfaces.has(name) || this.unions.has(name) || this.typeAliases.has(name)) {
      return true;
    }
    return false;
  };
  const hasUnknownTypeNames = (expr: AST.TypeExpression): boolean => {
    switch (expr.type) {
      case "SimpleTypeExpression":
        return !isKnownTypeName(expr.name.name);
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

  cls.prototype.ensureTypeSatisfiesInterface = function ensureTypeSatisfiesInterface(this: Interpreter, typeInfo: { name: string; typeArgs: AST.TypeExpression[] }, interfaceType: AST.TypeExpression, context: string, visited: Set<string>): void {
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

  cls.prototype.inferTypeArgumentsFromCall = function inferTypeArgumentsFromCall(this: Interpreter, funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, args: RuntimeValue[]): void {
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

  cls.prototype.bindTypeArgumentsIfAny = function bindTypeArgumentsIfAny(this: Interpreter, funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, env: Environment): void {
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

  cls.prototype.collectInterfaceConstraintExpressions = function collectInterfaceConstraintExpressions(this: Interpreter, typeExpr: AST.TypeExpression, memo: Set<string> = new Set()): AST.TypeExpression[] {
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

  let traceMethodReported = false;
  cls.prototype.findMethod = function findMethod(
    this: Interpreter,
    typeName: string,
    methodName: string,
    opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; interfaceName?: string; includeInherent?: boolean },
  ): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
  const includeInherent = opts?.includeInherent !== false;
  if (includeInherent) {
    const inherent = this.inherentMethods.get(typeName);
    if (inherent && inherent.has(methodName)) return inherent.get(methodName)!;
  }
  let interfaceNames: Set<string> | null = null;
  if (opts?.interfaceName) {
    const ifaceDef = this.interfaces.get(opts.interfaceName);
    if (ifaceDef?.baseInterfaces && ifaceDef.baseInterfaces.length > 0) {
      const expanded = this.collectInterfaceConstraintExpressions(AST.simpleTypeExpression(opts.interfaceName));
      interfaceNames = new Set<string>();
      for (const expr of expanded) {
        const info = interfaceInfoFromTypeExpression(expr);
        if (info) {
          interfaceNames.add(info.name);
        }
      }
    }
  }
  const subjectType = typeExpressionFromInfo(typeName, opts?.typeArgs);
  const entries = [
    ...(this.implMethods.get(typeName) ?? []),
    ...this.genericImplMethods,
  ];
  if (
    process.env.ABLE_TRACE_ERRORS &&
    !traceMethodReported &&
    methodName === "matches" &&
    typeName === "ContainMatcher"
  ) {
    const entryTypes = entries.map((entry) => this.typeExpressionToString(entry.def.targetType));
    console.error(`[trace] matches candidates for ${typeName}: ${entryTypes.join(", ") || "<none>"}`);
  }
  let constraintError: Error | null = null;
  let matches: Array<{
    method: Extract<RuntimeValue, { kind: "function" }>;
    score: number;
    entry: ImplMethodEntry;
    constraints: ConstraintSpec[];
    isConcreteTarget: boolean;
  }> = [];
  for (const entry of entries) {
    if (interfaceNames) {
      if (!interfaceNames.has(entry.def.interfaceName.name)) continue;
    } else if (opts?.interfaceName && entry.def.interfaceName.name !== opts.interfaceName) {
      continue;
    }
    const interfaceArgs = opts?.interfaceArgs && opts.interfaceArgs.length > 0 ? opts.interfaceArgs : undefined;
    const bindings = this.matchImplEntry(entry, {
      subjectType,
      typeArgs: opts?.typeArgs,
      typeArgMap: opts?.typeArgMap,
      interfaceArgs,
    });
    if (!bindings) continue;
    const method = entry.methods.get(methodName);
    if (!method) continue;
    const constraints = this.collectConstraintSpecs(entry.genericParams, entry.whereClause);
    if (constraints.length > 0) {
      try {
        this.enforceConstraintSpecs(constraints, bindings, `impl ${entry.def.interfaceName.name} for ${typeName}`);
      } catch (err) {
        if (!constraintError && err instanceof Error) constraintError = err;
        continue;
      }
    }
    const genericNames = collectImplGenericNames(this, entry);
    const score = this.measureTemplateSpecificity(entry.def.targetType, genericNames);
    const isConcreteTarget = !typeExpressionUsesGenerics(entry.def.targetType, genericNames);
    matches.push({ method, score, entry, constraints, isConcreteTarget });
  }
  if (opts?.interfaceName) {
    const directMatches = matches.filter((match) => match.entry.def.interfaceName.name === opts.interfaceName);
    if (directMatches.length > 0) {
      matches = directMatches;
    }
  }
  if (matches.length === 0) {
    if (constraintError) throw constraintError;
    return null;
  }
  const [firstMatch, ...remainingMatches] = matches;
  let best = firstMatch!;
  let contenders: typeof matches = [best];
  for (const candidate of remainingMatches) {
    const cmp = this.compareMethodMatches(candidate, best);
    if (cmp > 0) {
      best = candidate;
      contenders = [candidate];
      continue;
    }
    if (cmp === 0) {
      const reverse = this.compareMethodMatches(best, candidate);
      if (reverse < 0) {
        best = candidate;
        contenders = [candidate];
      } else if (reverse === 0) {
        contenders.push(candidate);
      }
    }
  }
  if (contenders.length > 1) {
    const ifaceName = contenders[0].entry.def.interfaceName.name || methodName;
    const detail = Array.from(new Set(contenders.map(c => `impl ${c.entry.def.interfaceName.name} for ${this.typeExpressionToString(c.entry.def.targetType)}`))).join(", ");
    throw new Error(`ambiguous implementations of ${ifaceName} for ${typeName}: ${detail}`);
  }
  if (
    process.env.ABLE_TRACE_ERRORS &&
    !traceMethodReported &&
    methodName === "matches" &&
    typeName === "ContainMatcher"
  ) {
    traceMethodReported = true;
    const entry = best.entry;
    console.error(`[trace] matches resolved to impl ${entry.def.interfaceName.name} for ${this.typeExpressionToString(entry.def.targetType)}`);
  }
  return best.method;
};

  cls.prototype.resolveInterfaceImplementation = function resolveInterfaceImplementation(
    this: Interpreter,
    typeName: string,
    interfaceName: string,
    opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> },
  ): { ok: boolean; error?: Error } {
  if (interfaceName === "Error" && typeName === "Error") {
    return { ok: true };
  }
  const ifaceDef = this.interfaces.get(interfaceName);
  if (ifaceDef?.baseInterfaces && ifaceDef.baseInterfaces.length > 0) {
    for (const base of ifaceDef.baseInterfaces) {
      const info = interfaceInfoFromTypeExpression(base);
      if (!info) continue;
      const baseInterfaceArgs = info.args && info.args.length > 0 ? info.args : undefined;
      const baseResult = this.resolveInterfaceImplementation(typeName, info.name, {
        typeArgs: opts?.typeArgs,
        interfaceArgs: baseInterfaceArgs,
        typeArgMap: opts?.typeArgMap,
      });
      if (!baseResult.ok) {
        return baseResult;
      }
    }
    if (!ifaceDef.signatures || ifaceDef.signatures.length === 0) {
      return { ok: true };
    }
  }
  const subjectType = typeExpressionFromInfo(typeName, opts?.typeArgs);
  const entries = [
    ...(this.implMethods.get(typeName) ?? []),
    ...this.genericImplMethods,
  ];
  if (entries.length === 0) return { ok: false };
  const matches: Array<{
    entry: ImplMethodEntry;
    constraints: ConstraintSpec[];
    score: number;
    isConcreteTarget: boolean;
  }> = [];
  let constraintError: Error | undefined;
  for (const entry of entries) {
    if (entry.def.interfaceName.name !== interfaceName) continue;
    const interfaceArgs = opts?.interfaceArgs && opts.interfaceArgs.length > 0 ? opts.interfaceArgs : undefined;
    const bindings = this.matchImplEntry(entry, {
      subjectType,
      typeArgs: opts?.typeArgs,
      interfaceArgs,
      typeArgMap: opts?.typeArgMap,
    });
    if (!bindings) continue;
    const constraints = this.collectConstraintSpecs(entry.genericParams, entry.whereClause);
    if (constraints.length > 0) {
      try {
        this.enforceConstraintSpecs(constraints, bindings, `impl ${entry.def.interfaceName.name} for ${typeName}`);
      } catch (err) {
        if (!constraintError && err instanceof Error) constraintError = err;
        continue;
      }
    }
    const genericNames = collectImplGenericNames(this, entry);
    matches.push({
      entry,
      constraints,
      score: this.measureTemplateSpecificity(entry.def.targetType, genericNames),
      isConcreteTarget: !typeExpressionUsesGenerics(entry.def.targetType, genericNames),
    });
  }
  if (matches.length === 0) {
    if (constraintError) {
      return { ok: false, error: constraintError };
    }
    const ifaceDef = this.interfaces.get(interfaceName);
    if (ifaceDef && Array.isArray(ifaceDef.signatures) && ifaceDef.signatures.length > 0) {
      const missing = ifaceDef.signatures[0]?.id?.name ?? "<unknown>";
      return { ok: false, error: new Error(`Type '${typeName}' does not satisfy interface '${interfaceName}': missing method '${missing}'`) };
    }
    return { ok: false };
  }
  let best = matches[0]!;
  let contenders: typeof matches = [best];
  for (const candidate of matches.slice(1)) {
    const cmp = this.compareMethodMatches(candidate, best);
    if (cmp > 0) {
      best = candidate;
      contenders = [candidate];
      continue;
    }
    if (cmp === 0) {
      const reverse = this.compareMethodMatches(best, candidate);
      if (reverse < 0) {
        best = candidate;
        contenders = [candidate];
      } else if (reverse === 0) {
        contenders.push(candidate);
      }
    }
  }
  if (contenders.length > 1) {
    const detail = Array.from(new Set(contenders.map(c => `impl ${c.entry.def.interfaceName.name} for ${this.typeExpressionToString(c.entry.def.targetType)}`))).join(", ");
    return { ok: false, error: new Error(`ambiguous implementations of ${interfaceName} for ${typeName}: ${detail}`) };
  }
  return { ok: true };
};

  cls.prototype.compareMethodMatches = function compareMethodMatches(this: Interpreter, a: { method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }>; score: number; entry: ImplMethodEntry; constraints: ConstraintSpec[]; isConcreteTarget: boolean }, b: { method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }>; score: number; entry: ImplMethodEntry; constraints: ConstraintSpec[]; isConcreteTarget: boolean }): number {
  if (a.isConcreteTarget && !b.isConcreteTarget) return 1;
  if (b.isConcreteTarget && !a.isConcreteTarget) return -1;
  const aConstraints = this.buildConstraintKeySet(a.constraints);
  const bConstraints = this.buildConstraintKeySet(b.constraints);
  if (this.isConstraintSuperset(aConstraints, bConstraints)) return 1;
  if (this.isConstraintSuperset(bConstraints, aConstraints)) return -1;
  const aUnion = a.entry.unionVariantSignatures;
  const bUnion = b.entry.unionVariantSignatures;
  const aUnionSize = aUnion?.length ?? 0;
  const bUnionSize = bUnion?.length ?? 0;
  if (aUnionSize !== bUnionSize) {
    if (aUnionSize === 0) return 1;
    if (bUnionSize === 0) return -1;
  }
  if (aUnion && bUnion) {
    if (this.isProperSubset(aUnion, bUnion)) return 1;
    if (this.isProperSubset(bUnion, aUnion)) return -1;
    if (aUnion.length !== bUnion.length) {
      return aUnion.length < bUnion.length ? 1 : -1;
    }
  }
  if (a.score > b.score) return 1;
  if (a.score < b.score) return -1;
  const aPriority = typeof (a.method as any)?.methodResolutionPriority === "number"
    ? (a.method as any).methodResolutionPriority
    : 0;
  const bPriority = typeof (b.method as any)?.methodResolutionPriority === "number"
    ? (b.method as any).methodResolutionPriority
    : 0;
  if (aPriority > bPriority) return 1;
  if (aPriority < bPriority) return -1;
  return 0;
};

  cls.prototype.buildConstraintKeySet = function buildConstraintKeySet(this: Interpreter, constraints: ConstraintSpec[]): Set<string> {
  const set = new Set<string>();
  for (const c of constraints) {
    const expanded = this.collectInterfaceConstraintExpressions(c.ifaceType);
    for (const expr of expanded) {
      set.add(`${this.typeExpressionToString(c.subjectExpr)}->${this.typeExpressionToString(expr)}`);
    }
  }
  return set;
};

  cls.prototype.isConstraintSuperset = function isConstraintSuperset(this: Interpreter, a: Set<string>, b: Set<string>): boolean {
  if (a.size <= b.size) return false;
  for (const key of b) {
    if (!a.has(key)) return false;
  }
  return true;
};

  cls.prototype.isProperSubset = function isProperSubset(this: Interpreter, a: string[], b: string[]): boolean {
  const aSet = new Set(a);
  const bSet = new Set(b);
  if (aSet.size >= bSet.size) return false;
  for (const val of aSet) {
    if (!bSet.has(val)) return false;
  }
  return true;
  };

  cls.prototype.matchImplEntry = function matchImplEntry(
    this: Interpreter,
    entry: ImplMethodEntry,
    opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; subjectType?: AST.TypeExpression },
  ): Map<string, AST.TypeExpression> | null {
  const bindings = new Map<string, AST.TypeExpression>();
  const genericNames = collectImplGenericNames(this, entry);
  const substituteBindings = (expr: AST.TypeExpression): AST.TypeExpression => {
    switch (expr.type) {
      case "SimpleTypeExpression": {
        const replacement = bindings.get(expr.name.name);
        return replacement ? this.cloneTypeExpression(replacement) : expr;
      }
      case "GenericTypeExpression":
        return {
          type: "GenericTypeExpression",
          base: substituteBindings(expr.base),
          arguments: (expr.arguments ?? []).map(arg => substituteBindings(arg)),
        };
      case "FunctionTypeExpression":
        return {
          type: "FunctionTypeExpression",
          paramTypes: expr.paramTypes.map(pt => substituteBindings(pt)),
          returnType: substituteBindings(expr.returnType),
        };
      case "NullableTypeExpression":
        return { type: "NullableTypeExpression", innerType: substituteBindings(expr.innerType) };
      case "ResultTypeExpression":
        return { type: "ResultTypeExpression", innerType: substituteBindings(expr.innerType) };
      case "UnionTypeExpression":
        return { type: "UnionTypeExpression", members: expr.members.map(member => substituteBindings(member)) };
      case "WildcardTypeExpression":
      default:
        return expr;
    }
  };
  const hasUnknownTypeName = (expr: AST.TypeExpression): boolean => {
    switch (expr.type) {
      case "SimpleTypeExpression":
        return !isKnownTypeName(this, expr.name.name);
      case "GenericTypeExpression":
        if (hasUnknownTypeName(expr.base)) return true;
        return (expr.arguments ?? []).some(arg => hasUnknownTypeName(arg));
      case "FunctionTypeExpression":
        if (hasUnknownTypeName(expr.returnType)) return true;
        return expr.paramTypes.some(pt => hasUnknownTypeName(pt));
      case "NullableTypeExpression":
      case "ResultTypeExpression":
        return hasUnknownTypeName(expr.innerType);
      case "UnionTypeExpression":
        return expr.members.some(member => hasUnknownTypeName(member));
      case "WildcardTypeExpression":
      default:
        return true;
    }
  };
  const isConcrete = (expr: AST.TypeExpression): boolean =>
    !typeExpressionUsesGenerics(expr, genericNames) && !hasUnknownTypeName(expr);
  const canonicalTemplate = this.expandTypeAliases(entry.def.targetType);
  const canonicalSubject = opts?.subjectType ? this.expandTypeAliases(opts.subjectType) : undefined;
  if (canonicalTemplate && canonicalSubject) {
    this.matchTypeExpressionTemplate(canonicalTemplate, canonicalSubject, genericNames, bindings);
  }
  const expectedArgs = entry.targetArgTemplates.map(t => this.expandTypeAliases(t));
  const paramUsedInTarget = (name: string): boolean => {
    if (!name) return false;
    const lookup = new Set([name]);
    if (entry.def?.targetType && typeExpressionUsesGenerics(this.expandTypeAliases(entry.def.targetType), lookup)) {
      return true;
    }
    return expectedArgs.some(arg => typeExpressionUsesGenerics(arg, lookup));
  };
  let actualArgs = opts?.typeArgs?.map(t => this.expandTypeAliases(t));
  const hasActualArgs = Boolean(actualArgs && actualArgs.length > 0);
  if (expectedArgs.length > 0) {
    if (!actualArgs) {
      actualArgs = expectedArgs.map(() => AST.wildcardTypeExpression());
    }
    if (actualArgs.length !== expectedArgs.length) return null;
    for (let i = 0; i < expectedArgs.length; i++) {
      const template = expectedArgs[i]!;
      const actual = actualArgs[i]!;
      if (!this.matchTypeExpressionTemplate(template, actual, genericNames, bindings)) return null;
    }
  }
  if (entry.def.interfaceArgs && entry.def.interfaceArgs.length > 0 && opts?.interfaceArgs) {
    const ifaceTemplates = entry.def.interfaceArgs.map(t => this.expandTypeAliases(t));
    const ifaceActualArgs = opts.interfaceArgs.map(t => substituteBindings(this.expandTypeAliases(t)));
    if (ifaceTemplates.length !== ifaceActualArgs.length) return null;
    const hasConcreteArgs = ifaceActualArgs.some(arg => isConcrete(arg));
    if (hasConcreteArgs) {
      for (let i = 0; i < ifaceTemplates.length; i++) {
        const template = ifaceTemplates[i]!;
        const actual = ifaceActualArgs[i]!;
        if (!this.matchTypeExpressionTemplate(template, actual, genericNames, bindings)) return null;
      }
    }
  }
  if (opts?.typeArgMap) {
    for (const [k, v] of opts.typeArgMap.entries()) {
      if (!bindings.has(k)) bindings.set(k, v);
    }
  }
  for (const gp of entry.genericParams) {
    if (!bindings.has(gp.name.name)) {
      if (!paramUsedInTarget(gp.name.name)) {
        continue;
      }
      if (!hasActualArgs && expectedArgs.length > 0) {
        bindings.set(gp.name.name, AST.wildcardTypeExpression());
        continue;
      }
      return null;
    }
  }
  return bindings;
};

  cls.prototype.matchTypeExpressionTemplate = function matchTypeExpressionTemplate(this: Interpreter, template: AST.TypeExpression, actual: AST.TypeExpression, genericNames: Set<string>, bindings: Map<string, AST.TypeExpression>): boolean {
  if (template.type === "WildcardTypeExpression" || actual.type === "WildcardTypeExpression") {
    return true;
  }
  if (template.type === "SimpleTypeExpression") {
    const name = template.name.name;
    if (genericNames.has(name)) {
      const existing = bindings.get(name);
      if (existing) return this.typeExpressionsEqual(existing, actual);
      bindings.set(name, actual);
      return true;
    }
    return this.typeExpressionsEqual(template, actual);
  }
  if (template.type === "GenericTypeExpression") {
    if (actual.type !== "GenericTypeExpression") return false;
    if (!this.matchTypeExpressionTemplate(template.base, actual.base, genericNames, bindings)) return false;
    const templateArgs = template.arguments ?? [];
    const actualArgs = actual.arguments ?? [];
    if (templateArgs.length !== actualArgs.length) return false;
    for (let i = 0; i < templateArgs.length; i++) {
      if (!this.matchTypeExpressionTemplate(templateArgs[i]!, actualArgs[i]!, genericNames, bindings)) return false;
    }
    return true;
  }
  return this.typeExpressionsEqual(template, actual);
};

  cls.prototype.expandImplementationTargetVariants = function expandImplementationTargetVariants(this: Interpreter, target: AST.TypeExpression): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> {
  const canonical = this.expandTypeAliases(target);
  if (canonical.type === "UnionTypeExpression") {
    const expanded: Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> = [];
    for (const member of canonical.members) {
      const memberVariants = this.expandImplementationTargetVariants(member);
      for (const variant of memberVariants) expanded.push(variant);
    }
    const seen = new Set<string>();
    const unique: Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> = [];
    for (const variant of expanded) {
      if (seen.has(variant.signature)) continue;
      seen.add(variant.signature);
      unique.push(variant);
    }
    if (unique.length === 0) {
      throw new Error("Union target must contain at least one concrete type");
    }
    return unique;
  }
  if (canonical.type === "SimpleTypeExpression") {
    const signature = this.typeExpressionToString(canonical);
    return [{ typeName: canonical.name.name, argTemplates: [], signature }];
  }
  if (canonical.type === "GenericTypeExpression") {
    const argTemplates: AST.TypeExpression[] = [];
    let current: AST.TypeExpression = canonical;
    while (current.type === "GenericTypeExpression") {
      if (current.arguments) argTemplates.unshift(...current.arguments);
      current = current.base;
    }
    if (current.type === "SimpleTypeExpression") {
      const signature = this.typeExpressionToString(canonical);
      return [{ typeName: current.name.name, argTemplates, signature }];
    }
  }
  throw new Error("Only simple, generic, or union target types supported in impl");
};

  cls.prototype.measureTemplateSpecificity = function measureTemplateSpecificity(this: Interpreter, t: AST.TypeExpression, genericNames: Set<string>): number {
  switch (t.type) {
    case "SimpleTypeExpression":
      return genericNames.has(t.name.name) ? 0 : 1;
    case "GenericTypeExpression": {
      let score = this.measureTemplateSpecificity(t.base, genericNames);
      for (const arg of t.arguments ?? []) {
        score += this.measureTemplateSpecificity(arg, genericNames);
      }
      return score;
    }
    case "NullableTypeExpression":
    case "ResultTypeExpression":
      return this.measureTemplateSpecificity(t.innerType, genericNames);
    case "UnionTypeExpression":
      return t.members.reduce((acc, member) => acc + this.measureTemplateSpecificity(member, genericNames), 0);
    default:
      return 0;
  }
};

  cls.prototype.buildSelfTypePatternBindings = function buildSelfTypePatternBindings(this: Interpreter, iface: AST.InterfaceDefinition, targetType: AST.TypeExpression): Map<string, AST.TypeExpression> {
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

  cls.prototype.attachDefaultInterfaceMethods = function attachDefaultInterfaceMethods(this: Interpreter, imp: AST.ImplementationDefinition, funcs: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>): void {
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

  cls.prototype.createDefaultMethodFunction = function createDefaultMethodFunction(this: Interpreter, sig: AST.FunctionSignature, env: Environment, targetType: AST.TypeExpression, typeBindings?: Map<string, AST.TypeExpression>): Extract<RuntimeValue, { kind: "function" }> | null {
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

  cls.prototype.substituteSelfTypeExpression = function substituteSelfTypeExpression(this: Interpreter, t: AST.TypeExpression | undefined, target: AST.TypeExpression): AST.TypeExpression | undefined {
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

  cls.prototype.substituteSelfInPattern = function substituteSelfInPattern(this: Interpreter, pattern: AST.Pattern, target: AST.TypeExpression): AST.Pattern {
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

function typeExpressionFromInfo(name: string, typeArgs?: AST.TypeExpression[]): AST.TypeExpression {
  const base: AST.SimpleTypeExpression = { type: "SimpleTypeExpression", name: AST.identifier(name) };
  if (!typeArgs || typeArgs.length === 0) return base;
  return { type: "GenericTypeExpression", base, arguments: typeArgs };
}

function interfaceInfoFromTypeExpression(expr: AST.TypeExpression | null | undefined): { name: string; args?: AST.TypeExpression[] } | null {
  if (!expr) return null;
  if (expr.type === "SimpleTypeExpression") {
    return { name: expr.name.name };
  }
  if (expr.type === "GenericTypeExpression" && expr.base.type === "SimpleTypeExpression") {
    return { name: expr.base.name.name, args: expr.arguments ?? [] };
  }
  return null;
}

function collectImplGenericNames(interp: Interpreter, entry: ImplMethodEntry): Set<string> {
  const genericNames = new Set<string>(entry.genericParams.map(g => g.name.name));
  const considerAsGeneric = (t: AST.TypeExpression | undefined): void => {
    if (!t) return;
    switch (t.type) {
      case "SimpleTypeExpression": {
        const name = t.name.name;
        if (/^[A-Z]$/.test(name) || !isKnownTypeName(interp, name)) {
          genericNames.add(name);
        }
        return;
      }
      case "GenericTypeExpression":
        considerAsGeneric(t.base);
        for (const arg of t.arguments ?? []) considerAsGeneric(arg);
        return;
      case "NullableTypeExpression":
      case "ResultTypeExpression":
        considerAsGeneric(t.innerType);
        return;
      case "UnionTypeExpression":
        for (const member of t.members) considerAsGeneric(member);
        return;
      default:
        return;
    }
  };
  for (const ifaceArg of entry.def.interfaceArgs ?? []) considerAsGeneric(ifaceArg);
  for (const template of entry.targetArgTemplates) considerAsGeneric(template);
  return genericNames;
}

function typeExpressionUsesGenerics(expr: AST.TypeExpression | undefined, genericNames: Set<string>): boolean {
  if (!expr) return false;
  switch (expr.type) {
    case "SimpleTypeExpression":
      return genericNames.has(expr.name.name);
    case "GenericTypeExpression":
      if (typeExpressionUsesGenerics(expr.base, genericNames)) return true;
      return (expr.arguments ?? []).some(arg => typeExpressionUsesGenerics(arg, genericNames));
    case "NullableTypeExpression":
    case "ResultTypeExpression":
      return typeExpressionUsesGenerics(expr.innerType, genericNames);
    case "UnionTypeExpression":
      return expr.members.some(member => typeExpressionUsesGenerics(member, genericNames));
    case "FunctionTypeExpression":
      if (typeExpressionUsesGenerics(expr.returnType, genericNames)) return true;
      return expr.paramTypes.some(param => typeExpressionUsesGenerics(param, genericNames));
    default:
      return false;
  }
}
