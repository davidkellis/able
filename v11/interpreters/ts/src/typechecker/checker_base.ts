import type * as AST from "../ast";
import { Environment } from "./environment";
import {
  describe,
  formatType,
  futureType,
  isBoolean,
  isNumeric,
  isUnknown,
  primitiveType,
  procType,
  rangeType,
  unknownType,
  type IntegerPrimitive,
  type TypeInfo,
} from "./types";
import {
  cloneFunctionInfoMap,
  clonePrelude,
  LocalTypeDeclaration,
  TypeCheckerOptions,
  TypeCheckerPrelude,
} from "./checker/core";
import type { StatementContext } from "./checker/expression-context";
import {
  buildCheckerContext,
  buildImplementationContext,
  buildImportContext,
  buildRegistryContext,
  buildTypeResolutionContext,
} from "./checker/context-builders";
import { collectFunctionDefinition as collectFunctionDefinitionHelper, type DeclarationsContext } from "./checker/declarations";
import {
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
  typeImplementsInterface,
  type ImplementationContext,
} from "./checker/implementations";
import { buildPackageSummary as buildPackageSummaryHelper } from "./checker/summary";
import {
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
import type { DiagnosticSeverity, TypecheckerDiagnostic, PackageSummary } from "./diagnostics";
import { createTypeResolutionHelpers, type TypeResolutionHelpers } from "./checker/type-resolution";
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
import { refineTypeWithExpected } from "./checker/expressions";
import { checkReturnStatement as checkReturnStatementHelper } from "./checker/statements";
import { getFunctionDefinitionName } from "./checker/names";
import {
  checkBreakStatement as checkBreakStatementHelper,
  checkContinueStatement as checkContinueStatementHelper,
} from "./checker/loops";
import { buildTypeCallTargetLabel } from "./checker/type-formatting";
import { normalizeUnionMembers } from "./checker/union-normalization";

type FunctionGenericContext = {
  label: string;
  inferred: Map<string, AST.GenericParameter>;
};

export class TypeCheckerBase {
  protected env: Environment;
  protected readonly options: TypeCheckerOptions;
  protected readonly packageSummaries: Map<string, PackageSummary>;
  protected readonly prelude?: TypeCheckerPrelude;
  protected diagnostics: TypecheckerDiagnostic[] = [];
  protected structDefinitions: Map<string, AST.StructDefinition> = new Map();
  protected interfaceDefinitions: Map<string, AST.InterfaceDefinition> = new Map();
  protected typeAliases: Map<string, AST.TypeAliasDefinition> = new Map();
  protected unionDefinitions: Map<string, AST.UnionDefinition> = new Map();
  protected functionInfos: Map<string, FunctionInfo[]> = new Map();
  protected methodSets: MethodSetRecord[] = [];
  protected implementationRecords: ImplementationRecord[] = [];
  protected implementationIndex: Map<string, ImplementationRecord[]> = new Map();
  protected declarationOrigins: Map<string, AST.Node> = new Map();
  protected functionGenericStack: FunctionGenericContext[] = [];
  protected packageAliases: Map<string, string> = new Map();
  protected reportedPackageMemberAccess = new WeakSet<AST.MemberAccessExpression>();
  protected asyncDepth = 0;
  protected returnTypeStack: TypeInfo[] = [];
  protected loopResultStack: TypeInfo[] = [];
  protected breakpointStack: string[] = [];
  protected allowDynamicLookups = false;
  protected currentPackageName = "<anonymous>";
  protected readonly typeResolution: TypeResolutionHelpers;
  protected readonly context: StatementContext;
  protected readonly declarationsContext: DeclarationsContext;
  protected readonly implementationContext: ImplementationContext;

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
  protected formatNodeOrigin(node: AST.Node | null | undefined): string {
    return formatNodeOriginHelper(node);
  }

  protected registerStructDefinition(definition: AST.StructDefinition): void {
    registerStructDefinitionHelper(this.registryContext(), definition);
  }

  protected registerInterfaceDefinition(definition: AST.InterfaceDefinition): void {
    registerInterfaceDefinitionHelper(this.registryContext(), definition);
  }

  protected registerTypeAlias(definition: AST.TypeAliasDefinition): void {
    registerTypeAliasHelper(this.registryContext(), definition);
    const target = definition.targetType;
    if (target?.type === "UnionTypeExpression" && Array.isArray(target.members)) {
      const members = target.members.map((member) => ({
        type: this.typeResolution.resolveTypeExpression(member),
        node: member ?? definition,
      }));
      this.normalizeUnionMembers(members, true);
    }
  }

  protected registerUnionDefinition(definition: AST.UnionDefinition): void {
    const name = definition.id?.name;
    if (!name) {
      return;
    }
    if (!this.ensureUniqueDeclaration(name, definition)) {
      return;
    }
    this.unionDefinitions.set(name, definition);
    if (Array.isArray(definition.variants) && definition.variants.length > 0) {
      const members = definition.variants.map((variant) => ({
        type: this.typeResolution.resolveTypeExpression(variant),
        node: variant ?? definition,
      }));
      this.normalizeUnionMembers(members, true);
    }
  }

  protected registerImplementationRecord(record: ImplementationRecord): void {
    registerImplementationRecordHelper(this.registryContext(), record);
  }

  protected collectFunctionDefinition(
    definition: AST.FunctionDefinition,
    context: FunctionContext | undefined,
  ): void {
    collectFunctionDefinitionHelper(this.declarationsContext, definition, context);
  }

  protected collectExternFunctionBody(extern: AST.ExternFunctionBody): void {
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

  protected formatImplementationLabel(interfaceName: string, targetName: string): string {
    return `impl ${interfaceName} for ${targetName}`;
  }

  protected formatImplementationTarget(targetType: AST.TypeExpression | null | undefined): string | null {
    if (!targetType) return null;
    return this.formatTypeExpression(targetType);
  }

  protected inferExpression(expression: AST.Expression | undefined | null): TypeInfo {
    return this.context.inferExpression(expression);
  }

  protected withForkedEnv<T>(fn: () => T): T {
    const previousEnv = this.env;
    this.env = this.env.fork();
    try {
      return fn();
    } finally {
      this.env = previousEnv;
    }
  }

  protected pushAsyncContext(): void {
    this.asyncDepth += 1;
  }

  protected popAsyncContext(): void {
    if (this.asyncDepth > 0) {
      this.asyncDepth -= 1;
    }
  }

  protected inAsyncContext(): boolean {
    return this.asyncDepth > 0;
  }

  protected pushReturnType(type: TypeInfo): void {
    this.returnTypeStack.push(type ?? unknownType);
  }

  protected popReturnType(): void {
    if (this.returnTypeStack.length > 0) {
      this.returnTypeStack.pop();
    }
  }

  protected currentReturnType(): TypeInfo | undefined {
    if (!this.returnTypeStack.length) return undefined;
    return this.returnTypeStack[this.returnTypeStack.length - 1];
  }

  protected getBuiltinCallName(callee: AST.Expression | undefined | null): string | undefined {
    if (!callee) return undefined;
    if (callee.type === "Identifier") {
      return callee.name;
    }
    return undefined;
  }

  protected checkBuiltinCallContext(name: string | undefined, call: AST.FunctionCall): void {
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

  protected checkFunctionCall(call: AST.FunctionCall): void {
    checkFunctionCallHelper(this.functionCallContext(), call);
  }

  protected checkFunctionDefinition(definition: AST.FunctionDefinition): void {
    if (!definition) return;
    const name = definition.id?.name ?? "<anonymous>";
    const paramTypes = Array.isArray(definition.params)
      ? definition.params.map((param) => this.resolveTypeExpression(param?.paramType))
      : [];
    const expectedReturn = this.resolveTypeExpression(definition.returnType);
    if (definition.id?.name) {
      this.env.define(definition.id.name, {
        kind: "function",
        parameters: paramTypes,
        returnType: expectedReturn ?? unknownType,
      });
    }
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
      const expectedVoid =
        expectedReturn?.kind === "primitive" && expectedReturn.name === "void";
      if (!expectedVoid && expectedReturn && expectedReturn.kind !== "unknown" && bodyType && bodyType.kind !== "unknown") {
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

  protected checkReturnStatement(statement: AST.ReturnStatement): void {
    if (!statement) return;
    const expected = this.currentReturnType();
    let actual = statement.argument ? this.inferExpression(statement.argument) : primitiveType("void");
    if (statement.argument?.type === "FunctionCall" && expected && expected.kind !== "unknown") {
      actual = refineTypeWithExpected(actual, expected);
    }
    if (!expected) {
      this.report("typechecker: return statement outside function", statement);
      return;
    }
    if (expected.kind === "unknown") {
      return;
    }
    if (expected.kind === "primitive" && expected.name === "void") {
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

  protected pushLoopContext(): void {
    this.loopResultStack.push(unknownType);
  }

  protected popLoopContext(): TypeInfo {
    if (this.loopResultStack.length === 0) {
      return unknownType;
    }
    return this.loopResultStack.pop() ?? unknownType;
  }

  protected inLoopContext(): boolean {
    return this.loopResultStack.length > 0;
  }

  protected recordLoopBreakType(breakType: TypeInfo): void {
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

  protected pushBreakpointLabel(label: string | null | undefined): void {
    if (!label) {
      return;
    }
    this.breakpointStack.push(label);
  }

  protected popBreakpointLabel(): void {
    if (this.breakpointStack.length === 0) {
      return;
    }
    this.breakpointStack.pop();
  }

  protected hasBreakpointLabel(label: string | null | undefined): boolean {
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

  protected checkBreakStatement(statement: AST.BreakStatement): void {
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

  protected checkContinueStatement(statement: AST.ContinueStatement): void {
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

  protected inferFunctionCallReturnType(call: AST.FunctionCall): TypeInfo {
    return inferFunctionCallReturnTypeHelper(this.functionCallContext(), call);
  }

  protected functionCallContext() {
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

  protected resolveStructDefinitionForPattern(pattern: AST.StructPattern, valueType: TypeInfo): AST.StructDefinition | undefined {
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

  protected resolveTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    if (expr?.type === "UnionTypeExpression") {
      const members = Array.isArray(expr.members)
        ? expr.members.map((member) => ({
            type: this.typeResolution.resolveTypeExpression(member, substitutions),
            node: member ?? expr,
          }))
        : [];
      return this.normalizeUnionMembers(members, true);
    }
    const resolved = this.typeResolution.resolveTypeExpression(expr, substitutions);
    if (resolved.kind === "union") {
      return this.normalizeUnionMembers(resolved.members.map((member) => ({ type: member })), false);
    }
    return resolved;
  }

  protected normalizeUnionType(members: TypeInfo[]): TypeInfo {
    const entries = members.map((member) => ({ type: member }));
    return this.normalizeUnionMembers(entries, false);
  }

  protected normalizeUnionMembers(
    members: Array<{ type: TypeInfo; node?: AST.Node | null }>,
    warnRedundant: boolean,
  ): TypeInfo {
    const normalized: Array<{ type: TypeInfo; node?: AST.Node | null }> = [];
    const nilType = primitiveType("nil");
    const equivalentForUnion = (left: TypeInfo, right: TypeInfo): boolean => {
      const leftUnknown = left.kind === "unknown";
      const rightUnknown = right.kind === "unknown";
      if (leftUnknown || rightUnknown) {
        return leftUnknown && rightUnknown;
      }
      return this.typeInfosEquivalent(left, right);
    };
    const addMember = (entry: { type: TypeInfo; node?: AST.Node | null }): void => {
      if (!entry.type) {
        return;
      }
      if (entry.type.kind === "union") {
        const innerMembers = entry.type.members ?? [];
        for (const inner of innerMembers) {
          addMember({ type: inner, node: entry.node });
        }
        return;
      }
      if (entry.type.kind === "nullable") {
        addMember({ type: nilType, node: entry.node });
        addMember({ type: entry.type.inner ?? unknownType, node: entry.node });
        return;
      }
      const exists = normalized.some((existing) => equivalentForUnion(existing.type, entry.type));
      if (exists) {
        if (warnRedundant && entry.node) {
          this.reportWarning(`typechecker: redundant union member ${formatType(entry.type)}`, entry.node);
        }
        return;
      }
      normalized.push(entry);
    };
    for (const entry of members) {
      addMember(entry);
    }
    if (normalized.length === 0) {
      return unknownType;
    }
    if (normalized.length === 1) {
      return normalized[0]!.type;
    }
    if (normalized.length === 2) {
      const nilIndex = normalized.findIndex(
        (entry) => entry.type.kind === "primitive" && entry.type.name === "nil",
      );
      if (nilIndex !== -1) {
        const other = normalized[nilIndex === 0 ? 1 : 0]!.type;
        return { kind: "nullable", inner: other };
      }
    }
    return { kind: "union", members: normalized.map((entry) => entry.type) };
  }

  protected instantiateTypeAlias(
    definition: AST.TypeAliasDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    return this.typeResolution.instantiateTypeAlias(definition, typeArguments, outerSubstitutions);
  }

  protected typeInfosEquivalent(a: TypeInfo | undefined, b: TypeInfo | undefined): boolean {
    return this.typeResolution.typeInfosEquivalent(a, b);
  }

  protected isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean {
    return this.typeResolution.isTypeAssignable(actual, expected);
  }

  public describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null {
    return this.typeResolution.describeLiteralMismatch(actual, expected);
  }

  protected canonicalizeStructuralType(type: TypeInfo): TypeInfo {
    return this.typeResolution.canonicalizeStructuralType(type);
  }

  protected typeExpressionsEquivalent(
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): boolean {
    return this.typeResolution.typeExpressionsEquivalent(a, b, substitutions);
  }

  protected describeTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): string {
    return this.typeResolution.describeTypeExpression(expr, substitutions);
  }

  protected formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string {
    return this.typeResolution.formatTypeExpression(expr, substitutions);
  }

  protected lookupSubstitution(name: string | null, substitutions?: Map<string, string>): string {
    return this.typeResolution.lookupSubstitution(name, substitutions);
  }

  protected describeTypeArgument(type: TypeInfo): string {
    return this.typeResolution.describeTypeArgument(type);
  }

  protected appendInterfaceArgsToLabel(label: string, args: string[]): string {
    return this.typeResolution.appendInterfaceArgsToLabel(label, args);
  }

  protected getInterfaceNameFromConstraint(constraint: AST.InterfaceConstraint | null | undefined): string | null {
    if (!constraint) return null;
    return this.getInterfaceNameFromTypeExpression(constraint.interfaceType);
  }

  protected getInterfaceNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null {
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

  protected getIdentifierName(node: AST.Identifier | null | undefined): string | null {
    if (!node) return null;
    return node.name ?? null;
  }

  protected getIdentifierNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null {
    if (!expr) return null;
    if (expr.type === "SimpleTypeExpression") {
      return this.getIdentifierName(expr.name);
    }
    if (expr.type === "GenericTypeExpression") {
      return this.getIdentifierNameFromTypeExpression(expr.base);
    }
    return null;
  }

  protected isExpression(node: AST.Statement | AST.Expression): node is AST.Expression {
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

  protected report(message: string, node?: AST.Node | null): void {
    this.pushDiagnostic("error", message, node);
  }

  protected reportWarning(message: string, node?: AST.Node | null): void {
    this.pushDiagnostic("warning", message, node);
  }

  protected pushDiagnostic(severity: DiagnosticSeverity, message: string, node?: AST.Node | null): void {
    const location = extractLocation(node);
    const diagnostic: TypecheckerDiagnostic = { severity, message };
    if (location) {
      diagnostic.location = location;
    }
    this.diagnostics.push(diagnostic);
  }

  protected applyImports(module: AST.Module): void {
    applyImportsHelper(buildImportContext(this), module);
  }

  protected registryContext() {
    return buildRegistryContext(this);
  }

  protected applyImportStatement(imp: AST.ImportStatement | null | undefined): void {
    applyImportStatementHelper(buildImportContext(this), imp);
  }

  protected applyDynImportStatement(statement: AST.DynImportStatement | null | undefined): void {
    if (statement?.isWildcard) {
      this.allowDynamicLookups = true;
      return;
    }
    applyDynImportStatementHelper(buildImportContext(this), statement);
  }

  protected handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean {
    return handlePackageMemberAccessHelper(buildImportContext(this), expression);
  }

  protected buildPackageSummary(module: AST.Module): PackageSummary | null {
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

  protected addFunctionInfo(key: string, info: FunctionInfo): void {
    const existing = this.functionInfos.get(key) ?? [];
    if (existing.some((entry) => entry.fullName === info.fullName)) {
      return;
    }
    this.functionInfos.set(key, [...existing, info]);
  }

  protected getFunctionInfos(key: string): FunctionInfo[] {
    return this.functionInfos.get(key) ?? [];
  }

  protected pushFunctionGenericContext(definition: AST.FunctionDefinition): void {
    if (!definition) {
      return;
    }
    const inferred = this.collectInferredGenericParameters(definition);
    const label = definition.id?.name ? `fn ${definition.id.name}` : "fn <anonymous>";
    this.functionGenericStack.push({ label, inferred });
  }

  protected popFunctionGenericContext(): void {
    if (this.functionGenericStack.length === 0) {
      return;
    }
    this.functionGenericStack.pop();
  }

  protected collectInferredGenericParameters(definition: AST.FunctionDefinition): Map<string, AST.GenericParameter> {
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

  protected checkLocalTypeDeclaration(node: LocalTypeDeclaration): void {
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
