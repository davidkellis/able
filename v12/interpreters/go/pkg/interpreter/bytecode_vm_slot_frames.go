package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) acquireSlotFrame(slotCount int) []runtime.Value {
	if slotCount <= 0 {
		return nil
	}
	if vm != nil {
		if vm.slotFrameHotSize == slotCount && len(vm.slotFrameHotPool) > 0 {
			idx := len(vm.slotFrameHotPool) - 1
			slots := vm.slotFrameHotPool[idx]
			vm.slotFrameHotPool = vm.slotFrameHotPool[:idx]
			return slots
		}
		if vm.slotFramePool != nil {
			if frames := vm.slotFramePool[slotCount]; len(frames) > 0 {
				idx := len(frames) - 1
				slots := frames[idx]
				vm.slotFramePool[slotCount] = frames[:idx]
				return slots
			}
		}
	}
	return make([]runtime.Value, slotCount)
}

func (vm *bytecodeVM) releaseSlotFrame(slots []runtime.Value) {
	if vm == nil || len(slots) == 0 {
		return
	}
	clear(slots)
	size := len(slots)
	if vm.slotFrameHotSize == 0 || vm.slotFrameHotSize == size {
		vm.slotFrameHotSize = size
		vm.slotFrameHotPool = append(vm.slotFrameHotPool, slots)
		return
	}
	if vm.slotFramePool == nil {
		vm.slotFramePool = make(map[int][][]runtime.Value, 1)
	}
	vm.slotFramePool[size] = append(vm.slotFramePool[size], slots)
}
