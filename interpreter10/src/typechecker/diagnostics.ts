export type DiagnosticSeverity = "error" | "warning";

export type TypecheckerDiagnostic = {
  severity: DiagnosticSeverity;
  message: string;
  code?: string;
};

export type TypecheckResult = {
  diagnostics: TypecheckerDiagnostic[];
};
