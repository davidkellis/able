package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

const numericEpsilon = 1e-9

func binaryOpForAssignment(op ast.AssignmentOperator) (string, bool) {
	switch op {
	case ast.AssignmentAdd:
		return "+", true
	case ast.AssignmentSub:
		return "-", true
	case ast.AssignmentMul:
		return "*", true
	case ast.AssignmentDiv:
		return "/", true
	case ast.AssignmentMod:
		return "%", true
	case ast.AssignmentBitAnd:
		return "&", true
	case ast.AssignmentBitOr:
		return "|", true
	case ast.AssignmentBitXor:
		return "^", true
	case ast.AssignmentShiftL:
		return "<<", true
	case ast.AssignmentShiftR:
		return ">>", true
	default:
		return "", false
	}
}

func applyBinaryOperator(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	switch op {
	case "+", "-", "*", "/":
		return evaluateArithmetic(op, left, right)
	case "%":
		return evaluateModulo(left, right)
	case "<", "<=", ">", ">=":
		return evaluateComparison(op, left, right)
	case "==":
		return runtime.BoolValue{Val: valuesEqual(left, right)}, nil
	case "!=":
		return runtime.BoolValue{Val: !valuesEqual(left, right)}, nil
	case "&", "|", "^", "<<", ">>":
		return evaluateBitwise(op, left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operator %s", op)
	}
}

func evaluateModulo(left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if lv, ok := left.(runtime.IntegerValue); ok {
		if rv, ok := right.(runtime.IntegerValue); ok {
			if rv.Val == nil || rv.Val.Sign() == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			targetType, err := promoteIntegerTypes(lv.TypeSuffix, rv.TypeSuffix)
			if err != nil {
				return nil, err
			}
			info, err := getIntegerInfo(targetType)
			if err != nil {
				return nil, err
			}
			result := new(big.Int).Rem(runtime.CloneBigInt(lv.Val), rv.Val)
			if err := ensureFitsInteger(info, result); err != nil {
				return nil, err
			}
			return runtime.IntegerValue{Val: result, TypeSuffix: targetType}, nil
		}
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	if rightFloat == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	targetFloatKind := floatResultKind(left, right)
	return runtime.FloatValue{Val: normalizeFloat(targetFloatKind, math.Mod(leftFloat, rightFloat)), TypeSuffix: targetFloatKind}, nil
}

func evaluateBitwise(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	lv, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires integer operands")
	}
	rv, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires integer operands")
	}
	targetType, err := promoteIntegerTypes(lv.TypeSuffix, rv.TypeSuffix)
	if err != nil {
		return nil, err
	}
	info, err := getIntegerInfo(targetType)
	if err != nil {
		return nil, err
	}
	lVal := runtime.CloneBigInt(lv.Val)
	rVal := runtime.CloneBigInt(rv.Val)
	var result *big.Int
	switch op {
	case "&":
		leftPattern := bitPattern(lVal, info)
		rightPattern := bitPattern(rVal, info)
		tmp := new(big.Int).And(leftPattern, rightPattern)
		result = patternToInteger(tmp, info)
	case "|":
		leftPattern := bitPattern(lVal, info)
		rightPattern := bitPattern(rVal, info)
		tmp := new(big.Int).Or(leftPattern, rightPattern)
		result = patternToInteger(tmp, info)
	case "^":
		leftPattern := bitPattern(lVal, info)
		rightPattern := bitPattern(rVal, info)
		tmp := new(big.Int).Xor(leftPattern, rightPattern)
		result = patternToInteger(tmp, info)
	case "<<":
		if !rVal.IsInt64() {
			return nil, fmt.Errorf("shift out of range")
		}
		count := int(rVal.Int64())
		shifted, err := shiftValueLeft(lVal, count, info)
		if err != nil {
			return nil, err
		}
		result = shifted
	case ">>":
		if !rVal.IsInt64() {
			return nil, fmt.Errorf("shift out of range")
		}
		count := int(rVal.Int64())
		shifted, err := shiftValueRight(lVal, count, info)
		if err != nil {
			return nil, err
		}
		result = shifted
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
	if err := ensureFitsInteger(info, result); err != nil {
		return nil, err
	}
	return runtime.IntegerValue{Val: result, TypeSuffix: targetType}, nil
}

func bigFromLiteral(val interface{}) *big.Int {
	switch v := val.(type) {
	case int:
		return big.NewInt(int64(v))
	case int64:
		return big.NewInt(v)
	case float64:
		return big.NewInt(int64(v))
	case string:
		if bi, ok := new(big.Int).SetString(v, 10); ok {
			return bi
		}
		return big.NewInt(0)
	case *big.Int:
		return runtime.CloneBigInt(v)
	default:
		return big.NewInt(0)
	}
}

func evaluateArithmetic(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	leftInt, leftIsInt := left.(runtime.IntegerValue)
	rightInt, rightIsInt := right.(runtime.IntegerValue)
	if leftIsInt && rightIsInt {
		targetType, err := promoteIntegerTypes(leftInt.TypeSuffix, rightInt.TypeSuffix)
		if err != nil {
			return nil, err
		}
		info, err := getIntegerInfo(targetType)
		if err != nil {
			return nil, err
		}
		lv := runtime.CloneBigInt(leftInt.Val)
		rv := runtime.CloneBigInt(rightInt.Val)
		result := new(big.Int)
		switch op {
		case "+":
			result.Add(lv, rv)
		case "-":
			result.Sub(lv, rv)
		case "*":
			result.Mul(lv, rv)
		case "/":
			if rv.Sign() == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result.Quo(lv, rv)
		default:
			return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
		}
		if err := ensureFitsInteger(info, result); err != nil {
			return nil, err
		}
		return runtime.IntegerValue{Val: result, TypeSuffix: targetType}, nil
	}
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	targetFloatKind := floatResultKind(left, right)
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	var val float64
	switch op {
	case "+":
		val = leftFloat + rightFloat
	case "-":
		val = leftFloat - rightFloat
	case "*":
		val = leftFloat * rightFloat
	case "/":
		if rightFloat == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		val = leftFloat / rightFloat
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
	val = normalizeFloat(targetFloatKind, val)
	return runtime.FloatValue{Val: val, TypeSuffix: targetFloatKind}, nil
}

func valuesEqual(left runtime.Value, right runtime.Value) bool {
	switch lv := left.(type) {
	case runtime.StringValue:
		if rv, ok := right.(runtime.StringValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.BoolValue:
		if rv, ok := right.(runtime.BoolValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.CharValue:
		if rv, ok := right.(runtime.CharValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.NilValue:
		_, ok := right.(runtime.NilValue)
		return ok
	case runtime.IntegerValue:
		switch rv := right.(type) {
		case runtime.IntegerValue:
			return lv.Val.Cmp(rv.Val) == 0
		case runtime.FloatValue:
			return math.Abs(bigIntToFloat(lv.Val)-rv.Val) < numericEpsilon
		}
	case runtime.FloatValue:
		switch rv := right.(type) {
		case runtime.FloatValue:
			return math.Abs(lv.Val-rv.Val) < numericEpsilon
		case runtime.IntegerValue:
			return math.Abs(lv.Val-bigIntToFloat(rv.Val)) < numericEpsilon
		}
	}
	return false
}

func evaluateComparison(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if li, ok := left.(runtime.IntegerValue); ok {
		if ri, ok := right.(runtime.IntegerValue); ok {
			cmp := li.Val.Cmp(ri.Val)
			return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
		}
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	cmp := 0
	diff := leftFloat - rightFloat
	if math.Abs(diff) < numericEpsilon {
		cmp = 0
	} else if diff < 0 {
		cmp = -1
	} else {
		cmp = 1
	}
	return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
}
