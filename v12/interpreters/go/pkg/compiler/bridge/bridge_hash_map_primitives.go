package bridge

import (
	"encoding/binary"
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func primitiveHashMapKeyEqual(left, right runtime.Value) (bool, bool) {
	if leftString, ok := hashMapStringValue(left); ok {
		rightString, ok := hashMapStringValue(right)
		if !ok {
			return false, false
		}
		return leftString == rightString, true
	}
	if leftBool, ok := hashMapBoolValue(left); ok {
		rightBool, ok := hashMapBoolValue(right)
		if !ok {
			return false, false
		}
		return leftBool == rightBool, true
	}
	if leftChar, ok := hashMapCharValue(left); ok {
		rightChar, ok := hashMapCharValue(right)
		if !ok {
			return false, false
		}
		return leftChar == rightChar, true
	}
	leftInt, ok := hashMapIntegerValue(left)
	if !ok {
		return false, false
	}
	rightInt, ok := hashMapIntegerValue(right)
	if !ok || leftInt.TypeSuffix != rightInt.TypeSuffix {
		return false, false
	}
	return leftInt.CmpInt(rightInt) == 0, true
}

func primitiveHashMapHash(val runtime.Value) (uint64, bool, error) {
	hasher := runtime.NewHasherValue()
	if str, ok := hashMapStringValue(val); ok {
		hasher.WriteString(str)
		return hasher.Finish(), true, nil
	}
	if b, ok := hashMapBoolValue(val); ok {
		hasher.WriteBool(b)
		return hasher.Finish(), true, nil
	}
	if ch, ok := hashMapCharValue(val); ok {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(ch))
		hasher.WriteBytes(buf[:])
		return hasher.Finish(), true, nil
	}
	intVal, ok := hashMapIntegerValue(val)
	if !ok {
		return 0, false, nil
	}
	switch intVal.TypeSuffix {
	case runtime.IntegerI8, runtime.IntegerI16, runtime.IntegerI32, runtime.IntegerI64, runtime.IntegerIsize:
		n, err := integerToInt64Value(intVal)
		if err != nil {
			return 0, true, err
		}
		switch intVal.TypeSuffix {
		case runtime.IntegerI8:
			hasher.WriteBytes([]byte{byte(int8(n))})
		case runtime.IntegerI16:
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(int16(n)))
			hasher.WriteBytes(buf[:])
		case runtime.IntegerI32:
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(int32(n)))
			hasher.WriteBytes(buf[:])
		default:
			hasher.WriteInt64(n)
		}
		return hasher.Finish(), true, nil
	case runtime.IntegerU8, runtime.IntegerU16, runtime.IntegerU32, runtime.IntegerU64, runtime.IntegerUsize:
		n, err := integerToUint64Value(intVal)
		if err != nil {
			return 0, true, err
		}
		switch intVal.TypeSuffix {
		case runtime.IntegerU8:
			hasher.WriteBytes([]byte{byte(n)})
		case runtime.IntegerU16:
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(n))
			hasher.WriteBytes(buf[:])
		case runtime.IntegerU32:
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(n))
			hasher.WriteBytes(buf[:])
		default:
			hasher.WriteUint64(n)
		}
		return hasher.Finish(), true, nil
	default:
		return 0, false, nil
	}
}

func hashMapStringValue(val runtime.Value) (string, bool) {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val, true
	case *runtime.StringValue:
		if v == nil {
			return "", false
		}
		return v.Val, true
	default:
		return "", false
	}
}

func hashMapBoolValue(val runtime.Value) (bool, bool) {
	switch v := val.(type) {
	case runtime.BoolValue:
		return v.Val, true
	case *runtime.BoolValue:
		if v == nil {
			return false, false
		}
		return v.Val, true
	default:
		return false, false
	}
}

func hashMapCharValue(val runtime.Value) (rune, bool) {
	switch v := val.(type) {
	case runtime.CharValue:
		return v.Val, true
	case *runtime.CharValue:
		if v == nil {
			return 0, false
		}
		return v.Val, true
	default:
		return 0, false
	}
}

func hashMapIntegerValue(val runtime.Value) (runtime.IntegerValue, bool) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		return v, true
	case *runtime.IntegerValue:
		if v == nil {
			return runtime.IntegerValue{}, false
		}
		return *v, true
	default:
		return runtime.IntegerValue{}, false
	}
}

func integerToUint64Value(val runtime.IntegerValue) (uint64, error) {
	if n, ok := val.ToInt64(); ok {
		if n < 0 {
			return 0, fmt.Errorf("expected non-negative integer")
		}
		return uint64(n), nil
	}
	if val.Sign() < 0 {
		return 0, fmt.Errorf("expected non-negative integer")
	}
	bi := val.BigInt()
	if !bi.IsUint64() {
		return 0, fmt.Errorf("integer out of range for u64")
	}
	return bi.Uint64(), nil
}

func integerToInt64Value(val runtime.IntegerValue) (int64, error) {
	if n, ok := val.ToInt64(); ok {
		return n, nil
	}
	return 0, fmt.Errorf("integer out of range for i64")
}
