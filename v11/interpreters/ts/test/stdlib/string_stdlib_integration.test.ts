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
const readArrayStrings = (value: any): string[] => {
  if (!value || value.kind !== "array" || !Array.isArray(value.elements)) return [];
  return value.elements.map(readString);
};
const collectStrings = (interp: V10.InterpreterV10, value: any): string[] => {
  if (!value) return [];
  if (value.kind === "array" && Array.isArray(value.elements)) return value.elements.map(readString);
  if (value.kind === "struct_instance" && value.def?.id?.name === "Array") {
    const sizeCall = memberAccessOnValue(interp as any, value, AST.identifier("size"), interp.globals, { preferMethods: true });
    const lenVal = callCallableValue(interp as any, sizeCall as any, [], interp.globals) as any;
    const len = readInteger(lenVal);
    const items: string[] = [];
    for (let i = 0; i < len; i++) {
      const getCall = memberAccessOnValue(interp as any, value, AST.identifier("get"), interp.globals, { preferMethods: true });
      const elemVal = callCallableValue(interp as any, getCall as any, [{ kind: "i32", value: BigInt(i) }], interp.globals) as any;
      if (elemVal && elemVal.kind !== "nil") items.push(readString(elemVal));
    }
    return items;
  }
  return [];
};
const orderingTag = (value: any): string => {
  if (!value) return "unknown";
  if (value.kind === "struct_instance") return value.def?.id?.name ?? "struct_instance";
  if (value.kind === "struct_def") return value.def?.id?.name ?? "struct_def";
  if (value.kind === "interface_value") return orderingTag(value.value);
  return value.kind ?? "unknown";
};

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

