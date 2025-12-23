import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import { callCallableValue } from "../../src/interpreter/functions";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { ModuleLoader } from "../../scripts/module-loader";
import { TypeChecker, V11 } from "../../index";

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
    const result = session.checkModule(mod.module);
    result.diagnostics.forEach((diag) => diagnostics.push(diag.message));
  }
  return diagnostics;
}

describe("v11 interpreter - bitwise hex literal parsing", () => {
  test("hex literal with E digits stays integer for bitwise ops", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-hex-bitwise-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: hex_bitwise\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

fn main() -> i32 {
  0xE0 .& 0xff
}
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics: string[] = [];
      for (const mod of program.modules) {
        const result = session.checkModule(mod.module);
        result.diagnostics.forEach((diag) => diagnostics.push(diag.message));
      }
      expect(diagnostics).toEqual([]);

      const interpreter = new V11.Interpreter();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);

      const pkg = interpreter.packageRegistry.get(program.entry.packageName);
      const mainFn = pkg?.get("main");
      if (!mainFn) throw new Error("entry module missing main");
      const result = callCallableValue(interpreter as any, mainFn as any, [], interpreter.globals) as any;
      expect(result).toEqual({ kind: "i32", value: 0xe0n });
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("octal literal with exponent marker is rejected", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-octal-e-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: octal_e\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

fn main() -> i32 { 0o17e1 }
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics = typecheckProgram(session, program);
      expect(diagnostics.some((diag) => diag.includes("undefined identifier 'e1'"))).toBe(true);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("binary literal with exponent marker is rejected", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-binary-e-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: binary_e\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

fn main() -> i32 { 0b10E2 }
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics = typecheckProgram(session, program);
      expect(diagnostics.some((diag) => diag.includes("undefined identifier 'E2'"))).toBe(true);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
