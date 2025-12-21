package typechecker

import (
	"fmt"
	"sort"
	"strings"
)

type interfaceResolution struct {
	iface InterfaceType
	args  []Type
	name  string
	err   string
}

type implementationMatch struct {
	spec           ImplementationSpec
	substitution   map[string]Type
	actualArgs     []Type
	specificity    int
	constraintKeys map[string]struct{}
	isConcrete     bool
}

func (c *Checker) resolveObligations() []Diagnostic {
	return c.evaluateObligations(c.obligations)
}

func (c *Checker) evaluateObligations(obligations []ConstraintObligation) []Diagnostic {
	if len(obligations) == 0 {
		return nil
	}
	methodObligation := map[string]bool{}
	for _, ob := range obligations {
		if strings.HasPrefix(ob.Owner, "methods for ") && strings.Contains(ob.Owner, "::") {
			label := strings.TrimSpace(strings.TrimPrefix(ob.Owner, "methods for "))
			if parts := strings.Split(label, "::"); len(parts) > 0 {
				methodObligation[parts[0]] = true
			}
		}
	}
	filtered := make([]ConstraintObligation, 0, len(obligations))
	for _, ob := range obligations {
		if strings.HasPrefix(ob.Owner, "methods for ") && !strings.Contains(ob.Owner, "::") && ob.Context == "via method set" {
			label := strings.TrimSpace(strings.TrimPrefix(ob.Owner, "methods for "))
			if methodObligation[label] {
				continue
			}
		}
		filtered = append(filtered, ob)
	}
	var diags []Diagnostic
	for _, ob := range filtered {
		diags = append(diags, c.evaluateObligation(ob)...)
	}
	return diags
}

func (c *Checker) evaluateObligation(ob ConstraintObligation) []Diagnostic {
	res := c.resolveConstraintInterfaceType(ob.Constraint)
	contextLabel := ""
	if ob.Context != "" {
		contextLabel = " (" + ob.Context + ")"
	}
	if ob.Context == "via method set" && strings.HasPrefix(ob.Owner, "methods for ") {
		contextLabel = ""
	}
	if res.err != "" {
		diags := []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s %s", ob.Owner, ob.TypeParam, contextLabel, res.err),
			Node:    ob.Node,
		}}
		if ob.Subject != nil && !isUnknownType(ob.Subject) && !isTypeParameter(ob.Subject) {
			subject := formatType(ob.Subject)
			interfaceLabel := res.name
			if interfaceLabel == "" && res.iface.InterfaceName != "" {
				interfaceLabel = res.iface.InterfaceName
			}
			if len(res.args) > 0 {
				interfaceLabel = formatInterfaceApplication(res.iface, res.args)
			}
			if interfaceLabel == "" {
				interfaceLabel = "<unknown>"
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: %s constraint on %s%s is not satisfied: %s does not implement %s", ob.Owner, ob.TypeParam, contextLabel, subject, interfaceLabel),
				Node:    ob.Node,
			})
		}
		return diags
	}
	expectedParams := len(res.iface.TypeParams)
	providedArgs := len(res.args)
	if expectedParams > 0 && providedArgs == 0 {
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s requires %d type argument(s) for interface '%s'", ob.Owner, ob.TypeParam, contextLabel, expectedParams, res.iface.InterfaceName),
			Node:    ob.Node,
		}}
	}
	if expectedParams != providedArgs && providedArgs != 0 {
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s expected %d type argument(s) for interface '%s', got %d", ob.Owner, ob.TypeParam, contextLabel, expectedParams, res.iface.InterfaceName, providedArgs),
			Node:    ob.Node,
		}}
	}
	if ok, detail := c.obligationSatisfied(ob, res); !ok {
		subject := formatType(ob.Subject)
		interfaceLabel := formatInterfaceApplication(res.iface, res.args)
		reason := ""
		if detail != "" {
			reason = ": " + detail
		}
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s is not satisfied: %s does not implement %s%s", ob.Owner, ob.TypeParam, contextLabel, subject, interfaceLabel, reason),
			Node:    ob.Node,
		}}
	}
	return nil
}

func (c *Checker) resolveConstraintInterfaceType(t Type) interfaceResolution {
	switch val := t.(type) {
	case InterfaceType:
		return interfaceResolution{iface: val, name: val.InterfaceName}
	case AppliedType:
		base := c.resolveConstraintInterfaceType(val.Base)
		if base.err != "" {
			return base
		}
		args := append([]Type{}, val.Arguments...)
		return interfaceResolution{iface: base.iface, args: args, name: base.iface.InterfaceName}
	case StructType:
		return c.interfaceFromName(val.StructName)
	case StructInstanceType:
		return c.interfaceFromName(val.StructName)
	default:
		return interfaceResolution{err: fmt.Sprintf("must reference an interface (got %s)", typeName(t)), name: typeName(t)}
	}
}

