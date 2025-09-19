// --- main.ts ---

import { Lexer } from './lexer';
import { Parser } from './parser';
import * as Ast from './ast'; // Assuming you might add an AST printer visitor

const sourceCode = `
fn greet(name: String) -> String {
  greeting = "Hello, " ## Assignment statement
  return greeting + name + "!" ## Example return
}

main = fn() { ## Assign lambda
   message = greet("Able")
   ## print(message) ## print not parsed yet
}
`;

// 1. Lexing
const lexer = new Lexer(sourceCode);
const tokens = lexer.scanTokens();
console.log("Tokens:", tokens);

// 2. Parsing
const parser = new Parser(tokens);
const syntaxTree = parser.parse(); // Returns array of Stmt
console.log("AST:", JSON.stringify(syntaxTree, null, 2)); // Basic AST view

// 3. Next steps would be semantic analysis, type checking, code generation...
