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
		var args []ast.TypeExpression
		for idx, arg := range t.Arguments {
			expandedArg := expandTypeAliases(arg, aliases, seen)
			if expandedArg != arg {
				if args == nil {
					args = append([]ast.TypeExpression(nil), t.Arguments...)
				}
				args[idx] = expandedArg
			}
		}
		currentArgs := t.Arguments
		if args != nil {
			currentArgs = args
		}
		if !ok {
			if base == t.Base && args == nil {
				return expr
			}
			return &ast.GenericTypeExpression{Base: base, Arguments: currentArgs}
		}
		if seen != nil {
			if _, exists := seen[baseName]; exists {
				if base == t.Base && args == nil {
					return expr
				}
				return &ast.GenericTypeExpression{Base: base, Arguments: currentArgs}
			}
		}
		alias, ok := aliases[baseName]
		if !ok || alias.TargetType == nil {
			if base == t.Base && args == nil {
				return expr
			}
			return &ast.GenericTypeExpression{Base: base, Arguments: currentArgs}
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
			if idx < len(currentArgs) && currentArgs[idx] != nil {
				subst[gp.Name.Name] = currentArgs[idx]
			}
		}
		substituted := substituteAliasTypeExpression(alias.TargetType, subst, aliases, seen)
		expanded := expandTypeAliases(substituted, aliases, seen)
		delete(seen, baseName)
		return expanded
	case *ast.NullableTypeExpression:
		inner := expandTypeAliases(t.InnerType, aliases, seen)
		if inner == t.InnerType {
			return expr
		}
		return &ast.NullableTypeExpression{InnerType: inner}
	case *ast.ResultTypeExpression:
		inner := expandTypeAliases(t.InnerType, aliases, seen)
		if inner == t.InnerType {
			return expr
		}
		return &ast.ResultTypeExpression{InnerType: inner}
	case *ast.UnionTypeExpression:
		var members []ast.TypeExpression
		for idx, member := range t.Members {
			expandedMember := expandTypeAliases(member, aliases, seen)
			if expandedMember != member {
				if members == nil {
					members = append([]ast.TypeExpression(nil), t.Members...)
				}
				members[idx] = expandedMember
			}
		}
		if members == nil {
			return expr
		}
		return &ast.UnionTypeExpression{Members: members}
	case *ast.FunctionTypeExpression:
		var params []ast.TypeExpression
		for idx, param := range t.ParamTypes {
			expandedParam := expandTypeAliases(param, aliases, seen)
			if expandedParam != param {
				if params == nil {
					params = append([]ast.TypeExpression(nil), t.ParamTypes...)
				}
				params[idx] = expandedParam
			}
		}
		returnType := expandTypeAliases(t.ReturnType, aliases, seen)
		if params == nil && returnType == t.ReturnType {
			return expr
		}
		if params == nil {
			params = t.ParamTypes
		}
		return &ast.FunctionTypeExpression{
			ParamTypes: params,
			ReturnType: returnType,
		}
	default:
		return expr
	}
}

func typeExpressionReferencesAlias(
	expr ast.TypeExpression,
	aliases map[string]*ast.TypeAliasDefinition,
) bool {
	if expr == nil || len(aliases) == 0 {
		return false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return false
		}
		_, ok := aliases[t.Name.Name]
		return ok
	case *ast.GenericTypeExpression:
		if typeExpressionReferencesAlias(t.Base, aliases) {
			return true
		}
		for _, arg := range t.Arguments {
			if typeExpressionReferencesAlias(arg, aliases) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return typeExpressionReferencesAlias(t.InnerType, aliases)
	case *ast.ResultTypeExpression:
		return typeExpressionReferencesAlias(t.InnerType, aliases)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if typeExpressionReferencesAlias(member, aliases) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		for _, param := range t.ParamTypes {
			if typeExpressionReferencesAlias(param, aliases) {
				return true
			}
		}
		return typeExpressionReferencesAlias(t.ReturnType, aliases)
	default:
		return false
	}
}
