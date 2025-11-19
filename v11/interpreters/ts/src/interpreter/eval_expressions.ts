import * as AST from "../ast";
import type { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import { evaluateAssignmentExpression } from "./assignments";
import {
  evaluateBlockExpression,
  evaluateBreakStatement,
  evaluateContinueStatement,
  evaluateForLoop,
  evaluateIfExpression,
  evaluateLoopExpression,
  evaluateReturnStatement,
  evaluateWhileLoop,
} from "./control_flow";
import { evaluateEnsureExpression, evaluateOrElseExpression, evaluatePropagationExpression, evaluateRescueExpression, evaluateRaiseStatement, evaluateRethrowStatement } from "./error_handling";
import { callCallableValue, evaluateFunctionCall, evaluateFunctionDefinition, evaluateLambdaExpression } from "./functions";
import { evaluateDynImportStatement, evaluateImportStatement, evaluateModule, evaluatePackageStatement } from "./imports";
import { evaluateImplementationDefinition, evaluateInterfaceDefinition, evaluateMethodsDefinition, evaluateUnionDefinition } from "./definitions";
import { evaluateMatchExpression } from "./match";
import { evaluateMemberAccessExpression, evaluateStructDefinition, evaluateStructLiteral, evaluateImplicitMemberExpression, memberAccessOnValue } from "./structs";
import { evaluateLiteral } from "./literals";
import { evaluateMapLiteral } from "./maps";
import { evaluateIndexExpression, evaluateRangeExpression, evaluateBinaryExpression, evaluateUnaryExpression, evaluateTopicReferenceExpression } from "./operations";
import { evaluateProcExpression, evaluateSpawnExpression, evaluateBreakpointExpression, evaluateStringInterpolation } from "./runtime_extras";
import { evaluateIteratorLiteral, evaluateYieldStatement } from "./iterators";
import { ProcContinuationContext } from "./proc_continuations";
import type { V10Value } from "./values";

declare module "./index" {
  interface InterpreterV10 {
    evaluate(node: AST.AstNode | null, env?: Environment): V10Value;
  }
}

const NIL: V10Value = { kind: "nil", value: null };
const EXPRESSION_TYPES = new Set<AST.AstNode["type"]>([
  "Identifier",
  "StringLiteral",
  "BooleanLiteral",
  "CharLiteral",
  "NilLiteral",
  "FloatLiteral",
  "IntegerLiteral",
  "ArrayLiteral",
  "UnaryExpression",
  "BinaryExpression",
  "FunctionCall",
  "BlockExpression",
  "AssignmentExpression",
  "RangeExpression",
  "StringInterpolation",
  "MemberAccessExpression",
  "IndexExpression",
  "LambdaExpression",
  "ProcExpression",
  "SpawnExpression",
  "AwaitExpression",
  "PropagationExpression",
  "OrElseExpression",
  "BreakpointExpression",
  "IteratorLiteral",
  "ImplicitMemberExpression",
  "PlaceholderExpression",
  "TopicReferenceExpression",
  "IfExpression",
  "MatchExpression",
  "StructLiteral",
  "MapLiteral",
  "RescueExpression",
  "EnsureExpression",
]);

export function applyEvaluationAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.evaluate = function evaluate(this: InterpreterV10, node: AST.AstNode | null, env: Environment = this.globals): V10Value {
    if (!node) return NIL;
    if (EXPRESSION_TYPES.has(node.type as AST.AstNode["type"]) && !this.hasPlaceholderFrame()) {
      const placeholderFn = this.tryBuildPlaceholderFunction(node as AST.Expression, env);
      if (placeholderFn) {
        return placeholderFn;
      }
    }
    switch (node.type) {
      case "StringLiteral":
      case "BooleanLiteral":
      case "CharLiteral":
      case "NilLiteral":
      case "FloatLiteral":
      case "IntegerLiteral":
      case "ArrayLiteral":
        return evaluateLiteral(this, node, env);
      case "UnaryExpression":
        return evaluateUnaryExpression(this, node as AST.UnaryExpression, env);
      case "BinaryExpression":
        return evaluateBinaryExpression(this, node as AST.BinaryExpression, env);
      case "RangeExpression":
        return evaluateRangeExpression(this, node as AST.RangeExpression, env);
      case "IndexExpression":
        return evaluateIndexExpression(this, node as AST.IndexExpression, env);
      case "IfExpression":
        return evaluateIfExpression(this, node as AST.IfExpression, env);
      case "WhileLoop":
        return evaluateWhileLoop(this, node as AST.WhileLoop, env);
      case "LoopExpression":
        return evaluateLoopExpression(this, node as AST.LoopExpression, env);
      case "BreakStatement":
        evaluateBreakStatement(this, node as AST.BreakStatement, env);
        return NIL;
      case "ContinueStatement":
        evaluateContinueStatement(this, node as AST.ContinueStatement);
        return NIL;
      case "ForLoop":
        return evaluateForLoop(this, node as AST.ForLoop, env);
      case "MatchExpression":
        return evaluateMatchExpression(this, node as AST.MatchExpression, env);
      case "Identifier":
        return env.get((node as AST.Identifier).name);
      case "BlockExpression":
        return evaluateBlockExpression(this, node as AST.BlockExpression, env);
      case "ReturnStatement":
        evaluateReturnStatement(this, node as AST.ReturnStatement, env);
        return NIL;
      case "AssignmentExpression":
        return evaluateAssignmentExpression(this, node as AST.AssignmentExpression, env);
      case "FunctionDefinition":
        return evaluateFunctionDefinition(this, node as AST.FunctionDefinition, env);
      case "LambdaExpression":
        return evaluateLambdaExpression(this, node as AST.LambdaExpression, env);
      case "FunctionCall":
        return evaluateFunctionCall(this, node as AST.FunctionCall, env);
      case "StructDefinition":
        return evaluateStructDefinition(this, node as AST.StructDefinition, env);
      case "StructLiteral":
        return evaluateStructLiteral(this, node as AST.StructLiteral, env);
      case "MapLiteral":
        return evaluateMapLiteral(this, node as AST.MapLiteral, env);
      case "MemberAccessExpression":
        return evaluateMemberAccessExpression(this, node as AST.MemberAccessExpression, env);
      case "ImplicitMemberExpression":
        return evaluateImplicitMemberExpression(this, node as AST.ImplicitMemberExpression, env);
      case "PlaceholderExpression":
        return this.evaluatePlaceholderExpression(node as AST.PlaceholderExpression, env);
      case "StringInterpolation":
        return evaluateStringInterpolation(this, node as AST.StringInterpolation, env);
      case "BreakpointExpression":
        return evaluateBreakpointExpression(this, node as AST.BreakpointExpression, env);
      case "RaiseStatement":
        evaluateRaiseStatement(this, node as AST.RaiseStatement, env);
        return NIL;
      case "RescueExpression":
        return evaluateRescueExpression(this, node as AST.RescueExpression, env);
      case "OrElseExpression":
        return evaluateOrElseExpression(this, node as AST.OrElseExpression, env);
      case "PropagationExpression":
        return evaluatePropagationExpression(this, node as AST.PropagationExpression, env);
      case "EnsureExpression":
        return evaluateEnsureExpression(this, node as AST.EnsureExpression, env);
      case "RethrowStatement":
        evaluateRethrowStatement(this, node as AST.RethrowStatement);
        return NIL;
      case "ProcExpression":
        return evaluateProcExpression(this, node as AST.ProcExpression, env);
      case "SpawnExpression":
        return evaluateSpawnExpression(this, node as AST.SpawnExpression, env);
      case "AwaitExpression":
        return evaluateAwaitExpression(this, node as AST.AwaitExpression, env);
      case "IteratorLiteral":
        return evaluateIteratorLiteral(this, node as AST.IteratorLiteral, env);
      case "YieldStatement":
        return evaluateYieldStatement(this, node as AST.YieldStatement, env);
      case "TopicReferenceExpression":
        return evaluateTopicReferenceExpression(this);
      case "Module":
        return evaluateModule(this, node as AST.Module);
      case "PackageStatement":
        return evaluatePackageStatement();
      case "ImportStatement":
        return evaluateImportStatement(this, node as AST.ImportStatement, env);
      case "DynImportStatement":
        return evaluateDynImportStatement(this, node as AST.DynImportStatement, env);
      case "InterfaceDefinition":
        return evaluateInterfaceDefinition(this, node as AST.InterfaceDefinition, env);
      case "UnionDefinition":
        return evaluateUnionDefinition(this, node as AST.UnionDefinition, env);
      case "TypeAliasDefinition":
        return NIL;
      case "MethodsDefinition":
        return evaluateMethodsDefinition(this, node as AST.MethodsDefinition, env);
      case "ImplementationDefinition":
        return evaluateImplementationDefinition(this, node as AST.ImplementationDefinition, env);
      default:
        throw new Error(`Not implemented in milestone: ${node.type}`);
    }
  };
}

