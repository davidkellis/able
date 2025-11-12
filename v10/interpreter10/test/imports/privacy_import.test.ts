import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - privacy in imports", () => {
  test("importing private function fails", () => {
    const I = new InterpreterV10();
    // define private function in globals
    const priv = AST.functionDefinition("secret", [], AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]), undefined, undefined, undefined, false, true);
    I.evaluate(priv);
    const imp = AST.importStatement(["core" as any], false, [AST.importSelector("secret", "alias")]);
    const mod = AST.module([imp]);
    expect(() => I.evaluate(mod as any)).toThrow();
  });

  test("importing private struct fails; public struct succeeds", () => {
    const I = new InterpreterV10();
    // private struct
    const pstruct = AST.structDefinition("Hidden", [], "named", undefined, undefined, true);
    I.evaluate(pstruct);
    // public struct
    const pub = AST.structDefinition("Public", [], "named", undefined, undefined, false);
    I.evaluate(pub);

    // attempt import private -> should throw
    const impPriv = AST.importStatement(["pkg" as any], false, [AST.importSelector("Hidden")] );
    const modPriv = AST.module([impPriv]);
    expect(() => I.evaluate(modPriv as any)).toThrow();

    // import public (with alias to avoid redefining global) -> should bind without error
    const impPub = AST.importStatement(["pkg" as any], false, [AST.importSelector("Public", "Pub")] );
    const modPub = AST.module([impPub]);
    expect(() => I.evaluate(modPub as any)).not.toThrow();
  });
});


