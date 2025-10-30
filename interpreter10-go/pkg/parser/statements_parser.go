package parser

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"able/interpreter10-go/pkg/ast"
)

func parseBlock(node *sitter.Node, source []byte) (*ast.BlockExpression, error) {
	if node == nil {
		return ast.NewBlockExpression(nil), nil
	}

	statements := make([]ast.Statement, 0)
	for i := uint(0); i < node.NamedChildCount(); {
		child := node.NamedChild(i)
		i++
		if child == nil || !child.IsNamed() {
			continue
		}
		if node.FieldNameForChild(uint32(i-1)) == "binding" {
			continue
		}
		var (
			stmt ast.Statement
			err  error
		)
		if child.Kind() == "break_statement" {
			stmt, err = parseStatement(child, source)
			if err != nil {
				return nil, err
			}
			if brk, ok := stmt.(*ast.BreakStatement); ok && brk != nil && brk.Value == nil {
				if next := nextNamedSibling(node, i-1); next != nil && next.Kind() == "expression_statement" {
					exprNode := firstNamedChild(next)
					if exprNode != nil {
						expr, exprErr := parseExpression(exprNode, source)
						if exprErr != nil {
							return nil, exprErr
						}
						brk.Value = expr
						i++
					}
				}
			}
		} else {
			stmt, err = parseStatement(child, source)
			if err != nil {
				return nil, err
			}
		}
		if stmt != nil {
			if lambda, ok := stmt.(*ast.LambdaExpression); ok && len(statements) > 0 {
				switch prev := statements[len(statements)-1].(type) {
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

	return ast.NewBlockExpression(statements), nil
}

func parseStatement(node *sitter.Node, source []byte) (ast.Statement, error) {
	switch node.Kind() {
	case "expression_statement":
		exprNode := firstNamedChild(node)
		if exprNode == nil {
			return nil, fmt.Errorf("parser: expression statement missing expression")
		}
		expr, err := parseExpression(exprNode, source)
		if err != nil {
			return nil, err
		}
		return expr, nil
	case "return_statement":
		valueNode := firstNamedChild(node)
		if valueNode == nil {
			return ast.NewReturnStatement(nil), nil
		}
		expr, err := parseExpression(valueNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewReturnStatement(expr), nil
	case "while_statement":
		if node.NamedChildCount() < 2 {
			return nil, fmt.Errorf("parser: malformed while statement")
		}
		conditionNode := node.NamedChild(0)
		bodyNode := node.NamedChild(1)
		condition, err := parseExpression(conditionNode, source)
		if err != nil {
			return nil, err
		}
		body, err := parseBlock(bodyNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewWhileLoop(condition, body), nil
	case "for_statement":
		if node.NamedChildCount() < 3 {
			return nil, fmt.Errorf("parser: malformed for statement")
		}
		patternNode := node.NamedChild(0)
		iterNode := node.NamedChild(1)
		bodyNode := node.NamedChild(2)
		pattern, err := parsePattern(patternNode, source)
		if err != nil {
			return nil, err
		}
		iterable, err := parseExpression(iterNode, source)
		if err != nil {
			return nil, err
		}
		body, err := parseBlock(bodyNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewForLoop(pattern, iterable, body), nil
	case "break_statement":
		labelNode := node.ChildByFieldName("label")
		var label *ast.Identifier
		if labelNode != nil {
			lbl, err := parseLabel(labelNode, source)
			if err != nil {
				return nil, err
			}
			label = lbl
		}
		valueNode := node.ChildByFieldName("value")
		var value ast.Expression
		if valueNode != nil {
			expr, err := parseExpression(valueNode, source)
			if err != nil {
				return nil, err
			}
			value = expr
		}
		return ast.NewBreakStatement(label, value), nil
	case "continue_statement":
		return ast.NewContinueStatement(nil), nil
	case "raise_statement":
		valueNode := firstNamedChild(node)
		if valueNode == nil {
			return nil, fmt.Errorf("parser: raise statement missing expression")
		}
		expr, err := parseExpression(valueNode, source)
		if err != nil {
			return nil, err
		}
		return ast.NewRaiseStatement(expr), nil
	case "struct_definition":
		return parseStructDefinition(node, source)
	case "methods_definition":
		return parseMethodsDefinition(node, source)
	case "implementation_definition":
		return parseImplementationDefinition(node, source)
	case "named_implementation_definition":
		return parseNamedImplementationDefinition(node, source)
	case "union_definition":
		return parseUnionDefinition(node, source)
	case "interface_definition":
		return parseInterfaceDefinition(node, source)
	case "prelude_statement":
		return parsePreludeStatement(node, source)
	case "extern_function":
		return parseExternFunction(node, source)
	default:
		// For now, ignore unsupported statements in blocks.
		return nil, nil
	}
}
