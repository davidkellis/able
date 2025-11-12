package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter10-go/pkg/ast"
)

var primitiveTypeNameSet = map[string]struct{}{
	"i8":     {},
	"i16":    {},
	"i32":    {},
	"i64":    {},
	"i128":   {},
	"u8":     {},
	"u16":    {},
	"u32":    {},
	"u64":    {},
	"u128":   {},
	"f32":    {},
	"f64":    {},
	"bool":   {},
	"string": {},
	"char":   {},
	"nil":    {},
	"void":   {},
}

func collectGenericParamNameSet(params []GenericParamSpec) map[string]struct{} {
	names := make(map[string]struct{}, len(params))
	for _, param := range params {
		if param.Name == "" {
			continue
		}
		names[param.Name] = struct{}{}
	}
	return names
}

func (c *declarationCollector) validateImplementationSelfTypePattern(
	def *ast.ImplementationDefinition,
	iface InterfaceType,
	interfaceName string,
	targetLabel string,
	implGenericNames map[string]struct{},
) bool {
	if def == nil {
		return false
	}
	interfaceLabel := nonEmpty(interfaceName)
	pattern := iface.SelfTypePattern
	if pattern != nil {
		interfaceGenerics := collectGenericParamNameSet(iface.TypeParams)
		if def.TargetType == nil || !c.doesSelfPatternMatchTarget(pattern, def.TargetType, interfaceGenerics) {
			expected := formatTypeExpressionNode(pattern)
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: impl %s for %s must match interface self type '%s'", interfaceLabel, targetLabel, expected),
				Node:    def,
			})
			return false
		}
		return true
	}

	if targetsBareTypeConstructor(def.TargetType, implGenericNames, c.env) {
		c.diags = append(c.diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: impl %s for %s cannot target a type constructor because the interface does not declare a self type (use 'for ...' to enable constructor implementations)", interfaceLabel, targetLabel),
			Node:    def,
		})
		return false
	}

	return true
}

func (c *declarationCollector) doesSelfPatternMatchTarget(
	pattern ast.TypeExpression,
	target ast.TypeExpression,
	interfaceGenericNames map[string]struct{},
) bool {
	if pattern == nil || target == nil {
		return false
	}
	return c.matchSelfTypePattern(pattern, target, interfaceGenericNames, make(map[string]ast.TypeExpression))
}

func (c *declarationCollector) matchSelfTypePattern(
	pattern ast.TypeExpression,
	target ast.TypeExpression,
	interfaceGenericNames map[string]struct{},
	bindings map[string]ast.TypeExpression,
) bool {
	switch pt := pattern.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		name := identifierName(pt.Name)
		if name == "" {
			return typeExpressionsEquivalent(pattern, target)
		}
		if c.isPatternPlaceholderName(name, interfaceGenericNames) {
			return bindSelfPatternPlaceholder(name, target, bindings)
		}
		actual, ok := target.(*ast.SimpleTypeExpression)
		if !ok {
			return false
		}
		return identifierName(actual.Name) == name
	case *ast.GenericTypeExpression:
		if patternAllowsBareConstructor(pt) {
			actual, ok := target.(*ast.SimpleTypeExpression)
			if !ok {
				return false
			}
			return c.matchSelfTypePattern(pt.Base, actual, interfaceGenericNames, bindings)
		}
		actual, ok := target.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if !c.matchSelfTypePattern(pt.Base, actual.Base, interfaceGenericNames, bindings) {
			return false
		}
		if len(pt.Arguments) != len(actual.Arguments) {
			return false
		}
		for idx := range pt.Arguments {
			expectedArg := pt.Arguments[idx]
			actualArg := actual.Arguments[idx]
			if expectedArg == nil || actualArg == nil {
				if expectedArg == nil && actualArg == nil {
					continue
				}
				return false
			}
			if isWildcardTypeExpression(expectedArg) {
				continue
			}
			if !c.matchSelfTypePattern(expectedArg, actualArg, interfaceGenericNames, bindings) {
				return false
			}
		}
		return true
	default:
		return typeExpressionsEquivalent(pattern, target)
	}
}

func bindSelfPatternPlaceholder(name string, target ast.TypeExpression, bindings map[string]ast.TypeExpression) bool {
	if existing, ok := bindings[name]; ok {
		return typeExpressionsEquivalent(existing, target)
	}
	bindings[name] = target
	return true
}

func (c *declarationCollector) isPatternPlaceholderName(name string, interfaceGenericNames map[string]struct{}) bool {
	if name == "" || name == "Self" {
		return false
	}
	if interfaceGenericNames != nil {
		if _, ok := interfaceGenericNames[name]; ok {
			return true
		}
	}
	if _, ok := primitiveTypeNameSet[name]; ok {
		return false
	}
	if c != nil && c.env != nil {
		if decl, ok := c.env.Lookup(name); ok {
			switch decl.(type) {
			case StructType, StructInstanceType, InterfaceType, UnionType:
				return false
			}
		}
	}
	return true
}

