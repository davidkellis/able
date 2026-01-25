import type * as AST from "../../ast";
import type { TypeInfo } from "../types";
import type { DeclarationsContext } from "./declarations";
import type { ImplementationRecord, MethodSetRecord } from "./types";

export interface ImplementationContext extends DeclarationsContext {
  getTypeAlias?(name: string): AST.TypeAliasDefinition | undefined;
  getUnionDefinition?(name: string): AST.UnionDefinition | undefined;
  formatImplementationTarget(expr: AST.TypeExpression | null | undefined): string | null;
  formatImplementationLabel(interfaceName: string, targetLabel: string): string;
  registerMethodSet(record: MethodSetRecord): void;
  getMethodSets(): Iterable<MethodSetRecord>;
  registerImplementationRecord(record: ImplementationRecord): void;
  getImplementationRecords(): Iterable<ImplementationRecord>;
  getImplementationBucket(key: string): ImplementationRecord[] | undefined;
  describeTypeArgument(type: TypeInfo): string;
  appendInterfaceArgsToLabel(label: string, args: string[]): string;
  formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string;
}
