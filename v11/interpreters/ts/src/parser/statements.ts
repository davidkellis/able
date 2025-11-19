import * as AST from "../ast";
import type { BlockExpression, Expression, FunctionCall, Statement } from "../ast";
import {
  annotate,
  annotateStatement,
  firstNamedChild,
  getActiveParseContext,
  inheritMetadata,
  isIgnorableNode,
  MapperError,
  MutableParseContext,
  nextNamedSibling,
  Node,
  parseLabel,
  sameNode,
} from "./shared";

export function registerStatementParsers(ctx: MutableParseContext): void {
  ctx.parseStatement = node => parseStatement(node, ctx.source);
  ctx.parseBlock = node => parseBlock(node, ctx.source);
}

function parseStatement(node: Node, source: string): Statement | null {
  const ctx = getActiveParseContext();
  switch (node.type) {
    case "expression_statement": {
      const exprNode = firstNamedChild(node);
      if (!exprNode) {
        throw new MapperError("parser: expression statement missing expression");
      }
      return annotateStatement(ctx.parseExpression(exprNode), node);
    }
    case "return_statement": {
      const valueNode = firstNamedChild(node);
      if (!valueNode) {
        return annotateStatement(AST.returnStatement(), node);
      }
      const expr = ctx.parseExpression(valueNode);
      return annotateStatement(AST.returnStatement(expr), node);
    }
    case "while_statement": {
      if (node.namedChildCount < 2) {
        throw new MapperError("parser: malformed while statement");
      }
      const condition = ctx.parseExpression(node.namedChild(0));
      const body = ctx.parseBlock(node.namedChild(1));
      return annotateStatement(AST.whileLoop(condition, body), node);
    }
    case "for_statement": {
      if (node.namedChildCount < 3) {
        throw new MapperError("parser: malformed for statement");
      }
      const pattern = ctx.parsePattern(node.namedChild(0));
      const iterable = ctx.parseExpression(node.namedChild(1));
      const body = ctx.parseBlock(node.namedChild(2));
      return annotateStatement(AST.forLoop(pattern, iterable, body), node);
    }
    case "break_statement": {
      const labelNode = node.childForFieldName("label");
      const valueNode = node.childForFieldName("value");
      const label = labelNode ? parseLabel(labelNode, source) : undefined;
      const value = valueNode ? ctx.parseExpression(valueNode) : undefined;
      return annotateStatement(AST.breakStatement(label, value), node);
    }
    case "continue_statement":
      return annotateStatement(AST.continueStatement(), node);
    case "raise_statement": {
      const valueNode = firstNamedChild(node);
      if (!valueNode) {
        throw new MapperError("parser: raise statement missing expression");
      }
      return annotateStatement(AST.raiseStatement(ctx.parseExpression(valueNode)), node);
    }
    case "rethrow_statement":
      return annotateStatement(AST.rethrowStatement(), node);
    case "struct_definition":
      return annotateStatement(ctx.parseStructDefinition(node), node);
    case "methods_definition":
      return annotateStatement(ctx.parseMethodsDefinition(node), node);
    case "implementation_definition":
      return annotateStatement(ctx.parseImplementationDefinition(node), node);
    case "named_implementation_definition":
      return annotateStatement(ctx.parseNamedImplementationDefinition(node), node);
    case "union_definition":
      return annotateStatement(ctx.parseUnionDefinition(node), node);
    case "interface_definition":
      return annotateStatement(ctx.parseInterfaceDefinition(node), node);
    case "type_alias_definition":
      return annotateStatement(ctx.parseTypeAliasDefinition(node), node);
    case "prelude_statement":
      return annotateStatement(ctx.parsePreludeStatement(node), node);
    case "extern_function":
      return annotateStatement(ctx.parseExternFunction(node), node);
    case "function_definition":
      return ctx.parseFunctionDefinition(node);
    default:
      return null;
  }
}

function parseBlock(node: Node | null | undefined, source: string): BlockExpression {
  if (!node) {
    return annotate(AST.blockExpression([]), node);
  }

  const ctx = getActiveParseContext();
  const statements: Statement[] = [];

  for (let i = 0; i < node.namedChildCount; ) {
    const child = node.namedChild(i);
    i++;
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    const fieldName = node.fieldNameForChild(i - 1);
    if (fieldName === "binding" && child.type === "identifier") continue;

    let stmt: Statement | null = null;
    if (child.type === "break_statement") {
      stmt = ctx.parseStatement(child);
      const breakStmt = stmt as AST.BreakStatement | null;
      if (breakStmt && breakStmt.type === "BreakStatement" && !breakStmt.value) {
        const next = nextNamedSibling(node, i - 1);
        if (next && next.type === "expression_statement") {
          const exprNode = firstNamedChild(next);
          if (exprNode) {
            breakStmt.value = ctx.parseExpression(exprNode);
            i++;
          }
        }
      }
    } else {
      stmt = ctx.parseStatement(child);
    }

    if (!stmt) continue;

    if (
      child.type === "expression_statement" &&
      stmt.type === "AssignmentExpression" &&
      (stmt.operator === "=" || stmt.operator === ":=")
    ) {
      const assignment = stmt as AST.AssignmentExpression;
      let anchorNode: Node = child;
      let anchorIndex = i - 1;
      while (true) {
        const next = nextNamedSibling(node, anchorIndex);
        if (!next || next.type !== "expression_statement") break;
        if (anchorNode.endPosition.row !== next.startPosition.row) break;
        if (hasSemicolonBetween(source, anchorNode, next)) break;
        const exprNode = firstNamedChild(next);
        if (!exprNode) break;
        const expr = ctx.parseExpression(exprNode);
        if (expr.type !== "UnaryExpression" || expr.operator !== "-") break;
        assignment.right = inheritMetadata(AST.binaryExpression("-", assignment.right, expr.operand), assignment.right, expr);
        i++;
        const nextIndex = findNamedChildIndex(node, next);
        if (nextIndex === -1) break;
        anchorIndex = nextIndex;
        anchorNode = next;
      }
    }

    if (stmt.type === "LambdaExpression" && statements.length > 0) {
      const prev = statements[statements.length - 1];
      if (prev.type === "FunctionCall") {
        const call = prev as FunctionCall;
        if (call.arguments.length === 0 || call.arguments[call.arguments.length - 1] !== stmt) {
          call.arguments.push(stmt);
        }
        call.isTrailingLambda = true;
        continue;
      }
      if ((prev as Expression).type) {
        const call = inheritMetadata(AST.functionCall(prev as Expression, [], undefined, true), prev as Expression, stmt);
        call.arguments.push(stmt);
        statements[statements.length - 1] = call;
        continue;
      }
    }

    statements.push(stmt);
  }

  return annotate(AST.blockExpression(statements), node) as BlockExpression;
}

function findNamedChildIndex(parent: Node | null | undefined, target: Node | null | undefined): number {
  if (!parent || !target) return -1;
  for (let idx = 0; idx < parent.namedChildCount; idx++) {
    const candidate = parent.namedChild(idx);
    if (!candidate || !candidate.isNamed || isIgnorableNode(candidate)) continue;
    if (sameNode(candidate, target)) return idx;
  }
  return -1;
}

function hasSemicolonBetween(source: string, left: Node, right: Node): boolean {
  const start = left.endIndex;
  const end = right.startIndex;
  if (start < 0 || end < start || end > source.length) return false;
  for (let idx = start; idx < end; idx++) {
    if (source[idx] === ";") {
      return true;
    }
  }
  return false;
}
