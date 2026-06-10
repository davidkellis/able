package interpreter

func (vm *bytecodeVM) switchRunProgram(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, next *bytecodeProgram) {
	vm.switchRunProgramWithActiveLookupState(program, instructions, validatedIntConsts, slotConstIntImmTable, next, bytecodeActiveLookupProgramState{})
}

func (vm *bytecodeVM) switchRunProgramI32RegisterFrame(next *bytecodeProgram) {
	if vm == nil || next == nil || vm.i32RegisterProgram == next {
		return
	}
	if vm.i32RegisterProgram == nil {
		layout := next.frameLayout
		if layout == nil || !layout.i32RegisterFrame || layout.slotCount <= 0 {
			return
		}
	}
	vm.activateI32RegisterFrame(next)
}

func bytecodeProgramNeedsI32RegisterFrame(program *bytecodeProgram) bool {
	if program == nil || program.frameLayout == nil {
		return false
	}
	layout := program.frameLayout
	return layout.i32RegisterFrame && layout.slotCount > 0
}

func (vm *bytecodeVM) switchRunProgramI32RegisterFrameIfNeeded(next *bytecodeProgram) {
	if vm == nil {
		return
	}
	if vm.i32RegisterProgram == nil && !bytecodeProgramNeedsI32RegisterFrame(next) {
		return
	}
	vm.switchRunProgramI32RegisterFrame(next)
}

func (vm *bytecodeVM) switchRunProgramWithActiveLookupState(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, next *bytecodeProgram, activeLookup bytecodeActiveLookupProgramState) {
	if vm == nil || next == nil {
		return
	}
	if vm.bytecodeProgramEntryPending {
		vm.interp.recordBytecodeProgramEntry(next)
		vm.bytecodeProgramEntryPending = false
	}
	if vm.interp != nil && vm.interp.bytecodeArrayOwnershipProfile != nil {
		vm.ensureBytecodeArrayOwnershipForProgram(next)
	}
	if program != nil && *program == next {
		vm.restoreOrSetActiveLookupProgram(next, activeLookup)
		vm.currentProgram = next
		vm.switchRunProgramI32RegisterFrameIfNeeded(next)
		return
	}
	if program != nil {
		*program = next
	}
	if instructions != nil {
		*instructions = next.instructions
	}
	instructionCount := 0
	if validatedIntConsts != nil {
		instructionCount = len(next.instructions)
		if next.integerConstValidationKnown && next.integerConstInstructionCount == instructionCount && !next.hasIntegerConstValidation {
			*validatedIntConsts = nil
		} else {
			*validatedIntConsts = vm.validatedIntegerConstSlots(next)
		}
	}
	if slotConstIntImmTable != nil {
		if instructionCount == 0 {
			instructionCount = len(next.instructions)
		}
		if next.slotConstIntImmTableKnown && next.slotConstIntImmInstructionCount == instructionCount {
			*slotConstIntImmTable = next.slotConstIntImmTable
		} else {
			*slotConstIntImmTable = vm.slotConstImmediateTable(next)
		}
	}
	vm.restoreOrSetActiveLookupProgram(next, activeLookup)
	vm.currentProgram = next
	vm.switchRunProgramI32RegisterFrameIfNeeded(next)
}

func (vm *bytecodeVM) restoreOrSetActiveLookupProgram(next *bytecodeProgram, activeLookup bytecodeActiveLookupProgramState) {
	if vm == nil || next == nil {
		return
	}
	if activeLookup.program == next {
		vm.activeLookup = activeLookup
		return
	}
	if vm.activeLookup.program == next {
		return
	}
	vm.activeLookup = bytecodeActiveLookupProgramState{program: next}
}
