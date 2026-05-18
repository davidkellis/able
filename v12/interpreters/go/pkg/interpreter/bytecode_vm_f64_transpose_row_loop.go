package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) tryExecF64TransposeRowLoop(program *bytecodeProgram, plan bytecodeF64TransposeRowLoopPlan) (bool, error) {
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
	start, end, ok := bytecodeF64TransposeRowLoopRange(index, bound)
	if !ok {
		return false, nil
	}
	col, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.colIndexSlot])
	if !ok || col < 0 {
		return false, nil
	}
	dest, ok := vm.slots[plan.receiverSlot].(*runtime.ArrayValue)
	if !ok || dest == nil {
		return false, nil
	}
	outer, ok := vm.slots[plan.outerSlot].(*runtime.ArrayValue)
	if !ok || outer == nil || bytecodeSameArrayStorage(dest, outer) {
		return false, nil
	}
	if !vm.arrayValueNoErrorForPropagation() || !vm.arrayGetPrimitiveNoError("f64") || !vm.canUseValidatedCanonicalArrayGet(program, outer) {
		return false, nil
	}
	instr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1}
	if !vm.canUseCanonicalArrayPushAt(program, plan.resultPushIP, instr, dest) {
		return false, nil
	}
	values := make([]float64, end-start)
	for rowIdx := start; rowIdx < end; rowIdx++ {
		rowValue, _, _, handled, err := vm.readCanonicalArrayGetValue(outer, int64(rowIdx))
		if err != nil || !handled {
			return false, nil
		}
		row, ok := rowValue.(*runtime.ArrayValue)
		if !ok || row == nil || bytecodeSameArrayStorage(dest, row) || !vm.canUseValidatedCanonicalArrayGet(program, row) {
			return false, nil
		}
		value, ok := vm.readCanonicalArrayGetF64Raw(row, col)
		if !ok {
			return false, nil
		}
		values[rowIdx-start] = value
	}
	if !vm.appendArrayF64ValuesFast(dest, values) {
		return false, nil
	}
	vm.storeI32Slot(plan.indexSlot, int64(end))
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.ip = plan.successTarget
	return true, nil
}

func bytecodeF64TransposeRowLoopRange(index int64, bound int64) (int, int, bool) {
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
