package typechecker

import (
	"fmt"
	"strings"
)

type interfaceResolution struct {
	iface InterfaceType
	args  []Type
	name  string
	err   string
}

func (c *Checker) resolveObligations() []Diagnostic {
	var diags []Diagnostic
	for _, ob := range c.obligations {
		res := c.resolveConstraintInterfaceType(ob.Constraint)
		contextLabel := ""
		if ob.Context != "" {
			contextLabel = " (" + ob.Context + ")"
		}
		if res.err != "" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: %s constraint on %s%s %s", ob.Owner, ob.TypeParam, contextLabel, res.err),
				Node:    ob.Node,
			})
			continue
		}
		expectedParams := len(res.iface.TypeParams)
		providedArgs := len(res.args)
		if expectedParams > 0 && providedArgs == 0 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: %s constraint on %s%s requires %d type argument(s) for interface '%s'", ob.Owner, ob.TypeParam, contextLabel, expectedParams, res.iface.InterfaceName),
				Node:    ob.Node,
			})
			continue
		}
		if expectedParams != providedArgs && providedArgs != 0 {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: %s constraint on %s%s expected %d type argument(s) for interface '%s', got %d", ob.Owner, ob.TypeParam, contextLabel, expectedParams, res.iface.InterfaceName, providedArgs),
				Node:    ob.Node,
			})
			continue
		}
		if ok, detail := c.obligationSatisfied(ob, res); !ok {
			subject := formatType(ob.Subject)
			interfaceLabel := formatInterfaceApplication(res.iface, res.args)
			reason := ""
			if detail != "" {
				reason = ": " + detail
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: %s constraint on %s%s is not satisfied: %s does not implement %s%s", ob.Owner, ob.TypeParam, contextLabel, subject, interfaceLabel, reason),
				Node:    ob.Node,
			})
		}
	}
	return diags
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
	switch val := subject.(type) {
	case NullableType:
		if c.implementationProvidesInterface(subject, iface, args) {
			return true, ""
		}
		return c.typeImplementsInterface(val.Inner, iface, args)
	case UnionLiteralType:
		if c.implementationProvidesInterface(subject, iface, args) {
			return true, ""
		}
		for _, member := range val.Members {
			ok, detail := c.typeImplementsInterface(member, iface, args)
			if !ok {
				return false, detail
			}
		}
		return true, ""
	}
	if subjectMatchesInterface(subject, iface, args) {
		return true, ""
	}
	if c.implementationProvidesInterface(subject, iface, args) {
		return true, ""
	}
	if ok, detail := c.methodSetProvidesInterface(subject, iface, args); ok {
		return true, ""
	} else if detail != "" {
		return false, detail
	}
	return false, ""
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

func (c *Checker) implementationProvidesInterface(subject Type, iface InterfaceType, args []Type) bool {
	if len(c.implementations) == 0 {
		return false
	}
	for _, spec := range c.implementations {
		if spec.InterfaceName != iface.InterfaceName {
			continue
		}
		subst, _, ok := matchMethodTarget(subject, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		if len(spec.InterfaceArgs) == 0 && len(args) == 0 {
			return true
		}
		actualArgs := make([]Type, len(spec.InterfaceArgs))
		for i, arg := range spec.InterfaceArgs {
			actualArgs[i] = substituteType(arg, subst)
		}
		if len(actualArgs) == 0 && len(args) == 0 {
			return true
		}
		if len(actualArgs) == 0 || len(actualArgs) != len(args) {
			continue
		}
		if interfaceArgsCompatible(actualArgs, args) {
			return true
		}
	}
	return false
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
	if trimmed == "" {
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

func methodSetSatisfiesInterface(spec MethodSetSpec, iface InterfaceType, args []Type, subject Type, subst map[string]Type) (bool, []ConstraintObligation, string) {
	if len(spec.Methods) == 0 || len(iface.Methods) == 0 {
		return false, nil, ""
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
			return false, nil, fmt.Sprintf("method '%s' not provided by methods block", name)
		}
		actualInst := substituteFunctionType(actual, combined)
		if !functionSignaturesCompatible(expected, actualInst) {
			return false, nil, fmt.Sprintf("method '%s' has incompatible signature", name)
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
