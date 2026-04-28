package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeIntegerValue(val runtime.Value) (runtime.IntegerValue, bool) {
	switch iv := val.(type) {
	case runtime.IntegerValue:
		return iv, true
	case *runtime.IntegerValue:
		if iv != nil {
			return *iv, true
		}
	}
	raw := unwrapScalarValue(unwrapInterfaceValue(val))
	switch iv := raw.(type) {
	case runtime.IntegerValue:
		return iv, true
	case *runtime.IntegerValue:
		if iv != nil {
			return *iv, true
		}
	}
	return runtime.IntegerValue{}, false
}

func bytecodeDirectIntegerValue(val runtime.Value) (runtime.IntegerValue, bool) {
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

func bytecodeDirectSameTypeSmallIntPair(left runtime.Value, right runtime.Value) (runtime.IntegerType, int64, int64, bool) {
	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if !lvRef.IsSmallRef() {
			return runtime.IntegerI32, 0, 0, false
		}
		switch rv := right.(type) {
		case runtime.IntegerValue:
			rvRef := &rv
			if rvRef.IsSmallRef() && lv.TypeSuffix == rv.TypeSuffix {
				return lv.TypeSuffix, lvRef.Int64FastRef(), rvRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.IsSmallRef() && lv.TypeSuffix == rv.TypeSuffix {
				return lv.TypeSuffix, lvRef.Int64FastRef(), rv.Int64FastRef(), true
			}
		}
	case *runtime.IntegerValue:
		if lv == nil || !lv.IsSmallRef() {
			return runtime.IntegerI32, 0, 0, false
		}
		switch rv := right.(type) {
		case runtime.IntegerValue:
			rvRef := &rv
			if rvRef.IsSmallRef() && lv.TypeSuffix == rv.TypeSuffix {
				return lv.TypeSuffix, lv.Int64FastRef(), rvRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.IsSmallRef() && lv.TypeSuffix == rv.TypeSuffix {
				return lv.TypeSuffix, lv.Int64FastRef(), rv.Int64FastRef(), true
			}
		}
	}
	return runtime.IntegerI32, 0, 0, false
}

func bytecodeDirectIntegerCompare(op string, left runtime.Value, right runtime.Value) (runtime.BoolValue, bool) {
	compare := func(l int64, r int64) (runtime.BoolValue, bool) {
		switch op {
		case "<":
			return runtime.BoolValue{Val: l < r}, true
		case "<=":
			return runtime.BoolValue{Val: l <= r}, true
		case ">":
			return runtime.BoolValue{Val: l > r}, true
		case ">=":
			return runtime.BoolValue{Val: l >= r}, true
		case "==":
			return runtime.BoolValue{Val: l == r}, true
		case "!=":
			return runtime.BoolValue{Val: l != r}, true
		default:
			return runtime.BoolValue{}, false
		}
	}

	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if !lvRef.IsSmallRef() {
			return runtime.BoolValue{}, false
		}
		switch rv := right.(type) {
		case runtime.IntegerValue:
			rvRef := &rv
			if rvRef.IsSmallRef() {
				return compare(lvRef.Int64FastRef(), rvRef.Int64FastRef())
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.IsSmallRef() {
				return compare(lvRef.Int64FastRef(), rv.Int64FastRef())
			}
		}
	case *runtime.IntegerValue:
		if lv == nil || !lv.IsSmallRef() {
			return runtime.BoolValue{}, false
		}
		switch rv := right.(type) {
		case runtime.IntegerValue:
			rvRef := &rv
			if rvRef.IsSmallRef() {
				return compare(lv.Int64FastRef(), rvRef.Int64FastRef())
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.IsSmallRef() {
				return compare(lv.Int64FastRef(), rv.Int64FastRef())
			}
		}
	}
	return runtime.BoolValue{}, false
}

func bytecodeDirectIntegerLessEqualImmediate(left runtime.Value, right runtime.IntegerValue) (bool, bool) {
	rightRef := &right
	if !rightRef.IsSmallRef() {
		return false, false
	}
	rightVal := rightRef.Int64FastRef()
	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if !lvRef.IsSmallRef() {
			return false, false
		}
		return lvRef.Int64FastRef() <= rightVal, true
	case *runtime.IntegerValue:
		if lv == nil || !lv.IsSmallRef() {
			return false, false
		}
		return lv.Int64FastRef() <= rightVal, true
	}
	return false, false
}

func bytecodeDirectIntegerLessEqualImmediateRaw(left runtime.Value, rightVal int64) (bool, bool) {
	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if !lvRef.IsSmallRef() {
			return false, false
		}
		return lvRef.Int64FastRef() <= rightVal, true
	case *runtime.IntegerValue:
		if lv == nil || !lv.IsSmallRef() {
			return false, false
		}
		return lv.Int64FastRef() <= rightVal, true
	}
	return false, false
}

func execBinaryDirectIntegerComparisonFast(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool) {
	switch op {
	case "<", "<=", ">", ">=", "==", "!=":
	default:
		return nil, false
	}
	if cmp, ok := bytecodeDirectIntegerCompare(op, left, right); ok {
		return cmp, true
	}
	leftInt, ok := bytecodeDirectIntegerValue(left)
	if !ok {
		return nil, false
	}
	rightInt, ok := bytecodeDirectIntegerValue(right)
	if !ok {
		return nil, false
	}
	return runtime.BoolValue{Val: integerComparisonResult(op, leftInt, rightInt)}, true
}

func (vm *bytecodeVM) execBinarySpecializedOpcode(instr *bytecodeInstruction, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	switch instr.op {
	case bytecodeOpBinaryIntAdd, bytecodeOpReturnBinaryIntAdd, bytecodeOpReturnBinaryIntAddI32:
		if fast, handled, err := bytecodeAddSmallI32PairFast(left, right); handled {
			return fast, true, err
		}
		if kind, l, r, ok := bytecodeDirectSameTypeSmallIntPair(left, right); ok {
			sum, overflow := addInt64Overflow(l, r)
			if !overflow {
				if err := ensureFitsInt64Type(kind, sum); err != nil {
					return nil, true, err
				}
				return boxedOrSmallIntegerValue(kind, sum), true, nil
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
		val, err := applyBinaryOperator(vm.interp, "+", left, right)
		return val, true, err
	case bytecodeOpBinaryIntSub:
		if fast, handled, err := bytecodeSubtractSmallI32PairFast(left, right); handled {
			return fast, true, err
		}
		if kind, l, r, ok := bytecodeDirectSameTypeSmallIntPair(left, right); ok {
			diff, overflow := subInt64Overflow(l, r)
			if !overflow {
				if err := ensureFitsInt64Type(kind, diff); err != nil {
					return nil, true, err
				}
				return boxedOrSmallIntegerValue(kind, diff), true, nil
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
		return boxedOrSmallIntegerValue(targetKind, quotient), true, nil
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

func (vm *bytecodeVM) execBinarySlotConst(instr *bytecodeInstruction, right runtime.IntegerValue, hasImmediate bool) (runtime.Value, bool, error) {
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntLessEqualSlotConst:
	default:
		return nil, false, nil
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, true, fmt.Errorf("bytecode slot out of range")
	}
	if !hasImmediate {
		return nil, true, fmt.Errorf("bytecode slot-const binary missing integer immediate")
	}
	left := vm.slots[instr.target]
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst:
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
		val, err := applyBinaryOperator(vm.interp, "+", left, right)
		return val, true, err
	case bytecodeOpBinaryIntSubSlotConst:
		if fast, handled, err := bytecodeSubtractIntegerImmediateFast(left, right); handled {
			return fast, true, err
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
		val, err := applyBinaryOperator(vm.interp, "-", left, right)
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
		val, err := applyBinaryOperator(vm.interp, "<=", left, right)
		return val, true, err
	default:
		return nil, false, nil
	}
}

func isBytecodeBinaryFastPathCandidate(op string) bool {
	normalized, _ := normalizeOperator(op)
	switch normalized {
	case "+", "-", "<", "<=", ">", ">=", "==", "!=":
		return true
	default:
		return false
	}
}

func (vm *bytecodeVM) execBinary(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (bool, error) {
	switch instr.op {
	case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntLessEqualSlotConst:
		rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
		if !hasImmediate {
			rightImmediate, hasImmediate = bytecodeImmediateIntegerValue(instr.value)
		}
		if !hasImmediate {
			rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
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
			vm.stack = append(vm.stack, bytecodeStackResultValue(fast))
			vm.ip++
			return false, nil
		}
	case bytecodeOpBinaryIntAdd,
		bytecodeOpBinaryIntSub,
		bytecodeOpBinaryIntLessEqual,
		bytecodeOpBinaryIntDivCast:
		if len(vm.stack) < 2 {
			return false, fmt.Errorf("bytecode stack underflow")
		}
		rightIdx := len(vm.stack) - 1
		right := vm.stack[rightIdx]
		leftIdx := rightIdx - 1
		left := vm.stack[leftIdx]
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
	if len(vm.stack) < 2 {
		return false, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := len(vm.stack) - 1
	right := vm.stack[rightIdx]
	leftIdx := rightIdx - 1
	left := vm.stack[leftIdx]
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

	left := vm.slots[slot]
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

func (vm *bytecodeVM) execReturnIfIntLessEqualSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (runtime.Value, bool, error) {
	conditionSlot := instr.argCount
	if conditionSlot < 0 || conditionSlot >= len(vm.slots) {
		return nil, false, fmt.Errorf("bytecode slot out of range")
	}
	if instr.target == conditionSlot && instr.hasIntImmediate {
		left := vm.slots[conditionSlot]
		right := instr.intImmediate
		rightRef := &right
		if rightRef.IsSmallRef() {
			rightVal := rightRef.Int64FastRef()
			switch lv := left.(type) {
			case runtime.IntegerValue:
				lvRef := &lv
				if lvRef.IsSmallRef() {
					if lvRef.Int64FastRef() <= rightVal {
						return left, true, nil
					}
					vm.ip++
					return nil, false, nil
				}
			case *runtime.IntegerValue:
				if lv != nil && lv.IsSmallRef() {
					if lv.Int64FastRef() <= rightVal {
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

	left := vm.slots[conditionSlot]
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
	return vm.slots[returnSlot], true, nil
}

func (vm *bytecodeVM) execReturnConstIfIntLessEqualSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable) (runtime.Value, bool, error) {
	conditionSlot := instr.argCount
	if conditionSlot < 0 || conditionSlot >= len(vm.slots) {
		return nil, false, fmt.Errorf("bytecode slot out of range")
	}
	rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
	}
	if !hasImmediate {
		return nil, false, fmt.Errorf("bytecode slot-const conditional missing integer immediate")
	}
	left := vm.slots[conditionSlot]
	if instr.hasIntRaw {
		if cmp, ok := bytecodeDirectIntegerLessEqualImmediateRaw(left, instr.intImmediateRaw); ok {
			if cmp {
				return instr.value, true, nil
			}
			vm.ip++
			return nil, false, nil
		}
	}
	if cmp, ok := bytecodeDirectIntegerLessEqualImmediate(left, rightImmediate); ok {
		if cmp {
			return instr.value, true, nil
		}
		vm.ip++
		return nil, false, nil
	}
	if leftInt, ok := bytecodeIntegerValue(left); ok {
		if integerComparisonResult("<=", leftInt, rightImmediate) {
			return instr.value, true, nil
		}
		vm.ip++
		return nil, false, nil
	}
	result, err := applyBinaryOperator(vm.interp, "<=", left, rightImmediate)
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
