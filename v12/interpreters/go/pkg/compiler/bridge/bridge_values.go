package bridge

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/runtime"
)

func AsString(value runtime.Value) (string, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val, nil
	case *runtime.StringValue:
		if v == nil {
			return "", fmt.Errorf("expected String, got nil")
		}
		return v.Val, nil
	case *runtime.StructInstanceValue:
		return stringFromStruct(v)
	default:
		return "", fmt.Errorf("expected String, got %T", value)
	}
}

func stringFromStruct(inst *runtime.StructInstanceValue) (string, error) {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return "", fmt.Errorf("expected String, got nil")
	}
	if inst.Definition.Node.ID.Name != "String" {
		return "", fmt.Errorf("expected String, got %T", inst)
	}
	var bytesVal runtime.Value
	if inst.Fields != nil {
		if field, ok := inst.Fields["bytes"]; ok {
			bytesVal = field
		}
	}
	if bytesVal == nil && len(inst.Positional) > 0 {
		bytesVal = inst.Positional[0]
	}
	if bytesVal == nil {
		return "", fmt.Errorf("string bytes are missing")
	}
	arr, err := arrayValueFromRuntime(bytesVal)
	if err != nil {
		return "", err
	}
	if arr == nil {
		return "", fmt.Errorf("string bytes are missing")
	}
	bytes := make([]byte, len(arr.Elements))
	maxByte := big.NewInt(0xff)
	for idx, elem := range arr.Elements {
		intVal, err := extractInteger(elem)
		if err != nil {
			return "", err
		}
		if intVal.Sign() < 0 || intVal.Cmp(maxByte) > 0 {
			return "", fmt.Errorf("string byte out of range")
		}
		bytes[idx] = byte(intVal.Int64())
	}
	return string(bytes), nil
}

func ToString(value string) runtime.Value {
	return runtime.StringValue{Val: value}
}

func AsBool(value runtime.Value) (bool, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.BoolValue:
		return v.Val, nil
	case *runtime.BoolValue:
		if v == nil {
			return false, fmt.Errorf("expected bool, got nil")
		}
		return v.Val, nil
	default:
		return false, fmt.Errorf("expected bool, got %T", value)
	}
}

func ToBool(value bool) runtime.Value {
	return runtime.BoolValue{Val: value}
}

func AsRune(value runtime.Value) (rune, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.CharValue:
		return v.Val, nil
	case *runtime.CharValue:
		if v == nil {
			return 0, fmt.Errorf("expected char, got nil")
		}
		return v.Val, nil
	default:
		return 0, fmt.Errorf("expected char, got %T", value)
	}
}

func ToRune(value rune) runtime.Value {
	return runtime.CharValue{Val: value}
}

func AsFloat(value runtime.Value) (float64, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.FloatValue:
		return v.Val, nil
	case *runtime.FloatValue:
		if v == nil {
			return 0, fmt.Errorf("expected float, got nil")
		}
		return v.Val, nil
	case runtime.IntegerValue:
		if n, ok := v.ToInt64(); ok {
			return float64(n), nil
		}
		return bigIntToFloat(v.BigInt()), nil
	case *runtime.IntegerValue:
		if v == nil {
			return 0, fmt.Errorf("expected float, got nil")
		}
		if n, ok := v.ToInt64(); ok {
			return float64(n), nil
		}
		return bigIntToFloat(v.BigInt()), nil
	default:
		return 0, fmt.Errorf("expected float, got %T", value)
	}
}

func ToFloat64(value float64) runtime.Value {
	return runtime.FloatValue{Val: value, TypeSuffix: runtime.FloatF64}
}

func ToFloat32(value float32) runtime.Value {
	return runtime.FloatValue{Val: float64(value), TypeSuffix: runtime.FloatF32}
}

func AsInt(value runtime.Value, bits int) (int64, error) {
	// Fast path: avoid big.Int allocation for small ints (hot in compiled loops).
	if iv, ok := unwrapInterface(value).(runtime.IntegerValue); ok {
		if n, fit := iv.ToInt64(); fit {
			if bits > 0 && bits <= 64 {
				lo, hi := signedRangeInt64(bits)
				if n < lo || n > hi {
					return 0, fmt.Errorf("integer %d overflows %d-bit signed", n, bits)
				}
			}
			return n, nil
		}
	}
	val, err := extractInteger(value)
	if err != nil {
		return 0, err
	}
	min, max := signedRange(bits)
	if val.Cmp(min) < 0 || val.Cmp(max) > 0 {
		return 0, fmt.Errorf("integer %s overflows %d-bit signed", val.String(), bits)
	}
	return val.Int64(), nil
}

