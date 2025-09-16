import * as AST from "../ast";

// Define a custom error struct
const divisionErrorStruct = AST.structDefinition("DivisionError", [
    AST.structFieldDefinition(AST.simpleTypeExpression("string"), "message")
], "named");

// Function that might raise an error
const divide = AST.functionDefinition(
    "divide",
    [AST.functionParameter("a", AST.simpleTypeExpression("i32")), AST.functionParameter("b", AST.simpleTypeExpression("i32"))],
    AST.blockExpression([
        AST.ifExpression(
            AST.binaryExpression("==", AST.identifier("b"), AST.integerLiteral(0)),
            // If b is 0, raise the custom error
            AST.blockExpression([
                AST.raiseStatement(
                    AST.structLiteral(
                        [AST.structFieldInitializer(AST.stringLiteral("Division by zero!"), "message")],
                        false, // named
                        "DivisionError"
                    )
                )
            ]),
            // Else (no condition needed for final 'or')
            [
                AST.orClause(
                    AST.blockExpression([
                         AST.returnStatement(AST.binaryExpression("/", AST.identifier("a"), AST.identifier("b")))
                    ])
                )
            ]
        )
    ]),
    AST.simpleTypeExpression("i32") // Return type (if successful)
);

// Main function
const main = AST.functionDefinition("main", [], AST.blockExpression([
    AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\n--- Exception Handling Tests ---")]),

    // --- Test Case 1: Successful division (no raise) ---
    AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nTest Case 1: 10 / 2")]),
    AST.assignmentExpression(":=", AST.identifier("result1"),
        AST.rescueExpression(
            AST.functionCall(AST.identifier("divide"), [AST.integerLiteral(10), AST.integerLiteral(2)]),
            [
                // This clause should not be reached
                AST.matchClause(
                    AST.structPattern([], false, "DivisionError"), // Match DivisionError (simplified pattern)
                    AST.stringLiteral("Caught DivisionError unexpectedly!")
                ),
                AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Caught unexpected error!"))
            ]
        )
    ),
    AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Result 1: "), AST.identifier("result1")])]),

    // --- Test Case 2: Division by zero, specific rescue ---
    AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nTest Case 2: 5 / 0")]),
    AST.assignmentExpression(":=", AST.identifier("result2"),
        AST.rescueExpression(
            AST.functionCall(AST.identifier("divide"), [AST.integerLiteral(5), AST.integerLiteral(0)]),
            [
                // Match specific error struct and bind message
                AST.matchClause(
                    AST.structPattern([AST.structPatternField(AST.identifier("msg"), "message")], false, "DivisionError"),
                    AST.stringInterpolation([AST.stringLiteral("Rescued successfully: "), AST.identifier("msg")])
                ),
                // Catch any other error (wildcard) - should not be reached here
                AST.matchClause(AST.wildcardPattern(), AST.stringLiteral("Caught some other error"))
            ]
        )
    ),
     AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Result 2: "), AST.identifier("result2")])]),

     // --- Test Case 3: Raise directly and rescue with wildcard ---
     AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nTest Case 3: raise string")]),
     AST.assignmentExpression(":=", AST.identifier("result3"),
        AST.rescueExpression(
            // Wrap the raise statement in a block expression
            AST.blockExpression([
                AST.raiseStatement(AST.stringLiteral("A manually raised string error"))
            ]),
            [
                 // Match DivisionError - won't match here
                 AST.matchClause(AST.structPattern([], false, "DivisionError"), AST.stringLiteral("Caught DivisionError (wrong type)")),
                 // Match any value and bind it
                 AST.matchClause(AST.identifier("err_val"),
                    AST.stringInterpolation([AST.stringLiteral("Caught generic error: "), AST.identifier("err_val")])
                 )
            ]
        )
    ),
     AST.functionCall(AST.identifier("print"), [AST.stringInterpolation([AST.stringLiteral("Result 3: "), AST.identifier("result3")])]),

    // --- Test Case 4: Unrescued error (demonstrates propagation) ---
    AST.functionCall(AST.identifier("print"), [AST.stringLiteral("\nTest Case 4: Unrescued 1 / 0 (will terminate if uncommented)")]),
    // Uncomment the following line to test unrescued propagation:
    // AST.functionCall(AST.identifier("divide"), [AST.integerLiteral(1), AST.integerLiteral(0)])

]));

// Export module
export default AST.module([
    divisionErrorStruct,
    divide,
    main,
    // Assuming print is a builtin provided by the interpreter
]);
