import type * as AST from "../../ast";

export type PlaceholderPlan = { paramCount: number };

export function placeholderFunctionPlan(expr: AST.Expression | null | undefined): PlaceholderPlan | null {
  if (!expr) return null;
  if (expr.type === "AssignmentExpression") return null;
  // Block expressions should be typechecked normally; placeholder presence inside a block
  // (e.g., via a pipe) must not coerce the whole block into a placeholder function.
  if (expr.type === "BlockExpression") return null;
  if (expr.type === "BinaryExpression" && (expr.operator === "|>" || expr.operator === "|>>")) {
    return null;
  }
  const plan = analyzePlaceholderExpression(expr);
  if (!plan) return null;
  if (expr.type === "FunctionCall") {
    const calleeHas = expressionContainsPlaceholder(expr.callee);
    const argsHave = expr.arguments.some((arg) => expressionContainsPlaceholder(arg));
    if (calleeHas && !argsHave) {
      return null;
    }
  }
  return plan;
}

export function expressionContainsPlaceholder(expr: AST.Expression | null | undefined): boolean {
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
      return expr.member?.type && isExpression(expr.member) ? expressionContainsPlaceholder(expr.member) : false;
    case "ImplicitMemberExpression":
      return false;
    case "IndexExpression":
      return expressionContainsPlaceholder(expr.object) || expressionContainsPlaceholder(expr.index);
    case "BlockExpression":
      return expr.body.some((stmt) => statementContainsPlaceholder(stmt));
    case "AssignmentExpression":
      if (expressionContainsPlaceholder(expr.right)) return true;
      return isExpression(expr.left) && expressionContainsPlaceholder(expr.left);
    case "StringInterpolation":
      return expr.parts.some((part) => expressionContainsPlaceholder(part));
    case "StructLiteral": {
      if (expr.fields.some((field) => field && expressionContainsPlaceholder(field.value))) return true;
      const legacySource = (expr as any).functionalUpdateSource as AST.Expression | undefined;
      const updateSources = expr.functionalUpdateSources ?? (legacySource ? [legacySource] : []);
      return updateSources.some((src) => expressionContainsPlaceholder(src));
    }
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
      if (
        expr.elseIfClauses.some(
          (clause) =>
            !!clause &&
            (expressionContainsPlaceholder(clause.condition) || expressionContainsPlaceholder(clause.body)),
        )
      ) {
        return true;
      }
      return expr.elseBody ? expressionContainsPlaceholder(expr.elseBody) : false;
    case "PropagationExpression":
      return expressionContainsPlaceholder(expr.expression);
    case "AwaitExpression":
      return expressionContainsPlaceholder(expr.expression);
    case "LoopExpression":
      return expressionContainsPlaceholder(expr.body);
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
      return expressionContainsPlaceholder(stmt.iterable) || expressionContainsPlaceholder(stmt.body);
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

function analyzePlaceholderExpression(expr: AST.Expression): PlaceholderPlan | null {
  const analyzer = new PlaceholderAnalyzer();
  analyzer.visitExpression(expr);
  if (!analyzer.hasPlaceholder) return null;
  const paramCount = analyzer.highestIndex > 0 ? analyzer.highestIndex : 1;
  return { paramCount };
}

class PlaceholderAnalyzer {
  highestIndex = 0;
  hasPlaceholder = false;

  visitExpression(expr: AST.Expression | null | undefined): void {
    if (!expr) return;
    switch (expr.type) {
      case "PlaceholderExpression": {
        this.hasPlaceholder = true;
        const idx = expr.index ?? 1;
        if (idx <= 0) {
          throw new Error(`Placeholder index must be positive, found @${idx}`);
        }
        if (idx > this.highestIndex) this.highestIndex = idx;
        return;
      }
      case "BinaryExpression":
        this.visitExpression(expr.left);
        this.visitExpression(expr.right);
        return;
      case "UnaryExpression":
        this.visitExpression(expr.operand);
        return;
      case "FunctionCall":
        this.visitExpression(expr.callee);
        expr.arguments.forEach((arg) => this.visitExpression(arg));
        return;
      case "MemberAccessExpression":
        this.visitExpression(expr.object);
        if (expr.member && isExpression(expr.member)) {
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
        expr.body.forEach((stmt) => this.visitStatement(stmt));
        return;
      case "AssignmentExpression":
        this.visitExpression(expr.right);
        if (isExpression(expr.left)) this.visitExpression(expr.left);
        return;
      case "StringInterpolation":
        expr.parts.forEach((part) => this.visitExpression(part));
        return;
      case "StructLiteral": {
        expr.fields.forEach((field) => field && this.visitExpression(field.value));
        const legacySource = (expr as any).functionalUpdateSource as AST.Expression | undefined;
        const updateSources = expr.functionalUpdateSources ?? (legacySource ? [legacySource] : []);
        updateSources.forEach((src) => this.visitExpression(src));
        return;
      }
      case "ArrayLiteral":
        expr.elements.forEach((el) => this.visitExpression(el));
        return;
      case "RangeExpression":
        this.visitExpression(expr.start);
        this.visitExpression(expr.end);
        return;
      case "MatchExpression":
        this.visitExpression(expr.subject);
        expr.clauses.forEach((clause) => {
          if (!clause) return;
          if (clause.guard) this.visitExpression(clause.guard);
          this.visitExpression(clause.body);
        });
        return;
      case "OrElseExpression":
        this.visitExpression(expr.expression);
        this.visitExpression(expr.handler);
        return;
      case "RescueExpression":
        this.visitExpression(expr.monitoredExpression);
        expr.clauses.forEach((clause) => {
          if (!clause) return;
          if (clause.guard) this.visitExpression(clause.guard);
          this.visitExpression(clause.body);
        });
        return;
      case "EnsureExpression":
        this.visitExpression(expr.tryExpression);
        this.visitExpression(expr.ensureBlock);
        return;
      case "IfExpression":
        this.visitExpression(expr.ifCondition);
        this.visitExpression(expr.ifBody);
        expr.elseIfClauses.forEach((clause) => {
          if (!clause) return;
          this.visitExpression(clause.condition);
          this.visitExpression(clause.body);
        });
        if (expr.elseBody) this.visitExpression(expr.elseBody);
        return;
      case "PropagationExpression":
        this.visitExpression(expr.expression);
        return;
      case "AwaitExpression":
        this.visitExpression(expr.expression);
        return;
      case "LoopExpression":
        this.visitExpression(expr.body);
        return;
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
