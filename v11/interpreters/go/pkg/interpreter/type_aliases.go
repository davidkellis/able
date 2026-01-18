package interpreter

import "able/interpreter-go/pkg/ast"

func substituteAliasTypeExpression(
	expr ast.TypeExpression,
	subst map[string]ast.TypeExpression,
	aliases map[string]*ast.TypeAliasDefinition,
	seen map[string]struct{},
) ast.TypeExpression {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			if replacement, ok := subst[t.Name.Name]; ok {
				return expandTypeAliases(replacement, aliases, seen)
			}
		}
		return expandTypeAliases(expr, aliases, seen)
	case *ast.GenericTypeExpression:
		base := substituteAliasTypeExpression(t.Base, subst, aliases, seen)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for i, arg := range t.Arguments {
			args[i] = substituteAliasTypeExpression(arg, subst, aliases, seen)
		}
		return &ast.GenericTypeExpression{Base: base, Arguments: args}
	case *ast.NullableTypeExpression:
		return &ast.NullableTypeExpression{InnerType: substituteAliasTypeExpression(t.InnerType, subst, aliases, seen)}
	case *ast.ResultTypeExpression:
		return &ast.ResultTypeExpression{InnerType: substituteAliasTypeExpression(t.InnerType, subst, aliases, seen)}
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(t.Members))
		for i, member := range t.Members {
			members[i] = substituteAliasTypeExpression(member, subst, aliases, seen)
		}
		return &ast.UnionTypeExpression{Members: members}
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for i, param := range t.ParamTypes {
			params[i] = substituteAliasTypeExpression(param, subst, aliases, seen)
		}
		return &ast.FunctionTypeExpression{
			ParamTypes: params,
			ReturnType: substituteAliasTypeExpression(t.ReturnType, subst, aliases, seen),
		}
	default:
		return expr
	}
}

func expandTypeAliases(
	expr ast.TypeExpression,
	aliases map[string]*ast.TypeAliasDefinition,
	seen map[string]struct{},
) ast.TypeExpression {
	if expr == nil || len(aliases) == 0 {
		return expr
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return expr
		}
		name := t.Name.Name
		if seen != nil {
			if _, ok := seen[name]; ok {
				return expr
			}
		}
		alias, ok := aliases[name]
		if !ok || alias.TargetType == nil {
			return expr
		}
		if seen == nil {
			seen = make(map[string]struct{})
		}
		seen[name] = struct{}{}
		expanded := expandTypeAliases(alias.TargetType, aliases, seen)
		delete(seen, name)
		return expanded
	case *ast.GenericTypeExpression:
		baseName, ok := simpleTypeName(t.Base)
		base := expandTypeAliases(t.Base, aliases, seen)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for i, arg := range t.Arguments {
			args[i] = expandTypeAliases(arg, aliases, seen)
		}
		if !ok {
			return &ast.GenericTypeExpression{Base: base, Arguments: args}
		}
		if seen != nil {
			if _, exists := seen[baseName]; exists {
				return &ast.GenericTypeExpression{Base: base, Arguments: args}
			}
		}
		alias, ok := aliases[baseName]
		if !ok || alias.TargetType == nil {
			return &ast.GenericTypeExpression{Base: base, Arguments: args}
		}
		if seen == nil {
			seen = make(map[string]struct{})
		}
		seen[baseName] = struct{}{}
		subst := make(map[string]ast.TypeExpression)
		for idx, gp := range alias.GenericParams {
			if gp == nil || gp.Name == nil {
				continue
			}
			if idx < len(args) && args[idx] != nil {
				subst[gp.Name.Name] = args[idx]
			}
		}
		substituted := substituteAliasTypeExpression(alias.TargetType, subst, aliases, seen)
		expanded := expandTypeAliases(substituted, aliases, seen)
		delete(seen, baseName)
		return expanded
	case *ast.NullableTypeExpression:
		return &ast.NullableTypeExpression{InnerType: expandTypeAliases(t.InnerType, aliases, seen)}
	case *ast.ResultTypeExpression:
		return &ast.ResultTypeExpression{InnerType: expandTypeAliases(t.InnerType, aliases, seen)}
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(t.Members))
		for i, member := range t.Members {
			members[i] = expandTypeAliases(member, aliases, seen)
		}
		return &ast.UnionTypeExpression{Members: members}
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for i, param := range t.ParamTypes {
			params[i] = expandTypeAliases(param, aliases, seen)
		}
		return &ast.FunctionTypeExpression{
			ParamTypes: params,
			ReturnType: expandTypeAliases(t.ReturnType, aliases, seen),
		}
	default:
		return expr
	}
}
