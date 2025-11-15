import * as AST from "../ast";
import type { Expression, Identifier, StructFieldInitializer } from "../ast";
import {
  annotate,
  annotateExpressionNode,
  firstNamedChild,
  identifiersToStrings,
  isIgnorableNode,
  MapperError,
  Node,
  ParseContext,
  parseIdentifier,
  sameNode,
  sliceText,
} from "./shared";

export function parseNumberLiteral(ctx: ParseContext, node: Node): Expression {
  const source = ctx.source;
  const content = sliceText(node, source);
  if (!content) {
    throw new MapperError("parser: empty number literal");
  }

  let base = content;
  let integerType: AST.IntegerLiteral["integerType"] | undefined;
  let floatType: AST.FloatLiteral["floatType"] | undefined;

  const underscoreIdx = content.lastIndexOf("_");
  if (underscoreIdx > 0) {
    const suffix = content.slice(underscoreIdx + 1);
    if (isNumericSuffix(suffix)) {
      base = content.slice(0, underscoreIdx);
      if (suffix === "f32" || suffix === "f64") {
        floatType = suffix as AST.FloatLiteral["floatType"];
      } else {
        integerType = suffix as AST.IntegerLiteral["integerType"];
      }
    }
  }

  const sanitized = base.replace(/_/g, "");
  if (base.includes(".") || /[eE]/.test(base) || floatType) {
    const value = Number(sanitized);
    if (!Number.isFinite(value)) {
      throw new MapperError(`parser: invalid number literal ${content}`);
    }
    return annotateExpressionNode(AST.floatLiteral(value, floatType), node);
  }

  let numberValue: number | bigint;
  try {
    numberValue = BigInt(sanitized);
  } catch {
    const value = Number(sanitized);
    if (!Number.isFinite(value)) {
      throw new MapperError(`parser: invalid number literal ${content}`);
    }
    numberValue = value;
  }

  if (typeof numberValue === "bigint") {
    if (numberValue <= BigInt(Number.MAX_SAFE_INTEGER) && numberValue >= BigInt(Number.MIN_SAFE_INTEGER)) {
      numberValue = Number(numberValue);
    }
  }

  return annotateExpressionNode(AST.integerLiteral(numberValue, integerType), node);
}

export function parseBooleanLiteral(ctx: ParseContext, node: Node): Expression {
  const value = sliceText(node, ctx.source).trim();
  if (value === "true") return annotateExpressionNode(AST.booleanLiteral(true), node);
  if (value === "false") return annotateExpressionNode(AST.booleanLiteral(false), node);
  throw new MapperError(`parser: invalid boolean literal ${value}`);
}

export function parseNilLiteral(ctx: ParseContext, node: Node): Expression {
  const value = sliceText(node, ctx.source).trim();
  if (value !== "nil") {
    throw new MapperError(`parser: invalid nil literal ${value}`);
  }
  return annotateExpressionNode(AST.nilLiteral(), node);
}

export function parseStringLiteral(ctx: ParseContext, node: Node): Expression {
  const raw = sliceText(node, ctx.source);
  try {
    return annotateExpressionNode(AST.stringLiteral(JSON.parse(raw)), node);
  } catch (error) {
    throw new MapperError(`parser: invalid string literal ${raw}: ${error}`);
  }
}

export function parseCharLiteral(ctx: ParseContext, node: Node): Expression {
  const raw = sliceText(node, ctx.source);
  let value: string;
  try {
    value = JSON.parse(raw.replace(/^'|'+$/g, match => (match === "'" ? '"' : match)));
  } catch (error) {
    throw new MapperError(`parser: invalid character literal ${raw}: ${error}`);
  }
  if (Array.from(value).length !== 1) {
    throw new MapperError(`parser: character literal ${raw} must resolve to a single rune`);
  }
  return annotateExpressionNode(AST.charLiteral(value), node);
}

export function parseArrayLiteral(ctx: ParseContext, node: Node): Expression {
  const elements: Expression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    elements.push(ctx.parseExpression(child));
  }
  return annotateExpressionNode(AST.arrayLiteral(elements), node);
}

