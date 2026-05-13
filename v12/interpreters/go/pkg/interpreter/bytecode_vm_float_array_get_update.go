package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeFusedArrayGetValue struct {
	value      runtime.Value
	typeToken  uint16
	tokenKnown bool
}

type bytecodeFusedArrayGetFloat struct {
	value float64
	kind  runtime.FloatType
}

func (vm *bytecodeVM) execStoreSlotFloatAddMulArrayGet(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction) (bool, error) {
	if instr == nil {
		return false, fmt.Errorf("bytecode float array-get slot update missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return false, fmt.Errorf("bytecode slot out of range")
	}
	if len(vm.stack) < 4 {
		return false, fmt.Errorf("bytecode stack underflow")
	}
	baseIdx := len(vm.stack) - 4
	base := vm.slots[instr.target]
	leftReceiver := vm.stack[baseIdx]
	leftIndex := vm.stack[baseIdx+1]
	rightReceiver := vm.stack[baseIdx+2]
	rightIndex := vm.stack[baseIdx+3]

	if handled, ok, err := vm.tryStoreSlotFloatAddMulRawArrayGet(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, baseIdx, base, leftReceiver, leftIndex, rightReceiver, rightIndex); handled || ok || err != nil {
		return handled, err
	}

	left, handled, err := vm.fusedArrayGetPropagatedValue(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, baseIdx, leftReceiver, leftIndex)
	if handled || err != nil {
		return handled, err
	}
	right, handled, err := vm.fusedArrayGetPropagatedValue(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, baseIdx, rightReceiver, rightIndex)
	if handled || err != nil {
		return handled, err
	}

	if result, ok := bytecodeDirectFloatAddMulValue(base, left, right); ok {
		vm.stack = vm.stack[:baseIdx]
		vm.storeOwnedFloatSlot(instr.target, result)
		if instr.target == 0 {
			vm.clearSelfFastSlot0I32()
		}
		if !instr.discardResult {
			vm.stack = append(vm.stack, result)
		}
		vm.ip++
		return false, nil
	}

	result, err := vm.storeSlotFloatAddMulResult(instr, base, left, right)
	if err != nil {
		return false, vm.wrapFloatAddMulArrayGetError(instr, err)
	}
	result = bytecodeStackResultValue(result)
	vm.stack = vm.stack[:baseIdx]
	vm.slots[instr.target] = result
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(result)
	}
	if !instr.discardResult {
		vm.stack = append(vm.stack, result)
	}
	vm.ip++
	return false, nil
}

func (vm *bytecodeVM) tryStoreSlotFloatAddMulRawArrayGet(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction, stackBase int, base runtime.Value, leftReceiver runtime.Value, leftIndex runtime.Value, rightReceiver runtime.Value, rightIndex runtime.Value) (bool, bool, error) {
	leftArr, leftIdx, ok := bytecodeFusedArrayGetArrayIndex(leftReceiver, leftIndex)
	if !ok {
		return false, false, nil
	}
	rightArr, rightIdx, ok := bytecodeFusedArrayGetArrayIndex(rightReceiver, rightIndex)
	if !ok || !vm.canUseValidatedCanonicalArrayGet(programValue(program), leftArr) {
		return false, false, nil
	}
	left, handled, ok, err := vm.fusedCanonicalArrayGetPropagatedFloat(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, stackBase, leftArr, leftIdx)
	if handled || err != nil || !ok {
		return handled, false, err
	}
	right, handled, ok, err := vm.fusedCanonicalArrayGetPropagatedFloat(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, stackBase, rightArr, rightIdx)
	if handled || err != nil || !ok {
		return handled, false, err
	}
	result, ok := bytecodeDirectFloatAddMulRaw(base, left.value, left.kind, right.value, right.kind)
	if !ok {
		return false, false, nil
	}
	vm.stack = vm.stack[:stackBase]
	vm.storeOwnedFloatSlot(instr.target, result)
	if instr.target == 0 {
		vm.clearSelfFastSlot0I32()
	}
	if !instr.discardResult {
		vm.stack = append(vm.stack, result)
	}
	vm.ip++
	return false, true, nil
}

func (vm *bytecodeVM) wrapFloatAddMulArrayGetError(instr *bytecodeInstruction, err error) error {
	if vm.interp != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
	}
	if instr != nil && instr.node != nil && vm.interp != nil {
		return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
	}
	return err
}

