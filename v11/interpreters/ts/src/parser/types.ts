import * as AST from "../ast";
import type {
  GenericParameter,
  Identifier,
  InterfaceConstraint,
  TypeExpression,
  WhereClauseConstraint,
} from "../ast";
import {
  annotate,
  annotateTypeExpressionNode,
  firstNamedChild,
  identifiersToStrings,
  inheritMetadata,
  isIgnorableNode,
  MapperError,
  MutableParseContext,
  Node,
  ParseContext,
  parseIdentifier,
  sameNode,
  sliceText,
  withActiveNode,
} from "./shared";

export function registerTypeParsers(ctx: MutableParseContext): void {
  ctx.parseReturnType = withActiveNode((node) => parseReturnType(ctx, node));
  ctx.parseTypeExpression = withActiveNode((node) => parseTypeExpression(ctx, node));
  ctx.parseTypeArgumentList = withActiveNode((node) => parseTypeArgumentList(ctx, node));
  ctx.parseTypeParameters = withActiveNode((node) => parseTypeParameters(ctx, node));
  ctx.parseTypeBoundList = withActiveNode((node) => parseTypeBoundList(ctx, node));
  ctx.parseWhereClause = withActiveNode((node) => parseWhereClause(ctx, node));
}

export function applyGenericType(base: TypeExpression | null, args: TypeExpression[]): TypeExpression | null {
  if (!base) return null;
  if (base.type === "NullableTypeExpression") {
    const inner = applyGenericType(base.innerType, args);
    const result = AST.nullableTypeExpression(inner ?? base.innerType);
    return inheritMetadata(result, base, inner ?? undefined);
  }
  if (base.type === "ResultTypeExpression") {
    const inner = applyGenericType(base.innerType, args);
    const result = AST.resultTypeExpression(inner ?? base.innerType);
    return inheritMetadata(result, base, inner ?? undefined);
  }
  let flattenedBase: TypeExpression = base;
  let flattenedArgs = args;
  if (base.type === "GenericTypeExpression") {
    const collected: TypeExpression[] = [];
    let current: TypeExpression = base;
    while (current.type === "GenericTypeExpression") {
      if (current.arguments && current.arguments.length > 0) {
        collected.unshift(...current.arguments);
      }
      current = current.base;
    }
    flattenedBase = current;
    flattenedArgs = [...collected, ...args];
  }
  const result = AST.genericTypeExpression(flattenedBase, flattenedArgs);
  return inheritMetadata(result, base);
}

export function parseReturnType(ctx: ParseContext, node: Node | null | undefined): TypeExpression | undefined {
  const expr = parseTypeExpression(ctx, node);
  return expr ?? undefined;
}

