import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";

export type PlaceholderFrame = {
  args: V10Value[];
  explicit: Set<number>;
  implicitAssigned: Set<number>;
  nextImplicit: number;
  paramCount: number;
};

type PlaceholderPlan = {
  explicitIndices: Set<number>;
  paramCount: number;
};

declare module "./index" {
  interface InterpreterV10 {
    placeholderFrames: PlaceholderFrame[];
    hasPlaceholderFrame(): boolean;
    pushPlaceholderFrame(plan: PlaceholderPlan, args: V10Value[]): void;
    popPlaceholderFrame(): void;
    currentPlaceholderFrame(): PlaceholderFrame | undefined;
    evaluatePlaceholderExpression(node: AST.PlaceholderExpression, env: Environment): V10Value;
    tryBuildPlaceholderFunction(node: AST.Expression, env: Environment): V10Value | null;
  }
}

export function applyPlaceholderAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.hasPlaceholderFrame = function hasPlaceholderFrame(this: InterpreterV10): boolean {
    return this.placeholderFrames.length > 0;
  };

  cls.prototype.pushPlaceholderFrame = function pushPlaceholderFrame(this: InterpreterV10, plan: PlaceholderPlan, args: V10Value[]): void {
    const frame: PlaceholderFrame = {
      args,
      explicit: new Set(plan.explicitIndices),
      implicitAssigned: new Set(),
      nextImplicit: 1,
      paramCount: plan.paramCount,
    };
    this.placeholderFrames.push(frame);
  };

  cls.prototype.popPlaceholderFrame = function popPlaceholderFrame(this: InterpreterV10): void {
    if (this.placeholderFrames.length === 0) return;
    this.placeholderFrames.pop();
  };

  cls.prototype.currentPlaceholderFrame = function currentPlaceholderFrame(this: InterpreterV10): PlaceholderFrame | undefined {
    if (this.placeholderFrames.length === 0) return undefined;
    return this.placeholderFrames[this.placeholderFrames.length - 1];
  };

  cls.prototype.evaluatePlaceholderExpression = function evaluatePlaceholderExpression(this: InterpreterV10, node: AST.PlaceholderExpression): V10Value {
    const frame = this.currentPlaceholderFrame();
    if (!frame) {
      throw new Error("Expression placeholder used outside of placeholder lambda");
    }
    if (node.index !== undefined) {
      const idx = node.index;
      if (idx <= 0) throw new Error(`Placeholder index must be positive, found @${idx}`);
      return placeholderValueAt(frame, idx);
    }
    const idx = nextImplicitIndex(frame);
    return placeholderValueAt(frame, idx);
  };

  cls.prototype.tryBuildPlaceholderFunction = function tryBuildPlaceholderFunction(this: InterpreterV10, node: AST.Expression, env: Environment): V10Value | null {
    if (this.hasPlaceholderFrame()) return null;
    if (node.type === "AssignmentExpression") {
      return null;
    }
    const plan = analyzePlaceholderExpression(node);
    if (!plan) return null;
    if (node.type === "BinaryExpression" && node.operator === "|>") {
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
    const func: Extract<V10Value, { kind: "native_function" }> = {
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

function placeholderValueAt(frame: PlaceholderFrame, index: number): V10Value {
  if (index <= 0 || index > frame.args.length) {
    throw new Error(`Placeholder index @${index} is out of range`);
  }
  const val = frame.args[index - 1];
  if (!val) {
    return { kind: "nil", value: null };
  }
  return val;
}

function nextImplicitIndex(frame: PlaceholderFrame): number {
  while (frame.nextImplicit <= frame.paramCount) {
    const idx = frame.nextImplicit;
    frame.nextImplicit += 1;
    if (frame.explicit.has(idx)) continue;
    if (frame.implicitAssigned.has(idx)) continue;
    frame.implicitAssigned.add(idx);
    return idx;
  }
  throw new Error("no implicit placeholder slots available");
}

function analyzePlaceholderExpression(expr: AST.Expression): PlaceholderPlan | null {
  const analyzer = new PlaceholderAnalyzer();
  analyzer.visitExpression(expr);
  if (!analyzer.hasPlaceholder) return null;
  let paramCount = analyzer.highestExplicit;
  const implicitTotal = analyzer.explicit.size + analyzer.implicitCount;
  if (implicitTotal > paramCount) paramCount = implicitTotal;
  return {
    explicitIndices: analyzer.explicit,
    paramCount,
  };
}

class PlaceholderAnalyzer {
  explicit = new Set<number>();
  implicitCount = 0;
  highestExplicit = 0;
  hasPlaceholder = false;

  visitExpression(expr: AST.Expression | null | undefined): void {
    if (!expr) return;
    switch (expr.type) {
      case "PlaceholderExpression":
        this.hasPlaceholder = true;
        if (expr.index !== undefined) {
          const idx = expr.index;
          if (idx <= 0) {
            throw new Error(`Placeholder index must be positive, found @${idx}`);
          }
          this.explicit.add(idx);
          if (idx > this.highestExplicit) this.highestExplicit = idx;
        } else {
          this.implicitCount += 1;
        }
        return;
      case "BinaryExpression":
        this.visitExpression(expr.left);
        this.visitExpression(expr.right);
        return;
      case "UnaryExpression":
        this.visitExpression(expr.operand);
        return;
      case "FunctionCall":
        this.visitExpression(expr.callee);
        for (const arg of expr.arguments) {
          this.visitExpression(arg);
        }
        return;
      case "MemberAccessExpression":
        this.visitExpression(expr.object);
        if (isExpression(expr.member)) {
          this.visitExpression(expr.member);
        }
        return;
      case "ImplicitMemberExpression":
        return;
      case "IndexExpression":
        this.visitExpression(expr.object);
        this.visitExpression(expr.index);
        return;
      case "BlockExpression":
        for (const stmt of expr.body) {
          this.visitStatement(stmt);
        }
        return;
      case "AssignmentExpression":
        this.visitExpression(expr.right);
        if (isExpression(expr.left)) this.visitExpression(expr.left);
        return;
      case "LoopExpression":
        this.visitExpression(expr.body);
        return;
      case "StringInterpolation":
        for (const part of expr.parts) {
          this.visitExpression(part);
        }
        return;
      case "StructLiteral":
        for (const field of expr.fields) {
          if (!field) continue;
          this.visitExpression(field.value);
        }
        const legacySource = (expr as any).functionalUpdateSource as AST.Expression | undefined;
        const updateSources = expr.functionalUpdateSources ?? (legacySource ? [legacySource] : []);
        for (const src of updateSources) {
          this.visitExpression(src);
        }
        return;
      case "ArrayLiteral":
        for (const el of expr.elements) {
          this.visitExpression(el);
        }
        return;
      case "RangeExpression":
        this.visitExpression(expr.start);
        this.visitExpression(expr.end);
        return;
      case "MatchExpression":
        this.visitExpression(expr.subject);
        for (const clause of expr.clauses) {
          if (!clause) continue;
          if (clause.guard) this.visitExpression(clause.guard);
          this.visitExpression(clause.body);
        }
        return;
      case "OrElseExpression":
        this.visitExpression(expr.expression);
        this.visitExpression(expr.handler);
        return;
      case "RescueExpression":
        this.visitExpression(expr.monitoredExpression);
        for (const clause of expr.clauses) {
          if (!clause) continue;
          if (clause.guard) this.visitExpression(clause.guard);
          this.visitExpression(clause.body);
        }
        return;
      case "EnsureExpression":
        this.visitExpression(expr.tryExpression);
        this.visitExpression(expr.ensureBlock);
        return;
      case "IfExpression":
        this.visitExpression(expr.ifCondition);
        this.visitExpression(expr.ifBody);
        for (const clause of expr.orClauses) {
          if (!clause) continue;
          if (clause.condition) this.visitExpression(clause.condition);
          this.visitExpression(clause.body);
        }
        return;
      case "PropagationExpression":
        this.visitExpression(expr.expression);
        return;
      case "IteratorLiteral":
      case "LambdaExpression":
      case "ProcExpression":
      case "SpawnExpression":
      case "TopicReferenceExpression":
      case "Identifier":
      case "IntegerLiteral":
      case "FloatLiteral":
      case "BooleanLiteral":
      case "StringLiteral":
      case "CharLiteral":
      case "NilLiteral":
        return;
      default:
        return;
    }
  }

  visitStatement(stmt: AST.Statement | null | undefined): void {
    if (!stmt) return;
    if (isExpression(stmt)) {
      this.visitExpression(stmt);
      return;
    }
    switch (stmt.type) {
      case "ReturnStatement":
        this.visitExpression(stmt.argument ?? null);
        return;
      case "RaiseStatement":
        this.visitExpression(stmt.expression ?? null);
        return;
      case "ForLoop":
        this.visitExpression(stmt.iterable);
        this.visitExpression(stmt.body);
        return;
      case "WhileLoop":
        this.visitExpression(stmt.condition);
        this.visitExpression(stmt.body);
        return;
      case "BreakStatement":
        this.visitExpression(stmt.value ?? null);
        return;
      case "ContinueStatement":
        return;
      case "YieldStatement":
        this.visitExpression(stmt.expression ?? null);
        return;
      default:
        return;
    }
  }
}

function expressionContainsPlaceholder(expr: AST.Expression | null | undefined): boolean {
  if (!expr) return false;
  switch (expr.type) {
    case "PlaceholderExpression":
      return true;
    case "BinaryExpression":
      return expressionContainsPlaceholder(expr.left) || expressionContainsPlaceholder(expr.right);
    case "UnaryExpression":
      return expressionContainsPlaceholder(expr.operand);
    case "FunctionCall":
      if (expressionContainsPlaceholder(expr.callee)) return true;
      return expr.arguments.some((arg) => expressionContainsPlaceholder(arg));
    case "MemberAccessExpression":
      if (expressionContainsPlaceholder(expr.object)) return true;
      if (isExpression(expr.member)) {
        return expressionContainsPlaceholder(expr.member);
      }
      return false;
    case "ImplicitMemberExpression":
      return false;
    case "IndexExpression":
      return expressionContainsPlaceholder(expr.object) || expressionContainsPlaceholder(expr.index);
    case "BlockExpression":
      return expr.body.some((stmt) => statementContainsPlaceholder(stmt));
    case "AssignmentExpression":
      if (expressionContainsPlaceholder(expr.right)) return true;
      if (isExpression(expr.left)) return expressionContainsPlaceholder(expr.left);
      return false;
    case "StringInterpolation":
      return expr.parts.some((part) => expressionContainsPlaceholder(part));
    case "StructLiteral":
      if (expr.fields.some((field) => field && expressionContainsPlaceholder(field.value))) return true;
      {
        const legacySource = (expr as any).functionalUpdateSource as AST.Expression | undefined;
        const updateSources = expr.functionalUpdateSources ?? (legacySource ? [legacySource] : []);
        for (const src of updateSources) {
          if (expressionContainsPlaceholder(src)) return true;
        }
      }
      return false;
    case "ArrayLiteral":
      return expr.elements.some((el) => expressionContainsPlaceholder(el));
    case "RangeExpression":
      return expressionContainsPlaceholder(expr.start) || expressionContainsPlaceholder(expr.end);
    case "MatchExpression":
      if (expressionContainsPlaceholder(expr.subject)) return true;
      return expr.clauses.some(
        (clause) =>
          !!clause &&
          ((clause.guard && expressionContainsPlaceholder(clause.guard)) ||
            expressionContainsPlaceholder(clause.body)),
      );
    case "OrElseExpression":
      return expressionContainsPlaceholder(expr.expression) || expressionContainsPlaceholder(expr.handler);
    case "RescueExpression":
      if (expressionContainsPlaceholder(expr.monitoredExpression)) return true;
      return expr.clauses.some(
        (clause) =>
          !!clause &&
          ((clause.guard && expressionContainsPlaceholder(clause.guard)) ||
            expressionContainsPlaceholder(clause.body)),
      );
    case "EnsureExpression":
      return expressionContainsPlaceholder(expr.tryExpression) || expressionContainsPlaceholder(expr.ensureBlock);
    case "IfExpression":
      if (expressionContainsPlaceholder(expr.ifCondition) || expressionContainsPlaceholder(expr.ifBody)) return true;
      return expr.orClauses.some(
        (clause) =>
          !!clause &&
          ((clause.condition && expressionContainsPlaceholder(clause.condition)) ||
            expressionContainsPlaceholder(clause.body)),
      );
    case "PropagationExpression":
      return expressionContainsPlaceholder(expr.expression);
    case "LoopExpression":
      return expressionContainsPlaceholder(expr.body);
    case "IteratorLiteral":
    case "LambdaExpression":
    case "ProcExpression":
    case "SpawnExpression":
    case "TopicReferenceExpression":
    case "Identifier":
    case "IntegerLiteral":
    case "FloatLiteral":
    case "BooleanLiteral":
    case "StringLiteral":
    case "CharLiteral":
    case "NilLiteral":
      return false;
    default:
      return false;
  }
}

function statementContainsPlaceholder(stmt: AST.Statement | null | undefined): boolean {
  if (!stmt) return false;
  if (isExpression(stmt)) {
    return expressionContainsPlaceholder(stmt);
  }
  switch (stmt.type) {
    case "ReturnStatement":
      return expressionContainsPlaceholder(stmt.argument ?? null);
    case "RaiseStatement":
      return expressionContainsPlaceholder(stmt.expression ?? null);
    case "ForLoop":
      return (
        expressionContainsPlaceholder(stmt.iterable) || expressionContainsPlaceholder(stmt.body)
      );
    case "WhileLoop":
      return expressionContainsPlaceholder(stmt.condition) || expressionContainsPlaceholder(stmt.body);
    case "BreakStatement":
      return expressionContainsPlaceholder(stmt.value ?? null);
    case "ContinueStatement":
      return false;
    case "YieldStatement":
      return expressionContainsPlaceholder(stmt.expression ?? null);
    default:
      return false;
  }
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
    case "TopicReferenceExpression":
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
