package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeRawI64SlotCell struct {
	Val int64
}

func (vm *bytecodeVM) mustUseImmutableRawIntegerCarriers() bool {
	return vm != nil && vm.interp != nil && !vm.bytecodeSingleThread()
}

func (v *bytecodeRawI64SlotCell) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (*bytecodeRawI64SlotCell) Kind() runtime.Kind {
	return runtime.KindInteger
}

func bytecodeRawI64Value(value runtime.Value) (int64, bool) {
	switch raw := value.(type) {
	case *bytecodeRawI64SlotCell:
		if raw != nil {
			return raw.Val, true
		}
	}
	return 0, false
}

func (vm *bytecodeVM) storeRawI64Slot(target int, value int64) runtime.Value {
	if target < 0 || target >= len(vm.slots) {
		return &bytecodeRawI64SlotCell{Val: value}
	}
	if vm.mustUseImmutableRawIntegerCarriers() {
		stored := bytecodeRawI64ResultValue(value)
		vm.slots[target] = stored
		return stored
	}
	if cell, ok := vm.slots[target].(*bytecodeRawI64SlotCell); ok && cell != nil {
		cell.Val = value
		vm.slots[target] = cell
		return cell
	}
	cell := vm.acquireRawI64SlotCell(value)
	vm.slots[target] = cell
	return cell
}

func (vm *bytecodeVM) acquireRawI64SlotCell(value int64) *bytecodeRawI64SlotCell {
	if vm == nil || len(vm.rawI64SlotCellPool) == 0 {
		return &bytecodeRawI64SlotCell{Val: value}
	}
	idx := len(vm.rawI64SlotCellPool) - 1
	cell := vm.rawI64SlotCellPool[idx]
	vm.rawI64SlotCellPool[idx] = nil
	vm.rawI64SlotCellPool = vm.rawI64SlotCellPool[:idx]
	if cell == nil {
		return &bytecodeRawI64SlotCell{Val: value}
	}
	cell.Val = value
	return cell
}

func (vm *bytecodeVM) releaseRawI64SlotCell(cell *bytecodeRawI64SlotCell) {
	if vm == nil || cell == nil {
		return
	}
	cell.Val = 0
	vm.rawI64SlotCellPool = append(vm.rawI64SlotCellPool, cell)
}

func bytecodeBoxRawI64Value(value *bytecodeRawI64SlotCell) runtime.Value {
	if value == nil {
		return runtime.NilValue{}
	}
	return boxedOrSmallIntegerValue(runtime.IntegerI64, value.Val)
}

func (vm *bytecodeVM) stackRawI64Value(index int, value int64) runtime.Value {
	if vm == nil || index < 0 {
		return &bytecodeRawI64SlotCell{Val: value}
	}
	if vm.mustUseImmutableRawIntegerCarriers() {
		return bytecodeRawI64ResultValue(value)
	}
	if index >= len(vm.stackI64Cells) {
		extra := make([]*bytecodeRawI64SlotCell, index-len(vm.stackI64Cells)+1)
		vm.stackI64Cells = append(vm.stackI64Cells, extra...)
	}
	cell := vm.stackI64Cells[index]
	if cell == nil {
		cell = &bytecodeRawI64SlotCell{}
		vm.stackI64Cells[index] = cell
	}
	cell.Val = value
	return cell
}

func (vm *bytecodeVM) appendRawI64Stack(value int64) {
	if vm == nil {
		return
	}
	vm.appendStackValue(vm.stackRawI64Value(vm.stackDepth(), value))
}

func (vm *bytecodeVM) replaceTop2RawI64Unchecked(value int64) {
	idx := vm.stackDepth() - 2
	vm.setStackValue(idx, vm.stackRawI64Value(idx, value))
	vm.truncateStack(idx + 1)
}