func (vm *bytecodeVM) fusedCanonicalArrayGetPropagatedFloat(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction, stackBase int, arr *runtime.ArrayValue, idx int64) (bytecodeFusedArrayGetFloat, bool, bool, error) {
	value, token, tokenKnown, handled, err := vm.readCanonicalArrayGetValue(arr, idx)
	if err != nil {
		if instr != nil && instr.node != nil && vm != nil && vm.interp != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return bytecodeFusedArrayGetFloat{}, false, false, err
	}
	if !handled {
		return bytecodeFusedArrayGetFloat{}, false, false, nil
	}
	if isNilRuntimeValue(value) {
		vm.stack = vm.stack[:stackBase]
		val := runtime.NilValue{}
		if vm.hasCallFrames() {
			if err := vm.finishInlineReturn(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, val, bytecodeSimpleTypeCheckUnknown); err != nil {
				return bytecodeFusedArrayGetFloat{}, false, false, err
			}
			return bytecodeFusedArrayGetFloat{}, true, false, nil
		}
		return bytecodeFusedArrayGetFloat{}, false, false, returnSignal{value: val, node: instr.node}
	}
	fv, ok := bytecodeFusedArrayGetFloatForToken(value, token, tokenKnown)
	if !ok || !vm.arrayGetPrimitiveNoErrorToken(token) {
		return bytecodeFusedArrayGetFloat{}, false, false, nil
	}
	return fv, false, true, nil
}

func (vm *bytecodeVM) fusedArrayGetPropagatedValue(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction, stackBase int, receiver runtime.Value, index runtime.Value) (runtime.Value, bool, error) {
	result, err := vm.fusedArrayGetValue(programValue(program), receiver, index)
	if err != nil {
		if instr != nil && instr.node != nil && vm != nil && vm.interp != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return nil, false, err
	}
	if isNilRuntimeValue(result.value) {
		vm.stack = vm.stack[:stackBase]
		val := runtime.NilValue{}
		if vm.hasCallFrames() {
			if err := vm.finishInlineReturn(program, instructions, validatedIntConsts, slotConstIntImmTable, instr, val, bytecodeSimpleTypeCheckUnknown); err != nil {
				return nil, false, err
			}
			return nil, true, nil
		}
		return nil, false, returnSignal{value: val, node: instr.node}
	}
	if !vm.fusedArrayGetCanSkipPropagationCheck(result) {
		if errVal, ok := vm.interp.propagationErrorValue(result.value, vm.env); ok {
			vm.stack = vm.stack[:stackBase]
			return nil, false, raiseSignal{value: errVal}
		}
	}
	return result.value, false, nil
}

func bytecodeFusedArrayGetArrayIndex(receiver runtime.Value, index runtime.Value) (*runtime.ArrayValue, int64, bool) {
	arr, ok := receiver.(*runtime.ArrayValue)
	if !ok || arr == nil {
		return nil, 0, false
	}
	if idx, ok := bytecodeDirectSmallI32Value(index); ok {
		return arr, idx, true
	}
	idx, ok := bytecodeArrayGetIndexI32(index)
	return arr, idx, ok
}

func bytecodeFusedArrayGetFloatForToken(value runtime.Value, token uint16, tokenKnown bool) (bytecodeFusedArrayGetFloat, bool) {
	if !tokenKnown {
		return bytecodeFusedArrayGetFloat{}, false
	}
	switch token {
	case bytecodeIndexTypeF32:
		switch fv := value.(type) {
		case runtime.FloatValue:
			if fv.TypeSuffix == runtime.FloatF32 {
				return bytecodeFusedArrayGetFloat{value: fv.Val, kind: runtime.FloatF32}, true
			}
		case *runtime.FloatValue:
			if fv != nil && fv.TypeSuffix == runtime.FloatF32 {
				return bytecodeFusedArrayGetFloat{value: fv.Val, kind: runtime.FloatF32}, true
			}
		}
	case bytecodeIndexTypeF64:
		switch fv := value.(type) {
		case runtime.FloatValue:
			if fv.TypeSuffix == runtime.FloatF64 {
				return bytecodeFusedArrayGetFloat{value: fv.Val, kind: runtime.FloatF64}, true
			}
		case *runtime.FloatValue:
			if fv != nil && fv.TypeSuffix == runtime.FloatF64 {
				return bytecodeFusedArrayGetFloat{value: fv.Val, kind: runtime.FloatF64}, true
			}
		}
	}
	return bytecodeFusedArrayGetFloat{}, false
}

func (vm *bytecodeVM) arrayGetPrimitiveNoErrorToken(token uint16) bool {
	switch token {
	case bytecodeIndexTypeF32:
		return vm.arrayGetPrimitiveNoError("f32")
	case bytecodeIndexTypeF64:
		return vm.arrayGetPrimitiveNoError("f64")
	default:
		return false
	}
}

