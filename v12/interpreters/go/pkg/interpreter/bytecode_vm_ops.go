package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execBinarySpecializedOpcode(instr *bytecodeInstruction, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	switch instr.op {
	case bytecodeOpBinaryIntAdd, bytecodeOpReturnBinaryIntAdd, bytecodeOpReturnBinaryIntAddI32:
		if fast, handled, err := bytecodeAddSmallI32PairFast(left, right); handled {
			return fast, true, err
		}
		if fast, handled := bytecodeDirectFloatArithmeticFast("+", left, right); handled {
			return fast, true, nil
		}
		if kind, l, r, ok := bytecodeDirectSameTypeSmallIntPair(left, right); ok {
			sum, overflow := addInt64Overflow(l, r)
			if !overflow {
				if err := ensureFitsInt64Type(kind, sum); err != nil {
					return nil, true, err
				}
				return bytecodeRawIntegerResultValue(kind, sum), true, nil
			}
		}
		if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
			if rightInt, ok := bytecodeDirectIntegerValue(right); ok {
				if fast, handled, err := addIntegerSameTypeFast(leftInt, rightInt); handled {
					return fast, true, err
				}
				val, err := evaluateIntegerArithmeticFast("+", leftInt, rightInt)
				return val, true, err
			}
		} else if leftInt, ok := bytecodeIntegerValue(left); ok {
			if rightInt, ok := bytecodeIntegerValue(right); ok {
				if fast, handled, err := addIntegerSameTypeFast(leftInt, rightInt); handled {
					return fast, true, err
				}
				val, err := evaluateIntegerArithmeticFast("+", leftInt, rightInt)
				return val, true, err
			}
		}
		left = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left)
		right = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, right)
		val, err := applyBinaryOperator(vm.interp, "+", left, right)
		return val, true, err
	case bytecodeOpBinaryIntSub:
		if fast, handled, err := bytecodeSubtractSmallI32PairFast(left, right); handled {
			return fast, true, err
		}
		if fast, handled := bytecodeDirectFloatArithmeticFast("-", left, right); handled {
			return fast, true, nil
		}
		if kind, l, r, ok := bytecodeDirectSameTypeSmallIntPair(left, right); ok {
			diff, overflow := subInt64Overflow(l, r)
			if !overflow {
				if err := ensureFitsInt64Type(kind, diff); err != nil {
					return nil, true, err
				}
				return bytecodeRawIntegerResultValue(kind, diff), true, nil
			}
		}
		if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
			if rightInt, ok := bytecodeDirectIntegerValue(right); ok {
				if fast, handled, err := subtractIntegerSameTypeFast(leftInt, rightInt); handled {
					return fast, true, err
				}
				val, err := evaluateIntegerArithmeticFast("-", leftInt, rightInt)
				return val, true, err
			}
		} else if leftInt, ok := bytecodeIntegerValue(left); ok {
			if rightInt, ok := bytecodeIntegerValue(right); ok {
				if fast, handled, err := subtractIntegerSameTypeFast(leftInt, rightInt); handled {
					return fast, true, err
				}
				val, err := evaluateIntegerArithmeticFast("-", leftInt, rightInt)
				return val, true, err
			}
		}
		left = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left)
		right = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, right)
		val, err := applyBinaryOperator(vm.interp, "-", left, right)
		return val, true, err
	case bytecodeOpBinaryIntLessEqual:
		if cmp, ok := bytecodeDirectIntegerCompare("<=", left, right); ok {
			return cmp, true, nil
		}
		if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
			if rightInt, ok := bytecodeDirectIntegerValue(right); ok {
				return runtime.BoolValue{Val: leftInt.BigInt().Cmp(rightInt.BigInt()) <= 0}, true, nil
			}
		} else if leftInt, ok := bytecodeIntegerValue(left); ok {
			if rightInt, ok := bytecodeIntegerValue(right); ok {
				if leftInt.IsSmall() && rightInt.IsSmall() {
					return runtime.BoolValue{Val: leftInt.Int64Fast() <= rightInt.Int64Fast()}, true, nil
				}
				return runtime.BoolValue{Val: leftInt.BigInt().Cmp(rightInt.BigInt()) <= 0}, true, nil
			}
		}
		left = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left)
		right = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, right)
		val, err := applyBinaryOperator(vm.interp, "<=", left, right)
		return val, true, err
	case bytecodeOpBinaryIntDivCast:
		targetKind := runtime.IntegerType(instr.operator)
		if _, ok := lookupIntegerInfo(targetKind); !ok {
			return nil, true, fmt.Errorf("bytecode integer-division cast missing integer target type")
		}
		if fast, ok, err := execBinaryIntDivCastFastPath(targetKind, left, right); ok {
			return fast, true, err
		}
		castTarget := ast.TypeExpression(ast.Ty(string(targetKind)))
		if castExpr, ok := instr.node.(*ast.TypeCastExpression); ok && castExpr != nil && castExpr.TargetType != nil {
			castTarget = castExpr.TargetType
		}
		castTarget = vm.canonicalRuntimeTypeExpression(castTarget)
		divResult, err := applyBinaryOperator(vm.interp, "/", left, right)
		if err != nil {
			return nil, true, err
		}
		casted, err := vm.interp.castValueToType(castTarget, divResult)
		return casted, true, err
	default:
		return nil, false, nil
	}
}

