import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { collectTypeDispatches } from "./type-dispatch";
import { callCallableValue } from "./functions";
import {
  applyArithmeticBinary,
  applyBitwiseBinary,
  applyBitwiseNot,
  applyComparisonBinary,
  applyNumericUnaryMinus,
  isNumericValue,
  numericToNumber,
} from "./numeric";
import { makeIntegerFromNumber } from "./numeric";
import { valuesEqual } from "./value_equals";

export function resolveIndexFunction(
  ctx: InterpreterV10,
  receiver: V10Value,
  methodName: string,
  interfaceName: string,
): Extract<V10Value, { kind: "function" | "function_overload" }> | null {
  const dispatches = collectTypeDispatches(ctx, receiver);
  for (const dispatch of dispatches) {
    const method = ctx.findMethod(dispatch.typeName, methodName, {
      typeArgs: dispatch.typeArgs,
      interfaceName,
    });
    if (method) return method;
  }
  return null;
}

declare module "./index" {
  interface InterpreterV10 {
    computeBinaryForCompound(op: string, left: V10Value, right: V10Value): V10Value;
    ensureDivModStruct(): AST.StructDefinition;
    divModStruct?: AST.StructDefinition;
  }
}

export function evaluateUnaryExpression(ctx: InterpreterV10, node: AST.UnaryExpression, env: Environment): V10Value {
  const v = ctx.evaluate(node.operand, env);
  if (node.operator === "-") {
    return applyNumericUnaryMinus(v);
  }
  if (node.operator === "!") {
    if (v.kind === "bool") return { kind: "bool", value: !v.value };
    throw new Error("Unary '!' requires boolean operand");
  }
  if (node.operator === ".~") {
    return applyBitwiseNot(v);
  }
  throw new Error(`Unknown unary operator ${node.operator}`);
}

export function evaluateBinaryExpression(ctx: InterpreterV10, node: AST.BinaryExpression, env: Environment): V10Value {
  const b = node;
  if (b.operator === "&&" || b.operator === "||") {
    const lv = ctx.evaluate(b.left, env);
    if (lv.kind !== "bool") throw new Error("Logical operands must be bool");
    if (b.operator === "&&") {
      if (!lv.value) return { kind: "bool", value: false };
      const rv = ctx.evaluate(b.right, env);
      if (rv.kind !== "bool") throw new Error("Logical operands must be bool");
      return { kind: "bool", value: lv.value && rv.value };
    }
    if (lv.value) return { kind: "bool", value: true };
    const rv = ctx.evaluate(b.right, env);
    if (rv.kind !== "bool") throw new Error("Logical operands must be bool");
    return { kind: "bool", value: lv.value || rv.value };
  }

  if (b.operator === "|>" || b.operator === "|>>") {
    const subject = ctx.evaluate(b.left, env);
    ctx.topicStack.push(subject);
    ctx.topicUsageStack.push(false);
    ctx.implicitReceiverStack.push(subject);
    try {
      const rhsVal = ctx.evaluate(b.right, env);
      const topicUsed = ctx.topicUsageStack[ctx.topicUsageStack.length - 1] ?? false;
      if (topicUsed) {
        return rhsVal;
      }
      const callArgs = (rhsVal.kind === "bound_method" || rhsVal.kind === "native_bound_method") ? [] : [subject];
      try {
        return callCallableValue(ctx, rhsVal, callArgs, env);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        throw new Error(`pipe RHS must be callable when '%' is not used: ${message}`);
      }
    } finally {
      ctx.implicitReceiverStack.pop();
      ctx.topicUsageStack.pop();
      ctx.topicStack.pop();
    }
  }

  const left = ctx.evaluate(b.left, env);
  const right = ctx.evaluate(b.right, env);

  if (b.operator === "+" && left.kind === "string" && right.kind === "string") {
    return { kind: "string", value: left.value + right.value };
  }

  if (["+","-","*","/","//","%%","/%"].includes(b.operator)) {
    return applyArithmeticBinary(b.operator, left, right, {
      makeDivMod: (kind, parts) => {
        const structDef = ctx.ensureDivModStruct();
        const typeArg = AST.simpleTypeExpression(kind);
        const typeArgMap = ctx.mapTypeArguments(structDef.genericParams ?? [], [typeArg], "DivMod");
        return {
          kind: "struct_instance",
          def: structDef,
          values: new Map([
            ["quotient", parts.quotient],
            ["remainder", parts.remainder],
          ]),
          typeArguments: [typeArg],
          typeArgMap,
        };
      },
    });
  }

  if ([">","<",">=","<=","==","!="].includes(b.operator)) {
    if (isNumericValue(left) && isNumericValue(right)) {
      return applyComparisonBinary(b.operator, left, right);
    }
    if (left.kind === "string" && right.kind === "string") {
      switch (b.operator) {
        case ">": return { kind: "bool", value: left.value > right.value };
        case "<": return { kind: "bool", value: left.value < right.value };
        case ">=": return { kind: "bool", value: left.value >= right.value };
        case "<=": return { kind: "bool", value: left.value <= right.value };
        case "==": return { kind: "bool", value: left.value === right.value };
        case "!=": return { kind: "bool", value: left.value !== right.value };
      }
    }
    if (b.operator === "==") return { kind: "bool", value: valuesEqual(left, right) };
    if (b.operator === "!=") return { kind: "bool", value: !valuesEqual(left, right) };
    throw new Error("Unsupported comparison operands");
  }

  if ([".&",".|",".^",".<<",".>>"].includes(b.operator)) {
    return applyBitwiseBinary(b.operator, left, right);
  }

  throw new Error(`Unknown binary operator ${b.operator}`);
}

