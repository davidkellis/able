package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotIntMulConstModConst(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode int mul/mod slot update missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if instr.discardResult {
		if err, handled := vm.execStoreSlotIntMulConstModConstDiscardSteadyStateFast(instr); handled {
			return err
		}
	}
	mulImmediate, hasMulImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasMulImmediate {
		return fmt.Errorf("bytecode int mul/mod slot update missing multiply immediate")
	}
	modImmediate, hasModImmediate := instr.intImmediate2, instr.hasIntImmediate2
	if !hasModImmediate {
		modImmediate, hasModImmediate = bytecodeImmediateIntegerValue(instr.value)
		if !hasModImmediate {
			return fmt.Errorf("bytecode int mul/mod slot update missing modulo immediate")
		}
	}
	if instr.discardResult {
		if err, handled := vm.execStoreSlotIntMulConstModConstDiscardFast(instr, mulImmediate, modImmediate); handled {
			return err
		}
	}
	base := vm.slots[instr.target]
	result, err := vm.storeSlotIntMulConstModConstResult(base, mulImmediate, modImmediate, instr.discardResult)
	return vm.finishStoreSlotBinaryIntSlotConstFastResult(instr, result, err)
}

func (vm *bytecodeVM) execStoreSlotIntMulConstModConstDiscardSteadyStateFast(instr *bytecodeInstruction) (error, bool) {
	if instr == nil || !instr.discardResult || !instr.hasIntImmediate || !instr.hasIntImmediate2 || !instr.hasIntRaw || !instr.hasIntRaw2 || instr.intImmediate.TypeSuffix != runtime.IntegerI64 || instr.intImmediate2.TypeSuffix != runtime.IntegerI64 || instr.intImmediateRaw <= 0 || instr.intImmediate2Raw <= 0 {
		return nil, false
	}
	sourceCell, ok := vm.slots[instr.target].(*bytecodeRawI64SlotCell)
	if !ok || sourceCell == nil || sourceCell.Val < 0 {
		return nil, false
	}
	rem, handled, err := bytecodePositiveIntMulConstModFast(sourceCell.Val, instr.intImmediateRaw, instr.intImmediate2Raw)
	if !handled || err != nil {
		return err, handled
	}
	sourceCell.Val = rem
	if instr.target == 0 {
		vm.clearSelfFastSlot0I32()
	}
	vm.ip++
	return nil, true
}

func (vm *bytecodeVM) execStoreSlotIntMulConstModConstDiscardFast(instr *bytecodeInstruction, mulImmediate runtime.IntegerValue, modImmediate runtime.IntegerValue) (error, bool) {
	if instr == nil || mulImmediate.TypeSuffix != runtime.IntegerI64 || modImmediate.TypeSuffix != runtime.IntegerI64 {
		return nil, false
	}
	baseRaw, ok := bytecodeDirectPositiveSmallIntOfKind(vm.slots[instr.target], runtime.IntegerI64)
	if !ok {
		return nil, false
	}
	mulRaw, modRaw, ok := bytecodeStoreSlotIntMulConstModConstImmediateRaws(instr, mulImmediate, modImmediate)
	if !ok {
		return nil, false
	}
	rem, handled, err := bytecodePositiveIntMulConstModFast(baseRaw, mulRaw, modRaw)
	if !handled || err != nil {
		return err, handled
	}
	vm.storeRawI64Slot(instr.target, rem)
	if instr.target == 0 {
		vm.clearSelfFastSlot0I32()
	}
	vm.ip++
	return nil, true
}

func bytecodeStoreSlotIntMulConstModConstImmediateRaws(instr *bytecodeInstruction, mulImmediate runtime.IntegerValue, modImmediate runtime.IntegerValue) (int64, int64, bool) {
	if instr != nil &&
		instr.hasIntImmediate &&
		instr.hasIntImmediate2 &&
		instr.hasIntRaw &&
		instr.hasIntRaw2 &&
		instr.intImmediate.TypeSuffix == runtime.IntegerI64 &&
		instr.intImmediate2.TypeSuffix == runtime.IntegerI64 &&
		instr.intImmediateRaw > 0 &&
		instr.intImmediate2Raw > 0 {
		return instr.intImmediateRaw, instr.intImmediate2Raw, true
	}
	mulRef := &mulImmediate
	modRef := &modImmediate
	if !mulRef.IsSmallRef() || !modRef.IsSmallRef() {
		return 0, 0, false
	}
	mulRaw := mulRef.Int64FastRef()
	modRaw := modRef.Int64FastRef()
	if mulRaw <= 0 || modRaw <= 0 {
		return 0, 0, false
	}
	return mulRaw, modRaw, true
}

