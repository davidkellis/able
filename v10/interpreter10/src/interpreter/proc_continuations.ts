import * as AST from "../ast";
import type { ContinuationContext, BlockState, ForLoopState, WhileLoopState, IfExpressionState, MatchExpressionState } from "./continuations";

export class ProcContinuationContext implements ContinuationContext {
  readonly kind = "proc" as const;

  private blockStates: WeakMap<AST.BlockExpression, BlockState> = new WeakMap();
  private forLoopStates: WeakMap<AST.ForLoop, ForLoopState> = new WeakMap();
  private whileLoopStates: WeakMap<AST.WhileLoop, WhileLoopState> = new WeakMap();
  private ifStates: WeakMap<AST.IfExpression, IfExpressionState> = new WeakMap();
  private matchStates: WeakMap<AST.MatchExpression, MatchExpressionState> = new WeakMap();

  markStatementIncomplete(): void {
    // Proc continuations resume from the same statement automatically via stored state.
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

  reset(): void {
    this.blockStates = new WeakMap();
    this.forLoopStates = new WeakMap();
    this.whileLoopStates = new WeakMap();
    this.ifStates = new WeakMap();
    this.matchStates = new WeakMap();
  }
}
