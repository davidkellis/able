import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - wildcard import", () => {
  test("wildcard imports bring in public names and skip private ones", () => {
    const I = new InterpreterV10();

    // simulate a package by evaluating a module with a package statement
    const pkg = AST.packageStatement(["my_pkg"]);
    const defs = [
      AST.functionDefinition("pub_fn", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(7))])),
      AST.functionDefinition("priv_fn", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]), undefined, undefined, undefined, false, true),
      AST.structDefinition("PubType", [], "named"),
      AST.structDefinition("HiddenType", [], "named", undefined, undefined, true),
      AST.interfaceDefinition("PublicI", [], undefined, undefined, undefined, undefined, false),
      AST.interfaceDefinition("HiddenI", [], undefined, undefined, undefined, undefined, true),
    ];
    const pkgMod = AST.module(defs, [], pkg);
    I.evaluate(pkgMod as any);

    // now wildcard import from that package
    const useMod = AST.module([
      AST.importStatement(["my_pkg"], true),
      AST.assignmentExpression(":=", AST.identifier("x"), AST.functionCall(AST.identifier("pub_fn"), [])),
    ]);
    const res = I.evaluate(useMod as any);
    expect(res).toEqual({ kind: "i32", value: 7n });

    // Ensure private ones are not defined
    expect(() => I.evaluate(AST.identifier("priv_fn"))).toThrow();
    expect(() => I.evaluate(AST.identifier("HiddenType"))).toThrow();
    expect(() => I.evaluate(AST.identifier("HiddenI"))).toThrow();

    // Public type/interface should be present
    const pubType = I.evaluate(AST.identifier("PubType"));
    expect(pubType.kind).toBe("struct_def");
    const pubI = I.evaluate(AST.identifier("PublicI"));
    expect(pubI.kind).toBe("interface_def");
  });
});


