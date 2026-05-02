package interpreter

import (
	"math"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) clearSelfFastSlot0I32() {
	if vm == nil {
		return
	}
	vm.selfFastSlot0I32Valid = false
}

func (vm *bytecodeVM) setSelfFastSlot0I32Raw(raw int32) {
	if vm == nil {
		return
	}
	vm.selfFastSlot0I32Raw = raw
	vm.selfFastSlot0I32Valid = true
}

func (vm *bytecodeVM) setSelfFastSlot0I32Raw64(raw int64) bool {
	if vm == nil || raw < math.MinInt32 || raw > math.MaxInt32 {
		if vm != nil {
			vm.selfFastSlot0I32Valid = false
		}
		return false
	}
	vm.setSelfFastSlot0I32Raw(int32(raw))
	return true
}

func (vm *bytecodeVM) setSelfFastSlot0I32Value(value runtime.Value) bool {
	raw, ok := bytecodeRawI32Value(value)
	if !ok {
		vm.clearSelfFastSlot0I32()
		return false
	}
	vm.setSelfFastSlot0I32Raw(raw)
	return true
}

func (vm *bytecodeVM) saveSelfFastSlot0I32(frame *bytecodeSelfFastMinimalCallFrame) {
	if vm == nil || frame == nil {
		return
	}
	frame.slot0I32Raw = vm.selfFastSlot0I32Raw
	frame.slot0I32Valid = vm.selfFastSlot0I32Valid
}

func (vm *bytecodeVM) restoreSelfFastSlot0I32(frame *bytecodeSelfFastMinimalCallFrame) {
	if vm == nil || frame == nil {
		return
	}
	if frame.slot0I32Valid {
		vm.setSelfFastSlot0I32Raw(frame.slot0I32Raw)
	} else {
		vm.clearSelfFastSlot0I32()
	}
	frame.slot0I32Raw = 0
	frame.slot0I32Valid = false
}
