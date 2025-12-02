import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - index expressions via Index impl", () => {
  test("invokes Index/IndexMut implementations on custom types", () => {
    const I = new InterpreterV10();

    const boxStruct = AST.structDefinition(
      "Box",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );
    I.evaluate(boxStruct);

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
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")],
    );
    I.evaluate(indexImpl);

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
          AST.blockExpression([
            AST.assignmentExpression("=", AST.implicitMemberExpression("value"), AST.identifier("value")),
            AST.returnStatement(AST.nilLiteral()),
          ]),
          AST.simpleTypeExpression("void"),
        ),
      ],
      undefined,
      undefined,
      [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")],
    );
    I.evaluate(indexMutImpl);

    const boxLiteral = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(1), "value")],
      false,
      "Box",
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("b"), boxLiteral));

    const initial = I.evaluate(AST.indexExpression(AST.identifier("b"), AST.integerLiteral(0)));
    expect(initial).toEqual({ kind: "i32", value: 1n });

    I.evaluate(
      AST.assignmentExpression(
        "=",
        AST.indexExpression(AST.identifier("b"), AST.integerLiteral(0)),
        AST.integerLiteral(9),
      ),
    );

    const updated = I.evaluate(AST.memberAccessExpression(AST.identifier("b"), "value"));
    expect(updated).toEqual({ kind: "i32", value: 9n });
  });
});
