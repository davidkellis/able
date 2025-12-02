import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { ReturnSignal } from "./signals";
import { memberAccessOnValue } from "./structs";
import { collectTypeDispatches } from "./type-dispatch";

function isGenericTypeReference(typeExpr: AST.TypeExpression | undefined, genericNames: Set<string>): boolean {
  if (!typeExpr || genericNames.size === 0) return false;
  if (typeExpr.type === "SimpleTypeExpression") {
    return genericNames.has(typeExpr.name.name);
  }
  return false;
}

export function evaluateFunctionDefinition(ctx: InterpreterV10, node: AST.FunctionDefinition, env: Environment): V10Value {
  const value: V10Value = { kind: "function", node, closureEnv: env };
  env.define(node.id.name, value);
  ctx.registerSymbol(node.id.name, value);
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, value); } catch {}
  }
  return { kind: "nil", value: null };
}

export function evaluateLambdaExpression(ctx: InterpreterV10, node: AST.LambdaExpression, env: Environment): V10Value {
  return { kind: "function", node, closureEnv: env };
}

function resolveApplyFunction(ctx: InterpreterV10, callee: V10Value): Extract<V10Value, { kind: "function" }> | null {
  const dispatches = collectTypeDispatches(ctx, callee);
  for (const dispatch of dispatches) {
    const method = ctx.findMethod(dispatch.typeName, "apply", {
      typeArgs: dispatch.typeArgs,
      interfaceName: "Apply",
    });
    if (method) return method;
  }
  return null;
}

export function evaluateFunctionCall(ctx: InterpreterV10, node: AST.FunctionCall, env: Environment): V10Value {
  if (node.callee.type === "MemberAccessExpression") {
    const receiver = ctx.evaluate(node.callee.object, env);
    if (node.callee.isSafe && receiver.kind === "nil") {
      return receiver;
    }
    const memberValue = memberAccessOnValue(ctx, receiver, node.callee.member, env, { preferMethods: true });
    const callArgs = node.arguments.map((arg) => ctx.evaluate(arg, env));
    return callCallableValue(ctx, memberValue, callArgs, env, node);
  }
  const calleeEvaluated = ctx.evaluate(node.callee, env);
  const callArgs = node.arguments.map(arg => ctx.evaluate(arg, env));
  return callCallableValue(ctx, calleeEvaluated, callArgs, env, node);
}

