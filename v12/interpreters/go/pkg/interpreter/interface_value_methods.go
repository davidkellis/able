package interpreter

import "able/interpreter-go/pkg/runtime"

func interfaceValueLookupLocalMethod(val *runtime.InterfaceValue, name string) (runtime.Value, bool) {
	if val == nil || name == "" || val.Methods == nil {
		return nil, false
	}
	method := val.Methods[name]
	return method, method != nil
}

func interfaceValueLookupBoundMethod(val *runtime.InterfaceValue, name string) (runtime.Value, bool) {
	if val == nil || name == "" || val.BoundMethodName != name || val.BoundMethod == nil {
		return nil, false
	}
	return val.BoundMethod, true
}

func interfaceValueLookupMethod(val *runtime.InterfaceValue, name string) (runtime.Value, bool) {
	if method, ok := interfaceValueLookupLocalMethod(val, name); ok {
		return method, true
	}
	if val == nil || name == "" {
		return nil, false
	}
	if val.SharedMethods != nil {
		if method := val.SharedMethods[name]; method != nil {
			return method, true
		}
	}
	return nil, false
}

func ensureOwnedInterfaceValueMethods(val *runtime.InterfaceValue, extraCapacity int) map[string]runtime.Value {
	if val == nil {
		return nil
	}
	if val.Methods == nil {
		size := extraCapacity
		if size < 1 {
			size = 1
		}
		// Methods is a per-interface-value overlay that can shadow SharedMethods
		// without copying the shared dictionary.
		val.Methods = make(map[string]runtime.Value, size)
	}
	return val.Methods
}

func interfaceValueSetMethod(val *runtime.InterfaceValue, name string, method runtime.Value) {
	if val == nil || name == "" || method == nil {
		return
	}
	val.BoundMethodName = ""
	val.BoundMethod = nil
	ensureOwnedInterfaceValueMethods(val, 1)[name] = method
}

func interfaceValueSetBoundMethod(val *runtime.InterfaceValue, name string, method runtime.Value) {
	if val == nil || name == "" || method == nil {
		return
	}
	val.BoundMethodName = name
	val.BoundMethod = method
}

func interfaceValueMethodIsBound(method runtime.Value) bool {
	switch method.(type) {
	case runtime.BoundMethodValue,
		*runtime.BoundMethodValue,
		runtime.NativeBoundMethodValue,
		*runtime.NativeBoundMethodValue:
		return true
	default:
		return false
	}
}
