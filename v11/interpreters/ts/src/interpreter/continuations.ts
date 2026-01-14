import type * as AST from "../ast";
import type { Environment } from "./environment";
import type { IteratorValue, RuntimeValue } from "./values";

export type BlockState = {
  env: Environment;
  index: number;
  result: RuntimeValue;
};

export type ModuleState = {
  env: Environment;
  index: number;
  result: RuntimeValue;
  initialized: boolean;
};

export type ForLoopState = {
  env: Environment;
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
  env: Environment;
  baseEnv: Environment;
  result: RuntimeValue;
  inBody: boolean;
  loopEnv?: Environment;
  conditionInProgress: boolean;
};

export type LoopExpressionState = {
  env: Environment;
  baseEnv: Environment;
  result: RuntimeValue;
  inBody: boolean;
  loopEnv?: Environment;
};

export type IfExpressionState = {
  env: Environment;
  stage: "if_condition" | "if_body" | "elseif_condition" | "elseif_body" | "else_body";
  elseIfIndex: number;
  result?: RuntimeValue;
};

export type MatchExpressionState = {
  env: Environment;
  stage: "subject" | "clause" | "guard" | "body";
  clauseIndex: number;
  subject?: RuntimeValue;
  matchEnv?: Environment;
};

export type StringInterpolationState = {
  env: Environment;
  index: number;
  output: string;
};

export type EnsureState = {
  env: Environment;
  stage: "try" | "ensure";
  result?: RuntimeValue;
  caught?: unknown;
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

  getStringInterpolationState(node: AST.StringInterpolation): StringInterpolationState | undefined;
  setStringInterpolationState(node: AST.StringInterpolation, state: StringInterpolationState): void;
  clearStringInterpolationState(node: AST.StringInterpolation): void;

  getEnsureState(node: AST.EnsureExpression): EnsureState | undefined;
  setEnsureState(node: AST.EnsureExpression, state: EnsureState): void;
  clearEnsureState(node: AST.EnsureExpression): void;
}
