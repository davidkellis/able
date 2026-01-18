import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import { callCallableValue } from "../../src/interpreter/functions";
import { ModuleLoader } from "../../scripts/module-loader";
import { collectModuleSearchPaths } from "../../scripts/module-search-paths";

const PROBE_ROOT = path.resolve(__dirname, "../../..");

function evaluateAllModules(interpreter: Interpreter, program: { modules: any[]; entry: any }): void {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    interpreter.evaluate(mod.module);
  }
  interpreter.evaluate(program.entry.module);
}

describe("hash helper builtins", () => {
  test("__able_f32_bits returns IEEE-754 bits", () => {
    const I = new Interpreter();
    const result = I.evaluate(AST.functionCall(AST.identifier("__able_f32_bits"), [AST.floatLiteral(1.5, "f32")])) as any;
    expect(result).toEqual({ kind: "u32", value: 0x3fc00000n });
  });

  test("__able_f64_bits returns IEEE-754 bits", () => {
    const I = new Interpreter();
    const result = I.evaluate(AST.functionCall(AST.identifier("__able_f64_bits"), [AST.floatLiteral(1.5, "f64")])) as any;
    expect(result).toEqual({ kind: "u64", value: 0x3ff8000000000000n });
  });

  test("__able_u64_mul wraps to 64 bits", () => {
    const I = new Interpreter();
    const max = (1n << 64n) - 1n;
    const result = I.evaluate(
      AST.functionCall(AST.identifier("__able_u64_mul"), [
        AST.integerLiteral(max, "u64"),
        AST.integerLiteral(2n, "u64"),
      ]),
    ) as any;
    const expected = (max * 2n) & ((1n << 64n) - 1n);
    expect(result).toEqual({ kind: "u64", value: expected });
  });
});

describe("kernel HashMap dispatch", () => {
  test("custom Hash/Eq impls drive kernel map lookups", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-kernel-hash-map-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: kernel_hash_map\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.kernel.{HashMap, Hash, Eq, Hasher}

struct Key { id: i32 }

impl Hash for Key {
  fn hash(self: Self, hasher: Hasher) -> void {
    hasher.write_i32(self.id)
  }
}

impl Eq for Key {
  fn eq(self: Self, other: Self) -> bool { self.id == other.id }
}

fn main() -> bool {
  map := HashMap.new()
  map.raw_set(Key { id: 1 }, 10)
  map.raw_set(Key { id: 2 }, 20)

  sum := 0
  map.raw_get(Key { id: 1 }) match {
    case value: i32 => { sum = sum + value },
    case _ => {}
  }
  map.raw_get(Key { id: 2 }) match {
    case value: i32 => { sum = sum + value },
    case _ => {}
  }

  sum == 30
}
`.trimStart(),
        "utf8",
      );

      const searchPaths = collectModuleSearchPaths({
        cwd: tmpRoot,
        probeFrom: [PROBE_ROOT],
      });
      const loader = new ModuleLoader(searchPaths);
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const interpreter = new Interpreter();
      evaluateAllModules(interpreter, program);

      const pkg = interpreter.packageRegistry.get(program.entry.packageName);
      const mainFn = pkg?.get("main");
      if (!mainFn) {
        throw new Error("entry module missing main");
      }

      const result = callCallableValue(interpreter as any, mainFn, [], interpreter.globals) as any;
      expect(result).toEqual({ kind: "bool", value: true });
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
