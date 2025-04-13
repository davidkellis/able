import * as ast from './ast';

// Helper to create simple type expressions
const type = (name: string) => ast.simpleTypeExpression(name);
const string = type('string');
const voidType = type('void');

// --- Main Program Logic ---

const mainFunction = ast.functionDefinition(
    'main',
    [], // No parameters for main
    ast.blockExpression([
        // Call the 'print' function with "hello world"
        ast.functionCall(
            ast.identifier('print'), // Assuming 'print' is available (imported)
            [ast.stringLiteral("hello world")]
        ),
        ast.returnStatement() // Explicit return void
    ]),
    voidType // Main returns void
);

// --- Module Definition ---
// Combine all definitions into a module
const sampleModule = ast.module(
    [
        // Main execution logic
        mainFunction
    ],
    [
        // Import 'print' from 'io' (adjust if print is located elsewhere, e.g., 'core')
        ast.importStatement(['io'], false, [ast.importSelector('print')]),
    ]
    // Optional package statement: ast.packageStatement(['sample2'])
);

// Export the final AST module
export default sampleModule;

// To use this file, you would typically import it and pass `sampleModule`
// to your interpreter or compiler.
// Example (in another file):
// import helloWorldAst from './sample2';
// interpret(helloWorldAst);
