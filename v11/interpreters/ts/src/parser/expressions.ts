import * as AST from "../ast";
import type {
  AssignmentExpression,
  BlockExpression,
  Expression,
  FunctionCall,
  Identifier,
  IteratorLiteral,
  LambdaExpression,
  MatchClause,
  OrClause,
  Pattern,
  Statement,
  TypeExpression,
} from "../ast";
import {
  annotate,
  annotateExpressionNode,
  annotateStatement,
  firstNamedChild,
  findIdentifier,
  getActiveParseContext,
  inheritMetadata,
  isIgnorableNode,
  MapperError,
  MutableParseContext,
  Node,
  parseIdentifier,
  parseLabel,
  sliceText,
} from "./shared";
import {
  parseArrayLiteral,
  parseBooleanLiteral,
  parseCharLiteral,
  parseMapLiteral,
  parseNilLiteral,
  parseNumberLiteral,
  parseStringLiteral,
  parseStructLiteral,
} from "./literals";

export function registerExpressionParsers(ctx: MutableParseContext): void {
  ctx.parseExpression = node => parseExpression(node, ctx.source);
}

const INFIX_OPERATOR_SETS = new Map<string, string[]>([
  ["logical_or_expression", ["||"]],
  ["logical_and_expression", ["&&"]],
  ["bitwise_or_expression", [".|"]],
  ["bitwise_xor_expression", [".^"]],
  ["bitwise_and_expression", [".&"]],
  ["equality_expression", ["==", "!="]],
  ["comparison_expression", [">", "<", ">=", "<="]],
  ["shift_expression", [".<<", ".>>"]],
  ["additive_expression", ["+", "-"]],
  ["multiplicative_expression", ["*", "/", "//", "%%", "/%"]],
  ["exponent_expression", ["^"]],
]);

