import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker return-context inference", () => {
  test("infers generic call arguments from explicit return context", () => {
    const checker = new TypeChecker();
    const hashSet = AST.structDefinition("HashSet", [], "named", [AST.genericParameter("T")]);
    const collect = AST.functionDefinition(
      "collect",
      [
        AST.functionParameter(
          "items",
          AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
        ),
      ],
      AST.blockExpression([]),
      AST.genericTypeExpression(AST.simpleTypeExpression("HashSet"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const call = AST.functionCall(AST.identifier("collect"), [AST.arrayLiteral([])]);
    const build = AST.functionDefinition(
      "build",
      [],
      AST.blockExpression([AST.returnStatement(call)]),
      AST.genericTypeExpression(AST.simpleTypeExpression("HashSet"), [AST.simpleTypeExpression("u8")]),
    );
    const module = AST.module([hashSet, collect, build]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
    expect(call.typeArguments).toHaveLength(1);
    const arg = call.typeArguments?.[0];
    expect(arg?.type).toBe("SimpleTypeExpression");
    if (arg && arg.type === "SimpleTypeExpression") {
      expect(arg.name?.name).toBe("u8");
    }
  });

  test("infers generic call arguments from implicit return context", () => {
    const checker = new TypeChecker();
    const hashSet = AST.structDefinition("HashSet", [], "named", [AST.genericParameter("T")]);
    const collect = AST.functionDefinition(
      "collect",
      [
        AST.functionParameter(
          "items",
          AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
        ),
      ],
      AST.blockExpression([]),
      AST.genericTypeExpression(AST.simpleTypeExpression("HashSet"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const call = AST.functionCall(AST.identifier("collect"), [AST.arrayLiteral([])]);
    const build = AST.functionDefinition(
      "build",
      [],
      AST.blockExpression([call]),
      AST.genericTypeExpression(AST.simpleTypeExpression("HashSet"), [AST.simpleTypeExpression("u8")]),
    );
    const module = AST.module([hashSet, collect, build]);

    const { diagnostics } = checker.checkModule(module);
    expect(diagnostics).toHaveLength(0);
    expect(call.typeArguments).toHaveLength(1);
    const arg = call.typeArguments?.[0];
    expect(arg?.type).toBe("SimpleTypeExpression");
    if (arg && arg.type === "SimpleTypeExpression") {
      expect(arg.name?.name).toBe("u8");
    }
  });
});