const bytecodeIntDivCastFastAbsMax = 2147483647

func execBinaryIntDivCastFastPath(targetKind runtime.IntegerType, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	try := func(leftInt runtime.IntegerValue, rightInt runtime.IntegerValue) (runtime.Value, bool, error) {
		l, lok := leftInt.ToInt64()
		r, rok := rightInt.ToInt64()
		if !lok || !rok {
			return nil, false, nil
		}
		// Keep this fast path in a value range where float / cast and integer
		// truncation are equivalent; fall back outside that range.
		if l < -bytecodeIntDivCastFastAbsMax || l > bytecodeIntDivCastFastAbsMax {
			return nil, false, nil
		}
		if r < -bytecodeIntDivCastFastAbsMax || r > bytecodeIntDivCastFastAbsMax {
			return nil, false, nil
		}
		if r == 0 {
			return nil, true, newDivisionByZeroError()
		}
		var quotient int64
		if r == 2 && l >= 0 {
			quotient = l >> 1
		} else {
			quotient = int64(float64(l) / float64(r))
		}
		if err := ensureFitsInt64Type(targetKind, quotient); err != nil {
			return nil, true, err
		}
		return bytecodeRawIntegerResultValue(targetKind, quotient), true, nil
	}
	if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
		if rightInt, ok := bytecodeDirectIntegerValue(right); ok {
			return try(leftInt, rightInt)
		}
	} else if leftInt, ok := bytecodeIntegerValue(left); ok {
		if rightInt, ok := bytecodeIntegerValue(right); ok {
			return try(leftInt, rightInt)
		}
	}
	return nil, false, nil
}

func bytecodeImmediateIntegerValue(val runtime.Value) (runtime.IntegerValue, bool) {
	switch iv := val.(type) {
	case runtime.IntegerValue:
		return iv, true
	case *runtime.IntegerValue:
		if iv != nil {
			return *iv, true
		}
	}
	return runtime.IntegerValue{}, false
}

func bytecodeMultiplyIntegerImmediateFast(left runtime.Value, right runtime.IntegerValue) (runtime.Value, bool, error) {
	kind, raw, handled, err := bytecodeIntegerImmediateRawFast(left, "*", right)
	if !handled || err != nil {
		return nil, handled, err
	}
	return bytecodeRawIntegerResultValue(kind, raw), true, nil
}

func (vm *bytecodeVM) execBinaryIntSlotConstRawFast(instr *bytecodeInstruction, right runtime.IntegerValue, hasImmediate bool) (bool, error) {
	if vm == nil || instr == nil || !hasImmediate {
		return false, nil
	}
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst,
		bytecodeOpBinaryIntSubSlotConst,
		bytecodeOpBinaryIntMulSlotConst,
		bytecodeOpBinaryIntModSlotConst:
	default:
		return false, nil
	}
	left := vm.slotStoredValue(instr.target)
	kind, raw, handled, err := bytecodeIntegerImmediateRawFast(left, instr.operator, right)
	if !handled {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	vm.appendRawIntegerStack(kind, raw)
	vm.ip++
	return true, nil
}

