package typechecker

import (
	"fmt"
	"strings"
)

func (c *Checker) lookupMethod(object Type, name string, allowMethodSets bool, allowTypeQualified bool) (FunctionType, bool, string) {
	bestFn, bestScore, found, detail := c.lookupMethodInMethodSets(object, name, allowMethodSets, allowTypeQualified)
	implFn, implScore, implFound, implDetail := c.lookupMethodInImplementations(object, name, true)
	if implFound && (!found || implScore > bestScore) {
		return implFn, true, ""
	}
	if found {
		return bestFn, true, ""
	}
	if detail != "" {
		return FunctionType{}, false, detail
	}
	return FunctionType{}, false, implDetail
}

func (c *Checker) lookupMethodInMethodSets(object Type, name string, allowMethodSets bool, allowTypeQualified bool) (FunctionType, int, bool, string) {
	if len(c.methodSets) == 0 || !allowMethodSets {
		return FunctionType{}, -1, false, ""
	}
	type candidate struct {
		fn    FunctionType
		score int
	}
	var candidates []candidate
	for _, spec := range c.methodSets {
		subst, score, ok := matchMethodTarget(object, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		substitution := cloneTypeMap(subst)
		if substitution == nil {
			substitution = make(map[string]Type)
		}
		if object != nil {
			substitution["Self"] = object
		}
		method, ok := spec.Methods[name]
		derivedFromConstraints := false
		if spec.TypeQualified != nil && spec.TypeQualified[name] && !allowTypeQualified {
			continue
		}
		if !ok {
			if inferred, inferredOK := c.methodFromMethodSetConstraints(spec.Obligations, substitution, name); inferredOK {
				method = inferred
				ok = true
				derivedFromConstraints = true
			}
		}
		if !ok {
			continue
		}
		if spec.TypeQualified != nil && spec.TypeQualified[name] && len(substitution) > 0 {
			for _, param := range spec.TypeParams {
				if param.Name == "" {
					continue
				}
				if val, exists := substitution[param.Name]; exists {
					if val == nil || isUnknownType(val) {
						delete(substitution, param.Name)
					}
				}
			}
		}
		if len(substitution) > 0 {
			method = substituteFunctionType(method, substitution)
		}
		if len(spec.Obligations) > 0 && !derivedFromConstraints {
			obligations := populateObligationSubjects(spec.Obligations, object)
			obligations = substituteObligations(obligations, substitution)
			if len(obligations) > 0 {
				ownerLabel := methodSetOwnerLabel(spec, substitution, name)
				for i := range obligations {
					if obligations[i].Owner == "" || !strings.Contains(obligations[i].Owner, "::") {
						obligations[i].Owner = ownerLabel
					}
				}
				method.Obligations = append(method.Obligations, obligations...)
			}
		}
		if shouldBindSelfParam(method, object) {
			method = bindMethodType(method)
		}
		candidates = append(candidates, candidate{fn: method, score: score})
	}
	if len(candidates) == 0 {
		return FunctionType{}, -1, false, ""
	}
	bestIdx := 0
	for idx := 1; idx < len(candidates); idx++ {
		if candidates[idx].score > candidates[bestIdx].score {
			bestIdx = idx
		}
	}
	best := candidates[bestIdx]
	for idx, cand := range candidates {
		if idx == bestIdx {
			continue
		}
		if cand.score != best.score {
			continue
		}
		if functionSignaturesEquivalent(cand.fn, best.fn) {
			return FunctionType{}, -1, false, fmt.Sprintf("ambiguous overload for %s", name)
		}
	}
	return best.fn, best.score, true, ""
}

func (c *Checker) methodFromMethodSetConstraints(obligations []ConstraintObligation, substitution map[string]Type, methodName string) (FunctionType, bool) {
	if len(obligations) == 0 {
		return FunctionType{}, false
	}
	for _, ob := range obligations {
		if ob.Constraint == nil {
			continue
		}
		constraint := ob.Constraint
		if len(substitution) > 0 {
			constraint = substituteType(constraint, substitution)
		}
		res := c.resolveConstraintInterfaceType(constraint)
		if res.err != "" {
			continue
		}
		if res.iface.Methods == nil {
			continue
		}
		method, ok := res.iface.Methods[methodName]
		if !ok {
			continue
		}
		replacement := cloneTypeMap(substitution)
		if len(res.iface.TypeParams) > 0 {
			if replacement == nil {
				replacement = make(map[string]Type)
			}
			for idx, param := range res.iface.TypeParams {
				var arg Type = UnknownType{}
				if idx < len(res.args) && res.args[idx] != nil {
					arg = res.args[idx]
				}
				replacement[param.Name] = arg
			}
		}
		if len(replacement) > 0 {
			method = substituteFunctionType(method, replacement)
		}
		return method, true
	}
	return FunctionType{}, false
}

func methodSetOwnerLabel(spec MethodSetSpec, substitution map[string]Type, methodName string) string {
	subject, ok := substitution["Self"]
	if !ok || subject == nil || isUnknownType(subject) {
		subject = spec.Target
	}
	label := nonEmpty(typeName(subject))
	return fmt.Sprintf("methods for %s::%s", label, methodName)
}

type implementationMethodBuild struct {
	fn           FunctionType
	substitution map[string]Type
	actualArgs   []Type
}

func (c *Checker) buildImplementationMethodCandidate(spec ImplementationSpec, object Type, name string) (implementationMethodBuild, int, bool, string) {
	if spec.ImplName != "" {
		return implementationMethodBuild{}, 0, false, ""
	}
	method, ok := spec.Methods[name]
	if !ok {
		return implementationMethodBuild{}, 0, false, ""
	}
	subst, score, ok := matchMethodTarget(object, spec.Target, spec.TypeParams)
	if !ok {
		return implementationMethodBuild{}, 0, false, ""
	}
	substitution := cloneTypeMap(subst)
	if substitution == nil {
		substitution = make(map[string]Type)
	}
	if object != nil {
		substitution["Self"] = object
	}
	if res := c.interfaceFromName(spec.InterfaceName); res.err == "" && res.iface.InterfaceName != "" {
		extendImplementationSubstitution(substitution, res.iface, spec.InterfaceArgs)
	}
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
	var substitutedObligations []ConstraintObligation
	if len(spec.Obligations) > 0 {
		populated := populateObligationSubjects(spec.Obligations, object)
		substitutedObligations = substituteObligations(populated, substitution)
		if ok, detail, ob := c.obligationSetSatisfied(substitutedObligations); !ok {
			annotated := annotateImplementationFailure(detail, spec, object, substitution, actualArgs, ob)
			return implementationMethodBuild{}, 0, false, annotated
		}
	}
	if len(substitution) > 0 {
		method = substituteFunctionType(method, substitution)
	}
	if len(substitutedObligations) > 0 {
		method.Obligations = append(method.Obligations, substitutedObligations...)
	}
	if shouldBindSelfParam(method, object) {
		method = bindMethodType(method)
	}
	return implementationMethodBuild{
		fn:           method,
		substitution: substitution,
		actualArgs:   actualArgs,
	}, score, true, ""
}

type implementationMethodCandidate struct {
	match  implementationMatch
	method FunctionType
	score  int
}

func (c *Checker) lookupMethodInImplementations(object Type, name string, allow bool) (FunctionType, int, bool, string) {
	if len(c.implementations) == 0 || !allow {
		return FunctionType{}, -1, false, ""
	}
	var (
		candidates []implementationMethodCandidate
		bestDetail string
	)
	for _, spec := range c.implementations {
		method, score, ok, detail := c.buildImplementationMethodCandidate(spec, object, name)
		if !ok {
			if detail != "" && len(detail) > len(bestDetail) {
				bestDetail = detail
			}
			continue
		}
		candidates = append(candidates, implementationMethodCandidate{
			match: implementationMatch{
				spec:           spec,
				substitution:   method.substitution,
				actualArgs:     method.actualArgs,
				specificity:    computeImplementationSpecificity(spec),
				constraintKeys: buildImplementationConstraintKeySet(spec),
				isConcrete:     !implementationTargetUsesTypeParams(spec.Target),
			},
			method: method.fn,
			score:  score,
		})
	}
	if len(candidates) == 0 {
		return FunctionType{}, -1, false, bestDetail
	}
	if len(candidates) == 1 {
		cand := candidates[0]
		return cand.method, cand.score, true, ""
	}
	best := candidates[0]
	contenders := []implementationMethodCandidate{best}
	for _, candidate := range candidates[1:] {
		cmp := compareImplementationMatches(candidate.match, best.match)
		if cmp > 0 {
			best = candidate
			contenders = []implementationMethodCandidate{candidate}
			continue
		}
		if cmp == 0 {
			reverse := compareImplementationMatches(best.match, candidate.match)
			if reverse < 0 {
				best = candidate
				contenders = []implementationMethodCandidate{candidate}
			} else if reverse == 0 {
				contenders = append(contenders, candidate)
			}
		}
	}
	if len(contenders) == 1 {
		return best.method, best.score, true, ""
	}
	iface := InterfaceType{InterfaceName: contenders[0].match.spec.InterfaceName}
	if res := c.interfaceFromName(iface.InterfaceName); res.err == "" && res.iface.InterfaceName != "" {
		iface = res.iface
	}
	matches := make([]implementationMatch, len(contenders))
	for i, cand := range contenders {
		match := cand.match
		match.substitution = nil
		match.actualArgs = nil
		matches[i] = match
	}
	detail := formatAmbiguousImplementationDetail(iface, object, matches)
	return FunctionType{}, -1, false, detail
}

func (c *Checker) lookupImplementationNamespaceMethod(ns ImplementationNamespaceType, name string) (FunctionType, string, bool) {
	if ns.Impl == nil {
		return FunctionType{}, "implementation has no methods", false
	}
	method, ok := ns.Impl.Methods[name]
	if !ok {
		label := "implementation"
		if ns.Impl.ImplName != "" {
			label = fmt.Sprintf("implementation '%s'", ns.Impl.ImplName)
		}
		return FunctionType{}, fmt.Sprintf("%s has no method '%s'", label, name), false
	}
	substitution := make(map[string]Type)
	target := ns.Impl.Target
	if target != nil {
		substitution["Self"] = target
	}
	if res := c.interfaceFromName(ns.Impl.InterfaceName); res.err == "" && res.iface.InterfaceName != "" {
		extendImplementationSubstitution(substitution, res.iface, ns.Impl.InterfaceArgs)
	}
	for _, param := range ns.Impl.TypeParams {
		if param.Name == "" {
			continue
		}
		if _, ok := substitution[param.Name]; !ok {
			substitution[param.Name] = TypeParameterType{ParameterName: param.Name}
		}
	}
	if len(substitution) > 0 {
		method = substituteFunctionType(method, substitution)
	}
	if len(ns.Impl.Obligations) > 0 {
		obligations := populateObligationSubjects(ns.Impl.Obligations, target)
		obligations = substituteObligations(obligations, substitution)
		method.Obligations = append(method.Obligations, obligations...)
	}
	return method, "", true
}

func (c *Checker) lookupUfcsInherentMethod(object Type, name string) (FunctionType, bool) {
	if len(c.methodSets) == 0 {
		return FunctionType{}, false
	}
	var (
		found     bool
		bestScore = -1
		bestFn    FunctionType
	)
	for _, spec := range c.methodSets {
		subst, score, ok := matchMethodTarget(object, spec.Target, spec.TypeParams)
		if !ok {
			continue
		}
		substitution := cloneTypeMap(subst)
		if substitution == nil {
			substitution = make(map[string]Type)
		}
		if object != nil {
			substitution["Self"] = object
		}
		method, ok := spec.Methods[name]
		derivedFromConstraints := false
		if !ok {
			if inferred, inferredOK := c.methodFromMethodSetConstraints(spec.Obligations, substitution, name); inferredOK {
				method = inferred
				ok = true
				derivedFromConstraints = true
			}
		}
		if !ok {
			continue
		}
		if len(substitution) > 0 {
			method = substituteFunctionType(method, substitution)
		}
		if len(spec.Obligations) > 0 && !derivedFromConstraints {
			obligations := populateObligationSubjects(spec.Obligations, object)
			obligations = substituteObligations(obligations, substitution)
			if len(obligations) > 0 {
				ownerLabel := methodSetOwnerLabel(spec, substitution, name)
				for i := range obligations {
					if obligations[i].Owner == "" || !strings.Contains(obligations[i].Owner, "::") {
						obligations[i].Owner = ownerLabel
					}
				}
				method.Obligations = append(method.Obligations, obligations...)
			}
		}
		if !shouldBindSelfParam(method, object) {
			continue
		}
		method = bindMethodType(method)
		if !found || score > bestScore {
			bestScore = score
			bestFn = method
			found = true
		}
	}
	return bestFn, found
}

func (c *Checker) lookupUfcsFreeFunction(env *Environment, object Type, name string) (FunctionType, bool) {
	if env == nil {
		return FunctionType{}, false
	}
	typ, ok := env.Lookup(name)
	if !ok {
		return FunctionType{}, false
	}
	fnType, ok := typ.(FunctionType)
	if !ok {
		return FunctionType{}, false
	}
	if !shouldBindSelfParam(fnType, object) {
		return FunctionType{}, false
	}
	return bindMethodType(fnType), true
}

func (c *Checker) lookupTypeParamMethod(paramName, methodName string) (FunctionType, bool) {
	constraints := c.typeParamConstraints(paramName)
	for _, constraint := range constraints {
		if fnType, ok := c.methodFromConstraint(paramName, constraint, methodName); ok {
			return bindMethodType(fnType), true
		}
	}
	return FunctionType{}, false
}

func (c *Checker) methodFromConstraint(paramName string, constraint Type, methodName string) (FunctionType, bool) {
	res := c.resolveConstraintInterfaceType(constraint)
	if res.err != "" {
		return FunctionType{}, false
	}
	if res.iface.Methods == nil {
		return FunctionType{}, false
	}
	method, ok := res.iface.Methods[methodName]
	if !ok {
		return FunctionType{}, false
	}
	subst := make(map[string]Type)
	subst["Self"] = TypeParameterType{ParameterName: paramName}
	for idx, spec := range res.iface.TypeParams {
		var arg Type = UnknownType{}
		if idx < len(res.args) && res.args[idx] != nil {
			arg = res.args[idx]
		}
		subst[spec.Name] = arg
	}
	if len(subst) > 0 {
		method = substituteFunctionType(method, subst)
	}
	return method, true
}

func bindMethodType(method FunctionType) FunctionType {
	if len(method.Params) == 0 {
		return method
	}
	params := make([]Type, len(method.Params)-1)
	copy(params, method.Params[1:])
	return FunctionType{
		Params:      params,
		Return:      method.Return,
		TypeParams:  method.TypeParams,
		Where:       method.Where,
		Obligations: method.Obligations,
	}
}

func functionSignaturesEquivalent(a, b FunctionType) bool {
	if len(a.Params) != len(b.Params) {
		return false
	}
	for i := range a.Params {
		if !typesEquivalentForSignature(a.Params[i], b.Params[i]) {
			return false
		}
	}
	return true
}

func shouldBindSelfParam(method FunctionType, subject Type) bool {
	if len(method.Params) == 0 {
		return false
	}
	first := method.Params[0]
	if typesEquivalentForSignature(first, subject) {
		return true
	}
	switch v := first.(type) {
	case TypeParameterType:
		return v.ParameterName == "Self"
	default:
		if infoParam, ok := structInfoFromType(first); ok {
			if infoSubject, ok := structInfoFromType(subject); ok && infoParam.name != "" && infoParam.name == infoSubject.name {
				return true
			}
		}
	}
	if subject != nil && !isUnknownType(subject) {
		if typeAssignable(subject, first) {
			return true
		}
	}
	return false
}
