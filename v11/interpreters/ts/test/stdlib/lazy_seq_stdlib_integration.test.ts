import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import { ModuleLoader } from "../../scripts/module-loader";
import { collectModuleSearchPaths } from "../../scripts/module-search-paths";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { callCallableValue } from "../../src/interpreter/functions";
import { TypeChecker, V11 } from "../../index";

const PROBE_ROOT = path.resolve(__dirname, "../../..");

const readInteger = (value: any): number => Number(value?.value ?? 0);

function evaluateAllModules(interpreter: V11.Interpreter, program: { modules: any[]; entry: any }): void {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    interpreter.evaluate(mod.module);
  }
  interpreter.evaluate(program.entry.module);
}

function typecheckProgram(
  session: TypeChecker.TypecheckerSession,
  program: { modules: any[]; entry: any },
): string[] {
  const diagnostics: string[] = [];
  for (const mod of program.modules) {
    if (mod.packageName !== program.entry.packageName) {
      const result = session.checkModule(mod.module);
      result.diagnostics.forEach((diag) => diagnostics.push(diag.message));
    }
  }
  const entryResult = session.checkModule(program.entry.module);
  entryResult.diagnostics.forEach((diag) => diagnostics.push(diag.message));
  return diagnostics;
}

describe("stdlib-backed lazy seq helpers", () => {
  test("LazySeq operations run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-lazy-seq-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: lazy_seq_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.kernel.*
import able.collections.array.*
import able.collections.lazy_seq.*
import able.core.iteration.{Iterator, IteratorEnd}

struct RecordingIterator {
  values: Array i32,
  index: i32,
  pulls: Array i32
}

impl Iterator i32 for RecordingIterator {
  fn next(self: Self) -> i32 | IteratorEnd {
    if self.index >= self.values.len() {
      return IteratorEnd {}
    }
    self.pulls.push(self.index)
    value := self.values.read_slot(self.index)
    self.index = self.index + 1
    value
  }
}

fn make_iterator(values: Array i32, pulls: Array i32) -> RecordingIterator {
  RecordingIterator { values: values, index: 0, pulls: pulls }
}

fn main() -> i32 {
  values := Array.new()
  values.push(2)
  values.push(4)
  values.push(6)
  pulls := Array.new()

  seq := LazySeq.from_iterator(make_iterator(values, pulls))
  score := 0

  seq.get(0) match {
    case value: i32 => score = score + value,
    case nil => { return -400 }
  }
  if pulls.len() == 1 { score = score + 10 } else { return -401 }

  seq.get(0) match {
    case _: i32 => {},
    case nil => { return -402 }
  }
  if pulls.len() == 1 { score = score + 10 } else { return -403 }

  seq.get(1) match {
    case value: i32 => score = score + value,
    case nil => { return -404 }
  }
  if pulls.len() == 2 { score = score + 10 } else { return -405 }

  taken := seq.take(2)
  score = score + taken.len()
  if pulls.len() != 2 { return -406 }

  arr := seq.to_array()
  arr_sum := 0
  for value in arr {
    arr_sum = arr_sum + value
  }
  score = score + arr_sum
  if pulls.len() == 3 { score = score + 10 } else { return -407 }

  score = score + arr.len()
  if pulls.len() == 3 { score = score + 10 } else { return -408 }

  score
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
      const diagnostics = typecheckProgram(session, program);
      expect(diagnostics).toEqual([]);

      const interpreter = new V11.Interpreter();
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
      expect(readInteger(result)).toBe(73);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
