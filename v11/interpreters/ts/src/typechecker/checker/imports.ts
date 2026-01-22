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
  registerImportedSymbol?: (name: string, packageName: string) => void;
  registerImportedTypeSymbol?: (name: string, packageName: string, canonicalName?: string) => void;
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
  const reexports: Record<string, { pkg: string; symbol: string }> = {
    "able.collections.array.Array": { pkg: "able.kernel", symbol: "Array" },
    "able.collections.range.Range": { pkg: "able.kernel", symbol: "Range" },
    "able.collections.range.RangeFactory": { pkg: "able.kernel", symbol: "RangeFactory" },
    "able.core.numeric.Ratio": { pkg: "able.kernel", symbol: "Ratio" },
    "able.concurrency.Channel": { pkg: "able.kernel", symbol: "Channel" },
    "able.concurrency.Mutex": { pkg: "able.kernel", symbol: "Mutex" },
    "able.concurrency.Awaitable": { pkg: "able.kernel", symbol: "Awaitable" },
    "able.concurrency.AwaitWaker": { pkg: "able.kernel", symbol: "AwaitWaker" },
    "able.concurrency.AwaitRegistration": { pkg: "able.kernel", symbol: "AwaitRegistration" },
  };
  const isTypeSymbol = (summaryInfo: PackageSummary | undefined, symbolName: string): boolean => {
    if (!summaryInfo) return false;
    if (summaryInfo.structs?.[symbolName]) return true;
    if (summaryInfo.unions?.[symbolName]) return true;
    if (summaryInfo.interfaces?.[symbolName]) return true;
    if (summaryInfo.functions?.[symbolName]) return false;
    return Boolean(summaryInfo.symbols?.[symbolName]);
  };
  const resolveImportedMeta = (
    symbolName: string,
  ): { originPackage: string | null; isTypeSymbol: boolean } => {
    const hasSymbol = !!summary?.symbols?.[symbolName];
    const reexportKey = packageName ? `${packageName}.${symbolName}` : symbolName;
    const fallback = reexportKey ? reexports[reexportKey] : undefined;
    const fallbackSummary = fallback ? ctx.packageSummaries.get(fallback.pkg) : undefined;
    const hasFallback = !!fallbackSummary?.symbols?.[fallback?.symbol ?? ""];
    if (hasSymbol) {
      return { originPackage: packageName ?? null, isTypeSymbol: isTypeSymbol(summary, symbolName) };
    }
    if (hasFallback && fallback) {
      return { originPackage: fallback.pkg, isTypeSymbol: isTypeSymbol(fallbackSummary, fallback.symbol) };
    }
    return { originPackage: packageName ?? null, isTypeSymbol: false };
  };
  const resolveImported = (
    symbolName: string,
  ): { type: TypeInfo; hasSymbol: boolean; hasFallback: boolean } => {
    const hasSymbol = !!summary?.symbols?.[symbolName];
    const resolved = hasSymbol && summary ? resolveImportedSymbolType(summary, symbolName) : unknownType;
    const reexportKey = packageName ? `${packageName}.${symbolName}` : symbolName;
    const fallback = reexportKey ? reexports[reexportKey] : undefined;
    const fallbackSummary = fallback ? ctx.packageSummaries.get(fallback.pkg) : undefined;
    const hasFallback = !!fallbackSummary?.symbols?.[fallback?.symbol ?? ""];
    let typeInfo: TypeInfo = resolved ?? unknownType;
    if ((typeInfo === unknownType || typeInfo?.kind === "unknown") && fallbackSummary && fallback) {
      const fallbackType = resolveImportedSymbolType(fallbackSummary, fallback.symbol);
      if (fallbackType) {
        typeInfo = fallbackType;
      }
    }
    return { type: typeInfo ?? unknownType, hasSymbol, hasFallback };
  };
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
        const alreadyDefined = ctx.env.has(symbolName);
        if (!alreadyDefined) {
          const resolved = resolveImported(symbolName);
          ctx.env.define(symbolName, resolved.type);
        }
        if (packageName) {
          ctx.registerImportedSymbol?.(symbolName, packageName);
        }
        const meta = resolveImportedMeta(symbolName);
        if (meta.isTypeSymbol && meta.originPackage) {
          ctx.registerImportedTypeSymbol?.(symbolName, meta.originPackage, symbolName);
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
      const resolved = resolveImported(selectorName);
      if (!resolved.hasSymbol && !resolved.hasFallback) {
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
        ctx.env.define(aliasName, resolved.type);
      }
      if (packageName) {
        ctx.registerImportedSymbol?.(aliasName, packageName);
      }
      const meta = resolveImportedMeta(selectorName);
      if (meta.isTypeSymbol && meta.originPackage) {
        ctx.registerImportedTypeSymbol?.(aliasName, meta.originPackage, selectorName);
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
): TypeInfo | null {
  if (!expression.object || expression.object.type !== "Identifier") {
    return null;
  }
  const aliasName = expression.object.name;
  if (!aliasName) {
    return null;
  }
  const packageName = ctx.packageAliases.get(aliasName);
  if (!packageName) {
    return null;
  }
  if (ctx.env.hasBindingOutsideGlobal(aliasName)) {
    return null;
  }
  const summary = ctx.packageSummaries.get(packageName);
  const memberName = ctx.getIdentifierName(expression.member as AST.Identifier);
  if (!memberName) {
    return unknownType;
  }
  const symbolType = summary?.symbolTypes?.[memberName];
  if (symbolType) {
    return symbolType;
  }
  if (!summary?.symbols || !summary.symbols[memberName]) {
    if (!ctx.reportedPackageMemberAccess.has(expression)) {
      ctx.report(
        `typechecker: package '${packageName}' has no symbol '${memberName}'`,
        expression.member ?? expression,
      );
      ctx.reportedPackageMemberAccess.add(expression);
    }
    return unknownType;
  }
  return resolveImportedSymbolType(summary, memberName);
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

function resolveImportedSymbolType(summary: PackageSummary, symbolName: string): TypeInfo {
  const symbolType = summary.symbolTypes?.[symbolName];
  if (symbolType) {
    return symbolType;
  }
  const struct = summary.structs?.[symbolName];
  if (struct) {
    const paramCount = Array.isArray(struct.typeParams) ? struct.typeParams.length : 0;
    const typeArguments = paramCount > 0 ? Array.from({ length: paramCount }, () => unknownType) : [];
    return { kind: "struct", name: symbolName, typeArguments };
  }
  const union = summary.unions?.[symbolName];
  if (union) {
    const members = Array.isArray(union.variants)
      ? union.variants.map((variant) => {
          const label = (variant ?? "").toString().trim();
          if (/^[A-Za-z_][A-Za-z0-9_]*$/.test(label)) {
            return { kind: "struct", name: label, typeArguments: [] };
          }
          return unknownType;
        })
      : [];
    return { kind: "union", members };
  }
  const iface = summary.interfaces?.[symbolName];
  if (iface) {
    const paramCount = Array.isArray(iface.typeParams) ? iface.typeParams.length : 0;
    const typeArguments = paramCount > 0 ? Array.from({ length: paramCount }, () => unknownType) : [];
    return { kind: "interface", name: symbolName, typeArguments };
  }
  return unknownType;
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
