package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func isCanonicalArrayReadSlotFunction(def *ast.FunctionDefinition) bool {
	return def != nil &&
		len(def.Params) == 2 &&
		typeExpressionToString(def.Params[1].ParamType) == "i32" &&
		typeExpressionToString(def.ReturnType) == "T"
}

func isCanonicalArrayWriteSlotFunction(def *ast.FunctionDefinition) bool {
	return def != nil &&
		len(def.Params) == 3 &&
		typeExpressionToString(def.Params[1].ParamType) == "i32" &&
		typeExpressionToString(def.ReturnType) == "void"
}

func bytecodeArraySlotIndexI32(val runtime.Value) (int, bool, error) {
	intVal, ok := bytecodeIntegerValue(val)
	if !ok {
		return 0, false, nil
	}
	var idx int64
	if intVal.IsSmall() {
		idx = intVal.Int64Fast()
	} else {
		var fits bool
		idx, fits = intVal.ToInt64()
		if !fits {
			return 0, false, nil
		}
	}
	if idx < -1<<31 || idx > 1<<31-1 {
		return 0, false, nil
	}
	if idx < 0 {
		return 0, true, fmt.Errorf("array index must be non-negative")
	}
	return int(idx), true, nil
}

func arraySlotIndexSmall(val runtime.Value) (int, bool) {
	switch idx := val.(type) {
	case runtime.IntegerValue:
		if idx.IsSmall() {
			raw := idx.Int64Fast()
			if raw >= 0 && raw <= 1<<31-1 {
				return int(raw), true
			}
			return 0, false
		}
	case *runtime.IntegerValue:
		if idx != nil && idx.IsSmallRef() {
			raw := idx.Int64FastRef()
			if raw >= 0 && raw <= 1<<31-1 {
				return int(raw), true
			}
			return 0, false
		}
	}
	_, idx, ok := bytecodeRawIntegerValueInfo(val)
	if !ok {
		return 0, false
	}
	if idx < 0 || idx > 1<<31-1 {
		return 0, false
	}
	return int(idx), true
}

func (vm *bytecodeVM) readArraySlotValueFast(arr *runtime.ArrayValue, index runtime.Value) (runtime.Value, string, bool, error) {
	if vm == nil || arr == nil {
		return nil, "", false, nil
	}
	return vm.readArraySlotValueFastChecked(arr, index)
}

func (vm *bytecodeVM) readArraySlotValueFastChecked(arr *runtime.ArrayValue, index runtime.Value) (runtime.Value, string, bool, error) {
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := arraySlotIndexSmall(index); ok && idx < len(state.Values) {
			result := state.Values[idx]
			if result == nil {
				return runtime.NilValue{}, "array_read_slot_tracked_fast", true, nil
			}
			return result, "array_read_slot_tracked_fast", true, nil
		}
	}
	idx, ok, err := bytecodeArraySlotIndexI32(index)
	if err != nil {
		return nil, "", true, err
	}
	if !ok {
		return nil, "", false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx < len(state.Values) {
			return state.Values[idx], "array_read_slot_tracked_fast", true, nil
		}
		return runtime.NilValue{}, "array_read_slot_tracked_fast", true, nil
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		return nil, "", true, err
	}
	if !ok {
		return nil, "", false, nil
	}
	var info runtime.ArrayStoreMonoPrimitiveReadInfo
	if ok, err := runtime.ArrayStoreMonoPrimitiveReadInfoInto(handle, idx, &info); err != nil {
		return nil, "", true, err
	} else if ok {
		dispatch := bytecodeMonoPrimitiveArrayDispatch(info, "array_read_slot")
		if !info.InBounds {
			return runtime.NilValue{}, dispatch, true, nil
		}
		if result, mono := bytecodeMonoPrimitiveArrayValue(info); mono {
			return result, dispatch, true, nil
		}
	}
	result, err := runtime.ArrayStoreRead(handle, idx)
	return result, "array_read_slot_fast", true, err
}

func (vm *bytecodeVM) readArraySlotValueFastAtSlot(arr *runtime.ArrayValue, indexSlot int) (runtime.Value, string, bool, error) {
	if vm == nil || arr == nil {
		return nil, "", false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := vm.slotArraySlotIndexSmall(indexSlot); ok && idx < len(state.Values) {
			result := state.Values[idx]
			if result == nil {
				return runtime.NilValue{}, "array_read_slot_tracked_fast", true, nil
			}
			return result, "array_read_slot_tracked_fast", true, nil
		}
	}
	return vm.readArraySlotValueFastChecked(arr, vm.slotMaterializedValue(indexSlot))
}

