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

describe("stdlib-backed deque helpers", () => {
  test("Deque operations run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-deque-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: deque_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.deque.{Deque}
import able.core.iteration.{IteratorEnd}

fn main() -> i32 {
  deque := Deque.new()
  deque.push_back(2)
  deque.push_front(1)
  deque.push_back(3)
  deque.push_back(4)

  total := deque.len()

  iter := deque.iterator()
  loop {
    iter.next() match {
      case value: i32 => total = total + value,
      case IteratorEnd => { break }
    }
  }

  for value in deque {
    total = total + (value * 2)
  }

  front := deque.pop_front()
  back := deque.pop_back()

  front match {
    case value: i32 => total = total + (value * 10),
    case nil => { total = -999 }
  }
  back match {
    case value: i32 => total = total + (value * 5),
    case nil => { total = -999 }
  }

  total = total + deque.len()
  total
}
`.trimStart(),
        "utf8",
      );

      const searchPaths = collectModuleSearchPaths({
        cwd: tmpRoot,
        probeStdlibFrom: [PROBE_ROOT],
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
      expect(readInteger(result)).toBe(66);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
