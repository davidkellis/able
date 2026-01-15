import type * as AST from "../ast";

export type RuntimeCallFrame = {
  node?: AST.FunctionCall | null;
};

export type RuntimeDiagnosticContext = {
  node?: AST.AstNode;
  callStack: RuntimeCallFrame[];
};

const RUNTIME_CONTEXT_KEY = "__able_runtime_context";

export function attachRuntimeDiagnosticContext(err: unknown, context: RuntimeDiagnosticContext): void {
  if (!err || typeof err !== "object") return;
  const target = err as Record<string, unknown>;
  if (target[RUNTIME_CONTEXT_KEY]) return;
  target[RUNTIME_CONTEXT_KEY] = context;
}

export function getRuntimeDiagnosticContext(err: unknown): RuntimeDiagnosticContext | undefined {
  if (!err || typeof err !== "object") return undefined;
  const target = err as Record<string, unknown>;
  return target[RUNTIME_CONTEXT_KEY] as RuntimeDiagnosticContext | undefined;
}
