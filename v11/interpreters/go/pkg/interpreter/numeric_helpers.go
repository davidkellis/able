package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/runtime"
)

type integerInfo struct {
	kind   runtime.IntegerType
	bits   int
	signed bool
	min    *big.Int
	max    *big.Int
	mask   *big.Int
}

func signedInfo(kind runtime.IntegerType, bits int) integerInfo {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)), big.NewInt(1))
	min := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)))
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	return integerInfo{kind: kind, bits: bits, signed: true, min: min, max: max, mask: mask}
}

func unsignedInfo(kind runtime.IntegerType, bits int) integerInfo {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	mask := new(big.Int).Set(max)
	return integerInfo{kind: kind, bits: bits, signed: false, min: big.NewInt(0), max: max, mask: mask}
}

var integerInfos = map[runtime.IntegerType]integerInfo{
	runtime.IntegerI8:   signedInfo(runtime.IntegerI8, 8),
	runtime.IntegerI16:  signedInfo(runtime.IntegerI16, 16),
	runtime.IntegerI32:  signedInfo(runtime.IntegerI32, 32),
	runtime.IntegerI64:  signedInfo(runtime.IntegerI64, 64),
	runtime.IntegerI128: signedInfo(runtime.IntegerI128, 128),
	runtime.IntegerU8:   unsignedInfo(runtime.IntegerU8, 8),
	runtime.IntegerU16:  unsignedInfo(runtime.IntegerU16, 16),
	runtime.IntegerU32:  unsignedInfo(runtime.IntegerU32, 32),
	runtime.IntegerU64:  unsignedInfo(runtime.IntegerU64, 64),
	runtime.IntegerU128: unsignedInfo(runtime.IntegerU128, 128),
}

var signedSequence = []runtime.IntegerType{
	runtime.IntegerI8,
	runtime.IntegerI16,
	runtime.IntegerI32,
	runtime.IntegerI64,
	runtime.IntegerI128,
}

var unsignedSequence = []runtime.IntegerType{
	runtime.IntegerU8,
	runtime.IntegerU16,
	runtime.IntegerU32,
	runtime.IntegerU64,
	runtime.IntegerU128,
}

func getIntegerInfo(kind runtime.IntegerType) (integerInfo, error) {
	info, ok := integerInfos[kind]
	if !ok {
		return integerInfo{}, fmt.Errorf("unsupported integer kind %s", kind)
	}
	return info, nil
}

func integerRangeWithinKinds(source runtime.IntegerType, target runtime.IntegerType) bool {
	sourceInfo, err := getIntegerInfo(source)
	if err != nil {
		return false
	}
	targetInfo, err := getIntegerInfo(target)
	if err != nil {
		return false
	}
	return sourceInfo.min.Cmp(targetInfo.min) >= 0 && sourceInfo.max.Cmp(targetInfo.max) <= 0
}

func integerValueWithinRange(val *big.Int, target runtime.IntegerType) bool {
	if val == nil {
		return false
	}
	info, err := getIntegerInfo(target)
	if err != nil {
		return false
	}
	return val.Cmp(info.min) >= 0 && val.Cmp(info.max) <= 0
}

func ensureFitsInteger(info integerInfo, value *big.Int) error {
	if value.Cmp(info.min) < 0 || value.Cmp(info.max) > 0 {
		return fmt.Errorf("integer overflow")
	}
	return nil
}

func smallestSigned(bits int) (integerInfo, bool) {
	for _, kind := range signedSequence {
		info := integerInfos[kind]
		if info.bits >= bits {
			return info, true
		}
	}
	return integerInfo{}, false
}

func smallestUnsigned(bits int) (integerInfo, bool) {
	for _, kind := range unsignedSequence {
		info := integerInfos[kind]
		if info.bits >= bits {
			return info, true
		}
	}
	return integerInfo{}, false
}

func widestUnsignedInfo(a integerInfo, b integerInfo) (integerInfo, bool) {
	candidates := []integerInfo{}
	if !a.signed {
		candidates = append(candidates, a)
	}
	if !b.signed {
		candidates = append(candidates, b)
	}
	if len(candidates) == 0 {
		return integerInfo{}, false
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.bits > best.bits {
			best = candidate
		}
	}
	return best, true
}

