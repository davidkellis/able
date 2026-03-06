package interpreter

import (
	"math"
	"math/big"
	"math/bits"

	"able/interpreter-go/pkg/runtime"
)

// Int64 fast-path arithmetic helpers. Each returns (result, overflow).
// When overflow is true the caller must fall back to big.Int.

func addInt64Overflow(a, b int64) (int64, bool) {
	sum, carry := bits.Add64(uint64(a), uint64(b), 0)
	// Overflow when operands have the same sign but result differs.
	signA := a >> 63
	signB := b >> 63
	signR := int64(sum) >> 63
	_ = carry
	overflow := (signA == signB) && (signA != signR)
	return int64(sum), overflow
}

func subInt64Overflow(a, b int64) (int64, bool) {
	diff := a - b
	// Overflow when operands differ in sign and result sign differs from a.
	overflow := ((a ^ b) & (a ^ diff)) < 0
	return diff, overflow
}

func mulInt64Overflow(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, false
	}
	result := a * b
	if result/a != b {
		return 0, true
	}
	return result, false
}

func expInt64Overflow(base, exp int64) (int64, bool) {
	if exp < 0 {
		// Negative exponents not supported for integers.
		return 0, true
	}
	if exp == 0 {
		return 1, false
	}
	if base == 0 {
		return 0, false
	}
	if base == 1 {
		return 1, false
	}
	if base == -1 {
		if exp%2 == 0 {
			return 1, false
		}
		return -1, false
	}
	// For exponents > 62, any base with |base| > 1 will overflow int64.
	if exp > 62 {
		return 0, true
	}
	result := int64(1)
	b := base
	e := exp
	for e > 0 {
		if e&1 == 1 {
			product, overflow := mulInt64Overflow(result, b)
			if overflow {
				return 0, true
			}
			result = product
		}
		e >>= 1
		if e > 0 {
			sq, overflow := mulInt64Overflow(b, b)
			if overflow {
				return 0, true
			}
			b = sq
		}
	}
	return result, false
}

// ensureFitsInt64 checks that the result fits in the target integer type.
// For int64 results we compare against the known min/max bounds without allocating.
func ensureFitsInt64(info integerInfo, value int64) error {
	if info.bits >= 64 && info.signed {
		// i64, i128, isize — any int64 fits.
		return nil
	}
	if info.bits > 64 {
		// u128, i128 — any int64 fits.
		return nil
	}
	// For smaller types, check against precomputed int64 bounds.
	minVal, maxVal := int64Bounds(info)
	if value < minVal || value > maxVal {
		return newOverflowError("integer overflow")
	}
	return nil
}

// ensureFitsInt64Type checks fit using the integer type suffix directly,
// avoiding integer-info map lookups on hot arithmetic paths.
func ensureFitsInt64Type(kind runtime.IntegerType, value int64) error {
	switch kind {
	case runtime.IntegerI8:
		if value < math.MinInt8 || value > math.MaxInt8 {
			return newOverflowError("integer overflow")
		}
	case runtime.IntegerI16:
		if value < math.MinInt16 || value > math.MaxInt16 {
			return newOverflowError("integer overflow")
		}
	case runtime.IntegerI32:
		if value < math.MinInt32 || value > math.MaxInt32 {
			return newOverflowError("integer overflow")
		}
	case runtime.IntegerI64, runtime.IntegerI128, runtime.IntegerIsize:
		// Any int64 fits.
	case runtime.IntegerU8:
		if value < 0 || value > math.MaxUint8 {
			return newOverflowError("integer overflow")
		}
	case runtime.IntegerU16:
		if value < 0 || value > math.MaxUint16 {
			return newOverflowError("integer overflow")
		}
	case runtime.IntegerU32:
		if value < 0 || value > math.MaxUint32 {
			return newOverflowError("integer overflow")
		}
	case runtime.IntegerU64, runtime.IntegerU128, runtime.IntegerUsize:
		if value < 0 {
			return newOverflowError("integer overflow")
		}
	default:
		return newOverflowError("integer overflow")
	}
	return nil
}

// int64Bounds returns the min/max that an int64 can represent for a given integer kind.
func int64Bounds(info integerInfo) (int64, int64) {
	if info.signed {
		switch info.bits {
		case 8:
			return math.MinInt8, math.MaxInt8
		case 16:
			return math.MinInt16, math.MaxInt16
		case 32:
			return math.MinInt32, math.MaxInt32
		case 64:
			return math.MinInt64, math.MaxInt64
		default:
			return math.MinInt64, math.MaxInt64
		}
	}
	// Unsigned
	switch info.bits {
	case 8:
		return 0, math.MaxUint8
	case 16:
		return 0, math.MaxUint16
	case 32:
		return 0, math.MaxUint32
	case 64:
		// MaxUint64 doesn't fit in int64; this bound means we can't
		// represent all u64 values in int64. But IsInt64() already
		// checks this — only values ≤ MaxInt64 reach here.
		return 0, math.MaxInt64
	default:
		return 0, math.MaxInt64
	}
}

// euclideanDivModInt64 performs Euclidean division/modulo on int64 values.
func euclideanDivModInt64(dividend, divisor int64) (int64, int64) {
	q := dividend / divisor
	r := dividend % divisor
	if r < 0 {
		if divisor > 0 {
			q--
			r += divisor
		} else {
			q++
			r -= divisor
		}
	}
	return q, r
}

// bigNewInt returns a *big.Int from an int64 value. This is just big.NewInt
// but named explicitly to signal it's the only alloc in the fast path.
func bigNewInt(v int64) *big.Int {
	return big.NewInt(v)
}

// smallIntWithinRange checks whether an int64 value fits within the range of
// the given integer type, without allocating big.Int.
func smallIntWithinRange(val int64, target runtime.IntegerType) bool {
	info, err := getIntegerInfo(target)
	if err != nil {
		return false
	}
	minVal, maxVal := int64Bounds(info)
	return val >= minVal && val <= maxVal
}
