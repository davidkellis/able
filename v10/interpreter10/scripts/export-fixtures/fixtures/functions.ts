import { AST } from "../../context";
import type { Fixture } from "../../types";

const functionsFixtures: Fixture[] = [
  {
      name: "functions/lambda_expression",
      module: AST.module([
        AST.assign(
          "adder",
          AST.lambdaExpression(
            [AST.param("x"), AST.param("y")],
            AST.bin("+", AST.id("x"), AST.id("y")),
          ),
        ),
        AST.functionCall(AST.id("adder"), [AST.int(2), AST.int(3)]),
      ]),
      manifest: {
        description: "Lambda expression returns computed sum",
        expect: {
          result: { kind: "i32", value: 5 },
        },
      },
    },

  {
      name: "functions/trailing_lambda_call",
      module: AST.module([
        AST.functionDefinition(
          "for_each",
          [AST.param("items"), AST.param("callback")],
          AST.blockExpression([
            AST.forIn(
              "item",
              AST.id("items"),
              AST.functionCall(AST.id("callback"), [AST.id("item")]),
            ),
          ]),
          undefined,
          undefined,
          undefined,
          false,
          false,
        ),
        AST.assign("numbers", AST.arr(AST.int(1), AST.int(2), AST.int(3))),
        AST.assign("total", AST.int(0)),
        AST.functionCall(
          AST.id("for_each"),
          [
            AST.id("numbers"),
            AST.lambdaExpression(
              [AST.param("n")],
              AST.assign("total", AST.id("n"), "+="),
            ),
          ],
          undefined,
          true,
        ),
        AST.id("total"),
      ]),
      manifest: {
        description: "Trailing lambda iterates array and accumulates values",
        expect: {
          result: { kind: "i32", value: 6 },
        },
      },
    },

  {
      name: "functions/hkt_interface_impl_ok",
      module: AST.module([
        AST.interfaceDefinition(
          "Wrapper",
          [
            AST.functionSignature(
              "wrap",
              [
                AST.functionParameter(
                  "self",
                  AST.genericTypeExpression(AST.simpleTypeExpression("Self"), [
                    AST.simpleTypeExpression("T"),
                  ]),
                ),
                AST.functionParameter("value", AST.simpleTypeExpression("T")),
              ],
              AST.genericTypeExpression(AST.simpleTypeExpression("Self"), [
                AST.simpleTypeExpression("T"),
              ]),
              [AST.genericParameter("T")],
            ),
          ],
          undefined,
          AST.genericTypeExpression(AST.simpleTypeExpression("F"), [AST.wildcardTypeExpression()]),
        ),
        AST.structDefinition(
          "Holder",
          [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
          "named",
          [AST.genericParameter("T")],
        ),
        AST.implementationDefinition(
          "Wrapper",
          AST.simpleTypeExpression("Holder"),
          [
            AST.functionDefinition(
              "wrap",
              [
                AST.functionParameter(
                  "self",
                  AST.genericTypeExpression(AST.simpleTypeExpression("Holder"), [
                    AST.simpleTypeExpression("T"),
                  ]),
                ),
                AST.functionParameter("value", AST.simpleTypeExpression("T")),
              ],
              AST.blockExpression([
                AST.structLiteral(
                  [AST.structFieldInitializer(AST.identifier("value"), "value")],
                  false,
                  "Holder",
                  undefined,
                  [AST.simpleTypeExpression("T")],
                ),
              ]),
              AST.genericTypeExpression(AST.simpleTypeExpression("Holder"), [
                AST.simpleTypeExpression("T"),
              ]),
              [AST.genericParameter("T")],
            ),
          ],
        ),
        AST.assign(
          "holder",
          AST.structLiteral(
            [AST.structFieldInitializer(AST.int(1), "value")],
            false,
            "Holder",
            undefined,
            [AST.simpleTypeExpression("i32")],
          ),
        ),
        AST.assign(
          "wrapped",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("holder"), "wrap"),
            [AST.int(7)],
            [AST.simpleTypeExpression("i32")],
          ),
        ),
        AST.memberAccessExpression(AST.identifier("wrapped"), "value"),
      ]),
      manifest: {
        description: "Higher-kinded interface impl accepts bare constructor when the interface declares 'for F _'",
        expect: {
          result: { kind: "i32", value: 7 },
        },
      },
    },
];

export default functionsFixtures;
