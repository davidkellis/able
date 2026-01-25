import * as AST from "../ast";
import type {
  AssignmentExpression,
  BlockExpression,
  ElseIfClause,
  Expression,
  FunctionParameter,
  FunctionCall,
  Identifier,
  LambdaExpression,
  MatchClause,
  Statement,
} from "../ast";
import {
  annotate,
  annotateExpressionNode,
  firstNamedChild,
  getActiveParseContext,
  inheritMetadata,
  isIgnorableNode,
  MapperError,
  Node,
  parseIdentifier,
  parseLabel,
} from "./shared";
import { parseParameterList } from "./definitions";

export function parseLambdaExpression(node: Node, source: string): LambdaExpression {
  if (node.type !== "lambda_expression") {
    throw new MapperError("parser: expected lambda expression");
  }

  const params: FunctionParameter[] = [];
  const paramsNode = node.childForFieldName("parameters");
  if (paramsNode) {
    for (let i = 0; i < paramsNode.namedChildCount; i++) {
      const paramNode = paramsNode.namedChild(i);
      if (!paramNode || paramNode.type !== "lambda_parameter") continue;
      params.push(parseLambdaParameter(paramNode, source));
    }
  }

  const returnNode = node.childForFieldName("return_type");
  const returnType = returnNode ? getActiveParseContext().parseReturnType(returnNode) : undefined;

  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: lambda missing body");
  }
  const bodyExpr = getActiveParseContext().parseExpression(bodyNode);

  return annotateExpressionNode(AST.lambdaExpression(params, bodyExpr, returnType, undefined, undefined, false), node) as LambdaExpression;
}

export function parseVerboseLambdaExpression(node: Node, source: string): LambdaExpression {
  if (node.type !== "verbose_lambda_expression") {
    throw new MapperError("parser: expected verbose lambda expression");
  }

  const ctx = getActiveParseContext();
  const params = parseParameterList(node.childForFieldName("parameters"), source, ctx);
  const returnType = ctx.parseReturnType(node.childForFieldName("return_type"));
  const generics = ctx.parseTypeParameters(node.childForFieldName("type_parameters"));
  const whereClause = ctx.parseWhereClause(node.childForFieldName("where_clause"));

  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: verbose lambda missing body");
  }
  const bodyExpr = ctx.parseBlock(bodyNode);

  return annotateExpressionNode(
    AST.lambdaExpression(params, bodyExpr, returnType, generics, whereClause, true),
    node,
  ) as LambdaExpression;
}

export function parseExpressionList(node: Node, _source: string): BlockExpression {
  const ctx = getActiveParseContext();
  const statements: Statement[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    const stmt = ctx.parseStatement(child);
    if (!stmt) continue;
    if (stmt.type === "LambdaExpression" && statements.length > 0) {
      const prev = statements[statements.length - 1];
      if (prev.type === "AssignmentExpression") {
        const rhs = prev.right;
        if (rhs.type === "FunctionCall") {
          if (rhs.arguments.length === 0 || rhs.arguments[rhs.arguments.length - 1] !== stmt) {
            rhs.arguments.push(stmt);
          }
          rhs.isTrailingLambda = true;
          continue;
        }
        if ((rhs as Expression).type) {
          const call = inheritMetadata(AST.functionCall(rhs as Expression, [], undefined, true), rhs as Expression, stmt);
          call.arguments.push(stmt);
          prev.right = call;
          continue;
        }
      }
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
  return annotateExpressionNode(AST.blockExpression(statements), node) as BlockExpression;
}

function parseLambdaParameter(node: Node, source: string): FunctionParameter {
  const nameNode = node.childForFieldName("name");
  if (!nameNode) {
    throw new MapperError("parser: lambda parameter missing name");
  }
  const id = parseIdentifier(nameNode, source);
  return annotate(AST.functionParameter(id), node) as FunctionParameter;
}

export function parseIfExpression(node: Node, _source: string): Expression {
  const ctx = getActiveParseContext();
  const conditionNode = node.childForFieldName("condition");
  if (!conditionNode) {
    throw new MapperError("parser: if expression missing condition");
  }
  const condition = ctx.parseExpression(conditionNode);

  const bodyNode = node.childForFieldName("consequence");
  if (!bodyNode) {
    throw new MapperError("parser: if expression missing body");
  }
  const body = ctx.parseBlock(bodyNode);

  const elseIfClauses: ElseIfClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child.type === "elsif_clause") {
      elseIfClauses.push(parseElseIfClause(child));
    }
  }

  const elseNode = node.childForFieldName("alternative");
  const elseBody = elseNode ? ctx.parseBlock(elseNode) : undefined;

  return annotateExpressionNode(AST.ifExpression(condition, body, elseIfClauses, elseBody), node);
}

