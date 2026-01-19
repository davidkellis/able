package interpreter

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type typeInfo struct {
	name     string
	typeArgs []ast.TypeExpression
}

type targetVariant struct {
	typeName     string
	argTemplates []ast.TypeExpression
	signature    string
}

func expandImplementationTargetVariants(target ast.TypeExpression, aliases map[string]*ast.TypeAliasDefinition) ([]targetVariant, []string, error) {
	target = expandTypeAliases(target, aliases, nil)
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return nil, nil, fmt.Errorf("Implementation target requires identifier")
		}
		signature := typeExpressionToString(t)
		return []targetVariant{{typeName: t.Name.Name, argTemplates: nil, signature: signature}}, nil, nil
	case *ast.GenericTypeExpression:
		simple, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || simple.Name == nil {
			return nil, nil, fmt.Errorf("Implementation target requires simple base type")
		}
		signature := typeExpressionToString(t)
		return []targetVariant{{
			typeName:     simple.Name.Name,
			argTemplates: append([]ast.TypeExpression(nil), t.Arguments...),
			signature:    signature,
		}}, nil, nil
	case *ast.UnionTypeExpression:
		var variants []targetVariant
		signatureSet := make(map[string]struct{})
		for _, member := range t.Members {
			childVariants, childSigs, err := expandImplementationTargetVariants(member, aliases)
			if err != nil {
				return nil, nil, err
			}
			for _, v := range childVariants {
				if _, seen := signatureSet[v.signature]; seen {
					continue
				}
				signatureSet[v.signature] = struct{}{}
				variants = append(variants, v)
			}
			for _, sig := range childSigs {
				signatureSet[sig] = struct{}{}
			}
		}
		if len(variants) == 0 {
			return nil, nil, fmt.Errorf("Union target must contain at least one concrete type")
		}
		unionSigs := make([]string, 0, len(signatureSet))
		for sig := range signatureSet {
			unionSigs = append(unionSigs, sig)
		}
		sort.Strings(unionSigs)
		return variants, unionSigs, nil
	default:
		return nil, nil, fmt.Errorf("Implementation target type %T is not supported", target)
	}
}

func collectConstraintSpecs(generics []*ast.GenericParameter, whereClause []*ast.WhereClauseConstraint) []constraintSpec {
	var specs []constraintSpec
	for _, gp := range generics {
		if gp == nil || gp.Name == nil {
			continue
		}
		for _, constraint := range gp.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			specs = append(specs, constraintSpec{typeParam: gp.Name.Name, ifaceType: constraint.InterfaceType})
		}
	}
	for _, clause := range whereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		for _, constraint := range clause.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			specs = append(specs, constraintSpec{typeParam: clause.TypeParam.Name, ifaceType: constraint.InterfaceType})
		}
	}
	return specs
}

