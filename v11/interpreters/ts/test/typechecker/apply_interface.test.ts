import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

function makeApplyInterface(): AST.InterfaceDefinition {
  return AST.interfaceDefinition(
    "Apply",
    [
      AST.functionSignature(
        "apply",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("args", AST.simpleTypeExpression("Args")),
        ],
        AST.simpleTypeExpression("Result"),
      ),
    ],
    [AST.genericParameter("Args"), AST.genericParameter("Result")],
  );
}

describe("TypeChecker Apply interface calls", () => {
  test("allows calls to values typed as Apply implementors", () => {
    const applyInterface = makeApplyInterface();
    const appenderStruct = AST.structDefinition(
      "Appender",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "prefix")],
      "named",
    );
    const applyImpl = AST.implementationDefinition(
      "Apply",
      AST.simpleTypeExpression("Appender"),
      [
        AST.functionDefinition(
          "apply",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("suffix", AST.simpleTypeExpression("String")),
          ],
          AST.blockExpression([
            AST.returnStatement(
              AST.binaryExpression("+", AST.implicitMemberExpression("prefix"), AST.identifier("suffix")),
            ),
          ]),
          AST.simpleTypeExpression("String"),
          undefined,
          undefined,
          true,
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("String"), AST.simpleTypeExpression("String")],
    );
    const bind = AST.assignmentExpression(
      ":=",
      AST.typedPattern(
        AST.identifier("callable"),
        AST.genericTypeExpression(AST.simpleTypeExpression("Apply"), [
          AST.simpleTypeExpression("String"),
          AST.simpleTypeExpression("String"),
        ]),
      ),
      AST.structLiteral(
        [AST.structFieldInitializer(AST.stringLiteral("hi "), "prefix")],
        false,
        "Appender",
      ),
    );
    const callExpr = AST.functionCall(AST.identifier("callable"), [AST.stringLiteral("there")]);
    const module = AST.module([applyInterface, appenderStruct, applyImpl, bind, callExpr], [], AST.packageStatement(["app"]));

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("reports missing Apply implementation for non-callable values", () => {
    const applyInterface = makeApplyInterface();
    const boxStruct = AST.structDefinition(
      "Box",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );
    const bind = AST.assignmentExpression(
      ":=",
      AST.identifier("b"),
      AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "value")], false, "Box"),
    );
    const callExpr = AST.functionCall(AST.identifier("b"), []);
    const module = AST.module([applyInterface, boxStruct, bind, callExpr], [], AST.packageStatement(["app"]));

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics.some((d) => d.message.includes("Apply"))).toBe(true);
  });

  test("allows calls to concrete types implementing Apply", () => {
    const applyInterface = makeApplyInterface();
    const multStruct = AST.structDefinition(
      "Multiplier",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "factor")],
      "named",
    );
    const applyImpl = AST.implementationDefinition(
      "Apply",
      AST.simpleTypeExpression("Multiplier"),
      [
        AST.functionDefinition(
          "apply",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("input", AST.simpleTypeExpression("i32")),
          ],
          AST.blockExpression([
            AST.returnStatement(
              AST.binaryExpression("*", AST.implicitMemberExpression("factor"), AST.identifier("input")),
            ),
          ]),
          AST.simpleTypeExpression("i32"),
          undefined,
          undefined,
          true,
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")],
    );
    const bind = AST.assignmentExpression(
      ":=",
      AST.identifier("m"),
      AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(4), "factor")], false, "Multiplier"),
    );
    const callExpr = AST.functionCall(AST.identifier("m"), [AST.integerLiteral(5)]);
    const annotate = AST.assignmentExpression(
      ":=",
      AST.typedPattern(AST.identifier("result"), AST.simpleTypeExpression("i32")),
      callExpr,
    );
    const module = AST.module([applyInterface, multStruct, applyImpl, bind, annotate], [], AST.packageStatement(["app"]));

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("reports argument mismatches for concrete Apply implementations", () => {
    const applyInterface = makeApplyInterface();
    const appenderStruct = AST.structDefinition(
      "Appender",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "prefix")],
      "named",
    );
    const applyImpl = AST.implementationDefinition(
      "Apply",
      AST.simpleTypeExpression("Appender"),
      [
        AST.functionDefinition(
          "apply",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("suffix", AST.simpleTypeExpression("String")),
          ],
          AST.blockExpression([
            AST.returnStatement(
              AST.binaryExpression("+", AST.implicitMemberExpression("prefix"), AST.identifier("suffix")),
            ),
          ]),
          AST.simpleTypeExpression("String"),
          undefined,
          undefined,
          true,
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("String"), AST.simpleTypeExpression("String")],
    );
    const bind = AST.assignmentExpression(
      ":=",
      AST.identifier("a"),
      AST.structLiteral([AST.structFieldInitializer(AST.stringLiteral("hi "), "prefix")], false, "Appender"),
    );
    const callExpr = AST.functionCall(AST.identifier("a"), [AST.integerLiteral(1)]);
    const module = AST.module([applyInterface, appenderStruct, applyImpl, bind, callExpr], [], AST.packageStatement(["app"]));

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics.some((d) => d.message.includes("argument 1"))).toBe(true);
  });
});
