package interpreter

func (vm *bytecodeVM) prepareRunProgram(program *bytecodeProgram, resume bool) *bytecodeProgram {
	if !resume {
		vm.stack = vm.stack[:0]
		vm.iterStack = vm.iterStack[:0]
		vm.loopStack = vm.loopStack[:0]
		vm.ensureStack = vm.ensureStack[:0]
		vm.ip = 0
	} else if vm.currentProgram != nil {
		// When resuming after a yield, restore the program that was active
		// at the time of the yield (may differ from the parameter if inline
		// call frames were in use).
		program = vm.currentProgram
	}
	vm.currentProgram = program
	return program
}
