import type * as AST from "../../ast";
import { inferFunctionSignatureGenerics, type DeclarationsContext } from "./declarations";
import type { ImplementationRecord } from "./types";

type RegistryContext = {
  structDefinitions: Map<string, AST.StructDefinition>;
  interfaceDefinitions: Map<string, AST.InterfaceDefinition>;
  typeAliases: Map<string, AST.TypeAliasDefinition>;
  implementationRecords: ImplementationRecord[];
  implementationIndex: Map<string, ImplementationRecord[]>;
  declarationOrigins: Map<string, AST.Node>;
  declarationsContext: DeclarationsContext;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  report(message: string, node?: AST.Node | null): void;
};

export function ensureUniqueDeclaration(
  ctx: RegistryContext,
  name: string | null | undefined,
  node: AST.Node | null | undefined,
): boolean {
  if (!name || !node) {
    return true;
  }
  const existing = ctx.declarationOrigins.get(name);
  if (existing) {
    if ((existing as unknown as { _builtin?: boolean })._builtin) {
      ctx.declarationOrigins.set(name, node);
      return true;
    }
    const existingOrigin = (existing as { origin?: string }).origin ?? "";
    const nextOrigin = (node as { origin?: string }).origin ?? "";
    // Allow non-builtin declarations to override kernel bridge symbols so stdlib/package
    // versions win without emitting duplicate errors.
    if (existingOrigin.includes("/kernel/") && existingOrigin !== nextOrigin) {
      ctx.declarationOrigins.set(name, node);
      return true;
    }
    const location = formatNodeOrigin(existing);
    const displayName = name.startsWith("<anonymous>::") ? name.slice("<anonymous>::".length) : name;
    ctx.report(`typechecker: duplicate declaration '${displayName}' (previous declaration at ${location})`, node);
    return false;
  }
  ctx.declarationOrigins.set(name, node);
  return true;
}

export function registerStructDefinition(ctx: RegistryContext, definition: AST.StructDefinition): void {
  const name = definition.id?.name;
  if (name) {
    if (!ensureUniqueDeclaration(ctx, name, definition)) {
      return;
    }
    ctx.structDefinitions.set(name, definition);
  }
}

export function registerInterfaceDefinition(ctx: RegistryContext, definition: AST.InterfaceDefinition): void {
  const name = definition.id?.name;
  if (name) {
    if (!ensureUniqueDeclaration(ctx, name, definition)) {
      return;
    }
    const parentGenerics = Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((paramName): paramName is string => Boolean(paramName))
      : [];
    if (Array.isArray(definition.signatures)) {
      for (const signature of definition.signatures) {
        inferFunctionSignatureGenerics(ctx.declarationsContext, signature, parentGenerics);
      }
    }
    ctx.interfaceDefinitions.set(name, definition);
  }
}

export function registerTypeAlias(ctx: RegistryContext, definition: AST.TypeAliasDefinition): void {
  const name = definition.id?.name;
  if (!name) return;
  if (!ensureUniqueDeclaration(ctx, name, definition)) {
    return;
  }
  ctx.typeAliases.set(name, definition);
}

export function registerImplementationRecord(ctx: RegistryContext, record: ImplementationRecord): void {
  ctx.implementationRecords.push(record);
  const bucket = ctx.implementationIndex.get(record.targetKey);
  if (bucket) {
    bucket.push(record);
  } else {
    ctx.implementationIndex.set(record.targetKey, [record]);
  }
}

export function formatNodeOrigin(node: AST.Node | null | undefined): string {
  if (!node) {
    return "<unknown location>";
  }
  const origin = (node as { origin?: string }).origin ?? "<unknown file>";
  const span = (node as { span?: { start?: { line?: number; column?: number } } }).span;
  const line = span?.start?.line ?? 0;
  const column = span?.start?.column ?? 0;
  return `${origin}:${line}:${column}`;
}
