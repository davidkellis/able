package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

// ApplyBinaryOperatorFast performs a best-effort operator evaluation without invoking interpreter dispatch.
// It returns handled=false when it cannot safely evaluate the operation.
func ApplyBinaryOperatorFast(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	rawOp := op
	op, dotted := normalizeOperator(op)
	rawLeft := unwrapInterfaceValue(left)
	rawRight := unwrapInterfaceValue(right)
	rawLeft = unwrapScalarValue(rawLeft)
	rawRight = unwrapScalarValue(rawRight)

	if isRatioValue(rawLeft) || isRatioValue(rawRight) {
		return nil, false, nil
	}

	switch op {
	case "+", "-", "*", "^":
		if op == "^" && dotted {
			if isIntegerValue(rawLeft) && isIntegerValue(rawRight) {
				val, err := evaluateBitwise(op, rawLeft, rawRight)
				return val, true, err
			}
			return nil, false, nil
		}
		if op == "+" {
			if ls, ok := stringFromValue(rawLeft); ok {
				if rs, ok := stringFromValue(rawRight); ok {
					return runtime.StringValue{Val: ls + rs}, true, nil
				}
				return nil, true, fmt.Errorf("Arithmetic requires numeric operands")
			}
			if _, ok := stringFromValue(rawRight); ok {
				return nil, true, fmt.Errorf("Arithmetic requires numeric operands")
			}
		}
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			val, err := evaluateArithmeticFast(op, rawLeft, rawRight)
			return val, true, err
		}
		return nil, false, nil
	case "/":
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			val, err := evaluateDivisionFast(rawLeft, rawRight)
			return val, true, err
		}
		return nil, false, nil
	case "//", "%":
		if isIntegerValue(rawLeft) && isIntegerValue(rawRight) {
			val, err := evaluateDivModFast(op, rawLeft, rawRight)
			return val, true, err
		}
		return nil, false, nil
	case "/%":
		return nil, false, nil
	case "<", "<=", ">", ">=", "==", "!=":
		_, leftIsString := stringFromValue(rawLeft)
		_, rightIsString := stringFromValue(rawRight)
		if (isNumericValue(rawLeft) && isNumericValue(rawRight)) || (leftIsString && rightIsString) {
			val, err := evaluateComparison(op, rawLeft, rawRight)
			return val, true, err
		}
		return nil, false, nil
	case "&", "|", "<<", ">>":
		if isIntegerValue(rawLeft) && isIntegerValue(rawRight) {
			val, err := evaluateBitwise(op, rawLeft, rawRight)
			return val, true, err
		}
		return nil, false, nil
	default:
		_ = rawOp
		return nil, false, nil
	}
}

// ApplyUnaryOperatorFast performs a best-effort unary operator evaluation without interpreter dispatch.
// It returns handled=false when it cannot safely evaluate the operation.
func ApplyUnaryOperatorFast(operator string, operand runtime.Value) (runtime.Value, bool, error) {
	rawOperand := unwrapInterfaceValue(operand)
	rawOperand = unwrapScalarValue(rawOperand)
	switch operator {
	case "-":
		switch v := rawOperand.(type) {
		case runtime.IntegerValue:
			neg := new(big.Int).Neg(v.Val)
			info, err := getIntegerInfo(v.TypeSuffix)
			if err != nil {
				return nil, true, err
			}
			if err := ensureFitsInteger(info, neg); err != nil {
				return nil, true, err
			}
			return runtime.IntegerValue{Val: neg, TypeSuffix: v.TypeSuffix}, true, nil
		case runtime.FloatValue:
			return runtime.FloatValue{Val: -v.Val, TypeSuffix: v.TypeSuffix}, true, nil
		default:
			return nil, false, nil
		}
	case "~", ".~":
		switch v := rawOperand.(type) {
		case runtime.IntegerValue:
			if strings.HasPrefix(string(v.TypeSuffix), "u") {
				width := integerBitWidth(v.TypeSuffix)
				if width <= 0 {
					return nil, true, fmt.Errorf("unsupported integer width for bitwise not")
				}
				mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(width)), big.NewInt(1))
				val := new(big.Int).Set(v.Val)
				if val.Sign() < 0 {
					return nil, true, fmt.Errorf("bitwise not on unsigned requires non-negative operand")
				}
				result := new(big.Int).Xor(mask, val)
				return runtime.IntegerValue{Val: result, TypeSuffix: v.TypeSuffix}, true, nil
			}
			neg := new(big.Int).Neg(new(big.Int).Add(v.Val, big.NewInt(1)))
			return runtime.IntegerValue{Val: neg, TypeSuffix: v.TypeSuffix}, true, nil
		default:
			return nil, false, nil
		}
	default:
		return nil, false, nil
	}
}

func unwrapScalarValue(val runtime.Value) runtime.Value {
	switch v := val.(type) {
	case *runtime.IntegerValue:
		if v != nil {
			return *v
		}
	case *runtime.FloatValue:
		if v != nil {
			return *v
		}
	case *runtime.StringValue:
		if v != nil {
			return *v
		}
	case *runtime.BoolValue:
		if v != nil {
			return *v
		}
	case *runtime.CharValue:
		if v != nil {
			return *v
		}
	}
	return val
}

func evaluateArithmeticFast(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
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
		case "^":
			if rv.Sign() < 0 {
				return nil, fmt.Errorf("Negative integer exponent is not supported")
			}
			result.Exp(lv, rv, nil)
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
	case "^":
		val = math.Pow(leftFloat, rightFloat)
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
	val = normalizeFloat(targetFloatKind, val)
	return runtime.FloatValue{Val: val, TypeSuffix: targetFloatKind}, nil
}

func evaluateDivisionFast(left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	leftInt, leftIsInt := left.(runtime.IntegerValue)
	rightInt, rightIsInt := right.(runtime.IntegerValue)
	if leftIsInt && rightIsInt {
		if rightInt.Val == nil || rightInt.Val.Sign() == 0 {
			return nil, newDivisionByZeroError()
		}
		leftFloat := bigIntToFloat(leftInt.Val)
		rightFloat := bigIntToFloat(rightInt.Val)
		if rightFloat == 0 {
			return nil, newDivisionByZeroError()
		}
		val := normalizeFloat(runtime.FloatF64, leftFloat/rightFloat)
		return runtime.FloatValue{Val: val, TypeSuffix: runtime.FloatF64}, nil
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
	if rightFloat == 0 {
		return nil, newDivisionByZeroError()
	}
	val := normalizeFloat(targetFloatKind, leftFloat/rightFloat)
	return runtime.FloatValue{Val: val, TypeSuffix: targetFloatKind}, nil
}

func evaluateDivModFast(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	lv, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Arithmetic requires integer operands")
	}
	rv, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Arithmetic requires integer operands")
	}
	quotient, remainder, _, err := computeDivMod(lv, rv)
	if err != nil {
		return nil, err
	}
	switch op {
	case "//":
		return quotient, nil
	case "%":
		return remainder, nil
	default:
		return nil, fmt.Errorf("unsupported div/mod operator %s", op)
	}
}
