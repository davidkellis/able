import { AST } from "../../context";
import type { Fixture } from "../../types";

const importsFixtures: Fixture[] = [
  {
      name: "imports/implicit_generic_import",
      setupModules: {
        "package.json": AST.module(
          [
            AST.functionDefinition(
              "wrap",
              [AST.functionParameter("value", AST.simpleTypeExpression("T"))],
              AST.blockExpression([AST.arr(AST.identifier("value"))]),
              AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
            ),
          ],
          [],
          AST.packageStatement(["wraplib"]),
        ),
      },
      module: AST.module(
        [AST.nil()],
        [AST.importStatement(["wraplib"], false, [AST.importSelector("wrap")])],
      ),
      manifest: {
        description: "Implicit generics remain available when functions are imported",
        setup: ["package.json"],
        expect: {
          result: { kind: "nil" },
        },
      },
    },

  {
      name: "imports/dynimport_selector_alias",
      setupModules: {
        "package.json": AST.module(
          [
            AST.functionDefinition(
              "f",
              [],
              AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))]),
            ),
            AST.functionDefinition(
              "hidden",
              [],
              AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
              undefined,
              undefined,
              undefined,
              false,
              true,
            ),
          ],
          [],
          AST.packageStatement(["dynp"]),
        ),
      },
      module: AST.module([
        AST.dynImportStatement(["dynp"], false, [AST.importSelector("f", "ff")]),
        AST.dynImportStatement(["dynp"], false, undefined, "D"),
        AST.assignmentExpression(
          ":=",
          AST.identifier("x"),
          AST.call(AST.identifier("ff")),
        ),
        AST.assignmentExpression(
          ":=",
          AST.identifier("y"),
          AST.call(AST.memberAccessExpression(AST.identifier("D"), "f")),
        ),
        AST.binaryExpression("+", AST.identifier("x"), AST.identifier("y")),
      ]),
      manifest: {
        description: "Dyn import selector and alias return callable references",
        expect: {
          result: { kind: "i32", value: 22n },
        },
      },
    },

  {
      name: "imports/dynimport_wildcard",
      setupModules: {
        "package.json": AST.module(
          [
            AST.functionDefinition(
              "f",
              [],
              AST.blockExpression([AST.returnStatement(AST.integerLiteral(11))]),
            ),
            AST.functionDefinition(
              "hidden",
              [],
              AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
              undefined,
              undefined,
              undefined,
              false,
              true,
            ),
          ],
          [],
          AST.packageStatement(["dynp"]),
        ),
      },
      module: AST.module([
        AST.dynImportStatement(["dynp"], true),
        AST.assignmentExpression(
          ":=",
          AST.identifier("value"),
          AST.call(AST.identifier("f")),
        ),
        AST.identifier("value"),
      ]),
      manifest: {
        description: "Dyn import wildcard exposes public symbols",
        expect: {
          result: { kind: "i32", value: 11n },
        },
      },
    },

  {
      name: "imports/dynimport_private_selector_error",
      setupModules: {
        "package.json": AST.module(
          [
            AST.functionDefinition(
              "hidden",
              [],
              AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
              undefined,
              undefined,
              undefined,
              false,
              true,
            ),
          ],
          [],
          AST.packageStatement(["dynp"]),
        ),
      },
      module: AST.module([
        AST.dynImportStatement(["dynp"], false, [AST.importSelector("hidden")]),
      ]),
      manifest: {
        description: "Dyn import selector rejects private symbols",
        expect: {
          errors: ["dynimport error: function 'hidden' is private"],
        },
      },
    },
];

export default importsFixtures;
