package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execArrayLenMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	arr, ok := vm.stackValue(receiverIndex).(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	size, ok, err := vm.arraySizeI32Fast(arr)
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !ok {
		return nil, false, nil
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_len_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(boxedOrSmallIntegerValue(runtime.IntegerI32, int64(size)), nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execArrayGetMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	arr, ok := vm.stackValue(receiverIndex).(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	idx, ok := bytecodeArrayGetIndexI32(vm.stackValue(argBase))
	if !ok {
		return nil, false, nil
	}
	return vm.finishArrayGetMemberFast(instr, arr, idx, receiverIndex, callNode)
}

func (vm *bytecodeVM) finishArrayGetMemberFast(instr bytecodeInstruction, arr *runtime.ArrayValue, idx int64, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if arr == nil || receiverIndex < 0 || receiverIndex >= vm.stackDepth() {
		return nil, false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		size := len(state.Values)
		if size > 1<<31-1 {
			return nil, false, nil
		}
		var result runtime.Value
		if idx < 0 || idx >= int64(size) {
			result = runtime.NilValue{}
		} else {
			result = state.Values[int(idx)]
		}
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_get_tracked_fast", instr.node)
		}
		vm.truncateStack(receiverIndex)
		if vm.canSkipArrayGetSuccessPropagation(result, state.ElementTypeToken, state.ElementTypeTokenKnown) {
			vm.appendStackValue(result)
			vm.ip += 2
			return nil, true, nil
		}
		newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
		return newProg, true, finishErr
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !ok {
		return nil, false, nil
	}
	if values, monoF64, monoErr := runtime.ArrayStoreMonoF64ValuesIfAvailable(handle); monoErr != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, monoErr, callNode, nil)
		return newProg, true, finishErr
	} else if monoF64 {
		if len(values) > 1<<31-1 {
			return nil, false, nil
		}
		var result runtime.Value
		if idx < 0 || idx >= int64(len(values)) {
			result = runtime.NilValue{}
		} else {
			result = runtime.FloatValue{Val: values[int(idx)], TypeSuffix: runtime.FloatF64}
		}
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_get_mono_f64_fast", instr.node)
		}
		vm.truncateStack(receiverIndex)
		if vm.canSkipArrayGetSuccessPropagation(result, bytecodeIndexTypeF64, true) {
			vm.appendStackValue(result)
			vm.ip += 2
			return nil, true, nil
		}
		newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
		return newProg, true, finishErr
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if size < 0 || size > 1<<31-1 {
		return nil, false, nil
	}
	var result runtime.Value
	if idx < 0 || idx >= int64(size) {
		result = runtime.NilValue{}
	} else {
		var info runtime.ArrayStoreMonoPrimitiveReadInfo
		if ok, monoErr := runtime.ArrayStoreMonoPrimitiveReadInfoInto(handle, int(idx), &info); monoErr != nil {
			vm.truncateStack(receiverIndex)
			newProg, finishErr := vm.finishCompletedCall(nil, monoErr, callNode, nil)
			return newProg, true, finishErr
		} else if ok {
			if result, mono := bytecodeMonoPrimitiveArrayValue(info); mono {
				if vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", bytecodeMonoPrimitiveArrayDispatch(info, "array_get"), instr.node)
				}
				vm.truncateStack(receiverIndex)
				newProg, finishErr := vm.finishCompletedCall(result, nil, callNode, nil)
				return newProg, true, finishErr
			}
		}
		result, err = runtime.ArrayStoreRead(handle, int(idx))
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_get_fast", instr.node)
	}
	vm.truncateStack(receiverIndex)
	token, tokenKnown := vm.arrayElementTypeTokenForPropagation(arr)
	if err == nil && vm.canSkipArrayGetSuccessPropagation(result, token, tokenKnown) {
		vm.appendStackValue(result)
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(result, err, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execArrayPushMemberFast(memberName string, argCount int, traceNode ast.Node, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	arr, ok := vm.stackValue(receiverIndex).(*runtime.ArrayValue)
	if !ok || arr == nil || vm == nil || vm.interp == nil {
		return nil, false, nil
	}
	value := vm.stackValue(argBase)
	vm.observeBytecodeArrayOwnershipArrayWrite(arr, value)
	finishFast := func(dispatch string) (*bytecodeProgram, bool, error) {
		if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
			vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", dispatch, traceNode)
		}
		vm.truncateStack(receiverIndex)
		if vm.canSkipFollowingPop(nil) {
			vm.ip += 2
			return nil, true, nil
		}
		newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
		return newProg, true, finishErr
	}
	switch typed := value.(type) {
	case runtime.CharValue:
		if vm.appendArrayCharValueFast(arr, typed.Val) {
			return finishFast("array_push_char_mono_fast")
		}
	case *runtime.CharValue:
		if typed != nil && vm.appendArrayCharValueFast(arr, typed.Val) {
			return finishFast("array_push_char_mono_fast")
		}
	}
	if raw, ok := bytecodeU8ValueFast(value); ok && vm.appendArrayU8ValueFast(arr, raw) {
		return finishFast("array_push_u8_mono_fast")
	}
	if raw, ok := bytecodeU32ValueFast(value); ok && vm.appendArrayU32ValueFast(arr, raw) {
		return finishFast("array_push_u32_mono_fast")
	}
	if raw, ok := bytecodeU64ValueFast(value); ok && vm.appendArrayU64ValueFast(arr, raw) {
		return finishFast("array_push_u64_mono_fast")
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked && !arr.TrackedAliases {
		handle, ok, err := vm.arrayHandleFast(arr)
		if err != nil {
			vm.truncateStack(receiverIndex)
			newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
			return newProg, true, finishErr
		}
		if ok {
			size, err := runtime.ArrayStoreSize(handle)
			if err != nil {
				vm.truncateStack(receiverIndex)
				newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
				return newProg, true, finishErr
			}
			storedValue := vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, value)
			if err := runtime.ArrayStoreWrite(handle, size, storedValue); err != nil {
				vm.truncateStack(receiverIndex)
				newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
				return newProg, true, finishErr
			}
			vm.interp.syncArrayHandleWriteAfterStore(handle, size, storedValue)
			if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
				vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "array_push_handle_fast", traceNode)
			}
			vm.truncateStack(receiverIndex)
			if vm.canSkipFollowingPop(nil) {
				vm.ip += 2
				return nil, true, nil
			}
			newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
			return newProg, true, finishErr
		}
	}
	if !tracked {
		var err error
		state, err = vm.interp.ensureArrayState(arr, 0)
		if err != nil {
			vm.truncateStack(receiverIndex)
			newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
			return newProg, true, finishErr
		}
	}
	vm.appendTrackedArrayValueFast(arr, state, value)
	if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "array_push_fast", traceNode)
	}
	vm.truncateStack(receiverIndex)
	if vm.canSkipFollowingPop(nil) {
		vm.ip += 2
		return nil, true, nil
	}
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) canSkipFollowingPop(program *bytecodeProgram) bool {
	if vm == nil {
		return false
	}
	if program == nil {
		program = vm.currentProgram
	}
	nextIP := vm.ip + 1
	return program != nil &&
		nextIP >= 0 &&
		nextIP < len(program.instructions) &&
		program.instructions[nextIP].op == bytecodeOpPop
}
