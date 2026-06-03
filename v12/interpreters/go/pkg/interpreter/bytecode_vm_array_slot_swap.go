package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execArraySlotSwapSlot(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode array slot swap missing instruction")
	}
	receiverSlot, firstSlot, secondSlot := instr.argCount, instr.loopBreak, instr.loopContinue
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		firstSlot < 0 || firstSlot >= len(vm.slots) ||
		secondSlot < 0 || secondSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array slot swap slot out of range")
	}
	receiver := vm.slots[receiverSlot]
	if !vm.hasI32RegisterFrame() {
		err := vm.resolveArraySlotSwapSlot(instr, program, receiver, vm.slots[firstSlot], vm.slots[secondSlot])
		if err != nil {
			return vm.attachArraySlotSwapSlotError(err, instr)
		}
		vm.stack = append(vm.stack, runtime.VoidValue{})
		vm.ip++
		return nil
	}
	var err error
	if arr, ok := receiver.(*runtime.ArrayValue); ok && arr != nil && vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		err = vm.resolveArraySlotSwapSlotAtSlots(instr, program, receiver, firstSlot, secondSlot)
	} else {
		err = vm.resolveArraySlotSwapSlot(instr, program, receiver, vm.slotMaterializedValue(firstSlot), vm.slotMaterializedValue(secondSlot))
	}
	if err != nil {
		return vm.attachArraySlotSwapSlotError(err, instr)
	}
	vm.stack = append(vm.stack, runtime.VoidValue{})
	vm.ip++
	return nil
}

func (vm *bytecodeVM) resolveArraySlotSwapSlotAtSlots(instr *bytecodeInstruction, program *bytecodeProgram, receiver runtime.Value, firstSlot int, secondSlot int) error {
	arr, ok := receiver.(*runtime.ArrayValue)
	if !ok || arr == nil || !vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		return vm.resolveArraySlotSwapSlot(instr, program, receiver, vm.slotMaterializedValue(firstSlot), vm.slotMaterializedValue(secondSlot))
	}
	if vm.lookupCachedCanonicalArraySlotCallForArray(program, vm.ip, bytecodeMemberMethodFastPathArrayReadWriteSlot) {
		if handled, err := vm.resolveArraySlotSwapSlotFastAtSlots(arr, firstSlot, secondSlot); handled || err != nil {
			return err
		}
	} else if ok, err := vm.proveCanonicalArrayReadWriteSlotCalls(program, vm.ip, receiver); err != nil {
		return err
	} else if ok {
		if handled, err := vm.resolveArraySlotSwapSlotFastAtSlots(arr, firstSlot, secondSlot); handled || err != nil {
			return err
		}
	}
	return vm.resolveArraySlotSwapSlotGeneric(receiver, vm.slotMaterializedValue(firstSlot), vm.slotMaterializedValue(secondSlot))
}

func (vm *bytecodeVM) resolveArraySlotSwapSlot(instr *bytecodeInstruction, program *bytecodeProgram, receiver runtime.Value, firstIdx runtime.Value, secondIdx runtime.Value) error {
	if arr, ok := receiver.(*runtime.ArrayValue); ok && arr != nil && vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		if vm.lookupCachedCanonicalArraySlotCallForArray(program, vm.ip, bytecodeMemberMethodFastPathArrayReadWriteSlot) {
			if handled, err := vm.resolveArraySlotSwapSlotFast(arr, firstIdx, secondIdx); handled || err != nil {
				return err
			}
		} else if ok, err := vm.proveCanonicalArrayReadWriteSlotCalls(program, vm.ip, receiver); err != nil {
			return err
		} else if ok {
			if handled, err := vm.resolveArraySlotSwapSlotFast(arr, firstIdx, secondIdx); handled || err != nil {
				return err
			}
		}
	}
	return vm.resolveArraySlotSwapSlotGeneric(receiver, firstIdx, secondIdx)
}

func (vm *bytecodeVM) resolveArraySlotSwapSlotFastAtSlots(arr *runtime.ArrayValue, firstSlot int, secondSlot int) (bool, error) {
	if handled := vm.resolveTrackedArraySlotSwapSlotFastAtSlots(arr, firstSlot, secondSlot); handled {
		return true, nil
	}
	left, _, handled, err := vm.readArraySlotValueFastAtSlot(arr, firstSlot)
	if err != nil || !handled {
		return handled, err
	}
	right, _, handled, err := vm.readArraySlotValueFastAtSlot(arr, secondSlot)
	if err != nil || !handled {
		return handled, err
	}
	if _, handled, err := vm.writeArraySlotValueFastAtSlot(arr, firstSlot, right); err != nil || !handled {
		return handled, err
	}
	_, handled, err = vm.writeArraySlotValueFastAtSlot(arr, secondSlot, left)
	return handled, err
}

