import * as AST from "../ast";
import type {
  AssignmentExpression,
  Expression,
  FunctionCall,
  Identifier,
  IteratorLiteral,
  Pattern,
  TypeExpression,
} from "../ast";
import {
  annotateExpressionNode,
  firstNamedChild,
  findIdentifier,
  getActiveParseContext,
  isIgnorableNode,
  MapperError,
  MutableParseContext,
  Node,
  parseIdentifier,
  sliceText,
  withActiveNode,
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
import { parsePlaceholderExpression } from "./placeholders";
import {
  parseBreakpointExpression,
  parseDoExpression,
  parseEnsureExpression,
  parseExpressionList,
  parseHandlingExpression,
  parseIfExpression,
  parseLambdaExpression,
  parseMatchExpression,
  parseProcExpression,
  parseRescueExpression,
  parseSpawnExpression,
  parseVerboseLambdaExpression,
} from "./expressions_control";

export function registerExpressionParsers(ctx: MutableParseContext): void {
  ctx.parseExpression = withActiveNode((node) => parseExpression(node, ctx.source));
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
  ["multiplicative_expression", ["*", "/", "//", "%", "/%"]],
  ["exponent_expression", ["^"]],
]);

const ASSIGNMENT_OPERATORS = new Set([
  ":=",
  "=",
  "+=",
  "-=",
  "*=",
  "/=",
  "%=",
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
    case "keyword_identifier":
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
    case "expression_list":
      return parseExpressionList(node, source);
    case "loop_expression": {
      const bodyNode = firstNamedChild(node);
      const body = getActiveParseContext().parseBlock(bodyNode);
      return annotateExpressionNode(AST.loopExpression(body), node);
    }
    case "do_expression":
      return parseDoExpression(node, source);
    case "verbose_lambda_expression":
      return parseVerboseLambdaExpression(node, source);
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
    case "cast_expression":
      return parseCastExpression(node, source);
    case "unary_expression":
      return parseUnaryExpression(node, source);
    case "implicit_member_expression":
      return parseImplicitMemberExpression(node, source);
    case "placeholder_expression":
      return parsePlaceholderExpression(node, source);
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


function parseInterpolatedString(node: Node, source: string): Expression {
  const parts: (AST.StringLiteral | Expression)[] = [];
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    switch (child.type) {
      case "interpolation_text": {
        const text = unescapeInterpolationText(sliceText(child, source));
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

function unescapeInterpolationText(text: string): string {
  if (text.indexOf("\\") === -1) return text;
  let out = "";
  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    if (ch !== "\\") {
      out += ch;
      continue;
    }
    if (i + 1 >= text.length) {
      out += "\\";
      break;
    }
    const next = text[i + 1];
    switch (next) {
      case "`":
        out += "`";
        i += 1;
        break;
      case "$":
        out += "$";
        i += 1;
        break;
      case "\\":
        out += "\\";
        i += 1;
        break;
      default:
        out += `\\${next}`;
        i += 1;
        break;
    }
  }
  return out;
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

  const startNode = node.namedChild(0);
  let result = parseExpression(startNode, source);
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
        const expr = annotateExpressionNode(AST.memberAccessExpression(result, memberExpr, { isSafe }), suffix);
        applySpanFromNodes(expr, startNode, suffix);
        result = expr;
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
        const expr = annotateExpressionNode(AST.indexExpression(result, indexExpr), suffix);
        applySpanFromNodes(expr, startNode, suffix);
        result = expr;
        lastCall = undefined;
        break;
      }
      case "call_suffix": {
        const args = parseCallArguments(suffix, source);
        const typeArgs = pendingTypeArgs ?? undefined;
        pendingTypeArgs = null;
        const callExpr = annotateExpressionNode(AST.functionCall(result, args, typeArgs, false), suffix);
        applySpanFromNodes(callExpr, startNode, suffix);
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
            applySpanFromNodes(callExpr, startNode, suffix);
            callExpr.arguments.push(lambdaExpr);
            result.right = callExpr;
            lastCall = callExpr;
          }
          break;
        }
        if (lastCall && !lastCall.isTrailingLambda) {
          lastCall.arguments.push(lambdaExpr);
          lastCall.isTrailingLambda = true;
          applySpanFromNodes(lastCall, startNode, suffix);
          result = lastCall;
        } else {
          const callExpr = annotateExpressionNode(AST.functionCall(result, [], typeArgs, true), suffix);
          applySpanFromNodes(callExpr, startNode, suffix);
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
        const expr = annotateExpressionNode(AST.propagationExpression(result), suffix);
        applySpanFromNodes(expr, startNode, suffix);
        result = expr;
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
  let previous = node.namedChild(0);
  for (let i = 1; i < node.namedChildCount; i++) {
    const stepNode = node.namedChild(i);
    const stepExpr = parseExpression(stepNode, source);
    const expr = annotateExpressionNode(AST.binaryExpression(operator, result, stepExpr), stepNode);
    applySpanFromNodes(expr, previous, stepNode);
    result = expr;
    previous = stepNode;
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
    const expr = annotateExpressionNode(AST.binaryExpression(operator, result, rightExpr), rightNode);
    applySpanFromNodes(expr, previous, rightNode);
    result = expr;
    previous = rightNode;
  }
  return annotateExpressionNode(result, node);
}

function applySpanFromNodes(
  expr: Expression | null | undefined,
  startNode: Node | null | undefined,
  endNode: Node | null | undefined,
): void {
  if (!expr || !startNode || !endNode) return;
  const start = startNode.startPosition;
  const end = endNode.endPosition;
  (expr as { span?: AST.Span }).span = {
    start: { line: start.row + 1, column: start.column + 1 },
    end: { line: end.row + 1, column: end.column + 1 },
  };
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
      if (expr.type === "MemberAccessExpression" || expr.type === "IndexExpression" || expr.type === "ImplicitMemberExpression") {
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
  if (operatorText === "-") {
    if (operand.type === "IntegerLiteral" && operand.integerType) {
      const value = operand.value;
      const negated = typeof value === "bigint" ? -value : -Number(value);
      return annotateExpressionNode(AST.integerLiteral(negated, operand.integerType), node);
    }
    if (operand.type === "FloatLiteral" && operand.floatType) {
      return annotateExpressionNode(AST.floatLiteral(-operand.value, operand.floatType), node);
    }
  }
  if (operatorText === "-" || operatorText === "!" || operatorText === ".~") {
    return annotateExpressionNode(AST.unaryExpression(operatorText as "-" | "!" | ".~", operand), node);
  }
  throw new MapperError(`parser: unsupported unary operator ${operatorText}`);
}

function parseCastExpression(node: Node, source: string): Expression {
  if (node.namedChildCount < 2) {
    const child = firstNamedChild(node);
    if (child) return parseExpression(child, source);
    throw new MapperError("parser: cast expression missing target type");
  }
  const ctx = getActiveParseContext();
  const baseExpr = parseExpression(node.namedChild(0), source);
  let result: Expression = baseExpr;
  for (let i = 1; i < node.namedChildCount; i++) {
    const typeNode = node.namedChild(i);
    if (!typeNode) continue;
    const targetType = ctx.parseTypeExpression(typeNode);
    if (!targetType) {
      throw new MapperError("parser: cast expression missing target type");
    }
    result = AST.typeCastExpression(result, targetType);
  }
  return annotateExpressionNode(result, node);
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
