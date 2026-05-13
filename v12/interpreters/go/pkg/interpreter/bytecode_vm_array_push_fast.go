package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) appendTrackedArrayValueFast(arr *runtime.ArrayValue, state *runtime.ArrayState, value runtime.Value) int {
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
