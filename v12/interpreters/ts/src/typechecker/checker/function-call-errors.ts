import * as AST from "../../ast";
import { formatType, type TypeInfo } from "../types";
import {
  ambiguousImplementationDetail,
  typeImplementsInterface,
} from "./implementations";
import { dropOptionalParam } from "./function-call-parse";
import type { FunctionCallContext } from "./function-calls";
import type { FunctionInfo } from "./types";

export function reportAmbiguousInterfaceMethod(
  ctx: FunctionCallContext,
  receiverType: TypeInfo,
  methodName: string,
  node: AST.Node,
): boolean {
  if (!receiverType || receiverType.kind === "unknown") {
    return false;
  }
  const implementations = ctx.implementationContext.getImplementationRecords?.();
  if (!implementations) {
    return false;
  }
  const interfaces = new Set<string>();
  for (const record of implementations) {
    if (!record?.definition || !Array.isArray(record.definition.definitions)) continue;
    const hasMethod = record.definition.definitions.some(
      (fn) => fn?.type === "FunctionDefinition" && fn.id?.name === methodName,
    );
    if (hasMethod) {
      interfaces.add(record.interfaceName);
    }
  }
  for (const interfaceName of interfaces) {
    const detail = ambiguousImplementationDetail(ctx.implementationContext, receiverType, interfaceName);
    if (detail) {
      ctx.report(`typechecker: ${detail}`, node);
      return true;
    }
    const result = typeImplementsInterface(ctx.implementationContext, receiverType, interfaceName);
    if (!result.ok && result.detail && result.detail.includes("ambiguous implementations")) {
      ctx.report(`typechecker: ${result.detail}`, node);
      return true;
    }
  }
  return false;
}

export function reportArgumentDiagnostics(
  ctx: FunctionCallContext,
  info: FunctionInfo,
  params: TypeInfo[],
  optionalLast: boolean,
  call: AST.FunctionCall,
  args: AST.Expression[],
  argTypes: TypeInfo[],
): boolean {
  if (params.length !== args.length && !(optionalLast && args.length === params.length - 1)) {
    ctx.report(`typechecker: function expects ${params.length} arguments, got ${args.length}`, call);
    return false;
  }
  const compareParams = dropOptionalParam(params, args.length, optionalLast);
  const compareCount = Math.min(compareParams.length, argTypes.length);
  for (let index = 0; index < compareCount; index += 1) {
    const expected = compareParams[index];
    const actual = argTypes[index];
    if (!expected || expected.kind === "unknown" || !actual || actual.kind === "unknown") {
      continue;
    }
    const literalMessage = ctx.describeLiteralMismatch(actual, expected);
    if (literalMessage) {
      ctx.report(literalMessage, args[index] ?? call);
      return false;
    }
    if (!ctx.isTypeAssignable(actual, expected)) {
      const interfaceArgs =
        expected.kind === "interface" && Array.isArray(expected.typeArguments)
          ? expected.typeArguments.map((arg) => (arg?.kind === "unknown" ? "Unknown" : formatType(arg)))
          : [];
      if (
        expected.kind === "interface" &&
        typeImplementsInterface(ctx.implementationContext, actual, expected.name, interfaceArgs).ok
      ) {
        continue;
      }
      if (
        expected.kind === "interface" &&
        ctx.implementationContext.getImplementationBucket?.(formatType(actual))?.some(
          (record) => record.interfaceName === expected.name,
        )
      ) {
        continue;
      }
      ctx.report(
        `typechecker: argument ${index + 1} has type ${formatType(actual)}, expected ${formatType(expected)}`,
        args[index] ?? call,
      );
      return false;
    }
  }
  return true;
}

export function formatCalleeLabel(callee: AST.Expression | undefined | null): string | null {
  if (!callee) return null;
  if (callee.type === "Identifier") {
    return callee.name;
  }
  if (callee.type === "MemberAccessExpression") {
    const member = callee.member;
    if (member?.type === "Identifier") {
      return member.name;
    }
  }
  return null;
}
