package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execLoadSlotOpcode(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot load missing instruction")
	}
	switch instr.op {
	case bytecodeOpLoadSlot:
		vm.stack = append(vm.stack, vm.slots[instr.target])
		vm.ip++
		return nil
	case bytecodeOpLoadSlotI32:
		return vm.execLoadSlotI32(instr)
	default:
		return fmt.Errorf("bytecode slot load opcode %d unsupported", instr.op)
	}
}

func (vm *bytecodeVM) execStoreSlotOpcode(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot store missing instruction")
	}
	switch instr.op {
	case bytecodeOpStoreSlot, bytecodeOpStoreSlotNew:
		return vm.execStoreSlot(instr)
	case bytecodeOpStoreSlotI32:
		return vm.execStoreSlotI32(instr)
	default:
		return fmt.Errorf("bytecode slot store opcode %d unsupported", instr.op)
	}
}

func (vm *bytecodeVM) execStoreSlotBinaryIntSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeImmediateIntegerValue(instr.value)
	}
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
	}
	if !hasImmediate {
		return fmt.Errorf("bytecode slot-const store missing integer immediate")
	}
	binaryInstr := *instr
	switch instr.operator {
	case "+":
		binaryInstr.op = bytecodeOpBinaryIntAddSlotConst
	case "-":
		binaryInstr.op = bytecodeOpBinaryIntSubSlotConst
	default:
		return fmt.Errorf("bytecode slot-const store unsupported operator %q", instr.operator)
	}
	result, handled, err := vm.execBinarySlotConst(&binaryInstr, rightImmediate, true)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if !handled {
		return fmt.Errorf("bytecode slot-const store was not handled")
	}
	result = bytecodeStackResultValue(result)
	vm.slots[instr.target] = result
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(result)
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil
}

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
		if instr.target == 0 {
			vm.setSelfFastSlot0I32Value(val)
		}
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
		if instr.target == 0 {
			vm.setSelfFastSlot0I32Value(storeVal)
		}
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
