package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) tryExecF64AffineRowLoop(program *bytecodeProgram, plan bytecodeF64AffineRowLoopPlan) (bool, error) {
	if vm == nil || program == nil || !plan.validForSlots(len(vm.slots)) || plan.successTarget <= vm.ip {
		return false, nil
	}
	index, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.indexSlot])
	if !ok {
		return false, nil
	}
	bound, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.boundSlot])
	if !ok {
		return false, nil
	}
	if index >= bound {
		vm.stack = append(vm.stack, runtime.NilValue{})
		vm.ip = plan.successTarget
		return true, nil
	}
	start, end, ok := bytecodeF64AffineRowLoopRange(index, bound)
	if !ok {
		return false, nil
	}
	dest, ok := vm.slots[plan.receiverSlot].(*runtime.ArrayValue)
	if !ok || dest == nil {
		return false, nil
	}
	instr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1}
	if !vm.canUseCanonicalArrayPushAt(program, plan.resultPushIP, instr, dest) {
		return false, nil
	}
	scale, ok := bytecodeDirectF64Value(vm.slots[plan.scaleSlot])
	if !ok {
		return false, nil
	}
	left, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.leftSlot])
	if !ok {
		return false, nil
	}
	values := make([]float64, end-start)
	for idx := start; idx < end; idx++ {
		right := int64(idx)
		values[idx-start] = scale * float64(left-right) * float64(left+right)
	}
	if !vm.appendArrayF64ValuesFast(dest, values) {
		return false, nil
	}
	vm.storeI32Slot(plan.indexSlot, int64(end))
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.ip = plan.successTarget
	return true, nil
}

func bytecodeF64AffineRowLoopRange(index int64, bound int64) (int, int, bool) {
	if index < 0 || bound < 0 {
		return 0, 0, false
	}
	start := int(index)
	end := int(bound)
	if start < 0 || end < start {
		return 0, 0, false
	}
	return start, end, true
}
