import * as ast from '../ast';

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value);
const bool = (value: boolean) => ast.booleanLiteral(value);
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const binaryOp = (op: string, left: ast.Expression, right: ast.Expression) => ast.binaryExpression(op, left, right);
const printCall = (expr: ast.Expression) => ast.functionCall(ident('print'), [expr]);
const assign = (target: string, value: ast.Expression, op: ':=' | '=' = ':=') => ast.assignmentExpression(op, ident(target), value);
const arrayLit = (elements: ast.Expression[]) => ast.arrayLiteral(elements);
const rangeLit = (start: ast.Expression, end: ast.Expression, inclusive: boolean) => ast.rangeExpression(start, end, inclusive);

const voidType = type('void');
const intType = type('i32'); // Assuming default int

// --- Test Cases ---

// 1. Simple While Loop
const test1 = ast.blockExpression([
    printCall(string("Test 1: While Loop Start")),
    assign("counter1", int(0)),
    ast.whileLoop(
        binaryOp('<', ident("counter1"), int(3)),
        ast.blockExpression([
            printCall(ast.stringInterpolation([string("  While counter: "), ident("counter1")])),
            assign("counter1", binaryOp('+', ident("counter1"), int(1)), '=') // Reassign with '='
        ])
    ),
    printCall(string("Test 1: While Loop End"))
]);

// 2. For Loop over Array
const test2 = ast.blockExpression([
    printCall(string("Test 2: For Loop (Array) Start")),
    assign("items", arrayLit([string("apple"), string("banana"), string("cherry")])),
    ast.forLoop(
        ident("item"), // Simple identifier pattern
        ident("items"),
        ast.blockExpression([
            printCall(ast.stringInterpolation([string("  For item: "), ident("item")]))
        ])
    ),
    printCall(string("Test 2: For Loop (Array) End"))
]);

// 3. For Loop over Inclusive Range
const test3 = ast.blockExpression([
    printCall(string("Test 3: For Loop (Inclusive Range 1..3) Start")),
    assign("sum3", int(0)),
    ast.forLoop(
        ident("i"),
        rangeLit(int(1), int(3), true), // 1..3
        ast.blockExpression([
            printCall(ast.stringInterpolation([string("  Range i: "), ident("i")])),
            assign("sum3", binaryOp('+', ident("sum3"), ident("i")), '=')
        ])
    ),
    printCall(ast.stringInterpolation([string("Test 3: Range Sum = "), ident("sum3")])), // Should be 6
    printCall(string("Test 3: For Loop (Range) End"))
]);

// 4. For Loop over Exclusive Range
const test4 = ast.blockExpression([
    printCall(string("Test 4: For Loop (Exclusive Range 5...8) Start")),
    assign("sum4", int(0)),
    ast.forLoop(
        ident("j"),
        rangeLit(int(5), int(8), false), // 5...8 (5, 6, 7)
        ast.blockExpression([
            printCall(ast.stringInterpolation([string("  Range j: "), ident("j")])),
            assign("sum4", binaryOp('+', ident("sum4"), ident("j")), '=')
        ])
    ),
    printCall(ast.stringInterpolation([string("Test 4: Range Sum = "), ident("sum4")])), // Should be 18 (5+6+7)
    printCall(string("Test 4: For Loop (Range) End"))
]);

// 5. For Loop with Destructuring (Placeholder - requires struct/tuple support)
// const test5 = ast.blockExpression([
//     printCall(string("Test 5: For Loop (Destructuring) Start")),
//     // assign("pairs", arrayLit([ ??? struct {1, "a"}, struct {2, "b"} ??? ])),
//     // ast.forLoop(
//     //     ast.structPattern([ast.structPatternField(ident("num")), ast.structPatternField(ident("char"))], true), // { num, char }
//     //     ident("pairs"),
//     //     ast.blockExpression([
//     //         printCall(ast.stringInterpolation([string("  Num: "), ident("num"), string(", Char: "), ident("char")]))
//     //     ])
//     // ),
//     printCall(string("Test 5: For Loop (Destructuring) End - SKIPPED"))
// ]);

// 6. While loop with break (Basic - requires break implementation)
const test6 = ast.blockExpression([
    printCall(string("Test 6: While Loop with Break Start")),
    assign("counter6", int(0)),
    ast.whileLoop(
        bool(true), // Infinite loop condition
        ast.blockExpression([
            printCall(ast.stringInterpolation([string("  Break counter: "), ident("counter6")])),
            assign("counter6", binaryOp('+', ident("counter6"), int(1)), '='),
            ast.ifExpression( // If condition to break
                binaryOp('>=', ident("counter6"), int(3)),
                ast.blockExpression([
                    printCall(string("    Breaking loop...")),
                    ast.breakStatement(ast.identifier("'implicit_loop'"), ast.nilLiteral()) // Placeholder break
                ]),
                []
            )
        ])
    ),
    printCall(string("Test 6: While Loop with Break End (Should execute if break works)"))
]);


// --- Main Function ---
const mainFunction = ast.functionDefinition(
    'main',
    [],
    ast.blockExpression([
        test1,
        test2,
        test3,
        test4,
        // test5, // Skipping destructuring test for now
        test6, // Add break test (will likely warn/error until break is fully done)
        ast.returnStatement() // Explicit return void
    ]),
    voidType
);

// --- Module Definition ---
const loopsModule = ast.module(
    [
        mainFunction
    ],
    [
        ast.importStatement(['io'], false, [ast.importSelector('print')]),
    ]
);

// Export the final AST module
export default loopsModule;
