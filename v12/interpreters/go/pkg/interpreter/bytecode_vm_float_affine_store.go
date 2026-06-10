package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotFloatAffine(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode float affine store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	program := vm.currentProgram
	if program == nil {
		return fmt.Errorf("bytecode float affine store missing active program")
	}
	plan, ok := program.floatAffineStores[vm.ip]
	if !ok {
		return fmt.Errorf("bytecode float affine store missing plan")
	}
	if !plan.validForSlots(len(vm.slots)) {
		return fmt.Errorf("bytecode float affine store invalid plan")
	}
	if resultVal, resultKind, ok, err := vm.bytecodeStoreSlotFloatAffineRawFast(program, plan); ok || err != nil {
		if err != nil {
			return vm.finishStoreSlotFloatResult(instr, nil, err)
		}
		return vm.finishStoreSlotFloatRawResult(instr, resultVal, resultKind)
	}
	result, err := vm.storeSlotFloatAffineGenericResult(program, plan)
	return vm.finishStoreSlotFloatResult(instr, result, err)
}

func (vm *bytecodeVM) bytecodeStoreSlotFloatAffineRawFast(program *bytecodeProgram, plan bytecodeStoreSlotFloatAffinePlan) (float64, runtime.FloatType, bool, error) {
	leftVal, ok := vm.bytecodeCastSlotFloatValue(plan.sourceSlot, plan.targetKind)
	if !ok {
		return 0, runtime.FloatF64, false, nil
	}
	divisorVal, ok, err := vm.bytecodeFloatAffineDivisorRaw(program, plan)
	if err != nil || !ok {
		return 0, runtime.FloatF64, ok, err
	}
	productVal, productKind, handled := bytecodeDirectFloatArithmeticRawFast("*", plan.scaleVal, plan.scaleKind, leftVal, plan.targetKind)
	if !handled {
		return 0, runtime.FloatF64, false, nil
	}
	quotientVal, quotientKind, handled, err := bytecodeDirectFloatDivisionRawFast(productVal, productKind, divisorVal, plan.targetKind)
	if err != nil {
		return 0, runtime.FloatF64, true, err
	}
	if !handled {
		return 0, runtime.FloatF64, false, nil
	}
	resultVal, resultKind, handled := bytecodeDirectFloatArithmeticRawFast("-", quotientVal, quotientKind, plan.offsetVal, plan.offsetKind)
	if !handled {
		return 0, runtime.FloatF64, false, nil
	}
	return resultVal, resultKind, true, nil
}

func (vm *bytecodeVM) bytecodeFloatAffineDivisorRaw(program *bytecodeProgram, plan bytecodeStoreSlotFloatAffinePlan) (float64, bool, error) {
	if plan.divisorSlot >= 0 {
		val, ok := vm.bytecodeCastSlotFloatValue(plan.divisorSlot, plan.targetKind)
		return val, ok, nil
	}
	if plan.divisorName == "" {
		return 0, false, nil
	}
	divisor, err := vm.resolveCachedIdentifierName(program, vm.ip, plan.divisorName)
	if err != nil {
		return 0, false, err
	}
	val, ok := bytecodeCastValueToFloatRaw(divisor, plan.targetKind)
	return val, ok, nil
}

func (vm *bytecodeVM) storeSlotFloatAffineGenericResult(program *bytecodeProgram, plan bytecodeStoreSlotFloatAffinePlan) (runtime.Value, error) {
	left := vm.slotRuntimeValue(plan.sourceSlot)
	castedLeft, err := vm.interp.castValueToType(ast.Ty(string(plan.targetKind)), left)
	if err != nil {
		return nil, err
	}
	divisor, err := vm.bytecodeFloatAffineDivisorValue(program, plan)
	if err != nil {
		return nil, err
	}
	castedDivisor, err := vm.interp.castValueToType(ast.Ty(string(plan.targetKind)), divisor)
	if err != nil {
		return nil, err
	}
	product, err := applyBinaryOperator(vm.interp, "*", runtime.FloatValue{Val: plan.scaleVal, TypeSuffix: plan.scaleKind}, castedLeft)
	if err != nil {
		return nil, err
	}
	quotient, err := applyBinaryOperator(vm.interp, "/", product, castedDivisor)
	if err != nil {
		return nil, err
	}
	return applyBinaryOperator(vm.interp, "-", quotient, runtime.FloatValue{Val: plan.offsetVal, TypeSuffix: plan.offsetKind})
}

func (vm *bytecodeVM) bytecodeFloatAffineDivisorValue(program *bytecodeProgram, plan bytecodeStoreSlotFloatAffinePlan) (runtime.Value, error) {
	if plan.divisorSlot >= 0 {
		return vm.slotRuntimeValue(plan.divisorSlot), nil
	}
	return vm.resolveCachedIdentifierName(program, vm.ip, plan.divisorName)
}