const ASSIGNMENT_OPERATORS = new Set([
  ":=",
  "=",
  "+=",
  "-=",
  "*=",
  "/=",
  ".&=",
  ".|=",
  ".^=",
  ".<<=",
  ".>>=",
]);
function parseExpression(node: Node | null | undefined, source: string): Expression {
  if (!node) {
    throw new MapperError("parser: nil expression node");
  }

  switch (node.type) {
    case "identifier":
      return parseIdentifier(node, source);
    case "number_literal":
      return parseNumberLiteral(getActiveParseContext(), node);
    case "boolean_literal":
      return parseBooleanLiteral(getActiveParseContext(), node);
    case "nil_literal":
      return parseNilLiteral(getActiveParseContext(), node);
    case "string_literal":
      return parseStringLiteral(getActiveParseContext(), node);
    case "character_literal":
      return parseCharLiteral(getActiveParseContext(), node);
    case "array_literal":
      return parseArrayLiteral(getActiveParseContext(), node);
    case "map_literal":
      return parseMapLiteral(getActiveParseContext(), node);
    case "struct_literal":
      return parseStructLiteral(getActiveParseContext(), node);
    case "block":
      return getActiveParseContext().parseBlock(node);
    case "loop_expression": {
      const bodyNode = firstNamedChild(node);
      const body = getActiveParseContext().parseBlock(bodyNode);
      return annotateExpressionNode(AST.loopExpression(body), node);
    }
    case "do_expression":
      return parseDoExpression(node, source);
    case "lambda_expression":
      return parseLambdaExpression(node, source);
    case "postfix_expression":
    case "call_target":
    case "rescue_postfix_expression":
      return parsePostfixExpression(node, source);
    case "member_access": {
      if (node.namedChildCount < 2) {
        throw new MapperError("parser: malformed member access");
      }
      const objectExpr = parseExpression(node.namedChild(0), source);
      const memberExpr = parseExpression(node.namedChild(1), source);
      const operatorNode = node.childForFieldName("operator");
      const operatorText = operatorNode ? sliceText(operatorNode, source).trim() : ".";
      const isSafe = operatorText === "?.";
      return annotateExpressionNode(AST.memberAccessExpression(objectExpr, memberExpr, { isSafe }), node);
    }
    case "proc_expression":
      return parseProcExpression(node, source);
    case "spawn_expression":
      return parseSpawnExpression(node, source);
    case "await_expression":
      return parseAwaitExpression(node, source);
    case "breakpoint_expression":
      return parseBreakpointExpression(node, source);
    case "handling_expression":
      return parseHandlingExpression(node, source);
    case "rescue_expression":
      return parseRescueExpression(node, source);
    case "ensure_expression":
      return parseEnsureExpression(node, source);
    case "if_expression":
      return parseIfExpression(node, source);
    case "match_expression":
      return parseMatchExpression(node, source);
    case "range_expression":
      return parseRangeExpression(node, source);
    case "assignment_expression":
      return parseAssignmentExpression(node, source);
    case "unary_expression":
      return parseUnaryExpression(node, source);
    case "implicit_member_expression":
      return parseImplicitMemberExpression(node, source);
    case "placeholder_expression":
      return parsePlaceholderExpression(node, source);
    case "topic_reference":
    return annotateExpressionNode(AST.topicReferenceExpression(), node);
    case "interpolated_string":
      return parseInterpolatedString(node, source);
    case "iterator_literal":
      return parseIteratorLiteral(node, source);
    case "parenthesized_expression": {
      const child = firstNamedChild(node);
      if (child) {
        return parseExpression(child, source);
      }
      throw new MapperError("parser: empty parenthesized expression");
    }
    case "pipe_expression":
      return parsePipeChain(node, source, "|>");
    case "low_precedence_pipe_expression":
      return parsePipeChain(node, source, "|>>");
    case "matchable_expression": {
      const child = firstNamedChild(node);
      if (child) return parseExpression(child, source);
      break;
    }
    case "pipe_operand_base": {
      const child = firstNamedChild(node);
      if (child) return parseExpression(child, source);
      break;
    }
  }

  if (INFIX_OPERATOR_SETS.has(node.type)) {
    const operators = INFIX_OPERATOR_SETS.get(node.type)!;
    return parseInfixExpression(node, source, operators);
  }

  const child = firstNamedChild(node);
  if (child && child !== node) {
    return parseExpression(child, source);
  }

  const fallbackId = findIdentifier(node, source);
  if (fallbackId) {
    return fallbackId;
  }

  throw new MapperError(`parser: unsupported expression kind ${node.type}`);
}


function parseImplicitMemberExpression(node: Node, source: string): Expression {
  const memberNode = node.childForFieldName("member") ?? firstNamedChild(node);
  if (!memberNode) {
    throw new MapperError("parser: implicit member missing identifier");
  }
  const member = parseIdentifier(memberNode, source);
  return annotateExpressionNode(AST.implicitMemberExpression(member), node);
}


function parsePlaceholderExpression(node: Node, source: string): Expression {
  const raw = sliceText(node, source).trim();
  if (raw === "@" || raw === "@0") {
    return annotateExpressionNode(AST.placeholderExpression(), node);
  }
  if (raw.startsWith("@")) {
    const value = raw.slice(1);
    if (value === "") {
      return annotateExpressionNode(AST.placeholderExpression(), node);
    }
    const index = Number.parseInt(value, 10);
    if (!Number.isInteger(index) || index <= 0) {
      throw new MapperError(`parser: invalid placeholder index ${raw}`);
    }
    return annotateExpressionNode(AST.placeholderExpression(index), node);
  }
  throw new MapperError(`parser: unsupported placeholder token ${raw}`);
}


