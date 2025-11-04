package parser

import (
	"fmt"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

var infixOperatorSets = map[string][]string{
	"logical_or_expression":     {"||"},
	"logical_and_expression":    {"&&"},
	"bitwise_or_expression":     {"|"},
	"bitwise_xor_expression":    {`\xor`},
	"bitwise_and_expression":    {"&"},
	"equality_expression":       {"==", "!="},
	"comparison_expression":     {">", "<", ">=", "<="},
	"shift_expression":          {"<<", ">>"},
	"additive_expression":       {"+", "-"},
	"multiplicative_expression": {"*", "/", "%"},
	"exponent_expression":       {"**"},
}

func parseImplicitMemberExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	memberNode := node.ChildByFieldName("member")
	if memberNode == nil {
		return nil, fmt.Errorf("parser: implicit member missing identifier")
	}
	member, err := parseIdentifier(memberNode, source)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewImplicitMemberExpression(member), node), nil
}

func parsePlaceholderExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	raw := strings.TrimSpace(sliceContent(node, source))
	if raw == "" {
		return nil, fmt.Errorf("parser: empty placeholder expression")
	}
	if raw == "@" {
		return annotateExpression(ast.NewPlaceholderExpression(nil), node), nil
	}
	if strings.HasPrefix(raw, "@") {
		value := raw[1:]
		if value == "" {
			return annotateExpression(ast.NewPlaceholderExpression(nil), node), nil
		}
		index, err := strconv.Atoi(value)
		if err != nil || index <= 0 {
			return nil, fmt.Errorf("parser: invalid placeholder index %q", raw)
		}
		return annotateExpression(ast.NewPlaceholderExpression(&index), node), nil
	}
	return nil, fmt.Errorf("parser: unsupported placeholder token %q", raw)
}

func parseInterpolatedString(node *sitter.Node, source []byte) (ast.Expression, error) {
	parts := make([]ast.Expression, 0)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		switch child.Kind() {
		case "interpolation_text":
			text := sliceContent(child, source)
			if text != "" {
				parts = append(parts, annotateExpression(ast.Str(text), child))
			}
		case "string_interpolation":
			exprNode := child.ChildByFieldName("expression")
			if exprNode == nil {
				return nil, fmt.Errorf("parser: interpolation missing expression")
			}
			expr, err := parseExpression(exprNode, source)
			if err != nil {
				return nil, err
			}
			parts = append(parts, expr)
		}
	}
	return annotateExpression(ast.NewStringInterpolation(parts), node), nil
}

func parseIteratorLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "iterator_literal" {
		return nil, fmt.Errorf("parser: expected iterator_literal node")
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: iterator literal missing body")
	}

	block, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}

	return annotateExpression(ast.NewIteratorLiteral(block.Body), node), nil
}

func parseExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil expression node")
	}

	switch node.Kind() {
	case "identifier":
		expr, err := parseIdentifier(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "number_literal":
		expr, err := parseNumberLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "boolean_literal":
		expr, err := parseBooleanLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "nil_literal":
		expr, err := parseNilLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "string_literal":
		expr, err := parseStringLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "character_literal":
		expr, err := parseCharLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "array_literal":
		expr, err := parseArrayLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "struct_literal":
		expr, err := parseStructLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "block":
		expr, err := parseBlock(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "do_expression":
		expr, err := parseDoExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "lambda_expression":
		expr, err := parseLambdaExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "postfix_expression":
		expr, err := parsePostfixExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "call_target":
		expr, err := parsePostfixExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "member_access":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed member access")
		}
		objectExpr, err := parseExpression(node.NamedChild(0), source)
		if err != nil {
			return nil, err
		}
		memberExpr, err := parseExpression(node.NamedChild(1), source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(ast.NewMemberAccessExpression(objectExpr, memberExpr), node), nil
	case "proc_expression":
		expr, err := parseProcExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "spawn_expression":
		expr, err := parseSpawnExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "breakpoint_expression":
		expr, err := parseBreakpointExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "handling_expression":
		expr, err := parseHandlingExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "rescue_expression":
		expr, err := parseRescueExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "ensure_expression":
		expr, err := parseEnsureExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "if_expression":
		expr, err := parseIfExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "match_expression":
		expr, err := parseMatchExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "range_expression":
		expr, err := parseRangeExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "assignment_expression":
		expr, err := parseAssignmentExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "unary_expression":
		expr, err := parseUnaryExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "implicit_member_expression":
		expr, err := parseImplicitMemberExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "placeholder_expression":
		expr, err := parsePlaceholderExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "topic_reference":
		return annotateExpression(ast.NewTopicReferenceExpression(), node), nil
	case "interpolated_string":
		expr, err := parseInterpolatedString(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "iterator_literal":
		expr, err := parseIteratorLiteral(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "parenthesized_expression":
		if child := firstNamedChild(node); child != nil {
			expr, err := parseExpression(child, source)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
		return nil, fmt.Errorf("parser: empty parenthesized expression")
	case "pipe_expression":
		expr, err := parsePipeExpression(node, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	case "matchable_expression":
		if child := firstNamedChild(node); child != nil {
			expr, err := parseExpression(child, source)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
	}

	if operators, ok := infixOperatorSets[node.Kind()]; ok {
		expr, err := parseInfixExpression(node, source, operators)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}

	if child := firstNamedChild(node); child != nil && child != node {
		expr, err := parseExpression(child, source)
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

func parsePostfixExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty postfix expression")
	}

	result, err := parseExpression(node.NamedChild(0), source)
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
				memberExpr, err = parseExpression(memberNode, source)
				if err != nil {
					return nil, err
				}
			}
			memberAccess := ast.NewMemberAccessExpression(prev, memberExpr)
			annotateCompositeExpression(memberAccess, prev, suffix)
			result = memberAccess
			lastCall = nil
		case "type_arguments":
			typeArgs, err := parseTypeArgumentList(suffix, source)
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
			indexExpr, err := parseExpression(suffix.NamedChild(0), source)
			if err != nil {
				return nil, err
			}
			indexed := ast.NewIndexExpression(prev, indexExpr)
			annotateCompositeExpression(indexed, prev, suffix)
			result = indexed
			lastCall = nil
		case "call_suffix":
			args, err := parseCallArguments(suffix, source)
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
			lambdaExpr, err := parseLambdaExpression(suffix, source)
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
			} else {
				callExpr := ast.NewFunctionCall(prev, nil, typeArgs, true)
				callExpr.Arguments = append(callExpr.Arguments, lambdaExpr)
				annotateCompositeExpression(callExpr, prev, suffix)
				result = callExpr
				lastCall = callExpr
			}
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

func parseCallArguments(node *sitter.Node, source []byte) ([]ast.Expression, error) {
	args := make([]ast.Expression, 0)

	for j := uint(0); j < node.NamedChildCount(); j++ {
		child := node.NamedChild(j)
		if child == nil || !child.IsNamed() || isIgnorableNode(child) {
			continue
		}
		argExpr, err := parseExpression(child, source)
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

func parseBooleanLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, source))
	switch value {
	case "true":
		return annotateExpression(ast.Bool(true), node), nil
	case "false":
		return annotateExpression(ast.Bool(false), node), nil
	default:
		return nil, fmt.Errorf("parser: invalid boolean literal %q", value)
	}
}

func parseNilLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, source))
	if value != "nil" {
		return nil, fmt.Errorf("parser: invalid nil literal %q", value)
	}
	return annotateExpression(ast.Nil(), node), nil
}

func parseArrayLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	elements := make([]ast.Expression, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || !child.IsNamed() || isIgnorableNode(child) {
			continue
		}
		element, err := parseExpression(child, source)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	return annotateExpression(ast.NewArrayLiteral(elements), node), nil
}

func parseStructLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "struct_literal" {
		return nil, fmt.Errorf("parser: expected struct literal node")
	}

	typeNode := node.ChildByFieldName("type")
	if typeNode == nil {
		return nil, fmt.Errorf("parser: struct literal missing type")
	}

	parts, err := parseQualifiedIdentifier(typeNode, source)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("parser: invalid struct literal type")
	}

	structType := collapseQualifiedIdentifier(parts)
	if structType == nil {
		return nil, fmt.Errorf("parser: struct literal missing type identifier")
	}

	typeArgs, err := parseTypeArgumentList(node.ChildByFieldName("type_arguments"), source)
	if err != nil {
		return nil, err
	}

	fields := make([]*ast.StructFieldInitializer, 0)
	var functionalUpdates []ast.Expression

	typeArgsNode := node.ChildByFieldName("type_arguments")

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		fieldName := node.FieldNameForChild(uint32(i))
		if fieldName == "type" || fieldName == "type_arguments" || sameNode(child, typeNode) || sameNode(child, typeArgsNode) {
			continue
		}

		var elem *sitter.Node
		if child.Kind() == "struct_literal_element" {
			elem = firstNamedChild(child)
		} else {
			elem = child
		}
		if elem == nil {
			continue
		}

		switch elem.Kind() {
		case "struct_literal_field":
			nameNode := elem.ChildByFieldName("name")
			if nameNode == nil {
				return nil, fmt.Errorf("parser: struct literal field missing name")
			}
			name, err := parseIdentifier(nameNode, source)
			if err != nil {
				return nil, err
			}
			valueNode := elem.ChildByFieldName("value")
			if valueNode == nil {
				return nil, fmt.Errorf("parser: struct literal field missing value")
			}
			value, err := parseExpression(valueNode, source)
			if err != nil {
				return nil, err
			}
			field := ast.NewStructFieldInitializer(value, name, false)
			annotateSpan(field, elem)
			fields = append(fields, field)
		case "struct_literal_shorthand_field":
			nameNode := elem.ChildByFieldName("name")
			if nameNode == nil {
				nameNode = firstNamedChild(elem)
			}
			name, err := parseIdentifier(nameNode, source)
			if err != nil {
				return nil, err
			}
			field := ast.NewStructFieldInitializer(nil, name, true)
			annotateSpan(field, elem)
			fields = append(fields, field)
		case "struct_literal_spread":
			exprNode := firstNamedChild(elem)
			if exprNode == nil {
				return nil, fmt.Errorf("parser: struct spread missing expression")
			}
			expr, err := parseExpression(exprNode, source)
			if err != nil {
				return nil, err
			}
			functionalUpdates = append(functionalUpdates, expr)
		default:
			expr, err := parseExpression(elem, source)
			if err != nil {
				return nil, err
			}
			field := ast.NewStructFieldInitializer(expr, nil, false)
			annotateSpan(field, elem)
			fields = append(fields, field)
		}
	}

	positional := false
	for _, field := range fields {
		if field.Name == nil {
			positional = true
			break
		}
	}

	return annotateExpression(ast.NewStructLiteral(fields, positional, structType, functionalUpdates, typeArgs), node), nil
}

func parseDoExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: do expression missing body")
	}
	block, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}
	return annotateExpression(block, node), nil
}

func parseProcExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: proc expression missing body")
	}
	body, err := parseExpression(bodyNode, source)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewProcExpression(body), node), nil
}

func parseSpawnExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: spawn expression missing body")
	}
	body, err := parseExpression(bodyNode, source)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewSpawnExpression(body), node), nil
}

func parseBreakpointExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "breakpoint_expression" {
		return nil, fmt.Errorf("parser: expected breakpoint expression node")
	}

	var label *ast.Identifier
	if labelNode := node.ChildByFieldName("label"); labelNode != nil {
		lbl, err := parseLabel(labelNode, source)
		if err != nil {
			return nil, err
		}
		label = lbl
	} else if identNode := fallbackBreakpointLabel(node); identNode != nil {
		lbl, err := parseIdentifier(identNode, source)
		if err != nil {
			return nil, err
		}
		label = lbl
	}

	var bodyNode *sitter.Node
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Kind() == "block" {
			bodyNode = child
			break
		}
	}
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: breakpoint expression missing body")
	}

	body, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}

	return annotateExpression(ast.NewBreakpointExpression(label, body), node), nil
}

func fallbackBreakpointLabel(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}
	childCount := uint(node.ChildCount())
	for i := uint(0); i < childCount; i++ {
		child := node.Child(i)
		if child == nil || isIgnorableNode(child) {
			continue
		}
		if child.Kind() == "identifier" {
			return child
		}
		if child.Kind() == "ERROR" && child.ChildCount() == 1 {
			grand := child.Child(0)
			if grand != nil && grand.Kind() == "identifier" {
				return grand
			}
		}
	}
	return nil
}

func parseHandlingExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "handling_expression" {
		return nil, fmt.Errorf("parser: expected handling_expression node")
	}
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: handling expression missing base expression")
	}

	baseExpr, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}

	current := baseExpr
	var assignment *ast.AssignmentExpression
	if assign, ok := baseExpr.(*ast.AssignmentExpression); ok {
		assignment = assign
		current = assign.Right
	}
	for i := uint(1); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() != "else_clause" {
			continue
		}
		handlerNode := child.ChildByFieldName("handler")
		if handlerNode == nil {
			return nil, fmt.Errorf("parser: else clause missing handler block")
		}
		handler, binding, err := parseHandlingBlock(handlerNode, source)
		if err != nil {
			return nil, err
		}
		prev := current
		orElse := ast.NewOrElseExpression(prev, handler, binding)
		annotateCompositeExpression(orElse, prev, child)
		current = orElse
		if assignment != nil {
			extendExpressionToNode(assignment, child)
		}
	}

	if assignment != nil {
		if current == nil {
			return nil, fmt.Errorf("parser: or-else assignment missing right-hand expression")
		}
		assignment.Right = current
		extendExpressionToNode(assignment, node)
		return assignment, nil
	}

	return current, nil
}

func parseHandlingBlock(node *sitter.Node, source []byte) (*ast.BlockExpression, *ast.Identifier, error) {
	if node == nil || node.Kind() != "handling_block" {
		return nil, nil, fmt.Errorf("parser: expected handling_block node")
	}

	var binding *ast.Identifier
	if bindingNode := node.ChildByFieldName("binding"); bindingNode != nil {
		id, err := parseIdentifier(bindingNode, source)
		if err != nil {
			return nil, nil, err
		}
		binding = id
	}

	statements := make([]ast.Statement, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || !child.IsNamed() {
			continue
		}
		if node.FieldNameForChild(uint32(i)) == "binding" {
			continue
		}
		stmt, err := parseStatement(child, source)
		if err != nil {
			return nil, nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	block := ast.NewBlockExpression(statements)
	annotateExpression(block, node)
	return block, binding, nil
}

func parseRescueExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "rescue_expression" {
		return nil, fmt.Errorf("parser: expected rescue_expression node")
	}
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: rescue expression missing monitored expression")
	}

	var monitoredNode *sitter.Node
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() == "rescue_block" {
			continue
		}
		monitoredNode = child
		break
	}

	if monitoredNode == nil {
		return nil, fmt.Errorf("parser: rescue expression missing monitored expression")
	}

	expr, err := parseExpression(monitoredNode, source)
	if err != nil {
		return nil, err
	}

	rescueNode := node.ChildByFieldName("rescue")
	if rescueNode == nil {
		return nil, fmt.Errorf("parser: rescue expression missing rescue block")
	}

	clauses, err := parseRescueBlock(rescueNode, source)
	if err != nil {
		return nil, err
	}

	if assignment, ok := expr.(*ast.AssignmentExpression); ok {
		if assignment.Right == nil {
			return nil, fmt.Errorf("parser: rescue assignment missing right-hand expression")
		}
		rescueExpr := annotateExpression(ast.NewRescueExpression(assignment.Right, clauses), node)
		assignment.Right = rescueExpr
		extendExpressionToNode(assignment, node)
		return assignment, nil
	}

	return annotateExpression(ast.NewRescueExpression(expr, clauses), node), nil
}

