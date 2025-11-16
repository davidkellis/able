import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - placeholders", () => {
  test("partial application with implicit placeholder", () => {
    const I = new InterpreterV10();
    const add = AST.functionDefinition(
      "add",
      [AST.functionParameter("left"), AST.functionParameter("right")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("left"), AST.identifier("right")))]),
    );
    I.evaluate(add);
    const call = AST.functionCall(
      AST.functionCall(AST.identifier("add"), [AST.placeholderExpression(), AST.integerLiteral(10)]),
      [AST.integerLiteral(5)],
    );
    expect(I.evaluate(call)).toEqual({ kind: "i32", value: 15n });
  });

  test("placeholder native function invocation", () => {
    const I = new InterpreterV10();
    const add = AST.functionDefinition(
      "add",
      [AST.functionParameter("left"), AST.functionParameter("right")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("left"), AST.identifier("right")))]),
    );
    I.evaluate(add);
    const partial = I.evaluate(
      AST.functionCall(AST.identifier("add"), [AST.placeholderExpression(), AST.integerLiteral(1)]),
    );
    expect(partial.kind).toBe("native_function");
    const nine = I.evaluate(AST.integerLiteral(9));
    const result = (partial.kind === "native_function") ? partial.impl(I, [nine]) : { kind: "nil", value: null };
    expect(result).toEqual({ kind: "i32", value: 10n });
  });

  test("mixed explicit and implicit placeholders", () => {
    const I = new InterpreterV10();
    const combine = AST.functionDefinition(
      "combine",
      [
        AST.functionParameter("a"),
        AST.functionParameter("b"),
        AST.functionParameter("c"),
      ],
      AST.blockExpression([
        AST.returnStatement(
          AST.binaryExpression(
            "+",
            AST.binaryExpression("*", AST.identifier("a"), AST.integerLiteral(100)),
            AST.binaryExpression(
              "+",
              AST.binaryExpression("*", AST.identifier("b"), AST.integerLiteral(10)),
              AST.identifier("c"),
            ),
          ),
        ),
      ]),
    );
    I.evaluate(combine);
    const call = AST.functionCall(
      AST.functionCall(
        AST.identifier("combine"),
        [
          AST.placeholderExpression(1),
          AST.placeholderExpression(),
          AST.placeholderExpression(3),
        ],
      ),
      [AST.integerLiteral(7), AST.integerLiteral(8), AST.integerLiteral(9)],
    );
    expect(I.evaluate(call)).toEqual({ kind: "i32", value: 789n });
  });

  test("lambda containing placeholder evaluates to callable", () => {
    const I = new InterpreterV10();
    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("builder"),
        AST.lambdaExpression([], AST.blockExpression([AST.placeholderExpression()])),
      ),
    );
    const builder = I.evaluate(AST.identifier("builder"));
    expect(builder?.kind).toBe("function");
    const result = I.evaluate(AST.functionCall(AST.identifier("builder"), []));
    expect(result.kind).toBe("native_function");
    expect(result.arity).toBe(1);
  });
});
