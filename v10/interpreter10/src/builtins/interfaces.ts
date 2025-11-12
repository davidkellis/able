import * as AST from "../ast";

export type BuiltinInterfaceBundle = {
  interfaces: AST.InterfaceDefinition[];
  implementations: AST.ImplementationDefinition[];
};

function displaySignature(): AST.FunctionSignature {
  return AST.functionSignature(
    "to_string",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.simpleTypeExpression("string"),
  );
}

function cloneSignature(): AST.FunctionSignature {
  return AST.functionSignature(
    "clone",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.simpleTypeExpression("Self"),
  );
}

function displayImpl(
  typeName: string,
  bodyExpression: AST.Expression,
): AST.ImplementationDefinition {
  return AST.implementationDefinition(
    "Display",
    AST.simpleTypeExpression(typeName),
    [
      AST.functionDefinition(
        "to_string",
        [AST.functionParameter("self", AST.simpleTypeExpression(typeName))],
        AST.blockExpression([AST.returnStatement(bodyExpression)]),
        AST.simpleTypeExpression("string"),
      ),
    ],
  );
}

function cloneImpl(typeName: string): AST.ImplementationDefinition {
  return AST.implementationDefinition(
    "Clone",
    AST.simpleTypeExpression(typeName),
    [
      AST.functionDefinition(
        "clone",
        [AST.functionParameter("self", AST.simpleTypeExpression(typeName))],
        AST.blockExpression([AST.returnStatement(AST.identifier("self"))]),
        AST.simpleTypeExpression(typeName),
      ),
    ],
  );
}

function interpolationOfSelf(): AST.StringInterpolation {
  return AST.stringInterpolation([AST.identifier("self")]);
}

export function buildStandardInterfaceBuiltins(): BuiltinInterfaceBundle {
  const interfaces = [
    AST.interfaceDefinition("Display", [displaySignature()]),
    AST.interfaceDefinition("Clone", [cloneSignature()]),
  ];

  const implementations: AST.ImplementationDefinition[] = [
    displayImpl("string", AST.identifier("self")),
    displayImpl("i32", interpolationOfSelf()),
    displayImpl("bool", interpolationOfSelf()),
    displayImpl("char", interpolationOfSelf()),
    displayImpl("f64", interpolationOfSelf()),
    cloneImpl("string"),
    cloneImpl("i32"),
    cloneImpl("bool"),
    cloneImpl("char"),
    cloneImpl("f64"),
  ];

  return { interfaces, implementations };
}
