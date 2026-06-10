package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execJumpIfArrayReadSlotCompareSlotFalse(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode array-slot compare jump missing instruction")
	}
	receiverSlot, indexSlot, rightSlot := instr.argCount, instr.loopBreak, instr.loopContinue
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		indexSlot < 0 || indexSlot >= len(vm.slots) ||
		rightSlot < 0 || rightSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array-slot compare slot out of range")
	}
	if cond, handled, err := vm.compareArrayReadSlotTrackedI32Condition(instr, program, receiverSlot, indexSlot, rightSlot); handled || err != nil {
		if err != nil {
			return err
		}
		if !cond {
			vm.ip = instr.target
			return nil
		}
		vm.ip++
		return nil
	}
	receiver := vm.slots[receiverSlot]
	index := vm.slotMaterializedValue(indexSlot)
	right := vm.slotMaterializedValue(rightSlot)
	left, err := vm.arrayReadSlotCompareValue(instr, program, receiver, index)
	if err != nil {
		return err
	}
	cond, err := vm.compareArrayReadSlotCondition(instr.operator, left, right)
	if err != nil {
		return err
	}
	if !cond {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) compareArrayReadSlotTrackedI32Condition(instr *bytecodeInstruction, program *bytecodeProgram, receiverSlot int, indexSlot int, rightSlot int) (bool, bool, error) {
	if vm == nil || instr == nil || program == nil {
		return false, false, nil
	}
	rightRaw, ok := vm.slotDirectTrackedCompareI32Raw(rightSlot)
	if !ok {
		return false, false, nil
	}
	arr, ok := vm.slots[receiverSlot].(*runtime.ArrayValue)
	if !ok || arr == nil || !vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		return false, false, nil
	}
	if !vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, vm.ip, bytecodeMemberMethodFastPathArrayReadSlot) {
		return false, false, nil
	}
	leftRaw, handled, err := vm.arrayReadSlotTrackedI32RawAtSlot(arr, indexSlot)
	if err != nil || !handled {
		return false, handled, err
	}
	cond, ok := bytecodeCompareI32(instr.operator, leftRaw, rightRaw)
	if !ok {
		return false, false, nil
	}
	return cond, true, nil
}

func (vm *bytecodeVM) slotDirectTrackedCompareI32Raw(slot int) (int32, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false
	}
	if value := vm.slots[slot]; value != nil {
		return trackedArrayCompareDirectSmallI32Raw(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return raw, true
	}
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		return raw, true
	}
	return 0, false
}

func (vm *bytecodeVM) trackedArrayCompareSlotIndexSmall(slot int) (int, bool) {
	return vm.slotArraySlotIndexSmall(slot)
}

func (vm *bytecodeVM) trackedArrayCompareI32RawAtSlot(state *runtime.ArrayState, indexSlot int) (int32, bool, error) {
	if vm == nil || state == nil {
		return 0, false, nil
	}
	idx, ok := vm.trackedArrayCompareSlotIndexSmall(indexSlot)
	if !ok || idx < 0 || idx >= len(state.Values) {
		return 0, false, nil
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, idx); ok {
		return int32(raw), true, nil
	}
	value := state.Values[idx]
	if value == nil {
		return 0, false, nil
	}
	raw, ok := trackedArrayCompareDirectSmallI32Raw(value)
	if !ok {
		return 0, false, nil
	}
	return raw, true, nil
}

func trackedArrayCompareDirectSmallI32Raw(value runtime.Value) (int32, bool) {
	switch iv := value.(type) {
	case bytecodeRawI32SlotValue:
		return int32(iv), true
	case runtime.IntegerValue:
		ivRef := &iv
		if iv.TypeSuffix == runtime.IntegerI32 && ivRef.IsSmallRef() {
			return int32(ivRef.Int64FastRef()), true
		}
	case *runtime.IntegerValue:
		if iv != nil && iv.TypeSuffix == runtime.IntegerI32 && iv.IsSmallRef() {
			return int32(iv.Int64FastRef()), true
		}
	}
	return 0, false
}

