package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotFloatAddMul(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode float slot update missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if vm.stackDepth() < 3 {
		return fmt.Errorf("bytecode stack underflow")
	}
	baseIdx := vm.stackDepth() - 3
	base := vm.stackValue(baseIdx)
	mulLeft := vm.stackValue(baseIdx + 1)
	mulRight := vm.stackValue(baseIdx + 2)
	if resultVal, resultKind, ok := bytecodeDirectFloatAddMulRawValue(base, mulLeft, mulRight); ok {
		vm.truncateStack(baseIdx)
		return vm.finishStoreSlotFloatRawResult(instr, resultVal, resultKind)
	}
	result, err := vm.storeSlotFloatAddMulResult(instr, base, mulLeft, mulRight)
	if err != nil {
		if vm.interp != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
		}
		if instr.node != nil && vm.interp != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.truncateStack(baseIdx)
	return vm.finishStoreSlotFloatResult(instr, result, nil)
}

func (vm *bytecodeVM) storeSlotFloatAddMulResult(instr *bytecodeInstruction, base runtime.Value, mulLeft runtime.Value, mulRight runtime.Value) (runtime.Value, error) {
	base = bytecodeSlotReadValue(base)
	mulLeft = bytecodeMaterializeRawFloatValue(mulLeft)
	mulRight = bytecodeMaterializeRawFloatValue(mulRight)
	product, err := applyBinaryOperator(vm.interp, "*", mulLeft, mulRight)
	if err != nil {
		return nil, err
	}
	return applyBinaryOperator(vm.interp, instr.operator, base, product)
}

func bytecodeDirectFloatAddMul(base runtime.Value, mulLeft runtime.Value, mulRight runtime.Value) (runtime.Value, bool) {
	result, ok := bytecodeDirectFloatAddMulValue(base, mulLeft, mulRight)
	if !ok {
		return nil, false
	}
	return result, true
}

func bytecodeDirectFloatAddMulValue(base runtime.Value, mulLeft runtime.Value, mulRight runtime.Value) (runtime.Value, bool) {
	resultVal, resultKind, ok := bytecodeDirectFloatAddMulRawValue(base, mulLeft, mulRight)
	if !ok {
		return nil, false
	}
	return bytecodeRawFloatSlotValue(resultVal, resultKind), true
}

func bytecodeDirectFloatAddMulRawValue(base runtime.Value, mulLeft runtime.Value, mulRight runtime.Value) (float64, runtime.FloatType, bool) {
	baseVal, baseKind, ok := bytecodeDirectFloatValue(base)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	leftVal, leftKind, ok := bytecodeDirectFloatValue(mulLeft)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	rightVal, rightKind, ok := bytecodeDirectFloatValue(mulRight)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	return bytecodeDirectFloatAddMulOperandsRawValue(baseVal, baseKind, leftVal, leftKind, rightVal, rightKind)
}

func bytecodeDirectFloatAddMulOperandsRawValue(baseVal float64, baseKind runtime.FloatType, leftVal float64, leftKind runtime.FloatType, rightVal float64, rightKind runtime.FloatType) (float64, runtime.FloatType, bool) {
	productKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		productKind = runtime.FloatF64
	}
	product := normalizeFloat(productKind, leftVal*rightVal)
	resultKind := runtime.FloatF32
	if baseKind == runtime.FloatF64 || productKind == runtime.FloatF64 {
		resultKind = runtime.FloatF64
	}
	return normalizeFloat(resultKind, baseVal+product), resultKind, true
}

func bytecodeDirectFloatAddMulRaw(base runtime.Value, leftVal float64, leftKind runtime.FloatType, rightVal float64, rightKind runtime.FloatType) (runtime.Value, bool) {
	baseVal, baseKind, ok := bytecodeDirectFloatValue(base)
	if !ok {
		return nil, false
	}
	resultVal, resultKind, ok := bytecodeDirectFloatAddMulOperandsRawValue(baseVal, baseKind, leftVal, leftKind, rightVal, rightKind)
	if !ok {
		return nil, false
	}
	return bytecodeRawFloatSlotValue(resultVal, resultKind), true
}

func bytecodeStackSnapshotValue(value runtime.Value) runtime.Value {
	switch v := value.(type) {
	case bytecodeRawF32SlotValue, bytecodeRawF64SlotValue:
		return value
	case *bytecodeRawIntegerSlotCell:
		if v == nil {
			return runtime.NilValue{}
		}
		return bytecodeRawIntegerResultValue(v.TypeSuffix, v.Raw)
	case *bytecodeRawI64SlotCell:
		if v == nil {
			return runtime.NilValue{}
		}
		return bytecodeRawI64ResultValue(v.Val)
	case bytecodeRawI32SlotValue,
		*bytecodeRawI32StackCell,
		bytecodeRawU8ResultValue,
		bytecodeRawU16ResultValue,
		bytecodeRawU32ResultValue,
		bytecodeRawU64ResultValue,
		bytecodeRawUsizeResultValue,
		bytecodeRawI64ResultValue,
		bytecodeRawIntegerValue:
		return value
	case *runtime.IntegerValue:
		return bytecodeStackSnapshotIntegerPointerValue(v)
	case *runtime.FloatValue:
		if v == nil {
			return runtime.NilValue{}
		}
		return runtime.FloatValue{Val: v.Val, TypeSuffix: v.TypeSuffix}
	default:
		return value
	}
}

