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
  type LiteralInfo,
  type PrimitiveName,
  type IntegerPrimitive,
} from "./types";
import {
  inferExpression as inferExpressionHelper,
  mergeBranchTypes as mergeBranchTypesHelper,
  type StatementContext,
} from "./checker/expressions";
import { checkStatement as checkStatementHelper } from "./checker/statements";
import {
  collectFunctionDefinition as collectFunctionDefinitionHelper,
  inferFunctionSignatureGenerics,
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
import { hasIntegerBounds, integerBounds, getIntegerTypeInfo } from "./numeric";

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
    const voidType = primitiveType("void");
    const boolType = primitiveType("bool");
    const i32Type = primitiveType("i32");
    const i64Type = primitiveType("i64");
    const u64Type = primitiveType("u64");
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
    register("proc_pending_tasks", [], i32Type);

    register("__able_channel_new", [i32Type], i64Type);
    register("__able_channel_send", [unknown, unknown], voidType);
    register("__able_channel_receive", [unknown], unknown);
    register("__able_channel_try_send", [unknown, unknown], boolType);
    register("__able_channel_try_receive", [unknown], unknown);
    register("__able_channel_await_try_recv", [unknown, unknown], unknown);
    register("__able_channel_await_try_send", [unknown, unknown, unknown], unknown);
    register("__able_channel_close", [unknown], voidType);
    register("__able_channel_is_closed", [unknown], boolType);

    register("__able_mutex_new", [], i64Type);
    register("__able_mutex_lock", [i64Type], voidType);
    register("__able_mutex_unlock", [i64Type], voidType);

    register("__able_array_new", [], i64Type);
    register("__able_array_with_capacity", [i32Type], i64Type);
    register("__able_array_size", [i64Type], u64Type);
    register("__able_array_capacity", [i64Type], u64Type);
    register("__able_array_set_len", [i64Type, i32Type], voidType);
    register("__able_array_read", [i64Type, i32Type], unknown);
    register("__able_array_write", [i64Type, i32Type, unknown], voidType);
    register("__able_array_reserve", [i64Type, i32Type], u64Type);
    register("__able_array_clone", [i64Type], i64Type);

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
      parameters: params,
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
    if (!name || !node) {
      return true;
    }
    const existing = this.declarationOrigins.get(name);
    if (existing) {
      const location = this.formatNodeOrigin(existing);
      this.report(
        `typechecker: duplicate declaration '${name}' (previous declaration at ${location})`,
        node,
      );
      return false;
    }
    this.declarationOrigins.set(name, node);
    return true;
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
    if (!node) {
      return "<unknown location>";
    }
    const origin = (node as { origin?: string }).origin ?? "<unknown file>";
    const span = (node as { span?: { start?: { line?: number; column?: number } } }).span;
    const line = span?.start?.line ?? 0;
    const column = span?.start?.column ?? 0;
    return `${origin}:${line}:${column}`;
  }

  private registerStructDefinition(definition: AST.StructDefinition): void {
    const name = definition.id?.name;
    if (name) {
      if (!this.ensureUniqueDeclaration(name, definition)) {
        return;
      }
      this.structDefinitions.set(name, definition);
    }
  }

  private registerInterfaceDefinition(definition: AST.InterfaceDefinition): void {
    const name = definition.id?.name;
    if (name) {
      if (!this.ensureUniqueDeclaration(name, definition)) {
        return;
      }
      const parentGenerics = Array.isArray(definition.genericParams)
        ? definition.genericParams
            .map((param) => this.getIdentifierName(param?.name))
            .filter((paramName): paramName is string => Boolean(paramName))
        : [];
      if (Array.isArray(definition.signatures)) {
        for (const signature of definition.signatures) {
          inferFunctionSignatureGenerics(this.declarationsContext, signature, parentGenerics);
        }
      }
      this.interfaceDefinitions.set(name, definition);
    }
  }

  private registerTypeAlias(definition: AST.TypeAliasDefinition): void {
    const name = definition.id?.name;
    if (!name) return;
    if (!this.ensureUniqueDeclaration(name, definition)) {
      return;
    }
    this.typeAliases.set(name, definition);
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
    const builtinName = this.getBuiltinCallName(call.callee);
    const args = Array.isArray(call.arguments) ? call.arguments : [];
    const argTypes = args.map((arg) => this.inferExpression(arg));
    this.checkBuiltinCallContext(builtinName, call);
    const infos = this.resolveFunctionInfos(call.callee);
    if (!infos.length) {
      return;
    }
    const info = infos[0];
    if (info) {
      const rawParams = Array.isArray(info.parameters) ? info.parameters : [];
      const implicitSelf =
        Boolean(info.structName && info.hasImplicitSelf) &&
        call.callee?.type === "MemberAccessExpression" &&
        rawParams.length > 0;
      let params = implicitSelf ? rawParams.slice(1) : rawParams;
      const optionalLast =
        params.length > 0 && params[params.length - 1]?.kind === "nullable";
      if (params.length !== args.length && !(optionalLast && args.length === params.length - 1)) {
        this.report(
          `typechecker: function expects ${params.length} arguments, got ${args.length}`,
          call,
        );
      }
      if (optionalLast && args.length === params.length - 1) {
        params = params.slice(0, params.length - 1);
      }
      const compareCount = Math.min(params.length, argTypes.length);
      for (let index = 0; index < compareCount; index += 1) {
        const expected = params[index];
        const actual = argTypes[index];
        if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
          continue;
        }
        const literalMessage = this.describeLiteralMismatch(actual, expected);
        if (literalMessage) {
          this.report(literalMessage, args[index] ?? call);
          continue;
        }
        if (!this.isTypeAssignable(actual, expected)) {
          this.report(
            `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
            args[index] ?? call,
          );
        }
      }
    }
    for (const info of infos) {
      enforceFunctionConstraintsHelper(this.implementationContext, info, call);
    }
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
      if (objectType.kind === "primitive" && objectType.name === "string") {
        const info = this.buildStringMethodInfo(memberName);
        if (info) return [info];
      }
    }
    return [];
  }

  private buildStringMethodInfo(memberName: string): FunctionInfo | null {
    const stringType = primitiveType("string");
    const u64 = primitiveType("u64");
    const boolType = primitiveType("bool");
    const base: Omit<FunctionInfo, "returnType" | "parameters"> = {
      name: `string.${memberName}`,
      fullName: `string.${memberName}`,
      genericConstraints: [],
      genericParamNames: [],
      whereClause: [],
    };
    switch (memberName) {
      case "len_bytes":
      case "len_chars":
      case "len_graphemes":
        return { ...base, parameters: [], returnType: u64 };
      case "bytes":
        return { ...base, parameters: [], returnType: iteratorType(primitiveType("u8")) };
      case "chars":
        return { ...base, parameters: [], returnType: iteratorType(primitiveType("char")) };
      case "graphemes":
        return { ...base, parameters: [], returnType: iteratorType(stringType) };
      case "substring":
        return {
          ...base,
          parameters: [u64, { kind: "nullable", inner: u64 }],
          returnType: { kind: "result", inner: stringType },
        };
      case "split":
        return { ...base, parameters: [stringType], returnType: arrayType(stringType) };
      case "replace":
        return { ...base, parameters: [stringType, stringType], returnType: stringType };
      case "starts_with":
      case "ends_with":
        return { ...base, parameters: [stringType], returnType: boolType };
      default:
        return null;
    }
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
    if (!expr) return unknownType;
    switch (expr.type) {
      case "SimpleTypeExpression": {
        const name = this.getIdentifierName(expr.name);
        if (!name) return unknownType;
        if (substitutions?.has(name)) {
          return substitutions.get(name) ?? unknownType;
        }
        switch (name) {
          case "i8":
          case "i16":
          case "i32":
          case "i64":
          case "i128":
          case "u8":
          case "u16":
          case "u32":
          case "u64":
          case "u128":
          case "f32":
          case "f64":
          case "bool":
          case "string":
          case "char":
          case "nil":
          case "void":
            return primitiveType(name as PrimitiveName);
          default: {
            const alias = this.typeAliases.get(name);
            if (alias) {
              return this.instantiateTypeAlias(alias, [], substitutions);
            }
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
          ? expr.arguments.map((arg) => this.resolveTypeExpression(arg, substitutions))
          : [];
        const alias = this.typeAliases.get(baseName);
        if (alias) {
          return this.instantiateTypeAlias(alias, typeArguments, substitutions);
        }
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
          inner: this.resolveTypeExpression(expr.innerType, substitutions),
        };
      case "ResultTypeExpression":
        return {
          kind: "result",
          inner: this.resolveTypeExpression(expr.innerType, substitutions),
        };
      case "UnionTypeExpression": {
        const members = Array.isArray(expr.members)
          ? expr.members.map((member) => this.resolveTypeExpression(member, substitutions))
          : [];
        return { kind: "union", members };
      }
      case "FunctionTypeExpression": {
        const parameters = Array.isArray(expr.paramTypes)
          ? expr.paramTypes.map((param) => this.resolveTypeExpression(param, substitutions))
          : [];
        const returnType = this.resolveTypeExpression(expr.returnType, substitutions);
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

  private instantiateTypeAlias(
    definition: AST.TypeAliasDefinition,
    typeArguments: TypeInfo[],
    outerSubstitutions?: Map<string, TypeInfo>,
  ): TypeInfo {
    const substitution = outerSubstitutions ? new Map(outerSubstitutions) : new Map<string, TypeInfo>();
    if (Array.isArray(definition.genericParams)) {
      definition.genericParams.forEach((param, index) => {
        const name = this.getIdentifierName(param?.name);
        if (!name) {
          return;
        }
        const arg = typeArguments[index] ?? unknownType;
        substitution.set(name, arg);
      });
    }
    return this.resolveTypeExpression(definition.targetType, substitution);
  }

  private typeInfosEquivalent(a: TypeInfo | undefined, b: TypeInfo | undefined): boolean {
    if (!a || a.kind === "unknown" || !b || b.kind === "unknown") {
      return true;
    }
    let left: TypeInfo = a;
    let right: TypeInfo = b;
    const normalizedLeft = this.canonicalizeStructuralType(left);
    const normalizedRight = this.canonicalizeStructuralType(right);
    if (normalizedLeft !== left || normalizedRight !== right) {
      return this.typeInfosEquivalent(normalizedLeft, normalizedRight);
    }
    left = normalizedLeft;
    right = normalizedRight;
    if (left.kind === "primitive" && right.kind === "primitive") {
      if (left.literal && this.literalFitsPrimitive(left.literal, right.name, left.name)) {
        return true;
      }
      if (right.literal && this.literalFitsPrimitive(right.literal, left.name, right.name)) {
        return true;
      }
      return left.name === right.name;
    }
    if (left.kind !== right.kind) {
      return false;
    }
    switch (left.kind) {
      case "array": {
        const other = right as Extract<TypeInfo, { kind: "array" }>;
        return this.typeInfosEquivalent(left.element, other.element);
      }
      case "map": {
        const other = right as Extract<TypeInfo, { kind: "map" }>;
        return this.typeInfosEquivalent(left.key, other.key) && this.typeInfosEquivalent(left.value, other.value);
      }
      case "iterator":
      case "range": {
        const other = right as Extract<TypeInfo, { kind: typeof left.kind }>;
        return this.typeInfosEquivalent(left.element, other.element);
      }
      case "proc":
      case "future": {
        const other = right as Extract<TypeInfo, { kind: typeof left.kind }>;
        return this.typeInfosEquivalent(left.result, other.result);
      }
      case "nullable":
      case "result": {
        const other = right as Extract<TypeInfo, { kind: typeof left.kind }>;
        return this.typeInfosEquivalent(left.inner, other.inner);
      }
      case "union": {
        const otherMembers = (right as typeof left).members ?? [];
        if (left.members.length !== otherMembers.length) {
          return false;
        }
        for (let i = 0; i < left.members.length; i += 1) {
          if (!this.typeInfosEquivalent(left.members[i], otherMembers[i])) {
            return false;
          }
        }
        return true;
      }
      default:
        return formatType(a) === formatType(b);
    }
  }

  private canWidenIntegerType(actual: PrimitiveName, expected: PrimitiveName): boolean {
    const actualInfo = getIntegerTypeInfo(actual);
    const expectedInfo = getIntegerTypeInfo(expected);
    if (!actualInfo || !expectedInfo) {
      return false;
    }
    return actualInfo.min >= expectedInfo.min && actualInfo.max <= expectedInfo.max;
  }

  private isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean {
    if (!actual || actual.kind === "unknown" || !expected || expected.kind === "unknown") {
      return true;
    }
    const normalizedActual = this.canonicalizeStructuralType(actual);
    const normalizedExpected = this.canonicalizeStructuralType(expected);
    if (this.typeInfosEquivalent(normalizedActual, normalizedExpected)) {
      return true;
    }
    if (normalizedExpected.kind === "nullable") {
      if (normalizedActual.kind === "primitive" && normalizedActual.name === "nil") {
        return true;
      }
      return this.isTypeAssignable(normalizedActual, normalizedExpected.inner);
    }
    if (normalizedExpected.kind === "result") {
      if (this.isTypeAssignable(normalizedActual, normalizedExpected.inner)) {
        return true;
      }
      return false;
    }
    if (normalizedExpected.kind === "union" && Array.isArray(normalizedExpected.members)) {
      return normalizedExpected.members.some((member) => this.isTypeAssignable(normalizedActual, member));
    }
    if (normalizedActual.kind === "primitive" && normalizedExpected.kind === "primitive") {
      if (this.canWidenIntegerType(normalizedActual.name, normalizedExpected.name)) {
        return true;
      }
    }
    return false;
  }

  private literalValueToBigInt(literal: LiteralInfo): bigint {
    if (typeof literal.value === "bigint") {
      return literal.value;
    }
    if (!Number.isFinite(literal.value)) {
      return BigInt(0);
    }
    return BigInt(Math.trunc(literal.value));
  }

  private literalFitsPrimitive(literal: LiteralInfo, expected: PrimitiveName, literalType: PrimitiveName): boolean {
    if (literal.literalKind === "integer") {
      if (literal.explicit) {
        return literalType === expected;
      }
      if (!hasIntegerBounds(expected)) {
        return literalType === expected;
      }
      const bounds = integerBounds(expected);
      const value = this.literalValueToBigInt(literal);
      return value >= bounds.min && value <= bounds.max;
    }
    if (literal.literalKind === "float") {
      if (literal.explicit) {
        return literalType === expected;
      }
      return expected === "f32" || expected === "f64";
    }
    return false;
  }

  public describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null {
    if (!actual || !expected) {
      return null;
    }
    let normalizedActual = actual;
    let normalizedExpected = expected;
    const nextActual = this.canonicalizeStructuralType(normalizedActual);
    const nextExpected = this.canonicalizeStructuralType(normalizedExpected);
    if (nextActual !== normalizedActual || nextExpected !== normalizedExpected) {
      return this.describeLiteralMismatch(nextActual, nextExpected);
    }
    normalizedActual = nextActual;
    normalizedExpected = nextExpected;
    if (normalizedActual.kind === "array" && normalizedExpected.kind === "array") {
      return this.describeLiteralMismatch(normalizedActual.element, normalizedExpected.element);
    }
    if (normalizedActual.kind === "map" && normalizedExpected.kind === "map") {
      return (
        this.describeLiteralMismatch(normalizedActual.key, normalizedExpected.key) ??
        this.describeLiteralMismatch(normalizedActual.value, normalizedExpected.value)
      );
    }
    if (normalizedActual.kind === "iterator" && normalizedExpected.kind === "iterator") {
      return this.describeLiteralMismatch(normalizedActual.element, normalizedExpected.element);
    }
    if (normalizedActual.kind === "range" && normalizedExpected.kind === "range") {
      const elementMessage = this.describeLiteralMismatch(
        normalizedActual.element,
        normalizedExpected.element,
      );
      if (elementMessage) {
        return elementMessage;
      }
      if (Array.isArray(normalizedActual.bounds)) {
        for (const bound of normalizedActual.bounds) {
          const boundMessage = this.describeLiteralMismatch(bound, normalizedExpected.element);
          if (boundMessage) {
            return boundMessage;
          }
        }
      }
      return null;
    }
    if (normalizedActual.kind === "proc" && normalizedExpected.kind === "proc") {
      return this.describeLiteralMismatch(normalizedActual.result, normalizedExpected.result);
    }
    if (normalizedActual.kind === "future" && normalizedExpected.kind === "future") {
      return this.describeLiteralMismatch(normalizedActual.result, normalizedExpected.result);
    }
    if (normalizedActual.kind === "nullable" && normalizedExpected.kind === "nullable") {
      return this.describeLiteralMismatch(normalizedActual.inner, normalizedExpected.inner);
    }
    if (normalizedActual.kind === "result" && normalizedExpected.kind === "result") {
      return this.describeLiteralMismatch(normalizedActual.inner, normalizedExpected.inner);
    }
    if (normalizedActual.kind === "union" && normalizedExpected.kind === "union") {
      const count = Math.min(normalizedActual.members.length, normalizedExpected.members.length);
      for (let i = 0; i < count; i += 1) {
        const message = this.describeLiteralMismatch(normalizedActual.members[i], normalizedExpected.members[i]);
        if (message) {
          return message;
        }
      }
      return null;
    }
    if (normalizedActual.kind === "union") {
      for (const member of normalizedActual.members) {
        const message = this.describeLiteralMismatch(member, normalizedExpected);
        if (message) {
          return message;
        }
      }
      return null;
    }
    if (normalizedExpected.kind === "union") {
      for (const member of normalizedExpected.members) {
        const message = this.describeLiteralMismatch(normalizedActual, member);
        if (message) {
          return message;
        }
      }
      return null;
    }
    if (normalizedActual.kind !== "primitive" || normalizedExpected.kind !== "primitive") {
      return null;
    }
    if (!normalizedActual.literal || normalizedActual.literal.literalKind !== "integer" || normalizedActual.literal.explicit) {
      return null;
    }
    if (!hasIntegerBounds(normalizedExpected.name)) {
      return null;
    }
    const bounds = integerBounds(normalizedExpected.name);
    const value = this.literalValueToBigInt(normalizedActual.literal);
    if (value < bounds.min || value > bounds.max) {
      return `typechecker: literal ${value.toString()} does not fit in ${normalizedExpected.name}`;
    }
    return null;
  }
  private canonicalizeStructuralType(type: TypeInfo): TypeInfo {
    if (!type || type.kind !== "struct") {
      return type;
    }
    const args = Array.isArray(type.typeArguments) ? type.typeArguments : [];
    const firstArg = args[0] ?? unknownType;
    switch (type.name) {
      case "Array":
        return { kind: "array", element: firstArg ?? unknownType };
      case "Iterator":
        return { kind: "iterator", element: firstArg ?? unknownType };
      case "Range":
        return { kind: "range", element: firstArg ?? unknownType };
      case "Proc":
        return { kind: "proc", result: firstArg ?? unknownType };
      case "Future":
        return { kind: "future", result: firstArg ?? unknownType };
      case "Map": {
        const key = args[0] ?? unknownType;
        const value = args[1] ?? unknownType;
        return { kind: "map", key, value };
      }
      default:
        return type;
    }
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
    if (!module || !Array.isArray(module.imports) || module.imports.length === 0) {
      return;
    }
    for (const imp of module.imports) {
      this.applyImportStatement(imp);
    }
  }

  private applyImportStatement(imp: AST.ImportStatement | null | undefined): void {
    if (!imp) {
      return;
    }
    const packageName = this.formatImportPath(imp.packagePath);
    const summary = packageName ? this.packageSummaries.get(packageName) : undefined;
    if (!summary) {
      const label = packageName ?? "<unknown>";
      this.report(`typechecker: import references unknown package '${label}'`, imp);
      return;
    }
    if (summary.visibility === "private" && summary.name !== this.currentPackageName) {
      this.report(`typechecker: package '${summary.name}' is private`, imp);
      return;
    }
    if (imp.isWildcard) {
      if (summary.symbols) {
        for (const symbolName of Object.keys(summary.symbols)) {
          if (!this.env.has(symbolName)) {
            this.env.define(symbolName, unknownType);
          }
        }
      }
      return;
    }
    if (Array.isArray(imp.selectors) && imp.selectors.length > 0) {
      for (const selector of imp.selectors) {
        if (!selector) continue;
        const selectorName = this.getIdentifierName(selector.name);
        if (!selectorName) continue;
        const aliasName = this.getIdentifierName(selector.alias) ?? selectorName;
        const hasSymbol = !!summary.symbols?.[selectorName];
        if (!hasSymbol) {
          if (summary.privateSymbols?.[selectorName]) {
            const label = packageName ?? "<unknown>";
            this.report(`typechecker: package '${label}' symbol '${selectorName}' is private`, selector);
          } else {
            const label = packageName ?? "<unknown>";
            this.report(`typechecker: package '${label}' has no symbol '${selectorName}'`, selector);
          }
          continue;
        }
        if (!this.env.has(aliasName)) {
          this.env.define(aliasName, unknownType);
        }
      }
      return;
    }
    const aliasName = this.getIdentifierName(imp.alias) ?? this.defaultPackageAlias(packageName);
    if (!aliasName) {
      return;
    }
    if (packageName) {
      this.packageAliases.set(aliasName, packageName);
    }
    if (!this.env.has(aliasName)) {
      this.env.define(aliasName, unknownType);
    }
  }

  private applyDynImportStatement(statement: AST.DynImportStatement | null | undefined): void {
    if (!statement) {
      return;
    }
    const placeholder = unknownType;
    if (statement.isWildcard) {
      this.allowDynamicLookups = true;
      return;
    }
    if (Array.isArray(statement.selectors) && statement.selectors.length > 0) {
      for (const selector of statement.selectors) {
        if (!selector) continue;
        const selectorName = this.getIdentifierName(selector.name);
        if (!selectorName) continue;
        const aliasName = this.getIdentifierName(selector.alias) ?? selectorName;
        if (!this.env.has(aliasName)) {
          this.env.define(aliasName, placeholder);
        }
      }
      return;
    }
    const aliasName = this.getIdentifierName(statement.alias);
    if (aliasName && !this.env.has(aliasName)) {
      this.env.define(aliasName, placeholder);
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
        this.report(
          `typechecker: package '${packageName}' has no symbol '${memberName}'`,
          expression.member ?? expression,
        );
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
    ctx.isKnownTypeName = (name: string) => this.isKnownTypeName(name);
    ctx.hasTypeDefinition = (name: string) =>
      this.structDefinitions.has(name) || this.interfaceDefinitions.has(name) || this.typeAliases.has(name);
    ctx.typeInfosEquivalent = this.typeInfosEquivalent.bind(this);
    ctx.isTypeAssignable = this.isTypeAssignable.bind(this);
    ctx.describeLiteralMismatch = this.describeLiteralMismatch.bind(this);
    ctx.resolveTypeExpression = this.resolveTypeExpression.bind(this);
    ctx.getStructDefinition = (name: string) => this.structDefinitions.get(name);
    ctx.getInterfaceDefinition = (name: string) => this.interfaceDefinitions.get(name);
    ctx.hasInterfaceDefinition = (name: string) => this.interfaceDefinitions.has(name);
    ctx.handlePackageMemberAccess = this.handlePackageMemberAccess.bind(this);
    ctx.pushAsyncContext = this.pushAsyncContext.bind(this);
    ctx.popAsyncContext = this.popAsyncContext.bind(this);
    ctx.checkReturnStatement = this.checkReturnStatement.bind(this);
    ctx.checkFunctionCall = this.checkFunctionCall.bind(this);
    ctx.inferFunctionCallReturnType = this.inferFunctionCallReturnType.bind(this);
    ctx.checkFunctionDefinition = this.checkFunctionDefinition.bind(this);
    ctx.pushLoopContext = this.pushLoopContext.bind(this);
    ctx.popLoopContext = () => this.popLoopContext();
    ctx.inLoopContext = () => this.inLoopContext();
    ctx.pushScope = () => this.env.pushScope();
    ctx.popScope = () => this.env.popScope();
    ctx.withForkedEnv = <T>(fn: () => T) => this.withForkedEnv(fn);
    ctx.lookupIdentifier = (name: string) => this.env.lookup(name);
    ctx.defineValue = (name: string, valueType: TypeInfo) => this.env.define(name, valueType);
    ctx.assignValue = (name: string, valueType: TypeInfo) => this.env.assign(name, valueType);
    ctx.hasBinding = (name: string) => this.env.has(name);
    ctx.hasBindingInCurrentScope = (name: string) => this.env.hasInCurrentScope(name);
    ctx.allowDynamicLookup = () => this.allowDynamicLookups;
    ctx.getFunctionInfo = (key: string) => this.functionInfos.get(key);
    ctx.setFunctionInfo = (key: string, info: FunctionInfo) => this.functionInfos.set(key, info);
    ctx.isExpression = (node: AST.Node | undefined | null): node is AST.Expression => this.isExpression(node);
    ctx.handleTypeDeclaration = (node) => this.checkLocalTypeDeclaration(node as LocalTypeDeclaration);
    ctx.pushBreakpointLabel = (label: string) => this.pushBreakpointLabel(label);
    ctx.popBreakpointLabel = () => this.popBreakpointLabel();
    ctx.hasBreakpointLabel = (label: string) => this.hasBreakpointLabel(label);
    ctx.handleBreakStatement = this.checkBreakStatement.bind(this);
    ctx.handleContinueStatement = this.checkContinueStatement.bind(this);

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
