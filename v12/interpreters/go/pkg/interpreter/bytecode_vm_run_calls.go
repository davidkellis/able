package interpreter

func (vm *bytecodeVM) execCallOpcode(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable, program *bytecodeProgram) (*bytecodeProgram, error) {
	if instr == nil {
		return nil, nil
	}
	switch instr.op {
	case bytecodeOpCall:
		return vm.execCall(*instr, program)
	case bytecodeOpCallName:
		return vm.execCallName(*instr, program)
	case bytecodeOpCallMember:
		return vm.execCallMember(*instr, program)
	case bytecodeOpCallGenericUnionMember:
		return vm.execCallGenericUnionMember(*instr, program)
	case bytecodeOpCallStaticMember:
		return vm.execCallStaticMember(*instr, program)
	case bytecodeOpCallMemberArrayGet:
		return vm.execCallMemberArrayGet(*instr, program)
	case bytecodeOpCallMemberNext:
		return vm.execCallMemberNext(*instr, program)
	case bytecodeOpCallMemberArrayNew:
		return vm.execCallMemberArrayNew(*instr, program)
	case bytecodeOpCallMemberArraySlot:
		return vm.execCallMemberArraySlot(instr, program)
	case bytecodeOpCallSelf:
		return vm.execCallSelf(*instr, program)
	case bytecodeOpCallSelfIntSubSlotConst:
		return vm.execCallSelfIntSubSlotConst(instr, slotConstIntImmTable, program)
	default:
		return nil, nil
	}
}
