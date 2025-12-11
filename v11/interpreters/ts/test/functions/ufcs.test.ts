import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { InterpreterV10 } from "../../src/interpreter";

describe("v11 interpreter - UFCS fallback", () => {
  test("free function add(a,b) callable as 4.add(5)", () => {
    const I = new InterpreterV10();
    const add = AST.functionDefinition(
      "add",
      [AST.functionParameter("a"), AST.functionParameter("b")],
      AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))])
    );
    I.evaluate(add);
    const call = AST.functionCall(AST.memberAccessExpression(AST.integerLiteral(4), "add"), [AST.integerLiteral(5)]);
    expect(I.evaluate(call)).toEqual({ kind: 'i32', value: 9n });
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
    expect(I.evaluate(AST.memberAccessExpression(AST.identifier("p"), "x"))).toEqual({ kind: 'i32', value: 4n });
    expect(res.kind).toBe('struct_instance');
  });

  test("prefers UFCS candidate whose first parameter matches the receiver type", () => {
    const I = new InterpreterV10();
    const point = AST.structDefinition("Point", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")], "named");
    I.evaluate(point);
    const pointTag = AST.functionDefinition(
      "tag",
      [AST.functionParameter("p", AST.simpleTypeExpression("Point"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("point"))]),
      AST.simpleTypeExpression("String"),
    );
    const stringTag = AST.functionDefinition(
      "tag",
      [AST.functionParameter("s", AST.simpleTypeExpression("String"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("String"))]),
      AST.simpleTypeExpression("String"),
    );
    I.evaluate(pointTag);
    I.evaluate(stringTag);

    I.evaluate(
      AST.assignmentExpression(
        ":=",
        AST.identifier("p"),
        AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point"),
      ),
    );
    const callOnPoint = AST.functionCall(AST.memberAccessExpression(AST.identifier("p"), "tag"), []);
    const callOnString = AST.functionCall(AST.memberAccessExpression(AST.stringLiteral("hi"), "tag"), []);

    expect(I.evaluate(callOnPoint)).toEqual({ kind: "String", value: "point" });
    expect(I.evaluate(callOnString)).toEqual({ kind: "String", value: "String" });
  });

  test("rejects UFCS binding when the receiver type mismatches the first parameter", () => {
    const I = new InterpreterV10();
    const point = AST.structDefinition("Point", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")], "named");
    I.evaluate(point);
    const onlyString = AST.functionDefinition(
      "label",
      [AST.functionParameter("s", AST.simpleTypeExpression("String"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("nope"))]),
      AST.simpleTypeExpression("String"),
    );
    I.evaluate(onlyString);

    const p = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point");
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), p));
    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("p"), "label"), []);

    expect(() => I.evaluate(call)).toThrow(/field or method/i);
  });

  test("callable field takes precedence over methods", () => {
    const I = new InterpreterV10();
    const box = AST.structDefinition(
      "Box",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("String"), "name"),
        AST.structFieldDefinition(undefined, "action"),
      ],
      "named",
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Box"), [
      AST.fn(
        "action",
        [],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("method"))]),
        AST.simpleTypeExpression("String"),
        undefined,
        undefined,
        true,
      ),
    ]);
    I.evaluate(box);
    I.evaluate(methods);
    const instance = AST.structLiteral(
      [
        AST.structFieldInitializer(AST.stringLiteral("ok"), "name"),
        AST.structFieldInitializer(AST.lambdaExpression([], AST.stringLiteral("field")), "action"),
      ],
      false,
      "Box",
    );
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("b"), instance));
    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("b"), "action"), []);

    expect(I.evaluate(call)).toEqual({ kind: "String", value: "field" });
  });

  test("identifier calls no longer bind via UFCS fallback", () => {
    const I = new InterpreterV10();
    const point = AST.structDefinition("Point", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")], "named");
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.fn(
        "tag",
        [],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("point"))]),
        AST.simpleTypeExpression("String"),
        undefined,
        undefined,
        true,
      ),
    ]);
    I.evaluate(point);
    I.evaluate(methods);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point")));
    const call = AST.functionCall(AST.identifier("tag"), [AST.identifier("p")]);

    expect(I.evaluate(call)).toEqual({ kind: "String", value: "point" });
  });

  test("reports ambiguity when inherent methods overlap with UFCS free functions", () => {
    const I = new InterpreterV10();
    const point = AST.structDefinition("Point", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")], "named");
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.fn(
        "describe",
        [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("method"))]),
        AST.simpleTypeExpression("String"),
        undefined,
        undefined,
        false,
      ),
    ]);
    const freeDescribe = AST.functionDefinition(
      "describe",
      [AST.functionParameter("p", AST.simpleTypeExpression("Point"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("free"))]),
      AST.simpleTypeExpression("String"),
    );
    I.evaluate(point);
    I.evaluate(methods);
    I.evaluate(freeDescribe);
    I.evaluate(AST.assignmentExpression(":=", AST.identifier("p"), AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point")));
    const call = AST.functionCall(AST.memberAccessExpression(AST.identifier("p"), "describe"), []);

    expect(I.evaluate(call)).toEqual({ kind: "String", value: "method" });
  });
});
