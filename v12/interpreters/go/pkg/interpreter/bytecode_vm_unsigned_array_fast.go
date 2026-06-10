package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeU32ValueFast(value runtime.Value) (uint32, bool) {
	intVal, ok := bytecodeIntegerValue(value)
	if !ok || intVal.TypeSuffix != runtime.IntegerU32 {
		return 0, false
	}
	raw, ok := intVal.ToInt64()
	if !ok || raw < 0 || raw > int64(^uint32(0)) {
		return 0, false
	}
	return uint32(raw), true
}

func bytecodeU64ValueFast(value runtime.Value) (uint64, bool) {
	intVal, ok := bytecodeIntegerValue(value)
	if !ok || intVal.TypeSuffix != runtime.IntegerU64 || intVal.Sign() < 0 {
		return 0, false
	}
	if intVal.IsSmall() {
		return uint64(intVal.Int64Fast()), true
	}
	raw := intVal.BigInt()
	if !raw.IsUint64() {
		return 0, false
	}
	return raw.Uint64(), true
}

func bytecodeReadMonoUnsignedArrayValue(handle int64, idx int) (runtime.Value, runtime.IntegerType, bool, error) {
	var info runtime.ArrayStoreMonoPrimitiveReadInfo
	ok, err := runtime.ArrayStoreMonoPrimitiveReadInfoInto(handle, idx, &info)
	if err != nil {
		return nil, "", true, err
	}
	if !ok || !info.InBounds {
		return nil, "", false, nil
	}
	switch info.Kind {
	case runtime.ArrayStoreMonoPrimitiveReadU32:
		return bytecodeRawIntegerResultValue(runtime.IntegerU32, int64(uint32(info.Uint64))), runtime.IntegerU32, true, nil
	case runtime.ArrayStoreMonoPrimitiveReadU64:
		return bytecodeUnsignedIntegerValue(runtime.IntegerU64, info.Uint64), runtime.IntegerU64, true, nil
	default:
		return nil, "", false, nil
	}
}
