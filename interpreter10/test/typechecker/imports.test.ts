import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";
import type { PackageSummary } from "../../src/typechecker/diagnostics";

describe("TypeChecker imports", () => {
  test("reports unknown packages", () => {
    const moduleAst = AST.module(
      [],
      [AST.importStatement(["missing"], false, undefined, "Pkg")],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: import references unknown package 'missing'");
  });

  test("reports missing selector symbols", () => {
    const moduleAst = AST.module(
      [],
      [AST.importStatement(["util"], false, [AST.importSelector("unknown")])],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker({
      packageSummaries: new Map([["util", sampleSummary()]]),
    });
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: package 'util' has no symbol 'unknown'");
  });

  test("reports alias member access for missing exports", () => {
    const moduleAst = AST.module(
      [AST.memberAccessExpression(AST.identifier("Util"), AST.identifier("hidden"))],
      [AST.importStatement(["util"], false, undefined, "Util")],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker({
      packageSummaries: new Map([["util", sampleSummary()]]),
    });
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: package 'util' has no symbol 'hidden'");
  });

  test("reports error when importing private package", () => {
    const moduleAst = AST.module(
      [],
      [AST.importStatement(["secret"], false)],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker({
      packageSummaries: new Map([["secret", { ...sampleSummary(), name: "secret", visibility: "private" }]]),
    });
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: package 'secret' is private");
  });
});

function sampleSummary(): PackageSummary {
  return {
    name: "util",
    visibility: "public",
    symbols: {
      foo: { type: "foo" },
    },
    structs: {},
    interfaces: {},
    functions: {},
    implementations: [],
    methodSets: [],
  };
}
