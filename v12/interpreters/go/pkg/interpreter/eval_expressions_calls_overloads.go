package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) reportOverloadMismatch(fn *runtime.FunctionValue, evalArgs []runtime.Value, call *ast.FunctionCall) error {
	if fn == nil || fn.Declaration == nil {
		return nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil {
		return nil
	}
	params := decl.Params
	paramCount := len(params)
	expectedArgs := paramCount
	if decl.IsMethodShorthand {
		expectedArgs++
	}
	optionalLast := paramCount > 0 && isNullableParam(params[paramCount-1])
	paramsForCheck := params
	argsForCheck := evalArgs
	if decl.IsMethodShorthand && len(argsForCheck) > 0 {
		argsForCheck = argsForCheck[1:]
	}
	if optionalLast && len(evalArgs) == expectedArgs-1 && len(paramsForCheck) > 0 {
		paramsForCheck = paramsForCheck[:len(paramsForCheck)-1]
	}
	if len(paramsForCheck) != len(argsForCheck) {
		return nil
	}
	generics := functionGenericNameSet(fn, decl)
	for idx, param := range paramsForCheck {
		if param == nil || param.ParamType == nil {
			continue
		}
		if paramUsesGeneric(param.ParamType, generics) {
			continue
		}
		if i.matchesParamTypeForOverload(fn, param.ParamType, argsForCheck[idx]) {
			continue
		}
		name := fmt.Sprintf("param_%d", idx)
		if id, ok := param.Name.(*ast.Identifier); ok && id != nil {
			name = id.Name
		}
		expected := typeExpressionToString(param.ParamType)
		actual := describeRuntimeValue(argsForCheck[idx])
		return fmt.Errorf("Parameter type mismatch for '%s': expected %s, got %s", name, expected, actual)
	}
	return nil
}

func (i *Interpreter) matchesParamTypeForOverload(fn *runtime.FunctionValue, param ast.TypeExpression, value runtime.Value) bool {
	if param == nil {
		return true
	}
	if simple, ok := param.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == "Self" {
		if fn != nil && fn.MethodSet != nil && fn.MethodSet.TargetType != nil {
			if len(fn.MethodSet.GenericParams) > 0 {
				genericNames := genericNameSet(fn.MethodSet.GenericParams)
				if typeExpressionUsesGenerics(fn.MethodSet.TargetType, genericNames) {
					return true
				}
			}
			checkVal := value
			if iv, ok := value.(*runtime.InterfaceValue); ok && iv != nil {
				checkVal = iv.Underlying
			}
			return i.matchesType(fn.MethodSet.TargetType, checkVal)
		}
	}
	return i.matchesType(param, value)
}

func (i *Interpreter) matchesSingleRuntimeOverload(fn *runtime.FunctionValue, evalArgs []runtime.Value) bool {
	if fn == nil || fn.Declaration == nil {
		return false
	}
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		params := decl.Params
		paramCount := len(params)
		optionalLast := paramCount > 0 && isNullableParam(params[paramCount-1])
		expectedArgs := paramCount
		if decl.IsMethodShorthand {
			expectedArgs++
		}
		if !arityMatchesRuntime(expectedArgs, len(evalArgs), optionalLast) {
			return false
		}
		paramsForCheck := params
		argsForCheck := evalArgs
		if optionalLast && len(evalArgs) == expectedArgs-1 {
			paramsForCheck = params[:paramCount-1]
		}
		if decl.IsMethodShorthand && len(argsForCheck) > 0 {
			argsForCheck = argsForCheck[1:]
		}
		if len(argsForCheck) != len(paramsForCheck) {
			return false
		}
		generics := functionGenericNameSet(fn, decl)
		for idx, param := range paramsForCheck {
			if param == nil {
				return false
			}
			if param.ParamType == nil {
				continue
			}
			if paramUsesGeneric(param.ParamType, generics) {
				continue
			}
			if !i.matchesParamTypeForOverload(fn, param.ParamType, argsForCheck[idx]) {
				return false
			}
		}
		return true
	case *ast.LambdaExpression:
		return len(decl.Params) == len(evalArgs)
	default:
		return true
	}
}

