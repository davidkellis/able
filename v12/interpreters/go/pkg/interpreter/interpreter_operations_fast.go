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
			// Int64 fast path for negation.
			if val, ok := v.ToInt64(); ok && val != math.MinInt64 {
				neg := -val
				info, err := getIntegerInfo(v.TypeSuffix)
				if err != nil {
					return nil, true, err
				}
				if err := ensureFitsInt64(info, neg); err != nil {
					return nil, true, err
				}
				return runtime.NewSmallInt(neg, v.TypeSuffix), true, nil
			}
			neg := new(big.Int).Neg(v.BigInt())
			info, err := getIntegerInfo(v.TypeSuffix)
			if err != nil {
				return nil, true, err
			}
			if err := ensureFitsInteger(info, neg); err != nil {
				return nil, true, err
			}
			return runtime.NewBigIntValue(neg, v.TypeSuffix), true, nil
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
				val := new(big.Int).Set(v.BigInt())
				if val.Sign() < 0 {
					return nil, true, fmt.Errorf("bitwise not on unsigned requires non-negative operand")
				}
				result := new(big.Int).Xor(mask, val)
				return runtime.NewBigIntValue(result, v.TypeSuffix), true, nil
			}
			neg := new(big.Int).Neg(new(big.Int).Add(v.BigInt(), big.NewInt(1)))
			return runtime.NewBigIntValue(neg, v.TypeSuffix), true, nil
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
		return evaluateIntegerArithmeticFast(op, leftInt, rightInt)
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

func evaluateIntegerArithmeticFast(op string, leftInt runtime.IntegerValue, rightInt runtime.IntegerValue) (runtime.Value, error) {
	if leftInt.TypeSuffix == rightInt.TypeSuffix {
		targetType := leftInt.TypeSuffix
		if l, lok := leftInt.ToInt64(); lok {
			if r, rok := rightInt.ToInt64(); rok {
				var result int64
				var overflow bool
				switch op {
				case "+":
					result, overflow = addInt64Overflow(l, r)
				case "-":
					result, overflow = subInt64Overflow(l, r)
				case "*":
					result, overflow = mulInt64Overflow(l, r)
				case "^":
					if r < 0 {
						return nil, fmt.Errorf("Negative integer exponent is not supported")
					}
					result, overflow = expInt64Overflow(l, r)
				default:
					return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
				}
				if !overflow {
					if err := ensureFitsInt64Type(targetType, result); err != nil {
						return nil, err
					}
					return boxedOrSmallIntegerValue(targetType, result), nil
				}
			}
		}
	}

	targetType, err := promoteIntegerTypes(leftInt.TypeSuffix, rightInt.TypeSuffix)
	if err != nil {
		return nil, err
	}
	// Int64 fast path: avoid big.Int allocation when both operands fit in int64.
	if l, lok := leftInt.ToInt64(); lok {
		if r, rok := rightInt.ToInt64(); rok {
			var result int64
			var overflow bool
			switch op {
			case "+":
				result, overflow = addInt64Overflow(l, r)
			case "-":
				result, overflow = subInt64Overflow(l, r)
			case "*":
				result, overflow = mulInt64Overflow(l, r)
			case "^":
				if r < 0 {
					return nil, fmt.Errorf("Negative integer exponent is not supported")
				}
				result, overflow = expInt64Overflow(l, r)
			default:
				return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
			}
			if !overflow {
				if err := ensureFitsInt64Type(targetType, result); err != nil {
					return nil, err
				}
				return boxedOrSmallIntegerValue(targetType, result), nil
			}
		}
	}
	info, err := getIntegerInfo(targetType)
	if err != nil {
		return nil, err
	}
	// Big.Int fallback.
	lv := runtime.CloneBigInt(leftInt.BigInt())
	rv := runtime.CloneBigInt(rightInt.BigInt())
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
	return runtime.NewBigIntValue(result, targetType), nil
}

func evaluateDivisionFast(left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	leftInt, leftIsInt := left.(runtime.IntegerValue)
	rightInt, rightIsInt := right.(runtime.IntegerValue)
	if leftIsInt && rightIsInt {
		if rightInt.IsZero() {
			return nil, newDivisionByZeroError()
		}
		// Int64 fast path for integer-to-float division.
		if l, lok := leftInt.ToInt64(); lok {
			if r, rok := rightInt.ToInt64(); rok {
				if r == 0 {
					return nil, newDivisionByZeroError()
				}
				val := normalizeFloat(runtime.FloatF64, float64(l)/float64(r))
				return runtime.FloatValue{Val: val, TypeSuffix: runtime.FloatF64}, nil
			}
		}
		leftFloat := bigIntToFloat(leftInt.BigInt())
		rightFloat := bigIntToFloat(rightInt.BigInt())
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
	// Int64 fast path.
	if l, lok := lv.ToInt64(); lok {
		if r, rok := rv.ToInt64(); rok {
			if r == 0 {
				return nil, newDivisionByZeroError()
			}
			targetType, err := promoteIntegerTypes(lv.TypeSuffix, rv.TypeSuffix)
			if err != nil {
				return nil, err
			}
			q, rem := euclideanDivModInt64(l, r)
			switch op {
			case "//":
				if err := ensureFitsInt64Type(targetType, q); err != nil {
					return nil, err
				}
				return boxedOrSmallIntegerValue(targetType, q), nil
			case "%":
				if err := ensureFitsInt64Type(targetType, rem); err != nil {
					return nil, err
				}
				return boxedOrSmallIntegerValue(targetType, rem), nil
			default:
				return nil, fmt.Errorf("unsupported div/mod operator %s", op)
			}
		}
	}
	quotient, remainder, _, err := computeDivMod(lv, rv)
	if err != nil {
		return nil, err
	}
	switch op {
	case "//":
		if boxed, ok := maybeBoxedIntegerValue(quotient); ok {
			return boxed, nil
		}
		return quotient, nil
	case "%":
		if boxed, ok := maybeBoxedIntegerValue(remainder); ok {
			return boxed, nil
		}
		return remainder, nil
	default:
		return nil, fmt.Errorf("unsupported div/mod operator %s", op)
	}
}