function parseElseIfClause(node: Node): ElseIfClause {
  const ctx = getActiveParseContext();
  const bodyNode = node.childForFieldName("consequence");
  if (!bodyNode) {
    throw new MapperError("parser: elsif clause missing body");
  }
  const body = ctx.parseBlock(bodyNode);

  const conditionNode = node.childForFieldName("condition");
  if (!conditionNode) {
    throw new MapperError("parser: elsif clause missing condition");
  }
  const condition = ctx.parseExpression(conditionNode);

  return annotate(AST.elseIfClause(condition, body), node) as ElseIfClause;
}

export function parseMatchExpression(node: Node, _source: string): Expression {
  const ctx = getActiveParseContext();
  const subjectNode = node.childForFieldName("subject");
  if (!subjectNode) {
    throw new MapperError("parser: match expression missing subject");
  }
  const subject = ctx.parseExpression(subjectNode);

  const clauses: MatchClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child.type === "match_clause") {
      clauses.push(parseMatchClause(child));
    }
  }

  if (clauses.length === 0) {
    throw new MapperError("parser: match expression requires at least one clause");
  }

  return annotateExpressionNode(AST.matchExpression(subject, clauses), node);
}

function parseMatchClause(node: Node): MatchClause {
  const ctx = getActiveParseContext();
  const patternNode = node.childForFieldName("pattern");
  if (!patternNode) {
    throw new MapperError("parser: match clause missing pattern");
  }
  const pattern = ctx.parsePattern(patternNode);

  let guardExpr: Expression | undefined;
  const guardNode = node.childForFieldName("guard");
  if (guardNode) {
    const guardChild = firstNamedChild(guardNode);
    if (!guardChild) {
      throw new MapperError("parser: match guard missing expression");
    }
    guardExpr = ctx.parseExpression(guardChild);
  }

  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: match clause missing body");
  }

  let body: Expression;
  if (bodyNode.type === "block") {
    body = ctx.parseBlock(bodyNode);
  } else {
    body = ctx.parseExpression(bodyNode);
  }

  return annotate(AST.matchClause(pattern, body, guardExpr), node) as MatchClause;
}

export function parseHandlingExpression(node: Node, _source: string): Expression {
  if (node.type !== "handling_expression") {
    throw new MapperError("parser: expected handling_expression node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: handling expression missing base expression");
  }

  const ctx = getActiveParseContext();
  const baseExpr = ctx.parseExpression(node.namedChild(0));

  let current: Expression | undefined = baseExpr;
  let assignment: AssignmentExpression | undefined;
  if (baseExpr.type === "AssignmentExpression") {
    assignment = baseExpr;
    current = baseExpr.right;
  } else if (baseExpr.type === "PropagationExpression" && baseExpr.expression.type === "AssignmentExpression") {
    assignment = baseExpr.expression;
    baseExpr.expression = assignment.right;
    assignment.right = baseExpr;
    current = baseExpr;
  }

  for (let i = 1; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type !== "or_handler_clause") continue;
    const handlerNode = child.childForFieldName("handler");
    if (!handlerNode) {
      throw new MapperError("parser: or clause missing handler block");
    }
    const { block, binding } = parseHandlingBlock(handlerNode, ctx);
    current = annotateExpressionNode(AST.orElseExpression(current, block, binding), child);
  }

  if (assignment) {
    if (!current) {
      throw new MapperError("parser: or-else assignment missing right-hand expression");
    }
    assignment.right = current;
    return annotateExpressionNode(assignment, node);
  }

  if (!current) {
    throw new MapperError("parser: handling expression missing result");
  }

  return annotateExpressionNode(current, node);
}

function parseHandlingBlock(
  node: Node,
  ctx: ReturnType<typeof getActiveParseContext>,
): { block: BlockExpression; binding?: Identifier } {
  if (node.type !== "handling_block") {
    throw new MapperError("parser: expected handling_block node");
  }

  let binding: Identifier | undefined;
  const bindingNode = node.childForFieldName("binding");
  if (bindingNode) {
    binding = parseIdentifier(bindingNode, ctx.source);
  }

  const statements: Statement[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed) continue;
    const fieldName = node.fieldNameForChild(i);
    if (fieldName === "binding" && child.type === "identifier") continue;
    const stmt = ctx.parseStatement(child);
    if (stmt) {
      statements.push(stmt);
    }
  }

  return { block: annotateExpressionNode(AST.blockExpression(statements), node) as BlockExpression, binding };
}