function parseInterpolatedString(node: Node, source: string): Expression {
  const parts: (AST.StringLiteral | Expression)[] = [];
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    switch (child.type) {
      case "interpolation_text": {
        const text = sliceText(child, source);
        if (text !== "") {
          parts.push(annotateExpressionNode(AST.stringLiteral(text), child) as AST.StringLiteral);
        }
        break;
      }
      case "string_interpolation": {
        const exprNode = child.childForFieldName("expression");
        if (!exprNode) {
          throw new MapperError("parser: interpolation missing expression");
        }
        parts.push(parseExpression(exprNode, source));
        break;
      }
      default:
        break;
    }
  }
  return annotateExpressionNode(AST.stringInterpolation(parts), node);
}


function parseIteratorLiteral(node: Node, source: string): IteratorLiteral {
  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: iterator literal missing body");
  }

  let binding: Identifier | undefined;
  const bindingNode = bodyNode.childForFieldName("binding");
  if (bindingNode) {
    binding = parseIdentifier(bindingNode, source);
  }

  const elementTypeNode = node.childForFieldName("element_type");
  const elementType = elementTypeNode ? getActiveParseContext().parseTypeExpression(elementTypeNode) : undefined;

  const block = getActiveParseContext().parseBlock(bodyNode);
  const literal = AST.iteratorLiteral(block.body, binding, elementType);
  return annotateExpressionNode(literal, node);
}

function parseAwaitExpression(node: Node, source: string): Expression {
  const expressionNode = firstNamedChild(node);
  if (!expressionNode) {
    throw new MapperError("parser: await expression missing operand");
  }
  const awaited = parseExpression(expressionNode, source);
  return annotateExpressionNode(AST.awaitExpression(awaited), node);
}


function parsePostfixExpression(node: Node, source: string): Expression {
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty postfix expression");
  }

  let result = parseExpression(node.namedChild(0), source);
  let pendingTypeArgs: TypeExpression[] | null = null;
  let lastCall: FunctionCall | undefined;

  for (let i = 1; i < node.namedChildCount; i++) {
    const suffix = node.namedChild(i);
    switch (suffix.type) {
      case "member_access": {
        let memberNode = suffix.childForFieldName("member") ?? firstNamedChild(suffix);
        if (!memberNode) {
          throw new MapperError("parser: member access missing member");
        }

        let memberExpr: Expression;
        if (memberNode.type === "numeric_member") {
          const valueText = sliceText(memberNode, source);
          if (valueText === "") {
            throw new MapperError("parser: empty numeric member access");
          }
          const intValue = Number.parseInt(valueText, 10);
          if (!Number.isInteger(intValue)) {
            throw new MapperError(`parser: invalid numeric member ${valueText}`);
          }
          memberExpr = annotateExpressionNode(AST.integerLiteral(intValue), memberNode);
        } else {
          memberExpr = parseExpression(memberNode, source);
        }
        const operatorNode = suffix.childForFieldName("operator");
        const operatorText = operatorNode ? sliceText(operatorNode, source).trim() : ".";
        const isSafe = operatorText === "?.";
        result = annotateExpressionNode(AST.memberAccessExpression(result, memberExpr, { isSafe }), suffix);
        lastCall = undefined;
        break;
      }
      case "type_arguments": {
        pendingTypeArgs = getActiveParseContext().parseTypeArgumentList(suffix);
        lastCall = undefined;
        break;
      }
      case "index_suffix": {
        if (suffix.namedChildCount === 0) {
          throw new MapperError("parser: index expression missing index value");
        }
        if (suffix.namedChildCount > 1) {
          throw new MapperError("parser: slice expressions are not supported yet");
        }
        const indexExpr = parseExpression(suffix.namedChild(0), source);
        result = annotateExpressionNode(AST.indexExpression(result, indexExpr), suffix);
        lastCall = undefined;
        break;
      }
      case "call_suffix": {
        const args = parseCallArguments(suffix, source);
        const typeArgs = pendingTypeArgs ?? undefined;
        pendingTypeArgs = null;
        const callExpr = annotateExpressionNode(AST.functionCall(result, args, typeArgs, false), suffix);
        result = callExpr;
        lastCall = callExpr;
        break;
      }
      case "lambda_expression": {
        const lambdaExpr = parseLambdaExpression(suffix, source);
        const typeArgs = pendingTypeArgs ?? undefined;
        pendingTypeArgs = null;
        if (result.type === "AssignmentExpression") {
          const rhs = result.right;
          if (rhs.type === "FunctionCall" && !rhs.isTrailingLambda) {
            rhs.arguments.push(lambdaExpr);
            rhs.isTrailingLambda = true;
            lastCall = rhs;
          } else {
            const callExpr = annotateExpressionNode(AST.functionCall(rhs, [], typeArgs, true), suffix);
            callExpr.arguments.push(lambdaExpr);
            result.right = callExpr;
            lastCall = callExpr;
          }
          break;
        }
        if (lastCall && !lastCall.isTrailingLambda) {
          lastCall.arguments.push(lambdaExpr);
          lastCall.isTrailingLambda = true;
          result = lastCall;
        } else {
          const callExpr = annotateExpressionNode(AST.functionCall(result, [], typeArgs, true), suffix);
          callExpr.arguments.push(lambdaExpr);
          result = callExpr;
          lastCall = callExpr;
        }
        break;
      }
      case "propagate_suffix": {
        if (pendingTypeArgs && pendingTypeArgs.length > 0) {
          throw new MapperError("parser: dangling type arguments before propagation");
        }
        result = annotateExpressionNode(AST.propagationExpression(result), suffix);
        lastCall = undefined;
        break;
      }
      default:
        throw new MapperError(`parser: unsupported postfix suffix ${suffix.type}`);
    }
  }

  if (pendingTypeArgs && pendingTypeArgs.length > 0) {
    throw new MapperError("parser: dangling type arguments in expression");
  }

  return annotateExpressionNode(result, node);
}


