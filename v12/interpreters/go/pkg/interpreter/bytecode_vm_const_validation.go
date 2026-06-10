package interpreter

func (vm *bytecodeVM) validatedIntegerConstSlots(program *bytecodeProgram) []bool {
	if vm == nil || program == nil {
		return nil
	}
	instructionCount := len(program.instructions)
	if program.integerConstValidationKnown && program.integerConstInstructionCount == instructionCount && !program.hasIntegerConstValidation {
		return nil
	}
	if vm.validatedIntConstsHotProgram == program && len(vm.validatedIntConstsHotValues) == instructionCount {
		return vm.validatedIntConstsHotValues
	}
	if vm.validatedIntConstsHotAltProgram == program && len(vm.validatedIntConstsHotAltValues) == instructionCount {
		vm.promoteValidatedIntegerConstSlotsHotAlt()
		return vm.validatedIntConstsHotValues
	}
	for _, cached := range vm.validatedIntConstsDirect {
		if cached.program != program || len(cached.values) != instructionCount {
			continue
		}
		vm.cacheValidatedIntegerConstSlotsHot(program, cached.values)
		return cached.values
	}
	if vm.validatedIntConsts == nil {
		vm.validatedIntConsts = make(map[*bytecodeProgram][]bool)
	}
	validated, ok := vm.validatedIntConsts[program]
	if ok && len(validated) == instructionCount {
		vm.cacheValidatedIntegerConstSlots(program, validated)
		return validated
	}
	validated = make([]bool, instructionCount)
	vm.validatedIntConsts[program] = validated
	vm.cacheValidatedIntegerConstSlots(program, validated)
	return validated
}

func (vm *bytecodeVM) cacheValidatedIntegerConstSlots(program *bytecodeProgram, values []bool) {
	if vm == nil || program == nil {
		return
	}
	vm.cacheValidatedIntegerConstSlotsHot(program, values)
	entry := bytecodeValidatedIntConstDirectCacheEntry{program: program, values: values}
	entries := vm.validatedIntConstsDirect[:]
	for idx := range entries {
		if entries[idx].program != program {
			continue
		}
		entries[idx] = entry
		return
	}
	insert := int(vm.validatedIntConstsDirectNext % bytecodeProgramMetadataDirectCacheSize)
	entries[insert] = entry
	vm.validatedIntConstsDirectNext = (vm.validatedIntConstsDirectNext + 1) % bytecodeProgramMetadataDirectCacheSize
}

func (vm *bytecodeVM) cacheValidatedIntegerConstSlotsHot(program *bytecodeProgram, values []bool) {
	if vm == nil || program == nil {
		return
	}
	if vm.validatedIntConstsHotProgram == program {
		vm.validatedIntConstsHotValues = values
		return
	}
	if vm.validatedIntConstsHotAltProgram == program {
		vm.validatedIntConstsHotAltValues = values
		vm.promoteValidatedIntegerConstSlotsHotAlt()
		return
	}
	vm.validatedIntConstsHotAltProgram = vm.validatedIntConstsHotProgram
	vm.validatedIntConstsHotAltValues = vm.validatedIntConstsHotValues
	vm.validatedIntConstsHotProgram = program
	vm.validatedIntConstsHotValues = values
}

func (vm *bytecodeVM) promoteValidatedIntegerConstSlotsHotAlt() {
	if vm == nil {
		return
	}
	vm.validatedIntConstsHotProgram, vm.validatedIntConstsHotAltProgram = vm.validatedIntConstsHotAltProgram, vm.validatedIntConstsHotProgram
	vm.validatedIntConstsHotValues, vm.validatedIntConstsHotAltValues = vm.validatedIntConstsHotAltValues, vm.validatedIntConstsHotValues
}
