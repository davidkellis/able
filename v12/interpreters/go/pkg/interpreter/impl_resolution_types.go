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

var integerTypes = map[string]struct{}{
	"i8": {}, "i16": {}, "i32": {}, "i64": {}, "i128": {},
	"u8": {}, "u16": {}, "u32": {}, "u64": {}, "u128": {},
	"isize": {}, "usize": {},
}

var floatTypes = map[string]struct{}{"f32": {}, "f64": {}}

func isPrimitiveTypeName(name string) bool {
	switch name {
	case "bool", "String", "string", "IoHandle", "ProcHandle", "char", "nil", "void":
		return true
	}
	if _, ok := integerTypes[name]; ok {
		return true
	}
	if _, ok := floatTypes[name]; ok {
		return true
	}
	return false
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
			specs = append(specs, constraintSpec{subject: ast.NewSimpleTypeExpression(gp.Name), ifaceType: constraint.InterfaceType})
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
			specs = append(specs, constraintSpec{subject: clause.TypeParam, ifaceType: constraint.InterfaceType})
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
		parts = append(parts, fmt.Sprintf("%s->%s", stringify(spec.subject), stringify(spec.ifaceType)))
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
				if _, ok := existing.(*ast.WildcardTypeExpression); ok {
					if actual != nil {
						if _, ok := actual.(*ast.WildcardTypeExpression); !ok {
							bindings[name] = actual
						}
					}
					return true
				}
				if _, ok := actual.(*ast.WildcardTypeExpression); ok {
					return true
				}
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
				for idx := range t.Arguments {
					if !matchTypeExpressionTemplate(t.Arguments[idx], cachedWildcardTypeExpression, genericNames, bindings) {
						return false
					}
				}
				return true
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
				if !matchTypeExpressionTemplate(t.Arguments[idx], cachedWildcardTypeExpression, genericNames, bindings) {
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
			return matchTypeExpressionTemplate(t.InnerType, other.InnerType, genericNames, bindings)
		}
		return matchTypeExpressionTemplate(t.InnerType, actual, genericNames, bindings)
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
	return mapTypeArgumentsByNames(buildGenericParamNamesByIndex(generics), len(generics), provided, context)
}

type indexedTypeArgumentLookup struct {
	namesByIndex  []string
	expectedCount int
	provided      []ast.TypeExpression
}

func (l indexedTypeArgumentLookup) Lookup(name string) (ast.TypeExpression, bool) {
	limit := l.expectedCount
	if len(l.namesByIndex) < limit {
		limit = len(l.namesByIndex)
	}
	if len(l.provided) < limit {
		limit = len(l.provided)
	}
	for idx := 0; idx < limit; idx++ {
		if l.namesByIndex[idx] != name {
			continue
		}
		value := l.provided[idx]
		if value == nil {
			return nil, false
		}
		return value, true
	}
	return nil, false
}

func validateTypeArgumentsByNames(namesByIndex []string, expectedCount int, provided []ast.TypeExpression, context string) error {
	if expectedCount == 0 {
		return nil
	}
	if len(provided) != expectedCount {
		return fmt.Errorf("Type arguments count mismatch %s: expected %d, got %d", context, expectedCount, len(provided))
	}
	for idx := 0; idx < expectedCount; idx++ {
		name := namesByIndex[idx]
		if name == "" {
			continue
		}
		if provided[idx] == nil {
			return fmt.Errorf("Missing type argument for '%s' required by %s", name, context)
		}
	}
	return nil
}

func indexedTypeArgumentsByNames(namesByIndex []string, expectedCount int, provided []ast.TypeExpression, context string) (indexedTypeArgumentLookup, error) {
	if err := validateTypeArgumentsByNames(namesByIndex, expectedCount, provided, context); err != nil {
		return indexedTypeArgumentLookup{}, err
	}
	return indexedTypeArgumentLookup{
		namesByIndex:  namesByIndex,
		expectedCount: expectedCount,
		provided:      provided,
	}, nil
}

