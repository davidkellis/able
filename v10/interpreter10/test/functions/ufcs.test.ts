import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v10 interpreter - UFCS fallback", () => {
  test("free function add(a,b) callable as 4.add(5)", () => {
    const I = new InterpreterV10();
    const add = AST.functionDefinition(
      "add",
      [AST.functionParameter("a"), AST.functionParameter("b")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))])
    );
    I.evaluate(add);
    const call = AST.functionCall(AST.memberAccessExpression(AST.integerLiteral(4), "add"), [AST.integerLiteral(5)]);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 9 });
  });

  test("struct receiver with free function move(Point, dx) called as p.move(dx)", () => {
    const I = new InterpreterV10();
    const Point = AST.structDefinition("Point", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")], "named");
    I.evaluate(Point);
    const move = AST.functionDefinition(
      "move",
      [AST.functionParameter("p"), AST.functionParameter("dx")],
      AST.blockExpression([
        AST.assignmentExpression("=", AST.memberAccessExpression(AST.identifier("p"), "x"), AST.binaryExpression("+", AST.memberAccessExpression(AST.identifier("p"), "x"), AST.identifier("dx"))),
        AST.returnStatement(AST.identifier("p"))
      ])
    );
    I.evaluate(move);
    const p = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), p));
    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("p"), "move"), [AST.integerLiteral(3)]);
    const res = I.evaluate(call);
    expect(I.evaluate(AST.memberAccessExpression(AST.identifier("p"), "x"))).toEqual({ kind: 'i32', value: 4 });
    expect(res.kind).toBe('struct_instance');
  });
});


