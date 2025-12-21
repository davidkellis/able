import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker, TypecheckerSession } from "../../src/typechecker";

function buildCoreModule() {
  const thing = AST.structDefinition(
    "Thing",
    [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
    "named",
  );
  return AST.module([thing], [], AST.packageStatement(["core"]));
}

function buildLibModule() {
  const importCore = AST.importStatement(["core"], false, [AST.importSelector("Thing", "AliasThing")]);
  const method = AST.functionDefinition(
    "alias_method",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
    AST.simpleTypeExpression("i32"),
  );
  const methods = AST.methodsDefinition(AST.simpleTypeExpression("AliasThing"), [method]);
  return AST.module([methods], [importCore], AST.packageStatement(["lib"]));
}

describe("typechecker import alias method propagation", () => {
  test("methods defined on imported aliases are keyed by the underlying type", () => {
    const coreChecker = new TypeChecker();
    const coreResult = coreChecker.checkModule(buildCoreModule());
    expect(coreResult.diagnostics).toHaveLength(0);

    const libChecker = new TypeChecker({
      packageSummaries: coreResult.summary ? new Map([[coreResult.summary.name, coreResult.summary]]) : undefined,
      prelude: coreChecker.exportPrelude(),
    });
    const libResult = libChecker.checkModule(buildLibModule());
    expect(libResult.diagnostics).toHaveLength(0);

    const prelude = libChecker.exportPrelude();

    const methodLabels = prelude?.methodSets.map((record) => record.label) ?? [];
    expect(methodLabels).toContain("methods for Thing");
    expect(methodLabels).not.toContain("methods for AliasThing");
  });

  test("type-qualified alias methods export under the canonical type name", () => {
    const session = new TypecheckerSession();

    const coreModule = AST.module(
      [
        AST.structDefinition(
          "Widget",
          [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
          "named",
        ),
      ],
      [],
      AST.packageStatement(["corealias"]),
    );
    const coreResult = session.checkModule(coreModule);
    expect(coreResult.diagnostics).toHaveLength(0);

    const extModule = AST.module(
      [
        AST.typeAliasDefinition(
          AST.identifier("AliasWidget"),
          AST.simpleTypeExpression("Widget"),
          undefined,
          undefined,
          true,
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("AliasWidget"),
          [
            AST.functionDefinition(
              "make",
              [AST.functionParameter("value", AST.simpleTypeExpression("i32"))],
              AST.blockExpression([
                AST.returnStatement(
                  AST.structLiteral(
                    [AST.structFieldInitializer(AST.identifier("value"), "value")],
                    false,
                    "Widget",
                  ),
                ),
              ]),
              AST.simpleTypeExpression("Widget"),
            ),
          ],
        ),
      ],
      [AST.importStatement(["corealias"], true)],
      AST.packageStatement(["extalias"]),
    );
    const extResult = session.checkModule(extModule);
    expect(extResult.diagnostics).toHaveLength(0);
    expect(extResult.summary?.symbols["Widget.make"]).toBeDefined();
    expect(extResult.summary?.functions["Widget.make"]).toBeDefined();

    const consumer = AST.module(
      [
        AST.assignmentExpression(
          ":=",
          AST.identifier("instance"),
          AST.functionCall(AST.memberAccessExpression(AST.identifier("Widget"), "make"), [AST.int(9)]),
        ),
        AST.memberAccessExpression(AST.identifier("instance"), "value"),
      ],
      [AST.importStatement(["corealias"], true), AST.importStatement(["extalias"], true)],
      AST.packageStatement(["appalias"]),
    );
    const consumerResult = session.checkModule(consumer);
    expect(consumerResult.diagnostics).toHaveLength(0);
  });
});
