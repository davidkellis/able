package parser

import (
	"able/interpreter10-go/pkg/ast"
	"fmt"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"strconv"
	"strings"
)

var infixOperatorSets = map[string][]string{
	"logical_or_expression":        {"||"},
	"logical_and_expression":       {"&&"},
	"bitwise_or_expression":        {".|"},
	"bitwise_xor_expression":       {".^"},
	"bitwise_and_expression":       {".&"},
	"equality_expression":          {"==", "!="},
	"comparison_expression":        {">", "<", ">=", "<="},
	"shift_expression":             {".<<", ".>>"},
	"additive_expression":          {"+", "-"},
	"multiplicative_expression":    {"*", "/", "//", "%%", "/%"},
	"exponent_expression":          {"^"},
	"topic_placeholder_expression": {"%"},
}

var assignmentOperatorMap = map[string]ast.AssignmentOperator{
	":=":     ast.AssignmentDeclare,
	"=":      ast.AssignmentAssign,
	"+=":     ast.AssignmentAdd,
	"-=":     ast.AssignmentSub,
	"*=":     ast.AssignmentMul,
	"/=":     ast.AssignmentDiv,
	".&=":    ast.AssignmentBitAnd,
	".|=":    ast.AssignmentBitOr,
	".^=":    ast.AssignmentBitXor,
	".<<=":   ast.AssignmentShiftL,
	".>>=":   ast.AssignmentShiftR,
}

