import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

function buildShowInterface(): AST.InterfaceDefinition {
  return AST.interfaceDefinition("Show", [
    AST.functionSignature(
      "to_string",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.simpleTypeExpression("string"),
    ),
  ]);
}

function buildPointStruct(): AST.StructDefinition {
  return AST.structDefinition("Point", [], "named");
}

function resultStringType(): AST.TypeExpression {
  return AST.resultTypeExpression(AST.simpleTypeExpression("string"));
}

function buildShowImplementation(): AST.ImplementationDefinition {
  return AST.implementationDefinition(
    AST.identifier("Show"),
    AST.simpleTypeExpression("Point"),
    [
      AST.functionDefinition(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
        AST.simpleTypeExpression("string"),
      ),
    ],
  );
}

function buildUseShowFunction(): AST.FunctionDefinition {
  return AST.functionDefinition(
    "use_show",
    [
      AST.functionParameter("value", AST.simpleTypeExpression("T")),
    ],
    AST.blockExpression([
      AST.returnStatement(
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("value"), AST.identifier("to_string")),
          [],
          [],
        ),
      ),
    ]),
    AST.simpleTypeExpression("string"),
    [
      AST.genericParameter(
        "T",
        [AST.interfaceConstraint(AST.simpleTypeExpression("Show"))],
      ),
    ],
  );
}

function buildCallExpression(): AST.Expression {
  return AST.functionCall(
    AST.identifier("use_show"),
    [AST.structLiteral([], false, "Point")],
    [AST.simpleTypeExpression("Point")],
  );
}

function buildModule(includeImpl: boolean): AST.Module {
  const body: AST.Statement[] = [
    buildShowInterface(),
    buildPointStruct(),
    buildUseShowFunction(),
  ];
  if (includeImpl) {
    body.splice(2, 0, buildShowImplementation());
  }
  body.push(buildCallExpression() as unknown as AST.Statement);
  return AST.module(body);
}

function buildResultShowImplementation(): AST.ImplementationDefinition {
  const typeExpr = resultStringType();
  return AST.implementationDefinition(
    AST.identifier("Show"),
    typeExpr,
    [
      AST.functionDefinition(
        "to_string",
        [AST.functionParameter("self", resultStringType())],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("result"))]),
        AST.simpleTypeExpression("string"),
      ),
    ],
  );
}

function buildResultCallExpression(): AST.Expression {
  return AST.functionCall(
    AST.identifier("use_show"),
    [AST.identifier("value")],
    [resultStringType()],
  );
}

function buildResultModule(includeImpl: boolean): AST.Module {
  const body: AST.Statement[] = [
    buildShowInterface(),
    buildUseShowFunction(),
  ];
  if (includeImpl) {
    body.splice(1, 0, buildResultShowImplementation());
  }
  body.push(buildResultCallExpression() as unknown as AST.Statement);
  return AST.module(body);
}

describe("TypeChecker constraint resolution", () => {
  test("honours interface implementations when enforcing generic constraints", () => {
    const moduleAst = buildModule(true);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics).toEqual([]);
  });

  test("reports unsatisfied constraints when no implementation is found", () => {
    const moduleAst = buildModule(false);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.message).toContain("constraint on T is not satisfied");
  });

  test("honours implementations for result types when enforcing generic constraints", () => {
    const moduleAst = buildResultModule(true);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics).toEqual([]);
  });

  test("reports unsatisfied constraints for result types without matching implementation", () => {
    const moduleAst = buildResultModule(false);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    const first = result.diagnostics[0]?.message ?? "";
    expect(first).toContain("constraint on T is not satisfied");
    expect(first).toContain("Result string");
  });
});
