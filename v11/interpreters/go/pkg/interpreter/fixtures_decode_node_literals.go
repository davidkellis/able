package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func decodeLiteralNodes(node map[string]any, typ string) (ast.Node, bool, error) {
	switch typ {
	case "StringLiteral":
		val, _ := node["value"].(string)
		return ast.NewStringLiteral(val), true, nil
	case "CharLiteral":
		val, _ := node["value"].(string)
		return ast.NewCharLiteral(val), true, nil
	case "NilLiteral":
		return ast.NewNilLiteral(), true, nil
	case "BooleanLiteral":
		val, _ := node["value"].(bool)
		return ast.NewBooleanLiteral(val), true, nil
	case "IntegerLiteral":
		bi := parseBigInt(node["value"])
		var suffixPtr *ast.IntegerType
		if s, ok := node["integerType"].(string); ok {
			stype := ast.IntegerType(s)
			suffixPtr = &stype
		}
		return ast.NewIntegerLiteral(bi, suffixPtr), true, nil
	case "FloatLiteral":
		val, _ := node["value"].(float64)
		var suffixPtr *ast.FloatType
		if s, ok := node["floatType"].(string); ok {
			stype := ast.FloatType(s)
			suffixPtr = &stype
		}
		return ast.NewFloatLiteral(val, suffixPtr), true, nil
	case "ArrayLiteral":
		elementsVal, _ := node["elements"].([]any)
		exprs := make([]ast.Expression, 0, len(elementsVal))
		for _, raw := range elementsVal {
			child, err := decodeNode(raw.(map[string]any))
			if err != nil {
				return nil, true, err
			}
			expr, ok := child.(ast.Expression)
			if !ok {
				return nil, true, fmt.Errorf("invalid array element %T", child)
			}
			exprs = append(exprs, expr)
		}
		return ast.NewArrayLiteral(exprs), true, nil
	case "MapLiteral":
		elementsRaw, _ := node["elements"].([]any)
		elements := make([]ast.MapLiteralElement, 0, len(elementsRaw))
		for _, raw := range elementsRaw {
			elemNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid map literal element %T", raw)
			}
			elem, err := decodeMapLiteralElement(elemNode)
			if err != nil {
				return nil, true, err
			}
			elements = append(elements, elem)
		}
		return ast.NewMapLiteral(elements), true, nil
	case "StructLiteral":
		fieldsVal, _ := node["fields"].([]any)
		fields := make([]*ast.StructFieldInitializer, 0, len(fieldsVal))
		for _, raw := range fieldsVal {
			fieldNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid struct field initializer %T", raw)
			}
			field, err := decodeStructFieldInitializer(fieldNode)
			if err != nil {
				return nil, true, err
			}
			fields = append(fields, field)
		}
		var structType *ast.Identifier
		if typeRaw, ok := node["structType"].(map[string]any); ok {
			typeNode, err := decodeNode(typeRaw)
			if err != nil {
				return nil, true, err
			}
			id, ok := typeNode.(*ast.Identifier)
			if !ok {
				return nil, true, fmt.Errorf("invalid struct type %T", typeNode)
			}
			structType = id
		}
		isPositional, _ := node["isPositional"].(bool)
		updateVals, _ := node["functionalUpdateSources"].([]any)
		updates := make([]ast.Expression, 0, len(updateVals))
		for _, raw := range updateVals {
			updateMap, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid functional update source %T", raw)
			}
			updNode, err := decodeNode(updateMap)
			if err != nil {
				return nil, true, err
			}
			expr, ok := updNode.(ast.Expression)
			if !ok {
				return nil, true, fmt.Errorf("invalid functional update source %T", updNode)
			}
			updates = append(updates, expr)
		}
		if len(updates) == 0 {
			if updateRaw, ok := node["functionalUpdateSource"].(map[string]any); ok {
				updNode, err := decodeNode(updateRaw)
				if err != nil {
					return nil, true, err
				}
				expr, ok := updNode.(ast.Expression)
				if !ok {
					return nil, true, fmt.Errorf("invalid functional update source %T", updNode)
				}
				updates = append(updates, expr)
			}
		}
		typeArgsVal, _ := node["typeArguments"].([]any)
		typeArgs := make([]ast.TypeExpression, 0, len(typeArgsVal))
		for _, raw := range typeArgsVal {
			typeNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid type argument %T", raw)
			}
			typeExpr, err := decodeTypeExpression(typeNode)
			if err != nil {
				return nil, true, err
			}
			typeArgs = append(typeArgs, typeExpr)
		}
		return ast.NewStructLiteral(fields, isPositional, structType, updates, typeArgs), true, nil
	case "IteratorLiteral":
		bodyVal, _ := node["body"].([]any)
		body := make([]ast.Statement, 0, len(bodyVal))
		for _, raw := range bodyVal {
			stmtNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid iterator body statement %T", raw)
			}
			decoded, err := decodeNode(stmtNode)
			if err != nil {
				return nil, true, err
			}
			stmt, ok := decoded.(ast.Statement)
			if !ok {
				return nil, true, fmt.Errorf("invalid iterator body node %T", decoded)
			}
			body = append(body, stmt)
		}
		literal := ast.NewIteratorLiteral(body)
		if bindingNode, ok := node["binding"].(map[string]any); ok {
			decoded, err := decodeNode(bindingNode)
			if err != nil {
				return nil, true, err
			}
			if id, ok := decoded.(*ast.Identifier); ok {
				literal.Binding = id
			} else {
				return nil, true, fmt.Errorf("invalid iterator binding %T", decoded)
			}
		}
		if elementNode, ok := node["elementType"].(map[string]any); ok {
			typExpr, err := decodeTypeExpression(elementNode)
			if err != nil {
				return nil, true, err
			}
			literal.ElementType = typExpr
		}
		return literal, true, nil
	default:
		return nil, false, nil
	}
}

func decodeMapLiteralElement(node map[string]any) (ast.MapLiteralElement, error) {
	typ, _ := node["type"].(string)
	switch typ {
	case "MapLiteralEntry":
		keyNode, _ := node["key"].(map[string]any)
		if keyNode == nil {
			return nil, fmt.Errorf("map literal entry missing key")
		}
		valueNode, _ := node["value"].(map[string]any)
		if valueNode == nil {
			return nil, fmt.Errorf("map literal entry missing value")
		}
		keyDecoded, err := decodeNode(keyNode)
		if err != nil {
			return nil, err
		}
		keyExpr, ok := keyDecoded.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid map literal key %T", keyDecoded)
		}
		valueDecoded, err := decodeNode(valueNode)
		if err != nil {
			return nil, err
		}
		valueExpr, ok := valueDecoded.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid map literal value %T", valueDecoded)
		}
		return ast.NewMapLiteralEntry(keyExpr, valueExpr), nil
	case "MapLiteralSpread":
		exprNode, _ := node["expression"].(map[string]any)
		if exprNode == nil {
			return nil, fmt.Errorf("map literal spread missing expression")
		}
		decodedNode, err := decodeNode(exprNode)
		if err != nil {
			return nil, err
		}
		decoded, ok := decodedNode.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid map literal spread expression %T", decodedNode)
		}
		return ast.NewMapLiteralSpread(decoded), nil
	default:
		return nil, fmt.Errorf("unsupported map literal element type %s", typ)
	}
}