func bytecodeStackSnapshotIntegerPointerValue(value *runtime.IntegerValue) runtime.Value {
	if value == nil {
		return value
	}
	if value.IsSmallRef() {
		raw := value.Int64FastRef()
		if boxed, ok := bytecodeBoxedIntegerValue(value.TypeSuffix, raw); ok {
			return boxed
		}
		return runtime.NewSmallInt(raw, value.TypeSuffix)
	}
	return runtime.NewBigIntValue(runtime.CloneBigInt(value.BigInt()), value.TypeSuffix)
}

func bytecodeSlotReadValue(value runtime.Value) runtime.Value {
	return bytecodeMaterializeRawFloatValue(bytecodeStackSnapshotValue(value))
}

func (vm *bytecodeVM) storeOwnedFloatSlot(target int, value runtime.FloatValue) runtime.Value {
	return vm.storeOwnedFloatSlotRaw(target, value.Val, value.TypeSuffix)
}

func (vm *bytecodeVM) storeRawFloatSlotRaw(target int, val float64, kind runtime.FloatType) runtime.Value {
	val = normalizeFloat(kind, val)
	if target >= 0 && target < len(vm.slots) {
		bytecodeSetNormalizedRawFloatValue(&vm.slots[target], val, kind)
		return vm.slots[target]
	}
	return bytecodeRawFloatSlotValue(val, kind)
}

func (vm *bytecodeVM) storeOwnedI32SlotRaw(target int, raw int32) runtime.Value {
	if raw >= bytecodeRawI32SlotCacheMin && raw <= bytecodeRawI32SlotCacheMax {
		value := bytecodeRawI32SlotCachedValue(raw)
		if target >= 0 && target < len(vm.slots) {
			vm.clearActiveValueSlotFloat(target)
			vm.slots[target] = value
		}
		return value
	}
	if target < 0 || target >= len(vm.slots) {
		return runtime.NewSmallInt(int64(raw), runtime.IntegerI32)
	}
	key := &vm.slots[target]
	if vm.ownedI32Slots != nil {
		if cell := vm.ownedI32Slots[key]; cell != nil {
			cell.ResetSmall(int64(raw), runtime.IntegerI32)
			vm.clearActiveValueSlotFloat(target)
			vm.slots[target] = cell
			return cell
		}
	} else {
		vm.ownedI32Slots = make(map[*runtime.Value]*runtime.IntegerValue, 4)
	}
	cell := &runtime.IntegerValue{}
	cell.ResetSmall(int64(raw), runtime.IntegerI32)
	vm.ownedI32Slots[key] = cell
	vm.clearActiveValueSlotFloat(target)
	vm.slots[target] = cell
	return cell
}

func (vm *bytecodeVM) storeOwnedFloatSlotRaw(target int, val float64, kind runtime.FloatType) runtime.Value {
	if target < 0 || target >= len(vm.slots) {
		return runtime.FloatValue{Val: val, TypeSuffix: kind}
	}
	vm.clearActiveValueSlotFloat(target)
	key := &vm.slots[target]
	if vm.ownedFloatSlots != nil {
		if cell := vm.ownedFloatSlots[key]; cell != nil {
			cell.Val = val
			cell.TypeSuffix = kind
			vm.slots[target] = cell
			return cell
		}
	} else {
		vm.ownedFloatSlots = make(map[*runtime.Value]*runtime.FloatValue, 4)
	}
	cell := &runtime.FloatValue{Val: val, TypeSuffix: kind}
	vm.ownedFloatSlots[key] = cell
	vm.slots[target] = cell
	return cell
}

func (vm *bytecodeVM) storeFloatSlotValue(target int, value runtime.Value) runtime.Value {
	value = bytecodeStackResultValue(value)
	if _, _, ok := bytecodeDirectRawFloatValue(value); ok {
		if target >= 0 && target < len(vm.slots) {
			vm.clearActiveValueSlotFloat(target)
			vm.slots[target] = value
		}
	} else if fv, ok := value.(runtime.FloatValue); ok {
		value = vm.storeOwnedFloatSlot(target, fv)
	} else if fv, ok := value.(*runtime.FloatValue); ok && fv != nil {
		value = vm.storeOwnedFloatSlot(target, *fv)
	} else if target >= 0 && target < len(vm.slots) {
		vm.clearActiveValueSlotFloat(target)
		vm.slots[target] = value
	}
	vm.clearActiveValueSlotI32(target)
	if vm.hasI32RegisterFrame() {
		vm.setI32RegisterValue(target, value)
	}
	if target == 0 {
		vm.setSelfFastSlot0I32Value(value)
	}
	return value
}
