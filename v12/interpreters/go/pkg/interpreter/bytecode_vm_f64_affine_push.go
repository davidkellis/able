package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) execTryArrayPushF64AffineProduct(program *bytecodeProgram, instr *bytecodeInstruction) error {
	if vm == nil || instr == nil {
		return nil
	}
	if program == nil {
		program = vm.currentProgram
	}
	plan, ok := bytecodeF64AffineProductPushPlanAt(program, vm.ip)
	if !ok || instr.target <= vm.ip || !plan.validForSlots(len(vm.slots)) {
		vm.ip++
		return nil
	}
	arr, ok := vm.slots[plan.receiverSlot].(*runtime.ArrayValue)
	if !ok || arr == nil || !vm.canUseCanonicalArrayPushForF64AffineProduct(program, *instr, arr) {
		vm.ip++
		return nil
	}
	scale, ok := bytecodeDirectF64Value(vm.slots[plan.scaleSlot])
	if !ok {
		vm.ip++
		return nil
	}
	left, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.leftSlot])
	if !ok {
		vm.ip++
		return nil
	}
	right, ok := bytecodeF64DotLoopI32Value(vm.slots[plan.rightSlot])
	if !ok {
		vm.ip++
		return nil
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	value := runtime.FloatValue{
		Val:        scale * float64(left-right) * float64(left+right),
		TypeSuffix: runtime.FloatF64,
	}
	if vm.appendArrayF64ValueFast(arr, value.Val) {
		if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
			vm.interp.recordBytecodeCallTrace("call_member", "push", "resolved_method", "array_push_f64_mono_fast", instr.node)
		}
		vm.stack = append(vm.stack, runtime.VoidValue{})
		vm.ip = instr.target
		return nil
	}
	if !tracked {
		if vm.interp == nil {
			vm.ip++
			return nil
		}
		var err error
		state, err = vm.interp.ensureArrayState(arr, 0)
		if err != nil {
			vm.ip++
			return nil
		}
	}
	vm.appendTrackedArrayValueFast(arr, state, value)
	if vm.interp != nil && vm.interp.bytecodeTraceEnabled {
		vm.interp.recordBytecodeCallTrace("call_member", "push", "resolved_method", "array_push_f64_affine_fast", instr.node)
	}
	vm.stack = append(vm.stack, runtime.VoidValue{})
	vm.ip = instr.target
	return nil
}

func (vm *bytecodeVM) appendArrayF64ValueFast(arr *runtime.ArrayValue, value float64) bool {
	if vm == nil || vm.interp == nil || arr == nil || arr.Handle == 0 || arr.TrackedAliases {
		return false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeF64 && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return false
		}
	}
	ok, err := runtime.ArrayStoreAppendF64Promote(arr.Handle, value)
	if err != nil || !ok {
		return false
	}
	arr.State = nil
	arr.Elements = nil
	arr.TrackedHandle = arr.Handle
	return true
}

func bytecodeF64AffineProductPushPlanAt(program *bytecodeProgram, ip int) (bytecodeF64AffineProductPushPlan, bool) {
	if program == nil || program.f64AffinePushes == nil {
		return bytecodeF64AffineProductPushPlan{}, false
	}
	plan, ok := program.f64AffinePushes[ip]
	return plan, ok
}

func (vm *bytecodeVM) canUseCanonicalArrayPushForF64AffineProduct(program *bytecodeProgram, instr bytecodeInstruction, arr *runtime.ArrayValue) bool {
	return vm.canUseCanonicalArrayPushAt(program, vm.ip, instr, arr)
}

func (vm *bytecodeVM) canUseCanonicalArrayPushAt(program *bytecodeProgram, ip int, instr bytecodeInstruction, arr *runtime.ArrayValue) bool {
	if vm == nil || vm.interp == nil || vm.env == nil || arr == nil || !vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		return false
	}
	const kind = bytecodeMemberMethodFastPathArrayPush
	if vm.lookupCachedCanonicalArraySlotCallForArray(program, ip, kind) {
		return true
	}
	callable, found, err := vm.interp.resolveMethodCallableFromPool(vm.env, "push", arr, "")
	if err != nil || !found {
		return false
	}
	fn, ok := bytecodeResolvedMemberFastPathFunction(callable)
	if !ok || vm.resolvedMemberMethodFastPath("push", arr, fn) != kind {
		return false
	}
	cacheInstr := bytecodeInstruction{
		op:       bytecodeOpCallMemberArraySlot,
		name:     "push",
		argCount: 1,
		node:     instr.node,
	}
	vm.storeCachedCanonicalArraySlotCall(program, ip, cacheInstr, arr, kind)
	return true
}