function parseCallArguments(node: Node, source: string): Expression[] {
  const args: Expression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    args.push(parseExpression(child, source));
  }
  return args;
}


function parsePipeChain(node: Node, source: string, operator: string): Expression {
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty pipe expression");
  }
  let result = parseExpression(node.namedChild(0), source);
  for (let i = 1; i < node.namedChildCount; i++) {
    const stepNode = node.namedChild(i);
    const stepExpr = parseExpression(stepNode, source);
    result = annotateExpressionNode(AST.binaryExpression(operator, result, stepExpr), stepNode);
  }
  return annotateExpressionNode(result, node);
}


function parseInfixExpression(node: Node, source: string, operators: string[]): Expression {
  const count = node.namedChildCount;
  if (count === 0) {
    throw new MapperError(`parser: empty ${node.type}`);
  }
  if (count === 1) {
    return parseExpression(node.namedChild(0), source);
  }
  let result = parseExpression(node.namedChild(0), source);
  let previous = node.namedChild(0);
  for (let i = 1; i < count; i++) {
    const rightNode = node.namedChild(i);
    const rightExpr = parseExpression(rightNode, source);
    const operator = extractOperatorBetween(previous, rightNode, source, operators);
    if (!operator) {
      throw new MapperError(`parser: could not determine operator between operands in ${node.type}`);
    }
    result = annotateExpressionNode(AST.binaryExpression(operator, result, rightExpr), rightNode);
    previous = rightNode;
  }
  return annotateExpressionNode(result, node);
}


function extractOperatorBetween(left: Node | null, right: Node | null, source: string, allowed: string[]): string {
  if (!left || !right) return "";
  const start = left.endIndex;
  const end = right.startIndex;
  if (start < 0 || end < start || end > source.length) {
    return "";
  }
  const segment = source.slice(start, end).trim();
  if (segment === "") {
    return "";
  }
  for (const op of allowed) {
    if (segment === op) return op;
  }
  for (const op of allowed) {
    if (segment.includes(op)) return op;
  }
  return "";
}

