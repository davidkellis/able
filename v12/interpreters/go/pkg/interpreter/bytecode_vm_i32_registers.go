package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeProgramUsesI32RegisterFrame(program *bytecodeProgram) bool {
	return program != nil &&
		program.frameLayout != nil &&
		program.frameLayout.i32RegisterFrame &&
		program.frameLayout.slotCount > 0
}

func (vm *bytecodeVM) hasI32RegisterFrame() bool {
	return vm != nil && len(vm.i32RegisterValid) > 0
}

func (vm *bytecodeVM) acquireI32RegisterFrame(slotCount int) ([]int32, []bool) {
	if slotCount <= 0 {
		return nil, nil
	}
	if vm != nil && vm.i32RegisterFramePool != nil {
		frames := vm.i32RegisterFramePool[slotCount]
		if len(frames) > 0 {
			idx := len(frames) - 1
			frame := frames[idx]
			vm.i32RegisterFramePool[slotCount] = frames[:idx]
			return frame.values, frame.valid
		}
	}
	return make([]int32, slotCount), make([]bool, slotCount)
}

func (vm *bytecodeVM) releaseI32RegisterFrame(values []int32, valid []bool) {
	if vm == nil || len(values) == 0 || len(values) != len(valid) {
		return
	}
	clear(valid)
	if vm.i32RegisterFramePool == nil {
		vm.i32RegisterFramePool = make(map[int][]bytecodeI32RegisterFrame, 2)
	}
	size := len(values)
	vm.i32RegisterFramePool[size] = append(vm.i32RegisterFramePool[size], bytecodeI32RegisterFrame{
		values: values,
		valid:  valid,
	})
}

func (vm *bytecodeVM) releaseActiveI32RegisterFrame() {
	if vm == nil || len(vm.i32Registers) == 0 {
		if vm != nil {
			vm.i32RegisterProgram = nil
		}
		return
	}
	values, valid := vm.i32Registers, vm.i32RegisterValid
	vm.i32RegisterProgram = nil
	vm.i32Registers = nil
	vm.i32RegisterValid = nil
	vm.releaseI32RegisterFrame(values, valid)
}

func (vm *bytecodeVM) detachActiveI32RegisterFrame() (*bytecodeProgram, []int32, []bool) {
	if vm == nil || len(vm.i32Registers) == 0 {
		if vm != nil {
			vm.i32RegisterProgram = nil
		}
		return nil, nil, nil
	}
	program := vm.i32RegisterProgram
	values, valid := vm.i32Registers, vm.i32RegisterValid
	vm.i32RegisterProgram = nil
	vm.i32Registers = nil
	vm.i32RegisterValid = nil
	return program, values, valid
}

func (vm *bytecodeVM) restoreI32RegisterFrame(program *bytecodeProgram, values []int32, valid []bool) {
	if vm == nil {
		return
	}
	vm.releaseActiveI32RegisterFrame()
	if len(values) == 0 || len(values) != len(valid) {
		vm.i32RegisterProgram = nil
		return
	}
	vm.i32RegisterProgram = program
	vm.i32Registers = values
	vm.i32RegisterValid = valid
}

func (vm *bytecodeVM) activateI32RegisterFrame(program *bytecodeProgram) {
	if vm == nil {
		return
	}
	if vm.i32RegisterProgram == program && bytecodeProgramUsesI32RegisterFrame(program) {
		return
	}
	vm.releaseActiveI32RegisterFrame()
	if !bytecodeProgramUsesI32RegisterFrame(program) {
		return
	}
	layout := program.frameLayout
	values, valid := vm.acquireI32RegisterFrame(layout.slotCount)
	for slot, kind := range layout.slotKinds {
		if kind != bytecodeCellKindI32 || slot >= len(vm.slots) || slot >= len(values) {
			continue
		}
		if raw, ok := bytecodeRawI32Value(vm.slots[slot]); ok {
			values[slot] = raw
			valid[slot] = true
		}
	}
	vm.i32Registers = values
	vm.i32RegisterValid = valid
	vm.i32RegisterProgram = program
}

func (vm *bytecodeVM) i32RegisterRaw(slot int) (int32, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.i32Registers) || slot >= len(vm.i32RegisterValid) {
		return 0, false
	}
	if !vm.i32RegisterValid[slot] {
		return 0, false
	}
	return vm.i32Registers[slot], true
}

func (vm *bytecodeVM) setI32RegisterRaw(slot int, raw int32) bool {
	if vm == nil || slot < 0 || slot >= len(vm.i32Registers) || slot >= len(vm.i32RegisterValid) {
		return false
	}
	vm.i32Registers[slot] = raw
	vm.i32RegisterValid[slot] = true
	return true
}

func (vm *bytecodeVM) setI32RegisterValue(slot int, value runtime.Value) bool {
	raw, ok := bytecodeRawI32Value(value)
	if !ok {
		vm.clearI32RegisterSlot(slot)
		return false
	}
	return vm.setI32RegisterRaw(slot, raw)
}

func (vm *bytecodeVM) clearI32RegisterSlot(slot int) {
	if vm == nil || slot < 0 || slot >= len(vm.i32RegisterValid) {
		return
	}
	vm.i32RegisterValid[slot] = false
}

func (vm *bytecodeVM) slotRuntimeValue(slot int) runtime.Value {
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return bytecodeBoxedIntegerI32Value(int64(raw))
	}
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return nil
	}
	return bytecodeSlotReadValue(vm.slots[slot])
}
