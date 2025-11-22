import type * as AST from "../../ast";
import { unknownType, type TypeInfo } from "../types";
import type { PackageSummary } from "../diagnostics";
import type { Environment } from "../environment";

type ImportContext = {
  env: Environment;
  packageSummaries: Map<string, PackageSummary>;
  packageAliases: Map<string, string>;
  reportedPackageMemberAccess: WeakSet<AST.MemberAccessExpression>;
  currentPackageName: string;
  report(message: string, node?: AST.Node | null): void;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
};

export function applyImports(ctx: ImportContext, module: AST.Module): void {
  if (!module || !Array.isArray(module.imports) || module.imports.length === 0) {
    return;
  }
  for (const imp of module.imports) {
    applyImportStatement(ctx, imp);
  }
}

export function applyImportStatement(ctx: ImportContext, imp: AST.ImportStatement | null | undefined): void {
  if (!imp) {
    return;
  }
  const packageName = formatImportPath(ctx, imp.packagePath);
  const summary = packageName ? ctx.packageSummaries.get(packageName) : undefined;
  if (!summary) {
    const label = packageName ?? "<unknown>";
    ctx.report(`typechecker: import references unknown package '${label}'`, imp);
    return;
  }
  if (summary.visibility === "private" && summary.name !== ctx.currentPackageName) {
    ctx.report(`typechecker: package '${summary.name}' is private`, imp);
    return;
  }
  if (imp.isWildcard) {
    if (summary.symbols) {
      for (const symbolName of Object.keys(summary.symbols)) {
        if (!ctx.env.has(symbolName)) {
          ctx.env.define(symbolName, unknownType);
        }
      }
    }
    return;
  }
  if (Array.isArray(imp.selectors) && imp.selectors.length > 0) {
    for (const selector of imp.selectors) {
      if (!selector) continue;
      const selectorName = ctx.getIdentifierName(selector.name);
      if (!selectorName) continue;
      const aliasName = ctx.getIdentifierName(selector.alias) ?? selectorName;
      const hasSymbol = !!summary.symbols?.[selectorName];
      if (!hasSymbol) {
        if (summary.privateSymbols?.[selectorName]) {
          const label = packageName ?? "<unknown>";
          ctx.report(`typechecker: package '${label}' symbol '${selectorName}' is private`, selector);
        } else {
          const label = packageName ?? "<unknown>";
          ctx.report(`typechecker: package '${label}' has no symbol '${selectorName}'`, selector);
        }
        continue;
      }
      if (!ctx.env.has(aliasName)) {
        ctx.env.define(aliasName, unknownType);
      }
    }
    return;
  }
  const aliasName = ctx.getIdentifierName(imp.alias) ?? defaultPackageAlias(packageName);
  if (!aliasName) {
    return;
  }
  if (packageName) {
    ctx.packageAliases.set(aliasName, packageName);
  }
  if (!ctx.env.has(aliasName)) {
    ctx.env.define(aliasName, unknownType);
  }
}

export function applyDynImportStatement(ctx: ImportContext, statement: AST.DynImportStatement | null | undefined): void {
  if (!statement) {
    return;
  }
  const placeholder: TypeInfo = unknownType;
  if (statement.isWildcard) {
    return;
  }
  if (Array.isArray(statement.selectors) && statement.selectors.length > 0) {
    for (const selector of statement.selectors) {
      if (!selector) continue;
      const selectorName = ctx.getIdentifierName(selector.name);
      if (!selectorName) continue;
      const aliasName = ctx.getIdentifierName(selector.alias) ?? selectorName;
      if (!ctx.env.has(aliasName)) {
        ctx.env.define(aliasName, placeholder);
      }
    }
    return;
  }
  const aliasName = ctx.getIdentifierName(statement.alias);
  if (aliasName && !ctx.env.has(aliasName)) {
    ctx.env.define(aliasName, placeholder);
  }
}

export function handlePackageMemberAccess(
  ctx: ImportContext,
  expression: AST.MemberAccessExpression,
): boolean {
  if (!expression.object || expression.object.type !== "Identifier") {
    return false;
  }
  const aliasName = expression.object.name;
  if (!aliasName) {
    return false;
  }
  const packageName = ctx.packageAliases.get(aliasName);
  if (!packageName) {
    return false;
  }
  const summary = ctx.packageSummaries.get(packageName);
  const memberName = ctx.getIdentifierName(expression.member as AST.Identifier);
  if (!memberName) {
    return true;
  }
  if (!summary?.symbols || !summary.symbols[memberName]) {
    if (!ctx.reportedPackageMemberAccess.has(expression)) {
      ctx.report(
        `typechecker: package '${packageName}' has no symbol '${memberName}'`,
        expression.member ?? expression,
      );
      ctx.reportedPackageMemberAccess.add(expression);
    }
  }
  return true;
}

export function clonePackageSummaries(
  summaries?: Map<string, PackageSummary> | Record<string, PackageSummary>,
): Map<string, PackageSummary> {
  if (!summaries) {
    return new Map();
  }
  if (summaries instanceof Map) {
    return new Map(summaries);
  }
  return new Map(Object.entries(summaries));
}

function formatImportPath(ctx: ImportContext, path: AST.Identifier[] | null | undefined): string | null {
  if (!Array.isArray(path) || path.length === 0) {
    return null;
  }
  const segments = path
    .map((segment) => ctx.getIdentifierName(segment))
    .filter((name): name is string => Boolean(name));
  if (segments.length === 0) {
    return null;
  }
  return segments.join(".");
}

function defaultPackageAlias(packageName: string | null): string | null {
  if (!packageName) {
    return null;
  }
  const segments = packageName.split(".");
  if (segments.length === 0) {
    return null;
  }
  return segments[segments.length - 1] ?? null;
}
