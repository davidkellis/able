package typechecker

import (
	"fmt"
	"strings"
)

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
