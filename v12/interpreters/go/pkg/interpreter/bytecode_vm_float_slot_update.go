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
	if len(vm.stack) < 3 {
		return fmt.Errorf("bytecode stack underflow")
	}
	baseIdx := len(vm.stack) - 3
	base := vm.stack[baseIdx]
	mulLeft := vm.stack[baseIdx+1]
	mulRight := vm.stack[baseIdx+2]
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
	result = bytecodeStackResultValue(result)
	vm.stack = vm.stack[:baseIdx]
	vm.slots[instr.target] = result
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(result)
	}
	if !instr.discardResult {
		vm.stack = append(vm.stack, result)
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) storeSlotFloatAddMulResult(instr *bytecodeInstruction, base runtime.Value, mulLeft runtime.Value, mulRight runtime.Value) (runtime.Value, error) {
	if result, ok := bytecodeDirectFloatAddMul(base, mulLeft, mulRight); ok {
		return result, nil
	}
	base = bytecodeSlotReadValue(base)
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

func bytecodeDirectFloatAddMulValue(base runtime.Value, mulLeft runtime.Value, mulRight runtime.Value) (runtime.FloatValue, bool) {
	baseVal, baseKind, ok := bytecodeDirectFloatValue(base)
	if !ok {
		return runtime.FloatValue{}, false
	}
	leftVal, leftKind, ok := bytecodeDirectFloatValue(mulLeft)
	if !ok {
		return runtime.FloatValue{}, false
	}
	rightVal, rightKind, ok := bytecodeDirectFloatValue(mulRight)
	if !ok {
		return runtime.FloatValue{}, false
	}
	productKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		productKind = runtime.FloatF64
	}
	product := normalizeFloat(productKind, leftVal*rightVal)
	resultKind := runtime.FloatF32
	if baseKind == runtime.FloatF64 || productKind == runtime.FloatF64 {
		resultKind = runtime.FloatF64
	}
	return runtime.FloatValue{
		Val:        normalizeFloat(resultKind, baseVal+product),
		TypeSuffix: resultKind,
	}, true
}

func bytecodeDirectFloatAddMulRaw(base runtime.Value, leftVal float64, leftKind runtime.FloatType, rightVal float64, rightKind runtime.FloatType) (runtime.FloatValue, bool) {
	baseVal, baseKind, ok := bytecodeDirectFloatValue(base)
	if !ok {
		return runtime.FloatValue{}, false
	}
	productKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		productKind = runtime.FloatF64
	}
	product := normalizeFloat(productKind, leftVal*rightVal)
	resultKind := runtime.FloatF32
	if baseKind == runtime.FloatF64 || productKind == runtime.FloatF64 {
		resultKind = runtime.FloatF64
	}
	return runtime.FloatValue{
		Val:        normalizeFloat(resultKind, baseVal+product),
		TypeSuffix: resultKind,
	}, true
}

func bytecodeSlotReadValue(value runtime.Value) runtime.Value {
	if fv, ok := value.(*runtime.FloatValue); ok && fv != nil {
		return runtime.FloatValue{Val: fv.Val, TypeSuffix: fv.TypeSuffix}
	}
	return value
}

func (vm *bytecodeVM) storeOwnedFloatSlot(target int, value runtime.FloatValue) runtime.Value {
	if target < 0 || target >= len(vm.slots) {
		return value
	}
	key := &vm.slots[target]
	if vm.ownedFloatSlots != nil {
		if cell := vm.ownedFloatSlots[key]; cell != nil {
			cell.Val = value.Val
			cell.TypeSuffix = value.TypeSuffix
			vm.slots[target] = cell
			return cell
		}
	} else {
		vm.ownedFloatSlots = make(map[*runtime.Value]*runtime.FloatValue, 4)
	}
	cell := &runtime.FloatValue{Val: value.Val, TypeSuffix: value.TypeSuffix}
	vm.ownedFloatSlots[key] = cell
	vm.slots[target] = cell
	return cell
}
