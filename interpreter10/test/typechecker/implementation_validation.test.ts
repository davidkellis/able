import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("TypeChecker implementation validation", () => {
  test("reports missing interface methods in implementation", () => {
    const interfaceDef = buildShowInterface();
    const moduleAst = buildModule(interfaceDef, []);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([
      { severity: "error", message: "typechecker: impl Show for Point missing method 'to_string'" },
    ]);
  });

  test("reports extraneous methods not declared on the interface", () => {
    const interfaceDef = buildShowInterface();
    const moduleAst = buildModule(interfaceDef, [
      buildToStringImplementation(),
      AST.functionDefinition(
        "extra",
        [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("extra"))]),
        AST.simpleTypeExpression("string"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([
      {
        severity: "error",
        message: "typechecker: impl Show for Point defines method 'extra' not declared in interface Show",
      },
    ]);
  });

  test("accepts implementations that cover the entire interface surface", () => {
    const interfaceDef = buildShowInterface();
    const moduleAst = buildModule(interfaceDef, [buildToStringImplementation()]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([]);
  });

  test("reports parameter count mismatches", () => {
    const interfaceDef = AST.interfaceDefinition("Comparable", [
      AST.functionSignature(
        "compare",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("other", AST.simpleTypeExpression("Self")),
        ],
        AST.simpleTypeExpression("bool"),
      ),
    ]);
    const moduleAst = buildModule(interfaceDef, [
      AST.functionDefinition(
        "compare",
        [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
        AST.blockExpression([AST.returnStatement(AST.booleanLiteral(true))]),
        AST.simpleTypeExpression("bool"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Comparable for Point method 'compare' expects 2 parameter(s), got 1",
    });
  });

  test("reports parameter type mismatches", () => {
    const interfaceDef = AST.interfaceDefinition("Formatter", [
      AST.functionSignature(
        "format",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("prefix", AST.simpleTypeExpression("string")),
        ],
        AST.simpleTypeExpression("string"),
      ),
    ]);
    const moduleAst = buildModule(interfaceDef, [
      AST.functionDefinition(
        "format",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Point")),
          AST.functionParameter("prefix", AST.simpleTypeExpression("i32")),
        ],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
        AST.simpleTypeExpression("string"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Formatter for Point method 'format' parameter 2 expected string, got i32",
    });
  });

  test("reports return type mismatches", () => {
    const interfaceDef = AST.interfaceDefinition("Cloner", [
      AST.functionSignature(
        "clone",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.resultTypeExpression(AST.simpleTypeExpression("Self")),
      ),
    ]);
    const moduleAst = buildModule(interfaceDef, [
      AST.functionDefinition(
        "clone",
        [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
        AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
        AST.simpleTypeExpression("string"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Cloner for Point method 'clone' return type expected Result Point, got string",
    });
  });

  test("reports generic parameter mismatches", () => {
    const interfaceDef = AST.interfaceDefinition("Wrapper", [
      AST.functionSignature(
        "wrap",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Self")),
          AST.functionParameter("value", AST.simpleTypeExpression("T")),
        ],
        AST.simpleTypeExpression("Self"),
        [AST.genericParameter("T")],
      ),
    ]);
    const moduleAst = buildModule(interfaceDef, [
      AST.functionDefinition(
        "wrap",
        [
          AST.functionParameter("self", AST.simpleTypeExpression("Point")),
          AST.functionParameter("value", AST.simpleTypeExpression("T")),
        ],
        AST.blockExpression([AST.returnStatement(AST.identifier("self"))]),
        AST.simpleTypeExpression("Point"),
        [AST.genericParameter("T"), AST.genericParameter("U")],
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Wrapper for Point method 'wrap' expects 1 generic parameter(s), got 2",
    });
  });

  test("reports missing interface type arguments", () => {
    const interfaceDef = AST.interfaceDefinition(
      "Container",
      [
        AST.functionSignature(
          "size",
          [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
          AST.simpleTypeExpression("i32"),
        ),
      ],
      [AST.genericParameter("T")],
    );
    const moduleAst = buildModule(interfaceDef, [
      AST.functionDefinition(
        "size",
        [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
        AST.blockExpression([AST.returnStatement(AST.integerLiteral(0))]),
        AST.simpleTypeExpression("i32"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Container for Point requires 1 interface type argument(s)",
    });
  });

  test("reports extra interface type arguments", () => {
    const interfaceDef = buildShowInterface();
    const moduleAst = buildModule(
      interfaceDef,
      [buildToStringImplementation()],
      [AST.simpleTypeExpression("string")],
    );
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Show does not accept type arguments",
    });
  });

  test("reports duplicate methods in implementation", () => {
    const interfaceDef = buildShowInterface();
    const moduleAst = buildModule(interfaceDef, [
      buildToStringImplementation(),
      buildToStringImplementation(),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Show for Point defines duplicate method 'to_string'",
    });
  });
});

function buildShowInterface(): AST.InterfaceDefinition {
  return AST.interfaceDefinition("Show", [
    AST.functionSignature(
      "to_string",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.simpleTypeExpression("string"),
    ),
  ]);
}

function buildToStringImplementation(): AST.FunctionDefinition {
  return AST.functionDefinition(
    "to_string",
    [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
    AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
    AST.simpleTypeExpression("string"),
  );
}

function buildPointStruct(): AST.StructDefinition {
  return AST.structDefinition(
    "Point",
    [
      AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
      AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
    ],
    "named",
  );
}

function buildImplementation(
  interfaceDef: AST.InterfaceDefinition,
  methods: AST.FunctionDefinition[],
  interfaceArgs: AST.TypeExpression[] = [],
  targetType: AST.TypeExpression = AST.simpleTypeExpression("Point"),
): AST.ImplementationDefinition {
  return AST.implementationDefinition(
    interfaceDef.id,
    targetType,
    methods,
    undefined,
    undefined,
    interfaceArgs,
  );
}

function buildModule(
  interfaceDef: AST.InterfaceDefinition,
  methods: AST.FunctionDefinition[],
  interfaceArgs: AST.TypeExpression[] = [],
  targetType: AST.TypeExpression = AST.simpleTypeExpression("Point"),
): AST.Module {
  return AST.module([
    interfaceDef,
    buildPointStruct(),
    buildImplementation(interfaceDef, methods, interfaceArgs, targetType),
  ]);
}
