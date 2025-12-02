import { describe, expect, test } from "bun:test";

import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - Apply interface calls", () => {
  test("calls apply implementation when invoking callable values", () => {
    const I = new InterpreterV10();

    const multStruct = AST.structDefinition(
      "Multiplier",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "factor")],
      "named",
    );
    I.evaluate(multStruct);

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
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")],
    );
    I.evaluate(applyImpl);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("m"),
        AST.structLiteral(
          [AST.structFieldInitializer(AST.integerLiteral(3), "factor")],
          false,
          "Multiplier",
        ),
      ),
    );

    const result = I.evaluate(AST.functionCall(AST.identifier("m"), [AST.integerLiteral(5)]));
    expect(result).toEqual({ kind: "i32", value: 15n });
  });
});
