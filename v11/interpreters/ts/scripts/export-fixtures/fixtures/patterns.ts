import { AST } from "../../context";
import type { Fixture } from "../../types";

const patternsFixtures: Fixture[] = [
  {
      name: "patterns/array_destructuring",
      module: AST.module([
        AST.assign("arr", AST.arr(AST.int(1), AST.int(2), AST.int(3), AST.int(4))),
        AST.assign(AST.arrP([AST.id("first"), AST.id("second")], AST.id("rest")), AST.id("arr")),
        AST.assign(AST.arrP([AST.id("third"), AST.id("fourth")]), AST.id("rest")),
        AST.bin("+", AST.id("first"), AST.bin("+", AST.id("second"), AST.id("third"))),
      ]),
      manifest: {
        description: "Array destructuring assignment extracts prefix and rest",
        expect: {
          result: { kind: "i32", value: 6n },
        },
      },
    },

  {
      name: "patterns/struct_positional_destructuring",
      module: AST.module([
        AST.structDefinition(
          "Pair",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32")),
          ],
          "positional",
        ),
        AST.assign(
          "pair",
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.integerLiteral(4)),
              AST.structFieldInitializer(AST.integerLiteral(8)),
            ],
            true,
            "Pair",
          ),
        ),
        AST.assign(
          AST.structPattern(
            [
              AST.structPatternField(AST.identifier("first")),
              AST.structPatternField(AST.identifier("second")),
            ],
            true,
            "Pair",
          ),
          AST.identifier("pair"),
        ),
        AST.binaryExpression(
          "+",
          AST.identifier("first"),
          AST.identifier("second"),
        ),
      ]),
      manifest: {
        description: "Positional struct destructuring assignment binds tuple fields",
        expect: {
          result: { kind: "i32", value: 12n },
        },
      },
    },

  {
      name: "patterns/nested_struct_destructuring",
      module: AST.module([
        AST.structDefinition(
          "Point",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
          ],
          "named",
        ),
        AST.structDefinition(
          "Wrapper",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("Point"), "left"),
            AST.structFieldDefinition(AST.simpleTypeExpression("Point"), "right"),
            AST.structFieldDefinition(
              AST.genericTypeExpression(
                AST.simpleTypeExpression("Array"),
                [AST.simpleTypeExpression("i32")],
              ),
              "values",
            ),
          ],
          "named",
        ),
        AST.assign(
          "wrapper",
          AST.structLiteral(
            [
              AST.structFieldInitializer(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.integerLiteral(1), "x"),
                    AST.structFieldInitializer(AST.integerLiteral(2), "y"),
                  ],
                  false,
                  "Point",
                ),
                "left",
              ),
              AST.structFieldInitializer(
                AST.structLiteral(
                  [
                    AST.structFieldInitializer(AST.integerLiteral(3), "x"),
                    AST.structFieldInitializer(AST.integerLiteral(4), "y"),
                  ],
                  false,
                  "Point",
                ),
                "right",
              ),
              AST.structFieldInitializer(
                AST.arrayLiteral([
                  AST.integerLiteral(10),
                  AST.integerLiteral(20),
                  AST.integerLiteral(30),
                ]),
                "values",
              ),
            ],
            false,
            "Wrapper",
          ),
        ),
        AST.assign(
          AST.structPattern(
            [
              AST.structPatternField(
                AST.structPattern(
                  [
                    AST.structPatternField(AST.identifier("left_x"), "x"),
                    AST.structPatternField(AST.identifier("left_y"), "y"),
                  ],
                  false,
                  "Point",
                ),
                "left",
              ),
              AST.structPatternField(
                AST.structPattern(
                  [
                    AST.structPatternField(AST.wildcardPattern(), "x"),
                    AST.structPatternField(AST.identifier("right_y"), "y"),
                  ],
                  false,
                  "Point",
                ),
                "right",
              ),
              AST.structPatternField(
                AST.arrayPattern(
                  [AST.identifier("first_value")],
                  "rest_values",
                ),
                "values",
              ),
            ],
            false,
            "Wrapper",
          ),
          AST.identifier("wrapper"),
        ),
        AST.binaryExpression(
          "+",
          AST.binaryExpression(
            "+",
            AST.binaryExpression(
              "+",
              AST.identifier("left_x"),
              AST.identifier("right_y"),
            ),
            AST.identifier("first_value"),
          ),
          AST.index(AST.identifier("rest_values"), AST.integerLiteral(0)),
        ),
      ]),
      manifest: {
        description: "Nested struct and array patterns destructure composite value",
        expect: {
          result: { kind: "i32", value: 35n },
        },
      },
    },

  {
      name: "patterns/for_array_pattern",
      module: AST.module([
        AST.assign("pairs", AST.arr(
          AST.arr(AST.int(1), AST.int(2)),
          AST.arr(AST.int(3), AST.int(4)),
        )),
        AST.assign("sum", AST.int(0)),
        AST.forIn(
          AST.arrP([AST.id("x"), AST.id("y")]),
          AST.id("pairs"),
          AST.block(
            AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("x")), "="),
            AST.assign("sum", AST.bin("+", AST.id("sum"), AST.id("y")), "="),
          ),
        ),
        AST.id("sum"),
      ]),
      manifest: {
        description: "For-in loop destructures array elements",
        expect: {
          result: { kind: "i32", value: 10n },
        },
      },
    },

  {
      name: "patterns/typed_assignment",
      module: AST.module([
        AST.assign("value", AST.int(42)),
        AST.assign(
          AST.typedP(AST.id("n"), AST.ty("i32")),
          AST.id("value"),
        ),
        AST.id("n"),
      ]),
      manifest: {
        description: "Typed pattern enforces simple type on assignment",
        expect: {
          result: { kind: "i32", value: 42n },
        },
      },
    },

  {
      name: "patterns/typed_assignment_error",
      module: AST.module([
        AST.assign("value", AST.str("nope")),
        AST.assign(
          AST.typedP(AST.id("n"), AST.ty("i32")),
          AST.id("value"),
        ),
      ]),
      manifest: {
        description: "Typed pattern mismatch raises error",
        expect: {
          errors: ["Typed pattern mismatch in assignment: expected i32, got string"],
        },
      },
    },
];

export default patternsFixtures;
