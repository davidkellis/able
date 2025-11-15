import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - pipes", () => {
  test("topic reference uses pipe subject", () => {
    const I = new InterpreterV10();
    const expr = AST.binaryExpression(
      "|>",
      AST.integerLiteral(5),
      AST.binaryExpression("+", AST.topicReferenceExpression(), AST.integerLiteral(3))
    );
    const result = I.evaluate(expr);
    expect(result).toEqual({ kind: "i32", value: 8n });
  });

  test("topic used inside function call", () => {
    const I = new InterpreterV10();
    const add = AST.functionDefinition(
      "add",
      [AST.functionParameter("left"), AST.functionParameter("right")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("left"), AST.identifier("right")))]),
    );
    I.evaluate(add);
    const expr = AST.binaryExpression(
      "|>",
      AST.integerLiteral(4),
      AST.functionCall(AST.identifier("add"), [AST.topicReferenceExpression(), AST.integerLiteral(1)]),
    );
    expect(I.evaluate(expr)).toEqual({ kind: "i32", value: 5n });
  });

  test("implicit member shorthand binds receiver once", () => {
    const I = new InterpreterV10();
    const Box = AST.structDefinition(
      "Box",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named"
    );
    I.evaluate(Box);

    const increment = AST.functionDefinition(
      "increment",
      [],
      AST.blockExpression([
        AST.returnStatement(
          AST.binaryExpression("+", AST.implicitMemberExpression("value"), AST.integerLiteral(1))
        ),
      ]),
      undefined,
      undefined,
      undefined,
      true
    );
    const doubleFn = AST.functionDefinition(
      "double",
      [],
      AST.blockExpression([
        AST.returnStatement(
          AST.binaryExpression("*", AST.implicitMemberExpression("value"), AST.integerLiteral(2))
        ),
      ]),
      undefined,
      undefined,
      undefined,
      true
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Box"), [increment, doubleFn]);
    I.evaluate(methods);

    const boxLiteral = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(5), "value")],
      false,
      "Box"
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("box"), boxLiteral));

    const first = I.evaluate(
      AST.binaryExpression("|>", AST.identifier("box"), AST.implicitMemberExpression("increment"))
    );
    const second = I.evaluate(
      AST.binaryExpression("|>", AST.identifier("box"), AST.implicitMemberExpression("double"))
    );

    expect(first).toEqual({ kind: "i32", value: 6n });
    expect(second).toEqual({ kind: "i32", value: 10n });
  });

  test("placeholder callable as pipe RHS", () => {
    const I = new InterpreterV10();
    const add = AST.functionDefinition(
      "add",
      [AST.functionParameter("left"), AST.functionParameter("right")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("left"), AST.identifier("right")))]),
    );
    I.evaluate(add);
    const expr = AST.binaryExpression(
      "|>",
      AST.integerLiteral(9),
      AST.functionCall(AST.identifier("add"), [AST.placeholderExpression(), AST.integerLiteral(1)]),
    );
    expect(I.evaluate(expr)).toEqual({ kind: "i32", value: 10n });
  });

  test("UFCS free function via pipe", () => {
    const I = new InterpreterV10();
    const Point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named"
    );
    I.evaluate(Point);
    const translate = AST.functionDefinition(
      "translate",
      [AST.functionParameter("point")],
      AST.blockExpression([
        AST.assignmentExpression(
          "=",
          AST.memberAccessExpression(AST.identifier("point"), "x"),
          AST.binaryExpression("+", AST.memberAccessExpression(AST.identifier("point"), "x"), AST.integerLiteral(5))
        ),
        AST.returnStatement(AST.identifier("point")),
      ]),
    );
    I.evaluate(translate);
    const pointLiteral = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(2), "x")],
      false,
      "Point"
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("point"), pointLiteral));
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("result"),
        AST.binaryExpression("|>", AST.identifier("point"), AST.identifier("translate")),
      ),
    );
    const result = I.evaluate(AST.memberAccessExpression(AST.identifier("result"), "x"));
    expect(result).toEqual({ kind: "i32", value: 7n });
  });
});
