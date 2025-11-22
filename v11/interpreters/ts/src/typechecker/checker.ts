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
import type { StatementContext } from "./checker/expression-context";
import { collectFunctionDefinition as collectFunctionDefinitionHelper, type DeclarationsContext } from "./checker/declarations";
import {
  collectImplementationDefinition as collectImplementationDefinitionHelper,
  collectMethodsDefinition as collectMethodsDefinitionHelper,
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
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
import { createCheckerContext as createCheckerContextHelper } from "./checker/context";
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

type LocalTypeDeclaration =
  | AST.StructDefinition
  | AST.UnionDefinition
  | AST.InterfaceDefinition
  | AST.TypeAliasDefinition;

type FunctionGenericContext = {
  label: string;
  inferred: Map<string, AST.GenericParameter>;
};

const RESERVED_TYPE_NAMES = new Set<string>([
  "Self",
  "bool",
  "string",
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
}

export class TypeChecker {
  private env: Environment;
  private readonly options: TypeCheckerOptions;
  private readonly packageSummaries: Map<string, PackageSummary>;
  private diagnostics: TypecheckerDiagnostic[] = [];
  private structDefinitions: Map<string, AST.StructDefinition> = new Map();
  private interfaceDefinitions: Map<string, AST.InterfaceDefinition> = new Map();
  private typeAliases: Map<string, AST.TypeAliasDefinition> = new Map();
  private functionInfos: Map<string, FunctionInfo> = new Map();
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
    this.typeResolution = createTypeResolutionHelpers(this.createTypeResolutionContext());
    this.context = this.createCheckerContext();
    this.declarationsContext = this.context as DeclarationsContext;
    this.implementationContext = this.createImplementationContext();
  }

  checkModule(module: AST.Module): TypecheckResult {
    this.env = new Environment();
    this.diagnostics = [];
    this.structDefinitions = new Map();
    this.interfaceDefinitions = new Map();
    this.typeAliases = new Map();
    this.functionInfos = new Map();
    this.methodSets = [];
    this.implementationRecords = [];
    this.implementationIndex = new Map();
    this.declarationOrigins = new Map();
    this.functionGenericStack = [];
    this.loopResultStack = [];
    this.breakpointStack = [];
    this.installBuiltins();
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
      registerInterfaceDefinition: (definition) => this.registerInterfaceDefinition(definition),
      collectImplementationDefinition: (definition) =>
        collectImplementationDefinitionHelper(this.implementationContext, definition),
    });
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
          statement.type === "TypeAliasDefinition")
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
      case "TypeAliasDefinition":
        this.registerTypeAlias(node);
        break;
      case "StructDefinition":
        this.registerStructDefinition(node);
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
    if (this.interfaceDefinitions.has(name)) {
      return true;
    }
    if (this.typeAliases.has(name)) {
      return true;
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

  private registerImplementationRecord(record: ImplementationRecord): void {
    registerImplementationRecordHelper(this.registryContext(), record);
  }

  private collectFunctionDefinition(
    definition: AST.FunctionDefinition,
    context: FunctionContext | undefined,
  ): void {
    collectFunctionDefinitionHelper(this.declarationsContext, definition, context);
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
    const info = this.functionInfos.get(name);
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
      describeLiteralMismatch: this.describeLiteralMismatch.bind(this),
      isTypeAssignable: this.isTypeAssignable.bind(this),
      report: this.report.bind(this),
      handlePackageMemberAccess: this.handlePackageMemberAccess.bind(this),
      getIdentifierName: this.getIdentifierName.bind(this),
      checkBuiltinCallContext: this.checkBuiltinCallContext.bind(this),
      getBuiltinCallName: this.getBuiltinCallName.bind(this),
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
    applyImportsHelper(this.importContext(), module);
  }

  private importContext() {
    return {
      env: this.env,
      packageSummaries: this.packageSummaries,
      packageAliases: this.packageAliases,
      reportedPackageMemberAccess: this.reportedPackageMemberAccess,
      currentPackageName: this.currentPackageName,
      report: this.report.bind(this),
      getIdentifierName: this.getIdentifierName.bind(this),
    };
  }

  private registryContext() {
    return {
      structDefinitions: this.structDefinitions,
      interfaceDefinitions: this.interfaceDefinitions,
      typeAliases: this.typeAliases,
      implementationRecords: this.implementationRecords,
      implementationIndex: this.implementationIndex,
      declarationOrigins: this.declarationOrigins,
      declarationsContext: this.declarationsContext,
      getIdentifierName: (identifier: AST.Identifier | null | undefined) => this.getIdentifierName(identifier),
      report: this.report.bind(this),
    };
  }

  private applyImportStatement(imp: AST.ImportStatement | null | undefined): void {
    applyImportStatementHelper(this.importContext(), imp);
  }

  private applyDynImportStatement(statement: AST.DynImportStatement | null | undefined): void {
    if (statement?.isWildcard) {
      this.allowDynamicLookups = true;
      return;
    }
    applyDynImportStatementHelper(this.importContext(), statement);
  }

  private handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean {
    return handlePackageMemberAccessHelper(this.importContext(), expression);
  }

  private createTypeResolutionContext() {
    return {
      getTypeAlias: (name: string) => this.typeAliases.get(name),
      getInterfaceDefinition: (name: string) => this.interfaceDefinitions.get(name),
      hasInterfaceDefinition: (name: string) => this.interfaceDefinitions.has(name),
      getStructDefinition: (name: string) => this.structDefinitions.get(name),
      getIdentifierName: (identifier: AST.Identifier | null | undefined) => this.getIdentifierName(identifier),
    };
  }

  private createCheckerContext(): StatementContext {
    return createCheckerContextHelper({
      resolveStructDefinitionForPattern: this.resolveStructDefinitionForPattern.bind(this),
      getIdentifierName: this.getIdentifierName.bind(this),
      getIdentifierNameFromTypeExpression: this.getIdentifierNameFromTypeExpression.bind(this),
      getInterfaceNameFromConstraint: this.getInterfaceNameFromConstraint.bind(this),
      getInterfaceNameFromTypeExpression: this.getInterfaceNameFromTypeExpression.bind(this),
      report: this.report.bind(this),
      describeTypeExpression: this.describeTypeExpression.bind(this),
      isKnownTypeName: (name: string) => this.isKnownTypeName(name),
      hasTypeDefinition: (name: string) =>
        this.structDefinitions.has(name) || this.interfaceDefinitions.has(name) || this.typeAliases.has(name),
      typeInfosEquivalent: this.typeInfosEquivalent.bind(this),
      isTypeAssignable: this.isTypeAssignable.bind(this),
      describeLiteralMismatch: this.describeLiteralMismatch.bind(this),
      resolveTypeExpression: this.resolveTypeExpression.bind(this),
      getStructDefinition: (name: string) => this.structDefinitions.get(name),
      getInterfaceDefinition: (name: string) => this.interfaceDefinitions.get(name),
      hasInterfaceDefinition: (name: string) => this.interfaceDefinitions.has(name),
      handlePackageMemberAccess: this.handlePackageMemberAccess.bind(this),
      pushAsyncContext: this.pushAsyncContext.bind(this),
      popAsyncContext: this.popAsyncContext.bind(this),
      checkReturnStatement: this.checkReturnStatement.bind(this),
      checkFunctionCall: this.checkFunctionCall.bind(this),
      inferFunctionCallReturnType: this.inferFunctionCallReturnType.bind(this),
      checkFunctionDefinition: this.checkFunctionDefinition.bind(this),
      pushLoopContext: this.pushLoopContext.bind(this),
      popLoopContext: () => this.popLoopContext(),
      inLoopContext: () => this.inLoopContext(),
      pushScope: () => this.env.pushScope(),
      popScope: () => this.env.popScope(),
      withForkedEnv: <T>(fn: () => T) => this.withForkedEnv(fn),
      lookupIdentifier: (name: string) => this.env.lookup(name),
      defineValue: (name: string, valueType: TypeInfo) => this.env.define(name, valueType),
      assignValue: (name: string, valueType: TypeInfo) => this.env.assign(name, valueType),
      hasBinding: (name: string) => this.env.has(name),
      hasBindingInCurrentScope: (name: string) => this.env.hasInCurrentScope(name),
      allowDynamicLookup: () => this.allowDynamicLookups,
      getFunctionInfo: (key: string) => this.functionInfos.get(key),
      setFunctionInfo: (key: string, info: FunctionInfo) => this.functionInfos.set(key, info),
      isExpression: (node: AST.Node | undefined | null): node is AST.Expression => this.isExpression(node),
      handleTypeDeclaration: (node) => this.checkLocalTypeDeclaration(node as LocalTypeDeclaration),
      pushBreakpointLabel: (label: string) => this.pushBreakpointLabel(label),
      popBreakpointLabel: () => this.popBreakpointLabel(),
      hasBreakpointLabel: (label: string) => this.hasBreakpointLabel(label),
      handleBreakStatement: this.checkBreakStatement.bind(this),
      handleContinueStatement: this.checkContinueStatement.bind(this),
    });
  }

  private createImplementationContext(): ImplementationContext {
    const ctx = this.declarationsContext as ImplementationContext;
    ctx.formatImplementationTarget = this.formatImplementationTarget.bind(this);
    ctx.formatImplementationLabel = this.formatImplementationLabel.bind(this);
    ctx.registerMethodSet = (record) => {
      this.methodSets.push(record);
    };
    ctx.getMethodSets = () => this.methodSets;
    ctx.registerImplementationRecord = (record) => this.registerImplementationRecord(record);
    ctx.getImplementationRecords = () => this.implementationRecords;
    ctx.getImplementationBucket = (key: string) => this.implementationIndex.get(key);
    ctx.describeTypeArgument = this.describeTypeArgument.bind(this);
    ctx.appendInterfaceArgsToLabel = this.appendInterfaceArgsToLabel.bind(this);
    ctx.formatTypeExpression = this.formatTypeExpression.bind(this);
    return ctx;
  }

  private buildPackageSummary(module: AST.Module): PackageSummary | null {
    return buildPackageSummaryHelper(this.implementationContext, module);
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
