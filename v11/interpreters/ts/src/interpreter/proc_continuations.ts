import * as AST from "../ast";
import { Environment } from "./environment";
import type { RuntimeValue } from "./values";
import type {
  ContinuationContext,
  BlockState,
  ModuleState,
  ForLoopState,
  WhileLoopState,
  LoopExpressionState,
  IfExpressionState,
  MatchExpressionState,
  StringInterpolationState,
  EnsureState,
} from "./continuations";

export class ProcContinuationContext implements ContinuationContext {
  readonly kind = "proc" as const;
  private repeatCurrentStatement = false;

  private blockStates: WeakMap<AST.BlockExpression, BlockState[]> = new WeakMap();
  private moduleStates: WeakMap<AST.Module, ModuleState[]> = new WeakMap();
  private forLoopStates: WeakMap<AST.ForLoop, ForLoopState[]> = new WeakMap();
  private whileLoopStates: WeakMap<AST.WhileLoop, WhileLoopState[]> = new WeakMap();
  private loopStates: WeakMap<AST.LoopExpression, LoopExpressionState[]> = new WeakMap();
  private ifStates: WeakMap<AST.IfExpression, IfExpressionState[]> = new WeakMap();
  private matchStates: WeakMap<AST.MatchExpression, MatchExpressionState[]> = new WeakMap();
  private stringInterpolationStates: WeakMap<AST.StringInterpolation, StringInterpolationState[]> = new WeakMap();
  private ensureStates: WeakMap<AST.EnsureExpression, EnsureState[]> = new WeakMap();
  private awaitStates: WeakMap<AST.AwaitExpression, unknown> = new WeakMap();
  private nativeCallStates: WeakMap<AST.FunctionCall, NativeCallState> = new WeakMap();
  private functionCallStates: WeakMap<AST.FunctionCall, FunctionCallState[]> = new WeakMap();

  markStatementIncomplete(): void {
    this.repeatCurrentStatement = true;
  }

  consumeRepeatCurrentStatement(): boolean {
    const repeat = this.repeatCurrentStatement;
    this.repeatCurrentStatement = false;
    return repeat;
  }

  getBlockState(node: AST.BlockExpression): BlockState | undefined {
    const stack = this.blockStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setBlockState(node: AST.BlockExpression, state: BlockState): void {
    const stack = this.blockStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.blockStates.set(node, [state]);
    }
  }

  clearBlockState(node: AST.BlockExpression): void {
    const stack = this.blockStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.blockStates.delete(node);
  }

  getModuleState(node: AST.Module): ModuleState | undefined {
    const stack = this.moduleStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setModuleState(node: AST.Module, state: ModuleState): void {
    const stack = this.moduleStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.moduleStates.set(node, [state]);
    }
  }

  clearModuleState(node: AST.Module): void {
    const stack = this.moduleStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.moduleStates.delete(node);
  }

  getForLoopState(node: AST.ForLoop): ForLoopState | undefined {
    const stack = this.forLoopStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setForLoopState(node: AST.ForLoop, state: ForLoopState): void {
    const stack = this.forLoopStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.forLoopStates.set(node, [state]);
    }
  }

  clearForLoopState(node: AST.ForLoop): void {
    const stack = this.forLoopStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.forLoopStates.delete(node);
  }

  getWhileLoopState(node: AST.WhileLoop): WhileLoopState | undefined {
    const stack = this.whileLoopStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setWhileLoopState(node: AST.WhileLoop, state: WhileLoopState): void {
    const stack = this.whileLoopStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.whileLoopStates.set(node, [state]);
    }
  }

  clearWhileLoopState(node: AST.WhileLoop): void {
    const stack = this.whileLoopStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.whileLoopStates.delete(node);
  }

  getLoopExpressionState(node: AST.LoopExpression): LoopExpressionState | undefined {
    const stack = this.loopStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setLoopExpressionState(node: AST.LoopExpression, state: LoopExpressionState): void {
    const stack = this.loopStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.loopStates.set(node, [state]);
    }
  }

  clearLoopExpressionState(node: AST.LoopExpression): void {
    const stack = this.loopStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.loopStates.delete(node);
  }

