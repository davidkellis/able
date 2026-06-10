package runtime

import (
	"math"
	"math/big"
	"math/bits"
)

// Uint128 is the native compiled carrier for Able's primitive u128 type.
// High and Low store the unsigned value in base 2^64.
type Uint128 struct {
	High uint64
	Low  uint64
}

// Int128 is the native compiled carrier for Able's primitive i128 type.
// High and Low store the two's-complement bit pattern in base 2^64.
type Int128 struct {
	High uint64
	Low  uint64
}

func Uint128FromUint64(value uint64) Uint128 { return Uint128{Low: value} }

func Uint128FromInt64(value int64) Uint128 {
	if value >= 0 {
		return Uint128{Low: uint64(value)}
	}
	return Uint128{High: ^uint64(0), Low: uint64(value)}
}

func Int128FromInt64(value int64) Int128 {
	if value >= 0 {
		return Int128{Low: uint64(value)}
	}
	return Int128{High: ^uint64(0), Low: uint64(value)}
}

func Int128FromUint64(value uint64) Int128 { return Int128{Low: value} }

func Uint128FromFloat64(value float64) (Uint128, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value >= math.Ldexp(1, 128) {
		return Uint128{}, false
	}
	integer, _ := new(big.Float).SetFloat64(math.Trunc(value)).Int(nil)
	return Uint128FromBig(integer)
}

func Int128FromFloat64(value float64) (Int128, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -math.Ldexp(1, 127) || value >= math.Ldexp(1, 127) {
		return Int128{}, false
	}
	integer, _ := new(big.Float).SetFloat64(math.Trunc(value)).Int(nil)
	return Int128FromBig(integer)
}

func Int128FromBits(value Uint128) Int128 { return Int128{High: value.High, Low: value.Low} }

func Uint128FromBits(value Int128) Uint128 { return Uint128{High: value.High, Low: value.Low} }

func Uint128FromBig(value *big.Int) (Uint128, bool) {
	if value == nil || value.Sign() < 0 || value.BitLen() > 128 {
		return Uint128{}, false
	}
	words := value.Bits()
	result := Uint128{}
	if bits.UintSize == 64 {
		if len(words) > 0 {
			result.Low = uint64(words[0])
		}
		if len(words) > 1 {
			result.High = uint64(words[1])
		}
		return result, true
	}
	for i := len(words) - 1; i >= 0; i-- {
		result = result.ShiftLeftUnchecked(32)
		result.Low |= uint64(words[i])
	}
	return result, true
}

func Int128FromBig(value *big.Int) (Int128, bool) {
	if value == nil {
		return Int128{}, false
	}
	min := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1))
	if value.Cmp(min) < 0 || value.Cmp(max) > 0 {
		return Int128{}, false
	}
	if value.Sign() < 0 {
		magnitude, ok := Uint128FromBig(new(big.Int).Neg(value))
		if !ok {
			return Int128{}, false
		}
		pattern := mustUintAdd(Uint128{High: ^magnitude.High, Low: ^magnitude.Low}, Uint128{Low: 1})
		return Int128FromBits(pattern), true
	}
	u, ok := Uint128FromBig(value)
	return Int128FromBits(u), ok
}

func (value Uint128) BigInt() *big.Int {
	result := new(big.Int).SetUint64(value.High)
	result.Lsh(result, 64)
	result.Or(result, new(big.Int).SetUint64(value.Low))
	return result
}

func (value Int128) BigInt() *big.Int {
	if value.IsNegative() {
		return new(big.Int).Neg(value.magnitude().BigInt())
	}
	return Uint128FromBits(value).BigInt()
}

func (value Uint128) IntegerValue() IntegerValue {
	if value.High == 0 && value.Low <= uint64(^uint64(0)>>1) {
		return NewSmallInt(int64(value.Low), IntegerU128)
	}
	return NewBigIntValue(value.BigInt(), IntegerU128)
}

func (value Int128) IntegerValue() IntegerValue {
	if (value.High == 0 && value.Low <= uint64(^uint64(0)>>1)) ||
		(value.High == ^uint64(0) && value.Low >= uint64(1)<<63) {
		return NewSmallInt(int64(value.Low), IntegerI128)
	}
	return NewBigIntValue(value.BigInt(), IntegerI128)
}

func Uint128FromValue(value Value) (Uint128, bool) {
	integer, ok := wideIntegerValue(value)
	if !ok {
		return Uint128{}, false
	}
	return Uint128FromBig(integer.BigInt())
}

