import * as AST from "../ast";

// --- Helpers ---
const type = (name: string) => AST.simpleTypeExpression(name);
const int = (value: number) => AST.integerLiteral(value);
const ident = (name: string) => AST.identifier(name);
const binaryOp = (op: string, left: AST.Expression, right: AST.Expression) =>
  AST.binaryExpression(op, left, right);
const returnStmt = (expr: AST.Expression) => AST.returnStatement(expr);
const funcDef = (
  name: string,
  params: AST.FunctionParameter[],
  body: AST.BlockExpression,
  returnType: AST.TypeExpression
) => AST.functionDefinition(name, params, body, returnType);
const funcCall = (name: string, args: AST.Expression[]) =>
  AST.functionCall(ident(name), args);

const voidType = type("void");
const intType = type("i32");

// --- Test Cases ---

// 1. Simple Function Definition and Call
const test1 = AST.blockExpression([
  AST.functionDefinition(
    "add",
    [
      AST.functionParameter("x", intType),
      AST.functionParameter("y", intType),
    ],
    AST.blockExpression([returnStmt(binaryOp("+", ident("x"), ident("y")))]),
    intType
  ),
  funcCall("print", [
    AST.stringInterpolation([
      AST.stringLiteral("Test 1: add(2, 3) = "),
      funcCall("add", [int(2), int(3)]),
    ]),
  ]),
]);

// 2. Nested Functions
const test2 = AST.blockExpression([
  AST.functionDefinition(
    "outer",
    [],
    AST.blockExpression([
      AST.functionDefinition(
        "inner",
        [AST.functionParameter("z", intType)],
        AST.blockExpression([
          returnStmt(binaryOp("*", ident("z"), int(2))),
        ]),
        intType
      ),
      funcCall("print", [
        AST.stringInterpolation([
          AST.stringLiteral("Test 2: inner(5) = "),
          funcCall("inner", [int(5)]),
        ]),
      ]),
    ]),
    voidType
  ),
  funcCall("outer", []),
]);

// 3. Recursive Function
const test3 = AST.blockExpression([
  AST.functionDefinition(
    "factorial",
    [AST.functionParameter("n", intType)],
    AST.blockExpression([
      AST.ifExpression(
        binaryOp("<=", ident("n"), int(1)),
        AST.blockExpression([returnStmt(int(1))]),
        [
          AST.orClause(
            AST.blockExpression([
              returnStmt(
                binaryOp(
                  "*",
                  ident("n"),
                  funcCall("factorial", [
                    binaryOp("-", ident("n"), int(1)),
                  ])
                )
              ),
            ])
          ),
        ]
      ),
    ]),
    intType
  ),
  funcCall("print", [
    AST.stringInterpolation([
      AST.stringLiteral("Test 3: factorial(5) = "),
      funcCall("factorial", [int(5)]),
    ]),
  ]),
]);

// 4. Function with Multiple Arguments
const test4 = AST.blockExpression([
  AST.functionDefinition(
    "sumThree",
    [
      AST.functionParameter("a", intType),
      AST.functionParameter("b", intType),
      AST.functionParameter("c", intType),
    ],
    AST.blockExpression([
      returnStmt(binaryOp("+", binaryOp("+", ident("a"), ident("b")), ident("c"))),
    ]),
    intType
  ),
  funcCall("print", [
    AST.stringInterpolation([
      AST.stringLiteral("Test 4: sumThree(1, 2, 3) = "),
      funcCall("sumThree", [int(1), int(2), int(3)]),
    ]),
  ]),
]);

// --- Main Function ---
const mainFunction = funcDef(
  "main",
  [],
  AST.blockExpression([
    test1,
    test2,
    test3,
    test4,
    AST.returnStatement(), // Explicit return void
  ]),
  voidType
);

// --- Module Definition ---
const functionsSampleModule: AST.Module = AST.module(
  [mainFunction],
  [AST.importStatement(["io"], false, [AST.importSelector("print")])]
);

export default functionsSampleModule;
