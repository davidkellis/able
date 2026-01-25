import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypecheckerSession } from "../../src/typechecker";

function buildPackage(pkg: string, returnType: string, returnExpr: AST.Expression): AST.Module {
  const foo = AST.structDefinition("Foo", [], "named");
  const make = AST.functionDefinition(
    "make",
    [],
    AST.blockExpression([AST.returnStatement(returnExpr)]),
    AST.simpleTypeExpression(returnType),
  );
  const methods = AST.methodsDefinition(AST.simpleTypeExpression("Foo"), [make]);
  return AST.module([foo, methods], [], AST.packageStatement([pkg]));
}

describe("typechecker type symbol scoping for methods", () => {
  test("type-qualified resolution prefers the explicitly imported type when names conflict", () => {
    const session = new TypecheckerSession();
    const pkgA = buildPackage("pkgA", "i32", AST.integerLiteral(1));
    const pkgB = buildPackage("pkgB", "String", AST.stringLiteral("ok"));
    expect(session.checkModule(pkgA).diagnostics).toHaveLength(0);
    expect(session.checkModule(pkgB).diagnostics).toHaveLength(0);

    const consumer = AST.module(
      [
        AST.functionCall(AST.memberAccessExpression(AST.identifier("Foo"), "make"), []),
      ],
      [AST.importStatement(["pkgA"], false, [AST.importSelector("Foo")])],
      AST.packageStatement(["app"]),
    );
    const result = session.checkModule(consumer);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("type-qualified calls require the type symbol to be in scope", () => {
    const session = new TypecheckerSession();
    const pkgA = buildPackage("pkgA", "i32", AST.integerLiteral(1));
    expect(session.checkModule(pkgA).diagnostics).toHaveLength(0);

    const consumer = AST.module(
      [AST.functionCall(AST.memberAccessExpression(AST.identifier("Foo"), "make"), [])],
      [],
      AST.packageStatement(["app"]),
    );
    const result = session.checkModule(consumer);
    expect(result.diagnostics.some((diag) => diag.message.includes("undefined identifier 'Foo'"))).toBe(true);
  });
});