describe("stdlib-backed string helpers", () => {
  test("method sets override native string helpers when defined", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-methods-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_override\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

methods string {
  fn len_bytes(self: Self) -> string { "custom-len" }
}

fn main() -> string {
  "hello".len_bytes()
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
      expect(readString(result)).toBe("custom-len");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("stdlib string methods run via ModuleLoader", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> i32 {
  parts := "one|two|three".split("|")
  parts.len()
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
      expect(readInteger(result)).toBe(3);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("substring out of range surfaces stdlib RangeError", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-range-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_range_error\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string
import able.core.errors.{RangeError}

fn main() {
  "hey".substring(10, nil)
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
      expect(orderingTag(result)).toBe("RangeError");
      const message = callCallableValue(
        interpreter as any,
        memberAccessOnValue(interpreter as any, result, AST.identifier("message"), interpreter.globals, {
          preferMethods: true,
        }) as any,
        [],
        interpreter.globals,
      );
      expect(readString(message)).toBe("substring start out of range");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("substring length overflow surfaces stdlib RangeError", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-range-len-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_range_len_error\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string
import able.core.errors.{RangeError}

fn main() {
  "hi".substring(0, 10)
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
      expect(orderingTag(result)).toBe("RangeError");
      const message = callCallableValue(
        interpreter as any,
        memberAccessOnValue(interpreter as any, result, AST.identifier("message"), interpreter.globals, {
          preferMethods: true,
        }) as any,
        [],
        interpreter.globals,
      );
      expect(readString(message)).toBe("substring range out of bounds");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("split with empty delimiter emits grapheme slices", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-split-empty-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_split_empty\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> Array string {
  "abc".split("")
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
      expect(collectStrings(interpreter, result)).toEqual(["a", "b", "c"]);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("replace with empty needle returns the receiver", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-replace-empty-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_replace_empty\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> string {
  "foobar".replace("", "X")
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
      expect(readString(result)).toBe("foobar");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("split with missing delimiter returns original string", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-split-missing-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_split_missing\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> Array string {
  "abc".split("|")
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
      expect(collectStrings(interpreter, result)).toEqual(["abc"]);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("replace with missing needle returns original string", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-replace-missing-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_replace_missing\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> string {
  "abc".replace("zzz", "x")
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
      expect(readString(result)).toBe("abc");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("split with multi-byte delimiter respects code points", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-split-utf8-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_split_utf8\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> Array string {
  "cafébar".split("é")
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
      expect(collectStrings(interpreter, result)).toEqual(["caf", "bar"]);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("replace with multi-byte needle swaps correctly", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-replace-utf8-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_replace_utf8\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> string {
  "abaéaba".replace("é", "δ")
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
      expect(readString(result)).toBe("abaδaba");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("for-loops over strings iterate bytes via stdlib iterator", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-iter-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_iterator\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.string

fn main() -> i32 {
  count := 0
  for b: u8 in "hey" {
    _ = b
    count = count + 1
  }
  count
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
      expect(readInteger(result)).toBe(3);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("Ord.cmp for strings returns stable ordering", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-string-ord-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: string_ord_cmp\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.core.interfaces
import able.core.interfaces.{Less, Equal, Greater}

fn cmp_lt() { "a".cmp("b") }
fn cmp_eq() { "mid".cmp("mid") }
fn cmp_gt() { "z".cmp("m") }
fn cmp_label(a: string, b: string) -> string {
  cmp := a.cmp(b)
  if cmp == Less { "less" }
  else {
    if cmp == Greater { "greater" }
    else { "equal" }
  }
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
      if (!pkg) {
        throw new Error("entry module missing package");
      }
      const cmpLt = pkg.get("cmp_lt");
      const cmpEq = pkg.get("cmp_eq");
      const cmpGt = pkg.get("cmp_gt");
      const cmpLabel = pkg.get("cmp_label");
      if (!cmpLt || !cmpEq || !cmpGt || !cmpLabel) {
        throw new Error("comparison helpers missing");
      }

      const lt = callCallableValue(interpreter as any, cmpLt as any, [], interpreter.globals) as any;
      const eq = callCallableValue(interpreter as any, cmpEq as any, [], interpreter.globals) as any;
      const gt = callCallableValue(interpreter as any, cmpGt as any, [], interpreter.globals) as any;

      expect(orderingTag(lt)).toBe("Less");
      expect(orderingTag(eq)).toBe("Equal");
      expect(orderingTag(gt)).toBe("Greater");

      const lessLabel = callCallableValue(
        interpreter as any,
        cmpLabel as any,
        [
          { kind: "string", value: "a" },
          { kind: "string", value: "b" },
        ],
        interpreter.globals,
      ) as any;
      const equalLabel = callCallableValue(
        interpreter as any,
        cmpLabel as any,
        [
          { kind: "string", value: "mid" },
          { kind: "string", value: "mid" },
        ],
        interpreter.globals,
      ) as any;
      const greaterLabel = callCallableValue(
        interpreter as any,
        cmpLabel as any,
        [
          { kind: "string", value: "z" },
          { kind: "string", value: "m" },
        ],
        interpreter.globals,
      ) as any;

      expect(readString(lessLabel)).toBe("less");
      expect(readString(equalLabel)).toBe("equal");
      expect(readString(greaterLabel)).toBe("greater");

      const cmpViaInterface = (receiver: any, other: string) => {
        const iface = (interpreter as any).toInterfaceValue("Ord", receiver);
        const method = memberAccessOnValue(
          interpreter as any,
          iface as any,
          AST.identifier("cmp"),
          interpreter.globals,
          { preferMethods: true },
        );
        return callCallableValue(interpreter as any, method as any, [{ kind: "string", value: other }], interpreter.globals);
      };

      const ltIface = cmpViaInterface({ kind: "string", value: "a" }, "b");
      const eqIface = cmpViaInterface({ kind: "string", value: "mid" }, "mid");
      const gtIface = cmpViaInterface({ kind: "string", value: "z" }, "m");

      expect(orderingTag(ltIface)).toBe("Less");
      expect(orderingTag(eqIface)).toBe("Equal");
      expect(orderingTag(gtIface)).toBe("Greater");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
