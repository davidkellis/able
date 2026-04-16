package interpreter

import "able/interpreter-go/pkg/runtime"

const (
	bytecodeSlotFrameBatchSize     = 8
	bytecodeSlotFrameBatchMaxSlots = 16
)

func (vm *bytecodeVM) spillHotSlotFrames() {
	if vm == nil || vm.slotFrameHotSize == 0 || len(vm.slotFrameHotPool) == 0 {
		return
	}
	if vm.slotFramePool == nil {
		vm.slotFramePool = make(map[int][][]runtime.Value, 1)
	}
	vm.slotFramePool[vm.slotFrameHotSize] = append(vm.slotFramePool[vm.slotFrameHotSize], vm.slotFrameHotPool...)
	vm.slotFrameHotPool = vm.slotFrameHotPool[:0]
}

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
		if slotCount <= bytecodeSlotFrameBatchMaxSlots {
			if vm.slotFrameHotSize != 0 && vm.slotFrameHotSize != slotCount {
				vm.spillHotSlotFrames()
			}
			backing := make([]runtime.Value, slotCount*bytecodeSlotFrameBatchSize)
			first := backing[:slotCount:slotCount]
			vm.slotFrameHotSize = slotCount
			for idx := bytecodeSlotFrameBatchSize - 1; idx >= 1; idx-- {
				start := idx * slotCount
				slots := backing[start : start+slotCount : start+slotCount]
				vm.slotFrameHotPool = append(vm.slotFrameHotPool, slots)
			}
			return first
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
