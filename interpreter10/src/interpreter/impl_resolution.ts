import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { ConstraintSpec, ImplMethodEntry, V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    enforceGenericConstraintsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall): void;
    collectConstraintSpecs(generics?: AST.GenericParameter[], where?: AST.WhereClauseConstraint[]): ConstraintSpec[];
    mapTypeArguments(generics: AST.GenericParameter[] | undefined, provided: AST.TypeExpression[] | undefined, context: string): Map<string, AST.TypeExpression>;
    enforceConstraintSpecs(constraints: ConstraintSpec[], typeArgMap: Map<string, AST.TypeExpression>, context: string): void;
    ensureTypeSatisfiesInterface(typeInfo: { name: string; typeArgs: AST.TypeExpression[] }, interfaceType: AST.TypeExpression, context: string, visited: Set<string>): void;
    bindTypeArgumentsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, env: Environment): void;
    collectInterfaceConstraintExpressions(typeExpr: AST.TypeExpression, memo?: Set<string>): AST.TypeExpression[];
    findMethod(typeName: string, methodName: string, opts?: { typeArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; interfaceName?: string }): Extract<V10Value, { kind: "function" }> | null;
    compareMethodMatches(a: { entry: ImplMethodEntry; bindings: Map<string, AST.TypeExpression>; constraints: ConstraintSpec[] }, b: { entry: ImplMethodEntry; bindings: Map<string, AST.TypeExpression>; constraints: ConstraintSpec[] }): number;
    buildConstraintKeySet(constraints: ConstraintSpec[]): Set<string>;
    isConstraintSuperset(a: Set<string>, b: Set<string>): boolean;
    isProperSubset(a: string[], b: string[]): boolean;
    matchImplEntry(entry: ImplMethodEntry, opts?: { typeArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> }): Map<string, AST.TypeExpression> | null;
    matchTypeExpressionTemplate(template: AST.TypeExpression, actual: AST.TypeExpression, genericNames: Set<string>, bindings: Map<string, AST.TypeExpression>): boolean;
    expandImplementationTargetVariants(target: AST.TypeExpression): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }>;
    computeImplSpecificity(entry: ImplMethodEntry, bindings: Map<string, AST.TypeExpression>, constraints: ConstraintSpec[]): number;
    measureTemplateSpecificity(t: AST.TypeExpression, genericNames: Set<string>): number;
    attachDefaultInterfaceMethods(imp: AST.ImplementationDefinition, funcs: Map<string, Extract<V10Value, { kind: "function" }>>): void;
    createDefaultMethodFunction(sig: AST.FunctionSignature, env: Environment, targetType: AST.TypeExpression): Extract<V10Value, { kind: "function" }> | null;
    substituteSelfTypeExpression(t: AST.TypeExpression | undefined, target: AST.TypeExpression): AST.TypeExpression | undefined;
    substituteSelfInPattern(pattern: AST.Pattern, target: AST.TypeExpression): AST.Pattern;
  }
}

