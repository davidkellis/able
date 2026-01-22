import type * as AST from "../ast";
import { Environment } from "./environment";
import { RESERVED_TYPE_NAMES, TypeCheckerOptions } from "./checker/core";
import type { TypecheckResult } from "./diagnostics";
import { collectImplementationDefinition as collectImplementationDefinitionHelper, collectMethodsDefinition as collectMethodsDefinitionHelper } from "./checker/implementations";
import { resolvePackageName } from "./checker/summary";
import { ensureUniqueDeclaration as ensureUniqueDeclarationHelper } from "./checker/registry";
import { installBuiltins as installBuiltinsHelper } from "./checker/builtins";
import { TypeCheckerBase } from "./checker_base";
export type {
  ImplementationObligation,
  ImplementationRecord,
  MethodSetRecord,
  FunctionContext,
  FunctionInfo,
} from "./checker/types";
export type { TypeCheckerOptions, TypeCheckerPrelude } from "./checker/core";

export class TypeChecker extends TypeCheckerBase {
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
    this.typeParamStack = [];
    this.loopResultStack = [];
    this.breakpointStack = [];
    this.installBuiltins();
    this.installPrelude();
    this.packageAliases.clear();
    this.importedPackages = new Set();
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

  protected installBuiltins(): void {
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

  protected installPrelude(): void {
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

  protected collectModuleDeclarations(module: AST.Module): void {
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

  protected registerPrimaryTypeDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
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

  protected collectPrimaryDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
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

  protected collectImplementationDeclaration(node: AST.Statement | AST.Expression | undefined | null): void {
    if (!node) return;
    if (node.type === "ImplementationDefinition") {
      collectImplementationDefinitionHelper(this.implementationContext, node);
    }
  }

  protected checkStatement(node: AST.Statement | AST.Expression | undefined | null): void {
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

  protected ensureUniqueDeclaration(name: string | null | undefined, node: AST.Node | null | undefined): boolean {
    return ensureUniqueDeclarationHelper(this.registryContext(), name, node);
  }

  protected isKnownTypeName(name: string | null | undefined): boolean {
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
}

export function createTypeChecker(options?: TypeCheckerOptions): TypeChecker {
  return new TypeChecker(options);
}
