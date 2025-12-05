import * as AST from "../ast";
import type { Identifier, ImportSelector } from "../ast";
import {
  annotate,
  MapperError,
  ParseContext,
  MutableParseContext,
  Node,
  isIgnorableNode,
  parseIdentifier,
} from "./shared";

export function registerImportParsers(ctx: MutableParseContext): void {
  ctx.parsePackageStatement = node => parsePackageStatement(ctx, node);
  ctx.parseQualifiedIdentifier = node => parseQualifiedIdentifier(ctx, node);
  ctx.parseImportClause = node => parseImportClause(ctx, node);
}

export function parsePackageStatement(ctx: ParseContext, node: Node): AST.PackageStatement {
  const { source } = ctx;
  const parts: Identifier[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    parts.push(parseIdentifier(child, source));
  }
  if (parts.length === 0) {
    throw new MapperError("parser: empty package statement");
  }
  return annotate(AST.packageStatement(parts, false), node) as AST.PackageStatement;
}

export function parseQualifiedIdentifier(ctx: ParseContext, node: Node | null | undefined): Identifier[] {
  if (!node || (node.type !== "qualified_identifier" && node.type !== "import_path")) {
    throw new MapperError("parser: expected qualified identifier");
  }
  const { source } = ctx;
  const parts: Identifier[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    parts.push(parseIdentifier(child, source));
  }
  if (parts.length === 0) {
    throw new MapperError("parser: empty qualified identifier");
  }
  return parts;
}

export function parseImportClause(
  ctx: ParseContext,
  node: Node | null | undefined,
): {
  isWildcard: boolean;
  selectors?: ImportSelector[];
} {
  if (!node) {
    return { isWildcard: false };
  }

  let isWildcard = false;
  const selectors: ImportSelector[] = [];

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    switch (child.type) {
      case "import_selector":
        selectors.push(parseImportSelector(ctx, child));
        break;
      case "import_wildcard_clause":
        isWildcard = true;
        break;
      default:
        throw new MapperError(`parser: unsupported import clause node ${child.type}`);
    }
  }

  if (isWildcard && selectors.length > 0) {
    throw new MapperError("parser: wildcard import cannot include selectors");
  }

  return { isWildcard, selectors: selectors.length ? selectors : undefined };
}

export function parseImportSelector(ctx: ParseContext, node: Node): ImportSelector {
  if (node.type !== "import_selector") {
    throw new MapperError("parser: expected import_selector node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty import selector");
  }
  const { source } = ctx;
  const name = parseIdentifier(node.namedChild(0), source);
  let alias: Identifier | undefined;
  if (node.namedChildCount > 1) {
    alias = parseIdentifier(node.namedChild(1), source);
  }
  return annotate(AST.importSelector(name, alias), node) as ImportSelector;
}
