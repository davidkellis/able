import type * as AST from "../../ast";
import type { DiagnosticLocation } from "../diagnostics";
import type { TypeInfo } from "../types";

export interface ImplementationObligation {
  typeParam: string;
  interfaceName: string;
  interfaceType?: AST.TypeExpression;
  context: string;
}

export interface ImplementationRecord {
  interfaceName: string;
  label: string;
  target: AST.TypeExpression;
  targetKey: string;
  resolvedTarget?: TypeInfo;
  genericParams: string[];
  obligations: ImplementationObligation[];
  interfaceArgs: AST.TypeExpression[];
  unionVariants?: string[];
  definition: AST.ImplementationDefinition;
}

export interface InterfaceCheckResult {
  ok: boolean;
  detail?: string;
}

export interface MethodSetRecord {
  label: string;
  target: AST.TypeExpression;
  resolvedTarget?: TypeInfo;
  genericParams: string[];
  obligations: ImplementationObligation[];
  definition: AST.MethodsDefinition;
}

export type FunctionContext = {
  structName?: string;
  structBaseName?: string;
  typeParamNames?: string[];
  fromMethodSet?: boolean;
};

export interface FunctionInfo {
  name: string;
  fullName: string;
  definition?: AST.FunctionDefinition;
  structName?: string;
  hasImplicitSelf?: boolean;
  isTypeQualified?: boolean;
  typeQualifier?: string;
  exportedName?: string;
  methodResolutionPriority?: number;
  fromMethodSet?: boolean;
  parameters: TypeInfo[];
  genericConstraints: Array<{
    paramName: string;
    interfaceName: string;
    interfaceDefined: boolean;
    interfaceType?: AST.TypeExpression;
  }>;
  genericParamNames: string[];
  whereClause: ImplementationObligation[];
  methodSetSubstitutions?: Array<[string, TypeInfo]>;
  returnType: TypeInfo;
}

export function extractLocation(node: AST.Node | null | undefined): DiagnosticLocation | undefined {
  if (!node) {
    return undefined;
  }
  const anyNode = node as unknown as {
    span?: { start?: { line?: number; column?: number }; end?: { line?: number; column?: number } };
    origin?: string | { path?: string };
    path?: string;
  };
  const span = anyNode.span;
  const location: DiagnosticLocation = {};
  if (typeof anyNode.origin === "string" && anyNode.origin) {
    location.path = anyNode.origin;
  } else if (anyNode.origin && typeof anyNode.origin === "object" && anyNode.origin?.path) {
    location.path = anyNode.origin.path;
  } else if (typeof anyNode.path === "string" && anyNode.path) {
    location.path = anyNode.path;
  }
  if (
    span &&
    span.start &&
    typeof span.start.line === "number" &&
    typeof span.start.column === "number"
  ) {
    location.line = span.start.line;
    location.column = span.start.column;
  }
  if (
    span &&
    span.end &&
    typeof span.end.line === "number" &&
    typeof span.end.column === "number"
  ) {
    location.endLine = span.end.line;
    location.endColumn = span.end.column;
  }
  if (location.path || location.line !== undefined || location.column !== undefined) {
    return location;
  }
  return undefined;
}
