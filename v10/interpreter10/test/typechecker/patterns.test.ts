import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker typed patterns", () => {
  test("does not emit diagnostics when annotation mismatches subject in assignment", () => {
    const checker = new TypeChecker();
    const module = AST.module([
      AST.assignmentExpression(
        ":=",
        AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("string")),
        AST.integerLiteral(1),
      ) as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });
});
