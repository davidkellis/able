package interpreter

func (vm *bytecodeVM) validatedIntegerConstSlots(program *bytecodeProgram) []bool {
	if vm == nil || program == nil {
		return nil
	}
	if vm.validatedIntConsts == nil {
		vm.validatedIntConsts = make(map[*bytecodeProgram][]bool)
	}
	validated, ok := vm.validatedIntConsts[program]
	if ok && len(validated) == len(program.instructions) {
		return validated
	}
	validated = make([]bool, len(program.instructions))
	vm.validatedIntConsts[program] = validated
	return validated
}
