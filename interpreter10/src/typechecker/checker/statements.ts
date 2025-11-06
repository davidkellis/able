import type * as AST from "../../ast";
import {
  formatType,
  isBoolean,
  isUnknown,
  type TypeInfo,
  primitiveType,
  unknownType,
} from "../types";
import { bindPatternToEnv, type StatementContext } from "./expressions";

export function checkStatement(ctx: StatementContext, node: AST.Statement | AST.Expression | undefined | null): void {
  if (!node) return;
  switch (node.type) {
    case "InterfaceDefinition":
    case "StructDefinition":
    case "ImplementationDefinition":
    case "MethodsDefinition":
    case "FunctionDefinition":
      return;
    case "AssignmentExpression":
      checkAssignment(ctx, node);
      return;
    case "RaiseStatement":
      checkRaiseStatement(ctx, node);
      return;
    case "RethrowStatement":
      checkRethrowStatement(ctx, node);
      return;
    case "WhileLoop":
      checkWhileLoop(ctx, node);
      return;
    case "ForLoop":
      checkForLoop(ctx, node);
      return;
    default:
      if (ctx.isExpression(node)) {
        ctx.inferExpression(node);
      }
      return;
  }
}

function checkAssignment(ctx: StatementContext, node: AST.AssignmentExpression): void {
  const valueType = ctx.inferExpression(node.right);
  if (node.left?.type === "StructPattern") {
    bindPatternToEnv(ctx, node.left, valueType, "assignment pattern");
    return;
  }
  if (node.left?.type === "Identifier" && node.left.name) {
    ctx.defineValue(node.left.name, valueType);
  }
}

function checkRaiseStatement(ctx: StatementContext, node: AST.RaiseStatement): void {
  if (node.expression) {
    ctx.inferExpression(node.expression);
  }
}

function checkRethrowStatement(ctx: StatementContext, node: AST.RethrowStatement): void {
  if (node.expression) {
    ctx.inferExpression(node.expression);
  }
}

function checkForLoop(ctx: StatementContext, loop: AST.ForLoop): void {
  if (!loop) return;
  const iterableType = ctx.inferExpression(loop.iterable);
  const { elementType, recognized } = resolveIterableElementType(iterableType);
  if (!recognized && !isUnknown(iterableType)) {
    ctx.report(
      `typechecker: for-loop iterable must be array, range, or iterator (got ${formatType(iterableType)})`,
      loop.iterable,
    );
  }
  ctx.pushScope();
  try {
    bindPatternToEnv(ctx, loop.pattern as AST.Pattern, elementType, "for-loop pattern");
    if (loop.body) {
      ctx.inferExpression(loop.body);
    }
  } finally {
    ctx.popScope();
  }
}

function checkWhileLoop(ctx: StatementContext, loop: AST.WhileLoop): void {
  if (!loop) return;
  const conditionType = ctx.inferExpression(loop.condition);
  if (!isBoolean(conditionType)) {
    ctx.report("typechecker: while condition must be bool", loop.condition);
  }
  ctx.pushScope();
  try {
    if (loop.body) {
      ctx.inferExpression(loop.body);
    }
  } finally {
    ctx.popScope();
  }
}

function resolveIterableElementType(type: TypeInfo): { elementType: TypeInfo; recognized: boolean } {
  if (!type || type.kind === "unknown") {
    return { elementType: unknownType, recognized: true };
  }
  if (type.kind === "iterator") {
    return { elementType: type.element ?? unknownType, recognized: true };
  }
  if (type.kind === "array") {
    return { elementType: type.element ?? unknownType, recognized: true };
  }
  if (type.kind === "range") {
    return { elementType: type.element ?? unknownType, recognized: true };
  }
  if (type.kind === "struct" && type.name === "Array") {
    const candidate =
      Array.isArray(type.typeArguments) && type.typeArguments.length > 0 ? type.typeArguments[0]! : unknownType;
    return { elementType: candidate ?? unknownType, recognized: true };
  }
  if (type.kind === "struct" && type.name === "Iterator") {
    const candidate =
      Array.isArray(type.typeArguments) && type.typeArguments.length > 0 ? type.typeArguments[0]! : unknownType;
    return { elementType: candidate ?? unknownType, recognized: true };
  }
  if (type.kind === "struct" && type.name === "Range") {
    return { elementType: unknownType, recognized: true };
  }
  return { elementType: unknownType, recognized: false };
}
