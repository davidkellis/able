import * as AST from "../../ast";
import type { TypeInfo } from "../types";

export function applyGenericTypeExpression(
  base: AST.TypeExpression,
  args: AST.TypeExpression[],
): AST.TypeExpression {
  if (args.length === 0) return base;
  if (base.type === "NullableTypeExpression") {
    return AST.nullableTypeExpression(applyGenericTypeExpression(base.innerType, args));
  }
  if (base.type === "ResultTypeExpression") {
    return AST.resultTypeExpression(applyGenericTypeExpression(base.innerType, args));
  }
  let flattenedBase = base;
  let flattenedArgs = args;
  if (base.type === "GenericTypeExpression") {
    const collected: AST.TypeExpression[] = [];
    let current: AST.TypeExpression = base;
    while (current.type === "GenericTypeExpression") {
      if (current.arguments && current.arguments.length > 0) {
        collected.unshift(...current.arguments);
      }
      current = current.base;
    }
    flattenedBase = current;
    flattenedArgs = [...collected, ...args];
  }
  return AST.genericTypeExpression(flattenedBase, flattenedArgs);
}

export function typeInfoToTypeExpression(type: TypeInfo | undefined): AST.TypeExpression | null {
  if (!type) return null;
  switch (type.kind) {
    case "primitive":
      return AST.simpleTypeExpression(type.name);
    case "type_parameter":
      return AST.simpleTypeExpression(type.name);
    case "struct": {
      const base = AST.simpleTypeExpression(type.name);
      const args =
        Array.isArray(type.typeArguments) && type.typeArguments.length > 0
          ? type.typeArguments.map((arg) => typeInfoToTypeExpression(arg) ?? AST.wildcardTypeExpression())
          : undefined;
      return args && args.length > 0 ? AST.genericTypeExpression(base, args) : base;
    }
    case "interface": {
      const base = AST.simpleTypeExpression(type.name);
      const args =
        Array.isArray(type.typeArguments) && type.typeArguments.length > 0
          ? type.typeArguments.map((arg) => typeInfoToTypeExpression(arg) ?? AST.wildcardTypeExpression())
          : undefined;
      return args && args.length > 0 ? AST.genericTypeExpression(base, args) : base;
    }
    case "array":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Array"),
        [typeInfoToTypeExpression(type.element) ?? AST.wildcardTypeExpression()],
      );
    case "map":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Map"),
        [
          typeInfoToTypeExpression(type.key) ?? AST.wildcardTypeExpression(),
          typeInfoToTypeExpression(type.value) ?? AST.wildcardTypeExpression(),
        ],
      );
    case "range":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Range"),
        [typeInfoToTypeExpression(type.element) ?? AST.wildcardTypeExpression()],
      );
    case "iterator":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Iterator"),
        [typeInfoToTypeExpression(type.element) ?? AST.wildcardTypeExpression()],
      );
    case "future":
      return AST.genericTypeExpression(
        AST.simpleTypeExpression("Future"),
        [typeInfoToTypeExpression(type.result) ?? AST.wildcardTypeExpression()],
      );
    case "nullable":
      return AST.nullableTypeExpression(typeInfoToTypeExpression(type.inner) ?? AST.wildcardTypeExpression());
    case "result":
      return AST.resultTypeExpression(typeInfoToTypeExpression(type.inner) ?? AST.wildcardTypeExpression());
    case "union":
      return AST.unionTypeExpression(
        type.members.map((member) => typeInfoToTypeExpression(member) ?? AST.wildcardTypeExpression()),
      );
    case "function": {
      const params = (type.parameters ?? []).map((param) => typeInfoToTypeExpression(param) ?? AST.wildcardTypeExpression());
      const returnType = typeInfoToTypeExpression(type.returnType) ?? AST.wildcardTypeExpression();
      return AST.functionTypeExpression(params, returnType);
    }
    default:
      return AST.wildcardTypeExpression();
  }
}
