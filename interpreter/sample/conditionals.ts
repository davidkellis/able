import * as ast from '../ast';

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value);
const bool = (value: boolean) => ast.booleanLiteral(value);
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const binaryOp = (op: string, left: ast.Expression, right: ast.Expression) => ast.binaryExpression(op, left, right);
const printCall = (expr: ast.Expression) => ast.functionCall(ident('print'), [expr]);
const assign = (target: string, value: ast.Expression) => ast.assignmentExpression(':=', ident(target), value); // Use := for declaration

const voidType = type('void');
const stringType = type('string');

// --- Test Cases ---

// 1. Simple if (true)
const test1 = printCall(
    ast.ifExpression(
        bool(true),
        ast.blockExpression([string("Test 1: Simple if (true) executed")]),
        [] // No 'or' clauses
    )
);

// 2. Simple if (false) - should print nil (or nothing if print(nil) does nothing)
const test2 = printCall(
    ast.ifExpression(
        bool(false),
        ast.blockExpression([string("Test 2: Simple if (false) - SHOULD NOT EXECUTE")]),
        []
    )
);

// 3. if/or (first condition true)
const test3 = printCall(
    ast.ifExpression(
        bool(true),
        ast.blockExpression([string("Test 3: if/or - First branch executed")]),
        [
            ast.orClause(ast.blockExpression([string("Test 3: if/or - Second branch SHOULD NOT EXECUTE")]), bool(true))
        ]
    )
);

// 4. if/or (second condition true)
const test4 = printCall(
    ast.ifExpression(
        bool(false),
        ast.blockExpression([string("Test 4: if/or - First branch SHOULD NOT EXECUTE")]),
        [
            ast.orClause(ast.blockExpression([string("Test 4: if/or - Second branch executed")]), bool(true))
        ]
    )
);

// 5. if/or/or {} (else)
const test5 = printCall(
    ast.ifExpression(
        binaryOp('>', int(5), int(10)), // false
        ast.blockExpression([string("Test 5: if/or/else - First branch SHOULD NOT EXECUTE")]),
        [
            ast.orClause(ast.blockExpression([string("Test 5: if/or/else - Second branch SHOULD NOT EXECUTE")]), binaryOp('==', int(1), int(2))), // false
            ast.orClause(ast.blockExpression([string("Test 5: if/or/else - Else branch executed")])) // Final else
        ]
    )
);

// 6. if expression result assignment
const test6_if = ast.ifExpression(
    binaryOp('<', int(3), int(5)), // true
    ast.blockExpression([string("Result A")]),
    [
        ast.orClause(ast.blockExpression([string("Result B")])) // else
    ]
);
const test6 = ast.blockExpression([
    assign("result6", test6_if),
    printCall(ast.stringInterpolation([string("Test 6: if result = "), ident("result6")]))
]);


// 7. Nested if
const test7_outer_if = ast.ifExpression(
    bool(true),
    ast.blockExpression([
        printCall(string("Test 7: Outer if true")),
        ast.ifExpression( // Inner if
            binaryOp('!=', int(10), int(10)), // false
            ast.blockExpression([string("Test 7: Inner if SHOULD NOT EXECUTE")]),
            [
                ast.orClause(
                  ast.blockExpression([
                    printCall(string("Test 7: Inner else executed"))
                  ])
                ) // Inner else
            ]
        )
    ]),
    [
        ast.orClause(ast.blockExpression([string("Test 7: Outer else SHOULD NOT EXECUTE")]))
    ]
);
const test7 = test7_outer_if; // Directly use the outer if as the statement


// --- Main Function ---
const mainFunction = ast.functionDefinition(
    'main',
    [],
    ast.blockExpression([
        test1,
        test2,
        test3,
        test4,
        test5,
        test6,
        test7,
        ast.returnStatement() // Explicit return void
    ]),
    voidType
);

// --- Module Definition ---
const conditionalsModule = ast.module(
    [
        mainFunction
    ],
    [
        ast.importStatement(['io'], false, [ast.importSelector('print')]),
    ]
);

// Export the final AST module
export default conditionalsModule;