type AwaitArmState = {
  awaitable: V10Value;
  isDefault: boolean;
  registration?: V10Value;
};

type AwaitEvaluationState = {
  env: Environment;
  arms: AwaitArmState[];
  defaultArm?: AwaitArmState;
  waiting: boolean;
  wakePending: boolean;
  waker?: Extract<V10Value, { kind: "struct_instance" }>;
};

function evaluateAwaitExpression(ctx: InterpreterV10, node: AST.AwaitExpression, env: Environment): V10Value {
  const asyncCtx = ctx.currentAsyncContext();
  if (!asyncCtx || asyncCtx.kind !== "proc") {
    throw new Error("await expressions must run inside a proc");
  }
  const procContext = ctx.currentProcContext();
  if (!procContext) {
    throw new Error("await expressions require a proc continuation context");
  }

  let state = procContext.getAwaitState(node) as AwaitEvaluationState | undefined;
  if (!state) {
    state = initializeAwaitState(ctx, node, env, asyncCtx.handle);
    procContext.setAwaitState(node, state);
  }

  const winner = selectReadyAwaitArm(ctx, state);
  if (winner) {
    return completeAwait(ctx, procContext, node, state, winner, asyncCtx.handle);
  }

  if (state.defaultArm) {
    return completeAwait(ctx, procContext, node, state, state.defaultArm, asyncCtx.handle);
  }

  if (asyncCtx.handle.cancelRequested) {
    cleanupAwaitState(ctx, procContext, node, state, asyncCtx.handle);
    throw new Error("Proc cancelled during await");
  }

  if (!state.waiting) {
    registerAwaitState(ctx, state);
  }
  asyncCtx.handle.awaitBlocked = true;
  ctx.procYield();
  return { kind: "nil", value: null };
}