func mapTypeArgumentsByNames(namesByIndex []string, expectedCount int, provided []ast.TypeExpression, context string) (map[string]ast.TypeExpression, error) {
	if err := validateTypeArgumentsByNames(namesByIndex, expectedCount, provided, context); err != nil {
		return nil, err
	}
	result := make(map[string]ast.TypeExpression, expectedCount)
	if expectedCount == 0 {
		return result, nil
	}
	for idx := 0; idx < expectedCount; idx++ {
		name := namesByIndex[idx]
		if name == "" {
			continue
		}
		ta := provided[idx]
		result[name] = ta
	}
	return result, nil
}

func (i *Interpreter) enforceConstraintSpecs(constraints []constraintSpec, typeArgMap map[string]ast.TypeExpression) error {
	for _, spec := range constraints {
		subject := substituteTypeParamsWithMap(spec.subject, typeArgMap)
		if err := i.enforceConstraintSpecSubject(spec, subject); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) enforceConstraintSpecsWithTypeArgs(constraints []constraintSpec, namesByIndex []string, expectedCount int, provided []ast.TypeExpression, context string) error {
	lookup, err := indexedTypeArgumentsByNames(namesByIndex, expectedCount, provided, context)
	if err != nil {
		return err
	}
	for _, spec := range constraints {
		subject := substituteTypeParamsWithIndexedArgs(spec.subject, lookup)
		if err := i.enforceConstraintSpecSubject(spec, subject); err != nil {
			return err
		}
	}
	return nil
}

func substituteTypeParamsWithMap(expr ast.TypeExpression, typeArgMap map[string]ast.TypeExpression) ast.TypeExpression {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			if replacement, ok := typeArgMap[t.Name.Name]; ok {
				return replacement
			}
		}
		return t
	case *ast.GenericTypeExpression:
		base := substituteTypeParamsWithMap(t.Base, typeArgMap)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = substituteTypeParamsWithMap(arg, typeArgMap)
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = substituteTypeParamsWithMap(param, typeArgMap)
		}
		return ast.NewFunctionTypeExpression(params, substituteTypeParamsWithMap(t.ReturnType, typeArgMap))
	case *ast.NullableTypeExpression:
		return ast.NewNullableTypeExpression(substituteTypeParamsWithMap(t.InnerType, typeArgMap))
	case *ast.ResultTypeExpression:
		return ast.NewResultTypeExpression(substituteTypeParamsWithMap(t.InnerType, typeArgMap))
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = substituteTypeParamsWithMap(member, typeArgMap)
		}
		return ast.NewUnionTypeExpression(members)
	default:
		return expr
	}
}

func substituteTypeParamsWithIndexedArgs(expr ast.TypeExpression, lookup indexedTypeArgumentLookup) ast.TypeExpression {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			if replacement, ok := lookup.Lookup(t.Name.Name); ok {
				return replacement
			}
		}
		return t
	case *ast.GenericTypeExpression:
		base := substituteTypeParamsWithIndexedArgs(t.Base, lookup)
		args := make([]ast.TypeExpression, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = substituteTypeParamsWithIndexedArgs(arg, lookup)
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.FunctionTypeExpression:
		params := make([]ast.TypeExpression, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = substituteTypeParamsWithIndexedArgs(param, lookup)
		}
		return ast.NewFunctionTypeExpression(params, substituteTypeParamsWithIndexedArgs(t.ReturnType, lookup))
	case *ast.NullableTypeExpression:
		return ast.NewNullableTypeExpression(substituteTypeParamsWithIndexedArgs(t.InnerType, lookup))
	case *ast.ResultTypeExpression:
		return ast.NewResultTypeExpression(substituteTypeParamsWithIndexedArgs(t.InnerType, lookup))
	case *ast.UnionTypeExpression:
		members := make([]ast.TypeExpression, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = substituteTypeParamsWithIndexedArgs(member, lookup)
		}
		return ast.NewUnionTypeExpression(members)
	default:
		return expr
	}
}

func (i *Interpreter) typeNameKnownForConstraint(name string) bool {
	if name == "" || name == "_" || name == "Self" {
		return false
	}
	if isPrimitiveTypeName(name) {
		return true
	}
	if _, ok := i.interfaces[name]; ok {
		return true
	}
	if _, ok := i.unionDefinitions[name]; ok {
		return true
	}
	if _, ok := i.typeAliases[name]; ok {
		return true
	}
	if _, ok := i.lookupStructDefinition(name); ok {
		return true
	}
	return false
}

