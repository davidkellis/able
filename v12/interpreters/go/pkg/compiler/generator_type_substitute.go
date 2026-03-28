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
		if applied, ok := substituteAppliedGenericType(base, args); ok {
			return applied
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

func substituteAppliedGenericType(base ast.TypeExpression, args []ast.TypeExpression) (ast.TypeExpression, bool) {
	generic, ok := base.(*ast.GenericTypeExpression)
	if !ok || generic == nil || len(args) == 0 {
		return nil, false
	}
	filled, remaining, replaced := substituteFillWildcardTypeArgs(generic.Arguments, args)
	if !replaced {
		return nil, false
	}
	if len(remaining) == 0 {
		return ast.NewGenericTypeExpression(generic.Base, filled), true
	}
	return ast.NewGenericTypeExpression(ast.NewGenericTypeExpression(generic.Base, filled), remaining), true
}

func substituteFillWildcardTypeArgs(existing []ast.TypeExpression, incoming []ast.TypeExpression) ([]ast.TypeExpression, []ast.TypeExpression, bool) {
	if len(existing) == 0 || len(incoming) == 0 {
		return existing, incoming, false
	}
	filled := make([]ast.TypeExpression, len(existing))
	copy(filled, existing)
	nextArg := 0
	replaced := false
	for idx, current := range filled {
		if _, ok := current.(*ast.WildcardTypeExpression); !ok {
			continue
		}
		if nextArg >= len(incoming) {
			break
		}
		filled[idx] = incoming[nextArg]
		nextArg++
		replaced = true
	}
	return filled, incoming[nextArg:], replaced
}
