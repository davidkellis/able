import type * as AST from "../ast";
import { Environment } from "./environment";
import {
  describe,
  isBoolean,
  isNumeric,
  formatType,
  primitiveType,
  unknownType,
  type TypeInfo,
} from "./types";
import type {
  DiagnosticLocation,
  TypecheckerDiagnostic,
  TypecheckResult,
  PackageSummary,
  ExportedFunctionSummary,
  ExportedGenericParamSummary,
  ExportedImplementationSummary,
  ExportedInterfaceSummary,
  ExportedMethodSetSummary,
  ExportedObligationSummary,
  ExportedStructSummary,
  ExportedSymbolSummary,
  ExportedWhereConstraintSummary,
} from "./diagnostics";

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

  constructor(options: TypeCheckerOptions = {}) {
    this.env = new Environment();
    this.options = options;
    this.packageSummaries = this.clonePackageSummaries(options.packageSummaries);
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
    this.env.define("print", primitiveType("void"));
    this.env.define("proc_yield", primitiveType("void"));
    this.env.define("proc_cancelled", primitiveType("bool"));
    this.env.define("proc_flush", primitiveType("void"));
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
        this.collectMethodsDefinition(node);
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
      this.collectImplementationDefinition(node);
    }
  }

  private checkStatement(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) return;
    switch (node.type) {
      case "InterfaceDefinition":
      case "StructDefinition":
      case "ImplementationDefinition":
      case "MethodsDefinition":
      case "FunctionDefinition":
        // Declarations are handled during collection.
        return;
      case "AssignmentExpression":
        this.checkAssignment(node);
        return;
      default:
        if (this.isExpression(node)) {
          this.inferExpression(node);
        }
        return;
    }
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

  private collectMethodsDefinition(definition: AST.MethodsDefinition): void {
    const structLabel =
      this.formatImplementationTarget(definition.targetType) ?? this.getIdentifierNameFromTypeExpression(definition.targetType);
    if (!structLabel) return;
    const record: MethodSetRecord = {
      label: `methods for ${structLabel}`,
      target: definition.targetType,
      genericParams: Array.isArray(definition.genericParams)
        ? definition.genericParams
            .map((param) => this.getIdentifierName(param?.name))
            .filter((name): name is string => Boolean(name))
        : [],
      obligations: this.extractMethodSetObligations(definition),
      definition,
    };
    this.methodSets.push(record);
    if (Array.isArray(definition.definitions)) {
      for (const entry of definition.definitions) {
        if (entry?.type === "FunctionDefinition") {
          this.collectFunctionDefinition(entry, { structName: structLabel });
        }
      }
    }
  }

  private collectImplementationDefinition(definition: AST.ImplementationDefinition): void {
    const interfaceName = this.getIdentifierName(definition.interfaceName);
    if (!interfaceName) {
      return;
    }
    const targetLabel = this.formatImplementationTarget(definition.targetType);
    const fallbackName = this.getIdentifierNameFromTypeExpression(definition.targetType);
    const contextName = targetLabel ?? fallbackName ?? "<unknown>";
    const targetKey = contextName;
    const interfaceDefinition = this.interfaceDefinitions.get(interfaceName);
    if (!interfaceDefinition) {
      const fallback = this.getIdentifierNameFromTypeExpression(definition.targetType);
      this.report(
        `typechecker: impl for ${fallback ?? "<unknown>"} references unknown interface '${interfaceName}'`,
        definition,
      );
      return;
    }
    this.validateImplementationInterfaceArguments(definition, interfaceDefinition, contextName, interfaceName);
    const hasRequiredMethods = this.ensureImplementationMethods(
      definition,
      interfaceDefinition,
      contextName,
      interfaceName,
    );
    if (hasRequiredMethods) {
      const record = this.createImplementationRecord(definition, interfaceName, contextName, targetKey);
      if (record) {
        this.registerImplementationRecord(record);
      }
    }

    if (Array.isArray(definition.definitions)) {
      for (const entry of definition.definitions) {
        if (entry?.type === "FunctionDefinition") {
          this.collectFunctionDefinition(entry, { structName: contextName });
        }
      }
    }
  }

  private createImplementationRecord(
    definition: AST.ImplementationDefinition,
    interfaceName: string,
    targetLabel: string,
    targetKey: string,
  ): ImplementationRecord | null {
    if (!definition.targetType) {
      return null;
    }
    const genericParams = Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => this.getIdentifierName(param?.name))
          .filter((name): name is string => Boolean(name))
      : [];
    const obligations = this.extractImplementationObligations(definition);
    const interfaceArgs = Array.isArray(definition.interfaceArgs)
      ? definition.interfaceArgs.filter((arg): arg is AST.TypeExpression => Boolean(arg))
      : [];
    return {
      interfaceName,
      label: this.formatImplementationLabel(interfaceName, targetLabel),
      target: definition.targetType,
      targetKey,
      genericParams,
      obligations,
      interfaceArgs,
      definition,
    };
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

  private extractImplementationObligations(definition: AST.ImplementationDefinition): ImplementationObligation[] {
    const obligations: ImplementationObligation[] = [];
    const appendObligation = (
      typeParam: string | null,
      interfaceType: AST.TypeExpression | null | undefined,
      context: string,
    ) => {
      const interfaceName = this.getInterfaceNameFromTypeExpression(interfaceType);
      if (!typeParam || !interfaceName) {
        return;
      }
      obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
    };

    if (Array.isArray(definition.genericParams)) {
      for (const param of definition.genericParams) {
        const paramName = this.getIdentifierName(param?.name);
        if (!paramName || !Array.isArray(param?.constraints)) continue;
        for (const constraint of param.constraints) {
          appendObligation(paramName, constraint?.interfaceType, "generic constraint");
        }
      }
    }

    if (Array.isArray(definition.whereClause)) {
      for (const clause of definition.whereClause) {
        const typeParamName = this.getIdentifierName(clause?.typeParam);
        if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
        for (const constraint of clause.constraints) {
          appendObligation(typeParamName, constraint?.interfaceType, "where clause");
        }
      }
    }

    return obligations;
  }

  private extractMethodSetObligations(definition: AST.MethodsDefinition): ImplementationObligation[] {
    const obligations: ImplementationObligation[] = [];
    const appendObligation = (
      typeParam: string | null,
      interfaceType: AST.TypeExpression | null | undefined,
      context: string,
    ) => {
      const interfaceName = this.getInterfaceNameFromTypeExpression(interfaceType);
      if (!typeParam || !interfaceName) {
        return;
      }
      obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
    };

    if (Array.isArray(definition.genericParams)) {
      for (const param of definition.genericParams) {
        const paramName = this.getIdentifierName(param?.name);
        if (!paramName || !Array.isArray(param?.constraints)) continue;
        for (const constraint of param.constraints) {
          appendObligation(paramName, constraint?.interfaceType, "generic constraint");
        }
      }
    }

    if (Array.isArray(definition.whereClause)) {
      for (const clause of definition.whereClause) {
        const typeParamName = this.getIdentifierName(clause?.typeParam);
        if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
        for (const constraint of clause.constraints) {
          appendObligation(typeParamName, constraint?.interfaceType, "where clause");
        }
      }
    }

    return obligations;
  }

  private extractFunctionWhereObligations(definition: AST.FunctionDefinition): ImplementationObligation[] {
    const obligations: ImplementationObligation[] = [];
    const appendObligation = (
      typeParam: string | null,
      interfaceType: AST.TypeExpression | null | undefined,
      context: string,
    ) => {
      const interfaceName = this.getInterfaceNameFromTypeExpression(interfaceType);
      if (!typeParam || !interfaceName) {
        return;
      }
      obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
    };

    if (Array.isArray(definition.genericParams)) {
      for (const param of definition.genericParams) {
        const paramName = this.getIdentifierName(param?.name);
        if (!paramName || !Array.isArray(param?.constraints)) continue;
        for (const constraint of param.constraints) {
          appendObligation(paramName, constraint?.interfaceType, "generic constraint");
        }
      }
    }

    if (Array.isArray(definition.whereClause)) {
      for (const clause of definition.whereClause) {
        const typeParamName = this.getIdentifierName(clause?.typeParam);
        if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
        for (const constraint of clause.constraints) {
          appendObligation(typeParamName, constraint?.interfaceType, "where clause");
        }
      }
    }

    return obligations;
  }

  private collectFunctionDefinition(
    definition: AST.FunctionDefinition,
    context: FunctionContext | undefined,
  ): void {
    const name = definition.id?.name ?? "<anonymous>";
    const structName = context?.structName;
    const fullName = structName ? `${structName}::${name}` : name;

    const info: FunctionInfo = {
      name,
      fullName,
      structName,
      genericConstraints: [],
      whereClause: this.extractFunctionWhereObligations(definition),
      genericParamNames: Array.isArray(definition.genericParams)
        ? definition.genericParams
            .map((param) => this.getIdentifierName(param?.name))
            .filter((name): name is string => Boolean(name))
        : [],
    };

    if (Array.isArray(definition.genericParams)) {
      for (const param of definition.genericParams) {
        const paramName = param.name?.name ?? "T";
        if (!Array.isArray(param.constraints)) continue;
        for (const constraint of param.constraints) {
          const interfaceName = this.getInterfaceNameFromConstraint(constraint);
          if (!interfaceName) continue;
          const interfaceDefined = this.interfaceDefinitions.has(interfaceName);
          if (!interfaceDefined) {
            const message = structName
              ? `typechecker: methods for ${structName}::${name} constraint on ${paramName} references unknown interface '${interfaceName}'`
              : `typechecker: fn ${name} constraint on ${paramName} references unknown interface '${interfaceName}'`;
            this.report(message, constraint?.interfaceType ?? constraint ?? definition);
          }
          info.genericConstraints.push({ paramName, interfaceName, interfaceDefined, interfaceType: constraint.interfaceType });
        }
      }
    }

    this.functionInfos.set(fullName, info);
    if (!structName) {
      this.functionInfos.set(name, info);
    }
  }

  private ensureImplementationMethods(
    implementation: AST.ImplementationDefinition,
    interfaceDefinition: AST.InterfaceDefinition,
    targetLabel: string,
    interfaceName: string,
  ): boolean {
    const provided = new Map<string, AST.FunctionDefinition>();
    if (Array.isArray(implementation.definitions)) {
      for (const fn of implementation.definitions) {
        if (!fn || fn.type !== "FunctionDefinition") continue;
        const methodName = fn.id?.name;
        if (!methodName) continue;
        if (provided.has(methodName)) {
          const label = this.formatImplementationLabel(interfaceName, targetLabel);
          this.report(`typechecker: ${label} defines duplicate method '${methodName}'`, fn);
          continue;
        }
        provided.set(methodName, fn);
      }
    }

    const signatures = Array.isArray(interfaceDefinition.signatures) ? interfaceDefinition.signatures : [];
    if (signatures.length === 0) {
      return true;
    }

    const label = this.formatImplementationLabel(interfaceName, targetLabel);
    let allRequiredPresent = true;

    for (const signature of signatures) {
      if (!signature) continue;
      const methodName = this.getIdentifierName(signature.name);
      if (!methodName) continue;
      if (!provided.has(methodName)) {
        this.report(`typechecker: ${label} missing method '${methodName}'`, implementation);
        allRequiredPresent = false;
        continue;
      }
        const method = provided.get(methodName);
        if (method) {
          const methodValid = this.validateImplementationMethod(
            interfaceDefinition,
          implementation,
          signature,
          method,
          label,
          targetLabel,
        );
        if (!methodValid) {
          allRequiredPresent = false;
        }
        provided.delete(methodName);
      }
    }

    for (const methodName of provided.keys()) {
      const extraMethod = provided.get(methodName);
      this.report(
        `typechecker: ${label} defines method '${methodName}' not declared in interface ${interfaceName}`,
        extraMethod ?? implementation,
      );
    }

    return allRequiredPresent;
  }

  private validateImplementationInterfaceArguments(
    implementation: AST.ImplementationDefinition,
    interfaceDefinition: AST.InterfaceDefinition,
    targetLabel: string,
    interfaceName: string,
  ): void {
    const expected = Array.isArray(interfaceDefinition.genericParams) ? interfaceDefinition.genericParams.length : 0;
    const provided = Array.isArray(implementation.interfaceArgs) ? implementation.interfaceArgs.length : 0;
    if (expected === 0 && provided > 0) {
      this.report(`typechecker: impl ${interfaceName} does not accept type arguments`, implementation);
      return;
    }
    if (expected > 0) {
      const targetDescription = targetLabel;
      if (provided === 0) {
        this.report(
          `typechecker: impl ${interfaceName} for ${targetDescription} requires ${expected} interface type argument(s)`,
          implementation,
        );
        return;
      }
      if (provided !== expected) {
        this.report(
          `typechecker: impl ${interfaceName} for ${targetDescription} expected ${expected} interface type argument(s), got ${provided}`,
          implementation,
        );
      }
    }
  }

  private formatImplementationLabel(interfaceName: string, targetName: string): string {
    return `impl ${interfaceName} for ${targetName}`;
  }

  private formatImplementationTarget(targetType: AST.TypeExpression | null | undefined): string | null {
    if (!targetType) return null;
    return this.formatTypeExpression(targetType);
  }

  private buildInterfaceSubstitutions(
    interfaceDefinition: AST.InterfaceDefinition,
    implementationDefinition: AST.ImplementationDefinition,
    targetName: string,
  ): Map<string, string> {
    const substitutions = new Map<string, string>();
    const targetDescription = implementationDefinition.targetType
      ? this.formatTypeExpression(implementationDefinition.targetType)
      : targetName;
    if (targetDescription) {
      substitutions.set("Self", targetDescription);
    }

    const interfaceParams = Array.isArray(interfaceDefinition.genericParams) ? interfaceDefinition.genericParams : [];
    const interfaceArgs = Array.isArray(implementationDefinition.interfaceArgs) ? implementationDefinition.interfaceArgs : [];
    interfaceParams.forEach((param, index) => {
      const paramName = param?.name?.name;
      if (!paramName) return;
      const arg = interfaceArgs[index];
      if (arg) {
        substitutions.set(paramName, this.formatTypeExpression(arg));
      }
    });
    return substitutions;
  }

  private validateImplementationMethod(
    interfaceDefinition: AST.InterfaceDefinition,
    implementationDefinition: AST.ImplementationDefinition,
    signature: AST.FunctionSignature,
    implementation: AST.FunctionDefinition,
    label: string,
    targetName: string,
  ): boolean {
    let valid = true;
    const substitutions = this.buildInterfaceSubstitutions(interfaceDefinition, implementationDefinition, targetName);

    const interfaceGenerics = Array.isArray(signature.genericParams) ? signature.genericParams.length : 0;
    const implementationGenerics = Array.isArray(implementation.genericParams) ? implementation.genericParams.length : 0;
    if (interfaceGenerics !== implementationGenerics) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceGenerics} generic parameter(s), got ${implementationGenerics}`,
        implementation,
      );
      valid = false;
    }

    const interfaceParams = Array.isArray(signature.params) ? signature.params : [];
    const implementationParams = Array.isArray(implementation.params) ? implementation.params : [];
    if (interfaceParams.length !== implementationParams.length) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceParams.length} parameter(s), got ${implementationParams.length}`,
        implementation,
      );
      valid = false;
    }

    const paramCount = Math.min(interfaceParams.length, implementationParams.length);
    for (let index = 0; index < paramCount; index += 1) {
      const expected = interfaceParams[index]?.paramType ?? null;
      const actual = implementationParams[index]?.paramType ?? null;
      if (!this.typeExpressionsEquivalent(expected, actual, substitutions)) {
        const expectedDescription = this.describeTypeExpression(expected, substitutions);
        const actualDescription = this.describeTypeExpression(actual, substitutions);
        this.report(
          `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' parameter ${index + 1} expected ${expectedDescription}, got ${actualDescription}`,
          implementation,
        );
        valid = false;
      }
    }

    const returnExpected = signature.returnType ?? null;
    const returnActual = implementation.returnType ?? null;
    if (!this.typeExpressionsEquivalent(returnExpected, returnActual, substitutions)) {
      const expectedDescription = this.describeTypeExpression(returnExpected, substitutions);
      const actualDescription = this.describeTypeExpression(returnActual, substitutions);
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' return type expected ${expectedDescription}, got ${actualDescription}`,
        implementation,
      );
      valid = false;
    }

    // Basic where-clause compatibility: ensure implementation does not omit required where clauses.
    const interfaceWhere = Array.isArray(signature.whereClause) ? signature.whereClause.length : 0;
    const implementationWhere = Array.isArray(implementation.whereClause) ? implementation.whereClause.length : 0;
    if (interfaceWhere !== implementationWhere) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceWhere} where-clause constraint(s), got ${implementationWhere}`,
        implementation,
      );
      valid = false;
    }

    // Placeholder: ensure method privacy matches interface expectations (interfaces methods are always public).
    if (implementation.isPrivate) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' must be public to satisfy interface`,
        implementation,
      );
      valid = false;
    }

    return valid;
  }

  private checkAssignment(node: AST.AssignmentExpression): void {
    const valueType = this.inferExpression(node.right);
    if (node.left?.type === "StructPattern") {
      this.checkStructPattern(node.left, valueType);
      return;
    }
    if (node.left?.type === "Identifier") {
      this.env.define(node.left.name, valueType);
    }
  }

  private checkStructPattern(pattern: AST.StructPattern, valueType: TypeInfo): void {
    const definition = this.resolveStructDefinitionForPattern(pattern, valueType);
    if (!definition) return;

    const knownFields = new Set<string>();
    if (Array.isArray(definition.fields)) {
      for (const field of definition.fields) {
        const fieldName = this.getIdentifierName(field?.name);
        if (fieldName) knownFields.add(fieldName);
      }
    }

    if (!Array.isArray(pattern.fields)) {
      return;
    }

    for (const field of pattern.fields) {
      if (!field) continue;
      const fieldName = this.getIdentifierName(field.fieldName);
      if (fieldName && !knownFields.has(fieldName)) {
        this.report(`typechecker: struct pattern field '${fieldName}' not found`, field ?? pattern);
      }
    }
  }

  private inferExpression(expression: AST.Expression | undefined | null): TypeInfo {
    if (!expression) return unknownType;
    switch (expression.type) {
      case "StringLiteral":
        return primitiveType("string");
      case "BooleanLiteral":
        return primitiveType("bool");
      case "IntegerLiteral":
        return primitiveType("i32");
      case "FloatLiteral":
        return primitiveType("f64");
      case "NilLiteral":
        return primitiveType("nil");
      case "Identifier": {
        const existing = this.env.lookup(expression.name);
        return existing ?? unknownType;
      }
      case "BinaryExpression": {
        const left = this.inferExpression(expression.left);
        const right = this.inferExpression(expression.right);
        if (expression.operator === "&&" || expression.operator === "||") {
          if (!isBoolean(left)) {
            this.report(`typechecker: '${expression.operator}' left operand must be bool (got ${describe(left)})`, expression);
          }
          if (!isBoolean(right)) {
            this.report(`typechecker: '${expression.operator}' right operand must be bool (got ${describe(right)})`, expression);
          }
          return primitiveType("bool");
        }
        return unknownType;
      }
      case "RangeExpression": {
        const start = this.inferExpression(expression.start);
        if (!isNumeric(start)) {
          this.report("typechecker: range start must be numeric", expression);
        }
        const end = this.inferExpression(expression.end);
        if (!isNumeric(end)) {
          this.report("typechecker: range end must be numeric", expression);
        }
        return unknownType;
      }
      case "FunctionCall":
        this.checkFunctionCall(expression);
        return unknownType;
      case "StructLiteral": {
        const structName = this.getIdentifierName(expression.structType);
        if (structName) {
          const typeArguments = Array.isArray(expression.typeArguments)
            ? expression.typeArguments.map((arg) => this.resolveTypeExpression(arg))
            : [];
          return {
            kind: "struct",
            name: structName,
            typeArguments,
            definition: this.structDefinitions.get(structName),
          };
        }
        return unknownType;
      }
      case "MemberAccessExpression":
        this.handlePackageMemberAccess(expression);
        return unknownType;
      default:
        return unknownType;
    }
  }

  private checkFunctionCall(call: AST.FunctionCall): void {
    const infos = this.resolveFunctionInfos(call.callee);
    if (!infos.length) {
      return;
    }
    for (const info of infos) {
      this.enforceFunctionConstraints(info, call);
    }
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
      const objectType = this.inferExpression(callee.object);
      if (objectType.kind === "struct") {
        const structLabel = formatType(objectType);
        const memberKey = `${structLabel}::${memberName}`;
        const infos: FunctionInfo[] = [];
        const seen = new Set<string>();
        const info = this.functionInfos.get(memberKey);
        const genericMatches = this.lookupMethodSetsForCall(structLabel, memberName, objectType);
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

  private lookupMethodSetsForCall(structLabel: string, methodName: string, objectType: TypeInfo): FunctionInfo[] {
    if (!this.methodSets.length) {
      return [];
    }
    const results: FunctionInfo[] = [];
    for (const record of this.methodSets) {
      const paramNames = new Set(record.genericParams);
      const substitutions = new Map<string, TypeInfo>();
      substitutions.set("Self", objectType);
      if (!this.matchImplementationTarget(objectType, record.target, paramNames, substitutions)) {
        continue;
      }
      const method = record.definition.definitions?.find(
        (fn): fn is AST.FunctionDefinition => fn?.type === "FunctionDefinition" && fn.id?.name === methodName,
      );
      if (!method) {
        continue;
      }
      const methodGenericNames = Array.isArray(method.genericParams)
        ? method.genericParams
            .map((param) => this.getIdentifierName(param?.name))
            .filter((name): name is string => Boolean(name))
        : [];
      const info: FunctionInfo = {
        name: methodName,
        fullName: `${record.label}::${methodName}`,
        structName: structLabel,
        genericConstraints: [],
        genericParamNames: methodGenericNames,
        whereClause: record.obligations,
        methodSetSubstitutions: Array.from(substitutions.entries()),
      };
      if (Array.isArray(method.genericParams)) {
        for (const param of method.genericParams) {
          const paramName = this.getIdentifierName(param?.name);
          if (!paramName || !Array.isArray(param?.constraints)) continue;
          for (const constraint of param.constraints) {
            const interfaceName = this.getInterfaceNameFromConstraint(constraint);
            info.genericConstraints.push({
              paramName,
              interfaceName: interfaceName ?? "<unknown>",
              interfaceDefined: !!interfaceName,
              interfaceType: constraint?.interfaceType,
            });
          }
        }
      }
      results.push(info);
    }
    return results;
  }

  private enforceFunctionConstraints(info: FunctionInfo, call: AST.FunctionCall): void {
    const typeArgs = Array.isArray(call.typeArguments) ? call.typeArguments : [];
    const substitutions = new Map<string, TypeInfo>();
    if (info.methodSetSubstitutions) {
      for (const [key, value] of info.methodSetSubstitutions) {
        substitutions.set(key, value);
      }
    } else if (call.callee?.type === "MemberAccessExpression") {
      const selfType = this.inferExpression(call.callee.object);
      if (selfType.kind !== "unknown") {
        substitutions.set("Self", selfType);
      }
    }
    info.genericParamNames.forEach((paramName, idx) => {
      const argExpr = typeArgs[idx];
      if (!paramName || !argExpr) return;
      substitutions.set(paramName, this.resolveTypeExpression(argExpr));
    });

    if (info.genericConstraints.length > 0) {
      info.genericConstraints.forEach((constraint, index) => {
        const typeArgExpr = typeArgs[index];
        const typeArg = this.resolveTypeExpression(typeArgExpr);
        if (!constraint.interfaceDefined) {
          const message = info.structName
            ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${constraint.paramName} references unknown interface '${constraint.interfaceName}'`
            : `typechecker: fn ${info.name} constraint on ${constraint.paramName} references unknown interface '${constraint.interfaceName}'`;
          this.report(message, typeArgExpr ?? call);
          return;
        }
        const expectedArgs = this.resolveInterfaceArgumentLabels(constraint.interfaceType);
        const result = this.typeImplementsInterface(typeArg, constraint.interfaceName, expectedArgs);
        if (!result.ok) {
          const typeName = this.describeTypeArgument(typeArg);
          const detailSuffix = result.detail ? `: ${result.detail}` : "";
          const message = info.structName
            ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${constraint.paramName} is not satisfied: ${typeName} does not implement ${constraint.interfaceName}${detailSuffix}`
            : `typechecker: fn ${info.name} constraint on ${constraint.paramName} is not satisfied: ${typeName} does not implement ${constraint.interfaceName}${detailSuffix}`;
          this.report(message, typeArgExpr ?? call);
        }
      });
    }

    if (info.whereClause.length > 0) {
      for (const obligation of info.whereClause) {
        const subject = substitutions.get(obligation.typeParam);
        if (!subject) {
          continue;
        }
        const expectedArgs = this.resolveInterfaceArgumentLabels(
          obligation.interfaceType,
          this.buildStringSubstitutionMap(substitutions),
        );
        const result = this.typeImplementsInterface(subject, obligation.interfaceName, expectedArgs);
        if (!result.ok) {
          const subjectLabel = formatType(subject);
          const detailSuffix = result.detail ? `: ${result.detail}` : "";
          const message = info.structName
            ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${obligation.typeParam} is not satisfied: ${subjectLabel} does not implement ${obligation.interfaceName}${detailSuffix}`
            : `typechecker: fn ${info.name} constraint on ${obligation.typeParam} is not satisfied: ${subjectLabel} does not implement ${obligation.interfaceName}${detailSuffix}`;
          const typeArgIndex = info.genericParamNames.indexOf(obligation.typeParam);
          const explicitTypeArg =
            typeArgIndex >= 0 && typeArgIndex < typeArgs.length ? typeArgs[typeArgIndex] : undefined;
          const locationNode =
            explicitTypeArg ??
            (call.callee && call.callee.type === "MemberAccessExpression"
              ? call.callee.member ?? call.callee
              : call.callee ?? call);
          this.report(message, locationNode);
        }
      }
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

  private typeImplementsInterface(type: TypeInfo, interfaceName: string, expectedArgs: string[] = []): InterfaceCheckResult {
    if (!type || type.kind === "unknown") {
      return { ok: true };
    }
    if (type.kind === "nullable") {
      const impl = this.implementationProvidesInterface(type, interfaceName, expectedArgs);
      if (impl.ok) {
        return impl;
      }
      const inner = this.typeImplementsInterface(type.inner, interfaceName, expectedArgs);
      if (!inner.ok) {
        if (inner.detail) {
          return inner;
        }
        if (impl.detail) {
          return { ok: false, detail: impl.detail };
        }
        return inner;
      }
      if (impl.detail) {
        return { ok: false, detail: impl.detail };
      }
      return { ok: true };
    }
    if (type.kind === "result") {
      const impl = this.implementationProvidesInterface(type, interfaceName, expectedArgs);
      if (impl.ok) {
        return impl;
      }
      const inner = this.typeImplementsInterface(type.inner, interfaceName, expectedArgs);
      if (!inner.ok) {
        if (inner.detail) {
          return inner;
        }
        if (impl.detail) {
          return { ok: false, detail: impl.detail };
        }
        return inner;
      }
      if (impl.detail) {
        return { ok: false, detail: impl.detail };
      }
      return { ok: true };
    }
    if (type.kind === "union") {
      const impl = this.implementationProvidesInterface(type, interfaceName, expectedArgs);
      if (impl.ok) {
        return impl;
      }
      for (const member of type.members) {
        const result = this.typeImplementsInterface(member, interfaceName, expectedArgs);
        if (!result.ok) {
          if (result.detail) {
            return result;
          }
          if (impl.detail) {
            return { ok: false, detail: impl.detail };
          }
          return result;
        }
      }
      if (impl.detail) {
        return { ok: false, detail: impl.detail };
      }
      return { ok: true };
    }
    if (type.kind === "interface" && type.name === interfaceName) {
      return { ok: true };
    }
    const impl = this.implementationProvidesInterface(type, interfaceName, expectedArgs);
    if (impl.ok) {
      return impl;
    }
    if (impl.detail) {
      return { ok: false, detail: impl.detail };
    }
    return { ok: false };
  }

  private implementationProvidesInterface(
    type: TypeInfo,
    interfaceName: string,
    expectedArgs: string[] = [],
  ): InterfaceCheckResult {
    if (this.implementationRecords.length === 0) {
      return { ok: false };
    }
    const candidates = this.lookupImplementationCandidates(type);
    let bestDetail: string | undefined;
    for (const record of candidates) {
      if (record.interfaceName !== interfaceName) {
        continue;
      }
      const paramNames = new Set(record.genericParams);
      const substitutions = new Map<string, TypeInfo>();
      substitutions.set("Self", type);
      if (!this.matchImplementationTarget(type, record.target, paramNames, substitutions)) {
        continue;
      }
      const actualArgs = record.interfaceArgs.length
        ? this.resolveInterfaceArgumentLabelsFromArray(record.interfaceArgs, substitutions)
        : [];
      if (!this.interfaceArgsCompatible(actualArgs, expectedArgs)) {
        const expectedLabel = expectedArgs.length > 0 ? expectedArgs.join(" ") : "(none)";
        const detail = `${this.appendInterfaceArgsToLabel(record.label, actualArgs)}: interface arguments do not match expected ${expectedLabel}`;
        if (!bestDetail || detail.length > bestDetail.length) {
          bestDetail = detail;
        }
        continue;
      }
      let failedDetail: string | undefined;
      for (const obligation of record.obligations) {
        const subject = this.lookupObligationSubject(obligation.typeParam, substitutions, type);
        if (!subject) {
          continue;
        }
        const obligationArgs = this.resolveInterfaceArgumentLabels(obligation.interfaceType, substitutions);
        const result = this.typeImplementsInterface(subject, obligation.interfaceName, obligationArgs);
        if (!result.ok) {
          const detail = this.annotateImplementationFailure(record, obligation, subject, result.detail, actualArgs, obligationArgs);
          if (!bestDetail || detail.length > bestDetail.length) {
            bestDetail = detail;
          }
          failedDetail = detail;
          break;
        }
      }
      if (failedDetail) {
        continue;
      }
      return { ok: true };
    }
    return bestDetail ? { ok: false, detail: bestDetail } : { ok: false };
  }

  private lookupImplementationCandidates(type: TypeInfo): ImplementationRecord[] {
    const key = formatType(type);
    const seen = new Set<ImplementationRecord>();
    const direct = this.implementationIndex.get(key);
    if (direct) {
      for (const record of direct) {
        seen.add(record);
      }
    }
    for (const record of this.implementationRecords) {
      seen.add(record);
    }
    return Array.from(seen);
  }

  private matchImplementationTarget(
    actual: TypeInfo,
    target: AST.TypeExpression,
    paramNames: Set<string>,
    substitutions: Map<string, TypeInfo>,
  ): boolean {
    if (!target) {
      return false;
    }
    if (!actual || actual.kind === "unknown") {
      return true;
    }
    switch (target.type) {
      case "SimpleTypeExpression": {
        const name = this.getIdentifierName(target.name);
        if (!name) {
          return false;
        }
        if (name === "Self") {
          const existing = substitutions.get("Self");
          if (existing) {
            return this.typeInfosEquivalent(existing, actual);
          }
          substitutions.set("Self", actual);
          return true;
        }
        if (paramNames.has(name)) {
          const existing = substitutions.get(name);
          if (existing) {
            return this.typeInfosEquivalent(existing, actual);
          }
          substitutions.set(name, actual);
          return true;
        }
        if (actual.kind === "primitive") {
          return actual.name === name;
        }
        if (actual.kind === "struct") {
          return actual.name === name && (actual.typeArguments?.length ?? 0) === 0;
        }
        if (actual.kind === "interface") {
          return actual.name === name && (actual.typeArguments?.length ?? 0) === 0;
        }
        return formatType(actual) === name;
      }
      case "GenericTypeExpression": {
        const baseName = this.getIdentifierNameFromTypeExpression(target.base);
        if (!baseName) {
          return false;
        }
        if (paramNames.has(baseName)) {
          const existing = substitutions.get(baseName);
          if (existing) {
            return this.typeInfosEquivalent(existing, actual);
          }
          substitutions.set(baseName, actual);
          return true;
        }
        if (actual.kind !== "struct" && actual.kind !== "interface") {
          return false;
        }
        if (actual.name !== baseName) {
          return false;
        }
        const expectedArgs = Array.isArray(target.arguments) ? target.arguments : [];
        const actualArgs = actual.typeArguments ?? [];
        if (expectedArgs.length !== actualArgs.length) {
          return false;
        }
        for (let index = 0; index < expectedArgs.length; index += 1) {
          const expectedArg = expectedArgs[index];
          const actualArg = actualArgs[index] ?? unknownType;
          if (!expectedArg) {
            return false;
          }
          if (!this.matchImplementationTarget(actualArg, expectedArg, paramNames, substitutions)) {
            return false;
          }
        }
        return true;
      }
      case "NullableTypeExpression":
        if (actual.kind !== "nullable") {
          return false;
        }
        return this.matchImplementationTarget(actual.inner, target.innerType, paramNames, substitutions);
      case "ResultTypeExpression":
        if (actual.kind !== "result") {
          return false;
        }
        return this.matchImplementationTarget(actual.inner, target.innerType, paramNames, substitutions);
      case "UnionTypeExpression": {
        if (actual.kind !== "union") {
          return false;
        }
        const expectedMembers = Array.isArray(target.members) ? target.members : [];
        if (expectedMembers.length !== actual.members.length) {
          return false;
        }
        for (let index = 0; index < expectedMembers.length; index += 1) {
          const expectedMember = expectedMembers[index];
          const actualMember = actual.members[index];
          if (!expectedMember) {
            return false;
          }
          if (!this.matchImplementationTarget(actualMember, expectedMember, paramNames, substitutions)) {
            return false;
          }
        }
        return true;
      }
      case "FunctionTypeExpression":
        return actual.kind === "function";
      default:
        return formatType(actual) === this.formatTypeExpression(target);
    }
  }

  private lookupObligationSubject(
    typeParam: string,
    substitutions: Map<string, TypeInfo>,
    selfType: TypeInfo,
  ): TypeInfo | null {
    if (typeParam === "Self") {
      return selfType;
    }
    if (substitutions.has(typeParam)) {
      return substitutions.get(typeParam) ?? unknownType;
    }
    return unknownType;
  }

  private annotateImplementationFailure(
    record: ImplementationRecord,
    obligation: ImplementationObligation,
    subject: TypeInfo,
    detail: string | undefined,
    actualArgs: string[],
    expectedArgs: string[],
  ): string {
    const label = this.appendInterfaceArgsToLabel(record.label, actualArgs);
    const contextSuffix = obligation.context ? ` (${obligation.context})` : "";
    const subjectLabel = subject && subject.kind !== "unknown" ? ` (got ${formatType(subject)})` : "";
    const expectedSuffix = expectedArgs.length ? ` expects ${expectedArgs.join(" ")}` : "";
    const detailSuffix = detail ? `: ${detail}` : "";
    return `${label}: constraint on ${obligation.typeParam}${contextSuffix} requires ${obligation.interfaceName}${expectedSuffix}${subjectLabel}${detailSuffix}`;
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

  private resolveInterfaceArgumentLabels(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
  ): string[] {
    if (!expr || expr.type !== "GenericTypeExpression") {
      return [];
    }
    return this.resolveInterfaceArgumentLabelsFromArray(expr.arguments ?? [], substitutions);
  }

  private resolveInterfaceArgumentLabelsFromArray(
    args: Array<AST.TypeExpression | null | undefined>,
    substitutions?: Map<string, TypeInfo>,
  ): string[] {
    if (!args || args.length === 0) {
      return [];
    }
    const stringSubs = substitutions ? this.buildStringSubstitutionMap(substitutions) : undefined;
    return args.map((arg) => (arg ? this.formatTypeExpression(arg, stringSubs) : "Unknown"));
  }

  private buildStringSubstitutionMap(substitutions: Map<string, TypeInfo>): Map<string, string> {
    const result = new Map<string, string>();
    substitutions.forEach((value, key) => {
      result.set(key, formatType(value));
    });
    return result;
  }

  private interfaceArgsCompatible(actual: string[], expected: string[]): boolean {
    if (actual.length !== expected.length) {
      return false;
    }
    for (let index = 0; index < expected.length; index += 1) {
      const exp = expected[index];
      const act = actual[index];
      if (exp === act) {
        continue;
      }
      if (exp === "Unknown" || act === "Unknown") {
        continue;
      }
      if (exp !== act) {
        return false;
      }
    }
    return true;
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
    for (const imp of module.imports) {
      if (!imp) continue;
      const packageName = this.formatImportPath(imp.packagePath);
      const summary = packageName ? this.packageSummaries.get(packageName) : undefined;
      if (!summary) {
        const label = packageName ?? "<unknown>";
        this.report(`typechecker: import references unknown package '${label}'`, imp);
      }
      if (imp.isWildcard) {
        if (summary?.symbols) {
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
          if (!summary?.symbols || !summary.symbols[selectorName]) {
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
      this.report(`typechecker: package '${packageName}' has no symbol '${memberName}'`, expression.member ?? expression);
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

  private buildPackageSummary(module: AST.Module): PackageSummary | null {
    const packageName = this.resolvePackageName(module);
    const symbols: Record<string, ExportedSymbolSummary> = {};
    const structs: Record<string, ExportedStructSummary> = {};
    const interfaces: Record<string, ExportedInterfaceSummary> = {};
    const functions: Record<string, ExportedFunctionSummary> = {};

    const statements = Array.isArray(module.body)
      ? (module.body as Array<AST.Statement | AST.Expression | null | undefined>)
      : [];
    for (const entry of statements) {
      if (!entry) continue;
      switch (entry.type) {
        case "StructDefinition": {
          if (entry.isPrivate) break;
          const name = entry.id?.name;
          if (!name) break;
          symbols[name] = { type: name };
          structs[name] = this.summarizeStructDefinition(entry);
          break;
        }
        case "InterfaceDefinition": {
          if (entry.isPrivate) break;
          const name = entry.id?.name;
          if (!name) break;
          symbols[name] = { type: name };
          interfaces[name] = this.summarizeInterfaceDefinition(entry);
          break;
        }
        case "FunctionDefinition": {
          if (entry.isPrivate || entry.isMethodShorthand) break;
          const name = entry.id?.name;
          if (!name) break;
          symbols[name] = { type: this.describeFunctionType(entry) };
          functions[name] = this.summarizeFunctionDefinition(entry);
          break;
        }
        default:
          break;
      }
    }

    const implementations: ExportedImplementationSummary[] = [];
    for (const record of this.implementationRecords) {
      if (record.definition?.isPrivate) {
        continue;
      }
      implementations.push(this.summarizeImplementationRecord(record));
    }

    const methodSets: ExportedMethodSetSummary[] = [];
    for (const record of this.methodSets) {
      methodSets.push(this.summarizeMethodSet(record));
    }

    return {
      name: packageName,
      symbols,
      structs,
      interfaces,
      functions,
      implementations,
      methodSets,
    };
  }

  private resolvePackageName(module: AST.Module): string {
    const path = module?.package?.namePath ?? [];
    const segments = path
      .map((segment) => segment?.name)
      .filter((segment): segment is string => Boolean(segment));
    if (segments.length > 0) {
      return segments.join(".");
    }
    return "<anonymous>";
  }

  private summarizeStructDefinition(definition: AST.StructDefinition): ExportedStructSummary {
    const summary: ExportedStructSummary = {
      typeParams: this.summarizeGenericParameters(definition.genericParams) ?? [],
      fields: {},
      positional: [],
      where: this.summarizeWhereClauses(definition.whereClause) ?? [],
    };

    if (Array.isArray(definition.fields)) {
      if (definition.kind === "named") {
        for (const field of definition.fields) {
          if (!field) continue;
          const fieldName = this.getIdentifierName(field.name);
          if (!fieldName) continue;
          summary.fields[fieldName] = this.formatTypeExpressionOrUnknown(field.fieldType);
        }
      } else if (definition.kind === "positional") {
        for (const field of definition.fields) {
          if (!field) continue;
          summary.positional.push(this.formatTypeExpressionOrUnknown(field.fieldType));
        }
      }
    }

    if (definition.kind !== "named") {
      summary.fields = {};
    }
    if (definition.kind !== "positional") {
      summary.positional = [];
    }
    return summary;
  }

  private summarizeInterfaceDefinition(definition: AST.InterfaceDefinition): ExportedInterfaceSummary {
    const methods: Record<string, ExportedFunctionSummary> = {};
    if (Array.isArray(definition.signatures)) {
      for (const signature of definition.signatures) {
        if (!signature?.name?.name) continue;
        methods[signature.name.name] = this.summarizeFunctionSignature(signature);
      }
    }
    return {
      typeParams: this.summarizeGenericParameters(definition.genericParams) ?? [],
      methods,
      where: this.summarizeWhereClauses(definition.whereClause) ?? [],
    };
  }

  private summarizeFunctionDefinition(definition: AST.FunctionDefinition): ExportedFunctionSummary {
    const info = definition.id?.name ? this.functionInfos.get(definition.id.name) : undefined;
    const obligations = info?.whereClause ?? [];
    return {
      parameters: this.summarizeParameters(definition.params),
      returnType: this.formatTypeExpressionOrUnknown(definition.returnType ?? null),
      typeParams: this.summarizeGenericParameters(definition.genericParams) ?? [],
      where: this.summarizeWhereClauses(definition.whereClause) ?? [],
      obligations: this.summarizeObligations(obligations, definition.id?.name) ?? [],
    };
  }

  private summarizeFunctionSignature(signature: AST.FunctionSignature): ExportedFunctionSummary {
    return {
      parameters: this.summarizeParameters(signature.params),
      returnType: this.formatTypeExpressionOrUnknown(signature.returnType ?? null),
      typeParams: this.summarizeGenericParameters(signature.genericParams) ?? [],
      where: this.summarizeWhereClauses(signature.whereClause) ?? [],
    };
  }

  private summarizeImplementationRecord(record: ImplementationRecord): ExportedImplementationSummary {
    const definition = record.definition;
    const interfaceArgs = this.summarizeInterfaceArgs(record.interfaceArgs);
    return {
      implName: definition.implName?.name ?? undefined,
      interface: record.interfaceName,
      target: this.formatTypeExpression(record.target),
      interfaceArgs: interfaceArgs ?? [],
      typeParams: this.summarizeGenericParameters(definition.genericParams) ?? [],
      methods: this.summarizeFunctionCollection(definition.definitions, { includeMethodShorthand: true }),
      where: this.summarizeWhereClauses(definition.whereClause) ?? [],
      obligations: this.summarizeObligations(record.obligations, record.label) ?? [],
    };
  }

  private summarizeMethodSet(record: MethodSetRecord): ExportedMethodSetSummary {
    return {
      typeParams: this.summarizeGenericParameters(record.definition.genericParams) ?? [],
      target: this.formatTypeExpression(record.target),
      methods: this.summarizeFunctionCollection(record.definition.definitions, { includeMethodShorthand: true }),
      where: this.summarizeWhereClauses(record.definition.whereClause) ?? [],
      obligations: this.summarizeObligations(record.obligations, record.label) ?? [],
    };
  }

  private summarizeFunctionCollection(
    definitions: Array<AST.FunctionDefinition | null | undefined> | undefined,
    options?: { includeMethodShorthand?: boolean },
  ): Record<string, ExportedFunctionSummary> {
    const methods: Record<string, ExportedFunctionSummary> = {};
    if (!Array.isArray(definitions)) {
      return methods;
    }
    for (const fn of definitions) {
      if (!fn || fn.isPrivate || !fn.id?.name) continue;
      if (!options?.includeMethodShorthand && fn.isMethodShorthand) continue;
      methods[fn.id.name] = this.summarizeFunctionDefinition(fn);
    }
    return methods;
  }

  private summarizeGenericParameters(
    params: Array<AST.GenericParameter | null | undefined> | undefined,
  ): ExportedGenericParamSummary[] | undefined {
    if (!Array.isArray(params) || params.length === 0) {
      return undefined;
    }
    const summaries: ExportedGenericParamSummary[] = [];
    for (const param of params) {
      if (!param) continue;
      const name = this.getIdentifierName(param.name);
      if (!name) continue;
      const constraints = this.summarizeInterfaceConstraints(param.constraints);
      summaries.push({ name, constraints });
    }
    return summaries.length ? summaries : undefined;
  }

  private summarizeInterfaceConstraints(
    constraints: Array<AST.InterfaceConstraint | null | undefined> | undefined,
  ): string[] | undefined {
    if (!Array.isArray(constraints) || constraints.length === 0) {
      return undefined;
    }
    const descriptions: string[] = [];
    for (const constraint of constraints) {
      if (!constraint?.interfaceType) continue;
      descriptions.push(this.formatTypeExpression(constraint.interfaceType));
    }
    return descriptions.length ? descriptions : undefined;
  }

  private summarizeWhereClauses(
    clauses: Array<AST.WhereClauseConstraint | null | undefined> | undefined,
  ): ExportedWhereConstraintSummary[] | undefined {
    if (!Array.isArray(clauses) || clauses.length === 0) {
      return undefined;
    }
    const summaries: ExportedWhereConstraintSummary[] = [];
    for (const clause of clauses) {
      if (!clause) continue;
      const typeParam = this.getIdentifierName(clause.typeParam);
      if (!typeParam) continue;
      const constraints = this.summarizeInterfaceConstraints(clause.constraints);
      summaries.push({ typeParam, constraints });
    }
    return summaries.length ? summaries : undefined;
  }

  private summarizeObligations(
    obligations: ImplementationObligation[] | undefined,
    owner?: string,
  ): ExportedObligationSummary[] | undefined {
    if (!obligations || obligations.length === 0) {
      return undefined;
    }
    return obligations.map((obligation) => ({
      owner,
      typeParam: obligation.typeParam,
      constraint: obligation.interfaceType
        ? this.formatTypeExpression(obligation.interfaceType)
        : obligation.interfaceName,
      subject: obligation.typeParam,
      context: obligation.context,
    }));
  }

  private summarizeInterfaceArgs(args: AST.TypeExpression[] | undefined): string[] | undefined {
    if (!Array.isArray(args) || args.length === 0) {
      return undefined;
    }
    const labels = args
      .filter((arg): arg is AST.TypeExpression => Boolean(arg))
      .map((arg) => this.formatTypeExpression(arg));
    return labels.length ? labels : undefined;
  }

  private summarizeParameters(params: Array<AST.FunctionParameter | null | undefined> | undefined): string[] {
    if (!Array.isArray(params) || params.length === 0) {
      return [];
    }
    return params.map((param) => this.formatTypeExpressionOrUnknown(param?.paramType ?? null));
  }

  private describeFunctionType(definition: AST.FunctionDefinition): string {
    const parameters = this.summarizeParameters(definition.params);
    const returnType = this.formatTypeExpressionOrUnknown(definition.returnType ?? null);
    return `fn(${parameters.join(", ")}) -> ${returnType}`;
  }

  private formatTypeExpressionOrUnknown(expr: AST.TypeExpression | null | undefined): string {
    if (!expr) {
      return "Unknown";
    }
    return this.formatTypeExpression(expr);
  }
}

export function createTypeChecker(options?: TypeCheckerOptions): TypeChecker {
  return new TypeChecker(options);
}

interface ImplementationObligation {
  typeParam: string;
  interfaceName: string;
  interfaceType?: AST.TypeExpression;
  context: string;
}

interface ImplementationRecord {
  interfaceName: string;
  label: string;
  target: AST.TypeExpression;
  targetKey: string;
  genericParams: string[];
  obligations: ImplementationObligation[];
  interfaceArgs: AST.TypeExpression[];
  definition: AST.ImplementationDefinition;
}

interface InterfaceCheckResult {
  ok: boolean;
  detail?: string;
}

interface MethodSetRecord {
  label: string;
  target: AST.TypeExpression;
  genericParams: string[];
  obligations: ImplementationObligation[];
  definition: AST.MethodsDefinition;
}

type FunctionContext = {
  structName?: string;
};

function extractLocation(node: AST.Node | null | undefined): DiagnosticLocation | undefined {
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
  if (location.path || location.line !== undefined || location.column !== undefined) {
    return location;
  }
  return undefined;
}

interface FunctionInfo {
  name: string;
  fullName: string;
  structName?: string;
  genericConstraints: Array<{
    paramName: string;
    interfaceName: string;
    interfaceDefined: boolean;
    interfaceType?: AST.TypeExpression;
  }>;
  genericParamNames: string[];
  whereClause: ImplementationObligation[];
  methodSetSubstitutions?: Array<[string, TypeInfo]>;
}
