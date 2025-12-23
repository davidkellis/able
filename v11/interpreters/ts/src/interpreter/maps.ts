import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { makeIntegerFromNumber, numericToNumber } from "./numeric";
import { hashMapEntries, insertHashMapEntry, type HashMapEntry } from "./hash_map_kernel";

function mergeTypeExpr(
  ctx: Interpreter,
  current: AST.TypeExpression | null,
  next: AST.TypeExpression | null,
): AST.TypeExpression | null {
  if (!next) return current;
  if (!current) return next;
  if (current.type === "WildcardTypeExpression") return current;
  if (next.type === "WildcardTypeExpression") return next;
  if (!ctx.typeExpressionsEqual(current, next)) return AST.wildcardTypeExpression();
  return current;
}

function hashMapHandleFromValue(ctx: Interpreter, value: RuntimeValue): number {
  if (value.kind !== "struct_instance" || value.def.id.name !== "HashMap") {
    throw new Error("Map literal spread requires a HashMap value");
  }
  if (!(value.values instanceof Map)) {
    throw new Error("HashMap storage is missing fields");
  }
  const handleValue = value.values.get("handle");
  if (!handleValue) {
    throw new Error("HashMap handle field is missing");
  }
  return Math.trunc(numericToNumber(handleValue, "hash map handle", { requireSafeInteger: true }));
}

function hashMapStructDef(env: Environment): AST.StructDefinition {
  const defVal = env.get("HashMap");
  if (!defVal || defVal.kind !== "struct_def") {
    throw new Error("HashMap struct is not available in scope");
  }
  return defVal.def;
}

export function evaluateMapLiteral(ctx: Interpreter, node: AST.MapLiteral, env: Environment): RuntimeValue {
  const handle = ctx.createHashMapHandle();
  const state = ctx.hashMapStateForHandle(handle);
  let keyType: AST.TypeExpression | null = null;
  let valueType: AST.TypeExpression | null = null;

  for (const entry of node.entries ?? []) {
    if (!entry) continue;
    if (entry.type === "MapLiteralEntry") {
      const keyValue = ctx.evaluate(entry.key, env);
      const value = ctx.evaluate(entry.value, env);
      insertHashMapEntry(state, keyValue, value);
      keyType = mergeTypeExpr(ctx, keyType, ctx.typeExpressionFromValue(keyValue));
      valueType = mergeTypeExpr(ctx, valueType, ctx.typeExpressionFromValue(value));
      continue;
    }
    const spreadValue = ctx.evaluate(entry.expression, env);
    const spreadHandle = hashMapHandleFromValue(ctx, spreadValue);
    const spreadState = ctx.hashMapStateForHandle(spreadHandle);
    for (const existing of hashMapEntries(spreadState)) {
      insertHashMapEntry(state, existing.key, existing.value);
    }
    if (spreadValue.kind === "struct_instance") {
      const typeArgs = spreadValue.typeArguments ?? [];
      if (typeArgs.length >= 2) {
        keyType = mergeTypeExpr(ctx, keyType, typeArgs[0] ?? null);
        valueType = mergeTypeExpr(ctx, valueType, typeArgs[1] ?? null);
      }
    }
  }

  const structDef = hashMapStructDef(env);
  const generics = structDef.genericParams ?? [];
  const typeArguments = generics.length > 0
    ? [
        keyType ?? AST.wildcardTypeExpression(),
        valueType ?? AST.wildcardTypeExpression(),
      ]
    : undefined;
  const typeArgMap = generics.length > 0 && typeArguments
    ? ctx.mapTypeArguments(generics, typeArguments, "instantiating HashMap")
    : undefined;
  const values = new Map<string, RuntimeValue>();
  values.set("handle", makeIntegerFromNumber("i64", handle));

  return {
    kind: "struct_instance",
    def: structDef,
    values,
    typeArguments,
    typeArgMap,
  };
}

export function serializeMapEntries(ctx: Interpreter, mapValue: RuntimeValue): HashMapEntry[] {
  const handle = hashMapHandleFromValue(ctx, mapValue);
  const state = ctx.hashMapStateForHandle(handle);
  return hashMapEntries(state);
}
