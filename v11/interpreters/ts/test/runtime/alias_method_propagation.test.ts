import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

function buildAliasDefs() {
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

  return [bagAlias, strListAlias, methods, displayImpl] as const;
}

describe("runtime alias method propagation", () => {
  test("private alias methods/impls attach to the underlying type", () => {
    const I = new InterpreterV10();
    const [bagAlias, strListAlias, methods, displayImpl] = buildAliasDefs();

    I.evaluate(AST.module([bagAlias, strListAlias, methods, displayImpl]));

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("arr"),
        AST.arrayLiteral([AST.stringLiteral("left"), AST.stringLiteral("right")]),
      ),
    );

    const headCall = AST.functionCall(AST.memberAccessExpression(AST.identifier("arr"), "head"), []);
    const toStringCall = AST.functionCall(AST.memberAccessExpression(AST.identifier("arr"), "to_string"), []);

    expect(I.evaluate(headCall)).toEqual({ kind: "String", value: "left" });
    expect(I.evaluate(toStringCall)).toEqual({ kind: "String", value: "<strlist>" });
  });

  test("methods defined via imported aliases register under the canonical type", () => {
    const I = new InterpreterV10();
    const widgetStruct = AST.structDefinition(
      "Widget",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );
    I.evaluate(widgetStruct);
    // Simulate an import alias binding to the same struct definition.
    I.globals.define("AliasWidget", { kind: "struct_def", def: widgetStruct });

    const describeMethod = AST.functionDefinition(
      "describe",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("ok"))]),
      AST.simpleTypeExpression("String"),
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("AliasWidget"), [describeMethod]);
    I.evaluate(methods);

    const widgetMethods = I.inherentMethods.get("Widget");
    expect(widgetMethods).toBeDefined();
    expect(widgetMethods?.has("describe")).toBe(true);
  });
});
