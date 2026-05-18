package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

type bytecodeF64ArrayCacheEntry struct {
	revision uint64
	length   int
	values   []float64
}

func (vm *bytecodeVM) execLoopEnterOpcode(program **bytecodeProgram, _ *[]bytecodeInstruction, _ *[]bool, _ **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction) (bool, error) {
	if instr == nil {
		return false, fmt.Errorf("bytecode loop enter missing instruction")
	}
	if active := programValue(program); active != nil {
		if active.f64AffineRowLoops != nil {
			if plan, ok := active.f64AffineRowLoops[vm.ip]; ok {
				if handled, err := vm.tryExecF64AffineRowLoop(active, plan); handled || err != nil {
					return handled, err
				}
			}
		}
		if active.f64TransposeRowLoops != nil {
			if plan, ok := active.f64TransposeRowLoops[vm.ip]; ok {
				if handled, err := vm.tryExecF64TransposeRowLoop(active, plan); handled || err != nil {
					return handled, err
				}
			}
		}
		if active.f64MatrixRowLoops != nil {
			if plan, ok := active.f64MatrixRowLoops[vm.ip]; ok {
				if handled, err := vm.tryExecF64MatrixRowLoop(active, plan); handled || err != nil {
					return handled, err
				}
			}
		}
		if active.f64DotLoops != nil {
			if plan, ok := active.f64DotLoops[vm.ip]; ok {
				if handled, err := vm.tryExecF64DotLoop(active, plan); handled || err != nil {
					return handled, err
				}
			}
		}
	}
	if err := vm.pushLoopFrame(instr.loopBreak, instr.loopContinue); err != nil {
		return false, err
	}
	vm.ip++
	return false, nil
}

