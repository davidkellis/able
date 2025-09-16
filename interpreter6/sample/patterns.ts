import * as AST from "../ast";

// Define a Point struct
const pointStruct = AST.structDefinition(
  "Point",
  [
    AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x"),
    AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "y"),
  ],
  "named"
);

// Main function to demonstrate pattern matching
const main = AST.functionDefinition(
  "main",
  [],
  AST.blockExpression([
    // --- Destructuring Assignment ---
    AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\n--- Destructuring Assignment ---")]),
    AST.assignmentExpression(
      ":=",
      AST.identifier("p1"),
      AST.structLiteral(
        [
          AST.structFieldInitializer(AST.integerLiteral(10), "x"),
          AST.structFieldInitializer(AST.integerLiteral(20), "y"),
        ],
        false, // named
        "Point"
      )
    ),
    // Destructure Point into variables x and y
    AST.assignmentExpression(
      ":=",
      AST.structPattern([
        AST.structPatternField(AST.identifier("x_val"), "x"),
        AST.structPatternField(AST.identifier("y_val"), "y"),
      ], false, "Point"),
      AST.identifier("p1")
    ),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Destructured p1: x="), AST.identifier("x_val"), AST.stringLiteral(", y="), AST.identifier("y_val")])]),

    // Destructure array
    AST.assignmentExpression(
      ":=",
      AST.identifier("my_array"),
      AST.arrayLiteral([
        AST.integerLiteral(1),
        AST.integerLiteral(2),
        AST.integerLiteral(3),
        AST.integerLiteral(4)
      ])
    ),
    AST.assignmentExpression(
      ":=",
      AST.arrayPattern(
        [AST.identifier("first"), AST.wildcardPattern()], // Match first, ignore second
        AST.identifier("rest") // Capture the rest
      ),
      AST.identifier("my_array")
    ),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Destructured array: first="), AST.identifier("first")])]),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Rest of array: "), AST.identifier("rest")])]), // Note: array printing needs improvement

    // --- Match Expression ---
    AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\n--- Match Expression ---")]),
    AST.assignmentExpression(":=", AST.identifier("val"), AST.integerLiteral(5)),

    AST.assignmentExpression(
        ":=",
        AST.identifier("match_result_1"),
        AST.matchExpression(
            AST.identifier("val"),
            [
                AST.matchClause(AST.literalPattern(AST.integerLiteral(1)), AST.stringLiteral("one")),
                AST.matchClause(AST.literalPattern(AST.integerLiteral(5)), AST.stringLiteral("five")),
                AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("other"))
            ]
        )
    ),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Match result 1 (val=5): "), AST.identifier("match_result_1")])]),

    // Match with identifier binding and guard
    AST.assignmentExpression(
        ":=",
        AST.identifier("match_result_2"),
        AST.matchExpression(
            AST.integerLiteral(10),
            [
                AST.matchClause(
                    AST.identifier("n"), // Bind to n
                    AST.stringInterpolation([AST.stringLiteral("matched n (<10): "), AST.identifier("n")]),
                    AST.binaryExpression("<", AST.identifier("n"), AST.integerLiteral(10)) // guard: n < 10
                ),
                 AST.matchClause(
                    AST.identifier("n"), // Bind to n
                    AST.stringInterpolation([AST.stringLiteral("matched n (>=10): "), AST.identifier("n")]),
                    undefined // No guard, default case for numbers
                )
            ]
        )
    ),
     AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Match result 2 (subj=10): "), AST.identifier("match_result_2")])]),

    // Match on struct
    AST.assignmentExpression(
        ":=",
        AST.identifier("match_result_3"),
        AST.matchExpression(
            AST.identifier("p1"),
            [
                 AST.matchClause(
                     AST.structPattern([AST.structPatternField(AST.literalPattern(AST.integerLiteral(0)),"x")], false, "Point"),
                     AST.stringLiteral("Origin X")
                 ),
                 AST.matchClause(
                     AST.structPattern([
                         AST.structPatternField(AST.identifier("px"), "x"),
                         AST.structPatternField(AST.identifier("py"), "y")
                     ], false, "Point"),
                     AST.stringInterpolation([AST.stringLiteral("Point at ("), AST.identifier("px"), AST.stringLiteral(", "), AST.identifier("py"), AST.stringLiteral(")")])
                 ),
                 AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Not a point"))
            ]
        )
    ),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Match result 3 (subj=p1): "), AST.identifier("match_result_3")])]),

    // Match on array
    AST.assignmentExpression(
        ":=",
        AST.identifier("match_result_4"),
        AST.matchExpression(
            AST.arrayLiteral([AST.integerLiteral(100), AST.integerLiteral(200)]),
            [
                AST.matchClause(AST.arrayPattern([]), AST.stringLiteral("Empty array")),
                AST.matchClause(AST.arrayPattern([AST.identifier("x")]), AST.stringInterpolation([AST.stringLiteral("Single element: "), AST.identifier("x")])),
                AST.matchClause(
                    AST.arrayPattern([AST.identifier("a"), AST.identifier("b")], AST.wildcardPattern()),
                    AST.stringInterpolation([AST.stringLiteral("Starts with: "), AST.identifier("a"), AST.stringLiteral(", "), AST.identifier("b")])
                ),
                AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Some other array"))
            ]
        )
    ),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Match result 4 (subj=[100, 200]): "), AST.identifier("match_result_4")])]),

  ])
);

// Create the module
export default AST.module(
  [
    pointStruct,
    main,
    // Need print definition if not globally available
    // Assuming print is a builtin provided by the interpreter
  ],
  // Assuming print is builtin and doesn't need explicit import here
  // [AST.importStatement(["io"], false, [AST.importSelector("print")])]
);
