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
    case "TypeAliasDefinition":
      if (node.type === "FunctionDefinition") {
        ctx.checkFunctionDefinition(node);
      }
      return;
    case "ReturnStatement":
      ctx.checkReturnStatement(node);
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
  let declarationNames: Set<string> | undefined;
  if (node.operator === ":=" && isPatternNode(node.left)) {
    const analysis = analyzePatternDeclarations(ctx, node.left);
    declarationNames = analysis.declarationNames;
    if (analysis.hasAny && declarationNames.size === 0) {
      ctx.report("typechecker: ':=' requires at least one new binding", node.left);
    }
    predeclarePattern(ctx, node.left, declarationNames);
  }
  const valueType = ctx.inferExpression(node.right);
  if (!node.left) {
    return;
  }
  if (node.left.type === "MemberAccessExpression" || node.left.type === "IndexExpression") {
    ctx.inferExpression(node.left);
    return;
  }
  const bindingOptions = {
    suppressMismatchReport: true,
    declarationNames,
    isDeclaration: node.operator === ":=",
    allowFallbackDeclaration: node.operator !== ":=",
  };
  bindPatternToEnv(ctx, node.left as AST.Pattern, valueType, "assignment pattern", bindingOptions);
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
  if (type.kind === "interface" && type.name === "Iterable") {
    const candidate = Array.isArray(type.typeArguments) && type.typeArguments.length > 0 ? type.typeArguments[0]! : null;
    return { elementType: candidate ?? unknownType, recognized: true };
  }
  return { elementType: unknownType, recognized: false };
}

function isPatternNode(node: AST.Node | undefined | null): node is AST.Pattern {
  if (!node) return false;
  switch (node.type) {
    case "Identifier":
    case "WildcardPattern":
    case "TypedPattern":
    case "StructPattern":
    case "ArrayPattern":
    case "LiteralPattern":
      return true;
    default:
      return false;
  }
}

function predeclarePattern(
  ctx: StatementContext,
  pattern: AST.Pattern | undefined | null,
  declarationNames?: Set<string>,
): void {
  if (!pattern) return;
  switch (pattern.type) {
    case "Identifier":
      if (pattern.name && (!declarationNames || declarationNames.has(pattern.name))) {
        ctx.defineValue(pattern.name, unknownType);
      }
      return;
    case "WildcardPattern":
    case "LiteralPattern":
      return;
    case "TypedPattern":
      if (pattern.pattern) {
        predeclarePattern(ctx, pattern.pattern as AST.Pattern, declarationNames);
      }
      return;
    case "StructPattern":
      if (!Array.isArray(pattern.fields)) {
        return;
      }
      for (const field of pattern.fields) {
        if (!field) continue;
        if (field.pattern) {
          predeclarePattern(ctx, field.pattern as AST.Pattern, declarationNames);
        }
        if (field.binding?.name && (!declarationNames || declarationNames.has(field.binding.name))) {
          ctx.defineValue(field.binding.name, unknownType);
        }
      }
      return;
    case "ArrayPattern":
      if (Array.isArray(pattern.elements)) {
        for (const element of pattern.elements) {
          predeclarePattern(ctx, element as AST.Pattern, declarationNames);
        }
      }
      if (
        pattern.restPattern &&
        pattern.restPattern.type === "Identifier" &&
        pattern.restPattern.name &&
        (!declarationNames || declarationNames.has(pattern.restPattern.name))
      ) {
        ctx.defineValue(pattern.restPattern.name, unknownType);
      }
      return;
    default:
      return;
  }
}

function collectPatternIdentifiers(pattern: AST.Pattern | undefined | null, into: Set<string>): void {
  if (!pattern) return;
  switch (pattern.type) {
    case "Identifier":
      if (pattern.name) into.add(pattern.name);
      return;
    case "StructPattern":
      if (!Array.isArray(pattern.fields)) return;
      for (const field of pattern.fields) {
        if (!field) continue;
        if (field.binding?.name) {
          into.add(field.binding.name);
        }
        collectPatternIdentifiers(field.pattern as AST.Pattern, into);
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
    default:
      return;
  }
}

function analyzePatternDeclarations(
  ctx: StatementContext,
  pattern: AST.Pattern,
): { declarationNames: Set<string>; hasAny: boolean } {
  const names = new Set<string>();
  collectPatternIdentifiers(pattern, names);
  const declarationNames = new Set<string>();
  for (const name of names) {
    if (!ctx.hasBindingInCurrentScope(name)) {
      declarationNames.add(name);
    }
  }
  return { declarationNames, hasAny: names.size > 0 };
}
