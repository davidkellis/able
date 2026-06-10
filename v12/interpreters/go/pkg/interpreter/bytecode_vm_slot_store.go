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
		if uint(instr.target) < uint(len(vm.slots)) {
			vm.appendSlotStackValueChecked(instr.target)
		} else {
			vm.appendStackValue(nil)
		}
		vm.ip++
		return nil
	case bytecodeOpLoadImplicitSlot:
		return vm.execLoadImplicitSlot(instr)
	case bytecodeOpLoadSlotI32:
		return vm.execLoadSlotI32(instr)
	case bytecodeOpLoadSlotStructField:
		return vm.execLoadSlotStructField(instr)
	default:
		return fmt.Errorf("bytecode slot load opcode %d unsupported", instr.op)
	}
}

func (vm *bytecodeVM) execLoadSlotStructField(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot struct field load missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	obj := vm.slotStackValueChecked(instr.target)
	if plan, ok := bytecodeNamedStructMemberPlanForInstruction(vm.currentProgram, vm.ip, instr); ok {
		if val, ok := bytecodeDirectPlannedStructMemberValue(obj, plan, false); ok {
			vm.appendStackValue(val)
			vm.ip++
			return nil
		}
	}
	obj = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, obj)
	if val, ok := bytecodeDirectStructMemberValue(obj, instr.name, false); ok {
		vm.appendStackValue(val)
		vm.ip++
		return nil
	}
	if instr.name == "" {
		return fmt.Errorf("bytecode slot struct field load requires field name")
	}
	return fmt.Errorf("Missing field '%s' during destructuring", instr.name)
}

func (vm *bytecodeVM) execStoreSlotOpcode(instr *bytecodeInstruction, program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable) (bool, error) {
	if instr == nil {
		return false, fmt.Errorf("bytecode slot store missing instruction")
	}
	switch instr.op {
	case bytecodeOpStoreSlot, bytecodeOpStoreSlotNew:
		return false, vm.execStoreSlot(instr)
	case bytecodeOpStoreImplicitSlot:
		return false, vm.execStoreImplicitSlot(instr)
	case bytecodeOpStoreSlotI32:
		return false, vm.execStoreSlotI32(instr)
	case bytecodeOpStoreSlotCastSlotFloatConstDiv:
		return false, vm.execStoreSlotCastSlotFloatConstDiv(instr)
	case bytecodeOpStoreSlotFloatAffine:
		return false, vm.execStoreSlotFloatAffine(instr)
	case bytecodeOpStoreSlotFloatRegion:
		return false, vm.execStoreSlotFloatRegion(instr)
	case bytecodeOpStoreSlotFloatBinary:
		return false, vm.execStoreSlotFloatBinary(instr)
	case bytecodeOpStoreSlotFloatAddSub:
		return false, vm.execStoreSlotFloatAddSub(instr)
	case bytecodeOpStoreSlotFloatAddMul:
		return false, vm.execStoreSlotFloatAddMul(instr)
	case bytecodeOpStoreSlotFloatAddMulSlot:
		return false, vm.execStoreSlotFloatAddMulSlot(instr)
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
	if instr.discardResult {
		if err, handled := vm.execStoreSlotBinaryIntSlotConstDiscardI32Fast(instr, rightImmediate); handled {
			return err
		}
	}
	if result, handled, err := vm.storeSlotBinaryIntSlotConstI32RawFastResult(instr, rightImmediate); handled {
		return vm.finishStoreSlotBinaryIntSlotConstFastResult(instr, result, err)
	}
	if result, handled, err := vm.storeSlotBinaryIntSlotConstFastResult(instr, rightImmediate); handled {
		return vm.finishStoreSlotBinaryIntSlotConstFastResult(instr, result, err)
	}
	binaryInstr := *instr
	switch instr.operator {
	case "+":
		binaryInstr.op = bytecodeOpBinaryIntAddSlotConst
	case "-":
		binaryInstr.op = bytecodeOpBinaryIntSubSlotConst
	case "*":
		binaryInstr.op = bytecodeOpBinaryIntMulSlotConst
	case "%":
		binaryInstr.op = bytecodeOpBinaryIntModSlotConst
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
	if raw, ok := result.(bytecodeRawI32SlotValue); ok && vm.hasI32RegisterFrame() && vm.setI32RegisterRaw(instr.target, int32(raw)) {
		vm.slots[instr.target] = nil
	} else {
		vm.slots[instr.target] = result
		if vm.hasI32RegisterFrame() {
			vm.setI32RegisterValue(instr.target, result)
		}
	}
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(result)
	}
	if !instr.discardResult {
		vm.appendStackValue(result)
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execStoreSlotBinaryIntSlotConstDiscardI32Fast(instr *bytecodeInstruction, right runtime.IntegerValue) (error, bool) {
	if instr == nil || !instr.discardResult || vm.hasI32RegisterFrame() {
		return nil, false
	}
	raw, handled, err := vm.storeSlotBinaryIntSlotConstI32RawFast(instr, right)
	if !handled {
		return nil, false
	}
	if err != nil {
		if vm.interp != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
		}
		if instr.node != nil && vm.interp != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env)), true
		}
		return err, true
	}
	if cell, ok := vm.slots[instr.target].(*runtime.IntegerValue); ok && cell != nil {
		cell.ResetSmall(int64(raw), runtime.IntegerI32)
	} else {
		vm.storeOwnedI32SlotRaw(instr.target, raw)
	}
	vm.clearActiveValueSlotFloat(instr.target)
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Raw(raw)
	}
	vm.ip++
	return nil, true
}