  getIfState(node: AST.IfExpression): IfExpressionState | undefined {
    const stack = this.ifStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setIfState(node: AST.IfExpression, state: IfExpressionState): void {
    const stack = this.ifStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.ifStates.set(node, [state]);
    }
  }

  clearIfState(node: AST.IfExpression): void {
    const stack = this.ifStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.ifStates.delete(node);
  }

  getMatchState(node: AST.MatchExpression): MatchExpressionState | undefined {
    const stack = this.matchStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setMatchState(node: AST.MatchExpression, state: MatchExpressionState): void {
    const stack = this.matchStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.matchStates.set(node, [state]);
    }
  }

  clearMatchState(node: AST.MatchExpression): void {
    const stack = this.matchStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.matchStates.delete(node);
  }

  getStringInterpolationState(node: AST.StringInterpolation): StringInterpolationState | undefined {
    const stack = this.stringInterpolationStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setStringInterpolationState(node: AST.StringInterpolation, state: StringInterpolationState): void {
    const stack = this.stringInterpolationStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.stringInterpolationStates.set(node, [state]);
    }
  }

  clearStringInterpolationState(node: AST.StringInterpolation): void {
    const stack = this.stringInterpolationStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.stringInterpolationStates.delete(node);
  }

  getEnsureState(node: AST.EnsureExpression): EnsureState | undefined {
    const stack = this.ensureStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  setEnsureState(node: AST.EnsureExpression, state: EnsureState): void {
    const stack = this.ensureStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.ensureStates.set(node, [state]);
    }
  }

  clearEnsureState(node: AST.EnsureExpression): void {
    const stack = this.ensureStates.get(node);
    if (!stack) return;
    stack.pop();
    if (stack.length === 0) this.ensureStates.delete(node);
  }

  getAwaitState(node: AST.AwaitExpression): unknown {
    return this.awaitStates.get(node);
  }

  setAwaitState(node: AST.AwaitExpression, state: unknown): void {
    this.awaitStates.set(node, state);
  }

  clearAwaitState(node: AST.AwaitExpression): void {
    this.awaitStates.delete(node);
  }

  getNativeCallState(node: AST.FunctionCall): NativeCallState | undefined {
    return this.nativeCallStates.get(node);
  }

  setNativeCallState(node: AST.FunctionCall, state: NativeCallState): void {
    this.nativeCallStates.set(node, state);
  }

  clearNativeCallState(node: AST.FunctionCall): void {
    this.nativeCallStates.delete(node);
  }

  getFunctionCallState(node: AST.FunctionCall): FunctionCallState | undefined {
    const stack = this.functionCallStates.get(node);
    return stack ? stack[stack.length - 1] : undefined;
  }

  pushFunctionCallState(node: AST.FunctionCall, state: FunctionCallState): void {
    const stack = this.functionCallStates.get(node);
    if (stack) {
      stack.push(state);
    } else {
      this.functionCallStates.set(node, [state]);
    }
  }

  popFunctionCallState(node: AST.FunctionCall): FunctionCallState | undefined {
    const stack = this.functionCallStates.get(node);
    if (!stack) return undefined;
    const state = stack.pop();
    if (stack.length === 0) {
      this.functionCallStates.delete(node);
    }
    return state;
  }

  reset(): void {
    this.repeatCurrentStatement = false;
    this.blockStates = new WeakMap();
    this.moduleStates = new WeakMap();
    this.forLoopStates = new WeakMap();
    this.whileLoopStates = new WeakMap();
    this.loopStates = new WeakMap();
    this.ifStates = new WeakMap();
    this.matchStates = new WeakMap();
    this.stringInterpolationStates = new WeakMap();
    this.ensureStates = new WeakMap();
    this.awaitStates = new WeakMap();
    this.nativeCallStates = new WeakMap();
    this.functionCallStates = new WeakMap();
  }
}

export type NativeCallState = {
  status: "pending" | "resolved" | "rejected";
  value?: RuntimeValue;
  error?: RuntimeValue;
};

export type FunctionCallState = {
  func: Extract<RuntimeValue, { kind: "function" }>;
  env: Environment;
  implicitReceiver: RuntimeValue | null;
  hasImplicit: boolean;
  suspended: boolean;
};
