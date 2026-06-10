package interpreter

import "able/interpreter-go/pkg/runtime"

const (
	// Reduced recursive bytecode benchmarks repeatedly churn very small slot
	// frames. A larger batch keeps those hot pools populated across the common
	// recursion depth instead of rebuilding them every run.
	bytecodeSlotFrameBatchSize        = 32
	bytecodeSlotFrameBatchMaxSlots    = 16
	bytecodeSlotFrameSmallHotMaxSlots = 4
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
		if slotCount <= bytecodeSlotFrameSmallHotMaxSlots {
			return vm.acquireSmallHotSlotFrame(slotCount)
		}
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

func (vm *bytecodeVM) acquireSlotFrame2() []runtime.Value {
	if vm == nil {
		return make([]runtime.Value, 2)
	}
	return vm.acquireSmallHotSlotFrame(2)
}

func (vm *bytecodeVM) acquireSmallHotSlotFrame(slotCount int) []runtime.Value {
	if vm == nil || slotCount <= 0 || slotCount > bytecodeSlotFrameSmallHotMaxSlots {
		return make([]runtime.Value, slotCount)
	}
	pool := vm.slotFrameSmallHotPools[slotCount]
	if len(pool) > 0 {
		idx := len(pool) - 1
		slots := pool[idx]
		vm.slotFrameSmallHotPools[slotCount] = pool[:idx]
		return slots
	}
	backing := make([]runtime.Value, slotCount*bytecodeSlotFrameBatchSize)
	first := backing[:slotCount:slotCount]
	for idx := bytecodeSlotFrameBatchSize - 1; idx >= 1; idx-- {
		start := idx * slotCount
		slots := backing[start : start+slotCount : start+slotCount]
		vm.slotFrameSmallHotPools[slotCount] = append(vm.slotFrameSmallHotPools[slotCount], slots)
	}
	return first
}

func (vm *bytecodeVM) releaseSlotFrame2(slots []runtime.Value) {
	if vm == nil {
		return
	}
	vm.releaseSlotFrameRawCells(slots)
	slots[0] = nil
	slots[1] = nil
	vm.slotFrameSmallHotPools[2] = append(vm.slotFrameSmallHotPools[2], slots)
}

func (vm *bytecodeVM) releaseSlotFrame4(slots []runtime.Value) {
	if vm == nil {
		return
	}
	vm.releaseSlotFrameRawCells(slots)
	slots[0] = nil
	slots[1] = nil
	slots[2] = nil
	slots[3] = nil
	vm.slotFrameSmallHotPools[4] = append(vm.slotFrameSmallHotPools[4], slots)
}

func (vm *bytecodeVM) releaseSlotFrame(slots []runtime.Value) {
	if vm == nil || len(slots) == 0 {
		return
	}
	size := len(slots)
	vm.releaseSlotFrameRawCells(slots)
	switch size {
	case 1:
		slots[0] = nil
	case 2:
		slots[0] = nil
		slots[1] = nil
	case 3:
		slots[0] = nil
		slots[1] = nil
		slots[2] = nil
	case 4:
		slots[0] = nil
		slots[1] = nil
		slots[2] = nil
		slots[3] = nil
	default:
		clear(slots)
	}
	if size <= bytecodeSlotFrameSmallHotMaxSlots {
		vm.slotFrameSmallHotPools[size] = append(vm.slotFrameSmallHotPools[size], slots)
		return
	}
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

func (vm *bytecodeVM) releaseSlotFrameRawCells(slots []runtime.Value) {
	if vm == nil {
		return
	}
	for idx, value := range slots {
		switch raw := value.(type) {
		case *bytecodeRawI64SlotCell:
			if vm.rawI64CellOwnedByStack(raw) || bytecodeRawI64CellSeen(slots[:idx], raw) {
				continue
			}
			vm.releaseRawI64SlotCell(raw)
		case *bytecodeRawIntegerSlotCell:
			if vm.rawIntegerCellOwnedByStack(raw) || bytecodeRawIntegerCellSeen(slots[:idx], raw) {
				continue
			}
			vm.releaseRawIntegerSlotCell(raw)
		}
	}
}

func (vm *bytecodeVM) rawI64CellOwnedByStack(cell *bytecodeRawI64SlotCell) bool {
	if vm == nil || cell == nil {
		return false
	}
	for _, stackCell := range vm.stackI64Cells {
		if stackCell == cell {
			return true
		}
	}
	return false
}

func bytecodeRawI64CellSeen(values []runtime.Value, cell *bytecodeRawI64SlotCell) bool {
	if cell == nil {
		return false
	}
	for _, value := range values {
		if value == cell {
			return true
		}
	}
	return false
}

func (vm *bytecodeVM) rawIntegerCellOwnedByStack(cell *bytecodeRawIntegerSlotCell) bool {
	if vm == nil || cell == nil {
		return false
	}
	for _, stackCell := range vm.stackIntegerCells {
		if stackCell == cell {
			return true
		}
	}
	return false
}

func bytecodeRawIntegerCellSeen(values []runtime.Value, cell *bytecodeRawIntegerSlotCell) bool {
	if cell == nil {
		return false
	}
	for _, value := range values {
		if value == cell {
			return true
		}
	}
	return false
}
