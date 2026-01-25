import { describe, expect, test } from "bun:test";
import * as AST from "../../src/ast";
import { TypeChecker } from "../../src/typechecker";

describe("constraint arity", () => {
  test("reports missing and mismatched interface type arguments", () => {
    const pairInterface = AST.interfaceDefinition(
      "Pair",
      [
        AST.functionSignature(
          "pair",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Self")),
            AST.functionParameter("a", AST.simpleTypeExpression("A")),
            AST.functionParameter("b", AST.simpleTypeExpression("B")),
          ],
          AST.simpleTypeExpression("String"),
        ),
      ],
      [AST.genericParameter("A"), AST.genericParameter("B")],
      AST.simpleTypeExpression("T"),
    );

    const boxStruct = AST.structDefinition(
      "Box",
      [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
      "named",
    );

    const pairImpl = AST.implementationDefinition(
      AST.identifier("Pair"),
      AST.simpleTypeExpression("Box"),
      [
        AST.functionDefinition(
          "pair",
          [
            AST.functionParameter("self", AST.simpleTypeExpression("Box")),
            AST.functionParameter("a", AST.simpleTypeExpression("i32")),
            AST.functionParameter("b", AST.simpleTypeExpression("i32")),
          ],
          AST.blockExpression([AST.returnStatement(AST.stringLiteral("ok"))]),
          AST.simpleTypeExpression("String"),
        ),
      ],
      undefined,
      [],
      [AST.simpleTypeExpression("i32"), AST.simpleTypeExpression("i32")],
    );

    const useMissing = AST.functionDefinition(
      "use_missing",
      [AST.functionParameter("value", AST.simpleTypeExpression("T"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
      [AST.genericParameter("T", [AST.interfaceConstraint(AST.simpleTypeExpression("Pair"))])],
    );

    const useMismatch = AST.functionDefinition(
      "use_mismatch",
      [AST.functionParameter("value", AST.simpleTypeExpression("T"))],
      AST.blockExpression([]),
      AST.simpleTypeExpression("void"),
      [
        AST.genericParameter(
          "T",
          [
            AST.interfaceConstraint(
              AST.genericTypeExpression(AST.simpleTypeExpression("Pair"), [AST.simpleTypeExpression("i32")]),
            ),
          ],
        ),
      ],
    );

    const boxLiteral = AST.structLiteral(
      [AST.structFieldInitializer(AST.integerLiteral(1), "value")],
      false,
      "Box",
    );
    const callMissing = AST.functionCall(
      AST.identifier("use_missing"),
      [boxLiteral],
      [AST.simpleTypeExpression("Box")],
    );
    const callMismatch = AST.functionCall(
      AST.identifier("use_mismatch"),
      [boxLiteral],
      [AST.simpleTypeExpression("Box")],
    );

    const moduleAst = AST.module([
      pairInterface,
      boxStruct,
      pairImpl,
      useMissing,
      useMismatch,
      callMissing as unknown as AST.Statement,
      callMismatch as unknown as AST.Statement,
    ]);

    const checker = new TypeChecker();
    const { diagnostics } = checker.checkModule(moduleAst);
    const messages = diagnostics.map((diag) => diag.message);

    expect(messages.some((message) => message.includes("requires 2 type argument(s) for interface 'Pair'"))).toBeTrue();
    expect(messages.some((message) => message.includes("expected 2 type argument(s) for interface 'Pair', got 1"))).toBeTrue();
  });
});
