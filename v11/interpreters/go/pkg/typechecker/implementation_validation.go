package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *Checker) validateImplementations() []Diagnostic {
	if len(c.implementations) == 0 || c.global == nil {
		return nil
	}

	var diags []Diagnostic
	start := c.preludeImplCount
	if start < 0 {
		start = 0
	}
	if start > len(c.implementations) {
		start = len(c.implementations)
	}
	for _, spec := range c.implementations[start:] {
		if spec.InterfaceName == "" {
			continue
		}

		decl, ok := c.global.Lookup(spec.InterfaceName)
		if !ok {
			continue
		}
		iface, _, ok := resolveInterfaceDecl(decl, spec.InterfaceArgs)
		if !ok {
			continue
		}
		if len(iface.Methods) == 0 {
			continue
		}

		subst := buildImplementationSubstitution(c, spec, iface)
		label := fmt.Sprintf("impl %s for %s", spec.InterfaceName, describeImplTarget(c, spec))

		for name, ifaceMethod := range iface.Methods {
			expected := substituteFunctionType(ifaceMethod, subst)
			actual, ok := spec.Methods[name]
			if !ok {
				if iface.DefaultMethods != nil && iface.DefaultMethods[name] {
					continue
				}
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s missing method '%s'", label, name),
					Node:    implementationMethodNode(spec.Definition, name),
				})
				continue
			}
			diags = append(diags, compareImplementationMethodSignature(label, spec, name, expected, actual)...)
		}
	}
	return diags
}

func buildImplementationSubstitution(c *Checker, spec ImplementationSpec, iface InterfaceType) map[string]Type {
	subst := make(map[string]Type, len(iface.TypeParams)+1)
	selfType := spec.Target
	if spec.Definition != nil && spec.Definition.TargetType != nil && len(spec.InterfaceArgs) > 0 {
		explicitParams := interfaceExplicitParamCountFromType(iface)
		if explicitParams > 0 {
			implNames := collectGenericParamNameSet(spec.TypeParams)
			if isTypeConstructorTarget(spec.Definition.TargetType, implNames, c.global) {
				selfType = applyInterfaceArgsToTargetType(selfType, spec.InterfaceArgs)
			}
		}
	}
	if selfType != nil && !isUnknownType(selfType) {
		subst["Self"] = selfType
	} else {
		subst["Self"] = UnknownType{}
	}
	for idx, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		var replacement Type = TypeParameterType{ParameterName: param.Name}
		if idx < len(spec.InterfaceArgs) && spec.InterfaceArgs[idx] != nil && !isUnknownType(spec.InterfaceArgs[idx]) {
			replacement = spec.InterfaceArgs[idx]
		}
		subst[param.Name] = replacement
	}
	applySelfPatternBindings(c, subst, spec, iface)
	return subst
}

func applySelfPatternBindings(c *Checker, subst map[string]Type, spec ImplementationSpec, iface InterfaceType) {
	if spec.Definition == nil || spec.Definition.TargetType == nil || iface.SelfTypePattern == nil {
		return
	}
	if c == nil || c.global == nil {
		return
	}
	interfaceGenerics := collectGenericParamNameSet(iface.TypeParams)
	bindings := make(map[string]ast.TypeExpression)
	matcher := &declarationCollector{env: c.global, localTypeNames: c.localTypeNames}
	if !matcher.matchSelfTypePattern(iface.SelfTypePattern, spec.Definition.TargetType, interfaceGenerics, bindings) {
		return
	}
	scope := make(map[string]Type, len(spec.TypeParams)+len(subst))
	for name, typ := range subst {
		scope[name] = typ
	}
	for _, param := range spec.TypeParams {
		if param.Name == "" {
			continue
		}
		if _, ok := scope[param.Name]; ok {
			continue
		}
		scope[param.Name] = TypeParameterType{ParameterName: param.Name}
	}
	for name, expr := range bindings {
		if name == "" {
			continue
		}
		if existing, ok := subst[name]; ok {
			if !isUnknownType(existing) {
				if param, ok := existing.(TypeParameterType); !ok || param.ParameterName != name {
					continue
				}
			}
		}
		subst[name] = matcher.resolveTypeExpressionWithOptions(expr, scope, typeResolutionOptions{allowTypeConstructors: true})
	}
}

