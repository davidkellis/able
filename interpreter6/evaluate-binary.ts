import * as AST from "./ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./interpreter";
import type { AbleValue } from "./runtime";
import { createError, hasKind } from "./runtime";

export function evaluateBinaryExpression(this: Interpreter, node: AST.BinaryExpression, environment: Environment): AbleValue {
  const left = this.evaluate(node.left, environment);
  if (node.operator === "&&") {
    if (!hasKind(left, "bool")) throw new Error("Interpreter Error: Left operand of && must be boolean");
    if (!left.value) return { kind: "bool", value: false };
    const right = this.evaluate(node.right, environment);
    if (!hasKind(right, "bool")) throw new Error("Interpreter Error: Right operand of && must be boolean");
    return right;
  }
  if (node.operator === "||") {
    if (!hasKind(left, "bool")) throw new Error("Interpreter Error: Left operand of || must be boolean");
    if (left.value) return { kind: "bool", value: true };
    const right = this.evaluate(node.right, environment);
    if (!hasKind(right, "bool")) throw new Error("Interpreter Error: Right operand of || must be boolean");
    return right;
  }

  const right = this.evaluate(node.right, environment);

  if (["+", "-", "*", "/", "%"].includes(node.operator)) {
    if ("value" in left && "value" in right) {
      if (
        typeof left.value === "number" &&
        typeof right.value === "number" &&
        typeof left.kind === "string" &&
        left.kind.match(/^(i(8|16|32)|u(8|16|32)|f(32|64))$/) &&
        left.kind === right.kind
      ) {
        const kind = left.kind;
        switch (node.operator) {
          case "+":
            return { kind, value: left.value + right.value };
          case "-":
            return { kind, value: left.value - right.value };
          case "*":
            return { kind, value: left.value * right.value };
          case "/":
            if (right.value === 0) throw createError("Division by zero");
            return { kind, value: kind.startsWith("f") ? left.value / right.value : Math.trunc(left.value / right.value) };
          case "%":
            if (right.value === 0) throw createError("Division by zero");
            return { kind, value: left.value % right.value };
        }
      }
      if (
        typeof left.value === "bigint" &&
        typeof right.value === "bigint" &&
        typeof left.kind === "string" &&
        left.kind.match(/^(i(64|128)|u(64|128))$/) &&
        left.kind === right.kind
      ) {
        const kind = left.kind;
        switch (node.operator) {
          case "+":
            return { kind, value: left.value + right.value };
          case "-":
            return { kind, value: left.value - right.value };
          case "*":
            return { kind, value: left.value * right.value };
          case "/":
            if (right.value === 0n) throw createError("Division by zero");
            return { kind, value: left.value / right.value };
          case "%":
            if (right.value === 0n) throw createError("Division by zero");
            return { kind, value: left.value % right.value };
        }
      }
      if (node.operator === "+" && left.kind === "string" && right.kind === "string") {
        return { kind: "string", value: left.value + right.value };
      }
    }
    throw new Error(
      `Interpreter Error: Operator '${node.operator}' not supported for types ${left?.kind ?? typeof left} and ${right?.kind ?? typeof right}`,
    );
  }

  if ([">", "<", ">=", "<=", "==", "!="].includes(node.operator)) {
    if ("value" in left && "value" in right) {
      if (left.kind === "nil" || right.kind === "nil") {
        if (node.operator === "==") return { kind: "bool", value: left.kind === right.kind };
        if (node.operator === "!=") return { kind: "bool", value: left.kind !== right.kind };
        throw new Error(`Interpreter Error: Operator '${node.operator}' not supported for nil.`);
      }
      const lVal = left.value!;
      const rVal = right.value!;
      if (typeof lVal === "number" && typeof rVal === "number" && left.kind === right.kind) {
        switch (node.operator) {
          case ">":
            return { kind: "bool", value: lVal > rVal };
          case "<":
            return { kind: "bool", value: lVal < rVal };
          case ">=":
            return { kind: "bool", value: lVal >= rVal };
          case "<=":
            return { kind: "bool", value: lVal <= rVal };
        }
      }
      if (typeof lVal === "bigint" && typeof rVal === "bigint" && left.kind === right.kind) {
        switch (node.operator) {
          case ">":
            return { kind: "bool", value: lVal > rVal };
          case "<":
            return { kind: "bool", value: lVal < rVal };
          case ">=":
            return { kind: "bool", value: lVal >= rVal };
          case "<=":
            return { kind: "bool", value: lVal <= rVal };
        }
      }
      if (typeof lVal === typeof rVal) {
        try {
          switch (node.operator) {
            case ">":
              return { kind: "bool", value: lVal > rVal };
            case "<":
              return { kind: "bool", value: lVal < rVal };
            case ">=":
              return { kind: "bool", value: lVal >= rVal };
            case "<=":
              return { kind: "bool", value: lVal <= rVal };
          }
        } catch {
          throw new Error(`Interpreter Error: Cannot compare ${left.kind} and ${right.kind} with ${node.operator}`);
        }
      }
      if (node.operator === "==") return { kind: "bool", value: left.kind === right.kind && lVal === rVal };
      if (node.operator === "!=") return { kind: "bool", value: left.kind !== right.kind || lVal !== rVal };
    } else {
      if (node.operator === "==") return { kind: "bool", value: false };
      if (node.operator === "!=") return { kind: "bool", value: true };
      throw new Error(
        `Interpreter Error: Comparison operator '${node.operator}' not supported for non-primitive types ${left?.kind ?? typeof left} and ${right?.kind ?? typeof right}`,
      );
    }
  }

  if (["&", "|", "^", "<<", ">>"].includes(node.operator)) {
    if ("value" in left && "value" in right) {
      if (typeof left.value === "number" && typeof right.value === "number" && left.kind.match(/^i|^u/) && left.kind === right.kind) {
        const kind = left.kind;
        switch (node.operator) {
          case "&":
            return { kind, value: left.value & right.value };
          case "|":
            return { kind, value: left.value | right.value };
          case "^":
            return { kind, value: left.value ^ right.value };
          case "<<":
            return { kind, value: left.value << right.value };
          case ">>":
            return { kind, value: left.value >> right.value };
        }
      }
      if (typeof left.value === "bigint" && typeof right.value === "bigint" && left.kind.match(/^i|^u/) && left.kind === right.kind) {
        const kind = left.kind;
        const shiftVal =
          node.operator === "<<" || node.operator === ">>"
            ? BigInt(Number(right.value))
            : right.value;
        if ((node.operator === "<<" || node.operator === ">>") && shiftVal < 0n) {
          throw new Error("Interpreter Error: Shift amount cannot be negative.");
        }
        switch (node.operator) {
          case "&":
            return { kind, value: left.value & right.value };
          case "|":
            return { kind, value: left.value | right.value };
          case "^":
            return { kind, value: left.value ^ right.value };
          case "<<":
            return { kind, value: left.value << shiftVal };
          case ">>":
            return { kind, value: left.value >> shiftVal };
        }
      }
    }
    throw new Error(
      `Interpreter Error: Bitwise operator '${node.operator}' not supported for types ${left?.kind ?? typeof left} and ${right?.kind ?? typeof right}`,
    );
  }

  throw new Error(`Interpreter Error: Unknown binary operator ${node.operator}`);
}
