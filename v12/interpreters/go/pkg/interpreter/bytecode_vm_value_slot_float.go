package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) activeValueSlotFloatFrameMatchesSlots(slots []runtime.Value) bool {
	if vm == nil {
		return false
	}
	if len(slots) == 0 {
		return len(vm.slotFloatValues) == 0 &&
			len(vm.slotFloatKinds) == 0 &&
			len(vm.slotFloatValid) == 0 &&
			vm.slotFloatOwner == nil
	}
	return len(vm.slotFloatValues) == len(slots) &&
		len(vm.slotFloatKinds) == len(slots) &&
		len(vm.slotFloatValid) == len(slots) &&
		vm.slotFloatOwner == bytecodeSlotFrameOwner(slots)
}

func (vm *bytecodeVM) acquireValueSlotFloatFrame(slotCount int) ([]float64, []runtime.FloatType, []bool) {
	if slotCount <= 0 {
		return nil, nil, nil
	}
	if vm != nil && vm.slotFloatFramePool != nil {
		frames := vm.slotFloatFramePool[slotCount]
		if len(frames) > 0 {
			idx := len(frames) - 1
			frame := frames[idx]
			vm.slotFloatFramePool[slotCount] = frames[:idx]
			return frame.values, frame.kinds, frame.valid
		}
	}
	return make([]float64, slotCount), make([]runtime.FloatType, slotCount), make([]bool, slotCount)
}

func (vm *bytecodeVM) releaseValueSlotFloatFrame(values []float64, kinds []runtime.FloatType, valid []bool) {
	if vm == nil || len(values) == 0 || len(values) != len(kinds) || len(values) != len(valid) {
		return
	}
	clear(valid)
	if vm.slotFloatFramePool == nil {
		vm.slotFloatFramePool = make(map[int][]bytecodeValueSlotFloatFrame, 2)
	}
	size := len(values)
	vm.slotFloatFramePool[size] = append(vm.slotFloatFramePool[size], bytecodeValueSlotFloatFrame{
		values: values,
		kinds:  kinds,
		valid:  valid,
	})
}

func (vm *bytecodeVM) releaseActiveValueSlotFloatFrame() {
	if vm == nil {
		return
	}
	values, kinds, valid := vm.slotFloatValues, vm.slotFloatKinds, vm.slotFloatValid
	vm.slotFloatOwner = nil
	vm.slotFloatValues = nil
	vm.slotFloatKinds = nil
	vm.slotFloatValid = nil
	vm.releaseValueSlotFloatFrame(values, kinds, valid)
}

func (vm *bytecodeVM) detachActiveValueSlotFloatFrame() ([]float64, []runtime.FloatType, []bool) {
	if vm == nil || len(vm.slotFloatValues) == 0 {
		if vm != nil {
			vm.slotFloatOwner = nil
		}
		return nil, nil, nil
	}
	values, kinds, valid := vm.slotFloatValues, vm.slotFloatKinds, vm.slotFloatValid
	vm.slotFloatOwner = nil
	vm.slotFloatValues = nil
	vm.slotFloatKinds = nil
	vm.slotFloatValid = nil
	return values, kinds, valid
}

func (vm *bytecodeVM) restoreValueSlotFloatFrame(slots []runtime.Value, values []float64, kinds []runtime.FloatType, valid []bool) {
	if vm == nil {
		return
	}
	if len(values) == 0 && len(kinds) == 0 && len(valid) == 0 {
		if vm.slotFloatOwner == nil && len(vm.slotFloatValues) == 0 && len(vm.slotFloatKinds) == 0 && len(vm.slotFloatValid) == 0 {
			return
		}
		vm.releaseActiveValueSlotFloatFrame()
		return
	}
	if vm.activeValueSlotFloatFrameMatchesSlots(slots) &&
		len(vm.slotFloatValues) == len(values) &&
		len(vm.slotFloatKinds) == len(kinds) &&
		len(vm.slotFloatValid) == len(valid) &&
		(len(values) == 0 ||
			(&vm.slotFloatValues[0] == &values[0] &&
				&vm.slotFloatKinds[0] == &kinds[0] &&
				&vm.slotFloatValid[0] == &valid[0])) {
		return
	}
	vm.releaseActiveValueSlotFloatFrame()
	if len(values) == 0 || len(values) != len(kinds) || len(values) != len(valid) || len(values) != len(slots) {
		return
	}
	vm.slotFloatOwner = bytecodeSlotFrameOwner(slots)
	vm.slotFloatValues = values
	vm.slotFloatKinds = kinds
	vm.slotFloatValid = valid
}

