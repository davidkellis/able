import type { TypeInfo } from "./types";

export type DiagnosticSeverity = "error" | "warning";

export type DiagnosticLocation = {
  path?: string;
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
};

export type DiagnosticNote = {
  message: string;
  location?: DiagnosticLocation;
};

export type TypecheckerDiagnostic = {
  severity: DiagnosticSeverity;
  message: string;
  code?: string;
  location?: DiagnosticLocation;
  notes?: DiagnosticNote[];
};

export type ExportedSymbolSummary = {
  type: string;
  visibility: "public" | "private";
};

export type ExportedGenericParamSummary = {
  name: string;
  constraints?: string[];
};

export type ExportedWhereConstraintSummary = {
  typeParam: string;
  constraints?: string[];
};

export type ExportedObligationSummary = {
  owner?: string;
  typeParam: string;
  constraint: string;
  subject: string;
  context?: string;
};

export type ExportedFunctionSummary = {
  parameters?: string[];
  returnType: string;
  typeParams?: ExportedGenericParamSummary[];
  where?: ExportedWhereConstraintSummary[];
  obligations?: ExportedObligationSummary[];
};

export type ExportedStructSummary = {
  typeParams?: ExportedGenericParamSummary[];
  fields?: Record<string, string>;
  positional?: string[];
  where?: ExportedWhereConstraintSummary[];
};

export type ExportedUnionSummary = {
  typeParams?: ExportedGenericParamSummary[];
  variants?: string[];
  where?: ExportedWhereConstraintSummary[];
};

export type ExportedInterfaceSummary = {
  typeParams?: ExportedGenericParamSummary[];
  methods?: Record<string, ExportedFunctionSummary>;
  where?: ExportedWhereConstraintSummary[];
};

export type ExportedImplementationSummary = {
  implName?: string;
  interface: string;
  target: string;
  interfaceArgs?: string[];
  typeParams?: ExportedGenericParamSummary[];
  methods?: Record<string, ExportedFunctionSummary>;
  where?: ExportedWhereConstraintSummary[];
  obligations?: ExportedObligationSummary[];
};

export type ExportedMethodSetSummary = {
  typeParams?: ExportedGenericParamSummary[];
  target: string;
  methods?: Record<string, ExportedFunctionSummary>;
  where?: ExportedWhereConstraintSummary[];
  obligations?: ExportedObligationSummary[];
};

export type PackageSummary = {
  name: string;
  visibility: "public" | "private";
  symbols: Record<string, ExportedSymbolSummary>;
  privateSymbols: Record<string, ExportedSymbolSummary>;
  symbolTypes?: Record<string, TypeInfo>;
  structs: Record<string, ExportedStructSummary>;
  unions: Record<string, ExportedUnionSummary>;
  interfaces: Record<string, ExportedInterfaceSummary>;
  functions: Record<string, ExportedFunctionSummary>;
  implementations: ExportedImplementationSummary[];
  methodSets: ExportedMethodSetSummary[];
};

export type TypecheckResult = {
  diagnostics: TypecheckerDiagnostic[];
  summary?: PackageSummary | null;
};