func (vm *bytecodeVM) arrayReadSlotTrackedI32RawAtSlot(arr *runtime.ArrayValue, indexSlot int) (int32, bool, error) {
	if vm == nil || arr == nil {
		return 0, false, nil
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked {
		return 0, false, nil
	}
	return vm.trackedArrayCompareI32RawAtSlot(state, indexSlot)
}

func bytecodeCompareI32(op string, left int32, right int32) (bool, bool) {
	switch op {
	case "==":
		return left == right, true
	case "!=":
		return left != right, true
	case "<":
		return left < right, true
	case "<=":
		return left <= right, true
	case ">":
		return left > right, true
	case ">=":
		return left >= right, true
	default:
		return false, false
	}
}

func (vm *bytecodeVM) execJumpIfArrayIndexSlotCompareSlotFalse(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode array-index compare jump missing instruction")
	}
	receiverSlot, indexSlot, rightSlot := instr.argCount, instr.loopBreak, instr.loopContinue
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		indexSlot < 0 || indexSlot >= len(vm.slots) ||
		rightSlot < 0 || rightSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array-index compare slot out of range")
	}
	if instr.name == "i32" {
		if cond, handled, err := vm.compareArrayIndexSlotI32ConditionAtSlots(instr, receiverSlot, indexSlot, rightSlot); handled || err != nil {
			if err != nil {
				return err
			}
			if !cond {
				vm.ip = instr.target
				return nil
			}
			vm.ip++
			return nil
		}
	}
	receiver := vm.slots[receiverSlot]
	index := vm.slotMaterializedValue(indexSlot)
	right := vm.slotMaterializedValue(rightSlot)
	left, err := vm.arrayIndexSlotCompareValue(instr, receiver, index)
	if err != nil {
		return err
	}
	cond, err := vm.compareArrayReadSlotCondition(instr.operator, left, right)
	if err != nil {
		return err
	}
	if !cond {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) arrayReadSlotCompareValue(instr *bytecodeInstruction, program *bytecodeProgram, receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	return vm.arrayReadSlotValue(instr, program, receiver, index)
}

func (vm *bytecodeVM) arrayIndexSlotCompareValue(instr *bytecodeInstruction, receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	var result runtime.Value
	var err error
	if vm.interp.canUseDirectArrayIndexGetFastPath() {
		if arr, ok := bytecodeArrayReceiverForIndexCache(receiver); ok {
			if idx, small := bytecodeDirectSmallArrayIndex(index); small {
				if state, tracked := bytecodeTrackedArrayState(arr); tracked {
					if idx < 0 || idx >= len(state.Values) {
						result = vm.interp.makeIndexErrorValue(idx, len(state.Values))
					} else if val := state.Values[idx]; val != nil {
						result = val
					} else {
						result = vm.interp.makeIndexErrorValue(idx, len(state.Values))
					}
				} else {
					result, err = vm.resolveDirectArrayIndexGetAt(arr, idx)
				}
			} else if value, handled, directErr := vm.resolveDirectArrayIndexGet(arr, index); handled {
				result, err = value, directErr
			}
		}
	}
	if result == nil && err == nil {
		result, err = vm.resolveIndexGet(receiver, index)
	}
	if err != nil {
		return nil, err
	}
	return vm.arrayIndexSlotCompareMaybeCast(instr, result)
}

func (vm *bytecodeVM) compareArrayIndexSlotI32Condition(instr *bytecodeInstruction, receiver runtime.Value, index runtime.Value, right runtime.Value) (bool, bool, error) {
	rightVal, ok := bytecodeDirectSmallI32Value(right)
	if !ok {
		return false, false, nil
	}
	leftVal, handled, err := vm.arrayIndexSlotCompareI32RawValue(receiver, index)
	if err != nil || !handled {
		return false, handled, err
	}
	cond, ok := bytecodeCompareInt64(instr.operator, leftVal, rightVal)
	if !ok {
		return false, false, nil
	}
	return cond, true, nil
}

