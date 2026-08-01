package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// validateCompositeInterfaceSelfPatterns runs after type declarations have
// been refreshed so bases declared later in the module have complete metadata.
func (c *declarationCollector) validateCompositeInterfaceSelfPatterns(stmts []ast.Statement) {
	if c == nil || c.env == nil {
		return
	}
	for _, stmt := range stmts {
		def, ok := stmt.(*ast.InterfaceDefinition)
		if !ok || def == nil || def.ID == nil || len(def.BaseInterfaces) == 0 {
			continue
		}
		c.validateCompositeInterfaceSelfPattern(def)
	}
}

func (c *declarationCollector) validateCompositeInterfaceSelfPattern(def *ast.InterfaceDefinition) {
	scope := compositeInterfaceTypeScope(def)
	c.addSelfPatternParamsToScope(def, scope)

	compositePattern := explicitSelfTypePattern(def.SelfTypePattern)
	for _, baseExpr := range def.BaseInterfaces {
		if baseExpr == nil {
			continue
		}
		baseType := c.resolveTypeExpression(baseExpr, scope)
		base, baseArgs, ok := resolveInterfaceDecl(baseType, nil)
		if !ok {
			// Method collection owns unresolved/non-interface base diagnostics.
			continue
		}
		basePattern := explicitSelfTypePattern(base.SelfTypePattern)
		if basePattern == nil {
			continue
		}
		basePattern = instantiateCompositeBaseSelfPattern(base, baseArgs)
		baseName := nonEmpty(base.InterfaceName)
		compositeName := nonEmpty(def.ID.Name)
		if compositePattern == nil {
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf(
					"typechecker: composite interface '%s' must declare a self type because base interface '%s' declares self type '%s'",
					compositeName,
					baseName,
					formatTypeExpressionNode(basePattern),
				),
				Node: baseExpr,
			})
			continue
		}
		baseGenerics := collectGenericParamNameSet(base.TypeParams)
		if c.doesSelfPatternMatchTarget(basePattern, compositePattern, baseGenerics) {
			continue
		}
		c.diags = append(c.diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: composite interface '%s' self type '%s' is incompatible with base interface '%s' self type '%s'",
				compositeName,
				formatTypeExpressionNode(compositePattern),
				baseName,
				formatTypeExpressionNode(basePattern),
			),
			Node: baseExpr,
		})
	}
}

func compositeInterfaceTypeScope(def *ast.InterfaceDefinition) map[string]Type {
	scope := map[string]Type{
		"Self": TypeParameterType{ParameterName: "Self"},
	}
	if def == nil {
		return scope
	}
	for _, param := range def.GenericParams {
		if param == nil || param.Name == nil || param.Name.Name == "" {
			continue
		}
		scope[param.Name.Name] = TypeParameterType{ParameterName: param.Name.Name}
	}
	if def.ID != nil && def.ID.Name != "" {
		scope[def.ID.Name] = InterfaceType{InterfaceName: def.ID.Name}
	}
	return scope
}

func explicitSelfTypePattern(pattern ast.TypeExpression) ast.TypeExpression {
	if pattern == nil || isTrivialSelfPattern(pattern) {
		return nil
	}
	return pattern
}

func instantiateCompositeBaseSelfPattern(base InterfaceType, baseArgs []Type) ast.TypeExpression {
	pattern := base.SelfTypePattern
	if len(baseArgs) == 0 || len(base.TypeParams) == 0 {
		return pattern
	}
	substitutions := make(map[string]ast.TypeExpression)
	for idx, param := range base.TypeParams {
		if param.Name == "" || idx >= len(baseArgs) || baseArgs[idx] == nil {
			continue
		}
		substitutions[param.Name] = compositeSelfPatternExpressionFromType(baseArgs[idx])
	}
	if len(substitutions) == 0 {
		return pattern
	}
	return substituteTypeExpressionForLabel(pattern, substitutions)
}

func compositeSelfPatternExpressionFromType(typ Type) ast.TypeExpression {
	if typ == nil {
		return ast.NewWildcardTypeExpression()
	}
	switch value := typ.(type) {
	case UnknownType:
		return ast.NewWildcardTypeExpression()
	case TypeParameterType:
		if value.ParameterName == "" {
			return ast.NewWildcardTypeExpression()
		}
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(value.ParameterName))
	case PrimitiveType, IntegerType, FloatType:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(formatType(typ)))
	case StructType:
		return typeExpressionWithWildcards(value.StructName, len(value.TypeParams))
	case StructInstanceType:
		if len(value.TypeArgs) == 0 {
			return ast.NewSimpleTypeExpression(ast.NewIdentifier(value.StructName))
		}
		args := make([]ast.TypeExpression, len(value.TypeArgs))
		for idx, arg := range value.TypeArgs {
			args[idx] = compositeSelfPatternExpressionFromType(arg)
		}
		return ast.NewGenericTypeExpression(
			ast.NewSimpleTypeExpression(ast.NewIdentifier(value.StructName)),
			args,
		)
	case InterfaceType:
		return typeExpressionWithWildcards(value.InterfaceName, len(value.TypeParams))
	case NullableType:
		return ast.NewNullableTypeExpression(compositeSelfPatternExpressionFromType(value.Inner))
	case UnionLiteralType:
		members := make([]ast.TypeExpression, len(value.Members))
		for idx, member := range value.Members {
			members[idx] = compositeSelfPatternExpressionFromType(member)
		}
		return ast.NewUnionTypeExpression(members)
	case FunctionType:
		params := make([]ast.TypeExpression, len(value.Params))
		for idx, param := range value.Params {
			params[idx] = compositeSelfPatternExpressionFromType(param)
		}
		return ast.NewFunctionTypeExpression(params, compositeSelfPatternExpressionFromType(value.Return))
	case AppliedType:
		base := compositeSelfPatternExpressionFromType(value.Base)
		if len(value.Arguments) == 0 {
			return base
		}
		args := make([]ast.TypeExpression, len(value.Arguments))
		for idx, arg := range value.Arguments {
			args[idx] = compositeSelfPatternExpressionFromType(arg)
		}
		return ast.NewGenericTypeExpression(base, args)
	case AliasType:
		if value.AliasName != "" {
			return ast.NewSimpleTypeExpression(ast.NewIdentifier(value.AliasName))
		}
	}
	return ast.NewSimpleTypeExpression(ast.NewIdentifier(formatType(typ)))
}