func promoteIntegerTypes(left runtime.IntegerType, right runtime.IntegerType) (runtime.IntegerType, error) {
	leftInfo, err := getIntegerInfo(left)
	if err != nil {
		return runtime.IntegerI32, err
	}
	rightInfo, err := getIntegerInfo(right)
	if err != nil {
		return runtime.IntegerI32, err
	}
	if leftInfo.signed == rightInfo.signed {
		targetBits := leftInfo.bits
		if rightInfo.bits > targetBits {
			targetBits = rightInfo.bits
		}
		if leftInfo.signed {
			if info, ok := smallestSigned(targetBits); ok {
				return info.kind, nil
			}
		} else {
			if info, ok := smallestUnsigned(targetBits); ok {
				return info.kind, nil
			}
		}
		return runtime.IntegerI32, fmt.Errorf("integer operands exceed supported widths")
	}
	bitsNeeded := leftInfo.bits + 1
	if rightInfo.bits+1 > bitsNeeded {
		bitsNeeded = rightInfo.bits + 1
	}
	if info, ok := smallestSigned(bitsNeeded); ok {
		return info.kind, nil
	}
	if info, ok := widestUnsignedInfo(leftInfo, rightInfo); ok && info.bits >= leftInfo.bits && info.bits >= rightInfo.bits {
		return info.kind, nil
	}
	return runtime.IntegerI32, fmt.Errorf("integer operands exceed supported widths")
}

func bitPattern(value *big.Int, info integerInfo) *big.Int {
	var pattern big.Int
	pattern.And(value, info.mask)
	return &pattern
}

func patternToInteger(pattern *big.Int, info integerInfo) *big.Int {
	var masked big.Int
	masked.And(pattern, info.mask)
	if !info.signed {
		return &masked
	}
	signBit := uint(info.bits - 1)
	if masked.Bit(int(signBit)) == 1 {
		var adjust big.Int
		adjust.Lsh(big.NewInt(1), uint(info.bits))
		masked.Sub(&masked, &adjust)
	}
	return &masked
}

func shiftValueLeft(value *big.Int, shiftCount int, info integerInfo) (*big.Int, error) {
	if shiftCount < 0 || shiftCount >= info.bits {
		return nil, fmt.Errorf("shift out of range")
	}
	var result big.Int
	result.Lsh(value, uint(shiftCount))
	return &result, ensureFitsInteger(info, &result)
}

func shiftValueRight(value *big.Int, shiftCount int, info integerInfo) (*big.Int, error) {
	if shiftCount < 0 || shiftCount >= info.bits {
		return nil, fmt.Errorf("shift out of range")
	}
	var result big.Int
	if info.signed {
		result.Rsh(value, uint(shiftCount))
	} else {
		pattern := bitPattern(value, info)
		result.Rsh(pattern, uint(shiftCount))
	}
	return &result, ensureFitsInteger(info, &result)
}

func bigIntToFloat(val *big.Int) float64 {
	f := new(big.Float).SetInt(val)
	result, _ := f.Float64()
	return result
}

func normalizeFloat(kind runtime.FloatType, value float64) float64 {
	if kind == runtime.FloatF32 {
		return float64(float32(value))
	}
	return value
}

func floatResultKind(left runtime.Value, right runtime.Value) runtime.FloatType {
	leftVal, leftIsFloat := left.(runtime.FloatValue)
	rightVal, rightIsFloat := right.(runtime.FloatValue)
	if (leftIsFloat && leftVal.TypeSuffix == runtime.FloatF64) || (rightIsFloat && rightVal.TypeSuffix == runtime.FloatF64) {
		return runtime.FloatF64
	}
	return runtime.FloatF32
}

func floatEquals(a, b runtime.Value) bool {
	left, ok := a.(runtime.FloatValue)
	if !ok {
		return false
	}
	right, ok := b.(runtime.FloatValue)
	if !ok {
		return false
	}
	diff := math.Abs(left.Val - right.Val)
	return diff < 1e-9
}
