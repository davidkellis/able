import type * as AST from "../../ast";
import { formatType, primitiveType, unknownType, type TypeInfo } from "../types";
import { refineTypeWithExpected } from "./expressions";
import { typeImplementsInterface } from "./implementations";
import type { ImplementationContext } from "./implementations";
import type { TypeResolutionOptions } from "./type-resolution";

export type CheckerBaseFunctionHost = {
  env: {
    define(name: string, value: TypeInfo): void;
    pushScope(): void;
    popScope(): void;
  };
  implementationContext: ImplementationContext;
  resolveTypeExpression(
    expr: AST.TypeExpression | null | undefined,
    substitutions?: Map<string, TypeInfo>,
    options?: TypeResolutionOptions,
  ): TypeInfo;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  inferExpressionWithExpected(expression: AST.Expression | undefined | null, expected: TypeInfo): TypeInfo;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean;
  report(message: string, node?: AST.Node | null): void;
  formatTypeExpression(expr: AST.TypeExpression, substitutions?: Map<string, string>): string;
  describeTypeArgument(type: TypeInfo): string;
  getInterfaceNameFromConstraint(constraint: AST.InterfaceConstraint | null | undefined): string | null;
  pushReturnType(type: TypeInfo): void;
  popReturnType(): void;
  currentReturnType(): TypeInfo | undefined;
  pushFunctionGenericContext(definition: AST.FunctionDefinition): void;
  popFunctionGenericContext(): void;
  pushTypeParamScope(definition: AST.FunctionDefinition): void;
  popTypeParamScope(): void;
};

export function checkFunctionDefinition(
  checker: CheckerBaseFunctionHost,
  definition: AST.FunctionDefinition,
): void {
  if (!definition) return;
  const name = definition.id?.name ?? "<anonymous>";
  const paramTypes = Array.isArray(definition.params)
    ? definition.params.map((param) => checker.resolveTypeExpression(param?.paramType))
    : [];
  const expectedReturn = checker.resolveTypeExpression(definition.returnType);
  if (definition.id?.name) {
    checker.env.define(definition.id.name, {
      kind: "function",
      parameters: paramTypes,
      returnType: expectedReturn ?? unknownType,
    });
  }
  checker.pushReturnType(expectedReturn ?? unknownType);
  checker.pushFunctionGenericContext(definition);
  checker.pushTypeParamScope(definition);
  checker.env.pushScope();
  try {
    if (Array.isArray(definition.params)) {
      definition.params.forEach((param, index) => {
        const paramName = checker.getIdentifierName(param?.name);
        if (!paramName) return;
        const paramType = paramTypes[index] ?? unknownType;
        checker.env.define(paramName, paramType ?? unknownType);
      });
    }
    const shouldUseExpected =
      expectedReturn &&
      expectedReturn.kind !== "unknown" &&
      !(expectedReturn.kind === "primitive" && expectedReturn.name === "void");
    const bodyType = shouldUseExpected
      ? checker.inferExpressionWithExpected(definition.body, expectedReturn)
      : checker.inferExpression(definition.body);
    const expectedVoid =
      expectedReturn?.kind === "primitive" && expectedReturn.name === "void";
    if (!expectedVoid && expectedReturn && expectedReturn.kind !== "unknown" && bodyType && bodyType.kind !== "unknown") {
      const literalMessage = checker.describeLiteralMismatch(bodyType, expectedReturn);
      if (literalMessage) {
        checker.report(literalMessage, definition.body ?? definition);
      } else if (!checker.isTypeAssignable(bodyType, expectedReturn)) {
        checker.report(
          `typechecker: function '${name}' body returns ${formatType(bodyType)}, expected ${formatType(expectedReturn)}`,
          definition.body ?? definition,
        );
      }
    }

    const genericNames = new Set<string>();
    const addGenericName = (param?: AST.GenericParameter | null) => {
      const paramName = checker.getIdentifierName(param?.name);
      if (paramName) {
        genericNames.add(paramName);
      }
    };
    (definition.genericParams ?? []).forEach(addGenericName);
    (definition.inferredGenericParams ?? []).forEach(addGenericName);
    if (Array.isArray(definition.whereClause) && definition.whereClause.length > 0) {
      const subjectUsesGeneric = (expr: AST.TypeExpression | null | undefined): boolean => {
        if (!expr) return false;
        switch (expr.type) {
          case "SimpleTypeExpression": {
            const id = checker.getIdentifierName(expr.name);
            return id ? genericNames.has(id) || id === "Self" : false;
          }
          case "GenericTypeExpression":
            if (subjectUsesGeneric(expr.base)) return true;
            return Array.isArray(expr.arguments) && expr.arguments.some((arg) => subjectUsesGeneric(arg));
          case "NullableTypeExpression":
          case "ResultTypeExpression":
            return subjectUsesGeneric(expr.innerType);
          case "UnionTypeExpression":
            return Array.isArray(expr.members) && expr.members.some((member) => subjectUsesGeneric(member));
          case "FunctionTypeExpression":
            if (Array.isArray(expr.paramTypes) && expr.paramTypes.some((param) => subjectUsesGeneric(param))) {
              return true;
            }
            return subjectUsesGeneric(expr.returnType);
          default:
            return false;
        }
      };
      for (const clause of definition.whereClause) {
        if (!clause?.typeParam || !Array.isArray(clause.constraints)) {
          continue;
        }
        if (subjectUsesGeneric(clause.typeParam)) {
          continue;
        }
        const subject = checker.resolveTypeExpression(clause.typeParam);
        for (const constraint of clause.constraints) {
          const interfaceName = checker.getInterfaceNameFromConstraint(constraint);
          if (!interfaceName) continue;
          const expectedArgs =
            constraint?.interfaceType?.type === "GenericTypeExpression"
              ? (constraint.interfaceType.arguments ?? []).map((arg) => checker.formatTypeExpression(arg))
              : [];
          const result = typeImplementsInterface(checker.implementationContext, subject, interfaceName, expectedArgs);
          if (!result.ok) {
            const subjectLabel = checker.describeTypeArgument(subject);
            const detailSuffix = result.detail ? `: ${result.detail}` : "";
            const message =
              `typechecker: fn ${name} constraint on ${checker.formatTypeExpression(clause.typeParam)} is not satisfied: ` +
              `${subjectLabel} does not implement ${interfaceName}${detailSuffix}`;
            checker.report(message, constraint?.interfaceType ?? clause.typeParam ?? definition);
          }
        }
      }
    }
  } finally {
    checker.env.popScope();
    checker.popTypeParamScope();
    checker.popFunctionGenericContext();
    checker.popReturnType();
  }
}

