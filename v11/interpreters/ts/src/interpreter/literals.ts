import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { FloatKind, IntegerKind, RuntimeValue } from "./values";
import { makeFloatValue, makeIntegerValue } from "./numeric";

export function evaluateLiteral(ctx: Interpreter, node: AST.AstNode, env: Environment): RuntimeValue {
  switch (node.type) {
    case "StringLiteral":
      return { kind: "String", value: (node as AST.StringLiteral).value };
    case "BooleanLiteral":
      return { kind: "bool", value: (node as AST.BooleanLiteral).value };
    case "CharLiteral":
      return { kind: "char", value: (node as AST.CharLiteral).value };
    case "NilLiteral":
      return { kind: "nil", value: null };
    case "FloatLiteral": {
      const floatNode = node as AST.FloatLiteral;
      const floatKind = (floatNode.floatType ?? "f64") as FloatKind;
      return makeFloatValue(floatKind, floatNode.value);
    }
    case "IntegerLiteral": {
      const intNode = node as AST.IntegerLiteral;
      const kind = (intNode.integerType ?? "i32") as IntegerKind;
      const raw = typeof intNode.value === "bigint" ? intNode.value : BigInt(Math.trunc(intNode.value ?? 0));
      return makeIntegerValue(kind, raw);
    }
    case "ArrayLiteral": {
      const arrNode = node as AST.ArrayLiteral;
      const elements = arrNode.elements.map(element => ctx.evaluate(element, env));
      return ctx.makeArrayValue(elements);
    }
    default:
      throw new Error(`Unsupported literal node: ${node.type}`);
  }
}
