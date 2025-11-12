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

  test("reports private selector symbols", () => {
    const moduleAst = AST.module(
      [],
      [AST.importStatement(["util"], false, [AST.importSelector("secret")])],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker({
      packageSummaries: new Map([
        [
          "util",
          sampleSummary({
            privateSymbols: {
              secret: { type: "secret", visibility: "private" },
            },
          }),
        ],
      ]),
    });
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: package 'util' symbol 'secret' is private");
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

  test("reports alias member access for private exports", () => {
    const moduleAst = AST.module(
      [AST.memberAccessExpression(AST.identifier("Util"), AST.identifier("secret"))],
      [AST.importStatement(["util"], false, undefined, "Util")],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker({
      packageSummaries: new Map([
        [
          "util",
          sampleSummary({
            privateSymbols: {
              secret: { type: "secret", visibility: "private" },
            },
          }),
        ],
      ]),
    });
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: package 'util' has no symbol 'secret'");
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

  test("reports undefined identifier when symbol is missing", () => {
    const moduleAst = AST.module([AST.identifier("missing")], [], AST.packageStatement(["app"]));
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toBe("typechecker: undefined identifier 'missing'");
  });

  test("dynimport wildcard allows late-bound identifiers", () => {
    const moduleAst = AST.module(
      [AST.dynImportStatement(["dynpkg"], true), AST.identifier("lateBound")],
      [],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(0);
  });

  test("dynimport selectors bind aliases", () => {
    const moduleAst = AST.module(
      [
        AST.dynImportStatement(["dynpkg"], false, [AST.importSelector("foo", "alias")]),
        AST.identifier("alias"),
      ],
      [],
      AST.packageStatement(["app"]),
    );
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toHaveLength(0);
  });
});

function sampleSummary(overrides?: Partial<PackageSummary>): PackageSummary {
  const base: PackageSummary = {
    name: "util",
    visibility: "public",
    symbols: {
      foo: { type: "foo", visibility: "public" },
    },
    privateSymbols: {},
    structs: {},
    interfaces: {},
    functions: {},
    implementations: [],
    methodSets: [],
  };
  return {
    ...base,
    ...overrides,
    symbols: overrides?.symbols ?? base.symbols,
    privateSymbols: overrides?.privateSymbols ?? base.privateSymbols,
  };
}
