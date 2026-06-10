package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func bytecodeInstructionFloatImmediateRaw(instr *bytecodeInstruction) (float64, runtime.FloatType, bool) {
	if instr != nil && instr.hasFloatImmediate {
		return instr.floatImmediateRaw, instr.floatImmediateKind, true
	}
	if instr == nil {
		return 0, runtime.FloatF64, false
	}
	right, ok := bytecodeFloatImmediateValue(instr.value)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	return right.Val, right.TypeSuffix, true
}

func (vm *bytecodeVM) execBinaryFloatMulSlotConst(instr *bytecodeInstruction) (runtime.Value, bool, error) {
	if instr == nil {
		return nil, true, fmt.Errorf("bytecode float slot-const multiply missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, true, fmt.Errorf("bytecode slot out of range")
	}
	rightVal, rightKind, ok := bytecodeInstructionFloatImmediateRaw(instr)
	if !ok {
		return nil, true, fmt.Errorf("bytecode float slot-const multiply missing float immediate")
	}
	if resultVal, resultKind, handled := vm.binaryFloatMulSlotConstFastRaw(instr.target, rightVal, rightKind); handled {
		return bytecodeRawFloatSlotValue(resultVal, resultKind), true, nil
	}
	right := runtime.FloatValue{Val: rightVal, TypeSuffix: rightKind}
	left := vm.slotRuntimeValue(instr.target)
	result, err := applyBinaryOperator(vm.interp, "*", left, right)
	return result, true, err
}

func (vm *bytecodeVM) binaryFloatMulSlotConstFast(slot int, right runtime.FloatValue) (runtime.Value, bool) {
	resultVal, resultKind, handled := vm.binaryFloatMulSlotConstFastRaw(slot, right.Val, right.TypeSuffix)
	if !handled {
		return nil, false
	}
	return bytecodeRawFloatSlotValue(resultVal, resultKind), true
}

func (vm *bytecodeVM) binaryFloatMulSlotConstFastRaw(slot int, rightVal float64, rightKind runtime.FloatType) (float64, runtime.FloatType, bool) {
	leftVal, leftKind, ok := vm.slotDirectFloatValue(slot)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	return bytecodeDirectFloatArithmeticRawFast("*", leftVal, leftKind, rightVal, rightKind)
}
