import * as ast from "../ast";

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value);
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const assign = (target: string, value: ast.Expression, op: ":=" | "=" = ":=") =>
  ast.assignmentExpression(op, ident(target), value);
const structPattern = ast.structPattern;
const structField = ast.structPatternField;
const arrayPattern = ast.arrayPattern;
const printCall = (expr: ast.Expression) => ast.functionCall(ident("print"), [expr]);

const voidType = type("void");

// Define the Point struct
const pointStruct = ast.structDefinition(
  "Point",
  [
    ast.structFieldDefinition(type("i32"), "x"),
    ast.structFieldDefinition(type("i32"), "y"),
  ],
  "named"
);

// --- Test Cases ---

// 1. Simple Assignment
const test1 = ast.blockExpression([
  printCall(string("Test 1: Simple Assignment")),
  assign("x", int(42)),
  printCall(ast.stringInterpolation([string("  x = "), ident("x")])),
]);

// 2. Reassignment
const test2 = ast.blockExpression([
  printCall(string("Test 2: Reassignment")),
  assign("y", int(10)),
  assign("y", int(20), "="),
  printCall(ast.stringInterpolation([string("  y = "), ident("y")])),
]);

// 3. Destructuring Assignment (Struct)
const test3 = ast.blockExpression([
  printCall(string("Test 3: Destructuring Assignment (Struct)")),
  assign(
    "point",
    ast.structLiteral(
      [
        ast.structFieldInitializer(int(1), "x"),
        ast.structFieldInitializer(int(2), "y"),
      ],
      false,
      "Point"
    )
  ),
  ast.assignmentExpression(
    ":=",
    structPattern(
      [structField(ident("a"), "x"), structField(ident("b"), "y")],
      false,
      "Point"
    ),
    ident("point")
  ),
  printCall(ast.stringInterpolation([string("  a = "), ident("a")])),
  printCall(ast.stringInterpolation([string("  b = "), ident("b")])),
]);

// 4. Destructuring Assignment (Array)
const test4 = ast.blockExpression([
  printCall(string("Test 4: Destructuring Assignment (Array)")),
  assign("arr", ast.arrayLiteral([int(3), int(4), int(5)])),
  ast.assignmentExpression(
    ":=",
    arrayPattern([ident("first"), ident("second")], ident("rest")),
    ident("arr")
  ),
  printCall(ast.stringInterpolation([string("  first = "), ident("first")])),
  printCall(ast.stringInterpolation([string("  second = "), ident("second")])),
  printCall(ast.stringInterpolation([string("  rest = "), ident("rest")])),
]);

// --- Main Function ---
const mainFunction = ast.functionDefinition(
  "main",
  [],
  ast.blockExpression([
    pointStruct, // Add the struct definition to the main function
    test1,
    test2,
    test3,
    test4,
    ast.returnStatement(), // Explicit return void
  ]),
  voidType
);

// --- Module Definition ---
const assignmentsModule = ast.module(
  [mainFunction],
  [ast.importStatement(["io"], false, [ast.importSelector("print")])]
);

// Export the final AST module
export default assignmentsModule;
