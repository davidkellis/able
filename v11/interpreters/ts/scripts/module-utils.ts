import { promises as fs } from "node:fs";
import path from "node:path";

import * as AST from "../src/ast";
import { mapSourceFile } from "../src/parser/tree-sitter-mapper";
import { getTreeSitterParser } from "../src/parser/tree-sitter-loader";
import { looksLikeKernelPath } from "./module-search-paths";

export type PackageLocation = {
  rootDir: string;
  rootName: string;
  files: string[];
};

export async function parseModuleFromSource(sourcePath: string): Promise<AST.Module | null> {
  try {
    const source = await fs.readFile(sourcePath, "utf8");
    const parser = await getTreeSitterParser();
    const tree = parser.parse(source);
    if (tree.rootNode.type !== "source_file") {
      throw new Error(`tree-sitter returned unexpected root ${tree.rootNode.type}`);
    }
    if ((tree.rootNode as unknown as { hasError?: boolean }).hasError) {
      throw new Error("tree-sitter reported syntax errors");
    }
    const moduleAst = mapSourceFile(tree.rootNode, source, sourcePath);
    if (moduleAst) {
      repairTypeAliasTargets(moduleAst, source);
    }
    return moduleAst;
  } catch (error) {
    console.error(`Failed to parse ${sourcePath}: ${extractErrorMessage(error)}`);
    return null;
  }
}

function repairTypeAliasTargets(moduleAst: AST.Module, source: string): void {
  const body = moduleAst.body ?? [];
  if (body.length === 0) return;
  const lines = source.split(/\r?\n/);
  const repaired: AST.Statement[] = [];
  for (let i = 0; i < body.length; i += 1) {
    const stmt = body[i];
    if (!stmt || stmt.type !== "TypeAliasDefinition") {
      repaired.push(stmt as AST.Statement);
      continue;
    }
    if (!stmt.targetType || !stmt.genericParams || stmt.genericParams.length === 0) {
      repaired.push(stmt);
      continue;
    }
    const genericNames = new Set(
      stmt.genericParams.map((param) => param?.name?.name).filter((name): name is string => !!name),
    );
    if (genericNames.size === 0 || !stmt.span?.end) {
      repaired.push(stmt);
      continue;
    }
    let targetType = stmt.targetType;
    let end = stmt.span.end;
    let consumed = 0;
    for (let j = i + 1; j < body.length; j += 1) {
      const next = body[j];
      if (!next || next.type !== "Identifier" || !next.span) break;
      if (next.span.start.line !== end.line) break;
      const line = lines[end.line - 1] ?? "";
      const startCol = Math.max(end.column - 1, 0);
      const endCol = Math.max(next.span.start.column - 1, 0);
      const between = line.slice(startCol, endCol);
      if (between.trim() !== "") break;
      if (!genericNames.has(next.name)) break;
      const extraArg = AST.simpleTypeExpression(next);
      if (targetType.type === "GenericTypeExpression") {
        targetType = { ...targetType, arguments: [...targetType.arguments, extraArg] };
      } else {
        targetType = AST.genericTypeExpression(targetType, [extraArg]);
      }
      end = next.span.end;
      consumed += 1;
    }
    if (consumed > 0) {
      stmt.targetType = targetType;
      stmt.span = stmt.span ? { ...stmt.span, end } : stmt.span;
    }
    repaired.push(stmt);
    i += consumed;
  }
  moduleAst.body = repaired;
}

export async function readJsonModule(filePath: string): Promise<AST.Module> {
  const raw = JSON.parse(await fs.readFile(filePath, "utf8"));
  const module = hydrateNode(raw) as AST.Module;
  annotateModuleOrigin(module, filePath);
  return module;
}

export function hydrateNode(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(hydrateNode);
  if (value && typeof value === "object") {
    const node = value as Record<string, unknown>;
    if (typeof node.type === "string") {
      switch (node.type) {
        case "IntegerLiteral":
          if (typeof node.value === "string") node.value = BigInt(node.value);
          break;
        case "FloatLiteral":
          if (typeof node.value === "string") node.value = Number(node.value);
          break;
        case "BooleanLiteral":
          if (typeof node.value === "string") node.value = node.value === "true";
          break;
        case "ArrayLiteral":
          node.elements = hydrateNode(node.elements) as unknown[];
          break;
        case "Module":
          node.imports = hydrateNode(node.imports) as unknown[];
          node.body = hydrateNode(node.body) as unknown[];
          break;
        default:
          for (const [key, val] of Object.entries(node)) {
            node[key] = hydrateNode(val) as never;
          }
          return node;
      }
    }
    for (const [key, val] of Object.entries(node)) {
      if (key !== "type") node[key] = hydrateNode(val) as never;
    }
    return node;
  }
  return value;
}

