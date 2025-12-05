import * as AST from "../ast";
import type { Expression, Identifier, Literal, Pattern, StructPattern, StructPatternField } from "../ast";
import {
  annotate,
  annotatePatternNode,
  getActiveParseContext,
  identifiersToStrings,
  isIgnorableNode,
  MapperError,
  MutableParseContext,
  Node,
  ParseContext,
  parseIdentifier,
  sameNode,
  sliceText,
} from "./shared";

export function registerPatternParsers(ctx: MutableParseContext): void {
  ctx.parsePattern = node => parsePattern(node, ctx.source);
}

function parsePattern(
  node: Node | null | undefined,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): Pattern {
  if (!node) {
    throw new MapperError("parser: nil pattern");
  }

  if (node.type === "pattern" || node.type === "pattern_base") {
    if (node.namedChildCount === 0) {
      const text = sliceText(node, source).trim();
      if (text === "_") {
        return annotatePatternNode(AST.wildcardPattern(), node);
      }
      for (let i = 0; i < node.childCount; i++) {
        const child = node.child(i);
        if (!child || isIgnorableNode(child)) continue;
        if (child.isNamed) {
          return parsePattern(child, source, ctx);
        }
        if (sliceText(child, source).trim() === "_") {
          return annotatePatternNode(AST.wildcardPattern(), child);
        }
      }
      throw new MapperError(`parser: empty ${node.type}`);
    }
    return parsePattern(node.namedChild(0), source, ctx);
  }

  switch (node.type) {
    case "identifier":
      return parseIdentifier(node, source);
    case "_":
      return annotatePatternNode(AST.wildcardPattern(), node);
    case "literal_pattern":
      return parseLiteralPattern(node, source, ctx);
    case "struct_pattern":
      return parseStructPattern(node, source, ctx);
    case "array_pattern":
      return parseArrayPattern(node, source, ctx);
    case "parenthesized_pattern": {
      const inner = node.namedChild(0);
      if (inner) {
        return parsePattern(inner, source, ctx);
      }
      throw new MapperError("parser: empty parenthesized pattern");
    }
    case "typed_pattern":
      if (node.namedChildCount < 2) {
        throw new MapperError("parser: malformed typed pattern");
      }
      const innerPattern = parsePattern(node.namedChild(0), source, ctx);
      const typeExpr = ctx.parseTypeExpression(node.namedChild(1));
      if (!typeExpr) {
        throw new MapperError("parser: typed pattern missing type expression");
      }
      return annotatePatternNode(AST.typedPattern(innerPattern, typeExpr), node);
    case "pattern":
    case "pattern_base":
      return parsePattern(node.namedChild(0), source, ctx);
    default:
      throw new MapperError(`parser: unsupported pattern kind ${node.type}`);
  }
}

function parseLiteralPattern(node: Node, source: string, ctx: ParseContext): Pattern {
  const literalNode = node.namedChild(0);
  if (!literalNode) {
    throw new MapperError("parser: literal pattern missing literal");
  }
  const literalExpr = ctx.parseExpression(literalNode);
  if (!isLiteralExpression(literalExpr)) {
    throw new MapperError(`parser: literal pattern must contain literal, found ${literalExpr.type}`);
  }
  return annotatePatternNode(AST.literalPattern(literalExpr), node);
}

