package interpreter

func (vm *bytecodeVM) switchRunProgram(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, next *bytecodeProgram) {
	if vm == nil || next == nil {
		return
	}
	if program != nil && *program == next {
		vm.currentProgram = next
		return
	}
	*program = next
	*instructions = next.instructions
	*validatedIntConsts = vm.validatedIntegerConstSlots(next)
	*slotConstIntImmTable = vm.slotConstImmediateTable(next)
	vm.currentProgram = next
}