func (vm *bytecodeVM) finishStoreSlotBinaryIntSlotConstFastResult(instr *bytecodeInstruction, result runtime.Value, err error) error {
	if err != nil {
		if vm.interp != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
		}
		if instr.node != nil && vm.interp != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if raw, ok := result.(bytecodeRawI32SlotValue); ok && vm.hasI32RegisterFrame() && vm.setI32RegisterRaw(instr.target, int32(raw)) {
		vm.clearActiveValueSlotI32(instr.target)
		vm.clearActiveValueSlotFloat(instr.target)
		vm.slots[instr.target] = nil
	} else {
		vm.clearActiveValueSlotI32(instr.target)
		vm.clearActiveValueSlotFloat(instr.target)
		vm.slots[instr.target] = result
		if vm.hasI32RegisterFrame() {
			vm.setI32RegisterValue(instr.target, result)
		}
	}
	if instr.target == 0 {
		vm.setSelfFastSlot0I32Value(result)
	}
	if !instr.discardResult {
		vm.appendStackValue(bytecodeSlotReadValue(result))
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) storeSlotBinaryIntSlotConstI32RawFastResult(instr *bytecodeInstruction, right runtime.IntegerValue) (runtime.Value, bool, error) {
	raw, handled, err := vm.storeSlotBinaryIntSlotConstI32RawFast(instr, right)
	if !handled || err != nil {
		return nil, handled, err
	}
	return bytecodeRawI32SlotCachedValue(raw), true, nil
}

func (vm *bytecodeVM) storeSlotBinaryIntSlotConstI32RawFast(instr *bytecodeInstruction, right runtime.IntegerValue) (int32, bool, error) {
	if instr == nil || !instr.hasIntImmediate || !instr.hasIntRaw || right.TypeSuffix != runtime.IntegerI32 {
		return 0, false, nil
	}
	if vm.hasI32RegisterFrame() {
		if leftRaw, ok := vm.i32RegisterRaw(instr.target); ok {
			return storeSlotBinaryIntSlotConstI32Raw(instr.operator, int64(leftRaw), instr.intImmediateRaw)
		}
	}
	leftVal, ok := vm.slotDirectSmallI32Value(instr.target)
	if !ok {
		return 0, false, nil
	}
	return storeSlotBinaryIntSlotConstI32Raw(instr.operator, leftVal, instr.intImmediateRaw)
}

func storeSlotBinaryIntSlotConstI32Raw(operator string, leftVal int64, rightVal int64) (int32, bool, error) {
	var result int64
	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "%":
		if rightVal == 0 {
			return 0, true, newDivisionByZeroError()
		}
		_, result = euclideanDivModInt64(leftVal, rightVal)
	default:
		return 0, false, nil
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		return 0, true, newOverflowError("integer overflow")
	}
	return int32(result), true, nil
}

