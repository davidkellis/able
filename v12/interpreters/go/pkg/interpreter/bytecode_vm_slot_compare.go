package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execJumpIfIntCompareSlotFalse(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot compare jump missing instruction")
	}
	leftSlot, rightSlot := instr.argCount, instr.loopBreak
	if leftSlot < 0 || leftSlot >= len(vm.slots) || rightSlot < 0 || rightSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode slot compare slot out of range")
	}
	if leftRaw, ok := vm.slotDirectSmallI32ValueValidated(leftSlot); ok {
		if rightRaw, ok := vm.slotDirectSmallI32ValueValidated(rightSlot); ok {
			if cond, ok := bytecodeCompareInt64(instr.operator, leftRaw, rightRaw); ok {
				if !cond {
					vm.ip = instr.target
					return nil
				}
				vm.ip++
				return nil
			}
		}
	}
	left, right := vm.slotRuntimeValue(leftSlot), vm.slotRuntimeValue(rightSlot)
	cond, err := vm.compareBytecodeCondition(instr.operator, left, right)
	if err != nil {
		return err
	}
	if !cond {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) compareBytecodeCondition(op string, left runtime.Value, right runtime.Value) (bool, error) {
	if cmp, ok := bytecodeDirectIntegerCompare(op, left, right); ok {
		return cmp.Val, nil
	}
	if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
		if rightInt, ok := bytecodeDirectIntegerValue(right); ok {
			return integerComparisonResult(op, leftInt, rightInt), nil
		}
	}
	if fast, handled := execBinaryDirectIntegerComparisonFast(op, left, right); handled {
		if b, ok := fast.(runtime.BoolValue); ok {
			return b.Val, nil
		}
		return vm.interp.isTruthy(fast), nil
	}
	if fast, handled := bytecodeDirectFloatCompareFast(op, left, right); handled {
		return fast.Val, nil
	}
	if isBytecodeBinaryFastPathCandidate(op) {
		if fast, handled, err := ApplyBinaryOperatorFast(op, left, right); handled {
			if err != nil {
				return false, err
			}
			return vm.interp.isTruthy(fast), nil
		}
	}
	result, err := applyBinaryOperator(vm.interp, op, left, right)
	if err != nil {
		return false, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	return vm.interp.isTruthy(result), nil
}
