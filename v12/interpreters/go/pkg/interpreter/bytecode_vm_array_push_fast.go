package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) appendTrackedArrayValueFast(arr *runtime.ArrayValue, state *runtime.ArrayState, value runtime.Value) int {
	value = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, value)
	idx := len(state.Values)
	if idx+1 > state.Capacity || idx == cap(state.Values) {
		runtime.ArrayEnsureCapacity(state, idx+1)
	}
	state.Values = append(state.Values, value)
	if state.Capacity < cap(state.Values) {
		state.Capacity = cap(state.Values)
	}
	if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
		vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
	}
	return idx
}

func (vm *bytecodeVM) appendArrayCharValueFast(arr *runtime.ArrayValue, value rune) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeChar && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendCharIfMono(arr.Handle, value)
	if err == nil && !ok {
		ok, err = runtime.ArrayStoreAppendCharPromote(arr.Handle, value)
	}
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}

func (vm *bytecodeVM) appendArrayU8ValueFast(arr *runtime.ArrayValue, value uint8) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeU8 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendU8Promote(arr.Handle, value)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}

func (vm *bytecodeVM) appendArrayU32ValueFast(arr *runtime.ArrayValue, value uint32) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeU32 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendU32Promote(arr.Handle, value)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}

func (vm *bytecodeVM) appendArrayU64ValueFast(arr *runtime.ArrayValue, value uint64) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeU64 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendU64Promote(arr.Handle, value)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}

func (vm *bytecodeVM) appendArrayU8BytesFast(arr *runtime.ArrayValue, value []byte) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeU8 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendU8BytesPromote(arr.Handle, value)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}

func (vm *bytecodeVM) appendArrayU8StringFast(arr *runtime.ArrayValue, value string) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeU8 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendU8StringPromote(arr.Handle, value)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}