func (c *Checker) interfaceFromName(name string) interfaceResolution {
	if name == "" {
		return interfaceResolution{err: "must reference an interface (got <unknown>)", name: "<unknown>"}
	}
	if c.global == nil {
		return interfaceResolution{err: fmt.Sprintf("references unknown interface '%s'", name), name: name}
	}
	decl, ok := c.global.Lookup(name)
	if !ok {
		return interfaceResolution{err: fmt.Sprintf("references unknown interface '%s'", name), name: name}
	}
	iface, ok := decl.(InterfaceType)
	if !ok {
		return interfaceResolution{err: fmt.Sprintf("references '%s' which is not an interface", name), name: name}
	}
	return interfaceResolution{iface: iface, name: iface.InterfaceName}
}

func (c *Checker) obligationSatisfied(ob ConstraintObligation, res interfaceResolution) (bool, string) {
	subject := ob.Subject
	if subject == nil || isUnknownType(subject) || isTypeParameter(subject) {
		return true, ""
	}
	ok, detail := c.typeImplementsInterface(subject, res.iface, res.args)
	if !ok {
		return false, detail
	}
	return true, ""
}

func (c *Checker) typeImplementsInterface(subject Type, iface InterfaceType, args []Type) (bool, string) {
	if implementsIntrinsicInterface(subject, iface.InterfaceName) {
		return true, ""
	}
	var implDetail string
	switch val := subject.(type) {
	case NullableType:
		if ok, detail := c.implementationProvidesInterface(subject, iface, args); ok {
			return true, ""
		} else if detail != "" {
			implDetail = detail
		}
		ok, detail := c.typeImplementsInterface(val.Inner, iface, args)
		if !ok {
			if detail != "" {
				return false, detail
			}
			if implDetail != "" {
				return false, implDetail
			}
			return false, ""
		}
		return true, ""
	case UnionLiteralType:
		if ok, detail := c.implementationProvidesInterface(subject, iface, args); ok {
			return true, ""
		} else if detail != "" {
			implDetail = detail
		}
		for _, member := range val.Members {
			ok, detail := c.typeImplementsInterface(member, iface, args)
			if !ok {
				if detail != "" {
					return false, detail
				}
				if implDetail != "" {
					return false, implDetail
				}
				return false, detail
			}
		}
		return true, ""
	}
	if subjectMatchesInterface(subject, iface, args) {
		return true, ""
	}
	if ok, detail := c.implementationProvidesInterface(subject, iface, args); ok {
		return true, ""
	} else if detail != "" {
		implDetail = detail
	}
	if ok, detail := c.methodSetProvidesInterface(subject, iface, args); ok {
		return true, ""
	} else if detail != "" {
		return false, detail
	}
	if implDetail != "" {
		return false, implDetail
	}
	return false, ""
}

func implementsIntrinsicInterface(subject Type, interfaceName string) bool {
	switch interfaceName {
	case "Hash", "Eq":
		switch val := subject.(type) {
		case PrimitiveType:
			return val.Kind == PrimitiveString || val.Kind == PrimitiveBool || val.Kind == PrimitiveChar
		case IntegerType:
			return val.Suffix == "i32"
		}
	}
	return false
}

func subjectMatchesInterface(subject Type, iface InterfaceType, args []Type) bool {
	switch val := subject.(type) {
	case InterfaceType:
		return val.InterfaceName == iface.InterfaceName
	case AppliedType:
		baseIface, ok := val.Base.(InterfaceType)
		if !ok || baseIface.InterfaceName != iface.InterfaceName {
			return false
		}
		if len(args) == 0 {
			return true
		}
		return interfaceArgsCompatible(val.Arguments, args)
	default:
		return false
	}
}

