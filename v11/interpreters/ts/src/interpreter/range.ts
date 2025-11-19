import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { callCallableValue } from "./functions";
import type { ImplMethodEntry, V10Value } from "./values";

export type RangeImplementationRecord = {
  entry: ImplMethodEntry;
  interfaceArgs: AST.TypeExpression[];
};

declare module "./index" {
  interface InterpreterV10 {
    rangeImplementations: RangeImplementationRecord[];
    registerRangeImplementation(entry: ImplMethodEntry, interfaceArgs?: (AST.TypeExpression | null)[]): void;
    tryInvokeRangeImplementation(start: V10Value, end: V10Value, inclusive: boolean, env: Environment): V10Value | null;
    typeExpressionForValue(value: V10Value): AST.TypeExpression | null;
    describeRuntimeType(value: V10Value): string;
  }
}

function cloneTypeArgs(ctx: InterpreterV10, args?: AST.TypeExpression[]): AST.TypeExpression[] {
  if (!Array.isArray(args)) return [];
  return args
    .filter((arg): arg is AST.TypeExpression => Boolean(arg))
    .map(arg => ctx.cloneTypeExpression(arg));
}

export function applyRangeAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.registerRangeImplementation = function registerRangeImplementation(
    this: InterpreterV10,
    entry: ImplMethodEntry,
    interfaceArgs?: (AST.TypeExpression | null)[],
  ): void {
    if (!Array.isArray(this.rangeImplementations)) {
      this.rangeImplementations = [];
    }
    const args = (interfaceArgs ?? []).filter((arg): arg is AST.TypeExpression => Boolean(arg));
    this.rangeImplementations.push({ entry, interfaceArgs: args });
  };

  cls.prototype.tryInvokeRangeImplementation = function tryInvokeRangeImplementation(
    this: InterpreterV10,
    start: V10Value,
    end: V10Value,
    inclusive: boolean,
    env: Environment,
  ): V10Value | null {
    if (!Array.isArray(this.rangeImplementations) || this.rangeImplementations.length === 0) {
      return null;
    }
    const startType = this.typeExpressionForValue(start);
    const endType = this.typeExpressionForValue(end);
    if (!startType || !endType) {
      return null;
    }
    for (const record of this.rangeImplementations) {
      const args = record.interfaceArgs;
      if (!Array.isArray(args) || args.length < 2) {
        continue;
      }
      const bindings = new Map<string, AST.TypeExpression>();
      const genericNames = new Set((record.entry.genericParams ?? []).map(param => param.name.name));
      if (!this.matchTypeExpressionTemplate(args[0]!, startType, genericNames, bindings)) {
        continue;
      }
      if (!this.matchTypeExpressionTemplate(args[1]!, endType, genericNames, bindings)) {
        continue;
      }
      if (record.entry.genericParams.length > 0) {
        const expected = record.entry.genericParams.map(param => param.name.name);
        const unresolved = expected.filter(param => !bindings.has(param));
        if (unresolved.length > 0) {
          continue;
        }
      }
      const constraints = this.collectConstraintSpecs(record.entry.genericParams, record.entry.whereClause);
      if (constraints.length > 0) {
        this.enforceConstraintSpecs(
          constraints,
          bindings,
          `impl ${record.entry.def.interfaceName.name} for ${this.typeExpressionToString(record.entry.def.targetType)}`,
        );
      }
      const methodName = inclusive ? "inclusive_range" : "exclusive_range";
      const method = record.entry.methods.get(methodName);
      if (!method) {
        throw new Error(`Range implementation missing method '${methodName}'`);
      }
      return callCallableValue(this, method, [start, end], env);
    }
    return null;
  };

  cls.prototype.typeExpressionForValue = function typeExpressionForValue(this: InterpreterV10, value: V10Value): AST.TypeExpression | null {
    if (value.kind === "struct_instance") {
      const typeName = value.def.id.name;
      const base = AST.simpleTypeExpression(typeName);
      if (value.typeArguments && value.typeArguments.length > 0) {
        const args = cloneTypeArgs(this, value.typeArguments);
        return AST.genericTypeExpression(base, args);
      }
      return base;
    }
    if (value.kind === "interface_value") {
      return AST.simpleTypeExpression(value.interfaceName);
    }
    const typeName = this.getTypeNameForValue(value);
    if (typeName) {
      return AST.simpleTypeExpression(typeName);
    }
    return null;
  };

  cls.prototype.describeRuntimeType = function describeRuntimeType(this: InterpreterV10, value: V10Value): string {
    const name = this.getTypeNameForValue(value);
    if (name) return name;
    if (value.kind === "struct_instance") {
      return value.def.id.name;
    }
    return value.kind;
  };
}