func (vm *bytecodeVM) compareArrayIndexSlotI32ConditionAtSlots(instr *bytecodeInstruction, receiverSlot int, indexSlot int, rightSlot int) (bool, bool, error) {
	if vm == nil || instr == nil {
		return false, false, nil
	}
	rightVal, ok := vm.slotDirectSmallI32Value(rightSlot)
	if !ok {
		return false, false, nil
	}
	receiver := vm.slots[receiverSlot]
	leftVal, handled, err := vm.arrayIndexSlotCompareI32RawAtSlot(receiver, indexSlot)
	if err != nil || !handled {
		return false, handled, err
	}
	cond, ok := bytecodeCompareInt64(instr.operator, leftVal, rightVal)
	if !ok {
		return false, false, nil
	}
	return cond, true, nil
}

func (vm *bytecodeVM) arrayIndexSlotCompareI32RawValue(receiver runtime.Value, index runtime.Value) (int64, bool, error) {
	if vm == nil || vm.interp == nil || !vm.interp.canUseDirectArrayIndexGetFastPath() {
		return 0, false, nil
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return 0, false, nil
	}
	if idx, small := bytecodeDirectSmallArrayIndex(index); small {
		if state, tracked := bytecodeTrackedArrayState(arr); tracked {
			if idx < 0 || idx >= len(state.Values) {
				return 0, false, nil
			}
			if raw, ok := trackedArrayCachedI32RawAt(state, idx); ok {
				return raw, true, nil
			}
			if val := state.Values[idx]; val != nil {
				raw, ok := bytecodeArrayIndexCastSmallI32Raw(val)
				return raw, ok, nil
			}
			return 0, false, nil
		}
		value, err := vm.resolveDirectArrayIndexGetAt(arr, idx)
		if err != nil {
			return 0, true, err
		}
		raw, ok := bytecodeArrayIndexCastSmallI32Raw(value)
		return raw, ok, nil
	}
	value, handled, err := vm.resolveDirectArrayIndexGet(arr, index)
	if err != nil || !handled {
		return 0, handled, err
	}
	raw, ok := bytecodeArrayIndexCastSmallI32Raw(value)
	return raw, ok, nil
}

func (vm *bytecodeVM) arrayIndexSlotCompareI32RawAtSlot(receiver runtime.Value, indexSlot int) (int64, bool, error) {
	if vm == nil || vm.interp == nil || !vm.interp.canUseDirectArrayIndexGetFastPath() {
		return 0, false, nil
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return 0, false, nil
	}
	if idx, small := vm.slotDirectSmallArrayIndex(indexSlot); small {
		if state, tracked := bytecodeTrackedArrayState(arr); tracked {
			if idx < 0 || idx >= len(state.Values) {
				return 0, false, nil
			}
			if raw, ok := trackedArrayCachedI32RawAt(state, idx); ok {
				return raw, true, nil
			}
			if val := state.Values[idx]; val != nil {
				raw, ok := bytecodeArrayIndexCastSmallI32Raw(val)
				return raw, ok, nil
			}
			return 0, false, nil
		}
		value, err := vm.resolveDirectArrayIndexGetAt(arr, idx)
		if err != nil {
			return 0, true, err
		}
		raw, ok := bytecodeArrayIndexCastSmallI32Raw(value)
		return raw, ok, nil
	}
	index := vm.slotMaterializedValue(indexSlot)
	if index == nil {
		return 0, false, nil
	}
	value, handled, err := vm.resolveDirectArrayIndexGet(arr, index)
	if err != nil || !handled {
		return 0, handled, err
	}
	raw, ok := bytecodeArrayIndexCastSmallI32Raw(value)
	return raw, ok, nil
}

func (vm *bytecodeVM) arrayIndexSlotCompareMaybeCast(instr *bytecodeInstruction, value runtime.Value) (runtime.Value, error) {
	if instr == nil || instr.typeExpr == nil {
		if value == nil {
			return runtime.NilValue{}, nil
		}
		return value, nil
	}
	if instr.name == "i32" && bytecodeValueIsI32(value) {
		return value, nil
	}
	casted, err := vm.interp.castValueToType(vm.canonicalRuntimeTypeExpression(instr.typeExpr), value)
	if err != nil {
		return nil, err
	}
	if casted == nil {
		return runtime.NilValue{}, nil
	}
	return casted, nil
}

