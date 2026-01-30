package parser

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter-go/pkg/ast"
)

func (ctx *parseContext) parseBlock(node *sitter.Node) (*ast.BlockExpression, error) {
	if node == nil {
		block := ast.NewBlockExpression(nil)
		annotateExpression(block, node)
		return block, nil
	}

	statements := make([]ast.Statement, 0)
	for i := uint(0); i < node.NamedChildCount(); {
		child := node.NamedChild(i)
		i++
		if child == nil || !child.IsNamed() {
			continue
		}
		if node.FieldNameForChild(uint32(i-1)) == "binding" && child.Kind() == "identifier" {
			continue
		}
		if child.Kind() == "elsif_clause_statement" || child.Kind() == "else_clause_statement" {
			if len(statements) == 0 {
				return nil, wrapParseError(child, fmt.Errorf("parser: %s without preceding if expression", child.Kind()))
			}
			target := findIfExpressionTarget(statements[len(statements)-1])
			if target == nil {
				return nil, wrapParseError(child, fmt.Errorf("parser: %s without preceding if expression", child.Kind()))
			}
			switch child.Kind() {
			case "elsif_clause_statement":
				if target.ElseBody != nil {
					return nil, wrapParseError(child, fmt.Errorf("parser: elsif clause after else"))
				}
				clause, err := ctx.parseElseIfClause(child)
				if err != nil {
					return nil, wrapParseError(child, err)
				}
				target.ElseIfClauses = append(target.ElseIfClauses, clause)
				extendExpressionToNode(target, child)
				if elseClause := child.ChildByFieldName("else_clause"); elseClause != nil {
					if target.ElseBody != nil {
						return nil, wrapParseError(elseClause, fmt.Errorf("parser: duplicate else clause"))
					}
					bodyNode := elseClause.ChildByFieldName("alternative")
					if bodyNode == nil {
						bodyNode = firstNamedChild(elseClause)
					}
					if bodyNode == nil {
						return nil, wrapParseError(elseClause, fmt.Errorf("parser: else clause missing body"))
					}
					body, err := ctx.parseBlock(bodyNode)
					if err != nil {
						return nil, wrapParseError(bodyNode, err)
					}
					target.ElseBody = body
					extendExpressionToNode(target, elseClause)
				}
			case "else_clause_statement":
				if target.ElseBody != nil {
					return nil, wrapParseError(child, fmt.Errorf("parser: duplicate else clause"))
				}
				bodyNode := child.ChildByFieldName("alternative")
				if bodyNode == nil {
					return nil, wrapParseError(child, fmt.Errorf("parser: else clause missing body"))
				}
				body, err := ctx.parseBlock(bodyNode)
				if err != nil {
					return nil, wrapParseError(bodyNode, err)
				}
				target.ElseBody = body
				extendExpressionToNode(target, child)
			}
			continue
		}
		var (
			stmt ast.Statement
			err  error
		)
		if child.Kind() == "break_statement" {
			stmt, err = ctx.parseStatement(child)
			if err != nil {
				return nil, wrapParseError(child, err)
			}
			if brk, ok := stmt.(*ast.BreakStatement); ok && brk != nil && brk.Value == nil {
				if next := nextNamedSibling(node, i-1); next != nil && next.Kind() == "expression_statement" {
					exprNode := firstNamedChild(next)
					if exprNode != nil {
						expr, exprErr := ctx.parseExpression(exprNode)
						if exprErr != nil {
							return nil, wrapParseError(exprNode, exprErr)
						}
						brk.Value = expr
						i++
					}
				}
			}
		} else {
			stmt, err = ctx.parseStatement(child)
			if err != nil {
				return nil, wrapParseError(child, err)
			}
		}
		if stmt != nil {
			if child.Kind() == "expression_statement" {
				if assignment, ok := stmt.(*ast.AssignmentExpression); ok && (assignment.Operator == ast.AssignmentAssign || assignment.Operator == ast.AssignmentDeclare) {
					anchorNode := child
					anchorIndex := i - 1
					currentRight := assignment.Right
					for {
						next := nextNamedSibling(node, anchorIndex)
						if next == nil || next.Kind() != "expression_statement" {
							break
						}
						if anchorNode.EndPosition().Row != next.StartPosition().Row {
							break
						}
						if hasSemicolonBetween(ctx.source, anchorNode, next) {
							break
						}
						exprNode := firstNamedChild(next)
						if exprNode == nil {
							break
						}
						expr, exprErr := ctx.parseExpression(exprNode)
						if exprErr != nil {
							return nil, wrapParseError(exprNode, exprErr)
						}
						unary, ok := expr.(*ast.UnaryExpression)
						if !ok || unary.Operator != ast.UnaryOperatorNegate {
							break
						}
						newRight := ast.NewBinaryExpression("-", currentRight, unary.Operand)
						assignment.Right = annotateCompositeExpression(newRight, currentRight, exprNode)
						currentRight = assignment.Right
						i++
						nextIndex := findNamedChildIndex(node, next)
						if nextIndex < 0 {
							break
						}
						anchorIndex = uint(nextIndex)
						anchorNode = next
					}
				}
			}
			if lambda, ok := stmt.(*ast.LambdaExpression); ok && len(statements) > 0 {
				switch prev := statements[len(statements)-1].(type) {
				case *ast.AssignmentExpression:
					switch rhs := prev.Right.(type) {
					case *ast.FunctionCall:
						if len(rhs.Arguments) == 0 || rhs.Arguments[len(rhs.Arguments)-1] != lambda {
							rhs.Arguments = append(rhs.Arguments, lambda)
						}
						rhs.IsTrailingLambda = true
						continue
					case ast.Expression:
						call := ast.NewFunctionCall(rhs, nil, nil, true)
						call.Arguments = []ast.Expression{lambda}
						prev.Right = call
						continue
					}
				case *ast.FunctionCall:
					if len(prev.Arguments) == 0 || prev.Arguments[len(prev.Arguments)-1] != lambda {
						prev.Arguments = append(prev.Arguments, lambda)
					}
					prev.IsTrailingLambda = true
					continue
				case ast.Expression:
					call := ast.NewFunctionCall(prev, nil, nil, true)
					call.Arguments = []ast.Expression{lambda}
					statements[len(statements)-1] = call
					continue
				}
			}
			statements = append(statements, stmt)
		}
	}

	block := ast.NewBlockExpression(statements)
	annotateExpression(block, node)
	return block, nil
}

