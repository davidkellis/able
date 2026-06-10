package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execArrayReadSlotI32(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode array read_slot i32 missing instruction")
	}
	receiverSlot, indexSlot := instr.argCount, instr.loopBreak
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		indexSlot < 0 || indexSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array read_slot i32 slot out of range")
	}
	if arr, ok := vm.slots[receiverSlot].(*runtime.ArrayValue); ok && arr != nil && vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		if vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, vm.ip, bytecodeMemberMethodFastPathArrayReadSlot) {
			if handled, err := vm.tryExecArrayReadSlotI32Fast(instr, arr, indexSlot); handled || err != nil {
				return err
			}
		} else if ok, err := vm.proveCanonicalArrayReadSlotCall(program, vm.ip, instr, vm.slots[receiverSlot]); err != nil {
			return err
		} else if ok {
			if handled, err := vm.tryExecArrayReadSlotI32Fast(instr, arr, indexSlot); handled || err != nil {
				return err
			}
		}
	}
	index := vm.slots[indexSlot]
	if index == nil && vm.hasI32RegisterFrame() {
		index = vm.slotMaterializedValue(indexSlot)
	}
	result, err := vm.arrayReadSlotValue(instr, program, vm.slots[receiverSlot], index)
	if err != nil {
		return err
	}
	raw, fallback, ok, err := vm.unboxAssignableI32Value(result)
	if err != nil {
		return err
	}
	if !ok {
		vm.i32UnboxFallbackValue = fallback
		vm.i32UnboxFallbackSet = true
		vm.ip++
		return nil
	}
	vm.i32UnboxFallbackValue = nil
	vm.i32UnboxFallbackSet = false
	vm.pushI32(raw)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) tryExecArrayReadSlotI32Fast(instr *bytecodeInstruction, arr *runtime.ArrayValue, indexSlot int) (bool, error) {
	if vm == nil || instr == nil || arr == nil {
		return false, nil
	}
	raw, handled, err := vm.arrayReadSlotTrackedI32RawAtSlot(arr, indexSlot)
	if err != nil || handled {
		if err != nil {
			return true, err
		}
		vm.i32UnboxFallbackValue = nil
		vm.i32UnboxFallbackSet = false
		vm.pushI32(int32(raw))
		vm.ip++
		return true, nil
	}
	return false, nil
}
