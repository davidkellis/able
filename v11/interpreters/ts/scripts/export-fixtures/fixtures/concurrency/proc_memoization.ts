import { AST } from "../../../context";
import type { Fixture } from "../../../types";

const procMemoizationFixtures: Fixture[] = [
  {
        name: "concurrency/proc_value_memoization",
        module: AST.module([
          AST.assign("count", AST.integerLiteral(0)),
          AST.assign(
            "handle",
            AST.procExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "+=",
                  AST.identifier("count"),
                  AST.integerLiteral(1),
                ),
                AST.integerLiteral(21),
              ]),
            ),
          ),
          AST.assign(
            "first",
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("handle"), "value"),
              [],
            ),
          ),
          AST.assign(
            "second",
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("handle"), "value"),
              [],
            ),
          ),
          AST.arrayLiteral([
            AST.identifier("first"),
            AST.identifier("second"),
            AST.identifier("count"),
          ]),
        ]),
        manifest: {
          description: "Proc handle value() memoises successful results",
          expect: {
            result: {
              kind: "array",
              length: 3,
              elements: [
                { kind: "i32", value: 21n },
                { kind: "i32", value: 21n },
                { kind: "i32", value: 1n },
              ],
            },
          },
        },
      },

  {
        name: "concurrency/proc_value_cancel_memoization",
        module: AST.module([
          AST.assign("ran", AST.integerLiteral(0)),
          AST.assign(
            "handle",
            AST.procExpression(
              AST.blockExpression([
                AST.assignmentExpression(
                  "+=",
                  AST.identifier("ran"),
                  AST.integerLiteral(1),
                ),
                AST.integerLiteral(7),
              ]),
            ),
          ),
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("handle"), "cancel"),
            [],
          ),
          AST.assign(
            "first_raw",
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("handle"), "value"),
              [],
            ),
          ),
          AST.assign(
            "second_raw",
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("handle"), "value"),
              [],
            ),
          ),
          AST.assign(
            "first",
            AST.stringInterpolation([AST.identifier("first_raw")]),
          ),
          AST.assign(
            "second",
            AST.stringInterpolation([AST.identifier("second_raw")]),
          ),
          AST.assign(
            "ran_string",
            AST.stringInterpolation([AST.identifier("ran")]),
          ),
          AST.arrayLiteral([
            AST.identifier("first"),
            AST.identifier("second"),
            AST.identifier("ran_string"),
          ]),
        ]),
        manifest: {
          description: "Cancelled proc handle value() calls return identical memoized errors",
          expect: {
            result: {
              kind: "array",
              length: 3,
              elements: [
                { kind: "string", value: "Proc cancelled" },
                { kind: "string", value: "Proc cancelled" },
                { kind: "string", value: "0" },
              ],
            },
          },
        },
      },
];

export default procMemoizationFixtures;
