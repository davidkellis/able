import * as ast from "../ast";

// --- Helpers ---
const type = (name: string) => ast.simpleTypeExpression(name);
const int = (value: number) => ast.integerLiteral(value);
const string = (value: string) => ast.stringLiteral(value);
const ident = (name: string) => ast.identifier(name);
const assign = (target: string, value: ast.Expression, op: ":=" | "=" = ":=") =>
  ast.assignmentExpression(op, ident(target), value);
const printCall = (expr: ast.Expression) => ast.functionCall(ident("print"), [expr]);
const returnStmt = (expr: ast.Expression) => ast.returnStatement(expr);
const funcDef = (
  name: string,
  params: ast.FunctionParameter[],
  body: ast.BlockExpression,
  returnType: ast.TypeExpression
) => ast.functionDefinition(name, params, body, returnType);

const voidType = type("void");
const intType = type("i32");

// --- Test Cases ---

// 1. Top-Level Assignment
const topLevelAssignment = assign("globalVar", int(100), ":="); // Ensure globalVar is defined in the global environment

// 2. Function with Assignment Expressions
const functionWithAssignments = funcDef(
  "modifyGlobalVar",
  [],
  ast.blockExpression([
    printCall(string("Inside modifyGlobalVar:")),
    assign("globalVar", ast.binaryExpression("+", ident("globalVar"), int(50)), "="),
    printCall(ast.stringInterpolation([string("  globalVar = "), ident("globalVar")])),
    returnStmt(ident("globalVar")),
  ]),
  intType
);

// 3. Closure Referencing Package-Level Variable
const closureTest = funcDef(
  "closureExample",
  [],
  ast.blockExpression([
    printCall(string("Inside closureExample:")),
    ast.assignmentExpression(
      ":=",
      ident("closureFunc"),
      ast.lambdaExpression(
        [],
        ast.blockExpression([
          printCall(string("  Inside closureFunc:")),
          printCall(ast.stringInterpolation([string("    globalVar = "), ident("globalVar")])),
          returnStmt(ast.binaryExpression("*", ident("globalVar"), int(2))),
        ]),
        intType
      )
    ),
    printCall(ast.stringInterpolation([string("  closureFunc() = "), ast.functionCall(ident("closureFunc"), [])])),
    returnStmt(ident("closureFunc")),
  ]),
  type("function")
);

// --- Main Function ---
const mainFunction = funcDef(
  "main",
  [],
  ast.blockExpression([
    printCall(ast.stringInterpolation([string("Initial globalVar = "), ident("globalVar")])),
    ast.functionCall(ident("modifyGlobalVar"), []),
    ast.functionCall(ident("closureExample"), []),
    ast.returnStatement(), // Explicit return void
  ]),
  voidType
);

// --- Module Definition ---
const assignmentsAndClosuresModule = ast.module(
  [topLevelAssignment, mainFunction, functionWithAssignments, closureTest],
  [ast.importStatement(["io"], false, [ast.importSelector("print")])]
);

// Export the final AST module
export default assignmentsAndClosuresModule;
