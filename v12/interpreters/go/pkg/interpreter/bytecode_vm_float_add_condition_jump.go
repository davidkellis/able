package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execJumpIfFloatAddCompareConstFalse(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode float add compare jump missing instruction")
	}
	if program == nil {
		return fmt.Errorf("bytecode float add compare jump missing program")
	}
	plan, ok := program.floatAddCompareConstJumps[vm.ip]
	if !ok {
		return fmt.Errorf("bytecode float add compare jump missing lowering plan")
	}
	if cond, handled := vm.floatAddCompareConstCondition(instr.operator, plan); handled {
		if !cond {
			vm.ip = instr.target
			return nil
		}
		vm.ip++
		return nil
	}
	left, err := vm.floatAddCompareConstFallbackValue(plan)
	if err != nil {
		return err
	}
	cond, err := vm.compareBytecodeCondition(instr.operator, left, plan.rightImmediate)
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

func (vm *bytecodeVM) floatAddCompareConstCondition(op string, plan bytecodeFloatAddCompareConstJumpPlan) (bool, bool) {
	leftVal, leftKind, ok := vm.slotDirectFloatValue(plan.leftSlot)
	if !ok {
		return false, false
	}
	rightVal, rightKind, ok := vm.slotDirectFloatValue(plan.rightSlot)
	if !ok {
		return false, false
	}
	sum, sumKind, ok := bytecodeDirectFloatArithmeticRawFast("+", leftVal, leftKind, rightVal, rightKind)
	if !ok {
		return false, false
	}
	cond, ok := bytecodeDirectFloatCompareRawFast(op, sum, sumKind, plan.rightImmediate.Val, plan.rightImmediate.TypeSuffix)
	if !ok {
		return false, false
	}
	return cond.Val, true
}

func (vm *bytecodeVM) floatAddCompareConstFallbackValue(plan bytecodeFloatAddCompareConstJumpPlan) (runtime.Value, error) {
	left := vm.slotMaterializedValue(plan.leftSlot)
	right := vm.slotMaterializedValue(plan.rightSlot)
	return applyBinaryOperator(vm.interp, "+", left, right)
}
