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
      { severity: "error", message: "typechecker: impl Show for Point missing method 'to_String'" },
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
        AST.simpleTypeExpression("String"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toEqual([]);
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
          AST.functionParameter("prefix", AST.simpleTypeExpression("String")),
        ],
        AST.simpleTypeExpression("String"),
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
        AST.simpleTypeExpression("String"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Formatter for Point method 'format' parameter 2 expected String, got i32",
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
        AST.simpleTypeExpression("String"),
      ),
    ]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics).toContainEqual({
      severity: "error",
      message: "typechecker: impl Cloner for Point method 'clone' return type expected Result Point, got String",
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
      [AST.simpleTypeExpression("String")],
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
      message: "typechecker: impl Show for Point defines duplicate method 'to_String'",
    });
  });

  test("reports error when implicit interface implementation targets a type constructor", () => {
    const interfaceDef = AST.interfaceDefinition("Display", []);
    const arrayStruct = AST.structDefinition(
      "Array",
      [],
      "named",
      [AST.genericParameter("T")],
    );
    const impl = AST.implementationDefinition(
      interfaceDef.id,
      AST.simpleTypeExpression("Array"),
      [],
    );
    const moduleAst = AST.module([interfaceDef, arrayStruct, impl]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics.length).toBeGreaterThan(0);
    const typeConstructorDiag = diagnostics.find((diag) => diag.message.includes("type constructor"));
    expect(typeConstructorDiag?.message).toContain("Display");
    expect(typeConstructorDiag?.message).toContain("Array");
  });

  test("accepts bare type constructor when interface declares higher-kinded self pattern", () => {
    const interfaceDef = AST.interfaceDefinition(
      "Mapper",
      [],
      undefined,
      AST.genericTypeExpression(AST.simpleTypeExpression("M"), [AST.wildcardTypeExpression()]),
    );
    const arrayStruct = AST.structDefinition(
      "Array",
      [],
      "named",
      [AST.genericParameter("T")],
    );
    const impl = AST.implementationDefinition(
      interfaceDef.id,
      AST.simpleTypeExpression("Array"),
      [],
    );
    const moduleAst = AST.module([interfaceDef, arrayStruct, impl]);
    const checker = new TypeChecker();
    expect(checker.checkModule(moduleAst).diagnostics).toEqual([]);
  });

  test("accepts partially applied constructor when interface declares higher-kinded self pattern", () => {
    const eachSignature = AST.functionSignature(
      "each",
      [
        AST.functionParameter(
          "self",
          AST.genericTypeExpression(AST.simpleTypeExpression("C"), [AST.simpleTypeExpression("A")]),
        ),
      ],
      AST.simpleTypeExpression("void"),
    );
    const interfaceDef = AST.interfaceDefinition(
      "Enumerable",
      [eachSignature],
      [AST.genericParameter("A")],
      AST.genericTypeExpression(AST.simpleTypeExpression("C"), [AST.wildcardTypeExpression()]),
    );
    const hashMapStruct = AST.structDefinition(
      "HashMap",
      [],
      "named",
      [AST.genericParameter("K"), AST.genericParameter("V")],
    );
    const eachDefinition = AST.functionDefinition(
      "each",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
    );
    const impl = AST.implementationDefinition(
      interfaceDef.id,
      AST.genericTypeExpression(AST.simpleTypeExpression("HashMap"), [AST.simpleTypeExpression("K")]),
      [eachDefinition],
      undefined,
      [AST.genericParameter("K"), AST.genericParameter("V")],
      [AST.simpleTypeExpression("V")],
    );
    const moduleAst = AST.module([interfaceDef, hashMapStruct, impl]);
    const checker = new TypeChecker();
    expect(checker.checkModule(moduleAst).diagnostics).toEqual([]);
  });

  test("accepts bare constructor when higher-kinded interface methods use generics", () => {
    const wrapSignature = AST.functionSignature(
      "wrap",
      [
        AST.functionParameter(
          "self",
          AST.genericTypeExpression(AST.simpleTypeExpression("Self"), [
            AST.simpleTypeExpression("T"),
          ]),
        ),
        AST.functionParameter("value", AST.simpleTypeExpression("T")),
      ],
      AST.genericTypeExpression(AST.simpleTypeExpression("Self"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const wrapperInterface = AST.interfaceDefinition(
      "Wrapper",
      [wrapSignature],
      undefined,
      AST.genericTypeExpression(AST.simpleTypeExpression("F"), [AST.wildcardTypeExpression()]),
    );
    const holderStruct = AST.structDefinition(
      "Holder",
      [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
      "named",
      [AST.genericParameter("T")],
    );
    const wrapDefinition = AST.functionDefinition(
      "wrap",
      [
        AST.functionParameter(
          "self",
          AST.genericTypeExpression(AST.simpleTypeExpression("Holder"), [AST.simpleTypeExpression("T")]),
        ),
        AST.functionParameter("value", AST.simpleTypeExpression("T")),
      ],
      AST.blockExpression([
        AST.returnStatement(
          AST.structLiteral(
            [AST.structFieldInitializer(AST.identifier("value"), "value")],
            false,
            "Holder",
            undefined,
            [AST.simpleTypeExpression("T")],
          ),
        ),
      ]),
      AST.genericTypeExpression(AST.simpleTypeExpression("Holder"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const implementation = AST.implementationDefinition(
      wrapperInterface.id,
      AST.simpleTypeExpression("Holder"),
      [wrapDefinition],
    );
    const moduleAst = AST.module([wrapperInterface, holderStruct, implementation]);
    const checker = new TypeChecker();
    expect(checker.checkModule(moduleAst).diagnostics).toEqual([]);
  });

  test("accepts higher-kinded self placeholders in interface method signatures", () => {
    const wrapSignature = AST.functionSignature(
      "wrap",
      [
        AST.functionParameter(
          "self",
          AST.genericTypeExpression(AST.simpleTypeExpression("M"), [AST.simpleTypeExpression("T")]),
        ),
        AST.functionParameter("value", AST.simpleTypeExpression("T")),
      ],
      AST.genericTypeExpression(AST.simpleTypeExpression("M"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const wrapperInterface = AST.interfaceDefinition(
      "Wrapper",
      [wrapSignature],
      undefined,
      AST.genericTypeExpression(AST.simpleTypeExpression("M"), [AST.wildcardTypeExpression()]),
    );
    const holderStruct = AST.structDefinition(
      "Holder",
      [],
      "named",
      [AST.genericParameter("T")],
    );
    const wrapDefinition = AST.functionDefinition(
      "wrap",
      [
        AST.functionParameter(
          "self",
          AST.genericTypeExpression(AST.simpleTypeExpression("Holder"), [AST.simpleTypeExpression("T")]),
        ),
        AST.functionParameter("value", AST.simpleTypeExpression("T")),
      ],
      AST.blockExpression([AST.returnStatement(AST.identifier("self"))]),
      AST.genericTypeExpression(AST.simpleTypeExpression("Holder"), [AST.simpleTypeExpression("T")]),
      [AST.genericParameter("T")],
    );
    const implementation = AST.implementationDefinition(
      wrapperInterface.id,
      AST.simpleTypeExpression("Holder"),
      [wrapDefinition],
    );
    const moduleAst = AST.module([wrapperInterface, holderStruct, implementation]);
    const checker = new TypeChecker();
    expect(checker.checkModule(moduleAst).diagnostics).toEqual([]);
  });

  test("reports mismatch when implementation target disagrees with explicit self type pattern", () => {
    const interfaceDef = AST.interfaceDefinition(
      "PointOnly",
      [],
      undefined,
      AST.simpleTypeExpression("Point"),
    );
    const pointStruct = buildPointStruct();
    const lineStruct = AST.structDefinition("Line", [], "named");
    const impl = AST.implementationDefinition(
      interfaceDef.id,
      AST.simpleTypeExpression("Line"),
      [],
    );
    const moduleAst = AST.module([interfaceDef, pointStruct, lineStruct, impl]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics.length).toBeGreaterThan(0);
    expect(diagnostics[0]?.message).toContain("self type 'Point'");
  });

  test("reports mismatch when implementation does not satisfy generic self type pattern", () => {
    const interfaceDef = AST.interfaceDefinition(
      "ArrayOnly",
      [],
      undefined,
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("T")]),
    );
    const arrayStruct = AST.structDefinition(
      "Array",
      [],
      "named",
      [AST.genericParameter("T")],
    );
    const resultStruct = AST.structDefinition(
      "Result",
      [],
      "named",
      [AST.genericParameter("T")],
    );
    const impl = AST.implementationDefinition(
      interfaceDef.id,
      AST.genericTypeExpression(AST.simpleTypeExpression("Result"), [AST.simpleTypeExpression("i32")]),
      [],
    );
    const moduleAst = AST.module([interfaceDef, arrayStruct, resultStruct, impl]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics.length).toBeGreaterThan(0);
    expect(diagnostics[0]?.message).toContain("self type 'Array T'");
  });

  test("reports mismatch when higher-kinded interface is implemented for a concrete application", () => {
    const interfaceDef = AST.interfaceDefinition(
      "Applicative",
      [],
      undefined,
      AST.genericTypeExpression(AST.simpleTypeExpression("F"), [AST.wildcardTypeExpression()]),
    );
    const arrayStruct = AST.structDefinition(
      "Array",
      [],
      "named",
      [AST.genericParameter("T")],
    );
    const impl = AST.implementationDefinition(
      interfaceDef.id,
      AST.genericTypeExpression(AST.simpleTypeExpression("Array"), [AST.simpleTypeExpression("i32")]),
      [],
    );
    const moduleAst = AST.module([interfaceDef, arrayStruct, impl]);
    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    expect(diagnostics.length).toBeGreaterThan(0);
    expect(diagnostics[0]?.message).toContain("self type 'F _'");
  });
});

function buildShowInterface(): AST.InterfaceDefinition {
  return AST.interfaceDefinition("Show", [
    AST.functionSignature(
      "to_String",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.simpleTypeExpression("String"),
    ),
  ]);
}

function buildToStringImplementation(): AST.FunctionDefinition {
  return AST.functionDefinition(
    "to_String",
    [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
    AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
    AST.simpleTypeExpression("String"),
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
