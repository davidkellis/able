import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST } from "../index";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../fixtures/ast");

interface Fixture {
  name: string; // folder relative to FIXTURE_ROOT
  module: AST.Module;
  setupModules?: Record<string, AST.Module>;
  manifest?: FixtureManifest;
}

interface FixtureManifest {
  description: string;
  entry?: string;
  expect?: Record<string, unknown>;
  setup?: string[];
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
    name: "strings/interpolation_basic",
    module: AST.module([
      AST.assign("x", AST.int(2)),
      AST.stringInterpolation([
        AST.stringLiteral("x = "),
        AST.identifier("x"),
        AST.stringLiteral(", sum = "),
        AST.binaryExpression("+", AST.integerLiteral(3), AST.integerLiteral(4)),
      ]),
    ]),
    manifest: {
      description: "Interpolates literals and expressions",
      expect: {
        result: { kind: "string", value: "x = 2, sum = 7" },
      },
    },
  },
  {
    name: "strings/interpolation_struct_to_string",
    module: AST.module([
      AST.structDefinition(
        "Point",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
        ],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Point"),
        [
          AST.functionDefinition(
            "to_string",
            [AST.functionParameter("self")],
            AST.blockExpression([
              AST.returnStatement(
                AST.stringInterpolation([
                  AST.stringLiteral("Point("),
                  AST.memberAccessExpression(AST.identifier("self"), "x"),
                  AST.stringLiteral(","),
                  AST.memberAccessExpression(AST.identifier("self"), "y"),
                  AST.stringLiteral(")"),
                ]),
              ),
            ]),
          ),
        ],
      ),
      AST.assign(
        "p",
        AST.structLiteral(
          [
            AST.structFieldInitializer(AST.integerLiteral(1), "x"),
            AST.structFieldInitializer(AST.integerLiteral(2), "y"),
          ],
          false,
          "Point",
        ),
      ),
      AST.stringInterpolation([
        AST.stringLiteral("P= "),
        AST.identifier("p"),
      ]),
    ]),
    manifest: {
      description: "Uses to_string method when interpolating struct instances",
      expect: {
        result: { kind: "string", value: "P= Point(1,2)" },
      },
    },
  },
  {
    name: "match/identifier_literal",
    module: AST.module([
      AST.matchExpression(
        AST.integerLiteral(2),
        [
          AST.matchClause(
            AST.literalPattern(AST.integerLiteral(1)),
            AST.integerLiteral(10),
          ),
          AST.matchClause(
            AST.identifier("x"),
            AST.binaryExpression("+", AST.identifier("x"), AST.integerLiteral(5)),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Match falls through literal clause and binds identifier",
      expect: {
        result: { kind: "i32", value: 7 },
      },
    },
  },
  {
    name: "match/struct_guard",
    module: AST.module([
      AST.structDefinition(
        "Point",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
        ],
        "named",
      ),
      AST.matchExpression(
        AST.structLiteral([
          AST.structFieldInitializer(AST.integerLiteral(1), "x"),
          AST.structFieldInitializer(AST.integerLiteral(2), "y"),
        ], false, "Point"),
        [
          AST.matchClause(
            AST.structPattern([
              AST.structPatternField(AST.identifier("a"), "x"),
              AST.structPatternField(AST.identifier("b"), "y"),
            ], false, "Point"),
            AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")),
            AST.binaryExpression(">", AST.identifier("b"), AST.identifier("a")),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Match struct pattern binds fields and guard filters clauses",
      expect: {
        result: { kind: "i32", value: 3 },
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
    name: "control/for_continue",
    module: AST.module([
      AST.assign("sum", AST.int(0)),
      AST.assign("items", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
      AST.forIn(
        "n",
        AST.id("items"),
        AST.ifExpression(
          AST.bin("==", AST.id("n"), AST.int(2)),
          AST.block(AST.continueStatement()),
        ),
        AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("n")), "="),
      ),
      AST.id("sum"),
    ]),
    manifest: {
      description: "For loop continue skips matching elements",
      expect: {
        result: { kind: "i32", value: 4 },
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
    name: "privacy/private_static_method",
    module: AST.module([
      AST.structDefinition(
        "Point",
        [
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
          AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
        ],
        "named",
      ),
      AST.methodsDefinition(
        AST.simpleTypeExpression("Point"),
        [
          AST.functionDefinition(
            "hidden_static",
            [],
            AST.blockExpression([
              AST.returnStatement(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.integerLiteral(0), "x"),
                    AST.structFieldInitializer(AST.integerLiteral(0), "y"),
                  ],
                  false,
                  "Point",
                ),
              ),
            ]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
          AST.functionDefinition(
            "origin",
            [],
            AST.blockExpression([
              AST.returnStatement(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.integerLiteral(0), "x"),
                    AST.structFieldInitializer(AST.integerLiteral(0), "y"),
                  ],
                  false,
                  "Point",
                ),
              ),
            ]),
          ),
        ],
      ),
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("Point"), "hidden_static"),
        [],
      ),
    ]),
    manifest: {
      description: "Calling a private static method raises an error",
      expect: {
        errors: ["Method 'hidden_static' on Point is private"],
      },
    },
  },
  {
    name: "imports/dynimport_selector_alias",
    setupModules: {
      "package.json": AST.module(
        [
          AST.functionDefinition(
            "f",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))]),
          ),
          AST.functionDefinition(
            "hidden",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
        ],
        [],
        AST.packageStatement(["dynp"]),
      ),
    },
    module: AST.module([
      AST.dynImportStatement(["dynp"], false, [AST.importSelector("f", "ff")]),
      AST.dynImportStatement(["dynp"], false, undefined, "D"),
      AST.assignmentExpression(
        ":=",
        AST.identifier("x"),
        AST.call(AST.identifier("ff")),
      ),
      AST.assignmentExpression(
        ":=",
        AST.identifier("y"),
        AST.call(AST.memberAccessExpression(AST.identifier("D"), "f")),
      ),
      AST.binaryExpression("+", AST.identifier("x"), AST.identifier("y")),
    ]),
    manifest: {
      description: "Dyn import selector and alias return callable references",
      expect: {
        result: { kind: "i32", value: 22 },
      },
    },
  },
  {
    name: "imports/dynimport_wildcard",
    setupModules: {
      "package.json": AST.module(
        [
          AST.functionDefinition(
            "f",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))]),
          ),
          AST.functionDefinition(
            "hidden",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
        ],
        [],
        AST.packageStatement(["dynp"]),
      ),
    },
    module: AST.module([
      AST.dynImportStatement(["dynp"], true),
      AST.assignmentExpression(
        ":=",
        AST.identifier("value"),
        AST.call(AST.identifier("f")),
      ),
      AST.identifier("value"),
    ]),
    manifest: {
      description: "Dyn import wildcard exposes public symbols",
      expect: {
        result: { kind: "i32", value: 11 },
      },
    },
  },
  {
    name: "imports/dynimport_private_selector_error",
    setupModules: {
      "package.json": AST.module(
        [
          AST.functionDefinition(
            "hidden",
            [],
            AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
            undefined,
            undefined,
            undefined,
            false,
            true,
          ),
        ],
        [],
        AST.packageStatement(["dynp"]),
      ),
    },
    module: AST.module([
      AST.dynImportStatement(["dynp"], false, [AST.importSelector("hidden")]),
    ]),
    manifest: {
      description: "Dyn import selector rejects private symbols",
      expect: {
        errors: ["dynimport error: function 'hidden' is private"],
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
  {
    name: "errors/rescue_typed_pattern",
    module: AST.module([
      AST.rescue(
        AST.block(AST.raise(AST.str("boom"))),
        [
          AST.matchClause(
            AST.typedPattern(AST.identifier("err"), AST.simpleTypeExpression("Error")),
            AST.stringLiteral("caught"),
          ),
        ],
      ),
    ]),
    manifest: {
      description: "Typed pattern catches raised error",
      expect: {
        result: { kind: "string", value: "caught" },
      },
    },
  },
  {
    name: "errors/or_else_handler",
    module: AST.module([
      AST.orElseExpression(
        AST.propagationExpression(
          AST.blockExpression([AST.raiseStatement(AST.stringLiteral("boom"))]),
        ),
        AST.blockExpression([AST.stringLiteral("handled")]),
        "err",
      ),
    ]),
    manifest: {
      description: "Or else handler runs when propagation raises",
      expect: {
        result: { kind: "string", value: "handled" },
      },
    },
  },
  {
    name: "errors/ensure_runs",
    module: AST.module([
      AST.ensureExpression(
        AST.rescueExpression(
          AST.blockExpression([AST.raiseStatement(AST.stringLiteral("oops"))]),
          [AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("rescued"))],
        ),
        AST.blockExpression([AST.call("print", AST.stringLiteral("ensure"))]),
      ),
    ]),
    manifest: {
      description: "Ensure block executes regardless of rescue",
      expect: {
        stdout: ["ensure"],
        result: { kind: "string", value: "rescued" },
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

  if (fixture.setupModules) {
    for (const [fileName, module] of Object.entries(fixture.setupModules)) {
      const filePath = path.join(outDir, fileName);
      await fs.writeFile(filePath, stringify(module), "utf8");
    }
  }

  const modulePath = path.join(outDir, "module.json");
  await fs.writeFile(modulePath, stringify(fixture.module), "utf8");

  if (fixture.manifest) {
    const manifestPath = path.join(outDir, "manifest.json");
    const entry = fixture.manifest.entry ?? "module.json";
    const setup = fixture.manifest.setup ?? (fixture.setupModules ? Object.keys(fixture.setupModules) : undefined);
    const manifest = { ...fixture.manifest, entry, ...(setup ? { setup } : {}) };
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
