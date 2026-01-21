package interpreter

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
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
	isBuiltin     bool
}

type implCandidate struct {
	entry       *implEntry
	bindings    map[string]ast.TypeExpression
	constraints []constraintSpec
	score       int
	isConcrete  bool
}

type methodMatch struct {
	candidate implCandidate
	method    runtime.Value
}

type constraintSpec struct {
	subject   ast.TypeExpression
	ifaceType ast.TypeExpression
}

func (i *Interpreter) registerUnnamedImpl(ifaceName string, ifaceArgs []ast.TypeExpression, variant targetVariant, unionSignatures []string, baseConstraintSig string, targetDescription string, isBuiltin bool) error {
	key := ifaceName + "::" + variant.typeName
	bucket, ok := i.unnamedImpls[key]
	if !ok {
		bucket = make(map[string]map[string]bool)
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
		constraintSet = make(map[string]bool)
		bucket[templateKey] = constraintSet
	}
	if existingBuiltin, exists := constraintSet[constraintKey]; exists {
		if existingBuiltin || isBuiltin {
			return nil
		}
		return fmt.Errorf("Unnamed impl for (%s, %s) already exists", ifaceName, targetDescription)
	}
	constraintSet[constraintKey] = isBuiltin
	return nil
}

func (i *Interpreter) matchImplEntry(entry *implEntry, info typeInfo) (map[string]ast.TypeExpression, bool) {
	if entry == nil {
		return nil, false
	}
	bindings := make(map[string]ast.TypeExpression)
	genericNames := collectImplGenericNames(entry)
	paramUsedInTarget := func(name string) bool {
		if name == "" {
			return false
		}
		lookup := map[string]struct{}{name: {}}
		if entry.definition != nil && typeExpressionUsesGenerics(entry.definition.TargetType, lookup) {
			return true
		}
		for _, tmpl := range entry.argTemplates {
			if typeExpressionUsesGenerics(tmpl, lookup) {
				return true
			}
		}
		return false
	}
	if entry.definition != nil {
		actual := typeExpressionFromInfo(info)
		if actual != nil {
			template := expandTypeAliases(entry.definition.TargetType, i.typeAliases, nil)
			matchTypeExpressionTemplate(template, expandTypeAliases(actual, i.typeAliases, nil), genericNames, bindings)
		}
	}
	if len(entry.argTemplates) > 0 {
		if len(info.typeArgs) != len(entry.argTemplates) {
			return nil, false
		}
		for idx, tmpl := range entry.argTemplates {
			if !matchTypeExpressionTemplate(expandTypeAliases(tmpl, i.typeAliases, nil), expandTypeAliases(info.typeArgs[idx], i.typeAliases, nil), genericNames, bindings) {
				return nil, false
			}
		}
	}
	for _, gp := range entry.genericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		if _, ok := bindings[gp.Name.Name]; !ok {
			if !paramUsedInTarget(gp.Name.Name) {
				continue
			}
			return nil, false
		}
	}
	return bindings, true
}

func (i *Interpreter) collectImplCandidates(info typeInfo, interfaceFilter string, methodFilter string) ([]implCandidate, error) {
	if info.name == "" {
		return nil, nil
	}
	typeNames := i.canonicalTypeNames(info.name)
	entries := make([]implEntry, 0)
	for _, name := range typeNames {
		if impls, ok := i.implMethods[name]; ok {
			entries = append(entries, impls...)
		}
	}
	if len(i.genericImpls) > 0 {
		entries = append(entries, i.genericImpls...)
	}
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
		entryInfo := info
		bindings, ok := i.matchImplEntry(entry, entryInfo)
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
		genericNames := collectImplGenericNames(entry)
		score := measureTemplateSpecificity(entry.definition.TargetType, genericNames)
		isConcrete := !typeExpressionUsesGenerics(entry.definition.TargetType, genericNames)
		matches = append(matches, implCandidate{
			entry:       entry,
			bindings:    bindings,
			constraints: constraints,
			score:       score,
			isConcrete:  isConcrete,
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
	if a.isConcrete && !b.isConcrete {
		return 1
	}
	if b.isConcrete && !a.isConcrete {
		return -1
	}
	aConstraints := i.buildConstraintKeySet(a.constraints)
	bConstraints := i.buildConstraintKeySet(b.constraints)
	if isConstraintSuperset(aConstraints, bConstraints) {
		return 1
	}
	if isConstraintSuperset(bConstraints, aConstraints) {
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
	if a.score > b.score {
		return 1
	}
	if a.score < b.score {
		return -1
	}
	if a.entry != nil && b.entry != nil && a.entry.isBuiltin != b.entry.isBuiltin {
		if a.entry.isBuiltin {
			return -1
		}
		return 1
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
			key := fmt.Sprintf("%s->%s", typeExpressionToString(c.subject), expr)
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
func descriptionsFromMethodMatches(matches []methodMatch) []string {
	set := make(map[string]struct{})
	for _, match := range matches {
		if match.candidate.entry == nil || match.candidate.entry.definition == nil {
			continue
		}
		target := typeExpressionToString(match.candidate.entry.definition.TargetType)
		label := fmt.Sprintf("impl %s for %s", match.candidate.entry.interfaceName, target)
		set[strings.TrimSpace(label)] = struct{}{}
	}
	return sortedKeys(set)
}

func descriptionsFromCandidates(matches []implCandidate) []string {
	set := make(map[string]struct{})
	for _, match := range matches {
		if match.entry == nil || match.entry.definition == nil {
			continue
		}
		target := typeExpressionToString(match.entry.definition.TargetType)
		label := fmt.Sprintf("impl %s for %s", match.entry.interfaceName, target)
		set[strings.TrimSpace(label)] = struct{}{}
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
