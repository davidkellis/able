package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execLoadSlotOpcode(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot load missing instruction")
	}
	switch instr.op {
	case bytecodeOpLoadSlot:
		vm.stack = append(vm.stack, bytecodeSlotReadValue(vm.slots[instr.target]))
		vm.ip++
		return nil
	case bytecodeOpLoadSlotI32:
		return vm.execLoadSlotI32(instr)
	default:
		return fmt.Errorf("bytecode slot load opcode %d unsupported", instr.op)
	}
}

func (vm *bytecodeVM) execStoreSlotOpcode(instr *bytecodeInstruction, program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable) (bool, error) {
	if instr == nil {
		return false, fmt.Errorf("bytecode slot store missing instruction")
	}
	switch instr.op {
	case bytecodeOpStoreSlot, bytecodeOpStoreSlotNew:
		return false, vm.execStoreSlot(instr)
	case bytecodeOpStoreSlotI32:
		return false, vm.execStoreSlotI32(instr)
	case bytecodeOpStoreSlotFloatAddMul:
		return false, vm.execStoreSlotFloatAddMul(instr)
	case bytecodeOpStoreSlotFloatAddMulArrayGet:
		return vm.execStoreSlotFloatAddMulArrayGet(program, instructions, validatedIntConsts, slotConstIntImmTable, instr)
	default:
		return false, fmt.Errorf("bytecode slot store opcode %d unsupported", instr.op)
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
	if result, handled, err := vm.storeSlotBinaryIntSlotConstFastResult(instr, rightImmediate); handled {
		if err != nil {
			if vm.interp != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
			}
			if instr.node != nil && vm.interp != nil {
				return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return err
		}
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
	binaryInstr := *instr
	switch instr.operator {
	case "+":
		binaryInstr.op = bytecodeOpBinaryIntAddSlotConst
	case "-":
		binaryInstr.op = bytecodeOpBinaryIntSubSlotConst
	case "*":
		binaryInstr.op = bytecodeOpBinaryIntMulSlotConst
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
	if !instr.discardResult {
		vm.stack = append(vm.stack, result)
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) storeSlotBinaryIntSlotConstFastResult(instr *bytecodeInstruction, right runtime.IntegerValue) (runtime.Value, bool, error) {
	rightRef := &right
	if instr == nil || !rightRef.IsSmallRef() {
		return nil, false, nil
	}
	if right.TypeSuffix == runtime.IntegerI32 {
		if leftVal, ok := bytecodeDirectSmallI32Value(vm.slots[instr.target]); ok {
			return storeSlotBinaryIntSlotConstI32FastResult(instr.operator, leftVal, rightRef.Int64FastRef())
		}
	}
	rightVal := rightRef.Int64FastRef()
	compute := func(kind runtime.IntegerType, leftVal int64) (runtime.Value, bool, error) {
		if kind != right.TypeSuffix {
			return nil, false, nil
		}
		var (
			result   int64
			overflow bool
		)
		switch instr.operator {
		case "+":
			result, overflow = addInt64Overflow(leftVal, rightVal)
		case "-":
			result, overflow = subInt64Overflow(leftVal, rightVal)
		case "*":
			result, overflow = mulInt64Overflow(leftVal, rightVal)
		default:
			return nil, false, nil
		}
		if overflow {
			return nil, false, nil
		}
		if err := ensureFitsInt64Type(kind, result); err != nil {
			return nil, true, err
		}
		return boxedOrSmallIntegerValue(kind, result), true, nil
	}
	switch left := vm.slots[instr.target].(type) {
	case runtime.IntegerValue:
		leftRef := &left
		if leftRef.IsSmallRef() {
			return compute(left.TypeSuffix, leftRef.Int64FastRef())
		}
	case *runtime.IntegerValue:
		if left != nil && left.IsSmallRef() {
			return compute(left.TypeSuffix, left.Int64FastRef())
		}
	}
	return nil, false, nil
}

func storeSlotBinaryIntSlotConstI32FastResult(operator string, leftVal int64, rightVal int64) (runtime.Value, bool, error) {
	var result int64
	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	default:
		return nil, false, nil
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		return nil, true, newOverflowError("integer overflow")
	}
	return bytecodeBoxedIntegerI32Value(result), true, nil
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
		if fv, ok := val.(runtime.FloatValue); ok {
			vm.storeOwnedFloatSlot(instr.target, fv)
		} else {
			vm.slots[instr.target] = val
		}
		if instr.target == 0 {
			vm.setSelfFastSlot0I32Value(vm.slots[instr.target])
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
		if fv, ok := storeVal.(runtime.FloatValue); ok {
			vm.storeOwnedFloatSlot(instr.target, fv)
		} else {
			vm.slots[instr.target] = storeVal
		}
		if instr.target == 0 {
			vm.setSelfFastSlot0I32Value(vm.slots[instr.target])
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