func (c *Checker) implementationProvidesInterface(subject Type, iface InterfaceType, args []Type) (bool, string) {
	if len(c.implementations) == 0 {
		return false, ""
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
		subst, _, ok := matchMethodTarget(subject, spec.Target, spec.TypeParams)
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

func (c *Checker) methodSetProvidesInterface(subject Type, iface InterfaceType, args []Type) (bool, string) {
	if len(c.methodSets) == 0 {
		return false, ""
	}
	bestDetail := ""
	bestScore := -1
	for _, spec := range c.methodSets {
		subst, score, ok := matchMethodTarget(subject, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		satisfied, obligations, detail := methodSetSatisfiesInterface(spec, iface, args, subject, subst)
		if detail != "" {
			annotated := annotateMethodSetFailure(detail, spec, subject, subst, ConstraintObligation{})
			if score > bestScore || (score == bestScore && len(annotated) > len(bestDetail)) {
				bestScore = score
				bestDetail = annotated
			}
			continue
		}
		if !satisfied {
			continue
		}
		ok, obligationDetail, ob := c.obligationSetSatisfied(obligations)
		if ok {
			return true, ""
		}
		annotated := annotateMethodSetFailure(obligationDetail, spec, subject, subst, ob)
		if score > bestScore || (score == bestScore && len(annotated) > len(bestDetail)) {
			bestScore = score
			bestDetail = annotated
		}
	}
	if bestDetail != "" {
		return false, bestDetail
	}
	return false, ""
}

func (c *Checker) obligationSetSatisfied(obligations []ConstraintObligation) (bool, string, ConstraintObligation) {
	if len(obligations) == 0 {
		return true, "", ConstraintObligation{}
	}
	appendContext := func(detail string, context string) string {
		if context == "" {
			return detail
		}
		if detail == "" {
			return context
		}
		return detail + " (" + context + ")"
	}
	for _, ob := range obligations {
		if ob.Constraint == nil {
			continue
		}
		res := c.resolveConstraintInterfaceType(ob.Constraint)
		if res.err != "" {
			return false, appendContext(res.err, ob.Context), ob
		}
		if ok, detail := c.obligationSatisfied(ob, res); !ok {
			if detail != "" {
				return false, appendContext(detail, ob.Context), ob
			}
			return false, appendContext(detail, ob.Context), ob
		}
	}
	return true, "", ConstraintObligation{}
}

func annotateMethodSetFailure(detail string, spec MethodSetSpec, subject Type, subst map[string]Type, ob ConstraintObligation) string {
	label := strings.TrimSpace(ob.Owner)
	if label == "" {
		label = formatMethodSetCandidateLabel(spec, subject, subst)
	}
	trimmed := strings.TrimSpace(detail)
	context := strings.TrimSpace(ob.Context)
	if trimmed == "" {
		if context != "" {
			return label + ": " + context
		}
		return label
	}
	if context != "" && !strings.Contains(trimmed, context) {
		trimmed = trimmed + " (" + context + ")"
	}
	if strings.HasPrefix(trimmed, label) {
		return trimmed
	}
	return label + ": " + trimmed
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

func formatMethodSetCandidateLabel(spec MethodSetSpec, subject Type, subst map[string]Type) string {
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
	return fmt.Sprintf("methods for %s", name)
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
	return false, formatAmbiguousImplementationDetail(iface, subject, contenders)
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
		label := formatImplementationCandidateLabel(match.spec, subject, match.substitution, match.actualArgs)
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

func methodSetSatisfiesInterface(spec MethodSetSpec, iface InterfaceType, args []Type, subject Type, subst map[string]Type) (bool, []ConstraintObligation, string) {
	if len(spec.Methods) == 0 || len(iface.Methods) == 0 {
		return false, nil, "method set is empty"
	}
	combined := cloneTypeMap(subst)
	if combined == nil {
		combined = make(map[string]Type)
	}
	combined["Self"] = subject

	for idx, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		if idx < len(args) && args[idx] != nil && !isUnknownType(args[idx]) {
			combined[param.Name] = args[idx]
		}
	}

	ifaceSubst := make(map[string]Type, len(combined)+1)
	ifaceSubst["Self"] = subject
	for idx, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		var replacement Type = TypeParameterType{ParameterName: param.Name}
		if idx < len(args) && args[idx] != nil && !isUnknownType(args[idx]) {
			replacement = args[idx]
		}
		ifaceSubst[param.Name] = replacement
	}

	var obligations []ConstraintObligation

	for name, ifaceMethod := range iface.Methods {
		expected := substituteFunctionType(ifaceMethod, ifaceSubst)
		actual, ok := spec.Methods[name]
		if !ok {
			return false, nil, fmt.Sprintf("method '%s' not provided", name)
		}
		actualInst := substituteFunctionType(actual, combined)
		if !functionSignaturesCompatible(expected, actualInst) {
			return false, nil, fmt.Sprintf("method '%s' signature does not satisfy interface", name)
		}
		if len(actualInst.Obligations) > 0 {
			populated := populateObligationSubjects(actualInst.Obligations, subject)
			for i := range populated {
				populated[i].Context = fmt.Sprintf("via method '%s'", name)
			}
			obligations = append(obligations, substituteObligations(populated, combined)...)
		}
	}

	if len(spec.Obligations) > 0 {
		populated := populateObligationSubjects(spec.Obligations, subject)
		for i := range populated {
			if populated[i].Context == "" {
				populated[i].Context = "via method set"
			}
		}
		obligations = append(obligations, substituteObligations(populated, combined)...)
	}

	if len(obligations) == 0 {
		return true, nil, ""
	}
	return true, obligations, ""
}

func functionSignaturesCompatible(expected, actual FunctionType) bool {
	if len(expected.TypeParams) != len(actual.TypeParams) {
		return false
	}
	if len(expected.Params) != len(actual.Params) {
		return false
	}
	for i := range expected.Params {
		if !typesEquivalentForSignature(expected.Params[i], actual.Params[i]) {
			return false
		}
	}
	return typesEquivalentForSignature(expected.Return, actual.Return)
}
