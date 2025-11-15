import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { numericToNumber } from "./numeric";

function isPatternNode(node: AST.Node | undefined | null): node is AST.Pattern {
  if (!node) return false;
  switch (node.type) {
    case "Identifier":
    case "WildcardPattern":
    case "LiteralPattern":
    case "StructPattern":
    case "ArrayPattern":
    case "TypedPattern":
      return true;
    default:
      return false;
  }
}

function collectPatternIdentifiers(pattern: AST.Pattern | undefined | null, into: Set<string>): void {
  if (!pattern) return;
  switch (pattern.type) {
    case "Identifier":
      if (pattern.name) into.add(pattern.name);
      return;
    case "StructPattern":
      if (Array.isArray(pattern.fields)) {
        for (const field of pattern.fields) {
          if (!field) continue;
          if (field.binding?.name) into.add(field.binding.name);
          collectPatternIdentifiers(field.pattern as AST.Pattern, into);
        }
      }
      return;
    case "ArrayPattern":
      if (Array.isArray(pattern.elements)) {
        for (const element of pattern.elements) {
          collectPatternIdentifiers(element as AST.Pattern, into);
        }
      }
      if (pattern.restPattern?.type === "Identifier" && pattern.restPattern.name) {
        into.add(pattern.restPattern.name);
      }
      return;
    case "TypedPattern":
      collectPatternIdentifiers(pattern.pattern as AST.Pattern, into);
      return;
    case "WildcardPattern":
    case "LiteralPattern":
    default:
      return;
  }
}

function analyzeDeclarationTargets(
  target: AST.Pattern,
  env: Environment,
): { declarationNames: Set<string>; hasAny: boolean } {
  const names = new Set<string>();
  collectPatternIdentifiers(target, names);
  const declarationNames = new Set<string>();
  for (const name of names) {
    if (!env.hasInCurrentScope(name)) {
      declarationNames.add(name);
    }
  }
  return { declarationNames, hasAny: names.size > 0 };
}

export function evaluateAssignmentExpression(ctx: InterpreterV10, node: AST.AssignmentExpression, env: Environment): V10Value {
  const value = ctx.evaluate(node.right, env);
  const isCompound = ["+=", "-=", "*=", "/=", "%=", "&=", "|=", "\\xor=", "<<=", ">>="].includes(node.operator);

  if (node.left.type === "Identifier") {
    if (node.operator === ":=") {
      if (env.hasInCurrentScope(node.left.name)) {
        throw new Error(":= requires at least one new binding");
      }
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
    if (!env.assignExisting(node.left.name, value)) {
      env.define(node.left.name, value);
    }
    return value;
  }

  if (isPatternNode(node.left as AST.Pattern)) {
    if (isCompound) throw new Error("Compound assignment not supported with destructuring");
    const pattern = node.left as AST.Pattern;
    if (node.operator === ":=") {
      const { declarationNames, hasAny } = analyzeDeclarationTargets(pattern, env);
      if (!hasAny || declarationNames.size === 0) {
        throw new Error(":= requires at least one new binding");
      }
      ctx.assignByPattern(pattern, value, env, true, { declarationNames });
      return value;
    }
    ctx.assignByPattern(pattern, value, env, false, { fallbackToDeclaration: true });
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
    const idx = Math.trunc(numericToNumber(idxVal, "Array index", { requireSafeInteger: true }));
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
