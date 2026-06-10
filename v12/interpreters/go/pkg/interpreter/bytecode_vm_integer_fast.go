package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeSimpleIntegerTargetKind(typeExpr ast.TypeExpression) (runtime.IntegerType, bool) {
	targetKind := runtime.IntegerType(cachedSimpleTypeName(typeExpr))
	if _, ok := lookupIntegerInfo(targetKind); !ok {
		return "", false
	}
	return targetKind, true
}

func (vm *bytecodeVM) tryExecCastFast(value runtime.Value, typeExpr ast.TypeExpression) (runtime.Value, bool, error) {
	targetKind, ok := bytecodeSimpleIntegerTargetKind(typeExpr)
	if ok {
		info, infoOK := lookupIntegerInfo(targetKind)
		if !infoOK {
			return nil, true, fmt.Errorf("unsupported integer cast target %s", targetKind)
		}
		if _, raw, rawOK := bytecodeRawIntegerValueInfo(value); rawOK {
			if result, handled := bytecodeCastSmallIntToIntegerKindRawResult(raw, targetKind, info); handled {
				return result, true, nil
			}
		}
		return nil, false, nil
	}

	typeName := cachedSimpleTypeName(typeExpr)
	if typeName != "f32" && typeName != "f64" {
		return nil, false, nil
	}
	if _, raw, ok := bytecodeRawIntegerValueInfo(value); !ok {
		return nil, false, nil
	} else {
		targetFloat := runtime.FloatType(typeName)
		return runtime.FloatValue{
			Val:        normalizeFloat(targetFloat, float64(raw)),
			TypeSuffix: targetFloat,
		}, true, nil
	}
}

func (vm *bytecodeVM) execCastOpcode(instr *bytecodeInstruction) error {
	if vm == nil || vm.interp == nil {
		return fmt.Errorf("bytecode cast missing VM or interpreter")
	}
	val, err := vm.pop()
	if err != nil {
		return err
	}
	castExpr, ok := instr.node.(*ast.TypeCastExpression)
	if !ok || castExpr == nil {
		return fmt.Errorf("bytecode cast expects node")
	}
	targetType := vm.canonicalRuntimeTypeExpression(castExpr.TargetType)
	if targetKind, ok := bytecodeSimpleIntegerTargetKind(targetType); ok {
		info, infoOK := lookupIntegerInfo(targetKind)
		if !infoOK {
			return fmt.Errorf("unsupported integer cast target %s", targetKind)
		}
		if _, raw, rawOK := bytecodeRawIntegerValueInfo(val); rawOK {
			if castRaw, handled := bytecodeCastSmallIntToIntegerKindRawBits(raw, targetKind, info); handled {
				vm.appendRawIntegerStack(targetKind, castRaw)
				vm.ip++
				return nil
			}
		}
	}
	result, handled, err := vm.tryExecCastFast(val, targetType)
	if err != nil {
		err = vm.interp.attachRuntimeContext(err, castExpr, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return nil
		}
		return err
	}
	if !handled {
		result, err = vm.interp.castValueToType(targetType, vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCast, val))
		if err != nil {
			err = vm.interp.attachRuntimeContext(err, castExpr, vm.interp.stateFromEnv(vm.env))
			if vm.handleLoopSignal(err) {
				return nil
			}
			return err
		}
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.appendStackValue(result)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) tryExecStoreTypedExactRawInteger(instr *bytecodeInstruction, value runtime.Value) (bool, error) {
	if vm == nil || instr == nil || !instr.storeTyped || instr.typeExpr == nil {
		return false, nil
	}
	targetKind, ok := bytecodeSimpleIntegerTargetKind(instr.typeExpr)
	if !ok || !bytecodeRawIntegerKindSupported(targetKind) {
		return false, nil
	}
	valueKind, raw, ok := bytecodeRawIntegerValueInfo(value)
	if !ok || valueKind != targetKind {
		return false, nil
	}
	if err := ensureFitsInt64Type(targetKind, raw); err != nil {
		return false, nil
	}
	storedValue := runtime.Value(nil)
	storedInRegister := false

	vm.clearActiveValueSlotI32(instr.target)
	vm.clearActiveValueSlotFloat(instr.target)

	if instr.discardResult &&
		targetKind == runtime.IntegerI32 &&
		raw >= math.MinInt32 &&
		raw <= math.MaxInt32 &&
		vm.hasI32RegisterFrame() &&
		vm.setI32RegisterRaw(instr.target, int32(raw)) {
		vm.slots[instr.target] = nil
		storedInRegister = true
	} else {
		storedValue = vm.storeRawIntegerSlot(instr.target, targetKind, raw)
		vm.setI32RegisterValue(instr.target, storedValue)
	}

	if instr.target == 0 {
		if storedInRegister {
			vm.setSelfFastSlot0I32Raw(int32(raw))
		} else {
			vm.setSelfFastSlot0I32Value(storedValue)
		}
	}

	if instr.discardResult {
		vm.truncateStack(vm.stackDepth() - 1)
		vm.ip++
		return true, nil
	}

	if targetKind == runtime.IntegerI64 {
		stackIndex := vm.stackDepth() - 1
		vm.setStackValue(stackIndex, vm.stackRawI64Value(stackIndex, raw))
	} else {
		vm.setStackValue(vm.stackDepth()-1, bytecodeRawIntegerResultValue(targetKind, raw))
	}
	vm.ip++
	return true, nil
}

func bytecodeIntegerImmediateRawFast(left runtime.Value, operator string, right runtime.IntegerValue) (runtime.IntegerType, int64, bool, error) {
	rightRef := &right
	if !rightRef.IsSmallRef() {
		return "", 0, false, nil
	}
	kind, leftVal, ok := bytecodeRawIntegerValueInfo(left)
	if !ok || kind != right.TypeSuffix {
		return "", 0, false, nil
	}
	rightVal := rightRef.Int64FastRef()
	var (
		result   int64
		overflow bool
	)
	switch operator {
	case "+":
		result, overflow = addInt64Overflow(leftVal, rightVal)
	case "-":
		result, overflow = subInt64Overflow(leftVal, rightVal)
	case "*":
		result, overflow = mulInt64Overflow(leftVal, rightVal)
	case "%":
		if rightVal == 0 {
			return kind, 0, true, newDivisionByZeroError()
		}
		_, result = euclideanDivModInt64(leftVal, rightVal)
	default:
		return "", 0, false, nil
	}
	if overflow {
		return "", 0, false, nil
	}
	if err := ensureFitsInt64Type(kind, result); err != nil {
		return kind, 0, true, err
	}
	return kind, result, true, nil
}
