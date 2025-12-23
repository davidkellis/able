import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";

describe("v11 interpreter - package alias import", () => {
  test("import package via alias and access public members; private hidden", () => {
    const I = new Interpreter();
    const pkg = AST.packageStatement(["util"]);
    const defs = [
      AST.functionDefinition("foo", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(3))])),
      AST.functionDefinition("hidden", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]), undefined, undefined, undefined, false, true),
      AST.structDefinition("Thing", [], "named"),
      AST.structDefinition("HiddenThing", [], "named", undefined, undefined, true),
    ];
    const pkgMod = AST.module(defs, [], pkg);
    I.evaluate(pkgMod as any);

    const use = AST.module([
      AST.importStatement(["util"], false, undefined, "U"),
      AST.assignmentExpression(":=", AST.identifier("x"), AST.functionCall(AST.memberAccessExpression(AST.identifier("U"), "foo"), [])),
    ]);
    const res = I.evaluate(use as any);
    expect(res).toEqual({ kind: "i32", value: 3n });

    // alias object kind
    const aliasVal = I.evaluate(AST.identifier("U"));
    expect(aliasVal.kind).toBe("package");

    // private members should not be present on alias
    expect(() => I.evaluate(AST.memberAccessExpression(AST.identifier("U"), "hidden"))).toThrow();
    expect(() => I.evaluate(AST.memberAccessExpression(AST.identifier("U"), "HiddenThing"))).toThrow();

    // public type available through alias
    const t = I.evaluate(AST.memberAccessExpression(AST.identifier("U"), "Thing"));
    expect(t.kind).toBe("struct_def");
  });
});

