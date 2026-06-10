package interpreter

// recordBytecodeStackDiagnostics samples VM stack and call-frame state at an
// instruction boundary. It is entirely disabled outside explicit bytecode
// statistics runs, so normal execution keeps its existing hot path.
func (vm *bytecodeVM) recordBytecodeStackDiagnostics() {
	if vm == nil || vm.interp == nil || !vm.interp.bytecodeStatsEnabled {
		return
	}

	stackCapacity := vm.stackCapacity()
	capacityGrew := vm.bytecodeStatsStackCapacityObserved && stackCapacity > vm.bytecodeStatsStackCapacity
	vm.bytecodeStatsStackCapacity = stackCapacity
	vm.bytecodeStatsStackCapacityObserved = true
	vm.interp.recordBytecodeVMDepths(
		vm.stackDepth(),
		stackCapacity,
		len(vm.callFrameKinds)+vm.selfFastMinimalSuffix,
		capacityGrew,
	)
}

func (vm *bytecodeVM) beginBytecodeInstructionDiagnostics(op bytecodeOp, ip int, instr *bytecodeInstruction) {
	if vm == nil || vm.interp == nil || !vm.interp.bytecodeStatsEnabled {
		return
	}
	vm.flushBytecodeInstructionStackDelta()
	vm.recordBytecodeStackDiagnostics()
	vm.bytecodeStatsPendingOp = op
	vm.bytecodeStatsPendingIP = ip
	vm.bytecodeStatsPendingProgram = vm.currentProgram
	vm.bytecodeStatsPendingStackDepth = vm.stackDepth()
	vm.bytecodeStatsPendingPeakSite = vm.interp.bytecodeStackPeakSite(op, ip, instr)
	vm.bytecodeStatsPendingCallOperand = vm.beginBytecodeCallOperandRegion(instr)
	vm.bytecodeStatsPendingOpObserved = true
}

func (vm *bytecodeVM) finishBytecodeInstructionDiagnostics() {
	if vm == nil || vm.interp == nil || !vm.interp.bytecodeStatsEnabled {
		return
	}
	vm.flushBytecodeInstructionStackDelta()
	vm.recordBytecodeStackDiagnostics()
}

func (vm *bytecodeVM) flushBytecodeInstructionStackDelta() {
	if vm == nil || vm.interp == nil || !vm.bytecodeStatsPendingOpObserved {
		return
	}
	delta := vm.stackDepth() - vm.bytecodeStatsPendingStackDepth
	vm.interp.recordBytecodeValueStackDelta(vm.bytecodeStatsPendingOp, vm.bytecodeStatsPendingStackDepth, vm.stackDepth())
	vm.interp.recordBytecodeValueStackDeltaSite(vm.bytecodeStatsPendingPeakSite, delta)
	if vm.bytecodeStatsPendingOp == bytecodeOpJump && vm.currentProgram == vm.bytecodeStatsPendingProgram && vm.ip < vm.bytecodeStatsPendingIP {
		vm.interp.recordBytecodeLoopBackedgeBalance(vm.bytecodeStatsPendingPeakSite, vm.bytecodeStatsPendingProgram, vm.stackDepth())
	}
	if stackDepth := vm.stackDepth(); stackDepth > vm.bytecodeStatsStackPeakDepth {
		vm.interp.recordBytecodeValueStackPeakSite(vm.bytecodeStatsPendingPeakSite, stackDepth-vm.bytecodeStatsStackPeakDepth)
		vm.bytecodeStatsStackPeakDepth = stackDepth
	}
	vm.bytecodeStatsPendingOpObserved = false
}

func (vm *bytecodeVM) beginBytecodeCallOperandRegion(instr *bytecodeInstruction) bytecodeCallOperandRegion {
	if vm == nil || vm.interp == nil || !vm.interp.bytecodeStatsEnabled || instr == nil || instr.argCount < 0 {
		return bytecodeCallOperandRegion{}
	}
	operandValues := instr.argCount
	switch instr.op {
	case bytecodeOpCall:
		operandValues++ // callee
	case bytecodeOpCallName:
		if instr.slotArgs {
			return bytecodeCallOperandRegion{
				site:            vm.interp.bytecodeStackPeakSite(instr.op, vm.ip, instr),
				base:            vm.stackDepth(),
				operandValues:   operandValues,
				expectedResults: 1,
				valid:           true,
			}
		}
	case bytecodeOpCallMember, bytecodeOpCallGenericUnionMember, bytecodeOpCallStaticMember, bytecodeOpCallMemberArrayGet, bytecodeOpCallMemberNext, bytecodeOpCallMemberArrayNew, bytecodeOpCallMemberArraySlot:
		operandValues++ // receiver
	case bytecodeOpCallSelfIntSubSlotConst:
		operandValues = 0 // operand is computed from a slot and an immediate
	case bytecodeOpCallSelf:
		// Explicit arguments are already present on the value stack.
	default:
		return bytecodeCallOperandRegion{}
	}
	base := vm.stackDepth() - operandValues
	if base < 0 {
		return bytecodeCallOperandRegion{}
	}
	return bytecodeCallOperandRegion{
		site:            vm.interp.bytecodeStackPeakSite(instr.op, vm.ip, instr),
		base:            base,
		operandValues:   operandValues,
		expectedResults: 1,
		valid:           true,
	}
}

func (vm *bytecodeVM) completeBytecodeCallOperandRegion(region bytecodeCallOperandRegion, stackDepth int) {
	if vm == nil || vm.interp == nil || !region.valid || stackDepth < region.base {
		return
	}
	vm.interp.recordBytecodeCallOperandBalance(region.site, region.operandValues, region.expectedResults, stackDepth-region.base)
}

func (vm *bytecodeVM) deferBytecodeInlineCallOperandRegion(region bytecodeCallOperandRegion) {
	if vm == nil || !region.valid {
		return
	}
	vm.bytecodeStatsInlineCallOperands = append(vm.bytecodeStatsInlineCallOperands, region)
}

func (vm *bytecodeVM) completeBytecodeInlineCallOperandRegion(stackDepth int) {
	if vm == nil || len(vm.bytecodeStatsInlineCallOperands) == 0 {
		return
	}
	idx := len(vm.bytecodeStatsInlineCallOperands) - 1
	region := vm.bytecodeStatsInlineCallOperands[idx]
	vm.bytecodeStatsInlineCallOperands = vm.bytecodeStatsInlineCallOperands[:idx]
	vm.completeBytecodeCallOperandRegion(region, stackDepth)
}

func (vm *bytecodeVM) finishBytecodeCallOperandRegion(newProgram *bytecodeProgram, err error) {
	if vm == nil {
		return
	}
	region := vm.bytecodeStatsPendingCallOperand
	vm.bytecodeStatsPendingCallOperand = bytecodeCallOperandRegion{}
	if newProgram != nil {
		vm.deferBytecodeInlineCallOperandRegion(region)
		return
	}
	if err == nil {
		vm.completeBytecodeCallOperandRegion(region, vm.stackDepth())
	}
}
