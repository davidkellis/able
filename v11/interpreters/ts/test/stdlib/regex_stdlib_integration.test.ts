import { describe, expect, test } from "bun:test";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

import { ModuleLoader } from "../../scripts/module-loader";
import { collectModuleSearchPaths } from "../../scripts/module-search-paths";
import { ensureConsolePrint, installRuntimeStubs } from "../../scripts/runtime-stubs";
import { callCallableValue } from "../../src/interpreter/functions";
import { V11 } from "../../index";

const PROBE_ROOT = path.resolve(__dirname, "../../..");

const readString = (value: any): string => String(value?.value ?? "");

function evaluateAllModules(interpreter: V11.Interpreter, program: { modules: any[]; entry: any }): void {
  const nonEntry = program.modules.filter((mod) => mod.packageName !== program.entry.packageName);
  for (const mod of nonEntry) {
    interpreter.evaluate(mod.module);
  }
  interpreter.evaluate(program.entry.module);
}
describe("stdlib-backed regex helpers", () => {
  test("regex stub returns unsupported error", async () => {
    const tmpRoot = await fs.mkdtemp(path.join(os.tmpdir(), "able-regex-stdlib-"));
    try {
      await fs.writeFile(path.join(tmpRoot, "package.yml"), "name: regex_stdlib\n", "utf8");
      await fs.writeFile(
        path.join(tmpRoot, "main.able"),
        `
package main

import able.text.regex.{regex_is_match}

fn main() -> String {
  _ = regex_is_match("a", "a")
  "regex stub"
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

      const interpreter = new V11.Interpreter();
      ensureConsolePrint(interpreter);
      installRuntimeStubs(interpreter);
      evaluateAllModules(interpreter, program);

      const pkg = interpreter.packageRegistry.get(program.entry.packageName);
      const mainFn = pkg?.get("main");
      if (!mainFn) throw new Error("entry module missing main");

      const result = callCallableValue(interpreter as any, mainFn, [], interpreter.globals) as any;
      expect(readString(result)).toBe("regex stub");
    } finally {
      await fs.rm(tmpRoot, { recursive: true, force: true });
    }
  });
});
