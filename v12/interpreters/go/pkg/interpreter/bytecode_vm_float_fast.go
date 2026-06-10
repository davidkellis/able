package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeDirectFloatValue(val runtime.Value) (float64, runtime.FloatType, bool) {
	if raw, kind, ok := bytecodeDirectRawFloatValue(val); ok {
		return raw, kind, true
	}
	switch fv := val.(type) {
	case runtime.FloatValue:
		return fv.Val, fv.TypeSuffix, true
	case *runtime.FloatValue:
		if fv != nil {
			return fv.Val, fv.TypeSuffix, true
		}
	}
	return 0, runtime.FloatF64, false
}

func bytecodeDirectFloatArithmeticFast(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool) {
	result, kind, ok := bytecodeDirectFloatArithmeticRawValue(op, left, right)
	if !ok {
		return nil, false
	}
	return bytecodeRawFloatSlotValue(result, kind), true
}

func bytecodeDirectFloatArithmeticRawValue(op string, left runtime.Value, right runtime.Value) (float64, runtime.FloatType, bool) {
	leftVal, leftKind, ok := bytecodeDirectFloatValue(left)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	rightVal, rightKind, ok := bytecodeDirectFloatValue(right)
	if !ok {
		return 0, runtime.FloatF64, false
	}
	return bytecodeDirectFloatArithmeticRawFast(op, leftVal, leftKind, rightVal, rightKind)
}

func bytecodeDirectFloatCompareFast(op string, left runtime.Value, right runtime.Value) (runtime.BoolValue, bool) {
	leftVal, leftKind, ok := bytecodeDirectFloatValue(left)
	if !ok {
		return runtime.BoolValue{}, false
	}
	rightVal, rightKind, ok := bytecodeDirectFloatValue(right)
	if !ok {
		return runtime.BoolValue{}, false
	}
	return bytecodeDirectFloatCompareRawFast(op, leftVal, leftKind, rightVal, rightKind)
}

func bytecodeDirectFloatCompareRawFast(op string, leftVal float64, leftKind runtime.FloatType, rightVal float64, rightKind runtime.FloatType) (runtime.BoolValue, bool) {
	if leftKind == runtime.FloatF32 {
		leftVal = normalizeFloat(runtime.FloatF32, leftVal)
	}
	if rightKind == runtime.FloatF32 {
		rightVal = normalizeFloat(runtime.FloatF32, rightVal)
	}
	cond, ok := bytecodeCompareFloat64(op, leftVal, rightVal)
	if !ok {
		return runtime.BoolValue{}, false
	}
	return runtime.BoolValue{Val: cond}, true
}

func bytecodeDirectFloatArithmeticRawFast(op string, leftVal float64, leftKind runtime.FloatType, rightVal float64, rightKind runtime.FloatType) (float64, runtime.FloatType, bool) {
	targetKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		targetKind = runtime.FloatF64
	}
	result := 0.0
	switch op {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	default:
		return 0, runtime.FloatF64, false
	}
	return normalizeFloat(targetKind, result), targetKind, true
}

func bytecodeDirectFloatDivisionRawFast(leftVal float64, leftKind runtime.FloatType, rightVal float64, rightKind runtime.FloatType) (float64, runtime.FloatType, bool, error) {
	targetKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		targetKind = runtime.FloatF64
	}
	return normalizeFloat(targetKind, leftVal/rightVal), targetKind, true, nil
}
