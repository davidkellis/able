import * as AST from "../../ast";
import type { Interpreter } from "../index";
import type { ConstraintSpec, ImplMethodEntry, RuntimeValue } from "../values";
import { formatAmbiguousImplementationError } from "./diagnostics";
import {
  collectImplGenericNames,
  interfaceInfoFromTypeExpression,
  isKnownTypeName,
  typeExpressionFromInfo,
  typeExpressionUsesGenerics,
} from "./helpers";

export function applyImplCandidateAugmentations(cls: typeof Interpreter): void {
  let traceMethodReported = false;
  cls.prototype.findMethod = function findMethod(
    this: Interpreter,
    typeName: string,
    methodName: string,
    opts?: {
      typeArgs?: AST.TypeExpression[];
      interfaceArgs?: AST.TypeExpression[];
      typeArgMap?: Map<string, AST.TypeExpression>;
      interfaceName?: string;
      includeInherent?: boolean;
    },
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
      throw new Error(formatAmbiguousImplementationError(this, ifaceName, typeName, contenders));
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
      return { ok: false, error: new Error(formatAmbiguousImplementationError(this, interfaceName, typeName, contenders)) };
    }
    return { ok: true };
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

  cls.prototype.matchTypeExpressionTemplate = function matchTypeExpressionTemplate(
    this: Interpreter,
    template: AST.TypeExpression,
    actual: AST.TypeExpression,
    genericNames: Set<string>,
    bindings: Map<string, AST.TypeExpression>,
  ): boolean {
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

  cls.prototype.expandImplementationTargetVariants = function expandImplementationTargetVariants(
    this: Interpreter,
    target: AST.TypeExpression,
  ): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }> {
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
}
