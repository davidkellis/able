package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// genericUnionMethodTarget expands a method-set target only when it names a
// generic union. Runtime values intentionally do not retain a union wrapper,
// so callers with a checked static receiver type use this structural template
// to select the correct generic-union method and bind its receiver arguments.
func (i *Interpreter) genericUnionMethodTarget(fn *runtime.FunctionValue) (ast.TypeExpression, bool) {
	if i == nil || fn == nil || fn.MethodSet == nil || fn.MethodSet.TargetType == nil || len(fn.MethodSet.GenericParams) == 0 {
		return nil, false
	}
	return i.expandGenericNamedUnionType(fn.MethodSet.TargetType)
}

// expandGenericNamedUnionType preserves the complete structural type of a
// named generic union. Runtime values deliberately do not carry that nominal
// wrapper, so static dispatch must expand both the method target and checked
// receiver through the same generic binding rule before comparing them.
func (i *Interpreter) expandGenericNamedUnionType(expr ast.TypeExpression) (ast.TypeExpression, bool) {
	if i == nil || expr == nil {
		return nil, false
	}
	target := i.expandTypeAliasesCached(expr)
	if target == nil {
		target = expr
	}

	var (
		name string
		args []ast.TypeExpression
	)
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t != nil && t.Name != nil {
			name = t.Name.Name
		}
	case *ast.GenericTypeExpression:
		if t != nil {
			if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
				name = base.Name.Name
			}
			args = t.Arguments
		}
	}
	if name == "" {
		return nil, false
	}
	union, ok := i.unionDefinitions[name]
	if !ok || union == nil || union.Node == nil || len(union.Node.GenericParams) == 0 {
		return nil, false
	}
	binding := make(map[string]ast.TypeExpression, len(union.Node.GenericParams))
	for index, param := range union.Node.GenericParams {
		if param == nil || param.Name == nil || param.Name.Name == "" || index >= len(args) {
			continue
		}
		binding[param.Name.Name] = args[index]
	}
	variants := make([]ast.TypeExpression, 0, len(union.Node.Variants))
	for _, variant := range union.Node.Variants {
		appendGenericUnionMethodVariant(&variants, substituteTypeParamsWithMap(variant, binding))
	}
	return ast.NewUnionTypeExpression(variants), true
}

func appendGenericUnionMethodVariant(variants *[]ast.TypeExpression, variant ast.TypeExpression) {
	if variants == nil || variant == nil {
		return
	}
	if union, ok := variant.(*ast.UnionTypeExpression); ok && union != nil {
		for _, member := range union.Members {
			appendGenericUnionMethodVariant(variants, member)
		}
		return
	}
	*variants = append(*variants, variant)
}

func (i *Interpreter) genericUnionMethodMatchesStaticReceiver(fn *runtime.FunctionValue, receiverType ast.TypeExpression) bool {
	if receiverType == nil {
		return false
	}
	target, ok := i.genericUnionMethodTarget(fn)
	if !ok {
		return false
	}
	actual, expanded := i.expandGenericNamedUnionType(receiverType)
	if !expanded {
		actual = i.expandTypeAliasesCached(receiverType)
		if actual == nil {
			actual = receiverType
		}
	}
	actual = genericUnionStructuralTypeExpression(actual)
	return matchTypeExpressionTemplate(target, actual, genericNameSet(fn.MethodSet.GenericParams), make(map[string]ast.TypeExpression))
}

// genericUnionStructuralTypeExpression expands the two language-level union
// shorthands before matching a named union definition structurally. This is
// not a stdlib name special case: ?T is nil | T and !T is Error | T wherever
// a checked type has retained shorthand syntax instead of its union form.
func genericUnionStructuralTypeExpression(expr ast.TypeExpression) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	switch value := expr.(type) {
	case *ast.NullableTypeExpression:
		if value == nil {
			return expr
		}
		return ast.NewUnionTypeExpression([]ast.TypeExpression{ast.Ty("nil"), genericUnionStructuralTypeExpression(value.InnerType)})
	case *ast.ResultTypeExpression:
		if value == nil {
			return expr
		}
		return ast.NewUnionTypeExpression([]ast.TypeExpression{ast.Ty("Error"), genericUnionStructuralTypeExpression(value.InnerType)})
	case *ast.UnionTypeExpression:
		if value == nil {
			return expr
		}
		members := make([]ast.TypeExpression, len(value.Members))
		for index, member := range value.Members {
			members[index] = genericUnionStructuralTypeExpression(member)
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.GenericTypeExpression:
		if value == nil {
			return expr
		}
		args := make([]ast.TypeExpression, len(value.Arguments))
		for index, arg := range value.Arguments {
			args[index] = genericUnionStructuralTypeExpression(arg)
		}
		return ast.NewGenericTypeExpression(genericUnionStructuralTypeExpression(value.Base), args)
	default:
		return expr
	}
}
