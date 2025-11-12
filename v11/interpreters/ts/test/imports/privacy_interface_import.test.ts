import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - privacy in interface imports", () => {
  test("importing private interface fails; public interface succeeds", () => {
    const I = new InterpreterV10();

    const privIface = AST.interfaceDefinition("HiddenI", [], undefined, undefined, undefined, undefined, true);
    I.evaluate(privIface);
    const pubIface = AST.interfaceDefinition("PublicI", [], undefined, undefined, undefined, undefined, false);
    I.evaluate(pubIface);

    const impPriv = AST.importStatement(["pkg" as any], false, [AST.importSelector("HiddenI")] );
    const modPriv = AST.module([impPriv]);
    expect(() => I.evaluate(modPriv as any)).toThrow();

    const impPub = AST.importStatement(["pkg" as any], false, [AST.importSelector("PublicI", "PI")] );
    const modPub = AST.module([impPub]);
    expect(() => I.evaluate(modPub as any)).not.toThrow();
  });
});


