import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { evaluateAssignmentExpression } from "./assignments";
import { evaluateBlockExpression, evaluateBreakStatement, evaluateContinueStatement, evaluateForLoop, evaluateIfExpression, evaluateReturnStatement, evaluateWhileLoop } from "./control_flow";
import { evaluateEnsureExpression, evaluateOrElseExpression, evaluatePropagationExpression, evaluateRescueExpression, evaluateRaiseStatement, evaluateRethrowStatement } from "./error_handling";
import { evaluateFunctionCall, evaluateFunctionDefinition, evaluateLambdaExpression } from "./functions";
import { evaluateDynImportStatement, evaluateImportStatement, evaluateModule, evaluatePackageStatement } from "./imports";
import { evaluateImplementationDefinition, evaluateInterfaceDefinition, evaluateMethodsDefinition, evaluateUnionDefinition } from "./definitions";
import { evaluateMatchExpression } from "./match";
import { evaluateMemberAccessExpression, evaluateStructDefinition, evaluateStructLiteral } from "./structs";
import { evaluateLiteral } from "./literals";
import { evaluateIndexExpression, evaluateRangeExpression, evaluateBinaryExpression, evaluateUnaryExpression } from "./operations";
import { evaluateProcExpression, evaluateSpawnExpression, evaluateBreakpointExpression, evaluateStringInterpolation } from "./runtime_extras";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    evaluate(node: AST.AstNode | null, env?: Environment): V10Value;
  }
}

const NIL: V10Value = { kind: "nil", value: null };

export function applyEvaluationAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.evaluate = function evaluate(this: InterpreterV10, node: AST.AstNode | null, env: Environment = this.globals): V10Value {
    if (!node) return NIL;
    switch (node.type) {
      case "StringLiteral":
      case "BooleanLiteral":
      case "CharLiteral":
      case "NilLiteral":
      case "FloatLiteral":
      case "IntegerLiteral":
      case "ArrayLiteral":
        return evaluateLiteral(this, node, env);
      case "UnaryExpression":
        return evaluateUnaryExpression(this, node as AST.UnaryExpression, env);
      case "BinaryExpression":
        return evaluateBinaryExpression(this, node as AST.BinaryExpression, env);
      case "RangeExpression":
        return evaluateRangeExpression(this, node as AST.RangeExpression, env);
      case "IndexExpression":
        return evaluateIndexExpression(this, node as AST.IndexExpression, env);
      case "IfExpression":
        return evaluateIfExpression(this, node as AST.IfExpression, env);
      case "WhileLoop":
        return evaluateWhileLoop(this, node as AST.WhileLoop, env);
      case "BreakStatement":
        evaluateBreakStatement(this, node as AST.BreakStatement, env);
        return NIL;
      case "ContinueStatement":
        evaluateContinueStatement(this, node as AST.ContinueStatement);
        return NIL;
      case "ForLoop":
        return evaluateForLoop(this, node as AST.ForLoop, env);
      case "MatchExpression":
        return evaluateMatchExpression(this, node as AST.MatchExpression, env);
      case "Identifier":
        return env.get((node as AST.Identifier).name);
      case "BlockExpression":
        return evaluateBlockExpression(this, node as AST.BlockExpression, env);
      case "ReturnStatement":
        evaluateReturnStatement(this, node as AST.ReturnStatement, env);
        return NIL;
      case "AssignmentExpression":
        return evaluateAssignmentExpression(this, node as AST.AssignmentExpression, env);
      case "FunctionDefinition":
        return evaluateFunctionDefinition(this, node as AST.FunctionDefinition, env);
      case "LambdaExpression":
        return evaluateLambdaExpression(this, node as AST.LambdaExpression, env);
      case "FunctionCall":
        return evaluateFunctionCall(this, node as AST.FunctionCall, env);
      case "StructDefinition":
        return evaluateStructDefinition(this, node as AST.StructDefinition, env);
      case "StructLiteral":
        return evaluateStructLiteral(this, node as AST.StructLiteral, env);
      case "MemberAccessExpression":
        return evaluateMemberAccessExpression(this, node as AST.MemberAccessExpression, env);
      case "StringInterpolation":
        return evaluateStringInterpolation(this, node as AST.StringInterpolation, env);
      case "BreakpointExpression":
        return evaluateBreakpointExpression(this, node as AST.BreakpointExpression, env);
      case "RaiseStatement":
        evaluateRaiseStatement(this, node as AST.RaiseStatement, env);
        return NIL;
      case "RescueExpression":
        return evaluateRescueExpression(this, node as AST.RescueExpression, env);
      case "OrElseExpression":
        return evaluateOrElseExpression(this, node as AST.OrElseExpression, env);
      case "PropagationExpression":
        return evaluatePropagationExpression(this, node as AST.PropagationExpression, env);
      case "EnsureExpression":
        return evaluateEnsureExpression(this, node as AST.EnsureExpression, env);
      case "RethrowStatement":
        evaluateRethrowStatement(this, node as AST.RethrowStatement);
        return NIL;
      case "ProcExpression":
        return evaluateProcExpression(this, node as AST.ProcExpression, env);
      case "SpawnExpression":
        return evaluateSpawnExpression(this, node as AST.SpawnExpression, env);
      case "Module":
        return evaluateModule(this, node as AST.Module);
      case "PackageStatement":
        return evaluatePackageStatement();
      case "ImportStatement":
        return evaluateImportStatement(this, node as AST.ImportStatement, env);
      case "DynImportStatement":
        return evaluateDynImportStatement(this, node as AST.DynImportStatement, env);
      case "InterfaceDefinition":
        return evaluateInterfaceDefinition(this, node as AST.InterfaceDefinition, env);
      case "UnionDefinition":
        return evaluateUnionDefinition(this, node as AST.UnionDefinition, env);
      case "MethodsDefinition":
        return evaluateMethodsDefinition(this, node as AST.MethodsDefinition, env);
      case "ImplementationDefinition":
        return evaluateImplementationDefinition(this, node as AST.ImplementationDefinition, env);
      default:
        throw new Error(`Not implemented in milestone: ${node.type}`);
    }
  };
}
