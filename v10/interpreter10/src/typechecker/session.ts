import type * as AST from "../ast";
import { TypeChecker, type TypeCheckerOptions } from "./checker";
import type { PackageSummary, TypecheckResult } from "./diagnostics";

export class TypecheckerSession {
  private readonly packages: Map<string, PackageSummary>;

  constructor(initial?: Map<string, PackageSummary> | Record<string, PackageSummary>) {
    this.packages = this.clone(initial);
  }

  checkModule(module: AST.Module, options?: Omit<TypeCheckerOptions, "packageSummaries">): TypecheckResult {
    const checker = new TypeChecker({
      ...options,
      packageSummaries: this.packages,
    });
    const result = checker.checkModule(module);
    if (result.summary) {
      this.packages.set(result.summary.name, result.summary);
    }
    return result;
  }

  getPackageSummaries(): Map<string, PackageSummary> {
    return new Map(this.packages);
  }

  private clone(
    source?: Map<string, PackageSummary> | Record<string, PackageSummary>,
  ): Map<string, PackageSummary> {
    if (!source) {
      return new Map();
    }
    if (source instanceof Map) {
      return new Map(source);
    }
    return new Map(Object.entries(source));
  }
}
