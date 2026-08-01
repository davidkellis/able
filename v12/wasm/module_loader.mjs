// loadStaticModuleClosure resolves the deliberately narrow static source
// module format: one package per <root>/<package>.able or
// <root>/<package>/main.able. The host supplies source bytes through
// sourceProvider; resolver ordering and package policy remain here so Node and
// browser hosts do not need to duplicate them.
export async function loadStaticModuleClosure({
  entryPath,
  moduleRoots = [],
  parseSource,
  sourceProvider,
}) {
  if (typeof parseSource !== "function") {
    throw new Error("module loader requires a parseSource function");
  }
  if (!entryPath) {
    throw new Error("module loader requires an entry source path");
  }
  if (!sourceProvider || typeof sourceProvider.readSource !== "function") {
    throw new Error("module loader requires a sourceProvider.readSource function");
  }

  const entryOrigin = normalizeOrigin(entryPath);
  const roots = uniqueOrigins([dirname(entryOrigin), ...moduleRoots]);
  const stateByOrigin = new Map();
  const sourceByOrigin = new Map();
  const modules = [];
  const visitStack = [];

  async function visit(origin, expectedPackage = "") {
    const normalizedOrigin = normalizeOrigin(origin);
    const previousState = stateByOrigin.get(normalizedOrigin);
    if (previousState === "done") {
      return;
    }
    if (previousState === "visiting") {
      const cycleStart = visitStack.indexOf(normalizedOrigin);
      const cycle = [...visitStack.slice(cycleStart), normalizedOrigin]
        .map((candidate) => displayOrigin(entryOrigin, candidate))
        .join(" -> ");
      throw new Error(`static import cycle: ${cycle}`);
    }

    stateByOrigin.set(normalizedOrigin, "visiting");
    visitStack.push(normalizedOrigin);
    try {
      const source = await requireSource(normalizedOrigin);
      let module;
      try {
        module = parseSource(source, normalizedOrigin);
      } catch (err) {
        throw new Error(`parse ${normalizedOrigin}: ${messageForError(err)}`);
      }
      validateModule(module, normalizedOrigin);

      const actualPackage = packageName(module);
      if (expectedPackage && actualPackage !== expectedPackage) {
        throw new Error(`module ${normalizedOrigin} declares package ${formatPackage(actualPackage)}, import requires ${expectedPackage}`);
      }

      for (const imp of module.imports) {
        const importedPackage = staticImportPackage(imp, normalizedOrigin);
        const dependencyOrigin = await resolvePackageSource(importedPackage, roots);
        await visit(dependencyOrigin, importedPackage);
      }

      modules.push({ origin: normalizedOrigin, module });
      stateByOrigin.set(normalizedOrigin, "done");
    } finally {
      visitStack.pop();
      if (stateByOrigin.get(normalizedOrigin) === "visiting") {
        stateByOrigin.delete(normalizedOrigin);
      }
    }
  }

  async function requireSource(origin) {
    const source = await readSource(origin);
    if (source === null) {
      throw new Error(`source ${origin} not found`);
    }
    return source;
  }

  async function resolvePackageSource(packageName, searchRoots) {
    const segments = packageName.split(".");
    const relativePath = joinOrigin(...segments);
    const candidates = [];
    for (const root of searchRoots) {
      candidates.push(joinOrigin(root, `${relativePath}.able`));
      candidates.push(joinOrigin(root, relativePath, "main.able"));
    }
    for (const candidate of candidates) {
      if (await readSource(candidate) !== null) {
        return candidate;
      }
    }
    throw new Error(`static import ${packageName} not found; searched ${candidates.join(", ")}`);
  }

  async function readSource(origin) {
    const normalizedOrigin = normalizeOrigin(origin);
    if (sourceByOrigin.has(normalizedOrigin)) {
      return sourceByOrigin.get(normalizedOrigin);
    }
    let source;
    try {
      source = await sourceProvider.readSource(normalizedOrigin);
    } catch (err) {
      throw new Error(`read ${normalizedOrigin}: ${messageForError(err)}`);
    }
    if (source === undefined || source === null) {
      sourceByOrigin.set(normalizedOrigin, null);
      return null;
    }
    if (typeof source !== "string") {
      throw new Error(`read ${normalizedOrigin}: source provider returned ${typeof source}, want string or null`);
    }
    sourceByOrigin.set(normalizedOrigin, source);
    return source;
  }

  await visit(entryOrigin);
  const entry = modules.at(-1);
  if (!entry) {
    throw new Error(`no entry module loaded from ${entryOrigin}`);
  }
  return {
    entry,
    setupModules: modules.slice(0, -1),
    moduleRoots: roots,
  };
}

