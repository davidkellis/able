import * as AST from "../ast";

const functionsSampleModule: AST.Module = {
  type: "Module",
  package: {
    type: "PackageStatement",
    namePath: [{ type: "Identifier", name: "sample_functions" }],
  },
  imports: [],
  body: [
    // --- Function Definition ---
    AST.functionDefinition(
      "add",
      [
        AST.functionParameter("x", AST.simpleTypeExpression("i32")),
        AST.functionParameter("y", AST.simpleTypeExpression("i32")),
      ],
      AST.blockExpression([
        AST.returnStatement(
          AST.binaryExpression(
            "+",
            AST.identifier("x"),
            AST.identifier("y")
          )
        ),
      ]),
      AST.simpleTypeExpression("i32")
    ),

    // --- Function Invocation ---
    AST.functionCall(
      AST.identifier("add"),
      [AST.integerLiteral(5), AST.integerLiteral(3)]
    ),
  ],
};

export default functionsSampleModule;
