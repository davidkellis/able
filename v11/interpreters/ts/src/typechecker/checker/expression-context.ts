import type * as AST from "../../ast";
import type { TypeInfo } from "../types";
import type { InterfaceCheckResult } from "./types";

export type TypeDeclarationNode =
  | AST.StructDefinition
  | AST.UnionDefinition
  | AST.InterfaceDefinition
  | AST.TypeAliasDefinition;

export interface ExpressionContext {
  resolveStructDefinitionForPattern(
    pattern: AST.StructPattern,
    valueType: TypeInfo,
  ): AST.StructDefinition | undefined;
  pushLoopContext(): void;
  popLoopContext(): TypeInfo;
  inLoopContext(): boolean;
  pushBreakpointLabel(label: string): void;
  popBreakpointLabel(): void;
  hasBreakpointLabel(label: string): boolean;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  report(message: string, node?: AST.Node | null | undefined): void;
  describeTypeExpression(expr: AST.TypeExpression | null | undefined): string | null;
  typeInfosEquivalent(a?: TypeInfo, b?: TypeInfo): boolean;
  isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  resolveTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, TypeInfo>): TypeInfo;
  getStructDefinition(name: string): AST.StructDefinition | undefined;
  handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  checkStatement(node: AST.Statement | AST.Expression | undefined | null): void;
  pushAsyncContext(): void;
  popAsyncContext(): void;
  checkFunctionCall(call: AST.FunctionCall): void;
  inferFunctionCallReturnType(call: AST.FunctionCall): TypeInfo;
  checkFunctionDefinition(definition: AST.FunctionDefinition): void;
  checkReturnStatement(statement: AST.ReturnStatement): void;
  pushScope(): void;
  popScope(): void;
  withForkedEnv<T>(fn: () => T): T;
  lookupIdentifier(name: string): TypeInfo | undefined;
  defineValue(name: string, valueType: TypeInfo): void;
  assignValue(name: string, valueType: TypeInfo): boolean;
  hasBinding(name: string): boolean;
  hasBindingInCurrentScope(name: string): boolean;
  allowDynamicLookup(): boolean;
  typeImplementsInterface?(
    type: TypeInfo,
    interfaceName: string,
    expectedArgs?: string[],
  ): InterfaceCheckResult;
}

export type StatementContext = ExpressionContext & {
  isExpression(node: AST.Node | undefined | null): node is AST.Expression;
  handleTypeDeclaration?(node: TypeDeclarationNode): void;
  handleBreakStatement?(node: AST.BreakStatement): void;
  handleContinueStatement?(node: AST.ContinueStatement): void;
};
