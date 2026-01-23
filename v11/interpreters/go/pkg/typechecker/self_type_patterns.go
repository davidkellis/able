package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
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
	"String": {},
	"IoHandle": {},
	"ProcHandle": {},
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
	if isTrivialSelfPattern(pattern) {
		pattern = nil
	}
	if pattern != nil {
		interfaceGenerics := collectGenericParamNameSet(iface.TypeParams)
		if patternAllowsBareConstructor(pattern) && !targetsBareTypeConstructor(def.TargetType, implGenericNames, c.env) {
			expected := formatTypeExpressionNode(pattern)
			c.diags = append(c.diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: impl %s for %s must match interface self type '%s'", interfaceLabel, targetLabel, expected),
				Node:    def,
			})
			return false
		}
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
			if actual, ok := target.(*ast.SimpleTypeExpression); ok {
				return c.matchSelfTypePattern(pt.Base, actual, interfaceGenericNames, bindings)
			}
		}
		actual, ok := target.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if patternAllowsBareConstructor(pt) {
			if simple, ok := pt.Base.(*ast.SimpleTypeExpression); ok {
				name := identifierName(simple.Name)
				if c.isPatternPlaceholderName(name, interfaceGenericNames) {
					if !bindSelfPatternPlaceholder(name, target, bindings) {
						return false
					}
				} else if !c.matchSelfTypePattern(pt.Base, actual.Base, interfaceGenericNames, bindings) {
					return false
				}
			} else if !c.matchSelfTypePattern(pt.Base, actual.Base, interfaceGenericNames, bindings) {
				return false
			}
		} else if !c.matchSelfTypePattern(pt.Base, actual.Base, interfaceGenericNames, bindings) {
			return false
		}
		if !selfPatternArgsCompatible(pt.Arguments, actual.Arguments) {
			return false
		}
		limit := len(pt.Arguments)
		if len(actual.Arguments) < limit {
			limit = len(actual.Arguments)
		}
		for idx := 0; idx < limit; idx++ {
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

func selfPatternArgsCompatible(patternArgs, targetArgs []ast.TypeExpression) bool {
	if len(patternArgs) == len(targetArgs) {
		return true
	}
	if len(patternArgs) > len(targetArgs) {
		return trailingSelfPatternWildcardOnly(patternArgs[len(targetArgs):])
	}
	return trailingSelfPatternWildcardOnly(targetArgs[len(patternArgs):])
}

func trailingSelfPatternWildcardOnly(args []ast.TypeExpression) bool {
	for _, arg := range args {
		if arg == nil || isWildcardTypeExpression(arg) {
			continue
		}
		return false
	}
	return true
}

func bindSelfPatternPlaceholder(name string, target ast.TypeExpression, bindings map[string]ast.TypeExpression) bool {
	if existing, ok := bindings[name]; ok {
		return typeExpressionsEquivalent(existing, target)
	}
	bindings[name] = target
	return true
}

func isTrivialSelfPattern(pattern ast.TypeExpression) bool {
	simple, ok := pattern.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false
	}
	return simple.Name.Name == "Self"
}

func (c *declarationCollector) isPatternPlaceholderName(name string, interfaceGenericNames map[string]struct{}) bool {
	if name == "" || name == "Self" || name == "_" {
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
	if _, ok := expr.(*ast.WildcardTypeExpression); ok {
		return true
	}
	if simple, ok := expr.(*ast.SimpleTypeExpression); ok {
		return identifierName(simple.Name) == "_"
	}
	return false
}

func targetsBareTypeConstructor(
	target ast.TypeExpression,
	implementationGenericNames map[string]struct{},
	env *Environment,
) bool {
	return isTypeConstructorTarget(target, implementationGenericNames, env)
}

func isTypeConstructorTarget(
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
		expected, ok := expectedTypeArgumentCount(name, lookupTypeName(env, name))
		return ok && expected > 0
	case *ast.GenericTypeExpression:
		if containsWildcardArgument(tt.Arguments) {
			return true
		}
		baseName := typeExpressionBaseName(tt.Base)
		if baseName == "" {
			return false
		}
		if implementationGenericNames != nil {
			if _, ok := implementationGenericNames[baseName]; ok {
				return false
			}
		}
		expected, ok := expectedTypeArgumentCount(baseName, lookupTypeName(env, baseName))
		if !ok {
			return false
		}
		return len(tt.Arguments) < expected
	default:
		return false
	}
}

