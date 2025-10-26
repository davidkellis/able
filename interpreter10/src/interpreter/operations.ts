import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { callCallableValue } from "./functions";

declare module "./index" {
  interface InterpreterV10 {
    computeBinaryForCompound(op: string, left: V10Value, right: V10Value): V10Value;
  }
}

export function evaluateUnaryExpression(ctx: InterpreterV10, node: AST.UnaryExpression, env: Environment): V10Value {
  const v = ctx.evaluate(node.operand, env);
  if (node.operator === "-") {
    if (v.kind === "i32") return { kind: "i32", value: -v.value };
    if (v.kind === "f64") return { kind: "f64", value: -v.value };
    throw new Error("Unary '-' requires numeric operand");
  }
  if (node.operator === "!") {
    if (v.kind === "bool") return { kind: "bool", value: !v.value };
    throw new Error("Unary '!' requires boolean operand");
  }
  if (node.operator === "~") {
    if (v.kind === "i32") return { kind: "i32", value: ~v.value };
    throw new Error("Unary '~' requires i32 operand");
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

  if (b.operator === "|>") {
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

  const isNum = (v: V10Value) => v.kind === "i32" || v.kind === "f64";
  const asNumber = (v: V10Value): number => (v.kind === "i32" || v.kind === "f64") ? v.value : NaN;
  const resultKind = (a: V10Value, c: V10Value): "i32" | "f64" => (a.kind === "f64" || c.kind === "f64") ? "f64" : "i32";

  if (["+","-","*","/","%"].includes(b.operator)) {
    if (!isNum(left) || !isNum(right)) throw new Error("Arithmetic requires numeric operands");
    const kind = resultKind(left, right);
    const l = asNumber(left);
    const r = asNumber(right);
    switch (b.operator) {
      case "+": return kind === "i32" ? { kind, value: (l + r) | 0 } : { kind, value: l + r };
      case "-": return kind === "i32" ? { kind, value: (l - r) | 0 } : { kind, value: l - r };
      case "*": return kind === "i32" ? { kind, value: (l * r) | 0 } : { kind, value: l * r };
      case "/": {
        if (r === 0) throw new Error("Division by zero");
        return kind === "i32" ? { kind, value: (l / r) | 0 } : { kind, value: l / r };
      }
      case "%": {
        if (r === 0) throw new Error("Division by zero");
        return { kind, value: kind === "i32" ? (l % r) | 0 : l % r };
      }
    }
  }

  if ([">","<",">=","<=","==","!="].includes(b.operator)) {
    if (isNum(left) && isNum(right)) {
      const l = asNumber(left);
      const r = asNumber(right);
      switch (b.operator) {
        case ">": return { kind: "bool", value: l > r };
        case "<": return { kind: "bool", value: l < r };
        case ">=": return { kind: "bool", value: l >= r };
        case "<=": return { kind: "bool", value: l <= r };
        case "==": return { kind: "bool", value: l === r };
        case "!=": return { kind: "bool", value: l !== r };
      }
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
    if (b.operator === "==") return { kind: "bool", value: JSON.stringify(left) === JSON.stringify(right) };
    if (b.operator === "!=") return { kind: "bool", value: JSON.stringify(left) !== JSON.stringify(right) };
    throw new Error("Unsupported comparison operands");
  }

  if (["&","|","^","<<",">>"] .includes(b.operator)) {
    if (left.kind !== "i32" || right.kind !== "i32") throw new Error("Bitwise requires i32 operands");
    switch (b.operator) {
      case "&": return { kind: "i32", value: left.value & right.value };
      case "|": return { kind: "i32", value: left.value | right.value };
      case "^": return { kind: "i32", value: left.value ^ right.value };
      case "<<": {
        const count = right.value;
        if (count < 0 || count >= 32) throw new Error("shift out of range");
        return { kind: "i32", value: left.value << count };
      }
      case ">>": {
        const count = right.value;
        if (count < 0 || count >= 32) throw new Error("shift out of range");
        return { kind: "i32", value: left.value >> count };
      }
    }
  }

  throw new Error(`Unknown binary operator ${b.operator}`);
}

export function evaluateRangeExpression(ctx: InterpreterV10, node: AST.RangeExpression, env: Environment): V10Value {
  const s = ctx.evaluate(node.start, env);
  const e = ctx.evaluate(node.end, env);
  const sNum = (s.kind === "i32" || s.kind === "f64") ? s.value : NaN;
  const eNum = (e.kind === "i32" || e.kind === "f64") ? e.value : NaN;
  if (Number.isNaN(sNum) || Number.isNaN(eNum)) throw new Error("Range boundaries must be numeric");
  return { kind: "range", start: sNum, end: eNum, inclusive: node.inclusive };
}

export function evaluateIndexExpression(ctx: InterpreterV10, node: AST.IndexExpression, env: Environment): V10Value {
  const obj = ctx.evaluate(node.object, env);
  const idxVal = ctx.evaluate(node.index, env);
  if (obj.kind !== "array") throw new Error("Indexing is only supported on arrays in this milestone");
  const idx = (idxVal.kind === "i32" || idxVal.kind === "f64") ? Math.trunc(idxVal.value) : NaN;
  if (!Number.isFinite(idx)) throw new Error("Array index must be a number");
  if (idx < 0 || idx >= obj.elements.length) throw new Error("Array index out of bounds");
  const el = obj.elements[idx];
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
  cls.prototype.computeBinaryForCompound = function computeBinaryForCompound(this: InterpreterV10, op: string, left: V10Value, right: V10Value): V10Value {
    const isNum = (v: V10Value) => v.kind === "i32" || v.kind === "f64";
    const asNumber = (v: V10Value) => (v.kind === "i32" || v.kind === "f64") ? v.value : NaN;
    const resultKind = (a: V10Value, c: V10Value): "i32" | "f64" => (a.kind === "f64" || c.kind === "f64") ? "f64" : "i32";

    if (["+","-","*","/","%"].includes(op)) {
      if (!isNum(left) || !isNum(right)) throw new Error("Arithmetic requires numeric operands");
      const kind = resultKind(left, right);
      const l = asNumber(left);
      const r = asNumber(right);
      switch (op) {
        case "+": return kind === "i32" ? { kind, value: (l + r) | 0 } : { kind, value: l + r };
        case "-": return kind === "i32" ? { kind, value: (l - r) | 0 } : { kind, value: l - r };
        case "*": return kind === "i32" ? { kind, value: (l * r) | 0 } : { kind, value: l * r };
        case "/": {
          if (r === 0) throw new Error("Division by zero");
          return kind === "i32" ? { kind, value: (l / r) | 0 } : { kind, value: l / r };
        }
        case "%": {
          if (r === 0) throw new Error("Division by zero");
          return { kind, value: kind === "i32" ? (l % r) | 0 : l % r };
        }
      }
    }

    if (["&","|","^","<<",">>"] .includes(op)) {
      if (left.kind !== "i32" || right.kind !== "i32") throw new Error("Bitwise requires i32 operands");
      switch (op) {
        case "&": return { kind: "i32", value: left.value & right.value };
        case "|": return { kind: "i32", value: left.value | right.value };
        case "^": return { kind: "i32", value: left.value ^ right.value };
        case "<<": {
          const count = right.value;
          if (count < 0 || count >= 32) throw new Error("shift out of range");
          return { kind: "i32", value: left.value << count };
        }
        case ">>": {
          const count = right.value;
          if (count < 0 || count >= 32) throw new Error("shift out of range");
          return { kind: "i32", value: left.value >> count };
        }
      }
    }

    throw new Error(`Unsupported compound operator ${op}`);
  };
}
