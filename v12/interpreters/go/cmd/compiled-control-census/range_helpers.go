package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
)

func isPrimitiveRangeHelper(name string) bool {
	switch name {
	case "__able_raise_overflow",
		"__able_checked_add_signed", "__able_checked_sub_signed", "__able_checked_mul_signed",
		"__able_checked_add_unsigned", "__able_checked_sub_unsigned", "__able_checked_mul_unsigned",
		"__able_divmod_signed", "__able_divmod_unsigned",
		"__able_shift_left_signed", "__able_shift_right_signed",
		"__able_shift_left_unsigned", "__able_shift_right_unsigned":
		return true
	default:
		return false
	}
}

func primitiveKind(name string) string {
	switch name {
	case "__able_raise_overflow":
		return "inline-overflow"
	case "__able_divmod_signed", "__able_divmod_unsigned":
		return "division-or-modulo"
	case "__able_shift_left_signed", "__able_shift_right_signed", "__able_shift_left_unsigned", "__able_shift_right_unsigned":
		return "shift"
	default:
		return "checked-arithmetic"
	}
}

func evaluatePrimitiveHelper(name string, expressions []ast.Expr, args []intRange) (intRange, bool, string) {
	if len(args) < 2 {
		return intRange{}, false, "helper operands are unavailable"
	}
	left, right := args[0], args[1]
	bits, bitsOK := helperBits(expressions, args)
	switch name {
	case "__able_checked_add_signed":
		return signedArithmeticResult("add", left, right, bits, bitsOK)
	case "__able_checked_sub_signed":
		return signedArithmeticResult("subtract", left, right, bits, bitsOK)
	case "__able_checked_mul_signed":
		return signedArithmeticResult("multiply", left, right, bits, bitsOK)
	case "__able_checked_add_unsigned":
		return unsignedArithmeticResult("add", left, right, bits, bitsOK)
	case "__able_checked_sub_unsigned":
		return unsignedArithmeticResult("subtract", left, right, bits, bitsOK)
	case "__able_checked_mul_unsigned":
		return unsignedArithmeticResult("multiply", left, right, bits, bitsOK)
	case "__able_divmod_signed":
		return signedDivResult(left, right, bits, bitsOK)
	case "__able_divmod_unsigned":
		return unsignedDivResult(left, right, bits, bitsOK)
	case "__able_shift_left_signed":
		return signedShiftResult(left, right, bits, bitsOK, true)
	case "__able_shift_right_signed":
		return signedShiftResult(left, right, bits, bitsOK, false)
	case "__able_shift_left_unsigned":
		return unsignedShiftResult(left, right, bits, bitsOK, true)
	case "__able_shift_right_unsigned":
		return unsignedShiftResult(left, right, bits, bitsOK, false)
	default:
		return intRange{}, false, "unsupported primitive helper"
	}
}

func helperBits(expressions []ast.Expr, args []intRange) (int64, bool) {
	if len(expressions) < 3 || len(args) < 3 {
		return 0, false
	}
	if value, ok := integerLiteral(expressions[2]); ok {
		return value, value > 0 && value <= 64
	}
	if args[2].Known && args[2].Min == args[2].Max {
		return args[2].Min, args[2].Min > 0 && args[2].Min <= 64
	}
	return 0, false
}

func signedBounds(bits int64) (intRange, bool) {
	if bits <= 0 || bits > 64 {
		return intRange{}, false
	}
	if bits == 64 {
		return knownRange(math.MinInt64, math.MaxInt64), true
	}
	maxValue := int64(1)<<(bits-1) - 1
	return knownRange(-maxValue-1, maxValue), true
}

func unsignedMax(bits int64) (int64, bool) {
	if bits <= 0 || bits > 64 {
		return 0, false
	}
	if bits >= 63 {
		// intRange can represent only the non-negative half of u64. Any
		// interval known to this analyzer is therefore below this conservative
		// ceiling and also below the actual unsigned bound.
		return math.MaxInt64, true
	}
	return int64(1)<<bits - 1, true
}

func signedArithmeticResult(operation string, left, right intRange, bits int64, bitsOK bool) (intRange, bool, string) {
	if !bitsOK || !left.Known || !right.Known {
		return intRange{}, false, fmt.Sprintf("%s and %s with unknown width or operand range", describeRange(left), describeRange(right))
	}
	var result intRange
	switch operation {
	case "add":
		result = addRange(left, right)
	case "subtract":
		result = subtractRange(left, right)
	case "multiply":
		result = multiplyRange(left, right)
	}
	bounds, ok := signedBounds(bits)
	safe := ok && result.Known && result.Min >= bounds.Min && result.Max <= bounds.Max
	return result, safe, fmt.Sprintf("%s result %s for signed %d-bit bounds", operation, describeRange(result), bits)
}

func unsignedArithmeticResult(operation string, left, right intRange, bits int64, bitsOK bool) (intRange, bool, string) {
	if !bitsOK || !left.Known || !right.Known || left.Min < 0 || right.Min < 0 {
		return intRange{}, false, fmt.Sprintf("%s and %s do not prove unsigned operands", describeRange(left), describeRange(right))
	}
	var result intRange
	switch operation {
	case "add":
		result = addRange(left, right)
	case "subtract":
		result = subtractRange(left, right)
	case "multiply":
		result = multiplyRange(left, right)
	}
	maxValue, ok := unsignedMax(bits)
	safe := ok && result.Known && result.Min >= 0 && result.Max <= maxValue
	return result, safe, fmt.Sprintf("%s result %s for unsigned %d-bit bounds", operation, describeRange(result), bits)
}