type overloadCacheKey struct {
	firstOverload *runtime.FunctionValue // identity of overload set
	argCount      int
	argSignature  overloadArgSignature
}

const overloadArgSignatureInlineLimit = 4

type overloadArgKey struct {
	kind               runtime.Kind
	typeName           string
	underlyingKind     runtime.Kind
	underlyingTypeName string
}

type overloadArgSignature struct {
	inline   [overloadArgSignatureInlineLimit]overloadArgKey
	overflow string
}

func overloadArgTypeName(arg runtime.Value) string {
	switch v := arg.(type) {
	case runtime.IntegerValue:
		return string(v.TypeSuffix)
	case *runtime.IntegerValue:
		if v == nil {
			return "nil"
		}
		return string(v.TypeSuffix)
	case runtime.FloatValue:
		return string(v.TypeSuffix)
	case *runtime.FloatValue:
		if v == nil {
			return "nil"
		}
		return string(v.TypeSuffix)
	case *runtime.StructInstanceValue:
		if v == nil {
			return "nil"
		}
		if v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil {
			return v.Definition.Node.ID.Name
		}
	case *runtime.HostHandleValue:
		if v == nil {
			return "nil"
		}
		if v.HandleType != "" {
			return v.HandleType
		}
		return "host_handle"
	case runtime.TypeRefValue:
		return v.TypeName
	case *runtime.TypeRefValue:
		if v == nil {
			return "nil"
		}
		return v.TypeName
	}
	return ""
}

func overloadArgKeyForValue(arg runtime.Value) overloadArgKey {
	if arg == nil {
		return overloadArgKey{typeName: "nil"}
	}
	key := overloadArgKey{
		kind:     arg.Kind(),
		typeName: overloadArgTypeName(arg),
	}
	if iface, ok := arg.(*runtime.InterfaceValue); ok && iface != nil && iface.Underlying != nil {
		key.underlyingKind = iface.Underlying.Kind()
		key.underlyingTypeName = overloadArgTypeName(iface.Underlying)
	}
	return key
}

func overloadArgKindsSlow(args []runtime.Value) string {
	if len(args) == 0 {
		return ""
	}
	buf := make([]byte, 0, len(args)*8)
	for _, arg := range args {
		if arg == nil {
			buf = append(buf, '?')
			continue
		}
		kind := arg.Kind().String()
		buf = append(buf, byte(len(kind)))
		buf = append(buf, kind...)
		// For struct instances, include the definition name to differentiate types.
		if inst, ok := arg.(*runtime.StructInstanceValue); ok && inst != nil && inst.Definition != nil && inst.Definition.Node != nil && inst.Definition.Node.ID != nil {
			name := inst.Definition.Node.ID.Name
			buf = append(buf, ':')
			buf = append(buf, name...)
		}
		// For interface values, include the underlying type.
		if iface, ok := arg.(*runtime.InterfaceValue); ok && iface != nil && iface.Underlying != nil {
			buf = append(buf, ':')
			buf = append(buf, iface.Underlying.Kind().String()...)
		}
		// For integer values, include the type suffix.
		if intVal, ok := arg.(runtime.IntegerValue); ok {
			buf = append(buf, ':')
			buf = append(buf, string(intVal.TypeSuffix)...)
		}
		// For float values, include the type suffix as well.
		if floatVal, ok := arg.(runtime.FloatValue); ok {
			buf = append(buf, ':')
			buf = append(buf, string(floatVal.TypeSuffix)...)
		}
	}
	return string(buf)
}

func overloadArgSignatureForValues(args []runtime.Value) overloadArgSignature {
	var sig overloadArgSignature
	if len(args) > overloadArgSignatureInlineLimit {
		sig.overflow = overloadArgKindsSlow(args)
		return sig
	}
	for idx, arg := range args {
		sig.inline[idx] = overloadArgKeyForValue(arg)
	}
	return sig
}

