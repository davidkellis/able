package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execStoreSlot(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if len(vm.stack) == 0 {
		return fmt.Errorf("bytecode stack underflow")
	}
	val := vm.stack[len(vm.stack)-1]
	if !instr.storeTyped || instr.typeExpr == nil {
		vm.slots[instr.target] = val
		vm.ip++
		return nil
	}
	storeVal, stackVal, shouldStore, err := vm.typedSlotAssignmentValues(*instr, val)
	if err != nil {
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if shouldStore {
		vm.slots[instr.target] = storeVal
	}
	if stackVal == nil {
		stackVal = runtime.NilValue{}
	}
	vm.stack[len(vm.stack)-1] = stackVal
	vm.ip++
	return nil
}

func (vm *bytecodeVM) typedSlotAssignmentValues(instr bytecodeInstruction, value runtime.Value) (runtime.Value, runtime.Value, bool, error) {
	if !instr.storeTyped || instr.typeExpr == nil {
		return value, value, true, nil
	}
	if !vm.interp.matchesType(instr.typeExpr, value) {
		expected := typeExpressionToString(instr.typeExpr)
		actualExpr := vm.interp.typeExpressionFromValue(value)
		actual := value.Kind().String()
		if actualExpr != nil {
			actual = typeExpressionToString(actualExpr)
		}
		return nil, runtime.ErrorValue{
			Message: fmt.Sprintf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual),
		}, false, nil
	}
	coerced, err := vm.interp.coerceValueToType(instr.typeExpr, value)
	if err != nil {
		return nil, nil, false, err
	}
	return coerced, value, true, nil
}

func typedIdentifierPatternFromTarget(target ast.AssignmentTarget) (*ast.TypedPattern, bool) {
	typedPattern, ok := target.(*ast.TypedPattern)
	if !ok || typedPattern == nil {
		return nil, false
	}
	_, ok = resolvePatternTargetName(typedPattern.Pattern)
	if !ok {
		return nil, false
	}
	return typedPattern, true
}
