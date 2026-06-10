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
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		return bytecodeBoxedIntegerI32Value(int64(raw))
	}
	return nil
}

func (vm *bytecodeVM) copySlotRawIntegerCellInto(dst []runtime.Value, dstSlot int, callerSlot int) bool {
	if vm == nil || dstSlot < 0 || dstSlot >= len(dst) || callerSlot < 0 || callerSlot >= len(vm.slots) {
		return false
	}
	switch raw := vm.slots[callerSlot].(type) {
	case *bytecodeRawI64SlotCell:
		if raw == nil {
			return false
		}
		if vm.mustUseImmutableRawIntegerCarriers() {
			dst[dstSlot] = bytecodeRawI64ResultValue(raw.Val)
		} else {
			dst[dstSlot] = vm.acquireRawI64SlotCell(raw.Val)
		}
		return true
	case *bytecodeRawIntegerSlotCell:
		if raw == nil {
			return false
		}
		if vm.mustUseImmutableRawIntegerCarriers() {
			dst[dstSlot] = bytecodeRawIntegerResultValue(raw.TypeSuffix, raw.Raw)
		} else {
			dst[dstSlot] = vm.acquireRawIntegerSlotCell(raw.TypeSuffix, raw.Raw)
		}
		return true
	default:
		return false
	}
}

func (vm *bytecodeVM) slotStackValue(slot int) runtime.Value {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return nil
	}
	return vm.slotStackValueChecked(slot)
}

func (vm *bytecodeVM) slotStackValueChecked(slot int) runtime.Value {
	if value := vm.slots[slot]; value != nil {
		switch raw := value.(type) {
		case *runtime.IntegerValue:
			return bytecodeStackSnapshotIntegerPointerValue(raw)
		case runtime.IntegerValue,
			*runtime.StructInstanceValue:
			return raw
		}
		return bytecodeStackSnapshotValue(value)
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return bytecodeRawI32SlotCachedValue(raw)
	}
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		return bytecodeRawI32SlotCachedValue(raw)
	}
	return nil
}

func (vm *bytecodeVM) appendSlotStackValueChecked(slot int) {
	if value := vm.slots[slot]; value != nil {
		switch raw := value.(type) {
		case *bytecodeRawI64SlotCell:
			if raw != nil {
				vm.appendRawI64Stack(raw.Val)
				return
			}
		case *bytecodeRawIntegerSlotCell:
			if raw != nil {
				vm.appendRawIntegerStack(raw.TypeSuffix, raw.Raw)
				return
			}
		case *runtime.IntegerValue:
			vm.appendStackValue(bytecodeStackSnapshotIntegerPointerValue(raw))
			return
		case runtime.IntegerValue,
			*runtime.StructInstanceValue:
			vm.appendStackValue(raw)
			return
		}
		vm.appendStackValue(bytecodeStackSnapshotValue(value))
		return
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		vm.appendStackValue(bytecodeRawI32SlotCachedValue(raw))
		return
	}
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		vm.appendStackValue(bytecodeRawI32SlotCachedValue(raw))
		return
	}
	vm.appendStackValue(nil)
}

func (vm *bytecodeVM) slotStoredValue(slot int) runtime.Value {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return nil
	}
	if value := vm.slots[slot]; value != nil {
		return value
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return bytecodeRawI32SlotCachedValue(raw)
	}
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		return bytecodeRawI32SlotCachedValue(raw)
	}
	return nil
}

func (vm *bytecodeVM) slotDirectFloatValue(slot int) (float64, runtime.FloatType, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, runtime.FloatF64, false
	}
	return vm.slotDirectFloatValueValidated(slot)
}

func (vm *bytecodeVM) slotDirectFloatValueValidated(slot int) (float64, runtime.FloatType, bool) {
	if value := vm.slots[slot]; value != nil {
		switch raw := value.(type) {
		case bytecodeRawF32SlotValue:
			return float64(raw), runtime.FloatF32, true
		case bytecodeRawF64SlotValue:
			return float64(raw), runtime.FloatF64, true
		case runtime.FloatValue:
			return raw.Val, raw.TypeSuffix, true
		case *runtime.FloatValue:
			if raw != nil {
				return raw.Val, raw.TypeSuffix, true
			}
		}
		return 0, runtime.FloatF64, false
	}
	if raw, kind, ok := vm.activeValueSlotFloatRaw(slot); ok {
		return raw, kind, true
	}
	return 0, runtime.FloatF64, false
}

