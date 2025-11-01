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
	return ast.NewImplicitMemberExpression(member), nil
}

func parsePlaceholderExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	raw := strings.TrimSpace(sliceContent(node, source))
	if raw == "" {
		return nil, fmt.Errorf("parser: empty placeholder expression")
	}
	if raw == "@" {
		return ast.NewPlaceholderExpression(nil), nil
	}
	if strings.HasPrefix(raw, "@") {
		value := raw[1:]
		if value == "" {
			return ast.NewPlaceholderExpression(nil), nil
		}
		index, err := strconv.Atoi(value)
		if err != nil || index <= 0 {
			return nil, fmt.Errorf("parser: invalid placeholder index %q", raw)
		}
		return ast.NewPlaceholderExpression(&index), nil
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
				parts = append(parts, ast.Str(text))
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
	return ast.NewStringInterpolation(parts), nil
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

	return ast.NewIteratorLiteral(block.Body), nil
}

func parseExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	if node == nil {
		return nil, fmt.Errorf("parser: nil expression node")
	}

	switch node.Kind() {
	case "identifier":
		return parseIdentifier(node, source)
	case "number_literal":
		return parseNumberLiteral(node, source)
	case "boolean_literal":
		return parseBooleanLiteral(node, source)
	case "nil_literal":
		return parseNilLiteral(node, source)
	case "string_literal":
		return parseStringLiteral(node, source)
	case "character_literal":
		return parseCharLiteral(node, source)
	case "array_literal":
		return parseArrayLiteral(node, source)
	case "struct_literal":
		return parseStructLiteral(node, source)
	case "block":
		return parseBlock(node, source)
	case "do_expression":
		return parseDoExpression(node, source)
	case "lambda_expression":
		return parseLambdaExpression(node, source)
	case "postfix_expression":
		return parsePostfixExpression(node, source)
	case "call_target":
		return parsePostfixExpression(node, source)
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
		return ast.NewMemberAccessExpression(objectExpr, memberExpr), nil
	case "proc_expression":
		return parseProcExpression(node, source)
	case "spawn_expression":
		return parseSpawnExpression(node, source)
	case "breakpoint_expression":
		return parseBreakpointExpression(node, source)
	case "handling_expression":
		return parseHandlingExpression(node, source)
	case "rescue_expression":
		return parseRescueExpression(node, source)
	case "ensure_expression":
		return parseEnsureExpression(node, source)
	case "if_expression":
		return parseIfExpression(node, source)
	case "match_expression":
		return parseMatchExpression(node, source)
	case "range_expression":
		return parseRangeExpression(node, source)
	case "assignment_expression":
		return parseAssignmentExpression(node, source)
	case "unary_expression":
		return parseUnaryExpression(node, source)
	case "implicit_member_expression":
		return parseImplicitMemberExpression(node, source)
	case "placeholder_expression":
		return parsePlaceholderExpression(node, source)
	case "topic_reference":
		return ast.NewTopicReferenceExpression(), nil
	case "interpolated_string":
		return parseInterpolatedString(node, source)
	case "iterator_literal":
		return parseIteratorLiteral(node, source)
	case "parenthesized_expression":
		if child := firstNamedChild(node); child != nil {
			return parseExpression(child, source)
		}
		return nil, fmt.Errorf("parser: empty parenthesized expression")
	case "pipe_expression":
		return parsePipeExpression(node, source)
	case "matchable_expression":
		if child := firstNamedChild(node); child != nil {
			return parseExpression(child, source)
		}
	}

	if operators, ok := infixOperatorSets[node.Kind()]; ok {
		return parseInfixExpression(node, source, operators)
	}

	if child := firstNamedChild(node); child != nil && child != node {
		return parseExpression(child, source)
	}

	if id, ok := findIdentifier(node, source); ok {
		return id, nil
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
				memberExpr = ast.Int(int64(intValue))
			} else {
				memberExpr, err = parseExpression(memberNode, source)
				if err != nil {
					return nil, err
				}
			}
			result = ast.NewMemberAccessExpression(result, memberExpr)
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
			result = ast.NewIndexExpression(result, indexExpr)
			lastCall = nil
		case "call_suffix":
			args, err := parseCallArguments(suffix, source)
			if err != nil {
				return nil, err
			}
			typeArgs := pendingTypeArgs
			pendingTypeArgs = nil

			callExpr := ast.NewFunctionCall(result, args, typeArgs, false)
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
				result = lastCall
			} else {
				callExpr := ast.NewFunctionCall(result, nil, typeArgs, true)
				callExpr.Arguments = append(callExpr.Arguments, lambdaExpr)
				result = callExpr
				lastCall = callExpr
			}
		case "propagate_suffix":
			if len(pendingTypeArgs) > 0 {
				return nil, fmt.Errorf("parser: dangling type arguments before propagation")
			}
			result = ast.NewPropagationExpression(result)
			lastCall = nil
		default:
			return nil, fmt.Errorf("parser: unsupported postfix suffix %q", suffix.Kind())
		}
	}

	if len(pendingTypeArgs) > 0 {
		return nil, fmt.Errorf("parser: dangling type arguments in expression")
	}

	return result, nil
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
		return ast.Bool(true), nil
	case "false":
		return ast.Bool(false), nil
	default:
		return nil, fmt.Errorf("parser: invalid boolean literal %q", value)
	}
}

