import * as ast from '../ast';

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value); // Default i32
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const binaryOp = (op: string, left: ast.Expression, right: ast.Expression) => ast.binaryExpression(op, left, right);
const unaryOp = (op: ast.UnaryExpression['operator'], operand: ast.Expression) => ast.unaryExpression(op, operand);
const printCall = (label: string, expr: ast.Expression) => ast.functionCall(
    ident('print'),
    [ast.stringInterpolation([
        string(label + ": "),
        expr // Expression to be evaluated and interpolated
    ])]
);

const voidType = type('void');

// --- Main Program Logic ---

const mainFunction = ast.functionDefinition(
    'main',
    [], // No parameters for main
    ast.blockExpression([
        // Simple operations
        printCall("1 + 2", binaryOp('+', int(1), int(2))),
        printCall("5 - 3", binaryOp('-', int(5), int(3))),
        printCall("4 * 6", binaryOp('*', int(4), int(6))),
        printCall("10 / 2", binaryOp('/', int(10), int(2))), // Integer division expected
        printCall("11 / 3", binaryOp('/', int(11), int(3))), // Integer division expected
        printCall("10 % 3", binaryOp('%', int(10), int(3))),
        printCall("-5", unaryOp('-', int(5))),

        // Precedence
        printCall("2 + 3 * 4", binaryOp('+', int(2), binaryOp('*', int(3), int(4)))), // Should be 2 + 12 = 14
        printCall("(2 + 3) * 4", binaryOp('*', binaryOp('+', int(2), int(3)), int(4))), // Should be 5 * 4 = 20

        // TODO: Add more tests (floats, bigints, different types, bitwise) when interpreter supports them fully

        ast.returnStatement() // Explicit return void
    ]),
    voidType // Main returns void
);

// --- Module Definition ---
const arithmeticModule = ast.module(
    [
        mainFunction
    ],
    [
        // Import 'print' from 'io' (or 'core' depending on stdlib structure)
        ast.importStatement(['io'], false, [ast.importSelector('print')]),
    ]
);

// Export the final AST module
export default arithmeticModule;
