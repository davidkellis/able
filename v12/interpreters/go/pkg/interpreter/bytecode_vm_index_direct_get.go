package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) resolveDirectArrayIndexGetAt(arr *runtime.ArrayValue, idx int) (runtime.Value, error) {
	if vm == nil || vm.interp == nil || arr == nil {
		return nil, nil
	}
	value, _, _, err := vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, 0)
	return value, err
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetAtReady(arr *runtime.ArrayValue, idx int) (runtime.Value, error) {
	return vm.resolveDirectArrayIndexGetAtReadyWithHandle(arr, idx, 0)
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetAtReadyWithHandle(arr *runtime.ArrayValue, idx int, handle int64) (runtime.Value, error) {
	value, _, _, err := vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
	return value, err
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr *runtime.ArrayValue, idx int, handle int64) (runtime.Value, uint16, bool, error) {
	if handle != 0 && arr != nil && arr.State == nil {
		return vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, idx)
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if tracked {
		if idx < 0 || idx >= len(state.Values) {
			return vm.interp.makeIndexErrorValue(idx, len(state.Values)), state.ElementTypeToken, state.ElementTypeTokenKnown, nil
		}
		val := state.Values[idx]
		if val == nil {
			return vm.interp.makeIndexErrorValue(idx, len(state.Values)), state.ElementTypeToken, state.ElementTypeTokenKnown, nil
		}
		return val, state.ElementTypeToken, state.ElementTypeTokenKnown, nil
	}
	if handle == 0 {
		handle = bytecodeArrayStorageHandle(arr)
		if handle == 0 {
			var ok bool
			var err error
			handle, ok, err = vm.arrayHandleFast(arr)
			if err != nil {
				return nil, bytecodeIndexTypeUnknown, false, err
			}
			if !ok {
				return nil, bytecodeIndexTypeUnknown, false, nil
			}
		}
	}
	return vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, idx)
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetAtHandleReady(handle int64, idx int) (runtime.Value, error) {
	value, _, _, err := vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, idx)
	return value, err
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetAtHandleReadyWithToken(handle int64, idx int) (runtime.Value, uint16, bool, error) {
	var info runtime.ArrayStoreMonoPrimitiveReadInfo
	if ok, err := runtime.ArrayStoreMonoPrimitiveReadInfoIntoFresh(handle, idx, &info); err != nil {
		return nil, bytecodeIndexTypeUnknown, false, err
	} else if ok {
		if !info.InBounds {
			token, tokenKnown := bytecodeMonoPrimitiveArrayToken(info.Kind)
			return vm.interp.makeIndexErrorValue(idx, info.Size), token, tokenKnown, nil
		}
		switch info.Kind {
		case runtime.ArrayStoreMonoPrimitiveReadI32:
			return bytecodeRawI32ResultValue(info.Int64), bytecodeIndexTypeI32, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadI64:
			return runtime.NewSmallInt(info.Int64, runtime.IntegerI64), bytecodeIndexTypeI64, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadBool:
			return runtime.BoolValue{Val: info.Bool}, bytecodeIndexTypeBool, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadChar:
			return runtime.CharValue{Val: rune(info.Int64)}, bytecodeIndexTypeChar, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadU8:
			return runtime.NewSmallInt(int64(info.Uint64), runtime.IntegerU8), bytecodeIndexTypeU8, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadU32:
			return bytecodeRawIntegerResultValue(runtime.IntegerU32, int64(uint32(info.Uint64))), bytecodeIndexTypeU32, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadU64:
			return bytecodeUnsignedIntegerValue(runtime.IntegerU64, info.Uint64), bytecodeIndexTypeU64, true, nil
		case runtime.ArrayStoreMonoPrimitiveReadF64:
			return runtime.FloatValue{Val: info.Float64, TypeSuffix: runtime.FloatF64}, bytecodeIndexTypeF64, true, nil
		}
		if result, mono := bytecodeMonoPrimitiveArrayValue(info); mono {
			token, tokenKnown := bytecodeMonoPrimitiveArrayToken(info.Kind)
			return result, token, tokenKnown, nil
		}
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		return nil, bytecodeIndexTypeUnknown, false, err
	}
	if idx < 0 || idx >= size {
		return vm.interp.makeIndexErrorValue(idx, size), bytecodeIndexTypeUnknown, false, nil
	}
	val, err := runtime.ArrayStoreRead(handle, idx)
	if err != nil {
		return nil, bytecodeIndexTypeUnknown, false, err
	}
	if val == nil {
		return vm.interp.makeIndexErrorValue(idx, size), bytecodeIndexTypeUnknown, false, nil
	}
	token, tokenKnown := bytecodeIndexValueTypeToken(val)
	return val, token, tokenKnown, nil
}

func (vm *bytecodeVM) resolveDirectArrayIndexGet(arr *runtime.ArrayValue, idxVal runtime.Value) (runtime.Value, bool, error) {
	return vm.resolveDirectArrayIndexGetWithHandle(arr, idxVal, 0)
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetWithHandle(arr *runtime.ArrayValue, idxVal runtime.Value, handle int64) (runtime.Value, bool, error) {
	value, _, _, handled, err := vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, handle)
	return value, handled, err
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetWithValidatedHandleAndToken(arr *runtime.ArrayValue, idxVal runtime.Value, handle int64) (runtime.Value, uint16, bool, bool, error) {
	if handle != 0 && arr != nil && arr.State == nil {
		if idx, ok := bytecodeDirectSmallArrayIndex(idxVal); ok {
			value, token, tokenKnown, err := vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, idx)
			return value, token, tokenKnown, true, err
		}
		if idx, ok, err := bytecodeDirectArrayIndex(idxVal); err != nil || ok {
			if err != nil {
				return nil, bytecodeIndexTypeUnknown, false, true, err
			}
			value, token, tokenKnown, err := vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, idx)
			return value, token, tokenKnown, true, err
		}
	}
	return vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, handle)
}

func (vm *bytecodeVM) resolveDirectArrayIndexGetWithHandleAndToken(arr *runtime.ArrayValue, idxVal runtime.Value, handle int64) (runtime.Value, uint16, bool, bool, error) {
	if vm == nil || vm.interp == nil || arr == nil {
		return nil, bytecodeIndexTypeUnknown, false, false, nil
	}
	if idx, ok := bytecodeDirectRawI32ArrayIndex(idxVal); ok {
		value, token, tokenKnown, err := vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
		return value, token, tokenKnown, true, err
	}
	idx, ok, err := bytecodeDirectArrayIndex(idxVal)
	if err != nil || !ok {
		return nil, bytecodeIndexTypeUnknown, false, ok, err
	}
	value, token, tokenKnown, err := vm.resolveDirectArrayIndexGetAtReadyWithHandleAndToken(arr, idx, handle)
	return value, token, tokenKnown, true, err
}

func bytecodeDirectRawI32ArrayIndex(idxVal runtime.Value) (int, bool) {
	switch idx := idxVal.(type) {
	case bytecodeRawI32SlotValue:
		return int(idx), true
	case *bytecodeRawI32StackCell:
		if idx != nil {
			return int(idx.Val), true
		}
	}
	return 0, false
}