export function parseTypeExpression(ctx: ParseContext, node: Node | null | undefined): TypeExpression | null {
  if (!node) return null;
  const { source } = ctx;
  switch (node.type) {
    case "return_type":
    case "type_expression":
    case "interface_type_expression":
    case "type_prefix":
    case "interface_type_prefix":
    case "type_atom":
    case "interface_type_atom": {
      if (node.namedChildCount === 0) break;
      const child = firstNamedChild(node);
      if (child && child !== node) {
        const expr = parseTypeExpression(ctx, child);
        if ((node.type === "type_prefix" || node.type === "interface_type_prefix") && expr) {
          let text = sliceText(node, source).trim();
          let result: TypeExpression | null = expr;
          while (text.startsWith("?") || text.startsWith("!")) {
            if (text.startsWith("?")) {
              text = text.slice(1);
              result = result ? AST.nullableTypeExpression(result) : result;
            } else if (text.startsWith("!")) {
              text = text.slice(1);
              result = result ? AST.resultTypeExpression(result) : result;
            } else {
              break;
            }
          }
          return result;
        }
        return expr;
      }
      break;
    }
    case "type_suffix":
    case "interface_type_suffix": {
      if (node.namedChildCount > 1) {
        const base = parseTypeExpression(ctx, node.namedChild(0));
        const args: TypeExpression[] = [];
        for (let i = 1; i < node.namedChildCount; i++) {
          const child = node.namedChild(i);
          if (!child || !child.isNamed || isIgnorableNode(child)) continue;
          if (child.type === "type_arguments") {
            const typeArgs = parseTypeArgumentList(ctx, child);
            if (typeArgs) args.push(...typeArgs);
            continue;
          }
          const arg = parseTypeExpression(ctx, child);
          if (arg) args.push(arg);
        }
        if (base && args.length > 0) {
          return applyGenericType(base, args);
        }
      }
      const child = firstNamedChild(node);
      if (child && child !== node) {
        return parseTypeExpression(ctx, child);
      }
      break;
    }
    case "type_arrow":
    case "interface_type_arrow": {
      if (node.namedChildCount >= 2) {
        const [paramTypes, ok] = parseFunctionParameterTypes(ctx, node.namedChild(0));
        if (ok && paramTypes) {
          const returnExpr = parseTypeExpression(ctx, node.namedChild(1));
          if (returnExpr) {
            return annotateTypeExpressionNode(AST.functionTypeExpression(paramTypes, returnExpr), node);
          }
        }
      }
      const child = firstNamedChild(node);
      if (child && child !== node) {
        return parseTypeExpression(ctx, child);
      }
      break;
    }
    case "type_generic_application": {
      if (node.namedChildCount === 0) break;
      const base = parseTypeExpression(ctx, node.namedChild(0));
      if (!base) return null;
      const args: TypeExpression[] = [];
      for (let i = 1; i < node.namedChildCount; i++) {
        const arg = parseTypeExpression(ctx, node.namedChild(i));
        if (arg) args.push(arg);
      }
      if (args.length === 0) return annotateTypeExpressionNode(base, node);
      return annotateTypeExpressionNode(applyGenericType(base, args) ?? base, node);
    }
    case "type_union":
    case "interface_type_union": {
      const members: TypeExpression[] = [];
      for (let i = 0; i < node.namedChildCount; i++) {
        const child = node.namedChild(i);
        const member = parseTypeExpression(ctx, child);
        if (member) members.push(member);
      }
      if (members.length === 1) return annotateTypeExpressionNode(members[0], node);
      if (members.length > 1) {
        return annotateTypeExpressionNode(AST.unionTypeExpression(members), node);
      }
      break;
    }
    case "type_identifier": {
      const child = firstNamedChild(node);
      if (child && child !== node) {
        return parseTypeExpression(ctx, child);
      }
      break;
    }
    case "wildcard_type":
      return annotateTypeExpressionNode(AST.wildcardTypeExpression(), node);
    case "identifier":
      return annotateTypeExpressionNode(AST.simpleTypeExpression(parseIdentifier(node, source)), node);
    case "qualified_identifier": {
      const parts = ctx.parseQualifiedIdentifier(node);
      if (parts.length === 0) return null;
      if (parts.length === 1) return annotateTypeExpressionNode(AST.simpleTypeExpression(parts[0]), node);
      const name = annotate(AST.identifier(identifiersToStrings(parts).join(".")), node) as Identifier;
      return annotateTypeExpressionNode(AST.simpleTypeExpression(name), node);
    }
    default: {
      const child = firstNamedChild(node);
      if (child && child !== node) {
        const expr = parseTypeExpression(ctx, child);
        if (expr) return expr;
      }
      break;
    }
  }
  const text = sliceText(node, source).trim();
  if (text === "") {
    return null;
  }
  return annotateTypeExpressionNode(AST.simpleTypeExpression(AST.identifier(text.replace(/\s+/g, ""))), node);
}

export function parseTypeArgumentList(ctx: ParseContext, node: Node | null | undefined): TypeExpression[] | null {
  if (!node) return null;
  const args: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    const typeExpr = parseTypeExpression(ctx, child);
    if (!typeExpr) {
      throw new MapperError(`parser: unsupported type argument kind ${child.type}`);
    }
    args.push(typeExpr);
  }
  return args;
}

export function parseFunctionParameterTypes(
  ctx: ParseContext,
  node: Node | null | undefined,
): [TypeExpression[] | null, boolean] {
  if (!node) {
    return [null, false];
  }

  let current: Node | null = node;
  while (current) {
    if (current.type === "parenthesized_type") {
      if (current.namedChildCount === 0) {
        return [[], true];
      }
      const params: TypeExpression[] = [];
      for (let i = 0; i < current.namedChildCount; i++) {
        const child = current.namedChild(i);
        if (!child || isIgnorableNode(child)) continue;
        const param = parseTypeExpression(ctx, child);
        if (!param) return [null, false];
        params.push(param);
      }
      return [params, true];
    }
    if (current.type !== "type_suffix" && current.type !== "type_prefix" && current.type !== "type_atom") {
      break;
    }
    const child = firstNamedChild(current);
    if (child && child !== current) {
      current = child;
      continue;
    }
    break;
  }

  const param = parseTypeExpression(ctx, node);
  if (!param) {
    return [null, false];
  }
  return [[param], true];
}

