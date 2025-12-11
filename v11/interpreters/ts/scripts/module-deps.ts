import fs from "node:fs";
import { promises as fsp } from "node:fs";
import os from "node:os";
import path from "node:path";

import { parse as parseYAML } from "yaml";

import type { ModuleSearchPath } from "./module-search-paths";
import { looksLikeKernelPath, looksLikeStdlibPath } from "./module-search-paths";

export type ManifestData = {
  path: string;
  name?: string | null;
  dependencies?: Record<string, unknown>;
  dev_dependencies?: Record<string, unknown>;
  build_dependencies?: Record<string, unknown>;
};

export type LockfileData = {
  path: string;
  root?: string | null;
  packages: LockedPackage[];
};

export type LockedPackage = {
  name?: string | null;
  version?: string | null;
  source?: string | null;
  checksum?: string | null;
  dependencies?: Array<{ name?: string | null; version?: string | null }> | null;
};

export type ManifestContext = {
  manifest: ManifestData | null;
  lock: LockfileData | null;
  manifestPath: string | null;
};

export async function loadManifestContext(startPath: string): Promise<ManifestContext> {
  const manifestPath = await findManifestPath(startPath);
  if (!manifestPath) {
    return { manifest: null, lock: null, manifestPath: null };
  }
  const manifest = await loadManifest(manifestPath);
  const lock = await loadLockfile(manifest);
  return { manifest, lock, manifestPath };
}

export async function findManifestPath(start: string): Promise<string | null> {
  let dir = path.resolve(start);
  try {
    const stat = await fsp.stat(dir);
    if (!stat.isDirectory()) {
      dir = path.dirname(dir);
    }
  } catch {
    // fall through and treat as directory
  }
  const origin = dir;
  while (true) {
    const candidate = path.join(dir, "package.yml");
    try {
      const stats = await fsp.stat(candidate);
      if (stats.isFile()) {
        return candidate;
      }
    } catch {
      // continue searching upwards
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      break;
    }
    dir = parent;
  }
  return null;
}

export async function loadManifest(manifestPath: string): Promise<ManifestData> {
  const abs = path.resolve(manifestPath);
  let contents: string;
  try {
    contents = await fsp.readFile(abs, "utf8");
  } catch (error) {
    const message = extractErrorMessage(error);
    throw new Error(`failed to read manifest ${abs}: ${message}`);
  }
  let parsed: any;
  try {
    parsed = parseYAML(contents) ?? {};
  } catch (error) {
    const message = extractErrorMessage(error);
    throw new Error(`failed to parse manifest ${abs}: ${message}`);
  }
  const manifest: ManifestData = {
    path: abs,
    name: typeof parsed?.name === "string" ? parsed.name : null,
    dependencies: isRecord(parsed?.dependencies) ? parsed.dependencies : undefined,
    dev_dependencies: isRecord(parsed?.dev_dependencies) ? parsed.dev_dependencies : undefined,
    build_dependencies: isRecord(parsed?.build_dependencies) ? parsed.build_dependencies : undefined,
  };
  return manifest;
}

export async function loadLockfile(manifest: ManifestData): Promise<LockfileData | null> {
  const manifestDir = path.dirname(manifest.path);
  const lockPath = path.join(manifestDir, "package.lock");
  let contents: string;
  try {
    contents = await fsp.readFile(lockPath, "utf8");
  } catch (error: any) {
    if (error && typeof error.code === "string" && error.code === "ENOENT") {
      if (manifestHasDependencies(manifest)) {
        const name = manifest.name ?? path.basename(manifestDir);
        throw new Error(`package.lock missing for "${name}"; run \`able deps install\``);
      }
      return null;
    }
    throw new Error(`failed to read lockfile ${lockPath}: ${extractErrorMessage(error)}`);
  }

  let parsed: any;
  try {
    parsed = parseYAML(contents) ?? {};
  } catch (error) {
    const message = extractErrorMessage(error);
    throw new Error(`failed to parse lockfile ${lockPath}: ${message}`);
  }

  const lock: LockfileData = {
    path: lockPath,
    root: typeof parsed?.root === "string" ? parsed.root : null,
    packages: Array.isArray(parsed?.packages)
      ? parsed.packages.map(normalizeLockedPackage).filter(Boolean) as LockedPackage[]
      : [],
  };
  if (manifest.name && lock.root && sanitizeName(lock.root) !== sanitizeName(manifest.name)) {
    throw new Error(
      `lockfile root ${String(lock.root)} does not match manifest name ${manifest.name}`,
    );
  }
  return lock;
}

