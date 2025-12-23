import type * as AST from "../ast";
import type { Environment } from "./environment";
import type { IteratorValue, RuntimeValue } from "./values";

export type BlockState = {
  env: Environment;
  index: number;
  result: RuntimeValue;
};

export type ForLoopState = {
  mode: "static" | "iterator";
  values?: RuntimeValue[];
  iterator?: IteratorValue;
  baseEnv: Environment;
  index: number;
  result: RuntimeValue;
  iterationEnv?: Environment;
  awaitingBody: boolean;
  pendingValue?: RuntimeValue;
  iteratorClosed?: boolean;
};

export type WhileLoopState = {
  baseEnv: Environment;
  result: RuntimeValue;
  inBody: boolean;
  loopEnv?: Environment;
  conditionInProgress: boolean;
};

export type LoopExpressionState = {
  baseEnv: Environment;
  result: RuntimeValue;
  inBody: boolean;
  loopEnv?: Environment;
};

export type IfExpressionState = {
  stage: "if_condition" | "if_body" | "or_condition" | "or_body";
  orIndex: number;
  result?: RuntimeValue;
};

export type MatchExpressionState = {
  stage: "subject" | "clause" | "guard" | "body";
  clauseIndex: number;
  subject?: RuntimeValue;
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
  getLoopExpressionState(node: AST.LoopExpression): LoopExpressionState | undefined;
  setLoopExpressionState(node: AST.LoopExpression, state: LoopExpressionState): void;
  clearLoopExpressionState(node: AST.LoopExpression): void;

  getIfState(node: AST.IfExpression): IfExpressionState | undefined;
  setIfState(node: AST.IfExpression, state: IfExpressionState): void;
  clearIfState(node: AST.IfExpression): void;

  getMatchState(node: AST.MatchExpression): MatchExpressionState | undefined;
  setMatchState(node: AST.MatchExpression, state: MatchExpressionState): void;
  clearMatchState(node: AST.MatchExpression): void;
}
