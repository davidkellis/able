package interpreter

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type implEntry struct {
	interfaceName string
	methods       map[string]runtime.Value
	definition    *ast.ImplementationDefinition
	argTemplates  []ast.TypeExpression
	genericParams []*ast.GenericParameter
	whereClause   []*ast.WhereClauseConstraint
	unionVariants []string
	defaultOnly   bool
}

type implCandidate struct {
	entry       *implEntry
	bindings    map[string]ast.TypeExpression
	constraints []constraintSpec
	score       int
}

type methodMatch struct {
	candidate implCandidate
	method    runtime.Value
}

type constraintSpec struct {
	typeParam string
	ifaceType ast.TypeExpression
}

type typeInfo struct {
	name     string
	typeArgs []ast.TypeExpression
}

type targetVariant struct {
	typeName     string
	argTemplates []ast.TypeExpression
	signature    string
}

func expandImplementationTargetVariants(target ast.TypeExpression) ([]targetVariant, []string, error) {
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
			childVariants, childSigs, err := expandImplementationTargetVariants(member)
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
		// Build union signature list for penalty/book-keeping
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

func constraintSignature(specs []constraintSpec) string {
	if len(specs) == 0 {
		return "<none>"
	}
	parts := make([]string, 0, len(specs))
	for _, spec := range specs {
		parts = append(parts, fmt.Sprintf("%s->%s", spec.typeParam, typeExpressionToString(spec.ifaceType)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

func (i *Interpreter) registerUnnamedImpl(ifaceName string, ifaceArgs []ast.TypeExpression, variant targetVariant, unionSignatures []string, baseConstraintSig string, targetDescription string) error {
	key := ifaceName + "::" + variant.typeName
	bucket, ok := i.unnamedImpls[key]
	if !ok {
		bucket = make(map[string]map[string]struct{})
		i.unnamedImpls[key] = bucket
	}
	ifaceKey := "<none>"
	if len(ifaceArgs) > 0 {
		parts := make([]string, 0, len(ifaceArgs))
		for _, arg := range ifaceArgs {
			parts = append(parts, typeExpressionToString(arg))
		}
		ifaceKey = strings.Join(parts, "|")
	}
	templateKey := "<none>"
	if len(variant.argTemplates) > 0 {
		parts := make([]string, 0, len(variant.argTemplates))
		for _, tmpl := range variant.argTemplates {
			parts = append(parts, typeExpressionToString(tmpl))
		}
		templateKey = strings.Join(parts, "|")
	}
	if ifaceKey != "" {
		templateKey = ifaceKey + "::" + templateKey
	}
	if len(unionSignatures) > 0 {
		prefix := strings.Join(unionSignatures, "::")
		templateKey = prefix + "::" + templateKey
	}
	constraintKey := baseConstraintSig
	if ifaceKey != "" && ifaceKey != "<none>" {
		constraintKey = ifaceKey + "::" + constraintKey
	}
	if len(unionSignatures) > 0 {
		prefix := strings.Join(unionSignatures, "::")
		constraintKey = prefix + "::" + constraintKey
	}
	constraintSet, ok := bucket[templateKey]
	if !ok {
		constraintSet = make(map[string]struct{})
		bucket[templateKey] = constraintSet
	}
	if _, exists := constraintSet[constraintKey]; exists {
		return fmt.Errorf("Unnamed impl for (%s, %s) already exists", ifaceName, targetDescription)
	}
	constraintSet[constraintKey] = struct{}{}
	return nil
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

func computeImplSpecificity(entry *implEntry, bindings map[string]ast.TypeExpression, constraints []constraintSpec) int {
	genericNames := genericNameSet(entry.genericParams)
	concreteScore := 0
	for _, tmpl := range entry.argTemplates {
		concreteScore += measureTemplateSpecificity(tmpl, genericNames)
	}
	constraintScore := len(constraints)
	bindingScore := len(bindings)
	unionPenalty := len(entry.unionVariants)
	defaultPenalty := 0
	if entry.defaultOnly {
		defaultPenalty = 1
	}
	return concreteScore*100 + constraintScore*10 + bindingScore - unionPenalty - defaultPenalty
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
		other, ok := actual.(*ast.GenericTypeExpression)
		if !ok {
			return false
		}
		if !matchTypeExpressionTemplate(t.Base, other.Base, genericNames, bindings) {
			return false
		}
		if len(t.Arguments) != len(other.Arguments) {
			return false
		}
		for idx := range t.Arguments {
			if !matchTypeExpressionTemplate(t.Arguments[idx], other.Arguments[idx], genericNames, bindings) {
				return false
			}
		}
		return true
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

func (i *Interpreter) matchImplEntry(entry *implEntry, info typeInfo) (map[string]ast.TypeExpression, bool) {
	if entry == nil {
		return nil, false
	}
	bindings := make(map[string]ast.TypeExpression)
	genericNames := genericNameSet(entry.genericParams)
	if len(entry.argTemplates) > 0 {
		if len(info.typeArgs) != len(entry.argTemplates) {
			return nil, false
		}
		for idx, tmpl := range entry.argTemplates {
			if !matchTypeExpressionTemplate(tmpl, info.typeArgs[idx], genericNames, bindings) {
				return nil, false
			}
		}
	}
	for _, gp := range entry.genericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		if _, ok := bindings[gp.Name.Name]; !ok {
			return nil, false
		}
	}
	return bindings, true
}

func (i *Interpreter) collectImplCandidates(info typeInfo, interfaceFilter string, methodFilter string) ([]implCandidate, error) {
	if info.name == "" {
		return nil, nil
	}
	entries := i.implMethods[info.name]
	matches := make([]implCandidate, 0)
	var constraintErr error
	for idx := range entries {
		entry := &entries[idx]
		if interfaceFilter != "" && entry.interfaceName != interfaceFilter {
			continue
		}
		if methodFilter != "" && !i.implProvidesMethod(entry, methodFilter) {
			continue
		}
		bindings, ok := i.matchImplEntry(entry, info)
		if !ok {
			continue
		}
		constraints := collectConstraintSpecs(entry.genericParams, entry.whereClause)
		if len(constraints) > 0 {
			if err := i.enforceConstraintSpecs(constraints, bindings); err != nil {
				if constraintErr == nil {
					constraintErr = err
				}
				continue
			}
		}
		score := computeImplSpecificity(entry, bindings, constraints)
		matches = append(matches, implCandidate{
			entry:       entry,
			bindings:    bindings,
			constraints: constraints,
			score:       score,
		})
	}
	if len(matches) == 0 {
		return nil, constraintErr
	}
	return matches, nil
}

func (i *Interpreter) implProvidesMethod(entry *implEntry, methodName string) bool {
	if entry == nil || methodName == "" {
		return true
	}
	if entry.methods != nil {
		if method := entry.methods[methodName]; method != nil {
			return true
		}
	}
	ifaceDef, ok := i.interfaces[entry.interfaceName]
	if !ok || ifaceDef == nil || ifaceDef.Node == nil {
		return false
	}
	for _, sig := range ifaceDef.Node.Signatures {
		if sig == nil || sig.Name == nil || sig.Name.Name != methodName {
			continue
		}
		return sig.DefaultImpl != nil
	}
	return false
}

func (i *Interpreter) compareMethodMatches(a, b implCandidate) int {
	if a.score > b.score {
		return 1
	}
	if a.score < b.score {
		return -1
	}
	aUnion := a.entry.unionVariants
	bUnion := b.entry.unionVariants
	if len(aUnion) > 0 && len(bUnion) == 0 {
		return -1
	}
	if len(aUnion) == 0 && len(bUnion) > 0 {
		return 1
	}
	if len(aUnion) > 0 && len(bUnion) > 0 {
		if isProperSubset(aUnion, bUnion) {
			return 1
		}
		if isProperSubset(bUnion, aUnion) {
			return -1
		}
		if len(aUnion) != len(bUnion) {
			if len(aUnion) < len(bUnion) {
				return 1
			}
			return -1
		}
	}
	aConstraints := i.buildConstraintKeySet(a.constraints)
	bConstraints := i.buildConstraintKeySet(b.constraints)
	if isConstraintSuperset(aConstraints, bConstraints) {
		return 1
	}
	if isConstraintSuperset(bConstraints, aConstraints) {
		return -1
	}
	return 0
}

func (i *Interpreter) buildConstraintKeySet(constraints []constraintSpec) map[string]struct{} {
	set := make(map[string]struct{})
	for _, c := range constraints {
		if c.ifaceType == nil {
			continue
		}
		expressions := i.collectInterfaceConstraintStrings(c.ifaceType, make(map[string]struct{}))
		for _, expr := range expressions {
			key := fmt.Sprintf("%s->%s", c.typeParam, expr)
			set[key] = struct{}{}
		}
	}
	return set
}

func (i *Interpreter) collectInterfaceConstraintStrings(typeExpr ast.TypeExpression, memo map[string]struct{}) []string {
	if typeExpr == nil {
		return nil
	}
	key := typeExpressionToString(typeExpr)
	if _, seen := memo[key]; seen {
		return nil
	}
	memo[key] = struct{}{}
	results := []string{key}
	if simple, ok := typeExpr.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
		if iface, exists := i.interfaces[simple.Name.Name]; exists && iface.Node != nil {
			for _, base := range iface.Node.BaseInterfaces {
				results = append(results, i.collectInterfaceConstraintStrings(base, memo)...)
			}
		}
	}
	return results
}

func isConstraintSuperset(a, b map[string]struct{}) bool {
	if len(a) <= len(b) {
		return false
	}
	for key := range b {
		if _, ok := a[key]; !ok {
			return false
		}
	}
	return true
}

func isProperSubset(a, b []string) bool {
	if len(a) == 0 {
		return len(b) > 0
	}
	setA := make(map[string]struct{}, len(a))
	for _, val := range a {
		setA[val] = struct{}{}
	}
	setB := make(map[string]struct{}, len(b))
	for _, val := range b {
		setB[val] = struct{}{}
	}
	if len(setA) >= len(setB) {
		return false
	}
	for val := range setA {
		if _, ok := setB[val]; !ok {
			return false
		}
	}
	return true
}

func (i *Interpreter) selectBestMethodCandidate(matches []methodMatch) (*methodMatch, []methodMatch) {
	if len(matches) == 0 {
		return nil, nil
	}
	bestIdx := 0
	contenders := []int{0}
	for idx := 1; idx < len(matches); idx++ {
		cmp := i.compareMethodMatches(matches[idx].candidate, matches[bestIdx].candidate)
		if cmp > 0 {
			bestIdx = idx
			contenders = []int{idx}
		} else if cmp == 0 {
			reverse := i.compareMethodMatches(matches[bestIdx].candidate, matches[idx].candidate)
			if reverse < 0 {
				bestIdx = idx
				contenders = []int{idx}
			} else if reverse == 0 {
				contenders = append(contenders, idx)
			}
		}
	}
	if len(contenders) > 1 {
		ambiguous := make([]methodMatch, 0, len(contenders))
		for _, idx := range contenders {
			ambiguous = append(ambiguous, matches[idx])
		}
		return nil, ambiguous
	}
	return &matches[bestIdx], nil
}

func (i *Interpreter) selectBestCandidate(matches []implCandidate) (*implCandidate, []implCandidate) {
	if len(matches) == 0 {
		return nil, nil
	}
	bestIdx := 0
	contenders := []int{0}
	for idx := 1; idx < len(matches); idx++ {
		cmp := i.compareMethodMatches(matches[idx], matches[bestIdx])
		if cmp > 0 {
			bestIdx = idx
			contenders = []int{idx}
		} else if cmp == 0 {
			reverse := i.compareMethodMatches(matches[bestIdx], matches[idx])
			if reverse < 0 {
				bestIdx = idx
				contenders = []int{idx}
			} else if reverse == 0 {
				contenders = append(contenders, idx)
			}
		}
	}
	if len(contenders) > 1 {
		ambiguous := make([]implCandidate, 0, len(contenders))
		for _, idx := range contenders {
			ambiguous = append(ambiguous, matches[idx])
		}
		return nil, ambiguous
	}
	return &matches[bestIdx], nil
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

func descriptionsFromMethodMatches(matches []methodMatch) []string {
	set := make(map[string]struct{})
	for _, match := range matches {
		if match.candidate.entry == nil || match.candidate.entry.definition == nil {
			continue
		}
		desc := typeExpressionToString(match.candidate.entry.definition.TargetType)
		set[desc] = struct{}{}
	}
	return sortedKeys(set)
}

func descriptionsFromCandidates(matches []implCandidate) []string {
	set := make(map[string]struct{})
	for _, match := range matches {
		if match.entry == nil || match.entry.definition == nil {
			continue
		}
		desc := typeExpressionToString(match.entry.definition.TargetType)
		set[desc] = struct{}{}
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
	if bucket, ok := i.inherentMethods[info.name]; ok {
		if _, exists := bucket[methodName]; exists {
			return true
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