export function manifestHasDependencies(manifest: ManifestData | null): boolean {
  if (!manifest) return false;
  const candidates = [manifest.dependencies, manifest.dev_dependencies, manifest.build_dependencies];
  return candidates.some((record) => record && Object.keys(record).length > 0);
}

export function buildExecutionSearchPaths(
  manifest: ManifestData | null,
  lock: LockfileData | null,
  ableHomeOverride?: string,
): ModuleSearchPath[] {
  const manifestRoot = manifest?.path ? path.dirname(manifest.path) : null;
  const ableHome = ableHomeOverride ?? resolveAbleHome();
  const extras: ModuleSearchPath[] = [];
  const seen = new Set<string>();

  const push = (entry: ModuleSearchPath) => {
    const abs = path.resolve(entry.path);
    if (seen.has(abs)) return;
    try {
      const stats = fs.statSync(abs);
      if (!stats.isDirectory()) return;
    } catch {
      return;
    }
    seen.add(abs);
    extras.push({ ...entry, path: abs });
  };

  if (manifestRoot) {
    push({ path: manifestRoot, source: "manifest" });
  }

  if (lock?.packages?.length) {
    for (const pkg of lock.packages) {
      const resolved = resolvePackageSourcePath(pkg, manifestRoot, ableHome);
      if (!resolved) continue;
      const pkgName = sanitizeName(pkg.name ?? "");
      const isStdlib = pkgName === "able" || pkgName === "kernel" || looksLikeStdlibPath(resolved) || looksLikeKernelPath(resolved);
      push({ path: resolved, isStdlib, source: "lock" });
    }
  }

  return extras;
}

function resolvePackageSourcePath(
  pkg: LockedPackage,
  manifestRoot: string | null,
  ableHome: string,
): string | null {
  if (!pkg) return null;
  const name = sanitizeName(pkg.name ?? "");
  const version = sanitizePathSegment(pkg.version ?? "");
  const source = typeof pkg.source === "string" ? pkg.source.trim() : "";

  if (source.startsWith("path:")) {
    const pathSpec = source.slice("path:".length).trim();
    if (!pathSpec) return null;
    if (path.isAbsolute(pathSpec)) {
      return pathSpec;
    }
    const base = manifestRoot ?? ableHome;
    return path.join(base, pathSpec);
  }

  if (source.startsWith("registry:")) {
    const pathSpec = source.slice("registry:".length).trim();
    if (!pathSpec) return null;
    const parts = pathSpec.split("/").filter(Boolean);
    if (parts.length >= 2) {
      const pkgName = sanitizeName(parts[parts.length - 2] ?? name);
      const pkgVersion = sanitizePathSegment(parts[parts.length - 1] ?? version);
      return path.join(ableHome, "pkg", "src", pkgName, pkgVersion);
    }
    return path.join(ableHome, "pkg", "src", name, version);
  }

  if (source.startsWith("git:")) {
    const pathSpec = source.slice("git:".length).trim();
    if (!pathSpec) return null;
    return path.join(ableHome, "pkg", "src", pathSpec);
  }

  if (name && version) {
    return path.join(ableHome, "pkg", "src", name, version);
  }

  return null;
}

function resolveAbleHome(): string {
  const env = process.env.ABLE_HOME;
  if (env && env.trim().length > 0) {
    return path.resolve(env.trim());
  }
  return path.join(os.homedir(), ".able");
}

function sanitizeName(value: string): string {
  return value.trim().replace(/-/g, "_");
}

function sanitizePathSegment(value: string): string {
  const cleaned = value.trim();
  if (!cleaned) return "head";
  let result = "";
  for (const ch of cleaned) {
    if (/[a-zA-Z0-9._-]/.test(ch)) {
      result += ch;
    } else {
      result += "_";
    }
  }
  return result || "head";
}

function normalizeLockedPackage(pkg: any): LockedPackage | null {
  if (!pkg || typeof pkg !== "object") return null;
  const normalized: LockedPackage = {
    name: typeof pkg.name === "string" ? pkg.name : null,
    version: typeof pkg.version === "string" ? pkg.version : null,
    source: typeof pkg.source === "string" ? pkg.source : null,
    checksum: typeof pkg.checksum === "string" ? pkg.checksum : null,
    dependencies: Array.isArray(pkg.dependencies)
      ? pkg.dependencies.map((dep: any) => ({
          name: typeof dep?.name === "string" ? dep.name : null,
          version: typeof dep?.version === "string" ? dep.version : null,
        }))
      : null,
  };
  return normalized;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function extractErrorMessage(error: unknown): string {
  if (!error) return "";
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  const anyErr = error as any;
  if (anyErr && typeof anyErr.message === "string") return anyErr.message;
  return String(error);
}
