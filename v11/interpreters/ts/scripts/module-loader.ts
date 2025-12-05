import { promises as fs } from "node:fs";
import path from "node:path";

import * as AST from "../src/ast";
import type { ImportStatement, Identifier } from "../src/ast";
import {
  annotateModuleOrigin,
  buildPackageSegmentsForModule,
  discoverRoot,
  discoverRootForPath,
  indexSourceFiles,
  parseModuleFromSource,
  type PackageLocation,
} from "./module-utils";
import type { ModuleSearchPath } from "./module-search-paths";

type FileModule = {
  path: string;
  packageName: string;
  ast: AST.Module;
  imports: string[];
};

export type LoadedModule = {
  packageName: string;
  module: AST.Module;
  files: string[];
  imports: string[];
};

export type Program = {
  entry: LoadedModule;
  modules: LoadedModule[];
};

type PackageOrigin = {
  root: string;
  isStdlib: boolean;
};

type RootInfo = {
  rootDir: string;
  rootName: string;
  isStdlib: boolean;
};

type NormalizedSearchPath = {
  path: string;
  isStdlib: boolean;
  source?: string;
};

type LoadOptions = {
  includePackages?: string[];
};

export class ModuleLoader {
  private readonly searchPaths: NormalizedSearchPath[];

  constructor(searchPaths: (ModuleSearchPath | string)[] = []) {
    this.searchPaths = normalizeSearchPaths(searchPaths);
  }

  async load(entryPath: string, options: LoadOptions = {}): Promise<Program> {
    if (!entryPath) {
      throw new Error("loader: empty entry path");
    }
    const resolvedEntry = await this.resolveEntryPath(entryPath);
    const entryInfo = await fs.stat(resolvedEntry);
    if (entryInfo.isDirectory()) {
      throw new Error(`loader: entry path ${resolvedEntry} is a directory`);
    }
    const { rootDir, rootName } = await discoverRoot(resolvedEntry);
    const pkgIndex = new Map<string, PackageLocation>();
    const packageOrigins = new Map<string, PackageOrigin>();
    const entryIsStdlib = this.searchPaths.some((sp) =>
      sp.isStdlib ? pathsOverlap(sp.path, rootDir) || pathsOverlap(rootDir, sp.path) : false,
    );
    const entryRoot: RootInfo = { rootDir, rootName, isStdlib: entryIsStdlib };
    ensureNamespaceAllowed(entryRoot);
    const { packages, fileToPackage } = await indexSourceFiles(rootDir, rootName);
    registerPackages(pkgIndex, packages, entryRoot, packageOrigins);
    await this.indexAdditionalRoots(pkgIndex, packageOrigins, entryRoot);

    const entryPackage = fileToPackage.get(resolvedEntry);
    if (!entryPackage) {
      throw new Error(`loader: failed to resolve package for entry file ${resolvedEntry}`);
    }

    const include = new Set<string>(options.includePackages ? options.includePackages : []);
    include.add(entryPackage);

    const loaded = new Map<string, LoadedModule>();
    const inProgress = new Set<string>();
    const ordered: LoadedModule[] = [];

    const loadPackage = async (pkgName: string): Promise<LoadedModule> => {
      if (loaded.has(pkgName)) {
        return loaded.get(pkgName)!;
      }
      if (inProgress.has(pkgName)) {
        throw new Error(`loader: import cycle detected at package ${pkgName}`);
      }
      const loc = pkgIndex.get(pkgName);
      if (!loc || loc.files.length === 0) {
        throw new Error(`loader: package ${pkgName} not found`);
      }
      inProgress.add(pkgName);
      const fileModules: FileModule[] = [];
      for (const filePath of loc.files) {
        const fm = await parseFile(filePath, loc.rootDir, loc.rootName);
        if (fm.packageName !== pkgName) {
          throw new Error(
            `loader: file ${filePath} resolves to package ${fm.packageName}, expected ${pkgName}`,
      );
    }
    fileModules.push(fm);
  }
  const combined = combinePackage(pkgName, fileModules);
  for (const dep of combined.imports) {
    if (dep === pkgName) continue;
    if (!pkgIndex.has(dep)) {
      throw new Error(`loader: package ${pkgName} imports unknown package ${dep}`);
    }
    await loadPackage(dep);
  }
  loaded.set(pkgName, combined);
  ordered.push(combined);
  inProgress.delete(pkgName);
  return combined;
};

    for (const pkgName of include) {
      await loadPackage(pkgName);
    }

    const entryModule = loaded.get(entryPackage);
    if (!entryModule) {
      throw new Error(`loader: failed to load entry package ${entryPackage}`);
    }

    return { entry: entryModule, modules: ordered };
  }

