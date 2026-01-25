import fs from "node:fs";
import path from "node:path";

export type ModuleSearchPath = {
  path: string;
  isStdlib?: boolean;
  source?: string;
};

type CollectOptions = {
  cwd?: string;
  ablePathEnv?: string;
  ableModulePathsEnv?: string;
  extras?: ModuleSearchPath[];
  probeFrom?: string[];
};

export function collectModuleSearchPaths(options: CollectOptions = {}): ModuleSearchPath[] {
  const cwd = options.cwd ? path.resolve(options.cwd) : process.cwd();
  const ordered: ModuleSearchPath[] = [];
  const seen = new Set<string>();
  const probes = options.probeFrom ?? [];

  const add = (candidate: ModuleSearchPath | string | undefined | null, isStdlib = false) => {
    if (!candidate) return;
    const normalized = normalizeSearchPath(candidate, isStdlib);
    if (!normalized) return;
    const abs = path.resolve(normalized.path);
    if (seen.has(abs)) return;
    seen.add(abs);
    const flagStdlib =
      normalized.isStdlib || looksLikeStdlibPath(abs) || looksLikeKernelPath(abs);
    ordered.push({ ...normalized, path: abs, isStdlib: flagStdlib });
  };

  add({ path: cwd, source: "cwd" });
  for (const extra of options.extras || []) {
    add(extra, Boolean(extra?.isStdlib));
  }

  for (const entry of parsePathList(options.ablePathEnv)) {
    add({ path: entry, source: "ABLE_PATH" });
  }
  for (const entry of parsePathList(options.ableModulePathsEnv)) {
    add({ path: entry, source: "ABLE_MODULE_PATHS" });
  }

  const bundledRoots = collectBundledRoots([cwd, ...probes]);
  for (const entry of bundledRoots) {
    add(entry, Boolean(entry.isStdlib));
  }

  return ordered;
}

export function parsePathList(raw: string | undefined): string[] {
  if (!raw) return [];
  return raw
    .split(path.delimiter)
    .map((segment) => segment.trim())
    .filter((segment) => segment.length > 0)
    .map((segment) => path.resolve(segment));
}

export function collectBundledRoots(probes: string[]): ModuleSearchPath[] {
  const ordered: ModuleSearchPath[] = [];
  const add = (candidate: string | null | undefined, isStdlib: boolean) => {
    if (!candidate) return;
    const abs = path.resolve(candidate);
    if (ordered.some((existing) => path.resolve(existing.path) === abs)) return;
    ordered.push({ path: abs, isStdlib, source: "auto" });
  };

  for (const probe of probes) {
    for (const candidate of findRootCandidates(probe)) {
      add(candidate.path, candidate.isStdlib);
    }
  }

  return ordered;
}

function findRootCandidates(start: string): { path: string; isStdlib: boolean }[] {
  if (!start) return [];
  const roots: { path: string; isStdlib: boolean }[] = [];
  let dir = path.resolve(start);
  while (true) {
    for (const candidate of [
      { path: path.join(dir, "kernel", "src"), isStdlib: true },
      { path: path.join(dir, "v11", "kernel", "src"), isStdlib: true },
      { path: path.join(dir, "stdlib", "src"), isStdlib: true },
      { path: path.join(dir, "v11", "stdlib", "src"), isStdlib: true },
      { path: path.join(dir, "stdlib", "v11", "src"), isStdlib: true },
    ]) {
      if (fsExists(candidate.path)) {
        roots.push(candidate);
      }
    }
    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  return roots;
}

function normalizeSearchPath(value: ModuleSearchPath | string, isStdlib = false): ModuleSearchPath | null {
  if (!value) return null;
  if (typeof value === "string") {
    return { path: value, isStdlib };
  }
  if (!value.path) return null;
  return { path: value.path, isStdlib: Boolean(value.isStdlib), source: value.source };
}

export function looksLikeStdlibPath(dir: string): boolean {
  return path
    .resolve(dir)
    .split(path.sep)
    .some((segment) => segment === "stdlib" || segment === "able_stdlib" || segment === "able-stdlib");
}

function fsExists(candidate: string): boolean {
  try {
    return fs.statSync(candidate).isDirectory();
  } catch {
    return false;
  }
}

export function looksLikeKernelPath(dir: string): boolean {
  return path
    .resolve(dir)
    .split(path.sep)
    .some((segment) => segment === "kernel" || segment === "ablekernel" || segment === "able_kernel");
}
