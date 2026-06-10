package bridge

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

// applyStaticPrimitiveBinaryOperator evaluates primitive Able operators when a
// compiled application intentionally has no interpreter. It only claims
// operations whose dispatch cannot involve a nominal operator implementation.
func applyStaticPrimitiveBinaryOperator(operator string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	operator, dotted := normalizeStaticBinaryOperator(operator)
	left = unwrapStaticPrimitive(left)
	right = unwrapStaticPrimitive(right)

	if leftInt, ok := left.(runtime.IntegerValue); ok {
		if rightInt, ok := right.(runtime.IntegerValue); ok {
			return applyStaticIntegerBinary(operator, dotted, leftInt, rightInt)
		}
	}
	if dotted {
		return nil, false, nil
	}
	if staticNumericValue(left) && staticNumericValue(right) {
		return applyStaticFloatBinary(operator, left, right)
	}
	if leftString, ok := left.(runtime.StringValue); ok {
		if rightString, ok := right.(runtime.StringValue); ok {
			return applyStaticStringBinary(operator, leftString.Val, rightString.Val)
		}
	}
	if leftBool, ok := left.(runtime.BoolValue); ok {
		if rightBool, ok := right.(runtime.BoolValue); ok {
			switch operator {
			case "==":
				return runtime.BoolValue{Val: leftBool.Val == rightBool.Val}, true, nil
			case "!=":
				return runtime.BoolValue{Val: leftBool.Val != rightBool.Val}, true, nil
			}
		}
	}
	if leftChar, ok := left.(runtime.CharValue); ok {
		if rightChar, ok := right.(runtime.CharValue); ok {
			switch operator {
			case "==":
				return runtime.BoolValue{Val: leftChar.Val == rightChar.Val}, true, nil
			case "!=":
				return runtime.BoolValue{Val: leftChar.Val != rightChar.Val}, true, nil
			}
		}
	}
	_, leftNil := left.(runtime.NilValue)
	_, rightNil := right.(runtime.NilValue)
	if leftNil || rightNil {
		switch operator {
		case "==":
			return runtime.BoolValue{Val: leftNil && rightNil}, true, nil
		case "!=":
			return runtime.BoolValue{Val: leftNil != rightNil}, true, nil
		}
	}
	return nil, false, nil
}

func normalizeStaticBinaryOperator(operator string) (string, bool) {
	switch operator {
	case ".&":
		return "&", true
	case ".|":
		return "|", true
	case ".^":
		return "^", true
	case ".<<":
		return "<<", true
	case ".>>":
		return ">>", true
	case "\\xor":
		return "^", false
	default:
		return operator, false
	}
}

func unwrapStaticPrimitive(value runtime.Value) runtime.Value {
	value = unwrapInterface(value)
	switch typed := value.(type) {
	case *runtime.IntegerValue:
		if typed != nil {
			return *typed
		}
	case *runtime.FloatValue:
		if typed != nil {
			return *typed
		}
	case *runtime.StringValue:
		if typed != nil {
			return *typed
		}
	case *runtime.BoolValue:
		if typed != nil {
			return *typed
		}
	case *runtime.CharValue:
		if typed != nil {
			return *typed
		}
	case *runtime.NilValue:
		if typed != nil {
			return *typed
		}
	}
	return value
}

func applyStaticIntegerBinary(operator string, dotted bool, left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	switch operator {
	case "<", "<=", ">", ">=", "==", "!=":
		return runtime.BoolValue{Val: staticComparisonResult(operator, left.CmpInt(right))}, true, nil
	case "+", "-", "*":
		return staticIntegerArithmetic(operator, left, right)
	case "^":
		if dotted {
			return staticIntegerBitwise(operator, left, right)
		}
		return staticIntegerArithmetic(operator, left, right)
	case "/":
		if right.IsZero() {
			return nil, true, fmt.Errorf("division by zero")
		}
		leftFloat := staticIntegerFloat(left)
		rightFloat := staticIntegerFloat(right)
		return runtime.FloatValue{Val: leftFloat / rightFloat, TypeSuffix: runtime.FloatF64}, true, nil
	case "//", "%":
		return staticIntegerDivMod(operator, left, right)
	case "&", "|":
		return staticIntegerBitwise(operator, left, right)
	case "<<", ">>":
		return staticIntegerShift(operator, left, right)
	default:
		return nil, false, nil
	}
}

