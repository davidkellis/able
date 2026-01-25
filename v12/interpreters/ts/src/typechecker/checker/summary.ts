import type * as AST from "../../ast";
import type {
  ExportedFunctionSummary,
  ExportedGenericParamSummary,
  ExportedImplementationSummary,
  ExportedInterfaceSummary,
  ExportedMethodSetSummary,
  ExportedObligationSummary,
  ExportedStructSummary,
  ExportedUnionSummary,
  ExportedSymbolSummary,
  ExportedWhereConstraintSummary,
  PackageSummary,
} from "../diagnostics";
import { unknownType, type TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import type { ImplementationObligation, ImplementationRecord, MethodSetRecord } from "./types";

export function buildPackageSummary(ctx: ImplementationContext, module: AST.Module): PackageSummary | null {
  const packageName = resolvePackageName(module);
  const visibility = resolvePackageVisibility(module);
  const symbols: Record<string, ExportedSymbolSummary> = {};
  const privateSymbols: Record<string, ExportedSymbolSummary> = {};
  const symbolTypes: Record<string, TypeInfo> = {};
  const structs: Record<string, ExportedStructSummary> = {};
  const unions: Record<string, ExportedUnionSummary> = {};
  const interfaces: Record<string, ExportedInterfaceSummary> = {};
  const functions: Record<string, ExportedFunctionSummary> = {};
  const implementationDefinitions = new Set<AST.ImplementationDefinition>();
  const methodSetDefinitions = new Set<AST.MethodsDefinition>();
  const methodSetRecordFor = (def: AST.MethodsDefinition): MethodSetRecord | undefined => {
    for (const record of ctx.getMethodSets()) {
      if (record.definition === def) {
        return record;
      }
    }
    return undefined;
  };

  const statements = Array.isArray(module.body)
    ? (module.body as Array<AST.Statement | AST.Expression | null | undefined>)
    : [];
  for (const entry of statements) {
    if (!entry) continue;
    switch (entry.type) {
      case "StructDefinition": {
        const name = entry.id?.name;
        if (!name) break;
        if (entry.isPrivate) {
          privateSymbols[name] = { type: name, visibility: "private" };
          break;
        }
        symbols[name] = { type: name, visibility: "public" };
        structs[name] = summarizeStructDefinition(ctx, entry);
        break;
      }
      case "InterfaceDefinition": {
        const name = entry.id?.name;
        if (!name) break;
        if (entry.isPrivate) {
          privateSymbols[name] = { type: name, visibility: "private" };
          break;
        }
        symbols[name] = { type: name, visibility: "public" };
        interfaces[name] = summarizeInterfaceDefinition(ctx, entry);
        break;
      }
      case "FunctionDefinition": {
        const name = entry.id?.name;
        if (!name) break;
        if (entry.isPrivate || entry.isMethodShorthand) {
          privateSymbols[name] = { type: describeFunctionType(ctx, entry), visibility: "private" };
          break;
        }
        symbols[name] = { type: describeFunctionType(ctx, entry), visibility: "public" };
        symbolTypes[name] = describeFunctionTypeInfo(ctx, entry, name);
        functions[name] = summarizeFunctionDefinition(ctx, entry);
        break;
      }
      case "ExternFunctionBody": {
        const def = entry.signature;
        const name = def?.id?.name;
        if (!name) break;
        if (def.isPrivate || def.isMethodShorthand) {
          privateSymbols[name] = { type: describeFunctionType(ctx, def), visibility: "private" };
          break;
        }
        symbols[name] = { type: describeFunctionType(ctx, def), visibility: "public" };
        symbolTypes[name] = describeFunctionTypeInfo(ctx, def, name);
        functions[name] = summarizeFunctionDefinition(ctx, def);
        break;
      }
      case "TypeAliasDefinition": {
        const name = entry.id?.name;
        if (!name) break;
        const label = `type alias -> ${formatTypeExpressionOrUnknown(ctx, entry.targetType)}`;
        if (entry.isPrivate) {
          privateSymbols[name] = { type: label, visibility: "private" };
          break;
        }
        symbols[name] = { type: label, visibility: "public" };
        break;
      }
      case "UnionDefinition": {
        const name = entry.id?.name;
        if (!name) break;
        if (entry.isPrivate) {
          privateSymbols[name] = { type: name, visibility: "private" };
        } else {
          symbols[name] = { type: name, visibility: "public" };
        }
        unions[name] = summarizeUnionDefinition(ctx, entry);
        break;
      }
      case "ImplementationDefinition": {
        implementationDefinitions.add(entry);
        if (Array.isArray(entry.definitions)) {
          for (const def of entry.definitions) {
            if (!def?.id?.name) continue;
            const fnName = def.id.name;
            if (def.isPrivate) {
              if (!symbols[fnName]) {
                privateSymbols[fnName] = { type: describeFunctionType(ctx, def), visibility: "private" };
              }
              continue;
            }
            symbols[fnName] = { type: describeFunctionType(ctx, def), visibility: "public" };
            symbolTypes[fnName] = describeFunctionTypeInfo(ctx, def, fnName);
            functions[fnName] = summarizeFunctionDefinition(ctx, def);
            delete privateSymbols[fnName];
          }
        }
        break;
      }
      case "MethodsDefinition": {
        methodSetDefinitions.add(entry);
        const record = methodSetRecordFor(entry);
        const targetBaseName =
          ctx.getIdentifierNameFromTypeExpression?.(record?.target ?? entry.targetType) ??
          ctx.getIdentifierNameFromTypeExpression?.(entry.targetType);
        if (Array.isArray(entry.definitions)) {
          for (const def of entry.definitions) {
            if (!def?.id?.name) continue;
            const fnName = def.id.name;
            const firstParam = Array.isArray(def.params) ? def.params[0] : undefined;
            const isSelfMethod =
              Boolean(def.isMethodShorthand) ||
              (firstParam?.name?.type === "Identifier" && firstParam.name.name?.toLowerCase() === "self") ||
              (firstParam?.paramType?.type === "SimpleTypeExpression" &&
                ctx.getIdentifierName(firstParam.paramType.name) === "Self");
            const exportName = isSelfMethod || !targetBaseName ? fnName : `${targetBaseName}.${fnName}`;
            if (def.isPrivate) {
              if (exportName && !symbols[exportName]) {
                privateSymbols[exportName] = { type: describeFunctionType(ctx, def), visibility: "private" };
              }
              continue;
            }
            if (!exportName) continue;
            symbols[exportName] = { type: describeFunctionType(ctx, def), visibility: "public" };
            symbolTypes[exportName] = describeFunctionTypeInfo(ctx, def, exportName);
            functions[exportName] = summarizeFunctionDefinition(ctx, def);
            delete privateSymbols[exportName];
          }
        }
        break;
      }
      default:
        break;
    }
  }

  const implementations: ExportedImplementationSummary[] = [];
  for (const record of ctx.getImplementationRecords()) {
    if (record.definition?.isPrivate) {
      continue;
    }
    if (!record.definition || !implementationDefinitions.has(record.definition)) {
      continue;
    }
    implementations.push(summarizeImplementationRecord(ctx, record));
  }

  const methodSets: ExportedMethodSetSummary[] = [];
  for (const record of ctx.getMethodSets()) {
    if (!methodSetDefinitions.has(record.definition)) {
      continue;
    }
    methodSets.push(summarizeMethodSet(ctx, record));
  }

  return {
    name: packageName,
    visibility,
    symbols,
    privateSymbols,
    structs,
    unions,
    interfaces,
    functions,
    symbolTypes,
    implementations,
    methodSets,
  };
}

export function resolvePackageName(module: AST.Module): string {
  const path = module?.package?.namePath ?? [];
  const segments = path
    .map((segment) => segment?.name)
    .filter((segment): segment is string => Boolean(segment));
  if (segments.length > 0) {
    return segments.join(".");
  }
  return "<anonymous>";
}

export function resolvePackageVisibility(module: AST.Module): "public" | "private" {
  if (module?.package?.isPrivate) {
    return "private";
  }
  return "public";
}

function summarizeStructDefinition(ctx: ImplementationContext, definition: AST.StructDefinition): ExportedStructSummary {
  const summary: ExportedStructSummary = {
    typeParams: summarizeGenericParameters(ctx, definition.genericParams) ?? [],
    fields: {},
    positional: [],
    where: summarizeWhereClauses(ctx, definition.whereClause) ?? [],
  };

  if (Array.isArray(definition.fields)) {
    if (definition.kind === "named") {
      for (const field of definition.fields) {
        if (!field) continue;
        const fieldName = ctx.getIdentifierName(field.name);
        if (!fieldName) continue;
        summary.fields[fieldName] = formatTypeExpressionOrUnknown(ctx, field.fieldType);
      }
    } else if (definition.kind === "positional") {
      for (const field of definition.fields) {
        if (!field) continue;
        summary.positional.push(formatTypeExpressionOrUnknown(ctx, field.fieldType));
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

function summarizeUnionDefinition(ctx: ImplementationContext, definition: AST.UnionDefinition): ExportedUnionSummary {
  const variants: string[] = [];
  if (Array.isArray(definition.variants)) {
    for (const variant of definition.variants) {
      variants.push(formatTypeExpressionOrUnknown(ctx, variant));
    }
  }
  return {
    typeParams: summarizeGenericParameters(ctx, definition.genericParams) ?? [],
    variants,
    where: summarizeWhereClauses(ctx, definition.whereClause) ?? [],
  };
}

function summarizeInterfaceDefinition(
  ctx: ImplementationContext,
  definition: AST.InterfaceDefinition,
): ExportedInterfaceSummary {
  const methods: Record<string, ExportedFunctionSummary> = {};
  if (Array.isArray(definition.signatures)) {
    for (const signature of definition.signatures) {
      if (!signature?.name?.name) continue;
      methods[signature.name.name] = summarizeFunctionSignature(ctx, signature);
    }
  }
  return {
    typeParams: summarizeGenericParameters(ctx, definition.genericParams) ?? [],
    methods,
    where: summarizeWhereClauses(ctx, definition.whereClause) ?? [],
  };
}

function summarizeFunctionDefinition(
  ctx: ImplementationContext,
  definition: AST.FunctionDefinition,
): ExportedFunctionSummary {
  const info = definition.id?.name ? ctx.getFunctionInfos(definition.id.name)?.[0] : undefined;
  const obligations = info?.whereClause ?? [];
  return {
    parameters: summarizeParameters(ctx, definition.params),
    returnType: formatTypeExpressionOrUnknown(ctx, definition.returnType ?? null),
    typeParams: summarizeGenericParameters(ctx, definition.genericParams) ?? [],
    where: summarizeWhereClauses(ctx, definition.whereClause) ?? [],
    obligations: summarizeObligations(ctx, obligations, info?.fullName) ?? [],
  };
}

function summarizeImplementationRecord(
  ctx: ImplementationContext,
  record: ImplementationRecord,
): ExportedImplementationSummary {
  return {
    implName: record.definition.id?.name,
    interface: record.interfaceName,
    target: ctx.formatTypeExpression(record.target),
    interfaceArgs: summarizeInterfaceArgs(ctx, record.interfaceArgs) ?? [],
    typeParams: summarizeGenericParameters(ctx, record.definition.genericParams) ?? [],
    methods: summarizeFunctionCollection(ctx, record.definition.definitions, { includeMethodShorthand: true }),
    where: summarizeWhereClauses(ctx, record.definition.whereClause) ?? [],
    obligations: summarizeObligations(ctx, record.obligations, record.label) ?? [],
  };
}

function summarizeMethodSet(ctx: ImplementationContext, record: MethodSetRecord): ExportedMethodSetSummary {
  const typeQualifier = ctx.getIdentifierNameFromTypeExpression?.(record.target);
  return {
    typeParams: summarizeGenericParameters(ctx, record.definition.genericParams) ?? [],
    target: ctx.formatTypeExpression(record.target),
    methods: summarizeFunctionCollection(ctx, record.definition.definitions, {
      includeMethodShorthand: true,
      typeQualifier,
    }),
    where: summarizeWhereClauses(ctx, record.definition.whereClause) ?? [],
    obligations: summarizeObligations(ctx, record.obligations, record.label) ?? [],
  };
}

function summarizeFunctionCollection(
  ctx: ImplementationContext,
  definitions: Array<AST.FunctionDefinition | null | undefined> | undefined,
  options?: { includeMethodShorthand?: boolean; typeQualifier?: string | null },
): Record<string, ExportedFunctionSummary> {
  const methods: Record<string, ExportedFunctionSummary> = {};
  if (!Array.isArray(definitions)) {
    return methods;
  }
  for (const fn of definitions) {
    if (!fn || fn.isPrivate || !fn.id?.name) continue;
    if (!options?.includeMethodShorthand && fn.isMethodShorthand) continue;
    const firstParam = Array.isArray(fn.params) ? fn.params[0] : undefined;
    const firstParamTypeName =
      firstParam?.paramType?.type === "SimpleTypeExpression"
        ? ctx.getIdentifierName(firstParam.paramType.name)
        : null;
    const firstParamBaseName =
      firstParam?.paramType?.type === "GenericTypeExpression"
        ? ctx.getIdentifierNameFromTypeExpression(firstParam.paramType.base)
        : firstParamTypeName;
    const expectsSelf =
      Boolean(fn.isMethodShorthand) ||
      (firstParam?.name?.type === "Identifier" && firstParam.name.name?.toLowerCase() === "self") ||
      (firstParam?.paramType?.type === "SimpleTypeExpression" &&
        ctx.getIdentifierName(firstParam.paramType.name) === "Self") ||
      (Boolean(options?.typeQualifier) &&
        Boolean(firstParamBaseName) &&
        firstParamBaseName === options?.typeQualifier);
    const exportName = !options?.typeQualifier || expectsSelf ? fn.id.name : `${options.typeQualifier}.${fn.id.name}`;
    methods[exportName] = summarizeFunctionDefinition(ctx, fn);
  }
  return methods;
}

function summarizeGenericParameters(
  ctx: ImplementationContext,
  params: Array<AST.GenericParameter | null | undefined> | undefined,
): ExportedGenericParamSummary[] | undefined {
  if (!Array.isArray(params) || params.length === 0) {
    return undefined;
  }
  const summaries: ExportedGenericParamSummary[] = [];
  for (const param of params) {
    if (!param) continue;
    const name = ctx.getIdentifierName(param.name);
    if (!name) continue;
    const constraints = summarizeInterfaceConstraints(ctx, param.constraints);
    summaries.push({ name, constraints });
  }
  return summaries.length ? summaries : undefined;
}

function summarizeInterfaceConstraints(
  ctx: ImplementationContext,
  constraints: Array<AST.InterfaceConstraint | null | undefined> | undefined,
): string[] | undefined {
  if (!Array.isArray(constraints) || constraints.length === 0) {
    return undefined;
  }
  const descriptions: string[] = [];
  for (const constraint of constraints) {
    if (!constraint?.interfaceType) continue;
    descriptions.push(ctx.formatTypeExpression(constraint.interfaceType));
  }
  return descriptions.length ? descriptions : undefined;
}

function summarizeWhereClauses(
  ctx: ImplementationContext,
  clauses: Array<AST.WhereClauseConstraint | null | undefined> | undefined,
): ExportedWhereConstraintSummary[] | undefined {
  if (!Array.isArray(clauses) || clauses.length === 0) {
    return undefined;
  }
  const summaries: ExportedWhereConstraintSummary[] = [];
  for (const clause of clauses) {
    if (!clause) continue;
    const typeParam = clause.typeParam ? ctx.formatTypeExpression(clause.typeParam) : null;
    if (!typeParam) continue;
    const constraints = summarizeInterfaceConstraints(ctx, clause.constraints);
    summaries.push({ typeParam, constraints });
  }
  return summaries.length ? summaries : undefined;
}

function summarizeObligations(
  ctx: ImplementationContext,
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
      ? ctx.formatTypeExpression(obligation.interfaceType)
      : obligation.interfaceName,
    subject: obligation.subjectExpr
      ? ctx.formatTypeExpression(obligation.subjectExpr)
      : obligation.typeParam,
    context: obligation.context,
  }));
}

function summarizeInterfaceArgs(
  ctx: ImplementationContext,
  args: AST.TypeExpression[] | undefined,
): string[] | undefined {
  if (!Array.isArray(args) || args.length === 0) {
    return undefined;
  }
  const labels = args
    .filter((arg): arg is AST.TypeExpression => Boolean(arg))
    .map((arg) => ctx.formatTypeExpression(arg));
  return labels.length ? labels : undefined;
}

function summarizeParameters(
  ctx: ImplementationContext,
  params: Array<AST.FunctionParameter | null | undefined> | undefined,
): string[] {
  if (!Array.isArray(params) || params.length === 0) {
    return [];
  }
  return params.map((param) => formatTypeExpressionOrUnknown(ctx, param?.paramType ?? null));
}

function describeFunctionType(ctx: ImplementationContext, definition: AST.FunctionDefinition): string {
  const parameters = summarizeParameters(ctx, definition.params);
  const returnType = formatTypeExpressionOrUnknown(ctx, definition.returnType ?? null);
  return `fn(${parameters.join(", ")}) -> ${returnType}`;
}

function describeFunctionTypeInfo(
  ctx: ImplementationContext,
  definition: AST.FunctionDefinition,
  exportedName: string,
): TypeInfo {
  const name = definition.id?.name ?? exportedName;
  if (name) {
    const infos = ctx.getFunctionInfos(name) ?? [];
    const match = infos.find((info) => info.definition === definition);
    if (match) {
      return {
        kind: "function",
        parameters: match.parameters ?? [],
        returnType: match.returnType ?? unknownType,
      };
    }
  }
  const params = Array.isArray(definition.params)
    ? definition.params.map((param) => ctx.resolveTypeExpression(param?.paramType))
    : [];
  const returnType = definition.returnType ? ctx.resolveTypeExpression(definition.returnType) : unknownType;
  return { kind: "function", parameters: params, returnType };
}

function summarizeFunctionSignature(
  ctx: ImplementationContext,
  signature: AST.FunctionSignature,
): ExportedFunctionSummary {
  return {
    parameters: summarizeParameters(ctx, signature.params),
    returnType: formatTypeExpressionOrUnknown(ctx, signature.returnType ?? null),
    typeParams: summarizeGenericParameters(ctx, signature.genericParams) ?? [],
    where: summarizeWhereClauses(ctx, signature.whereClause) ?? [],
    obligations: [],
  };
}

function formatTypeExpressionOrUnknown(
  ctx: ImplementationContext,
  expr: AST.TypeExpression | null | undefined,
): string {
  if (!expr) {
    return "Unknown";
  }
  return ctx.formatTypeExpression(expr);
}
