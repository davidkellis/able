import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST } from "../index";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../fixtures/ast");

interface Fixture {
  name: string; // folder relative to FIXTURE_ROOT
  module: AST.Module;
  manifest?: FixtureManifest;
}

interface FixtureManifest {
  description: string;
  entry?: string;
  expect?: Record<string, unknown>;
}

const fixtures: Fixture[] = [
  {
    name: "basics/string_literal",
    module: AST.module([AST.str("hello")]),
    manifest: {
      description: "Evaluates a simple string literal module",
      expect: {
        result: { kind: "string", value: "hello" },
      },
    },
  },
  {
    name: "expressions/int_addition",
    module: AST.module([
      AST.assign("a", AST.int(1)),
      AST.assign("b", AST.int(2)),
      AST.bin("+", AST.id("a"), AST.id("b")),
    ]),
    manifest: {
      description: "Adds two integers",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "control/while_sum",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.assign("i", AST.int(0)),
      AST.assign("limit", AST.int(3)),
      AST.wloop(
        AST.bin("<", AST.id("i"), AST.id("limit")),
        AST.block(
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("i")), "="),
          AST.assign("i", AST.bin("+", AST.id("i"), AST.int(1)), "="),
        ),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "Sums integers using a while loop",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "control/if_stdout",
    module: AST.module([
      AST.ifExpression(
        AST.bool(true),
        AST.block(
          AST.call("print", AST.str("branch")),
        ),
      ),
      AST.str("done"),
    ]),
    manifest: {
      description: "If expression emits stdout",
      expect: {
        stdout: ["branch"],
        result: { kind: "string", value: "done" },
      },
    },
  },
  {
    name: "control/if_else_branch",
    module: AST.module([
      AST.ifExpression(
        AST.bool(false),
        AST.block(AST.call("print", AST.str("true"))),
        [AST.orClause(AST.block(AST.call("print", AST.str("false"))))],
      ),
      AST.str("after"),
    ]),
    manifest: {
      description: "Else branch executes when condition false",
      expect: {
        stdout: ["false"],
        result: { kind: "string", value: "after" },
      },
    },
  },
  {
    name: "control/for_sum",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.assign("items", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
      AST.forIn("n", AST.id("items"), AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "=")),
      AST.id("sum"),
    ]),
    manifest: {
      description: "For loop iterates over array",
      expect: {
        result: { kind: "i32", value: 6 },
      },
    },
  },
  {
    name: "control/for_range_break",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.forIn(
        "n",
        AST.range(AST.int(0), AST.int(5), false),
        AST.block(
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "="),
          AST.ifExpression(
            AST.bin(">=", AST.id("n"), AST.int(2)),
            AST.block(AST.brk(undefined, AST.id("sum"))),
          ),
        ),
      ),
    ]),
    manifest: {
      description: "For loop over range with break",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "patterns/array_destructuring",
    module: AST.module([
      AST.assign("arr", AST.arr(AST.int(1), AST.int(2), AST.int(3), AST.int(4))),
      AST.assign(AST.arrP([AST.id("first"), AST.id("second")], AST.id("rest")), AST.id("arr")),
      AST.assign(AST.arrP([AST.id("third"), AST.id("fourth")]), AST.id("rest")),
      AST.bin("+", AST.id("first"), AST.bin("+", AST.id("second"), AST.id("third"))),
    ]),
    manifest: {
      description: "Array destructuring assignment extracts prefix and rest",
      expect: {
        result: { kind: "i32", value: 6 },
      },
    },
  },
  {
    name: "patterns/for_array_pattern",
    module: AST.module([
      AST.assign("pairs", AST.arr(
        AST.arr(AST.int(1), AST.int(2)),
        AST.arr(AST.int(3), AST.int(4)),
      )),
      AST.assign("sum", AST.int(0)),
      AST.forIn(
        AST.arrP([AST.id("x"), AST.id("y")]),
        AST.id("pairs"),
        AST.block(
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("x")), "="),
          AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("y")), "="),
        ),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "For-in loop destructures array elements",
      expect: {
        result: { kind: "i32", value: 10 },
      },
    },
  },
  {
    name: "patterns/typed_assignment",
    module: AST.module([
      AST.assign("value", AST.int(42)),
      AST.assign(
        AST.typedP(AST.id("n"), AST.ty("i32")),
        AST.id("value"),
      ),
      AST.id("n"),
    ]),
    manifest: {
      description: "Typed pattern enforces simple type on assignment",
      expect: {
        result: { kind: "i32", value: 42 },
      },
    },
  },
  {
    name: "patterns/typed_assignment_error",
    module: AST.module([
      AST.assign("value", AST.str("nope")),
      AST.assign(
        AST.typedP(AST.id("n"), AST.ty("i32")),
        AST.id("value"),
      ),
    ]),
    manifest: {
      description: "Typed pattern mismatch raises error",
      expect: {
        errors: ["Typed pattern mismatch in assignment"],
      },
    },
  },
  {
    name: "structs/named_literal",
    module: AST.module([
      AST.structDefinition(
        "Point",
        [
          AST.fieldDef(AST.ty("i32"), "x"),
          AST.fieldDef(AST.ty("i32"), "y"),
        ],
        "named",
      ),
      AST.assign(
        "point",
        AST.structLiteral(
          [
            AST.fieldInit(AST.int(3), "x"),
            AST.fieldInit(AST.int(4), "y"),
          ],
          false,
          "Point",
        ),
      ),
      AST.member(AST.id("point"), "x"),
    ]),
    manifest: {
      description: "Named struct literal evaluates and exposes fields",
      expect: {
        result: { kind: "i32", value: 3 },
      },
    },
  },
  {
    name: "structs/positional_literal",
    module: AST.module([
      AST.structDefinition(
        "Pair",
        [
          AST.fieldDef(AST.ty("i32")),
          AST.fieldDef(AST.ty("i32")),
        ],
        "positional",
      ),
      AST.assign(
        "pair",
        AST.structLiteral(
          [
            AST.fieldInit(AST.int(7)),
            AST.fieldInit(AST.int(9)),
          ],
          true,
          "Pair",
        ),
      ),
      AST.member(AST.id("pair"), AST.int(1)),
    ]),
    manifest: {
      description: "Positional struct literal supports numeric member access",
      expect: {
        result: { kind: "i32", value: 9 },
      },
    },
  },
  {
    name: "errors/rescue_guard",
    module: AST.module([
      AST.rescue(
        AST.block(AST.raise(AST.str("boom"))),
        [
          AST.mc(AST.litP(AST.str("ignore")), AST.str("ignored")),
          AST.mc(
            AST.id("msg"),
            AST.block(
              AST.ifExpression(
                AST.bin("==", AST.id("msg"), AST.str("boom")),
                AST.block(AST.str("handled")),
              ),
            ),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Rescue guard selects matching clause",
      expect: {
        result: { kind: "nil" },
      },
    },
  },
  {
    name: "errors/raise_manifest",
    module: AST.module([
      AST.raise(AST.str("boom")),
    ]),
    manifest: {
      description: "Fixture raises error",
      expect: {
        errors: ["boom"],
      },
    },
  },
  {
    name: "errors/rescue_catch",
    module: AST.module([
      AST.rescue(
        AST.block(AST.raise(AST.str("boom"))),
        [
          AST.mc(
            AST.id("err"),
            AST.block(AST.call("print", AST.id("err")), AST.str("handled")),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Rescue expression catches raise",
      expect: {
        stdout: ["[error]"],
        result: { kind: "string", value: "handled" },
      },
    },
  },
];

async function main() {
  for (const fixture of fixtures) {
    await writeFixture(fixture);
  }
  console.log(`Wrote ${fixtures.length} fixture(s) to ${FIXTURE_ROOT}`);
}

async function writeFixture(fixture: Fixture) {
  const outDir = path.join(FIXTURE_ROOT, fixture.name);
  await fs.mkdir(outDir, { recursive: true });

  const modulePath = path.join(outDir, "module.json");
  await fs.writeFile(modulePath, stringify(fixture.module), "utf8");

  if (fixture.manifest) {
    const manifestPath = path.join(outDir, "manifest.json");
    const entry = fixture.manifest.entry ?? "module.json";
    const manifest = { ...fixture.manifest, entry };
    await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2), "utf8");
  }
}

function stringify(value: unknown): string {
  return JSON.stringify(
    value,
    (_key, val) => (typeof val === "bigint" ? val.toString() : val),
    2,
  );
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
