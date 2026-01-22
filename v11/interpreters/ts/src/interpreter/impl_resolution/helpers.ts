import * as AST from "../../ast";
import type { Interpreter } from "../index";
import type { ImplMethodEntry } from "../values";

const INTEGER_TYPES = new Set([
  "i8", "i16", "i32", "i64", "i128",
  "u8", "u16", "u32", "u64", "u128",
]);

const FLOAT_TYPES = new Set(["f32", "f64"]);

export function isPrimitiveTypeName(name: string): boolean {
  if (name === "bool" || name === "String" || name === "IoHandle" || name === "ProcHandle" || name === "char" || name === "nil" || name === "void") {
    return true;
  }
  return INTEGER_TYPES.has(name) || FLOAT_TYPES.has(name);
}

export function isKnownTypeName(interp: Interpreter, name: string): boolean {
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

export function collectSelfPatternPlaceholderNames(
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

export function primitiveImplementsInterfaceMethod(typeName: string, ifaceName: string, methodName: string): boolean {
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

export function typeExpressionFromInfo(name: string, typeArgs?: AST.TypeExpression[]): AST.TypeExpression {
  const base: AST.SimpleTypeExpression = { type: "SimpleTypeExpression", name: AST.identifier(name) };
  if (!typeArgs || typeArgs.length === 0) return base;
  return { type: "GenericTypeExpression", base, arguments: typeArgs };
}

export function interfaceInfoFromTypeExpression(expr: AST.TypeExpression | null | undefined): { name: string; args?: AST.TypeExpression[] } | null {
  if (!expr) return null;
  if (expr.type === "SimpleTypeExpression") {
    return { name: expr.name.name };
  }
  if (expr.type === "GenericTypeExpression" && expr.base.type === "SimpleTypeExpression") {
    return { name: expr.base.name.name, args: expr.arguments ?? [] };
  }
  return null;
}

export function collectImplGenericNames(interp: Interpreter, entry: ImplMethodEntry): Set<string> {
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

export function typeExpressionUsesGenerics(expr: AST.TypeExpression | undefined, genericNames: Set<string>): boolean {
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