export function parseTypeParameters(ctx: ParseContext, node: Node | null | undefined): GenericParameter[] | undefined {
  if (!node) return undefined;
  const { source } = ctx;
  switch (node.type) {
    case "declaration_type_parameters":
      if (node.namedChildCount === 0) return undefined;
      return parseTypeParameters(ctx, node.namedChild(0));
    case "type_parameter_list": {
      const params: GenericParameter[] = [];
      for (let i = 0; i < node.namedChildCount; i++) {
        const child = node.namedChild(i);
        if (!child || isIgnorableNode(child) || child.type !== "type_parameter") continue;
        params.push(parseTypeParameter(ctx, child));
      }
      return params.length ? params : undefined;
    }
    case "generic_parameter_list": {
      const params: GenericParameter[] = [];
      for (let i = 0; i < node.namedChildCount; i++) {
        const child = node.namedChild(i);
        if (!child || isIgnorableNode(child) || child.type !== "generic_parameter") continue;
        params.push(parseGenericParameter(ctx, child));
      }
      return params.length ? params : undefined;
    }
    default:
      throw new MapperError(`parser: unsupported type parameter node ${node.type}`);
  }
}

export function parseTypeParameter(ctx: ParseContext, node: Node): GenericParameter {
  if (node.type !== "type_parameter") {
    throw new MapperError("parser: expected type_parameter node");
  }
  return buildGenericParameter(ctx, node);
}

export function parseGenericParameter(ctx: ParseContext, node: Node): GenericParameter {
  if (node.type !== "generic_parameter") {
    throw new MapperError("parser: expected generic_parameter node");
  }
  return buildGenericParameter(ctx, node);
}

function buildGenericParameter(ctx: ParseContext, node: Node): GenericParameter {
  const { source } = ctx;
  let nameNode: Node | null = firstNamedChild(node);
  if (!nameNode || nameNode.type !== "identifier") {
    throw new MapperError("parser: generic parameter missing identifier");
  }
  const name = parseIdentifier(nameNode, source);

  let constraints: InterfaceConstraint[] | undefined;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child) || sameNode(child, nameNode)) continue;
    const typeExprs = parseTypeBoundList(ctx, child);
    constraints = typeExprs?.map(expr => inheritMetadata(AST.interfaceConstraint(expr), expr));
    if (constraints && constraints.length > 0) break;
  }

  return annotate(AST.genericParameter(name, constraints), node) as GenericParameter;
}

export function parseTypeBoundList(ctx: ParseContext, node: Node | null | undefined): TypeExpression[] | undefined {
  if (!node) return undefined;
  const bounds: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "type_bound_list") {
      const nested = parseTypeBoundList(ctx, child);
      if (nested) bounds.push(...nested);
      continue;
    }
    const expr = parseReturnType(ctx, child);
    if (expr) bounds.push(expr);
  }
  if (bounds.length === 0) {
    throw new MapperError("parser: empty type bound list");
  }
  return bounds;
}

export function parseWhereClause(ctx: ParseContext, node: Node | null | undefined): WhereClauseConstraint[] | undefined {
  if (!node) return undefined;
  if (node.type !== "where_clause") {
    throw new MapperError(`parser: expected where clause, found ${node.type}`);
  }
  const constraints: WhereClauseConstraint[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child) || child.type !== "where_constraint") continue;
    constraints.push(parseWhereConstraint(ctx, child));
  }
  return constraints.length ? constraints : undefined;
}

export function parseWhereConstraint(ctx: ParseContext, node: Node): WhereClauseConstraint {
  if (node.type !== "where_constraint") {
    throw new MapperError("parser: expected where_constraint node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty where constraint");
  }
  const { source } = ctx;
  const nameNode = firstNamedChild(node);
  if (!nameNode || nameNode.type !== "identifier") {
    throw new MapperError("parser: where constraint missing identifier");
  }
  const name = parseIdentifier(nameNode, source);
  let constraintNode: Node | null = null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child) || sameNode(child, nameNode)) continue;
    constraintNode = child;
    break;
  }
  const typeExprs = parseTypeBoundList(ctx, constraintNode);
  if (!typeExprs || typeExprs.length === 0) {
    throw new MapperError("parser: where constraint missing bounds");
  }
  const interfaceConstraints = typeExprs.map(expr => inheritMetadata(AST.interfaceConstraint(expr), expr));
  return annotate(AST.whereClauseConstraint(name, interfaceConstraints), node) as WhereClauseConstraint;
}
