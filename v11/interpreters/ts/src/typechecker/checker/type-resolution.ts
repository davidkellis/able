import type * as AST from "../../ast";
import {
  formatType,
  primitiveType,
  unknownType,
  type LiteralInfo,
  type PrimitiveName,
  type TypeInfo,
} from "../types";
import { getIntegerTypeInfo, hasIntegerBounds, integerBounds } from "../numeric";

const PRIMITIVE_TYPE_NAMES = new Set<PrimitiveName>([
  "i8",
  "i16",
  "i32",
  "i64",
  "i128",
  "u8",
  "u16",
  "u32",
  "u64",
  "u128",
  "f32",
  "f64",
  "bool",
  "String",
  "IoHandle",
  "ProcHandle",
  "char",
  "nil",
  "void",
]);

const BUILTIN_TYPE_ARITY = new Map<string, number>([
  ["Array", 1],
  ["Iterator", 1],
  ["Range", 1],
  ["Proc", 1],
  ["Future", 1],
  ["Map", 2],
  ["HashMap", 2],
  ["Channel", 1],
  ["Mutex", 0],
]);

export type TypeResolutionContext = {
  getTypeAlias(name: string): AST.TypeAliasDefinition | undefined;
  getInterfaceDefinition(name: string): AST.InterfaceDefinition | undefined;
  hasInterfaceDefinition(name: string): boolean;
  getStructDefinition(name: string): AST.StructDefinition | undefined;
  getUnionDefinition(name: string): AST.UnionDefinition | undefined;
  hasUnionDefinition(name: string): boolean;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  report?: (message: string, node?: AST.Node | null) => void;
};

export type TypeResolutionHelpers = {
  resolveTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
  ): TypeInfo;
  instantiateTypeAlias(
    definition: AST.TypeAliasDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo;
  instantiateUnionDefinition(
    definition: AST.UnionDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo;
  typeInfosEquivalent(a: TypeInfo | undefined, b: TypeInfo | undefined): boolean;
  isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  canonicalizeStructuralType(type: TypeInfo): TypeInfo;
  typeExpressionsEquivalent(
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): boolean;
  describeTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, string>): string;
  formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string;
  lookupSubstitution(name: string | null, substitutions?: Map<string, string>): string;
  describeTypeArgument(type: TypeInfo): string;
  appendInterfaceArgsToLabel(label: string, args: string[]): string;
};

