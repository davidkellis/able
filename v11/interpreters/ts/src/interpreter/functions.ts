import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";
import { ProcYieldSignal, RaiseSignal, ReturnSignal } from "./signals";
import type { FunctionCallState, NativeCallState } from "./proc_continuations";
import { memberAccessOnValue } from "./structs";
import { collectTypeDispatches } from "./type-dispatch";
import { attachRuntimeDiagnosticContext, getRuntimeDiagnosticContext } from "./runtime_diagnostics";
import {
  enforceMethodSetConstraints,
  functionArityRange,
  makePartialFunction,
  overloadArityRange,
  parameterSpecificity,
  resolveMethodSetReceiver,
  selectRuntimeOverload,
  substituteSelfType,
} from "./functions_overloads";

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

function canonicalizeExpandedTypeExpression(ctx: Interpreter, env: Environment, expr: AST.TypeExpression): AST.TypeExpression {
  switch (expr.type) {
    case "SimpleTypeExpression": {
      const name = expr.name.name;
      let binding: RuntimeValue | null = null;
      try {
        binding = env.get(name);
      } catch {}
      if (!binding) return expr;
      if (binding.kind === "struct_def" || binding.kind === "interface_def" || binding.kind === "union_def") {
        const canonical = binding.def.id.name;
        if (canonical && canonical !== name) {
          return AST.simpleTypeExpression(canonical);
        }
      }
      return expr;
    }
    case "GenericTypeExpression":
      return AST.genericTypeExpression(
        canonicalizeExpandedTypeExpression(ctx, env, expr.base),
        (expr.arguments ?? []).map((arg) => (arg ? canonicalizeExpandedTypeExpression(ctx, env, arg) : arg)),
      );
    case "NullableTypeExpression":
      return AST.nullableTypeExpression(canonicalizeExpandedTypeExpression(ctx, env, expr.innerType));
    case "ResultTypeExpression":
      return AST.resultTypeExpression(canonicalizeExpandedTypeExpression(ctx, env, expr.innerType));
    case "UnionTypeExpression":
      return AST.unionTypeExpression((expr.members ?? []).map((member) => canonicalizeExpandedTypeExpression(ctx, env, member)));
    case "FunctionTypeExpression":
      return AST.functionTypeExpression(
        (expr.paramTypes ?? []).map((param) => canonicalizeExpandedTypeExpression(ctx, env, param)),
        canonicalizeExpandedTypeExpression(ctx, env, expr.returnType),
      );
    default:
      return expr;
  }
}

function canonicalizeTypeExpression(ctx: Interpreter, env: Environment, expr: AST.TypeExpression): AST.TypeExpression {
  const expanded = ctx.expandTypeAliases(expr);
  return canonicalizeExpandedTypeExpression(ctx, env, expanded);
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
  let runtimeFramePushed = false;
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
  if (callNode) {
    ctx.runtimeCallStack.push({ node: callNode });
    runtimeFramePushed = true;
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
        try {
          return coerceReturnValue(ctx, funcNode.returnType, e.value, genericNames, funcValue.closureEnv);
        } catch (err) {
          const context = getRuntimeDiagnosticContext(e);
          if (context) {
            attachRuntimeDiagnosticContext(err, context);
          }
          throw err;
        }
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
    if (runtimeFramePushed) {
      ctx.runtimeCallStack.pop();
    }
    if (traceEnabled) {
      callTraceStack.pop();
    }
  }
}
