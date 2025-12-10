import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";
import { serializeMapEntries } from "../../src/interpreter/maps";
import type { HashMapValue } from "../../src/interpreter/values";

describe("runtime map literals", () => {
  test("evaluates basic map literal", () => {
    const interpreter = new InterpreterV10();
    const mapValue = interpreter.evaluate(
      AST.mapLit([
        AST.mapEntry(AST.stringLiteral("host"), AST.stringLiteral("api")),
        AST.mapEntry(AST.stringLiteral("port"), AST.int(443)),
      ]),
    );
    expect(mapValue.kind).toBe("hash_map");
    const entries = toSimpleEntries(mapValue as HashMapValue);
    expect(entries).toEqual([
      { key: { kind: "String", value: "host" }, value: { kind: "String", value: "api" } },
      { key: { kind: "String", value: "port" }, value: { kind: "i32", value: 443n } },
    ]);
  });

  test("applies spreads and overrides duplicates", () => {
    const interpreter = new InterpreterV10();
    interpreter.evaluate(
      AST.assign(
        "defaults",
        AST.mapLit([
          AST.mapEntry(AST.stringLiteral("accept"), AST.stringLiteral("application/json")),
          AST.mapEntry(AST.stringLiteral("cache"), AST.stringLiteral("no-store")),
        ]),
      ),
    );
    const mapValue = interpreter.evaluate(
      AST.mapLit([
        AST.mapEntry(AST.stringLiteral("content-type"), AST.stringLiteral("application/json")),
        AST.mapSpread(AST.identifier("defaults")),
        AST.mapEntry(AST.stringLiteral("cache"), AST.stringLiteral("max-age=0")),
      ]),
    );
    expect(mapValue.kind).toBe("hash_map");
    const entries = toSimpleEntries(mapValue as HashMapValue);
    expect(entries).toEqual([
      { key: { kind: "String", value: "content-type" }, value: { kind: "String", value: "application/json" } },
      { key: { kind: "String", value: "accept" }, value: { kind: "String", value: "application/json" } },
      { key: { kind: "String", value: "cache" }, value: { kind: "String", value: "max-age=0" } },
    ]);
  });
});

function toSimpleEntries(map: HashMapValue): Array<{ key: { kind: string; value: unknown }; value: { kind: string; value: unknown } }> {
  return serializeMapEntries(map).map(({ key, value }) => ({
    key: simplifyValue(key),
    value: simplifyValue(value),
  }));
}

function simplifyValue(value: any): { kind: string; value: unknown } {
  if (value.kind === "nil") {
    return { kind: "nil", value: null };
  }
  return { kind: value.kind, value: (value as any).value };
}
