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
import type { TypecheckerDiagnostic, TypecheckResult } from "./diagnostics";

export interface TypeCheckerOptions {
  /**
   * When true, the checker will attempt to continue after diagnostics instead of
   * aborting immediately. The checker currently always continues.
   */
  continueAfterDiagnostics?: boolean;
}

export class TypeChecker {
  private env: Environment;
  private readonly options: TypeCheckerOptions;
  private diagnostics: TypecheckerDiagnostic[] = [];
  private structDefinitions: Map<string, AST.StructDefinition> = new Map();
  private interfaceDefinitions: Map<string, AST.InterfaceDefinition> = new Map();
  private functionInfos: Map<string, FunctionInfo> = new Map();
  private implementations: Map<string, Set<string>> = new Map();

  constructor(options: TypeCheckerOptions = {}) {
    this.env = new Environment();
    this.options = options;
  }

  checkModule(module: AST.Module): TypecheckResult {
    this.env = new Environment();
    this.diagnostics = [];
    this.structDefinitions = new Map();
    this.interfaceDefinitions = new Map();
    this.functionInfos = new Map();
    this.implementations = new Map();
    this.installBuiltins();
    this.collectModuleDeclarations(module);

    if (Array.isArray(module.body)) {
      for (const statement of module.body) {
        this.checkStatement(statement as AST.Statement | AST.Expression);
      }
    }

    return { diagnostics: [...this.diagnostics] };
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
    const structLabel = this.formatImplementationTarget(definition.targetType) ?? this.getIdentifierNameFromTypeExpression(definition.targetType);
    if (!structLabel) return;
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
      let entries = this.implementations.get(targetKey);
      if (!entries) {
        entries = new Set<string>();
        this.implementations.set(targetKey, entries);
      }
      entries.add(interfaceName);
    }