export function evaluateRangeExpression(ctx: InterpreterV10, node: AST.RangeExpression, env: Environment): V10Value {
  const start = ctx.evaluate(node.start, env);
  const end = ctx.evaluate(node.end, env);
  const viaInterface = ctx.tryInvokeRangeImplementation(start, end, node.inclusive, env);
  if (viaInterface) {
    return viaInterface;
  }
  let startNum: number;
  let endNum: number;
  try {
    startNum = numericToNumber(start, "Range start");
    endNum = numericToNumber(end, "Range end");
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    if (message.includes("Range start must be numeric") || message.includes("Range end must be numeric")) {
      throw new Error("Range boundaries must be numeric");
    }
    throw err;
  }
  if (!Number.isFinite(startNum) || !Number.isFinite(endNum)) {
    throw new Error("Range endpoint must be finite");
  }
  const startInt = Math.trunc(startNum);
  const endInt = Math.trunc(endNum);
  const step = startInt <= endInt ? 1 : -1;
  const elements: V10Value[] = [];
  for (let current = startInt; ; current += step) {
    if (step > 0) {
      if (node.inclusive) {
        if (current > endInt) break;
      } else if (current >= endInt) {
        break;
      }
    } else if (node.inclusive) {
      if (current < endInt) break;
    } else if (current <= endInt) {
      break;
    }
    elements.push(makeIntegerFromNumber("i32", current));
  }
  return ctx.makeArrayValue(elements);
}

export function evaluateIndexExpression(ctx: InterpreterV10, node: AST.IndexExpression, env: Environment): V10Value {
  const obj = ctx.evaluate(node.object, env);
  const idxVal = ctx.evaluate(node.index, env);
  const viaInterface = resolveIndexFunction(ctx, obj, "get", "Index");
  if (viaInterface) {
    return callCallableValue(ctx, viaInterface, [obj, idxVal], env);
  }
  if (obj.kind !== "array") throw new Error("Indexing is only supported on arrays in this milestone");
  const state = ctx.ensureArrayState(obj);
  const idx = Math.trunc(numericToNumber(idxVal, "Array index", { requireSafeInteger: true }));
  if (idx < 0 || idx >= state.values.length) throw new Error("Array index out of bounds");
  const el = state.values[idx];
  if (el === undefined) throw new Error("Internal error: array element undefined");
  return el;
}

export function evaluateTopicReferenceExpression(ctx: InterpreterV10): V10Value {
  if (ctx.topicStack.length === 0 || ctx.topicUsageStack.length === 0) {
    throw new Error("Topic reference '%' used outside of pipe expression");
  }
  ctx.topicUsageStack[ctx.topicUsageStack.length - 1] = true;
  const current = ctx.topicStack[ctx.topicStack.length - 1];
  return current;
}

export function applyOperationsAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.ensureDivModStruct = function ensureDivModStruct(this: InterpreterV10): AST.StructDefinition {
    if (this.divModStruct) return this.divModStruct;
    const divModDef = AST.structDefinition(
      "DivMod",
      [
        AST.structFieldDefinition(AST.simpleTypeExpression("T"), "quotient"),
        AST.structFieldDefinition(AST.simpleTypeExpression("T"), "remainder"),
      ],
      "named",
      [AST.genericParameter("T")],
    );
    this.evaluate(divModDef, this.globals);
    const resolved = this.globals.get("DivMod");
    if (resolved && resolved.kind === "struct_def") {
      this.divModStruct = resolved.def;
    } else {
      this.divModStruct = divModDef;
    }
    return this.divModStruct;
  };

  cls.prototype.computeBinaryForCompound = function computeBinaryForCompound(this: InterpreterV10, op: string, left: V10Value, right: V10Value): V10Value {
    if (["+","-","*","/"].includes(op)) {
      return applyArithmeticBinary(op, left, right);
    }

    if ([".&",".|",".^",".<<",".>>"].includes(op)) {
      return applyBitwiseBinary(op, left, right);
    }

    throw new Error(`Unsupported compound operator ${op}`);
  };
}