func applyInterfaceArgsToTargetType(target Type, args []Type) Type {
	if target == nil || len(args) == 0 {
		return target
	}
	if isUnknownType(target) {
		return target
	}
	base, baseArgs, ok := flattenAppliedType(target)
	if !ok {
		return AppliedType{Base: target, Arguments: append([]Type{}, args...)}
	}
	combined := append([]Type{}, baseArgs...)
	argIdx := 0
	for i := range combined {
		if argIdx >= len(args) {
			break
		}
		if isUnknownType(combined[i]) {
			combined[i] = args[argIdx]
			argIdx++
		}
	}
	if argIdx < len(args) {
		combined = append(combined, args[argIdx:]...)
	}
	return AppliedType{Base: base, Arguments: combined}
}

func flattenAppliedType(target Type) (Type, []Type, bool) {
	applied, ok := target.(AppliedType)
	if !ok {
		return target, nil, false
	}
	base := applied.Base
	args := append([]Type{}, applied.Arguments...)
	for {
		next, ok := base.(AppliedType)
		if !ok {
			return base, args, true
		}
		args = append(append([]Type{}, next.Arguments...), args...)
		base = next.Base
	}
}

func compareImplementationMethodSignature(label string, spec ImplementationSpec, methodName string, expected, actual FunctionType) []Diagnostic {
	var diags []Diagnostic
	node := implementationMethodNode(spec.Definition, methodName)

	expectedParams := expected.TypeParams
	actualParams := actual.TypeParams
	if len(spec.TypeParams) > 0 && len(actualParams) > 0 {
		implParams := collectGenericParamNameSet(spec.TypeParams)
		filtered := make([]GenericParamSpec, 0, len(actualParams))
		for _, param := range actualParams {
			if param.Name == "" {
				filtered = append(filtered, param)
				continue
			}
			if _, ok := implParams[param.Name]; ok {
				continue
			}
			filtered = append(filtered, param)
		}
		actualParams = filtered
	}

	if len(expectedParams) != len(actualParams) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' expects %d generic parameter(s), got %d",
				label, methodName, len(expectedParams), len(actualParams),
			),
			Node: node,
		})
	}

	if len(expected.Params) != len(actual.Params) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' expects %d parameter(s), got %d",
				label, methodName, len(expected.Params), len(actual.Params),
			),
			Node: node,
		})
	} else {
		for idx := range expected.Params {
			if !typesEquivalentForSignature(expected.Params[idx], actual.Params[idx]) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf(
						"typechecker: %s method '%s' parameter %d expected %s, got %s",
						label, methodName, idx+1,
						formatTypeForMessage(expected.Params[idx]),
						formatTypeForMessage(actual.Params[idx]),
					),
					Node: node,
				})
			}
		}
	}

	if !typesEquivalentForSignature(expected.Return, actual.Return) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' return type expected %s, got %s",
				label, methodName,
				formatTypeForMessage(expected.Return),
				formatTypeForMessage(actual.Return),
			),
			Node: node,
		})
	}

	methodWhereCount := 0
	if spec.MethodWhereClauseCounts != nil {
		methodWhereCount = spec.MethodWhereClauseCounts[methodName]
	}
	if len(expected.Where) != methodWhereCount {
		diagNode := node
		if spec.Definition != nil {
			diagNode = spec.Definition
		}
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf(
				"typechecker: %s method '%s' expects %d where-clause constraint(s), got %d",
				label, methodName, len(expected.Where), methodWhereCount,
			),
			Node: diagNode,
		})
	}

	return diags
}

func implementationMethodNode(def *ast.ImplementationDefinition, name string) ast.Node {
	if def == nil {
		return nil
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			continue
		}
		if fn.ID.Name == name {
			return fn
		}
	}
	return def
}

func describeImplTarget(c *Checker, spec ImplementationSpec) string {
	if c != nil && spec.Definition != nil && spec.Definition.TargetType != nil {
		implParams := collectGenericParamNameSet(spec.TypeParams)
		expr := c.expandTypeAliasesForLabel(spec.Definition.TargetType, make(map[string]struct{}))
		expr = c.canonicalizeTypeExpressionForLabel(expr, implParams)
		label := formatTypeExpressionNode(expr)
		if label != "" && label != "unknown" {
			return label
		}
	}
	if spec.Target == nil || isUnknownType(spec.Target) {
		return "<unknown>"
	}
	return formatType(spec.Target)
}

