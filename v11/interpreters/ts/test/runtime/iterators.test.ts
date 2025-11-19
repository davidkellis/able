import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";
import type { V10Value } from "../../src/interpreter";

const callNext = (target: string) =>
  AST.functionCall(AST.memberAccessExpression(AST.identifier(target), "next"), []);

const assign = (name: string, expr: AST.Expression) =>
  AST.assignmentExpression(":=", AST.identifier(name), expr);

const update = (name: string, expr: AST.Expression) =>
  AST.assignmentExpression("=", AST.identifier(name), expr);

describe("v11 interpreter - iterator literals", () => {
  test("iterator literal yields values lazily", () => {
    const I = new InterpreterV10();

    I.evaluate(assign("count", AST.integerLiteral(0)));
    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          update(
            "count",
            AST.binaryExpression("+", AST.identifier("count"), AST.integerLiteral(1))
          ),
          AST.yieldStatement(AST.identifier("count")),
          update(
            "count",
            AST.binaryExpression("+", AST.identifier("count"), AST.integerLiteral(1))
          ),
          AST.yieldStatement(AST.identifier("count")),
        ])
      )
    );

    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 0n });

    const first = I.evaluate(callNext("iter"));
    expect(first).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 1n });

    const second = I.evaluate(callNext("iter"));
    expect(second).toEqual({ kind: "i32", value: 2n });
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 2n });

    const third = I.evaluate(callNext("iter")) as V10Value;
    expect(third.kind).toBe("iterator_end");

    const fourth = I.evaluate(callNext("iter")) as V10Value;
    expect(fourth.kind).toBe("iterator_end");
  });

  test("for loop drives iterator lazily", () => {
    const I = new InterpreterV10();

    I.evaluate(assign("sum", AST.integerLiteral(0)));
    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.yieldStatement(AST.integerLiteral(10)),
          AST.yieldStatement(AST.integerLiteral(20)),
          AST.yieldStatement(AST.integerLiteral(30)),
        ])
      )
    );

    const loop = AST.forLoop(
      AST.identifier("item"),
      AST.identifier("iter"),
      AST.blockExpression([
        update(
          "sum",
          AST.binaryExpression("+", AST.identifier("sum"), AST.identifier("item"))
        ),
        AST.identifier("item"),
      ])
    );

    const last = I.evaluate(loop);
    expect(last).toEqual({ kind: "nil", value: null });
    expect(I.evaluate(AST.identifier("sum"))).toEqual({ kind: "i32", value: 60n });
  });

  test("gen.stop terminates iteration", () => {
    const I = new InterpreterV10();

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.yieldStatement(AST.integerLiteral(1)),
          AST.functionCall(AST.memberAccessExpression(AST.identifier("gen"), "stop"), []),
          AST.yieldStatement(AST.integerLiteral(99)),
        ])
      )
    );

    const first = I.evaluate(callNext("iter"));
    expect(first).toEqual({ kind: "i32", value: 1n });

    const second = I.evaluate(callNext("iter")) as V10Value;
    expect(second.kind).toBe("iterator_end");
  });

  test("iterator next propagates runtime errors", () => {
    const I = new InterpreterV10();

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.raiseStatement(AST.stringLiteral("boom")),
        ])
      )
    );

    expect(() => I.evaluate(callNext("iter"))).toThrow("RaiseSignal");
  });

  test("iterator close terminates subsequent next calls", () => {
    const I = new InterpreterV10();

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.yieldStatement(AST.integerLiteral(1)),
        ])
      )
    );

    I.evaluate(AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "close"), []));

    const value = I.evaluate(callNext("iter")) as V10Value;
    expect(value.kind).toBe("iterator_end");
  });

  test("for loop inside generator yields each element once", () => {
    const I = new InterpreterV10();

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.forLoop(
            AST.identifier("item"),
            AST.arrayLiteral([AST.integerLiteral(1), AST.integerLiteral(2), AST.integerLiteral(3)]),
            AST.blockExpression([
              AST.yieldStatement(AST.identifier("item")),
            ])
          ),
        ])
      )
    );

    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 2n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 3n });
    const done = I.evaluate(callNext("iter")) as V10Value;
    expect(done.kind).toBe("iterator_end");
  });

  test("while loop resumes after yield within body", () => {
    const I = new InterpreterV10();

    I.evaluate(assign("count", AST.integerLiteral(0)));

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.whileLoop(
            AST.binaryExpression("<", AST.identifier("count"), AST.integerLiteral(3)),
            AST.blockExpression([
              AST.yieldStatement(AST.identifier("count")),
              AST.assignmentExpression(
                "=",
                AST.identifier("count"),
                AST.binaryExpression("+", AST.identifier("count"), AST.integerLiteral(1))
              ),
            ])
          ),
        ])
      )
    );

    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 0n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 2n });
    const done = I.evaluate(callNext("iter")) as V10Value;
    expect(done.kind).toBe("iterator_end");
    expect(I.evaluate(AST.identifier("count"))).toEqual({ kind: "i32", value: 3n });
  });

  test("if expression resumes without re-running condition", () => {
    const I = new InterpreterV10();

    I.evaluate(assign("calls", AST.integerLiteral(0)));

    I.evaluate(
      AST.functionDefinition(
        "tick",
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("calls"),
            AST.binaryExpression("+", AST.identifier("calls"), AST.integerLiteral(1))
          ),
          AST.returnStatement(AST.booleanLiteral(true)),
        ])
      )
    );

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.ifExpression(
            AST.functionCall(AST.identifier("tick"), []),
            AST.blockExpression([
              AST.yieldStatement(AST.integerLiteral(1)),
              AST.yieldStatement(AST.integerLiteral(2)),
            ])
          ),
        ])
      )
    );

    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 2n });
    const done = I.evaluate(callNext("iter")) as V10Value;
    expect(done.kind).toBe("iterator_end");
    expect(I.evaluate(AST.identifier("calls"))).toEqual({ kind: "i32", value: 1n });
  });

  test("match expression resumes without re-running subject or guard", () => {
    const I = new InterpreterV10();

    I.evaluate(assign("subject_calls", AST.integerLiteral(0)));
    I.evaluate(assign("guard_calls", AST.integerLiteral(0)));

    I.evaluate(
      AST.functionDefinition(
        "getSubject",
        [],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("subject_calls"),
            AST.binaryExpression("+", AST.identifier("subject_calls"), AST.integerLiteral(1))
          ),
          AST.returnStatement(AST.integerLiteral(1)),
        ])
      )
    );

    I.evaluate(
      AST.functionDefinition(
        "guardCheck",
        [AST.functionParameter("value")],
        AST.blockExpression([
          AST.assignmentExpression(
            "=",
            AST.identifier("guard_calls"),
            AST.binaryExpression("+", AST.identifier("guard_calls"), AST.integerLiteral(1))
          ),
          AST.returnStatement(AST.booleanLiteral(true)),
        ])
      )
    );

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral([
          AST.matchExpression(
            AST.functionCall(AST.identifier("getSubject"), []),
            [
              AST.matchClause(
                AST.literalPattern(AST.integerLiteral(1)),
                AST.blockExpression([
                  AST.yieldStatement(AST.integerLiteral(10)),
                  AST.yieldStatement(AST.integerLiteral(11)),
                ]),
                AST.functionCall(AST.identifier("guardCheck"), [AST.integerLiteral(1)])
              ),
            ]
          ),
        ])
      )
    );

    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 10n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 11n });
    const done = I.evaluate(callNext("iter")) as V10Value;
    expect(done.kind).toBe("iterator_end");
    expect(I.evaluate(AST.identifier("subject_calls"))).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(AST.identifier("guard_calls"))).toEqual({ kind: "i32", value: 1n });
  });

  test("iterator literal honors custom binding names and annotations", () => {
    const I = new InterpreterV10();

    I.evaluate(
      assign(
        "iter",
        AST.iteratorLiteral(
          [
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("driver"), "yield"),
              [AST.integerLiteral(1)]
            ),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("driver"), "yield"),
              [AST.integerLiteral(2)]
            ),
          ],
          "driver",
          AST.simpleTypeExpression("i32")
        )
      )
    );

    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 1n });
    expect(I.evaluate(callNext("iter"))).toEqual({ kind: "i32", value: 2n });
    const done = I.evaluate(callNext("iter")) as V10Value;
    expect(done.kind).toBe("iterator_end");
  });
});
