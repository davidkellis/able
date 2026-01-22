import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

function buildLibraryModule() {
  const displayInterface = AST.interfaceDefinition(
    "Display",
    [
      AST.functionSignature(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String"),
      ),
    ],
  );
  const bagAlias = AST.typeAliasDefinition(
    AST.identifier("Bag"),
    AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
    [AST.genericParameter("T")],
    undefined,
    true,
  );

  const strListAlias = AST.typeAliasDefinition(
    AST.identifier("StrList"),
    AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("String")]),
    undefined,
    undefined,
    true,
  );

  const headMethod = AST.functionDefinition(
    "head",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.blockExpression([
      AST.returnStatement(
        AST.indexExpression(AST.identifier("self"), AST.integerLiteral(0)),
      ),
    ]),
    AST.nullableTypeExpression(AST.simpleTypeExpression("T")),
    [AST.genericParameter("T")],
  );

  const methods = AST.methodsDefinition(
    AST.genericTypeExpression(AST.simpleTypeExpression("Bag"), [AST.simpleTypeExpression("T")]),
    [headMethod],
    [AST.genericParameter("T")],
  );

  const displayImpl = AST.implementationDefinition(
    "Display",
    AST.simpleTypeExpression("StrList"),
    [
      AST.functionDefinition(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("<strlist>"))]),
        AST.simpleTypeExpression("String"),
      ),
    ],
  );

  return AST.module(
    [displayInterface, bagAlias, strListAlias, methods, displayImpl],
    undefined,
    AST.packageStatement(["pkg"]),
  );
}

function buildConsumerModule() {
  const arrBinding = AST.assignmentExpression(
    ":=",
    AST.identifier("arr"),
    AST.arrayLiteral([AST.stringLiteral("a"), AST.stringLiteral("b")]),
  );
  const callHead = AST.functionCall(AST.memberAccessExpression(AST.identifier("arr"), "head"), []);
  const callToString = AST.functionCall(AST.memberAccessExpression(AST.identifier("arr"), "to_string"), []);

  return AST.module(
    [arrBinding, callHead, callToString],
    [AST.importStatement(["pkg"], true)],
    AST.packageStatement(["app"]),
  );
}

describe("typechecker alias method propagation", () => {
  test("methods and impls defined on private aliases attach to underlying types", () => {
    const libChecker = new TypeChecker();
    const libModule = buildLibraryModule();
    const { diagnostics: libDiagnostics, summary } = libChecker.checkModule(libModule);
    expect(libDiagnostics).toHaveLength(0);
    expect(summary).toBeDefined();

    const consumer = new TypeChecker({ packageSummaries: new Map(summary ? [["pkg", summary]] : []) });
    const consumerModule = buildConsumerModule();
    const { diagnostics } = consumer.checkModule(consumerModule);
    expect(diagnostics).toHaveLength(0);
  });
});
