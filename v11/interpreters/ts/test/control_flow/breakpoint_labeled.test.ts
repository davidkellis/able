import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - labeled breakpoint/break", () => {
  test("break to label returns value and unwinds through loop", () => {
    const I = new InterpreterV10();
    const body = AST.blockExpression([
      AST.assignmentExpression(":=", AST.identifier("sum"), AST.integerLiteral(0)),
      AST.forLoop(
        AST.identifier("n"),
        AST.rangeExpression(AST.integerLiteral(1), AST.integerLiteral(5), true),
        AST.blockExpression([
          AST.assignmentExpression("=", AST.identifier("sum"), AST.binaryExpression("+", AST.identifier("sum"), AST.identifier("n"))),
          AST.ifExpression(
            AST.binaryExpression("==", AST.identifier("n"), AST.integerLiteral(3)),
            AST.blockExpression([AST.breakStatement("exit", AST.stringLiteral("done"))]),
            []
          )
        ])
      ),
      AST.stringLiteral("fallthrough")
    ]);
    const bp = AST.breakpointExpression("exit", body);
    const res = I.evaluate(bp as any);
    expect(res).toEqual({ kind: "String", value: "done" });
    // ensure loop didn't error
  });

  test("no break: returns last body expr", () => {
    const I = new InterpreterV10();
    const body = AST.blockExpression([
      AST.integerLiteral(1), AST.integerLiteral(2)
    ]);
    const bp = AST.breakpointExpression("label", body);
    const res = I.evaluate(bp as any);
    expect(res).toEqual({ kind: "i32", value: 2n });
  });
});