function parseAssignmentExpression(node: Node, source: string): Expression {
  const operatorNode = node.childForFieldName("operator");
  if (!operatorNode) {
    const child = firstNamedChild(node);
    if (!child) {
      throw new MapperError("parser: empty assignment expression");
    }
    return parseExpression(child, source);
  }
  const leftNode = node.childForFieldName("left");
  const rightNode = node.childForFieldName("right");
  if (!leftNode || !rightNode) {
    throw new MapperError("parser: malformed assignment expression");
  }
  const left = parseAssignmentTarget(leftNode, source);
  let right = parseExpression(rightNode, source);
  // Trailing lambdas after assignments should bind to the right-hand call.
  const trailingLambdaNode = node.namedChildren.find(
    (child) => child !== leftNode && child !== rightNode && child !== operatorNode && child.type === "lambda_expression",
  );
  if (trailingLambdaNode) {
    const lambdaExpr = parseLambdaExpression(trailingLambdaNode, source);
    if (right.type === "FunctionCall" && !right.isTrailingLambda) {
      right.arguments.push(lambdaExpr);
      right.isTrailingLambda = true;
    } else {
      const callExpr = annotateExpressionNode(AST.functionCall(right, [], undefined, true), trailingLambdaNode);
      callExpr.arguments.push(lambdaExpr);
      right = callExpr;
    }
  }
  const operatorText = sliceText(operatorNode, source).trim();
  if (!ASSIGNMENT_OPERATORS.has(operatorText)) {
    throw new MapperError(`parser: unsupported assignment operator ${operatorText}`);
  }
  return annotateExpressionNode(
    AST.assignmentExpression(operatorText as AssignmentExpression["operator"], left, right),
    node,
  );
}


function parseAssignmentTarget(node: Node, source: string): AssignmentExpression["left"] {
  switch (node.type) {
    case "assignment_target": {
      const child = firstNamedChild(node);
      if (!child) {
        throw new MapperError("parser: empty assignment target");
      }
      return parseAssignmentTarget(child, source);
    }
    case "pattern":
    case "pattern_base":
    case "typed_pattern":
    case "struct_pattern":
    case "array_pattern":
      return getActiveParseContext().parsePattern(node);
    default: {
      const expr = parseExpression(node, source);
      if (expr.type === "MemberAccessExpression" || expr.type === "IndexExpression") {
        return expr;
      }
      if (
        expr.type === "Identifier" ||
        expr.type === "StructPattern" ||
        expr.type === "ArrayPattern" ||
        expr.type === "TypedPattern" ||
        expr.type === "WildcardPattern" ||
        expr.type === "LiteralPattern"
      ) {
        return expr as Pattern;
      }
      throw new MapperError(`parser: expression cannot be used as assignment target: ${expr.type}`);
    }
  }
}


function parseUnaryExpression(node: Node, source: string): Expression {
  const operandNode = firstNamedChild(node);
  if (!operandNode) {
    throw new MapperError("parser: unary expression missing operand");
  }
  if (node.startIndex === operandNode.startIndex) {
    return parseExpression(operandNode, source);
  }
  const operatorText = source.slice(node.startIndex, operandNode.startIndex).trim();
  if (operatorText === "") {
    return parseExpression(operandNode, source);
  }
  const operand = parseExpression(operandNode, source);
  if (operatorText === "-" || operatorText === "!" || operatorText === ".~") {
    return annotateExpressionNode(AST.unaryExpression(operatorText as "-" | "!" | ".~", operand), node);
  }
  throw new MapperError(`parser: unsupported unary operator ${operatorText}`);
}


function parseRangeExpression(node: Node, source: string): Expression {
  const operatorNode = node.childForFieldName("operator");
  if (!operatorNode || node.namedChildCount < 2) {
    const child = firstNamedChild(node);
    if (child) {
      return parseExpression(child, source);
    }
    throw new MapperError("parser: malformed range expression");
  }
  const startExpr = parseExpression(node.namedChild(0), source);
  const endExpr = parseExpression(node.namedChild(1), source);
  const operatorText = sliceText(operatorNode, source).trim();
  if (operatorText !== ".." && operatorText !== "...") {
    throw new MapperError(`parser: unsupported range operator ${operatorText}`);
  }
  return annotateExpressionNode(AST.rangeExpression(startExpr, endExpr, operatorText === ".."), node);
}


