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
  options: { ignoreNonEntryDiagnostics?: boolean } = {},
): string[] {
  const diagnostics: string[] = [];
  for (const mod of program.modules) {
    if (mod.packageName !== program.entry.packageName) {
      const result = session.checkModule(mod.module);
      if (!options.ignoreNonEntryDiagnostics) {
        result.diagnostics.forEach((diag) => {
          const { location } = diag;
          const loc = location ? `${location.path ?? ""}:${location.line ?? "?"}:${location.column ?? "?"}` : "";
          diagnostics.push(loc ? `${diag.message} (${loc})` : diag.message);
        });
      }
    }
  }
  const entryResult = session.checkModule(program.entry.module);
  entryResult.diagnostics.forEach((diag) => {
    const { location } = diag;
    const loc = location ? `${location.path ?? ""}:${location.line ?? "?"}:${location.column ?? "?"}` : "";
    diagnostics.push(loc ? `${diag.message} (${loc})` : diag.message);
  });
  return diagnostics;
}

describe("stdlib-backed bit set helpers", () => {
  test("BitSet operations run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-bit-set-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: bit_set_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.bit_set.*

fn main() -> i32 {
  bits: BitSet := BitSet.new()
  bits.set(0)
  bits.set(1)
  bits.set(64)
  bits.flip(1)
  bits.reset(0)
  bits.set(5)
  bits.set(70)

  total := 0
  if bits.contains(0) { total = total + 100 }
  if bits.contains(5) { total = total + 5 }
  if bits.contains(70) { total = total + 7 }
  if !bits.contains(1) { total = total + 11 }

  for bit in bits {
    total = total + bit
  }
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
      expect(readInteger(result)).toBe(162);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