func (vm *bytecodeVM) slotDirectF64Value(slot int) (float64, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false
	}
	if value := vm.slots[slot]; value != nil {
		switch raw := value.(type) {
		case bytecodeRawF64SlotValue:
			return float64(raw), true
		case runtime.FloatValue:
			return raw.Val, raw.TypeSuffix == runtime.FloatF64
		case *runtime.FloatValue:
			if raw != nil {
				return raw.Val, raw.TypeSuffix == runtime.FloatF64
			}
		}
		return 0, false
	}
	raw, kind, ok := vm.activeValueSlotFloatRaw(slot)
	return raw, ok && kind == runtime.FloatF64
}

func (vm *bytecodeVM) slotDirectSmallI32Value(slot int) (int64, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false
	}
	return vm.slotDirectSmallI32ValueValidated(slot)
}

func (vm *bytecodeVM) slotDirectSmallI32ValueValidated(slot int) (int64, bool) {
	if value := vm.slots[slot]; value != nil {
		return bytecodeDirectSmallI32Value(value)
	}
	if uint(slot) < uint(len(vm.i32Registers)) &&
		uint(slot) < uint(len(vm.i32RegisterValid)) &&
		vm.i32RegisterValid[slot] {
		return int64(vm.i32Registers[slot]), true
	}
	if uint(slot) < uint(len(vm.slotI32Values)) &&
		uint(slot) < uint(len(vm.slotI32Valid)) &&
		vm.slotI32Valid[slot] &&
		vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		return int64(vm.slotI32Values[slot]), true
	}
	return 0, false
}

func (vm *bytecodeVM) slotI32Value(slot int) (int64, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false
	}
	if raw, ok := vm.i32RegisterRaw(slot); ok {
		return int64(raw), true
	}
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		return int64(raw), true
	}
	if raw, ok := bytecodeRawI32Value(vm.slots[slot]); ok {
		return int64(raw), true
	}
	return 0, false
}

func (vm *bytecodeVM) slotDirectSmallArrayIndex(slot int) (int, bool) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) || strconv.IntSize != 64 {
		return 0, false
	}
	return vm.slotDirectSmallArrayIndexValidated(slot)
}

func (vm *bytecodeVM) slotDirectSmallArrayIndexValidated(slot int) (int, bool) {
	if strconv.IntSize != 64 {
		return 0, false
	}
	if value := vm.slots[slot]; value != nil {
		return bytecodeDirectSmallArrayIndex(value)
	}
	if uint(slot) < uint(len(vm.i32Registers)) &&
		uint(slot) < uint(len(vm.i32RegisterValid)) &&
		vm.i32RegisterValid[slot] {
		return int(vm.i32Registers[slot]), true
	}
	if uint(slot) < uint(len(vm.slotI32Values)) &&
		uint(slot) < uint(len(vm.slotI32Valid)) &&
		vm.slotI32Valid[slot] &&
		vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		return int(vm.slotI32Values[slot]), true
	}
	return 0, false
}

func (vm *bytecodeVM) slotDirectArrayIndex(slot int) (int, bool, error) {
	if vm == nil || slot < 0 || slot >= len(vm.slots) {
		return 0, false, nil
	}
	return vm.slotDirectArrayIndexValidated(slot)
}

func (vm *bytecodeVM) slotDirectArrayIndexValidated(slot int) (int, bool, error) {
	if value := vm.slots[slot]; value != nil {
		return bytecodeDirectArrayIndex(value)
	}
	if uint(slot) < uint(len(vm.i32Registers)) &&
		uint(slot) < uint(len(vm.i32RegisterValid)) &&
		vm.i32RegisterValid[slot] {
		return int(vm.i32Registers[slot]), true, nil
	}
	if uint(slot) < uint(len(vm.slotI32Values)) &&
		uint(slot) < uint(len(vm.slotI32Valid)) &&
		vm.slotI32Valid[slot] &&
		vm.activeValueSlotI32FrameMatchesSlots(vm.slots) {
		return int(vm.slotI32Values[slot]), true, nil
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
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok && raw >= 0 {
		return int(raw), true
	}
	return 0, false
}