export function annotateModuleOrigin(node: unknown, origin: string, seen = new Set<object>()): void {
  if (!node || typeof node !== "object") {
    return;
  }
  if (seen.has(node as object)) {
    return;
  }
  seen.add(node as object);

  if (Array.isArray(node)) {
    for (const entry of node) {
      annotateModuleOrigin(entry, origin, seen);
    }
    return;
  }

  const candidate = node as Partial<AST.AstNode>;
  if (typeof candidate.type === "string" && typeof candidate.origin !== "string") {
    candidate.origin = origin;
  }

  for (const value of Object.values(node)) {
    annotateModuleOrigin(value, origin, seen);
  }
}

export async function discoverRoot(entryPath: string): Promise<{ rootDir: string; rootName: string }> {
  let dir = path.dirname(entryPath);
  while (true) {
    const candidate = path.join(dir, "package.yml");
    try {
      const stats = await fs.stat(candidate);
      if (stats.isFile()) {
        const name = await readPackageName(candidate);
        const fallbackName = name !== undefined && name !== null ? name : path.basename(dir) || "pkg";
        const sanitized = sanitizeSegment(fallbackName) || "pkg";
        return { rootDir: dir, rootName: sanitized };
      }
    } catch (err: any) {
      if (err && err.code !== "ENOENT") {
        const message = err && typeof err.message === "string" ? err.message : String(err);
        throw new Error(`loader: stat ${candidate}: ${message}`);
      }
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      break;
    }
    dir = parent;
  }
  const fallbackDir = path.dirname(entryPath);
  const fallbackName = sanitizeSegment(path.basename(fallbackDir) || "pkg") || "pkg";
  return { rootDir: fallbackDir, rootName: fallbackName };
}

export async function discoverRootForPath(searchPath: string): Promise<{ abs: string; rootName: string }> {
  if (!searchPath) {
    throw new Error("loader: empty search path");
  }
  const abs = path.resolve(searchPath);
  const info = await fs.stat(abs);
  if (!info.isDirectory()) {
    throw new Error(`loader: search path ${abs} is not a directory`);
  }
  const manifestName = await findManifestName(abs);
  let rootName = manifestName !== undefined && manifestName !== null ? manifestName : path.basename(abs) || "pkg";
  rootName = sanitizeSegment(rootName) || "pkg";
  return { abs, rootName };
}

export async function findManifestName(start: string): Promise<string | null> {
  let dir = path.resolve(start);
  while (true) {
    const candidate = path.join(dir, "package.yml");
    try {
      const stats = await fs.stat(candidate);
      if (stats.isFile()) {
        const name = await readPackageName(candidate);
        return name !== undefined && name !== null ? name : null;
      }
    } catch (err: any) {
      if (err && err.code !== "ENOENT") {
        const message = typeof err.message === "string" ? err.message : String(err);
        throw new Error(`loader: stat ${candidate}: ${message}`);
      }
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      break;
    }
    dir = parent;
  }
  return null;
}

