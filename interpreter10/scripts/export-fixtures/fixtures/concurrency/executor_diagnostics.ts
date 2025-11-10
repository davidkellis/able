import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const executorDiagnosticsFixtures: Fixture[] = [
  {
    name: "concurrency/proc_executor_diagnostics",
    module: AST.module([
      AST.assign("initial_pending", AST.functionCall(AST.identifier("proc_pending_tasks"), [])),
      AST.iff(
        AST.binaryExpression("!=", AST.identifier("initial_pending"), AST.integerLiteral(0)),
        AST.raiseStatement(AST.stringLiteral("executor should start empty")),
      ),
      AST.assign(
        "_first",
        AST.spawnExpression(
          AST.blockExpression([
            AST.integerLiteral(1),
          ]),
        ),
      ),
      AST.assign(
        "_second",
        AST.spawnExpression(
          AST.blockExpression([
            AST.integerLiteral(2),
          ]),
        ),
      ),
      AST.assign("pending_mid", AST.functionCall(AST.identifier("proc_pending_tasks"), [])),
      AST.iff(
        AST.binaryExpression("<=", AST.identifier("pending_mid"), AST.integerLiteral(0)),
        AST.raiseStatement(AST.stringLiteral("expected pending tasks after spawn")),
      ),
      AST.functionCall(AST.identifier("proc_flush"), []),
      AST.assign("pending_end", AST.functionCall(AST.identifier("proc_pending_tasks"), [])),
      AST.assign("attempts", AST.integerLiteral(0)),
      AST.assign("max_attempts", AST.integerLiteral(8)),
      AST.whileLoop(
        AST.binaryExpression(
          "&&",
          AST.binaryExpression(">", AST.identifier("pending_end"), AST.integerLiteral(0)),
          AST.binaryExpression("<", AST.identifier("attempts"), AST.identifier("max_attempts")),
        ),
        AST.blockExpression([
          AST.functionCall(AST.identifier("proc_flush"), []),
          AST.assignmentExpression(
            "=",
            AST.identifier("pending_end"),
            AST.functionCall(AST.identifier("proc_pending_tasks"), []),
          ),
          AST.assignmentExpression("+=", AST.identifier("attempts"), AST.integerLiteral(1)),
        ]),
      ),
      AST.iff(
        AST.binaryExpression("!=", AST.identifier("pending_end"), AST.integerLiteral(0)),
        AST.raiseStatement(AST.stringLiteral("executor should be drained after proc_flush")),
      ),
      AST.bool(true),
    ]),
    manifest: {
      description: "proc_pending_tasks exposes cooperative executor queue state",
      expect: {
        result: { kind: "bool", value: true },
      },
    },
  },
];

export default executorDiagnosticsFixtures;
