import type { SyntaxNode } from "web-tree-sitter";

import * as AST from "../ast";
import type {
  BlockExpression,
  ExternFunctionBody,
  Expression,
  FunctionDefinition,
  GenericParameter,
  Identifier,
  ImplementationDefinition,
  ImportSelector,
  InterfaceDefinition,
  MethodsDefinition,
  Pattern,
  PreludeStatement,
  Statement,
  StructDefinition,
  TypeAliasDefinition,
  TypeExpression,
  UnionDefinition,
  WhereClauseConstraint,
} from "../ast";

export type Node = SyntaxNode;

export class MapperError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "MapperError";
  }
}

interface MutableAstNode extends AST.AstNode {
  span?: AST.Span;
  origin?: string;
}

let CURRENT_ORIGIN: string | undefined;
let ACTIVE_CONTEXT: ParseContext | undefined;

export function setMapperOrigin(origin: string | undefined): void {
  CURRENT_ORIGIN = origin;
}

export function clearMapperOrigin(): void {
  CURRENT_ORIGIN = undefined;
}

export function toSpan(node: Node): AST.Span {
  const start = node.startPosition;
  const end = node.endPosition;
  return {
    start: { line: start.row + 1, column: start.column + 1 },
    end: { line: end.row + 1, column: end.column + 1 },
  };
}

export function annotate<T extends AST.AstNode | null | undefined>(value: T, tsNode: Node | null | undefined): T {
  if (!value || !tsNode) {
    return value;
  }
  const mutable = value as MutableAstNode;
  if (!mutable.span) {
    mutable.span = toSpan(tsNode);
  }
  if (CURRENT_ORIGIN && !mutable.origin) {
    mutable.origin = CURRENT_ORIGIN;
  }
  return value;
}

export function annotateStatement<T extends Statement | null | undefined>(stmt: T, node: Node | null | undefined): T {
  return annotate(stmt, node) as T;
}

export function annotateExpressionNode<T extends Expression | null | undefined>(expr: T, node: Node | null | undefined): T {
  return annotate(expr, node) as T;
}

export function annotatePatternNode<T extends Pattern | null | undefined>(pattern: T, node: Node | null | undefined): T {
  return annotate(pattern, node) as T;
}

export function annotateTypeExpressionNode<T extends TypeExpression | null | undefined>(
  typeExpr: T,
  node: Node | null | undefined,
): T {
  return annotate(typeExpr, node) as T;
}

export function inheritMetadata<T extends AST.AstNode | null | undefined>(
  value: T,
  ...sources: (AST.AstNode | null | undefined)[]
): T {
  if (!value) {
    return value;
  }
  const target = value as MutableAstNode;
  if (!target.span) {
    for (const source of sources) {
      if (!source) continue;
      const src = source as MutableAstNode;
      if (src.span) {
        target.span = src.span;
        break;
      }
    }
  }
  if (!target.origin) {
    for (const source of sources) {
      if (!source) continue;
      const src = source as MutableAstNode;
      if (src.origin) {
        target.origin = src.origin;
        break;
      }
    }
  }
  if (CURRENT_ORIGIN && !target.origin) {
    target.origin = CURRENT_ORIGIN;
  }
  return value;
}

export function sliceText(node: Node | null | undefined, source: string): string {
  if (!node) return "";
  const start = node.startIndex;
  const end = node.endIndex;
  if (start < 0 || end < start || end > source.length) {
    return "";
  }
  return source.slice(start, end);
}

export function firstNamedChild(node: Node | null | undefined): Node | null {
  if (!node) return null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    return child;
  }
  return null;
}

export function nextNamedSibling(parent: Node | null | undefined, currentIndex: number): Node | null {
  if (!parent) return null;
  for (let i = currentIndex + 1; i < parent.namedChildCount; i++) {
    const sibling = parent.namedChild(i);
    if (!sibling || !sibling.isNamed || isIgnorableNode(sibling)) continue;
    return sibling;
  }
  return null;
}

export function hasLeadingPrivate(node: Node | null | undefined): boolean {
  if (!node) return false;
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "private") {
      return true;
    }
    if (child.isNamed) break;
  }
  return false;
}

export function sameNode(a: Node | null | undefined, b: Node | null | undefined): boolean {
  if (!a || !b) return false;
  return a.type === b.type && a.startIndex === b.startIndex && a.endIndex === b.endIndex;
}

export function isIgnorableNode(node: Node | null | undefined): boolean {
  if (!node) return false;
  switch (node.type) {
    case "comment":
    case "line_comment":
    case "block_comment":
      return true;
    default:
      return false;
  }
}

export function findIdentifier(node: Node | null | undefined, source: string): Identifier | null {
  if (!node) return null;
  if (isIgnorableNode(node)) return null;
  if (node.type === "identifier") {
    return annotate(AST.identifier(sliceText(node, source)), node);
  }
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "identifier") {
      return annotate(AST.identifier(sliceText(child, source)), child);
    }
    const nested = findIdentifier(child, source);
    if (nested) return nested;
  }
  return null;
}

export function parseIdentifier(node: Node | null | undefined, source: string): Identifier {
  if (!node || node.type !== "identifier") {
    throw new MapperError("parser: expected identifier node");
  }
  const id = annotate(AST.identifier(sliceText(node, source)), node);
  if (!id) {
    throw new MapperError("parser: failed to build identifier");
  }
  return id;
}