function initializeAwaitState(
  ctx: InterpreterV10,
  node: AST.AwaitExpression,
  env: Environment,
  handle: Extract<V10Value, { kind: "proc_handle" }>,
): AwaitEvaluationState {
  const iterableValue = ctx.evaluate(node.expression, env);
  const arms = collectAwaitArms(ctx, iterableValue, env);
  if (arms.length === 0) {
    throw new Error("await requires at least one arm");
  }
  const defaultArms = arms.filter((arm) => arm.isDefault);
  if (defaultArms.length > 1) {
    throw new Error("await accepts at most one default arm");
  }
  const state: AwaitEvaluationState = {
    env,
    arms,
    defaultArm: defaultArms[0],
    waiting: false,
    wakePending: false,
  };
  state.waker = ctx.createAwaitWaker(handle, state);
  return state;
}

function collectAwaitArms(ctx: InterpreterV10, iterable: V10Value, env: Environment): AwaitArmState[] {
  if (iterable.kind !== "array") {
    throw new Error("await currently expects an array of Awaitable values");
  }
  return iterable.elements.map((value) => ({
    awaitable: value,
    isDefault: checkAwaitArmIsDefault(ctx, value, env),
  }));
}

function checkAwaitArmIsDefault(ctx: InterpreterV10, awaitable: V10Value, env: Environment): boolean {
  try {
    const member = memberAccessOnValue(ctx, awaitable, AST.identifier("is_default"), env);
    const result = callCallableValue(ctx, member, [], env);
    return ctx.isTruthy(result);
  } catch {
    return false;
  }
}