export async function readPackageName(manifestPath: string): Promise<string | null> {
  try {
    const contents = await fs.readFile(manifestPath, "utf8");
    const lines = contents.split(/\r?\n/);
    for (const rawLine of lines) {
      const line = rawLine.trim();
      if (!line || line.startsWith("#")) continue;
      if (line.startsWith("name:")) {
        const value = line.slice("name:".length).trim().replace(/^['"]|['"]$/g, "");
        return value || null;
      }
    }
    return null;
  } catch (err: any) {
    const message = err && typeof err.message === "string" ? err.message : String(err);
    throw new Error(`loader: read package.yml ${manifestPath}: ${message}`);
  }
}

export async function indexSourceFiles(
  rootDir: String,
  rootName: String,
): Promise<{ packages: Map<String, PackageLocation>; fileToPackage: Map<String, String> }> {
  const packages = new Map<String, PackageLocation>();
  const fileToPackage = new Map<String, String>();

  async function walk(current: String) {
    const entries = await fs.readdir(current, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(current, entry.name);
      if (entry.isDirectory() && entry.name === "quarantine") {
        continue;
      }
      if (entry.isDirectory()) {
        await walk(fullPath);
        continue;
      }
      if (!entry.isFile() || path.extname(entry.name) !== ".able") {
        continue;
      }
      const declared = await scanPackageDeclaration(fullPath);
      const segments = buildPackageSegments(rootDir, rootName, fullPath, declared || []);
      const pkgName = segments.join(".");
      const absPath = path.resolve(fullPath);
      const existing = packages.get(pkgName);
      const loc = existing
        ? existing
        : {
            rootDir,
            rootName,
            files: [],
          };
      loc.files.push(absPath);
      packages.set(pkgName, loc);
      fileToPackage.set(absPath, pkgName);
    }
  }

  await walk(rootDir);
  for (const loc of packages.values()) {
    loc.files.sort();
  }
  return { packages, fileToPackage };
}

export async function scanPackageDeclaration(filePath: String): Promise<String[] | null> {
  try {
    const data = await fs.readFile(filePath, "utf8");
    const lines = data.split(/\r?\n/);
    for (const rawLine of lines) {
      let line = rawLine;
      const commentIdx = line.indexOf("##");
      if (commentIdx >= 0) {
        line = line.slice(0, commentIdx);
      }
      const trimmed = line.trim();
      if (!trimmed) continue;
      let body = trimmed;
      if (body.startsWith("private ")) {
        body = body.slice("private ".length).trim();
      }
      if (!body.startsWith("package ")) {
        continue;
      }
      body = body.slice("package ".length).trim();
      if (body.endsWith(";")) {
        body = body.slice(0, -1).trim();
      }
      if (!body) return null;
      const parts = body.split(".");
      if (parts.length > 1) {
        throw new Error(`loader: package declaration must be unqualified in ${filePath}`);
      }
      const segments = parts.map((part) => sanitizeSegment(part)).filter(Boolean);
      return segments.length > 0 ? segments : null;
    }
    return null;
  } catch (err: any) {
    const message = err && typeof err.message === "string" ? err.message : String(err);
    throw new Error(`loader: read ${filePath}: ${message}`);
  }
}

export function buildPackageSegments(
  rootDir: String,
  rootPackage: String,
  filePath: String,
  declared: String[],
): String[] {
  const base =
    rootPackage === "kernel" || looksLikeKernelPath(String(rootDir))
      ? [sanitizeSegment("able"), sanitizeSegment("kernel")]
      : [sanitizeSegment(rootPackage) || "pkg"];
  return buildPackageSegmentsWithBase(base, rootDir, filePath, declared);
}

function buildPackageSegmentsWithBase(
  base: String[],
  rootDir: String,
  filePath: String,
  declared: String[],
): String[] {
  const segments: String[] = [...base];
  const declaredSegments = declared.map((seg) => sanitizeSegment(seg)).filter(Boolean);
  const rel = path.relative(rootDir, filePath);
  const relDir = path.dirname(rel);
  if (relDir && relDir !== "." && relDir !== path.sep) {
    for (const part of relDir.split(path.sep)) {
      const trimmed = sanitizeSegment(part);
      if (trimmed && trimmed !== ".") {
        segments.push(trimmed);
      }
    }
  }
  segments.push(...declaredSegments);
  return segments;
}

export function buildPackageSegmentsForModule(
  rootDir: String,
  rootPackage: String,
  filePath: String,
  module: AST.Module,
): { segments: String[]; isPrivate: boolean } {
  const declaredSegments: String[] = [];
  let isPrivate = false;
  if (module.package) {
    isPrivate = Boolean(module.package.isPrivate);
    for (const part of module.package.namePath) {
      if (part && part.name) {
        declaredSegments.push(part.name);
      }
    }
  }
  const segments = buildPackageSegments(rootDir, rootPackage, filePath, declaredSegments);
  return { segments, isPrivate };
}

export function sanitizeSegment(value: String): String {
  return value.trim().replace(/-/g, "_");
}

function extractErrorMessage(err: unknown): String {
  if (!err) return "";
  if (typeof err === "string") return err;
  if (err instanceof Error) return err.message;
  if (typeof err === "object" && err) {
    const anyErr = err as Record<String, unknown>;
    if (typeof anyErr.message === "string") return anyErr.message;
  }
  return String(err);
}
