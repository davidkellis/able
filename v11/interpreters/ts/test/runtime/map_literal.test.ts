import { describe, expect, test } from "bun:test";
import path from "node:path";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import { serializeMapEntries } from "../../src/interpreter/maps";
import { loadModuleFromPath } from "../../scripts/fixture-utils";

const KERNEL_ENTRY = path.resolve(__dirname, "../../../../kernel/src/kernel.able");

describe("runtime map literals", () => {
  test("evaluates basic map literal", async () => {
    const interpreter = new Interpreter();
    await loadKernelModule(interpreter);
    ensureHashMapStruct(interpreter);
    const mapValue = interpreter.evaluate(
      AST.mapLit([
        AST.mapEntry(AST.stringLiteral("host"), AST.int(1)),
        AST.mapEntry(AST.stringLiteral("port"), AST.int(443)),
      ]),
    );
    expect(mapValue.kind).toBe("struct_instance");
    const entries = toSimpleEntries(interpreter, mapValue);
    expect(entries).toEqual([
      { key: { kind: "String", value: "host" }, value: { kind: "i32", value: 1n } },
      { key: { kind: "String", value: "port" }, value: { kind: "i32", value: 443n } },
    ]);
  });

  test("applies spreads and overrides duplicates", async () => {
    const interpreter = new Interpreter();
    await loadKernelModule(interpreter);
    ensureHashMapStruct(interpreter);
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
    expect(mapValue.kind).toBe("struct_instance");
    const entries = toSimpleEntries(interpreter, mapValue);
    expect(entries).toEqual([
      { key: { kind: "String", value: "content-type" }, value: { kind: "String", value: "application/json" } },
      { key: { kind: "String", value: "accept" }, value: { kind: "String", value: "application/json" } },
      { key: { kind: "String", value: "cache" }, value: { kind: "String", value: "max-age=0" } },
    ]);
  });
});

function ensureHashMapStruct(interpreter: Interpreter): void {
  try {
    const existing = interpreter.globals.get("HashMap");
    if (existing?.kind === "struct_def") return;
  } catch {}
  try {
    const kernelHashMap = interpreter.globals.get("able.kernel.HashMap");
    if (kernelHashMap?.kind === "struct_def") {
      interpreter.globals.define("HashMap", kernelHashMap);
      return;
    }
  } catch {}
  interpreter.evaluate(
    AST.structDefinition(
      "HashMap",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle")],
      "named",
      [AST.genericParameter("K"), AST.genericParameter("V")],
    ),
  );
}

async function loadKernelModule(interpreter: Interpreter): Promise<void> {
  if (interpreter.packageRegistry.has("able.kernel")) {
    return;
  }
  const kernelModule = await loadModuleFromPath(KERNEL_ENTRY);
  kernelModule.package = AST.packageStatement(["able", "kernel"]);
  interpreter.evaluate(kernelModule);
}

function toSimpleEntries(
  interpreter: Interpreter,
  mapValue: any,
): Array<{ key: { kind: string; value: unknown }; value: { kind: string; value: unknown } }> {
  return serializeMapEntries(interpreter, mapValue).map(({ key, value }) => ({
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
