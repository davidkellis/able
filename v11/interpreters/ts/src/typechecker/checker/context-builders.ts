import type * as AST from "../../ast";
import type { TypeInfo } from "../types";
import { createCheckerContext as createCheckerContextHelper } from "./context";
import type { StatementContext } from "./expression-context";
import type { ImplementationContext } from "./implementations";
import { typeImplementsInterface } from "./implementations";
import type { FunctionInfo } from "./types";
import type { LocalTypeDeclaration } from "./core";

export function buildImportContext(checker: any) {
  return {
    env: checker.env,
    packageSummaries: checker.packageSummaries,
    packageAliases: checker.packageAliases,
    reportedPackageMemberAccess: checker.reportedPackageMemberAccess,
    currentPackageName: checker.currentPackageName,
    report: checker.report.bind(checker),
    getIdentifierName: checker.getIdentifierName.bind(checker),
  };
}

export function buildRegistryContext(checker: any) {
  return {
    structDefinitions: checker.structDefinitions,
    interfaceDefinitions: checker.interfaceDefinitions,
    typeAliases: checker.typeAliases,
    implementationRecords: checker.implementationRecords,
    implementationIndex: checker.implementationIndex,
    declarationOrigins: checker.declarationOrigins,
    declarationsContext: checker.declarationsContext,
    getIdentifierName: (identifier: AST.Identifier | null | undefined) => checker.getIdentifierName(identifier),
    report: checker.report.bind(checker),
  };
}

export function buildTypeResolutionContext(checker: any) {
  return {
    getTypeAlias: (name: string) => checker.typeAliases.get(name),
    getInterfaceDefinition: (name: string) => checker.interfaceDefinitions.get(name),
    hasInterfaceDefinition: (name: string) => checker.interfaceDefinitions.has(name),
    getStructDefinition: (name: string) => checker.structDefinitions.get(name),
    getUnionDefinition: (name: string) => checker.unionDefinitions.get(name),
    hasUnionDefinition: (name: string) => checker.unionDefinitions.has(name),
    getIdentifierName: (identifier: AST.Identifier | null | undefined) => checker.getIdentifierName(identifier),
  };
}

export function buildCheckerContext(checker: any): StatementContext {
  return createCheckerContextHelper({
    resolveStructDefinitionForPattern: checker.resolveStructDefinitionForPattern.bind(checker),
    getIdentifierName: checker.getIdentifierName.bind(checker),
    getIdentifierNameFromTypeExpression: checker.getIdentifierNameFromTypeExpression.bind(checker),
    getInterfaceNameFromConstraint: checker.getInterfaceNameFromConstraint.bind(checker),
    getInterfaceNameFromTypeExpression: checker.getInterfaceNameFromTypeExpression.bind(checker),
    report: checker.report.bind(checker),
    reportWarning: checker.reportWarning.bind(checker),
    describeTypeExpression: checker.describeTypeExpression.bind(checker),
    isKnownTypeName: (name: string) => checker.isKnownTypeName(name),
    hasTypeDefinition: (name: string) =>
      checker.structDefinitions.has(name) ||
      checker.unionDefinitions.has(name) ||
      checker.interfaceDefinitions.has(name) ||
      checker.typeAliases.has(name),
    typeInfosEquivalent: checker.typeInfosEquivalent.bind(checker),
    isTypeAssignable: checker.isTypeAssignable.bind(checker),
    describeLiteralMismatch: checker.describeLiteralMismatch.bind(checker),
    resolveTypeExpression: checker.resolveTypeExpression.bind(checker),
    normalizeUnionType: checker.normalizeUnionType.bind(checker),
    getStructDefinition: (name: string) => checker.structDefinitions.get(name),
    getInterfaceDefinition: (name: string) => checker.interfaceDefinitions.get(name),
    hasInterfaceDefinition: (name: string) => checker.interfaceDefinitions.has(name),
    handlePackageMemberAccess: checker.handlePackageMemberAccess.bind(checker),
    pushAsyncContext: checker.pushAsyncContext.bind(checker),
    popAsyncContext: checker.popAsyncContext.bind(checker),
    checkReturnStatement: checker.checkReturnStatement.bind(checker),
    checkFunctionCall: checker.checkFunctionCall.bind(checker),
    inferFunctionCallReturnType: checker.inferFunctionCallReturnType.bind(checker),
    checkFunctionDefinition: checker.checkFunctionDefinition.bind(checker),
    pushLoopContext: checker.pushLoopContext.bind(checker),
    popLoopContext: () => checker.popLoopContext(),
    inLoopContext: () => checker.inLoopContext(),
    pushScope: () => checker.env.pushScope(),
    popScope: () => checker.env.popScope(),
    withForkedEnv: <T>(fn: () => T) => checker.withForkedEnv(fn),
    lookupIdentifier: (name: string) => checker.env.lookup(name),
    defineValue: (name: string, valueType: TypeInfo) => checker.env.define(name, valueType),
    assignValue: (name: string, valueType: TypeInfo) => checker.env.assign(name, valueType),
    hasBinding: (name: string) => checker.env.has(name),
    hasBindingInCurrentScope: (name: string) => checker.env.hasInCurrentScope(name),
    allowDynamicLookup: () => checker.allowDynamicLookups,
    getFunctionInfos: (key: string) => checker.getFunctionInfos(key),
    addFunctionInfo: (key: string, info: FunctionInfo) => checker.addFunctionInfo(key, info),
    isExpression: (node: AST.Node | undefined | null): node is AST.Expression => checker.isExpression(node),
    handleTypeDeclaration: (node) => checker.checkLocalTypeDeclaration(node as LocalTypeDeclaration),
    pushBreakpointLabel: (label: string) => checker.pushBreakpointLabel(label),
    popBreakpointLabel: () => checker.popBreakpointLabel(),
    hasBreakpointLabel: (label: string) => checker.hasBreakpointLabel(label),
    handleBreakStatement: checker.checkBreakStatement.bind(checker),
    handleContinueStatement: checker.checkContinueStatement.bind(checker),
    typeImplementsInterface: (type, interfaceName, expectedArgs) =>
      typeImplementsInterface(checker.implementationContext, type, interfaceName, expectedArgs ?? []),
  });
}

export function buildImplementationContext(checker: any): ImplementationContext {
  const ctx = checker.declarationsContext as ImplementationContext;
  ctx.formatImplementationTarget = checker.formatImplementationTarget.bind(checker);
  ctx.formatImplementationLabel = checker.formatImplementationLabel.bind(checker);
  ctx.registerMethodSet = (record) => {
    checker.methodSets.push(record);
  };
  ctx.getMethodSets = () => checker.methodSets;
  ctx.registerImplementationRecord = (record) => checker.registerImplementationRecord(record);
  ctx.getImplementationRecords = () => checker.implementationRecords;
  ctx.getImplementationBucket = (key: string) => checker.implementationIndex.get(key);
  ctx.describeTypeArgument = checker.describeTypeArgument.bind(checker);
  ctx.appendInterfaceArgsToLabel = checker.appendInterfaceArgsToLabel.bind(checker);
  ctx.formatTypeExpression = checker.formatTypeExpression.bind(checker);
  ctx.getTypeAlias = (name: string) => checker.typeAliases.get(name);
  return ctx;
}
