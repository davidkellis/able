package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) restoreValueSlotSidecarFrames(slots []runtime.Value, i32Values []int32, i32Valid []bool, floatValues []float64, floatKinds []runtime.FloatType, floatValid []bool) {
	if vm == nil {
		return
	}
	if len(i32Values) != 0 ||
		len(i32Valid) != 0 ||
		vm.slotI32Owner != nil ||
		len(vm.slotI32Values) != 0 ||
		len(vm.slotI32Valid) != 0 {
		vm.restoreValueSlotI32Frame(slots, i32Values, i32Valid)
	}
	if len(floatValues) != 0 ||
		len(floatKinds) != 0 ||
		len(floatValid) != 0 ||
		vm.slotFloatOwner != nil ||
		len(vm.slotFloatValues) != 0 ||
		len(vm.slotFloatKinds) != 0 ||
		len(vm.slotFloatValid) != 0 {
		vm.restoreValueSlotFloatFrame(slots, floatValues, floatKinds, floatValid)
	}
}
