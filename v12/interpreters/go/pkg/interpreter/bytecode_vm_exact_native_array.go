package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeArrayHandleI64(val runtime.Value) (int64, bool) {
	intVal, ok := bytecodeIntegerValue(val)
	if !ok {
		return 0, false
	}
	if intVal.IsSmall() {
		return intVal.Int64Fast(), true
	}
	handle, ok := intVal.ToInt64()
	return handle, ok
}

func (vm *bytecodeVM) tryExecExactNativeArrayReadFast(name string, explicitArgs []runtime.Value) (runtime.Value, bool, error) {
	if vm == nil || name != "__able_array_read" || len(explicitArgs) != 2 {
		return nil, false, nil
	}
	handle, ok := bytecodeArrayHandleI64(explicitArgs[0])
	if !ok || handle == 0 {
		return nil, false, nil
	}
	idx, ok, err := bytecodeArraySlotIndexI32(explicitArgs[1])
	if err != nil {
		return nil, true, err
	}
	if !ok {
		return nil, false, nil
	}
	var info runtime.ArrayStoreMonoPrimitiveReadInfo
	if ok, err := runtime.ArrayStoreMonoPrimitiveReadInfoInto(handle, idx, &info); err != nil {
		return nil, true, err
	} else if ok {
		if !info.InBounds {
			return runtime.NilValue{}, true, nil
		}
		if result, mono := bytecodeMonoPrimitiveArrayValue(info); mono {
			return result, true, nil
		}
	}
	result, err := runtime.ArrayStoreRead(handle, idx)
	return result, true, err
}
