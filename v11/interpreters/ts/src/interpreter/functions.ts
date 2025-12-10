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

function resolveApplyFunction(
  ctx: InterpreterV10,
  callee: V10Value,
): Extract<V10Value, { kind: "function" | "function_overload" }> | null {
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
  if (node.callee.type === "Identifier") {
    let calleeValue: V10Value | null = null;
    let lookupError: unknown;
    try {
      calleeValue = env.get(node.callee.name);
    } catch (err) {
      lookupError = err;
    }
    const callArgs = node.arguments.map((arg) => ctx.evaluate(arg, env));
    if (!calleeValue) {
      if (lookupError) {
        throw lookupError;
      }
      throw new Error(`Undefined variable '${node.callee.name}'`);
    }
    return callCallableValue(ctx, calleeValue, callArgs, env, node);
  }
  const calleeEvaluated = ctx.evaluate(node.callee, env);
  const callArgs = node.arguments.map((arg) => ctx.evaluate(arg, env));
  return callCallableValue(ctx, calleeEvaluated, callArgs, env, node);
}

export function callCallableValue(ctx: InterpreterV10, callee: V10Value, args: V10Value[], env: Environment, callNode?: AST.FunctionCall): V10Value {
  if (callee.kind === "partial_function") {
    const mergedCall = callNode ?? callee.callNode;
    return callCallableValue(ctx, callee.target, [...callee.boundArgs, ...args], env, mergedCall);
  }

  let funcValue: Extract<V10Value, { kind: "function" }> | null = null;
  let overloadSet: Extract<V10Value, { kind: "function_overload" }> | null = null;
  let nativeFunc: Extract<V10Value, { kind: "native_function" }> | null = null;
  let injectedArgs: V10Value[] = [];

  if (callee.kind === "bound_method") {
    if (callee.func.kind === "function_overload") {
      overloadSet = callee.func;
    } else {
      funcValue = callee.func;
    }
    injectedArgs = [callee.self];
  } else if (callee.kind === "function") {
    funcValue = callee;
  } else if (callee.kind === "function_overload") {
    overloadSet = callee;
  } else if (callee.kind === "dyn_ref") {
    const bucket = ctx.packageRegistry.get(callee.pkg);
    const sym = bucket?.get(callee.name);
    if (!sym || (sym.kind !== "function" && sym.kind !== "function_overload")) {
      throw new Error(`dyn ref '${callee.pkg}.${callee.name}' is not callable`);
    }
    if (sym.kind === "function_overload") {
      overloadSet = sym;
    } else {
      funcValue = sym;
    }
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

  let evalArgs = [...injectedArgs, ...args];
  if (overloadSet) {
    const range = overloadArityRange(overloadSet.overloads);
    if (evalArgs.length > range.maxArgs) {
      const name = callNode?.callee?.type === "Identifier" ? callNode.callee.name : "(overload)";
      throw new Error(`Arity mismatch calling ${name}: expected at most ${range.maxArgs}, got ${evalArgs.length}`);
    }
    if (evalArgs.length < range.minArgs) {
      return makePartialFunction(overloadSet, evalArgs, callNode);
    }
    const selected = selectRuntimeOverload(ctx, overloadSet.overloads, evalArgs, callNode);
    if (!selected) {
      const name = callNode?.callee?.type === "Identifier" ? callNode.callee.name : "(overload)";
      throw new Error(`No overloads of ${name} match provided arguments`);
    }
    funcValue = selected;
  }
  if (nativeFunc) {
    if (nativeFunc.arity >= 0) {
      if (evalArgs.length > nativeFunc.arity) {
        throw new Error(`Arity mismatch calling ${nativeFunc.name}: expected ${nativeFunc.arity}, got ${evalArgs.length}`);
      }
      if (evalArgs.length < nativeFunc.arity) {
        return makePartialFunction(nativeFunc, evalArgs, callNode);
      }
    }
    return nativeFunc.impl(ctx, evalArgs);
  }

  if (!funcValue) throw new Error("Callable target missing function value");
  const funcNode = funcValue.node;
  const arity = functionArityRange(funcNode);
  if (evalArgs.length > arity.maxArgs) {
    const name = funcNode.type === "FunctionDefinition" ? funcNode.id?.name ?? "(anonymous)" : "(lambda)";
    throw new Error(`Arity mismatch calling ${name}: expected at most ${arity.maxArgs}, got ${evalArgs.length}`);
  }
  if (evalArgs.length < arity.minArgs) {
    return makePartialFunction(funcValue, evalArgs, callNode);
  }
  if (arity.optionalLast && evalArgs.length === arity.maxArgs - 1) {
    evalArgs = [...evalArgs, { kind: "nil", value: null }];
  }
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

function makePartialFunction(target: V10Value, boundArgs: V10Value[], callNode?: AST.FunctionCall): Extract<V10Value, { kind: "partial_function" }> {
  return { kind: "partial_function", target, boundArgs, callNode };
}

function functionArityRange(funcNode: AST.FunctionDefinition | AST.LambdaExpression): { minArgs: number; maxArgs: number; optionalLast: boolean } {
  const params = funcNode.params ?? [];
  const paramCount = params.length;
  const optionalLast = paramCount > 0 && params[paramCount - 1]?.paramType?.type === "NullableTypeExpression";
  const maxArgs = funcNode.type === "FunctionDefinition" && funcNode.isMethodShorthand ? paramCount + 1 : paramCount;
  const minArgs = optionalLast ? Math.max(0, maxArgs - 1) : maxArgs;
  return { minArgs, maxArgs, optionalLast };
}

function overloadArityRange(overloads: Array<Extract<V10Value, { kind: "function" }>>): { minArgs: number; maxArgs: number } {
  let minArgs = Number.POSITIVE_INFINITY;
  let maxArgs = 0;
  for (const fn of overloads) {
    const arity = functionArityRange(fn.node);
    if (arity.minArgs < minArgs) minArgs = arity.minArgs;
    if (arity.maxArgs > maxArgs) maxArgs = arity.maxArgs;
  }
  if (!Number.isFinite(minArgs)) minArgs = 0;
  return { minArgs, maxArgs };
}

function selectRuntimeOverload(
  ctx: InterpreterV10,
  overloads: Array<Extract<V10Value, { kind: "function" }>>,
  evalArgs: V10Value[],
  callNode?: AST.FunctionCall,
): Extract<V10Value, { kind: "function" }> | null {
  const candidates: Array<{
    fn: Extract<V10Value, { kind: "function" }>;
    params: AST.FunctionParameter[];
    optionalLast: boolean;
    score: number;
    priority: number;
  }> = [];
  for (const fn of overloads) {
    const funcNode = fn.node;
    if (funcNode.type !== "FunctionDefinition") {
      candidates.push({ fn, params: funcNode.params ?? [], optionalLast: false, score: 0 });
      continue;
    }
    const params = funcNode.params ?? [];
    const arity = functionArityRange(funcNode);
    if (evalArgs.length < arity.minArgs || evalArgs.length > arity.maxArgs) {
      continue;
    }
    const paramsForCheck =
      arity.optionalLast && evalArgs.length === arity.maxArgs - 1 ? params.slice(0, Math.max(0, params.length - 1)) : params;
    const argsForCheck = funcNode.isMethodShorthand ? evalArgs.slice(1) : evalArgs;
    if (argsForCheck.length !== paramsForCheck.length) {
      continue;
    }
    const generics = new Set((funcNode.genericParams ?? []).map((gp) => gp.name.name));
    let score = arity.optionalLast && evalArgs.length === arity.maxArgs - 1 ? -0.5 : 0;
    let compatible = true;
    for (let i = 0; i < paramsForCheck.length; i += 1) {
      const param = paramsForCheck[i];
      const arg = argsForCheck[i];
      if (!param || arg === undefined) {
        compatible = false;
        break;
      }
      if (param.paramType) {
        if (!ctx.matchesType(param.paramType, arg)) {
          compatible = false;
          break;
        }
        score += parameterSpecificity(param.paramType, generics);
      }
    }
    if (compatible) {
      const priority = typeof (fn as any).methodResolutionPriority === "number" ? (fn as any).methodResolutionPriority : 0;
      candidates.push({ fn, params: paramsForCheck, optionalLast: arity.optionalLast, score, priority });
    }
  }
  if (!candidates.length) return null;
  candidates.sort((a, b) => {
    if (a.score !== b.score) return b.score - a.score;
    return b.priority - a.priority;
  });
  const best = candidates[0]!;
  const ties = candidates.filter((c) => c.score === best.score && c.priority === best.priority);
  if (ties.length > 1) {
    const name =
      callNode?.callee?.type === "Identifier"
        ? callNode.callee.name
        : callNode?.callee?.type === "MemberAccessExpression"
          ? callNode.callee.member?.type === "Identifier"
            ? callNode.callee.member.name
            : "(function)"
          : "(function)";
    throw new Error(`Ambiguous overload for ${name}`);
  }
  return best.fn;
}

function parameterSpecificity(typeExpr: AST.TypeExpression, generics: Set<string>): number {
  if (!typeExpr) return 0;
  switch (typeExpr.type) {
    case "WildcardTypeExpression":
      return 0;
    case "SimpleTypeExpression": {
      const name = typeExpr.name.name;
      if (generics.has(name) || /^[A-Z]$/.test(name)) {
        return 1;
      }
      return 3;
    }
    case "NullableTypeExpression":
      return 1 + parameterSpecificity(typeExpr.innerType, generics);
    case "GenericTypeExpression":
      return 2 + (typeExpr.arguments ?? []).reduce((acc, arg) => acc + parameterSpecificity(arg, generics), 0);
    case "FunctionTypeExpression":
    case "UnionTypeExpression":
      return 2;
    default:
      return 1;
  }
}
