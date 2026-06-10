package interpreter

import "able/interpreter-go/pkg/ast"

func inlineBindingsSlice(inline []ast.TypeExpression, expectedCount int) []ast.TypeExpression {
	if expectedCount <= 0 {
		return nil
	}
	if expectedCount <= len(inline) {
		return inline[:expectedCount]
	}
	return make([]ast.TypeExpression, expectedCount)
}

func indexedTypeExpressionBindingsComplete(bindings []ast.TypeExpression) bool {
	if len(bindings) == 0 {
		return true
	}
	for _, binding := range bindings {
		if binding == nil {
			return false
		}
	}
	return true
}

func (i *Interpreter) cachedTypeArgumentsFromIndexedBindings(bindings []ast.TypeExpression) []ast.TypeExpression {
	switch len(bindings) {
	case 0:
		return nil
	case 1:
		return i.cachedTypeExpressionTuple1(indexedBindingOrWildcard(bindings, 0))
	case 2:
		return i.cachedTypeExpressionTuple2(
			indexedBindingOrWildcard(bindings, 0),
			indexedBindingOrWildcard(bindings, 1),
		)
	case 3:
		return i.cachedTypeExpressionTuple3(
			indexedBindingOrWildcard(bindings, 0),
			indexedBindingOrWildcard(bindings, 1),
			indexedBindingOrWildcard(bindings, 2),
		)
	default:
		typeArgs := make([]ast.TypeExpression, len(bindings))
		for idx, binding := range bindings {
			if binding != nil {
				typeArgs[idx] = binding
				continue
			}
			typeArgs[idx] = cachedWildcardTypeExpression
		}
		return i.cachedTypeExpressionTuple(typeArgs)
	}
}

func indexedBindingOrWildcard(bindings []ast.TypeExpression, idx int) ast.TypeExpression {
	if idx >= 0 && idx < len(bindings) && bindings[idx] != nil {
		return bindings[idx]
	}
	return cachedWildcardTypeExpression
}

func matchTypeExpressionTemplateIndexed(template, actual ast.TypeExpression, genericIndex map[string]int, bindings []ast.TypeExpression) bool {
	switch t := template.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return actual == nil
		}
		name := t.Name.Name
		if slot, isGeneric := genericIndex[name]; isGeneric && slot >= 0 && slot < len(bindings) {
			existing := bindings[slot]
			if existing != nil {
				if _, ok := existing.(*ast.WildcardTypeExpression); ok {
					if actual != nil {
						if _, ok := actual.(*ast.WildcardTypeExpression); !ok {
							bindings[slot] = actual
						}
					}
					return true
				}
				if _, ok := actual.(*ast.WildcardTypeExpression); ok {
					return true
				}
				return typeExpressionsEqual(existing, actual)
			}
			bindings[slot] = actual
			return true
		}
		return typeExpressionsEqual(template, actual)
	case *ast.GenericTypeExpression:
		switch other := actual.(type) {
		case *ast.GenericTypeExpression:
			if !matchTypeExpressionTemplateIndexed(t.Base, other.Base, genericIndex, bindings) {
				return false
			}
			otherArgs := other.Arguments
			if len(otherArgs) == 0 && len(t.Arguments) > 0 {
				for idx := range t.Arguments {
					if !matchTypeExpressionTemplateIndexed(t.Arguments[idx], cachedWildcardTypeExpression, genericIndex, bindings) {
						return false
					}
				}
				return true
			}
			if len(t.Arguments) != len(otherArgs) {
				return false
			}
			for idx := range t.Arguments {
				if !matchTypeExpressionTemplateIndexed(t.Arguments[idx], otherArgs[idx], genericIndex, bindings) {
					return false
				}
			}
			return true
		case *ast.SimpleTypeExpression:
			if !matchTypeExpressionTemplateIndexed(t.Base, other, genericIndex, bindings) {
				return false
			}
			for idx := range t.Arguments {
				if !matchTypeExpressionTemplateIndexed(t.Arguments[idx], cachedWildcardTypeExpression, genericIndex, bindings) {
					return false
				}
			}
			return true
		default:
			return false
		}
	case *ast.NullableTypeExpression:
		if typeExpressionIsNilOrWildcard(actual) {
			return true
		}
		if other, ok := actual.(*ast.NullableTypeExpression); ok {
			return matchTypeExpressionTemplateIndexed(t.InnerType, other.InnerType, genericIndex, bindings)
		}
		return matchTypeExpressionTemplateIndexed(t.InnerType, actual, genericIndex, bindings)
	case *ast.ResultTypeExpression:
		other, ok := actual.(*ast.ResultTypeExpression)
		if !ok {
			return false
		}
		return matchTypeExpressionTemplateIndexed(t.InnerType, other.InnerType, genericIndex, bindings)
	case *ast.UnionTypeExpression:
		other, ok := actual.(*ast.UnionTypeExpression)
		if !ok || len(t.Members) != len(other.Members) {
			return false
		}
		for idx := range t.Members {
			if !matchTypeExpressionTemplateIndexed(t.Members[idx], other.Members[idx], genericIndex, bindings) {
				return false
			}
		}
		return true
	default:
		return typeExpressionsEqual(template, actual)
	}
}

func typeExpressionIsNilOrWildcard(expr ast.TypeExpression) bool {
	if expr == nil {
		return true
	}
	if _, ok := expr.(*ast.WildcardTypeExpression); ok {
		return true
	}
	simple, ok := expr.(*ast.SimpleTypeExpression)
	return ok && simple != nil && simple.Name != nil && simple.Name.Name == "nil"
}
