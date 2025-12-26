import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("typechecker control flow", () => {
  test("accepts truthy values in if conditions", () => {
    const checker = new TypeChecker();
    const ifExpr = AST.ifExpression(
      AST.integerLiteral(1),
      AST.blockExpression([AST.integerLiteral(0) as unknown as AST.Statement]),
    );
    const module = AST.module([ifExpr as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("accepts truthy values in while conditions", () => {
    const checker = new TypeChecker();
    const loop = AST.whileLoop(
      AST.integerLiteral(1),
      AST.blockExpression([]),
    );
    const module = AST.module([loop]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
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

  test("reports diagnostic when break appears outside loop", () => {
    const checker = new TypeChecker();
    const module = AST.module([AST.breakStatement()]);

    const result = checker.checkModule(module);
    const hasDiag = result.diagnostics.some((diag) =>
      diag.message.includes("break statement must appear inside a loop"),
    );
    expect(hasDiag).toBe(true);
  });

  test("reports diagnostic when break label is unknown", () => {
    const checker = new TypeChecker();
    const loop = AST.loopExpression(
      AST.blockExpression([
        AST.breakStatement("missing") as unknown as AST.Statement,
      ]),
    );
    const module = AST.module([loop as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    const hasDiag = result.diagnostics.some((diag) =>
      diag.message.includes("unknown break label 'missing'"),
    );
    expect(hasDiag).toBe(true);
  });

  test("allows break with matching breakpoint label", () => {
    const checker = new TypeChecker();
    const bp = AST.breakpointExpression(
      "exit",
      AST.blockExpression([
        AST.breakStatement("exit", AST.integerLiteral(1)),
      ]),
    );
    const module = AST.module([bp as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });

  test("reports diagnostic when continue appears outside loop", () => {
    const checker = new TypeChecker();
    const module = AST.module([AST.continueStatement()]);

    const result = checker.checkModule(module);
    const hasDiag = result.diagnostics.some((diag) =>
      diag.message.includes("continue statement must appear inside a loop"),
    );
    expect(hasDiag).toBe(true);
  });

  test("reports diagnostic when continue statement is labeled", () => {
    const checker = new TypeChecker();
    const loop = AST.loopExpression(
      AST.blockExpression([
        AST.continueStatement("next") as unknown as AST.Statement,
      ]),
    );
    const module = AST.module([loop as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    const hasDiag = result.diagnostics.some((diag) =>
      diag.message.includes("labeled continue is not supported"),
    );
    expect(hasDiag).toBe(true);
  });

  test("loop expression type matches break payload", () => {
    const checker = new TypeChecker();
    const loopExpr = AST.loopExpression(
      AST.blockExpression([
        AST.breakStatement(undefined, AST.integerLiteral(5)),
      ]),
    );
    const assignment = AST.assignmentExpression(
      ":=",
      AST.typedPattern(AST.identifier("value"), AST.simpleTypeExpression("i32")),
      loopExpr,
    );
    const module = AST.module([assignment as unknown as AST.Statement]);

    const result = checker.checkModule(module);
    expect(result.diagnostics).toHaveLength(0);
  });
});