export function identifiersToStrings(ids: Identifier[]): string[] {
  return ids.map(id => id.name);
}

export function pruneUndefined<T>(value: T): T {
  if (Array.isArray(value)) {
    for (let i = 0; i < value.length; i++) {
      value[i] = pruneUndefined(value[i]);
    }
    return value;
  }
  if (value && typeof value === "object") {
    const record = value as Record<string, unknown>;
    for (const key of Object.keys(record)) {
      const entry = record[key];
      if (entry === undefined) {
        delete record[key];
      } else {
        record[key] = pruneUndefined(entry);
      }
    }
    return value;
  }
  return value;
}

type ContextFns = {
  parseExpression: (node: Node | null | undefined) => Expression;
  parsePattern: (node: Node | null | undefined) => Pattern;
  parseBlock: (node: Node | null | undefined) => BlockExpression;
  parseStatement: (node: Node) => Statement | null;
  parseTypeExpression: (node: Node | null | undefined) => TypeExpression | null;
  parseReturnType: (node: Node | null | undefined) => TypeExpression | undefined;
  parseTypeArgumentList: (node: Node | null | undefined) => TypeExpression[] | null;
  parseTypeParameters: (node: Node | null | undefined) => GenericParameter[] | undefined;
  parseTypeBoundList: (node: Node | null | undefined) => TypeExpression[] | undefined;
  parseWhereClause: (node: Node | null | undefined) => WhereClauseConstraint[] | undefined;
  parseQualifiedIdentifier: (node: Node | null | undefined) => Identifier[];
  parseImportClause: (
    node: Node | null | undefined,
  ) => { isWildcard: boolean; selectors?: ImportSelector[] };
  parsePackageStatement: (node: Node) => AST.PackageStatement;
  parseFunctionDefinition: (node: Node) => FunctionDefinition;
  parseStructDefinition: (node: Node) => StructDefinition;
  parseMethodsDefinition: (node: Node) => MethodsDefinition;
  parseImplementationDefinition: (node: Node) => ImplementationDefinition;
  parseNamedImplementationDefinition: (node: Node) => ImplementationDefinition;
  parseUnionDefinition: (node: Node) => UnionDefinition;
  parseInterfaceDefinition: (node: Node) => InterfaceDefinition;
  parseTypeAliasDefinition: (node: Node) => TypeAliasDefinition;
  parsePreludeStatement: (node: Node) => PreludeStatement;
  parseExternFunction: (node: Node) => ExternFunctionBody;
};

export interface ParseContext extends ContextFns {
  readonly source: string;
  readonly structKinds: Map<string, AST.StructDefinition["kind"]>;
}

export type MutableParseContext = ContextFns & { source: string; structKinds: Map<string, AST.StructDefinition["kind"]> };

export function createParseContext(source: string): MutableParseContext {
  const uninitialized = (name: string) => (): never => {
    throw new MapperError(`parser: ${name} has not been configured on the ParseContext`);
  };
  return {
    source,
    structKinds: new Map(),
    parseExpression: uninitialized("parseExpression"),
    parsePattern: uninitialized("parsePattern"),
    parseBlock: uninitialized("parseBlock"),
    parseStatement: uninitialized("parseStatement"),
    parseTypeExpression: uninitialized("parseTypeExpression"),
    parseReturnType: uninitialized("parseReturnType"),
    parseTypeArgumentList: uninitialized("parseTypeArgumentList"),
    parseTypeParameters: uninitialized("parseTypeParameters"),
    parseTypeBoundList: uninitialized("parseTypeBoundList"),
    parseWhereClause: uninitialized("parseWhereClause"),
    parseQualifiedIdentifier: uninitialized("parseQualifiedIdentifier"),
    parseImportClause: uninitialized("parseImportClause"),
    parsePackageStatement: uninitialized("parsePackageStatement"),
    parseFunctionDefinition: uninitialized("parseFunctionDefinition"),
    parseStructDefinition: uninitialized("parseStructDefinition"),
    parseMethodsDefinition: uninitialized("parseMethodsDefinition"),
    parseImplementationDefinition: uninitialized("parseImplementationDefinition"),
    parseNamedImplementationDefinition: uninitialized("parseNamedImplementationDefinition"),
    parseUnionDefinition: uninitialized("parseUnionDefinition"),
    parseInterfaceDefinition: uninitialized("parseInterfaceDefinition"),
    parseTypeAliasDefinition: uninitialized("parseTypeAliasDefinition"),
    parsePreludeStatement: uninitialized("parsePreludeStatement"),
    parseExternFunction: uninitialized("parseExternFunction"),
  };
}

export function setActiveParseContext(ctx: ParseContext | undefined): void {
  ACTIVE_CONTEXT = ctx;
}

export function getActiveParseContext(): ParseContext {
  if (!ACTIVE_CONTEXT) {
    throw new MapperError("parser: ParseContext requested outside of mapSourceFile");
  }
  return ACTIVE_CONTEXT;
}

export function parseLabel(node: Node, source: string): Identifier {
  if (node.type !== "label") {
    throw new MapperError("parser: expected label");
  }
  let content = sliceText(node, source).trim();
  if (content.startsWith("'")) {
    content = content.slice(1);
  }
  if (!content) {
    throw new MapperError("parser: empty label");
  }
  return annotate(AST.identifier(content), node) as Identifier;
}