func (vm *bytecodeVM) execArrayReadSlotMemberFast(memberName string, argCount int, traceNode ast.Node, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	arr, ok := vm.stackValue(receiverIndex).(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	return vm.finishArrayReadSlotMemberFast(memberName, argCount, traceNode, arr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) finishArrayReadSlotMemberFast(memberName string, argCount int, traceNode ast.Node, arr *runtime.ArrayValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || arr == nil || argCount != 1 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase >= vm.stackDepth() {
		return nil, false, nil
	}
	indexVal := vm.stackValue(argBase)
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := arraySlotIndexSmall(indexVal); ok && idx < len(state.Values) {
			result := state.Values[idx]
			if result == nil {
				result = runtime.NilValue{}
			}
			vm.truncateStack(receiverIndex)
			vm.appendStackValue(result)
			vm.ip++
			return nil, true, nil
		}
	}
	result, mode, handled, err := vm.readArraySlotValueFastChecked(arr, indexVal)
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !handled {
		return nil, false, nil
	}
	if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", mode, traceNode)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedCall(result, err, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execArrayWriteSlotMemberFast(memberName string, argCount int, traceNode ast.Node, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || argCount != 2 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase+1 >= vm.stackDepth() || vm.interp == nil {
		return nil, false, nil
	}
	arr, ok := vm.stackValue(receiverIndex).(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	return vm.finishArrayWriteSlotMemberFast(memberName, argCount, traceNode, arr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) finishArrayWriteSlotMemberFast(memberName string, argCount int, traceNode ast.Node, arr *runtime.ArrayValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || arr == nil || argCount != 2 || receiverIndex < 0 || receiverIndex >= vm.stackDepth() || argBase < 0 || argBase+1 >= vm.stackDepth() || vm.interp == nil {
		return nil, false, nil
	}
	indexVal := vm.stackValue(argBase)
	value := vm.stackValue(argBase + 1)
	vm.observeBytecodeArrayOwnershipArrayWrite(arr, value)
	mode, handled, err := vm.writeArraySlotValueFastChecked(arr, indexVal, value)
	if err != nil {
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !handled {
		return nil, false, nil
	}
	if vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", mode, traceNode)
	}
	vm.truncateStack(receiverIndex)
	newProg, finishErr := vm.finishCompletedVoidCallFast()
	return newProg, true, finishErr
}

func (vm *bytecodeVM) writeArraySlotValueFast(arr *runtime.ArrayValue, indexVal runtime.Value, value runtime.Value) (string, bool, error) {
	if vm == nil || arr == nil || vm.interp == nil {
		return "", false, nil
	}
	return vm.writeArraySlotValueFastChecked(arr, indexVal, value)
}

func (vm *bytecodeVM) writeArraySlotValueFastChecked(arr *runtime.ArrayValue, indexVal runtime.Value, value runtime.Value) (string, bool, error) {
	value = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, value)
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := arraySlotIndexSmall(indexVal); ok {
			switch length := len(state.Values); {
			case idx == length:
				vm.appendTrackedArrayValueFast(arr, state, value)
			case idx > length:
				runtime.ArrayEnsureCapacity(state, idx+1)
				setArrayLength(state, idx+1)
				state.Values[idx] = value
				if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
					vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
				}
			default:
				state.Values[idx] = value
				if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
					vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
				}
			}
			return "array_write_slot_tracked_fast", true, nil
		}
	}
	idx, ok, err := bytecodeArraySlotIndexI32(indexVal)
	if err != nil {
		return "", true, err
	}
	if !ok {
		return "", false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		switch length := len(state.Values); {
		case idx == length:
			vm.appendTrackedArrayValueFast(arr, state, value)
		case idx > length:
			runtime.ArrayEnsureCapacity(state, idx+1)
			setArrayLength(state, idx+1)
			state.Values[idx] = value
			if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
				vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
			}
		default:
			state.Values[idx] = value
			if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
				vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
			}
		}
		return "array_write_slot_tracked_fast", true, nil
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		return "", true, err
	}
	if !ok {
		return "", false, nil
	}
	storedValue := vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, value)
	err = runtime.ArrayStoreWrite(handle, idx, storedValue)
	if err == nil {
		vm.interp.syncArrayHandleWriteAfterStore(handle, idx, storedValue)
	}
	return "array_write_slot_fast", true, err
}

func (vm *bytecodeVM) writeArraySlotValueFastAtSlot(arr *runtime.ArrayValue, indexSlot int, value runtime.Value) (string, bool, error) {
	if vm == nil || arr == nil || vm.interp == nil {
		return "", false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := vm.slotArraySlotIndexSmall(indexSlot); ok {
			switch length := len(state.Values); {
			case idx == length:
				vm.appendTrackedArrayValueFast(arr, state, value)
			case idx > length:
				runtime.ArrayEnsureCapacity(state, idx+1)
				setArrayLength(state, idx+1)
				state.Values[idx] = value
				vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
			default:
				state.Values[idx] = value
				vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
			}
			return "array_write_slot_tracked_fast", true, nil
		}
	}
	return vm.writeArraySlotValueFastChecked(arr, vm.slotMaterializedValue(indexSlot), value)
}
