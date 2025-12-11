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

describe("stdlib-backed queue helpers", () => {
  test("Queue operations run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-queue-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: queue_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.queue.{Queue}
import able.core.iteration.{IteratorEnd}

fn main() -> i32 {
  queue := Queue.new()
  queue.enqueue(10)
  queue.enqueue(20)
  queue.enqueue(30)

  total := 0
  queue.peek() match {
    case value: i32 => total = total + value,
    case nil => { total = -500 }
  }

  iter := queue.iterator()
  iter_sum := 0
  loop {
    iter.next() match {
      case value: i32 => iter_sum = iter_sum + value,
      case IteratorEnd => { break }
    }
  }
  total = total + iter_sum

  sum_dequeue := 0
  queue.dequeue() match { case value: i32 => sum_dequeue = sum_dequeue + value, case nil => { sum_dequeue = -100 } }
  queue.dequeue() match { case value: i32 => sum_dequeue = sum_dequeue + value, case nil => { sum_dequeue = -100 } }
  queue.dequeue() match { case value: i32 => sum_dequeue = sum_dequeue + value, case nil => { sum_dequeue = -100 } }
  total = total + sum_dequeue

  queue.dequeue() match {
    case nil => total = total + 1,
    case _ => { total = -500 }
  }

  queue.enqueue(5)
  queue.enqueue(7)

  for value in queue {
    total = total + (value * 3)
  }

  total = total + queue.size()
  total
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
      expect(readInteger(result)).toBe(169);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
