package runtime

import "math/big"

type RawValueKind uint8

const (
	RawValueMaterialized RawValueKind = iota
	RawValueInteger
	RawValueFloat
)

type RawValue struct {
	value       Value
	integerKind IntegerType
	integerRaw  int64
	floatKind   FloatType
	floatRaw    float64
	kind        RawValueKind
}

func NewRawValue(value Value) RawValue {
	return RawValue{value: value}
}

func NewRawIntegerValue(kind IntegerType, raw int64) RawValue {
	return RawValue{kind: RawValueInteger, integerKind: kind, integerRaw: raw}
}

func NewRawFloatValue(kind FloatType, raw float64) RawValue {
	return RawValue{kind: RawValueFloat, floatKind: kind, floatRaw: raw}
}

func (v RawValue) Kind() RawValueKind {
	return v.kind
}

func (v RawValue) Value() Value {
	return v.value
}

func (v RawValue) Integer() (IntegerType, int64, bool) {
	if v.kind != RawValueInteger {
		return "", 0, false
	}
	return v.integerKind, v.integerRaw, true
}

func (v RawValue) Float() (FloatType, float64, bool) {
	if v.kind != RawValueFloat {
		return "", 0, false
	}
	return v.floatKind, v.floatRaw, true
}

func (v RawValue) Materialize() Value {
	if v.value != nil {
		return v.value
	}
	switch v.kind {
	case RawValueInteger:
		switch v.integerKind {
		case IntegerU64, IntegerU128, IntegerUsize:
			if v.integerRaw < 0 {
				return NewBigIntValue(new(big.Int).SetUint64(uint64(v.integerRaw)), v.integerKind)
			}
		}
		return NewSmallInt(v.integerRaw, v.integerKind)
	case RawValueFloat:
		return FloatValue{Val: v.floatRaw, TypeSuffix: v.floatKind}
	default:
		return nil
	}
}