func parseRescueBlock(node *sitter.Node, source []byte) ([]*ast.MatchClause, error) {
	if node == nil || node.Kind() != "rescue_block" {
		return nil, fmt.Errorf("parser: expected rescue_block node")
	}

	var clauses []*ast.MatchClause
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() != "match_clause" {
			continue
		}
		clause, err := parseMatchClause(child, source)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}

	if len(clauses) == 0 {
		return nil, fmt.Errorf("parser: rescue block requires at least one clause")
	}

	return clauses, nil
}

func parseEnsureExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "ensure_expression" {
		return nil, fmt.Errorf("parser: expected ensure_expression node")
	}

	var tryNode *sitter.Node
	ensureNode := node.ChildByFieldName("ensure")
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child == ensureNode {
			continue
		}
		tryNode = child
		break
	}

	if tryNode == nil {
		return nil, fmt.Errorf("parser: ensure expression missing try expression")
	}

	tryExpr, err := parseExpression(tryNode, source)
	if err != nil {
		return nil, err
	}

	if ensureNode == nil {
		return nil, fmt.Errorf("parser: ensure expression missing ensure block")
	}

	ensureBlock, err := parseBlock(ensureNode, source)
	if err != nil {
		return nil, err
	}

	if assignment, ok := tryExpr.(*ast.AssignmentExpression); ok {
		if assignment.Right == nil {
			return nil, fmt.Errorf("parser: ensure assignment missing right-hand expression")
		}
		ensureExpr := annotateExpression(ast.NewEnsureExpression(assignment.Right, ensureBlock), node)
		assignment.Right = ensureExpr
		extendExpressionToNode(assignment, node)
		return assignment, nil
	}

	return annotateExpression(ast.NewEnsureExpression(tryExpr, ensureBlock), node), nil
}

func parseMatchExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil || node.Kind() != "match_expression" {
		return nil, fmt.Errorf("parser: expected match_expression node")
	}

	subjectNode := node.ChildByFieldName("subject")
	if subjectNode == nil {
		return nil, fmt.Errorf("parser: match expression missing subject")
	}

	subject, err := parseExpression(subjectNode, source)
	if err != nil {
		return nil, err
	}

	var clauses []*ast.MatchClause
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() != "match_clause" {
			continue
		}
		clause, err := parseMatchClause(child, source)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}

	if len(clauses) == 0 {
		return nil, fmt.Errorf("parser: match expression requires at least one clause")
	}

	return annotateExpression(ast.NewMatchExpression(subject, clauses), node), nil
}

func parseMatchClause(node *sitter.Node, source []byte) (*ast.MatchClause, error) {
	if node == nil || node.Kind() != "match_clause" {
		return nil, fmt.Errorf("parser: expected match_clause node")
	}

	patternNode := node.ChildByFieldName("pattern")
	if patternNode == nil {
		return nil, fmt.Errorf("parser: match clause missing pattern")
	}
	pattern, err := parsePattern(patternNode, source)
	if err != nil {
		return nil, err
	}

	var guardExpr ast.Expression
	if guardNode := node.ChildByFieldName("guard"); guardNode != nil {
		guardChild := firstNamedChild(guardNode)
		if guardChild == nil {
			return nil, fmt.Errorf("parser: match guard missing expression")
		}
		expr, err := parseExpression(guardChild, source)
		if err != nil {
			return nil, err
		}
		guardExpr = expr
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: match clause missing body")
	}

	var body ast.Expression
	if bodyNode.Kind() == "block" {
		block, err := parseBlock(bodyNode, source)
		if err != nil {
			return nil, err
		}
		body = block
	} else {
		expr, err := parseExpression(bodyNode, source)
		if err != nil {
			return nil, err
		}
		body = expr
	}

	clause := ast.NewMatchClause(pattern, body, guardExpr)
	annotateSpan(clause, node)
	return clause, nil
}

func parseIfExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	conditionNode := node.ChildByFieldName("condition")
	if conditionNode == nil {
		return nil, fmt.Errorf("parser: if expression missing condition")
	}
	condition, err := parseExpression(conditionNode, source)
	if err != nil {
		return nil, err
	}
	bodyNode := node.ChildByFieldName("consequence")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: if expression missing body")
	}
	body, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}
	clauses := make([]*ast.OrClause, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Kind() == "or_clause" {
			clause, err := parseOrClause(child, source)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
		}
	}
	if elseNode := findElseBlock(node, bodyNode); elseNode != nil {
		elseBody, err := parseBlock(elseNode, source)
		if err != nil {
			return nil, err
		}
		elseClause := ast.NewOrClause(elseBody, nil)
		annotateSpan(elseClause, elseNode)
		clauses = append(clauses, elseClause)
	}
	return annotateExpression(ast.NewIfExpression(condition, body, clauses), node), nil
}

func parseOrClause(node *sitter.Node, source []byte) (*ast.OrClause, error) {
	bodyNode := node.ChildByFieldName("consequence")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: or clause missing body")
	}
	body, err := parseBlock(bodyNode, source)
	if err != nil {
		return nil, err
	}
	conditionNode := node.ChildByFieldName("condition")
	var condition ast.Expression
	if conditionNode != nil {
		condExpr, err := parseExpression(conditionNode, source)
		if err != nil {
			return nil, err
		}
		condition = condExpr
	}
	clause := ast.NewOrClause(body, condition)
	annotateSpan(clause, node)
	return clause, nil
}

func findElseBlock(ifNode *sitter.Node, consequence *sitter.Node) *sitter.Node {
	if ifNode == nil {
		return nil
	}
	var consequenceRangeStart, consequenceRangeEnd uint
	if consequence != nil {
		consequenceRangeStart = consequence.StartByte()
		consequenceRangeEnd = consequence.EndByte()
	}
	for i := uint(0); i < ifNode.NamedChildCount(); i++ {
		child := ifNode.NamedChild(i)
		if child.Kind() != "block" {
			continue
		}
		if consequence != nil && child.StartByte() == consequenceRangeStart && child.EndByte() == consequenceRangeEnd {
			continue
		}
		return child
	}
	return nil
}

func parseRangeExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil || node.NamedChildCount() < 2 {
		if child := firstNamedChild(node); child != nil {
			expr, err := parseExpression(child, source)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
		return nil, fmt.Errorf("parser: malformed range expression")
	}
	startExpr, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}
	endExpr, err := parseExpression(node.NamedChild(1), source)
	if err != nil {
		return nil, err
	}
	operatorText := strings.TrimSpace(sliceContent(operatorNode, source))
	inclusive := operatorText == "..."
	if operatorText != ".." && operatorText != "..." {
		return nil, fmt.Errorf("parser: unsupported range operator %q", operatorText)
	}
	return annotateExpression(ast.NewRangeExpression(startExpr, endExpr, inclusive), node), nil
}

func parseAssignmentExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil {
		child := firstNamedChild(node)
		if child == nil {
			return nil, fmt.Errorf("parser: empty assignment expression")
		}
		expr, err := parseExpression(child, source)
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
	left, err := parseAssignmentTarget(leftNode, source)
	if err != nil {
		return nil, err
	}
	right, err := parseExpression(rightNode, source)
	if err != nil {
		return nil, err
	}
	operatorText := strings.TrimSpace(sliceContent(operatorNode, source))
	operator, err := mapAssignmentOperator(operatorText)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewAssignmentExpression(operator, left, right), node), nil
}