  private async resolveEntryPath(entry: string): Promise<string> {
    const candidate = path.resolve(entry);
    try {
      await fs.access(candidate);
      return candidate;
    } catch {
      throw new Error(`loader: unable to access ${candidate}`);
    }
  }

  private async indexAdditionalRoots(
    pkgIndex: Map<string, PackageLocation>,
    packageOrigins: Map<string, PackageOrigin>,
    primaryRoot: RootInfo,
  ): Promise<void> {
    if (this.searchPaths.length === 0) {
      return;
    }
    const used = new Set<string>([path.resolve(primaryRoot.rootDir)]);
    for (const searchPath of this.searchPaths) {
      const resolvedSearchPath = path.resolve(searchPath.path);
      let overlaps = false;
      for (const seen of used) {
        if (pathsOverlap(seen, resolvedSearchPath) || pathsOverlap(resolvedSearchPath, seen)) {
          overlaps = true;
          break;
        }
      }
      if (overlaps) {
        continue;
      }
      if (used.has(resolvedSearchPath)) {
        continue;
      }
      let rootInfo: { abs: string; rootName: string };
      try {
        rootInfo = await discoverRootForPath(resolvedSearchPath);
      } catch (error) {
        warnInvalidSearchPath(resolvedSearchPath, error);
        continue;
      }
      const clean = path.resolve(rootInfo.abs);
      if (used.has(clean)) {
        continue;
      }
      used.add(clean);
      const indexedRoot: RootInfo = {
        rootDir: rootInfo.abs,
        rootName: rootInfo.rootName,
        isStdlib: Boolean(searchPath.isStdlib),
      };
      if (!ensureNamespaceAllowed(indexedRoot, true)) {
        continue;
      }
      const { packages } = await indexSourceFiles(indexedRoot.rootDir, indexedRoot.rootName);
      registerPackages(pkgIndex, packages, indexedRoot, packageOrigins);
    }
  }
}

function normalizeSearchPaths(paths: (ModuleSearchPath | string)[]): NormalizedSearchPath[] {
  const uniques = new Map<string, NormalizedSearchPath>();
  for (const entry of paths) {
    if (!entry) continue;
    const normalized =
      typeof entry === "string"
        ? { path: entry, isStdlib: false }
        : { path: entry.path, isStdlib: Boolean(entry.isStdlib), source: entry.source };
    if (!normalized.path) continue;
    const abs = path.resolve(normalized.path);
    if (uniques.has(abs)) continue;
    uniques.set(abs, { ...normalized, path: abs });
  }
  return [...uniques.values()];
}

function pathsOverlap(a: string, b: string): boolean {
  const rel = path.relative(path.resolve(a), path.resolve(b));
  return rel === "" || rel === "." || (!rel.startsWith("..") && !path.isAbsolute(rel));
}

function looksLikeStdlibPath(dir: string): boolean {
  return path
    .resolve(dir)
    .split(path.sep)
    .some((segment) => segment === "stdlib" || segment === "stdlib_v11" || segment === "stdlib_v10");
}

function ensureNamespaceAllowed(root: RootInfo, allowSkip = false): boolean {
  if (root.rootName === "able" && !root.isStdlib) {
    if (looksLikeStdlibPath(root.rootDir)) {
      root.isStdlib = true;
      return true;
    }
    const message = `loader: package namespace 'able.*' is reserved for the standard library (path: ${root.rootDir})`;
    if (allowSkip) {
      return false;
    }
    throw new Error(message);
  }
  return true;
}

function registerPackages(
  pkgIndex: Map<string, PackageLocation>,
  packages: Map<string, PackageLocation>,
  root: RootInfo,
  origins: Map<string, PackageOrigin>,
): void {
  for (const [name, loc] of packages.entries()) {
    if (loc.files.length === 0) continue;
    const existing = origins.get(name);
    if (existing) {
      throw new Error(
        `loader: package ${name} found in multiple roots (${existing.root}, ${root.rootDir})`,
      );
    }
    origins.set(name, { root: root.rootDir, isStdlib: root.isStdlib });
    pkgIndex.set(name, loc);
  }
}

