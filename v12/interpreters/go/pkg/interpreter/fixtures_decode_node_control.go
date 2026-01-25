package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func decodeControlFlowNodes(node map[string]any, typ string) (ast.Node, bool, error) {
	switch typ {
	case "ContinueStatement":
		var label *ast.Identifier
		if labelRaw, ok := node["label"].(map[string]any); ok {
			decoded, err := decodeNode(labelRaw)
			if err != nil {
				return nil, true, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, true, fmt.Errorf("invalid continue label %T", decoded)
			}
			label = id
		}
		return ast.NewContinueStatement(label), true, nil
	case "ReturnStatement":
		var argument ast.Expression
		if argRaw, ok := node["argument"].(map[string]any); ok {
			decoded, err := decodeNode(argRaw)
			if err != nil {
				return nil, true, err
			}
			expr, ok := decoded.(ast.Expression)
			if !ok {
				return nil, true, fmt.Errorf("invalid return argument %T", decoded)
			}
			argument = expr
		}
		return ast.NewReturnStatement(argument), true, nil
	case "RaiseStatement":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		expr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid raise expression %T", exprNode)
		}
		return ast.NewRaiseStatement(expr), true, nil
	case "BreakStatement":
		var label *ast.Identifier
		if labelRaw, ok := node["label"].(map[string]any); ok {
			decoded, err := decodeNode(labelRaw)
			if err != nil {
				return nil, true, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, true, fmt.Errorf("invalid break label %T", decoded)
			}
			label = id
		}
		var value ast.Expression
		if valueRaw, ok := node["value"].(map[string]any); ok {
			decoded, err := decodeNode(valueRaw)
			if err != nil {
				return nil, true, err
			}
			expr, ok := decoded.(ast.Expression)
			if !ok {
				return nil, true, fmt.Errorf("invalid break value %T", decoded)
			}
			value = expr
		}
		return ast.NewBreakStatement(label, value), true, nil
	case "RescueExpression":
		monNode, err := decodeNode(node["monitoredExpression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		monExpr, ok := monNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid rescue expression %T", monNode)
		}
		clausesVal, _ := node["clauses"].([]any)
		clauses := make([]*ast.MatchClause, 0, len(clausesVal))
		for _, raw := range clausesVal {
			clauseNode, err := decodeMatchClause(raw.(map[string]any))
			if err != nil {
				return nil, true, err
			}
			clauses = append(clauses, clauseNode)
		}
		return ast.NewRescueExpression(monExpr, clauses), true, nil
	case "BreakpointExpression":
		var label *ast.Identifier
		if labelRaw, ok := node["label"].(map[string]any); ok {
			decoded, err := decodeNode(labelRaw)
			if err != nil {
				return nil, true, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, true, fmt.Errorf("invalid breakpoint label %T", decoded)
			}
			label = id
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("breakpoint body must be block expression, got %T", bodyNode)
		}
		return ast.NewBreakpointExpression(label, body), true, nil
	case "YieldStatement":
		var expr ast.Expression
		if exprRaw, ok := node["expression"].(map[string]any); ok {
			decoded, err := decodeNode(exprRaw)
			if err != nil {
				return nil, true, err
			}
			exprVal, ok := decoded.(ast.Expression)
			if !ok {
				return nil, true, fmt.Errorf("invalid yield expression %T", decoded)
			}
			expr = exprVal
		}
		return ast.NewYieldStatement(expr), true, nil
	case "RethrowStatement":
		return ast.NewRethrowStatement(), true, nil
	default:
		return nil, false, nil
	}
}