func parseAssignmentTarget(node *sitter.Node, source []byte) (ast.AssignmentTarget, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil assignment target")
	}
	switch node.Kind() {
	case "assignment_target":
		if child := firstNamedChild(node); child != nil {
			return parseAssignmentTarget(child, source)
		}
		return nil, fmt.Errorf("parser: empty assignment target")
	case "pattern", "pattern_base", "typed_pattern", "struct_pattern", "array_pattern":
		pattern, err := parsePattern(node, source)
		if err != nil {
			return nil, err
		}
		target, ok := pattern.(ast.AssignmentTarget)
		if !ok {
			return nil, fmt.Errorf("parser: pattern cannot be used as assignment target: %T", pattern)
		}
		return target, nil
	default:
		expr, err := parseExpression(node, source)
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

func parseUnaryExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operandNode := firstNamedChild(node)
	if operandNode == nil {
		return nil, fmt.Errorf("parser: unary expression missing operand")
	}
	if int(node.StartByte()) == int(operandNode.StartByte()) {
		expr, err := parseExpression(operandNode, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	operatorText := strings.TrimSpace(string(source[int(node.StartByte()):int(operandNode.StartByte())]))
	if operatorText == "" {
		expr, err := parseExpression(operandNode, source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	operand, err := parseExpression(operandNode, source)
	if err != nil {
		return nil, err
	}
	switch operatorText {
	case "-":
		return annotateExpression(ast.NewUnaryExpression(ast.UnaryOperatorNegate, operand), node), nil
	case "!":
		return annotateExpression(ast.NewUnaryExpression(ast.UnaryOperatorNot, operand), node), nil
	case "~":
		return annotateExpression(ast.NewUnaryExpression(ast.UnaryOperatorBitNot, operand), node), nil
	default:
		return nil, fmt.Errorf("parser: unsupported unary operator %q", operatorText)
	}
}

func parsePipeExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: empty pipe expression")
	}
	result, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}
	for i := uint(1); i < node.NamedChildCount(); i++ {
		stepNode := node.NamedChild(i)
		stepExpr, err := parseExpression(stepNode, source)
		if err != nil {
			return nil, err
		}
		prev := result
		result = annotateCompositeExpression(ast.NewBinaryExpression("|>", result, stepExpr), prev, stepNode)
	}
	return annotateExpression(result, node), nil
}

func parseInfixExpression(node *sitter.Node, source []byte, operators []string) (ast.Expression, error) {
	count := node.NamedChildCount()
	if count == 0 {
		return nil, fmt.Errorf("parser: empty %s", node.Kind())
	}
	if count == 1 {
		expr, err := parseExpression(node.NamedChild(0), source)
		if err != nil {
			return nil, err
		}
		return annotateExpression(expr, node), nil
	}
	result, err := parseExpression(node.NamedChild(0), source)
	if err != nil {
		return nil, err
	}
	prevNode := node.NamedChild(0)
	for i := uint(1); i < count; i++ {
		rightNode := node.NamedChild(i)
		rightExpr, err := parseExpression(rightNode, source)
		if err != nil {
			return nil, err
		}
		operator := extractOperatorBetween(prevNode, rightNode, source, operators)
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

var assignmentOperatorMap = map[string]ast.AssignmentOperator{
	":=":    ast.AssignmentDeclare,
	"=":     ast.AssignmentAssign,
	"+=":    ast.AssignmentAdd,
	"-=":    ast.AssignmentSub,
	"*=":    ast.AssignmentMul,
	"/=":    ast.AssignmentDiv,
	"%=":    ast.AssignmentMod,
	"&=":    ast.AssignmentBitAnd,
	"|=":    ast.AssignmentBitOr,
	`\xor=`: ast.AssignmentBitXor,
	"<<=":   ast.AssignmentShiftL,
	">>=":   ast.AssignmentShiftR,
}

func mapAssignmentOperator(op string) (ast.AssignmentOperator, error) {
	if operator, ok := assignmentOperatorMap[op]; ok {
		return operator, nil
	}
	return "", fmt.Errorf("parser: unsupported assignment operator %q", op)
}

func parseLambdaExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
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
			param, err := parseLambdaParameter(paramNode, source)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
	}

	var returnType ast.TypeExpression
	if returnNode := node.ChildByFieldName("return_type"); returnNode != nil {
		returnType = parseReturnType(returnNode, source)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: lambda missing body")
	}

	bodyExpr, err := parseExpression(bodyNode, source)
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
