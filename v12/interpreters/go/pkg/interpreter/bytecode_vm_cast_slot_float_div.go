package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execBinaryCastSlotFloatConstDiv(instr *bytecodeInstruction) (runtime.Value, bool, error) {
	if instr == nil {
		return nil, true, fmt.Errorf("bytecode cast-slot-float-const div missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, true, fmt.Errorf("bytecode slot out of range")
	}
	targetKind, ok := bytecodeFloatCastTargetKind(instr.typeExpr)
	if !ok {
		return nil, true, fmt.Errorf("bytecode cast-slot-float-const div missing float target type")
	}
	right, ok := bytecodeFloatImmediateValue(instr.value)
	if !ok {
		return nil, true, fmt.Errorf("bytecode cast-slot-float-const div missing float immediate")
	}
	if fast, ok, err := vm.bytecodeCastSlotFloatConstDivFast(instr.target, targetKind, right); ok {
		return fast, true, err
	}
	rawLeft := vm.slotRuntimeValue(instr.target)
	casted, err := vm.interp.castValueToType(ast.Ty(string(targetKind)), rawLeft)
	if err != nil {
		return nil, true, err
	}
	result, err := applyBinaryOperator(vm.interp, "/", casted, right)
	return result, true, err
}

func (vm *bytecodeVM) bytecodeCastSlotFloatConstDivFast(slot int, targetKind runtime.FloatType, right runtime.FloatValue) (runtime.Value, bool, error) {
	resultVal, resultKind, ok, err := vm.bytecodeCastSlotFloatConstDivRawFast(slot, targetKind, right)
	if !ok || err != nil {
		return nil, ok, err
	}
	return bytecodeRawFloatSlotValue(resultVal, resultKind), true, nil
}

func (vm *bytecodeVM) bytecodeCastSlotFloatConstDivRawFast(slot int, targetKind runtime.FloatType, right runtime.FloatValue) (float64, runtime.FloatType, bool, error) {
	leftVal, ok := vm.bytecodeCastSlotFloatValue(slot, targetKind)
	if !ok {
		return 0, runtime.FloatF64, false, nil
	}
	resultVal, resultKind, handled, err := bytecodeDirectFloatDivisionRawFast(leftVal, targetKind, right.Val, right.TypeSuffix)
	if err != nil {
		return 0, resultKind, true, err
	}
	if !handled {
		return 0, runtime.FloatF64, false, nil
	}
	return resultVal, resultKind, true, nil
}

func (vm *bytecodeVM) bytecodeCastSlotFloatValue(slot int, targetKind runtime.FloatType) (float64, bool) {
	if vm.hasI32RegisterFrame() {
		if raw, ok := vm.i32RegisterRaw(slot); ok {
			return normalizeFloat(targetKind, float64(raw)), true
		}
	}
	if raw, ok := vm.activeValueSlotI32Raw(slot); ok {
		return normalizeFloat(targetKind, float64(raw)), true
	}
	if raw, _, ok := vm.activeValueSlotFloatRaw(slot); ok {
		return normalizeFloat(targetKind, raw), true
	}
	return bytecodeCastValueToFloatRaw(vm.slots[slot], targetKind)
}

func bytecodeCastValueToFloatRaw(rawValue runtime.Value, targetKind runtime.FloatType) (float64, bool) {
	switch val := rawValue.(type) {
	case runtime.IntegerValue:
		return normalizeFloat(targetKind, integerValueToFloat64Fast(val)), true
	case *runtime.IntegerValue:
		if val != nil {
			return normalizeFloat(targetKind, integerRefToFloat64Fast(val)), true
		}
	case *bytecodeRawI64SlotCell:
		if val != nil {
			return normalizeFloat(targetKind, float64(val.Val)), true
		}
	case bytecodeRawF32SlotValue:
		return normalizeFloat(targetKind, float64(val)), true
	case bytecodeRawF64SlotValue:
		return normalizeFloat(targetKind, float64(val)), true
	case runtime.FloatValue:
		return normalizeFloat(targetKind, val.Val), true
	case *runtime.FloatValue:
		if val != nil {
			return normalizeFloat(targetKind, val.Val), true
		}
	case bytecodeRawI32SlotValue:
		return normalizeFloat(targetKind, float64(val)), true
	}
	return 0, false
}

func bytecodeFloatImmediateValue(val runtime.Value) (runtime.FloatValue, bool) {
	if raw, kind, ok := bytecodeDirectRawFloatValue(val); ok {
		return runtime.FloatValue{Val: raw, TypeSuffix: kind}, true
	}
	switch fv := val.(type) {
	case runtime.FloatValue:
		return fv, true
	case *runtime.FloatValue:
		if fv != nil {
			return *fv, true
		}
	}
	return runtime.FloatValue{}, false
}