export function callCallableValue(ctx: InterpreterV10, callee: V10Value, args: V10Value[], env: Environment, callNode?: AST.FunctionCall): V10Value {
  let funcValue: Extract<V10Value, { kind: "function" }> | null = null;
  let nativeFunc: Extract<V10Value, { kind: "native_function" }> | null = null;
  let injectedArgs: V10Value[] = [];

  if (callee.kind === "bound_method") {
    funcValue = callee.func;
    injectedArgs = [callee.self];
  } else if (callee.kind === "function") {
    funcValue = callee;
  } else if (callee.kind === "dyn_ref") {
    const bucket = ctx.packageRegistry.get(callee.pkg);
    const sym = bucket?.get(callee.name);
    if (!sym || sym.kind !== "function") throw new Error(`dyn ref '${callee.pkg}.${callee.name}' is not callable`);
    funcValue = sym;
  } else if (callee.kind === "native_bound_method") {
    nativeFunc = callee.func;
    injectedArgs = [callee.self];
  } else if (callee.kind === "native_function") {
    nativeFunc = callee;
  } else {
    const location =
      callNode && (callNode as any).span && (callNode as any).origin
        ? `${(callNode as any).origin}:${(callNode as any).span.start.line + 1}:${(callNode as any).span.start.column + 1}`
        : "";
    const suffix = location ? ` at ${location}` : "";
    const applyFn = resolveApplyFunction(ctx, callee);
    if (applyFn) {
      return callCallableValue(ctx, applyFn, [callee, ...args], env, callNode);
    }
    throw new Error(`Cannot call non-function (kind ${callee.kind})${suffix}`);
  }

  const evalArgs = [...injectedArgs, ...args];
  if (nativeFunc) {
    if (nativeFunc.arity >= 0 && evalArgs.length !== nativeFunc.arity) {
      throw new Error(`Arity mismatch calling ${nativeFunc.name}: expected ${nativeFunc.arity}, got ${evalArgs.length}`);
    }
    return nativeFunc.impl(ctx, evalArgs);
  }

  if (!funcValue) throw new Error("Callable target missing function value");
  const funcNode = funcValue.node;
  if (callNode) {
    ctx.inferTypeArgumentsFromCall(funcNode, callNode, evalArgs);
    ctx.enforceGenericConstraintsIfAny(funcNode, callNode);
  }
  const funcEnv = new Environment(funcValue.closureEnv);
  if (callNode) {
    ctx.bindTypeArgumentsIfAny(funcNode, callNode, funcEnv);
  }

  if (funcNode.type === "FunctionDefinition") {
    const genericNames = new Set((funcNode.genericParams ?? []).map((gp) => gp.name.name));
    const params = funcNode.params;
    const paramCount = params.length;
    const expectedArgs = funcNode.isMethodShorthand ? paramCount + 1 : paramCount;
    if (evalArgs.length !== expectedArgs) {
      const name = funcNode.id?.name ?? "(anonymous)";
      throw new Error(`Arity mismatch calling ${name}: expected ${expectedArgs}, got ${evalArgs.length}`);
    }
    let bindArgs = evalArgs;
    let implicitReceiver: V10Value | null = null;
    let hasImplicit = false;
    if (funcNode.isMethodShorthand) {
      implicitReceiver = evalArgs[0]!;
      bindArgs = evalArgs.slice(1);
      hasImplicit = true;
    } else if (paramCount > 0 && evalArgs.length > 0) {
      implicitReceiver = evalArgs[0]!;
      hasImplicit = true;
    }
    if (bindArgs.length !== paramCount) {
      const name = funcNode.id?.name ?? "(anonymous)";
      throw new Error(`Arity mismatch calling ${name}: expected ${paramCount}, got ${bindArgs.length}`);
    }
    for (let i = 0; i < params.length; i++) {
      const param = params[i];
      const argVal = bindArgs[i];
      if (!param) throw new Error(`Parameter missing at index ${i}`);
      if (argVal === undefined) throw new Error(`Argument missing at index ${i}`);
      let coerced = argVal;
      const skipRuntimeTypeCheck = isGenericTypeReference(param.paramType, genericNames);
      if (param.paramType && !skipRuntimeTypeCheck) {
        if (!ctx.matchesType(param.paramType, argVal)) {
          const pname = (param.name as any).name ?? `param_${i}`;
          const expected = ctx.typeExpressionToString(param.paramType);
          const actual = ctx.getTypeNameForValue(argVal) ?? argVal.kind;
          const origin =
            callNode && (callNode as any).span && (callNode as any).origin
              ? `${(callNode as any).origin}:${(callNode as any).span.start.line + 1}:${(callNode as any).span.start.column + 1}`
              : null;
          const suffix = origin ? ` at ${origin}` : "";
          throw new Error(`Parameter type mismatch for '${pname}': expected ${expected}, got ${actual}${suffix}`);
        }
        coerced = ctx.coerceValueToType(param.paramType, argVal);
        bindArgs[i] = coerced;
      } else if (skipRuntimeTypeCheck) {
        coerced = argVal;
      }
      if (param.name.type === "Identifier") {
        funcEnv.define(param.name.name, coerced);
      } else if (param.name.type === "WildcardPattern") {
        // ignore
      } else if (param.name.type === "StructPattern" || param.name.type === "ArrayPattern" || param.name.type === "LiteralPattern" || param.name.type === "TypedPattern") {
        ctx.assignByPattern(param.name as AST.Pattern, coerced, funcEnv, true);
      } else {
        throw new Error("Only simple identifier and destructuring params supported for now");
      }
    }
    if (funcNode.isMethodShorthand && implicitReceiver && !funcEnv.hasInCurrentScope("self")) {
      funcEnv.define("self", implicitReceiver);
    }
    let pushedImplicit = false;
    if (hasImplicit && implicitReceiver) {
      ctx.implicitReceiverStack.push(implicitReceiver);
      pushedImplicit = true;
    }
    try {
      return ctx.evaluate(funcNode.body, funcEnv);
    } catch (e) {
      if (e instanceof ReturnSignal) return e.value;
      throw e;
    } finally {
      if (pushedImplicit) {
        ctx.implicitReceiverStack.pop();
      }
    }
  }

  if (funcNode.type === "LambdaExpression") {
    const params = funcNode.params;
    if (evalArgs.length !== params.length) {
      throw new Error(`Lambda expects ${params.length} arguments, got ${evalArgs.length}`);
    }
    for (let i = 0; i < params.length; i++) {
      const param = params[i];
      const argVal = evalArgs[i];
      if (!param) throw new Error(`Lambda parameter missing at index ${i}`);
      if (argVal === undefined) throw new Error(`Argument missing at index ${i}`);
      let coerced = argVal;
      if (param.paramType) {
        if (!ctx.matchesType(param.paramType, argVal)) {
          const pname = (param.name as any).name ?? `param_${i}`;
          throw new Error(`Parameter type mismatch for '${pname}'`);
        }
        coerced = ctx.coerceValueToType(param.paramType, argVal);
        evalArgs[i] = coerced;
      }
      if (param.name.type === "Identifier") {
        funcEnv.define(param.name.name, coerced);
      } else if (param.name.type === "WildcardPattern") {
        // ignore
      } else if (param.name.type === "StructPattern" || param.name.type === "ArrayPattern" || param.name.type === "LiteralPattern" || param.name.type === "TypedPattern") {
        ctx.assignByPattern(param.name as AST.Pattern, coerced, funcEnv, true);
      } else {
        throw new Error("Only simple identifier and destructuring params supported for now");
      }
    }
    let pushedImplicit = false;
    if (params.length > 0 && evalArgs.length > 0) {
      ctx.implicitReceiverStack.push(evalArgs[0]!);
      pushedImplicit = true;
    }
    try {
      return ctx.evaluate(funcNode.body as AST.AstNode, funcEnv);
    } finally {
      if (pushedImplicit) {
        ctx.implicitReceiverStack.pop();
      }
    }
  }

  throw new Error("calling unsupported function declaration");
}
