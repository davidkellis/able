package interpreter

import (
	"math"
	"math/big"

	"able/interpreter-go/pkg/runtime"
)

type bytecodeRawIntegerValue struct {
	Raw        int64
	TypeSuffix runtime.IntegerType
}

func (bytecodeRawIntegerValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawU8ResultValue uint8

func (bytecodeRawU8ResultValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawU16ResultValue uint16

func (bytecodeRawU16ResultValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawU32ResultValue uint32

func (bytecodeRawU32ResultValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawU64ResultValue uint64

func (bytecodeRawU64ResultValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawUsizeResultValue uint64

func (bytecodeRawUsizeResultValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawI64ResultValue int64

func (bytecodeRawI64ResultValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawIntegerSlotCell struct {
	Raw        int64
	TypeSuffix runtime.IntegerType
}

func (*bytecodeRawIntegerSlotCell) Kind() runtime.Kind {
	return runtime.KindInteger
}

type bytecodeRawIntegerReturnScratch struct {
	Raw        int64
	TypeSuffix runtime.IntegerType
}

func (v bytecodeRawIntegerValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawU8ResultValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawU16ResultValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawU32ResultValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawU64ResultValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawUsizeResultValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawI64ResultValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v *bytecodeRawIntegerSlotCell) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v *bytecodeRawIntegerReturnScratch) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (*bytecodeRawIntegerReturnScratch) Kind() runtime.Kind {
	return runtime.KindInteger
}

func (vm *bytecodeVM) acquireRawIntegerSlotCell(kind runtime.IntegerType, raw int64) *bytecodeRawIntegerSlotCell {
	if vm == nil || len(vm.rawIntegerSlotCellPool) == 0 {
		return &bytecodeRawIntegerSlotCell{Raw: raw, TypeSuffix: kind}
	}
	idx := len(vm.rawIntegerSlotCellPool) - 1
	cell := vm.rawIntegerSlotCellPool[idx]
	vm.rawIntegerSlotCellPool[idx] = nil
	vm.rawIntegerSlotCellPool = vm.rawIntegerSlotCellPool[:idx]
	if cell == nil {
		return &bytecodeRawIntegerSlotCell{Raw: raw, TypeSuffix: kind}
	}
	cell.Raw = raw
	cell.TypeSuffix = kind
	return cell
}

func (vm *bytecodeVM) releaseRawIntegerSlotCell(cell *bytecodeRawIntegerSlotCell) {
	if vm == nil || cell == nil {
		return
	}
	cell.Raw = 0
	cell.TypeSuffix = ""
	vm.rawIntegerSlotCellPool = append(vm.rawIntegerSlotCellPool, cell)
}

func bytecodeRawIntegerKindSupported(kind runtime.IntegerType) bool {
	switch kind {
	case runtime.IntegerI8,
		runtime.IntegerI16,
		runtime.IntegerI32,
		runtime.IntegerI64,
		runtime.IntegerU8,
		runtime.IntegerU16,
		runtime.IntegerU32,
		runtime.IntegerU64,
		runtime.IntegerIsize,
		runtime.IntegerUsize:
		return true
	default:
		return false
	}
}

func bytecodeRawIntegerStoreKindFits(kind runtime.IntegerType, raw int64) bool {
	switch kind {
	case runtime.IntegerI8:
		return raw >= math.MinInt8 && raw <= math.MaxInt8
	case runtime.IntegerI16:
		return raw >= math.MinInt16 && raw <= math.MaxInt16
	case runtime.IntegerI32:
		return raw >= math.MinInt32 && raw <= math.MaxInt32
	case runtime.IntegerI64, runtime.IntegerIsize:
		return true
	case runtime.IntegerU8:
		return raw >= 0 && raw <= math.MaxUint8
	case runtime.IntegerU16:
		return raw >= 0 && raw <= math.MaxUint16
	case runtime.IntegerU32:
		return raw >= 0 && raw <= math.MaxUint32
	case runtime.IntegerU64, runtime.IntegerUsize:
		return raw >= 0
	default:
		return false
	}
}

func bytecodeRawIntegerValueInfo(value runtime.Value) (runtime.IntegerType, int64, bool) {
	switch raw := value.(type) {
	case bytecodeRawI32SlotValue:
		return runtime.IntegerI32, int64(raw), true
	case *bytecodeRawI32StackCell:
		if raw != nil {
			return runtime.IntegerI32, int64(raw.Val), true
		}
	case bytecodeRawU8ResultValue:
		return runtime.IntegerU8, int64(raw), true
	case bytecodeRawU16ResultValue:
		return runtime.IntegerU16, int64(raw), true
	case bytecodeRawU32ResultValue:
		return runtime.IntegerU32, int64(raw), true
	case bytecodeRawU64ResultValue:
		return runtime.IntegerU64, int64(raw), true
	case bytecodeRawUsizeResultValue:
		return runtime.IntegerUsize, int64(raw), true
	case bytecodeRawI64ResultValue:
		return runtime.IntegerI64, int64(raw), true
	case bytecodeRawIntegerValue:
		return raw.TypeSuffix, raw.Raw, true
	case *bytecodeRawIntegerSlotCell:
		if raw != nil {
			return raw.TypeSuffix, raw.Raw, true
		}
	case *bytecodeRawIntegerReturnScratch:
		if raw != nil {
			return raw.TypeSuffix, raw.Raw, true
		}
	case *bytecodeRawI64SlotCell:
		if raw != nil {
			return runtime.IntegerI64, raw.Val, true
		}
	case runtime.IntegerValue:
		if raw.IsSmall() {
			return raw.TypeSuffix, raw.Int64Fast(), true
		}
	case *runtime.IntegerValue:
		if raw != nil && raw.IsSmallRef() {
			return raw.TypeSuffix, raw.Int64FastRef(), true
		}
	}
	return "", 0, false
}

func bytecodeRawIntegerResultValue(kind runtime.IntegerType, raw int64) runtime.Value {
	switch kind {
	case runtime.IntegerI32:
		if raw >= math.MinInt32 && raw <= math.MaxInt32 {
			return bytecodeRawI32SlotCachedValue(int32(raw))
		}
	case runtime.IntegerU8:
		return bytecodeRawU8ResultValue(uint8(raw))
	case runtime.IntegerU16:
		return bytecodeRawU16ResultValue(uint16(raw))
	case runtime.IntegerU32:
		return bytecodeRawU32ResultValue(uint32(raw))
	case runtime.IntegerU64:
		return bytecodeRawU64ResultValue(uint64(raw))
	case runtime.IntegerUsize:
		return bytecodeRawUsizeResultValue(uint64(raw))
	case runtime.IntegerI64:
		return bytecodeRawI64ResultValue(raw)
	case runtime.IntegerI8,
		runtime.IntegerI16,
		runtime.IntegerIsize:
		return bytecodeRawIntegerValue{Raw: raw, TypeSuffix: kind}
	}
	return bytecodeBoxRawIntegerValue(kind, raw)
}

func bytecodeBoxRawIntegerValue(kind runtime.IntegerType, raw int64) runtime.Value {
	if (kind == runtime.IntegerU64 || kind == runtime.IntegerUsize) && raw < 0 {
		return runtime.NewBigIntValue(new(big.Int).SetUint64(uint64(raw)), kind)
	}
	return boxedOrSmallIntegerValue(kind, raw)
}

func bytecodeMaterializeRawIntegerValue(value runtime.Value) runtime.Value {
	switch raw := value.(type) {
	case bytecodeRawI32SlotValue:
		return bytecodeBoxRawI32Value(raw)
	case *bytecodeRawI32StackCell:
		if raw == nil {
			return runtime.NilValue{}
		}
		return bytecodeBoxedIntegerI32Value(int64(raw.Val))
	case bytecodeRawU8ResultValue:
		return bytecodeBoxRawIntegerValue(runtime.IntegerU8, int64(raw))
	case bytecodeRawU16ResultValue:
		return bytecodeBoxRawIntegerValue(runtime.IntegerU16, int64(raw))
	case bytecodeRawU32ResultValue:
		return bytecodeBoxRawIntegerValue(runtime.IntegerU32, int64(raw))
	case bytecodeRawU64ResultValue:
		return bytecodeBoxRawIntegerValue(runtime.IntegerU64, int64(raw))
	case bytecodeRawUsizeResultValue:
		return bytecodeBoxRawIntegerValue(runtime.IntegerUsize, int64(raw))
	case bytecodeRawI64ResultValue:
		return boxedOrSmallIntegerValue(runtime.IntegerI64, int64(raw))
	case bytecodeRawIntegerValue:
		return bytecodeBoxRawIntegerValue(raw.TypeSuffix, raw.Raw)
	case *bytecodeRawIntegerSlotCell:
		if raw == nil {
			return runtime.NilValue{}
		}
		return bytecodeBoxRawIntegerValue(raw.TypeSuffix, raw.Raw)
	case *bytecodeRawIntegerReturnScratch:
		if raw == nil {
			return runtime.NilValue{}
		}
		return bytecodeBoxRawIntegerValue(raw.TypeSuffix, raw.Raw)
	case *bytecodeRawI64SlotCell:
		return bytecodeBoxRawI64Value(raw)
	default:
		return value
	}
}

func bytecodeMaterializeRawValue(value runtime.Value) runtime.Value {
	return bytecodeMaterializeRawFloatValue(bytecodeMaterializeRawIntegerValue(value))
}

func bytecodeIsRawIntegerCarrier(value runtime.Value) bool {
	switch value.(type) {
	case bytecodeRawI32SlotValue,
		*bytecodeRawI32StackCell,
		bytecodeRawU8ResultValue,
		bytecodeRawU16ResultValue,
		bytecodeRawU32ResultValue,
		bytecodeRawU64ResultValue,
		bytecodeRawUsizeResultValue,
		bytecodeRawI64ResultValue,
		bytecodeRawIntegerValue,
		*bytecodeRawIntegerSlotCell,
		*bytecodeRawIntegerReturnScratch,
		*bytecodeRawI64SlotCell:
		return true
	default:
		return false
	}
}

func bytecodeCastSmallIntToIntegerKindRawResult(value int64, targetKind runtime.IntegerType, info integerInfo) (runtime.Value, bool) {
	if raw, ok := bytecodeCastSmallIntToIntegerKindRawBits(value, targetKind, info); ok {
		return bytecodeRawIntegerResultValue(targetKind, raw), true
	}
	if !bytecodeRawIntegerKindSupported(targetKind) || info.bits <= 0 {
		return nil, false
	}
	if !info.signed && info.bits >= 64 &&
		(targetKind == runtime.IntegerU64 || targetKind == runtime.IntegerUsize) {
		bits := uint64(value)
		if bits > math.MaxInt64 {
			return runtime.NewBigIntValue(new(big.Int).SetUint64(bits), targetKind), true
		}
	}
	return nil, false
}

func bytecodeCastSmallIntToIntegerKindRawBits(value int64, targetKind runtime.IntegerType, info integerInfo) (int64, bool) {
	if !bytecodeRawIntegerKindSupported(targetKind) || info.bits <= 0 {
		return 0, false
	}
	if info.signed {
		if info.bits >= 64 {
			return value, true
		}
		mask := (uint64(1) << uint(info.bits)) - 1
		bits := uint64(value) & mask
		signBit := uint64(1) << uint(info.bits-1)
		if bits&signBit != 0 {
			bits |= ^mask
		}
		return int64(bits), true
	}
	if info.bits < 64 {
		mask := (uint64(1) << uint(info.bits)) - 1
		return int64(uint64(value) & mask), true
	}
	if value >= 0 {
		return value, true
	}
	bits := uint64(value)
	if bits <= math.MaxInt64 {
		return int64(bits), true
	}
	return 0, false
}

func bytecodeCachedRawIntegerSlotValue(kind runtime.IntegerType, raw int64) (runtime.Value, bool) {
	if kind == runtime.IntegerI32 {
		return nil, false
	}
	return boxedSmallIntValue(kind, raw)
}

func (vm *bytecodeVM) storeRawIntegerSlot(target int, kind runtime.IntegerType, raw int64) runtime.Value {
	switch kind {
	case runtime.IntegerI32:
		return vm.storeOwnedI32SlotRaw(target, int32(raw))
	default:
		if boxed, ok := bytecodeCachedRawIntegerSlotValue(kind, raw); ok {
			if target >= 0 && target < len(vm.slots) {
				vm.slots[target] = boxed
			}
			return boxed
		}
		if kind == runtime.IntegerI64 {
			return vm.storeRawI64Slot(target, raw)
		}
		if !bytecodeRawIntegerKindSupported(kind) {
			value := bytecodeBoxRawIntegerValue(kind, raw)
			if target >= 0 && target < len(vm.slots) {
				vm.slots[target] = value
			}
			return value
		}
		if target < 0 || target >= len(vm.slots) {
			return bytecodeRawIntegerResultValue(kind, raw)
		}
		if vm.mustUseImmutableRawIntegerCarriers() {
			value := bytecodeRawIntegerResultValue(kind, raw)
			vm.slots[target] = value
			return value
		}
		if cell, ok := vm.slots[target].(*bytecodeRawIntegerSlotCell); ok && cell != nil && cell.TypeSuffix == kind {
			cell.Raw = raw
			vm.slots[target] = cell
			return cell
		}
		cell := vm.acquireRawIntegerSlotCell(kind, raw)
		vm.slots[target] = cell
		return cell
	}
}

func (vm *bytecodeVM) stackRawIntegerValue(index int, kind runtime.IntegerType, raw int64) runtime.Value {
	switch kind {
	case runtime.IntegerI32:
		if raw >= math.MinInt32 && raw <= math.MaxInt32 {
			return vm.stackRawI32Value(index, int32(raw))
		}
	case runtime.IntegerI64:
		return vm.stackRawI64Value(index, raw)
	}
	if !bytecodeRawIntegerKindSupported(kind) {
		return bytecodeBoxRawIntegerValue(kind, raw)
	}
	if vm == nil || index < 0 {
		return bytecodeRawIntegerResultValue(kind, raw)
	}
	if vm.mustUseImmutableRawIntegerCarriers() {
		return bytecodeRawIntegerResultValue(kind, raw)
	}
	if index >= len(vm.stackIntegerCells) {
		extra := make([]*bytecodeRawIntegerSlotCell, index-len(vm.stackIntegerCells)+1)
		vm.stackIntegerCells = append(vm.stackIntegerCells, extra...)
	}
	cell := vm.stackIntegerCells[index]
	if cell == nil {
		cell = &bytecodeRawIntegerSlotCell{}
		vm.stackIntegerCells[index] = cell
	}
	cell.Raw = raw
	cell.TypeSuffix = kind
	return cell
}

func (vm *bytecodeVM) appendRawIntegerStack(kind runtime.IntegerType, raw int64) runtime.Value {
	result := vm.stackRawIntegerValue(vm.stackDepth(), kind, raw)
	vm.appendStackValue(result)
	return result
}

func (vm *bytecodeVM) rawIntegerReturnValue(kind runtime.IntegerType, raw int64) runtime.Value {
	if vm == nil {
		return bytecodeRawIntegerResultValue(kind, raw)
	}
	if vm.mustUseImmutableRawIntegerCarriers() {
		return bytecodeRawIntegerResultValue(kind, raw)
	}
	vm.rawIntegerReturnScratch.Raw = raw
	vm.rawIntegerReturnScratch.TypeSuffix = kind
	return &vm.rawIntegerReturnScratch
}

func (vm *bytecodeVM) replaceTop2RawIntegerUnchecked(kind runtime.IntegerType, raw int64) {
	idx := vm.stackDepth() - 2
	vm.setStackValue(idx, vm.stackRawIntegerValue(idx, kind, raw))
	vm.truncateStack(idx + 1)
}

func (vm *bytecodeVM) tryStoreRawIntegerSlotValue(target int, value runtime.Value) (runtime.Value, bool) {
	var (
		kind runtime.IntegerType
		raw  int64
	)
	switch value := value.(type) {
	case bytecodeRawI32SlotValue:
		return vm.storeOwnedI32SlotRaw(target, int32(value)), true
	case *bytecodeRawI32StackCell:
		if value == nil {
			return nil, false
		}
		return vm.storeOwnedI32SlotRaw(target, value.Val), true
	case bytecodeRawI64ResultValue:
		return vm.storeRawI64Slot(target, int64(value)), true
	case *bytecodeRawI64SlotCell:
		if value == nil {
			return nil, false
		}
		return vm.storeRawI64Slot(target, value.Val), true
	case bytecodeRawU8ResultValue:
		kind, raw = runtime.IntegerU8, int64(value)
	case bytecodeRawU16ResultValue:
		kind, raw = runtime.IntegerU16, int64(value)
	case bytecodeRawU32ResultValue:
		kind, raw = runtime.IntegerU32, int64(value)
	case bytecodeRawU64ResultValue:
		kind, raw = runtime.IntegerU64, int64(value)
	case bytecodeRawUsizeResultValue:
		kind, raw = runtime.IntegerUsize, int64(value)
	case bytecodeRawIntegerValue:
		kind, raw = value.TypeSuffix, value.Raw
	case *bytecodeRawIntegerSlotCell:
		if value == nil {
			return nil, false
		}
		kind, raw = value.TypeSuffix, value.Raw
	case *bytecodeRawIntegerReturnScratch:
		if value == nil {
			return nil, false
		}
		kind, raw = value.TypeSuffix, value.Raw
	case runtime.IntegerValue:
		valueRef := &value
		if !valueRef.IsSmallRef() {
			return nil, false
		}
		kind, raw = value.TypeSuffix, valueRef.Int64FastRef()
	case *runtime.IntegerValue:
		if value == nil || !value.IsSmallRef() {
			return nil, false
		}
		kind, raw = value.TypeSuffix, value.Int64FastRef()
	default:
		return nil, false
	}
	if !bytecodeRawIntegerStoreKindFits(kind, raw) {
		return nil, false
	}
	return vm.storeRawIntegerSlot(target, kind, raw), true
}
