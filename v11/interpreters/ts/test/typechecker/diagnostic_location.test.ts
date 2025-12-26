import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("TypeChecker diagnostic locations", () => {
  test("includes location metadata when span information is present", () => {
    const expr = AST.identifier("missing");
    (expr as any).span = { start: { line: 5, column: 3 }, end: { line: 5, column: 15 } };
    (expr as any).origin = "example.able";

    const moduleAst = AST.module([expr]);
    (moduleAst as any).span = { start: { line: 1, column: 1 }, end: { line: 6, column: 1 } };
    (moduleAst as any).origin = "example.able";

    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    const diagnostic = result.diagnostics[0]!;
    expect(diagnostic.location?.path).toBe("example.able");
    expect(diagnostic.location?.line).toBe(5);
    expect(diagnostic.location?.column).toBe(3);
  });
});