function staticImportPackage(imp, origin) {
  if (imp?.type === "DynImportStatement") {
    throw new Error(`dynamic import in ${origin} is unavailable in browser source module loading`);
  }
  if (imp?.type !== "ImportStatement") {
    throw new Error(`unsupported import node ${JSON.stringify(imp?.type)} in ${origin}`);
  }
  if (!Array.isArray(imp.packagePath) || imp.packagePath.length === 0) {
    throw new Error(`static import in ${origin} has no package path`);
  }
  const segments = imp.packagePath.map((segment) => segment?.name);
  for (const segment of segments) {
    if (typeof segment !== "string" || !segment || segment === "." || segment === ".." || segment.includes("/") || segment.includes("\\")) {
      throw new Error(`static import in ${origin} has an unsafe package segment ${JSON.stringify(segment)}`);
    }
  }
  return segments.join(".");
}

function validateModule(module, origin) {
  if (!module || module.type !== "Module") {
    throw new Error(`parse ${origin} did not produce an AST Module`);
  }
  if (!Array.isArray(module.imports)) {
    throw new Error(`module ${origin} has no imports array`);
  }
  if (!Array.isArray(module.body)) {
    throw new Error(`module ${origin} has no body array`);
  }
}

function packageName(module) {
  const namePath = module?.package?.namePath;
  if (!Array.isArray(namePath) || namePath.length === 0) {
    return "";
  }
  const segments = namePath.map((segment) => segment?.name);
  if (segments.some((segment) => typeof segment !== "string" || !segment)) {
    return "";
  }
  return segments.join(".");
}

function formatPackage(name) {
  return name ? name : "<none>";
}

function uniqueOrigins(origins) {
  const seen = new Set();
  const unique = [];
  for (const candidate of origins) {
    if (!candidate) {
      continue;
    }
    const normalized = normalizeOrigin(candidate);
    if (!seen.has(normalized)) {
      seen.add(normalized);
      unique.push(normalized);
    }
  }
  return unique;
}

// Origins use a portable slash-separated virtual-path form. Hosts may map
// them to real files, sandbox handles, or in-memory source maps. Import paths
// cannot add dot segments, so this normalisation governs only caller-provided
// entry and root labels.
function normalizeOrigin(origin) {
  if (typeof origin !== "string" || origin === "") {
    throw new Error("module loader origin must be a non-empty string");
  }
  const slashSeparated = origin.replaceAll("\\", "/");
  const absolute = slashSeparated.startsWith("/");
  const parts = [];
  for (const segment of slashSeparated.split("/")) {
    if (!segment || segment === ".") {
      continue;
    }
    if (segment === "..") {
      if (parts.length > 0 && parts.at(-1) !== "..") {
        parts.pop();
      } else if (!absolute) {
        parts.push(segment);
      }
      continue;
    }
    parts.push(segment);
  }
  const joined = parts.join("/");
  if (absolute) {
    return joined ? `/${joined}` : "/";
  }
  return joined || ".";
}

function dirname(origin) {
  const normalized = normalizeOrigin(origin);
  const separator = normalized.lastIndexOf("/");
  if (separator < 0) {
    return ".";
  }
  if (separator === 0) {
    return "/";
  }
  return normalized.slice(0, separator);
}

function joinOrigin(...parts) {
  return normalizeOrigin(parts.filter(Boolean).join("/"));
}

function displayOrigin(entryOrigin, candidate) {
  const entryDirectory = dirname(entryOrigin);
  const prefix = entryDirectory === "/" ? "/" : `${entryDirectory}/`;
  if (candidate.startsWith(prefix)) {
    return candidate.slice(prefix.length) || ".";
  }
  return candidate;
}

function messageForError(err) {
  return err instanceof Error ? err.message : String(err);
}
