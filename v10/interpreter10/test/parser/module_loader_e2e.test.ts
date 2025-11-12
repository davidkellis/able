import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import { ModuleLoader, type Program as LoadedProgram } from "../../scripts/module-loader";
import { discoverRoot, indexSourceFiles } from "../../scripts/module-utils";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { callCallableValue } from "../../src/interpreter/functions";
import { TypeChecker, V10 } from "../../index";
import { getTreeSitterParser } from "../../src/parser/tree-sitter-loader";
import { format } from "node:util";

describe("ModuleLoader end-to-end", () => {
  test("parses and runs a multi-file package from source", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-module-loader-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: sample\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import sample.shared.{greet}

fn main() -> void {
  print(greet("Able"))
}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "shared.able"),
        `
package shared

fn greet(name: string) -> string {
  \`hello \${name}\`
}
`.trimStart(),
        "utf8",
      );

      const parser = await getTreeSitterParser();
      const mainSource = await fs.readFile(path.join(tmpRoot, "main.able"), "utf8");
      const parsed = parser.parse(mainSource);
      if (parsed.rootNode.hasError) {
        throw new Error(`tree-sitter reported syntax errors for main.able: ${parsed.rootNode.toString()}`);
      }

      const loader = new ModuleLoader();
      const entryPath = path.join(tmpRoot, "main.able");
      const program = await loader.load(entryPath);

      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);

      evaluateAllModules(interpreter, program);

      const entryPackage = interpreter.packageRegistry.get(program.entry.packageName);
      if (!entryPackage) {
        throw new Error(`entry package '${program.entry.packageName}' not registered`);
      }
      const mainValue = entryPackage.get("main");
      if (!mainValue) {
        throw new Error("entry package missing main function");
      }

      const observed: string[] = [];
      const originalLog = console.log;
      console.log = (...args: unknown[]) => {
        observed.push(args.map((arg) => String(arg)).join(" "));
      };
      try {
        callCallableValue(interpreter as any, mainValue, [], interpreter.globals);
      } finally {
        console.log = originalLog;
      }

      expect(observed).toEqual(["hello Able"]);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});