func (vm *bytecodeVM) storeSlotBinaryIntSlotConstFastResult(instr *bytecodeInstruction, right runtime.IntegerValue) (runtime.Value, bool, error) {
	rightRef := &right
	if instr == nil || !rightRef.IsSmallRef() {
		return nil, false, nil
	}
	if right.TypeSuffix == runtime.IntegerI32 {
		if vm.hasI32RegisterFrame() {
			if leftRaw, ok := vm.i32RegisterRaw(instr.target); ok {
				return storeSlotBinaryIntSlotConstI32FastResult(instr.operator, int64(leftRaw), rightRef.Int64FastRef(), instr.discardResult)
			}
		}
		if leftVal, ok := vm.slotDirectSmallI32Value(instr.target); ok {
			return storeSlotBinaryIntSlotConstI32FastResult(instr.operator, leftVal, rightRef.Int64FastRef(), instr.discardResult)
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

func storeSlotBinaryIntSlotConstI32FastResult(operator string, leftVal int64, rightVal int64, discardResult bool) (runtime.Value, bool, error) {
	return storeSlotBinaryIntSlotConstI32RawResult(operator, leftVal, rightVal, discardResult)
}

func storeSlotBinaryIntSlotConstI32RawResult(operator string, leftVal int64, rightVal int64, discardResult bool) (runtime.Value, bool, error) {
	var result int64
	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "%":
		if rightVal == 0 {
			return nil, true, newDivisionByZeroError()
		}
		_, result = euclideanDivModInt64(leftVal, rightVal)
	default:
		return nil, false, nil
	}
	if result < math.MinInt32 || result > math.MaxInt32 {
		return nil, true, newOverflowError("integer overflow")
	}
	return bytecodeRawI32ResultValue(result), true, nil
}

func (vm *bytecodeVM) execStoreSlot(instr *bytecodeInstruction) error {
	if instr == nil {
		return fmt.Errorf("bytecode slot store missing instruction")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return fmt.Errorf("bytecode slot out of range")
	}
	if vm.stackDepth() == 0 {
		return fmt.Errorf("bytecode stack underflow")
	}
	val := vm.stackValue(vm.stackDepth() - 1)
	if !instr.storeTyped || instr.typeExpr == nil {
		storedRaw := false
		storedInteger := false
		if instr.storeRawI32Sidecar && !vm.hasI32RegisterFrame() {
			if raw, ok := bytecodeRawI32Value(val); ok {
				vm.clearActiveValueSlotFloat(instr.target)
				storedRaw = vm.storeActiveValueSlotI32Raw(instr.target, raw)
			} else {
				vm.clearActiveValueSlotI32(instr.target)
			}
		}
		if !storedRaw {
			vm.clearActiveValueSlotI32(instr.target)
		}
		if !storedRaw {
			if _, ok := vm.tryStoreRawIntegerSlotValue(instr.target, val); ok {
				vm.clearActiveValueSlotFloat(instr.target)
				storedInteger = true
			}
		}
		if !storedRaw && !storedInteger {
			if fv, ok := val.(runtime.FloatValue); ok {
				vm.storeOwnedFloatSlot(instr.target, fv)
			} else {
				vm.clearActiveValueSlotFloat(instr.target)
				vm.slots[instr.target] = val
			}
		}
		if !storedRaw && vm.hasI32RegisterFrame() {
			vm.setI32RegisterValue(instr.target, vm.slots[instr.target])
		}
		if instr.target == 0 {
			if storedRaw {
				vm.setSelfFastSlot0I32Value(val)
			} else {
				vm.setSelfFastSlot0I32Value(vm.slots[instr.target])
			}
		}
		if instr.discardResult {
			vm.truncateStack(vm.stackDepth() - 1)
		}
		vm.ip++
		return nil
	}
	if handled, err := vm.tryExecStoreTypedExactRawInteger(instr, val); handled || err != nil {
		return err
	}
	storeVal, stackVal, shouldStore, err := vm.typedSlotAssignmentValues(*instr, val)
	if err != nil {
		if instr.node != nil {
			return vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	if shouldStore {
		storedInRegister := false
		vm.clearActiveValueSlotI32(instr.target)
		vm.clearActiveValueSlotFloat(instr.target)
		if instr.discardResult && vm.hasI32RegisterFrame() && vm.setI32RegisterValue(instr.target, storeVal) {
			vm.slots[instr.target] = nil
			storedInRegister = true
		} else {
			vm.clearActiveValueSlotI32(instr.target)
			if fv, ok := storeVal.(runtime.FloatValue); ok {
				vm.storeOwnedFloatSlot(instr.target, fv)
			} else {
				vm.clearActiveValueSlotFloat(instr.target)
				vm.slots[instr.target] = storeVal
			}
			if vm.hasI32RegisterFrame() {
				vm.setI32RegisterValue(instr.target, vm.slots[instr.target])
			}
		}
		if instr.target == 0 {
			if storedInRegister {
				vm.setSelfFastSlot0I32Value(storeVal)
			} else {
				vm.setSelfFastSlot0I32Value(vm.slots[instr.target])
			}
		}
	}
	if instr.discardResult {
		vm.truncateStack(vm.stackDepth() - 1)
		vm.ip++
		return nil
	}
	if stackVal == nil {
		stackVal = runtime.NilValue{}
	}
	vm.setStackValue(vm.stackDepth()-1, stackVal)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) typedSlotAssignmentValues(instr bytecodeInstruction, value runtime.Value) (runtime.Value, runtime.Value, bool, error) {
	if !instr.storeTyped || instr.typeExpr == nil {
		return value, value, true, nil
	}
	value = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonPattern, value)
	typeExpr := vm.canonicalRuntimeTypeExpression(instr.typeExpr)
	if !vm.interp.matchesType(typeExpr, value) {
		expected := typeExpressionToString(typeExpr)
		actualExpr := vm.interp.typeExpressionFromValue(value)
		actual := value.Kind().String()
		if actualExpr != nil {
			actual = typeExpressionToString(actualExpr)
		}
		return nil, runtime.ErrorValue{
			Message: fmt.Sprintf("Typed pattern mismatch in assignment: expected %s, got %s", expected, actual),
		}, false, nil
	}
	coerced, err := vm.interp.coerceValueToType(typeExpr, value)
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