export function parseRescueExpression(node: Node, _source: string): Expression {
  if (node.type !== "rescue_expression") {
    throw new MapperError("parser: expected rescue_expression node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: rescue expression missing monitored expression");
  }

  let monitoredNode: Node | null = null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type === "rescue_block") continue;
    monitoredNode = child;
    break;
  }

  if (!monitoredNode) {
    throw new MapperError("parser: rescue expression missing monitored expression");
  }

  const ctx = getActiveParseContext();
  const expr = ctx.parseExpression(monitoredNode);
  const rescueNode = node.childForFieldName("rescue");
  if (!rescueNode) {
    throw new MapperError("parser: rescue expression missing rescue block");
  }

  const clauses = parseRescueBlock(rescueNode);

  if (expr.type === "AssignmentExpression") {
    if (!expr.right) {
      throw new MapperError("parser: rescue assignment missing right-hand expression");
    }
    const rescueExpr = annotateExpressionNode(AST.rescueExpression(expr.right, clauses), node);
    expr.right = rescueExpr;
    return annotateExpressionNode(expr, node);
  }

  return annotateExpressionNode(AST.rescueExpression(expr, clauses), node);
}

function parseRescueBlock(node: Node): MatchClause[] {
  if (node.type !== "rescue_block") {
    throw new MapperError("parser: expected rescue_block node");
  }
  const clauses: MatchClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type !== "match_clause") continue;
    clauses.push(parseMatchClause(child));
  }
  if (clauses.length === 0) {
    throw new MapperError("parser: rescue block requires at least one clause");
  }
  return clauses;
}

export function parseEnsureExpression(node: Node, _source: string): Expression {
  if (node.type !== "ensure_expression") {
    throw new MapperError("parser: expected ensure_expression node");
  }

  let tryNode: Node | null = null;
  const ensureNode = node.childForFieldName("ensure");
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child === ensureNode) continue;
    tryNode = child;
    break;
  }

  if (!tryNode) {
    throw new MapperError("parser: ensure expression missing try expression");
  }

  const ctx = getActiveParseContext();
  const tryExpr = ctx.parseExpression(tryNode);
  if (!ensureNode) {
    throw new MapperError("parser: ensure expression missing ensure block");
  }
  const ensureBlock = ctx.parseBlock(ensureNode);

  if (tryExpr.type === "AssignmentExpression") {
    if (!tryExpr.right) {
      throw new MapperError("parser: ensure assignment missing right-hand expression");
    }
    const ensureExpr = annotateExpressionNode(AST.ensureExpression(tryExpr.right, ensureBlock), node);
    tryExpr.right = ensureExpr;
    return annotateExpressionNode(tryExpr, node);
  }

  return annotateExpressionNode(AST.ensureExpression(tryExpr, ensureBlock), node);
}

export function parseBreakpointExpression(node: Node, source: string): Expression {
  if (node.type !== "breakpoint_expression") {
    throw new MapperError("parser: expected breakpoint_expression node");
  }

  let labelNode = node.childForFieldName("label");
  if (!labelNode) {
    labelNode = fallbackBreakpointLabel(node);
  }
  if (!labelNode) {
    throw new MapperError("parser: breakpoint expression missing label");
  }
  let label: Identifier;
  if (labelNode.type === "label") {
    label = parseLabel(labelNode, source);
  } else {
    label = parseIdentifier(labelNode, source);
  }

  let bodyNode: Node | null = null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child && child.type === "block") {
      bodyNode = child;
      break;
    }
  }
  if (!bodyNode) {
    throw new MapperError("parser: breakpoint expression missing body");
  }

  const body = getActiveParseContext().parseBlock(bodyNode);
  return annotateExpressionNode(AST.breakpointExpression(label, body), node);
}

function fallbackBreakpointLabel(node: Node): Node | null {
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "identifier" || child.type === "label") {
      return child;
    }
    if (child.type === "ERROR" && child.childCount === 1) {
      const grand = child.child(0);
      if (grand && (grand.type === "identifier" || grand.type === "label")) {
        return grand;
      }
    }
  }
  return null;
}

export function parseDoExpression(node: Node, _source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: do expression missing body");
  }
  return getActiveParseContext().parseBlock(bodyNode);
}

export function parseSpawnExpression(node: Node, _source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: spawn expression missing body");
  }
  const expr = getActiveParseContext().parseExpression(bodyNode);
  if (expr.type !== "FunctionCall" && expr.type !== "BlockExpression") {
    throw new MapperError("parser: spawn expression requires function call or block");
  }
  return annotateExpressionNode(AST.spawnExpression(expr as FunctionCall | BlockExpression), node);
}