describe("ModuleLoader alias imports & privacy", () => {
  test("alias imports expose only public members", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-module-loader-alias-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: sample\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "helpers.able"),
        `
package helpers

private fn secret() -> string {
  "hidden"
}

fn greet(name: string) -> string {
  \`hi \${name}\`
}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import sample.helpers as Helpers

fn call_secret() -> string {
  Helpers.secret()
}

fn main() -> void {
  print(Helpers.greet("Able"))
}
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);

      const entryPackage = interpreter.packageRegistry.get(program.entry.packageName);
      if (!entryPackage) throw new Error(`entry package '${program.entry.packageName}' not found`);
      const mainFn = entryPackage.get("main");
      if (!mainFn) throw new Error("expected main function");

      const logs: string[] = [];
      const originalLog = console.log;
      console.log = (...args: unknown[]) => {
        logs.push(args.map((arg) => format("%s", arg)).join(" "));
      };
      try {
        callCallableValue(interpreter as any, mainFn, [], interpreter.globals);
      } finally {
        console.log = originalLog;
      }
      expect(logs).toEqual(["hi Able"]);

      const callSecretFn = entryPackage.get("call_secret");
      if (!callSecretFn) throw new Error("expected call_secret helper");
      let observedError: unknown;
      try {
        callCallableValue(interpreter as any, callSecretFn, [], interpreter.globals);
      } catch (err) {
        observedError = err;
      }
      expect(observedError).toBeInstanceOf(Error);
      expect((observedError as Error).message).toContain("No public member 'secret'");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});

describe("ModuleLoader dynimport scenarios", () => {
  test("dynimport exposes public members via dyn packages", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-module-loader-dyn-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: sample\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "shared.able"),
        `
package shared

fn greet(name: string) -> string {
  \`hey \${name}\`
}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "caller.able"),
        `
package caller

dynimport sample.shared.{greet}

fn call(name: string) -> string {
  greet(name)
}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import sample.caller.{call}

fn main() -> void {
  print(call("Able"))
}
`.trimStart(),
        "utf8",
      );

      const entryPath = path.join(tmpRoot, "main.able");
      const { rootDir, rootName } = await discoverRoot(entryPath);
      const { packages } = await indexSourceFiles(rootDir, rootName);
      const packageNames = [...packages.keys()];
      const sharedPackageName = packageNames.find((name) => name.endsWith(".shared"));
      if (!sharedPackageName) {
        throw new Error(`shared package missing for dynimport test (found: ${packageNames.join(", ")})`);
      }

      const loader = new ModuleLoader();
      const program = await loader.load(entryPath, { includePackages: [sharedPackageName] });
      expect(program.modules.some((mod) => mod.packageName === sharedPackageName)).toBe(true);
      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);
      expect(interpreter.packageRegistry.has(sharedPackageName)).toBe(true);

      const entryPackage = interpreter.packageRegistry.get(program.entry.packageName);
      if (!entryPackage) throw new Error("entry package missing");
      const mainFn = entryPackage.get("main");
      if (!mainFn) throw new Error("main function missing");
      const logs: string[] = [];
      const originalLog = console.log;
      console.log = (...args: unknown[]) => logs.push(args.map((arg) => format("%s", arg)).join(" "));
      try {
        callCallableValue(interpreter as any, mainFn, [], interpreter.globals);
      } finally {
        console.log = originalLog;
      }
      expect(logs).toEqual(["hey Able"]);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("dynimport rejects private members", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-module-loader-dyn-private-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: sample\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "shared.able"),
        `
package shared

private fn secret() -> string {
  "nope"
}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

dynimport sample.shared.{secret}

fn main() -> void {
  print(secret())
}
`.trimStart(),
        "utf8",
      );

      const entryPath = path.join(tmpRoot, "main.able");
      const { rootDir, rootName } = await discoverRoot(entryPath);
      const { packages } = await indexSourceFiles(rootDir, rootName);
      const packageNames = [...packages.keys()];
      const sharedPackageName = packageNames.find((name) => name.endsWith(".shared"));
      if (!sharedPackageName) {
        throw new Error(
          `shared package missing for dynimport privacy test (found: ${packageNames.join(", ")})`,
        );
      }

      const loader = new ModuleLoader();
      const program = await loader.load(entryPath, { includePackages: [sharedPackageName] });
      expect(program.modules.some((mod) => mod.packageName === sharedPackageName)).toBe(true);
      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);

      let caught: unknown;
      try {
        evaluateAllModules(interpreter, program);
      } catch (err) {
        caught = err;
      }
      expect(interpreter.packageRegistry.has(sharedPackageName)).toBe(true);
      expect(caught).toBeInstanceOf(Error);
      expect((caught as Error).message).toContain("function 'secret' is private");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});

describe("ModuleLoader pipeline with typechecker", () => {
  test("typechecks and evaluates parsed modules from source", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-module-loader-pipeline-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: sample\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "shared.able"),
        `
package shared

fn welcome(name: string) -> string {
  \`welcome \${name}\`
}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import sample.shared.{welcome}

fn main() -> void {
  print(welcome("Able"))
}
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const entryPath = path.join(tmpRoot, "main.able");
      const program = await loader.load(entryPath);

      const session = new TypeChecker.TypecheckerSession();
      const diagnostics: string[] = [];
      for (const mod of program.modules) {
        const result = session.checkModule(mod.module);
        for (const diag of result.diagnostics) {
          diagnostics.push(`${mod.packageName}: ${diag.message}`);
        }
      }
      expect(diagnostics).toEqual([]);

      const interpreter = new V10.InterpreterV10();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);

      const entryPackage = interpreter.packageRegistry.get(program.entry.packageName);
      if (!entryPackage) {
        throw new Error(`entry package ${program.entry.packageName} missing`);
      }
      const mainFn = entryPackage.get("main");
      if (!mainFn) {
        throw new Error("entry module missing main");
      }

      const observed: string[] = [];
      const originalLog = console.log;
      console.log = (...args: unknown[]) => {
        observed.push(args.map((arg) => format("%s", arg)).join(" "));
      };
      try {
        callCallableValue(interpreter as any, mainFn, [], interpreter.globals);
      } finally {
        console.log = originalLog;
      }
      expect(observed).toEqual(["welcome Able"]);
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });

  test("reports typechecker diagnostics for invalid modules", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-module-loader-diag-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: diag_demo\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import diag_demo.support.{missing_symbol}

fn main() -> void {}
`.trimStart(),
        "utf8",
      );
      await fs.writeFile(
        path.join(tmpRoot, "support.able"),
        `
package support

fn available() -> void {}
`.trimStart(),
        "utf8",
      );

      const loader = new ModuleLoader();
      const program = await loader.load(path.join(tmpRoot, "main.able"));

      const session = new TypeChecker.TypecheckerSession();
      for (const mod of program.modules) {
        if (mod.packageName !== program.entry.packageName) {
          session.checkModule(mod.module);
        }
      }
      const result = session.checkModule(program.entry.module);
      expect(result.diagnostics.length).toBeGreaterThan(0);
      expect(result.diagnostics[0]?.message).toContain("package 'diag_demo.support' has no symbol 'missing_symbol'");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});

function evaluateAllModules(interpreter: V10.InterpreterV10, program: LoadedProgram): void {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    interpreter.evaluate(mod.module);
  }
  interpreter.evaluate(program.entry.module);
}