func Int128FromValue(value Value) (Int128, bool) {
	integer, ok := wideIntegerValue(value)
	if !ok {
		return Int128{}, false
	}
	return Int128FromBig(integer.BigInt())
}

func wideIntegerValue(value Value) (IntegerValue, bool) {
	for {
		switch typed := value.(type) {
		case InterfaceValue:
			value = typed.Underlying
			continue
		case *InterfaceValue:
			if typed == nil {
				return IntegerValue{}, false
			}
			value = typed.Underlying
			continue
		case IntegerValue:
			return typed, true
		case *IntegerValue:
			if typed != nil {
				return *typed, true
			}
		}
		return IntegerValue{}, false
	}
}

func (value Uint128) IsZero() bool { return value.High == 0 && value.Low == 0 }

func (value Int128) IsZero() bool { return value.High == 0 && value.Low == 0 }

func (value Int128) IsNegative() bool { return value.High&(uint64(1)<<63) != 0 }

func (value Uint128) Compare(other Uint128) int {
	if value.High < other.High {
		return -1
	}
	if value.High > other.High {
		return 1
	}
	if value.Low < other.Low {
		return -1
	}
	if value.Low > other.Low {
		return 1
	}
	return 0
}

func (value Int128) Compare(other Int128) int {
	leftHigh := int64(value.High)
	rightHigh := int64(other.High)
	if leftHigh < rightHigh {
		return -1
	}
	if leftHigh > rightHigh {
		return 1
	}
	if value.Low < other.Low {
		return -1
	}
	if value.Low > other.Low {
		return 1
	}
	return 0
}

func (value Uint128) AddChecked(other Uint128) (Uint128, bool) {
	low, carry := bits.Add64(value.Low, other.Low, 0)
	high, overflow := bits.Add64(value.High, other.High, carry)
	return Uint128{High: high, Low: low}, overflow == 0
}

func (value Uint128) SubChecked(other Uint128) (Uint128, bool) {
	low, borrow := bits.Sub64(value.Low, other.Low, 0)
	high, underflow := bits.Sub64(value.High, other.High, borrow)
	return Uint128{High: high, Low: low}, underflow == 0
}

func (value Uint128) MulChecked(other Uint128) (Uint128, bool) {
	p00High, low := bits.Mul64(value.Low, other.Low)
	p01High, p01Low := bits.Mul64(value.Low, other.High)
	p10High, p10Low := bits.Mul64(value.High, other.Low)
	p11High, p11Low := bits.Mul64(value.High, other.High)
	high, carry1 := bits.Add64(p00High, p01Low, 0)
	high, carry2 := bits.Add64(high, p10Low, 0)
	upper, carry3 := bits.Add64(p01High, p10High, carry1)
	upper, carry4 := bits.Add64(upper, p11Low, carry2)
	overflow := p11High != 0 || carry3 != 0 || carry4 != 0 || upper != 0
	return Uint128{High: high, Low: low}, !overflow
}

func (value Int128) AddChecked(other Int128) (Int128, bool) {
	result := Int128FromBits(mustUintAdd(Uint128FromBits(value), Uint128FromBits(other)))
	overflow := value.IsNegative() == other.IsNegative() && result.IsNegative() != value.IsNegative()
	return result, !overflow
}

func (value Int128) SubChecked(other Int128) (Int128, bool) {
	result := Int128FromBits(mustUintSub(Uint128FromBits(value), Uint128FromBits(other)))
	overflow := value.IsNegative() != other.IsNegative() && result.IsNegative() != value.IsNegative()
	return result, !overflow
}

func (value Int128) NegateChecked() (Int128, bool) {
	if value.High == uint64(1)<<63 && value.Low == 0 {
		return Int128{}, false
	}
	pattern := Uint128FromBits(value)
	negated := mustUintAdd(Uint128{High: ^pattern.High, Low: ^pattern.Low}, Uint128{Low: 1})
	return Int128FromBits(negated), true
}

func (value Int128) magnitude() Uint128 {
	pattern := Uint128FromBits(value)
	if !value.IsNegative() {
		return pattern
	}
	return mustUintAdd(Uint128{High: ^pattern.High, Low: ^pattern.Low}, Uint128{Low: 1})
}

func int128FromMagnitude(value Uint128, negative bool) (Int128, bool) {
	limit := Uint128{High: uint64(1) << 63}
	if negative {
		if value.Compare(limit) > 0 {
			return Int128{}, false
		}
		pattern := mustUintAdd(Uint128{High: ^value.High, Low: ^value.Low}, Uint128{Low: 1})
		return Int128FromBits(pattern), true
	}
	if value.Compare(limit) >= 0 {
		return Int128{}, false
	}
	return Int128FromBits(value), true
}

