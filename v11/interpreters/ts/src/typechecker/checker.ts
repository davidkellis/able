import type * as AST from "../ast";
import { Environment } from "./environment";
import {
  describe,
  isBoolean,
  isNumeric,
  isUnknown,
  formatType,
  primitiveType,
  rangeType,
  procType,
  futureType,
  unknownType,
  type TypeInfo,
  type IntegerPrimitive,
} from "./types";
import {
  cloneFunctionInfoMap,
  clonePrelude,
  LocalTypeDeclaration,
  RESERVED_TYPE_NAMES,
  TypeCheckerOptions,
  TypeCheckerPrelude,
} from "./checker/core";
import type { StatementContext } from "./checker/expression-context";
import {
  buildCheckerContext,
  buildImportContext,
  buildImplementationContext,
  buildRegistryContext,
  buildTypeResolutionContext,
} from "./checker/context-builders";
import { collectFunctionDefinition as collectFunctionDefinitionHelper, type DeclarationsContext } from "./checker/declarations";
import {
  collectImplementationDefinition as collectImplementationDefinitionHelper,
  collectMethodsDefinition as collectMethodsDefinitionHelper,
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
  typeImplementsInterface,
  type ImplementationContext,
} from "./checker/implementations";
import { buildPackageSummary as buildPackageSummaryHelper, resolvePackageName } from "./checker/summary";
import {
  ensureUniqueDeclaration as ensureUniqueDeclarationHelper,
  formatNodeOrigin as formatNodeOriginHelper,
  registerImplementationRecord as registerImplementationRecordHelper,
  registerInterfaceDefinition as registerInterfaceDefinitionHelper,
  registerStructDefinition as registerStructDefinitionHelper,
  registerTypeAlias as registerTypeAliasHelper,
} from "./checker/registry";
import {
  FunctionContext,
  FunctionInfo,
  ImplementationObligation,
  ImplementationRecord,
  InterfaceCheckResult,
  MethodSetRecord,
  extractLocation,
} from "./checker/types";
export type {
  ImplementationObligation,
  ImplementationRecord,
  MethodSetRecord,
  FunctionContext,
  FunctionInfo,
} from "./checker/types";
export type { TypeCheckerOptions, TypeCheckerPrelude } from "./checker/core";
import type { DiagnosticLocation, TypecheckerDiagnostic, TypecheckResult, PackageSummary } from "./diagnostics";
import { createTypeResolutionHelpers, type TypeResolutionHelpers } from "./checker/type-resolution";
import { installBuiltins as installBuiltinsHelper } from "./checker/builtins";
import {
  applyDynImportStatement as applyDynImportStatementHelper,
  applyImportStatement as applyImportStatementHelper,
  applyImports as applyImportsHelper,
  clonePackageSummaries,
  handlePackageMemberAccess as handlePackageMemberAccessHelper,
} from "./checker/imports";
import {
  checkFunctionCall as checkFunctionCallHelper,
  inferFunctionCallReturnType as inferFunctionCallReturnTypeHelper,
} from "./checker/function-calls";

type FunctionGenericContext = {
  label: string;
  inferred: Map<string, AST.GenericParameter>;
};

export class TypeChecker {
  private env: Environment;
  private readonly options: TypeCheckerOptions;
  private readonly packageSummaries: Map<string, PackageSummary>;
  private readonly prelude?: TypeCheckerPrelude;
  private diagnostics: TypecheckerDiagnostic[] = [];
  private structDefinitions: Map<string, AST.StructDefinition> = new Map();
  private interfaceDefinitions: Map<string, AST.InterfaceDefinition> = new Map();
  private typeAliases: Map<string, AST.TypeAliasDefinition> = new Map();
  private unionDefinitions: Map<string, AST.UnionDefinition> = new Map();
  private functionInfos: Map<string, FunctionInfo[]> = new Map();
  private methodSets: MethodSetRecord[] = [];
  private implementationRecords: ImplementationRecord[] = [];
  private implementationIndex: Map<string, ImplementationRecord[]> = new Map();
  private declarationOrigins: Map<string, AST.Node> = new Map();
  private functionGenericStack: FunctionGenericContext[] = [];
  private packageAliases: Map<string, string> = new Map();
  private reportedPackageMemberAccess = new WeakSet<AST.MemberAccessExpression>();
  private asyncDepth = 0;
  private returnTypeStack: TypeInfo[] = [];
  private loopResultStack: TypeInfo[] = [];
  private breakpointStack: string[] = [];
  private allowDynamicLookups = false;
  private currentPackageName = "<anonymous>";
  private readonly typeResolution: TypeResolutionHelpers;
  private readonly context: StatementContext;
  private readonly declarationsContext: DeclarationsContext;
  private readonly implementationContext: ImplementationContext;

