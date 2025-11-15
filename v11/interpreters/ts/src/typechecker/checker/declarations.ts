import * as AST from "../../ast";
import type { TypeInfo } from "../types";
import type { FunctionContext, FunctionInfo, ImplementationObligation } from "./types";
import type { StatementContext } from "./expressions";

export interface DeclarationsContext extends StatementContext {
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
  setFunctionInfo(key: string, info: FunctionInfo): void;
  getFunctionInfo(key: string): FunctionInfo | undefined;
}

export function collectFunctionDefinition(
  ctx: DeclarationsContext,
  definition: AST.FunctionDefinition,
  scope: FunctionContext | undefined,
): void {
  const name = definition.id?.name ?? "<anonymous>";
  const structName = scope?.structName;
  const fullName = structName ? `${structName}::${name}` : name;
  const returnType = ctx.resolveTypeExpression(definition.returnType);
  injectImplicitSelfParameter(definition, scope);
  const parameterTypes = resolveFunctionParameterTypes(ctx, definition);

  const info: FunctionInfo = {
    name,
    fullName,
    structName,
    genericConstraints: [],
    whereClause: extractFunctionWhereObligations(ctx, definition),
    genericParamNames: Array.isArray(definition.genericParams)
      ? definition.genericParams
          .map((param) => ctx.getIdentifierName(param?.name))
          .filter((paramName): paramName is string => Boolean(paramName))
      : [],
    returnType,
  };

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

  ctx.setFunctionInfo(fullName, info);
  if (!structName) {
    ctx.setFunctionInfo(name, info);
    if (definition.id?.name) {
      ctx.defineValue(definition.id.name, {
        kind: "function",
        parameters: parameterTypes,
        returnType,
      });
    }
  }
}

function resolveFunctionParameterTypes(ctx: DeclarationsContext, definition: AST.FunctionDefinition): TypeInfo[] {
  if (!Array.isArray(definition.params)) {
    return [];
  }
  return definition.params.map((param) => ctx.resolveTypeExpression(param?.paramType));
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
  ) => {
    const interfaceName = ctx.getInterfaceNameFromTypeExpression(interfaceType);
    if (!typeParam || !interfaceName) {
      return;
    }
    obligations.push({ typeParam, interfaceName, interfaceType: interfaceType ?? undefined, context });
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
      const typeParamName = ctx.getIdentifierName(clause?.typeParam);
      if (!typeParamName || !Array.isArray(clause?.constraints)) continue;
      for (const constraint of clause.constraints) {
        appendObligation(typeParamName, constraint?.interfaceType, "where clause");
      }
    }
  }

  return obligations;
}

function injectImplicitSelfParameter(definition: AST.FunctionDefinition, scope: FunctionContext | undefined): void {
  if (!scope?.structName || !Array.isArray(definition.params) || definition.params.length === 0) {
    return;
  }
  const firstParam = definition.params[0];
  if (!firstParam || firstParam.paramType) {
    return;
  }
  if (firstParam.name?.type !== "Identifier") {
    return;
  }
  const paramName = firstParam.name.name?.toLowerCase();
  if (paramName !== "self") {
    return;
  }
  firstParam.paramType = AST.simpleTypeExpression("Self");
}
