import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { ReturnSignal } from "./signals";

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

export function evaluateFunctionCall(ctx: InterpreterV10, node: AST.FunctionCall, env: Environment): V10Value {
  const calleeEvaluated = ctx.evaluate(node.callee, env);
  let funcValue: Extract<V10Value, { kind: "function" }> | null = null;
  let nativeFunc: Extract<V10Value, { kind: "native_function" }> | null = null;
  let injectedArgs: V10Value[] = [];

  if (calleeEvaluated.kind === "bound_method") {
    funcValue = calleeEvaluated.func;
    injectedArgs = [calleeEvaluated.self];
  } else if (calleeEvaluated.kind === "function") {
    funcValue = calleeEvaluated;
  } else if (calleeEvaluated.kind === "dyn_ref") {
    const bucket = ctx.packageRegistry.get(calleeEvaluated.pkg);
    const sym = bucket?.get(calleeEvaluated.name);
    if (!sym || sym.kind !== "function") throw new Error(`dyn ref '${calleeEvaluated.pkg}.${calleeEvaluated.name}' is not callable`);
    funcValue = sym;
  } else if (calleeEvaluated.kind === "native_bound_method") {
    nativeFunc = calleeEvaluated.func;
    injectedArgs = [calleeEvaluated.self];
  } else if (calleeEvaluated.kind === "native_function") {
    nativeFunc = calleeEvaluated;
  } else {
    throw new Error("Cannot call non-function");
  }

  const callArgs = node.arguments.map(arg => ctx.evaluate(arg, env));
  if (nativeFunc) {
    const evalArgs = [...injectedArgs, ...callArgs];
    if (evalArgs.length !== nativeFunc.arity) {
      throw new Error(`Arity mismatch calling ${nativeFunc.name}: expected ${nativeFunc.arity}, got ${evalArgs.length}`);
    }
    return nativeFunc.impl(ctx, evalArgs);
  }

  if (!funcValue) throw new Error("Callable target missing function value");
  const funcNode = funcValue.node;
  ctx.enforceGenericConstraintsIfAny(funcNode, node);
  const funcEnv = new Environment(funcValue.closureEnv);
  ctx.bindTypeArgumentsIfAny(funcNode, node, funcEnv);
  const params = funcNode.type === "FunctionDefinition" ? funcNode.params : funcNode.params;
  const evalArgs: V10Value[] = [...injectedArgs, ...callArgs];
  if (evalArgs.length !== params.length) {
    const name = (funcNode as any).id?.name ?? "(lambda)";
    throw new Error(`Arity mismatch calling ${name}: expected ${params.length}, got ${evalArgs.length}`);
  }
  for (let i = 0; i < params.length; i++) {
    const param = params[i];
    const argVal = evalArgs[i];
    if (!param) throw new Error(`Parameter missing at index ${i}`);
    if (argVal === undefined) throw new Error(`Argument missing at index ${i}`);
    if (param.paramType) {
      if (!ctx.matchesType(param.paramType, argVal)) {
        const pname = (param.name as any).name ?? `param_${i}`;
        throw new Error(`Parameter type mismatch for '${pname}'`);
      }
    }
    const coerced = param.paramType ? ctx.coerceValueToType(param.paramType, argVal) : argVal;
    evalArgs[i] = coerced;
    if (param.name.type === "Identifier") {
      funcEnv.define(param.name.name, coerced);
    } else if (param.name.type === "WildcardPattern") {
      // ignore
    } else if (param.name.type === "StructPattern" || param.name.type === "ArrayPattern" || param.name.type === "LiteralPattern") {
      ctx.assignByPattern(param.name as AST.Pattern, coerced, funcEnv, true);
    } else {
      throw new Error("Only simple identifier and destructuring params supported for now");
    }
  }

  try {
    if (funcNode.type === "FunctionDefinition") {
      return ctx.evaluate(funcNode.body, funcEnv);
    }
    return ctx.evaluate(funcNode.body as AST.AstNode, funcEnv);
  } catch (e) {
    if (e instanceof ReturnSignal) return e.value;
    throw e;
  }
}
