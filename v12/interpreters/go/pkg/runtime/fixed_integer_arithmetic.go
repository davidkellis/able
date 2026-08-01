package runtime

import (
	"fmt"
	"math/big"
)

// FixedIntegerOperation identifies one of the non-raising primitive
// arithmetic operations exposed by the language.
type FixedIntegerOperation uint8

const (
	FixedIntegerAdd FixedIntegerOperation = iota
	FixedIntegerSub
	FixedIntegerMul
)

// FixedIntegerMode selects the overflow result contract.
type FixedIntegerMode uint8

const (
	FixedIntegerWrapping FixedIntegerMode = iota
	FixedIntegerSaturating
	FixedIntegerChecked
)

// FixedIntegerArithmetic applies one alternative arithmetic operation to
// same-typed fixed-width integer values. The boolean is false only when
// checked arithmetic overflows.
func FixedIntegerArithmetic(
	left IntegerValue,
	right IntegerValue,
	operation FixedIntegerOperation,
	mode FixedIntegerMode,
) (IntegerValue, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return IntegerValue{}, false, fmt.Errorf(
			"fixed integer arithmetic requires matching types, got %s and %s",
			left.TypeSuffix,
			right.TypeSuffix,
		)
	}
	bits, signed, ok := fixedIntegerShape(left.TypeSuffix)
	if !ok {
		return IntegerValue{}, false, fmt.Errorf(
			"fixed integer arithmetic does not support %s",
			left.TypeSuffix,
		)
	}

	result := new(big.Int)
	switch operation {
	case FixedIntegerAdd:
		result.Add(left.BigInt(), right.BigInt())
	case FixedIntegerSub:
		result.Sub(left.BigInt(), right.BigInt())
	case FixedIntegerMul:
		result.Mul(left.BigInt(), right.BigInt())
	default:
		return IntegerValue{}, false, fmt.Errorf(
			"unknown fixed integer arithmetic operation %d",
			operation,
		)
	}

	min, max, modulus := fixedIntegerBounds(bits, signed)
	if result.Cmp(min) >= 0 && result.Cmp(max) <= 0 {
		return compactIntegerValue(result, left.TypeSuffix), true, nil
	}

	switch mode {
	case FixedIntegerChecked:
		return IntegerValue{}, false, nil
	case FixedIntegerSaturating:
		if result.Cmp(min) < 0 {
			return compactIntegerValue(min, left.TypeSuffix), true, nil
		}
		return compactIntegerValue(max, left.TypeSuffix), true, nil
	case FixedIntegerWrapping:
		result.Mod(result, modulus)
		if signed && result.Bit(bits-1) != 0 {
			result.Sub(result, modulus)
		}
		return compactIntegerValue(result, left.TypeSuffix), true, nil
	default:
		return IntegerValue{}, false, fmt.Errorf(
			"unknown fixed integer arithmetic mode %d",
			mode,
		)
	}
}

func fixedIntegerShape(kind IntegerType) (bits int, signed bool, ok bool) {
	switch kind {
	case IntegerI8:
		return 8, true, true
	case IntegerI16:
		return 16, true, true
	case IntegerI32:
		return 32, true, true
	case IntegerI64:
		return 64, true, true
	case IntegerI128:
		return 128, true, true
	case IntegerU8:
		return 8, false, true
	case IntegerU16:
		return 16, false, true
	case IntegerU32:
		return 32, false, true
	case IntegerU64:
		return 64, false, true
	case IntegerU128:
		return 128, false, true
	default:
		return 0, false, false
	}
}

func fixedIntegerBounds(bits int, signed bool) (min *big.Int, max *big.Int, modulus *big.Int) {
	modulus = new(big.Int).Lsh(big.NewInt(1), uint(bits))
	if !signed {
		return big.NewInt(0), new(big.Int).Sub(new(big.Int).Set(modulus), big.NewInt(1)), modulus
	}
	half := new(big.Int).Rsh(new(big.Int).Set(modulus), 1)
	min = new(big.Int).Neg(new(big.Int).Set(half))
	max = new(big.Int).Sub(half, big.NewInt(1))
	return min, max, modulus
}

func compactIntegerValue(value *big.Int, kind IntegerType) IntegerValue {
	if value.IsInt64() {
		return NewSmallInt(value.Int64(), kind)
	}
	return NewBigIntValue(new(big.Int).Set(value), kind)
}
