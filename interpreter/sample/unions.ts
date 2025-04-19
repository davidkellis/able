import * as AST from "../ast";

// --- Define Variant Structs ---
const circleStruct = AST.structDefinition(
  "Circle",
  [AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "radius")],
  "named"
);

const rectangleStruct = AST.structDefinition(
  "Rectangle",
  [
    AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "width"),
    AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "height"),
  ],
  "named"
);

const pointStruct = AST.structDefinition(
    "Point",
    [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "x")],
    "named"
);

// --- Define Union ---
const shapeUnion = AST.unionDefinition(
  "Shape",
  [
    AST.simpleTypeExpression("Circle"),
    AST.simpleTypeExpression("Rectangle"),
    // Note: Can include primitives or other types too, e.g.:
    // AST.simpleTypeExpression("string")
  ]
);

// --- Function to process a Shape ---
const processShape = AST.functionDefinition(
  "processShape",
  [AST.functionParameter("s", AST.simpleTypeExpression("Shape"))], // Parameter typed as the union
  AST.blockExpression([
    AST.assignmentExpression(
      ":=",
      AST.identifier("description"),
      AST.matchExpression(
        AST.identifier("s"), // Match on the union-typed parameter
        [
          // Match Circle variant using struct pattern
          AST.matchClause(
            AST.structPattern([AST.structPatternField(AST.identifier("r"), "radius")], false, "Circle"),
            AST.stringInterpolation([AST.stringLiteral("It is a Circle with radius: "), AST.identifier("r")])
          ),
          // Match Rectangle variant using struct pattern
          AST.matchClause(
            AST.structPattern([
              AST.structPatternField(AST.identifier("w"), "width"),
              AST.structPatternField(AST.identifier("h"), "height")
            ], false, "Rectangle"),
            AST.stringInterpolation([AST.stringLiteral("It is a Rectangle with width "), AST.identifier("w"), AST.stringLiteral(" and height "), AST.identifier("h")])
          ),
          // Wildcard for any other variants (if any were added to the union)
          AST.matchClause(
            AST.wildcardPattern(),
            AST.stringLiteral("It is some other shape or value.")
          )
        ]
      )
    ),
    AST.functionCall(AST.identifier("print"), [AST.identifier("description")])
  ])
);

// --- Main Function ---
const main = AST.functionDefinition("main", [], AST.blockExpression([
  AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\n--- Union Tests ---")]),

  // Create instances of the variants
  AST.assignmentExpression(
    ":=",
    AST.identifier("c"),
    AST.structLiteral([AST.structFieldInitializer(AST.floatLiteral(5.0), "radius")], false, "Circle")
  ),
  AST.assignmentExpression(
    ":=",
    AST.identifier("r"),
    AST.structLiteral([
      AST.structFieldInitializer(AST.floatLiteral(4.0), "width"),
      AST.structFieldInitializer(AST.floatLiteral(6.0), "height")
    ], false, "Rectangle")
  ),
   AST.assignmentExpression(
    ":=",
    AST.identifier("p"), // Not actually part of the Shape union
    AST.structLiteral([AST.structFieldInitializer(AST.integerLiteral(1), "x")], false, "Point")
  ),

  // Process the shapes (which conform to the union type)
  AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nProcessing Circle:")]),
  AST.functionCall(AST.identifier("processShape"), [AST.identifier("c")]),

  AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nProcessing Rectangle:")]),
  AST.functionCall(AST.identifier("processShape"), [AST.identifier("r")]),

  // Process something not explicitly in the union (will hit wildcard)
  // Note: Without type checking, the interpreter relies solely on pattern matching.
  // A real type checker would flag passing 'p' to processShape if 'Point' isn't in 'Shape'.
  AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nProcessing Point (falls through):")]),
  AST.functionCall(AST.identifier("processShape"), [AST.identifier("p")])

]));

// Export module
export default AST.module([
  circleStruct,
  rectangleStruct,
  pointStruct, // Define Point even if not in union for the example
  shapeUnion,
  processShape,
  main
  // Assuming print is a builtin
]);
