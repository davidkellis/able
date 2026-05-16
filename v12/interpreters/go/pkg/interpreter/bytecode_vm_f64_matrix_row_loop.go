package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) tryExecF64MatrixRowLoop(program *bytecodeProgram, plan bytecodeF64MatrixRowLoopPlan) (bool, error) {
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
	dest, ok := vm.slots[plan.resultReceiverSlot].(*runtime.ArrayValue)
	if !ok || dest == nil {
		return false, nil
	}
	left, ok := vm.slots[plan.leftReceiverSlot].(*runtime.ArrayValue)
	if !ok || left == nil {
		return false, nil
	}
	outer, ok := vm.slots[plan.rightOuterSlot].(*runtime.ArrayValue)
	if !ok || outer == nil {
		return false, nil
	}
	if bytecodeSameArrayStorage(dest, left) || bytecodeSameArrayStorage(dest, outer) {
		return false, nil
	}
	if !vm.arrayValueNoErrorForPropagation() || !vm.canUseValidatedCanonicalArrayGet(program, outer) || !vm.canUseValidatedCanonicalArrayGet(program, left) {
		return false, nil
	}
	leftValues, ok := vm.f64DotLoopFloatValues(left)
	if !ok {
		return false, nil
	}
	start, end, ok := bytecodeF64MatrixRowLoopRange(index, bound, len(leftValues), outer)
	if !ok {
		return false, nil
	}
	results := make([]float64, end-start)
	for rowIdx := start; rowIdx < end; rowIdx++ {
		rowValue, _, _, handled, err := vm.readCanonicalArrayGetValue(outer, int64(rowIdx))
		if err != nil {
			return false, err
		}
		if !handled {
			return false, nil
		}
		row, ok := rowValue.(*runtime.ArrayValue)
		if !ok || row == nil || !vm.canUseValidatedCanonicalArrayGet(program, row) {
			return false, nil
		}
		if bytecodeSameArrayStorage(dest, row) {
			return false, nil
		}
		rowValues, ok := vm.f64DotLoopFloatValues(row)
		if !ok || len(rowValues) < int(bound) {
			return false, nil
		}
		acc := 0.0
		for colIdx := 0; colIdx < int(bound); colIdx++ {
			acc += leftValues[colIdx] * rowValues[colIdx]
		}
		results[rowIdx-start] = acc
	}
	if !vm.appendF64MatrixRowLoopResults(program, plan, dest, results) {
		return false, nil
	}
	vm.storeI32Slot(plan.indexSlot, int64(end))
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.ip = plan.successTarget
	return true, nil
}

func bytecodeSameArrayStorage(left *runtime.ArrayValue, right *runtime.ArrayValue) bool {
	if left == nil || right == nil {
		return false
	}
	if left == right {
		return true
	}
	leftHandle := left.Handle
	if leftHandle == 0 {
		leftHandle = left.TrackedHandle
	}
	rightHandle := right.Handle
	if rightHandle == 0 {
		rightHandle = right.TrackedHandle
	}
	return leftHandle != 0 && leftHandle == rightHandle
}

func bytecodeF64MatrixRowLoopRange(index, bound int64, leftLen int, outer *runtime.ArrayValue) (int, int, bool) {
	if index < 0 || bound < 0 || bound > int64(leftLen) {
		return 0, 0, false
	}
	start := int(index)
	end := int(bound)
	if start < 0 || end < start {
		return 0, 0, false
	}
	if state, tracked := bytecodeTrackedArrayState(outer); tracked {
		return start, end, state != nil && end <= len(state.Values)
	}
	if outer == nil || outer.Handle == 0 {
		return 0, 0, false
	}
	size, err := runtime.ArrayStoreSize(outer.Handle)
	if err != nil || size < 0 {
		return 0, 0, false
	}
	return start, end, end <= size
}

func (vm *bytecodeVM) appendF64MatrixRowLoopResults(program *bytecodeProgram, plan bytecodeF64MatrixRowLoopPlan, dest *runtime.ArrayValue, values []float64) bool {
	if dest == nil {
		return false
	}
	instr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1}
	if !vm.canUseCanonicalArrayPushAt(program, plan.resultPushIP, instr, dest) {
		return false
	}
	return vm.appendArrayF64ValuesFast(dest, values)
}

func (vm *bytecodeVM) appendArrayF64ValuesFast(arr *runtime.ArrayValue, values []float64) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if len(values) == 0 {
		return true
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeF64 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendF64ValuesPromote(arr.Handle, values)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}