func staticIntegerArithmetic(operator string, left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return nil, false, nil
	}
	leftValue := new(big.Int).Set(left.BigInt())
	rightValue := new(big.Int).Set(right.BigInt())
	result := new(big.Int)
	switch operator {
	case "+":
		result.Add(leftValue, rightValue)
	case "-":
		result.Sub(leftValue, rightValue)
	case "*":
		result.Mul(leftValue, rightValue)
	case "^":
		if rightValue.Sign() < 0 {
			return nil, true, fmt.Errorf("Negative integer exponent is not supported")
		}
		result.Exp(leftValue, rightValue, nil)
	default:
		return nil, false, nil
	}
	return staticCheckedInteger(result, left.TypeSuffix)
}

func staticIntegerDivMod(operator string, left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return nil, false, nil
	}
	divisor := new(big.Int).Set(right.BigInt())
	if divisor.Sign() == 0 {
		return nil, true, fmt.Errorf("division by zero")
	}
	dividend := new(big.Int).Set(left.BigInt())
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(dividend, divisor, remainder)
	if remainder.Sign() < 0 {
		if divisor.Sign() > 0 {
			quotient.Sub(quotient, big.NewInt(1))
			remainder.Add(remainder, divisor)
		} else {
			quotient.Add(quotient, big.NewInt(1))
			remainder.Sub(remainder, divisor)
		}
	}
	if operator == "//" {
		return staticCheckedInteger(quotient, left.TypeSuffix)
	}
	return staticCheckedInteger(remainder, left.TypeSuffix)
}

func staticIntegerBitwise(operator string, left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return nil, false, nil
	}
	bits, signed, ok := staticIntegerShape(left.TypeSuffix)
	if !ok {
		return nil, true, fmt.Errorf("unsupported integer kind %s", left.TypeSuffix)
	}
	leftPattern := staticIntegerPattern(left.BigInt(), bits)
	rightPattern := staticIntegerPattern(right.BigInt(), bits)
	result := new(big.Int)
	switch operator {
	case "&":
		result.And(leftPattern, rightPattern)
	case "|":
		result.Or(leftPattern, rightPattern)
	case "^":
		result.Xor(leftPattern, rightPattern)
	default:
		return nil, false, nil
	}
	result = staticIntegerFromPattern(result, bits, signed)
	return staticCheckedInteger(result, left.TypeSuffix)
}

func staticIntegerShift(operator string, left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return nil, false, nil
	}
	bits, signed, ok := staticIntegerShape(left.TypeSuffix)
	if !ok {
		return nil, true, fmt.Errorf("unsupported integer kind %s", left.TypeSuffix)
	}
	shift, fits := right.ToInt64()
	if !fits || shift < 0 || shift >= int64(bits) {
		return nil, true, fmt.Errorf("shift count out of range")
	}
	value := new(big.Int).Set(left.BigInt())
	result := new(big.Int)
	if operator == "<<" {
		result.Lsh(value, uint(shift))
	} else if signed {
		result.Rsh(value, uint(shift))
	} else {
		result.Rsh(staticIntegerPattern(value, bits), uint(shift))
	}
	return staticCheckedInteger(result, left.TypeSuffix)
}

func staticCheckedInteger(value *big.Int, kind runtime.IntegerType) (runtime.Value, bool, error) {
	bits, signed, ok := staticIntegerShape(kind)
	if !ok {
		return nil, true, fmt.Errorf("unsupported integer kind %s", kind)
	}
	var min, max *big.Int
	if signed {
		max = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)), big.NewInt(1))
		min = new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)))
	} else {
		min = big.NewInt(0)
		max = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	}
	if value.Cmp(min) < 0 || value.Cmp(max) > 0 {
		return nil, true, fmt.Errorf("integer overflow")
	}
	if small, ok := value.Int64(), value.IsInt64(); ok {
		return ToInt(small, kind), true, nil
	}
	return runtime.NewBigIntValue(new(big.Int).Set(value), kind), true, nil
}