func patternAllowsBareConstructor(pattern ast.TypeExpression) bool {
	generic, ok := pattern.(*ast.GenericTypeExpression)
	if !ok || len(generic.Arguments) == 0 {
		return false
	}
	for _, arg := range generic.Arguments {
		if isWildcardTypeExpression(arg) {
			return true
		}
	}
	return false
}

func isWildcardTypeExpression(expr ast.TypeExpression) bool {
	if expr == nil {
		return false
	}
	_, ok := expr.(*ast.WildcardTypeExpression)
	return ok
}

func targetsBareTypeConstructor(
	target ast.TypeExpression,
	implementationGenericNames map[string]struct{},
	env *Environment,
) bool {
	switch tt := target.(type) {
	case *ast.SimpleTypeExpression:
		name := identifierName(tt.Name)
		if name == "" {
			return false
		}
		if implementationGenericNames != nil {
			if _, ok := implementationGenericNames[name]; ok {
				return false
			}
		}
		if _, ok := primitiveTypeNameSet[name]; ok {
			return false
		}
		if env == nil {
			return false
		}
		decl, ok := env.Lookup(name)
		if !ok {
			return false
		}
		structType, ok := decl.(StructType)
		if !ok {
			return false
		}
		return len(structType.TypeParams) > 0
	case *ast.GenericTypeExpression:
		for _, arg := range tt.Arguments {
			if isWildcardTypeExpression(arg) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func formatTypeExpressionNode(expr ast.TypeExpression) string {
	if expr == nil {
		return "unknown"
	}
	switch node := expr.(type) {
	case *ast.SimpleTypeExpression:
		name := identifierName(node.Name)
		if name == "" {
			return "unknown"
		}
		return name
	case *ast.GenericTypeExpression:
		base := formatTypeExpressionNode(node.Base)
		if len(node.Arguments) == 0 {
			return base
		}
		args := make([]string, len(node.Arguments))
		for i, arg := range node.Arguments {
			args[i] = formatTypeExpressionNode(arg)
		}
		return strings.TrimSpace(base + " " + strings.Join(args, " "))
	case *ast.FunctionTypeExpression:
		params := make([]string, len(node.ParamTypes))
		for i, param := range node.ParamTypes {
			params[i] = formatTypeExpressionNode(param)
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(params, ", "), formatTypeExpressionNode(node.ReturnType))
	case *ast.NullableTypeExpression:
		return fmt.Sprintf("%s?", formatTypeExpressionNode(node.InnerType))
	case *ast.ResultTypeExpression:
		return strings.TrimSpace("Result " + formatTypeExpressionNode(node.InnerType))
	case *ast.UnionTypeExpression:
		if len(node.Members) == 0 {
			return "Union"
		}
		members := make([]string, len(node.Members))
		for i, member := range node.Members {
			members[i] = formatTypeExpressionNode(member)
		}
		return strings.Join(members, " | ")
	case *ast.WildcardTypeExpression:
		return "_"
	default:
		return "unknown"
	}
}

func typeExpressionsEquivalent(a, b ast.TypeExpression) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	switch av := a.(type) {
	case *ast.SimpleTypeExpression:
		bv, ok := b.(*ast.SimpleTypeExpression)
		if !ok {
			return false
		}
		return identifierName(av.Name) == identifierName(bv.Name)
	case *ast.GenericTypeExpression:
		bv, ok := b.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if !typeExpressionsEquivalent(av.Base, bv.Base) {
			return false
		}
		if len(av.Arguments) != len(bv.Arguments) {
			return false
		}
		for i := range av.Arguments {
			if !typeExpressionsEquivalent(av.Arguments[i], bv.Arguments[i]) {
				return false
			}
		}
		return true
	case *ast.FunctionTypeExpression:
		bv, ok := b.(*ast.FunctionTypeExpression)
		if !ok || len(av.ParamTypes) != len(bv.ParamTypes) {
			return false
		}
		for i := range av.ParamTypes {
			if !typeExpressionsEquivalent(av.ParamTypes[i], bv.ParamTypes[i]) {
				return false
			}
		}
		return typeExpressionsEquivalent(av.ReturnType, bv.ReturnType)
	case *ast.NullableTypeExpression:
		bv, ok := b.(*ast.NullableTypeExpression)
		if !ok {
			return false
		}
		return typeExpressionsEquivalent(av.InnerType, bv.InnerType)
	case *ast.ResultTypeExpression:
		bv, ok := b.(*ast.ResultTypeExpression)
		if !ok {
			return false
		}
		return typeExpressionsEquivalent(av.InnerType, bv.InnerType)
	case *ast.UnionTypeExpression:
		bv, ok := b.(*ast.UnionTypeExpression)
		if !ok || len(av.Members) != len(bv.Members) {
			return false
		}
		for i := range av.Members {
			if !typeExpressionsEquivalent(av.Members[i], bv.Members[i]) {
				return false
			}
		}
		return true
	case *ast.WildcardTypeExpression:
		_, ok := b.(*ast.WildcardTypeExpression)
		return ok
	default:
		return formatTypeExpressionNode(a) == formatTypeExpressionNode(b)
	}
}