export function checkReturnStatement(
  checker: CheckerBaseFunctionHost,
  statement: AST.ReturnStatement,
): void {
  if (!statement) return;
  const expected = checker.currentReturnType();
  const shouldUseExpected =
    expected &&
    expected.kind !== "unknown" &&
    !(expected.kind === "primitive" && expected.name === "void");
  let actual = statement.argument
    ? shouldUseExpected
      ? checker.inferExpressionWithExpected(statement.argument, expected)
      : checker.inferExpression(statement.argument)
    : primitiveType("void");
  if (statement.argument?.type === "FunctionCall" && expected && expected.kind !== "unknown") {
    actual = refineTypeWithExpected(actual, expected);
  }
  if (!expected) {
    checker.report("typechecker: return statement outside function", statement);
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
  const literalMessage = checker.describeLiteralMismatch(actual, expected);
  if (literalMessage) {
    checker.report(literalMessage, statement.argument ?? statement);
    return;
  }
  if (expected.kind === "result") {
    if (checker.isTypeAssignable(actual, expected.inner)) {
      return;
    }
    if (typeImplementsInterface(checker.implementationContext, actual, "Error", [] as string[]).ok) {
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
      return typeImplementsInterface(checker.implementationContext, actual, member.name, args).ok;
    }
    if (member.kind === "struct" && member.name === "Error") {
      return typeImplementsInterface(checker.implementationContext, actual, "Error", []).ok;
    }
    return false;
  };
  if (expected.kind === "union" && expected.members?.some((member) => unionMatches(member))) {
    return;
  }
  if (!checker.isTypeAssignable(actual, expected)) {
    checker.report(
      `typechecker: return expects ${formatType(expected)}, got ${formatType(actual)}`,
      statement.argument ?? statement,
    );
  }
}