func typeExpressionBaseName(expr ast.TypeExpression) string {
	switch node := expr.(type) {
	case *ast.SimpleTypeExpression:
		return identifierName(node.Name)
	case *ast.GenericTypeExpression:
		return typeExpressionBaseName(node.Base)
	default:
		return ""
	}
}

func lookupTypeName(env *Environment, name string) Type {
	if env == nil || name == "" {
		return nil
	}
	if decl, ok := env.Lookup(name); ok {
		return decl
	}
	return nil
}

func containsWildcardArgument(args []ast.TypeExpression) bool {
	for _, arg := range args {
		if isWildcardTypeExpression(arg) {
			return true
		}
	}
	return false
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

func (c *Checker) applySelfPatternConstructorSubstitution(subst map[string]Type, iface InterfaceType, self Type) {
	if c == nil || subst == nil || iface.SelfTypePattern == nil || self == nil {
		return
	}
	names := c.collectSelfPatternConstructorPlaceholders(iface)
	if len(names) == 0 {
		return
	}
	constructor := c.selfTypeConstructor(self)
	if constructor == nil || isUnknownType(constructor) {
		return
	}
	for name := range names {
		if existing, ok := subst[name]; ok && !shouldOverrideSelfPatternPlaceholder(existing, name) {
			continue
		}
		subst[name] = constructor
	}
}

func shouldOverrideSelfPatternPlaceholder(existing Type, name string) bool {
	if existing == nil || isUnknownType(existing) {
		return true
	}
	if param, ok := existing.(TypeParameterType); ok && param.ParameterName == name {
		return true
	}
	return false
}

func (c *Checker) collectSelfPatternConstructorPlaceholders(iface InterfaceType) map[string]struct{} {
	if c == nil || iface.SelfTypePattern == nil {
		return nil
	}
	interfaceGenerics := collectGenericParamNameSet(iface.TypeParams)
	matcher := &declarationCollector{env: c.global, localTypeNames: c.localTypeNames}
	placeholders := map[string]struct{}{}
	var walk func(ast.TypeExpression)
	walk = func(expr ast.TypeExpression) {
		if expr == nil {
			return
		}
		switch t := expr.(type) {
		case *ast.GenericTypeExpression:
			if simple, ok := t.Base.(*ast.SimpleTypeExpression); ok {
				name := identifierName(simple.Name)
				if matcher.isPatternPlaceholderName(name, interfaceGenerics) {
					placeholders[name] = struct{}{}
				}
			} else {
				walk(t.Base)
			}
			for _, arg := range t.Arguments {
				walk(arg)
			}
		case *ast.FunctionTypeExpression:
			for _, param := range t.ParamTypes {
				walk(param)
			}
			walk(t.ReturnType)
		case *ast.NullableTypeExpression:
			walk(t.InnerType)
		case *ast.ResultTypeExpression:
			walk(t.InnerType)
		case *ast.UnionTypeExpression:
			for _, member := range t.Members {
				walk(member)
			}
		}
	}
	walk(iface.SelfTypePattern)
	if len(placeholders) == 0 {
		return nil
	}
	return placeholders
}

func (c *Checker) selfTypeConstructor(self Type) Type {
	if self == nil || isUnknownType(self) {
		return UnknownType{}
	}
	if base, _, ok := flattenAppliedType(self); ok {
		return c.selfTypeConstructor(base)
	}
	switch v := self.(type) {
	case AliasType:
		return c.selfTypeConstructor(v.Target)
	case InterfaceType:
		if c.global != nil {
			if decl, ok := c.global.Lookup(v.InterfaceName); ok {
				if iface, ok := decl.(InterfaceType); ok {
					return iface
				}
			}
		}
		return InterfaceType{InterfaceName: v.InterfaceName}
	}
	if info, ok := structInfoFromType(self); ok && info.name != "" {
		if c.global != nil {
			if decl, ok := c.global.Lookup(info.name); ok {
				switch typed := decl.(type) {
				case StructType:
					return typed
				case InterfaceType:
					return typed
				case UnionType:
					return typed
				}
			}
		}
		if info.isUnion {
			return UnionType{UnionName: info.name}
		}
		return StructType{StructName: info.name}
	}
	if name, ok := unionName(self); ok && name != "" {
		if c.global != nil {
			if decl, ok := c.global.Lookup(name); ok {
				if union, ok := decl.(UnionType); ok {
					return union
				}
			}
		}
		return UnionType{UnionName: name}
	}
	return self
}
