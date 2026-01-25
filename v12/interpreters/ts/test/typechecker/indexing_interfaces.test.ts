import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

function makeIndexInterfaces(): AST.InterfaceDefinition[] {
  const indexGenerics = [AST.genericParameter("Idx"), AST.genericParameter("Val")];
  const index = AST.interfaceDefinition(
    "Index",
    [
      AST.functionSignature(
        "get",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("idx", AST.simpleTypeExpression("Idx")),
        ],
        AST.simpleTypeExpression("Val"),
      ),
    ],
    indexGenerics,
  );

  const indexMut = AST.interfaceDefinition(
    "IndexMut",
    [
      AST.functionSignature(
        "set",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("idx", AST.simpleTypeExpression("Idx")),
          AST.functionParameter("value", AST.simpleTypeExpression("Val")),
        ],
        AST.simpleTypeExpression("void"),
      ),
    ],
    indexGenerics,
  );

  return [index, indexMut];
}

function makeBoxStruct(): AST.StructDefinition {
  return AST.structDefinition(
    "Box",
    [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
    "named",
  );
}

describe("TypeChecker index interfaces", () => {
  test("assigns through IndexMut implementation", () => {
    const [index, indexMut] = makeIndexInterfaces();
    const boxStruct = makeBoxStruct();
    const indexImpl = AST.implementationDefinition(
      "Index",
      AST.simpleTypeExpression("Box"),
      [
        AST.functionDefinition(
          "get",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("idx", AST.simpleTypeExpression("i32")),
          ],
          AST.blockExpression([AST.returnStatement(AST.implicitMemberExpression("value"))]),
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
    const indexMutImpl = AST.implementationDefinition(
      "IndexMut",
      AST.simpleTypeExpression("Box"),
      [
        AST.functionDefinition(
          "set",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("idx", AST.simpleTypeExpression("i32")),
            AST.functionParameter("value", AST.simpleTypeExpression("i32")),
          ],
          AST.blockExpression([AST.returnStatement(AST.nilLiteral())]),
          AST.simpleTypeExpression("void"),
          undefined,
          undefined,
          true,
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")],
    );
    const bindBox = AST.assignmentExpression(
      ":=",
      AST.typedPattern(
        AST.identifier("b"),
        AST.genericTypeExpression(AST.simpleTypeExpression("IndexMut"), [
          AST.simpleTypeExpression("i32"),
          AST.simpleTypeExpression("i32"),
        ]),
      ),
      AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "value")], false, "Box"),
    );
    const writeIndex = AST.assignmentExpression(
      "=",
      AST.indexExpression(AST.identifier("b"), AST.integerLiteral(0)),
      AST.integerLiteral(5),
    );
    const module = AST.module(
      [index, indexMut, boxStruct, indexImpl, indexMutImpl, bindBox, writeIndex],
      [],
      AST.packageStatement(["app"]),
    );

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("requires IndexMut implementation for [] assignment", () => {
    const [index] = makeIndexInterfaces();
    const boxStruct = makeBoxStruct();
    const indexImpl = AST.implementationDefinition(
      "Index",
      AST.simpleTypeExpression("Box"),
      [
        AST.functionDefinition(
          "get",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("idx", AST.simpleTypeExpression("i32")),
          ],
          AST.blockExpression([AST.returnStatement(AST.implicitMemberExpression("value"))]),
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
    const bindBox = AST.assignmentExpression(
      ":=",
      AST.typedPattern(
        AST.identifier("b"),
        AST.genericTypeExpression(AST.simpleTypeExpression("Index"), [
          AST.simpleTypeExpression("i32"),
          AST.simpleTypeExpression("i32"),
        ]),
      ),
      AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "value")], false, "Box"),
    );
    const writeIndex = AST.assignmentExpression(
      "=",
      AST.indexExpression(AST.identifier("b"), AST.integerLiteral(0)),
      AST.integerLiteral(2),
    );
    const module = AST.module(
      [index, boxStruct, indexImpl, bindBox, writeIndex],
      [],
      AST.packageStatement(["app"]),
    );

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    const diag = diagnostics.find((d) => d.message.includes("IndexMut implementation"));
    expect(diag?.message ?? "").toContain("IndexMut");
  });

  test("rejects := on index assignment", () => {
    const [index, indexMut] = makeIndexInterfaces();
    const boxStruct = makeBoxStruct();
    const module = AST.module(
      [
        index,
        indexMut,
        boxStruct,
        AST.assignmentExpression(
          ":=",
          AST.indexExpression(AST.identifier("b"), AST.integerLiteral(0)),
          AST.integerLiteral(1),
        ),
      ],
      [],
      AST.packageStatement(["app"]),
    );

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(module);
    const diag = diagnostics.find((d) => d.message.includes("cannot use :="));
    expect(diag?.message).toContain("index assignment");
  });
});
