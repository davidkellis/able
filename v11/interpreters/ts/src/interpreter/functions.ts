import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { ConstraintSpec, RuntimeValue } from "./values";
import { ProcYieldSignal, RaiseSignal, ReturnSignal } from "./signals";
import type { FunctionCallState, NativeCallState } from "./proc_continuations";
import { memberAccessOnValue } from "./structs";
import { collectTypeDispatches } from "./type-dispatch";

const CALL_TRACE_THRESHOLD = 120;
const callTraceStack: string[] = [];
let callTraceReported = false;
let matchesTraceCount = 0;

function describeCallTarget(
  ctx: Interpreter | null,
  callee: RuntimeValue,
  callNode?: AST.FunctionCall,
): string {
  const memberName = (() => {
    if (!callNode) return null;
    const calleeNode = callNode.callee;
    if (calleeNode.type === "MemberAccessExpression") {
      const member = calleeNode.member;
      if (member.type === "Identifier") return member.name;
    }
    if (calleeNode.type === "Identifier") return calleeNode.name;
    return null;
  })();
  const receiverName = (() => {
    if (!ctx) return null;
    if (callee.kind === "bound_method") {
      return ctx.getTypeNameForValue(callee.self);
    }
    if (callee.kind === "native_bound_method") {
      return ctx.getTypeNameForValue(callee.self);
    }
    return null;
  })();

  switch (callee.kind) {
    case "native_function":
      return callee.name;
    case "native_bound_method":
      return receiverName ? `${receiverName}.${callee.func.name}` : callee.func.name;
    case "function": {
      const node = callee.node;
      const name = node.type === "FunctionDefinition" ? node.id?.name ?? "(fn)" : "(lambda)";
      return memberName ? name : name;
    }
    case "function_overload": {
      const first = callee.overloads[0];
      const name = first?.node.type === "FunctionDefinition" ? first.node.id?.name ?? "(overload)" : "(overload)";
      return name;
    }
    case "bound_method": {
      const name = memberName ?? describeCallTarget(ctx, callee.func);
      return receiverName ? `${receiverName}.${name}` : name;
    }
    case "partial_function":
      return describeCallTarget(ctx, callee.target, callNode);
    case "dyn_ref":
      return `${callee.pkg}.${callee.name}`;
    default:
      if (memberName && receiverName) return `${receiverName}.${memberName}`;
      if (memberName) return memberName;
      return callee.kind;
  }
}

function isGenericTypeReference(typeExpr: AST.TypeExpression | undefined, genericNames: Set<string>): boolean {
  if (!typeExpr || genericNames.size === 0) return false;
  if (typeExpr.type === "SimpleTypeExpression") {
    return genericNames.has(typeExpr.name.name);
  }
  return false;
}

function isVoidTypeExpression(ctx: Interpreter, expr: AST.TypeExpression): boolean {
  const expanded = ctx.expandTypeAliases(expr);
  return expanded.type === "SimpleTypeExpression" && expanded.name.name === "void";
}

function canonicalizeTypeExpression(ctx: Interpreter, env: Environment, expr: AST.TypeExpression): AST.TypeExpression {
  const expanded = ctx.expandTypeAliases(expr);
  switch (expanded.type) {
    case "SimpleTypeExpression": {
      const name = expanded.name.name;
      let binding: RuntimeValue | null = null;
      try {
        binding = env.get(name);
      } catch {}
      if (!binding) return expanded;
      if (binding.kind === "struct_def" || binding.kind === "interface_def" || binding.kind === "union_def") {
        const canonical = binding.def.id.name;
        if (canonical && canonical !== name) {
          return AST.simpleTypeExpression(canonical);
        }
      }
      return expanded;
    }
    case "GenericTypeExpression":
      return AST.genericTypeExpression(
        canonicalizeTypeExpression(ctx, env, expanded.base),
        (expanded.arguments ?? []).map((arg) => (arg ? canonicalizeTypeExpression(ctx, env, arg) : arg)),
      );
    case "NullableTypeExpression":
      return AST.nullableTypeExpression(canonicalizeTypeExpression(ctx, env, expanded.innerType));
    case "ResultTypeExpression":
      return AST.resultTypeExpression(canonicalizeTypeExpression(ctx, env, expanded.innerType));
    case "UnionTypeExpression":
      return AST.unionTypeExpression((expanded.members ?? []).map((member) => canonicalizeTypeExpression(ctx, env, member)));
    case "FunctionTypeExpression":
      return AST.functionTypeExpression(
        (expanded.paramTypes ?? []).map((param) => canonicalizeTypeExpression(ctx, env, param)),
        canonicalizeTypeExpression(ctx, env, expanded.returnType),
      );
    default:
      return expanded;
  }
}

