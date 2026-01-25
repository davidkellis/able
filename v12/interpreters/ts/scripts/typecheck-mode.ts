export type TypecheckMode = "off" | "warn" | "strict";

export function resolveTypecheckMode(raw: string | undefined): TypecheckMode {
  if (raw === undefined) return "strict";
  const normalized = raw.trim().toLowerCase();
  if (normalized === "" || normalized === "0" || normalized === "off" || normalized === "false") {
    return "off";
  }
  if (normalized === "strict" || normalized === "fail" || normalized === "error" || normalized === "1" || normalized === "true") {
    return "strict";
  }
  if (normalized === "warn" || normalized === "warning") {
    return "warn";
  }
  return "warn";
}