func (vm *bytecodeVM) resolveArraySlotSwapSlotFast(arr *runtime.ArrayValue, firstIdx runtime.Value, secondIdx runtime.Value) (bool, error) {
	if handled := vm.resolveTrackedArraySlotSwapSlotFast(arr, firstIdx, secondIdx); handled {
		return true, nil
	}
	left, _, handled, err := vm.readArraySlotValueFast(arr, firstIdx)
	if err != nil || !handled {
		return handled, err
	}
	right, _, handled, err := vm.readArraySlotValueFast(arr, secondIdx)
	if err != nil || !handled {
		return handled, err
	}
	if _, handled, err := vm.writeArraySlotValueFast(arr, firstIdx, right); err != nil || !handled {
		return handled, err
	}
	_, handled, err = vm.writeArraySlotValueFast(arr, secondIdx, left)
	return handled, err
}

func (vm *bytecodeVM) resolveTrackedArraySlotSwapSlotFastAtSlots(arr *runtime.ArrayValue, firstSlot int, secondSlot int) bool {
	if arr == nil {
		return false
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked || state == nil {
		return false
	}
	first, firstOK := vm.slotArraySlotIndexSmall(firstSlot)
	second, secondOK := vm.slotArraySlotIndexSmall(secondSlot)
	if !firstOK || !secondOK || first >= len(state.Values) || second >= len(state.Values) {
		return false
	}
	left := state.Values[first]
	right := state.Values[second]
	if left == nil {
		left = runtime.NilValue{}
	}
	if right == nil {
		right = runtime.NilValue{}
	}
	state.Values[first] = right
	state.Values[second] = left
	vm.syncTrackedArrayIndexSwapSlot(arr, state, first, right, second, left)
	return true
}

func (vm *bytecodeVM) resolveTrackedArraySlotSwapSlotFast(arr *runtime.ArrayValue, firstIdx runtime.Value, secondIdx runtime.Value) bool {
	if arr == nil {
		return false
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked || state == nil {
		return false
	}
	first, firstOK := arraySlotIndexSmall(firstIdx)
	second, secondOK := arraySlotIndexSmall(secondIdx)
	if !firstOK || !secondOK || first >= len(state.Values) || second >= len(state.Values) {
		return false
	}
	left := state.Values[first]
	right := state.Values[second]
	if left == nil {
		left = runtime.NilValue{}
	}
	if right == nil {
		right = runtime.NilValue{}
	}
	state.Values[first] = right
	state.Values[second] = left
	vm.syncTrackedArrayIndexSwapSlot(arr, state, first, right, second, left)
	return true
}

func (vm *bytecodeVM) resolveArraySlotSwapSlotGeneric(receiver runtime.Value, firstIdx runtime.Value, secondIdx runtime.Value) error {
	left, err := vm.genericArrayReadSlotCompareValue(receiver, firstIdx)
	if err != nil {
		return err
	}
	right, err := vm.genericArrayReadSlotCompareValue(receiver, secondIdx)
	if err != nil {
		return err
	}
	if err := vm.genericArrayWriteSlotValue(receiver, firstIdx, right); err != nil {
		return err
	}
	return vm.genericArrayWriteSlotValue(receiver, secondIdx, left)
}

func (vm *bytecodeVM) genericArrayWriteSlotValue(receiver runtime.Value, index runtime.Value, value runtime.Value) error {
	if vm == nil || vm.interp == nil {
		return fmt.Errorf("bytecode VM is nil")
	}
	callee, err := vm.interp.memberAccessOnValueWithOptions(receiver, ast.ID("write_slot"), vm.env, true)
	if err != nil {
		return err
	}
	args := [2]runtime.Value{index, value}
	_, err = vm.interp.callCallableValueMutable(callee, args[:], vm.env, nil)
	return err
}

func (vm *bytecodeVM) proveCanonicalArrayReadWriteSlotCalls(program *bytecodeProgram, ip int, receiver runtime.Value) (bool, error) {
	arr, ok := receiver.(*runtime.ArrayValue)
	if !ok || arr == nil || !vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		return false, nil
	}
	readOK, err := vm.proveCanonicalArraySlotMethod("read_slot", receiver, bytecodeMemberMethodFastPathArrayReadSlot)
	if err != nil || !readOK {
		return false, err
	}
	writeOK, err := vm.proveCanonicalArraySlotMethod("write_slot", receiver, bytecodeMemberMethodFastPathArrayWriteSlot)
	if err != nil || !writeOK {
		return false, err
	}
	vm.storeCachedCanonicalArraySlotCallForArray(program, ip, arr, bytecodeMemberMethodFastPathArrayReadWriteSlot)
	return true, nil
}

func (vm *bytecodeVM) proveCanonicalArraySlotMethod(name string, receiver runtime.Value, expected bytecodeMemberMethodFastPathKind) (bool, error) {
	if vm == nil || vm.interp == nil {
		return false, nil
	}
	callable, found, err := vm.interp.resolveMethodCallableFromPool(vm.env, name, receiver, "")
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
	return vm.resolvedMemberMethodFastPath(name, receiver, fn) == expected, nil
}

func (vm *bytecodeVM) attachArraySlotSwapSlotError(err error, instr *bytecodeInstruction) error {
	if err == nil || vm == nil || vm.interp == nil {
		return err
	}
	err = vm.interp.wrapStandardRuntimeError(err)
	if instr != nil && instr.node != nil {
		err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
	}
	return err
}
