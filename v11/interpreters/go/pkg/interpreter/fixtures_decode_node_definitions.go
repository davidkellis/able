package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func decodeDefinitionNodes(node map[string]any, typ string) (ast.Node, bool, error) {
	switch typ {
	case "FunctionSignature":
		sig, err := decodeFunctionSignature(node)
		if err != nil {
			return nil, true, err
		}
		return sig, true, nil
	case "FunctionParameter":
		nameMap, ok := node["name"].(map[string]any)
		if !ok {
			return nil, true, fmt.Errorf("function parameter missing name")
		}
		pattern, err := decodePattern(nameMap)
		if err != nil {
			return nil, true, err
		}
		var paramType ast.TypeExpression
		if typeRaw, ok := node["paramType"].(map[string]any); ok {
			typeExpr, err := decodeTypeExpression(typeRaw)
			if err != nil {
				return nil, true, err
			}
			paramType = typeExpr
		}
		return ast.NewFunctionParameter(pattern, paramType), true, nil
	case "FunctionDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, true, fmt.Errorf("invalid function identifier %T", idNode)
		}
		paramsVal, _ := node["params"].([]any)
		params := make([]*ast.FunctionParameter, 0, len(paramsVal))
		for _, raw := range paramsVal {
			paramNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid function parameter node %T", raw)
			}
			decoded, err := decodeNode(paramNode)
			if err != nil {
				return nil, true, err
			}
			param, ok := decoded.(*ast.FunctionParameter)
			if !ok {
				return nil, true, fmt.Errorf("invalid function parameter %T", decoded)
			}
			params = append(params, param)
		}
		bodyNode, err := decodeNode(node["body"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		body, ok := bodyNode.(*ast.BlockExpression)
		if !ok {
			return nil, true, fmt.Errorf("function body must be block, got %T", bodyNode)
		}
		var returnType ast.TypeExpression
		if retRaw, ok := node["returnType"].(map[string]any); ok {
			rt, err := decodeTypeExpression(retRaw)
			if err != nil {
				return nil, true, err
			}
			returnType = rt
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid generic parameter node %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				generics = append(generics, gp)
			}
		}
		var inferredGenerics []*ast.GenericParameter
		if igRaw, ok := node["inferredGenericParams"].([]any); ok {
			inferredGenerics = make([]*ast.GenericParameter, 0, len(igRaw))
			for _, raw := range igRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid inferred generic parameter node %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				inferredGenerics = append(inferredGenerics, gp)
			}
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid where clause node %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, true, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		isMethodShorthand, _ := node["isMethodShorthand"].(bool)
		isPrivate, _ := node["isPrivate"].(bool)
		fn := ast.NewFunctionDefinition(id, params, body, returnType, generics, whereClause, isMethodShorthand, isPrivate)
		if len(inferredGenerics) > 0 {
			fn.InferredGenericParams = attachInferredGenericParams(inferredGenerics, generics)
		}
		return fn, true, nil
	case "ImplementationDefinition":
		ifaceNode, err := decodeNode(node["interfaceName"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		ifaceID, ok := ifaceNode.(*ast.Identifier)
		if !ok {
			return nil, true, fmt.Errorf("invalid implementation interface name %T", ifaceNode)
		}
		targetRaw, ok := node["targetType"].(map[string]any)
		if !ok {
			return nil, true, fmt.Errorf("implementation definition missing target type")
		}
		targetType, err := decodeTypeExpression(targetRaw)
		if err != nil {
			return nil, true, err
		}
		defsVal, _ := node["definitions"].([]any)
		defs := make([]*ast.FunctionDefinition, 0, len(defsVal))
		for _, raw := range defsVal {
			fnNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid implementation function %T", raw)
			}
			decoded, err := decodeNode(fnNode)
			if err != nil {
				return nil, true, err
			}
			fn, ok := decoded.(*ast.FunctionDefinition)
			if !ok {
				return nil, true, fmt.Errorf("invalid implementation function %T", decoded)
			}
			defs = append(defs, fn)
		}
		var implName *ast.Identifier
		if implRaw, ok := node["implName"].(map[string]any); ok {
			decoded, err := decodeNode(implRaw)
			if err != nil {
				return nil, true, err
			}
			id, ok := decoded.(*ast.Identifier)
			if !ok {
				return nil, true, fmt.Errorf("invalid implementation name %T", decoded)
			}
			implName = id
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid implementation generic param %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				generics = append(generics, gp)
			}
		}
		var interfaceArgs []ast.TypeExpression
		if iaRaw, ok := node["interfaceArgs"].([]any); ok {
			interfaceArgs = make([]ast.TypeExpression, 0, len(iaRaw))
			for _, raw := range iaRaw {
				argNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid implementation interface arg %T", raw)
				}
				expr, err := decodeTypeExpression(argNode)
				if err != nil {
					return nil, true, err
				}
				interfaceArgs = append(interfaceArgs, expr)
			}
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid implementation where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, true, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewImplementationDefinition(ifaceID, targetType, defs, implName, generics, interfaceArgs, whereClause, isPrivate), true, nil
	case "StructDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, true, fmt.Errorf("invalid struct identifier %T", idNode)
		}
		fieldsVal, _ := node["fields"].([]any)
		fields := make([]*ast.StructFieldDefinition, 0, len(fieldsVal))
		for _, raw := range fieldsVal {
			fieldNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid struct field %T", raw)
			}
			field, err := decodeStructFieldDefinition(fieldNode)
			if err != nil {
				return nil, true, err
			}
			fields = append(fields, field)
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid struct generic param %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				generics = append(generics, gp)
			}
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid struct where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, true, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		kindStr, _ := node["kind"].(string)
		kind := ast.StructKind(kindStr)
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewStructDefinition(id, fields, kind, generics, whereClause, isPrivate), true, nil
	case "UnionDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, true, fmt.Errorf("invalid union identifier %T", idNode)
		}
		variantsVal, _ := node["variants"].([]any)
		variants := make([]ast.TypeExpression, 0, len(variantsVal))
		for _, raw := range variantsVal {
			variantNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid union variant %T", raw)
			}
			variant, err := decodeTypeExpression(variantNode)
			if err != nil {
				return nil, true, err
			}
			variants = append(variants, variant)
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid union generic param %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				generics = append(generics, gp)
			}
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid union where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, true, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewUnionDefinition(id, variants, generics, whereClause, isPrivate), true, nil
	case "MethodsDefinition":
		targetRaw, ok := node["targetType"].(map[string]any)
		if !ok {
			return nil, true, fmt.Errorf("methods definition missing target type")
		}
		targetType, err := decodeTypeExpression(targetRaw)
		if err != nil {
			return nil, true, err
		}
		defsVal, _ := node["definitions"].([]any)
		defs := make([]*ast.FunctionDefinition, 0, len(defsVal))
		for _, raw := range defsVal {
			defNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid methods definition entry %T", raw)
			}
			decoded, err := decodeNode(defNode)
			if err != nil {
				return nil, true, err
			}
			fn, ok := decoded.(*ast.FunctionDefinition)
			if !ok {
				return nil, true, fmt.Errorf("invalid methods definition function %T", decoded)
			}
			defs = append(defs, fn)
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid methods generic %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				generics = append(generics, gp)
			}
		}
		var where []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			where = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid methods where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, true, err
				}
				where = append(where, wc)
			}
		}
		return ast.NewMethodsDefinition(targetType, defs, generics, where), true, nil
	case "InterfaceDefinition":
		idNode, err := decodeNode(node["id"].(map[string]any))
		if err != nil {
			return nil, true, err
		}
		id, ok := idNode.(*ast.Identifier)
		if !ok {
			return nil, true, fmt.Errorf("invalid interface identifier %T", idNode)
		}
		sigsVal, _ := node["signatures"].([]any)
		signatures := make([]*ast.FunctionSignature, 0, len(sigsVal))
		for _, raw := range sigsVal {
			sigNode, ok := raw.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("invalid interface signature %T", raw)
			}
			sig, err := decodeFunctionSignature(sigNode)
			if err != nil {
				return nil, true, err
			}
			signatures = append(signatures, sig)
		}
		var generics []*ast.GenericParameter
		if gpRaw, ok := node["genericParams"].([]any); ok {
			generics = make([]*ast.GenericParameter, 0, len(gpRaw))
			for _, raw := range gpRaw {
				gpNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid interface generic %T", raw)
				}
				gp, err := decodeGenericParameter(gpNode)
				if err != nil {
					return nil, true, err
				}
				generics = append(generics, gp)
			}
		}
		var selfType ast.TypeExpression
		if selfRaw, ok := node["selfTypePattern"].(map[string]any); ok {
			typeExpr, err := decodeTypeExpression(selfRaw)
			if err != nil {
				return nil, true, err
			}
			selfType = typeExpr
		}
		var whereClause []*ast.WhereClauseConstraint
		if wcRaw, ok := node["whereClause"].([]any); ok {
			whereClause = make([]*ast.WhereClauseConstraint, 0, len(wcRaw))
			for _, raw := range wcRaw {
				wcNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid where clause %T", raw)
				}
				wc, err := decodeWhereClauseConstraint(wcNode)
				if err != nil {
					return nil, true, err
				}
				whereClause = append(whereClause, wc)
			}
		}
		var baseInterfaces []ast.TypeExpression
		if baseRaw, ok := node["baseInterfaces"].([]any); ok {
			baseInterfaces = make([]ast.TypeExpression, 0, len(baseRaw))
			for _, raw := range baseRaw {
				baseNode, ok := raw.(map[string]any)
				if !ok {
					return nil, true, fmt.Errorf("invalid base interface %T", raw)
				}
				typeExpr, err := decodeTypeExpression(baseNode)
				if err != nil {
					return nil, true, err
				}
				baseInterfaces = append(baseInterfaces, typeExpr)
			}
		}
		isPrivate, _ := node["isPrivate"].(bool)
		return ast.NewInterfaceDefinition(id, signatures, generics, selfType, whereClause, baseInterfaces, isPrivate), true, nil
	case "StructFieldDefinition":
		field, err := decodeStructFieldDefinition(node)
		if err != nil {
			return nil, true, err
		}
		return field, true, nil
	case "StructFieldInitializer":
		field, err := decodeStructFieldInitializer(node)
		if err != nil {
			return nil, true, err
		}
		return field, true, nil
	default:
		return nil, false, nil
	}
}

func attachInferredGenericParams(inferred []*ast.GenericParameter, generics []*ast.GenericParameter) []*ast.GenericParameter {
	if len(inferred) == 0 {
		return nil
	}
	if len(generics) == 0 {
		return inferred
	}
	lookup := make(map[string]*ast.GenericParameter, len(generics))
	for _, param := range generics {
		if param == nil || param.Name == nil {
			continue
		}
		name := param.Name.Name
		if name == "" {
			continue
		}
		lookup[name] = param
	}
	result := make([]*ast.GenericParameter, 0, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			result = append(result, param)
			continue
		}
		name := param.Name.Name
		if existing, ok := lookup[name]; ok {
			if param.IsInferred && !existing.IsInferred {
				existing.IsInferred = true
			}
			result = append(result, existing)
			continue
		}
		result = append(result, param)
	}
	return result
}