function coerceReturnValue(
  ctx: Interpreter,
  returnType: AST.TypeExpression | undefined,
  value: RuntimeValue,
  genericNames: Set<string>,
  env: Environment,
): RuntimeValue {
  if (!returnType) return value;
  const canonicalReturn = canonicalizeTypeExpression(ctx, env, returnType);
  if (canonicalReturn.type === "SimpleTypeExpression" && canonicalReturn.name.name === "void") {
    return { kind: "void" };
  }
  if (value.kind === "void") {
    if (canonicalReturn.type === "ResultTypeExpression" && isVoidTypeExpression(ctx, canonicalReturn.innerType)) {
      return value;
    }
    if (ctx.matchesType(canonicalReturn, value)) {
      return value;
    }
    const expected = ctx.typeExpressionToString(returnType);
    throw new Error(`Return type mismatch: expected ${expected}, got void`);
  }
  const skipRuntimeTypeCheck = isGenericTypeReference(returnType, genericNames);
  if (!skipRuntimeTypeCheck && !ctx.matchesType(returnType, value)) {
    const expected = ctx.typeExpressionToString(returnType);
    const actual = ctx.getTypeNameForValue(value) ?? value.kind;
    throw new Error(`Return type mismatch: expected ${expected}, got ${actual}`);
  }
  if (skipRuntimeTypeCheck) return value;
  return ctx.coerceValueToType(returnType, value);
}

function isThenable(value: unknown): value is Promise<RuntimeValue> {
  return !!value && typeof (value as any).then === "function";
}

function coerceAsyncError(ctx: Interpreter, err: unknown): RuntimeValue {
  if (err instanceof RaiseSignal) return err.value;
  if (err && typeof err === "object" && (err as any).kind === "error") {
    return err as RuntimeValue;
  }
  if (err instanceof Error) {
    return ctx.makeRuntimeError(err.message);
  }
  return ctx.makeRuntimeError(String(err));
}

export function evaluateFunctionDefinition(ctx: Interpreter, node: AST.FunctionDefinition, env: Environment): RuntimeValue {
  const value: RuntimeValue = { kind: "function", node, closureEnv: env };
  ctx.defineInEnv(env, node.id.name, value);
  ctx.registerSymbol(node.id.name, value);
  const qn = ctx.qualifiedName(node.id.name);
  if (qn) {
    try { ctx.globals.define(qn, value); } catch {}
  }
  return { kind: "nil", value: null };
}

export function evaluateLambdaExpression(ctx: Interpreter, node: AST.LambdaExpression, env: Environment): RuntimeValue {
  return { kind: "function", node, closureEnv: env };
}