function selectReadyAwaitArm(ctx: InterpreterV10, state: AwaitEvaluationState): AwaitArmState | undefined {
  const ready: AwaitArmState[] = [];
  for (const arm of state.arms) {
    if (arm.isDefault) continue;
    const result = invokeAwaitableMethod(ctx, arm.awaitable, "is_ready", [], state.env);
    if (ctx.isTruthy(result)) {
      ready.push(arm);
    }
  }
  if (ready.length === 0) return undefined;
  const start = ready.length > 0 ? ctx.awaitRoundRobinIndex % ready.length : 0;
  ctx.awaitRoundRobinIndex = (ctx.awaitRoundRobinIndex + 1) % ready.length;
  return ready[start];
}

function registerAwaitState(ctx: InterpreterV10, state: AwaitEvaluationState): void {
  const waker = state.waker;
  if (!waker) return;
  for (const arm of state.arms) {
    if (arm.isDefault) continue;
    if (arm.registration) continue;
    const registration = invokeAwaitableMethod(ctx, arm.awaitable, "register", [waker], state.env);
    arm.registration = registration;
  }
  state.waiting = true;
  state.wakePending = false;
}

function cleanupAwaitState(
  ctx: InterpreterV10,
  procContext: ProcContinuationContext,
  node: AST.AwaitExpression,
  state: AwaitEvaluationState,
  handle?: Extract<V10Value, { kind: "proc_handle" }>,
): void {
  procContext.clearAwaitState(node);
  for (const arm of state.arms) {
    if (arm.registration) {
      cancelAwaitRegistration(ctx, arm.registration, state.env);
      arm.registration = undefined;
    }
  }
  state.waiting = false;
  if (handle) {
    handle.awaitBlocked = false;
  }
}

function completeAwait(
  ctx: InterpreterV10,
  procContext: ProcContinuationContext,
  node: AST.AwaitExpression,
  state: AwaitEvaluationState,
  winner: AwaitArmState,
  handle: Extract<V10Value, { kind: "proc_handle" }>,
): V10Value {
  for (const arm of state.arms) {
    if (arm === winner) continue;
    if (arm.registration) {
      cancelAwaitRegistration(ctx, arm.registration, state.env);
      arm.registration = undefined;
    }
  }
  const result = invokeAwaitableMethod(ctx, winner.awaitable, "commit", [], state.env);
  cleanupAwaitState(ctx, procContext, node, state, handle);
  handle.awaitBlocked = false;
  return result;
}

function cancelAwaitRegistration(ctx: InterpreterV10, registration: V10Value, env: Environment): void {
  try {
    const member = memberAccessOnValue(ctx, registration, AST.identifier("cancel"), env);
    callCallableValue(ctx, member, [], env);
  } catch {
    // ignore cancellation errors
  }
}

function invokeAwaitableMethod(
  ctx: InterpreterV10,
  awaitable: V10Value,
  methodName: string,
  args: V10Value[],
  env: Environment,
): V10Value {
  const member = memberAccessOnValue(ctx, awaitable, AST.identifier(methodName), env);
  return callCallableValue(ctx, member, args, env);
}