func (vm *bytecodeVM) storeSlotIntMulConstModConstResult(base runtime.Value, mulImmediate runtime.IntegerValue, modImmediate runtime.IntegerValue, discardResult bool) (runtime.Value, error) {
	if fast, handled, err := bytecodeIntMulConstModConstFast(base, mulImmediate, modImmediate, discardResult); handled {
		return fast, err
	}
	product, err := applyBinaryOperator(vm.interp, "*", base, mulImmediate)
	if err != nil {
		return nil, err
	}
	return evaluateDivMod(vm.interp, "%", product, modImmediate)
}

func bytecodeIntMulConstModConstFast(base runtime.Value, mulImmediate runtime.IntegerValue, modImmediate runtime.IntegerValue, discardResult bool) (runtime.Value, bool, error) {
	mulRef := &mulImmediate
	modRef := &modImmediate
	if !mulRef.IsSmallRef() || !modRef.IsSmallRef() || mulImmediate.TypeSuffix != modImmediate.TypeSuffix {
		return nil, false, nil
	}
	switch mulImmediate.TypeSuffix {
	case runtime.IntegerI32, runtime.IntegerI64:
	default:
		return nil, false, nil
	}
	mulRaw := mulRef.Int64FastRef()
	modRaw := modRef.Int64FastRef()
	if modRaw == 0 {
		return nil, true, newDivisionByZeroError()
	}
	if mulRaw > 0 && modRaw > 0 {
		if baseRaw, ok := bytecodeDirectPositiveSmallIntOfKind(base, mulImmediate.TypeSuffix); ok {
			if rem, ok, err := bytecodePositiveIntMulConstModFast(baseRaw, mulRaw, modRaw); ok {
				if err != nil {
					return nil, true, err
				}
				if mulImmediate.TypeSuffix == runtime.IntegerI32 {
					return bytecodeRawI32ResultValue(rem), true, nil
				}
				return boxedOrSmallIntegerValue(mulImmediate.TypeSuffix, rem), true, nil
			}
		}
	}
	baseInt, ok := bytecodeDirectIntegerValue(base)
	if !ok || baseInt.TypeSuffix != mulImmediate.TypeSuffix {
		return nil, false, nil
	}
	baseRef := &baseInt
	if !baseRef.IsSmallRef() {
		return nil, false, nil
	}
	product, overflow := mulInt64Overflow(baseRef.Int64FastRef(), mulRaw)
	if overflow {
		return nil, true, newOverflowError("integer overflow")
	}
	_, rem := euclideanDivModInt64(product, modRaw)
	if err := ensureFitsInt64Type(mulImmediate.TypeSuffix, rem); err != nil {
		return nil, true, err
	}
	if mulImmediate.TypeSuffix == runtime.IntegerI32 {
		return bytecodeRawI32ResultValue(rem), true, nil
	}
	return boxedOrSmallIntegerValue(mulImmediate.TypeSuffix, rem), true, nil
}

func bytecodeDirectPositiveSmallIntOfKind(base runtime.Value, kind runtime.IntegerType) (int64, bool) {
	gotKind, raw, ok := bytecodeRawIntegerValueInfo(base)
	if !ok || gotKind != kind || raw < 0 {
		return 0, false
	}
	return raw, true
}

func bytecodePositiveIntMulConstModFast(baseRaw int64, mulRaw int64, modRaw int64) (int64, bool, error) {
	if baseRaw < 0 || mulRaw <= 0 || modRaw <= 0 {
		return 0, false, nil
	}
	if baseRaw != 0 && mulRaw > math.MaxInt64/baseRaw {
		return 0, true, newOverflowError("integer overflow")
	}
	product := baseRaw * mulRaw
	return product % modRaw, true, nil
}

func bytecodeDirectPositiveSmallI64ImmediateValue(value runtime.Value) (int64, bool) {
	switch iv := value.(type) {
	case runtime.IntegerValue:
		ref := &iv
		if iv.TypeSuffix != runtime.IntegerI64 || !ref.IsSmallRef() {
			return 0, false
		}
		raw := ref.Int64FastRef()
		if raw <= 0 {
			return 0, false
		}
		return raw, true
	case *runtime.IntegerValue:
		if iv == nil || iv.TypeSuffix != runtime.IntegerI64 || !iv.IsSmallRef() {
			return 0, false
		}
		raw := iv.Int64FastRef()
		if raw <= 0 {
			return 0, false
		}
		return raw, true
	default:
		return 0, false
	}
}