func (c *Checker) expandTypeAliasesForLabel(expr ast.TypeExpression, seen map[string]struct{}) ast.TypeExpression {
	if expr == nil || c == nil {
		return expr
	}
	switch node := expr.(type) {
	case *ast.SimpleTypeExpression:
		name := identifierName(node.Name)
		if name == "" || c.global == nil {
			return expr
		}
		typ, ok := c.global.Lookup(name)
		if !ok {
			return expr
		}
		alias, ok := typ.(AliasType)
		if !ok || alias.Definition == nil || alias.Definition.TargetType == nil {
			return expr
		}
		if _, ok := seen[name]; ok {
			return expr
		}
		seen[name] = struct{}{}
		expanded := c.expandTypeAliasesForLabel(alias.Definition.TargetType, seen)
		delete(seen, name)
		if expanded != nil {
			return expanded
		}
		return expr
	case *ast.GenericTypeExpression:
		baseName := ""
		if base, ok := node.Base.(*ast.SimpleTypeExpression); ok {
			baseName = identifierName(base.Name)
		}
		expandedBase := c.expandTypeAliasesForLabel(node.Base, seen)
		expandedArgs := make([]ast.TypeExpression, len(node.Arguments))
		for i, arg := range node.Arguments {
			expandedArgs[i] = c.expandTypeAliasesForLabel(arg, seen)
		}
		if baseName != "" && c.global != nil {
			if typ, ok := c.global.Lookup(baseName); ok {
				if alias, ok := typ.(AliasType); ok && alias.Definition != nil && alias.Definition.TargetType != nil {
					if _, ok := seen[baseName]; !ok {
						substitutions := make(map[string]ast.TypeExpression)
						for i, param := range alias.Definition.GenericParams {
							paramName := identifierName(param.Name)
							if paramName == "" {
								continue
							}
							if i < len(expandedArgs) && expandedArgs[i] != nil {
								substitutions[paramName] = expandedArgs[i]
							} else {
								substitutions[paramName] = ast.NewWildcardTypeExpression()
							}
						}
						seen[baseName] = struct{}{}
						substituted := substituteTypeExpressionForLabel(alias.Definition.TargetType, substitutions)
						expanded := c.expandTypeAliasesForLabel(substituted, seen)
						delete(seen, baseName)
						if expanded != nil {
							return expanded
						}
						return substituted
					}
				}
			}
		}
		if expandedBase == nil {
			return expr
		}
		return ast.NewGenericTypeExpression(expandedBase, expandedArgs)
	case *ast.NullableTypeExpression:
		inner := c.expandTypeAliasesForLabel(node.InnerType, seen)
		if inner == nil {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		inner := c.expandTypeAliasesForLabel(node.InnerType, seen)
		if inner == nil {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(node.Members))
		for i, member := range node.Members {
			members[i] = c.expandTypeAliasesForLabel(member, seen)
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(node.ParamTypes))
		for i, param := range node.ParamTypes {
			params[i] = c.expandTypeAliasesForLabel(param, seen)
		}
		ret := c.expandTypeAliasesForLabel(node.ReturnType, seen)
		if ret == nil {
			return expr
		}
		return ast.NewFunctionTypeExpression(params, ret)
	default:
		return expr
	}
}

func substituteTypeExpressionForLabel(expr ast.TypeExpression, substitutions map[string]ast.TypeExpression) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	switch node := expr.(type) {
	case *ast.SimpleTypeExpression:
		name := identifierName(node.Name)
		if name != "" {
			if sub, ok := substitutions[name]; ok && sub != nil {
				return sub
			}
		}
		return expr
	case *ast.GenericTypeExpression:
		base := substituteTypeExpressionForLabel(node.Base, substitutions)
		args := make([]ast.TypeExpression, len(node.Arguments))
		for i, arg := range node.Arguments {
			args[i] = substituteTypeExpressionForLabel(arg, substitutions)
		}
		if base == nil {
			return expr
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		inner := substituteTypeExpressionForLabel(node.InnerType, substitutions)
		if inner == nil {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		inner := substituteTypeExpressionForLabel(node.InnerType, substitutions)
		if inner == nil {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(node.Members))
		for i, member := range node.Members {
			members[i] = substituteTypeExpressionForLabel(member, substitutions)
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(node.ParamTypes))
		for i, param := range node.ParamTypes {
			params[i] = substituteTypeExpressionForLabel(param, substitutions)
		}
		ret := substituteTypeExpressionForLabel(node.ReturnType, substitutions)
		if ret == nil {
			return expr
		}
		return ast.NewFunctionTypeExpression(params, ret)
	default:
		return expr
	}
}

func (c *Checker) canonicalizeTypeExpressionForLabel(expr ast.TypeExpression, implParams map[string]struct{}) ast.TypeExpression {
	if expr == nil || c == nil {
		return expr
	}
	switch node := expr.(type) {
	case *ast.SimpleTypeExpression:
		name := identifierName(node.Name)
		if name == "" {
			return expr
		}
		if name == "_" {
			return ast.NewWildcardTypeExpression()
		}
		if _, ok := implParams[name]; ok {
			return ast.NewWildcardTypeExpression()
		}
		if c.localTypeNames != nil {
			if _, ok := c.localTypeNames[name]; ok {
				return expr
			}
		}
		if c.global != nil {
			if typ, ok := c.global.Lookup(name); ok {
				if resolved := c.typeExpressionForLabelFromType(typ); resolved != nil {
					return resolved
				}
			}
		}
		return expr
	case *ast.GenericTypeExpression:
		base := c.canonicalizeTypeExpressionForLabel(node.Base, implParams)
		args := make([]ast.TypeExpression, len(node.Arguments))
		for i, arg := range node.Arguments {
			args[i] = c.canonicalizeTypeExpressionForLabel(arg, implParams)
		}
		if base == nil {
			return expr
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		inner := c.canonicalizeTypeExpressionForLabel(node.InnerType, implParams)
		if inner == nil {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		inner := c.canonicalizeTypeExpressionForLabel(node.InnerType, implParams)
		if inner == nil {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(node.Members))
		for i, member := range node.Members {
			members[i] = c.canonicalizeTypeExpressionForLabel(member, implParams)
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(node.ParamTypes))
		for i, param := range node.ParamTypes {
			params[i] = c.canonicalizeTypeExpressionForLabel(param, implParams)
		}
		ret := c.canonicalizeTypeExpressionForLabel(node.ReturnType, implParams)
		if ret == nil {
			return expr
		}
		return ast.NewFunctionTypeExpression(params, ret)
	default:
		return expr
	}
}

func (c *Checker) typeExpressionForLabelFromType(t Type) ast.TypeExpression {
	if t == nil {
		return ast.NewWildcardTypeExpression()
	}
	switch v := t.(type) {
	case UnknownType:
		return ast.NewWildcardTypeExpression()
	case TypeParameterType:
		return ast.NewWildcardTypeExpression()
	case PrimitiveType:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(formatType(v)))
	case IntegerType:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(formatType(v)))
	case FloatType:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(formatType(v)))
	case StructType:
		return typeExpressionWithWildcards(v.StructName, len(v.TypeParams))
	case StructInstanceType:
		if len(v.TypeArgs) > 0 {
			args := make([]ast.TypeExpression, len(v.TypeArgs))
			for i, arg := range v.TypeArgs {
				args[i] = c.typeExpressionForLabelFromType(arg)
			}
			return ast.NewGenericTypeExpression(ast.NewSimpleTypeExpression(ast.NewIdentifier(v.StructName)), args)
		}
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(v.StructName))
	case InterfaceType:
		return typeExpressionWithWildcards(v.InterfaceName, len(v.TypeParams))
	case ArrayType:
		return ast.NewGenericTypeExpression(
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Array")),
			[]ast.TypeExpression{c.typeExpressionForLabelFromType(v.Element)},
		)
	case RangeType:
		return ast.NewGenericTypeExpression(
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Range")),
			[]ast.TypeExpression{c.typeExpressionForLabelFromType(v.Element)},
		)
	case IteratorType:
		return ast.NewGenericTypeExpression(
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Iterator")),
			[]ast.TypeExpression{c.typeExpressionForLabelFromType(v.Element)},
		)
	case FutureType:
		return ast.NewGenericTypeExpression(
			ast.NewSimpleTypeExpression(ast.NewIdentifier("Future")),
			[]ast.TypeExpression{c.typeExpressionForLabelFromType(v.Result)},
		)
	case NullableType:
		return ast.NewNullableTypeExpression(c.typeExpressionForLabelFromType(v.Inner))
	case UnionLiteralType:
		members := make([]ast.TypeExpression, len(v.Members))
		for i, member := range v.Members {
			members[i] = c.typeExpressionForLabelFromType(member)
		}
		return ast.NewUnionTypeExpression(members)
	case FunctionType:
		params := make([]ast.TypeExpression, len(v.Params))
		for i, param := range v.Params {
			params[i] = c.typeExpressionForLabelFromType(param)
		}
		return ast.NewFunctionTypeExpression(params, c.typeExpressionForLabelFromType(v.Return))
	case AppliedType:
		base := c.typeExpressionForLabelFromType(v.Base)
		if len(v.Arguments) == 0 {
			return base
		}
		args := make([]ast.TypeExpression, len(v.Arguments))
		for i, arg := range v.Arguments {
			args[i] = c.typeExpressionForLabelFromType(arg)
		}
		if base == nil {
			return ast.NewWildcardTypeExpression()
		}
		return ast.NewGenericTypeExpression(base, args)
	case AliasType:
		if v.Definition != nil && v.Definition.TargetType != nil {
			return c.expandTypeAliasesForLabel(v.Definition.TargetType, make(map[string]struct{}))
		}
		if v.AliasName != "" {
			return ast.NewSimpleTypeExpression(ast.NewIdentifier(v.AliasName))
		}
		return ast.NewWildcardTypeExpression()
	default:
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(formatType(t)))
	}
}

