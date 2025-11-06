import * as AST from "./ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./interpreter";
import type { AblePrimitive, AbleValue } from "./runtime";
import { BreakSignal, RaiseSignal, ReturnSignal } from "./signals";

export function evaluateNode(this: Interpreter, node: AST.AstNode | null, environment: Environment): AbleValue {
  const self = this as any;
  try {
    if (!node) return { kind: "nil", value: null };

    switch (node.type) {
      case "StringLiteral":
        return { kind: "string", value: (node as AST.StringLiteral).value };
      case "IntegerLiteral": {
        const typedNode = node as AST.IntegerLiteral;
        const intType = typedNode.integerType || "i32";
        if (["i64", "i128", "u64", "u128"].includes(intType)) {
          return { kind: intType as any, value: BigInt(typedNode.value.toString()) };
        } else {
          return { kind: intType as any, value: Number(typedNode.value) };
        }
      }
      case "FloatLiteral":
        return { kind: (node as AST.FloatLiteral).floatType || "f64", value: (node as AST.FloatLiteral).value };
      case "BooleanLiteral":
        return { kind: "bool", value: (node as AST.BooleanLiteral).value };
      case "NilLiteral":
        return { kind: "nil", value: null };
      case "CharLiteral":
        return { kind: "char", value: (node as AST.CharLiteral).value };
      case "ArrayLiteral": {
        const typedNode = node as AST.ArrayLiteral;
        const elements = typedNode.elements.map((el) => self.evaluate(el, environment));
        return { kind: "array", elements };
      }
      case "Identifier":
        return environment.get((node as AST.Identifier).name);
      case "BlockExpression":
        return self.evaluateBlockExpression(node as AST.BlockExpression, environment);
      case "UnaryExpression":
        return self.evaluateUnaryExpression(node as AST.UnaryExpression, environment);
      case "BinaryExpression":
        return self.evaluateBinaryExpression(node as AST.BinaryExpression, environment);
      case "AssignmentExpression":
        return self.evaluateAssignmentExpression(node as AST.AssignmentExpression, environment);
      case "FunctionCall":
        return self.evaluateFunctionCall(node as AST.FunctionCall, environment);
      case "IfExpression":
        return self.evaluateIfExpression(node as AST.IfExpression, environment);
      case "StructLiteral":
        return self.evaluateStructLiteral(node as AST.StructLiteral, environment);
      case "MemberAccessExpression":
        return self.evaluateMemberAccess(node as AST.MemberAccessExpression, environment);
      case "StringInterpolation":
        return self.evaluateStringInterpolation(node as AST.StringInterpolation, environment);
      case "LambdaExpression":
        return self.evaluateLambdaExpression(node as AST.LambdaExpression, environment);
      case "RangeExpression":
        return self.evaluateRangeExpression(node as AST.RangeExpression, environment);
      case "MatchExpression":
        return self.evaluateMatchExpression(node as AST.MatchExpression, environment);
      case "ProcExpression":
      case "SpawnExpression":
      case "BreakpointExpression":
      case "PropagationExpression":
      case "OrElseExpression":
        return { kind: "nil", value: null };
      case "RescueExpression":
        return self.evaluateRescueExpression(node as AST.RescueExpression, environment);
      case "FunctionDefinition":
        self.evaluateFunctionDefinition(node as AST.FunctionDefinition, environment);
        return { kind: "nil", value: null };
      case "StructDefinition":
        self.evaluateStructDefinition(node as AST.StructDefinition, environment);
        return { kind: "nil", value: null };
      case "UnionDefinition":
        self.evaluateUnionDefinition(node as AST.UnionDefinition, environment);
        return { kind: "nil", value: null };
      case "InterfaceDefinition":
        self.evaluateInterfaceDefinition(node as AST.InterfaceDefinition, environment);
        return { kind: "nil", value: null };
      case "ImplementationDefinition":
        self.evaluateImplementationDefinition(node as AST.ImplementationDefinition, environment);
        return { kind: "nil", value: null };
      case "MethodsDefinition":
        self.evaluateMethodsDefinition(node as AST.MethodsDefinition, environment);
        return { kind: "nil", value: null };
      case "ReturnStatement": {
        const typedNode = node as AST.ReturnStatement;
        const returnValue = typedNode.argument
          ? self.evaluate(typedNode.argument, environment)
          : ({ kind: "void", value: undefined } as AblePrimitive);
        throw new ReturnSignal(returnValue);
      }
      case "RaiseStatement":
        self.evaluateRaiseStatement(node as AST.RaiseStatement, environment);
        return { kind: "nil", value: null };
      case "BreakStatement":
        return self.evaluateBreakStatement(node as AST.BreakStatement, environment);
      case "WhileLoop":
        return self.evaluateWhileLoop(node as AST.WhileLoop, environment);
      case "ForLoop":
        return self.evaluateForLoop(node as AST.ForLoop, environment);
      case "PackageStatement":
      case "ImportStatement":
      case "ImportSelector":
        return { kind: "nil", value: null };
      default:
        return { kind: "nil", value: null };
    }
  } catch (e) {
    if (e instanceof Error && !(e instanceof ReturnSignal || e instanceof RaiseSignal || e instanceof BreakSignal)) {
      console.error(`Error during evaluation of ${node?.type} node:`, node);
    }
    throw e;
  }
}
