import * as AST from "../ast";

// Define a Counter struct
const counterStruct = AST.structDefinition(
  "Counter",
  [AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "value")],
  "named"
);

// Define methods for Counter
const counterMethods = AST.methodsDefinition(
  AST.simpleTypeExpression("Counter"), // Target type
  [
    // Static method: origin() -> Counter
    AST.functionDefinition(
      "origin",
      [], // No 'self' parameter indicates static
      AST.blockExpression([
        AST.returnStatement(
          AST.structLiteral(
            [AST.structFieldInitializer(AST.integerLiteral(0), "value")],
            false, // named
            "Counter"
          )
        )
      ]),
      AST.simpleTypeExpression("Counter") // Return type
    ),

    // Instance method: increment(self)
    AST.functionDefinition(
      "increment",
      [AST.functionParameter("self")], // 'self' parameter indicates instance method
      AST.blockExpression([
        // self.value += 1
        AST.assignmentExpression(
          "+=",
          AST.memberAccessExpression(AST.identifier("self"), "value"),
          AST.integerLiteral(1)
        )
        // No explicit return, defaults to nil/void
      ])
      // No explicit return type needed if void/nil is intended
    )
  ]
);

// Main function
const main = AST.functionDefinition("main", [], AST.blockExpression([
  AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\n--- Static Method Tests ---")]),

  // Call static method
  AST.assignmentExpression(
    ":=",
    AST.identifier("c1"),
    AST.functionCall(AST.memberAccessExpression(AST.identifier("Counter"), "origin"), [])
  ),
  AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Counter created via static origin(): "), AST.identifier("c1")])]), // Needs valueToString for structs

  // Call instance method
  AST.functionCall(AST.memberAccessExpression(AST.identifier("c1"), "increment"), []),
  AST.functionCall(AST.memberAccessExpression(AST.identifier("c1"), "increment"), []),
  AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Counter after c1.increment() x2: "), AST.identifier("c1")])]),

]));

// Export module
export default AST.module([
  counterStruct,
  counterMethods,
  main
  // Assuming print is a builtin
]);
