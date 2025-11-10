import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const procReentrancyFixtures: Fixture[] = [
  {
        name: "concurrency/future_value_reentrancy",
        module: AST.module([
          AST.assign("inner_started", AST.bool(false)),
          AST.assign("inner_completed", AST.bool(false)),
          AST.assign("outer_started", AST.bool(false)),
          AST.assign("outer_completed", AST.bool(false)),
          AST.assign("inner_value_seen", AST.bool(false)),
          AST.assign(
            "inner",
            AST.spawnExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("inner_started"),
                  AST.bool(true),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("inner_completed"),
                  AST.bool(true),
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
                  AST.identifier("outer_started"),
                  AST.bool(true),
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
                  AST.identifier("inner_value_seen"),
                  AST.binaryExpression(
                    "==",
                    AST.identifier("result"),
                    AST.stringLiteral("X"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("outer_completed"),
                  AST.bool(true),
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
          AST.arrayLiteral([
            AST.identifier("inner_started"),
            AST.identifier("inner_completed"),
            AST.identifier("outer_started"),
            AST.identifier("inner_value_seen"),
            AST.identifier("outer_completed"),
            AST.binaryExpression(
              "==",
              AST.identifier("final"),
              AST.stringLiteral("done"),
            ),
          ]),
        ]),
        manifest: {
          description: "Nested future.value() calls (re-entrancy) resolve without deadlock",
          expect: {
            result: {
              kind: "array",
              elements: [
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
              ],
            },
          },
        },
      },

  {
        name: "concurrency/proc_value_reentrancy",
        module: AST.module([
          AST.assign("inner_started", AST.bool(false)),
          AST.assign("inner_completed", AST.bool(false)),
          AST.assign("outer_started", AST.bool(false)),
          AST.assign("outer_completed", AST.bool(false)),
          AST.assign("inner_value_seen", AST.bool(false)),
          AST.assign(
            "inner",
            AST.procExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "=",
                  AST.identifier("inner_started"),
                  AST.bool(true),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("inner_completed"),
                  AST.bool(true),
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
                  AST.identifier("outer_started"),
                  AST.bool(true),
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
                  AST.identifier("inner_value_seen"),
                  AST.binaryExpression(
                    "==",
                    AST.identifier("result"),
                    AST.stringLiteral("X"),
                  ),
                ),
                AST.assignmentExpression(
                  "=",
                  AST.identifier("outer_completed"),
                  AST.bool(true),
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
          AST.arrayLiteral([
            AST.identifier("inner_started"),
            AST.identifier("inner_completed"),
            AST.identifier("outer_started"),
            AST.identifier("inner_value_seen"),
            AST.identifier("outer_completed"),
            AST.binaryExpression(
              "==",
              AST.identifier("final"),
              AST.stringLiteral("done"),
            ),
          ]),
        ]),
        manifest: {
          description: "Nested proc value() calls resolve without deadlock under the serial executor",
          expect: {
            result: {
              kind: "array",
              elements: [
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
                { kind: "bool", value: true },
              ],
            },
          },
        },
      },
];

export default procReentrancyFixtures;