async function parseFile(filePath: string, rootDir: string, rootPackage: string): Promise<FileModule> {
  const moduleAST = await parseModuleFromSource(filePath);
  if (!moduleAST) {
    throw new Error(`loader: failed to parse ${filePath}`);
  }
  const { segments, isPrivate } = buildPackageSegmentsForModule(rootDir, rootPackage, filePath, moduleAST);
  const pkgName = segments.join(".");
  moduleAST.package = AST.packageStatement(segments, isPrivate);
  annotateModuleOrigin(moduleAST, filePath);

  const importSet = new Set<string>();
  for (const imp of (moduleAST.imports || [])) {
    const name = formatImportPath(imp);
    if (!name) {
      continue;
    }
    importSet.add(name);
  }
  collectDynImportPackages(moduleAST, importSet);
  const imports = [...importSet].sort();
  return { path: filePath, packageName: pkgName, ast: moduleAST, imports };
}

function combinePackage(packageName: string, files: FileModule[]): LoadedModule {
  if (files.length === 0) {
    throw new Error("loader: combinePackage called with no files");
  }
  const sortedFiles = [...files].sort((a, b) => a.path.localeCompare(b.path));
  const primaryPath = sortedFiles[0] && sortedFiles[0].path ? sortedFiles[0].path : "";
  const body: AST.Statement[] = [];
  const importNodes: ImportStatement[] = [];
  const importNodeKeys = new Set<string>();
  const importNames = new Set<string>();
  let pkgStmt: AST.PackageStatement | undefined;

  for (const file of sortedFiles) {
    if (file.ast.package && !pkgStmt) {
      pkgStmt = AST.packageStatement(
        file.ast.package.namePath.map((id) => id.name),
        file.ast.package.isPrivate,
      );
    }
    for (const imp of (file.ast.imports || [])) {
      const key = importKey(imp);
      if (key && !importNodeKeys.has(key)) {
        importNodeKeys.add(key);
        importNodes.push(imp);
      }
    }
    for (const name of file.imports) {
      if (name === packageName) continue;
      importNames.add(name);
    }
    body.push(...file.ast.body);
  }
  if (!pkgStmt) {
    const segments = packageName.split(".").filter(Boolean);
    pkgStmt = AST.packageStatement(segments);
  }
  const module = AST.module(body, importNodes, pkgStmt);
  const fallbackOrigin = primaryPath || (files[0] ? files[0].path : "");
  annotateModuleOrigin(module, fallbackOrigin || "<unknown>");
  return {
    packageName,
    module,
    files: sortedFiles.map((f) => f.path),
    imports: [...importNames].sort(),
  };
}


function collectDynImportPackages(module: AST.Module, importSet: Set<string>): void {
  collectDynImportsFromNode(module, importSet, new Set<object>());
}

function collectDynImportsFromNode(node: unknown, imports: Set<string>, seen: Set<object>): void {
  if (!node || typeof node !== "object") return;
  if (seen.has(node as object)) return;
  seen.add(node as object);

  const candidate = node as Partial<AST.AstNode>;
  if (candidate && candidate.type === "DynImportStatement") {
    const dynPath = formatDynImportPath(candidate as AST.DynImportStatement);
    if (dynPath) {
      imports.add(dynPath);
    }
  }

  for (const value of Object.values(node)) {
    if (!value) continue;
    if (typeof value === "object") {
      collectDynImportsFromNode(value, imports, seen);
    }
  }
}

function formatDynImportPath(imp: AST.DynImportStatement): string {
  if (!imp || !Array.isArray(imp.packagePath)) return "";
  return imp.packagePath
    .map((id: Identifier) => (id && id.name ? id.name : ""))
    .filter(Boolean)
    .join(".");
}

function formatImportPath(imp: ImportStatement): string {
  if (!imp || !imp.packagePath || imp.packagePath.length === 0) {
    return "";
  }
  return imp.packagePath
    .map((id: Identifier) => (id && id.name ? id.name : ""))
    .filter(Boolean)
    .join(".");
}

function importKey(imp: ImportStatement): string {
  if (!imp) return "";
  const aliasName = imp.alias && imp.alias.name ? imp.alias.name : "";
  const parts = [formatImportPath(imp), imp.isWildcard ? "*" : "", aliasName];
  if (imp.selectors && imp.selectors.length > 0) {
    const selectorParts = imp.selectors
      .map((sel) => {
        if (!sel?.name?.name) return "";
        return sel.alias?.name ? `${sel.name.name}::${sel.alias.name}` : sel.name.name;
      })
      .filter(Boolean)
      .join(",");
    parts.push(selectorParts);
  }
  return parts.join("|");
}

function warnInvalidSearchPath(searchPath: string, error: unknown): void {
  const reason =
    error instanceof Error ? error.message : typeof error === "string" ? error : String(error);
  console.warn(`loader: skipping search path ${searchPath}: ${reason}`);
}
