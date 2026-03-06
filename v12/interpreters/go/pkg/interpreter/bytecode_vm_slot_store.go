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
	assignExpr, ok := instr.node.(*ast.AssignmentExpression)
	if !ok || assignExpr == nil {
		return value, value, true, nil
	}
	typedPattern, ok := typedIdentifierPatternFromTarget(assignExpr.Left)
	if !ok {
		return value, value, true, nil
	}
	bindings := make([]patternBinding, 0, 1)
	if err := vm.interp.collectPatternBindings(typedPattern, value, vm.env, &bindings); err != nil {
		if msg, ok := asPatternMismatch(err); ok {
			return nil, runtime.ErrorValue{Message: msg}, false, nil
		}
		return nil, nil, false, err
	}
	if len(bindings) == 1 {
		return bindings[0].value, value, true, nil
	}
	return value, value, true, nil
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
