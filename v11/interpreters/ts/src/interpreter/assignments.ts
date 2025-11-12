import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

export function evaluateAssignmentExpression(ctx: InterpreterV10, node: AST.AssignmentExpression, env: Environment): V10Value {
  const value = ctx.evaluate(node.right, env);
  const isCompound = ["+=", "-=", "*=", "/=", "%=", "&=", "|=", "\\xor=", "<<=", ">>="].includes(node.operator);

  if (node.left.type === "Identifier") {
    if (node.operator === ":=") {
      env.define(node.left.name, value);
      return value;
    }
    if (isCompound) {
      const current = env.get(node.left.name);
      const op = node.operator.slice(0, -1);
      const computed = ctx.computeBinaryForCompound(op, current, value);
      env.assign(node.left.name, computed);
      return computed;
    }
    env.assign(node.left.name, value);
    return value;
  }

  if (node.left.type === "StructPattern" || node.left.type === "ArrayPattern" || node.left.type === "WildcardPattern" || node.left.type === "LiteralPattern" || (node.left as any).type === "TypedPattern") {
    if (isCompound) throw new Error("Compound assignment not supported with destructuring");
    const isDeclaration = node.operator === ":=";
    ctx.assignByPattern(node.left as AST.Pattern, value, env, isDeclaration);
    return value;
  }

  if (node.left.type === "MemberAccessExpression") {
    if (node.operator === ":=") throw new Error("Cannot use := on member access");
    const targetObj = ctx.evaluate(node.left.object, env);
    if (targetObj.kind === "struct_instance") {
      if (node.left.member.type === "Identifier") {
        if (!(targetObj.values instanceof Map)) throw new Error("Expected named struct instance");
        if (!targetObj.values.has(node.left.member.name)) throw new Error(`No field named '${node.left.member.name}'`);
        if (isCompound) {
          const current = targetObj.values.get(node.left.member.name)!;
          const op = node.operator.slice(0, -1);
          const computed = ctx.computeBinaryForCompound(op, current, value);
          targetObj.values.set(node.left.member.name, computed);
          return computed;
        }
        targetObj.values.set(node.left.member.name, value);
        return value;
      }
      if (!Array.isArray(targetObj.values)) throw new Error("Expected positional struct instance");
      const idx = Number(node.left.member.value);
      if (idx < 0 || idx >= targetObj.values.length) throw new Error("Struct field index out of bounds");
      if (isCompound) {
        const current = targetObj.values[idx] as V10Value;
        const op = node.operator.slice(0, -1);
        const computed = ctx.computeBinaryForCompound(op, current, value);
        targetObj.values[idx] = computed;
        return computed;
      }
      targetObj.values[idx] = value;
      return value;
    }
    if (targetObj.kind === "array") {
      if (node.left.member.type !== "IntegerLiteral") throw new Error("Array member assignment requires integer member");
      const idx = Number(node.left.member.value);
      if (idx < 0 || idx >= targetObj.elements.length) throw new Error("Array index out of bounds");
      if (isCompound) {
        const current = targetObj.elements[idx]!;
        const op = node.operator.slice(0, -1);
        const computed = ctx.computeBinaryForCompound(op, current, value);
        targetObj.elements[idx] = computed;
        return computed;
      }
      targetObj.elements[idx] = value;
      return value;
    }
    throw new Error("Member assignment requires struct or array");
  }

  if (node.left.type === "IndexExpression") {
    if (node.operator === ":=") throw new Error("Cannot use := on index assignment");
    const obj = ctx.evaluate(node.left.object, env);
    const idxVal = ctx.evaluate(node.left.index, env);
    if (obj.kind !== "array") throw new Error("Index assignment requires array");
    const idx = (idxVal.kind === "i32" || idxVal.kind === "f64") ? Math.trunc(idxVal.value) : NaN;
    if (!Number.isFinite(idx)) throw new Error("Array index must be a number");
    if (idx < 0 || idx >= obj.elements.length) throw new Error("Array index out of bounds");
    if (isCompound) {
      const current = obj.elements[idx]!;
      const op = node.operator.slice(0, -1);
      const computed = ctx.computeBinaryForCompound(op, current, value);
      obj.elements[idx] = computed;
      return computed;
    }
    obj.elements[idx] = value;
    return value;
  }

  throw new Error("Unsupported assignment target");
}
