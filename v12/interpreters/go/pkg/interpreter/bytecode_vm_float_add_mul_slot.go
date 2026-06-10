package interpreter

import (
	"fmt"
)

func (vm *bytecodeVM) execStoreSlotFloatAddMulSlot(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode float add-mul slot store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	baseSlot := instr.argCount
	mulSlot := instr.loopBreak
	if baseSlot < 0 || baseSlot >= len(vm.slots) || mulSlot < 0 || mulSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode float add-mul slot source out of range")
	}
	if vm.stackDepth() < 1 {
		return fmt.Errorf("bytecode float add-mul slot store missing stack operand")
	}

	stackIdx := vm.stackDepth() - 1
	stackVal := vm.stackValue(stackIdx)
	baseVal, baseKind, baseOK := vm.slotDirectFloatValue(baseSlot)
	stackRaw, stackKind, stackOK := bytecodeDirectFloatValue(stackVal)
	mulVal, mulKind, mulOK := vm.slotDirectFloatValue(mulSlot)
	if baseOK && stackOK && mulOK {
		if resultVal, resultKind, ok := bytecodeDirectFloatAddMulOperandsRawValue(baseVal, baseKind, stackRaw, stackKind, mulVal, mulKind); ok {
			vm.truncateStack(stackIdx)
			return vm.finishStoreSlotFloatRawResult(instr, resultVal, resultKind)
		}
	}

	base := vm.slotRuntimeValue(baseSlot)
	mulRight := vm.slotRuntimeValue(mulSlot)
	result, err := vm.storeSlotFloatAddMulResult(instr, base, stackVal, mulRight)
	vm.truncateStack(stackIdx)
	return vm.finishStoreSlotFloatResult(instr, result, err)
}
