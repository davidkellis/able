package interpreter

import (
	"strconv"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) slotMaterializedValue(slot int) runtime.Value {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return nil
	}
	if value := vm.slots[slot]; value != nil {
		return bytecodeSlotReadValue(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return bytecodeBoxedIntegerI32Value(int64(raw))
	}
	return nil
}

func (vm *bytecodeVM) slotDirectSmallI32Value(slot int) (int64, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false
	}
	if value := vm.slots[slot]; value != nil {
		return bytecodeDirectSmallI32Value(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return int64(raw), true
	}
	return 0, false
}

func (vm *bytecodeVM) slotDirectSmallArrayIndex(slot int) (int, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) || strconv.IntSize != 64 {
		return 0, false
	}
	if value := vm.slots[slot]; value != nil {
		return bytecodeDirectSmallArrayIndex(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return int(raw), true
	}
	return 0, false
}

func (vm *bytecodeVM) slotDirectArrayIndex(slot int) (int, bool, error) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false, nil
	}
	if value := vm.slots[slot]; value != nil {
		return bytecodeDirectArrayIndex(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return int(raw), true, nil
	}
	return 0, false, nil
}

func (vm *bytecodeVM) slotArraySlotIndexSmall(slot int) (int, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false
	}
	if value := vm.slots[slot]; value != nil {
		return arraySlotIndexSmall(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok && raw >= 0 {
		return int(raw), true
	}
	return 0, false
}
