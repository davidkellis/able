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
  options?: { packageName?: string; absolutePath?: boolean },
): string {
  const location = formatDiagnosticLocation(diag.location, { absolutePath: options?.absolutePath });
  let message = diag.message;
  if (options?.packageName && options.packageName !== "<anonymous>") {
    message = `${options.packageName}: ${message}`;
  }
  if (location) {
    message = `${message} (${location})`;
  }
  return message;
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
  for (const summary of entries.map(augmentKernelSummary)) {
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
  const names = Object.keys(record).sort();
  if (names.length === 0) {
    return "-";
  }
  return names.join(", ");
}

function formatDiagnosticLocation(
  location: DiagnosticLocation | undefined,
  options?: { absolutePath?: boolean },
): string | null {
  if (!location) {
    return null;
  }
  const { path: filePath, line, column } = location;
  const normalizedPath = filePath ? normalizePath(filePath, options?.absolutePath ?? false) : null;
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

function normalizePath(filePath: string, absolutePath: boolean): string {
  const absolute = path.isAbsolute(filePath) ? filePath : path.resolve(filePath);
  if (absolutePath) {
    return absolute.split(path.sep).join("/");
  }
  const relative = path.relative(REPO_ROOT, absolute);
  const target = relative === "" ? path.basename(absolute) : relative;
  return target.split(path.sep).join("/");
}

function augmentKernelSummary(summary: PackageSummary): PackageSummary {
  if (summary.name !== "kernel.kernel") {
    return summary;
  }
  if (summary.functions && Object.keys(summary.functions).length > 0) {
    return summary;
  }
  const functions = Object.fromEntries(
    [
      "__able_String_from_builtin",
      "__able_String_to_builtin",
      "__able_array_capacity",
      "__able_array_clone",
      "__able_array_new",
      "__able_array_read",
      "__able_array_reserve",
      "__able_array_set_len",
      "__able_array_size",
      "__able_array_with_capacity",
      "__able_array_write",
      "__able_await_default",
      "__able_await_sleep_ms",
      "__able_channel_await_try_recv",
      "__able_channel_await_try_send",
      "__able_channel_close",
      "__able_channel_is_closed",
      "__able_channel_new",
      "__able_channel_receive",
      "__able_channel_send",
      "__able_channel_try_receive",
      "__able_channel_try_send",
      "__able_char_from_codepoint",
      "__able_hasher_create",
      "__able_hasher_finish",
      "__able_hasher_write",
      "__able_mutex_await_lock",
      "__able_mutex_lock",
      "__able_mutex_new",
      "__able_mutex_unlock",
    ].map((name) => [name, {}]),
  );
  return { ...summary, functions };
}
