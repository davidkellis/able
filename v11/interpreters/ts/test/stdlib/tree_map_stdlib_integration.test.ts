import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import { ModuleLoader } from "../../scripts/module-loader";
import { collectModuleSearchPaths } from "../../scripts/module-search-paths";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { callCallableValue } from "../../src/interpreter/functions";
import { TypeChecker, V10 } from "../../index";

const PROBE_ROOT = path.resolve(__dirname, "../../..");

const readInteger = (value: any): number => Number(value?.value ?? 0);

function evaluateAllModules(interpreter: V10.InterpreterV10, program: { modules: any[]; entry: any }): void {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    interpreter.evaluate(mod.module);
  }
  interpreter.evaluate(program.entry.module);
}

function typecheckProgram(
  session: TypeChecker.TypecheckerSession,
  program: { modules: any[]; entry: any },
  options: { ignoreNonEntryDiagnostics?: boolean } = {},
): string[] {
  const diagnostics: string[] = [];
  for (const mod of program.modules) {
    if (mod.packageName !== program.entry.packageName) {
      const result = session.checkModule(mod.module);
      if (!options.ignoreNonEntryDiagnostics) {
        result.diagnostics.forEach((diag) => diagnostics.push(diag.message));
      }
    }
  }
  const entryResult = session.checkModule(program.entry.module);
  entryResult.diagnostics.forEach((diag) => diagnostics.push(diag.message));
  return diagnostics;
}

describe("stdlib-backed tree map helpers", () => {
  test("TreeMap operations run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-tree-map-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: tree_map_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.tree_map.{TreeMap, TreeEntry}

fn main() -> i32 {
  map: TreeMap String i32 := TreeMap.new()
  map.set("beta", 2)
  map.set("alpha", 1)
  map.set("gamma", 3)
  map.set("alpha", 4) ## overwrite one entry

  iter_sum := 0
  values := map.values()
  idx := 0
  while idx < values.len() {
    values.get(idx) match {
      case nil => {},
      case value: i32 => { iter_sum = iter_sum + value }
    }
    idx = idx + 1
  }

  first_value := 0
  last_value := 0
  map.first() match {
    case nil => {},
    case entry => { first_value = entry.value }
  }
  map.last() match {
    case nil => {},
    case entry => { last_value = entry.value }
  }

  removed := 0
  if map.remove("beta") == true { removed = 5 }
  missing := if map.contains("missing") == true { 100 } else { 1 }

  iter_sum + first_value + last_value + removed + map.len() + missing
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

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics = typecheckProgram(session, program, { ignoreNonEntryDiagnostics: true });
      expect(diagnostics).toEqual([]);

      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);

      const pkg = interpreter.packageRegistry.get(program.entry.packageName);
      const mainFn = pkg?.get("main");
      if (!mainFn) {
        throw new Error("entry module missing main");
      }

      const result = callCallableValue(interpreter as any, mainFn, [], interpreter.globals) as any;
      expect(result?.kind).toBe("i32");
      expect(readInteger(result)).toBe(24);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("TreeMap handles custom Ord keys", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-tree-map-custom-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: tree_map_custom\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.tree_map.{TreeMap}
import able.core.interfaces.{Ordering}

struct Key {
  raw: String
}

methods Key {
  fn cmp(self: Self, other: Key) -> Ordering { self.raw.cmp(other.raw) }
  fn clone(self: Self) -> Key { Key { raw: self.raw } }
}

fn main() -> String {
  map: TreeMap Key i32 := TreeMap.new()
  map.set(Key { raw: "delta" }, 4)
  map.set(Key { raw: "alpha" }, 1)
  map.set(Key { raw: "charlie" }, 3)
  map.set(Key { raw: "beta" }, 2)

  keys := map.keys()
  out := ""
  idx := 0
  while idx < keys.len() {
    keys.get(idx) match {
      case nil => {},
      case key => { out = out + key.raw + ";" }
    }
    idx = idx + 1
  }
  out
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

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics = typecheckProgram(session, program, { ignoreNonEntryDiagnostics: true });
      expect(diagnostics).toEqual([]);

      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);

      const pkg = interpreter.packageRegistry.get(program.entry.packageName);
      const mainFn = pkg?.get("main");
      if (!mainFn) {
        throw new Error("entry module missing main");
      }

      const result = callCallableValue(interpreter as any, mainFn, [], interpreter.globals) as any;
      expect(result?.kind).toBe("String");
      expect(result.value).toBe("alpha;beta;charlie;delta;");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
