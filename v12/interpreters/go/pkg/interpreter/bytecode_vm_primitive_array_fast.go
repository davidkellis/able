package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeMonoPrimitiveArrayValue(info runtime.ArrayStoreMonoPrimitiveReadInfo) (runtime.Value, bool) {
	if !info.InBounds {
		return nil, false
	}
	switch info.Kind {
	case runtime.ArrayStoreMonoPrimitiveReadI32:
		return bytecodeRawI32ResultValue(info.Int64), true
	case runtime.ArrayStoreMonoPrimitiveReadI64:
		return runtime.NewSmallInt(info.Int64, runtime.IntegerI64), true
	case runtime.ArrayStoreMonoPrimitiveReadBool:
		return runtime.BoolValue{Val: info.Bool}, true
	case runtime.ArrayStoreMonoPrimitiveReadChar:
		return runtime.CharValue{Val: rune(info.Int64)}, true
	case runtime.ArrayStoreMonoPrimitiveReadU8:
		return bytecodeRawU8ResultValue(info.Uint64), true
	case runtime.ArrayStoreMonoPrimitiveReadU32:
		return bytecodeRawIntegerResultValue(runtime.IntegerU32, int64(uint32(info.Uint64))), true
	case runtime.ArrayStoreMonoPrimitiveReadU64:
		return bytecodeUnsignedIntegerValue(runtime.IntegerU64, info.Uint64), true
	case runtime.ArrayStoreMonoPrimitiveReadF64:
		return runtime.FloatValue{Val: info.Float64, TypeSuffix: runtime.FloatF64}, true
	default:
		return nil, false
	}
}

var bytecodeMonoPrimitiveArrayTokenTable = [...]uint16{
	runtime.ArrayStoreMonoPrimitiveReadI32:  bytecodeIndexTypeI32,
	runtime.ArrayStoreMonoPrimitiveReadI64:  bytecodeIndexTypeI64,
	runtime.ArrayStoreMonoPrimitiveReadBool: bytecodeIndexTypeBool,
	runtime.ArrayStoreMonoPrimitiveReadChar: bytecodeIndexTypeChar,
	runtime.ArrayStoreMonoPrimitiveReadU8:   bytecodeIndexTypeU8,
	runtime.ArrayStoreMonoPrimitiveReadU32:  bytecodeIndexTypeU32,
	runtime.ArrayStoreMonoPrimitiveReadU64:  bytecodeIndexTypeU64,
	runtime.ArrayStoreMonoPrimitiveReadF64:  bytecodeIndexTypeF64,
}

func bytecodeMonoPrimitiveArrayToken(kind runtime.ArrayStoreMonoPrimitiveReadKind) (uint16, bool) {
	if int(kind) < len(bytecodeMonoPrimitiveArrayTokenTable) {
		token := bytecodeMonoPrimitiveArrayTokenTable[kind]
		return token, token != bytecodeIndexTypeUnknown
	}
	return bytecodeIndexTypeUnknown, false
}

func bytecodeMonoPrimitiveArrayDispatch(info runtime.ArrayStoreMonoPrimitiveReadInfo, prefix string) string {
	switch prefix {
	case "array_get":
		switch info.Kind {
		case runtime.ArrayStoreMonoPrimitiveReadI32:
			return "array_get_mono_i32_fast"
		case runtime.ArrayStoreMonoPrimitiveReadI64:
			return "array_get_mono_i64_fast"
		case runtime.ArrayStoreMonoPrimitiveReadBool:
			return "array_get_mono_bool_fast"
		case runtime.ArrayStoreMonoPrimitiveReadChar:
			return "array_get_mono_char_fast"
		case runtime.ArrayStoreMonoPrimitiveReadU8:
			return "array_get_mono_u8_fast"
		case runtime.ArrayStoreMonoPrimitiveReadU32:
			return "array_get_mono_u32_fast"
		case runtime.ArrayStoreMonoPrimitiveReadU64:
			return "array_get_mono_u64_fast"
		case runtime.ArrayStoreMonoPrimitiveReadF64:
			return "array_get_mono_f64_fast"
		default:
			return "array_get_fast"
		}
	case "array_read_slot":
		switch info.Kind {
		case runtime.ArrayStoreMonoPrimitiveReadI32:
			return "array_read_slot_mono_i32_fast"
		case runtime.ArrayStoreMonoPrimitiveReadI64:
			return "array_read_slot_mono_i64_fast"
		case runtime.ArrayStoreMonoPrimitiveReadBool:
			return "array_read_slot_mono_bool_fast"
		case runtime.ArrayStoreMonoPrimitiveReadChar:
			return "array_read_slot_mono_char_fast"
		case runtime.ArrayStoreMonoPrimitiveReadU8:
			return "array_read_slot_mono_u8_fast"
		case runtime.ArrayStoreMonoPrimitiveReadU32:
			return "array_read_slot_mono_u32_fast"
		case runtime.ArrayStoreMonoPrimitiveReadU64:
			return "array_read_slot_mono_u64_fast"
		case runtime.ArrayStoreMonoPrimitiveReadF64:
			return "array_read_slot_mono_f64_fast"
		default:
			return "array_read_slot_fast"
		}
	}
	switch info.Kind {
	case runtime.ArrayStoreMonoPrimitiveReadI32:
		return prefix + "_mono_i32_fast"
	case runtime.ArrayStoreMonoPrimitiveReadI64:
		return prefix + "_mono_i64_fast"
	case runtime.ArrayStoreMonoPrimitiveReadBool:
		return prefix + "_mono_bool_fast"
	case runtime.ArrayStoreMonoPrimitiveReadChar:
		return prefix + "_mono_char_fast"
	case runtime.ArrayStoreMonoPrimitiveReadU8:
		return prefix + "_mono_u8_fast"
	case runtime.ArrayStoreMonoPrimitiveReadU32:
		return prefix + "_mono_u32_fast"
	case runtime.ArrayStoreMonoPrimitiveReadU64:
		return prefix + "_mono_u64_fast"
	case runtime.ArrayStoreMonoPrimitiveReadF64:
		return prefix + "_mono_f64_fast"
	default:
		return prefix + "_fast"
	}
}
