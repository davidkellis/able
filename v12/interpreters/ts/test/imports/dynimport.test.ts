import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - dynimport", () => {
  test("dynimport selectors and alias work; wildcard binds dynamic refs", () => {
    const I = new Interpreter();
    const pkg = AST.packageStatement(["dynp"]);
    const defs = [
      AST.functionDefinition("f", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))])),
      AST.functionDefinition("hidden", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]), undefined, undefined, undefined, false, true),
    ];
    const pkgMod = AST.module(defs, [], pkg);
    I.evaluate(pkgMod as any);

    // selector + alias
    const m1 = AST.module([
      AST.dynImportStatement(["dynp"], false, [AST.importSelector("f", "ff")] ),
      AST.assignmentExpression(":=", AST.identifier("x"), AST.functionCall(AST.identifier("ff"), [])),
    ]);
    const r1 = I.evaluate(m1 as any);
    expect(r1).toEqual({ kind: "i32", value: 11n });

    // alias package dyn
    const m2 = AST.module([
      AST.dynImportStatement(["dynp"], false, undefined, "D"),
      AST.assignmentExpression(":=", AST.identifier("y"), AST.functionCall(AST.memberAccessExpression(AST.identifier("D"), "f"), [])),
    ]);
    const r2 = I.evaluate(m2 as any);
    expect(r2).toEqual({ kind: "i32", value: 11n });

    // wildcard: binds dynamic refs; hidden should not bind
    const m3 = AST.module([
      AST.dynImportStatement(["dynp"], true),
      AST.assignmentExpression(":=", AST.identifier("z"), AST.functionCall(AST.identifier("f"), [])),
    ]);
    const r3 = I.evaluate(m3 as any);
    expect(r3).toEqual({ kind: "i32", value: 11n });
    expect(() => I.evaluate(AST.identifier("hidden"))).toThrow();
  });
});