export function applyImplResolutionAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.enforceGenericConstraintsIfAny = function enforceGenericConstraintsIfAny(this: InterpreterV10, funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall): void {
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

  cls.prototype.collectConstraintSpecs = function collectConstraintSpecs(this: InterpreterV10, generics?: AST.GenericParameter[], where?: AST.WhereClauseConstraint[]): ConstraintSpec[] {
  const all: ConstraintSpec[] = [];
  if (generics) {
    for (const gp of generics) {
      if (!gp.constraints) continue;
      for (const c of gp.constraints) {
        all.push({ typeParam: gp.name.name, ifaceType: c.interfaceType });
      }
    }
  }
  if (where) {
    for (const clause of where) {
      for (const c of clause.constraints) {
        all.push({ typeParam: clause.typeParam.name, ifaceType: c.interfaceType });
      }
    }
  }
  return all;
};

  cls.prototype.mapTypeArguments = function mapTypeArguments(this: InterpreterV10, generics: AST.GenericParameter[] | undefined, provided: AST.TypeExpression[] | undefined, context: string): Map<string, AST.TypeExpression> {
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

  cls.prototype.enforceConstraintSpecs = function enforceConstraintSpecs(this: InterpreterV10, constraints: ConstraintSpec[], typeArgMap: Map<string, AST.TypeExpression>, context: string): void {
  for (const c of constraints) {
    const actual = typeArgMap.get(c.typeParam);
    if (!actual) {
      throw new Error(`Missing type argument for '${c.typeParam}' required by constraints`);
    }
    const typeInfo = this.parseTypeExpression(actual);
    if (!typeInfo) continue;
    this.ensureTypeSatisfiesInterface(typeInfo, c.ifaceType, c.typeParam, new Set());
  }
};

  cls.prototype.ensureTypeSatisfiesInterface = function ensureTypeSatisfiesInterface(this: InterpreterV10, typeInfo: { name: string; typeArgs: AST.TypeExpression[] }, interfaceType: AST.TypeExpression, context: string, visited: Set<string>): void {
  const ifaceInfo = this.parseTypeExpression(interfaceType);
  if (!ifaceInfo) return;
  if (visited.has(ifaceInfo.name)) return;
  visited.add(ifaceInfo.name);
  const iface = this.interfaces.get(ifaceInfo.name);
  if (!iface) throw new Error(`Unknown interface '${ifaceInfo.name}' in constraint on '${context}'`);
  for (const base of iface.baseInterfaces ?? []) {
    this.ensureTypeSatisfiesInterface(typeInfo, base, context, visited);
  }
  for (const sig of iface.signatures) {
    const methodName = sig.name.name;
    const method = this.findMethod(typeInfo.name, methodName, { typeArgs: typeInfo.typeArgs, interfaceName: ifaceInfo.name });
    if (!method) {
      throw new Error(`Type '${typeInfo.name}' does not satisfy interface '${ifaceInfo.name}': missing method '${methodName}'`);
    }
  }
};

  cls.prototype.bindTypeArgumentsIfAny = function bindTypeArgumentsIfAny(this: InterpreterV10, funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, env: Environment): void {
  const generics = (funcNode as any).genericParams as AST.GenericParameter[] | undefined;
  if (!generics || generics.length === 0) return;
  const args = call.typeArguments ?? [];
  const count = Math.min(generics.length, args.length);
  for (let i = 0; i < count; i++) {
    const gp = generics[i]!;
    const ta = args[i]!;
    const name = `${gp.name.name}_type`;
    const s = this.typeExpressionToString(ta);
    try { env.define(name, { kind: "string", value: s }); } catch {}
  }
};

  cls.prototype.collectInterfaceConstraintExpressions = function collectInterfaceConstraintExpressions(this: InterpreterV10, typeExpr: AST.TypeExpression, memo: Set<string> = new Set()): AST.TypeExpression[] {
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

  cls.prototype.findMethod = function findMethod(this: InterpreterV10, typeName: string, methodName: string, opts?: { typeArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; interfaceName?: string }): Extract<V10Value, { kind: "function" }> | null {
  const inherent = this.inherentMethods.get(typeName);
  if (inherent && inherent.has(methodName)) return inherent.get(methodName)!;
  const entries = this.implMethods.get(typeName);
  let constraintError: Error | null = null;
  const matches: Array<{
    method: Extract<V10Value, { kind: "function" }>;
    score: number;
    entry: ImplMethodEntry;
    constraints: ConstraintSpec[];
  }> = [];
  if (entries) {
    for (const entry of entries) {
      if (opts?.interfaceName && entry.def.interfaceName.name !== opts.interfaceName) continue;
      const bindings = this.matchImplEntry(entry, opts);
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
      const method = entry.methods.get(methodName);
      if (!method) continue;
      const score = this.computeImplSpecificity(entry, bindings, constraints);
      matches.push({ method, score, entry, constraints });
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
    const detail = Array.from(new Set(contenders.map(c => this.typeExpressionToString(c.entry.def.targetType)))).join(", ");
    throw new Error(`Ambiguous method '${methodName}' for type '${typeName}' (candidates: ${detail})`);
  }
  return best.method;
};

  cls.prototype.compareMethodMatches = function compareMethodMatches(this: InterpreterV10, a: { method: Extract<V10Value, { kind: "function" }>; score: number; entry: ImplMethodEntry; constraints: ConstraintSpec[] }, b: { method: Extract<V10Value, { kind: "function" }>; score: number; entry: ImplMethodEntry; constraints: ConstraintSpec[] }): number {
  if (a.score > b.score) return 1;
  if (a.score < b.score) return -1;
  const aUnion = a.entry.unionVariantSignatures;
  const bUnion = b.entry.unionVariantSignatures;
  if (aUnion && !bUnion) return -1;
  if (!aUnion && bUnion) return 1;
  if (aUnion && bUnion) {
    if (this.isProperSubset(aUnion, bUnion)) return 1;
    if (this.isProperSubset(bUnion, aUnion)) return -1;
    if (aUnion.length !== bUnion.length) {
      return aUnion.length < bUnion.length ? 1 : -1;
    }
  }
  const aConstraints = this.buildConstraintKeySet(a.constraints);
  const bConstraints = this.buildConstraintKeySet(b.constraints);
  if (this.isConstraintSuperset(aConstraints, bConstraints)) return 1;
  if (this.isConstraintSuperset(bConstraints, aConstraints)) return -1;
  return 0;
};

  cls.prototype.buildConstraintKeySet = function buildConstraintKeySet(this: InterpreterV10, constraints: ConstraintSpec[]): Set<string> {
  const set = new Set<string>();
  for (const c of constraints) {
    const expanded = this.collectInterfaceConstraintExpressions(c.ifaceType);
    for (const expr of expanded) {
      set.add(`${c.typeParam}->${this.typeExpressionToString(expr)}`);
    }
  }
  return set;
};

  cls.prototype.isConstraintSuperset = function isConstraintSuperset(this: InterpreterV10, a: Set<string>, b: Set<string>): boolean {
  if (a.size <= b.size) return false;
  for (const key of b) {
    if (!a.has(key)) return false;
  }
  return true;
};

  cls.prototype.isProperSubset = function isProperSubset(this: InterpreterV10, a: string[], b: string[]): boolean {
  const aSet = new Set(a);
  const bSet = new Set(b);
  if (aSet.size >= bSet.size) return false;
  for (const val of aSet) {
    if (!bSet.has(val)) return false;
  }
  return true;
};

  cls.prototype.matchImplEntry = function matchImplEntry(this: InterpreterV10, entry: ImplMethodEntry, opts?: { typeArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> }): Map<string, AST.TypeExpression> | null {
  const bindings = new Map<string, AST.TypeExpression>();
  const genericNames = new Set(entry.genericParams.map(g => g.name.name));
  const expectedArgs = entry.targetArgTemplates;
  const actualArgs = opts?.typeArgs;
  if (expectedArgs.length > 0) {
    if (!actualArgs || actualArgs.length !== expectedArgs.length) return null;
    for (let i = 0; i < expectedArgs.length; i++) {
      const template = expectedArgs[i]!;
      const actual = actualArgs[i]!;
      if (!this.matchTypeExpressionTemplate(template, actual, genericNames, bindings)) return null;
    }
  }
  if (opts?.typeArgMap) {
    for (const [k, v] of opts.typeArgMap.entries()) {
      if (!bindings.has(k)) bindings.set(k, v);
    }
  }
  for (const gp of entry.genericParams) {
    if (!bindings.has(gp.name.name)) return null;
  }
  return bindings;
};

  cls.prototype.matchTypeExpressionTemplate = function matchTypeExpressionTemplate(this: InterpreterV10, template: AST.TypeExpression, actual: AST.TypeExpression, genericNames: Set<string>, bindings: Map<string, AST.TypeExpression>): boolean {
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

  cls.prototype.expandImplementationTargetVariants = function expandImplementationTargetVariants(this: InterpreterV10, target: AST.TypeExpression): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> {
  if (target.type === "UnionTypeExpression") {
    const expanded: Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> = [];
    for (const member of target.members) {
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
  if (target.type === "SimpleTypeExpression") {
    const signature = this.typeExpressionToString(target);
    return [{ typeName: target.name.name, argTemplates: [], signature }];
  }
  if (target.type === "GenericTypeExpression" && target.base.type === "SimpleTypeExpression") {
    const signature = this.typeExpressionToString(target);
    return [{ typeName: target.base.name.name, argTemplates: target.arguments ?? [], signature }];
  }
  throw new Error("Only simple, generic, or union target types supported in impl");
};

  cls.prototype.computeImplSpecificity = function computeImplSpecificity(this: InterpreterV10, entry: ImplMethodEntry, bindings: Map<string, AST.TypeExpression>, constraints: ConstraintSpec[]): number {
  const genericNames = new Set(entry.genericParams.map(g => g.name.name));
  let concreteScore = 0;
  for (const template of entry.targetArgTemplates) {
    concreteScore += this.measureTemplateSpecificity(template, genericNames);
  }
  const constraintScore = constraints.length;
  const bindingScore = bindings.size;
  const unionPenalty = entry.unionVariantSignatures ? entry.unionVariantSignatures.length : 0;
  return concreteScore * 100 + constraintScore * 10 + bindingScore - unionPenalty;
};

  cls.prototype.measureTemplateSpecificity = function measureTemplateSpecificity(this: InterpreterV10, t: AST.TypeExpression, genericNames: Set<string>): number {
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

  cls.prototype.attachDefaultInterfaceMethods = function attachDefaultInterfaceMethods(this: InterpreterV10, imp: AST.ImplementationDefinition, funcs: Map<string, Extract<V10Value, { kind: "function" }>>): void {
  const interfaceName = imp.interfaceName.name;
  const iface = this.interfaces.get(interfaceName);
  if (!iface) return;
  const ifaceEnv = this.interfaceEnvs.get(interfaceName) ?? this.globals;
  const targetType = imp.targetType;
  for (const sig of iface.signatures) {
    if (!sig.defaultImpl) continue;
    const methodName = sig.name.name;
    if (funcs.has(methodName)) continue;
    const defaultFunc = this.createDefaultMethodFunction(sig, ifaceEnv, targetType);
    if (defaultFunc) funcs.set(methodName, defaultFunc);
  }
};

  cls.prototype.createDefaultMethodFunction = function createDefaultMethodFunction(this: InterpreterV10, sig: AST.FunctionSignature, env: Environment, targetType: AST.TypeExpression): Extract<V10Value, { kind: "function" }> | null {
  if (!sig.defaultImpl) return null;
  const params = sig.params.map(param => {
    const substitutedPattern = this.substituteSelfInPattern(param.name as AST.Pattern, targetType);
    const substitutedType = this.substituteSelfTypeExpression(param.paramType, targetType);
    if (substitutedPattern === param.name && substitutedType === param.paramType) return param;
    return { type: "FunctionParameter", name: substitutedPattern, paramType: substitutedType } as AST.FunctionParameter;
  });
  const returnType = this.substituteSelfTypeExpression(sig.returnType, targetType) ?? sig.returnType;
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
  return { kind: "function", node: fnDef, closureEnv: env };
};

  cls.prototype.substituteSelfTypeExpression = function substituteSelfTypeExpression(this: InterpreterV10, t: AST.TypeExpression | undefined, target: AST.TypeExpression): AST.TypeExpression | undefined {
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

  cls.prototype.substituteSelfInPattern = function substituteSelfInPattern(this: InterpreterV10, pattern: AST.Pattern, target: AST.TypeExpression): AST.Pattern {
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