func typeExpressionWithWildcards(name string, count int) ast.TypeExpression {
	if name == "" {
		return ast.NewWildcardTypeExpression()
	}
	if count <= 0 {
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(name))
	}
	args := make([]ast.TypeExpression, count)
	for i := range args {
		args[i] = ast.NewWildcardTypeExpression()
	}
	return ast.NewGenericTypeExpression(ast.NewSimpleTypeExpression(ast.NewIdentifier(name)), args)
}

func formatTypeForMessage(t Type) string {
	return formatType(t)
}

func typesEquivalentForSignature(a, b Type) bool {
	if a == nil || b == nil {
		return isUnknownType(a) || isUnknownType(b)
	}
	if isUnknownType(a) || isUnknownType(b) {
		return true
	}

	switch av := a.(type) {
	case TypeParameterType:
		_, ok := b.(TypeParameterType)
		return ok
	case StructType:
		switch bv := b.(type) {
		case StructType:
			return av.StructName == bv.StructName
		case StructInstanceType:
			return av.StructName == bv.StructName
		}
	case StructInstanceType:
		switch bv := b.(type) {
		case StructType:
			return av.StructName == bv.StructName
		case StructInstanceType:
			return av.StructName == bv.StructName
		case AppliedType:
			return typesEquivalentForSignature(av, bv.Base)
		}
	case AppliedType:
		switch bv := b.(type) {
		case AppliedType:
			if !typesEquivalentForSignature(av.Base, bv.Base) {
				return false
			}
			if len(av.Arguments) != len(bv.Arguments) {
				return false
			}
			for i := range av.Arguments {
				if !typesEquivalentForSignature(av.Arguments[i], bv.Arguments[i]) {
					return false
				}
			}
			return true
		case StructType, StructInstanceType:
			return typesEquivalentForSignature(av.Base, bv)
		}
		return false
	case ArrayType:
		if bv, ok := b.(ArrayType); ok {
			return typesEquivalentForSignature(av.Element, bv.Element)
		}
	case NullableType:
		if bv, ok := b.(NullableType); ok {
			return typesEquivalentForSignature(av.Inner, bv.Inner)
		}
	case RangeType:
		if bv, ok := b.(RangeType); ok {
			return typesEquivalentForSignature(av.Element, bv.Element)
		}
	case UnionLiteralType:
		if bv, ok := b.(UnionLiteralType); ok {
			if len(av.Members) != len(bv.Members) {
				return false
			}
			for i := range av.Members {
				if !typesEquivalentForSignature(av.Members[i], bv.Members[i]) {
					return false
				}
			}
			return true
		}
	case FunctionType:
		bv, ok := b.(FunctionType)
		if !ok {
			return false
		}
		if len(av.Params) != len(bv.Params) {
			return false
		}
		for i := range av.Params {
			if !typesEquivalentForSignature(av.Params[i], bv.Params[i]) {
				return false
			}
		}
		return typesEquivalentForSignature(av.Return, bv.Return)
	case FutureType:
		if bv, ok := b.(FutureType); ok {
			return typesEquivalentForSignature(av.Result, bv.Result)
		}
	}

	return a.Name() == b.Name()
}
