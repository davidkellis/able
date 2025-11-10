import type * as AST from "../ast";
import { buildStandardInterfaceBuiltins } from "../builtins/interfaces";
import { Environment } from "./environment";
import {
  describe,
  isBoolean,
  isNumeric,
  isUnknown,
  formatType,
  primitiveType,
  iteratorType,
  arrayType,
  rangeType,
  procType,
  futureType,
  unknownType,
  type TypeInfo,
} from "./types";
import {
  inferExpression as inferExpressionHelper,
  mergeBranchTypes as mergeBranchTypesHelper,
  type StatementContext,
} from "./checker/expressions";
import { checkStatement as checkStatementHelper } from "./checker/statements";
import {
  collectFunctionDefinition as collectFunctionDefinitionHelper,
  type DeclarationsContext,
} from "./checker/declarations";
import {
  collectImplementationDefinition as collectImplementationDefinitionHelper,
  collectMethodsDefinition as collectMethodsDefinitionHelper,
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
  type ImplementationContext,
} from "./checker/implementations";
import { buildPackageSummary as buildPackageSummaryHelper, resolvePackageName } from "./checker/summary";
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
  private functionInfos: Map<string, FunctionInfo> = new Map();
  private methodSets: MethodSetRecord[] = [];
  private implementationRecords: ImplementationRecord[] = [];
  private implementationIndex: Map<string, ImplementationRecord[]> = new Map();
  private packageAliases: Map<string, string> = new Map();
  private reportedPackageMemberAccess = new WeakSet<AST.MemberAccessExpression>();
  private asyncDepth = 0;
  private readonly context: StatementContext;
  private readonly declarationsContext: DeclarationsContext;
  private readonly implementationContext: ImplementationContext;

  constructor(options: TypeCheckerOptions = {}) {
    this.env = new Environment();
    this.options = options;
    this.packageSummaries = this.clonePackageSummaries(options.packageSummaries);
    this.context = this.createCheckerContext();
    this.declarationsContext = this.context as DeclarationsContext;
    this.implementationContext = this.createImplementationContext();
  }

  checkModule(module: AST.Module): TypecheckResult {
    this.env = new Environment();
    this.diagnostics = [];
    this.structDefinitions = new Map();
    this.interfaceDefinitions = new Map();
    this.functionInfos = new Map();
    this.methodSets = [];
    this.implementationRecords = [];
    this.implementationIndex = new Map();
    this.installBuiltins();
    this.packageAliases.clear();
    this.reportedPackageMemberAccess = new WeakSet();
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
    const voidType = primitiveType("void");
    const boolType = primitiveType("bool");
    const i32Type = primitiveType("i32");
    const i64Type = primitiveType("i64");
    const stringType = primitiveType("string");
    const charType = primitiveType("char");
    const unknown = unknownType;

    const register = (name: string, params: TypeInfo[], returnType: TypeInfo) => {
      this.registerBuiltinFunction(name, params, returnType);
    };

    register("print", [unknown], voidType);
    register("proc_yield", [], voidType);
    register("proc_cancelled", [], boolType);
    register("proc_flush", [], voidType);

    register("__able_channel_new", [i32Type], i64Type);
    register("__able_channel_send", [unknown, unknown], voidType);
    register("__able_channel_receive", [unknown], unknown);
    register("__able_channel_try_send", [unknown, unknown], boolType);
    register("__able_channel_try_receive", [unknown], unknown);
    register("__able_channel_close", [unknown], voidType);
    register("__able_channel_is_closed", [unknown], boolType);

    register("__able_mutex_new", [], i64Type);
    register("__able_mutex_lock", [i64Type], voidType);
    register("__able_mutex_unlock", [i64Type], voidType);

    register("__able_string_from_builtin", [stringType], arrayType(i32Type));
    register("__able_string_to_builtin", [arrayType(i32Type)], stringType);
    register("__able_char_from_codepoint", [i32Type], charType);

    register("__able_hasher_create", [], i64Type);
    register("__able_hasher_write", [i64Type, stringType], voidType);
    register("__able_hasher_finish", [i64Type], i64Type);
    this.installBuiltinInterfaces();
  }

  private registerBuiltinFunction(name: string, params: TypeInfo[], returnType: TypeInfo): void {
    const fnType: TypeInfo = {
      kind: "function",
      parameters: params,
      returnType,
    };
    this.env.define(name, fnType);
    this.functionInfos.set(name, {
      name,
      fullName: name,
      genericConstraints: [],
      genericParamNames: [],
      whereClause: [],
      returnType,
    });
  }

  private collectModuleDeclarations(module: AST.Module): void {
    if (!Array.isArray(module.body)) {
      return;
    }
    const statements = module.body as Array<AST.Statement | AST.Expression>;
    for (const statement of statements) {
      this.collectPrimaryDeclaration(statement);
    }
    for (const statement of statements) {
      this.collectImplementationDeclaration(statement);
    }
  }

  private collectPrimaryDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) return;
    switch (node.type) {
      case "InterfaceDefinition":
        this.registerInterfaceDefinition(node);
        break;
      case "StructDefinition":
        this.registerStructDefinition(node);
        break;
      case "MethodsDefinition":
        collectMethodsDefinitionHelper(this.implementationContext, node);
        break;
      case "FunctionDefinition":
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
    this.context.checkStatement(node);
  }

  private registerStructDefinition(definition: AST.StructDefinition): void {
    const name = definition.id?.name;
    if (name) {
      this.structDefinitions.set(name, definition);
    }
  }

  private registerInterfaceDefinition(definition: AST.InterfaceDefinition): void {
    const name = definition.id?.name;
    if (name) {
      this.interfaceDefinitions.set(name, definition);
    }
  }

  private registerImplementationRecord(record: ImplementationRecord): void {
    this.implementationRecords.push(record);
    const bucket = this.implementationIndex.get(record.targetKey);
    if (bucket) {
      bucket.push(record);
    } else {
      this.implementationIndex.set(record.targetKey, [record]);
    }
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
    const builtinName = this.getBuiltinCallName(call.callee);
    this.checkBuiltinCallContext(builtinName, call);
    const infos = this.resolveFunctionInfos(call.callee);
    if (!infos.length) {
      return;
    }
    for (const info of infos) {
      enforceFunctionConstraintsHelper(this.implementationContext, info, call);
    }
  }

  private inferFunctionCallReturnType(call: AST.FunctionCall): TypeInfo {
    const infos = this.resolveFunctionInfos(call.callee);
    if (!infos.length) {
      return unknownType;
    }
    const returnTypes = infos
      .map((info) => info.returnType ?? unknownType)
      .filter((type) => type && type.kind !== "unknown");
    if (!returnTypes.length) {
      return unknownType;
    }
    return mergeBranchTypesHelper(this.context, returnTypes);
  }

  private resolveFunctionInfos(callee: AST.Expression | undefined | null): FunctionInfo[] {
    if (!callee) return [];
    if (callee.type === "Identifier") {
      const info = this.functionInfos.get(callee.name);
      return info ? [info] : [];
    }
    if (callee.type === "MemberAccessExpression") {
      if (this.handlePackageMemberAccess(callee)) {
        return [];
      }
      const memberName = this.getIdentifierName(callee.member);
      if (!memberName) return [];
      let objectType = this.inferExpression(callee.object);
      if (
        objectType.kind !== "struct" &&
        callee.object?.type === "Identifier" &&
        callee.object.name &&
        this.structDefinitions.has(callee.object.name)
      ) {
        objectType = {
          kind: "struct",
          name: callee.object.name,
          typeArguments: [],
          definition: this.structDefinitions.get(callee.object.name),
        };
      }
      if (objectType.kind === "struct") {
        const structLabel = formatType(objectType);
        const memberKey = `${structLabel}::${memberName}`;
        const infos: FunctionInfo[] = [];
        const seen = new Set<string>();
        const info = this.functionInfos.get(memberKey);
        const genericMatches = lookupMethodSetsForCallHelper(
          this.implementationContext,
          structLabel,
          memberName,
          objectType,
        );
        if (genericMatches.length) {
          for (const match of genericMatches) {
            if (seen.has(match.fullName)) continue;
            infos.push(match);
            seen.add(match.fullName);
          }
        }
        if (!infos.length && info) {
          infos.push(info);
          seen.add(info.fullName);
        }
        if (!infos.length) {
          const fallback = this.functionInfos.get(memberName);
          if (fallback && !seen.has(fallback.fullName)) {
            infos.push(fallback);
          }
        }
        return infos;
      }
    }
    return [];
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

  private resolveTypeExpression(expr: AST.TypeExpression | null | undefined): TypeInfo {
    if (!expr) return unknownType;
    switch (expr.type) {
      case "SimpleTypeExpression": {
        const name = this.getIdentifierName(expr.name);
        if (!name) return unknownType;
        switch (name) {
          case "i32":
          case "f64":
          case "bool":
          case "string":
          case "char":
          case "nil":
          case "void":
            return primitiveType(name as any);
          default: {
            if (this.interfaceDefinitions.has(name)) {
              return { kind: "interface", name, typeArguments: [] };
            }
            return {
              kind: "struct",
              name,
              typeArguments: [],
              definition: this.structDefinitions.get(name),
            };
          }
        }
      }
      case "GenericTypeExpression": {
        const baseName = this.getIdentifierNameFromTypeExpression(expr.base);
        if (!baseName) return unknownType;
        const typeArguments = Array.isArray(expr.arguments)
          ? expr.arguments.map((arg) => this.resolveTypeExpression(arg))
          : [];
        if (this.interfaceDefinitions.has(baseName)) {
          return { kind: "interface", name: baseName, typeArguments };
        }
        return {
          kind: "struct",
          name: baseName,
          typeArguments,
          definition: this.structDefinitions.get(baseName),
        };
      }
      case "NullableTypeExpression":
        return {
          kind: "nullable",
          inner: this.resolveTypeExpression(expr.innerType),
        };
      case "ResultTypeExpression":
        return {
          kind: "result",
          inner: this.resolveTypeExpression(expr.innerType),
        };
      case "UnionTypeExpression": {
        const members = Array.isArray(expr.members)
          ? expr.members.map((member) => this.resolveTypeExpression(member))
          : [];
        return { kind: "union", members };
      }
      case "FunctionTypeExpression": {
        const parameters = Array.isArray(expr.paramTypes)
          ? expr.paramTypes.map((param) => this.resolveTypeExpression(param))
          : [];
        const returnType = this.resolveTypeExpression(expr.returnType);
        return {
          kind: "function",
          parameters,
          returnType,
        };
      }
      default:
        return unknownType;
    }
  }

  private typeInfosEquivalent(a: TypeInfo | undefined, b: TypeInfo | undefined): boolean {
    if (!a || a.kind === "unknown" || !b || b.kind === "unknown") {
      return true;
    }
    return formatType(a) === formatType(b);
  }

  private typeExpressionsEquivalent(
    a: AST.TypeExpression | null | undefined,
    b: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): boolean {
    if (!a && !b) return true;
    if (!a || !b) return false;
    return this.formatTypeExpression(a, substitutions) === this.formatTypeExpression(b, substitutions);
  }

  private describeTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, string>,
  ): string {
    if (!expr) return "unspecified";
    return this.formatTypeExpression(expr, substitutions);
  }

  private formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string {
    switch (expr.type) {
      case "SimpleTypeExpression":
        return this.lookupSubstitution(this.getIdentifierName(expr.name), substitutions);
      case "GenericTypeExpression": {
        const base = expr.base ? this.formatTypeExpression(expr.base, substitutions) : "Unknown";
        const args = Array.isArray(expr.arguments)
          ? expr.arguments
              .map((arg) => (arg ? this.formatTypeExpression(arg, substitutions) : "Unknown"))
              .filter(Boolean)
          : [];
        return args.length > 0 ? [base, ...args].join(" ") : base;
      }
      case "FunctionTypeExpression": {
        const params = Array.isArray(expr.paramTypes)
          ? expr.paramTypes.map((param) => (param ? this.formatTypeExpression(param, substitutions) : "Unknown"))
          : [];
        const ret = expr.returnType ? this.formatTypeExpression(expr.returnType, substitutions) : "void";
        return `fn(${params.join(", ")}) -> ${ret}`;
      }
      case "NullableTypeExpression":
        return `${this.formatTypeExpression(expr.innerType, substitutions)}?`;
      case "ResultTypeExpression":
        return `Result ${this.formatTypeExpression(expr.innerType, substitutions)}`;
      case "UnionTypeExpression": {
        const members = Array.isArray(expr.members)
          ? expr.members.map((member) => (member ? this.formatTypeExpression(member, substitutions) : "Unknown"))
          : [];
        return members.length > 0 ? members.join(" | ") : "Union";
      }
      case "WildcardTypeExpression":
        return "_";
      default:
        return "Unknown";
    }
  }

  private lookupSubstitution(name: string | null, substitutions?: Map<string, string>): string {
    if (!name) return "Unknown";
    if (substitutions && substitutions.has(name)) {
      return substitutions.get(name) ?? "Unknown";
    }
    return name;
  }

  private describeTypeArgument(type: TypeInfo): string {
    return formatType(type);
  }

  private appendInterfaceArgsToLabel(label: string, args: string[]): string {
    if (!args.length) {
      return label;
    }
    return `${label} ${args.join(" ")}`.trim();
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
      case "ImportStatement":
      case "DynImportStatement":
      case "MethodsDefinition":
      case "PreludeStatement":
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
    if (!module || !Array.isArray(module.imports) || module.imports.length === 0) {
      return;
    }
    const currentPackage = resolvePackageName(module);
    for (const imp of module.imports) {
      if (!imp) continue;
      const packageName = this.formatImportPath(imp.packagePath);
      const summary = packageName ? this.packageSummaries.get(packageName) : undefined;
      if (!summary) {
        const label = packageName ?? "<unknown>";
        this.report(`typechecker: import references unknown package '${label}'`, imp);
        continue;
      }
      if (summary.visibility === "private" && summary.name !== currentPackage) {
        this.report(`typechecker: package '${summary.name}' is private`, imp);
        continue;
      }
      if (imp.isWildcard) {
        if (summary.symbols) {
          for (const symbolName of Object.keys(summary.symbols)) {
            if (!this.env.has(symbolName)) {
              this.env.define(symbolName, unknownType);
            }
          }
        }
        continue;
      }
      if (Array.isArray(imp.selectors) && imp.selectors.length > 0) {
        for (const selector of imp.selectors) {
          if (!selector) continue;
          const selectorName = this.getIdentifierName(selector.name);
          if (!selectorName) continue;
          const aliasName = this.getIdentifierName(selector.alias) ?? selectorName;
          if (!summary.symbols || !summary.symbols[selectorName]) {
            const label = packageName ?? "<unknown>";
            this.report(`typechecker: package '${label}' has no symbol '${selectorName}'`, selector);
          }
          if (!this.env.has(aliasName)) {
            this.env.define(aliasName, unknownType);
          }
        }
        continue;
      }
      const aliasName = this.getIdentifierName(imp.alias) ?? this.defaultPackageAlias(packageName);
      if (!aliasName) {
        continue;
      }
      if (packageName) {
        this.packageAliases.set(aliasName, packageName);
      }
      if (!this.env.has(aliasName)) {
        this.env.define(aliasName, unknownType);
      }
    }
  }

  private formatImportPath(path: AST.Identifier[] | null | undefined): string | null {
    if (!Array.isArray(path) || path.length === 0) {
      return null;
    }
    const segments = path
      .map((segment) => this.getIdentifierName(segment))
      .filter((name): name is string => Boolean(name));
    if (segments.length === 0) {
      return null;
    }
    return segments.join(".");
  }

  private defaultPackageAlias(packageName: string | null): string | null {
    if (!packageName) {
      return null;
    }
    const segments = packageName.split(".");
    if (segments.length === 0) {
      return null;
    }
    return segments[segments.length - 1] ?? null;
  }

  private handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean {
    if (!expression.object || expression.object.type !== "Identifier") {
      return false;
    }
    const aliasName = expression.object.name;
    if (!aliasName) {
      return false;
    }
    const packageName = this.packageAliases.get(aliasName);
    if (!packageName) {
      return false;
    }
    const summary = this.packageSummaries.get(packageName);
    const memberName = this.getIdentifierName(expression.member as AST.Identifier);
    if (!memberName) {
      return true;
    }
    if (!summary?.symbols || !summary.symbols[memberName]) {
      if (!this.reportedPackageMemberAccess.has(expression)) {
        this.report(`typechecker: package '${packageName}' has no symbol '${memberName}'`, expression.member ?? expression);
        this.reportedPackageMemberAccess.add(expression);
      }
    }
    return true;
  }

  private clonePackageSummaries(
    summaries?: Map<string, PackageSummary> | Record<string, PackageSummary>,
  ): Map<string, PackageSummary> {
    if (!summaries) {
      return new Map();
    }
    if (summaries instanceof Map) {
      return new Map(summaries);
    }
    return new Map(Object.entries(summaries));
  }

  private createCheckerContext(): StatementContext {
    const ctx: Partial<StatementContext> = {};
    ctx.resolveStructDefinitionForPattern = this.resolveStructDefinitionForPattern.bind(this);
    ctx.getIdentifierName = this.getIdentifierName.bind(this);
    ctx.getIdentifierNameFromTypeExpression = this.getIdentifierNameFromTypeExpression.bind(this);
    ctx.getInterfaceNameFromConstraint = this.getInterfaceNameFromConstraint.bind(this);
    ctx.getInterfaceNameFromTypeExpression = this.getInterfaceNameFromTypeExpression.bind(this);
    ctx.report = this.report.bind(this);
    ctx.describeTypeExpression = this.describeTypeExpression.bind(this);
    ctx.typeInfosEquivalent = this.typeInfosEquivalent.bind(this);
    ctx.resolveTypeExpression = this.resolveTypeExpression.bind(this);
    ctx.getStructDefinition = (name: string) => this.structDefinitions.get(name);
    ctx.getInterfaceDefinition = (name: string) => this.interfaceDefinitions.get(name);
    ctx.hasInterfaceDefinition = (name: string) => this.interfaceDefinitions.has(name);
    ctx.handlePackageMemberAccess = this.handlePackageMemberAccess.bind(this);
    ctx.pushAsyncContext = this.pushAsyncContext.bind(this);
    ctx.popAsyncContext = this.popAsyncContext.bind(this);
    ctx.checkFunctionCall = this.checkFunctionCall.bind(this);
    ctx.inferFunctionCallReturnType = this.inferFunctionCallReturnType.bind(this);
    ctx.pushScope = () => this.env.pushScope();
    ctx.popScope = () => this.env.popScope();
    ctx.withForkedEnv = <T>(fn: () => T) => this.withForkedEnv(fn);
    ctx.lookupIdentifier = (name: string) => this.env.lookup(name);
    ctx.defineValue = (name: string, valueType: TypeInfo) => this.env.define(name, valueType);
    ctx.getFunctionInfo = (key: string) => this.functionInfos.get(key);
    ctx.setFunctionInfo = (key: string, info: FunctionInfo) => this.functionInfos.set(key, info);
    ctx.isExpression = (node: AST.Node | undefined | null): node is AST.Expression => this.isExpression(node);

    const expressionCtx = ctx as StatementContext;
    ctx.inferExpression = (expression) => inferExpressionHelper(expressionCtx, expression);
    ctx.checkStatement = (node) => checkStatementHelper(expressionCtx, node);
    return expressionCtx;
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

  private installBuiltinInterfaces(): void {
    const { interfaces, implementations } = buildStandardInterfaceBuiltins();
    for (const iface of interfaces) {
      this.registerInterfaceDefinition(iface);
    }
    for (const impl of implementations) {
      collectImplementationDefinitionHelper(this.implementationContext, impl);
    }
  }

}

export function createTypeChecker(options?: TypeCheckerOptions): TypeChecker {
  return new TypeChecker(options);
}
