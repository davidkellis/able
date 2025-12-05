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
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0]?.message).toContain("function expects 2 arguments, got 1");
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
      AST.simpleTypeExpression("string"),
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
});
