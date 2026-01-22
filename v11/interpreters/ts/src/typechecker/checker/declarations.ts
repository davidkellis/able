import * as AST from "../../ast";
import { unknownType } from "../types";
import type { TypeInfo } from "../types";
import type { FunctionContext, FunctionInfo, ImplementationObligation } from "./types";
import { formatNodeOrigin } from "./registry";
import type { StatementContext } from "./expression-context";

export interface DeclarationsContext extends StatementContext {
  getCurrentPackageName?: () => string;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  getIdentifierNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null;
  getInterfaceNameFromConstraint(constraint: AST.GenericConstraint | null | undefined): string | null;
  getInterfaceNameFromTypeExpression(expr: AST.TypeExpression | null | undefined): string | null;
  resolveTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, TypeInfo>): TypeInfo;
  describeTypeExpression(expr: AST.TypeExpression | null | undefined, substitutions?: Map<string, string>): string | null;
  report(message: string, node?: AST.Node | null | undefined): void;
  defineValue(name: string, type: TypeInfo): void;
  getInterfaceDefinition(name: string): AST.InterfaceDefinition | undefined;
  hasInterfaceDefinition(name: string): boolean;
  addFunctionInfo(key: string, info: FunctionInfo): void;
  getFunctionInfos(key: string): FunctionInfo[];
  isKnownTypeName(name: string): boolean;
  hasTypeDefinition(name: string): boolean;
}

export function collectFunctionDefinition(
  ctx: DeclarationsContext,
  definition: AST.FunctionDefinition,
  scope: FunctionContext | undefined,
): void {
  inferFunctionGenerics(ctx, definition, scope);
  const currentPackageName = ctx.getCurrentPackageName?.();
  const name = definition.id?.name ?? "<anonymous>";
  const structName = scope?.structName;
  const structBaseName = scope?.structBaseName ?? structName;
  const fullName = structName ? `${structName}::${name}` : name;
  let substitutions = buildGenericSubstitutions(ctx, definition, scope);
  const returnType = ctx.resolveTypeExpression(definition.returnType, substitutions);
  const selfType =
    structName !== undefined
      ? structTypeFromScope(structName, scope?.typeParamNames)
      : unknownType;
  if (structName) {
    substitutions = substitutions ?? new Map<string, TypeInfo>();
    substitutions.set("Self", selfType);
  }
  const firstParam = Array.isArray(definition.params) ? definition.params[0] : undefined;
  const injectedSelf = injectImplicitSelfParameter(definition, scope);
  const firstParamTypeName =
    firstParam?.paramType?.type === "SimpleTypeExpression"
      ? ctx.getIdentifierName(firstParam.paramType.name)
      : null;
  const firstParamBaseName =
    firstParam?.paramType?.type === "GenericTypeExpression"
      ? ctx.getIdentifierNameFromTypeExpression(firstParam.paramType.base)
      : firstParamTypeName;
  const paramNameIsSelf = firstParam?.name?.type === "Identifier" && firstParam.name.name?.toLowerCase() === "self";
  const matchesSelfParam =
    Boolean(structName) &&
    (paramNameIsSelf ||
      (Boolean(firstParamBaseName) &&
        (firstParamBaseName === "Self" || firstParamBaseName === structBaseName || firstParamBaseName === structName)));
  const expectsSelfParam =
    Boolean(definition.isMethodShorthand) ||
    injectedSelf ||
    matchesSelfParam;
  let parameterTypes = resolveFunctionParameterTypes(ctx, definition, substitutions);
  if (expectsSelfParam && structName) {
    if (parameterTypes.length > 0) {
      parameterTypes = [selfType, ...parameterTypes.slice(1)];
    } else {
      parameterTypes = [selfType];
    }
  }
  const isTypeQualified = Boolean(structName) && !expectsSelfParam;
  const typeQualifier = isTypeQualified ? structBaseName ?? structName : undefined;
  const exportedName = isTypeQualified && typeQualifier ? `${typeQualifier}.${name}` : undefined;

  const info: FunctionInfo = {
    name,
    fullName,
    packageName: currentPackageName,
    definition,
    structName,
    hasImplicitSelf: expectsSelfParam,
    isTypeQualified,
    typeQualifier,
    exportedName,
    parameters: parameterTypes,
    genericConstraints: [],
    whereClause: extractFunctionWhereObligations(ctx, definition),
    genericParamNames: Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((paramName): paramName is string => Boolean(paramName))
      : [],
    returnType,
  };
  if (scope?.fromMethodSet) {
    info.fromMethodSet = true;
  }

  if (!structName && definition.id?.name) {
    const existing = ctx.getFunctionInfos(definition.id.name);
    const matches = existing.find(
      (entry) =>
        !entry.structName &&
        (!entry.packageName || !currentPackageName || entry.packageName === currentPackageName) &&
        functionParamsEquivalent(ctx, entry.parameters, info.parameters),
    );
    if (matches?.definition) {
      const location = formatNodeOrigin(matches.definition);
      const displayName = name.startsWith("<anonymous>::") ? name.slice("<anonymous>::".length) : name;
      ctx.report(`typechecker: duplicate declaration '${displayName}' (previous declaration at ${location})`, definition);
      return;
    }
  }

  if (Array.isArray(definition.genericParams)) {
    for (const param of definition.genericParams) {
      const paramName = param.name?.name ?? "T";
      if (!Array.isArray(param.constraints)) continue;
      for (const constraint of param.constraints) {
        const interfaceName = ctx.getInterfaceNameFromConstraint(constraint);
        if (!interfaceName) continue;
        const interfaceDefined = ctx.hasInterfaceDefinition(interfaceName);
        if (!interfaceDefined) {
          const message = structName
            ? `typechecker: methods for ${structName}::${name} constraint on ${paramName} references unknown interface '${interfaceName}'`
            : `typechecker: fn ${name} constraint on ${paramName} references unknown interface '${interfaceName}'`;
          ctx.report(message, constraint?.interfaceType ?? constraint ?? definition);
        }
        info.genericConstraints.push({
          paramName,
          interfaceName,
          interfaceDefined,
          interfaceType: constraint.interfaceType,
        });
      }
    }
  }

  ctx.addFunctionInfo(fullName, info);
  if (definition.id?.name) {
    if (isTypeQualified) {
      if (exportedName) {
        ctx.addFunctionInfo(exportedName, info);
        ctx.defineValue(exportedName, {
          kind: "function",
          parameters: parameterTypes,
          returnType,
        });
      }
    } else {
      ctx.addFunctionInfo(name, info);
      ctx.defineValue(definition.id.name, {
        kind: "function",
        parameters: parameterTypes,
        returnType,
      });
    }
  }
}