func (vm *bytecodeVM) execBinarySlotConst(instr *bytecodeInstruction, right runtime.IntegerValue, hasImmediate bool) (runtime.Value, bool, error) {
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntMulSlotConst, bytecodeOpBinaryIntModSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpBinaryIntCompareSlotConst:
	default:
		return nil, false, nil
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, true, fmt.Errorf("bytecode slot out of range")
	}
	if !hasImmediate {
		return nil, true, fmt.Errorf("bytecode slot-const binary missing integer immediate")
	}
	left := vm.slotStoredValue(instr.target)
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst:
		if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
			if fast, handled, err := addIntegerSameTypeFast(leftInt, right); handled {
				return fast, true, err
			}
		}
		switch lv := left.(type) {
		case runtime.IntegerValue:
			if fast, handled, err := addIntegerSameTypeFast(lv, right); handled {
				return fast, true, err
			}
			val, err := evaluateIntegerArithmeticFast("+", lv, right)
			return val, true, err
		case *runtime.IntegerValue:
			if lv != nil {
				if fast, handled, err := addIntegerSameTypeFast(*lv, right); handled {
					return fast, true, err
				}
				val, err := evaluateIntegerArithmeticFast("+", *lv, right)
				return val, true, err
			}
		}
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			val, err := evaluateIntegerArithmeticFast("+", leftInt, right)
			return val, true, err
		}
		val, err := applyBinaryOperator(vm.interp, "+", vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left), right)
		return val, true, err
	case bytecodeOpBinaryIntSubSlotConst:
		if fast, handled, err := bytecodeSubtractIntegerImmediateFast(left, right); handled {
			return fast, true, err
		}
		if leftInt, ok := bytecodeDirectIntegerValue(left); ok {
			if fast, handled, err := subtractIntegerSameTypeFast(leftInt, right); handled {
				return fast, true, err
			}
		}
		switch lv := left.(type) {
		case runtime.IntegerValue:
			if fast, handled, err := subtractIntegerSameTypeFast(lv, right); handled {
				return fast, true, err
			}
			val, err := evaluateIntegerArithmeticFast("-", lv, right)
			return val, true, err
		case *runtime.IntegerValue:
			if lv != nil {
				if fast, handled, err := subtractIntegerSameTypeFast(*lv, right); handled {
					return fast, true, err
				}
				val, err := evaluateIntegerArithmeticFast("-", *lv, right)
				return val, true, err
			}
		}
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			val, err := evaluateIntegerArithmeticFast("-", leftInt, right)
			return val, true, err
		}
		val, err := applyBinaryOperator(vm.interp, "-", vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left), right)
		return val, true, err
	case bytecodeOpBinaryIntMulSlotConst:
		if fast, handled, err := bytecodeMultiplyIntegerImmediateFast(left, right); handled {
			return fast, true, err
		}
		switch lv := left.(type) {
		case runtime.IntegerValue:
			val, err := evaluateIntegerArithmeticFast("*", lv, right)
			return val, true, err
		case *runtime.IntegerValue:
			if lv != nil {
				val, err := evaluateIntegerArithmeticFast("*", *lv, right)
				return val, true, err
			}
		}
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			val, err := evaluateIntegerArithmeticFast("*", leftInt, right)
			return val, true, err
		}
		val, err := applyBinaryOperator(vm.interp, "*", vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left), right)
		return val, true, err
	case bytecodeOpBinaryIntModSlotConst:
		rightInt, ok := bytecodeIntegerValue(right)
		if ok {
			if rightInt.IsSmall() && rightInt.Int64Fast() == 0 {
				return nil, true, newDivisionByZeroError()
			}
			switch lv := left.(type) {
			case runtime.IntegerValue:
				if lv.IsSmall() && rightInt.IsSmall() {
					_, rem := euclideanDivModInt64(lv.Int64Fast(), rightInt.Int64Fast())
					return bytecodeRawIntegerResultValue(lv.TypeSuffix, rem), true, nil
				}
			case *runtime.IntegerValue:
				if lv != nil && lv.IsSmallRef() && rightInt.IsSmall() {
					_, rem := euclideanDivModInt64(lv.Int64FastRef(), rightInt.Int64Fast())
					return bytecodeRawIntegerResultValue(lv.TypeSuffix, rem), true, nil
				}
			}
			if leftInt, ok := bytecodeIntegerValue(left); ok {
				if leftInt.IsSmall() && rightInt.IsSmall() {
					_, rem := euclideanDivModInt64(leftInt.Int64Fast(), rightInt.Int64Fast())
					return bytecodeRawIntegerResultValue(leftInt.TypeSuffix, rem), true, nil
				}
				val, err := evaluateDivMod(vm.interp, "%", leftInt, right)
				return val, true, err
			}
		}
		val, err := evaluateDivMod(vm.interp, "%", vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left), right)
		return val, true, err
	case bytecodeOpBinaryIntLessEqualSlotConst:
		rightRef := &right
		switch lv := left.(type) {
		case runtime.IntegerValue:
			lvRef := &lv
			if lvRef.IsSmallRef() && rightRef.IsSmallRef() {
				return runtime.BoolValue{Val: lvRef.Int64FastRef() <= rightRef.Int64FastRef()}, true, nil
			}
			return runtime.BoolValue{Val: lv.BigInt().Cmp(right.BigInt()) <= 0}, true, nil
		case *runtime.IntegerValue:
			if lv != nil {
				if lv.IsSmallRef() && rightRef.IsSmallRef() {
					return runtime.BoolValue{Val: lv.Int64FastRef() <= rightRef.Int64FastRef()}, true, nil
				}
				return runtime.BoolValue{Val: lv.BigInt().Cmp(right.BigInt()) <= 0}, true, nil
			}
		}
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			leftIntRef := &leftInt
			if leftIntRef.IsSmallRef() && rightRef.IsSmallRef() {
				return runtime.BoolValue{Val: leftIntRef.Int64FastRef() <= rightRef.Int64FastRef()}, true, nil
			}
			return runtime.BoolValue{Val: leftInt.BigInt().Cmp(right.BigInt()) <= 0}, true, nil
		}
		val, err := applyBinaryOperator(vm.interp, "<=", vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left), right)
		return val, true, err
	case bytecodeOpBinaryIntCompareSlotConst:
		if instr.hasIntRaw {
			if cmp, ok := bytecodeDirectIntegerCompareImmediateRaw(instr.operator, left, instr.intImmediateRaw); ok {
				return runtime.BoolValue{Val: cmp}, true, nil
			}
		}
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			return runtime.BoolValue{Val: integerComparisonResult(instr.operator, leftInt, right)}, true, nil
		}
		val, err := applyBinaryOperator(vm.interp, instr.operator, vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonDynamicOperation, left), right)
		return val, true, err
	default:
		return nil, false, nil
	}
}

