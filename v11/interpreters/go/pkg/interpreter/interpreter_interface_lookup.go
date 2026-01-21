package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) lookupImplEntry(info typeInfo, interfaceName string) (*implCandidate, error) {
	matches, err := i.collectImplCandidates(info, interfaceName, "")
	if len(matches) == 0 {
		return nil, err
	}
	best, ambiguous := i.selectBestCandidate(matches)
	if ambiguous != nil {
		detail := descriptionsFromCandidates(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("ambiguous implementations of %s for %s: %s", interfaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	return best, nil
}

func (i *Interpreter) findMethod(info typeInfo, methodName string, interfaceFilter string) (runtime.Value, error) {
	var matches []implCandidate
	var err error
	if interfaceFilter == "" {
		matches, err = i.collectImplCandidates(info, "", methodName)
	} else {
		names := i.interfaceSearchNames(interfaceFilter, make(map[string]struct{}))
		if len(names) == 0 {
			names = []string{interfaceFilter}
		}
		var constraintErr error
		for _, name := range names {
			candidates, candErr := i.collectImplCandidates(info, name, methodName)
			if candErr != nil && constraintErr == nil {
				constraintErr = candErr
			}
			if len(candidates) > 0 {
				matches = append(matches, candidates...)
			}
		}
		if len(matches) == 0 {
			return nil, constraintErr
		}
	}
	if len(matches) == 0 {
		return nil, err
	}
	if interfaceFilter != "" {
		direct := make([]implCandidate, 0, len(matches))
		for _, cand := range matches {
			if cand.entry != nil && cand.entry.interfaceName == interfaceFilter {
				direct = append(direct, cand)
			}
		}
		if len(direct) > 0 {
			matches = direct
		}
	}
	methodMatches := make([]methodMatch, 0, len(matches))
	for _, cand := range matches {
		method := cand.entry.methods[methodName]
		if method == nil {
			if ifaceDef, ok := i.interfaces[cand.entry.interfaceName]; ok && ifaceDef.Node != nil {
				for _, sig := range ifaceDef.Node.Signatures {
					if sig == nil || sig.Name == nil || sig.Name.Name != methodName || sig.DefaultImpl == nil {
						continue
					}
					defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
					method = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env, MethodPriority: -1}
					if cand.entry.methods == nil {
						cand.entry.methods = make(map[string]runtime.Value)
					}
					mergeFunctionLike(cand.entry.methods, methodName, method)
					break
				}
			}
		}
		if method == nil {
			continue
		}
		methodMatches = append(methodMatches, methodMatch{candidate: cand, method: method})
	}
	if len(methodMatches) == 0 {
		return nil, err
	}
	if len(methodMatches) > 1 {
		explicit := make([]methodMatch, 0, len(methodMatches))
		for _, match := range methodMatches {
			if implDefinesMethod(match.candidate.entry, methodName) {
				explicit = append(explicit, match)
			}
		}
		if len(explicit) > 0 {
			methodMatches = explicit
		}
	}
	best, ambiguous := i.selectBestMethodCandidate(methodMatches)
	if ambiguous != nil {
		detail := descriptionsFromMethodMatches(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		ifaceName := methodName
		if len(ambiguous) > 0 && ambiguous[0].candidate.entry != nil && ambiguous[0].candidate.entry.interfaceName != "" {
			ifaceName = ambiguous[0].candidate.entry.interfaceName
		}
		if len(detail) == 0 {
			detail = []string{"<unknown>"}
		}
		return nil, fmt.Errorf("ambiguous implementations of %s for %s: %s", ifaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	if fnVal := firstFunction(best.method); fnVal != nil {
		if fnDef, ok := fnVal.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", methodName, info.name)
		}
	}
	return best.method, nil
}

func implDefinesMethod(entry *implEntry, methodName string) bool {
	if entry == nil || entry.definition == nil || methodName == "" {
		return false
	}
	for _, fn := range entry.definition.Definitions {
		if fn == nil || fn.ID == nil {
			continue
		}
		if fn.ID.Name == methodName {
			return true
		}
	}
	return false
}

func (i *Interpreter) interfaceSearchNames(interfaceName string, visited map[string]struct{}) []string {
	if interfaceName == "" {
		return nil
	}
	if _, seen := visited[interfaceName]; seen {
		return nil
	}
	visited[interfaceName] = struct{}{}
	names := []string{interfaceName}
	ifaceDef, ok := i.interfaces[interfaceName]
	if !ok || ifaceDef == nil || ifaceDef.Node == nil {
		return names
	}
	for _, base := range ifaceDef.Node.BaseInterfaces {
		info, ok := parseTypeExpression(base)
		if !ok || info.name == "" {
			continue
		}
		names = append(names, i.interfaceSearchNames(info.name, visited)...)
	}
	return names
}

func (i *Interpreter) typeImplementsInterface(info typeInfo, interfaceName string, visited map[string]struct{}) (bool, error) {
	if info.name == "" || interfaceName == "" {
		return false, nil
	}
	if interfaceName == "Error" && info.name == "Error" {
		return true, nil
	}
	key := interfaceName + "::" + typeInfoToString(info)
	if _, seen := visited[key]; seen {
		return true, nil
	}
	visited[key] = struct{}{}
	ifaceDef, ok := i.interfaces[interfaceName]
	if ok && ifaceDef != nil && ifaceDef.Node != nil && len(ifaceDef.Node.BaseInterfaces) > 0 {
		for _, base := range ifaceDef.Node.BaseInterfaces {
			baseInfo, ok := parseTypeExpression(base)
			if !ok || baseInfo.name == "" {
				return false, nil
			}
			okImpl, err := i.typeImplementsInterface(info, baseInfo.name, visited)
			if err != nil || !okImpl {
				return okImpl, err
			}
		}
		if len(ifaceDef.Node.Signatures) == 0 {
			return true, nil
		}
	}
	entry, err := i.lookupImplEntry(info, interfaceName)
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}

func (i *Interpreter) interfaceMatches(val *runtime.InterfaceValue, interfaceName string) bool {
	if val == nil {
		return false
	}
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		if val.Interface.Node.ID.Name == interfaceName {
			return true
		}
	}
	info, ok := i.getTypeInfoForValue(val.Underlying)
	if !ok {
		return false
	}
	okImpl, err := i.typeImplementsInterface(info, interfaceName, make(map[string]struct{}))
	return err == nil && okImpl
}

func (i *Interpreter) selectStructMethod(inst *runtime.StructInstanceValue, methodName string) (runtime.Value, error) {
	if inst == nil {
		return nil, nil
	}
	info, ok := i.typeInfoFromStructInstance(inst)
	if !ok {
		return nil, nil
	}
	return i.findMethod(info, methodName, "")
}
