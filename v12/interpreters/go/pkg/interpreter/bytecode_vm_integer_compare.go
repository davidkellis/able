package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeIntegerValue(val runtime.Value) (runtime.IntegerValue, bool) {
	if kind, raw, ok := bytecodeRawIntegerValueInfo(val); ok {
		return runtime.NewSmallInt(raw, kind), true
	}
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
	if kind, raw, ok := bytecodeRawIntegerValueInfo(val); ok {
		return runtime.NewSmallInt(raw, kind), true
	}
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
	switch leftRaw := left.(type) {
	case bytecodeRawI32SlotValue:
		switch rightRaw := right.(type) {
		case bytecodeRawI32SlotValue:
			return runtime.IntegerI32, int64(leftRaw), int64(rightRaw), true
		case *bytecodeRawI32StackCell:
			if rightRaw != nil {
				return runtime.IntegerI32, int64(leftRaw), int64(rightRaw.Val), true
			}
		case runtime.IntegerValue:
			rightRef := &rightRaw
			if rightRaw.TypeSuffix == runtime.IntegerI32 && rightRef.IsSmallRef() {
				return runtime.IntegerI32, int64(leftRaw), rightRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rightRaw != nil && rightRaw.TypeSuffix == runtime.IntegerI32 && rightRaw.IsSmallRef() {
				return runtime.IntegerI32, int64(leftRaw), rightRaw.Int64FastRef(), true
			}
		}
	case *bytecodeRawI32StackCell:
		if leftRaw == nil {
			return runtime.IntegerI32, 0, 0, false
		}
		switch rightRaw := right.(type) {
		case bytecodeRawI32SlotValue:
			return runtime.IntegerI32, int64(leftRaw.Val), int64(rightRaw), true
		case *bytecodeRawI32StackCell:
			if rightRaw != nil {
				return runtime.IntegerI32, int64(leftRaw.Val), int64(rightRaw.Val), true
			}
		case runtime.IntegerValue:
			rightRef := &rightRaw
			if rightRaw.TypeSuffix == runtime.IntegerI32 && rightRef.IsSmallRef() {
				return runtime.IntegerI32, int64(leftRaw.Val), rightRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rightRaw != nil && rightRaw.TypeSuffix == runtime.IntegerI32 && rightRaw.IsSmallRef() {
				return runtime.IntegerI32, int64(leftRaw.Val), rightRaw.Int64FastRef(), true
			}
		}
	case bytecodeRawI64ResultValue:
		switch rightRaw := right.(type) {
		case bytecodeRawI64ResultValue:
			return runtime.IntegerI64, int64(leftRaw), int64(rightRaw), true
		case *bytecodeRawI64SlotCell:
			if rightRaw != nil {
				return runtime.IntegerI64, int64(leftRaw), rightRaw.Val, true
			}
		case runtime.IntegerValue:
			rightRef := &rightRaw
			if rightRaw.TypeSuffix == runtime.IntegerI64 && rightRef.IsSmallRef() {
				return runtime.IntegerI64, int64(leftRaw), rightRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rightRaw != nil && rightRaw.TypeSuffix == runtime.IntegerI64 && rightRaw.IsSmallRef() {
				return runtime.IntegerI64, int64(leftRaw), rightRaw.Int64FastRef(), true
			}
		}
	case *bytecodeRawI64SlotCell:
		if leftRaw == nil {
			return runtime.IntegerI32, 0, 0, false
		}
		switch rightRaw := right.(type) {
		case bytecodeRawI64ResultValue:
			return runtime.IntegerI64, leftRaw.Val, int64(rightRaw), true
		case *bytecodeRawI64SlotCell:
			if rightRaw != nil {
				return runtime.IntegerI64, leftRaw.Val, rightRaw.Val, true
			}
		case runtime.IntegerValue:
			rightRef := &rightRaw
			if rightRaw.TypeSuffix == runtime.IntegerI64 && rightRef.IsSmallRef() {
				return runtime.IntegerI64, leftRaw.Val, rightRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rightRaw != nil && rightRaw.TypeSuffix == runtime.IntegerI64 && rightRaw.IsSmallRef() {
				return runtime.IntegerI64, leftRaw.Val, rightRaw.Int64FastRef(), true
			}
		}
	}

	leftKind, leftRaw, leftOK := bytecodeRawIntegerValueInfo(left)
	if !leftOK {
		return runtime.IntegerI32, 0, 0, false
	}
	rightKind, rightRaw, rightOK := bytecodeRawIntegerValueInfo(right)
	if !rightOK || leftKind != rightKind {
		return runtime.IntegerI32, 0, 0, false
	}
	return leftKind, leftRaw, rightRaw, true
}

func bytecodeDirectSameTypeRawIntegerBitwise(op string, left runtime.Value, right runtime.Value) (runtime.IntegerType, int64, bool, error) {
	op, dotted := normalizeOperator(op)
	if !dotted {
		return runtime.IntegerI32, 0, false, nil
	}
	switch op {
	case "&", "|", "^", "<<", ">>":
	default:
		return runtime.IntegerI32, 0, false, nil
	}
	kind, leftRaw, rightRaw, ok := bytecodeDirectSameTypeSmallIntPair(left, right)
	if !ok {
		return runtime.IntegerI32, 0, false, nil
	}
	info, ok := lookupIntegerInfo(kind)
	if !ok || info.bits <= 0 || info.bits > 64 {
		return runtime.IntegerI32, 0, false, nil
	}
	leftPattern, ok := integerBitPatternUint64(leftRaw, info)
	if !ok {
		return runtime.IntegerI32, 0, false, nil
	}

	switch op {
	case "&", "|", "^":
		rightPattern, ok := integerBitPatternUint64(rightRaw, info)
		if !ok {
			return runtime.IntegerI32, 0, false, nil
		}
		var resultPattern uint64
		switch op {
		case "&":
			resultPattern = leftPattern & rightPattern
		case "|":
			resultPattern = leftPattern | rightPattern
		case "^":
			resultPattern = leftPattern ^ rightPattern
		}
		result, ok := integerFromBitPatternUint64(resultPattern, info)
		if !ok {
			return runtime.IntegerI32, 0, false, nil
		}
		return kind, result, true, nil
	case "<<", ">>":
		if rightRaw < 0 || rightRaw >= int64(info.bits) {
			return runtime.IntegerI32, 0, true, newShiftOutOfRangeError(rightRaw)
		}
		count := uint(rightRaw)
		if op == ">>" {
			if info.signed {
				return kind, leftRaw >> count, true, nil
			}
			result, ok := integerFromBitPatternUint64(leftPattern>>count, info)
			if !ok {
				return runtime.IntegerI32, 0, false, nil
			}
			return kind, result, true, nil
		}

		mask := integerMaskUint64(info.bits)
		if info.signed {
			result, ok := integerFromBitPatternUint64((leftPattern<<count)&mask, info)
			if !ok {
				return runtime.IntegerI32, 0, false, nil
			}
			if result>>count != leftRaw {
				return runtime.IntegerI32, 0, true, newOverflowError("integer overflow")
			}
			return kind, result, true, nil
		}
		if count > 0 && leftPattern > (mask>>count) {
			return runtime.IntegerI32, 0, true, newOverflowError("integer overflow")
		}
		result, ok := integerFromBitPatternUint64((leftPattern<<count)&mask, info)
		if !ok {
			return runtime.IntegerI32, 0, false, nil
		}
		return kind, result, true, nil
	}
	return runtime.IntegerI32, 0, false, nil
}

func bytecodeDottedBitwiseOperator(op string) bool {
	op, dotted := normalizeOperator(op)
	if !dotted {
		return false
	}
	switch op {
	case "&", "|", "^", "<<", ">>":
		return true
	default:
		return false
	}
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
	case bytecodeRawI32SlotValue:
		switch rv := right.(type) {
		case bytecodeRawI32SlotValue:
			return compare(int64(lv), int64(rv))
		case *bytecodeRawI32StackCell:
			if rv != nil {
				return compare(int64(lv), int64(rv.Val))
			}
		case runtime.IntegerValue:
			rvRef := &rv
			if rvRef.IsSmallRef() {
				return compare(int64(lv), rvRef.Int64FastRef())
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.IsSmallRef() {
				return compare(int64(lv), rv.Int64FastRef())
			}
		}
	case *bytecodeRawI32StackCell:
		if lv == nil {
			return runtime.BoolValue{}, false
		}
		switch rv := right.(type) {
		case bytecodeRawI32SlotValue:
			return compare(int64(lv.Val), int64(rv))
		case *bytecodeRawI32StackCell:
			if rv != nil {
				return compare(int64(lv.Val), int64(rv.Val))
			}
		case runtime.IntegerValue:
			rvRef := &rv
			if rvRef.IsSmallRef() {
				return compare(int64(lv.Val), rvRef.Int64FastRef())
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.IsSmallRef() {
				return compare(int64(lv.Val), rv.Int64FastRef())
			}
		}
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
	case bytecodeRawI32SlotValue:
		return int64(lv) <= rightVal, true
	case *bytecodeRawI32StackCell:
		if lv != nil {
			return int64(lv.Val) <= rightVal, true
		}
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
	if _, raw, ok := bytecodeRawIntegerValueInfo(left); ok {
		return raw <= rightVal, true
	}
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

func bytecodeDirectIntegerCompareImmediateRaw(op string, left runtime.Value, rightVal int64) (bool, bool) {
	if _, raw, ok := bytecodeRawIntegerValueInfo(left); ok {
		return bytecodeCompareInt64(op, raw, rightVal)
	}
	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if !lvRef.IsSmallRef() {
			return false, false
		}
		return bytecodeCompareInt64(op, lvRef.Int64FastRef(), rightVal)
	case *runtime.IntegerValue:
		if lv == nil || !lv.IsSmallRef() {
			return false, false
		}
		return bytecodeCompareInt64(op, lv.Int64FastRef(), rightVal)
	}
	return false, false
}

func bytecodeCompareInt64(op string, left int64, right int64) (bool, bool) {
	switch op {
	case "<":
		return left < right, true
	case "<=":
		return left <= right, true
	case ">":
		return left > right, true
	case ">=":
		return left >= right, true
	case "==":
		return left == right, true
	case "!=":
		return left != right, true
	default:
		return false, false
	}
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
	switch left.(type) {
	case runtime.BoolValue,
		runtime.CharValue,
		runtime.StringValue,
		runtime.FloatValue,
		bytecodeRawF32SlotValue,
		bytecodeRawF64SlotValue:
		return nil, false
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