function functionParamsEquivalent(
  ctx: DeclarationsContext,
  leftParams: TypeInfo[] | undefined,
  rightParams: TypeInfo[] | undefined,
): boolean {
  const left = Array.isArray(leftParams) ? leftParams : [];
  const right = Array.isArray(rightParams) ? rightParams : [];
  if (left.length !== right.length) {
    return false;
  }
  for (let index = 0; index < left.length; index += 1) {
    if (!signatureTypesEquivalent(ctx, left[index], right[index])) {
      return false;
    }
  }
  return true;
}

function signatureTypesEquivalent(
  ctx: DeclarationsContext,
  left: TypeInfo | undefined,
  right: TypeInfo | undefined,
): boolean {
  if (!left || left.kind === "unknown" || !right || right.kind === "unknown") {
    return true;
  }
  if (left.kind === "type_parameter" || right.kind === "type_parameter") {
    return left.kind === "type_parameter" && right.kind === "type_parameter";
  }
  return ctx.typeInfosEquivalent(left, right);
}

function resolveFunctionParameterTypes(
  ctx: DeclarationsContext,
  definition: AST.FunctionDefinition,
  substitutions?: Map<string, TypeInfo>,
): TypeInfo[] {
  if (!Array.isArray(definition.params)) {
    return [];
  }
  return definition.params.map((param) => ctx.resolveTypeExpression(param?.paramType, substitutions));
}

function extractFunctionWhereObligations(
  ctx: DeclarationsContext,
  definition: AST.FunctionDefinition,
): ImplementationObligation[] {
  const obligations: ImplementationObligation[] = [];
  const appendObligation = (
    typeParam: string | null,
    interfaceType: AST.TypeExpression | null | undefined,
    context: string,
    subjectExpr?: AST.TypeExpression,
  ) => {
    const interfaceName = ctx.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context, subjectExpr });
  };

  if (Array.isArray(definition.genericParams)) {
    for (const param of definition.genericParams) {
      const paramName = ctx.getIdentifierName(param?.name);
      if (!paramName || !Array.isArray(param?.constraints)) continue;
      for (const constraint of param.constraints) {
        appendObligation(paramName, constraint?.interfaceType, "generic constraint");
      }
    }
  }

  if (Array.isArray(definition.whereClause)) {
    for (const clause of definition.whereClause) {
      const subjectExpr = clause?.typeParam;
      const typeParamName = subjectExpr ? ctx.describeTypeExpression(subjectExpr) : null;
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause", subjectExpr);
      }
    }
  }

  return obligations;
}

