import * as AST from "../ast";
import { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { RuntimeValue } from "./values";

export type PlaceholderFrame = {
  args: RuntimeValue[];
  paramCount: number;
};

type PlaceholderPlan = {
  paramCount: number;
};

declare module "./index" {
  interface Interpreter {
    placeholderFrames: PlaceholderFrame[];
    hasPlaceholderFrame(): boolean;
    pushPlaceholderFrame(plan: PlaceholderPlan, args: RuntimeValue[]): void;
    popPlaceholderFrame(): void;
    currentPlaceholderFrame(): PlaceholderFrame | undefined;
    evaluatePlaceholderExpression(node: AST.PlaceholderExpression, env: Environment): RuntimeValue;
    tryBuildPlaceholderFunction(node: AST.Expression, env: Environment): RuntimeValue | null;
  }
}

export function applyPlaceholderAugmentations(cls: typeof Interpreter): void {
  cls.prototype.hasPlaceholderFrame = function hasPlaceholderFrame(this: Interpreter): boolean {
    return this.placeholderFrames.length > 0;
  };

  cls.prototype.pushPlaceholderFrame = function pushPlaceholderFrame(this: Interpreter, plan: PlaceholderPlan, args: RuntimeValue[]): void {
    const frame: PlaceholderFrame = { args, paramCount: plan.paramCount };
    this.placeholderFrames.push(frame);
  };

  cls.prototype.popPlaceholderFrame = function popPlaceholderFrame(this: Interpreter): void {
    if (this.placeholderFrames.length === 0) return;
    this.placeholderFrames.pop();
  };

  cls.prototype.currentPlaceholderFrame = function currentPlaceholderFrame(this: Interpreter): PlaceholderFrame | undefined {
    if (this.placeholderFrames.length === 0) return undefined;
    return this.placeholderFrames[this.placeholderFrames.length - 1];
  };

  cls.prototype.evaluatePlaceholderExpression = function evaluatePlaceholderExpression(this: Interpreter, node: AST.PlaceholderExpression): RuntimeValue {
    const frame = this.currentPlaceholderFrame();
    if (!frame) {
      throw new Error("Expression placeholder used outside of placeholder lambda");
    }
    const idx = (node.index ?? 1);
    if (idx <= 0) throw new Error(`Placeholder index must be positive, found @${idx}`);
    return placeholderValueAt(frame, idx);
  };

  cls.prototype.tryBuildPlaceholderFunction = function tryBuildPlaceholderFunction(this: Interpreter, node: AST.Expression, env: Environment): RuntimeValue | null {
    if (this.hasPlaceholderFrame()) return null;
    if (node.type === "AssignmentExpression") {
      return null;
    }
    const plan = analyzePlaceholderExpression(node);
    if (!plan) return null;
    if (node.type === "BinaryExpression" && (node.operator === "|>" || node.operator === "|>>")) {
      return null;
    }
    if (node.type === "FunctionCall") {
      const calleeHas = expressionContainsPlaceholder(node.callee);
      let argsHave = false;
      for (const arg of node.arguments) {
        if (expressionContainsPlaceholder(arg)) {
          argsHave = true;
          break;
        }
      }
      if (calleeHas && !argsHave) {
        return null;
      }
    }
    const placeholderExpr = node;
    const placeholderEnv = env;
    const func: Extract<RuntimeValue, { kind: "native_function" }> = {
      kind: "native_function",
      name: "<placeholder>",
      arity: plan.paramCount,
      impl: (runtimeInterp, args) => {
        const activeInterp = runtimeInterp ?? this;
        if (args.length !== plan.paramCount) {
          throw new Error(`Placeholder lambda expects ${plan.paramCount} arguments, got ${args.length}`);
        }
        activeInterp.pushPlaceholderFrame(plan, args);
        try {
          const callEnv = new Environment(placeholderEnv);
          return activeInterp.evaluate(placeholderExpr, callEnv);
        } finally {
          activeInterp.popPlaceholderFrame();
        }
      },
    };
    return func;
  };
}

function placeholderValueAt(frame: PlaceholderFrame, index: number): RuntimeValue {
  if (index <= 0 || index > frame.args.length) {
    throw new Error(`Placeholder index @${index} is out of range`);
  }
  const val = frame.args[index - 1];
  if (!val) {
    return { kind: "nil", value: null };
  }
  return val;
}

function analyzePlaceholderExpression(expr: AST.Expression): PlaceholderPlan | null {
  const info = collectPlaceholderInfo(expr);
  if (!info.hasPlaceholder) return null;
  const paramCount = info.highestExplicit > 0 ? info.highestExplicit : 1;
  return { paramCount };
}

function collectPlaceholderInfo(
  root: AST.AstNode | null | undefined,
): { hasPlaceholder: boolean; highestExplicit: number } {
  if (!root) return { hasPlaceholder: false, highestExplicit: 0 };
  const visited = new Set<AST.AstNode>();
  const stack: AST.AstNode[] = [root];
  let hasPlaceholder = false;
  let highestExplicit = 0;

  while (stack.length > 0) {
    const node = stack.pop();
    if (!node || visited.has(node)) continue;
    visited.add(node);
    switch (node.type) {
      case "PlaceholderExpression": {
        hasPlaceholder = true;
        const idx = node.index ?? 1;
        if (idx <= 0) {
          throw new Error(`Placeholder index must be positive, found @${idx}`);
        }
        if (idx > highestExplicit) highestExplicit = idx;
        break;
      }
      case "BinaryExpression":
        stack.push(node.left, node.right);
        break;
      case "UnaryExpression":
        stack.push(node.operand);
        break;
      case "TypeCastExpression":
        stack.push(node.expression);
        break;
      case "FunctionCall":
        stack.push(node.callee, ...node.arguments);
        break;
      case "MemberAccessExpression":
        stack.push(node.object);
        if (isExpression(node.member)) {
          stack.push(node.member);
        }
        break;
      case "IndexExpression":
        stack.push(node.object, node.index);
        break;
      case "BlockExpression":
        stack.push(...node.body);
        break;
      case "AssignmentExpression":
        stack.push(node.right);
        if (isExpression(node.left)) stack.push(node.left);
        break;
      case "LoopExpression":
        stack.push(node.body);
        break;
      case "StringInterpolation":
        stack.push(...node.parts);
        break;
      case "StructLiteral": {
        for (const field of node.fields) {
          if (!field) continue;
          stack.push(field.value);
        }
        const legacySource = (node as any).functionalUpdateSource as AST.Expression | undefined;
        const updateSources = node.functionalUpdateSources ?? (legacySource ? [legacySource] : []);
        stack.push(...updateSources);
        break;
      }
      case "ArrayLiteral":
        stack.push(...node.elements);
        break;
      case "RangeExpression":
        stack.push(node.start, node.end);
        break;
      case "MatchExpression":
        stack.push(node.subject);
        for (const clause of node.clauses) {
          if (!clause) continue;
          if (clause.guard) stack.push(clause.guard);
          stack.push(clause.body);
        }
        break;
      case "OrElseExpression":
        stack.push(node.expression, node.handler);
        break;
      case "RescueExpression":
        stack.push(node.monitoredExpression);
        for (const clause of node.clauses) {
          if (!clause) continue;
          if (clause.guard) stack.push(clause.guard);
          stack.push(clause.body);
        }
        break;
      case "EnsureExpression":
        stack.push(node.tryExpression, node.ensureBlock);
        break;
      case "IfExpression":
        stack.push(node.ifCondition, node.ifBody);
        for (const clause of node.elseIfClauses) {
          if (!clause) continue;
          stack.push(clause.condition, clause.body);
        }
        if (node.elseBody) stack.push(node.elseBody);
        break;
      case "PropagationExpression":
        stack.push(node.expression);
        break;
      case "AwaitExpression":
        stack.push(node.expression);
        break;
      case "ReturnStatement":
        if (node.argument) stack.push(node.argument);
        break;
      case "RaiseStatement":
        if (node.expression) stack.push(node.expression);
        break;
      case "ForLoop":
        stack.push(node.iterable, node.body);
        break;
      case "WhileLoop":
        stack.push(node.condition, node.body);
        break;
      case "BreakStatement":
        if (node.value) stack.push(node.value);
        break;
      case "YieldStatement":
        if (node.expression) stack.push(node.expression);
        break;
      case "ImplicitMemberExpression":
      case "IteratorLiteral":
      case "LambdaExpression":
      case "ProcExpression":
      case "SpawnExpression":
      case "Identifier":
      case "IntegerLiteral":
      case "FloatLiteral":
      case "BooleanLiteral":
      case "StringLiteral":
      case "CharLiteral":
      case "NilLiteral":
      case "ContinueStatement":
        break;
      default:
        break;
    }
  }

  return { hasPlaceholder, highestExplicit };
}

function expressionContainsPlaceholder(expr: AST.Expression | null | undefined): boolean {
  return collectPlaceholderInfo(expr).hasPlaceholder;
}

function statementContainsPlaceholder(stmt: AST.Statement | null | undefined): boolean {
  return collectPlaceholderInfo(stmt).hasPlaceholder;
}

function isExpression(node: AST.AstNode | null | undefined): node is AST.Expression {
  if (!node) return false;
  switch (node.type) {
    case "Identifier":
    case "StringLiteral":
    case "BooleanLiteral":
    case "CharLiteral":
    case "NilLiteral":
    case "FloatLiteral":
    case "IntegerLiteral":
    case "ArrayLiteral":
    case "UnaryExpression":
    case "BinaryExpression":
    case "FunctionCall":
    case "BlockExpression":
    case "AssignmentExpression":
    case "RangeExpression":
    case "StringInterpolation":
    case "MemberAccessExpression":
    case "IndexExpression":
    case "LambdaExpression":
    case "ProcExpression":
    case "SpawnExpression":
    case "PropagationExpression":
    case "OrElseExpression":
    case "BreakpointExpression":
    case "IteratorLiteral":
    case "LoopExpression":
    case "ImplicitMemberExpression":
    case "PlaceholderExpression":
    case "AwaitExpression":
    case "IfExpression":
    case "MatchExpression":
    case "StructLiteral":
    case "RescueExpression":
    case "EnsureExpression":
      return true;
    default:
      return false;
  }
}
