import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker, TypecheckerSession } from "../../src/typechecker";

describe("TypeChecker package summary", () => {
  test("captures exported program surface", () => {
    const moduleAst = buildSampleModule();
    const checker = new TypeChecker();
    const { diagnostics, summary } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([]);
    expect(summary).not.toBeNull();
    if (!summary) {
      throw new Error("expected summary to be present");
    }

    expect(summary.name).toBe("demo.pkg");
    expect(summary.symbols).toHaveProperty("Point");
    expect(summary.symbols).not.toHaveProperty("Hidden");

    expect(summary.structs.Point.fields).toEqual({ x: "i32", y: "i32" });
    expect(summary.functions.make_point.parameters).toEqual(["i32", "i32"]);
    expect(summary.functions.make_point.returnType).toBe("Point");
    expect(summary.interfaces.Show.methods.to_string.returnType).toBe("string");

    expect(summary.implementations).toHaveLength(1);
    expect(summary.implementations[0].interface).toBe("Show");
    expect(summary.methodSets).toHaveLength(1);
    expect(summary.methodSets[0].target).toBe("Point");
    expect(summary.methodSets[0].methods.display.returnType).toBe("string");
  });

  test("falls back to anonymous package name", () => {
    const moduleAst = AST.module([
      AST.functionDefinition(
        "main",
        [],
        AST.blockExpression([AST.returnStatement()]),
      ),
    ]);
    const checker = new TypeChecker();
    const { summary } = checker.checkModule(moduleAst);
    expect(summary).not.toBeNull();
    expect(summary?.name).toBe("<anonymous>");
  });

  test("session aggregates package summaries", () => {
    const moduleAst = buildSampleModule();
    const session = new TypecheckerSession();
    const { diagnostics } = session.checkModule(moduleAst);
    expect(diagnostics).toEqual([]);
    const summaries = session.getPackageSummaries();
    expect([...summaries.keys()]).toContain("demo.pkg");
  });
});

function buildSampleModule(): AST.Module {
  const pointStruct = AST.structDefinition(
    "Point",
    [
      AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
      AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
    ],
    "named",
  );

  const hiddenStruct = AST.structDefinition("Hidden", [], "singleton", undefined, undefined, true);

  const showInterface = AST.interfaceDefinition("Show", [
    AST.functionSignature(
      "to_string",
      [
        AST.functionParameter("self", AST.simpleTypeExpression("Self")),
      ],
      AST.simpleTypeExpression("string"),
    ),
  ]);

  const makePointFn = AST.functionDefinition(
    "make_point",
    [
      AST.functionParameter("x", AST.simpleTypeExpression("i32")),
      AST.functionParameter("y", AST.simpleTypeExpression("i32")),
    ],
    AST.blockExpression([
      AST.returnStatement(
        AST.structLiteral(
          [
            AST.structFieldInitializer(AST.identifier("x"), "x"),
            AST.structFieldInitializer(AST.identifier("y"), "y"),
          ],
          false,
          "Point",
        ),
      ),
    ]),
    AST.simpleTypeExpression("Point"),
  );

  const helperFn = AST.functionDefinition(
    "helper",
    [],
    AST.blockExpression([AST.returnStatement(AST.integerLiteral(0))]),
    AST.simpleTypeExpression("i32"),
    undefined,
    undefined,
    false,
    true,
  );

  const impl = AST.implementationDefinition(
    "Show",
    AST.simpleTypeExpression("Point"),
    [
      AST.functionDefinition(
        "to_string",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Point")),
        ],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
        AST.simpleTypeExpression("string"),
      ),
    ],
  );

  const methods = AST.methodsDefinition(
    AST.simpleTypeExpression("Point"),
    [
      AST.functionDefinition(
        "display",
        [AST.functionParameter("prefix", AST.simpleTypeExpression("string"))],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("display"))]),
        AST.simpleTypeExpression("string"),
        undefined,
        undefined,
        true,
      ),
    ],
  );

  const pkg = AST.packageStatement(["demo", "pkg"]);

  return AST.module(
    [pointStruct, hiddenStruct, showInterface, makePointFn, helperFn, impl, methods],
    [],
    pkg,
  );
}
