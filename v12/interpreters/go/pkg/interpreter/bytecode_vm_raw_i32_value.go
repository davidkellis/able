package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeRawI32SlotValue int32

const (
	bytecodeRawI32SlotCacheMin = -1024
	bytecodeRawI32SlotCacheMax = 262143
)

var bytecodeRawI32SlotCache = func() [bytecodeRawI32SlotCacheMax - bytecodeRawI32SlotCacheMin + 1]runtime.Value {
	var cache [bytecodeRawI32SlotCacheMax - bytecodeRawI32SlotCacheMin + 1]runtime.Value
	for idx := range cache {
		cache[idx] = bytecodeRawI32SlotValue(int32(idx + bytecodeRawI32SlotCacheMin))
	}
	return cache
}()

type bytecodeRawI32StackCell struct {
	Val int32
}

func (v bytecodeRawI32SlotValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

func (*bytecodeRawI32StackCell) Kind() runtime.Kind {
	return runtime.KindInteger
}

func (v bytecodeRawI32SlotValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v *bytecodeRawI32StackCell) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func bytecodeRawI32SlotCachedValue(value int32) runtime.Value {
	if value >= bytecodeRawI32SlotCacheMin && value <= bytecodeRawI32SlotCacheMax {
		return bytecodeRawI32SlotCache[int(value-bytecodeRawI32SlotCacheMin)]
	}
	return bytecodeRawI32SlotValue(value)
}

func bytecodeBoxRawI32Value(value bytecodeRawI32SlotValue) runtime.Value {
	return bytecodeBoxedIntegerI32Value(int64(value))
}

func bytecodeRawI32ResultValue(value int64) runtime.Value {
	return bytecodeRawI32SlotCachedValue(int32(value))
}

func (vm *bytecodeVM) stackRawI32Value(index int, value int32) runtime.Value {
	if vm == nil || index < 0 {
		return bytecodeRawI32SlotCachedValue(value)
	}
	if vm.mustUseImmutableRawIntegerCarriers() {
		return bytecodeRawI32SlotCachedValue(value)
	}
	if value >= bytecodeRawI32SlotCacheMin && value <= bytecodeRawI32SlotCacheMax {
		return bytecodeRawI32SlotCachedValue(value)
	}
	if index >= len(vm.stackI32Cells) {
		extra := make([]*bytecodeRawI32StackCell, index-len(vm.stackI32Cells)+1)
		vm.stackI32Cells = append(vm.stackI32Cells, extra...)
	}
	cell := vm.stackI32Cells[index]
	if cell == nil {
		cell = &bytecodeRawI32StackCell{}
		vm.stackI32Cells[index] = cell
	}
	cell.Val = value
	return cell
}

func (vm *bytecodeVM) replaceTop2RawI32Unchecked(value int32) {
	idx := vm.stackDepth() - 2
	vm.setStackValue(idx, vm.stackRawI32Value(idx, value))
	vm.truncateStack(idx + 1)
}