func (vm *bytecodeVM) prepareValueSlotFloatFrame(_ *bytecodeProgram) {
	if vm == nil {
		return
	}
	if len(vm.slots) == 0 {
		vm.releaseActiveValueSlotFloatFrame()
		return
	}
	if !vm.activeValueSlotFloatFrameMatchesSlots(vm.slots) {
		vm.releaseActiveValueSlotFloatFrame()
	}
}

func (vm *bytecodeVM) ensureActiveValueSlotFloatFrame() bool {
	if vm == nil || len(vm.slots) == 0 {
		return false
	}
	if vm.activeValueSlotFloatFrameMatchesSlots(vm.slots) {
		return true
	}
	values, kinds, valid := vm.acquireValueSlotFloatFrame(len(vm.slots))
	for idx, value := range vm.slots {
		if raw, kind, ok := bytecodeDirectRawFloatValue(value); ok {
			values[idx] = raw
			kinds[idx] = kind
			valid[idx] = true
		}
	}
	vm.restoreValueSlotFloatFrame(vm.slots, values, kinds, valid)
	return vm.activeValueSlotFloatFrameMatchesSlots(vm.slots)
}

func (vm *bytecodeVM) activeValueSlotFloatRaw(slot int) (float64, runtime.FloatType, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) || slot >= len(vm.slotFloatValues) || slot >= len(vm.slotFloatKinds) || slot >= len(vm.slotFloatValid) {
		return 0, runtime.FloatF64, false
	}
	if !vm.activeValueSlotFloatFrameMatchesSlots(vm.slots) || !vm.slotFloatValid[slot] {
		return 0, runtime.FloatF64, false
	}
	return vm.slotFloatValues[slot], vm.slotFloatKinds[slot], true
}

func (vm *bytecodeVM) setActiveValueSlotFloatRaw(slot int, raw float64, kind runtime.FloatType) bool {
	if vm == nil || slot < 0 || slot >= len(vm.slots) || slot >= len(vm.slotFloatValues) || slot >= len(vm.slotFloatKinds) || slot >= len(vm.slotFloatValid) {
		return false
	}
	if !vm.activeValueSlotFloatFrameMatchesSlots(vm.slots) {
		return false
	}
	vm.slotFloatValues[slot] = normalizeFloat(kind, raw)
	vm.slotFloatKinds[slot] = kind
	vm.slotFloatValid[slot] = true
	return true
}

func (vm *bytecodeVM) clearActiveValueSlotFloat(slot int) {
	if vm == nil || slot < 0 || slot >= len(vm.slotFloatValid) || !vm.activeValueSlotFloatFrameMatchesSlots(vm.slots) {
		return
	}
	vm.slotFloatValid[slot] = false
}

func (vm *bytecodeVM) storeActiveValueSlotFloatRaw(slot int, raw float64, kind runtime.FloatType) bool {
	if !vm.ensureActiveValueSlotFloatFrame() {
		return false
	}
	if !vm.setActiveValueSlotFloatRaw(slot, raw, kind) {
		return false
	}
	vm.slots[slot] = nil
	return true
}

func (vm *bytecodeVM) saveReusedSlot0ValueSlotFloat(frame *bytecodeSelfFastMinimalCallFrame) {
	if vm == nil || frame == nil {
		return
	}
	if raw, kind, ok := vm.activeValueSlotFloatRaw(0); ok {
		frame.slot0FloatRaw = raw
		frame.slot0FloatKind = kind
		frame.slot0FloatValid = true
		return
	}
	frame.slot0FloatRaw = 0
	frame.slot0FloatKind = runtime.FloatF64
	frame.slot0FloatValid = false
}

func (vm *bytecodeVM) restoreReusedSlot0ValueSlotFloat(raw float64, kind runtime.FloatType, valid bool) {
	if !valid {
		vm.clearActiveValueSlotFloat(0)
		return
	}
	vm.setActiveValueSlotFloatRaw(0, raw, kind)
}
