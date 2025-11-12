import path from "node:path";
import { fileURLToPath } from "node:url";

import type {
  DiagnosticLocation,
  PackageSummary,
  TypecheckerDiagnostic,
} from "../src/typechecker/diagnostics";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../..");

export function formatTypecheckerDiagnostic(
  diag: TypecheckerDiagnostic,
  options?: { packageName?: string },
): string {
  const location = formatDiagnosticLocation(diag.location);
  const segments: string[] = [];
  if (options?.packageName && options.packageName !== "<anonymous>") {
    segments.push(options.packageName);
  }
  if (location) {
    segments.push(location);
  }
  const prefix = segments.length > 0 ? `${segments.join(" ")} ` : "";
  return `typechecker: ${prefix}${diag.message}`;
}

export function printPackageSummaries(summaries: Map<string, PackageSummary>): void {
  if (summaries.size === 0) {
    return;
  }
  const entries = [...summaries.values()].sort((a, b) => a.name.localeCompare(b.name));
  if (entries.length === 0) {
    return;
  }
  console.warn("---- package export summary ----");
  for (const summary of entries) {
    const label =
      summary.visibility && summary.visibility === "private"
        ? `${summary.name} (private)`
        : summary.name;
    const structs = formatSummaryList(summary.structs);
    const interfaces = formatSummaryList(summary.interfaces);
    const functions = formatSummaryList(summary.functions);
    console.warn(
      `package ${label} exports: structs=${structs}; interfaces=${interfaces}; functions=${functions}; impls=${summary.implementations.length}; method sets=${summary.methodSets.length}`,
    );
  }
}

function formatSummaryList(record?: Record<string, unknown>): string {
  if (!record) {
    return "-";
  }
  const names = Object.keys(record).sort((a, b) => a.localeCompare(b));
  if (names.length === 0) {
    return "-";
  }
  return names.join(", ");
}

function formatDiagnosticLocation(location: DiagnosticLocation | undefined): string | null {
  if (!location) {
    return null;
  }
  const { path: filePath, line, column } = location;
  const normalizedPath = filePath ? normalizePath(filePath) : null;
  if (normalizedPath && line && column) {
    return `${normalizedPath}:${line}:${column}`;
  }
  if (normalizedPath && line) {
    return `${normalizedPath}:${line}`;
  }
  if (normalizedPath) {
    return normalizedPath;
  }
  if (line && column) {
    return `line ${line}, column ${column}`;
  }
  if (line) {
    return `line ${line}`;
  }
  return null;
}

function normalizePath(filePath: string): string {
  const absolute = path.isAbsolute(filePath) ? filePath : path.resolve(filePath);
  const relative = path.relative(REPO_ROOT, absolute);
  const target = relative === "" ? path.basename(absolute) : relative;
  return target.split(path.sep).join("/");
}
