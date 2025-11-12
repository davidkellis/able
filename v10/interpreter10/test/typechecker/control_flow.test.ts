import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker control flow", () => {
  test("reports diagnostic when if condition is not bool", () => {
    const checker = new TypeChecker();
    const ifExpr = AST.ifExpression(
      AST.integerLiteral(1),
      AST.blockExpression([AST.integerLiteral(0) as unknown as AST.Statement]),
    );
    const module = AST.module([ifExpr as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("if condition must be bool");
  });

  test("reports diagnostic when while condition is not bool", () => {
    const checker = new TypeChecker();
    const loop = AST.whileLoop(
      AST.integerLiteral(1),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("while condition must be bool");
  });

  test("accepts equality expressions in if conditions", () => {
    const checker = new TypeChecker();
    const assign = AST.assignmentExpression(
      ":=",
      AST.identifier("value"),
      AST.integerLiteral(1),
    );
    const comparison = AST.binaryExpression("==", AST.identifier("value"), AST.integerLiteral(1));
    const ifExpr = AST.ifExpression(
      comparison,
      AST.blockExpression([]),
    );
    const module = AST.module([
      assign as unknown as AST.Statement,
      ifExpr as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("uses function return types in conditions", () => {
    const checker = new TypeChecker();
    const boolType = AST.simpleTypeExpression("bool");
    const fnBody = AST.blockExpression([AST.returnStatement(AST.booleanLiteral(true))]);
    const fnDef = AST.functionDefinition("is_ready", [], fnBody, boolType);
    const ifExpr = AST.ifExpression(
      AST.functionCall(AST.identifier("is_ready"), []),
      AST.blockExpression([]),
    );
    const module = AST.module([
      fnDef,
      ifExpr as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("respects method return types in conditions", () => {
    const checker = new TypeChecker();
    const channelStruct = AST.structDefinition(
      "Channel",
      [AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "flag")],
      "named",
    );
    const methods = AST.methodsDefinition(
      AST.simpleTypeExpression("Channel"),
      [
        AST.functionDefinition(
          "is_ready",
          [AST.functionParameter("self", AST.simpleTypeExpression("Channel"))],
          AST.blockExpression([
            AST.returnStatement(AST.memberAccessExpression(AST.identifier("self"), "flag")),
          ]),
          AST.simpleTypeExpression("bool"),
        ),
      ],
    );
    const assignChannel = AST.assign(
      "channel",
      AST.structLiteral(
        [AST.structFieldInitializer(AST.booleanLiteral(true), "flag")],
        false,
        "Channel",
      ),
    );
    const ifExpr = AST.ifExpression(
      AST.functionCall(AST.memberAccessExpression(AST.identifier("channel"), "is_ready"), []),
      AST.blockExpression([]),
    );
    const module = AST.module([
      channelStruct,
      methods,
      assignChannel as unknown as AST.Statement,
      ifExpr as unknown as AST.Statement,
    ]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });
});