func overloadsHaveGenerics(overloads []*runtime.FunctionValue) bool {
	for _, fn := range overloads {
		if fn == nil || fn.Declaration == nil {
			continue
		}
		if decl, ok := fn.Declaration.(*ast.FunctionDefinition); ok && len(decl.GenericParams) > 0 {
			return true
		}
	}
	return false
}

func (i *Interpreter) selectRuntimeOverload(overloads []*runtime.FunctionValue, evalArgs []runtime.Value, call *ast.FunctionCall) (*runtime.FunctionValue, error) {
	// Cache lookup: use first overload pointer as set identity.
	// Only cache when no overload has generic parameters (generics need full type checking).
	var cacheKey overloadCacheKey
	useCache := len(overloads) > 1 && len(overloads) <= 32 && !overloadsHaveGenerics(overloads)
	if useCache {
		cacheKey = overloadCacheKey{
			firstOverload: overloads[0],
			argCount:      len(evalArgs),
			argSignature:  overloadArgSignatureForValues(evalArgs),
		}
		if cached, ok := i.overloadCache[cacheKey]; ok {
			return cached, nil
		}
	}

	type candidate struct {
		fn       *runtime.FunctionValue
		score    float64
		priority float64
	}
	candidates := make([]candidate, 0, len(overloads))
	for _, fn := range overloads {
		if fn == nil || fn.Declaration == nil {
			continue
		}
		switch decl := fn.Declaration.(type) {
		case *ast.FunctionDefinition:
			params := decl.Params
			paramCount := len(params)
			optionalLast := paramCount > 0 && isNullableParam(params[paramCount-1])
			expectedArgs := paramCount
			if decl.IsMethodShorthand {
				expectedArgs++
			}
			if !arityMatchesRuntime(expectedArgs, len(evalArgs), optionalLast) {
				continue
			}
			paramsForCheck := params
			argsForCheck := evalArgs
			if optionalLast && len(evalArgs) == expectedArgs-1 {
				paramsForCheck = params[:paramCount-1]
			}
			if decl.IsMethodShorthand && len(argsForCheck) > 0 {
				argsForCheck = argsForCheck[1:]
			}
			if len(argsForCheck) != len(paramsForCheck) {
				continue
			}
			generics := functionGenericNameSet(fn, decl)
			score := 0.0
			if optionalLast && len(evalArgs) == expectedArgs-1 {
				score -= 0.5
			}
			compatible := true
			for idx, param := range paramsForCheck {
				if param == nil {
					compatible = false
					break
				}
				if param.ParamType != nil {
					if paramUsesGeneric(param.ParamType, generics) {
						continue
					}
					if !i.matchesParamTypeForOverload(fn, param.ParamType, argsForCheck[idx]) {
						compatible = false
						break
					}
					score += float64(parameterSpecificity(param.ParamType, generics))
				}
			}
			if compatible {
				candidates = append(candidates, candidate{fn: fn, score: score, priority: fn.MethodPriority})
			}
		case *ast.LambdaExpression:
			if len(decl.Params) != len(evalArgs) {
				continue
			}
			candidates = append(candidates, candidate{fn: fn, score: 0, priority: fn.MethodPriority})
		default:
			candidates = append(candidates, candidate{fn: fn, score: 0, priority: fn.MethodPriority})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	best := candidates[0]
	ties := []candidate{best}
	for _, cand := range candidates[1:] {
		if cand.score > best.score || (cand.score == best.score && cand.priority > best.priority) {
			best = cand
			ties = []candidate{cand}
		} else if cand.score == best.score && cand.priority == best.priority {
			ties = append(ties, cand)
		}
	}
	if len(ties) > 1 {
		return nil, fmt.Errorf("Ambiguous overload for %s", overloadName(call))
	}
	if useCache {
		i.overloadCache[cacheKey] = best.fn
	}
	return best.fn, nil
}