function resolveApplyFunction(
  ctx: Interpreter,
  callee: RuntimeValue,
): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null {
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

export function evaluateFunctionCall(ctx: Interpreter, node: AST.FunctionCall, env: Environment): RuntimeValue {
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  if (procContext) {
    const state = procContext.getNativeCallState(node);
    if (state) {
      if (state.status === "resolved") {
        procContext.clearNativeCallState(node);
        return state.value ?? { kind: "nil", value: null };
      }
      if (state.status === "rejected") {
        procContext.clearNativeCallState(node);
        throw new RaiseSignal(state.error ?? ctx.makeRuntimeError("Async native call failed"));
      }
      const asyncCtx = ctx.currentAsyncContext ? ctx.currentAsyncContext() : null;
      if (asyncCtx) {
        (asyncCtx.handle as any).awaitBlocked = true;
      }
      return ctx.procYield(true);
    }
  }

  if (node.callee.type === "MemberAccessExpression") {
    const receiver = ctx.evaluate(node.callee.object, env);
    if (
      process.env.ABLE_TRACE_ERRORS &&
      node.callee.object.type === "Identifier" &&
      node.callee.object.name === "matcher" &&
      node.callee.member.type === "Identifier" &&
      node.callee.member.name === "matches"
    ) {
      const receiverType = ctx.getTypeNameForValue(receiver) ?? receiver.kind;
      console.error(`[trace] matcher.matches receiver ${receiverType} kind=${receiver.kind}`);
    }
    if (node.callee.isSafe && receiver.kind === "nil") {
      return receiver;
    }
    const memberValue = memberAccessOnValue(ctx, receiver, node.callee.member, env, { preferMethods: true });
    const callArgs = node.arguments.map((arg) => ctx.evaluate(arg, env));
    return callCallableValue(ctx, memberValue, callArgs, env, node);
  }
  if (node.callee.type === "Identifier") {
    let calleeValue: RuntimeValue | null = null;
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

export function callCallableValue(ctx: Interpreter, callee: RuntimeValue, args: RuntimeValue[], env: Environment, callNode?: AST.FunctionCall): RuntimeValue {
  const traceEnabled = !!process.env.ABLE_TRACE_ERRORS;
  if (traceEnabled) {
    callTraceStack.push(describeCallTarget(ctx, callee, callNode));
    if (!callTraceReported && callTraceStack.length >= CALL_TRACE_THRESHOLD) {
      callTraceReported = true;
      console.error(`[trace] call depth ${callTraceStack.length}: ${callTraceStack.slice(-30).join(" -> ")}`);
    }
    if (
      callNode?.callee.type === "MemberAccessExpression" &&
      callNode.callee.member.type === "Identifier" &&
      callNode.callee.member.name === "matches"
    ) {
      const receiver = callee.kind === "bound_method" ? callee.self : null;
      const receiverName = receiver ? ctx.getTypeNameForValue(receiver) ?? receiver.kind : "<none>";
      if ((receiverName === "ContainMatcher" || receiverName === "RaiseErrorMatcher") && matchesTraceCount < 50) {
        matchesTraceCount += 1;
        console.error(`[trace] matches call receiver ${receiverName} kind=${receiver?.kind ?? "none"}`);
      }
    }
    if (callNode?.callee.type === "Identifier" && callNode.callee.name === "expect" && args.length > 0) {
      const actual = args[0]!;
      const actualType = ctx.getTypeNameForValue(actual) ?? actual.kind;
      console.error(`[trace] expect arg ${actualType} kind=${actual.kind}`);
    }
  }
  try {
  if (callee.kind === "partial_function") {
    const mergedCall = callNode ?? callee.callNode;
    return callCallableValue(ctx, callee.target, [...callee.boundArgs, ...args], env, mergedCall);
  }

  let funcValue: Extract<RuntimeValue, { kind: "function" }> | null = null;
  let overloadSet: Extract<RuntimeValue, { kind: "function_overload" }> | null = null;
  let nativeFunc: Extract<RuntimeValue, { kind: "native_function" }> | null = null;
  let injectedArgs: RuntimeValue[] = [];

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
    if (sym.kind === "function") {
      if (sym.node.type === "FunctionDefinition" && sym.node.isPrivate) {
        throw new Error(`dyn ref '${callee.pkg}.${callee.name}' is private`);
      }
    } else {
      const first = sym.overloads[0];
      if (first?.node.type === "FunctionDefinition" && first.node.isPrivate) {
        throw new Error(`dyn ref '${callee.pkg}.${callee.name}' is private`);
      }
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
    const result = nativeFunc.impl(ctx, evalArgs);
    if (isThenable(result)) {
      if (!callNode) {
        throw new Error(`Async native call '${nativeFunc.name}' requires a call site`);
      }
      const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
      if (!procContext) {
        throw new Error(`Async native call '${nativeFunc.name}' requires an async context`);
      }
      const asyncCtx = ctx.currentAsyncContext ? ctx.currentAsyncContext() : null;
      if (!asyncCtx) {
        throw new Error(`Async native call '${nativeFunc.name}' requires an async task`);
      }
      const state: NativeCallState = { status: "pending" };
      procContext.setNativeCallState(callNode, state);
      const handle = asyncCtx.handle as any;
      handle.awaitBlocked = true;
      const scheduleResume = () => {
        if (handle.state !== "pending") return;
        handle.awaitBlocked = false;
        if (!handle.runner) {
          handle.runner = () => {
            if (asyncCtx.kind === "proc") {
              ctx.runProcHandle(handle);
            } else {
              ctx.runFuture(handle);
            }
          };
        }
        ctx.scheduleAsync(handle.runner);
      };
      result.then(
        (value) => {
          state.status = "resolved";
          state.value = value;
          scheduleResume();
        },
        (err) => {
          state.status = "rejected";
          state.error = coerceAsyncError(ctx, err);
          scheduleResume();
        },
      );
      return ctx.procYield(true);
    }
    return result;
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
  const methodReceiver = resolveMethodSetReceiver(funcNode, evalArgs);
  if (methodReceiver) {
    enforceMethodSetConstraints(ctx, funcValue, methodReceiver);
  }
  const procContext = ctx.currentProcContext ? ctx.currentProcContext() : null;
  let callState: FunctionCallState | undefined;
  if (procContext && callNode) {
    const existing = procContext.getFunctionCallState(callNode);
    if (existing && existing.suspended && existing.func === funcValue) {
      callState = existing;
      callState.suspended = false;
    }
  }
  if (!callState && callNode) {
    ctx.inferTypeArgumentsFromCall(funcNode, callNode, evalArgs);
    ctx.enforceGenericConstraintsIfAny(funcNode, callNode);
  }
  let funcEnv: Environment;
  if (callState) {
    funcEnv = callState.env;
  } else {
    funcEnv = new Environment(funcValue.closureEnv);
    if (callNode) {
      ctx.bindTypeArgumentsIfAny(funcNode, callNode, funcEnv);
    }
  }

  if (funcNode.type === "FunctionDefinition") {
    if (
      traceEnabled &&
      funcNode.id?.name === "matches" &&
      evalArgs.length >= 2 &&
      evalArgs[0] &&
      ctx.getTypeNameForValue(evalArgs[0]) === "RaiseErrorMatcher"
    ) {
      const invocation = evalArgs[1]!;
      const invocationType = ctx.getTypeNameForValue(invocation) ?? invocation.kind;
      console.error(`[trace] RaiseErrorMatcher.matches invocation ${invocationType} kind=${invocation.kind}`);
    }
    if (
      traceEnabled &&
      funcNode.id?.name === "matches" &&
      evalArgs.length >= 2 &&
      evalArgs[0] &&
      ctx.getTypeNameForValue(evalArgs[0]) === "ContainMatcher"
    ) {
      console.error("[trace] ContainMatcher.matches invoked");
    }
    if (
      traceEnabled &&
      funcNode.id?.name === "to" &&
      (funcNode as any).structName === "Expectation" &&
      evalArgs.length >= 2
    ) {
      const matcher = evalArgs[1]!;
      const matcherType = ctx.getTypeNameForValue(matcher) ?? matcher.kind;
      const typeArgs = matcher.kind === "struct_instance" ? matcher.typeArguments : undefined;
      const typeArgInfo = typeArgs ? ` typeArgs=${typeArgs.length}` : "";
      console.error(`[trace] Expectation.to matcher ${matcherType} kind=${matcher.kind}${typeArgInfo}`);
    }
    const genericNames = new Set((funcNode.genericParams ?? []).map((gp) => gp.name.name));
    const params = funcNode.params;
    const paramCount = params.length;
    let implicitReceiver: RuntimeValue | null = callState ? callState.implicitReceiver : null;
    let hasImplicit = callState ? callState.hasImplicit : false;
    if (!callState) {
      let bindArgs = evalArgs;
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
          if (
            traceEnabled &&
            funcNode.id?.name === "to" &&
            (funcNode as any).structName === "Expectation" &&
            param.name.name === "matcher"
          ) {
            const boundType = ctx.getTypeNameForValue(coerced) ?? coerced.kind;
            console.error(`[trace] Expectation.to bound matcher ${boundType} kind=${coerced.kind}`);
          }
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
      if (procContext && callNode) {
        callState = {
          func: funcValue,
          env: funcEnv,
          implicitReceiver,
          hasImplicit,
          suspended: false,
        };
        procContext.pushFunctionCallState(callNode, callState);
      }
    }
    let pushedImplicit = false;
    if (hasImplicit && implicitReceiver) {
      ctx.implicitReceiverStack.push(implicitReceiver);
      pushedImplicit = true;
    }
    try {
      const result = ctx.evaluate(funcNode.body, funcEnv);
      if (traceEnabled && funcNode.id?.name === "contain") {
        const resultType = ctx.getTypeNameForValue(result) ?? result.kind;
        console.error(`[trace] contain() result ${resultType} kind=${result.kind}`);
      }
      return coerceReturnValue(ctx, funcNode.returnType, result, genericNames, funcValue.closureEnv);
    } catch (e) {
      if (e instanceof ReturnSignal) {
        return coerceReturnValue(ctx, funcNode.returnType, e.value, genericNames, funcValue.closureEnv);
      }
      if (e instanceof ProcYieldSignal) {
        if (callState) {
          callState.suspended = true;
        }
        throw e;
      }
      throw e;
    } finally {
      if (pushedImplicit) {
        ctx.implicitReceiverStack.pop();
      }
      if (callState && procContext && callNode && !callState.suspended) {
        procContext.popFunctionCallState(callNode);
      }
    }
  }

  if (funcNode.type === "LambdaExpression") {
    const genericNames = new Set((funcNode.genericParams ?? []).map((gp) => gp.name.name));
    const params = funcNode.params;
    let implicitReceiver: RuntimeValue | null = callState ? callState.implicitReceiver : null;
    let hasImplicit = callState ? callState.hasImplicit : false;
    if (!callState) {
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
      if (params.length > 0 && evalArgs.length > 0) {
        implicitReceiver = evalArgs[0]!;
        hasImplicit = true;
      }
      if (procContext && callNode) {
        callState = {
          func: funcValue,
          env: funcEnv,
          implicitReceiver,
          hasImplicit,
          suspended: false,
        };
        procContext.pushFunctionCallState(callNode, callState);
      }
    }
    let pushedImplicit = false;
    if (hasImplicit && implicitReceiver) {
      ctx.implicitReceiverStack.push(implicitReceiver);
      pushedImplicit = true;
    }
    try {
      const result = ctx.evaluate(funcNode.body as AST.AstNode, funcEnv);
      return coerceReturnValue(ctx, funcNode.returnType, result, genericNames, funcValue.closureEnv);
    } catch (e) {
      if (e instanceof ProcYieldSignal) {
        if (callState) {
          callState.suspended = true;
        }
        throw e;
      }
      throw e;
    } finally {
      if (pushedImplicit) {
        ctx.implicitReceiverStack.pop();
      }
      if (callState && procContext && callNode && !callState.suspended) {
        procContext.popFunctionCallState(callNode);
      }
    }
  }

  throw new Error("calling unsupported function declaration");
  } finally {
    if (traceEnabled) {
      callTraceStack.pop();
    }
  }
}

function makePartialFunction(target: RuntimeValue, boundArgs: RuntimeValue[], callNode?: AST.FunctionCall): Extract<RuntimeValue, { kind: "partial_function" }> {
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

function overloadArityRange(overloads: Array<Extract<RuntimeValue, { kind: "function" }>>): { minArgs: number; maxArgs: number } {
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
  ctx: Interpreter,
  overloads: Array<Extract<RuntimeValue, { kind: "function" }>>,
  evalArgs: RuntimeValue[],
  callNode?: AST.FunctionCall,
): Extract<RuntimeValue, { kind: "function" }> | null {
  const candidates: Array<{
    fn: Extract<RuntimeValue, { kind: "function" }>;
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
        const paramType = substituteSelfType(param.paramType, fn);
        if (!ctx.matchesType(paramType, arg)) {
          compatible = false;
          break;
        }
        score += parameterSpecificity(paramType, generics);
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

function substituteSelfType(typeExpr: AST.TypeExpression, fn: Extract<RuntimeValue, { kind: "function" }>): AST.TypeExpression {
  const target = (fn as any).methodSetTargetType as AST.TypeExpression | undefined;
  if (!target) return typeExpr;
  return replaceSelfTypeExpr(typeExpr, target);
}

function replaceSelfTypeExpr(typeExpr: AST.TypeExpression, replacement: AST.TypeExpression): AST.TypeExpression {
  switch (typeExpr.type) {
    case "SimpleTypeExpression":
      if (typeExpr.name.name === "Self") return replacement;
      return typeExpr;
    case "GenericTypeExpression": {
      const base = replaceSelfTypeExpr(typeExpr.base, replacement);
      const args = (typeExpr.arguments ?? []).map((arg) => (arg ? replaceSelfTypeExpr(arg, replacement) : arg));
      if (base === typeExpr.base && args.every((arg, idx) => arg === (typeExpr.arguments ?? [])[idx])) {
        return typeExpr;
      }
      return AST.genericTypeExpression(base, args);
    }
    case "NullableTypeExpression": {
      const inner = replaceSelfTypeExpr(typeExpr.innerType, replacement);
      if (inner === typeExpr.innerType) return typeExpr;
      return AST.nullableTypeExpression(inner);
    }
    case "ResultTypeExpression": {
      const inner = replaceSelfTypeExpr(typeExpr.innerType, replacement);
      if (inner === typeExpr.innerType) return typeExpr;
      return AST.resultTypeExpression(inner);
    }
    case "UnionTypeExpression": {
      const members = typeExpr.members.map((member) => replaceSelfTypeExpr(member, replacement));
      if (members.every((member, idx) => member === typeExpr.members[idx])) return typeExpr;
      return AST.unionTypeExpression(members);
    }
    case "FunctionTypeExpression": {
      const params = (typeExpr.paramTypes ?? []).map((param) => replaceSelfTypeExpr(param, replacement));
      const ret = replaceSelfTypeExpr(typeExpr.returnType, replacement);
      if (ret === typeExpr.returnType && params.every((param, idx) => param === (typeExpr.paramTypes ?? [])[idx])) {
        return typeExpr;
      }
      return AST.functionTypeExpression(params, ret);
    }
    default:
      return typeExpr;
  }
}

function resolveMethodSetReceiver(
  funcNode: AST.FunctionDefinition | AST.LambdaExpression,
  evalArgs: RuntimeValue[],
): RuntimeValue | null {
  if (funcNode.type !== "FunctionDefinition") return null;
  if (funcNode.isMethodShorthand) {
    return evalArgs[0] ?? null;
  }
  const params = funcNode.params ?? [];
  if (params.length === 0) return null;
  const first = params[0];
  if (!first) return null;
  if (first.name?.type === "Identifier" && first.name.name.toLowerCase() === "self") {
    return evalArgs[0] ?? null;
  }
  if (first.paramType?.type === "SimpleTypeExpression" && first.paramType.name.name === "Self") {
    return evalArgs[0] ?? null;
  }
  return null;
}

function enforceMethodSetConstraints(
  ctx: Interpreter,
  funcValue: Extract<RuntimeValue, { kind: "function" }>,
  receiver: RuntimeValue,
): void {
  const constraints = (funcValue as any).methodSetConstraints as ConstraintSpec[] | undefined;
  if (!constraints || constraints.length === 0) return;
  const targetType = (funcValue as any).methodSetTargetType as AST.TypeExpression | undefined;
  const genericParams = (funcValue as any).methodSetGenericParams as AST.GenericParameter[] | undefined;
  const actualTypeExpr = ctx.typeExpressionFromValue(receiver);
  if (!actualTypeExpr) return;
  const bindings = new Map<string, AST.TypeExpression>();
  const genericNames = new Set((genericParams ?? []).map((gp) => gp.name.name));
  if (targetType && genericNames.size > 0) {
    const canonicalTarget = ctx.expandTypeAliases(targetType);
    const canonicalActual = ctx.expandTypeAliases(actualTypeExpr);
    ctx.matchTypeExpressionTemplate(canonicalTarget, canonicalActual, genericNames, bindings);
    bindings.set("Self", canonicalActual);
  } else {
    bindings.set("Self", ctx.expandTypeAliases(actualTypeExpr));
  }
  if (receiver.kind === "struct_instance" && receiver.typeArgMap) {
    for (const [key, value] of receiver.typeArgMap.entries()) {
      if (!bindings.has(key)) bindings.set(key, value);
    }
  }
  ctx.enforceConstraintSpecs(constraints, bindings, "method set");
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
