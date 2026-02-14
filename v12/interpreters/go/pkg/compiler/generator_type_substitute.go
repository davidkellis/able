package compiler

import "able/interpreter-go/pkg/ast"

func substituteTypeParams(expr ast.TypeExpression, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if expr == nil || len(bindings) == 0 {
		return expr
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return expr
		}
		if bound, ok := bindings[t.Name.Name]; ok && bound != nil {
			return bound
		}
		return expr
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		base := substituteTypeParams(t.Base, bindings)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = substituteTypeParams(arg, bindings)
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = substituteTypeParams(param, bindings)
		}
		return ast.NewFunctionTypeExpression(params, substituteTypeParams(t.ReturnType, bindings))
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		return ast.NewNullableTypeExpression(substituteTypeParams(t.InnerType, bindings))
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		return ast.NewResultTypeExpression(substituteTypeParams(t.InnerType, bindings))
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		members := make([]ast.TypeExpression, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = substituteTypeParams(member, bindings)
		}
		return ast.NewUnionTypeExpression(members)
	default:
		return expr
	}
}
