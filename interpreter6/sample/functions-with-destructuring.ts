import * as ast from "../ast";

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value);
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const structPattern = ast.structPattern;
const structField = ast.structPatternField;
const arrayPattern = ast.arrayPattern;
const printCall = (expr: ast.Expression) => ast.functionCall(ident("print"), [expr]);
const returnStmt = (expr: ast.Expression) => ast.returnStatement(expr);
const funcDef = (name: string, params: ast.FunctionParameter[], body: ast.BlockExpression, returnType: ast.TypeExpression) =>
  ast.functionDefinition(name, params, body, returnType);

const voidType = type("void");
const intType = type("i32");

// --- Test Cases ---

// 1. Function with Struct Destructuring
const structDestructuringFunc = funcDef(
  "processPoint",
  [ast.functionParameter(structPattern([structField(ident("x"), "x"), structField(ident("y"), "y")], false, "Point"), type("Point"))],
  ast.blockExpression([
    printCall(ast.stringInterpolation([string("x = "), ident("x")])),
    printCall(ast.stringInterpolation([string("y = "), ident("y")])),
    returnStmt(ast.binaryExpression("+", ident("x"), ident("y"))),
  ]),
  intType
);

// 2. Function with Array Destructuring
const arrayDestructuringFunc = funcDef(
  "sumFirstTwo",
  [
    ast.functionParameter(
      arrayPattern([ident("first"), ident("second")], ident("rest")), // Add rest pattern
      type("Array")
    ),
  ],
  ast.blockExpression([
    printCall(ast.stringInterpolation([string("first = "), ident("first")])),
    printCall(ast.stringInterpolation([string("second = "), ident("second")])),
    printCall(ast.stringInterpolation([string("rest = "), ident("rest")])), // Print the rest of the array
    returnStmt(ast.binaryExpression("+", ident("first"), ident("second"))),
  ]),
  intType
);

// 3. Struct Definition for Point
const pointStruct = ast.structDefinition("Point", [ast.structFieldDefinition(intType, "x"), ast.structFieldDefinition(intType, "y")], "named");

// --- Main Function ---
const mainFunction = funcDef(
  "main",
  [],
  ast.blockExpression([
    pointStruct, // Define the Point struct
    printCall(string("Test 1: Struct Destructuring")),
    ast.functionCall(ident("processPoint"), [ast.structLiteral([ast.structFieldInitializer(int(3), "x"), ast.structFieldInitializer(int(4), "y")], false, "Point")]),
    printCall(string("Test 2: Array Destructuring")),
    ast.functionCall(ident("sumFirstTwo"), [ast.arrayLiteral([int(5), int(6), int(7)])]),
    ast.returnStatement(), // Explicit return void
  ]),
  voidType
);

// --- Module Definition ---
const functionsWithDestructuringModule = ast.module(
  [mainFunction, structDestructuringFunc, arrayDestructuringFunc],
  [ast.importStatement(["io"], false, [ast.importSelector("print")])]
);

// Export the final AST module
export default functionsWithDestructuringModule;