function injectImplicitSelfParameter(definition: AST.FunctionDefinition, scope: FunctionContext | undefined): boolean {
  if (!scope?.structName || !Array.isArray(definition.params) || definition.params.length === 0) {
    return false;
  }
  const firstParam = definition.params[0];
  if (!firstParam || firstParam.name?.type !== "Identifier") {
    return false;
  }
  const paramName = firstParam.name.name?.toLowerCase();
  if (paramName !== "self") {
    return false;
  }
  if (!firstParam.paramType) {
    firstParam.paramType = AST.simpleTypeExpression("Self");
  }
  return firstParam.paramType?.type === "SimpleTypeExpression" && firstParam.paramType.name?.name === "Self";
}

function buildGenericSubstitutions(
  ctx: DeclarationsContext,
  definition: AST.FunctionDefinition,
  scope: FunctionContext | undefined,
): Map<string, TypeInfo> | undefined {
  const paramNames = new Set<string>();
  if (Array.isArray(scope?.typeParamNames)) {
    scope.typeParamNames.forEach((name) => {
      if (name) {
        paramNames.add(name);
      }
    });
  }
  if (Array.isArray(definition.genericParams)) {
    for (const param of definition.genericParams) {
      const name = ctx.getIdentifierName(param?.name);
      if (name) {
        paramNames.add(name);
      }
    }
  }
  if (paramNames.size === 0) {
    return undefined;
  }
  const substitutions = new Map<string, TypeInfo>();
  for (const name of paramNames) {
    substitutions.set(name, unknownType);
  }
  return substitutions;
}

type FunctionLikeNode = Pick<
  AST.FunctionDefinition,
  "genericParams" | "inferredGenericParams" | "params" | "returnType" | "whereClause"
>;

type TypeOccurrence = {
  name: string;
  node?: AST.Node | null;
  kind: "type" | "where";
};

export function inferFunctionSignatureGenerics(
  ctx: DeclarationsContext,
  signature: AST.FunctionSignature,
  parentTypeParams: string[] = [],
): void {
  inferGenericsForNode(ctx, signature, parentTypeParams);
}

function inferFunctionGenerics(
  ctx: DeclarationsContext,
  definition: AST.FunctionDefinition,
  scope: FunctionContext | undefined,
): void {
  const parentTypeParams = scope?.typeParamNames ?? [];
  inferGenericsForNode(ctx, definition, parentTypeParams);
}

function inferGenericsForNode(
  ctx: DeclarationsContext,
  node: FunctionLikeNode,
  parentTypeParams: string[],
): void {
  const occurrences = collectFunctionLikeOccurrences(node);
  if (occurrences.length === 0) {
    return;
  }
  const known = new Set<string>();
  parentTypeParams.forEach((name) => {
    if (name) {
      known.add(name);
    }
  });
  if (Array.isArray(node.genericParams)) {
    for (const param of node.genericParams) {
      const name = ctx.getIdentifierName(param?.name);
      if (name) {
        known.add(name);
      }
    }
  }
  const inferred: AST.GenericParameter[] = [];
  const inferredMap = new Map<string, AST.GenericParameter>();
  for (const occurrence of occurrences) {
    const decision = classifyInferenceCandidate(ctx, occurrence.name, known);
    if (decision === "known-type") {
      continue;
    }
    if (decision !== "infer") {
      continue;
    }
    const param = AST.genericParameter(occurrence.name, undefined, { isInferred: true }) as AST.GenericParameter;
    if (occurrence.node) {
      param.origin = (occurrence.node as AST.AstNode).origin;
      param.span = (occurrence.node as AST.AstNode).span;
    }
    inferred.push(param);
    inferredMap.set(occurrence.name, param);
    known.add(occurrence.name);
  }
  if (inferred.length === 0) {
    return;
  }
  node.genericParams = Array.isArray(node.genericParams) ? [...node.genericParams, ...inferred] : inferred;
  node.whereClause = hoistWhereClauses(node.whereClause, inferredMap);
  node.inferredGenericParams = Array.isArray(node.inferredGenericParams)
    ? [...node.inferredGenericParams, ...inferred]
    : inferred;
}

