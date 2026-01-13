import * as AST from "../ast";
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
} from "./continuations";

export class ProcContinuationContext implements ContinuationContext {
  readonly kind = "proc" as const;
  private repeatCurrentStatement = false;

  private blockStates: WeakMap<AST.BlockExpression, BlockState> = new WeakMap();
  private moduleStates: WeakMap<AST.Module, ModuleState> = new WeakMap();
  private forLoopStates: WeakMap<AST.ForLoop, ForLoopState> = new WeakMap();
  private whileLoopStates: WeakMap<AST.WhileLoop, WhileLoopState> = new WeakMap();
  private loopStates: WeakMap<AST.LoopExpression, LoopExpressionState> = new WeakMap();
  private ifStates: WeakMap<AST.IfExpression, IfExpressionState> = new WeakMap();
  private matchStates: WeakMap<AST.MatchExpression, MatchExpressionState> = new WeakMap();
  private stringInterpolationStates: WeakMap<AST.StringInterpolation, StringInterpolationState> = new WeakMap();
  private awaitStates: WeakMap<AST.AwaitExpression, unknown> = new WeakMap();
  private nativeCallStates: WeakMap<AST.FunctionCall, NativeCallState> = new WeakMap();

  markStatementIncomplete(): void {
    this.repeatCurrentStatement = true;
  }

  consumeRepeatCurrentStatement(): boolean {
    const repeat = this.repeatCurrentStatement;
    this.repeatCurrentStatement = false;
    return repeat;
  }

  getBlockState(node: AST.BlockExpression): BlockState | undefined {
    return this.blockStates.get(node);
  }

  setBlockState(node: AST.BlockExpression, state: BlockState): void {
    this.blockStates.set(node, state);
  }

  clearBlockState(node: AST.BlockExpression): void {
    this.blockStates.delete(node);
  }

  getModuleState(node: AST.Module): ModuleState | undefined {
    return this.moduleStates.get(node);
  }

  setModuleState(node: AST.Module, state: ModuleState): void {
    this.moduleStates.set(node, state);
  }

  clearModuleState(node: AST.Module): void {
    this.moduleStates.delete(node);
  }

  getForLoopState(node: AST.ForLoop): ForLoopState | undefined {
    return this.forLoopStates.get(node);
  }

  setForLoopState(node: AST.ForLoop, state: ForLoopState): void {
    this.forLoopStates.set(node, state);
  }

  clearForLoopState(node: AST.ForLoop): void {
    this.forLoopStates.delete(node);
  }

  getWhileLoopState(node: AST.WhileLoop): WhileLoopState | undefined {
    return this.whileLoopStates.get(node);
  }

  setWhileLoopState(node: AST.WhileLoop, state: WhileLoopState): void {
    this.whileLoopStates.set(node, state);
  }

  clearWhileLoopState(node: AST.WhileLoop): void {
    this.whileLoopStates.delete(node);
  }

  getLoopExpressionState(node: AST.LoopExpression): LoopExpressionState | undefined {
    return this.loopStates.get(node);
  }

  setLoopExpressionState(node: AST.LoopExpression, state: LoopExpressionState): void {
    this.loopStates.set(node, state);
  }

  clearLoopExpressionState(node: AST.LoopExpression): void {
    this.loopStates.delete(node);
  }

  getIfState(node: AST.IfExpression): IfExpressionState | undefined {
    return this.ifStates.get(node);
  }

  setIfState(node: AST.IfExpression, state: IfExpressionState): void {
    this.ifStates.set(node, state);
  }

  clearIfState(node: AST.IfExpression): void {
    this.ifStates.delete(node);
  }

  getMatchState(node: AST.MatchExpression): MatchExpressionState | undefined {
    return this.matchStates.get(node);
  }

  setMatchState(node: AST.MatchExpression, state: MatchExpressionState): void {
    this.matchStates.set(node, state);
  }

  clearMatchState(node: AST.MatchExpression): void {
    this.matchStates.delete(node);
  }

  getStringInterpolationState(node: AST.StringInterpolation): StringInterpolationState | undefined {
    return this.stringInterpolationStates.get(node);
  }

  setStringInterpolationState(node: AST.StringInterpolation, state: StringInterpolationState): void {
    this.stringInterpolationStates.set(node, state);
  }

  clearStringInterpolationState(node: AST.StringInterpolation): void {
    this.stringInterpolationStates.delete(node);
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
    this.awaitStates = new WeakMap();
    this.nativeCallStates = new WeakMap();
  }
}

export type NativeCallState = {
  status: "pending" | "resolved" | "rejected";
  value?: RuntimeValue;
  error?: RuntimeValue;
};
