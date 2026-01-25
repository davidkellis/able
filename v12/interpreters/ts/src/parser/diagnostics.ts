import type { DiagnosticLocation, DiagnosticNote, DiagnosticSeverity } from "../typechecker/diagnostics";
import type { Node } from "./shared";
import { MapperError } from "./shared";

export type ParserDiagnostic = {
  severity: DiagnosticSeverity;
  message: string;
  code?: string;
  location?: DiagnosticLocation;
  notes?: DiagnosticNote[];
};

export class ParserDiagnosticError extends Error {
  readonly diagnostic: ParserDiagnostic;

  constructor(diagnostic: ParserDiagnostic) {
    super(diagnostic.message);
    this.name = "ParserDiagnosticError";
    this.diagnostic = diagnostic;
  }
}

export function buildSyntaxErrorDiagnostic(root: Node, origin?: string): ParserDiagnostic {
  const missing = findFirstMissingNode(root);
  const errorNode = missing ?? findFirstErrorNode(root) ?? root;
  const location = toDiagnosticLocation(errorNode, origin);
  const expected = missing ? formatExpectedKind(missing.type) : null;
  const message = expected ? `parser: syntax error: expected ${expected}` : "parser: syntax error";
  return {
    severity: "error",
    message,
    location,
  };
}

export function buildMapperErrorDiagnostic(error: MapperError, origin?: string): ParserDiagnostic {
  const location = error.node ? toDiagnosticLocation(error.node, error.origin ?? origin) : origin ? { path: origin } : undefined;
  return {
    severity: "error",
    message: error.message,
    location,
  };
}

export function buildParserDiagnostic(error: unknown, origin?: string): ParserDiagnostic {
  if (error instanceof ParserDiagnosticError) {
    return error.diagnostic;
  }
  if (error instanceof MapperError) {
    return buildMapperErrorDiagnostic(error, origin);
  }
  if (error instanceof Error) {
    return {
      severity: "error",
      message: error.message,
      location: origin ? { path: origin } : undefined,
    };
  }
  return {
    severity: "error",
    message: String(error),
    location: origin ? { path: origin } : undefined,
  };
}

function toDiagnosticLocation(node: Node | null | undefined, origin?: string): DiagnosticLocation | undefined {
  if (!node && !origin) return undefined;
  if (!node) return origin ? { path: origin } : undefined;
  const start = node.startPosition;
  const end = node.endPosition;
  return {
    path: origin,
    line: start.row + 1,
    column: start.column + 1,
    endLine: end.row + 1,
    endColumn: end.column + 1,
  };
}

function findFirstMissingNode(root: Node): Node | null {
  let best: Node | null = null;
  walkNodes(root, (node) => {
    if (!nodeIsMissing(node)) return;
    if (!best || node.startIndex < best.startIndex) {
      best = node;
    }
  });
  return best;
}

function findFirstErrorNode(root: Node): Node | null {
  let best: Node | null = null;
  walkNodes(root, (node) => {
    if (!nodeIsError(node)) return;
    if (!best || node.startIndex < best.startIndex) {
      best = node;
    }
  });
  return best;
}

function walkNodes(root: Node, visit: (node: Node) => void): void {
  visit(root);
  for (let i = 0; i < root.childCount; i += 1) {
    const child = root.child(i);
    if (child) {
      walkNodes(child, visit);
    }
  }
}

function nodeIsMissing(node: Node): boolean {
  return Boolean((node as unknown as { isMissing?: boolean }).isMissing);
}

function nodeIsError(node: Node): boolean {
  const asAny = node as unknown as { isError?: boolean };
  return node.type === "ERROR" || Boolean(asAny.isError);
}

function formatExpectedKind(kind: string): string {
  const trimmed = kind.trim();
  if (!trimmed) {
    return "token";
  }
  const isSymbol = /^[^A-Za-z0-9_]+$/.test(trimmed);
  if (trimmed.length === 1 || isSymbol) {
    return `'${trimmed}'`;
  }
  return trimmed.replace(/_/g, " ");
}
