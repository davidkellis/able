import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

function blockReturningString(): AST.BlockExpression {
  return AST.blockExpression([AST.stringLiteral("done") as unknown as AST.Statement]);
}

describe("typechecker concurrency expressions", () => {
  test("reports diagnostic when iterating over proc handles", () => {
    const checker = new TypeChecker();
    const loop = AST.forLoop(
      AST.identifier("value"),
      AST.procExpression(blockReturningString()),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("Proc");
  });

  test("reports diagnostic when iterating over future handles", () => {
    const checker = new TypeChecker();
    const loop = AST.forLoop(
      AST.identifier("value"),
      AST.spawnExpression(blockReturningString()),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("Future");
  });

  test("reports diagnostic when calling proc_yield outside async context", () => {
    const checker = new TypeChecker();
    const call = AST.functionCall(AST.identifier("proc_yield"), []);
    const module = AST.module([call as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("proc_yield() may only be called");
  });

  test("allows proc_yield inside proc expression", () => {
    const checker = new TypeChecker();
    const call = AST.functionCall(AST.identifier("proc_yield"), []);
    const procExpr = AST.procExpression(call as unknown as AST.FunctionCall);
    const module = AST.module([procExpr as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });
});