func isBytecodeBinaryFastPathCandidate(op string) bool {
	normalized, _ := normalizeOperator(op)
	switch normalized {
	case "+", "-", "*", "^", "/", "//", "%", "<", "<=", ">", ">=", "==", "!=", "&", "|", "<<", ">>":
		return true
	default:
		return false
	}
}

func (vm *bytecodeVM) execBinary(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (bool, error) {
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntMulSlotConst, bytecodeOpBinaryIntModSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpBinaryIntCompareSlotConst:
		rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
		if !hasImmediate {
			rightImmediate, hasImmediate = bytecodeImmediateIntegerValue(instr.value)
		}
		if !hasImmediate {
			rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
		}
		if handled, err := vm.execBinaryIntSlotConstRawFast(instr, rightImmediate, hasImmediate); handled {
			if err != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						return true, nil
					}
				}
				return false, err
			}
			return false, nil
		}
		if fast, handled, err := vm.execBinarySlotConst(instr, rightImmediate, hasImmediate); handled {
			if err != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						return true, nil
					}
				}
				return false, err
			}
			vm.appendStackValue(bytecodeStackResultValue(fast))
			vm.ip++
			return false, nil
		}
	case bytecodeOpBinaryCastSlotFloatConstDiv:
		if fast, handled, err := vm.execBinaryCastSlotFloatConstDiv(instr); handled {
			if err != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						return true, nil
					}
				}
				return false, err
			}
			vm.appendStackValue(bytecodeStackResultValue(fast))
			vm.ip++
			return false, nil
		}
	case bytecodeOpBinaryFloatMulSlotConst:
		if fast, handled, err := vm.execBinaryFloatMulSlotConst(instr); handled {
			if err != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						return true, nil
					}
				}
				return false, err
			}
			vm.appendStackValue(bytecodeStackResultValue(fast))
			vm.ip++
			return false, nil
		}
	case bytecodeOpBinaryIntAdd,
		bytecodeOpBinaryIntSub,
		bytecodeOpBinaryIntLessEqual,
		bytecodeOpBinaryIntDivCast:
		if vm.stackDepth() < 2 {
			return false, fmt.Errorf("bytecode stack underflow")
		}
		rightIdx := vm.stackDepth() - 1
		leftIdx := rightIdx - 1
		right := vm.stackValue(rightIdx)
		left := vm.stackValue(leftIdx)
		if instr.op == bytecodeOpBinaryIntAdd || instr.op == bytecodeOpBinaryIntSub {
			op := "+"
			if instr.op == bytecodeOpBinaryIntSub {
				op = "-"
			}
			if raw, kind, handled := bytecodeDirectFloatArithmeticRawValue(op, left, right); handled {
				vm.replaceTop2RawFloatUnchecked(raw, kind)
				vm.ip++
				return false, nil
			}
			if kind, l, r, ok := bytecodeDirectSameTypeSmallIntPair(left, right); ok {
				var (
					result   int64
					overflow bool
				)
				if instr.op == bytecodeOpBinaryIntAdd {
					result, overflow = addInt64Overflow(l, r)
				} else {
					result, overflow = subInt64Overflow(l, r)
				}
				if !overflow {
					if err := ensureFitsInt64Type(kind, result); err != nil {
						return false, err
					}
					vm.replaceTop2RawIntegerUnchecked(kind, result)
					vm.ip++
					return false, nil
				}
			}
		}
		fast, _, err := vm.execBinarySpecializedOpcode(instr, left, right)
		if err != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
			if instr.node != nil {
				err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
				if vm.handleLoopSignal(err) {
					return true, nil
				}
			}
			return false, err
		}
		vm.replaceTop2Unchecked(fast)
		vm.ip++
		return false, nil
	}
	if vm.stackDepth() < 2 {
		return false, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := vm.stackDepth() - 1
	leftIdx := rightIdx - 1
	right := vm.stackValue(rightIdx)
	left := vm.stackValue(leftIdx)
	if instr.operator == "+" {
		rawLeft := unwrapInterfaceValue(left)
		rawRight := unwrapInterfaceValue(right)
		if ls, ok := rawLeft.(runtime.StringValue); ok {
			rs, ok := rawRight.(runtime.StringValue)
			if !ok {
				err := fmt.Errorf("Arithmetic requires numeric operands")
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
				}
				return false, err
			}
			vm.replaceTop2Unchecked(runtime.StringValue{Val: ls.Val + rs.Val})
			vm.ip++
			return false, nil
		}
		if _, ok := rawRight.(runtime.StringValue); ok {
			err := fmt.Errorf("Arithmetic requires numeric operands")
			if instr.node != nil {
				err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return false, err
		}
	}
	if fast, handled := execBinaryDirectIntegerComparisonFast(instr.operator, left, right); handled {
		vm.replaceTop2Unchecked(fast)
		vm.ip++
		return false, nil
	}
	if fast, handled := bytecodeDirectFloatCompareFast(instr.operator, left, right); handled {
		vm.replaceTop2Unchecked(fast)
		vm.ip++
		return false, nil
	}
	if raw, kind, handled := bytecodeDirectFloatArithmeticRawValue(instr.operator, left, right); handled {
		vm.replaceTop2RawFloatUnchecked(raw, kind)
		vm.ip++
		return false, nil
	}
	if instr.bitwiseRawCandidate {
		if kind, raw, handled, err := bytecodeDirectSameTypeRawIntegerBitwise(instr.operator, left, right); handled {
			if err != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						return true, nil
					}
				}
				return false, err
			}
			vm.replaceTop2RawIntegerUnchecked(kind, raw)
			vm.ip++
			return false, nil
		}
	}
	if isBytecodeBinaryFastPathCandidate(instr.operator) {
		if fast, handled, err := ApplyBinaryOperatorFast(instr.operator, left, right); handled {
			if err != nil {
				err = vm.interp.wrapStandardRuntimeError(err)
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						return true, nil
					}
				}
				return false, err
			}
			vm.replaceTop2Unchecked(fast)
			vm.ip++
			return false, nil
		}
	}
	result, err := applyBinaryOperator(vm.interp, instr.operator, left, right)
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			if vm.handleLoopSignal(err) {
				return true, nil
			}
		}
		return false, err
	}
	vm.replaceTop2Unchecked(result)
	vm.ip++
	return false, nil
}