func parseExpressionInternal(ctx *parseContext, node *sitter.Node) (ast.Expression, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil expression node")
	}

	source := ctx.source

	switch node.Kind() {
	case "identifier":
		expr, err := parseIdentifier(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "number_literal":
		expr, err := ctx.parseNumberLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "boolean_literal":
		expr, err := ctx.parseBooleanLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "nil_literal":
		expr, err := ctx.parseNilLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "string_literal":
		expr, err := ctx.parseStringLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "character_literal":
		expr, err := ctx.parseCharLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "array_literal":
		expr, err := ctx.parseArrayLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "map_literal":
		expr, err := ctx.parseMapLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "struct_literal":
		expr, err := ctx.parseStructLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "block":
		expr, err := ctx.parseBlock(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "do_expression":
		expr, err := ctx.parseDoExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "lambda_expression":
		expr, err := ctx.parseLambdaExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "postfix_expression":
		expr, err := ctx.parsePostfixExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "call_target":
		expr, err := ctx.parsePostfixExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "member_access":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed member access")
		}
		objectExpr, err := parseExpressionInternal(ctx, node.NamedChild(0))
		if err != nil {
			return nil, err
		}
		memberExpr, err := parseExpressionInternal(ctx, node.NamedChild(1))
		if err != nil {
			return nil, err
		}
		memberAccess := ast.NewMemberAccessExpression(objectExpr, memberExpr)
		if opNode := node.ChildByFieldName("operator"); opNode != nil {
			operatorText := strings.TrimSpace(sliceContent(opNode, source))
			if operatorText == "?." {
				memberAccess.Safe = true
			}
		}
		return annotateExpression(memberAccess, node), nil
	case "proc_expression":
		expr, err := ctx.parseProcExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "spawn_expression":
		expr, err := ctx.parseSpawnExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "await_expression":
		expr, err := ctx.parseAwaitExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "loop_expression":
		expr, err := ctx.parseLoopExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "breakpoint_expression":
		expr, err := ctx.parseBreakpointExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "handling_expression":
		expr, err := ctx.parseHandlingExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "rescue_expression":
		expr, err := ctx.parseRescueExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "ensure_expression":
		expr, err := ctx.parseEnsureExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "if_expression":
		expr, err := ctx.parseIfExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "match_expression":
		expr, err := ctx.parseMatchExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "range_expression":
		expr, err := ctx.parseRangeExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "assignment_expression":
		expr, err := ctx.parseAssignmentExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "unary_expression":
		expr, err := ctx.parseUnaryExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "implicit_member_expression":
		expr, err := ctx.parseImplicitMemberExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "placeholder_expression":
		expr, err := ctx.parsePlaceholderExpression(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "topic_reference":
		return annotateExpression(ast.NewTopicReferenceExpression(), node), nil
	case "interpolated_string":
		expr, err := ctx.parseInterpolatedString(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "iterator_literal":
		expr, err := ctx.parseIteratorLiteral(node)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "parenthesized_expression":
		if child := firstNamedChild(node); child != nil {
			expr, err := parseExpressionInternal(ctx, child)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
		return nil, fmt.Errorf("parser: empty parenthesized expression")
	case "pipe_expression":
		expr, err := ctx.parsePipeExpression(node, "|>")
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "low_precedence_pipe_expression":
		expr, err := ctx.parsePipeExpression(node, "|>>")
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "matchable_expression":
		if child := firstNamedChild(node); child != nil {
			expr, err := parseExpressionInternal(ctx, child)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
	case "pipe_operand_base":
		if child := firstNamedChild(node); child != nil {
			expr, err := parseExpressionInternal(ctx, child)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
	}

	if operators, ok := infixOperatorSets[node.Kind()]; ok {
		expr, err := ctx.parseInfixExpression(node, operators)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}

	if child := firstNamedChild(node); child != nil && child != node {
		expr, err := parseExpressionInternal(ctx, child)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}

	if id, ok := findIdentifier(node, source); ok {
		return annotateExpression(id, node), nil
	}

	return nil, fmt.Errorf("parser: unsupported expression kind %q", node.Kind())
}

func (ctx *parseContext) parsePostfixExpression(node *sitter.Node) (ast.Expression, error) {
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty postfix expression")
	}

	source := ctx.source

	result, err := ctx.parseExpression(node.NamedChild(0))
	if err != nil {
		return nil, err
	}

	var (
		pendingTypeArgs []ast.TypeExpression
		lastCall        *ast.FunctionCall
	)

	for i := uint(1); i < node.NamedChildCount(); i++ {
		suffix := node.NamedChild(i)
		prev := result
		switch suffix.Kind() {
		case "member_access":
			memberNode := suffix.ChildByFieldName("member")
			if memberNode == nil {
				memberNode = firstNamedChild(suffix)
			}
			if memberNode == nil {
				return nil, fmt.Errorf("parser: member access missing member")
			}

			var (
				memberExpr ast.Expression
				err        error
			)

			if memberNode.Kind() == "numeric_member" {
				valueText := sliceContent(memberNode, source)
				if valueText == "" {
					return nil, fmt.Errorf("parser: empty numeric member access")
				}
				intValue, convErr := strconv.Atoi(valueText)
				if convErr != nil {
					return nil, fmt.Errorf("parser: invalid numeric member %q", valueText)
				}
				memberExpr = annotateExpression(ast.Int(int64(intValue)), memberNode)
			} else {
				memberExpr, err = ctx.parseExpression(memberNode)
				if err != nil {
					return nil, err
				}
			}
			memberAccess := ast.NewMemberAccessExpression(prev, memberExpr)
			if opNode := suffix.ChildByFieldName("operator"); opNode != nil {
				operatorText := strings.TrimSpace(sliceContent(opNode, source))
				if operatorText == "?." {
					memberAccess.Safe = true
				}
			}
			annotateCompositeExpression(memberAccess, prev, suffix)
			result = memberAccess
			lastCall = nil
		case "type_arguments":
			typeArgs, err := ctx.parseTypeArgumentList(suffix)
			if err != nil {
				return nil, err
			}
			pendingTypeArgs = typeArgs
			lastCall = nil
		case "index_suffix":
			if suffix.NamedChildCount() == 0 {
				return nil, fmt.Errorf("parser: index expression missing index value")
			}
			if suffix.NamedChildCount() > 1 {
				return nil, fmt.Errorf("parser: slice expressions are not supported yet")
			}
			indexExpr, err := ctx.parseExpression(suffix.NamedChild(0))
			if err != nil {
				return nil, err
			}
			indexed := ast.NewIndexExpression(prev, indexExpr)
			annotateCompositeExpression(indexed, prev, suffix)
			result = indexed
			lastCall = nil
		case "call_suffix":
			args, err := ctx.parseCallArguments(suffix)
			if err != nil {
				return nil, err
			}
			typeArgs := pendingTypeArgs
			pendingTypeArgs = nil

			callExpr := ast.NewFunctionCall(prev, args, typeArgs, false)
			annotateCompositeExpression(callExpr, prev, suffix)
			result = callExpr
			lastCall = callExpr
		case "lambda_expression":
			lambdaExpr, err := ctx.parseLambdaExpression(suffix)
			if err != nil {
				return nil, err
			}

			typeArgs := pendingTypeArgs
			pendingTypeArgs = nil

			if lastCall != nil && !lastCall.IsTrailingLambda {
				lastCall.Arguments = append(lastCall.Arguments, lambdaExpr)
				lastCall.IsTrailingLambda = true
				extendExpressionToNode(lastCall, suffix)
				result = lastCall
				break
			}

			// Bind trailing lambda to RHS when the current expression is an assignment whose
			// right-hand side is (or becomes) a call.
			if assign, ok := result.(*ast.AssignmentExpression); ok {
				switch rhs := assign.Right.(type) {
				case *ast.FunctionCall:
					if !rhs.IsTrailingLambda {
						rhs.Arguments = append(rhs.Arguments, lambdaExpr)
						rhs.IsTrailingLambda = true
						extendExpressionToNode(rhs, suffix)
						lastCall = rhs
						break
					}
				default:
					callExpr := ast.NewFunctionCall(rhs, nil, typeArgs, true)
					callExpr.Arguments = append(callExpr.Arguments, lambdaExpr)
					annotateCompositeExpression(callExpr, rhs, suffix)
					assign.Right = callExpr
					lastCall = callExpr
					break
				}
				break
			}

			callExpr := ast.NewFunctionCall(prev, nil, typeArgs, true)
			callExpr.Arguments = append(callExpr.Arguments, lambdaExpr)
			annotateCompositeExpression(callExpr, prev, suffix)
			result = callExpr
			lastCall = callExpr
		case "propagate_suffix":
			if len(pendingTypeArgs) > 0 {
				return nil, fmt.Errorf("parser: dangling type arguments before propagation")
			}
			prop := ast.NewPropagationExpression(prev)
			annotateCompositeExpression(prop, prev, suffix)
			result = prop
			lastCall = nil
		default:
			return nil, fmt.Errorf("parser: unsupported postfix suffix %q", suffix.Kind())
		}
	}

	if len(pendingTypeArgs) > 0 {
		return nil, fmt.Errorf("parser: dangling type arguments in expression")
	}

	return annotateExpression(result, node), nil
}

func (ctx *parseContext) parseCallArguments(node *sitter.Node) ([]ast.Expression, error) {
	args := make([]ast.Expression, 0)

	for j := uint(0); j < node.NamedChildCount(); j++ {
		child := node.NamedChild(j)
		if child == nil || !child.IsNamed() || isIgnorableNode(child) {
			continue
		}
		argExpr, err := ctx.parseExpression(child)
		if err != nil {
			return nil, err
		}
		args = append(args, argExpr)
	}

	return args, nil
}

func parseTypeArgumentList(node *sitter.Node, source []byte) ([]ast.TypeExpression, error) {
	if node == nil {
		return nil, nil
	}

	var args []ast.TypeExpression
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || !child.IsNamed() || isIgnorableNode(child) {
			continue
		}
		typeExpr := parseTypeExpression(child, source)
		if typeExpr == nil {
			return nil, fmt.Errorf("parser: unsupported type argument kind %q", child.Kind())
		}
		args = append(args, typeExpr)
	}

	return args, nil
}

func (ctx *parseContext) parseAssignmentExpression(node *sitter.Node) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil {
		child := firstNamedChild(node)
		if child == nil {
			return nil, fmt.Errorf("parser: empty assignment expression")
		}
		expr, err := ctx.parseExpression(child)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	leftNode := node.ChildByFieldName("left")
	rightNode := node.ChildByFieldName("right")
	if leftNode == nil || rightNode == nil {
		return nil, fmt.Errorf("parser: malformed assignment expression")
	}
	left, err := ctx.parseAssignmentTarget(leftNode)
	if err != nil {
		return nil, err
	}
	right, err := ctx.parseExpression(rightNode)
	if err != nil {
		return nil, err
	}
	operatorText := strings.TrimSpace(sliceContent(operatorNode, ctx.source))
	operator, err := mapAssignmentOperator(operatorText)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewAssignmentExpression(operator, left, right), node), nil
}

func (ctx *parseContext) parseAssignmentTarget(node *sitter.Node) (ast.AssignmentTarget, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil assignment target")
	}
	switch node.Kind() {
	case "assignment_target":
		if child := firstNamedChild(node); child != nil {
			return ctx.parseAssignmentTarget(child)
		}
		return nil, fmt.Errorf("parser: empty assignment target")
	case "pattern", "pattern_base", "typed_pattern", "struct_pattern", "array_pattern":
		pattern, err := ctx.parsePattern(node)
		if err != nil {
			return nil, err
		}
		target, ok := pattern.(ast.AssignmentTarget)
		if !ok {
			return nil, fmt.Errorf("parser: pattern cannot be used as assignment target: %T", pattern)
		}
		return target, nil
	default:
		expr, err := ctx.parseExpression(node)
		if err != nil {
			return nil, err
		}
		target, ok := expr.(ast.AssignmentTarget)
		if !ok {
			return nil, fmt.Errorf("parser: expression cannot be used as assignment target: %T", expr)
		}
		return target, nil
	}
}

func (ctx *parseContext) parseUnaryExpression(node *sitter.Node) (ast.Expression, error) {
	operandNode := firstNamedChild(node)
	if operandNode == nil {
		return nil, fmt.Errorf("parser: unary expression missing operand")
	}
	if int(node.StartByte()) == int(operandNode.StartByte()) {
		expr, err := ctx.parseExpression(operandNode)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	operatorText := strings.TrimSpace(string(ctx.source[int(node.StartByte()):int(operandNode.StartByte())]))
	if operatorText == "" {
		expr, err := ctx.parseExpression(operandNode)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	operand, err := ctx.parseExpression(operandNode)
	if err != nil {
		return nil, err
	}
	switch operatorText {
	case "-":
		return annotateExpression(ast.NewUnaryExpression(ast.UnaryOperatorNegate, operand), node), nil
	case "!":
		return annotateExpression(ast.NewUnaryExpression(ast.UnaryOperatorNot, operand), node), nil
	case ".~":
		return annotateExpression(ast.NewUnaryExpression(ast.UnaryOperatorBitNot, operand), node), nil
	default:
		return nil, fmt.Errorf("parser: unsupported unary operator %q", operatorText)
	}
}

func (ctx *parseContext) parsePipeExpression(node *sitter.Node, operator string) (ast.Expression, error) {
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty pipe expression")
	}
	result, err := ctx.parseExpression(node.NamedChild(0))
	if err != nil {
		return nil, err
	}
	for i := uint(1); i < node.NamedChildCount(); i++ {
		stepNode := node.NamedChild(i)
		stepExpr, err := ctx.parseExpression(stepNode)
		if err != nil {
			return nil, err
		}
		prev := result
		result = annotateCompositeExpression(ast.NewBinaryExpression(operator, result, stepExpr), prev, stepNode)
	}
	return annotateExpression(result, node), nil
}

func (ctx *parseContext) parseInfixExpression(node *sitter.Node, operators []string) (ast.Expression, error) {
	count := node.NamedChildCount()
	if count == 0 {
		return nil, fmt.Errorf("parser: empty %s", node.Kind())
	}
	if count == 1 {
		expr, err := ctx.parseExpression(node.NamedChild(0))
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	result, err := ctx.parseExpression(node.NamedChild(0))
	if err != nil {
		return nil, err
	}
	prevNode := node.NamedChild(0)
	for i := uint(1); i < count; i++ {
		rightNode := node.NamedChild(i)
		rightExpr, err := ctx.parseExpression(rightNode)
		if err != nil {
			return nil, err
		}
		operator := extractOperatorBetween(prevNode, rightNode, ctx.source, operators)
		if operator == "" {
			return nil, fmt.Errorf("parser: could not determine operator between operands in %s", node.Kind())
		}
		prev := result
		result = annotateCompositeExpression(ast.NewBinaryExpression(operator, result, rightExpr), prev, rightNode)
		prevNode = rightNode
	}
	return annotateExpression(result, node), nil
}

func extractOperatorBetween(left, right *sitter.Node, source []byte, allowed []string) string {
	if left == nil || right == nil {
		return ""
	}
	start := int(left.EndByte())
	end := int(right.StartByte())
	if start < 0 || end < start || end > len(source) {
		return ""
	}
	segment := strings.TrimSpace(string(source[start:end]))
	if segment == "" {
		return ""
	}
	for _, op := range allowed {
		if segment == op {
			return op
		}
	}
	for _, op := range allowed {
		if strings.Contains(segment, op) {
			return op
		}
	}
	return ""
}

func mapAssignmentOperator(op string) (ast.AssignmentOperator, error) {
	if operator, ok := assignmentOperatorMap[op]; ok {
		return operator, nil
	}
	return "", fmt.Errorf("parser: unsupported assignment operator %q", op)
}

func (ctx *parseContext) parseLambdaExpression(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "lambda_expression" {
		return nil, fmt.Errorf("parser: expected lambda expression")
	}

	var params []*ast.FunctionParameter
	if paramsNode := node.ChildByFieldName("parameters"); paramsNode != nil {
		for i := uint(0); i < paramsNode.NamedChildCount(); i++ {
			paramNode := paramsNode.NamedChild(i)
			if paramNode == nil || paramNode.Kind() != "lambda_parameter" {
				continue
			}
			param, err := parseLambdaParameter(paramNode, ctx.source)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
	}

	var returnType ast.TypeExpression
	if returnNode := node.ChildByFieldName("return_type"); returnNode != nil {
		returnType = ctx.parseReturnType(returnNode)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: lambda missing body")
	}

	bodyExpr, err := ctx.parseExpression(bodyNode)
	if err != nil {
		return nil, err
	}

	return annotateExpression(ast.NewLambdaExpression(params, bodyExpr, returnType, nil, nil, false), node), nil
}

func parseLambdaParameter(node *sitter.Node, source []byte) (*ast.FunctionParameter, error) {
	if node == nil || node.Kind() != "lambda_parameter" {
		return nil, fmt.Errorf("parser: expected lambda parameter")
	}

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil, fmt.Errorf("parser: lambda parameter missing name")
	}

	id, err := parseIdentifier(nameNode, source)
	if err != nil {
		return nil, err
	}

	param := ast.NewFunctionParameter(id, nil)
	annotateSpan(param, node)
	return param, nil
}

func (ctx *parseContext) parseExpression(node *sitter.Node) (ast.Expression, error) {
	return parseExpressionInternal(ctx, node)
}