function collectFunctionLikeOccurrences(node: FunctionLikeNode): TypeOccurrence[] {
  const occurrences: TypeOccurrence[] = [];
  if (Array.isArray(node.params)) {
    for (const param of node.params) {
      if (!param?.paramType) continue;
      collectTypeExpressionOccurrences(param.paramType, occurrences);
    }
  }
  if (node.returnType) {
    collectTypeExpressionOccurrences(node.returnType, occurrences);
  }
  if (Array.isArray(node.whereClause)) {
    for (const clause of node.whereClause) {
      if (clause?.typeParam) {
        collectTypeExpressionOccurrences(clause.typeParam, occurrences, "where");
      }
    }
  }
  return occurrences;
}

function collectTypeExpressionOccurrences(
  expr: AST.TypeExpression | null | undefined,
  acc: TypeOccurrence[],
  kind: TypeOccurrence["kind"] = "type",
): void {
  if (!expr) {
    return;
  }
  switch (expr.type) {
    case "SimpleTypeExpression":
      if (expr.name?.name) {
        acc.push({ name: expr.name.name, node: expr, kind });
      }
      break;
    case "GenericTypeExpression":
      collectTypeExpressionOccurrences(expr.base, acc, kind);
      if (Array.isArray(expr.arguments)) {
        for (const arg of expr.arguments) {
          collectTypeExpressionOccurrences(arg, acc, kind);
        }
      }
      break;
    case "FunctionTypeExpression":
      if (Array.isArray(expr.paramTypes)) {
        for (const paramType of expr.paramTypes) {
          collectTypeExpressionOccurrences(paramType, acc, kind);
        }
      }
      collectTypeExpressionOccurrences(expr.returnType, acc, kind);
      break;
    case "NullableTypeExpression":
      collectTypeExpressionOccurrences(expr.innerType, acc, kind);
      break;
    case "ResultTypeExpression":
      collectTypeExpressionOccurrences(expr.innerType, acc, kind);
      break;
    case "UnionTypeExpression":
      if (Array.isArray(expr.members)) {
        for (const member of expr.members) {
          collectTypeExpressionOccurrences(member, acc, kind);
        }
      }
      break;
    default:
      break;
  }
}

function classifyInferenceCandidate(
  ctx: DeclarationsContext,
  name: string | undefined,
  known: Set<string>,
): "infer" | "known-type" | "skip" {
  if (!name) {
    return "skip";
  }
  if (known.has(name)) {
    return "skip";
  }
  if (name.includes(".")) {
    return "skip";
  }
  if (ctx.hasTypeDefinition(name)) {
    return "known-type";
  }
  if (ctx.isKnownTypeName(name)) {
    return "skip";
  }
  return "infer";
}

function structTypeFromScope(structName: string, typeParamNames?: string[]): TypeInfo {
  const baseName = structName.split("<")[0]?.trim() ?? structName;
  const args = Array.isArray(typeParamNames)
    ? typeParamNames.map((paramName) => (paramName ? unknownType : unknownType))
    : [];
  return {
    kind: "struct",
    name: baseName,
    typeArguments: args,
  };
}

function hoistWhereClauses(
  clauses: AST.WhereClauseConstraint[] | undefined,
  inferred: Map<string, AST.GenericParameter>,
): AST.WhereClauseConstraint[] | undefined {
  if (!Array.isArray(clauses) || inferred.size === 0) {
    return clauses;
  }
  const remaining: AST.WhereClauseConstraint[] = [];
  for (const clause of clauses) {
    const clauseName = clause?.typeParam?.type === "SimpleTypeExpression"
      ? clause.typeParam.name?.name
      : null;
    if (!clauseName) {
      remaining.push(clause);
      continue;
    }
    const target = inferred.get(clauseName);
    if (!target) {
      remaining.push(clause);
      continue;
    }
    if (Array.isArray(clause.constraints) && clause.constraints.length > 0) {
      target.constraints = [...(target.constraints ?? []), ...clause.constraints];
    }
  }
  return remaining.length > 0 ? remaining : undefined;
}
