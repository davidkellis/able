import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { HashMapEntry, HashMapValue, V10Value } from "./values";
import { isFloatValue, isIntegerValue } from "./numeric";

function createHashMap(): HashMapValue {
  return { kind: "hash_map", entries: new Map(), order: [] };
}

function keyLabel(value: V10Value): string {
  switch (value.kind) {
    case "String":
      return `s:${value.value}`;
    case "bool":
      return `b:${value.value ? 1 : 0}`;
    case "char":
      return `c:${value.value}`;
    case "nil":
      return "n:";
    default:
      if (isIntegerValue(value)) {
        return `i:${value.value.toString()}`;
      }
      if (isFloatValue(value)) {
        if (!Number.isFinite(value.value)) {
          throw new Error("Map literal keys cannot be NaN or infinite numbers");
        }
        return `f:${value.value}`;
      }
      throw new Error("Map literal keys must be primitives (String, bool, char, nil, numeric)");
  }
}

function insertEntry(target: HashMapValue, key: V10Value, value: V10Value): void {
  const label = keyLabel(key);
  if (target.entries.has(label)) {
    target.entries.set(label, { key, value });
    return;
  }
  target.entries.set(label, { key, value });
  target.order.push(label);
}

function copyEntries(map: HashMapValue): HashMapEntry[] {
  return map.order.map((label) => {
    const entry = map.entries.get(label);
    if (!entry) {
      throw new Error("Map entry missing during spread");
    }
    return entry;
  });
}

export function evaluateMapLiteral(ctx: InterpreterV10, node: AST.MapLiteral, env: Environment): HashMapValue {
  const map = createHashMap();
  if (!node.entries) {
    return map;
  }
  for (const entry of node.entries) {
    if (!entry) {
      continue;
    }
    if (entry.type === "MapLiteralEntry") {
      const keyValue = ctx.evaluate(entry.key, env);
      const value = ctx.evaluate(entry.value, env);
      insertEntry(map, keyValue, value);
      continue;
    }
    const spreadValue = ctx.evaluate(entry.expression, env);
    if (spreadValue.kind !== "hash_map") {
      throw new Error("Map literal spread requires a map value");
    }
    for (const existing of copyEntries(spreadValue)) {
      insertEntry(map, existing.key, existing.value);
    }
  }
  return map;
}

export function serializeMapEntries(map: HashMapValue): HashMapEntry[] {
  return copyEntries(map);
}
