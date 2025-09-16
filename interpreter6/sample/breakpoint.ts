import * as ast from '../ast';

// --- Helpers from other samples (assuming they exist in ast) ---
// If these helpers don't exist in ../ast, they would need to be added there
// or defined locally in this file.
const ident = (name: string) => ast.identifier(name);
const int = (value: number) => ast.integerLiteral(value);

// --- Module Definition ---
const breakpointModule = ast.module(
    [
        // print("Before breakpoint");
        ast.functionCall(
            ident('print'),
            [ast.stringLiteral("Before breakpoint")],
            [], // Type args
            false // isTrailingLambda
        ),
        // let x := 10;
        ast.assignmentExpression(
            ':=',
            ident('x'),
            int(10)
        ),
        // breakpoint;
        // Note: The AST builder might not require label/body if they are optional
        // or if the builder provides defaults. Adjust if ast.breakpointExpression
        // has a different signature.
        ast.breakpointExpression(
            // Provide the required label/body arguments
            ident('_breakpoint'),
            ast.blockExpression([])
        ),
        // print(x);
        ast.functionCall(
            ident('print'),
            [ident('x')],
            [], // Add empty array for type arguments
            false // isTrailingLambda
        ),
    ],
    [
        // Import 'print'
        // ast.importStatement(['io'], false, [ast.importSelector('print')]),
        // Relying on global print for simplicity in this test
    ]
);

export default breakpointModule;
