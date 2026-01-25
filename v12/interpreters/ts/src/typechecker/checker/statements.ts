import type * as AST from "../../ast";
import {
  formatType,
  isBoolean,
  isIntegerPrimitiveType,
  isUnknown,
  type TypeInfo,
  primitiveType,
  unknownType,
} from "../types";
import type { StatementContext } from "./expression-context";
import { refineTypeWithExpected } from "./expressions";
import { bindPatternToEnv } from "./patterns";

export function checkStatement(ctx: StatementContext, node: AST.Statement | AST.Expression | undefined | null): void {
  if (!node) return;
  switch (node.type) {
    case "InterfaceDefinition":
    case "StructDefinition":
    case "UnionDefinition":
    case "TypeAliasDefinition":
      ctx.handleTypeDeclaration?.(node);
      return;
    case "ImplementationDefinition":
    case "MethodsDefinition":
      return;
    case "FunctionDefinition":
      ctx.checkFunctionDefinition(node);
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
    case "BreakStatement":
      ctx.handleBreakStatement?.(node);
      return;
    case "ContinueStatement":
      ctx.handleContinueStatement?.(node);
      return;
    default:
      if (ctx.isExpression(node)) {
        ctx.inferExpression(node);
      }
      return;
  }
}

export function checkAssignment(ctx: StatementContext, node: AST.AssignmentExpression): TypeInfo {
  const declarationAssignmentsSeen: WeakSet<AST.AssignmentExpression> =
    (checkAssignment as any)._seen ?? new WeakSet<AST.AssignmentExpression>();
  (checkAssignment as any)._seen = declarationAssignmentsSeen;
  const alreadyProcessed = declarationAssignmentsSeen.has(node);
  const isDeclaration = node.operator === ":=" && !alreadyProcessed;
  if (node.operator === ":=" && !alreadyProcessed) {
    declarationAssignmentsSeen.add(node);
  }
  let declarationNames: Set<string> | undefined;
  if (isDeclaration && isPatternNode(node.left)) {
    const analysis = analyzePatternDeclarations(ctx, node.left);
    declarationNames = analysis.declarationNames;
    if (analysis.hasAny && declarationNames.size === 0) {
      ctx.report("typechecker: ':=' requires at least one new binding", node.left);
    }
    predeclarePattern(ctx, node.left, declarationNames);
  }
  let expectedType: TypeInfo | null = null;
  if (node.left?.type === "TypedPattern" && node.left.typeAnnotation) {
    expectedType = ctx.resolveTypeExpression(node.left.typeAnnotation);
  } else if (node.operator === "=" && node.left?.type === "Identifier") {
    const existing = ctx.lookupIdentifier(node.left.name);
    if (existing && existing.kind !== "unknown") {
      expectedType = existing;
    }
  }
  let valueType = ctx.inferExpression(node.right);
  if (expectedType && expectedType.kind !== "unknown" && node.right?.type === "FunctionCall") {
    valueType = refineTypeWithExpected(valueType, expectedType);
  }
  let targetType = valueType;
  if (node.operator === "=" && node.left?.type === "Identifier") {
    const existing = ctx.lookupIdentifier(node.left.name);
    if (existing && existing.kind !== "unknown") {
      if (!ctx.isTypeAssignable(valueType, existing)) {
        ctx.report(
          `typechecker: assignment expects type ${formatType(existing)}, got ${formatType(valueType)}`,
          node.right ?? node,
        );
      }
      targetType = existing;
    }
  }
  if (!node.left) {
    return valueType;
  }
  if (node.left.type === "IndexExpression") {
    checkIndexAssignment(ctx, node, valueType);
    return valueType;
  }
  if (node.left.type === "MemberAccessExpression") {
    ctx.inferExpression(node.left);
    return valueType;
  }
  if (node.left.type === "ImplicitMemberExpression") {
    if (node.operator === ":=") {
      ctx.report("typechecker: cannot use := on implicit member access", node.left);
    }
    ctx.inferExpression(node.left);
    return valueType;
  }
  const bindingOptions = {
    suppressMismatchReport: true,
    declarationNames,
    isDeclaration,
    allowFallbackDeclaration: !isDeclaration,
  };
  bindPatternToEnv(ctx, node.left as AST.Pattern, targetType, "assignment pattern", bindingOptions);
  return valueType;
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
  const { elementType, recognized } = resolveIterableElementType(ctx, iterableType);
  if (!recognized && !isUnknown(iterableType)) {
    ctx.report(
      `typechecker: for-loop iterable must implement Iterable (got ${formatType(iterableType)})`,
      loop.iterable,
    );
  }
  ctx.pushLoopContext();
  ctx.pushScope();
  try {
    bindPatternToEnv(ctx, loop.pattern as AST.Pattern, elementType, "for-loop pattern");
    if (loop.body) {
      ctx.inferExpression(loop.body);
    }
  } finally {
    ctx.popScope();
    ctx.popLoopContext();
  }
}

function checkWhileLoop(ctx: StatementContext, loop: AST.WhileLoop): void {
  if (!loop) return;
    ctx.inferExpression(loop.condition);
  ctx.pushLoopContext();
  ctx.pushScope();
  try {
    if (loop.body) {
      ctx.inferExpression(loop.body);
    }
  } finally {
    ctx.popScope();
    ctx.popLoopContext();
  }
}

