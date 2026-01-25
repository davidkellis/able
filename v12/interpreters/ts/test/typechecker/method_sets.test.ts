import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

function buildShowcaseInterface(): AST.InterfaceDefinition {
  return AST.interfaceDefinition("Showcase", [
    AST.functionSignature(
      "show",
      [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
      AST.simpleTypeExpression("String"),
    ),
  ]);
}

function buildFormatterMethodSet(): AST.MethodsDefinition {
  const selfType = AST.genericTypeExpression(AST.simpleTypeExpression("Wrapper"), [AST.simpleTypeExpression("T")]);
  const describeMethod = AST.functionDefinition(
    "describe",
    [AST.functionParameter("self", selfType)],
    AST.blockExpression([AST.returnStatement(AST.stringLiteral("ok"))]),
    AST.simpleTypeExpression("String"),
  );
  return AST.methodsDefinition(
    selfType,
    [describeMethod],
    [AST.genericParameter("T")],
    [AST.whereClauseConstraint("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Showcase"))])],
  );
}

function buildWrapperStruct(): AST.StructDefinition {
  return AST.structDefinition(
    "Wrapper",
    [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
    "named",
    [AST.genericParameter("T")],
  );
}

function buildWrapperLiteral(typeArg: string, valueExpression: AST.Expression): AST.StructLiteral {
  return AST.structLiteral(
    [AST.structFieldInitializer(valueExpression, "value")],
    false,
    "Wrapper",
    undefined,
    [AST.simpleTypeExpression(typeArg)],
  );
}

describe("typechecker method-set obligations", () => {
  test("reports missing method-set constraint for inferred subject", () => {
    const checker = new TypeChecker();
    const wrapperValue = buildWrapperLiteral("i32", AST.integerLiteral(1));
    const call = AST.functionCall(AST.memberAccessExpression(wrapperValue, AST.identifier("describe")), []);
    const module = AST.module([
      buildShowcaseInterface(),
      buildWrapperStruct(),
      buildFormatterMethodSet(),
      call as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    const messages = result.diagnostics.map((diag) => diag.message);
    expect(messages).toContainEqual(expect.stringMatching(/constraint on T/));
    expect(messages).toContainEqual(expect.stringMatching(/does not implement Showcase/));
  });

  test("accepts satisfied method-set constraint", () => {
    const checker = new TypeChecker();
    const pointStruct = AST.structDefinition(
      "Point",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
      ],
      "named",
    );
    const displayImpl = AST.implementationDefinition(
      AST.identifier("Showcase"),
      AST.simpleTypeExpression("Point"),
      [
        AST.functionDefinition(
          "show",
          [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
          AST.blockExpression([AST.returnStatement(AST.stringLiteral("<point>"))]),
          AST.simpleTypeExpression("String"),
        ),
      ],
    );
    const wrapperValue = buildWrapperLiteral(
      "Point",
      AST.structLiteral(
        [
          AST.structFieldInitializer(AST.integerLiteral(1), "x"),
          AST.structFieldInitializer(AST.integerLiteral(2), "y"),
        ],
        false,
        "Point",
      ),
    );
    const call = AST.functionCall(AST.memberAccessExpression(wrapperValue, AST.identifier("describe")), []);
    const module = AST.module([
      buildShowcaseInterface(),
      pointStruct,
      displayImpl,
      buildWrapperStruct(),
      buildFormatterMethodSet(),
      call as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("reports missing method-set constraint for Self requirement", () => {
    const checker = new TypeChecker();
    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Wrapper"),
      [
        AST.functionDefinition(
          "describe",
          [AST.functionParameter("self", AST.simpleTypeExpression("Wrapper"))],
          AST.blockExpression([
            AST.returnStatement(
              AST.functionCall(
                AST.memberAccessExpression(AST.identifier("self"), AST.identifier("show")),
                [],
              ),
            ),
          ]),
          AST.simpleTypeExpression("String"),
        ),
      ],
      [],
      [
        AST.whereClauseConstraint("Self", [
          AST.interfaceConstraint(AST.simpleTypeExpression("Showcase")),
        ]),
      ],
    );
    const module = AST.module([
      buildShowcaseInterface(),
      AST.structDefinition(
        "Wrapper",
        [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
        "named",
      ),
      methods,
      AST.assignmentExpression(
        ":=",
        AST.identifier("wrapper"),
        AST.structLiteral(
          [AST.structFieldInitializer(AST.integerLiteral(1), "value")],
          false,
          "Wrapper",
        ),
      ) as unknown as AST.Statement,
      AST.functionCall(
        AST.memberAccessExpression(AST.identifier("wrapper"), AST.identifier("describe")),
        [],
      ) as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    const messages = result.diagnostics.map((diag) => diag.message);
    expect(messages).toContainEqual(expect.stringMatching(/constraint on Self/));
    expect(messages).toContainEqual(expect.stringMatching(/Wrapper does not implement Showcase/));
  });

  test("methods allow implicit self parameter annotations", () => {
    const checker = new TypeChecker();
    const channelStruct = AST.structDefinition(
      "Channel",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle")],
      "named",
    );
    const sendMethod = AST.functionDefinition(
      "send",
      [
        AST.functionParameter("self"),
        AST.functionParameter("value", AST.simpleTypeExpression("i32")),
      ],
      AST.blockExpression([
        AST.returnStatement(
          AST.memberAccessExpression(AST.identifier("self"), AST.identifier("handle")),
        ),
      ]),
      AST.simpleTypeExpression("i64"),
    );
    const methods = AST.methodsDefinition(AST.simpleTypeExpression("Channel"), [sendMethod]);
    const module = AST.module([channelStruct, methods]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });

  test("implementations allow implicit self parameter annotations", () => {
    const checker = new TypeChecker();
    const showInterface = AST.interfaceDefinition("Show", [
      AST.functionSignature(
        "value",
        [AST.functionParameter("self", AST.simpleTypeExpression("Self"))],
        AST.simpleTypeExpression("i32"),
      ),
    ]);
    const meterStruct = AST.structDefinition(
      "Meter",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "reading")],
      "named",
    );
    const valueMethod = AST.functionDefinition(
      "value",
      [AST.functionParameter("self")],
      AST.blockExpression([
        AST.returnStatement(
          AST.memberAccessExpression(AST.identifier("self"), AST.identifier("reading")),
        ),
      ]),
      AST.simpleTypeExpression("i32"),
    );
    const impl = AST.implementationDefinition(
      AST.identifier("Show"),
      AST.simpleTypeExpression("Meter"),
      [valueMethod],
    );
    const module = AST.module([showInterface, meterStruct, impl]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toEqual([]);
  });
});