export function createTypeResolutionHelpers(context: TypeResolutionContext): TypeResolutionHelpers {
  const reportedTypeArgumentArity = new WeakSet<AST.TypeExpression>();

  const expectedTypeArgumentCount = (name: string): number | null => {
    const alias = context.getTypeAlias(name);
    if (alias) {
      return Array.isArray(alias.genericParams) ? alias.genericParams.length : 0;
    }
    const structDef = context.getStructDefinition(name);
    if (structDef) {
      return Array.isArray(structDef.genericParams) ? structDef.genericParams.length : 0;
    }
    const unionDef = context.getUnionDefinition(name);
    if (unionDef) {
      return Array.isArray(unionDef.genericParams) ? unionDef.genericParams.length : 0;
    }
    const ifaceDef = context.getInterfaceDefinition(name);
    if (ifaceDef) {
      return Array.isArray(ifaceDef.genericParams) ? ifaceDef.genericParams.length : 0;
    }
    const builtinArity = BUILTIN_TYPE_ARITY.get(name);
    if (builtinArity !== undefined) {
      return builtinArity;
    }
    if (PRIMITIVE_TYPE_NAMES.has(name as PrimitiveName)) {
      return 0;
    }
    return null;
  };

  const reportTypeArgumentArity = (
    name: string,
    expected: number,
    actual: number,
    node: AST.TypeExpression,
  ) => {
    if (!context.report || reportedTypeArgumentArity.has(node)) {
      return;
    }
    reportedTypeArgumentArity.add(node);
    context.report(
      `typechecker: type '${name}' expects ${expected} type argument(s), got ${actual}`,
      node,
    );
  };

  function resolveTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    if (!expr) return unknownType;
    switch (expr.type) {
      case "SimpleTypeExpression": {
        const name = context.getIdentifierName(expr.name);
        if (!name) return unknownType;
        if (substitutions?.has(name)) {
          return substitutions.get(name) ?? unknownType;
        }
        switch (name) {
          case "i8":
          case "i16":
          case "i32":
          case "i64":
          case "i128":
          case "u8":
          case "u16":
          case "u32":
          case "u64":
          case "u128":
          case "f32":
          case "f64":
          case "bool":
          case "String":
          case "IoHandle":
          case "ProcHandle":
          case "char":
          case "nil":
          case "void":
            return primitiveType(name as PrimitiveName);
      default: {
        const structDef = context.getStructDefinition(name);
        if (structDef) {
          return {
            kind: "struct",
            name,
            typeArguments: [],
            definition: structDef,
          };
        }
        const unionDef = context.getUnionDefinition(name);
        if (unionDef) {
          return instantiateUnionDefinition(unionDef, [], substitutions);
        }
        if (context.hasInterfaceDefinition(name)) {
          return { kind: "interface", name, typeArguments: [] };
        }
        const alias = context.getTypeAlias(name);
        if (alias) {
          return instantiateTypeAlias(alias, [], substitutions);
        }
        const builtin = resolveBuiltinStructuralType(name, []);
        if (builtin) {
          return builtin;
        }
        return unknownType;
        }
      }
      }
      case "GenericTypeExpression": {
        const baseName = getIdentifierNameFromTypeExpression(expr.base);
        if (!baseName) return unknownType;
        const baseArgs =
          expr.base?.type === "GenericTypeExpression" && Array.isArray(expr.base.arguments) ? expr.base.arguments : [];
        const rawArgs = [
          ...baseArgs,
          ...(Array.isArray(expr.arguments) ? expr.arguments : []),
        ] as Array<AST.TypeExpression | null | undefined>;
        const typeArguments = rawArgs.map((arg) => resolveTypeExpression(arg, substitutions));
        if (!substitutions?.has(baseName) && baseName !== "Self") {
          const expected = expectedTypeArgumentCount(baseName);
          if (expected !== null && typeArguments.length > expected) {
            reportTypeArgumentArity(baseName, expected, typeArguments.length, expr);
          }
        }
        const substitutedBase = substitutions?.get(baseName);
        if (substitutedBase) {
          if (substitutedBase.kind === "struct") {
            const def = substitutedBase.definition ?? context.getStructDefinition(substitutedBase.name);
            return { kind: "struct", name: substitutedBase.name, typeArguments, definition: def };
          }
          if (substitutedBase.kind === "interface") {
            const def = substitutedBase.definition ?? context.getInterfaceDefinition(substitutedBase.name);
            return { kind: "interface", name: substitutedBase.name, typeArguments, definition: def };
          }
          if (substitutedBase.kind === "type_parameter") {
            return unknownType;
          }
        }
        const structDef = context.getStructDefinition(baseName);
        if (structDef) {
          return {
            kind: "struct",
            name: baseName,
            typeArguments,
            definition: structDef,
          };
        }
        const unionDef = context.getUnionDefinition(baseName);
        if (unionDef) {
          return instantiateUnionDefinition(unionDef, typeArguments, substitutions);
        }
        if (context.hasInterfaceDefinition(baseName)) {
          return { kind: "interface", name: baseName, typeArguments };
        }
        const alias = context.getTypeAlias(baseName);
        if (alias) {
          return instantiateTypeAlias(alias, typeArguments, substitutions);
        }
        const builtin = resolveBuiltinStructuralType(baseName, typeArguments);
        if (builtin) {
          return builtin;
        }
        return unknownType;
      }
      case "NullableTypeExpression":
        return {
          kind: "nullable",
          inner: resolveTypeExpression(expr.innerType, substitutions),
        };
      case "ResultTypeExpression":
        return {
          kind: "result",
          inner: resolveTypeExpression(expr.innerType, substitutions),
        };
      case "UnionTypeExpression": {
        const members = Array.isArray(expr.members)
          ? expr.members.map((member) => resolveTypeExpression(member, substitutions))
          : [];
        return { kind: "union", members };
      }
      case "FunctionTypeExpression": {
        const parameters = Array.isArray(expr.paramTypes)
          ? expr.paramTypes.map((param) => resolveTypeExpression(param, substitutions))
          : [];
        const returnType = resolveTypeExpression(expr.returnType, substitutions);
        return {
          kind: "function",
          parameters,
          returnType,
        };
      }
      default:
        return unknownType;
    }
  }

  function instantiateTypeAlias(
    definition: AST.TypeAliasDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    const substitution = outerSubstitutions ? new Map(outerSubstitutions) : new Map<string, TypeInfo>();
    if (Array.isArray(definition.genericParams)) {
      definition.genericParams.forEach((param, index) => {
        const name = context.getIdentifierName(param?.name);
        if (!name) {
          return;
        }
        const arg = typeArguments[index] ?? unknownType;
        substitution.set(name, arg);
      });
    }
    return resolveTypeExpression(definition.targetType, substitution);
  }

  function instantiateUnionDefinition(
    definition: AST.UnionDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    const substitutions = new Map<string, TypeInfo>();
    if (outerSubstitutions) {
      for (const [key, value] of outerSubstitutions.entries()) {
        substitutions.set(key, value);
      }
    }
    const paramNames = Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => context.getIdentifierName(param?.name))
          .filter((name): name is string => Boolean(name))
      : [];
    paramNames.forEach((name, index) => {
      substitutions.set(name, typeArguments[index] ?? unknownType);
    });
    const members = Array.isArray(definition.variants)
      ? definition.variants.map((variant) => resolveTypeExpression(variant, substitutions))
      : [];
    return { kind: "union", members };
  }

  function canonicalizeStructuralType(type: TypeInfo): TypeInfo {
    if (!type || type.kind !== "struct") {
      return type;
    }
    const args = Array.isArray(type.typeArguments) ? type.typeArguments : [];
    const firstArg = args[0] ?? unknownType;
    switch (type.name) {
      case "Array":
        return { kind: "array", element: firstArg ?? unknownType };
      case "Iterator":
        return { kind: "iterator", element: firstArg ?? unknownType };
      case "Range":
        return { kind: "range", element: firstArg ?? unknownType };
      case "Proc":
        return { kind: "proc", result: firstArg ?? unknownType };
      case "Future":
        return { kind: "future", result: firstArg ?? unknownType };
      case "Map": {
        const key = args[0] ?? unknownType;
        const value = args[1] ?? unknownType;
        return { kind: "map", key, value };
      }
      default:
        return type;
    }
  }

  function resolveBuiltinStructuralType(name: string, typeArguments: TypeInfo[]): TypeInfo | null {
    switch (name) {
      case "Array":
        return { kind: "array", element: typeArguments[0] ?? unknownType };
      case "Iterator":
        return { kind: "iterator", element: typeArguments[0] ?? unknownType };
      case "Range":
        return { kind: "range", element: typeArguments[0] ?? unknownType };
      case "Proc":
        return { kind: "proc", result: typeArguments[0] ?? unknownType };
      case "Future":
        return { kind: "future", result: typeArguments[0] ?? unknownType };
      case "Map":
        return {
          kind: "map",
          key: typeArguments[0] ?? unknownType,
          value: typeArguments[1] ?? unknownType,
        };
      case "Channel":
      case "Mutex":
        return { kind: "struct", name, typeArguments };
      default:
        return null;
    }
  }

  function literalValueToBigInt(literal: LiteralInfo): bigint {
    if (typeof literal.value === "bigint") {
      return literal.value;
    }
    if (!Number.isFinite(literal.value)) {
      return BigInt(0);
    }
    return BigInt(Math.trunc(literal.value));
  }

  function literalFitsPrimitive(literal: LiteralInfo, expected: PrimitiveName, literalType: PrimitiveName): boolean {
    if (literal.literalKind === "integer") {
      if (literal.explicit) {
        return literalType === expected;
      }
      if (expected === "f32" || expected === "f64") {
        return true;
      }
      if (!hasIntegerBounds(expected)) {
        return literalType === expected;
      }
      const bounds = integerBounds(expected);
      const value = literalValueToBigInt(literal);
      return value >= bounds.min && value <= bounds.max;
    }
    if (literal.literalKind === "float") {
      if (literal.explicit) {
        return literalType === expected;
      }
      return expected === "f32" || expected === "f64";
    }
    return false;
  }

  function typeInfosEquivalent(a: TypeInfo | undefined, b: TypeInfo | undefined): boolean {
    if (!a || a.kind === "unknown" || !b || b.kind === "unknown") {
      return true;
    }
    if (a.kind === "type_parameter" || b.kind === "type_parameter") {
      if (a.kind === "type_parameter" && b.kind === "type_parameter") {
        return a.name === b.name;
      }
      return true;
    }
    let left: TypeInfo = a;
    let right: TypeInfo = b;
    const normalizedLeft = canonicalizeStructuralType(left);
    const normalizedRight = canonicalizeStructuralType(right);
    if (normalizedLeft !== left || normalizedRight !== right) {
      return typeInfosEquivalent(normalizedLeft, normalizedRight);
    }
    left = normalizedLeft;
    right = normalizedRight;
    if (left.kind === "primitive" && right.kind === "primitive") {
      if (left.literal && literalFitsPrimitive(left.literal, right.name, left.name)) {
        return true;
      }
      if (right.literal && literalFitsPrimitive(right.literal, left.name, right.name)) {
        return true;
      }
      return left.name === right.name;
    }
    if (left.kind !== right.kind) {
      return false;
    }
    switch (left.kind) {
      case "array": {
        const other = right as Extract<TypeInfo, { kind: "array" }>;
        return typeInfosEquivalent(left.element, other.element);
      }
      case "map": {
        const other = right as Extract<TypeInfo, { kind: "map" }>;
        return typeInfosEquivalent(left.key, other.key) && typeInfosEquivalent(left.value, other.value);
      }
      case "iterator":
      case "range": {
        const other = right as Extract<TypeInfo, { kind: typeof left.kind }>;
        return typeInfosEquivalent(left.element, other.element);
      }
      case "proc":
      case "future": {
        const other = right as Extract<TypeInfo, { kind: typeof left.kind }>;
        return typeInfosEquivalent(left.result, other.result);
      }
      case "nullable":
      case "result": {
        const other = right as Extract<TypeInfo, { kind: typeof left.kind }>;
        return typeInfosEquivalent(left.inner, other.inner);
      }
      case "union": {
        const otherMembers = (right as typeof left).members ?? [];
        if (left.members.length !== otherMembers.length) {
          return false;
        }
        for (let i = 0; i < left.members.length; i += 1) {
          if (!typeInfosEquivalent(left.members[i], otherMembers[i])) {
            return false;
          }
        }
        return true;
      }
      default:
        return formatType(a) === formatType(b);
    }
  }

  function canWidenIntegerType(actual: PrimitiveName, expected: PrimitiveName): boolean {
    const actualInfo = getIntegerTypeInfo(actual);
    const expectedInfo = getIntegerTypeInfo(expected);
    if (!actualInfo || !expectedInfo) {
      return false;
    }
    return actualInfo.min >= expectedInfo.min && actualInfo.max <= expectedInfo.max;
  }

  function isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean {
    if (!actual || actual.kind === "unknown" || !expected || expected.kind === "unknown") {
      return true;
    }
    if (actual.kind === "type_parameter" || expected.kind === "type_parameter") {
      return true;
    }
    const normalizedActual = canonicalizeStructuralType(actual);
    const normalizedExpected = canonicalizeStructuralType(expected);
    if (typeInfosEquivalent(normalizedActual, normalizedExpected)) {
      return true;
    }
    if (
      normalizedExpected.kind === "map" &&
      normalizedActual.kind === "struct" &&
      normalizedActual.name === "HashMap"
    ) {
      const args = normalizedActual.typeArguments ?? [];
      const keyArg = args[0] ?? unknownType;
      const valueArg = args[1] ?? unknownType;
      return isTypeAssignable(keyArg, normalizedExpected.key) && isTypeAssignable(valueArg, normalizedExpected.value);
    }
    if (normalizedExpected.kind === "nullable") {
      if (normalizedActual.kind === "primitive" && normalizedActual.name === "nil") {
        return true;
      }
      return isTypeAssignable(normalizedActual, normalizedExpected.inner);
    }
    if (normalizedExpected.kind === "result") {
      if (isTypeAssignable(normalizedActual, normalizedExpected.inner)) {
        return true;
      }
      return false;
    }
    if (normalizedExpected.kind === "union" && normalizedActual.kind === "union") {
      return normalizedActual.members.every((member) => isTypeAssignable(member, normalizedExpected));
    }
    if (normalizedExpected.kind === "union" && Array.isArray(normalizedExpected.members)) {
      return normalizedExpected.members.some((member) => isTypeAssignable(normalizedActual, member));
    }
    if (normalizedActual.kind === "nullable") {
      return isTypeAssignable(normalizedActual.inner, normalizedExpected);
    }
    if (normalizedActual.kind === "function" && normalizedExpected.kind === "function") {
      const actualParams = Array.isArray(normalizedActual.parameters) ? normalizedActual.parameters : [];
      const expectedParams = Array.isArray(normalizedExpected.parameters) ? normalizedExpected.parameters : [];
      if (actualParams.length !== expectedParams.length) {
        return false;
      }
      for (let index = 0; index < expectedParams.length; index += 1) {
        if (!isTypeAssignable(actualParams[index], expectedParams[index])) {
          return false;
        }
      }
      return isTypeAssignable(normalizedActual.returnType, normalizedExpected.returnType);
    }
    if (normalizedActual.kind === "struct" && normalizedExpected.kind === "struct") {
      if (normalizedActual.name !== normalizedExpected.name) {
        return false;
      }
      const actualArgs = normalizedActual.typeArguments ?? [];
      const expectedArgs = normalizedExpected.typeArguments ?? [];
      if (actualArgs.length !== expectedArgs.length) {
        return expectedArgs.length === 0;
      }
      for (let index = 0; index < expectedArgs.length; index += 1) {
        if (!isTypeAssignable(actualArgs[index], expectedArgs[index])) {
          return false;
        }
      }
      return true;
    }
    if (normalizedActual.kind === "interface" && normalizedExpected.kind === "interface") {
      if (normalizedActual.name !== normalizedExpected.name) {
        return false;
      }
      const actualArgs = normalizedActual.typeArguments ?? [];
      const expectedArgs = normalizedExpected.typeArguments ?? [];
      if (actualArgs.length !== expectedArgs.length) {
        return expectedArgs.length === 0;
      }
      for (let index = 0; index < expectedArgs.length; index += 1) {
        if (!isTypeAssignable(actualArgs[index], expectedArgs[index])) {
          return false;
        }
      }
      return true;
    }
    if (normalizedActual.kind === "primitive" && normalizedExpected.kind === "primitive") {
      if (canWidenIntegerType(normalizedActual.name, normalizedExpected.name)) {
        return true;
      }
    }
    return false;
  }

  function canonicalizeLiteralComparison(type: TypeInfo): TypeInfo {
    if (!type) return type;
    if (type.kind === "interface" && type.name === "Iterator") {
      const element = Array.isArray(type.typeArguments) ? type.typeArguments[0] ?? unknownType : unknownType;
      return { kind: "iterator", element };
    }
    return type;
  }

  function describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null {
    if (!actual || !expected) {
      return null;
    }
    let normalizedActual = canonicalizeLiteralComparison(actual);
    let normalizedExpected = canonicalizeLiteralComparison(expected);
    if (normalizedActual.kind === "struct" && normalizedActual.name === "HashMap") {
      const args = normalizedActual.typeArguments ?? [];
      normalizedActual = { kind: "map", key: args[0] ?? unknownType, value: args[1] ?? unknownType };
    }
    if (normalizedExpected.kind === "struct" && normalizedExpected.name === "HashMap") {
      const args = normalizedExpected.typeArguments ?? [];
      normalizedExpected = { kind: "map", key: args[0] ?? unknownType, value: args[1] ?? unknownType };
    }
    const nextActual = canonicalizeStructuralType(normalizedActual);
    const nextExpected = canonicalizeStructuralType(normalizedExpected);
    if (nextActual !== normalizedActual || nextExpected !== normalizedExpected) {
      return describeLiteralMismatch(nextActual, nextExpected);
    }
    normalizedActual = nextActual;
    normalizedExpected = nextExpected;
    if (normalizedActual.kind === "array" && normalizedExpected.kind === "array") {
      return describeLiteralMismatch(normalizedActual.element, normalizedExpected.element);
    }
    if (normalizedActual.kind === "map" && normalizedExpected.kind === "map") {
      return (
        describeLiteralMismatch(normalizedActual.key, normalizedExpected.key) ??
        describeLiteralMismatch(normalizedActual.value, normalizedExpected.value)
      );
    }
    if (normalizedActual.kind === "iterator" && normalizedExpected.kind === "iterator") {
      return describeLiteralMismatch(normalizedActual.element, normalizedExpected.element);
    }
    if (normalizedActual.kind === "range" && normalizedExpected.kind === "range") {
      const elementMessage = describeLiteralMismatch(normalizedActual.element, normalizedExpected.element);
      if (elementMessage) {
        return elementMessage;
      }
      if (Array.isArray(normalizedActual.bounds)) {
        for (const bound of normalizedActual.bounds) {
          const boundMessage = describeLiteralMismatch(bound, normalizedExpected.element);
          if (boundMessage) {
            return boundMessage;
          }
        }
      }
      return null;
    }
    if (normalizedActual.kind === "proc" && normalizedExpected.kind === "proc") {
      return describeLiteralMismatch(normalizedActual.result, normalizedExpected.result);
    }
    if (normalizedActual.kind === "future" && normalizedExpected.kind === "future") {
      return describeLiteralMismatch(normalizedActual.result, normalizedExpected.result);
    }
    if (normalizedActual.kind === "nullable" && normalizedExpected.kind === "nullable") {
      return describeLiteralMismatch(normalizedActual.inner, normalizedExpected.inner);
    }
    if (normalizedActual.kind === "result" && normalizedExpected.kind === "result") {
      return describeLiteralMismatch(normalizedActual.inner, normalizedExpected.inner);
    }
    if (normalizedActual.kind === "union" && normalizedExpected.kind === "union") {
      const count = Math.min(normalizedActual.members.length, normalizedExpected.members.length);
      for (let i = 0; i < count; i += 1) {
        const message = describeLiteralMismatch(normalizedActual.members[i], normalizedExpected.members[i]);
        if (message) {
          return message;
        }
      }
      return null;
    }
    if (normalizedActual.kind === "union") {
      for (const member of normalizedActual.members) {
        const message = describeLiteralMismatch(member, normalizedExpected);
        if (message) {
          return message;
        }
      }
      return null;
    }
    if (normalizedExpected.kind === "union") {
      for (const member of normalizedExpected.members) {
        const message = describeLiteralMismatch(normalizedActual, member);
        if (message) {
          return message;
        }
      }
      return null;
    }
    if (normalizedActual.kind !== "primitive" || normalizedExpected.kind !== "primitive") {
      return null;
    }
    if (!normalizedActual.literal || normalizedActual.literal.literalKind !== "integer" || normalizedActual.literal.explicit) {
      return null;
    }
    if (!hasIntegerBounds(normalizedExpected.name)) {
      return null;
    }
    const bounds = integerBounds(normalizedExpected.name);
    const value = literalValueToBigInt(normalizedActual.literal);
    if (value < bounds.min || value > bounds.max) {
      return `typechecker: literal ${value.toString()} does not fit in ${normalizedExpected.name}`;
    }
    return null;
  }

  function formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string {
    switch (expr.type) {
      case "SimpleTypeExpression":
        return lookupSubstitution(context.getIdentifierName(expr.name), substitutions);
      case "GenericTypeExpression": {
        const base = expr.base ? formatTypeExpression(expr.base, substitutions) : "Unknown";
        const args = Array.isArray(expr.arguments)
          ? expr.arguments
              .map((arg) => (arg ? formatTypeExpression(arg, substitutions) : "Unknown"))
              .filter(Boolean)
          : [];
        return args.length > 0 ? [base, ...args].join(" ") : base;
      }
      case "FunctionTypeExpression": {
        const params = Array.isArray(expr.paramTypes)
          ? expr.paramTypes.map((param) => (param ? formatTypeExpression(param, substitutions) : "Unknown"))
          : [];
        const ret = expr.returnType ? formatTypeExpression(expr.returnType, substitutions) : "void";
        return `fn(${params.join(", ")}) -> ${ret}`;
      }
      case "NullableTypeExpression":
        return `${formatTypeExpression(expr.innerType, substitutions)}?`;
      case "ResultTypeExpression":
        return `Result ${formatTypeExpression(expr.innerType, substitutions)}`;
      case "UnionTypeExpression": {
        const members = Array.isArray(expr.members)
          ? expr.members.map((member) => (member ? formatTypeExpression(member, substitutions) : "Unknown"))
          : [];
        return members.length > 0 ? members.join(" | ") : "Union";
      }
      case "WildcardTypeExpression":
        return "_";
      default:
        return "Unknown";
    }
  }

  function typeExpressionsEquivalent(
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): boolean {
    if (!a && !b) return true;
    if (!a || !b) return false;
    return formatTypeExpression(a, substitutions) === formatTypeExpression(b, substitutions);
  }

  function describeTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): string {
    if (!expr) return "unspecified";
    return formatTypeExpression(expr, substitutions);
  }

  function lookupSubstitution(name: string | null, substitutions?: Map<string, string>): string {
    if (!name) return "Unknown";
    if (substitutions && substitutions.has(name)) {
      return substitutions.get(name) ?? "Unknown";
    }
    return name;
  }

  function describeTypeArgument(type: TypeInfo): string {
    return formatType(type);
  }

  function appendInterfaceArgsToLabel(label: string, args: string[]): string {
    if (!args.length) {
      return label;
    }
    return `${label} ${args.join(" ")}`.trim();
  }

  function getIdentifierNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null {
    if (!expr) return null;
    if (expr.type === "SimpleTypeExpression") {
      return context.getIdentifierName(expr.name);
    }
    if (expr.type === "GenericTypeExpression") {
      return getIdentifierNameFromTypeExpression(expr.base);
    }
    return null;
  }

  return {
    resolveTypeExpression,
    instantiateTypeAlias,
    typeInfosEquivalent,
    isTypeAssignable,
    describeLiteralMismatch,
    canonicalizeStructuralType,
    typeExpressionsEquivalent,
    describeTypeExpression,
    formatTypeExpression,
    lookupSubstitution,
    describeTypeArgument,
    instantiateUnionDefinition,
    appendInterfaceArgsToLabel,
  };
}