func staticIntegerShape(kind runtime.IntegerType) (bits int, signed bool, ok bool) {
	switch kind {
	case runtime.IntegerI8:
		return 8, true, true
	case runtime.IntegerI16:
		return 16, true, true
	case runtime.IntegerI32:
		return 32, true, true
	case runtime.IntegerI64, runtime.IntegerIsize:
		return 64, true, true
	case runtime.IntegerI128:
		return 128, true, true
	case runtime.IntegerU8:
		return 8, false, true
	case runtime.IntegerU16:
		return 16, false, true
	case runtime.IntegerU32:
		return 32, false, true
	case runtime.IntegerU64, runtime.IntegerUsize:
		return 64, false, true
	case runtime.IntegerU128:
		return 128, false, true
	default:
		return 0, false, false
	}
}

func staticIntegerPattern(value *big.Int, bits int) *big.Int {
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	return new(big.Int).And(new(big.Int).Set(value), mask)
}

func staticIntegerFromPattern(pattern *big.Int, bits int, signed bool) *big.Int {
	if !signed || pattern.Bit(bits-1) == 0 {
		return pattern
	}
	modulus := new(big.Int).Lsh(big.NewInt(1), uint(bits))
	return new(big.Int).Sub(pattern, modulus)
}

func staticIntegerFloat(value runtime.IntegerValue) float64 {
	if small, ok := value.ToInt64(); ok {
		return float64(small)
	}
	result, _ := new(big.Float).SetInt(value.BigInt()).Float64()
	return result
}

func staticNumericValue(value runtime.Value) bool {
	switch value.(type) {
	case runtime.IntegerValue, runtime.FloatValue:
		return true
	default:
		return false
	}
}

func applyStaticFloatBinary(operator string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	leftFloat, leftKind := staticFloatValue(left)
	rightFloat, rightKind := staticFloatValue(right)
	resultKind := runtime.FloatF32
	if leftKind == runtime.FloatF64 || rightKind == runtime.FloatF64 {
		resultKind = runtime.FloatF64
	}
	switch operator {
	case "<", "<=", ">", ">=", "==", "!=":
		if math.IsNaN(leftFloat) || math.IsNaN(rightFloat) {
			return runtime.BoolValue{Val: operator == "!="}, true, nil
		}
		comparison := 0
		if leftFloat < rightFloat {
			comparison = -1
		} else if leftFloat > rightFloat {
			comparison = 1
		}
		return runtime.BoolValue{Val: staticComparisonResult(operator, comparison)}, true, nil
	case "+", "-", "*", "/", "^":
		var result float64
		switch operator {
		case "+":
			result = leftFloat + rightFloat
		case "-":
			result = leftFloat - rightFloat
		case "*":
			result = leftFloat * rightFloat
		case "/":
			result = leftFloat / rightFloat
		case "^":
			result = math.Pow(leftFloat, rightFloat)
		}
		if resultKind == runtime.FloatF32 {
			result = float64(float32(result))
		}
		return runtime.FloatValue{Val: result, TypeSuffix: resultKind}, true, nil
	default:
		return nil, false, nil
	}
}

func staticFloatValue(value runtime.Value) (float64, runtime.FloatType) {
	switch typed := value.(type) {
	case runtime.FloatValue:
		return typed.Val, typed.TypeSuffix
	case runtime.IntegerValue:
		return staticIntegerFloat(typed), runtime.FloatF64
	default:
		return 0, runtime.FloatF64
	}
}

func applyStaticStringBinary(operator string, left string, right string) (runtime.Value, bool, error) {
	if operator == "+" {
		return runtime.StringValue{Val: left + right}, true, nil
	}
	switch operator {
	case "<", "<=", ">", ">=", "==", "!=":
		return runtime.BoolValue{Val: staticComparisonResult(operator, strings.Compare(left, right))}, true, nil
	default:
		return nil, false, nil
	}
}

func staticComparisonResult(operator string, comparison int) bool {
	switch operator {
	case "<":
		return comparison < 0
	case "<=":
		return comparison <= 0
	case ">":
		return comparison > 0
	case ">=":
		return comparison >= 0
	case "==":
		return comparison == 0
	case "!=":
		return comparison != 0
	default:
		return false
	}
}
