import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";

describe("AST builders", () => {
  test("identifier and literals", () => {
    expect(AST.identifier("x")).toEqual({ type: "Identifier", name: "x" });
    expect(AST.stringLiteral("hi")).toEqual({ type: "StringLiteral", value: "hi" });
    expect(AST.integerLiteral(42)).toEqual({ type: "IntegerLiteral", value: 42 });
    expect(AST.nilLiteral()).toEqual({ type: "NilLiteral", value: null });
    expect(AST.arrayLiteral([AST.stringLiteral("a")])).toEqual({ type: "ArrayLiteral", elements: [AST.stringLiteral("a")] });
  });

  test("types", () => {
    const T = AST.simpleTypeExpression("String");
    const U = AST.simpleTypeExpression("i32");
    expect(AST.functionTypeExpression([T], U)).toEqual({ type: "FunctionTypeExpression", paramTypes: [T], returnType: U });
    expect(AST.nullableTypeExpression(T)).toEqual({ type: "NullableTypeExpression", innerType: T });
    expect(AST.resultTypeExpression(T)).toEqual({ type: "ResultTypeExpression", innerType: T });
    expect(AST.wildcardTypeExpression()).toEqual({ type: "WildcardTypeExpression" });
  });

  test("expressions & statements", () => {
    const x = AST.identifier("x");
    const one = AST.integerLiteral(1);
    const two = AST.integerLiteral(2);
    expect(AST.binaryExpression("+", one, two)).toEqual({ type: "BinaryExpression", operator: "+", left: one, right: two });
    const call = AST.functionCall(AST.identifier("add"), [one, two]);
    expect(call).toEqual({ type: "FunctionCall", callee: AST.identifier("add"), arguments: [one, two], isTrailingLambda: false });
    const blk = AST.blockExpression([call]);
    expect(blk).toEqual({ type: "BlockExpression", body: [call] });
    expect(AST.returnStatement(x)).toEqual({ type: "ReturnStatement", argument: x });
  });

  test("definitions", () => {
    const intT = AST.simpleTypeExpression("i32");
    const Point = AST.structDefinition("Point", [AST.structFieldDefinition(intT, "x"), AST.structFieldDefinition(intT, "y")], "named");
    expect(Point.type).toBe("StructDefinition");
    const fn = AST.functionDefinition("sum", [AST.functionParameter("a", intT), AST.functionParameter("b", intT)], AST.blockExpression([AST.returnStatement(AST.binaryExpression("+", AST.identifier("a"), AST.identifier("b")))]), intT);
    expect(fn.type).toBe("FunctionDefinition");
    const mod = AST.module([Point, fn]);
    expect(mod).toEqual({ type: "Module", body: [Point, fn], imports: [], package: undefined });
  });
});

describe("DSL helpers", () => {
  test("aliases for literals and identifier", () => {
    expect(AST.id("n")).toEqual(AST.identifier("n"));
    expect(AST.str("s")).toEqual(AST.stringLiteral("s"));
    expect(AST.int(3)).toEqual(AST.integerLiteral(3));
    expect(AST.nil()).toEqual(AST.nilLiteral());
  });

  test("call, block, assign, member/index", () => {
    const call = AST.call("print", AST.str("hello"));
    expect(call).toEqual(AST.functionCall(AST.identifier("print"), [AST.stringLiteral("hello")]));

    const b = AST.block(AST.ret());
    expect(b).toEqual(AST.blockExpression([AST.returnStatement()]));

    const assignId = AST.assign("x", AST.int(1));
    expect(assignId).toEqual(AST.assignmentExpression(":=", AST.identifier("x"), AST.integerLiteral(1)));

    const assignMem = AST.assignMember("obj", "field", AST.int(2));
    expect(assignMem.type).toBe("AssignmentExpression");

    const idxAssign = AST.assignIndex("arr", AST.int(0), AST.int(9));
    expect(idxAssign.type).toBe("AssignmentExpression");
  });

  test("iff and forIn helpers", () => {
    const ie = AST.iff(AST.bool(true), AST.ret());
    expect(ie.type).toBe("IfExpression");

    const fe = AST.forIn("x", AST.id("xs"), AST.ret());
    expect(fe.type).toBe("ForLoop");
  });
});


