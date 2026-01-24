package parser

import (
	"able/interpreter-go/pkg/ast"
	"fmt"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"strings"
)

func (ctx *parseContext) parseDoExpression(node *sitter.Node) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: do expression missing body")
	}
	block, err := ctx.parseBlock(bodyNode)
	if err != nil {
		return nil, err
	}
	return annotateExpression(block, node), nil
}

func (ctx *parseContext) parseLoopExpression(node *sitter.Node) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: loop expression missing body")
	}
	body, err := ctx.parseBlock(bodyNode)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewLoopExpression(body), node), nil
}

func (ctx *parseContext) parseSpawnExpression(node *sitter.Node) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: spawn expression missing body")
	}
	body, err := ctx.parseExpression(bodyNode)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewSpawnExpression(body), node), nil
}

func (ctx *parseContext) parseAwaitExpression(node *sitter.Node) (ast.Expression, error) {
	bodyNode := firstNamedChild(node)
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: await expression missing operand")
	}
	body, err := ctx.parseExpression(bodyNode)
	if err != nil {
		return nil, err
	}
	return annotateExpression(ast.NewAwaitExpression(body), node), nil
}

func (ctx *parseContext) parseBreakpointExpression(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "breakpoint_expression" {
		return nil, fmt.Errorf("parser: expected breakpoint expression node")
	}

	var label *ast.Identifier
	if labelNode := node.ChildByFieldName("label"); labelNode != nil {
		lbl, err := parseLabel(labelNode, ctx.source)
		if err != nil {
			return nil, err
		}
		label = lbl
	} else if identNode := fallbackBreakpointLabel(node); identNode != nil {
		lbl, err := parseIdentifier(identNode, ctx.source)
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

	body, err := ctx.parseBlock(bodyNode)
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

func (ctx *parseContext) parseHandlingExpression(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "handling_expression" {
		return nil, fmt.Errorf("parser: expected handling_expression node")
	}
	if node.NamedChildCount() == 0 {
		return nil, fmt.Errorf("parser: handling expression missing base expression")
	}

	baseExpr, err := ctx.parseExpression(node.NamedChild(0))
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
		if child == nil || child.Kind() != "or_handler_clause" {
			continue
		}
		handlerNode := child.ChildByFieldName("handler")
		if handlerNode == nil {
			return nil, fmt.Errorf("parser: or clause missing handler block")
		}
		handler, binding, err := ctx.parseHandlingBlock(handlerNode)
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

func (ctx *parseContext) parseHandlingBlock(node *sitter.Node) (*ast.BlockExpression, *ast.Identifier, error) {
	if node == nil || node.Kind() != "handling_block" {
		return nil, nil, fmt.Errorf("parser: expected handling_block node")
	}

	var binding *ast.Identifier
	if bindingNode := node.ChildByFieldName("binding"); bindingNode != nil {
		id, err := parseIdentifier(bindingNode, ctx.source)
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
		if node.FieldNameForChild(uint32(i)) == "binding" && child.Kind() == "identifier" {
			continue
		}
		stmt, err := ctx.parseStatement(child)
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

func (ctx *parseContext) parseRescueExpression(node *sitter.Node) (ast.Expression, error) {
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

	expr, err := ctx.parseExpression(monitoredNode)
	if err != nil {
		return nil, err
	}

	rescueNode := node.ChildByFieldName("rescue")
	if rescueNode == nil {
		return nil, fmt.Errorf("parser: rescue expression missing rescue block")
	}

	clauses, err := ctx.parseRescueBlock(rescueNode)
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

func (ctx *parseContext) parseRescueBlock(node *sitter.Node) ([]*ast.MatchClause, error) {
	if node == nil || node.Kind() != "rescue_block" {
		return nil, fmt.Errorf("parser: expected rescue_block node")
	}

	var clauses []*ast.MatchClause
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() != "match_clause" {
			continue
		}
		clause, err := ctx.parseMatchClause(child)
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

func (ctx *parseContext) parseEnsureExpression(node *sitter.Node) (ast.Expression, error) {
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

	tryExpr, err := ctx.parseExpression(tryNode)
	if err != nil {
		return nil, err
	}

	if ensureNode == nil {
		return nil, fmt.Errorf("parser: ensure expression missing ensure block")
	}

	ensureBlock, err := ctx.parseBlock(ensureNode)
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

func (ctx *parseContext) parseMatchExpression(node *sitter.Node) (ast.Expression, error) {
	if node == nil || node.Kind() != "match_expression" {
		return nil, fmt.Errorf("parser: expected match_expression node")
	}

	subjectNode := node.ChildByFieldName("subject")
	if subjectNode == nil {
		return nil, fmt.Errorf("parser: match expression missing subject")
	}

	subject, err := ctx.parseExpression(subjectNode)
	if err != nil {
		return nil, err
	}

	var clauses []*ast.MatchClause
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child == nil || child.Kind() != "match_clause" {
			continue
		}
		clause, err := ctx.parseMatchClause(child)
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

func (ctx *parseContext) parseMatchClause(node *sitter.Node) (*ast.MatchClause, error) {
	if node == nil || node.Kind() != "match_clause" {
		return nil, fmt.Errorf("parser: expected match_clause node")
	}

	patternNode := node.ChildByFieldName("pattern")
	if patternNode == nil {
		return nil, fmt.Errorf("parser: match clause missing pattern")
	}
	pattern, err := ctx.parsePattern(patternNode)
	if err != nil {
		return nil, err
	}

	var guardExpr ast.Expression
	if guardNode := node.ChildByFieldName("guard"); guardNode != nil {
		guardChild := firstNamedChild(guardNode)
		if guardChild == nil {
			return nil, fmt.Errorf("parser: match guard missing expression")
		}
		expr, err := ctx.parseExpression(guardChild)
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
		block, err := ctx.parseBlock(bodyNode)
		if err != nil {
			return nil, err
		}
		body = block
	} else {
		expr, err := ctx.parseExpression(bodyNode)
		if err != nil {
			return nil, err
		}
		body = expr
	}

	clause := ast.NewMatchClause(pattern, body, guardExpr)
	annotateSpan(clause, node)
	return clause, nil
}

func (ctx *parseContext) parseIfExpression(node *sitter.Node) (ast.Expression, error) {
	conditionNode := node.ChildByFieldName("condition")
	if conditionNode == nil {
		return nil, fmt.Errorf("parser: if expression missing condition")
	}
	condition, err := ctx.parseExpression(conditionNode)
	if err != nil {
		return nil, err
	}
	bodyNode := node.ChildByFieldName("consequence")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: if expression missing body")
	}
	body, err := ctx.parseBlock(bodyNode)
	if err != nil {
		return nil, err
	}
	clauses := make([]*ast.ElseIfClause, 0)
	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Kind() == "elsif_clause" {
			clause, err := ctx.parseElseIfClause(child)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
		}
	}
	var elseBody *ast.BlockExpression
	if elseNode := node.ChildByFieldName("alternative"); elseNode != nil {
		block, err := ctx.parseBlock(elseNode)
		if err != nil {
			return nil, err
		}
		elseBody = block
	}
	return annotateExpression(ast.NewIfExpression(condition, body, clauses, elseBody), node), nil
}

func (ctx *parseContext) parseElseIfClause(node *sitter.Node) (*ast.ElseIfClause, error) {
	bodyNode := node.ChildByFieldName("consequence")
	if bodyNode == nil {
		return nil, fmt.Errorf("parser: elsif clause missing body")
	}
	body, err := ctx.parseBlock(bodyNode)
	if err != nil {
		return nil, err
	}
	conditionNode := node.ChildByFieldName("condition")
	if conditionNode == nil {
		return nil, fmt.Errorf("parser: elsif clause missing condition")
	}
	condExpr, err := ctx.parseExpression(conditionNode)
	if err != nil {
		return nil, err
	}
	clause := ast.NewElseIfClause(body, condExpr)
	annotateSpan(clause, node)
	return clause, nil
}

func (ctx *parseContext) parseRangeExpression(node *sitter.Node) (ast.Expression, error) {
	operatorNode := node.ChildByFieldName("operator")
	if operatorNode == nil || node.NamedChildCount() < 2 {
		if child := firstNamedChild(node); child != nil {
			expr, err := ctx.parseExpression(child)
			if err != nil {
				return nil, err
			}
			return annotateExpression(expr, node), nil
		}
		return nil, fmt.Errorf("parser: malformed range expression")
	}
	startExpr, err := ctx.parseExpression(node.NamedChild(0))
	if err != nil {
		return nil, err
	}
	endExpr, err := ctx.parseExpression(node.NamedChild(1))
	if err != nil {
		return nil, err
	}
	operatorText := strings.TrimSpace(sliceContent(operatorNode, ctx.source))
	inclusive := operatorText == ".."
	if operatorText != ".." && operatorText != "..." {
		return nil, fmt.Errorf("parser: unsupported range operator %q", operatorText)
	}
	return annotateExpression(ast.NewRangeExpression(startExpr, endExpr, inclusive), node), nil
}
