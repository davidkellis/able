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
	var iv *runtime.IntegerValue
	switch value := val.(type) {
	case bytecodeRawI32SlotValue:
		idx := int64(value)
		if idx < 0 || idx > 1<<31-1 {
			return 0, false
		}
		return int(idx), true
	case runtime.IntegerValue:
		iv = &value
	case *runtime.IntegerValue:
		iv = value
	default:
		return 0, false
	}
	if iv == nil || !iv.IsSmallRef() {
		return 0, false
	}
	idx := iv.Int64FastRef()
	if idx < 0 || idx > 1<<31-1 {
		return 0, false
	}
	return int(idx), true
}

func (vm *bytecodeVM) readArraySlotValueFast(arr *runtime.ArrayValue, index runtime.Value) (runtime.Value, string, bool, error) {
	if vm == nil || arr == nil {
		return nil, "", false, nil
	}
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
	if rawByte, ok, err := runtime.ArrayStoreMonoReadU8IfAvailable(handle, idx); err != nil {
		return nil, "", true, err
	} else if ok {
		return boxedOrSmallIntegerValue(runtime.IntegerU8, int64(rawByte)), "array_read_slot_mono_u8_fast", true, nil
	}
	result, err := runtime.ArrayStoreRead(handle, idx)
	return result, "array_read_slot_fast", true, err
}

func (vm *bytecodeVM) execArrayReadSlotMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack) {
		return nil, false, nil
	}
	arr, ok := vm.stack[receiverIndex].(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	return vm.finishArrayReadSlotMemberFast(instr, arr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) finishArrayReadSlotMemberFast(instr bytecodeInstruction, arr *runtime.ArrayValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || arr == nil || instr.argCount != 1 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase >= len(vm.stack) {
		return nil, false, nil
	}
	indexVal := vm.stack[argBase]
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := arraySlotIndexSmall(indexVal); ok && idx < len(state.Values) {
			result := state.Values[idx]
			if result == nil {
				result = runtime.NilValue{}
			}
			vm.stack = vm.stack[:receiverIndex]
			vm.stack = append(vm.stack, result)
			vm.ip++
			return nil, true, nil
		}
	}
	result, mode, handled, err := vm.readArraySlotValueFast(arr, indexVal)
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !handled {
		return nil, false, nil
	}
	if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", mode, instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(result, err, callNode, nil)
	return newProg, true, finishErr
}

func (vm *bytecodeVM) execArrayWriteSlotMemberFast(instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || instr.argCount != 2 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase+1 >= len(vm.stack) || vm.interp == nil {
		return nil, false, nil
	}
	arr, ok := vm.stack[receiverIndex].(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, false, nil
	}
	return vm.finishArrayWriteSlotMemberFast(instr, arr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) finishArrayWriteSlotMemberFast(instr bytecodeInstruction, arr *runtime.ArrayValue, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || arr == nil || instr.argCount != 2 || receiverIndex < 0 || receiverIndex >= len(vm.stack) || argBase < 0 || argBase+1 >= len(vm.stack) || vm.interp == nil {
		return nil, false, nil
	}
	indexVal := vm.stack[argBase]
	value := vm.stack[argBase+1]
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if idx, ok := arraySlotIndexSmall(indexVal); ok {
			switch length := len(state.Values); {
			case idx == length:
				vm.appendTrackedArrayValueFast(arr, state, value)
			case idx > length:
				runtime.ArrayEnsureCapacity(state, idx+1)
				runtime.ArraySetLength(state, idx+1)
				state.Values[idx] = value
				vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
			default:
				state.Values[idx] = value
				vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
			}
			vm.stack = vm.stack[:receiverIndex]
			vm.stack = append(vm.stack, runtime.VoidValue{})
			vm.ip++
			return nil, true, nil
		}
	}
	idx, ok, err := bytecodeArraySlotIndexI32(indexVal)
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !ok {
		return nil, false, nil
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		switch length := len(state.Values); {
		case idx == length:
			vm.appendTrackedArrayValueFast(arr, state, value)
		case idx > length:
			runtime.ArrayEnsureCapacity(state, idx+1)
			runtime.ArraySetLength(state, idx+1)
			state.Values[idx] = value
			vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
		default:
			state.Values[idx] = value
			vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
		}
		if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_write_slot_tracked_fast", instr.node)
		}
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, nil, callNode, nil)
		return newProg, true, finishErr
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		vm.stack = vm.stack[:receiverIndex]
		newProg, finishErr := vm.finishCompletedCall(nil, err, callNode, nil)
		return newProg, true, finishErr
	}
	if !ok {
		return nil, false, nil
	}
	err = runtime.ArrayStoreWrite(handle, idx, value)
	if err == nil {
		if state, stateErr := runtime.ArrayStoreState(handle); stateErr == nil {
			vm.interp.syncArrayValues(handle, state)
		}
	}
	if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "array_write_slot_fast", instr.node)
	}
	vm.stack = vm.stack[:receiverIndex]
	newProg, finishErr := vm.finishCompletedCall(runtime.VoidValue{}, err, callNode, nil)
	return newProg, true, finishErr
}
