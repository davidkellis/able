package typechecker

import "able/interpreter-go/pkg/ast"

// resolveLocalTypeReference resolves an expression annotation against the
// lexical environment so renamed imports retain their canonical type.
func (c *Checker) resolveLocalTypeReference(
	env *Environment,
	expr ast.TypeExpression,
) Type {
	if expr == nil {
		return UnknownType{}
	}
	switch typed := expr.(type) {
	case *ast.SimpleTypeExpression:
		if typed.Name == nil || typed.Name.Name == "" {
			return UnknownType{}
		}
		name := typed.Name.Name
		if c.typeParamInScope(name) || name == "Self" || name == "_" {
			return c.resolveTypeReference(expr)
		}
		if env != nil {
			if resolved, ok := env.Lookup(name); ok {
				if alias, ok := resolved.(AliasType); ok {
					instantiated, substitution := instantiateAlias(alias, nil)
					c.verifyAliasConstraints(alias, substitution, typed)
					return instantiated
				}
				return resolved
			}
		}
		return c.resolveTypeReference(expr)

	case *ast.GenericTypeExpression:
		var base Type
		if simple, ok := typed.Base.(*ast.SimpleTypeExpression); ok &&
			simple != nil && simple.Name != nil && env != nil {
			base, _ = env.Lookup(simple.Name.Name)
		}
		if base == nil {
			base = c.resolveTypeReferenceWithOptions(typed.Base, true)
		}
		arguments := make([]Type, len(typed.Arguments))
		for idx, argument := range typed.Arguments {
			arguments[idx] = c.resolveLocalTypeReference(env, argument)
		}
		if alias, ok := base.(AliasType); ok {
			instantiated, substitution := instantiateAlias(alias, arguments)
			c.verifyAliasConstraints(alias, substitution, typed)
			return instantiated
		}
		if union, ok := base.(UnionType); ok {
			return instantiateUnionTypeArgs(union, arguments)
		}
		if base == nil {
			return UnknownType{}
		}
		return AppliedType{Base: base, Arguments: arguments}

	case *ast.FunctionTypeExpression:
		params := make([]Type, len(typed.ParamTypes))
		for idx, param := range typed.ParamTypes {
			params[idx] = c.resolveLocalTypeReference(env, param)
		}
		return FunctionType{
			Params: params,
			Return: c.resolveLocalTypeReference(env, typed.ReturnType),
		}

	case *ast.NullableTypeExpression:
		return NullableType{
			Inner: c.resolveLocalTypeReference(env, typed.InnerType),
		}

	case *ast.ResultTypeExpression:
		inner := c.resolveLocalTypeReference(env, typed.InnerType)
		if env != nil {
			if result, ok := env.Lookup("Result"); ok {
				if union, ok := result.(UnionType); ok {
					return instantiateUnionTypeArgs(union, []Type{inner})
				}
				if alias, ok := result.(AliasType); ok {
					instantiated, substitution := instantiateAlias(alias, []Type{inner})
					c.verifyAliasConstraints(alias, substitution, typed)
					return instantiated
				}
			}
		}
		return AppliedType{
			Base:      StructType{StructName: "Result"},
			Arguments: []Type{inner},
		}

	case *ast.UnionTypeExpression:
		members := make([]Type, len(typed.Members))
		for idx, member := range typed.Members {
			members[idx] = c.resolveLocalTypeReference(env, member)
		}
		return buildUnionType(members...)
	}
	return c.resolveTypeReference(expr)
}
