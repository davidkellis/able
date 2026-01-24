import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { Interpreter } from "../../src/interpreter";
import type { RuntimeValue } from "../../src/interpreter";

const NIL: RuntimeValue = { kind: "nil", value: null };

describe("v11 interpreter - async native suspension", () => {
  test("async native calls suspend only the current proc", async () => {
    const I = new Interpreter();
    I.evaluate(AST.assign("record", AST.stringLiteral("")));

    const asyncWait = I.makeNativeFunction("async_wait", 0, () => {
      return new Promise<RuntimeValue>((resolve) => {
        setTimeout(() => resolve(NIL), 0);
      });
    });
    I.globals.define("async_wait", asyncWait);

    const append = (value: string) =>
      AST.assignmentExpression(
        "=",
        AST.identifier("record"),
        AST.binaryExpression("+", AST.identifier("record"), AST.stringLiteral(value)),
      );

    I.evaluate(
      AST.spawnExpression(
        AST.blockExpression([
          append("A"),
          AST.functionCall(AST.identifier("async_wait"), []),
          append("a"),
          AST.integerLiteral(0),
        ]),
      ),
    );

    I.evaluate(
      AST.spawnExpression(
        AST.blockExpression([
          append("B"),
          AST.integerLiteral(0),
        ]),
      ),
    );

    I.procFlush();
    const first = I.evaluate(AST.identifier("record"));
    expect(first).toEqual({ kind: "String", value: "AB" });

    await new Promise((resolve) => setTimeout(resolve, 0));
    I.procFlush();
    const second = I.evaluate(AST.identifier("record"));
    expect(second).toEqual({ kind: "String", value: "ABa" });
  });
});
