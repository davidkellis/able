import type * as AST from "../../ast";
import { arrayType, formatType, iteratorType, primitiveType, unknownType, type TypeInfo } from "../types";
import type { ImplementationContext } from "./implementations";
import {
  enforceFunctionConstraints as enforceFunctionConstraintsHelper,
  lookupMethodSetsForCall as lookupMethodSetsForCallHelper,
} from "./implementations";
import { mergeBranchTypes as mergeBranchTypesHelper } from "./expressions";
import type { StatementContext } from "./expression-context";
import type { FunctionInfo } from "./types";

type FunctionCallContext = {
  implementationContext: ImplementationContext;
  functionInfos: Map<string, FunctionInfo>;
  structDefinitions: Map<string, AST.StructDefinition>;
  inferExpression(expression: AST.Expression | undefined | null): TypeInfo;
  describeLiteralMismatch(actual?: TypeInfo, expected?: TypeInfo): string | null;
  isTypeAssignable(actual?: TypeInfo, expected?: TypeInfo): boolean;
  report(message: string, node?: AST.Node | null): void;
  handlePackageMemberAccess(expression: AST.MemberAccessExpression): boolean;
  getIdentifierName(node: AST.Identifier | null | undefined): string | null;
  checkBuiltinCallContext(name: string | undefined, call: AST.FunctionCall): void;
  getBuiltinCallName(callee: AST.Expression | undefined | null): string | undefined;
  statementContext: StatementContext;
};

export function checkFunctionCall(ctx: FunctionCallContext, call: AST.FunctionCall): void {
  const builtinName = ctx.getBuiltinCallName(call.callee);
  const args = Array.isArray(call.arguments) ? call.arguments : [];
  const argTypes = args.map((arg) => ctx.inferExpression(arg));
  ctx.checkBuiltinCallContext(builtinName, call);
  const infos = resolveFunctionInfos(ctx, call.callee);
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
    const optionalLast = params.length > 0 && params[params.length - 1]?.kind === "nullable";
    if (params.length !== args.length && !(optionalLast && args.length === params.length - 1)) {
      ctx.report(`typechecker: function expects ${params.length} arguments, got ${args.length}`, call);
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
      const literalMessage = ctx.describeLiteralMismatch(actual, expected);
      if (literalMessage) {
        ctx.report(literalMessage, args[index] ?? call);
        continue;
      }
      if (!ctx.isTypeAssignable(actual, expected)) {
        ctx.report(
          `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
          args[index] ?? call,
        );
      }
    }
  }
  for (const info of infos) {
    enforceFunctionConstraintsHelper(ctx.implementationContext, info, call);
  }
}

export function inferFunctionCallReturnType(ctx: FunctionCallContext, call: AST.FunctionCall): TypeInfo {
  const infos = resolveFunctionInfos(ctx, call.callee);
  if (!infos.length) {
    return unknownType;
  }
  const returnTypes = infos
    .map((info) => info.returnType ?? unknownType)
    .filter((type) => type && type.kind !== "unknown");
  if (!returnTypes.length) {
    return unknownType;
  }
  return mergeBranchTypesHelper(ctx.statementContext, returnTypes);
}

function resolveFunctionInfos(ctx: FunctionCallContext, callee: AST.Expression | undefined | null): FunctionInfo[] {
  if (!callee) return [];
  if (callee.type === "Identifier") {
    const info = ctx.functionInfos.get(callee.name);
    return info ? [info] : [];
  }
  if (callee.type === "MemberAccessExpression") {
    if (ctx.handlePackageMemberAccess(callee)) {
      return [];
    }
    const memberName = ctx.getIdentifierName(callee.member);
    if (!memberName) return [];
    let objectType = ctx.inferExpression(callee.object);
    if (
      objectType.kind !== "struct" &&
      callee.object?.type === "Identifier" &&
      callee.object.name &&
      ctx.structDefinitions.has(callee.object.name)
    ) {
      objectType = {
        kind: "struct",
        name: callee.object.name,
        typeArguments: [],
        definition: ctx.structDefinitions.get(callee.object.name),
      };
    }
    if (objectType.kind === "struct") {
      const structLabel = formatType(objectType);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey);
      const genericMatches = lookupMethodSetsForCallHelper(
        ctx.implementationContext,
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
        const fallback = ctx.functionInfos.get(memberName);
        if (fallback && !seen.has(fallback.fullName)) {
          infos.push(fallback);
        }
      }
      return infos;
    }
    if (objectType.kind === "array") {
      const arrayStruct: TypeInfo = {
        kind: "struct",
        name: "Array",
        typeArguments: [objectType.element ?? unknownType],
      };
      const structLabel = formatType(arrayStruct);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey);
      const genericMatches = lookupMethodSetsForCallHelper(
        ctx.implementationContext,
        structLabel,
        memberName,
        arrayStruct,
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
        const fallback = ctx.functionInfos.get(memberName);
        if (fallback && !seen.has(fallback.fullName)) {
          infos.push(fallback);
        }
      }
      return infos;
    }
    if (objectType.kind === "primitive" && objectType.name === "string") {
      const structLabel = formatType(objectType);
      const memberKey = `${structLabel}::${memberName}`;
      const infos: FunctionInfo[] = [];
      const seen = new Set<string>();
      const info = ctx.functionInfos.get(memberKey);
      const genericMatches = lookupMethodSetsForCallHelper(
        ctx.implementationContext,
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
        const fallback = ctx.functionInfos.get(memberName);
        if (fallback && !seen.has(fallback.fullName)) {
          infos.push(fallback);
        }
      }
      if (!infos.length) {
        const builtin = buildStringMethodInfo(memberName);
        if (builtin) infos.push(builtin);
      }
      return infos;
    }
  }
  return [];
}

function buildStringMethodInfo(memberName: string): FunctionInfo | null {
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