func (vm *bytecodeVM) fusedArrayGetCanSkipPropagationCheck(result bytecodeFusedArrayGetValue) bool {
	if !result.tokenKnown {
		return false
	}
	if !bytecodeArrayGetResultMatchesFloatToken(result.value, result.typeToken) {
		return false
	}
	switch result.typeToken {
	case bytecodeIndexTypeF32:
		return vm.arrayGetPrimitiveNoError("f32")
	case bytecodeIndexTypeF64:
		return vm.arrayGetPrimitiveNoError("f64")
	default:
		return false
	}
}

func (vm *bytecodeVM) fusedArrayGetValue(program *bytecodeProgram, receiver runtime.Value, index runtime.Value) (bytecodeFusedArrayGetValue, error) {
	if arr, ok := receiver.(*runtime.ArrayValue); ok && arr != nil {
		idx, idxOK := bytecodeArrayGetIndexI32(index)
		canUse := idxOK && vm.canUseValidatedCanonicalArrayGet(program, arr)
		if idxOK && canUse {
			value, token, tokenKnown, handled, err := vm.readCanonicalArrayGetValue(arr, idx)
			if err != nil {
				return bytecodeFusedArrayGetValue{}, err
			}
			if handled {
				return bytecodeFusedArrayGetValue{value: value, typeToken: token, tokenKnown: tokenKnown}, nil
			}
		}
	}
	value, err := vm.callArrayGetFallback(receiver, index)
	return bytecodeFusedArrayGetValue{value: value}, err
}

func (vm *bytecodeVM) canUseValidatedCanonicalArrayGet(program *bytecodeProgram, arr *runtime.ArrayValue) bool {
	if vm == nil || vm.interp == nil || arr == nil || !vm.canUseCanonicalArrayGetCallCacheForArray(arr) {
		return false
	}
	if vm.lookupCachedCanonicalArrayGetCallForArray(program, vm.ip) {
		return true
	}
	callee, err := vm.interp.memberAccessOnValueWithOptions(arr, ast.ID("get"), vm.env, true)
	canonical := err == nil && vm.isCanonicalNullableArrayGetOverload(callee)
	if !canonical {
		return false
	}
	vm.storeCachedCanonicalArrayGetCall(program, vm.ip, bytecodeInstruction{name: "get", argCount: 1}, arr)
	return true
}

func (vm *bytecodeVM) readCanonicalArrayGetValue(arr *runtime.ArrayValue, idx int64) (runtime.Value, uint16, bool, bool, error) {
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		size := len(state.Values)
		if size > 1<<31-1 {
			return nil, bytecodeIndexTypeUnknown, false, false, nil
		}
		if idx < 0 || idx >= int64(size) {
			return runtime.NilValue{}, state.ElementTypeToken, state.ElementTypeTokenKnown, true, nil
		}
		return state.Values[int(idx)], state.ElementTypeToken, state.ElementTypeTokenKnown, true, nil
	}
	handle, ok, err := vm.arrayHandleFast(arr)
	if err != nil {
		return nil, bytecodeIndexTypeUnknown, false, true, err
	}
	if !ok {
		return nil, bytecodeIndexTypeUnknown, false, false, nil
	}
	if values, monoF64, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(handle); err != nil {
		return nil, bytecodeIndexTypeUnknown, false, true, err
	} else if monoF64 {
		if len(values) > 1<<31-1 {
			return nil, bytecodeIndexTypeUnknown, false, false, nil
		}
		if idx < 0 || idx >= int64(len(values)) {
			return runtime.NilValue{}, bytecodeIndexTypeF64, true, true, nil
		}
		return runtime.FloatValue{Val: values[int(idx)], TypeSuffix: runtime.FloatF64}, bytecodeIndexTypeF64, true, true, nil
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		return nil, bytecodeIndexTypeUnknown, false, true, err
	}
	if size < 0 || size > 1<<31-1 {
		return nil, bytecodeIndexTypeUnknown, false, false, nil
	}
	token, tokenKnown := bytecodeArrayElementTypeToken(arr)
	if idx < 0 || idx >= int64(size) {
		return runtime.NilValue{}, token, tokenKnown, true, nil
	}
	value, err := runtime.ArrayStoreRead(handle, int(idx))
	return value, token, tokenKnown, true, err
}

func (vm *bytecodeVM) callArrayGetFallback(receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	callee, err := vm.interp.memberAccessOnValueWithOptions(receiver, ast.ID("get"), vm.env, true)
	if err != nil {
		return nil, err
	}
	return vm.interp.callCallableValueMutable(callee, []runtime.Value{index}, vm.env, nil)
}

func programValue(program **bytecodeProgram) *bytecodeProgram {
	if program == nil {
		return nil
	}
	return *program
}
