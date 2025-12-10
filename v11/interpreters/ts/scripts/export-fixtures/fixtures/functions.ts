import { AST } from "../../context";
import type { Fixture } from "../../types";

const functionsFixtures: Fixture[] = [
  {
      name: "functions/implicit_generic_inference",
      module: AST.module([
        AST.functionDefinition(
          "choose",
          [
            AST.functionParameter("first", AST.simpleTypeExpression("T")),
            AST.functionParameter("second", AST.simpleTypeExpression("U")),
          ],
          AST.blockExpression([AST.identifier("first")]),
          AST.simpleTypeExpression("T"),
        ),
        AST.nil(),
      ]),
      manifest: {
        description: "Implicitly inferred generic parameters propagate types",
        expect: {
          result: { kind: "nil" },
        },
      },
    },

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
          result: { kind: "i32", value: 5n },
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
          result: { kind: "i32", value: 6n },
        },
      },
    },

  {
      name: "functions/bitwise_xor_operator",
      module: AST.module([
        AST.assign("lhs", AST.int(0b1010)),
        AST.assign("rhs", AST.int(0b1100)),
        AST.assign("mask", AST.bin(".^", AST.id("lhs"), AST.id("rhs"))),
        AST.identifier("mask"),
      ]),
      manifest: {
        description: "Bitwise xor operator is supported and returns integer result",
        expect: {
          result: { kind: "i32", value: 0b0110n },
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
          result: { kind: "i32", value: 7n },
        },
      },
    },

  {
      name: "functions/overload_resolution_success",
      module: AST.module([
        AST.functionDefinition(
          "pick",
          [AST.functionParameter("value", AST.simpleTypeExpression("String"))],
          AST.blockExpression([AST.int(20)]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.functionDefinition(
          "pick",
          [
            AST.functionParameter("value", AST.simpleTypeExpression("i32")),
            AST.functionParameter("note", AST.nullableTypeExpression(AST.simpleTypeExpression("String"))),
          ],
          AST.blockExpression([
            AST.matchExpression(AST.id("note"), [
              AST.matchClause(AST.literalPattern(AST.nil()), AST.int(30)),
              AST.matchClause(AST.wildcardPattern(), AST.int(40)),
            ]),
          ]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.functionDefinition(
          "pick",
          [AST.functionParameter("value", AST.simpleTypeExpression("bool"))],
          AST.blockExpression([AST.int(50)]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.structDefinition("Box", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")], "named"),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Box"),
          [
            AST.functionDefinition(
              "mark",
              [
                AST.functionParameter("self", AST.simpleTypeExpression("Box")),
                AST.functionParameter("delta", AST.simpleTypeExpression("i32")),
              ],
              AST.blockExpression([
                AST.bin("+", AST.member(AST.id("self"), "value"), AST.id("delta")),
              ]),
              AST.simpleTypeExpression("i32"),
            ),
            AST.functionDefinition(
              "mark",
              [
                AST.functionParameter("self", AST.simpleTypeExpression("Box")),
                AST.functionParameter("tag", AST.nullableTypeExpression(AST.simpleTypeExpression("String"))),
              ],
              AST.blockExpression([
                AST.matchExpression(AST.id("tag"), [
                  AST.matchClause(
                    AST.literalPattern(AST.nil()),
                    AST.bin("+", AST.member(AST.id("self"), "value"), AST.int(100)),
                  ),
                  AST.matchClause(
                    AST.wildcardPattern(),
                    AST.bin("+", AST.member(AST.id("self"), "value"), AST.int(200)),
                  ),
                ]),
              ]),
              AST.simpleTypeExpression("i32"),
            ),
          ],
        ),
        AST.assign("box", AST.structLiteral([AST.structFieldInitializer(AST.int(5), "value")], false, "Box")),
        AST.arrayLiteral([
          AST.functionCall(AST.id("pick"), [AST.int(5)]),
          AST.functionCall(AST.id("pick"), [AST.int(5), AST.str("note")]),
          AST.functionCall(AST.id("pick"), [AST.str("ok")]),
          AST.functionCall(AST.id("pick"), [AST.bool(true)]),
          AST.functionCall(AST.member(AST.id("box"), "mark"), [AST.int(3)]),
          AST.functionCall(AST.member(AST.id("box"), "mark"), []),
          AST.functionCall(AST.member(AST.id("box"), "mark"), [AST.str("hey")]),
        ]),
      ]),
      manifest: {
        description: "Runtime overload resolution selects best matches for functions and methods (including nullable tails)",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 30n },
              { kind: "i32", value: 40n },
              { kind: "i32", value: 20n },
              { kind: "i32", value: 50n },
              { kind: "i32", value: 8n },
              { kind: "i32", value: 105n },
              { kind: "i32", value: 205n },
            ],
          },
          typecheckDiagnostics: [
            "typechecker: ../../../fixtures/ast/functions/overload_resolution_success/source.able:4:1 typechecker: duplicate declaration 'pick' (previous declaration at ../../../../fixtures/ast/functions/overload_resolution_success/source.able:1:1)",
            "typechecker: ../../../fixtures/ast/functions/overload_resolution_success/source.able:10:1 typechecker: duplicate declaration 'pick' (previous declaration at ../../../../fixtures/ast/functions/overload_resolution_success/source.able:1:1)",
            "typechecker: ../../../fixtures/ast/functions/overload_resolution_success/source.able:28:7 typechecker: argument 1 has type i32, expected String",
            "typechecker: ../../../fixtures/ast/functions/overload_resolution_success/source.able:28:15 typechecker: function expects 1 arguments, got 2",
            "typechecker: ../../../fixtures/ast/functions/overload_resolution_success/source.able:28:45 typechecker: argument 1 has type bool, expected String",
          ],
        },
      },
    },

  {
      name: "functions/overload_function_ambiguity",
      module: AST.module([
        AST.functionDefinition(
          "collide",
          [AST.functionParameter("x", AST.simpleTypeExpression("i32"))],
          AST.blockExpression([AST.int(1)]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.functionDefinition(
          "collide",
          [AST.functionParameter("x", AST.simpleTypeExpression("i32"))],
          AST.blockExpression([AST.int(2)]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.functionCall(AST.id("collide"), [AST.int(1)]),
      ]),
      manifest: {
        description: "Duplicate free function overloads surface an ambiguity error",
        expect: {
          errors: ["Ambiguous overload for collide"],
          typecheckDiagnostics: [
            "typechecker: ../../../fixtures/ast/functions/overload_function_ambiguity/source.able:4:1 typechecker: duplicate declaration 'collide' (previous declaration at ../../../../fixtures/ast/functions/overload_function_ambiguity/source.able:1:1)",
          ],
        },
      },
    },

  {
      name: "functions/overload_method_ambiguity",
      module: AST.module([
        AST.structDefinition("Bag", [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")], "named"),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Bag"),
          [
            AST.functionDefinition(
              "do",
              [
                AST.functionParameter("self", AST.simpleTypeExpression("Bag")),
                AST.functionParameter("x", AST.simpleTypeExpression("i32")),
              ],
              AST.blockExpression([AST.bin("+", AST.member(AST.id("self"), "value"), AST.id("x"))]),
              AST.simpleTypeExpression("i32"),
            ),
            AST.functionDefinition(
              "do",
              [
                AST.functionParameter("self", AST.simpleTypeExpression("Bag")),
                AST.functionParameter("x", AST.simpleTypeExpression("i32")),
              ],
              AST.blockExpression([AST.member(AST.id("self"), "value")]),
              AST.simpleTypeExpression("i32"),
            ),
          ],
        ),
        AST.assign("bag", AST.structLiteral([AST.structFieldInitializer(AST.int(10), "value")], false, "Bag")),
        AST.functionCall(AST.member(AST.id("bag"), "do"), [AST.int(1)]),
      ]),
      manifest: {
        description: "Duplicate method overloads produce an ambiguity error on invocation",
        expect: {
          errors: ["Ambiguous overload for do"],
        },
      },
    },

  {
      name: "functions/ufcs_inherent_methods",
      module: AST.module([
        AST.structDefinition(
          "Point",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
          ],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Point"),
          [
            AST.functionDefinition(
              "norm",
              [AST.functionParameter("self", AST.simpleTypeExpression("Point"))],
              AST.blockExpression([
                AST.bin(
                  "+",
                  AST.bin("*", AST.member(AST.id("self"), "x"), AST.member(AST.id("self"), "x")),
                  AST.bin("*", AST.member(AST.id("self"), "y"), AST.member(AST.id("self"), "y")),
                ),
              ]),
              AST.simpleTypeExpression("i32"),
            ),
            AST.functionDefinition(
              "scale",
              [
                AST.functionParameter("self", AST.simpleTypeExpression("Point")),
                AST.functionParameter("factor", AST.simpleTypeExpression("i32")),
              ],
              AST.blockExpression([
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(
                      AST.bin("*", AST.member(AST.id("self"), "x"), AST.id("factor")),
                      "x",
                    ),
                    AST.structFieldInitializer(
                      AST.bin("*", AST.member(AST.id("self"), "y"), AST.id("factor")),
                      "y",
                    ),
                  ],
                  false,
                  "Point",
                ),
              ]),
              AST.simpleTypeExpression("Point"),
            ),
          ],
        ),
        AST.assign(
          "p",
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.int(3), "x"),
              AST.structFieldInitializer(AST.int(4), "y"),
            ],
            false,
            "Point",
          ),
        ),
        AST.assign("scaled", AST.functionCall(AST.memberAccessExpression(AST.id("p"), "scale"), [AST.int(2)])),
        AST.assign("pipeScaled", AST.functionCall(AST.memberAccessExpression(AST.id("p"), "scale"), [AST.int(3)])),
        AST.arrayLiteral([
          AST.functionCall(AST.memberAccessExpression(AST.id("p"), "norm"), []),
          AST.member(AST.id("scaled"), "x"),
          AST.member(AST.id("scaled"), "y"),
          AST.binaryExpression("|>", AST.id("p"), AST.implicitMemberExpression("norm")),
          AST.member(AST.id("pipeScaled"), "x"),
          AST.member(AST.id("pipeScaled"), "y"),
        ]),
      ]),
      manifest: {
        description: "UFCS resolves inherent instance methods (including pipeline usage)",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 25n },
              { kind: "i32", value: 6n },
              { kind: "i32", value: 8n },
              { kind: "i32", value: 25n },
              { kind: "i32", value: 9n },
              { kind: "i32", value: 12n },
            ],
          },
        },
      },
    },

  {
      name: "functions/ufcs_generic_overloads",
      module: AST.module([
        AST.structDefinition(
          "Box",
          [AST.structFieldDefinition(AST.simpleTypeExpression("T"), "value")],
          "named",
          [AST.genericParameter("T")],
        ),
        AST.functionDefinition(
          "describe",
          [
            AST.functionParameter(
              "box",
              AST.genericTypeExpression(AST.simpleTypeExpression("Box"), [AST.simpleTypeExpression("T")]),
            ),
          ],
          AST.blockExpression([AST.int(1)]),
          AST.simpleTypeExpression("i32"),
          [AST.genericParameter("T")],
        ),
        AST.functionDefinition(
          "describe",
          [
            AST.functionParameter(
              "box",
              AST.genericTypeExpression(AST.simpleTypeExpression("Box"), [AST.simpleTypeExpression("i32")]),
            ),
          ],
          AST.blockExpression([
            AST.binaryExpression("+", AST.memberAccessExpression(AST.id("box"), "value"), AST.int(1)),
          ]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.functionDefinition(
          "sum",
          [
            AST.functionParameter(
              "box",
              AST.genericTypeExpression(AST.simpleTypeExpression("Box"), [AST.simpleTypeExpression("i32")]),
            ),
            AST.functionParameter("delta", AST.simpleTypeExpression("i32")),
          ],
          AST.blockExpression([
            AST.binaryExpression("+", AST.memberAccessExpression(AST.id("box"), "value"), AST.id("delta")),
          ]),
          AST.simpleTypeExpression("i32"),
        ),
        AST.assign(
          "intBox",
          AST.structLiteral(
            [AST.structFieldInitializer(AST.int(5), "value")],
            false,
            "Box",
            undefined,
            [AST.simpleTypeExpression("i32")],
          ),
        ),
        AST.assign(
          "boolBox",
          AST.structLiteral(
            [AST.structFieldInitializer(AST.booleanLiteral(false), "value")],
            false,
            "Box",
            undefined,
            [AST.simpleTypeExpression("bool")],
          ),
        ),
        AST.arrayLiteral([
          AST.functionCall(AST.memberAccessExpression(AST.id("intBox"), "describe"), []),
          AST.functionCall(AST.memberAccessExpression(AST.id("boolBox"), "describe"), []),
          AST.binaryExpression("|>", AST.id("intBox"), AST.id("describe")),
          AST.functionCall(AST.id("sum"), [AST.id("intBox"), AST.int(2)]),
          AST.functionCall(AST.memberAccessExpression(AST.id("intBox"), "sum"), [AST.int(3)]),
          AST.binaryExpression("|>", AST.id("intBox"), AST.functionCall(AST.id("sum"), [AST.int(4)])),
          AST.binaryExpression("|>", AST.id("boolBox"), AST.id("describe")),
        ]),
      ]),
      manifest: {
        description: "UFCS resolves overloaded and generic free functions, matching pipe and method-style calls",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 6n },
              { kind: "i32", value: 1n },
              { kind: "i32", value: 6n },
              { kind: "i32", value: 7n },
              { kind: "i32", value: 8n },
              { kind: "i32", value: 9n },
              { kind: "i32", value: 1n },
            ],
          },
          typecheckDiagnostics: [
            "typechecker: ../../../fixtures/ast/functions/ufcs_generic_overloads/source.able:7:1 typechecker: duplicate declaration 'describe' (previous declaration at ../../../../fixtures/ast/functions/ufcs_generic_overloads/source.able:4:1)",
          ],
        },
      },
    },
];

export default functionsFixtures;
