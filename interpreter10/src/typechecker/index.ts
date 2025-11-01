import type * as AST from "../ast";
import { TypeChecker, createTypeChecker } from "./checker";
import type { TypeCheckerOptions } from "./checker";
import type { TypecheckResult, TypecheckerDiagnostic, DiagnosticSeverity } from "./diagnostics";
import type { TypeInfo, PrimitiveName } from "./types";

export type { TypeCheckerOptions, TypecheckResult, TypecheckerDiagnostic, DiagnosticSeverity, TypeInfo, PrimitiveName };
export { TypeChecker, createTypeChecker };

export function checkModule(module: AST.Module, options?: TypeCheckerOptions): TypecheckResult {
  const checker = createTypeChecker(options);
  return checker.checkModule(module);
}
