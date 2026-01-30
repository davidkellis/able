package typechecker

import (
	"fmt"
	"sort"
	"strings"
)

type implementationMatch struct {
	spec           ImplementationSpec
	substitution   map[string]Type
	actualArgs     []Type
	specificity    int
	constraintKeys map[string]struct{}
	isConcrete     bool
}

func (c *Checker) implementationProvidesInterface(subject Type, iface InterfaceType, args []Type) (bool, string) {
	if len(c.implementations) == 0 {
		return false, ""
	}
	matches, bestDetail := c.collectImplementationMatches(subject, iface, args)
	if len(matches) == 0 {
		return false, bestDetail
	}
	if len(matches) == 1 {
		return true, ""
	}
	ok, detail := c.selectMostSpecificImplementationMatch(matches, iface, subject)
	if ok {
		return true, ""
	}
	if detail != "" {
		return false, detail
	}
	return false, bestDetail
}

func (c *Checker) matchImplementationTarget(subject Type, target Type, params []GenericParamSpec) (map[string]Type, int, bool) {
	if iface, args, ok := interfaceTargetFromType(target); ok {
		return c.matchInterfaceTarget(subject, iface, args, params)
	}
	return matchMethodTarget(subject, target, params)
}

func interfaceTargetFromType(t Type) (InterfaceType, []Type, bool) {
	if t == nil {
		return InterfaceType{}, nil, false
	}
	iface, args, ok := resolveInterfaceDecl(t, nil)
	if !ok || iface.InterfaceName == "" {
		return InterfaceType{}, nil, false
	}
	return iface, args, true
}

func (c *Checker) matchInterfaceTarget(subject Type, iface InterfaceType, targetArgs []Type, params []GenericParamSpec) (map[string]Type, int, bool) {
	if subject == nil || isUnknownType(subject) {
		return nil, 0, false
	}
	if subjIface, subjArgs, ok := interfaceTargetFromType(subject); ok && subjIface.InterfaceName == iface.InterfaceName {
		return matchInterfaceArgs(subjArgs, targetArgs, params)
	}
	if info, ok := structInfoFromType(subject); ok && info.name == iface.InterfaceName && !info.isUnion && !info.isNullable {
		return matchInterfaceArgs(info.args, targetArgs, params)
	}
	return c.matchInterfaceTargetFromImplementations(subject, iface, targetArgs, params)
}

func matchInterfaceArgs(actualArgs []Type, targetArgs []Type, params []GenericParamSpec) (map[string]Type, int, bool) {
	if len(targetArgs) == 0 {
		return finalizeMatchResult(nil, params, 0)
	}
	if len(actualArgs) != len(targetArgs) {
		return nil, 0, false
	}
	subst := make(map[string]Type)
	score := 0
	for i := range targetArgs {
		ok, s := matchTypeArgument(actualArgs[i], targetArgs[i], subst)
		if !ok {
			return nil, 0, false
		}
		score += s
	}
	return finalizeMatchResult(subst, params, score)
}

func (c *Checker) matchInterfaceTargetFromImplementations(subject Type, iface InterfaceType, targetArgs []Type, params []GenericParamSpec) (map[string]Type, int, bool) {
	matches, _ := c.collectImplementationMatches(subject, iface, nil)
	if len(matches) == 0 {
		return nil, 0, false
	}
	type candidate struct {
		match        implementationMatch
		substitution map[string]Type
		score        int
	}
	var candidates []candidate
	for _, match := range matches {
		subst, score, ok := matchInterfaceArgs(match.actualArgs, targetArgs, params)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate{
			match:        match,
			substitution: subst,
			score:        score,
		})
	}
	if len(candidates) == 0 {
		return nil, 0, false
	}
	if len(candidates) == 1 {
		return candidates[0].substitution, candidates[0].score, true
	}
	best := candidates[0]
	contenders := []candidate{best}
	for _, cand := range candidates[1:] {
		cmp := compareImplementationMatches(cand.match, best.match)
		if cmp > 0 {
			best = cand
			contenders = []candidate{cand}
			continue
		}
		if cmp == 0 {
			reverse := compareImplementationMatches(best.match, cand.match)
			if reverse < 0 {
				best = cand
				contenders = []candidate{cand}
			} else if reverse == 0 {
				contenders = append(contenders, cand)
			}
		}
	}
	if len(contenders) == 1 {
		return best.substitution, best.score, true
	}
	return nil, 0, false
}

