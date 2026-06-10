package interpreter

import "able/interpreter-go/pkg/runtime"

type primitiveMethodCacheKey uint8

type bytecodeMemberPrimitiveCacheKey = primitiveMethodCacheKey
type boundMethodPrimitiveCacheKey = primitiveMethodCacheKey

const (
	primitiveMethodCacheKeyInvalid primitiveMethodCacheKey = iota
	primitiveMethodCacheKeyString
	primitiveMethodCacheKeyBool
	primitiveMethodCacheKeyChar
	primitiveMethodCacheKeyNil
	primitiveMethodCacheKeyIntI8
	primitiveMethodCacheKeyIntI16
	primitiveMethodCacheKeyIntI32
	primitiveMethodCacheKeyIntI64
	primitiveMethodCacheKeyIntI128
	primitiveMethodCacheKeyIntU8
	primitiveMethodCacheKeyIntU16
	primitiveMethodCacheKeyIntU32
	primitiveMethodCacheKeyIntU64
	primitiveMethodCacheKeyIntU128
	primitiveMethodCacheKeyIntIsize
	primitiveMethodCacheKeyIntUsize
	primitiveMethodCacheKeyFloatF32
	primitiveMethodCacheKeyFloatF64
)

func primitiveMethodCacheKeyForIntegerSuffix(suffix runtime.IntegerType) (primitiveMethodCacheKey, bool) {
	switch suffix {
	case runtime.IntegerI8:
		return primitiveMethodCacheKeyIntI8, true
	case runtime.IntegerI16:
		return primitiveMethodCacheKeyIntI16, true
	case runtime.IntegerI32:
		return primitiveMethodCacheKeyIntI32, true
	case runtime.IntegerI64:
		return primitiveMethodCacheKeyIntI64, true
	case runtime.IntegerI128:
		return primitiveMethodCacheKeyIntI128, true
	case runtime.IntegerU8:
		return primitiveMethodCacheKeyIntU8, true
	case runtime.IntegerU16:
		return primitiveMethodCacheKeyIntU16, true
	case runtime.IntegerU32:
		return primitiveMethodCacheKeyIntU32, true
	case runtime.IntegerU64:
		return primitiveMethodCacheKeyIntU64, true
	case runtime.IntegerU128:
		return primitiveMethodCacheKeyIntU128, true
	case runtime.IntegerIsize:
		return primitiveMethodCacheKeyIntIsize, true
	case runtime.IntegerUsize:
		return primitiveMethodCacheKeyIntUsize, true
	default:
		return primitiveMethodCacheKeyInvalid, false
	}
}

func primitiveMethodCacheKeyForFloatSuffix(suffix runtime.FloatType) (primitiveMethodCacheKey, bool) {
	switch suffix {
	case runtime.FloatF32:
		return primitiveMethodCacheKeyFloatF32, true
	case runtime.FloatF64:
		return primitiveMethodCacheKeyFloatF64, true
	default:
		return primitiveMethodCacheKeyInvalid, false
	}
}