func bytecodeArrayIndexCastSmallI32Raw(value runtime.Value) (int64, bool) {
	switch v := value.(type) {
	case bytecodeRawI32SlotValue:
		return int64(v), true
	case runtime.IntegerValue:
		ref := &v
		if ref.IsSmallRef() {
			return int64(int32(uint32(ref.Int64FastRef()))), true
		}
	case *runtime.IntegerValue:
		if v != nil && v.IsSmallRef() {
			return int64(int32(uint32(v.Int64FastRef()))), true
		}
	}
	return 0, false
}

func bytecodeValueIsI32(value runtime.Value) bool {
	switch v := value.(type) {
	case bytecodeRawI32SlotValue:
		return true
	case runtime.IntegerValue:
		return v.TypeSuffix == runtime.IntegerI32
	case *runtime.IntegerValue:
		return v != nil && v.TypeSuffix == runtime.IntegerI32
	default:
		return false
	}
}

func (vm *bytecodeVM) arrayReadSlotValue(instr *bytecodeInstruction, program *bytecodeProgram, receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	if arr, ok := receiver.(*runtime.ArrayValue); ok && arr != nil && vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		if vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, vm.ip, bytecodeMemberMethodFastPathArrayReadSlot) {
			if value, mode, handled, err := vm.readArraySlotValueFastChecked(arr, index); handled || err != nil {
				if err == nil && vm.interp != nil && vm.interp.bytecodeTraceEnabled {
					vm.interp.recordBytecodeCallTrace("call_member", "read_slot", "resolved_method", mode, instr.node)
				}
				return value, err
			}
		} else if ok, err := vm.proveCanonicalArrayReadSlotCall(program, vm.ip, instr, receiver); err != nil {
			return nil, err
		} else if ok {
			if value, mode, handled, err := vm.readArraySlotValueFastChecked(arr, index); handled || err != nil {
				if err == nil && vm.interp != nil && vm.interp.bytecodeTraceEnabled {
					vm.interp.recordBytecodeCallTrace("call_member", "read_slot", "resolved_method", mode, instr.node)
				}
				return value, err
			}
		}
	}
	return vm.genericArrayReadSlotCompareValue(receiver, index)
}

func (vm *bytecodeVM) execArrayReadSlot(instr *bytecodeInstruction, program *bytecodeProgram) (*bytecodeProgram, error) {
	if instr == nil {
		return nil, fmt.Errorf("bytecode array read_slot missing instruction")
	}
	receiverSlot, indexSlot := instr.argCount, instr.loopBreak
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		indexSlot < 0 || indexSlot >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode array read_slot slot out of range")
	}
	index := vm.slots[indexSlot]
	if index == nil {
		index = vm.slotMaterializedValue(indexSlot)
	}
	result, err := vm.arrayReadSlotValue(instr, program, vm.slots[receiverSlot], index)
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func (vm *bytecodeVM) proveCanonicalArrayReadSlotCall(program *bytecodeProgram, ip int, instr *bytecodeInstruction, receiver runtime.Value) (bool, error) {
	if vm == nil || vm.interp == nil || instr == nil {
		return false, nil
	}
	callable, found, err := vm.interp.resolveMethodCallableFromPool(vm.env, "read_slot", receiver, "")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	fn, ok := bytecodeResolvedMemberFastPathFunction(callable)
	if !ok {
		return false, nil
	}
	kind := vm.resolvedMemberMethodFastPath("read_slot", receiver, fn)
	if kind != bytecodeMemberMethodFastPathArrayReadSlot {
		return false, nil
	}
	vm.storeCachedCanonicalArraySlotCall(program, ip, bytecodeInstruction{name: "read_slot", argCount: 1}, receiver, kind)
	return true, nil
}

func (vm *bytecodeVM) genericArrayReadSlotCompareValue(receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	callee, err := vm.interp.memberAccessOnValueWithOptions(receiver, ast.ID("read_slot"), vm.env, true)
	if err != nil {
		return nil, err
	}
	args := [1]runtime.Value{index}
	value, err := vm.callCallableValueMutable(callee, args[:], nil)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return runtime.NilValue{}, nil
	}
	return value, nil
}

func (vm *bytecodeVM) compareArrayReadSlotCondition(op string, left runtime.Value, right runtime.Value) (bool, error) {
	return vm.compareBytecodeCondition(op, left, right)
}
