package interpreter

import "fmt"

func (vm *bytecodeVM) execStoreSlotFloatBinary(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode float slot binary store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	leftSlot := instr.argCount
	rightSlot := instr.loopBreak
	if leftSlot < 0 || leftSlot >= len(vm.slots) || rightSlot < 0 || rightSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode float slot binary source out of range")
	}

	leftVal, leftKind, leftOK := vm.slotDirectFloatValueValidated(leftSlot)
	rightVal, rightKind, rightOK := leftVal, leftKind, leftOK
	if leftSlot != rightSlot {
		rightVal, rightKind, rightOK = vm.slotDirectFloatValueValidated(rightSlot)
	}
	if leftOK && rightOK {
		if resultVal, resultKind, handled := bytecodeDirectFloatArithmeticRawFast(instr.operator, leftVal, leftKind, rightVal, rightKind); handled {
			return vm.finishStoreSlotFloatRawResult(instr, resultVal, resultKind)
		}
	}

	left := vm.slotRuntimeValue(leftSlot)
	right := vm.slotRuntimeValue(rightSlot)
	result, err := applyBinaryOperator(vm.interp, instr.operator, left, right)
	return vm.finishStoreSlotFloatResult(instr, result, err)
}
