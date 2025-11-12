import type * as AST from "../ast";
import type { Environment } from "./environment";
import type { IteratorValue, V10Value } from "./values";

export type BlockState = {
  env: Environment;
  index: number;
  result: V10Value;
};

export type ForLoopState = {
  mode: "static" | "iterator";
  values?: V10Value[];
  iterator?: IteratorValue;
  baseEnv: Environment;
  index: number;
  result: V10Value;
  iterationEnv?: Environment;
  awaitingBody: boolean;
  pendingValue?: V10Value;
  iteratorClosed?: boolean;
};

export type WhileLoopState = {
  baseEnv: Environment;
  result: V10Value;
  inBody: boolean;
  loopEnv?: Environment;
  conditionInProgress: boolean;
};

export type IfExpressionState = {
  stage: "if_condition" | "if_body" | "or_condition" | "or_body";
  orIndex: number;
  result?: V10Value;
};

export type MatchExpressionState = {
  stage: "subject" | "clause" | "guard" | "body";
  clauseIndex: number;
  subject?: V10Value;
  matchEnv?: Environment;
};

export interface ContinuationContext {
  kind: "generator" | "proc";
  markStatementIncomplete(): void;

  getBlockState(node: AST.BlockExpression): BlockState | undefined;
  setBlockState(node: AST.BlockExpression, state: BlockState): void;
  clearBlockState(node: AST.BlockExpression): void;

  getForLoopState(node: AST.ForLoop): ForLoopState | undefined;
  setForLoopState(node: AST.ForLoop, state: ForLoopState): void;
  clearForLoopState(node: AST.ForLoop): void;

  getWhileLoopState(node: AST.WhileLoop): WhileLoopState | undefined;
  setWhileLoopState(node: AST.WhileLoop, state: WhileLoopState): void;
  clearWhileLoopState(node: AST.WhileLoop): void;

  getIfState(node: AST.IfExpression): IfExpressionState | undefined;
  setIfState(node: AST.IfExpression, state: IfExpressionState): void;
  clearIfState(node: AST.IfExpression): void;

  getMatchState(node: AST.MatchExpression): MatchExpressionState | undefined;
  setMatchState(node: AST.MatchExpression, state: MatchExpressionState): void;
  clearMatchState(node: AST.MatchExpression): void;
}