func (vm *bytecodeVM) execJumpIfIntLessEqualSlotConstFalse(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) error {
	slot := instr.argCount
	if slot < 0 || slot >= len(vm.slots) {
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
		return fmt.Errorf("bytecode slot-const conditional missing integer immediate")
	}
	if instr.hasIntRaw {
		if raw, ok := vm.slotDirectSmallI32ValueValidated(slot); ok {
			condValue := raw <= instr.intImmediateRaw
			if !condValue {
				vm.ip = instr.target
				return nil
			}
			vm.ip++
			return nil
		}
	}
	left := vm.slotRuntimeValue(slot)
	condKnown := false
	condValue := false
	if instr.hasIntRaw {
		if cmp, ok := bytecodeDirectIntegerLessEqualImmediateRaw(left, instr.intImmediateRaw); ok {
			condKnown = true
			condValue = cmp
		}
	}
	if !condKnown {
		if cmp, ok := bytecodeDirectIntegerLessEqualImmediate(left, rightImmediate); ok {
			condKnown = true
			condValue = cmp
		}
	}
	if !condKnown {
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			condKnown = true
			condValue = integerComparisonResult("<=", leftInt, rightImmediate)
		}
	}
	if !condKnown {
		result, err := applyBinaryOperator(vm.interp, "<=", left, rightImmediate)
		if err != nil {
			return err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		condValue = vm.interp.isTruthy(result)
	}
	if !condValue {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execJumpIfIntCompareSlotConstFalse(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) error {
	slot := instr.argCount
	if slot < 0 || slot >= len(vm.slots) {
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
		return fmt.Errorf("bytecode slot-const conditional missing integer immediate")
	}
	if instr.hasIntRaw {
		if raw, ok := vm.slotDirectSmallI32ValueValidated(slot); ok {
			if condValue, ok := bytecodeCompareInt64(instr.operator, raw, instr.intImmediateRaw); ok {
				if !condValue {
					vm.ip = instr.target
					return nil
				}
				vm.ip++
				return nil
			}
		}
	}
	left := vm.slotRuntimeValue(slot)
	condKnown := false
	condValue := false
	if instr.hasIntRaw {
		if cmp, ok := bytecodeDirectIntegerCompareImmediateRaw(instr.operator, left, instr.intImmediateRaw); ok {
			condKnown = true
			condValue = cmp
		}
	}
	if !condKnown {
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			condKnown = true
			condValue = integerComparisonResult(instr.operator, leftInt, rightImmediate)
		}
	}
	if !condKnown {
		result, err := applyBinaryOperator(vm.interp, instr.operator, left, rightImmediate)
		if err != nil {
			return err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		condValue = vm.interp.isTruthy(result)
	}
	if !condValue {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execReturnIfIntLessEqualSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (runtime.Value, bool, error) {
	conditionSlot := instr.argCount
	if conditionSlot < 0 || conditionSlot >= len(vm.slots) {
		return nil, false, fmt.Errorf("bytecode slot out of range")
	}
	operator := instr.operator
	if operator == "" {
		operator = "<="
	}
	if instr.target == conditionSlot && instr.hasIntImmediate {
		if vm.hasI32RegisterFrame() {
			if raw, ok := vm.i32RegisterRaw(conditionSlot); ok {
				right := instr.intImmediate
				rightRef := &right
				if rightRef.IsSmallRef() {
					if condValue, ok := bytecodeCompareInt64(operator, int64(raw), rightRef.Int64FastRef()); ok && condValue {
						return bytecodeRawI32ResultValue(int64(raw)), true, nil
					}
					vm.ip++
					return nil, false, nil
				}
			}
		}
		left := vm.slotRuntimeValue(conditionSlot)
		right := instr.intImmediate
		rightRef := &right
		if rightRef.IsSmallRef() {
			rightVal := rightRef.Int64FastRef()
			switch lv := left.(type) {
			case runtime.IntegerValue:
				lvRef := &lv
				if lvRef.IsSmallRef() {
					if condValue, ok := bytecodeCompareInt64(operator, lvRef.Int64FastRef(), rightVal); ok && condValue {
						return left, true, nil
					}
					vm.ip++
					return nil, false, nil
				}
			case *runtime.IntegerValue:
				if lv != nil && lv.IsSmallRef() {
					if condValue, ok := bytecodeCompareInt64(operator, lv.Int64FastRef(), rightVal); ok && condValue {
						return left, true, nil
					}
					vm.ip++
					return nil, false, nil
				}
			}
		}
	}
	returnSlot := instr.target
	if returnSlot < 0 || returnSlot >= len(vm.slots) {
		return nil, false, fmt.Errorf("bytecode return slot out of range")
	}
	rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeImmediateIntegerValue(instr.value)
	}
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
	}
	if !hasImmediate {
		return nil, false, fmt.Errorf("bytecode slot-const conditional missing integer immediate")
	}
	left := vm.slotRuntimeValue(conditionSlot)
	condKnown := false
	condValue := false
	if instr.hasIntRaw {
		if operator == "<=" {
			if cmp, ok := bytecodeDirectIntegerLessEqualImmediateRaw(left, instr.intImmediateRaw); ok {
				condKnown = true
				condValue = cmp
			}
		} else if cmp, ok := bytecodeDirectIntegerCompareImmediateRaw(operator, left, instr.intImmediateRaw); ok {
			condKnown = true
			condValue = cmp
		}
	}
	if !condKnown {
		if operator == "<=" {
			if cmp, ok := bytecodeDirectIntegerLessEqualImmediate(left, rightImmediate); ok {
				condKnown = true
				condValue = cmp
			}
		}
	}
	if !condKnown {
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			condKnown = true
			condValue = integerComparisonResult(operator, leftInt, rightImmediate)
		}
	}
	if !condKnown {
		result, err := applyBinaryOperator(vm.interp, operator, left, rightImmediate)
		if err != nil {
			return nil, false, err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		condValue = vm.interp.isTruthy(result)
	}
	if !condValue {
		vm.ip++
		return nil, false, nil
	}
	return vm.slotRuntimeValue(returnSlot), true, nil
}

func (vm *bytecodeVM) execReturnConstIfIntLessEqualSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (runtime.Value, bool, error) {
	conditionSlot := instr.argCount
	if conditionSlot < 0 || conditionSlot >= len(vm.slots) {
		return nil, false, fmt.Errorf("bytecode slot out of range")
	}
	operator := instr.operator
	if operator == "" {
		operator = "<="
	}
	rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
	}
	if !hasImmediate {
		return nil, false, fmt.Errorf("bytecode slot-const conditional missing integer immediate")
	}
	if instr.hasIntRaw && conditionSlot == 0 && vm.selfFastSlot0I32Valid {
		if condValue, ok := bytecodeCompareInt64(operator, int64(vm.selfFastSlot0I32Raw), instr.intImmediateRaw); ok && condValue {
			return instr.value, true, nil
		}
		vm.ip++
		return nil, false, nil
	}
	left := vm.slotRuntimeValue(conditionSlot)
	if instr.hasIntRaw {
		if operator == "<=" {
			if cmp, ok := bytecodeDirectIntegerLessEqualImmediateRaw(left, instr.intImmediateRaw); ok {
				if cmp {
					return instr.value, true, nil
				}
				vm.ip++
				return nil, false, nil
			}
		} else if cmp, ok := bytecodeDirectIntegerCompareImmediateRaw(operator, left, instr.intImmediateRaw); ok {
			if cmp {
				return instr.value, true, nil
			}
			vm.ip++
			return nil, false, nil
		}
	}
	if operator == "<=" {
		if cmp, ok := bytecodeDirectIntegerLessEqualImmediate(left, rightImmediate); ok {
			if cmp {
				return instr.value, true, nil
			}
			vm.ip++
			return nil, false, nil
		}
	}
	if leftInt, ok := bytecodeIntegerValue(left); ok {
		if integerComparisonResult(operator, leftInt, rightImmediate) {
			return instr.value, true, nil
		}
		vm.ip++
		return nil, false, nil
	}
	result, err := applyBinaryOperator(vm.interp, operator, left, rightImmediate)
	if err != nil {
		return nil, false, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	if !vm.interp.isTruthy(result) {
		vm.ip++
		return nil, false, nil
	}
	return instr.value, true, nil
}