func parseNilLiteral(node *sitter.Node, source []byte) (ast.Expression, error) {
	value := strings.TrimSpace(sliceContent(node, source))
	if value != "nil" {
		return nil, fmt.Errorf("parser: invalid nil literal %q", value)
	}
	return ast.Nil(), nil
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
	return ast.NewArrayLiteral(elements), nil
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

	structType := parts[len(parts)-1]
	if len(parts) > 1 {
		structType = ast.ID(strings.Join(identifiersToStrings(parts), "."))
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
			fields = append(fields, ast.NewStructFieldInitializer(value, name, false))
		case "struct_literal_shorthand_field":
			nameNode := elem.ChildByFieldName("name")
			if nameNode == nil {
				nameNode = firstNamedChild(elem)
			}
			name, err := parseIdentifier(nameNode, source)
			if err != nil {
				return nil, err
			}
			fields = append(fields, ast.NewStructFieldInitializer(nil, name, true))
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
			fields = append(fields, ast.NewStructFieldInitializer(expr, nil, false))
		}
	}

	positional := false
	for _, field := range fields {
		if field.Name == nil {
			positional = true
			break
		}
	}

	return ast.NewStructLiteral(fields, positional, structType, functionalUpdates, typeArgs), nil
}

func parseDoExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: do expression missing body")
	}
	return parseBlock(bodyNode, source)
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
	return ast.NewProcExpression(body), nil
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
	return ast.NewSpawnExpression(body), nil
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

	return ast.NewBreakpointExpression(label, body), nil
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
		current = ast.NewOrElseExpression(current, handler, binding)
	}

	if assignment != nil {
		if current == nil {
			return nil, fmt.Errorf("parser: or-else assignment missing right-hand expression")
		}
		assignment.Right = current
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

	return ast.NewBlockExpression(statements), binding, nil
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
		rescueExpr := ast.NewRescueExpression(assignment.Right, clauses)
		assignment.Right = rescueExpr
		return assignment, nil
	}

	return ast.NewRescueExpression(expr, clauses), nil
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
		ensureExpr := ast.NewEnsureExpression(assignment.Right, ensureBlock)
		assignment.Right = ensureExpr
		return assignment, nil
	}

	return ast.NewEnsureExpression(tryExpr, ensureBlock), nil
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

	return ast.NewMatchExpression(subject, clauses), nil
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

	return ast.NewMatchClause(pattern, body, guardExpr), nil
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
		clauses = append(clauses, ast.NewOrClause(elseBody, nil))
	}
	return ast.NewIfExpression(condition, body, clauses), nil
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
	return ast.NewOrClause(body, condition), nil
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
			return parseExpression(child, source)
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
	return ast.NewRangeExpression(startExpr, endExpr, inclusive), nil
}

func parseAssignmentExpression(node *sitter.Node, source []byte) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil {
		child := firstNamedChild(node)
		if child == nil {
			return nil, fmt.Errorf("parser: empty assignment expression")
		}
		return parseExpression(child, source)
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
	return ast.NewAssignmentExpression(operator, left, right), nil
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
		return parseExpression(operandNode, source)
	}
	operatorText := strings.TrimSpace(string(source[int(node.StartByte()):int(operandNode.StartByte())]))
	if operatorText == "" {
		return parseExpression(operandNode, source)
	}
	operand, err := parseExpression(operandNode, source)
	if err != nil {
		return nil, err
	}
	switch operatorText {
	case "-":
		return ast.NewUnaryExpression(ast.UnaryOperatorNegate, operand), nil
	case "!":
		return ast.NewUnaryExpression(ast.UnaryOperatorNot, operand), nil
	case "~":
		return ast.NewUnaryExpression(ast.UnaryOperatorBitNot, operand), nil
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
		result = ast.NewBinaryExpression("|>", result, stepExpr)
	}
	return result, nil
}

func parseInfixExpression(node *sitter.Node, source []byte, operators []string) (ast.Expression, error) {
	count := node.NamedChildCount()
	if count == 0 {
		return nil, fmt.Errorf("parser: empty %s", node.Kind())
	}
	if count == 1 {
		return parseExpression(node.NamedChild(0), source)
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
		result = ast.NewBinaryExpression(operator, result, rightExpr)
		prevNode = rightNode
	}
	return result, nil
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

	return ast.NewLambdaExpression(params, bodyExpr, returnType, nil, nil, false), nil
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

	return ast.NewFunctionParameter(id, nil), nil
}
