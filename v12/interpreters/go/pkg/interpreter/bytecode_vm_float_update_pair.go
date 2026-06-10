package interpreter

func (vm *bytecodeVM) execTryFloatUpdatePair(program *bytecodeProgram, instr *bytecodeInstruction) error {
	if vm == nil || instr == nil {
		return nil
	}
	if program == nil {
		program = vm.currentProgram
	}
	plan, ok := bytecodeFloatUpdatePairPlanAt(program, vm.ip)
	if !ok || instr.target <= vm.ip || !plan.validForSlots(len(vm.slots)) {
		vm.ip++
		return nil
	}

	firstScaleRaw, firstScaleKind, ok := vm.binaryFloatMulSlotConstFastRaw(plan.firstScaleSourceSlot, plan.firstScaleImmediate.Val, plan.firstScaleImmediate.TypeSuffix)
	if !ok {
		vm.ip++
		return nil
	}
	firstBaseRaw, firstBaseKind, ok := vm.slotDirectFloatValue(plan.firstBaseSlot)
	if !ok {
		vm.ip++
		return nil
	}
	firstMulRaw, firstMulKind, ok := vm.slotDirectFloatValue(plan.firstMulSlot)
	if !ok {
		vm.ip++
		return nil
	}
	firstRaw, firstKind, ok := bytecodeDirectFloatAddMulOperandsRawValue(firstBaseRaw, firstBaseKind, firstScaleRaw, firstScaleKind, firstMulRaw, firstMulKind)
	if !ok {
		vm.ip++
		return nil
	}

	secondBaseRaw, secondBaseKind, ok := vm.slotDirectFloatValue(plan.secondBaseSlot)
	if !ok {
		vm.ip++
		return nil
	}
	secondLeftRaw, secondLeftKind, ok := vm.slotDirectFloatValue(plan.secondSubLeftSlot)
	if !ok {
		vm.ip++
		return nil
	}
	secondRightRaw, secondRightKind, ok := vm.slotDirectFloatValue(plan.secondSubRightSlot)
	if !ok {
		vm.ip++
		return nil
	}
	secondRaw, secondKind, ok := bytecodeDirectFloatAddSubOperandsRawValue(secondBaseRaw, secondBaseKind, secondLeftRaw, secondLeftKind, secondRightRaw, secondRightKind)
	if !ok {
		vm.ip++
		return nil
	}

	vm.storeReusableNormalizedFloatSlotRawDiscard(plan.firstTargetSlot, firstRaw, firstKind)
	vm.storeReusableNormalizedFloatSlotRawDiscard(plan.secondTargetSlot, secondRaw, secondKind)
	vm.ip = instr.target
	return nil
}

func bytecodeFloatUpdatePairPlanAt(program *bytecodeProgram, ip int) (bytecodeFloatUpdatePairPlan, bool) {
	if program == nil || program.floatUpdatePairs == nil {
		return bytecodeFloatUpdatePairPlan{}, false
	}
	plan, ok := program.floatUpdatePairs[ip]
	return plan, ok
}
