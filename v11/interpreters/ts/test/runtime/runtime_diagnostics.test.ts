import * as AST from "../../src/ast";
import { attachRuntimeDiagnosticContext } from "../../src/interpreter/runtime_diagnostics";
import { buildRuntimeDiagnostic, formatRuntimeDiagnostic } from "../../scripts/typecheck-utils";

test("formats runtime diagnostics with call-site notes", () => {
  const err = new Error("division by zero");
  const expr = AST.binaryExpression("/", AST.integerLiteral(1), AST.integerLiteral(0));
  expr.span = { start: { line: 3, column: 5 }, end: { line: 3, column: 9 } };
  expr.origin = "v11/interpreters/ts/src/main.able";

  const callNode = AST.functionCall(AST.identifier("foo"), []);
  callNode.span = { start: { line: 8, column: 2 }, end: { line: 8, column: 6 } };
  callNode.origin = "v11/interpreters/ts/src/main.able";

  attachRuntimeDiagnosticContext(err, { node: expr, callStack: [{ node: callNode }] });

  const formatted = formatRuntimeDiagnostic(buildRuntimeDiagnostic(err));
  expect(formatted).toBe(
    "runtime: v11/interpreters/ts/src/main.able:3:5 division by zero\nnote: v11/interpreters/ts/src/main.able:8:2 called from here",
  );
});
