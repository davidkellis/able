package interpreter

import "able/interpreter-go/pkg/runtime"

func (i *Interpreter) propagationErrorValue(val runtime.Value, env *runtime.Environment) (runtime.ErrorValue, bool) {
	if errVal, ok := asErrorValue(val); ok {
		return errVal, true
	}
	if !i.propagationValueMayImplementError(val) {
		return runtime.ErrorValue{}, false
	}
	if i.matchesType(cachedSimpleTypeExpression("Error"), val) {
		return i.makeErrorValue(val, env), true
	}
	return runtime.ErrorValue{}, false
}

func (i *Interpreter) propagationValueMayImplementError(val runtime.Value) bool {
	if i == nil || val == nil {
		return false
	}
	switch v := val.(type) {
	case runtime.ErrorValue:
		return true
	case *runtime.ErrorValue:
		return v != nil
	case runtime.InterfaceValue:
		return i.interfaceValueMayImplementError(&v)
	case *runtime.InterfaceValue:
		return i.interfaceValueMayImplementError(v)
	case *runtime.StructInstanceValue:
		return v != nil
	case runtime.StructDefinitionValue:
		return true
	case *runtime.StructDefinitionValue:
		return v != nil
	case runtime.StringValue:
		return i.typeNameMayImplementError("String")
	case *runtime.StringValue:
		return v != nil && i.typeNameMayImplementError("String")
	case runtime.BoolValue:
		return i.typeNameMayImplementError("bool")
	case *runtime.BoolValue:
		return v != nil && i.typeNameMayImplementError("bool")
	case runtime.CharValue:
		return i.typeNameMayImplementError("char")
	case *runtime.CharValue:
		return v != nil && i.typeNameMayImplementError("char")
	case runtime.NilValue:
		return i.typeNameMayImplementError("nil")
	case *runtime.NilValue:
		return v != nil && i.typeNameMayImplementError("nil")
	case runtime.VoidValue:
		return i.typeNameMayImplementError("void")
	case *runtime.VoidValue:
		return v != nil && i.typeNameMayImplementError("void")
	case runtime.IntegerValue:
		return i.typeNameMayImplementError(string(v.TypeSuffix))
	case *runtime.IntegerValue:
		return v != nil && i.typeNameMayImplementError(string(v.TypeSuffix))
	case runtime.FloatValue:
		return i.typeNameMayImplementError(string(v.TypeSuffix))
	case *runtime.FloatValue:
		return v != nil && i.typeNameMayImplementError(string(v.TypeSuffix))
	case *runtime.ArrayValue:
		return v != nil && i.typeNameMayImplementError("Array")
	case *runtime.HashMapValue:
		return v != nil && i.typeNameMayImplementError("HashMap")
	case *runtime.IteratorValue:
		return v != nil && i.typeNameMayImplementError("Iterator")
	case runtime.IteratorEndValue:
		return i.typeNameMayImplementError("IteratorEnd")
	case *runtime.IteratorEndValue:
		return v != nil && i.typeNameMayImplementError("IteratorEnd")
	case *runtime.FutureValue:
		return v != nil && i.typeNameMayImplementError("Future")
	default:
		return false
	}
}

func (i *Interpreter) interfaceValueMayImplementError(val *runtime.InterfaceValue) bool {
	if i == nil || val == nil {
		return false
	}
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		if val.Interface.Node.ID.Name == "Error" {
			return true
		}
	}
	if val.Underlying != nil {
		return i.propagationValueMayImplementError(val.Underlying)
	}
	return true
}

func (i *Interpreter) typeNameMayImplementError(typeName string) bool {
	if i == nil || typeName == "" {
		return false
	}
	typeName = normalizeKernelAliasName(typeName)
	if typeName == "Error" {
		return true
	}
	i.methodCacheMu.RLock()
	if cached, ok := i.propagationErrorCache[typeName]; ok {
		i.methodCacheMu.RUnlock()
		return cached
	}
	i.methodCacheMu.RUnlock()

	mayImplement := false
	if i.compiledImplChecker != nil && i.compiledImplChecker(typeName, "Error") {
		mayImplement = true
	}
	if !mayImplement {
		for _, entry := range i.implMethods[typeName] {
			if entry.interfaceName == "Error" {
				mayImplement = true
				break
			}
		}
	}
	if !mayImplement {
		for _, entry := range i.genericImpls {
			if entry.interfaceName == "Error" {
				mayImplement = true
				break
			}
		}
	}

	i.methodCacheMu.Lock()
	if i.propagationErrorCache == nil {
		i.propagationErrorCache = make(map[string]bool)
	}
	i.propagationErrorCache[typeName] = mayImplement
	i.methodCacheMu.Unlock()
	return mayImplement
}