function resolveIterableElementType(
  ctx: StatementContext,
  type: TypeInfo,
): { elementType: TypeInfo; recognized: boolean } {
  if (!type || type.kind === "unknown") {
    return { elementType: unknownType, recognized: true };
  }
  if (type.kind === "primitive" && type.name === "String") {
    return { elementType: primitiveType("u8"), recognized: true };
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
  if (
    type.kind === "struct" &&
    (type.name === "List" ||
      type.name === "LinkedList" ||
      type.name === "LazySeq" ||
      type.name === "Vector" ||
      type.name === "HashSet" ||
      type.name === "Deque" ||
      type.name === "Queue" ||
      type.name === "Channel")
  ) {
    const candidate =
      Array.isArray(type.typeArguments) && type.typeArguments.length > 0 ? type.typeArguments[0]! : unknownType;
    return { elementType: candidate ?? unknownType, recognized: true };
  }
  if (type.kind === "struct" && type.name === "BitSet") {
    return { elementType: primitiveType("i32"), recognized: true };
  }
  if (type.kind === "struct" && type.name === "Range") {
    return { elementType: unknownType, recognized: true };
  }
  if (type.kind === "interface" && type.name === "Iterable") {
    const candidate = Array.isArray(type.typeArguments) && type.typeArguments.length > 0 ? type.typeArguments[0]! : null;
    return { elementType: candidate ?? unknownType, recognized: true };
  }
  if (ctx.typeImplementsInterface?.(type, "Iterable", ["Unknown"])?.ok) {
    if (type.kind === "struct" && Array.isArray(type.typeArguments) && type.typeArguments.length > 0) {
      return { elementType: type.typeArguments[0] ?? unknownType, recognized: true };
    }
    return { elementType: unknownType, recognized: true };
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

function checkIndexAssignment(ctx: StatementContext, node: AST.AssignmentExpression, valueType: TypeInfo): void {
  const target = node.left as AST.IndexExpression;
  const objectType = ctx.inferExpression(target.object);
  const indexType = ctx.inferExpression(target.index);
  if (node.operator === ":=") {
    ctx.report("typechecker: cannot use := on index assignment", target);
    return;
  }
  if (!objectType || objectType.kind === "unknown") {
    return;
  }
  const reportValueMismatch = (expected: TypeInfo): void => {
    if (
      expected &&
      expected.kind !== "unknown" &&
      valueType &&
      valueType.kind !== "unknown" &&
      !ctx.isTypeAssignable(valueType, expected)
    ) {
      ctx.report(
        `typechecker: index assignment expects value type ${formatType(expected)}, got ${formatType(valueType)}`,
        node.right ?? target,
      );
    }
  };
  const requireIntegerIndex = (): void => {
    if (!isIntegerPrimitiveType(indexType) && indexType.kind !== "unknown") {
      ctx.report("typechecker: index must be an integer", target.index);
    }
  };
  if (objectType.kind === "array") {
    requireIntegerIndex();
    reportValueMismatch(objectType.element ?? unknownType);
    return;
  }
  if (objectType.kind === "struct" && objectType.name === "Array") {
    requireIntegerIndex();
    reportValueMismatch((objectType.typeArguments ?? [])[0] ?? unknownType);
    return;
  }
  if (objectType.kind === "map") {
    reportValueMismatch(objectType.value ?? unknownType);
    return;
  }
  if (objectType.kind === "struct" && objectType.name === "HashMap") {
    const args = objectType.typeArguments ?? [];
    const keyType = args[0];
    const value = args[1];
    if (keyType && !ctx.isTypeAssignable(indexType, keyType) && indexType.kind !== "unknown") {
      ctx.report(
        `typechecker: index expects type ${formatType(keyType)}, got ${formatType(indexType)}`,
        target.index,
      );
    }
    reportValueMismatch(value ?? unknownType);
    return;
  }
  if (objectType.kind === "interface" && objectType.name === "IndexMut") {
    const args = objectType.typeArguments ?? [];
    const keyType = args[0];
    const value = args[1];
    if (keyType && !ctx.isTypeAssignable(indexType, keyType) && indexType.kind !== "unknown") {
      ctx.report(
        `typechecker: index expects type ${formatType(keyType)}, got ${formatType(indexType)}`,
        target.index,
      );
    }
    reportValueMismatch(value ?? unknownType);
    return;
  }
  if (objectType.kind === "interface" && objectType.name === "Index") {
    ctx.report(
      `typechecker: cannot assign via [] without IndexMut implementation on type ${formatType(objectType)}`,
      target,
    );
    return;
  }
  if (ctx.typeImplementsInterface?.(objectType, "IndexMut", ["Unknown", "Unknown"])?.ok) {
    return;
  }
  if (ctx.typeImplementsInterface?.(objectType, "Index", ["Unknown", "Unknown"])?.ok) {
    ctx.report(
      `typechecker: cannot assign via [] without IndexMut implementation on type ${formatType(objectType)}`,
      target,
    );
    return;
  }
  ctx.report(`typechecker: cannot assign into type ${formatType(objectType)}`, target);
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
