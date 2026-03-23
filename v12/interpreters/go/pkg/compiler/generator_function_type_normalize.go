package compiler

import "able/interpreter-go/pkg/ast"

func normalizeCallableSyntaxTypeExpr(expr ast.TypeExpression) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		params, changed := normalizeCallableSyntaxFunctionParams(t.ParamTypes)
		ret := normalizeCallableSyntaxTypeExpr(t.ReturnType)
		if ret != t.ReturnType {
			changed = true
		}
		if !changed {
			return expr
		}
		return ast.NewFunctionTypeExpression(params, ret)
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		base := normalizeCallableSyntaxTypeExpr(t.Base)
		changed := base != t.Base
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			next := normalizeCallableSyntaxTypeExpr(arg)
			args = append(args, next)
			if next != arg {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		inner := normalizeCallableSyntaxTypeExpr(t.InnerType)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		inner := normalizeCallableSyntaxTypeExpr(t.InnerType)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			next := normalizeCallableSyntaxTypeExpr(member)
			members = append(members, next)
			if next != member {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewUnionTypeExpression(members)
	default:
		return expr
	}
}

func normalizeCallableSyntaxFunctionParams(params []ast.TypeExpression) ([]ast.TypeExpression, bool) {
	if len(params) == 1 {
		if expanded, ok := normalizeCallableSyntaxParamList(params[0]); ok {
			return expanded, true
		}
	}
	changed := false
	out := make([]ast.TypeExpression, 0, len(params))
	for _, param := range params {
		next := normalizeCallableSyntaxTypeExpr(param)
		out = append(out, next)
		if next != param {
			changed = true
		}
	}
	return out, changed
}

func normalizeCallableSyntaxParamList(expr ast.TypeExpression) ([]ast.TypeExpression, bool) {
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil {
		return nil, false
	}
	base, ok := generic.Base.(*ast.SimpleTypeExpression)
	if !ok || base == nil || base.Name == nil || base.Name.Name != "fn" {
		return nil, false
	}
	// The parser currently represents `fn(...) -> R` as a function type whose
	// lone parameter is the generic application `fn<...>`. Canonicalize that
	// encoding back to the actual callable parameter list so later lowering sees
	// the right arity.
	if len(generic.Arguments) == 1 && isCallableUnitSentinelTypeExpr(generic.Arguments[0]) {
		return []ast.TypeExpression{}, true
	}
	out := make([]ast.TypeExpression, 0, len(generic.Arguments))
	for _, arg := range generic.Arguments {
		out = append(out, normalizeCallableSyntaxTypeExpr(arg))
	}
	return out, true
}

func isCallableUnitSentinelTypeExpr(expr ast.TypeExpression) bool {
	simple, ok := expr.(*ast.SimpleTypeExpression)
	return ok && simple != nil && simple.Name != nil && simple.Name.Name == "()"
}