function parseLambdaExpression(node: Node, source: string): LambdaExpression {
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
  const bodyExpr = parseExpression(bodyNode, source);

  return annotateExpressionNode(AST.lambdaExpression(params, bodyExpr, returnType, undefined, undefined, false), node) as LambdaExpression;
}


function parseLambdaParameter(node: Node, source: string): FunctionParameter {
  const nameNode = node.childForFieldName("name");
  if (!nameNode) {
    throw new MapperError("parser: lambda parameter missing name");
  }
  const id = parseIdentifier(nameNode, source);
  return annotate(AST.functionParameter(id), node) as FunctionParameter;
}


function parseIfExpression(node: Node, source: string): Expression {
  const conditionNode = node.childForFieldName("condition");
  if (!conditionNode) {
    throw new MapperError("parser: if expression missing condition");
  }
  const condition = parseExpression(conditionNode, source);

  const bodyNode = node.childForFieldName("consequence");
  if (!bodyNode) {
    throw new MapperError("parser: if expression missing body");
  }
  const body = getActiveParseContext().parseBlock(bodyNode);

  const clauses: OrClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child.type === "or_clause") {
      clauses.push(parseOrClause(child, source));
    }
  }

  const elseNode = findElseBlock(node, bodyNode);
  if (elseNode) {
    const elseBody = getActiveParseContext().parseBlock(elseNode);
    clauses.push(annotate(AST.orClause(elseBody, undefined), elseNode) as OrClause);
  }

  return annotateExpressionNode(AST.ifExpression(condition, body, clauses), node);
}


function parseOrClause(node: Node, source: string): OrClause {
  const bodyNode = node.childForFieldName("consequence");
  if (!bodyNode) {
    throw new MapperError("parser: or clause missing body");
  }
  const body = getActiveParseContext().parseBlock(bodyNode);

  const conditionNode = node.childForFieldName("condition");
  let condition: Expression | undefined;
  if (conditionNode) {
    condition = parseExpression(conditionNode, source);
  }

  return annotate(AST.orClause(body, condition), node) as OrClause;
}


function findElseBlock(ifNode: Node, consequence: Node): Node | null {
  const consequenceStart = consequence.startIndex;
  const consequenceEnd = consequence.endIndex;
  for (let i = 0; i < ifNode.namedChildCount; i++) {
    const child = ifNode.namedChild(i);
    if (child.type !== "block") continue;
    if (child.startIndex === consequenceStart && child.endIndex === consequenceEnd) {
      continue;
    }
    return child;
  }
  return null;
}


function parseMatchExpression(node: Node, source: string): Expression {
  const subjectNode = node.childForFieldName("subject");
  if (!subjectNode) {
    throw new MapperError("parser: match expression missing subject");
  }
  const subject = parseExpression(subjectNode, source);

  const clauses: MatchClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child.type === "match_clause") {
      clauses.push(parseMatchClause(child, source));
    }
  }

  if (clauses.length === 0) {
    throw new MapperError("parser: match expression requires at least one clause");
  }

  return annotateExpressionNode(AST.matchExpression(subject, clauses), node);
}


function parseMatchClause(node: Node, source: string): MatchClause {
  const patternNode = node.childForFieldName("pattern");
  if (!patternNode) {
    throw new MapperError("parser: match clause missing pattern");
  }
  const pattern = getActiveParseContext().parsePattern(patternNode);

  let guardExpr: Expression | undefined;
  const guardNode = node.childForFieldName("guard");
  if (guardNode) {
    const guardChild = firstNamedChild(guardNode);
    if (!guardChild) {
      throw new MapperError("parser: match guard missing expression");
    }
    guardExpr = parseExpression(guardChild, source);
  }

  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: match clause missing body");
  }

  let body: Expression;
  if (bodyNode.type === "block") {
    body = getActiveParseContext().parseBlock(bodyNode);
  } else {
    body = parseExpression(bodyNode, source);
  }

  return annotate(AST.matchClause(pattern, body, guardExpr), node) as MatchClause;
}