func (i *Interpreter) constraintSubjectHasUnknownNames(expr ast.TypeExpression) bool {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return true
		}
		return !i.typeNameKnownForConstraint(t.Name.Name)
	case *ast.GenericTypeExpression:
		if i.constraintSubjectHasUnknownNames(t.Base) {
			return true
		}
		for _, arg := range t.Arguments {
			if i.constraintSubjectHasUnknownNames(arg) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if i.constraintSubjectHasUnknownNames(t.ReturnType) {
			return true
		}
		for _, param := range t.ParamTypes {
			if i.constraintSubjectHasUnknownNames(param) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return i.constraintSubjectHasUnknownNames(t.InnerType)
	case *ast.ResultTypeExpression:
		return i.constraintSubjectHasUnknownNames(t.InnerType)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if i.constraintSubjectHasUnknownNames(member) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func (i *Interpreter) enforceConstraintSpecSubject(spec constraintSpec, subject ast.TypeExpression) error {
	if i.constraintSubjectHasUnknownNames(subject) {
		return nil
	}
	tInfo, ok := parseTypeExpression(subject)
	if !ok {
		return nil
	}
	context := typeInfoToString(tInfo)
	return i.ensureTypeSatisfiesInterface(tInfo, spec.ifaceType, context, nil)
}

func (i *Interpreter) ensureTypeSatisfiesInterface(tInfo typeInfo, ifaceExpr ast.TypeExpression, context string, visited map[string]struct{}) error {
	ifaceInfo, ok := parseTypeExpression(ifaceExpr)
	if !ok {
		return nil
	}
	if visited != nil {
		if _, seen := visited[ifaceInfo.name]; seen {
			return nil
		}
		visited[ifaceInfo.name] = struct{}{}
	}
	ifaceDef, ok := i.interfaces[ifaceInfo.name]
	if !ok {
		// In compiled no-bootstrap mode, trust the compiled dispatch table
		if i.compiledImplChecker != nil && i.compiledImplChecker(tInfo.name, ifaceInfo.name) {
			return nil
		}
		return fmt.Errorf("Unknown interface '%s' in constraint on '%s'", ifaceInfo.name, context)
	}
	if ifaceDef.Node != nil {
		if visited == nil && len(ifaceDef.Node.BaseInterfaces) > 0 {
			visited = make(map[string]struct{}, 4)
			visited[ifaceInfo.name] = struct{}{}
		}
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
	method, err := i.findMethodCached(info, methodName, ifaceName)
	if err == nil && method != nil {
		return true
	}
	if i.compiledImplChecker != nil && ifaceName != "" && i.compiledImplChecker(info.name, ifaceName) {
		return true
	}
	return false
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
		// Type expression argument slices are treated as immutable by runtime
		// resolution paths; reusing them avoids per-parse copy churn.
		tInfo.typeArgs = t.Arguments
		return tInfo, true
	default:
		return typeInfo{}, false
	}
}

func typeExpressionToString(expr ast.TypeExpression) string {
	var b strings.Builder
	appendTypeExpressionString(&b, expr)
	return b.String()
}

func appendTypeExpressionString(b *strings.Builder, expr ast.TypeExpression) {
	if b == nil {
		return
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			b.WriteString("<?>")
			return
		}
		b.WriteString(t.Name.Name)
	case *ast.GenericTypeExpression:
		if t == nil {
			b.WriteString("<?>")
			return
		}
		appendTypeExpressionString(b, t.Base)
		b.WriteByte('<')
		for idx, arg := range t.Arguments {
			if idx > 0 {
				b.WriteString(", ")
			}
			appendTypeExpressionString(b, arg)
		}
		b.WriteByte('>')
	case *ast.NullableTypeExpression:
		if t == nil {
			b.WriteString("<?>")
			return
		}
		appendTypeExpressionString(b, t.InnerType)
		b.WriteByte('?')
	case *ast.ResultTypeExpression:
		if t == nil {
			b.WriteString("<?>")
			return
		}
		b.WriteByte('!')
		appendTypeExpressionString(b, t.InnerType)
	case *ast.FunctionTypeExpression:
		if t == nil {
			b.WriteString("<?>")
			return
		}
		b.WriteString("fn(")
		for idx, p := range t.ParamTypes {
			if idx > 0 {
				b.WriteString(", ")
			}
			appendTypeExpressionString(b, p)
		}
		b.WriteString(") -> ")
		appendTypeExpressionString(b, t.ReturnType)
	case *ast.UnionTypeExpression:
		if t == nil {
			b.WriteString("<?>")
			return
		}
		for idx, member := range t.Members {
			if idx > 0 {
				b.WriteString(" | ")
			}
			appendTypeExpressionString(b, member)
		}
	default:
		b.WriteString("<?>")
	}
}

func (i *Interpreter) typeInfoFromStructInstance(inst *runtime.StructInstanceValue) (typeInfo, bool) {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return typeInfo{}, false
	}
	node := inst.Definition.Node
	name := node.ID.Name
	if name == "Array" {
		if arr, err := i.arrayValueFromStructInstance(inst); err == nil && arr != nil {
			if typeExpr := i.typeExpressionFromValue(arr); typeExpr != nil {
				if info, ok := parseTypeExpression(typeExpr); ok {
					return info, true
				}
			}
		}
	}
	info := typeInfo{name: name}
	if len(inst.TypeArguments) == 0 && len(node.GenericParams) == 0 {
		return info, true
	}
	if structTypeArgsConcreteForDefinition(node, inst.TypeArguments) {
		info.typeArgs = inst.TypeArguments
		return info, true
	}
	if typeArgs := i.resolvedStructInstanceTypeArgumentsWithSeen(inst, nil); len(typeArgs) > 0 {
		info.typeArgs = typeArgs
	} else if len(inst.TypeArguments) > 0 {
		info.typeArgs = inst.TypeArguments
	}
	return info, true
}

