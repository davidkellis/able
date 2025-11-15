import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker map literals", () => {
  test("infers key/value types", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.identifier("headers"),
        AST.mapLit([
          AST.mapEntry(AST.stringLiteral("content-type"), AST.stringLiteral("application/json")),
          AST.mapEntry(AST.stringLiteral("authorization"), AST.stringLiteral("token")),
        ]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("reports incompatible spreads", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.identifier("headers"),
        AST.mapLit([
          AST.mapEntry(AST.stringLiteral("accept"), AST.stringLiteral("json")),
          AST.mapSpread(AST.stringLiteral("oops")),
        ]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("map spread expects Map");
  });

  test("reports mixed key types", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.identifier("codes"),
        AST.mapLit([
          AST.mapEntry(AST.stringLiteral("ok"), AST.int(200)),
          AST.mapEntry(AST.booleanLiteral(true), AST.int(201)),
        ]),
      ) as unknown as AST.Statement,
    ]);
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("map key expects");
  });
});
