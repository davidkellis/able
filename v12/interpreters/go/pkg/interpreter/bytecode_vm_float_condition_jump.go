package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

type bytecodeFloatMulAddMulCompareConstConditionResult struct {
	condition bool
	leftRaw   float64
	leftKind  runtime.FloatType
	rightRaw  float64
	rightKind runtime.FloatType
}

func (vm *bytecodeVM) execJumpIfFloatMulAddMulCompareConstFalse(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode float mul-add compare jump missing instruction")
	}
	if program == nil {
		return fmt.Errorf("bytecode float mul-add compare jump missing program")
	}
	plan, ok := program.floatMulAddMulJumps[vm.ip]
	if !ok {
		return fmt.Errorf("bytecode float mul-add compare jump missing lowering plan")
	}
	if result, handled, err := vm.floatMulAddMulCompareConstCondition(instr.operator, plan); handled || err != nil {
		if err != nil {
			return err
		}
		if !result.condition {
			vm.storeFloatMulAddMulCompareConstProductsRaw(plan, result)
			vm.ip = instr.target
			return nil
		}
		vm.ip++
		return nil
	}
	leftProduct, rightProduct, left, err := vm.floatMulAddMulCompareConstFallbackValues(plan)
	if err != nil {
		return err
	}
	cond, err := vm.compareBytecodeCondition(instr.operator, left, plan.rightImmediate)
	if err != nil {
		return err
	}
	if !cond {
		vm.storeFloatMulAddMulCompareConstProducts(plan, leftProduct, rightProduct)
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) floatMulAddMulCompareConstCondition(op string, plan bytecodeFloatMulAddMulCompareConstJumpPlan) (bytecodeFloatMulAddMulCompareConstConditionResult, bool, error) {
	leftProduct, leftProductKind, ok := vm.floatMulTermConditionValue(plan.leftMulLeftSlot, plan.leftMulRightSlot)
	if !ok {
		return bytecodeFloatMulAddMulCompareConstConditionResult{}, false, nil
	}
	rightProduct, rightProductKind, ok := vm.floatMulTermConditionValue(plan.rightMulLeftSlot, plan.rightMulRightSlot)
	if !ok {
		return bytecodeFloatMulAddMulCompareConstConditionResult{}, false, nil
	}
	sumKind := runtime.FloatF32
	if leftProductKind == runtime.FloatF64 || rightProductKind == runtime.FloatF64 {
		sumKind = runtime.FloatF64
	}
	sum := normalizeFloat(sumKind, leftProduct+rightProduct)
	cond, ok := bytecodeCompareFloat64(op, sum, plan.rightImmediate.Val)
	if !ok {
		return bytecodeFloatMulAddMulCompareConstConditionResult{}, false, nil
	}
	return bytecodeFloatMulAddMulCompareConstConditionResult{
		condition: cond,
		leftRaw:   leftProduct,
		leftKind:  leftProductKind,
		rightRaw:  rightProduct,
		rightKind: rightProductKind,
	}, true, nil
}

func (vm *bytecodeVM) floatMulTermConditionValue(leftSlot int, rightSlot int) (float64, runtime.FloatType, bool) {
	if leftSlot == rightSlot {
		val, kind, ok := vm.slotDirectFloatValue(leftSlot)
		if !ok {
			return 0, runtime.FloatType(""), false
		}
		return normalizeFloat(kind, val*val), kind, true
	}
	leftVal, leftKind, ok := vm.slotDirectFloatValue(leftSlot)
	if !ok {
		return 0, runtime.FloatType(""), false
	}
	rightVal, rightKind, ok := vm.slotDirectFloatValue(rightSlot)
	if !ok {
		return 0, runtime.FloatType(""), false
	}
	productKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		productKind = runtime.FloatF64
	}
	return normalizeFloat(productKind, leftVal*rightVal), productKind, true
}

func (vm *bytecodeVM) floatMulAddMulCompareConstFallbackValues(plan bytecodeFloatMulAddMulCompareConstJumpPlan) (runtime.Value, runtime.Value, runtime.Value, error) {
	leftLeft := vm.slotMaterializedValue(plan.leftMulLeftSlot)
	leftRight := vm.slotMaterializedValue(plan.leftMulRightSlot)
	leftProduct, err := applyBinaryOperator(vm.interp, "*", leftLeft, leftRight)
	if err != nil {
		return nil, nil, nil, err
	}
	rightLeft := vm.slotMaterializedValue(plan.rightMulLeftSlot)
	rightRight := vm.slotMaterializedValue(plan.rightMulRightSlot)
	rightProduct, err := applyBinaryOperator(vm.interp, "*", rightLeft, rightRight)
	if err != nil {
		return nil, nil, nil, err
	}
	sum, err := applyBinaryOperator(vm.interp, "+", leftProduct, rightProduct)
	if err != nil {
		return nil, nil, nil, err
	}
	return leftProduct, rightProduct, sum, nil
}

func (vm *bytecodeVM) storeFloatMulAddMulCompareConstProductsRaw(plan bytecodeFloatMulAddMulCompareConstJumpPlan, result bytecodeFloatMulAddMulCompareConstConditionResult) {
	if !plan.storeProducts {
		return
	}
	vm.storeReusableNormalizedFloatSlotRawDiscard(plan.leftTargetSlot, result.leftRaw, result.leftKind)
	vm.storeReusableNormalizedFloatSlotRawDiscard(plan.rightTargetSlot, result.rightRaw, result.rightKind)
}

func (vm *bytecodeVM) storeFloatMulAddMulCompareConstProducts(plan bytecodeFloatMulAddMulCompareConstJumpPlan, left runtime.Value, right runtime.Value) {
	if !plan.storeProducts {
		return
	}
	vm.storeFloatSlotValue(plan.leftTargetSlot, left)
	vm.storeFloatSlotValue(plan.rightTargetSlot, right)
}

func bytecodeCompareFloat64(op string, left float64, right float64) (bool, bool) {
	switch op {
	case "<":
		return left < right, true
	case "<=":
		return left <= right, true
	case ">":
		return left > right, true
	case ">=":
		return left >= right, true
	case "==":
		return left == right, true
	case "!=":
		return left != right, true
	default:
		return false, false
	}
}