func (i *Interpreter) resolvedStructInstanceTypeArgumentsWithSeen(inst *runtime.StructInstanceValue, seen map[*runtime.StructInstanceValue]struct{}) []ast.TypeExpression {
	return i.resolvedStructInstanceTypeArgumentsWithSeenMemo(inst, seen, true)
}

func (i *Interpreter) resolvedStructInstanceTypeArgumentsWithSeenMemo(inst *runtime.StructInstanceValue, seen map[*runtime.StructInstanceValue]struct{}, memoize bool) []ast.TypeExpression {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil {
		return nil
	}
	typeArgs := inst.TypeArguments
	if structTypeArgsConcreteForDefinition(inst.Definition.Node, typeArgs) {
		return typeArgs
	}
	plan := i.structGenericInferencePlan(inst.Definition.Node)
	if plan == nil || plan.expectedCount == 0 {
		return inst.TypeArguments
	}
	if !structTypeArgsNeedInference(plan, typeArgs) {
		return typeArgs
	}
	inferSeen := seen
	if inferSeen == nil {
		inferSeen = map[*runtime.StructInstanceValue]struct{}{inst: {}}
	} else if _, ok := inferSeen[inst]; !ok {
		inferSeen[inst] = struct{}{}
		defer delete(inferSeen, inst)
	}
	typeArgs = i.inferStructTypeArgumentsWithSeen(inst.Definition.Node, inst.Fields, inst.Positional, inferSeen)
	if len(typeArgs) > 0 && memoize {
		inst.TypeArguments = typeArgs
	}
	return typeArgs
}

func typeInfoToString(info typeInfo) string {
	if info.name == "" {
		return "<unknown>"
	}
	if len(info.typeArgs) == 0 {
		return info.name
	}
	var b strings.Builder
	b.WriteString(info.name)
	b.WriteByte('<')
	for idx, arg := range info.typeArgs {
		if idx > 0 {
			b.WriteString(", ")
		}
		appendTypeExpressionString(&b, arg)
	}
	b.WriteByte('>')
	return b.String()
}

func typeExpressionFromInfo(info typeInfo) ast.TypeExpression {
	if info.name == "" {
		return nil
	}
	base := cachedSimpleTypeExpression(info.name)
	if len(info.typeArgs) == 0 {
		return base
	}
	args := append([]ast.TypeExpression(nil), info.typeArgs...)
	return ast.NewGenericTypeExpression(base, args)
}
