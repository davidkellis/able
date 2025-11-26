import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import * as AST from "../../src/ast";
import { ModuleLoader } from "../../scripts/module-loader";
import { collectModuleSearchPaths } from "../../scripts/module-search-paths";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { callCallableValue } from "../../src/interpreter/functions";
import { memberAccessOnValue } from "../../src/interpreter/structs";
import { TypeChecker, V10 } from "../../index";

const PROBE_ROOT = path.resolve(__dirname, "../../..");

const readInteger = (value: any): number => Number(value?.value ?? 0);
const readString = (value: any): string => String(value?.value ?? "");

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

describe("stdlib-backed array helpers", () => {
  test("method sets override native array helpers when defined", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-array-methods-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: array_override\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

methods Array T {
  fn size(self: Self) -> string { "custom-size" }
}

fn main() -> string {
  arr := [1, 2, 3]
  arr.size()
}
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics = typecheckProgram(session, program);
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
      expect(result?.kind).toBe("string");
      expect(readString(result)).toBe("custom-size");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("stdlib Array methods run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-array-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: array_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.array.{Array}

fn double(value: i32) -> i32 { value * 2 }

fn main() -> i32 {
  values := Array.new()
  values.push(1)
  values.push(2)
  extras := Array.new()
  extras.push(3)
  values.push_all(extras)
  baseline := values.len()
  if values.is_empty() {
    values.push(99)
  }

  count := values.len()
  if !values.is_empty() {
    count = count + 10
  }

  count + baseline
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
      expect(readInteger(result)).toBe(16);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("pop on empty array returns nil via stdlib", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-array-pop-empty-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: array_pop_empty\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.array.{Array}

fn main() {
  values := Array.new()
  values.pop()
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
      expect(result?.kind).toBe("nil");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("bounds errors route through stdlib IndexError", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-array-index-error-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: array_index_error\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.array.{Array}
import able.core.errors.{IndexError}

fn main() {
  values := Array.new()
  values.push(1)
  values.set(5, 9)
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
      expect(result?.kind).toBe("struct_instance");
      const message = callCallableValue(
        interpreter as any,
        memberAccessOnValue(interpreter as any, result, AST.identifier("message"), interpreter.globals) as any,
        [],
        interpreter.globals,
      );
      expect(readString(message)).toBe("index 5 out of bounds for length 1");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("get out of bounds returns nil via stdlib", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-array-get-oob-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: array_get_oob\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.array.{Array}

fn main() {
  values := Array.new()
  values.push(1)
  values.get(10)
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
      expect(result?.kind).toBe("nil");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("negative index surfaces stdlib IndexError", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-array-index-negative-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: array_index_negative\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.collections.array.{Array}
import able.core.errors.{IndexError}

fn main() {
  values := Array.new()
  values.push(1)
  values.set(-1, 9)
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
      expect(result?.kind).toBe("struct_instance");
      const message = callCallableValue(
        interpreter as any,
        memberAccessOnValue(interpreter as any, result, AST.identifier("message"), interpreter.globals) as any,
        [],
        interpreter.globals,
      );
      expect(readString(message)).toBe("index -1 out of bounds for length 1");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
