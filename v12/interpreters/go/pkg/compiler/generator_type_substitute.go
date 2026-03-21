package compiler

import "able/interpreter-go/pkg/ast"

func substituteTypeParams(expr ast.TypeExpression, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	return substituteTypeParamsSeen(expr, bindings, nil)
}

func substituteTypeParamsSeen(expr ast.TypeExpression, bindings map[string]ast.TypeExpression, seen map[string]struct{}) ast.TypeExpression {
	if expr == nil || len(bindings) == 0 {
		return expr
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return expr
		}
		if bound, ok := bindings[t.Name.Name]; ok && bound != nil {
			if seen == nil {
				seen = make(map[string]struct{}, 1)
			}
			if _, exists := seen[t.Name.Name]; exists {
				return expr
			}
			seen[t.Name.Name] = struct{}{}
			return substituteTypeParamsSeen(bound, bindings, seen)
		}
		return expr
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		base := substituteTypeParamsSeen(t.Base, bindings, seen)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = substituteTypeParamsSeen(arg, bindings, seen)
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = substituteTypeParamsSeen(param, bindings, seen)
		}
		return ast.NewFunctionTypeExpression(params, substituteTypeParamsSeen(t.ReturnType, bindings, seen))
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		return ast.NewNullableTypeExpression(substituteTypeParamsSeen(t.InnerType, bindings, seen))
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		return ast.NewResultTypeExpression(substituteTypeParamsSeen(t.InnerType, bindings, seen))
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		members := make([]ast.TypeExpression, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = substituteTypeParamsSeen(member, bindings, seen)
		}
		return ast.NewUnionTypeExpression(members)
	default:
		return expr
	}
}
