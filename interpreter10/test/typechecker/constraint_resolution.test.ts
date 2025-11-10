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
function buildDisplayInterface(): AST.InterfaceDefinition {
  return AST.interfaceDefinition("Display", [
    AST.functionSignature(
      "to_string",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.simpleTypeExpression("string"),
    ),
  ]);
}
function buildNullableDisplayImplementation(): AST.ImplementationDefinition {
  const genericParam = AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Display"))]);
  const nullableType = AST.nullableTypeExpression(AST.simpleTypeExpression("T"));
  const method = AST.functionDefinition(
    "to_string",
    [AST.functionParameter("self", nullableType)],
    AST.blockExpression([AST.returnStatement(AST.stringLiteral("nullable"))]),
    AST.simpleTypeExpression("string"),
    [],
    [],
  );
  return AST.implementationDefinition(
    AST.identifier("Display"),
    nullableType,
    [method],
    undefined,
    [genericParam],
    [],
    [],
  );
}
function buildPointDisplayImplementation(): AST.ImplementationDefinition {
  const method = AST.functionDefinition(
    "to_string",
    [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
    AST.blockExpression([AST.returnStatement(AST.stringLiteral("point"))]),
    AST.simpleTypeExpression("string"),
  );
  return AST.implementationDefinition(
    AST.identifier("Display"),
    AST.simpleTypeExpression("Point"),
    [method],
  );
}
function buildUseDisplayNullableFunction(): AST.FunctionDefinition {
  return AST.functionDefinition(
    "use_display",
    [
      AST.functionParameter(
        "value",
        AST.nullableTypeExpression(AST.simpleTypeExpression("T")),
      ),
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
    [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Display"))])],
  );
}
function buildNullableCallExpression(): AST.Expression {
  return AST.functionCall(
    AST.identifier("use_display"),
    [AST.nilLiteral()],
    [AST.nullableTypeExpression(AST.simpleTypeExpression("Point"))],
  );
}
function buildNullableModule(includePointImpl: boolean): AST.Module {
  const body: AST.Statement[] = [
    buildDisplayInterface(),
    buildPointStruct(),
    buildNullableDisplayImplementation(),
    buildUseDisplayNullableFunction(),
  ];
  if (includePointImpl) {
    body.splice(2, 0, buildPointDisplayImplementation());
  }
  body.push(buildNullableCallExpression() as unknown as AST.Statement);
  return AST.module(body);
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
  test("reports unsatisfied constraints for nullable implementations without inner implementation", () => {
    const moduleAst = buildNullableModule(false);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    const message = result.diagnostics[0]?.message ?? "";
    expect(message).toContain("constraint on T");
    expect(message).toContain("impl Display");
  });
  test("honours nullable implementations when obligations are satisfied", () => {
    const moduleAst = buildNullableModule(true);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics).toEqual([]);
  });
  test("built-in Display/Clone interfaces satisfy constraints without explicit declarations", () => {
    const chooseFirst = AST.functionDefinition(
      "choose_first",
      [
        AST.functionParameter("first", AST.simpleTypeExpression("T")),
        AST.functionParameter("second", AST.simpleTypeExpression("U")),
      ],
      AST.blockExpression([AST.returnStatement(AST.identifier("first"))]),
      AST.simpleTypeExpression("T"),
      [AST.genericParameter("T"), AST.genericParameter("U")],
      [
        AST.whereClauseConstraint("T", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
          AST.interfaceConstraint(AST.simpleTypeExpression("Clone")),
        ]),
        AST.whereClauseConstraint("U", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Display")),
        ]),
      ],
    );
    const invocation = AST.functionCall(
      AST.identifier("choose_first"),
      [AST.stringLiteral("winner"), AST.integerLiteral(1)],
      [AST.simpleTypeExpression("string"), AST.simpleTypeExpression("i32")],
    ) as unknown as AST.Statement;
    const moduleAst = AST.module([chooseFirst, invocation]);
    const checker = new TypeChecker();
    const result = checker.checkModule(moduleAst);
    expect(result.diagnostics).toEqual([]);
  });
});
