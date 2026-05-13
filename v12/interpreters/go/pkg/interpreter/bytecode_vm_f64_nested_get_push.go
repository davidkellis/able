package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) execTryArrayPushF64NestedGet(program *bytecodeProgram, instr *bytecodeInstruction) error {
	if vm == nil || instr == nil {
		return nil
	}
	if program == nil {
		program = vm.currentProgram
	}
	plan, ok := bytecodeF64NestedArrayGetPushPlanAt(program, vm.ip)
	if !ok || instr.target <= vm.ip || !plan.validForSlots(len(vm.slots)) {
		vm.ip++
		return nil
	}
	dest, ok := vm.slots[plan.receiverSlot].(*runtime.ArrayValue)
	if !ok || dest == nil || !vm.canUseCanonicalArrayPushForF64AffineProduct(program, *instr, dest) {
		vm.ip++
		return nil
	}
	outer, ok := vm.slots[plan.outerSlot].(*runtime.ArrayValue)
	if !ok || outer == nil || !vm.arrayValueNoErrorForPropagation() || !vm.arrayGetPrimitiveNoError("f64") {
		vm.ip++
		return nil
	}
	rowIdx, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.rowIndexSlot])
	if !ok {
		vm.ip++
		return nil
	}
	colIdx, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.colIndexSlot])
	if !ok {
		vm.ip++
		return nil
	}
	if !vm.canUseValidatedCanonicalArrayGet(program, outer) {
		vm.ip++
		return nil
	}
	rowValue, _, _, handled, err := vm.readCanonicalArrayGetValue(outer, rowIdx)
	if err != nil || !handled {
		vm.ip++
		return nil
	}
	row, ok := rowValue.(*runtime.ArrayValue)
	if !ok || row == nil || !vm.canUseValidatedCanonicalArrayGet(program, row) {
		vm.ip++
		return nil
	}
	value, ok := vm.readCanonicalArrayGetF64Raw(row, colIdx)
	if !ok || !vm.appendArrayF64ValueFast(dest, value) {
		vm.ip++
		return nil
	}
	if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", "push", "resolved_method", "array_push_f64_nested_get_fast", instr.node)
	}
	vm.stack = append(vm.stack, runtime.VoidValue{})
	vm.ip = instr.target
	return nil
}

func bytecodeF64NestedArrayGetPushPlanAt(program *bytecodeProgram, ip int) (bytecodeF64NestedArrayGetPushPlan, bool) {
	if program == nil || program.f64NestedGetPushes == nil {
		return bytecodeF64NestedArrayGetPushPlan{}, false
	}
	plan, ok := program.f64NestedGetPushes[ip]
	return plan, ok
}

func (vm *bytecodeVM) readCanonicalArrayGetF64Raw(arr *runtime.ArrayValue, idx int64) (float64, bool) {
	if arr == nil || idx < 0 || idx > 1<<31-1 {
		return 0, false
	}
	if arr.Handle != 0 {
		values, monoF64, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(arr.Handle)
		if err != nil {
			return 0, false
		}
		if monoF64 {
			if idx >= int64(len(values)) {
				return 0, false
			}
			return values[int(idx)], true
		}
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil || idx >= int64(len(state.Values)) {
			return 0, false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeF64 {
			return 0, false
		}
		return bytecodeDirectF64Value(state.Values[int(idx)])
	}
	return 0, false
}