    if (Array.isArray(definition.definitions)) {
      for (const entry of definition.definitions) {
        if (entry?.type === "FunctionDefinition") {
          this.collectFunctionDefinition(entry, { structName: contextName });
        }
      }
    }
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
            this.report(message);
          }
          info.genericConstraints.push({ paramName, interfaceName, interfaceDefined });
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
          this.report(`typechecker: ${label} defines duplicate method '${methodName}'`);
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
        this.report(`typechecker: ${label} missing method '${methodName}'`);
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
      this.report(
        `typechecker: ${label} defines method '${methodName}' not declared in interface ${interfaceName}`,
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
      this.report(`typechecker: impl ${interfaceName} does not accept type arguments`);
      return;
    }
    if (expected > 0) {
      const targetDescription = targetLabel;
      if (provided === 0) {
        this.report(
          `typechecker: impl ${interfaceName} for ${targetDescription} requires ${expected} interface type argument(s)`,
        );
        return;
      }
      if (provided !== expected) {
        this.report(
          `typechecker: impl ${interfaceName} for ${targetDescription} expected ${expected} interface type argument(s), got ${provided}`,
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
      );
      valid = false;
    }

    const interfaceParams = Array.isArray(signature.params) ? signature.params : [];
    const implementationParams = Array.isArray(implementation.params) ? implementation.params : [];
    if (interfaceParams.length !== implementationParams.length) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceParams.length} parameter(s), got ${implementationParams.length}`,
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
      );
      valid = false;
    }

    // Basic where-clause compatibility: ensure implementation does not omit required where clauses.
    const interfaceWhere = Array.isArray(signature.whereClause) ? signature.whereClause.length : 0;
    const implementationWhere = Array.isArray(implementation.whereClause) ? implementation.whereClause.length : 0;
    if (interfaceWhere !== implementationWhere) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' expects ${interfaceWhere} where-clause constraint(s), got ${implementationWhere}`,
      );
      valid = false;
    }

    // Placeholder: ensure method privacy matches interface expectations (interfaces methods are always public).
    if (implementation.isPrivate) {
      this.report(
        `typechecker: ${label} method '${signature.name?.name ?? "<anonymous>"}' must be public to satisfy interface`,
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
        this.report(`typechecker: struct pattern field '${fieldName}' not found`);
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
            this.report(`typechecker: '${expression.operator}' left operand must be bool (got ${describe(left)})`);
          }
          if (!isBoolean(right)) {
            this.report(`typechecker: '${expression.operator}' right operand must be bool (got ${describe(right)})`);
          }
          return primitiveType("bool");
        }
        return unknownType;
      }
      case "RangeExpression": {
        const start = this.inferExpression(expression.start);
        if (!isNumeric(start)) {
          this.report("typechecker: range start must be numeric");
        }
        const end = this.inferExpression(expression.end);
        if (!isNumeric(end)) {
          this.report("typechecker: range end must be numeric");
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
      const memberName = this.getIdentifierName(callee.member);
      if (!memberName) return [];
      const objectType = this.inferExpression(callee.object);
      if (objectType.kind === "struct") {
        const structLabel = formatType(objectType);
        const info = this.functionInfos.get(`${structLabel}::${memberName}`) ?? this.functionInfos.get(memberName);
        return info ? [info] : [];
      }
    }
    return [];
  }

  private enforceFunctionConstraints(info: FunctionInfo, call: AST.FunctionCall): void {
    if (!info.genericConstraints.length) return;
    const typeArgs = Array.isArray(call.typeArguments) ? call.typeArguments : [];
    info.genericConstraints.forEach((constraint, index) => {
      const typeArgExpr = typeArgs[index];
      const typeArg = this.resolveTypeExpression(typeArgExpr);
      if (!constraint.interfaceDefined) {
        const message = info.structName
          ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${constraint.paramName} references unknown interface '${constraint.interfaceName}'`
          : `typechecker: fn ${info.name} constraint on ${constraint.paramName} references unknown interface '${constraint.interfaceName}'`;
        this.report(message);
        return;
      }
      if (!this.typeImplementsInterface(typeArg, constraint.interfaceName)) {
        const typeName = this.describeTypeArgument(typeArg);
        const message = info.structName
          ? `typechecker: methods for ${info.structName}::${info.name} constraint on ${constraint.paramName} is not satisfied: ${typeName} does not implement ${constraint.interfaceName}`
          : `typechecker: fn ${info.name} constraint on ${constraint.paramName} is not satisfied: ${typeName} does not implement ${constraint.interfaceName}`;
        this.report(message);
      }
    });
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

  private typeImplementsInterface(type: TypeInfo, interfaceName: string): boolean {
    if (type.kind === "unknown") {
      return true;
    }
    if (type.kind === "nullable") {
      if (this.typeHasImplementation(type, interfaceName)) {
        return true;
      }
      return this.typeImplementsInterface(type.inner, interfaceName);
    }
    if (type.kind === "result") {
      if (this.typeHasImplementation(type, interfaceName)) {
        return true;
      }
      return this.typeImplementsInterface(type.inner, interfaceName);
    }
    if (type.kind === "union") {
      return type.members.every((member) => this.typeImplementsInterface(member, interfaceName));
    }
    if (type.kind === "interface" && type.name === interfaceName) {
      return true;
    }
    return this.typeHasImplementation(type, interfaceName);
  }

  private typeHasImplementation(type: TypeInfo, interfaceName: string): boolean {
    const key = this.typeKeyFromTypeInfo(type);
    if (!key) {
      return false;
    }
    const implementations = this.implementations.get(key);
    if (!implementations) {
      return false;
    }
    return implementations.has(interfaceName);
  }

  private typeKeyFromTypeInfo(type: TypeInfo): string | null {
    if (!type || type.kind === "unknown") {
      return null;
    }
    return formatType(type);
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

  private getInterfaceNameFromConstraint(constraint: AST.InterfaceConstraint | null | undefined): string | null {
    if (!constraint) return null;
    return this.getIdentifierNameFromTypeExpression(constraint.interfaceType);
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

  private report(message: string): void {
    this.diagnostics.push({ severity: "error", message });
  }
}

export function createTypeChecker(options?: TypeCheckerOptions): TypeChecker {
  return new TypeChecker(options);
}

type FunctionContext = {
  structName?: string;
};

interface FunctionInfo {
  name: string;
  fullName: string;
  structName?: string;
  genericConstraints: Array<{
    paramName: string;
    interfaceName: string;
    interfaceDefined: boolean;
  }>;
}
