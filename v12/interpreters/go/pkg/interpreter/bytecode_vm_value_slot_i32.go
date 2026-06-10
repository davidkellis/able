package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeSlotFrameOwner(slots []runtime.Value) *runtime.Value {
	if len(slots) == 0 {
		return nil
	}
	return &slots[0]
}

func (vm *bytecodeVM) activeValueSlotI32FrameMatchesSlots(slots []runtime.Value) bool {
	if vm == nil {
		return false
	}
	if len(slots) == 0 {
		return len(vm.slotI32Values) == 0 && len(vm.slotI32Valid) == 0 && vm.slotI32Owner == nil
	}
	return len(vm.slotI32Values) == len(slots) &&
		len(vm.slotI32Valid) == len(slots) &&
		vm.slotI32Owner == bytecodeSlotFrameOwner(slots)
}

func (vm *bytecodeVM) acquireValueSlotI32Frame(slotCount int) ([]int32, []bool) {
	if slotCount <= 0 {
		return nil, nil
	}
	if vm != nil && vm.slotI32FramePool != nil {
		frames := vm.slotI32FramePool[slotCount]
		if len(frames) > 0 {
			idx := len(frames) - 1
			frame := frames[idx]
			vm.slotI32FramePool[slotCount] = frames[:idx]
			return frame.values, frame.valid
		}
	}
	return make([]int32, slotCount), make([]bool, slotCount)
}

func (vm *bytecodeVM) releaseValueSlotI32Frame(values []int32, valid []bool) {
	if vm == nil || len(values) == 0 || len(values) != len(valid) {
		return
	}
	clear(valid)
	if vm.slotI32FramePool == nil {
		vm.slotI32FramePool = make(map[int][]bytecodeValueSlotI32Frame, 2)
	}
	size := len(values)
	vm.slotI32FramePool[size] = append(vm.slotI32FramePool[size], bytecodeValueSlotI32Frame{
		values: values,
		valid:  valid,
	})
}

func (vm *bytecodeVM) releaseActiveValueSlotI32Frame() {
	if vm == nil {
		return
	}
	values, valid := vm.slotI32Values, vm.slotI32Valid
	vm.slotI32Owner = nil
	vm.slotI32Values = nil
	vm.slotI32Valid = nil
	vm.releaseValueSlotI32Frame(values, valid)
}

func (vm *bytecodeVM) detachActiveValueSlotI32Frame() ([]int32, []bool) {
	if vm == nil || len(vm.slotI32Values) == 0 {
		if vm != nil {
			vm.slotI32Owner = nil
		}
		return nil, nil
	}
	values, valid := vm.slotI32Values, vm.slotI32Valid
	vm.slotI32Owner = nil
	vm.slotI32Values = nil
	vm.slotI32Valid = nil
	return values, valid
}

func (vm *bytecodeVM) restoreValueSlotI32Frame(slots []runtime.Value, values []int32, valid []bool) {
	if vm == nil {
		return
	}
	if len(values) == 0 && len(valid) == 0 {
		if vm.slotI32Owner == nil && len(vm.slotI32Values) == 0 && len(vm.slotI32Valid) == 0 {
			return
		}
		vm.releaseActiveValueSlotI32Frame()
		return
	}
	if vm.activeValueSlotI32FrameMatchesSlots(slots) &&
		len(vm.slotI32Values) == len(values) &&
		len(vm.slotI32Valid) == len(valid) &&
		(len(values) == 0 || (&vm.slotI32Values[0] == &values[0] && &vm.slotI32Valid[0] == &valid[0])) {
		return
	}
	vm.releaseActiveValueSlotI32Frame()
	if len(values) == 0 || len(values) != len(valid) || len(values) != len(slots) {
		return
	}
	vm.slotI32Owner = bytecodeSlotFrameOwner(slots)
	vm.slotI32Values = values
	vm.slotI32Valid = valid
}

func (vm *bytecodeVM) prepareValueSlotI32Frame(program *bytecodeProgram) {
	if vm == nil {
		return
	}
	if bytecodeProgramUsesI32RegisterFrame(program) || len(vm.slots) == 0 {
		vm.releaseActiveValueSlotI32Frame()
		return
	}
	if !vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		vm.releaseActiveValueSlotI32Frame()
	}
}

func (vm *bytecodeVM) ensureActiveValueSlotI32Frame() bool {
	if vm == nil || len(vm.slots) == 0 || bytecodeProgramUsesI32RegisterFrame(vm.currentProgram) {
		return false
	}
	if vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		return true
	}
	values, valid := vm.acquireValueSlotI32Frame(len(vm.slots))
	for idx, value := range vm.slots {
		if raw, ok := bytecodeRawI32Value(value); ok {
			values[idx] = raw
			valid[idx] = true
		}
	}
	vm.restoreValueSlotI32Frame(vm.slots, values, valid)
	return vm.activeValueSlotI32FrameMatchesSlots(vm.slots)
}

func (vm *bytecodeVM) activeValueSlotI32Raw(slot int) (int32, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) || slot >= len(vm.slotI32Values) || slot >= len(vm.slotI32Valid) {
		return 0, false
	}
	if !vm.activeValueSlotI32FrameMatchesSlots(vm.slots) || !vm.slotI32Valid[slot] {
		return 0, false
	}
	return vm.slotI32Values[slot], true
}

func (vm *bytecodeVM) setActiveValueSlotI32Raw(slot int, raw int32) bool {
	if vm == nil || slot < 0 || slot >= len(vm.slots) || slot >= len(vm.slotI32Values) || slot >= len(vm.slotI32Valid) {
		return false
	}
	if !vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		return false
	}
	vm.slotI32Values[slot] = raw
	vm.slotI32Valid[slot] = true
	return true
}

func (vm *bytecodeVM) clearActiveValueSlotI32(slot int) {
	if vm == nil || slot < 0 || slot >= len(vm.slotI32Valid) || !vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		return
	}
	vm.slotI32Valid[slot] = false
}

func (vm *bytecodeVM) storeActiveValueSlotI32Raw(slot int, raw int32) bool {
	if !vm.ensureActiveValueSlotI32Frame() {
		return false
	}
	if !vm.setActiveValueSlotI32Raw(slot, raw) {
		return false
	}
	vm.slots[slot] = nil
	return true
}

func (vm *bytecodeVM) restoreReusedSlot0ValueSlotI32(raw int32, valid bool) {
	if !valid {
		vm.clearActiveValueSlotI32(0)
		return
	}
	vm.setActiveValueSlotI32Raw(0, raw)
}