func AsUint(value runtime.Value, bits int) (uint64, error) {
	// Fast path: avoid big.Int allocation for small ints (hot in compiled loops).
	if iv, ok := unwrapInterface(value).(runtime.IntegerValue); ok {
		if n, fit := iv.ToInt64(); fit {
			if n < 0 {
				return 0, fmt.Errorf("integer %d is negative for unsigned", n)
			}
			if bits > 0 && bits <= 64 {
				_, hi := unsignedRangeUint64(bits)
				if uint64(n) > hi {
					return 0, fmt.Errorf("integer %d overflows %d-bit unsigned", n, bits)
				}
			}
			return uint64(n), nil
		}
	}
	val, err := extractInteger(value)
	if err != nil {
		return 0, err
	}
	if val.Sign() < 0 {
		return 0, fmt.Errorf("integer %s is negative for unsigned", val.String())
	}
	_, max := unsignedRange(bits)
	if val.Cmp(max) > 0 {
		return 0, fmt.Errorf("integer %s overflows %d-bit unsigned", val.String(), bits)
	}
	return val.Uint64(), nil
}

func ToInt(value int64, suffix runtime.IntegerType) runtime.Value {
	return runtime.NewSmallInt(value, suffix)
}

func ToUint(value uint64, suffix runtime.IntegerType) runtime.Value {
	if value <= math.MaxInt64 {
		return runtime.NewSmallInt(int64(value), suffix)
	}
	val := new(big.Int)
	val.SetUint64(value)
	return runtime.NewBigIntValue(val, suffix)
}

func unwrapInterface(value runtime.Value) runtime.Value {
	for {
		switch v := value.(type) {
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.InterfaceValue:
			if v != nil {
				value = v.Underlying
				continue
			}
		}
		break
	}
	return value
}

func arrayValueFromRuntime(value runtime.Value) (*runtime.ArrayValue, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case *runtime.ArrayValue:
		if v == nil {
			return nil, fmt.Errorf("string bytes are missing")
		}
		return v, nil
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil || v.Definition.Node.ID.Name != "Array" {
			return nil, fmt.Errorf("string bytes must be an array (got %T)", value)
		}
		var handleVal runtime.Value
		if v.Fields != nil {
			handleVal = v.Fields["storage_handle"]
		}
		if handleVal == nil && len(v.Positional) >= 3 {
			handleVal = v.Positional[2]
		}
		if handleVal == nil {
			return nil, fmt.Errorf("array value missing storage_handle")
		}
		handleInt, err := extractInteger(handleVal)
		if err != nil {
			return nil, err
		}
		if !handleInt.IsInt64() {
			return nil, fmt.Errorf("array handle is out of range")
		}
		handle := handleInt.Int64()
		arr, _, err := runtime.ArrayStoreValueFromHandle(handle, 0, 0)
		if err != nil {
			return nil, err
		}
		return arr, nil
	default:
		return nil, fmt.Errorf("string bytes must be an array (got %T)", value)
	}
}

func bigIntToFloat(val *big.Int) float64 {
	if val == nil {
		return 0
	}
	f := new(big.Float).SetInt(val)
	result, _ := f.Float64()
	return result
}

func extractInteger(value runtime.Value) (*big.Int, error) {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case runtime.IntegerValue:
		return v.BigInt(), nil
	case *runtime.IntegerValue:
		if v == nil {
			return nil, fmt.Errorf("expected integer, got nil")
		}
		return v.BigInt(), nil
	default:
		return nil, fmt.Errorf("expected integer, got %T", value)
	}
}

func signedRangeInt64(bits int) (int64, int64) {
	switch bits {
	case 8:
		return -128, 127
	case 16:
		return -32768, 32767
	case 32:
		return -2147483648, 2147483647
	default:
		return math.MinInt64, math.MaxInt64
	}
}

func unsignedRangeUint64(bits int) (uint64, uint64) {
	switch bits {
	case 8:
		return 0, 255
	case 16:
		return 0, 65535
	case 32:
		return 0, 4294967295
	default:
		return 0, math.MaxUint64
	}
}

func signedRange(bits int) (*big.Int, *big.Int) {
	if bits <= 0 || bits > 64 {
		bits = 64
	}
	one := big.NewInt(1)
	max := new(big.Int).Lsh(one, uint(bits-1))
	max.Sub(max, one)
	min := new(big.Int).Neg(new(big.Int).Lsh(one, uint(bits-1)))
	return min, max
}

func unsignedRange(bits int) (*big.Int, *big.Int) {
	if bits <= 0 || bits > 64 {
		bits = 64
	}
	one := big.NewInt(1)
	max := new(big.Int).Lsh(one, uint(bits))
	max.Sub(max, one)
	return big.NewInt(0), max
}