function parseHandlingExpression(node: Node, source: string): Expression {
  if (node.type !== "handling_expression") {
    throw new MapperError("parser: expected handling_expression node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: handling expression missing base expression");
  }
  const baseExpr = parseExpression(node.namedChild(0), source);

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
    if (!child || child.type !== "else_clause") continue;
    const handlerNode = child.childForFieldName("handler");
    if (!handlerNode) {
      throw new MapperError("parser: else clause missing handler block");
    }
    const { block, binding } = parseHandlingBlock(handlerNode, source);
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


function parseHandlingBlock(node: Node, source: string): { block: BlockExpression; binding?: Identifier } {
  if (node.type !== "handling_block") {
    throw new MapperError("parser: expected handling_block node");
  }

  let binding: Identifier | undefined;
  const bindingNode = node.childForFieldName("binding");
  if (bindingNode) {
    binding = parseIdentifier(bindingNode, source);
  }

  const statements: Statement[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed) continue;
    const fieldName = node.fieldNameForChild(i);
    if (fieldName === "binding" && child.type === "identifier") continue;
    const stmt = getActiveParseContext().parseStatement(child);
    if (stmt) {
      statements.push(stmt);
    }
  }

  return { block: annotateExpressionNode(AST.blockExpression(statements), node) as BlockExpression, binding };
}


function parseRescueExpression(node: Node, source: string): Expression {
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

  const expr = parseExpression(monitoredNode, source);
  const rescueNode = node.childForFieldName("rescue");
  if (!rescueNode) {
    throw new MapperError("parser: rescue expression missing rescue block");
  }

  const clauses = parseRescueBlock(rescueNode, source);

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


function parseRescueBlock(node: Node, source: string): MatchClause[] {
  if (node.type !== "rescue_block") {
    throw new MapperError("parser: expected rescue_block node");
  }
  const clauses: MatchClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type !== "match_clause") continue;
    clauses.push(parseMatchClause(child, source));
  }
  if (clauses.length === 0) {
    throw new MapperError("parser: rescue block requires at least one clause");
  }
  return clauses;
}


function parseEnsureExpression(node: Node, source: string): Expression {
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

  const tryExpr = parseExpression(tryNode, source);
  if (!ensureNode) {
    throw new MapperError("parser: ensure expression missing ensure block");
  }
  const ensureBlock = getActiveParseContext().parseBlock(ensureNode);

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


function parseBreakpointExpression(node: Node, source: string): Expression {
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


function parseDoExpression(node: Node, source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: do expression missing body");
  }
  return getActiveParseContext().parseBlock(bodyNode);
}


function parseProcExpression(node: Node, source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: proc expression missing body");
  }
  const expr = parseExpression(bodyNode, source);
  if (expr.type !== "FunctionCall" && expr.type !== "BlockExpression") {
    throw new MapperError("parser: proc expression requires function call or block");
  }
  return annotateExpressionNode(AST.procExpression(expr as FunctionCall | BlockExpression), node);
}


function parseSpawnExpression(node: Node, source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: spawn expression missing body");
  }
  const expr = parseExpression(bodyNode, source);
  if (expr.type !== "FunctionCall" && expr.type !== "BlockExpression") {
    throw new MapperError("parser: spawn expression requires function call or block");
  }
  return annotateExpressionNode(AST.spawnExpression(expr as FunctionCall | BlockExpression), node);
}

// --- Numerous helpers omitted for brevity; the completed mapper will port every parser helper from Go. ---