func (value Int128) MulChecked(other Int128) (Int128, bool) {
	product, ok := value.magnitude().MulChecked(other.magnitude())
	if !ok {
		return Int128{}, false
	}
	return int128FromMagnitude(product, value.IsNegative() != other.IsNegative())
}

func (value Uint128) PowChecked(exponent Uint128) (Uint128, bool) {
	result := Uint128{Low: 1}
	current := value
	for !exponent.IsZero() {
		if exponent.Low&1 != 0 {
			var ok bool
			result, ok = result.MulChecked(current)
			if !ok {
				return Uint128{}, false
			}
		}
		exponent = exponent.ShiftRightUnchecked(1)
		if !exponent.IsZero() {
			var ok bool
			current, ok = current.MulChecked(current)
			if !ok {
				return Uint128{}, false
			}
		}
	}
	return result, true
}

// PowChecked returns status 1 for a negative exponent and status 2 for
// overflow. Status zero is success.
func (value Int128) PowChecked(exponent Int128) (Int128, int) {
	if exponent.IsNegative() {
		return Int128{}, 1
	}
	result := Int128{Low: 1}
	current := value
	exponentBits := Uint128FromBits(exponent)
	for !exponentBits.IsZero() {
		if exponentBits.Low&1 != 0 {
			var ok bool
			result, ok = result.MulChecked(current)
			if !ok {
				return Int128{}, 2
			}
		}
		exponentBits = exponentBits.ShiftRightUnchecked(1)
		if !exponentBits.IsZero() {
			var ok bool
			current, ok = current.MulChecked(current)
			if !ok {
				return Int128{}, 2
			}
		}
	}
	return result, 0
}

func mustUintAdd(left Uint128, right Uint128) Uint128 {
	low, carry := bits.Add64(left.Low, right.Low, 0)
	high, _ := bits.Add64(left.High, right.High, carry)
	return Uint128{High: high, Low: low}
}

func mustUintSub(left Uint128, right Uint128) Uint128 {
	low, borrow := bits.Sub64(left.Low, right.Low, 0)
	high, _ := bits.Sub64(left.High, right.High, borrow)
	return Uint128{High: high, Low: low}
}

func (value Uint128) DivMod(other Uint128) (Uint128, Uint128, bool) {
	if other.IsZero() {
		return Uint128{}, Uint128{}, false
	}
	if other.High == 0 {
		quotientHigh := value.High / other.Low
		remainderHigh := value.High % other.Low
		quotientLow, remainder := bits.Div64(remainderHigh, value.Low, other.Low)
		return Uint128{High: quotientHigh, Low: quotientLow}, Uint128{Low: remainder}, true
	}
	if value.Compare(other) < 0 {
		return Uint128{}, value, true
	}
	quotient := divideUint128ByWide(value, other)
	product, ok := other.MulChecked(Uint128{Low: quotient})
	if !ok || product.Compare(value) > 0 {
		// The normalized estimate can be one too large.
		quotient--
		product, _ = other.MulChecked(Uint128{Low: quotient})
	}
	return Uint128{Low: quotient}, mustUintSub(value, product), true
}

func divideUint128ByWide(dividend Uint128, divisor Uint128) uint64 {
	shift := uint(bits.LeadingZeros64(divisor.High))
	var divisorHigh, divisorLow uint64
	var dividendExtra, dividendHigh, dividendLow uint64
	if shift == 0 {
		divisorHigh, divisorLow = divisor.High, divisor.Low
		dividendHigh, dividendLow = dividend.High, dividend.Low
	} else {
		divisorHigh = divisor.High<<shift | divisor.Low>>(64-shift)
		divisorLow = divisor.Low << shift
		dividendExtra = dividend.High >> (64 - shift)
		dividendHigh = dividend.High<<shift | dividend.Low>>(64-shift)
		dividendLow = dividend.Low << shift
	}
	quotient, remainderHigh := bits.Div64(dividendExtra, dividendHigh, divisorHigh)
	for {
		productHigh, productLow := bits.Mul64(quotient, divisorLow)
		if productHigh < remainderHigh || (productHigh == remainderHigh && productLow <= dividendLow) {
			return quotient
		}
		quotient--
		var carry uint64
		remainderHigh, carry = bits.Add64(remainderHigh, divisorHigh, 0)
		if carry != 0 {
			return quotient
		}
	}
}

