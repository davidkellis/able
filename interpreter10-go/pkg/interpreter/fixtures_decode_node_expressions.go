package interpreter

import (
	"encoding/json"
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func decodeExpressionNodes(node map[string]any, typ string) (ast.Node, bool, error) {
	switch typ {
	case "Identifier":
		name, _ := node["name"].(string)
		return ast.NewIdentifier(name), true, nil
	case "AssignmentExpression":
		op, _ := node["operator"].(string)
		leftNode, err := decodeNode(node["left"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		rightNode, err := decodeNode(node["right"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		left, ok := leftNode.(ast.AssignmentTarget)
		if !ok {
			return nil, true, fmt.Errorf("invalid assignment target %T", leftNode)
		}
		right, ok := rightNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid assignment expression right %T", rightNode)
		}
		return ast.NewAssignmentExpression(ast.AssignmentOperator(op), left, right), true, nil
	case "UnaryExpression":
		op, _ := node["operator"].(string)
		operandNode, ok := node["operand"].(map[string]any)
		if !ok {
			return nil, true, fmt.Errorf("unary expression missing operand")
		}
		decoded, err := decodeNode(operandNode)
		if err != nil {
			return nil, true, err
		}
		expr, ok := decoded.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid unary operand %T", decoded)
		}
		return ast.NewUnaryExpression(ast.UnaryOperator(op), expr), true, nil
	case "BlockExpression":
		bodyVal, _ := node["body"].([]any)
		stmts := make([]ast.Statement, 0, len(bodyVal))
		for _, raw := range bodyVal {
			child, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, true, err
			}
			stmt, ok := child.(ast.Statement)
			if !ok {
				return nil, true, fmt.Errorf("invalid block statement %T", child)
			}
			stmts = append(stmts, stmt)
		}
		return ast.NewBlockExpression(stmts), true, nil
	case "BinaryExpression":
		op, _ := node["operator"].(string)
		leftNode, err := decodeNode(node["left"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		rightNode, err := decodeNode(node["right"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		left, ok := leftNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid binary left operand %T", leftNode)
		}
		right, ok := rightNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid binary right operand %T", rightNode)
		}
		return ast.NewBinaryExpression(op, left, right), true, nil
	case "StringInterpolation":
		partsVal, _ := node["parts"].([]any)
		parts := make([]ast.Expression, 0, len(partsVal))
		for _, raw := range partsVal {
			partNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid interpolation part %T", raw)
			}
			kind, _ := partNode["kind"].(string)
			switch kind {
			case "text":
				val, _ := partNode["value"].(string)
				parts = append(parts, ast.NewStringLiteral(val))
			case "expression":
				exprNode, err := decodeNode(partNode["expression"].(map[string]any))
				if err != nil {
					return nil, true, err
				}
				expr, ok := exprNode.(ast.Expression)
				if !ok {
					return nil, true, fmt.Errorf("invalid interpolation expression %T", exprNode)
				}
				parts = append(parts, expr)
			default:
				return nil, true, fmt.Errorf("unknown interpolation part kind %s", kind)
			}
		}
		return ast.NewStringInterpolation(parts), true, nil
	case "RangeExpression":
		startNode, err := decodeNode(node["start"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		endNode, err := decodeNode(node["end"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		startExpr, ok := startNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid range start %T", startNode)
		}
		endExpr, ok := endNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid range end %T", endNode)
		}
		exclusive, _ := node["exclusive"].(bool)
		return ast.NewRangeExpression(startExpr, endExpr, exclusive), true, nil
	case "MatchExpression":
		scrutNode, err := decodeNode(node["scrutinee"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		scrutinee, ok := scrutNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid match scrutinee %T", scrutNode)
		}
		clausesVal, _ := node["clauses"].([]any)
		clauses := make([]*ast.MatchClause, 0, len(clausesVal))
		for _, raw := range clausesVal {
			clauseNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid match clause %T", raw)
			}
			clause, err := decodeMatchClause(clauseNode)
			if err != nil {
				return nil, true, err
			}
			clauses = append(clauses, clause)
		}
		return ast.NewMatchExpression(scrutinee, clauses), true, nil
	case "PropagationExpression":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		expr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid propagation expression %T", exprNode)
		}
		return ast.NewPropagationExpression(expr), true, nil
	case "OrElseExpression":
		exprNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		valueExpr, ok := exprNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid or-else expression %T", exprNode)
		}
		handlerNode, err := decodeNode(node["handler"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		handler, ok := handlerNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("invalid or-else handler %T", handlerNode)
		}
		var binding *ast.Identifier
		if bindingRaw, ok := node["errorBinding"].(map[string]any); ok {
			decoded, err := decodeNode(bindingRaw)
			if err != nil {
				return nil, true, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, true, fmt.Errorf("invalid or-else error binding %T", decoded)
			}
			binding = id
		}
		return ast.NewOrElseExpression(valueExpr, handler, binding), true, nil
	case "EnsureExpression":
		tryNode, err := decodeNode(node["tryExpression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		tryExpr, ok := tryNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid ensure try expression %T", tryNode)
		}
		ensureNode, err := decodeNode(node["ensureBlock"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		ensureBlock, ok := ensureNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("invalid ensure block %T", ensureNode)
		}
		return ast.NewEnsureExpression(tryExpr, ensureBlock), true, nil
	case "ProcExpression":
		bodyNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		body, ok := bodyNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid proc expression body %T", bodyNode)
		}
		return ast.NewProcExpression(body), true, nil
	case "SpawnExpression":
		bodyNode, err := decodeNode(node["expression"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		bodyExpr, ok := bodyNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid spawn expression %T", bodyNode)
		}
		return ast.NewSpawnExpression(bodyExpr), true, nil
	case "WhileLoop":
		condNode, err := decodeNode(node["condition"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		condExpr, ok := condNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid while condition %T", condNode)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("invalid while body %T", bodyNode)
		}
		return ast.NewWhileLoop(condExpr, body), true, nil
	case "ForLoop":
		iterNode, err := decodeNode(node["iterable"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		iterExpr, ok := iterNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid for iterable %T", iterNode)
		}
		pattern, err := decodePattern(node["pattern"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("invalid for body %T", bodyNode)
		}
		return ast.NewForLoop(pattern, iterExpr, body), true, nil
	case "IfExpression":
		condNode, err := decodeNode(node["ifCondition"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		ifCondition, ok := condNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid if condition %T", condNode)
		}
		bodyNode, err := decodeNode(node["ifBody"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		ifBody, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("invalid if body %T", bodyNode)
		}
		var orClauses []*ast.OrClause
		if clausesRaw, ok := node["orClauses"].([]any); ok {
			orClauses = make([]*ast.OrClause, 0, len(clausesRaw))
			for _, raw := range clausesRaw {
				clauseNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid or clause %T", raw)
				}
				bodyRaw, ok := clauseNode["body"].(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("or clause missing body")
				}
				bodyNode, err := decodeNode(bodyRaw)
				if err != nil {
					return nil, true, err
				}
				body, ok := bodyNode.(*ast.BlockExpression)
				if !ok {
					return nil, true, fmt.Errorf("invalid or clause body %T", bodyNode)
				}
				var condition ast.Expression
				if condRaw, ok := clauseNode["condition"].(map[string]any); ok {
					condNode, err := decodeNode(condRaw)
					if err != nil {
						return nil, true, err
					}
					condExpr, ok := condNode.(ast.Expression)
					if !ok {
						return nil, true, fmt.Errorf("invalid or clause condition %T", condNode)
					}
					condition = condExpr
				}
				orClauses = append(orClauses, ast.NewOrClause(body, condition))
			}
		}
		return ast.NewIfExpression(ifCondition, ifBody, orClauses), true, nil
	case "LambdaExpression":
		paramsVal, _ := node["parameters"].([]any)
		params := make([]*ast.FunctionParameter, 0, len(paramsVal))
		for _, raw := range paramsVal {
			paramNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid lambda parameter %T", raw)
			}
			decoded, err := decodeNode(paramNode)
			if err != nil {
				return nil, true, err
			}
			param, ok := decoded.(*ast.FunctionParameter)
			if !ok {
				return nil, true, fmt.Errorf("invalid lambda parameter %T", decoded)
			}
			params = append(params, param)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		bodyExpr, ok := bodyNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid lambda body %T", bodyNode)
		}
		var returnType ast.TypeExpression
		if retRaw, ok := node["returnType"].(map[string]any); ok {
			ret, err := decodeTypeExpression(retRaw)
			if err != nil {
				return nil, true, err
			}
			returnType = ret
		}
		return ast.NewLambdaExpression(params, bodyExpr, returnType, nil, nil, false), true, nil
	case "FunctionCall":
		calleeNode, err := decodeNode(node["callee"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		callee, ok := calleeNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid callee %T", calleeNode)
		}
		argsVal, _ := node["arguments"].([]any)
		args := make([]ast.Expression, 0, len(argsVal))
		for _, raw := range argsVal {
			argMap, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid argument node %T", raw)
			}
			argNode, err := decodeNode(argMap)
			if err != nil {
				return nil, true, err
			}
			expr, ok := argNode.(ast.Expression)
			if !ok {
				return nil, true, fmt.Errorf("invalid argument %T", argNode)
			}
			args = append(args, expr)
		}
		var typeArgs []ast.TypeExpression
		if taRaw, ok := node["typeArguments"].([]any); ok {
			typeArgs = make([]ast.TypeExpression, 0, len(taRaw))
			for _, raw := range taRaw {
				taNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid type argument %T", raw)
				}
				taExpr, err := decodeTypeExpression(taNode)
				if err != nil {
					return nil, true, err
				}
				typeArgs = append(typeArgs, taExpr)
			}
		}
		isTrailing, _ := node["isTrailingLambda"].(bool)
		return ast.NewFunctionCall(callee, args, typeArgs, isTrailing), true, nil
	case "MemberAccessExpression":
		objectNode, err := decodeNode(node["object"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		object, ok := objectNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid member access object %T", objectNode)
		}
		memberNode, err := decodeNode(node["member"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		memberExpr, ok := memberNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid member access member %T", memberNode)
		}
		return ast.NewMemberAccessExpression(object, memberExpr), true, nil
	case "ImplicitMemberExpression":
		name, _ := node["name"].(string)
		return ast.NewImplicitMemberExpression(ast.NewIdentifier(name)), true, nil
	case "IndexExpression":
		objectNode, err := decodeNode(node["object"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		object, ok := objectNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid index object %T", objectNode)
		}
		indexNode, err := decodeNode(node["index"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		indexExpr, ok := indexNode.(ast.Expression)
		if !ok {
			return nil, true, fmt.Errorf("invalid index expression %T", indexNode)
		}
		return ast.NewIndexExpression(object, indexExpr), true, nil
	case "TopicReferenceExpression":
		return ast.NewTopicReferenceExpression(), true, nil
	case "PlaceholderExpression":
		if idxRaw, ok := node["index"]; ok {
			switch v := idxRaw.(type) {
			case float64:
				index := int(v)
				return ast.NewPlaceholderExpression(&index), true, nil
			case json.Number:
				if i, err := v.Int64(); err == nil {
					index := int(i)
					return ast.NewPlaceholderExpression(&index), true, nil
				}
			}
		}
		return ast.NewPlaceholderExpression(nil), true, nil
	default:
		return nil, false, nil
	}
}