  constructor(options: TypeCheckerOptions = {}) {
    this.env = new Environment();
    this.options = options;
    this.packageSummaries = clonePackageSummaries(options.packageSummaries);
    this.prelude = clonePrelude(options.prelude);
    this.typeResolution = createTypeResolutionHelpers(buildTypeResolutionContext(this));
    this.context = buildCheckerContext(this);
    this.declarationsContext = this.context as DeclarationsContext;
    this.implementationContext = buildImplementationContext(this);
  }

  checkModule(module: AST.Module): TypecheckResult {
    this.env = new Environment();
    this.diagnostics = [];
    this.structDefinitions = new Map();
    this.interfaceDefinitions = new Map();
    this.typeAliases = new Map();
    this.unionDefinitions = new Map();
    this.functionInfos = new Map();
    this.methodSets = [];
    this.implementationRecords = [];
    this.implementationIndex = new Map();
    this.declarationOrigins = new Map();
    this.functionGenericStack = [];
    this.loopResultStack = [];
    this.breakpointStack = [];
    this.installBuiltins();
    this.installPrelude();
    this.packageAliases.clear();
    this.reportedPackageMemberAccess = new WeakSet();
    this.allowDynamicLookups = false;
    this.currentPackageName = resolvePackageName(module);
    this.applyImports(module);
    this.collectModuleDeclarations(module);

    if (Array.isArray(module.body)) {
      for (const statement of module.body) {
        this.checkStatement(statement as AST.Statement | AST.Expression);
      }
    }

    const summary = this.buildPackageSummary(module);
    return { diagnostics: [...this.diagnostics], summary };
  }

  private installBuiltins(): void {
    installBuiltinsHelper({
      env: this.env,
      functionInfos: this.functionInfos,
      implementationContext: this.implementationContext,
      registerStructDefinition: (definition) => this.registerStructDefinition(definition),
      registerTypeAlias: (definition) => this.registerTypeAlias(definition),
      registerInterfaceDefinition: (definition) => this.registerInterfaceDefinition(definition),
      collectImplementationDefinition: (definition) =>
        collectImplementationDefinitionHelper(this.implementationContext, definition),
      collectMethodsDefinition: (definition) => collectMethodsDefinitionHelper(this.implementationContext, definition),
    });
  }

  private installPrelude(): void {
    if (!this.prelude) {
      return;
    }
    for (const definition of this.prelude.structs.values()) {
      this.registerStructDefinition(definition);
    }
    for (const definition of this.prelude.interfaces.values()) {
      this.registerInterfaceDefinition(definition);
    }
    for (const definition of this.prelude.unions.values()) {
      this.registerUnionDefinition(definition);
    }
    for (const definition of this.prelude.typeAliases.values()) {
      this.registerTypeAlias(definition);
    }
    for (const [key, infos] of this.prelude.functionInfos.entries()) {
      for (const info of infos) {
        this.addFunctionInfo(key, info);
      }
    }
    this.methodSets.push(...this.prelude.methodSets);
    for (const record of this.prelude.implementationRecords) {
      this.registerImplementationRecord(record);
    }
  }

  private collectModuleDeclarations(module: AST.Module): void {
    if (!Array.isArray(module.body)) {
      return;
    }
    const statements = module.body as Array<AST.Statement | AST.Expression>;
      for (const statement of statements) {
        this.registerPrimaryTypeDeclaration(statement);
      }
      for (const statement of statements) {
        if (
          statement &&
          (statement.type === "StructDefinition" ||
            statement.type === "InterfaceDefinition" ||
            statement.type === "TypeAliasDefinition" ||
            statement.type === "UnionDefinition")
        ) {
          continue;
        }
        this.collectPrimaryDeclaration(statement);
      }
    for (const statement of statements) {
      this.collectImplementationDeclaration(statement);
    }
  }