func (value Int128) DivMod(other Int128) (Int128, Int128, bool, bool) {
	if other.IsZero() {
		return Int128{}, Int128{}, false, false
	}
	quotientMagnitude, remainderMagnitude, _ := value.magnitude().DivMod(other.magnitude())
	quotient, ok := int128FromMagnitude(quotientMagnitude, value.IsNegative() != other.IsNegative())
	if !ok {
		return Int128{}, Int128{}, true, false
	}
	remainder, _ := int128FromMagnitude(remainderMagnitude, value.IsNegative() && !remainderMagnitude.IsZero())
	if remainder.IsNegative() {
		one := Int128{Low: 1}
		if other.IsNegative() {
			quotient, _ = quotient.AddChecked(one)
			remainder, _ = remainder.SubChecked(other)
		} else {
			quotient, _ = quotient.SubChecked(one)
			remainder, _ = remainder.AddChecked(other)
		}
	}
	return quotient, remainder, true, true
}

func (value Uint128) And(other Uint128) Uint128 {
	return Uint128{High: value.High & other.High, Low: value.Low & other.Low}
}

func (value Uint128) Or(other Uint128) Uint128 {
	return Uint128{High: value.High | other.High, Low: value.Low | other.Low}
}

func (value Uint128) Xor(other Uint128) Uint128 {
	return Uint128{High: value.High ^ other.High, Low: value.Low ^ other.Low}
}

func (value Uint128) Not() Uint128 { return Uint128{High: ^value.High, Low: ^value.Low} }

func (value Int128) And(other Int128) Int128 {
	return Int128{High: value.High & other.High, Low: value.Low & other.Low}
}

func (value Int128) Or(other Int128) Int128 {
	return Int128{High: value.High | other.High, Low: value.Low | other.Low}
}

func (value Int128) Xor(other Int128) Int128 {
	return Int128{High: value.High ^ other.High, Low: value.Low ^ other.Low}
}

func (value Int128) Not() Int128 { return Int128{High: ^value.High, Low: ^value.Low} }

func (value Uint128) shiftCount() (uint, bool) {
	if value.High != 0 || value.Low >= 128 {
		return 0, false
	}
	return uint(value.Low), true
}

func (value Int128) shiftCount() (uint, bool) {
	if value.IsNegative() || value.High != 0 || value.Low >= 128 {
		return 0, false
	}
	return uint(value.Low), true
}

func (value Uint128) ShiftLeftUnchecked(shift uint) Uint128 {
	switch {
	case shift == 0:
		return value
	case shift < 64:
		return Uint128{High: value.High<<shift | value.Low>>(64-shift), Low: value.Low << shift}
	case shift < 128:
		return Uint128{High: value.Low << (shift - 64)}
	default:
		return Uint128{}
	}
}

func (value Uint128) ShiftRightUnchecked(shift uint) Uint128 {
	switch {
	case shift == 0:
		return value
	case shift < 64:
		return Uint128{High: value.High >> shift, Low: value.Low>>shift | value.High<<(64-shift)}
	case shift < 128:
		return Uint128{Low: value.High >> (shift - 64)}
	default:
		return Uint128{}
	}
}

func (value Uint128) ShiftLeft(count Uint128) (Uint128, bool) {
	shift, ok := count.shiftCount()
	return value.ShiftLeftUnchecked(shift), ok
}

func (value Uint128) ShiftRight(count Uint128) (Uint128, bool) {
	shift, ok := count.shiftCount()
	return value.ShiftRightUnchecked(shift), ok
}

func (value Int128) ShiftLeft(count Int128) (Int128, bool) {
	shift, ok := count.shiftCount()
	return Int128FromBits(Uint128FromBits(value).ShiftLeftUnchecked(shift)), ok
}

func (value Int128) ShiftRight(count Int128) (Int128, bool) {
	shift, ok := count.shiftCount()
	if !ok {
		return Int128{}, false
	}
	if shift == 0 {
		return value, true
	}
	unsigned := Uint128FromBits(value).ShiftRightUnchecked(shift)
	if value.IsNegative() {
		if shift < 64 {
			unsigned.High |= ^uint64(0) << (64 - shift)
		} else {
			unsigned.High = ^uint64(0)
			unsigned.Low |= ^uint64(0) << (128 - shift)
		}
	}
	return Int128FromBits(unsigned), true
}

func (value Uint128) Float64() float64 {
	return float64(value.High)*18446744073709551616.0 + float64(value.Low)
}

func (value Int128) Float64() float64 {
	if !value.IsNegative() {
		return Uint128FromBits(value).Float64()
	}
	return -value.magnitude().Float64()
}
