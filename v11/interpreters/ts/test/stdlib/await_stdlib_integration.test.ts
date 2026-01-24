import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import * as AST from "../../src/ast";
import { ModuleLoader } from "../../scripts/module-loader";
import { collectModuleSearchPaths } from "../../scripts/module-search-paths";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { callCallableValue } from "../../src/interpreter/functions";
import { TypeChecker, V11 } from "../../index";

const PROBE_ROOT = path.resolve(__dirname, "../../..");

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
      continue;
    }
    const entryResult = session.checkModule(mod.module);
    entryResult.diagnostics.forEach((diag) => diagnostics.push(diag.message));
  }
  return diagnostics;
}

describe("Await stdlib helpers", () => {
  test("Await.default and sleep wrappers drive awaitable arms", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-await-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: await_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.concurrency.{Await}

fn main() -> !Array bool {
  runner := spawn {
    first := await [Await.default({ => "fallback"})]
    second := await [Await.sleep(0.seconds(), { => "timer"})]
    third := await [Await.sleep_ms(0, { => "timer"})]
    [first == "fallback", second == "timer", third == "timer"]
  }
  future_flush()
  runner.value()
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

      const secondsValue = interpreter.evaluate(
        AST.functionCall(AST.memberAccessExpression(AST.integerLiteral(0), "seconds"), []),
      ) as any;
      expect(["i32", "i64", "u32", "u64"]).toContain(secondsValue?.kind);
      expect(secondsValue?.value).toBe(0n);

      const pkg = interpreter.packageRegistry.get(program.entry.packageName);
      const mainFn = pkg?.get("main");
      if (!mainFn) {
        throw new Error("entry module missing main");
      }

      const result = callCallableValue(interpreter as any, mainFn, [], interpreter.globals) as any;
      expect(result).toEqual({
        kind: "array",
        elements: [
          { kind: "bool", value: true },
          { kind: "bool", value: true },
          { kind: "bool", value: true },
        ],
      });
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
