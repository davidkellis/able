import * as ast from '../ast';

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value);
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const printCall = (expr: ast.Expression) => ast.functionCall(ident('print'), [expr]);
const assign = (target: string, value: ast.Expression, op: ':=' | '=' = ':=') => ast.assignmentExpression(op, ident(target), value);
const memberAccess = (obj: string | ast.Expression, member: string | number) => ast.memberAccessExpression(
    typeof obj === 'string' ? ident(obj) : obj,
    typeof member === 'string' ? ident(member) : int(member) // Use int() for positional index
);
const assignMember = (obj: string, member: string | number, value: ast.Expression) => ast.assignmentExpression('=', memberAccess(obj, member), value);

const voidType = type('void');
const intType = type('i32');
const stringType = type('string');

// --- Struct Definitions ---

// 1. Singleton Struct
const singletonDef = ast.structDefinition("MySingleton", [], 'singleton');

// 2. Named Fields Struct
const namedStructDef = ast.structDefinition(
    "Point",
    [
        ast.structFieldDefinition(intType, "x"),
        ast.structFieldDefinition(intType, "y")
    ],
    'named'
);

// 3. Positional Fields Struct (Named Tuple)
const positionalStructDef = ast.structDefinition(
    "Color",
    [
        ast.structFieldDefinition(intType), // r
        ast.structFieldDefinition(intType), // g
        ast.structFieldDefinition(intType)  // b
    ],
    'positional'
);

// --- Test Logic ---

const testLogic = ast.blockExpression([
    // --- Singleton ---
    printCall(string("--- Singleton Struct ---")),
    assign("s1", ident("MySingleton")), // Instantiate by using the name
    printCall(ast.stringInterpolation([string("Singleton instance: "), ident("s1")])), // Should print <struct MySingleton> or similar

    // --- Named Fields ---
    printCall(string("--- Named Fields Struct ---")),
    // Instantiation
    assign("p1", ast.structLiteral(
        [
            ast.structFieldInitializer(int(10), "x"),
            ast.structFieldInitializer(int(20), "y")
        ],
        false, // Not positional
        "Point"
    )),
    printCall(ast.stringInterpolation([string("p1: "), ident("p1")])),
    // Field Access
    printCall(ast.stringInterpolation([string("p1.x: "), memberAccess("p1", "x")])),
    printCall(ast.stringInterpolation([string("p1.y: "), memberAccess("p1", "y")])),
    // Mutation
    assignMember("p1", "x", int(15)),
    printCall(ast.stringInterpolation([string("p1 after mutation: "), ident("p1")])),
    // Functional Update
    assign("p2", ast.structLiteral(
        [
            ast.structFieldInitializer(int(99), "y") // Override y
        ],
        false,
        "Point",
        ident("p1") // Source for update ...p1
    )),
    printCall(ast.stringInterpolation([string("p2 (functional update from p1): "), ident("p2")])),
    printCall(ast.stringInterpolation([string("p1 (should be unchanged): "), ident("p1")])), // Verify p1 wasn't mutated by functional update

    // --- Positional Fields ---
    printCall(string("--- Positional Fields Struct ---")),
    // Instantiation
    assign("c1", ast.structLiteral(
        [
            ast.structFieldInitializer(int(255)), // r
            ast.structFieldInitializer(int(0)),   // g
            ast.structFieldInitializer(int(128))  // b
        ],
        true, // Positional
        "Color"
    )),
    printCall(ast.stringInterpolation([string("c1: "), ident("c1")])),
    // Field Access (by index)
    printCall(ast.stringInterpolation([string("c1.0 (r): "), memberAccess("c1", 0)])),
    printCall(ast.stringInterpolation([string("c1.1 (g): "), memberAccess("c1", 1)])),
    printCall(ast.stringInterpolation([string("c1.2 (b): "), memberAccess("c1", 2)])),
    // Mutation (by index)
    assignMember("c1", 1, int(100)), // Change green component
    printCall(ast.stringInterpolation([string("c1 after mutation: "), ident("c1")])),

    ast.returnStatement() // Explicit return void
]);


// --- Main Function ---
const mainFunction = ast.functionDefinition(
    'main',
    [],
    ast.blockExpression([
        // Define structs first
        singletonDef,
        namedStructDef,
        positionalStructDef,
        // Then run test logic
        testLogic
    ]),
    voidType
);

// --- Module Definition ---
const structsModule = ast.module(
    [
        mainFunction
    ],
    [
        ast.importStatement(['io'], false, [ast.importSelector('print')]),
    ]
);

// Export the final AST module
export default structsModule;