function parseStructPattern(node: Node, source: string, ctx: ParseContext): Pattern {
  let structType: Identifier | undefined;
  const typeNode = node.childForFieldName("type");
  if (typeNode) {
    const parts = ctx.parseQualifiedIdentifier(typeNode);
    if (parts.length === 0) {
      throw new MapperError("parser: struct pattern type missing identifier");
    }
    structType = parts[parts.length - 1];
    if (parts.length > 1) {
      structType = annotate(AST.identifier(identifiersToStrings(parts).join(".")), typeNode) as Identifier;
    }
  }

  const fields: StructPatternField[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    const fieldName = node.fieldNameForChild(i);
    if (fieldName === "type" || (typeNode && sameNode(child, typeNode))) {
      continue;
    }
    let elem: Node | null = child;
    if (child.type === "struct_pattern_element") {
      elem = child.namedChild(0);
      if (!elem) continue;
    }

    if (elem.type === "struct_pattern_field") {
      fields.push(parseStructPatternField(elem, source, ctx));
      continue;
    }

    const pattern = parsePattern(elem, source, ctx);
    fields.push(annotatePatternNode(AST.structPatternField(pattern, undefined, undefined), elem) as StructPatternField);
  }

  const structKind = resolveStructKind(structType, ctx);
  if (structKind === "positional") {
    for (const field of fields) {
      if (field.fieldName) {
        field.fieldName = undefined;
      }
    }
  }

  const isPositional = structKind === "positional" ? true : fields.some((field) => !field.fieldName);

  return annotatePatternNode(AST.structPattern(fields, isPositional, structType), node) as StructPattern;
}

function parseStructPatternField(node: Node, source: string, ctx: ParseContext): StructPatternField {
  if (node.type !== "struct_pattern_field") {
    throw new MapperError("parser: expected struct_pattern_field node");
  }

  let fieldName: Identifier | undefined;
  const nameNode = node.childForFieldName("field");
  if (nameNode) {
    fieldName = parseIdentifier(nameNode, source);
  }

  let binding: Identifier | undefined;
  const bindingNode = node.childForFieldName("binding");
  if (bindingNode) {
    binding = parseIdentifier(bindingNode, source);
  }

  const typeNode = node.childForFieldName("type");
  const typeAnnotation = typeNode ? ctx.parseTypeExpression(typeNode) ?? undefined : undefined;

  const valueNode = node.childForFieldName("value");
  const alias = binding ?? fieldName;

  let pattern: Pattern | undefined;
  if (valueNode) {
    pattern = parsePattern(valueNode, source, ctx);
  } else if (alias) {
    pattern = alias;
  } else {
    pattern = annotatePatternNode(AST.wildcardPattern(), node) as Pattern;
  }

  if (
    pattern &&
    pattern.type === "StructPattern" &&
    !pattern.structType &&
    typeAnnotation &&
    typeAnnotation.type === "SimpleTypeExpression" &&
    typeAnnotation.name
  ) {
    pattern.structType = typeAnnotation.name as Identifier;
  }

  const bindingForField = valueNode && binding ? binding : undefined;

  return annotatePatternNode(AST.structPatternField(pattern, fieldName, bindingForField, typeAnnotation), node) as StructPatternField;
}

function resolveStructKind(structType: Identifier | undefined, ctx: ParseContext): AST.StructDefinition["kind"] | undefined {
  if (!structType) return undefined;
  const direct = ctx.structKinds.get(structType.name);
  if (direct) return direct;
  const parts = structType.name.split(".");
  if (parts.length > 1) {
    const base = parts[parts.length - 1];
    return ctx.structKinds.get(base);
  }
  return undefined;
}

function parseArrayPattern(node: Node, source: string, ctx: ParseContext): Pattern {
  const elements: Pattern[] = [];
  let rest: Pattern | undefined;

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "array_pattern_rest") {
      if (rest) {
        throw new MapperError("parser: multiple array rest patterns");
      }
      rest = parseArrayPatternRest(child, source, ctx);
      continue;
    }
    elements.push(parsePattern(child, source, ctx));
  }

  return annotatePatternNode(AST.arrayPattern(elements, rest), node) as Pattern;
}

function parseArrayPatternRest(node: Node, source: string, ctx: ParseContext): Pattern {
  if (node.namedChildCount === 0) {
    return annotatePatternNode(AST.wildcardPattern(), node) as Pattern;
  }
  return parsePattern(node.namedChild(0), source, ctx);
}

function isLiteralExpression(expr: Expression): expr is Literal {
  switch (expr.type) {
    case "StringLiteral":
    case "IntegerLiteral":
    case "FloatLiteral":
    case "BooleanLiteral":
    case "NilLiteral":
    case "CharLiteral":
    case "ArrayLiteral":
      return true;
    default:
      return false;
  }
}