func findIfExpressionTarget(stmt ast.Statement) *ast.IfExpression {
	switch s := stmt.(type) {
	case *ast.IfExpression:
		return s
	case *ast.AssignmentExpression:
		if expr, ok := s.Right.(*ast.IfExpression); ok {
			return expr
		}
	case *ast.ReturnStatement:
		if expr, ok := s.Argument.(*ast.IfExpression); ok {
			return expr
		}
	case *ast.BreakStatement:
		if expr, ok := s.Value.(*ast.IfExpression); ok {
			return expr
		}
	}
	return nil
}

func (ctx *parseContext) parseStatement(node *sitter.Node) (ast.Statement, error) {
	switch node.Kind() {
	case "expression_statement":
		exprNode := firstNamedChild(node)
		if exprNode == nil {
			return nil, fmt.Errorf("parser: expression statement missing expression")
		}
		expr, err := ctx.parseExpression(exprNode)
		if err != nil {
			return nil, wrapParseError(exprNode, err)
		}
		return annotateStatement(expr, node), nil
	case "return_statement":
		valueNode := firstNamedChild(node)
		if valueNode == nil {
			return annotateStatement(ast.NewReturnStatement(nil), node), nil
		}
		expr, err := ctx.parseExpression(valueNode)
		if err != nil {
			return nil, wrapParseError(valueNode, err)
		}
		return annotateStatement(ast.NewReturnStatement(expr), node), nil
	case "while_statement":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed while statement")
		}
		conditionNode := node.NamedChild(0)
		bodyNode := node.NamedChild(1)
		condition, err := ctx.parseExpression(conditionNode)
		if err != nil {
			return nil, wrapParseError(conditionNode, err)
		}
		body, err := ctx.parseBlock(bodyNode)
		if err != nil {
			return nil, wrapParseError(bodyNode, err)
		}
		return annotateStatement(ast.NewWhileLoop(condition, body), node), nil
	case "loop_statement":
		bodyNode := firstNamedChild(node)
		if bodyNode == nil {
			return nil, fmt.Errorf("parser: loop statement missing body")
		}
		body, err := ctx.parseBlock(bodyNode)
		if err != nil {
			return nil, wrapParseError(bodyNode, err)
		}
		return annotateStatement(ast.NewLoopExpression(body), node), nil
	case "for_statement":
		if node.NamedChildCount() < 3 {
			return nil, fmt.Errorf("parser: malformed for statement")
		}
		patternNode := node.NamedChild(0)
		iterNode := node.NamedChild(1)
		bodyNode := node.NamedChild(2)
		pattern, err := ctx.parsePattern(patternNode)
		if err != nil {
			return nil, wrapParseError(patternNode, err)
		}
		iterable, err := ctx.parseExpression(iterNode)
		if err != nil {
			return nil, wrapParseError(iterNode, err)
		}
		body, err := ctx.parseBlock(bodyNode)
		if err != nil {
			return nil, wrapParseError(bodyNode, err)
		}
		return annotateStatement(ast.NewForLoop(pattern, iterable, body), node), nil
	case "break_statement":
		labelNode := node.ChildByFieldName("label")
		var label *ast.Identifier
		if labelNode != nil {
			lbl, err := parseLabel(labelNode, ctx.source)
			if err != nil {
				return nil, err
			}
			label = lbl
		}
		valueNode := node.ChildByFieldName("value")
		var value ast.Expression
		if valueNode != nil {
			expr, err := ctx.parseExpression(valueNode)
			if err != nil {
				return nil, wrapParseError(valueNode, err)
			}
			value = expr
		}
		return annotateStatement(ast.NewBreakStatement(label, value), node), nil
	case "continue_statement":
		return annotateStatement(ast.NewContinueStatement(nil), node), nil
	case "raise_statement":
		valueNode := firstNamedChild(node)
		if valueNode == nil {
			return nil, fmt.Errorf("parser: raise statement missing expression")
		}
		expr, err := ctx.parseExpression(valueNode)
		if err != nil {
			return nil, wrapParseError(valueNode, err)
		}
		return annotateStatement(ast.NewRaiseStatement(expr), node), nil
	case "rethrow_statement":
		return annotateStatement(ast.NewRethrowStatement(), node), nil
	case "struct_definition":
		stmt, err := ctx.parseStructDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "methods_definition":
		stmt, err := ctx.parseMethodsDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "implementation_definition":
		stmt, err := ctx.parseImplementationDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "named_implementation_definition":
		stmt, err := ctx.parseNamedImplementationDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "union_definition":
		stmt, err := ctx.parseUnionDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "interface_definition":
		stmt, err := ctx.parseInterfaceDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "type_alias_definition":
		stmt, err := ctx.parseTypeAliasDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "function_definition":
		stmt, err := ctx.parseFunctionDefinition(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "prelude_statement":
		stmt, err := ctx.parsePreludeStatement(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	case "extern_function":
		stmt, err := ctx.parseExternFunction(node)
		if err != nil {
			return nil, err
		}
		return annotateStatement(stmt, node), nil
	default:
		// For now, ignore unsupported statements in blocks.
		return nil, nil
	}
}