  private registerPrimaryTypeDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) return;
    switch (node.type) {
      case "StructDefinition":
        this.registerStructDefinition(node);
        break;
      case "InterfaceDefinition":
        this.registerInterfaceDefinition(node);
        break;
      case "UnionDefinition":
        this.registerUnionDefinition(node);
        break;
      case "TypeAliasDefinition":
        this.registerTypeAlias(node);
        break;
      default:
        break;
    }
  }

  private collectPrimaryDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) return;
    switch (node.type) {
      case "InterfaceDefinition":
        this.registerInterfaceDefinition(node);
        break;
      case "UnionDefinition":
        this.registerUnionDefinition(node);
        break;
      case "TypeAliasDefinition":
        this.registerTypeAlias(node);
        break;
      case "StructDefinition":
        this.registerStructDefinition(node);
        break;
      case "ExternFunctionBody":
        this.collectExternFunctionBody(node);
        break;
      case "MethodsDefinition":
        collectMethodsDefinitionHelper(this.implementationContext, node);
        break;
      case "FunctionDefinition":
        if (node.id?.name && !node.isMethodShorthand) {
          if (!this.ensureUniqueDeclaration(node.id.name, node)) {
            return;
          }
        }
        this.collectFunctionDefinition(node, undefined);
        break;
      default:
        break;
    }
  }

  private collectImplementationDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) return;
    if (node.type === "ImplementationDefinition") {
      collectImplementationDefinitionHelper(this.implementationContext, node);
    }
  }

  private checkStatement(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) {
      return;
    }
    if (node.type === "ImportStatement") {
      this.applyImportStatement(node);
      return;
    }
    if (node.type === "DynImportStatement") {
      this.applyDynImportStatement(node);
      return;
    }
    this.context.checkStatement(node);
  }

  private ensureUniqueDeclaration(name: string | null | undefined, node: AST.Node | null | undefined): boolean {
    return ensureUniqueDeclarationHelper(this.registryContext(), name, node);
  }

  private isKnownTypeName(name: string | null | undefined): boolean {
    if (!name) {
      return false;
    }
    if (RESERVED_TYPE_NAMES.has(name)) {
      return true;
    }
    if (this.structDefinitions.has(name)) {
      return true;
    }
    if (this.unionDefinitions.has(name)) {
      return true;
    }
    if (this.interfaceDefinitions.has(name)) {
      return true;
    }
    if (this.typeAliases.has(name)) {
      return true;
    }
    for (const record of this.methodSets) {
      const targetName = this.getIdentifierNameFromTypeExpression(record.target);
      if (targetName === name) {
        return true;
      }
    }
    for (const summary of this.packageSummaries.values()) {
      if (summary.structs?.[name] || summary.interfaces?.[name] || summary.symbols?.[name]) {
        return true;
      }
    }
    return false;
  }

  private formatNodeOrigin(node: AST.Node | null | undefined): string {
    return formatNodeOriginHelper(node);
  }

  private registerStructDefinition(definition: AST.StructDefinition): void {
    registerStructDefinitionHelper(this.registryContext(), definition);
  }

  private registerInterfaceDefinition(definition: AST.InterfaceDefinition): void {
    registerInterfaceDefinitionHelper(this.registryContext(), definition);
  }

  private registerTypeAlias(definition: AST.TypeAliasDefinition): void {
    registerTypeAliasHelper(this.registryContext(), definition);
  }

  private registerUnionDefinition(definition: AST.UnionDefinition): void {
    const name = definition.id?.name;
    if (!name) {
      return;
    }
    if (!this.ensureUniqueDeclaration(name, definition)) {
      return;
    }
    this.unionDefinitions.set(name, definition);
  }

  private registerImplementationRecord(record: ImplementationRecord): void {
    registerImplementationRecordHelper(this.registryContext(), record);
  }

  private collectFunctionDefinition(
    definition: AST.FunctionDefinition,
    context: FunctionContext | undefined,
  ): void {
    collectFunctionDefinitionHelper(this.declarationsContext, definition, context);
  }

  private collectExternFunctionBody(extern: AST.ExternFunctionBody): void {
    const name = extern.signature?.id?.name;
    if (!name) {
      return;
    }
    const existing = this.declarationOrigins.get(name);
    if (!existing) {
      this.declarationOrigins.set(name, extern);
      this.collectFunctionDefinition(extern.signature, undefined);
      return;
    }
    if (!this.env.has(name)) {
      this.collectFunctionDefinition(extern.signature, undefined);
    }
  }

  private formatImplementationLabel(interfaceName: string, targetName: string): string {
    return `impl ${interfaceName} for ${targetName}`;
  }

  private formatImplementationTarget(targetType: AST.TypeExpression | null | undefined): string | null {
    if (!targetType) return null;
    return this.formatTypeExpression(targetType);
  }

  private inferExpression(expression: AST.Expression | undefined | null): TypeInfo {
    return this.context.inferExpression(expression);
  }

  private withForkedEnv<T>(fn: () => T): T {
    const previousEnv = this.env;
    this.env = this.env.fork();
    try {
      return fn();
    } finally {
      this.env = previousEnv;
    }
  }

  private pushAsyncContext(): void {
    this.asyncDepth += 1;
  }

  private popAsyncContext(): void {
    if (this.asyncDepth > 0) {
      this.asyncDepth -= 1;
    }
  }

  private inAsyncContext(): boolean {
    return this.asyncDepth > 0;
  }

  private pushReturnType(type: TypeInfo): void {
    this.returnTypeStack.push(type ?? unknownType);
  }

  private popReturnType(): void {
    if (this.returnTypeStack.length > 0) {
      this.returnTypeStack.pop();
    }
  }

  private currentReturnType(): TypeInfo | undefined {
    if (!this.returnTypeStack.length) return undefined;
    return this.returnTypeStack[this.returnTypeStack.length - 1];
  }

  private getBuiltinCallName(callee: AST.Expression | undefined | null): string | undefined {
    if (!callee) return undefined;
    if (callee.type === "Identifier") {
      return callee.name;
    }
    return undefined;
  }

  private checkBuiltinCallContext(name: string | undefined, call: AST.FunctionCall): void {
    if (!name) return;
    switch (name) {
      case "proc_yield":
        if (!this.inAsyncContext()) {
          this.report("typechecker: proc_yield() may only be called from within proc or spawn bodies", call);
        }
        break;
      default:
        break;
    }
  }

  private checkFunctionCall(call: AST.FunctionCall): void {
    checkFunctionCallHelper(this.functionCallContext(), call);
  }

  private checkFunctionDefinition(definition: AST.FunctionDefinition): void {
    if (!definition) return;
    const name = definition.id?.name ?? "<anonymous>";
    const info = this.getFunctionInfos(name)[0];
    const paramTypes = Array.isArray(definition.params)
      ? definition.params.map((param) => this.resolveTypeExpression(param?.paramType))
      : [];
    const expectedReturn =
      (info?.returnType && info.returnType.kind !== "unknown" && info.returnType) ||
      this.resolveTypeExpression(definition.returnType);
    this.pushReturnType(expectedReturn ?? unknownType);
    this.pushFunctionGenericContext(definition);
    this.env.pushScope();
    try {
      if (Array.isArray(definition.params)) {
        definition.params.forEach((param, index) => {
          const paramName = this.getIdentifierName(param?.name);
          if (!paramName) return;
          const paramType = paramTypes[index] ?? unknownType;
          this.env.define(paramName, paramType ?? unknownType);
        });
      }
      const bodyType = this.inferExpression(definition.body);
      if (expectedReturn && expectedReturn.kind !== "unknown" && bodyType && bodyType.kind !== "unknown") {
        const literalMessage = this.describeLiteralMismatch(bodyType, expectedReturn);
        if (literalMessage) {
          this.report(literalMessage, definition.body ?? definition);
        } else if (!this.isTypeAssignable(bodyType, expectedReturn)) {
          this.report(
            `typechecker: function '${name}' body returns ${formatType(bodyType)}, expected ${formatType(expectedReturn)}`,
            definition.body ?? definition,
          );
        }
      }
    } finally {
      this.env.popScope();
      this.popFunctionGenericContext();
      this.popReturnType();
    }
  }

  private checkReturnStatement(statement: AST.ReturnStatement): void {
    if (!statement) return;
    const expected = this.currentReturnType();
    const actual = statement.argument ? this.inferExpression(statement.argument) : primitiveType("nil");
    if (!expected) {
      this.report("typechecker: return statement outside function", statement);
      return;
    }
    if (expected.kind === "unknown") {
      return;
    }
    if (!actual || actual.kind === "unknown") {
      return;
    }
    const literalMessage = this.describeLiteralMismatch(actual, expected);
    if (literalMessage) {
      this.report(literalMessage, statement.argument ?? statement);
      return;
    }
    if (expected.kind === "result") {
      if (this.isTypeAssignable(actual, expected.inner)) {
        return;
      }
      if (typeImplementsInterface(this.implementationContext, actual, "Error", [] as string[]).ok) {
        return;
      }
    }
    const unionMatches = (member: TypeInfo | undefined): boolean => {
      if (!member) return false;
      if (member.kind === "union") {
        return member.members?.some((inner) => unionMatches(inner)) ?? false;
      }
      if (member.kind === "interface") {
        const args = (member.typeArguments ?? []).map((arg) => formatType(arg));
        return typeImplementsInterface(this.implementationContext, actual, member.name, args).ok;
      }
      if (member.kind === "struct" && member.name === "Error") {
        return typeImplementsInterface(this.implementationContext, actual, "Error", []).ok;
      }
      return false;
    };
    if (expected.kind === "union" && expected.members?.some((member) => unionMatches(member))) {
      return;
    }
    if (!this.isTypeAssignable(actual, expected)) {
      this.report(
        `typechecker: return expects ${formatType(expected)}, got ${formatType(actual)}`,
        statement.argument ?? statement,
      );
    }
  }

  private pushLoopContext(): void {
    this.loopResultStack.push(unknownType);
  }

  private popLoopContext(): TypeInfo {
    if (this.loopResultStack.length === 0) {
      return unknownType;
    }
    return this.loopResultStack.pop() ?? unknownType;
  }

  private inLoopContext(): boolean {
    return this.loopResultStack.length > 0;
  }

  private recordLoopBreakType(breakType: TypeInfo): void {
    if (this.loopResultStack.length === 0) {
      return;
    }
    const idx = this.loopResultStack.length - 1;
    const normalized = breakType ?? primitiveType("nil");
    const current = this.loopResultStack[idx];
    if (!normalized || normalized.kind === "unknown") {
      this.loopResultStack[idx] = unknownType;
      return;
    }
    if (!current || current.kind === "unknown") {
      this.loopResultStack[idx] = normalized;
      return;
    }
    if (!this.typeInfosEquivalent(current, normalized)) {
      this.loopResultStack[idx] = unknownType;
    }
  }

  private pushBreakpointLabel(label: string | null | undefined): void {
    if (!label) {
      return;
    }
    this.breakpointStack.push(label);
  }

  private popBreakpointLabel(): void {
    if (this.breakpointStack.length === 0) {
      return;
    }
    this.breakpointStack.pop();
  }

  private hasBreakpointLabel(label: string | null | undefined): boolean {
    if (!label) {
      return false;
    }
    for (let index = this.breakpointStack.length - 1; index >= 0; index -= 1) {
      if (this.breakpointStack[index] === label) {
        return true;
      }
    }
    return false;
  }

  private checkBreakStatement(statement: AST.BreakStatement): void {
    if (!statement) {
      return;
    }
    const hasLabel = Boolean(statement.label);
    const labelName = statement.label ? this.getIdentifierName(statement.label) : null;
    const inLoop = this.inLoopContext();
    if (!inLoop && !hasLabel) {
      this.report("typechecker: break statement must appear inside a loop", statement);
    }
    if (hasLabel) {
      if (!labelName) {
        this.report("typechecker: break label cannot be empty", statement.label ?? statement);
      } else if (!this.hasBreakpointLabel(labelName)) {
        this.report(`typechecker: unknown break label '${labelName}'`, statement);
      }
    }
    const valueType = statement.value ? this.inferExpression(statement.value) : primitiveType("nil");
    if (inLoop && !hasLabel) {
      this.recordLoopBreakType(valueType);
    }
  }

  private checkContinueStatement(statement: AST.ContinueStatement): void {
    if (!statement) {
      return;
    }
    if (!this.inLoopContext()) {
      this.report("typechecker: continue statement must appear inside a loop", statement);
    }
    if (statement.label) {
      this.report("typechecker: labeled continue is not supported", statement);
    }
  }

  private inferFunctionCallReturnType(call: AST.FunctionCall): TypeInfo {
    return inferFunctionCallReturnTypeHelper(this.functionCallContext(), call);
  }

  private functionCallContext() {
    return {
      implementationContext: this.implementationContext,
      functionInfos: this.functionInfos,
      structDefinitions: this.structDefinitions,
      inferExpression: (expression: AST.Expression | undefined | null) => this.inferExpression(expression),
      resolveTypeExpression: (expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, TypeInfo>) =>
        this.resolveTypeExpression(expr, substitutions),
      describeLiteralMismatch: this.describeLiteralMismatch.bind(this),
      isTypeAssignable: this.isTypeAssignable.bind(this),
      report: this.report.bind(this),
      handlePackageMemberAccess: this.handlePackageMemberAccess.bind(this),
      getIdentifierName: this.getIdentifierName.bind(this),
      checkBuiltinCallContext: this.checkBuiltinCallContext.bind(this),
      getBuiltinCallName: this.getBuiltinCallName.bind(this),
      typeImplementsInterface: (type, interfaceName, expectedArgs) =>
        typeImplementsInterface(this.implementationContext, type, interfaceName, expectedArgs ?? []),
      statementContext: this.context,
    };
  }

  private resolveStructDefinitionForPattern(pattern: AST.StructPattern, valueType: TypeInfo): AST.StructDefinition | undefined {
    if (valueType.kind === "struct" && valueType.definition) {
      return valueType.definition;
    }
    if (valueType.kind === "struct") {
      const definition = this.structDefinitions.get(valueType.name);
      if (definition) return definition;
    }
    const structName = this.getIdentifierName(pattern.structType);
    if (!structName) return undefined;
    return this.structDefinitions.get(structName);
  }

  private resolveTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    return this.typeResolution.resolveTypeExpression(expr, substitutions);
  }

  private instantiateTypeAlias(
    definition: AST.TypeAliasDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    return this.typeResolution.instantiateTypeAlias(definition, typeArguments, outerSubstitutions);
  }

  private typeInfosEquivalent(a: TypeInfo | undefined, b: TypeInfo | undefined): boolean {
    return this.typeResolution.typeInfosEquivalent(a, b);
  }

  private isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean {
    return this.typeResolution.isTypeAssignable(actual, expected);
  }

  public describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null {
    return this.typeResolution.describeLiteralMismatch(actual, expected);
  }

  private canonicalizeStructuralType(type: TypeInfo): TypeInfo {
    return this.typeResolution.canonicalizeStructuralType(type);
  }

  private typeExpressionsEquivalent(
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): boolean {
    return this.typeResolution.typeExpressionsEquivalent(a, b, substitutions);
  }

  private describeTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): string {
    return this.typeResolution.describeTypeExpression(expr, substitutions);
  }

  private formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string {
    return this.typeResolution.formatTypeExpression(expr, substitutions);
  }

  private lookupSubstitution(name: string | null, substitutions?: Map<string, string>): string {
    return this.typeResolution.lookupSubstitution(name, substitutions);
  }

  private describeTypeArgument(type: TypeInfo): string {
    return this.typeResolution.describeTypeArgument(type);
  }

  private appendInterfaceArgsToLabel(label: string, args: string[]): string {
    return this.typeResolution.appendInterfaceArgsToLabel(label, args);
  }

  private getInterfaceNameFromConstraint(constraint: AST.InterfaceConstraint | null | undefined): string | null {
    if (!constraint) return null;
    return this.getInterfaceNameFromTypeExpression(constraint.interfaceType);
  }

  private getInterfaceNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null {
    if (!expr) return null;
    switch (expr.type) {
      case "SimpleTypeExpression":
        return this.getIdentifierName(expr.name);
      case "GenericTypeExpression":
        return this.getInterfaceNameFromTypeExpression(expr.base);
      default:
        return null;
    }
  }

  private getIdentifierName(node: AST.Identifier | null | undefined): string | null {
    if (!node) return null;
    return node.name ?? null;
  }

  private getIdentifierNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null {
    if (!expr) return null;
    if (expr.type === "SimpleTypeExpression") {
      return this.getIdentifierName(expr.name);
    }
    if (expr.type === "GenericTypeExpression") {
      return this.getIdentifierNameFromTypeExpression(expr.base);
    }
    return null;
  }

  private isExpression(node: AST.Statement | AST.Expression): node is AST.Expression {
    switch (node.type) {
      case "StructDefinition":
      case "FunctionDefinition":
      case "ImplementationDefinition":
      case "TypeAliasDefinition":
      case "ImportStatement":
      case "DynImportStatement":
      case "MethodsDefinition":
      case "PreludeStatement":
      case "ReturnStatement":
      case "RaiseStatement":
      case "RethrowStatement":
      case "BreakStatement":
      case "ContinueStatement":
        return false;
      default:
        return true;
    }
  }

  private report(message: string, node?: AST.Node | null): void {
    const location = extractLocation(node);
    const diagnostic: TypecheckerDiagnostic = { severity: "error", message };
    if (location) {
      diagnostic.location = location;
    }
    this.diagnostics.push(diagnostic);
  }

  private applyImports(module: AST.Module): void {
    applyImportsHelper(buildImportContext(this), module);
  }

  private registryContext() {
    return buildRegistryContext(this);
  }

  private applyImportStatement(imp: AST.ImportStatement | null | undefined): void {
    applyImportStatementHelper(buildImportContext(this), imp);
  }

  private applyDynImportStatement(statement: AST.DynImportStatement | null | undefined): void {
    if (statement?.isWildcard) {
      this.allowDynamicLookups = true;
      return;
    }
    applyDynImportStatementHelper(buildImportContext(this), statement);
  }

  private handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean {
    return handlePackageMemberAccessHelper(buildImportContext(this), expression);
  }

  private buildPackageSummary(module: AST.Module): PackageSummary | null {
    return buildPackageSummaryHelper(this.implementationContext, module);
  }

  exportPrelude(): TypeCheckerPrelude {
    return {
      structs: new Map(this.structDefinitions),
      interfaces: new Map(this.interfaceDefinitions),
      typeAliases: new Map(this.typeAliases),
      unions: new Map(this.unionDefinitions),
      functionInfos: cloneFunctionInfoMap(this.functionInfos),
      methodSets: [...this.methodSets],
      implementationRecords: [...this.implementationRecords],
    };
  }

  private addFunctionInfo(key: string, info: FunctionInfo): void {
    const existing = this.functionInfos.get(key) ?? [];
    if (existing.some((entry) => entry.fullName === info.fullName)) {
      return;
    }
    this.functionInfos.set(key, [...existing, info]);
  }

  private getFunctionInfos(key: string): FunctionInfo[] {
    return this.functionInfos.get(key) ?? [];
  }

  private pushFunctionGenericContext(definition: AST.FunctionDefinition): void {
    if (!definition) {
      return;
    }
    const inferred = this.collectInferredGenericParameters(definition);
    const label = definition.id?.name ? `fn ${definition.id.name}` : "fn <anonymous>";
    this.functionGenericStack.push({ label, inferred });
  }

  private popFunctionGenericContext(): void {
    if (this.functionGenericStack.length === 0) {
      return;
    }
    this.functionGenericStack.pop();
  }

  private collectInferredGenericParameters(definition: AST.FunctionDefinition): Map<string, AST.GenericParameter> {
    const inferred = new Map<string, AST.GenericParameter>();
    if (!definition) {
      return inferred;
    }
    const params = Array.isArray(definition.inferredGenericParams)
      ? definition.inferredGenericParams
      : definition.genericParams;
    if (!Array.isArray(params)) {
      return inferred;
    }
    for (const param of params) {
      if (!param?.isInferred) {
        continue;
      }
      const name = this.getIdentifierName(param.name);
      if (!name) {
        continue;
      }
      inferred.set(name, param);
    }
    return inferred;
  }

  private checkLocalTypeDeclaration(node: LocalTypeDeclaration): void {
    if (!node || this.functionGenericStack.length === 0) {
      return;
    }
    const current = this.functionGenericStack[this.functionGenericStack.length - 1];
    if (!current || current.inferred.size === 0) {
      return;
    }
    const name = this.getIdentifierName((node as { id?: AST.Identifier })?.id);
    if (!name) {
      return;
    }
    const param = current.inferred.get(name);
    if (!param) {
      return;
    }
    const location = this.formatNodeOrigin(param);
    this.report(
      `typechecker: cannot redeclare inferred type parameter '${name}' inside ${current.label} (inferred at ${location})`,
      node,
    );
  }
}

export function createTypeChecker(options?: TypeCheckerOptions): TypeChecker {
  return new TypeChecker(options);
}