func constraintSignature(specs []constraintSpec, stringify func(ast.TypeExpression) string) string {
	if len(specs) == 0 {
		return "<none>"
	}
	parts := make([]string, 0, len(specs))
	for _, spec := range specs {
		parts = append(parts, fmt.Sprintf("%s->%s", spec.typeParam, stringify(spec.ifaceType)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

func genericNameSet(params []*ast.GenericParameter) map[string]struct{} {
	set := make(map[string]struct{})
	for _, gp := range params {
		if gp == nil || gp.Name == nil {
			continue
		}
		set[gp.Name.Name] = struct{}{}
	}
	return set
}

func measureTemplateSpecificity(expr ast.TypeExpression, genericNames map[string]struct{}) int {
	if expr == nil {
		return 0
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return 0
		}
		if _, ok := genericNames[t.Name.Name]; ok {
			return 0
		}
		return 1
	case *ast.GenericTypeExpression:
		score := measureTemplateSpecificity(t.Base, genericNames)
		for _, arg := range t.Arguments {
			score += measureTemplateSpecificity(arg, genericNames)
		}
		return score
	case *ast.NullableTypeExpression:
		return measureTemplateSpecificity(t.InnerType, genericNames)
	case *ast.ResultTypeExpression:
		return measureTemplateSpecificity(t.InnerType, genericNames)
	case *ast.UnionTypeExpression:
		score := 0
		for _, member := range t.Members {
			score += measureTemplateSpecificity(member, genericNames)
		}
		return score
	default:
		return 0
	}
}

func collectImplGenericNames(entry *implEntry) map[string]struct{} {
	names := genericNameSet(entry.genericParams)
	if entry == nil || entry.definition == nil {
		return names
	}
	var consider func(ast.TypeExpression)
	consider = func(expr ast.TypeExpression) {
		switch val := expr.(type) {
		case *ast.SimpleTypeExpression:
			if val.Name != nil && len(val.Name.Name) == 1 && val.Name.Name[0] >= 'A' && val.Name.Name[0] <= 'Z' {
				names[val.Name.Name] = struct{}{}
			}
		case *ast.GenericTypeExpression:
			consider(val.Base)
			for _, arg := range val.Arguments {
				if arg == nil {
					continue
				}
				consider(arg)
			}
		case *ast.NullableTypeExpression:
			consider(val.InnerType)
		case *ast.ResultTypeExpression:
			consider(val.InnerType)
		case *ast.UnionTypeExpression:
			for _, member := range val.Members {
				if member == nil {
					continue
				}
				consider(member)
			}
		}
	}
	for _, ifaceArg := range entry.definition.InterfaceArgs {
		if ifaceArg != nil {
			consider(ifaceArg)
		}
	}
	for _, tmpl := range entry.argTemplates {
		if tmpl != nil {
			consider(tmpl)
		}
	}
	return names
}

func typeExpressionUsesGenerics(expr ast.TypeExpression, genericNames map[string]struct{}) bool {
	switch val := expr.(type) {
	case nil:
		return false
	case *ast.SimpleTypeExpression:
		if val.Name == nil {
			return false
		}
		_, ok := genericNames[val.Name.Name]
		return ok
	case *ast.GenericTypeExpression:
		if typeExpressionUsesGenerics(val.Base, genericNames) {
			return true
		}
		for _, arg := range val.Arguments {
			if typeExpressionUsesGenerics(arg, genericNames) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return typeExpressionUsesGenerics(val.InnerType, genericNames)
	case *ast.ResultTypeExpression:
		return typeExpressionUsesGenerics(val.InnerType, genericNames)
	case *ast.UnionTypeExpression:
		for _, member := range val.Members {
			if typeExpressionUsesGenerics(member, genericNames) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if typeExpressionUsesGenerics(val.ReturnType, genericNames) {
			return true
		}
		for _, param := range val.ParamTypes {
			if typeExpressionUsesGenerics(param, genericNames) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func typeExpressionsEqual(a, b ast.TypeExpression) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch ta := a.(type) {
	case *ast.SimpleTypeExpression:
		tb, ok := b.(*ast.SimpleTypeExpression)
		if !ok {
			return false
		}
		if ta.Name == nil || tb.Name == nil {
			return ta.Name == nil && tb.Name == nil
		}
		return ta.Name.Name == tb.Name.Name
	case *ast.GenericTypeExpression:
		tb, ok := b.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if !typeExpressionsEqual(ta.Base, tb.Base) {
			return false
		}
		if len(ta.Arguments) != len(tb.Arguments) {
			return false
		}
		for idx := range ta.Arguments {
			if !typeExpressionsEqual(ta.Arguments[idx], tb.Arguments[idx]) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		tb, ok := b.(*ast.NullableTypeExpression)
		if !ok {
			return false
		}
		return typeExpressionsEqual(ta.InnerType, tb.InnerType)
	case *ast.ResultTypeExpression:
		tb, ok := b.(*ast.ResultTypeExpression)
		if !ok {
			return false
		}
		return typeExpressionsEqual(ta.InnerType, tb.InnerType)
	case *ast.FunctionTypeExpression:
		tb, ok := b.(*ast.FunctionTypeExpression)
		if !ok {
			return false
		}
		if len(ta.ParamTypes) != len(tb.ParamTypes) {
			return false
		}
		for idx := range ta.ParamTypes {
			if !typeExpressionsEqual(ta.ParamTypes[idx], tb.ParamTypes[idx]) {
				return false
			}
		}
		return typeExpressionsEqual(ta.ReturnType, tb.ReturnType)
	case *ast.UnionTypeExpression:
		tb, ok := b.(*ast.UnionTypeExpression)
		if !ok || len(ta.Members) != len(tb.Members) {
			return false
		}
		for idx := range ta.Members {
			if !typeExpressionsEqual(ta.Members[idx], tb.Members[idx]) {
				return false
			}
		}
		return true
	case *ast.WildcardTypeExpression:
		_, ok := b.(*ast.WildcardTypeExpression)
		return ok
	default:
		return false
	}
}

func matchTypeExpressionTemplate(template, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	switch t := template.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return actual == nil
		}
		name := t.Name.Name
		if _, isGeneric := genericNames[name]; isGeneric {
			if existing, ok := bindings[name]; ok {
				return typeExpressionsEqual(existing, actual)
			}
			bindings[name] = actual
			return true
		}
		return typeExpressionsEqual(template, actual)
	case *ast.GenericTypeExpression:
		switch other := actual.(type) {
		case *ast.GenericTypeExpression:
			if !matchTypeExpressionTemplate(t.Base, other.Base, genericNames, bindings) {
				return false
			}
			otherArgs := other.Arguments
			if len(otherArgs) == 0 && len(t.Arguments) > 0 {
				otherArgs = make([]ast.TypeExpression, len(t.Arguments))
				for idx := range otherArgs {
					otherArgs[idx] = ast.NewWildcardTypeExpression()
				}
			}
			if len(t.Arguments) != len(otherArgs) {
				return false
			}
			for idx := range t.Arguments {
				if !matchTypeExpressionTemplate(t.Arguments[idx], otherArgs[idx], genericNames, bindings) {
					return false
				}
			}
			return true
		case *ast.SimpleTypeExpression:
			if !matchTypeExpressionTemplate(t.Base, other, genericNames, bindings) {
				return false
			}
			for idx := range t.Arguments {
				if !matchTypeExpressionTemplate(t.Arguments[idx], ast.NewWildcardTypeExpression(), genericNames, bindings) {
					return false
				}
			}
			return true
		default:
			return false
		}
	case *ast.NullableTypeExpression:
		other, ok := actual.(*ast.NullableTypeExpression)
		if !ok {
			return false
		}
		return matchTypeExpressionTemplate(t.InnerType, other.InnerType, genericNames, bindings)
	case *ast.ResultTypeExpression:
		other, ok := actual.(*ast.ResultTypeExpression)
		if !ok {
			return false
		}
		return matchTypeExpressionTemplate(t.InnerType, other.InnerType, genericNames, bindings)
	case *ast.UnionTypeExpression:
		other, ok := actual.(*ast.UnionTypeExpression)
		if !ok || len(t.Members) != len(other.Members) {
			return false
		}
		for idx := range t.Members {
			if !matchTypeExpressionTemplate(t.Members[idx], other.Members[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	default:
		return typeExpressionsEqual(template, actual)
	}
}

func mapTypeArguments(generics []*ast.GenericParameter, provided []ast.TypeExpression, context string) (map[string]ast.TypeExpression, error) {
	result := make(map[string]ast.TypeExpression)
	if len(generics) == 0 {
		return result, nil
	}
	if len(provided) != len(generics) {
		return nil, fmt.Errorf("Type arguments count mismatch %s: expected %d, got %d", context, len(generics), len(provided))
	}
	for idx, gp := range generics {
		if gp == nil || gp.Name == nil {
			continue
		}
		ta := provided[idx]
		if ta == nil {
			return nil, fmt.Errorf("Missing type argument for '%s' required by %s", gp.Name.Name, context)
		}
		result[gp.Name.Name] = ta
	}
	return result, nil
}

func (i *Interpreter) enforceConstraintSpecs(constraints []constraintSpec, typeArgMap map[string]ast.TypeExpression) error {
	for _, spec := range constraints {
		actual, ok := typeArgMap[spec.typeParam]
		if !ok {
			return fmt.Errorf("Missing type argument for '%s' required by constraints", spec.typeParam)
		}
		tInfo, ok := parseTypeExpression(actual)
		if !ok {
			continue
		}
		if err := i.ensureTypeSatisfiesInterface(tInfo, spec.ifaceType, spec.typeParam, make(map[string]struct{})); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) ensureTypeSatisfiesInterface(tInfo typeInfo, ifaceExpr ast.TypeExpression, context string, visited map[string]struct{}) error {
	ifaceInfo, ok := parseTypeExpression(ifaceExpr)
	if !ok {
		return nil
	}
	if _, seen := visited[ifaceInfo.name]; seen {
		return nil
	}
	visited[ifaceInfo.name] = struct{}{}
	ifaceDef, ok := i.interfaces[ifaceInfo.name]
	if !ok {
		return fmt.Errorf("Unknown interface '%s' in constraint on '%s'", ifaceInfo.name, context)
	}
	if ifaceDef.Node != nil {
		for _, base := range ifaceDef.Node.BaseInterfaces {
			if err := i.ensureTypeSatisfiesInterface(tInfo, base, context, visited); err != nil {
				return err
			}
		}
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			methodName := sig.Name.Name
			if !i.typeHasMethod(tInfo, methodName, ifaceInfo.name) {
				return fmt.Errorf("Type '%s' does not satisfy interface '%s': missing method '%s'", tInfo.name, ifaceInfo.name, methodName)
			}
		}
	}
	return nil
}

func (i *Interpreter) typeHasMethod(info typeInfo, methodName, ifaceName string) bool {
	if info.name == "" {
		return false
	}
	if primitiveImplementsInterfaceMethod(info.name, ifaceName, methodName) {
		return true
	}
	for _, name := range i.canonicalTypeNames(info.name) {
		if bucket, ok := i.inherentMethods[name]; ok {
			if _, exists := bucket[methodName]; exists {
				return true
			}
		}
	}
	method, err := i.findMethod(info, methodName, ifaceName)
	return err == nil && method != nil
}

func parseTypeExpression(expr ast.TypeExpression) (typeInfo, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return typeInfo{}, false
		}
		return typeInfo{name: t.Name.Name, typeArgs: nil}, true
	case *ast.GenericTypeExpression:
		tInfo, ok := parseTypeExpression(t.Base)
		if !ok {
			return typeInfo{}, false
		}
		tInfo.typeArgs = append([]ast.TypeExpression(nil), t.Arguments...)
		return tInfo, true
	default:
		return typeInfo{}, false
	}
}

func typeExpressionToString(expr ast.TypeExpression) string {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return "<?>"
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		base := typeExpressionToString(t.Base)
		args := make([]string, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			args = append(args, typeExpressionToString(arg))
		}
		return fmt.Sprintf("%s<%s>", base, strings.Join(args, ", "))
	case *ast.NullableTypeExpression:
		return typeExpressionToString(t.InnerType) + "?"
	case *ast.FunctionTypeExpression:
		parts := make([]string, 0, len(t.ParamTypes))
		for _, p := range t.ParamTypes {
			parts = append(parts, typeExpressionToString(p))
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(parts, ", "), typeExpressionToString(t.ReturnType))
	case *ast.UnionTypeExpression:
		parts := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			parts = append(parts, typeExpressionToString(member))
		}
		return strings.Join(parts, " | ")
	default:
		return "<?>"
	}
}

func (i *Interpreter) typeInfoFromStructInstance(inst *runtime.StructInstanceValue) (typeInfo, bool) {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return typeInfo{}, false
	}
	info := typeInfo{name: inst.Definition.Node.ID.Name}
	if len(inst.TypeArguments) > 0 {
		info.typeArgs = append([]ast.TypeExpression(nil), inst.TypeArguments...)
	}
	return info, true
}

func typeInfoToString(info typeInfo) string {
	if info.name == "" {
		return "<unknown>"
	}
	if len(info.typeArgs) == 0 {
		return info.name
	}
	parts := make([]string, 0, len(info.typeArgs))
	for _, arg := range info.typeArgs {
		parts = append(parts, typeExpressionToString(arg))
	}
	return fmt.Sprintf("%s<%s>", info.name, strings.Join(parts, ", "))
}

func typeExpressionFromInfo(info typeInfo) ast.TypeExpression {
	if info.name == "" {
		return nil
	}
	base := ast.Ty(info.name)
	if len(info.typeArgs) == 0 {
		return base
	}
	return ast.Gen(base, info.typeArgs...)
}
