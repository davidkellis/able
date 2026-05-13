package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeDirectFloatValue(val runtime.Value) (float64, runtime.FloatType, bool) {
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
	leftVal, leftKind, ok := bytecodeDirectFloatValue(left)
	if !ok {
		return nil, false
	}
	rightVal, rightKind, ok := bytecodeDirectFloatValue(right)
	if !ok {
		return nil, false
	}
	targetKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		targetKind = runtime.FloatF64
	}
	var result float64
	switch op {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	default:
		return nil, false
	}
	return runtime.FloatValue{Val: normalizeFloat(targetKind, result), TypeSuffix: targetKind}, true
}
