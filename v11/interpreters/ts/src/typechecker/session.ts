import type * as AST from "../ast";
import { TypeChecker, type TypeCheckerOptions, type TypeCheckerPrelude } from "./checker";
import type { PackageSummary, TypecheckResult } from "./diagnostics";

export class TypecheckerSession {
  private readonly packages: Map<string, PackageSummary>;
  private prelude?: TypeCheckerPrelude;

  constructor(initial?: Map<string, PackageSummary> | Record<string, PackageSummary>) {
    this.packages = this.clone(initial);
  }

  checkModule(module: AST.Module, options?: Omit<TypeCheckerOptions, "packageSummaries">): TypecheckResult {
    const checker = new TypeChecker({
      ...options,
      packageSummaries: this.packages,
      prelude: this.prelude,
    });
    const result = checker.checkModule(module);
    if (result.summary) {
      this.packages.set(result.summary.name, result.summary);
    }
    this.mergePrelude(checker.exportPrelude());
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

  private mergePrelude(next: TypeCheckerPrelude): void {
    if (!this.prelude) {
      this.prelude = {
        structs: new Map(next.structs ?? new Map()),
        interfaces: new Map(next.interfaces ?? new Map()),
        typeAliases: new Map(next.typeAliases ?? new Map()),
        unions: new Map(next.unions ?? new Map()),
        functionInfos: new Map(next.functionInfos ?? new Map()),
        methodSets: [...(next.methodSets ?? [])],
        implementationRecords: [...(next.implementationRecords ?? [])],
      };
      return;
    }
    const mergeMaps = <T>(target: Map<string, T>, source?: Map<string, T>): void => {
      if (!source) return;
      for (const [key, value] of source.entries()) {
        target.set(key, value);
      }
    };
    mergeMaps(this.prelude.structs, next.structs);
    mergeMaps(this.prelude.interfaces, next.interfaces);
    mergeMaps(this.prelude.typeAliases, next.typeAliases);
    mergeMaps(this.prelude.unions, next.unions);
    mergeMaps(this.prelude.functionInfos, next.functionInfos);
    this.prelude.methodSets.push(...(next.methodSets ?? []));
    this.prelude.implementationRecords.push(...(next.implementationRecords ?? []));
  }
}