export function parseStructLiteral(ctx: ParseContext, node: Node): Expression {
  const source = ctx.source;
  const typeNode = node.childForFieldName("type");
  if (!typeNode) {
    throw new MapperError("parser: struct literal missing type");
  }
  const parts = ctx.parseQualifiedIdentifier(typeNode);
  if (parts.length === 0) {
    throw new MapperError("parser: invalid struct literal type");
  }
  let structType = parts[parts.length - 1];
  if (parts.length > 1) {
    structType = annotate(AST.identifier(identifiersToStrings(parts).join(".")), typeNode) as Identifier;
  }

  const typeArgsNode = node.childForFieldName("type_arguments");
  const typeArguments = typeArgsNode ? ctx.parseTypeArgumentList(typeArgsNode) ?? undefined : undefined;

  const fields: StructFieldInitializer[] = [];
  const functionalUpdates: Expression[] = [];

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    const fieldName = node.fieldNameForChild(i);
    if (
      fieldName === "type" ||
      fieldName === "type_arguments" ||
      sameNode(child, typeNode) ||
      (typeArgsNode && sameNode(child, typeArgsNode))
    ) {
      continue;
    }

    let elem: Node | null = child;
    if (child.type === "struct_literal_element") {
      elem = firstNamedChild(child);
      if (!elem) continue;
    }

    switch (elem.type) {
      case "struct_literal_field": {
        const nameNode = elem.childForFieldName("name");
        if (!nameNode) {
          throw new MapperError("parser: struct literal field missing name");
        }
        const name = parseIdentifier(nameNode, source);
        const valueNode = elem.childForFieldName("value");
        if (!valueNode) {
          throw new MapperError("parser: struct literal field missing value");
        }
        const value = ctx.parseExpression(valueNode);
        fields.push(annotateExpressionNode(AST.structFieldInitializer(value, name, false), elem) as StructFieldInitializer);
        break;
      }
      case "struct_literal_shorthand_field": {
        let nameNode = elem.childForFieldName("name");
        if (!nameNode) {
          nameNode = firstNamedChild(elem);
        }
        if (!nameNode) {
          throw new MapperError("parser: struct literal shorthand missing name");
        }
        const name = parseIdentifier(nameNode, source);
        fields.push(
          annotateExpressionNode(AST.structFieldInitializer(AST.identifier(name.name), name, true), elem) as StructFieldInitializer,
        );
        break;
      }
      case "struct_literal_spread": {
        const exprNode = firstNamedChild(elem);
        if (!exprNode) {
          throw new MapperError("parser: struct spread missing expression");
        }
        functionalUpdates.push(ctx.parseExpression(exprNode));
        break;
      }
      default: {
        fields.push(
          annotateExpressionNode(AST.structFieldInitializer(ctx.parseExpression(elem), undefined, false), elem) as StructFieldInitializer,
        );
        break;
      }
    }
  }

  const positional = fields.some(field => !field.name);

  return annotateExpressionNode(
    AST.structLiteral(fields, positional, structType, functionalUpdates.length ? functionalUpdates : undefined, typeArguments ?? undefined),
    node,
  );
}

export function parseMapLiteral(ctx: ParseContext, node: Node): Expression {
  const entries: (AST.MapLiteralEntry | AST.MapLiteralSpread)[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    let elem: Node | null = child;
    if (child.type === "map_literal_element") {
      elem = firstNamedChild(child);
    }
    if (!elem) continue;
    switch (elem.type) {
      case "map_literal_entry": {
        const keyNode = elem.childForFieldName("key") ?? firstNamedChild(elem);
        const valueNode = elem.childForFieldName("value") ?? elem.namedChild(elem.namedChildCount - 1);
        if (!keyNode || !valueNode) {
          throw new MapperError("parser: map literal entry missing key or value");
        }
        const keyExpr = ctx.parseExpression(keyNode);
        const valueExpr = ctx.parseExpression(valueNode);
        entries.push(annotateExpressionNode(AST.mapLiteralEntry(keyExpr, valueExpr), elem) as AST.MapLiteralEntry);
        break;
      }
      case "map_literal_spread": {
        const exprNode = elem.childForFieldName("expression") ?? firstNamedChild(elem);
        if (!exprNode) {
          throw new MapperError("parser: map literal spread missing expression");
        }
        entries.push(annotateExpressionNode(AST.mapLiteralSpread(ctx.parseExpression(exprNode)), elem) as AST.MapLiteralSpread);
        break;
      }
      default:
        throw new MapperError(`parser: unsupported map literal element ${elem.type}`);
    }
  }
  return annotateExpressionNode(AST.mapLiteral(entries), node);
}

export function isNumericSuffix(value: string): boolean {
  switch (value) {
    case "i8":
    case "i16":
    case "i32":
    case "i64":
    case "i128":
    case "u8":
    case "u16":
    case "u32":
    case "u64":
    case "u128":
    case "f32":
    case "f64":
      return true;
    default:
      return false;
  }
}
