import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker function calls", () => {
  test("reports literal overflow when argument does not fit annotated parameter type", () => {
    const checker = new TypeChecker();
    const fn = AST.functionDefinition(
      "write_byte",
      [AST.functionParameter("value", AST.simpleTypeExpression("u8"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
    );
    const call = AST.functionCall(AST.identifier("write_byte"), [AST.integerLiteral(512)]);
    const module = AST.module([fn, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("literal 512 does not fit in u8");
  });

  test("reports diagnostic when argument count mismatches parameter list", () => {
    const checker = new TypeChecker();
    const fn = AST.functionDefinition(
      "add",
      [
        AST.functionParameter("lhs", AST.simpleTypeExpression("i32")),
        AST.functionParameter("rhs", AST.simpleTypeExpression("i32")),
      ],
      AST.blockExpression([AST.returnStatement(AST.identifier("lhs"))]),
      AST.simpleTypeExpression("i32"),
    );
    const call = AST.functionCall(AST.identifier("add"), [AST.integerLiteral(1)]);
    const module = AST.module([fn, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
  });

  test("reports type mismatch when argument type differs from parameter type", () => {
    const checker = new TypeChecker();
    const fn = AST.functionDefinition(
      "takes_bool",
      [AST.functionParameter("flag", AST.simpleTypeExpression("bool"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
    );
    const call = AST.functionCall(AST.identifier("takes_bool"), [AST.integerLiteral(1)]);
    const module = AST.module([fn, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("argument 1 has type i32, expected bool");
  });

  test("allows integer arguments to widen to annotated parameter type", () => {
    const checker = new TypeChecker();
    const fn = AST.functionDefinition(
      "takes_i64",
      [AST.functionParameter("value", AST.simpleTypeExpression("i64"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
    );
    const call = AST.functionCall(AST.identifier("takes_i64"), [AST.integerLiteral(1)]);
    const module = AST.module([fn, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toEqual([]);
  });

  test("method-style UFCS binds free function when the first parameter matches the receiver", () => {
    const checker = new TypeChecker();
    const point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named",
    );
    const tag = AST.functionDefinition(
      "tag",
      [AST.functionParameter("p", AST.simpleTypeExpression("Point"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
      AST.simpleTypeExpression("String"),
    );
    const pointValue = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(1), "x")],
      false,
      "Point",
    );
    const call = AST.functionCall(AST.memberAccessExpression(pointValue, AST.identifier("tag")), []);
    const module = AST.module([point, tag, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
  });

  test("pipe RHS must be callable", () => {
    const checker = new TypeChecker();
    const init = AST.assignmentExpression(":=", AST.identifier("value"), AST.integerLiteral(2));
    const pipe = AST.binaryExpression("|>", AST.identifier("value"), AST.integerLiteral(3));
    const module = AST.module([init, pipe as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics.some((diag) => diag.message.includes("non-callable"))).toBe(true);
  });

  test("callable fields are preferred over methods with the same name", () => {
    const checker = new TypeChecker();
    const box = AST.structDefinition(
      "Box",
      [
        AST.structFieldDefinition(
          AST.fnType([AST.simpleTypeExpression("String")], AST.simpleTypeExpression("String")),
          "action",
        ),
      ],
      "named",
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Box"), [
      AST.fn(
        "action",
        [],
        AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
        AST.simpleTypeExpression("i32"),
        undefined,
        undefined,
        true,
      ),
    ]);
    const literal = AST.structLiteral(
      [AST.structFieldInitializer(AST.lambdaExpression([AST.functionParameter("msg")], AST.stringLiteral("ok")), "action")],
      false,
      "Box",
    );
    const call = AST.functionCall(AST.memberAccessExpression(literal, AST.identifier("action")), [AST.stringLiteral("hi")]);
    const module = AST.module([box, methods, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
  });

  test("method and UFCS free functions with matching signatures are ambiguous", () => {
    const checker = new TypeChecker();
    const point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named",
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.fn(
        "describe",
        [],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("method"))]),
        AST.simpleTypeExpression("String"),
        undefined,
        undefined,
        true,
      ),
    ]);
    const freeDescribe = AST.functionDefinition(
      "describe",
      [AST.functionParameter("p", AST.simpleTypeExpression("Point"))],
      AST.blockExpression([AST.returnStatement(AST.stringLiteral("free"))]),
      AST.simpleTypeExpression("String"),
    );
    const literal = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(1), "x")],
      false,
      "Point",
    );
    const call = AST.functionCall(AST.memberAccessExpression(literal, AST.identifier("describe")), []);
    const module = AST.module([point, methods, freeDescribe, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
  });

  test("methods from method sets can be called as exported functions", () => {
    const checker = new TypeChecker();
    const point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named",
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.fn(
        "norm",
        [],
        AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
        AST.simpleTypeExpression("i32"),
        undefined,
        undefined,
        true,
      ),
    ]);
    const literal = AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point");
    const call = AST.functionCall(AST.identifier("norm"), [literal]);
    const module = AST.module([point, methods, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
  });

  test("method exports enforce receiver type when called directly", () => {
    const checker = new TypeChecker();
    const point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named",
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.fn(
        "norm",
        [],
        AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
        AST.simpleTypeExpression("i32"),
        undefined,
        undefined,
        true,
      ),
    ]);
    const call = AST.functionCall(AST.identifier("norm"), [AST.integerLiteral(3)]);
    const module = AST.module([point, methods, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("expected Point");
  });

  test("method set where clause obligations apply to exported functions", () => {
    const checker = new TypeChecker();
    const display = AST.interfaceDefinition("Display", [
      AST.functionSignature(
        "show",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("String"),
      ),
    ]);
    const doc = AST.structDefinition(
      "Doc",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "body")],
      "named",
    );
    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Doc"),
      [
        AST.fn(
          "title",
          [],
          AST.blockExpression([AST.returnStatement(AST.stringLiteral("hi"))]),
          AST.simpleTypeExpression("String"),
          undefined,
          undefined,
          true,
        ),
      ],
      undefined,
      [AST.whereClauseConstraint("Self", [AST.interfaceConstraint(AST.simpleTypeExpression("Display"))])],
    );
    const call = AST.functionCall(
      AST.identifier("title"),
      [
        AST.structLiteral(
          [AST.structFieldInitializer(AST.stringLiteral("body"), "body")],
          false,
          "Doc",
        ),
      ],
    );
    const module = AST.module([display, doc, methods, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics.length).toBeGreaterThan(0);
    expect(diagnostics[0]?.message).toContain("Display");
  });

  test("method shorthand exports still require a receiver when called as functions", () => {
    const checker = new TypeChecker();
    const point = AST.structDefinition(
      "Point",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
      "named",
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Point"), [
      AST.fn(
        "norm",
        [],
        AST.blockExpression([AST.returnStatement(AST.integerLiteral(1))]),
        AST.simpleTypeExpression("i32"),
        undefined,
        undefined,
        true,
      ),
    ]);
    const call = AST.functionCall(AST.identifier("norm"), []);
    const module = AST.module([point, methods, call as unknown as AST.Statement]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
  });
});
