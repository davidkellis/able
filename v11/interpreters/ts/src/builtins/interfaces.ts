import * as AST from "../ast";

export type BuiltinInterfaceBundle = {
  interfaces: AST.InterfaceDefinition[];
  implementations: AST.ImplementationDefinition[];
};

function displaySignature(): AST.FunctionSignature {
  return AST.functionSignature(
    "to_string",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.simpleTypeExpression("String"),
  );
}

function cloneSignature(): AST.FunctionSignature {
  return AST.functionSignature(
    "clone",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.simpleTypeExpression("Self"),
  );
}

function errorMessageSignature(): AST.FunctionSignature {
  return AST.functionSignature(
    "message",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.simpleTypeExpression("String"),
  );
}

function errorCauseSignature(): AST.FunctionSignature {
  return AST.functionSignature(
    "cause",
    [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
    AST.nullableTypeExpression(AST.simpleTypeExpression("Error")),
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
        AST.simpleTypeExpression("String"),
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

function errorImplForErrorValue(): AST.ImplementationDefinition {
  const selfIdent = AST.identifier("self");
  const messageFn = AST.functionDefinition(
    "message",
    [AST.functionParameter("self", AST.simpleTypeExpression("Error"))],
    AST.blockExpression([
      AST.returnStatement(AST.functionCall(AST.memberAccessExpression(selfIdent, AST.identifier("message")), [])),
    ]),
    AST.simpleTypeExpression("String"),
  );
  const causeFn = AST.functionDefinition(
    "cause",
    [AST.functionParameter("self", AST.simpleTypeExpression("Error"))],
    AST.blockExpression([
      AST.returnStatement(AST.functionCall(AST.memberAccessExpression(selfIdent, AST.identifier("cause")), [])),
    ]),
    AST.nullableTypeExpression(AST.simpleTypeExpression("Error")),
  );
  return AST.implementationDefinition(
    "Error",
    AST.simpleTypeExpression("Error"),
    [messageFn, causeFn],
  );
}

function errorImplForProcError(): AST.ImplementationDefinition {
  const selfIdent = AST.identifier("self");
  const messageFn = AST.functionDefinition(
    "message",
    [AST.functionParameter("self", AST.simpleTypeExpression("ProcError"))],
    AST.blockExpression([
      AST.returnStatement(AST.memberAccessExpression(selfIdent, AST.identifier("details"))),
    ]),
    AST.simpleTypeExpression("String"),
  );
  const causeFn = AST.functionDefinition(
    "cause",
    [AST.functionParameter("self", AST.simpleTypeExpression("ProcError"))],
    AST.blockExpression([AST.returnStatement(AST.nilLiteral())]),
    AST.nullableTypeExpression(AST.simpleTypeExpression("Error")),
  );
  return AST.implementationDefinition(
    "Error",
    AST.simpleTypeExpression("ProcError"),
    [messageFn, causeFn],
  );
}

function interpolationOfSelf(): AST.StringInterpolation {
  return AST.stringInterpolation([AST.identifier("self")]);
}

export function buildStandardInterfaceBuiltins(): BuiltinInterfaceBundle {
  const interfaces = [
    AST.interfaceDefinition("Display", [displaySignature()]),
    AST.interfaceDefinition("Clone", [cloneSignature()]),
    AST.interfaceDefinition("Error", [errorMessageSignature(), errorCauseSignature()]),
  ];

  const implementations: AST.ImplementationDefinition[] = [
    displayImpl("String", AST.identifier("self")),
    displayImpl("i32", interpolationOfSelf()),
    displayImpl("bool", interpolationOfSelf()),
    displayImpl("char", interpolationOfSelf()),
    displayImpl("f64", interpolationOfSelf()),
    cloneImpl("String"),
    cloneImpl("i32"),
    cloneImpl("bool"),
    cloneImpl("char"),
    cloneImpl("f64"),
    errorImplForProcError(),
  ];

  return { interfaces, implementations };
}
