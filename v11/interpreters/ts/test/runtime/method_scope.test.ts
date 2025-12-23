import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import type { RuntimeValue } from "../../src/interpreter/values";

function buildMethodScopePackage(): AST.Module {
  const widget = AST.structDefinition(
    "Widget",
    [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
    "named",
  );

  const bump = AST.functionDefinition(
    "bump",
    [
      AST.functionParameter("self", AST.simpleTypeExpression("Self")),
      AST.functionParameter("delta", AST.simpleTypeExpression("i32")),
    ],
    AST.blockExpression([
      AST.returnStatement(
        AST.binaryExpression("+", AST.memberAccessExpression(AST.identifier("self"), "value"), AST.identifier("delta")),
      ),
    ]),
    AST.simpleTypeExpression("i32"),
  );

  const augment = AST.functionDefinition(
    "augment",
    [
      AST.functionParameter("item", AST.simpleTypeExpression("Widget")),
      AST.functionParameter("delta", AST.simpleTypeExpression("i32")),
    ],
    AST.blockExpression([
      AST.returnStatement(
        AST.binaryExpression("+", AST.memberAccessExpression(AST.identifier("item"), "value"), AST.identifier("delta")),
      ),
    ]),
    AST.simpleTypeExpression("i32"),
  );

  const make = AST.functionDefinition(
    "make",
    [AST.functionParameter("start", AST.simpleTypeExpression("i32"))],
    AST.blockExpression([
      AST.returnStatement(
        AST.structLiteral(
          [AST.structFieldInitializer(AST.identifier("start"), "value")],
          false,
          "Widget",
        ),
      ),
    ]),
    AST.simpleTypeExpression("Widget"),
  );

  const methods = AST.methodsDefinition(AST.simpleTypeExpression("Widget"), [bump, augment, make]);

  return AST.module([widget, methods], [], AST.packageStatement([AST.identifier("pkgmethods")]));
}

describe("method import scoping and type-qualified UFCS exclusions", () => {
  test("receiver sugar requires the method name to be imported", () => {
    const I = new Interpreter();
    I.evaluate(buildMethodScopePackage());

    const entry = AST.module(
      [
        AST.assignmentExpression(
          ":=",
          AST.identifier("inst"),
          AST.functionCall(
            AST.memberAccessExpression(AST.memberAccessExpression(AST.identifier("pkgmethods"), "Widget"), "make"),
            [AST.integerLiteral(3)],
          ),
        ),
        AST.functionCall(AST.memberAccessExpression(AST.identifier("inst"), "bump"), [AST.integerLiteral(2)]),
      ],
      [AST.importStatement([AST.identifier("pkgmethods")])],
    );

    expect(() => I.evaluate(entry)).toThrow("No field or method named 'bump'");
  });

  test("type-qualified functions are callable explicitly but excluded from UFCS", () => {
    const I = new Interpreter();
    I.evaluate(buildMethodScopePackage());

    const entry = AST.module(
      [
        AST.assignmentExpression(
          ":=",
          AST.identifier("inst"),
          AST.functionCall(
            AST.memberAccessExpression(AST.memberAccessExpression(AST.identifier("pkgmethods"), "Widget"), "make"),
            [AST.integerLiteral(4)],
          ),
        ),
        AST.assignmentExpression(
          ":=",
          AST.identifier("augment"),
          AST.memberAccessExpression(AST.identifier("pkgmethods"), "Widget.augment"),
        ),
        AST.assignmentExpression(
          ":=",
          AST.identifier("direct"),
          AST.functionCall(
            AST.memberAccessExpression(AST.memberAccessExpression(AST.identifier("pkgmethods"), "Widget"), "augment"),
            [AST.identifier("inst"), AST.integerLiteral(5)],
          ),
        ),
        AST.functionCall(AST.memberAccessExpression(AST.identifier("inst"), "augment"), [AST.integerLiteral(1)]),
      ],
      [AST.importStatement([AST.identifier("pkgmethods")])],
    );

    expect(() => I.evaluate(entry)).toThrow("No field or method named 'augment'");
  });

  test("wildcard import surfaces type-qualified exports", () => {
    const I = new Interpreter();
    I.evaluate(buildMethodScopePackage());

    const entry = AST.module(
      [
        AST.assignmentExpression(
          ":=",
          AST.identifier("inst"),
          AST.functionCall(AST.memberAccessExpression(AST.identifier("Widget"), "make"), [AST.integerLiteral(11)]),
        ),
        AST.memberAccessExpression(AST.identifier("inst"), "value"),
      ],
      [AST.importStatement([AST.identifier("pkgmethods")], true)],
    );

    const val = I.evaluate(entry);
    expect(val).toEqual({ kind: "i32", value: 11n });
  });
});