func (c *Checker) collectImplementationMatches(subject Type, iface InterfaceType, args []Type) ([]implementationMatch, string) {
	if len(c.implementations) == 0 {
		return nil, ""
	}
	var matches []implementationMatch
	bestDetail := ""
	for _, spec := range c.implementations {
		if spec.ImplName != "" {
			continue
		}
		if spec.InterfaceName != iface.InterfaceName {
			continue
		}
		subst, _, ok := c.matchImplementationTarget(subject, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		substitution := cloneTypeMap(subst)
		if substitution == nil {
			substitution = make(map[string]Type)
		}
		if subject != nil {
			substitution["Self"] = subject
		}
		extendImplementationSubstitution(substitution, iface, spec.InterfaceArgs)
		for _, param := range spec.TypeParams {
			if param.Name == "" {
				continue
			}
			if _, ok := substitution[param.Name]; !ok {
				substitution[param.Name] = UnknownType{}
			}
		}
		actualArgs := make([]Type, len(spec.InterfaceArgs))
		for i, arg := range spec.InterfaceArgs {
			actualArgs[i] = substituteType(arg, substitution)
		}
		if !interfaceArgsCompatible(actualArgs, args) {
			expected := formatInterfaceApplication(iface, args)
			if expected == "" {
				expected = "(none)"
			}
			label := formatImplementationCandidateLabel(spec, subject, substitution, actualArgs)
			detail := fmt.Sprintf("%s: interface arguments do not match expected %s", label, expected)
			if len(detail) > len(bestDetail) {
				bestDetail = detail
			}
			continue
		}
		if len(spec.Obligations) > 0 {
			populated := populateObligationSubjects(spec.Obligations, subject)
			substituted := substituteObligations(populated, substitution)
			if ok, detail, ob := c.obligationSetSatisfied(substituted); !ok {
				annotated := annotateImplementationFailure(detail, spec, subject, substitution, actualArgs, ob)
				if len(annotated) > len(bestDetail) {
					bestDetail = annotated
				}
				continue
			}
		}
		matches = append(matches, implementationMatch{
			spec:           spec,
			substitution:   substitution,
			actualArgs:     actualArgs,
			specificity:    computeImplementationSpecificity(spec),
			constraintKeys: buildImplementationConstraintKeySet(spec),
			isConcrete:     !implementationTargetUsesTypeParams(spec.Target),
		})
	}
	return matches, bestDetail
}

func interfaceArgsCompatible(actual []Type, expected []Type) bool {
	if len(expected) == 0 {
		return true
	}
	if len(actual) != len(expected) {
		return false
	}
	for i := range expected {
		a := actual[i]
		b := expected[i]
		if a == nil || isUnknownType(a) || b == nil || isUnknownType(b) {
			continue
		}
		if isTypeParameter(a) || isTypeParameter(b) {
			continue
		}
		if !typesEquivalentForSignature(a, b) {
			return false
		}
	}
	return true
}

func formatInterfaceApplication(iface InterfaceType, args []Type) string {
	name := iface.InterfaceName
	if name == "" {
		name = "<unknown>"
	}
	if len(args) == 0 {
		return name
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatType(arg)
	}
	return strings.TrimSpace(name + " " + strings.Join(parts, " "))
}

func annotateImplementationFailure(detail string, spec ImplementationSpec, subject Type, subst map[string]Type, args []Type, ob ConstraintObligation) string {
	label := strings.TrimSpace(spec.ImplName)
	if label == "" {
		label = formatImplementationCandidateLabel(spec, subject, subst, args)
	}
	trimmed := strings.TrimSpace(detail)
	if trimmed == "" {
		context := strings.TrimSpace(ob.Context)
		if context != "" {
			return label + ": " + context
		}
		return label
	}
	if strings.HasPrefix(trimmed, label) {
		return trimmed
	}
	return label + ": " + trimmed
}

func formatImplementationCandidateLabel(spec ImplementationSpec, subject Type, subst map[string]Type, args []Type) string {
	target := spec.Target
	if len(subst) > 0 {
		target = substituteType(target, subst)
	}
	if (target == nil || isUnknownType(target)) && subject != nil && !isUnknownType(subject) {
		target = subject
	}
	name := formatType(target)
	if name == "" || name == "<unknown>" {
		name = typeName(target)
	}
	if name == "" {
		name = "<unknown>"
	}
	interfaceName := spec.InterfaceName
	if interfaceName == "" {
		interfaceName = "<unknown>"
	}
	var parts []string
	if len(args) > 0 {
		parts = make([]string, len(args))
		for i, arg := range args {
			parts[i] = formatType(arg)
		}
	} else if len(spec.InterfaceArgs) > 0 {
		parts = make([]string, len(spec.InterfaceArgs))
		for i, arg := range spec.InterfaceArgs {
			parts[i] = formatType(substituteType(arg, subst))
		}
	}
	argSuffix := ""
	if len(parts) > 0 {
		argSuffix = " " + strings.Join(parts, " ")
	}
	return fmt.Sprintf("impl %s%s for %s", interfaceName, argSuffix, name)
}

func formatImplementationCandidateLabelForAmbiguity(spec ImplementationSpec, subject Type, subst map[string]Type) string {
	target := spec.Target
	if len(subst) > 0 {
		target = substituteType(target, subst)
	}
	if (target == nil || isUnknownType(target)) && subject != nil && !isUnknownType(subject) {
		target = subject
	}
	name := formatType(target)
	if name == "" || name == "<unknown>" {
		name = typeName(target)
	}
	if name == "" {
		name = "<unknown>"
	}
	interfaceName := spec.InterfaceName
	if interfaceName == "" {
		interfaceName = "<unknown>"
	}
	return fmt.Sprintf("impl %s for %s", interfaceName, name)
}

func (c *Checker) selectMostSpecificImplementationMatch(matches []implementationMatch, iface InterfaceType, subject Type) (bool, string) {
	best := matches[0]
	contenders := []implementationMatch{best}
	for _, candidate := range matches[1:] {
		cmp := compareImplementationMatches(candidate, best)
		if cmp > 0 {
			best = candidate
			contenders = []implementationMatch{candidate}
			continue
		}
		if cmp == 0 {
			reverse := compareImplementationMatches(best, candidate)
			if reverse < 0 {
				best = candidate
				contenders = []implementationMatch{candidate}
			} else if reverse == 0 {
				contenders = append(contenders, candidate)
			}
		}
	}
	if len(contenders) == 1 {
		return true, ""
	}
	if chosen := pickInterfaceSpecificityPair(contenders, "Eq", "PartialEq"); chosen != nil {
		return true, ""
	}
	if chosen := pickInterfaceSpecificityPair(contenders, "Ord", "PartialOrd"); chosen != nil {
		return true, ""
	}
	return false, formatAmbiguousImplementationDetail(iface, subject, contenders)
}

func pickInterfaceSpecificityPair(contenders []implementationMatch, preferred, fallback string) *implementationMatch {
	var preferredMatch *implementationMatch
	foundFallback := false
	for i := range contenders {
		name := contenders[i].spec.InterfaceName
		if name == preferred {
			if preferredMatch != nil {
				return nil
			}
			preferredMatch = &contenders[i]
			continue
		}
		if name == fallback {
			foundFallback = true
		}
	}
	if preferredMatch != nil && foundFallback {
		return preferredMatch
	}
	return nil
}

func compareImplementationMatches(a, b implementationMatch) int {
	if a.isConcrete && !b.isConcrete {
		return 1
	}
	if b.isConcrete && !a.isConcrete {
		return -1
	}
	if isConstraintSupersetMap(a.constraintKeys, b.constraintKeys) {
		return 1
	}
	if isConstraintSupersetMap(b.constraintKeys, a.constraintKeys) {
		return -1
	}
	aUnion := a.spec.UnionVariants
	bUnion := b.spec.UnionVariants
	if len(aUnion) > 0 && len(bUnion) == 0 {
		return -1
	}
	if len(bUnion) > 0 && len(aUnion) == 0 {
		return 1
	}
	if len(aUnion) > 0 && len(bUnion) > 0 {
		if isProperSubsetStrings(aUnion, bUnion) {
			return 1
		}
		if isProperSubsetStrings(bUnion, aUnion) {
			return -1
		}
		if len(aUnion) != len(bUnion) {
			if len(aUnion) < len(bUnion) {
				return 1
			}
			return -1
		}
	}
	if a.specificity > b.specificity {
		return 1
	}
	if a.specificity < b.specificity {
		return -1
	}
	if a.spec.IsBuiltin != b.spec.IsBuiltin {
		if a.spec.IsBuiltin {
			return -1
		}
		return 1
	}
	if cmp := compareInterfaceSpecificity(a.spec.InterfaceName, b.spec.InterfaceName); cmp != 0 {
		return cmp
	}
	return 0
}

func compareInterfaceSpecificity(a, b string) int {
	if a == b {
		return 0
	}
	switch a {
	case "Eq":
		if b == "PartialEq" {
			return 1
		}
	case "PartialEq":
		if b == "Eq" {
			return -1
		}
	case "Ord":
		if b == "PartialOrd" {
			return 1
		}
	case "PartialOrd":
		if b == "Ord" {
			return -1
		}
	}
	return 0
}

func formatAmbiguousImplementationDetail(iface InterfaceType, subject Type, matches []implementationMatch) string {
	typeLabel := formatType(subject)
	if typeLabel == "" {
		typeLabel = typeName(subject)
	}
	if typeLabel == "" {
		typeLabel = "<unknown>"
	}
	interfaceLabel := iface.InterfaceName
	if len(matches) > 0 {
		interfaceLabel = matches[0].spec.InterfaceName
	}
	if interfaceLabel == "" {
		interfaceLabel = "<unknown>"
	}
	labels := make([]string, 0, len(matches))
	for _, match := range matches {
		label := formatImplementationCandidateLabelForAmbiguity(match.spec, subject, match.substitution)
		labels = append(labels, label)
	}
	unique := uniqueSortedStrings(labels)
	return fmt.Sprintf("ambiguous implementations of %s for %s: %s", interfaceLabel, typeLabel, strings.Join(unique, ", "))
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func computeImplementationSpecificity(spec ImplementationSpec) int {
	return typeSpecificityScore(spec.Target)
}

func implementationTargetUsesTypeParams(t Type) bool {
	switch val := t.(type) {
	case TypeParameterType:
		return true
	case AppliedType:
		switch base := val.Base.(type) {
		case StructType:
			if len(val.Arguments) == 0 && len(base.TypeParams) > 0 {
				return true
			}
		case UnionType:
			if len(val.Arguments) == 0 && len(base.TypeParams) > 0 {
				return true
			}
		case InterfaceType:
			if len(val.Arguments) == 0 && len(base.TypeParams) > 0 {
				return true
			}
		default:
			if implementationTargetUsesTypeParams(val.Base) {
				return true
			}
		}
		for _, arg := range val.Arguments {
			if implementationTargetUsesTypeParams(arg) {
				return true
			}
		}
		return false
	case NullableType:
		return implementationTargetUsesTypeParams(val.Inner)
	case UnionLiteralType:
		for _, member := range val.Members {
			if implementationTargetUsesTypeParams(member) {
				return true
			}
		}
		return false
	case StructInstanceType:
		for _, arg := range val.TypeArgs {
			if implementationTargetUsesTypeParams(arg) {
				return true
			}
		}
		return false
	case FunctionType:
		for _, p := range val.Params {
			if implementationTargetUsesTypeParams(p) {
				return true
			}
		}
		return implementationTargetUsesTypeParams(val.Return)
	case ArrayType:
		return implementationTargetUsesTypeParams(val.Element)
	case MapType:
		return implementationTargetUsesTypeParams(val.Key) || implementationTargetUsesTypeParams(val.Value)
	case AliasType:
		return implementationTargetUsesTypeParams(val.Target)
	case StructType:
		return len(val.TypeParams) > 0
	case InterfaceType:
		return len(val.TypeParams) > 0
	default:
		return false
	}
}

func typeSpecificityScore(t Type) int {
	switch val := t.(type) {
	case StructType, StructInstanceType, PrimitiveType, IntegerType, FloatType:
		return 1
	case AppliedType:
		score := typeSpecificityScore(val.Base)
		for _, arg := range val.Arguments {
			score += typeSpecificityScore(arg)
		}
		return score
	case NullableType:
		return typeSpecificityScore(val.Inner)
	case UnionLiteralType:
		total := 0
		for _, member := range val.Members {
			total += typeSpecificityScore(member)
		}
		return total
	case TypeParameterType:
		return 0
	default:
		return 0
	}
}

func buildImplementationConstraintKeySet(spec ImplementationSpec) map[string]struct{} {
	if len(spec.Obligations) == 0 {
		return nil
	}
	result := make(map[string]struct{}, len(spec.Obligations))
	for _, ob := range spec.Obligations {
		if ob.Constraint == nil {
			continue
		}
		label := formatType(ob.Constraint)
		if label == "" {
			label = typeName(ob.Constraint)
		}
		if label == "" {
			label = "<unknown>"
		}
		key := fmt.Sprintf("%s->%s", ob.TypeParam, label)
		result[key] = struct{}{}
	}
	return result
}

func isConstraintSupersetMap(a, b map[string]struct{}) bool {
	if len(a) == 0 || len(a) <= len(b) {
		return false
	}
	for key := range b {
		if _, ok := a[key]; !ok {
			return false
		}
	}
	return true
}

func isProperSubsetStrings(a, b []string) bool {
	if len(a) == 0 || len(a) >= len(b) {
		return false
	}
	set := make(map[string]struct{}, len(b))
	for _, value := range b {
		set[value] = struct{}{}
	}
	for _, value := range a {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}
