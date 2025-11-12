import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

export function evaluateLiteral(ctx: InterpreterV10, node: AST.AstNode, env: Environment): V10Value {
  switch (node.type) {
    case "StringLiteral":
      return { kind: "string", value: (node as AST.StringLiteral).value };
    case "BooleanLiteral":
      return { kind: "bool", value: (node as AST.BooleanLiteral).value };
    case "CharLiteral":
      return { kind: "char", value: (node as AST.CharLiteral).value };
    case "NilLiteral":
      return { kind: "nil", value: null };
    case "FloatLiteral": {
      const n = (node as AST.FloatLiteral).value;
      return { kind: "f64", value: n };
    }
    case "IntegerLiteral": {
      const intNode = node as AST.IntegerLiteral;
      const kind = intNode.integerType ?? "i32";
      if (["i64", "i128", "u64", "u128"].includes(kind)) {
        return { kind: "i32", value: Number(intNode.value) };
      }
      return { kind: "i32", value: Number(intNode.value) };
    }
    case "ArrayLiteral": {
      const arrNode = node as AST.ArrayLiteral;
      const elements = arrNode.elements.map(element => ctx.evaluate(element, env));
      return { kind: "array", elements };
    }
    default:
      throw new Error(`Unsupported literal node: ${node.type}`);
  }
}