func (vm *bytecodeVM) tryExecF64DotLoop(program *bytecodeProgram, plan bytecodeF64DotLoopPlan) (bool, error) {
	if !plan.validForSlots(len(vm.slots)) || plan.successTarget <= vm.ip {
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
	resultAppend := plan.resultAppend
	acc, ok := bytecodeDirectF64Value(vm.slots[plan.accumulatorSlot])
	if !ok {
		return false, nil
	}
	if index >= bound {
		if resultAppend {
			if !vm.appendF64DotLoopResult(program, plan, acc) {
				return false, nil
			}
			vm.ip = plan.resultTarget
			return true, nil
		}
		vm.stack = append(vm.stack, runtime.NilValue{})
		vm.ip = plan.successTarget
		return true, nil
	}
	leftArr, ok := vm.slots[plan.leftReceiverSlot].(*runtime.ArrayValue)
	if !ok || leftArr == nil {
		return false, nil
	}
	rightArr, ok := vm.slots[plan.rightReceiverSlot].(*runtime.ArrayValue)
	if !ok || rightArr == nil {
		return false, nil
	}
	if !vm.canUseValidatedCanonicalArrayGet(program, leftArr) || !vm.canUseValidatedCanonicalArrayGet(program, rightArr) {
		return false, nil
	}
	leftValues, ok := vm.f64DotLoopFloatValues(leftArr)
	if !ok {
		return false, nil
	}
	rightValues, ok := vm.f64DotLoopFloatValues(rightArr)
	if !ok {
		return false, nil
	}
	start, end, ok := bytecodeF64DotLoopRange(index, bound, len(leftValues), len(rightValues))
	if !ok {
		return false, nil
	}
	for idx := start; idx < end; idx++ {
		acc += leftValues[idx] * rightValues[idx]
	}
	index = bound
	iterated := start < end
	if iterated && !resultAppend {
		vm.storeOwnedFloatSlot(plan.accumulatorSlot, runtime.FloatValue{Val: acc, TypeSuffix: runtime.FloatF64})
		if plan.accumulatorSlot == 0 {
			vm.clearSelfFastSlot0I32()
		}
		vm.storeI32Slot(plan.indexSlot, index)
	}
	if index < bound {
		return false, nil
	}
	if resultAppend {
		if !vm.appendF64DotLoopResult(program, plan, acc) {
			return false, nil
		}
		if iterated {
			vm.storeOwnedFloatSlot(plan.accumulatorSlot, runtime.FloatValue{Val: acc, TypeSuffix: runtime.FloatF64})
			if plan.accumulatorSlot == 0 {
				vm.clearSelfFastSlot0I32()
			}
			vm.storeI32Slot(plan.indexSlot, index)
		}
		vm.ip = plan.resultTarget
		return true, nil
	}
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.ip = plan.successTarget
	return true, nil
}

func bytecodeF64DotLoopRange(index, bound int64, leftLen, rightLen int) (int, int, bool) {
	if index < 0 || bound < 0 {
		return 0, 0, false
	}
	start := int(index)
	end := int(bound)
	if start < 0 || end < start || end > leftLen || end > rightLen {
		return 0, 0, false
	}
	return start, end, true
}

func (vm *bytecodeVM) appendF64DotLoopResult(program *bytecodeProgram, plan bytecodeF64DotLoopPlan, value float64) bool {
	if !plan.resultAppend || plan.resultReceiverSlot < 0 || plan.resultReceiverSlot >= len(vm.slots) {
		return false
	}
	dest, ok := vm.slots[plan.resultReceiverSlot].(*runtime.ArrayValue)
	if !ok || dest == nil {
		return false
	}
	instr := bytecodeInstruction{op: bytecodeOpCallMemberArraySlot, name: "push", argCount: 1}
	if !vm.canUseCanonicalArrayPushAt(program, plan.resultPushIP, instr, dest) {
		return false
	}
	return vm.appendArrayF64ValueFast(dest, value)
}

func bytecodeF64DotLoopI32Value(value runtime.Value) (int64, bool) {
	if raw, ok := bytecodeDirectSmallI32Value(value); ok {
		return raw, true
	}
	raw, ok := bytecodeRawI32Value(value)
	return int64(raw), ok
}

func (plan bytecodeF64DotLoopPlan) validForSlots(slotCount int) bool {
	ok := plan.accumulatorSlot >= 0 && plan.accumulatorSlot < slotCount &&
		plan.indexSlot >= 0 && plan.indexSlot < slotCount &&
		plan.boundSlot >= 0 && plan.boundSlot < slotCount &&
		plan.leftReceiverSlot >= 0 && plan.leftReceiverSlot < slotCount &&
		plan.rightReceiverSlot >= 0 && plan.rightReceiverSlot < slotCount
	if !ok {
		return false
	}
	if plan.resultAppend {
		return plan.resultReceiverSlot >= 0 &&
			plan.resultReceiverSlot < slotCount &&
			plan.resultPushIP >= 0 &&
			plan.resultTarget > 0
	}
	return true
}

func (vm *bytecodeVM) f64DotLoopFloatValues(arr *runtime.ArrayValue) ([]float64, bool) {
	if arr == nil || !vm.arrayGetPrimitiveNoError("f64") {
		return nil, false
	}
	if arr.Handle != 0 {
		values, ok, err := runtime.ArrayStoreMonoF64ValuesIfAvailable(arr.Handle)
		if err != nil {
			return nil, false
		}
		if ok {
			return values, true
		}
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state == nil {
			return nil, false
		}
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeF64 {
			return nil, false
		}
		if cached, ok := vm.f64ArrayCache[state]; ok && cached.revision == state.Revision && cached.length == len(state.Values) {
			return cached.values, true
		}
		values := make([]float64, len(state.Values))
		for idx, value := range state.Values {
			raw, ok := bytecodeDirectF64Value(value)
			if !ok {
				return nil, false
			}
			values[idx] = raw
		}
		if vm.f64ArrayCache == nil {
			vm.f64ArrayCache = make(map[*runtime.ArrayState]bytecodeF64ArrayCacheEntry, 8)
		}
		vm.f64ArrayCache[state] = bytecodeF64ArrayCacheEntry{
			revision: state.Revision,
			length:   len(state.Values),
			values:   values,
		}
		return values, true
	}
	return nil, false
}

func bytecodeDirectF64Value(value runtime.Value) (float64, bool) {
	switch fv := value.(type) {
	case runtime.FloatValue:
		if fv.TypeSuffix == runtime.FloatF64 {
			return fv.Val, true
		}
	case *runtime.FloatValue:
		if fv != nil && fv.TypeSuffix == runtime.FloatF64 {
			return fv.Val, true
		}
	}
	return 0, false
}

func (vm *bytecodeVM) storeI32Slot(slot int, value int64) {
	boxed := bytecodeBoxedIntegerI32Value(value)
	vm.slots[slot] = boxed
	if slot == 0 {
		vm.setSelfFastSlot0I32Value(boxed)
	}
}
