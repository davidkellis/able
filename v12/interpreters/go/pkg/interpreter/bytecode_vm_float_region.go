package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlotFloatRegion(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode float region store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode float region target out of range")
	}
	plan, err := vm.floatRegionPlan(instr)
	if err != nil {
		return err
	}
	raw, kind, ok, err := vm.evalFloatRegionRaw(plan)
	if err != nil {
		return vm.finishStoreSlotFloatResult(instr, nil, err)
	}
	if !ok {
		result, fallbackErr := vm.evalFloatRegionFallback(plan)
		return vm.finishStoreSlotFloatResult(instr, result, fallbackErr)
	}
	return vm.finishStoreSlotFloatRegionRawResult(instr, raw, kind)
}

func (vm *bytecodeVM) floatRegionPlan(instr *bytecodeInstruction) (bytecodeFloatRegionPlan, error) {
	if vm == nil || vm.currentProgram == nil || instr == nil || instr.argCount < 0 || instr.argCount >= len(vm.currentProgram.floatRegions) {
		return bytecodeFloatRegionPlan{}, fmt.Errorf("bytecode float region plan out of range")
	}
	plan := vm.currentProgram.floatRegions[instr.argCount]
	if len(plan.steps) == 0 || plan.maxDepth == 0 || plan.maxDepth > bytecodeFloatRegionMaxDepth {
		return bytecodeFloatRegionPlan{}, fmt.Errorf("bytecode float region plan is invalid")
	}
	return plan, nil
}

func (vm *bytecodeVM) evalFloatRegionRaw(plan bytecodeFloatRegionPlan) (float64, runtime.FloatType, bool, error) {
	var values [bytecodeFloatRegionMaxDepth]float64
	var kinds [bytecodeFloatRegionMaxDepth]runtime.FloatType
	depth := 0
	for _, step := range plan.steps {
		switch step.kind {
		case bytecodeFloatRegionLoadSlot:
			if depth >= len(values) {
				return 0, runtime.FloatF64, false, fmt.Errorf("bytecode float region stack overflow")
			}
			value, kind, ok := vm.slotDirectFloatValueValidated(step.slot)
			if !ok {
				return 0, runtime.FloatF64, false, nil
			}
			values[depth], kinds[depth] = value, kind
			depth++
		case bytecodeFloatRegionConst:
			if depth >= len(values) {
				return 0, runtime.FloatF64, false, fmt.Errorf("bytecode float region stack overflow")
			}
			values[depth], kinds[depth] = step.value, step.floatKind
			depth++
		default:
			if depth < 2 {
				return 0, runtime.FloatF64, false, fmt.Errorf("bytecode float region stack underflow")
			}
			right := depth - 1
			left := right - 1
			value, kind, err := bytecodeFloatRegionRawBinary(step.kind, values[left], kinds[left], values[right], kinds[right])
			if err != nil {
				return 0, runtime.FloatF64, false, err
			}
			values[left], kinds[left] = value, kind
			depth--
		}
	}
	if depth != 1 {
		return 0, runtime.FloatF64, false, fmt.Errorf("bytecode float region result depth %d, want 1", depth)
	}
	return values[0], kinds[0], true, nil
}

func (vm *bytecodeVM) finishStoreSlotFloatRegionRawResult(instr *bytecodeInstruction, raw float64, kind runtime.FloatType) error {
	stored := vm.storeFloatSlotValue(instr.target, bytecodeRawFloatSlotValue(normalizeFloat(kind, raw), kind))
	if !instr.discardResult {
		vm.appendStackValue(stored)
	}
	vm.ip++
	return nil
}

func bytecodeFloatRegionRawBinary(kind bytecodeFloatRegionStepKind, left float64, leftKind runtime.FloatType, right float64, rightKind runtime.FloatType) (float64, runtime.FloatType, error) {
	switch kind {
	case bytecodeFloatRegionAdd:
		value, resultKind, _ := bytecodeDirectFloatArithmeticRawFast("+", left, leftKind, right, rightKind)
		return value, resultKind, nil
	case bytecodeFloatRegionSub:
		value, resultKind, _ := bytecodeDirectFloatArithmeticRawFast("-", left, leftKind, right, rightKind)
		return value, resultKind, nil
	case bytecodeFloatRegionMul:
		value, resultKind, _ := bytecodeDirectFloatArithmeticRawFast("*", left, leftKind, right, rightKind)
		return value, resultKind, nil
	case bytecodeFloatRegionDiv:
		value, resultKind, handled, err := bytecodeDirectFloatDivisionRawFast(left, leftKind, right, rightKind)
		if !handled && err == nil {
			return 0, runtime.FloatF64, fmt.Errorf("bytecode float region division was not handled")
		}
		return value, resultKind, err
	default:
		return 0, runtime.FloatF64, fmt.Errorf("bytecode float region operator %d unsupported", kind)
	}
}

func (vm *bytecodeVM) evalFloatRegionFallback(plan bytecodeFloatRegionPlan) (runtime.Value, error) {
	var values [bytecodeFloatRegionMaxDepth]runtime.Value
	depth := 0
	for _, step := range plan.steps {
		switch step.kind {
		case bytecodeFloatRegionLoadSlot:
			if depth >= len(values) {
				return nil, fmt.Errorf("bytecode float region fallback stack overflow")
			}
			if step.slot < 0 || step.slot >= len(vm.slots) {
				return nil, fmt.Errorf("bytecode float region source out of range")
			}
			values[depth] = vm.slotRuntimeValue(step.slot)
			depth++
		case bytecodeFloatRegionConst:
			if depth >= len(values) {
				return nil, fmt.Errorf("bytecode float region fallback stack overflow")
			}
			values[depth] = runtime.FloatValue{Val: step.value, TypeSuffix: step.floatKind}
			depth++
		default:
			if depth < 2 {
				return nil, fmt.Errorf("bytecode float region fallback stack underflow")
			}
			right := depth - 1
			left := right - 1
			result, err := applyBinaryOperator(vm.interp, bytecodeFloatRegionOperator(step.kind), values[left], values[right])
			if err != nil {
				return nil, err
			}
			values[left] = result
			depth--
		}
	}
	if depth != 1 {
		return nil, fmt.Errorf("bytecode float region fallback result depth %d, want 1", depth)
	}
	return values[0], nil
}

func bytecodeFloatRegionOperator(kind bytecodeFloatRegionStepKind) string {
	switch kind {
	case bytecodeFloatRegionAdd:
		return "+"
	case bytecodeFloatRegionSub:
		return "-"
	case bytecodeFloatRegionMul:
		return "*"
	case bytecodeFloatRegionDiv:
		return "/"
	default:
		return ""
	}
}
