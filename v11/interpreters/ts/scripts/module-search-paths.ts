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
  ableStdLibEnv?: string;
  extras?: ModuleSearchPath[];
  probeStdlibFrom?: string[];
};

export function collectModuleSearchPaths(options: CollectOptions = {}): ModuleSearchPath[] {
  const cwd = options.cwd ? path.resolve(options.cwd) : process.cwd();
  const ordered: ModuleSearchPath[] = [];
  const seen = new Set<string>();

  const add = (candidate: ModuleSearchPath | string | undefined | null, isStdlib = false) => {
    if (!candidate) return;
    const normalized = normalizeSearchPath(candidate, isStdlib);
    if (!normalized) return;
    const abs = path.resolve(normalized.path);
    if (seen.has(abs)) return;
    seen.add(abs);
    ordered.push({ ...normalized, path: abs });
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

  const stdlibPaths = collectStdlibPaths(options.ableStdLibEnv, options.probeStdlibFrom || []);
  for (const entry of stdlibPaths) {
    add({ path: entry, isStdlib: true, source: "ABLE_STD_LIB" }, true);
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

export function collectStdlibPaths(envValue: string | undefined, probes: string[]): string[] {
  const ordered: string[] = [];
  const add = (candidate: string | null | undefined) => {
    if (!candidate) return;
    const abs = path.resolve(candidate);
    if (ordered.some((existing) => path.resolve(existing) === abs)) return;
    ordered.push(abs);
  };

  const envPaths = parsePathList(envValue);
  if (envPaths.length > 0) {
    for (const entry of envPaths) add(entry);
    return ordered;
  }

  for (const probe of probes) {
    add(findStdlibRoot(probe));
  }

  return ordered;
}

function findStdlibRoot(start: string): string | null {
  if (!start) return null;
  let dir = path.resolve(start);
  while (true) {
    for (const candidate of [
      path.join(dir, "stdlib", "src"),
      path.join(dir, "stdlib", "v11", "src"),
      path.join(dir, "stdlib", "v10", "src"),
    ]) {
      if (fsExists(candidate)) {
        return candidate;
      }
    }
    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  return null;
}

function normalizeSearchPath(value: ModuleSearchPath | string, isStdlib = false): ModuleSearchPath | null {
  if (!value) return null;
  if (typeof value === "string") {
    return { path: value, isStdlib };
  }
  if (!value.path) return null;
  return { path: value.path, isStdlib: Boolean(value.isStdlib), source: value.source };
}

function fsExists(candidate: string): boolean {
  try {
    return fs.statSync(candidate).isDirectory();
  } catch {
    return false;
  }
}
