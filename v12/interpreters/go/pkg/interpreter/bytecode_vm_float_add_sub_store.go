package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotFloatAddSub(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode float add-sub store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if vm.stackDepth() < 3 {
		return fmt.Errorf("bytecode float add-sub store missing operands")
	}

	baseIdx := vm.stackDepth() - 3
	base := vm.stackValue(baseIdx)
	subLeft := vm.stackValue(baseIdx + 1)
	subRight := vm.stackValue(baseIdx + 2)
	if resultVal, resultKind, ok := bytecodeDirectFloatAddSubRawValue(base, subLeft, subRight); ok {
		vm.truncateStack(baseIdx)
		return vm.finishStoreSlotFloatRawResult(instr, resultVal, resultKind)
	}

	result, err := vm.storeSlotFloatAddSubResult(base, subLeft, subRight)
	vm.truncateStack(baseIdx)
	return vm.finishStoreSlotFloatResult(instr, result, err)
}

func (vm *bytecodeVM) storeSlotFloatAddSubResult(base runtime.Value, subLeft runtime.Value, subRight runtime.Value) (runtime.Value, error) {
	base = bytecodeSlotReadValue(base)
	subLeft = bytecodeMaterializeRawFloatValue(subLeft)
	subRight = bytecodeMaterializeRawFloatValue(subRight)
	diff, err := applyBinaryOperator(vm.interp, "-", subLeft, subRight)
	if err != nil {
		return nil, err
	}
	return applyBinaryOperator(vm.interp, "+", diff, base)
}

func bytecodeDirectFloatAddSub(base runtime.Value, subLeft runtime.Value, subRight runtime.Value) (runtime.Value, bool) {
	resultVal, resultKind, ok := bytecodeDirectFloatAddSubRawValue(base, subLeft, subRight)
	if !ok {
		return nil, false
	}
	return bytecodeRawFloatSlotValue(resultVal, resultKind), true
}

func bytecodeDirectFloatAddSubRawValue(base runtime.Value, subLeft runtime.Value, subRight runtime.Value) (float64, runtime.FloatType, bool) {
	baseVal, baseKind, ok := bytecodeDirectFloatValue(base)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	subLeftVal, subLeftKind, ok := bytecodeDirectFloatValue(subLeft)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	subRightVal, subRightKind, ok := bytecodeDirectFloatValue(subRight)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	return bytecodeDirectFloatAddSubOperandsRawValue(baseVal, baseKind, subLeftVal, subLeftKind, subRightVal, subRightKind)
}

func bytecodeDirectFloatAddSubOperandsRawValue(baseVal float64, baseKind runtime.FloatType, subLeftVal float64, subLeftKind runtime.FloatType, subRightVal float64, subRightKind runtime.FloatType) (float64, runtime.FloatType, bool) {
	diffVal, diffKind, ok := bytecodeDirectFloatArithmeticRawFast("-", subLeftVal, subLeftKind, subRightVal, subRightKind)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	return bytecodeDirectFloatArithmeticRawFast("+", diffVal, diffKind, baseVal, baseKind)
}