func signedDivResult(left, right intRange, bits int64, bitsOK bool) (intRange, bool, string) {
	bounds, boundsOK := signedBounds(bits)
	nonzero := right.Known && (right.Min > 0 || right.Max < 0)
	minOverflow := left.Known && right.Known && left.Min <= bounds.Min && left.Max >= bounds.Min && right.Min <= -1 && right.Max >= -1
	safe := bitsOK && boundsOK && left.Known && nonzero && !minOverflow
	result := intRange{}
	if safe && left.Min >= 0 && right.Min > 0 {
		result = knownRange(left.Min/right.Max, left.Max/right.Min)
	}
	return result, safe, fmt.Sprintf("dividend %s, divisor %s, nonzero=%t, min/-1 possible=%t", describeRange(left), describeRange(right), nonzero, minOverflow)
}

func unsignedDivResult(left, right intRange, bits int64, bitsOK bool) (intRange, bool, string) {
	nonzero := right.Known && right.Min > 0
	maxValue, maxOK := unsignedMax(bits)
	safe := bitsOK && maxOK && left.Known && left.Min >= 0 && left.Max <= maxValue && nonzero
	result := intRange{}
	if safe {
		result = knownRange(left.Min/right.Max, left.Max/right.Min)
	}
	return result, safe, fmt.Sprintf("unsigned dividend %s and divisor %s", describeRange(left), describeRange(right))
}

func signedShiftResult(value, shift intRange, bits int64, bitsOK bool, left bool) (intRange, bool, string) {
	shiftSafe := bitsOK && shift.Known && shift.Min >= 0 && shift.Max < bits
	if !shiftSafe || !value.Known {
		return intRange{}, false, fmt.Sprintf("value %s and shift %s for %d-bit width", describeRange(value), describeRange(shift), bits)
	}
	if !left {
		return intRange{}, true, fmt.Sprintf("right shift %s is within [0,%d]", describeRange(shift), bits-1)
	}
	if value.Min < 0 || shift.Max >= 63 {
		return intRange{}, false, "signed left-shift range is not non-negative and representable"
	}
	factor := int64(1) << shift.Max
	maxResult, ok := mul64(value.Max, factor)
	bounds, boundsOK := signedBounds(bits)
	safe := ok && boundsOK && maxResult <= bounds.Max
	result := intRange{}
	if safe {
		result = knownRange(value.Min<<shift.Min, maxResult)
	}
	return result, safe, fmt.Sprintf("left shift of %s by %s yields maximum %d", describeRange(value), describeRange(shift), maxResult)
}

func unsignedShiftResult(value, shift intRange, bits int64, bitsOK bool, left bool) (intRange, bool, string) {
	shiftSafe := bitsOK && shift.Known && shift.Min >= 0 && shift.Max < bits
	maxValue, maxOK := unsignedMax(bits)
	if !shiftSafe || !maxOK || !value.Known || value.Min < 0 {
		return intRange{}, false, fmt.Sprintf("unsigned value %s and shift %s for %d-bit width", describeRange(value), describeRange(shift), bits)
	}
	if !left {
		return intRange{}, true, fmt.Sprintf("right shift %s is within [0,%d]", describeRange(shift), bits-1)
	}
	factor := int64(1) << shift.Max
	maxResult, ok := mul64(value.Max, factor)
	safe := ok && maxResult <= maxValue
	result := intRange{}
	if safe {
		result = knownRange(value.Min<<shift.Min, maxResult)
	}
	return result, safe, fmt.Sprintf("unsigned left shift yields maximum %d", maxResult)
}

func refineRangeEnv(env rangeEnv, condition ast.Expr, truth bool) rangeEnv {
	refined := cloneRangeEnv(env)
	binary, ok := condition.(*ast.BinaryExpr)
	if !ok {
		if unary, ok := condition.(*ast.UnaryExpr); ok && unary.Op == token.NOT {
			return refineRangeEnv(refined, unary.X, !truth)
		}
		return refined
	}
	if binary.Op == token.LAND && truth {
		return refineRangeEnv(refineRangeEnv(refined, binary.X, true), binary.Y, true)
	}
	if binary.Op == token.LOR && !truth {
		return refineRangeEnv(refineRangeEnv(refined, binary.X, false), binary.Y, false)
	}
	name := identName(binary.X)
	value, literalOK := integerLiteral(binary.Y)
	if name == "" || !literalOK {
		return refined
	}
	current := refined[name]
	if !current.Known {
		return refined
	}
	operator := binary.Op
	if !truth {
		operator = inverseComparison(operator)
	}
	switch operator {
	case token.LSS:
		if value > math.MinInt64 {
			current = intersectRange(current, knownRange(math.MinInt64, value-1))
		}
	case token.LEQ:
		current = intersectRange(current, knownRange(math.MinInt64, value))
	case token.GTR:
		if value < math.MaxInt64 {
			current = intersectRange(current, knownRange(value+1, math.MaxInt64))
		}
	case token.GEQ:
		current = intersectRange(current, knownRange(value, math.MaxInt64))
	case token.EQL:
		current = intersectRange(current, knownRange(value, value))
	case token.NEQ:
		if current.Min == value && current.Max > value {
			current.Min++
		} else if current.Max == value && current.Min < value {
			current.Max--
		}
	}
	refined[name] = current
	return refined
}

func inverseComparison(operator token.Token) token.Token {
	switch operator {
	case token.LSS:
		return token.GEQ
	case token.LEQ:
		return token.GTR
	case token.GTR:
		return token.LEQ
	case token.GEQ:
		return token.LSS
	case token.EQL:
		return token.NEQ
	case token.NEQ:
		return token.EQL
	default:
		return operator
	}
}
