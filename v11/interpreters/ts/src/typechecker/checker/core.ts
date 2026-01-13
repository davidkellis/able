import type * as AST from "../../ast";
import type { PackageSummary } from "../diagnostics";
import type { FunctionInfo, ImplementationRecord, MethodSetRecord } from "./types";

export type LocalTypeDeclaration =
  | AST.StructDefinition
  | AST.UnionDefinition
  | AST.InterfaceDefinition
  | AST.TypeAliasDefinition;

export interface TypeCheckerPrelude {
  structs: Map<string, AST.StructDefinition>;
  interfaces: Map<string, AST.InterfaceDefinition>;
  typeAliases: Map<string, AST.TypeAliasDefinition>;
  unions: Map<string, AST.UnionDefinition>;
  functionInfos: Map<string, FunctionInfo[]>;
  methodSets: MethodSetRecord[];
  implementationRecords: ImplementationRecord[];
}

export interface TypeCheckerOptions {
  /**
   * When true, the checker will attempt to continue after diagnostics instead of
   * aborting immediately. The checker currently always continues.
   */
  continueAfterDiagnostics?: boolean;
  /**
   * Package summaries collected from previously-checked modules. Used to
   * resolve imports and surface package metadata to consumers.
   */
  packageSummaries?: Map<string, PackageSummary> | Record<string, PackageSummary>;
  /**
   * Prelude of already-checked symbols and methods that should be available to
   * the current module (e.g., stdlib packages loaded earlier in the session).
   */
  prelude?: TypeCheckerPrelude;
}

export const RESERVED_TYPE_NAMES = new Set<string>([
  "Self",
  "bool",
  "String",
  "IoHandle",
  "ProcHandle",
  "char",
  "nil",
  "void",
  "i8",
  "i16",
  "i32",
  "i64",
  "i128",
  "isize",
  "u8",
  "u16",
  "u32",
  "u64",
  "u128",
  "usize",
  "f32",
  "f64",
  "Array",
  "Map",
  "Range",
  "Iterator",
  "Result",
  "Option",
  "Proc",
  "Future",
  "Channel",
  "Mutex",
  "Error",
]);

export function clonePrelude(prelude?: TypeCheckerPrelude): TypeCheckerPrelude | undefined {
  if (!prelude) {
    return undefined;
  }
  return {
    structs: new Map(prelude.structs ?? new Map()),
    interfaces: new Map(prelude.interfaces ?? new Map()),
    typeAliases: new Map(prelude.typeAliases ?? new Map()),
    unions: new Map(prelude.unions ?? new Map()),
    functionInfos: cloneFunctionInfoMap(prelude.functionInfos),
    methodSets: [...(prelude.methodSets ?? [])],
    implementationRecords: [...(prelude.implementationRecords ?? [])],
  };
}

export function cloneFunctionInfoMap(source?: Map<string, FunctionInfo[]>): Map<string, FunctionInfo[]> {
  const cloned = new Map<string, FunctionInfo[]>();
  if (!source) {
    return cloned;
  }
  for (const [key, infos] of source.entries()) {
    cloned.set(key, [...infos]);
  }
  return cloned;
}
