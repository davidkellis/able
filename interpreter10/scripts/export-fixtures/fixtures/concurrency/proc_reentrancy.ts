import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const procReentrancyFixtures: Fixture[] = [
  {
        name: "concurrency/future_value_reentrancy",
        module: AST.module([
          AST.assign("trace", AST.stringLiteral("")),
          AST.assign(
            "inner",
            AST.spawnExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("I"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("J"),
                  ),
                ),
                AST.stringLiteral("X"),
              ]),
            ),
          ),
          AST.assign(
            "outer",
            AST.spawnExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("O"),
                  ),
                ),
                AST.assign(
                  "result",
                  AST.functionCall(
                    AST.memberAccessExpression(AST.identifier("inner"), "value"),
                    [],
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.stringInterpolation([
                    AST.identifier("trace"),
                    AST.identifier("result"),
                  ]),
                ),
                AST.stringLiteral("done"),
              ]),
            ),
          ),
          AST.assign(
            "final",
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("outer"), "value"),
              [],
            ),
          ),
          AST.stringInterpolation([
            AST.identifier("trace"),
            AST.identifier("final"),
          ]),
        ]),
        manifest: {
          description: "Nested future.value() calls (re-entrancy) resolve without deadlock",
          expect: {
            result: { kind: "string", value: "OIJXdone" },
          },
        },
      },

  {
        name: "concurrency/proc_value_reentrancy",
        module: AST.module([
          AST.assign("trace", AST.stringLiteral("")),
          AST.assign(
            "inner",
            AST.procExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("I"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("J"),
                  ),
                ),
                AST.stringLiteral("X"),
              ]),
            ),
          ),
          AST.assign(
            "outer",
            AST.procExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.binaryExpression(
                    "+",
                    AST.identifier("trace"),
                    AST.stringLiteral("O"),
                  ),
                ),
                AST.assign(
                  "result",
                  AST.functionCall(
                    AST.memberAccessExpression(AST.identifier("inner"), "value"),
                    [],
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("trace"),
                  AST.stringInterpolation([
                    AST.identifier("trace"),
                    AST.identifier("result"),
                  ]),
                ),
                AST.stringLiteral("done"),
              ]),
            ),
          ),
          AST.assign(
            "final",
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("outer"), "value"),
              [],
            ),
          ),
          AST.stringInterpolation([
            AST.identifier("trace"),
            AST.identifier("final"),
          ]),
        ]),
        manifest: {
          description: "Nested proc value() calls resolve without deadlock under the serial executor",
          expect: {
            result: { kind: "string", value: "OIJXdone" },
          },
        },
      },
];

export default procReentrancyFixtures;
