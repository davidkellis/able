import type * as AST from "../ast";
import { TypeChecker, createTypeChecker } from "./checker";
import type { TypeCheckerOptions } from "./checker";
import { TypecheckerSession } from "./session";
import type {
  TypecheckResult,
  TypecheckerDiagnostic,
  DiagnosticSeverity,
  DiagnosticLocation,
  PackageSummary,
  ExportedSymbolSummary,
  ExportedStructSummary,
  ExportedInterfaceSummary,
  ExportedFunctionSummary,
  ExportedImplementationSummary,
  ExportedMethodSetSummary,
  ExportedGenericParamSummary,
  ExportedWhereConstraintSummary,
  ExportedObligationSummary,
} from "./diagnostics";
import type { TypeInfo, PrimitiveName } from "./types";

export type {
  TypeCheckerOptions,
  TypecheckResult,
  TypecheckerDiagnostic,
  DiagnosticSeverity,
  DiagnosticLocation,
  PackageSummary,
  ExportedSymbolSummary,
  ExportedStructSummary,
  ExportedInterfaceSummary,
  ExportedFunctionSummary,
  ExportedImplementationSummary,
  ExportedMethodSetSummary,
  ExportedGenericParamSummary,
  ExportedWhereConstraintSummary,
  ExportedObligationSummary,
  TypeInfo,
  PrimitiveName,
};
export { TypeChecker, createTypeChecker, TypecheckerSession };

export function checkModule(module: AST.Module, options?: TypeCheckerOptions): TypecheckResult {
  const checker = createTypeChecker(options);
  return checker.checkModule(module);
}
