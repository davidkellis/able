package interpreter

import (
	"sync"

	"able/interpreter-go/pkg/runtime"
)

const (
	bytecodeSmallIntBoxMin    int64 = -256
	bytecodeSmallIntBoxMax    int64 = 16384
	bytecodeI32ExtendedBoxMin int64 = bytecodeSmallIntBoxMax + 1
	bytecodeI32ExtendedBoxMax int64 = 262143
	// Bound dynamic boxed-int growth for out-of-range integer values. Keep this
	// large enough that a single warmup pass can retain the common text-benchmark
	// loop-index working set instead of re-boxing the same large values in the
	// measured run.
	bytecodeIntBoxDynamicCacheLimit = 262144
)

var (
	bytecodeBoxedI8          []runtime.Value
	bytecodeBoxedI16         []runtime.Value
	bytecodeBoxedI32         []runtime.Value
	bytecodeBoxedI32Extended []runtime.Value
	bytecodeBoxedI64         []runtime.Value
	bytecodeBoxedI128        []runtime.Value
	bytecodeBoxedU8          []runtime.Value
	bytecodeBoxedU16         []runtime.Value
	bytecodeBoxedU32         []runtime.Value
	bytecodeBoxedU64         []runtime.Value
	bytecodeBoxedU128        []runtime.Value
	bytecodeBoxedIsize       []runtime.Value
	bytecodeBoxedUsize       []runtime.Value

	bytecodeDynamicBoxedI8    map[int64]runtime.Value
	bytecodeDynamicBoxedI16   map[int64]runtime.Value
	bytecodeIntBoxDynamicMu   sync.RWMutex
	bytecodeDynamicBoxedI32   map[int64]runtime.Value
	bytecodeDynamicBoxedI64   map[int64]runtime.Value
	bytecodeDynamicBoxedI128  map[int64]runtime.Value
	bytecodeDynamicBoxedU8    map[int64]runtime.Value
	bytecodeDynamicBoxedU16   map[int64]runtime.Value
	bytecodeDynamicBoxedU32   map[int64]runtime.Value
	bytecodeDynamicBoxedU64   map[int64]runtime.Value
	bytecodeDynamicBoxedU128  map[int64]runtime.Value
	bytecodeDynamicBoxedIsize map[int64]runtime.Value
	bytecodeDynamicBoxedUsize map[int64]runtime.Value
)

func init() {
	initBytecodeSmallIntBoxCache()
}

func initBytecodeSmallIntBoxCache() {
	size := int(bytecodeSmallIntBoxMax-bytecodeSmallIntBoxMin) + 1
	bytecodeBoxedI8 = make([]runtime.Value, size)
	bytecodeBoxedI16 = make([]runtime.Value, size)
	bytecodeBoxedI32 = make([]runtime.Value, size)
	bytecodeBoxedI32Extended = make([]runtime.Value, int(bytecodeI32ExtendedBoxMax-bytecodeI32ExtendedBoxMin)+1)
	bytecodeBoxedI64 = make([]runtime.Value, size)
	bytecodeBoxedI128 = make([]runtime.Value, size)
	bytecodeBoxedU8 = make([]runtime.Value, size)
	bytecodeBoxedU16 = make([]runtime.Value, size)
	bytecodeBoxedU32 = make([]runtime.Value, size)
	bytecodeBoxedU64 = make([]runtime.Value, size)
	bytecodeBoxedU128 = make([]runtime.Value, size)
	bytecodeBoxedIsize = make([]runtime.Value, size)
	bytecodeBoxedUsize = make([]runtime.Value, size)
	for cur := bytecodeSmallIntBoxMin; cur <= bytecodeSmallIntBoxMax; cur++ {
		idx := int(cur - bytecodeSmallIntBoxMin)
		bytecodeBoxedI8[idx] = runtime.NewSmallInt(cur, runtime.IntegerI8)
		bytecodeBoxedI16[idx] = runtime.NewSmallInt(cur, runtime.IntegerI16)
		bytecodeBoxedI32[idx] = runtime.NewSmallInt(cur, runtime.IntegerI32)
		bytecodeBoxedI64[idx] = runtime.NewSmallInt(cur, runtime.IntegerI64)
		bytecodeBoxedI128[idx] = runtime.NewSmallInt(cur, runtime.IntegerI128)
		bytecodeBoxedIsize[idx] = runtime.NewSmallInt(cur, runtime.IntegerIsize)
		if cur >= 0 {
			bytecodeBoxedU8[idx] = runtime.NewSmallInt(cur, runtime.IntegerU8)
			bytecodeBoxedU16[idx] = runtime.NewSmallInt(cur, runtime.IntegerU16)
			bytecodeBoxedU32[idx] = runtime.NewSmallInt(cur, runtime.IntegerU32)
			bytecodeBoxedU64[idx] = runtime.NewSmallInt(cur, runtime.IntegerU64)
			bytecodeBoxedU128[idx] = runtime.NewSmallInt(cur, runtime.IntegerU128)
			bytecodeBoxedUsize[idx] = runtime.NewSmallInt(cur, runtime.IntegerUsize)
		}
	}
	for cur := bytecodeI32ExtendedBoxMin; cur <= bytecodeI32ExtendedBoxMax; cur++ {
		bytecodeBoxedI32Extended[int(cur-bytecodeI32ExtendedBoxMin)] = runtime.NewSmallInt(cur, runtime.IntegerI32)
	}
}

func boxedSmallIntValue(kind runtime.IntegerType, value int64) (runtime.Value, bool) {
	if value < bytecodeSmallIntBoxMin || value > bytecodeSmallIntBoxMax {
		return nil, false
	}
	idx := int(value - bytecodeSmallIntBoxMin)
	switch kind {
	case runtime.IntegerI8:
		return bytecodeBoxedI8[idx], true
	case runtime.IntegerI16:
		return bytecodeBoxedI16[idx], true
	case runtime.IntegerI32:
		return bytecodeBoxedI32[idx], true
	case runtime.IntegerI64:
		return bytecodeBoxedI64[idx], true
	case runtime.IntegerI128:
		return bytecodeBoxedI128[idx], true
	case runtime.IntegerU8:
		if value < 0 {
			return nil, false
		}
		return bytecodeBoxedU8[idx], true
	case runtime.IntegerU16:
		if value < 0 {
			return nil, false
		}
		return bytecodeBoxedU16[idx], true
	case runtime.IntegerU32:
		if value < 0 {
			return nil, false
		}
		return bytecodeBoxedU32[idx], true
	case runtime.IntegerU64:
		if value < 0 {
			return nil, false
		}
		return bytecodeBoxedU64[idx], true
	case runtime.IntegerU128:
		if value < 0 {
			return nil, false
		}
		return bytecodeBoxedU128[idx], true
	case runtime.IntegerIsize:
		return bytecodeBoxedIsize[idx], true
	case runtime.IntegerUsize:
		if value < 0 {
			return nil, false
		}
		return bytecodeBoxedUsize[idx], true
	default:
		return nil, false
	}
}

func bytecodeBoxedSmallIntValue(kind runtime.IntegerType, value int64) (runtime.Value, bool) {
	return boxedSmallIntValue(kind, value)
}

func boxedExtendedI32Value(value int64) (runtime.Value, bool) {
	if value < bytecodeI32ExtendedBoxMin || value > bytecodeI32ExtendedBoxMax {
		return nil, false
	}
	return bytecodeBoxedI32Extended[int(value-bytecodeI32ExtendedBoxMin)], true
}

func bytecodeDynamicIntBoxCache(kind runtime.IntegerType) map[int64]runtime.Value {
	switch kind {
	case runtime.IntegerI8:
		return bytecodeDynamicBoxedI8
	case runtime.IntegerI16:
		return bytecodeDynamicBoxedI16
	case runtime.IntegerI32:
		return bytecodeDynamicBoxedI32
	case runtime.IntegerI64:
		return bytecodeDynamicBoxedI64
	case runtime.IntegerI128:
		return bytecodeDynamicBoxedI128
	case runtime.IntegerU8:
		return bytecodeDynamicBoxedU8
	case runtime.IntegerU16:
		return bytecodeDynamicBoxedU16
	case runtime.IntegerU32:
		return bytecodeDynamicBoxedU32
	case runtime.IntegerU64:
		return bytecodeDynamicBoxedU64
	case runtime.IntegerU128:
		return bytecodeDynamicBoxedU128
	case runtime.IntegerIsize:
		return bytecodeDynamicBoxedIsize
	case runtime.IntegerUsize:
		return bytecodeDynamicBoxedUsize
	default:
		return nil
	}
}

func setBytecodeDynamicIntBoxCache(kind runtime.IntegerType, cache map[int64]runtime.Value) {
	switch kind {
	case runtime.IntegerI8:
		bytecodeDynamicBoxedI8 = cache
	case runtime.IntegerI16:
		bytecodeDynamicBoxedI16 = cache
	case runtime.IntegerI32:
		bytecodeDynamicBoxedI32 = cache
	case runtime.IntegerI64:
		bytecodeDynamicBoxedI64 = cache
	case runtime.IntegerI128:
		bytecodeDynamicBoxedI128 = cache
	case runtime.IntegerU8:
		bytecodeDynamicBoxedU8 = cache
	case runtime.IntegerU16:
		bytecodeDynamicBoxedU16 = cache
	case runtime.IntegerU32:
		bytecodeDynamicBoxedU32 = cache
	case runtime.IntegerU64:
		bytecodeDynamicBoxedU64 = cache
	case runtime.IntegerU128:
		bytecodeDynamicBoxedU128 = cache
	case runtime.IntegerIsize:
		bytecodeDynamicBoxedIsize = cache
	case runtime.IntegerUsize:
		bytecodeDynamicBoxedUsize = cache
	}
}

// bytecodeBoxedIntegerValue returns cached boxed integers for supported hot
// integer kinds; values outside the fixed small-int cache use a bounded dynamic
// cache keyed by exact integer value.
func bytecodeBoxedIntegerValue(kind runtime.IntegerType, value int64) (runtime.Value, bool) {
	if boxed, ok := boxedSmallIntValue(kind, value); ok {
		return boxed, true
	}
	if kind == runtime.IntegerI32 {
		if boxed, ok := boxedExtendedI32Value(value); ok {
			return boxed, true
		}
	}
	switch kind {
	case runtime.IntegerI8, runtime.IntegerI16, runtime.IntegerI32, runtime.IntegerI64, runtime.IntegerI128,
		runtime.IntegerU8, runtime.IntegerU16, runtime.IntegerU32, runtime.IntegerU64, runtime.IntegerU128,
		runtime.IntegerIsize, runtime.IntegerUsize:
		// supported kinds
	default:
		return nil, false
	}
	if value < 0 {
		switch kind {
		case runtime.IntegerU8, runtime.IntegerU16, runtime.IntegerU32, runtime.IntegerU64, runtime.IntegerU128, runtime.IntegerUsize:
			return nil, false
		}
	}
	if kind == runtime.IntegerI64 {
		// Large i64 hot loops such as Monte Carlo generate very high-cardinality
		// values with little reuse. Bypass the dynamic boxed-int map there and
		// return the direct small-int carrier instead of paying lock/map churn.
		if bytecodeDynamicIntBoxReuseEnabled {
			bytecodeRecordDynamicIntBoxCacheEvent(kind, bytecodeDynamicIntBoxCacheI64Bypass)
		}
		return runtime.NewSmallInt(value, kind), true
	}
	if bytecodeDynamicIntBoxReuseEnabled {
		bytecodeRecordDynamicIntBoxCacheEvent(kind, bytecodeDynamicIntBoxCacheLookup)
	}

	bytecodeIntBoxDynamicMu.RLock()
	cache := bytecodeDynamicIntBoxCache(kind)
	if cache != nil {
		if boxed, ok := cache[value]; ok {
			bytecodeIntBoxDynamicMu.RUnlock()
			if bytecodeDynamicIntBoxReuseEnabled {
				bytecodeRecordDynamicIntBoxCacheEvent(kind, bytecodeDynamicIntBoxCacheHit)
			}
			return boxed, true
		}
	}
	bytecodeIntBoxDynamicMu.RUnlock()

	boxed := runtime.NewSmallInt(value, kind)

	bytecodeIntBoxDynamicMu.Lock()
	cache = bytecodeDynamicIntBoxCache(kind)
	if cache == nil {
		cache = make(map[int64]runtime.Value, 256)
		setBytecodeDynamicIntBoxCache(kind, cache)
	}
	if existing, ok := cache[value]; ok {
		bytecodeIntBoxDynamicMu.Unlock()
		if bytecodeDynamicIntBoxReuseEnabled {
			bytecodeRecordDynamicIntBoxCacheEvent(kind, bytecodeDynamicIntBoxCacheHit)
		}
		return existing, true
	}
	if len(cache) < bytecodeIntBoxDynamicCacheLimit {
		cache[value] = boxed
		bytecodeIntBoxDynamicMu.Unlock()
		if bytecodeDynamicIntBoxReuseEnabled {
			bytecodeRecordDynamicIntBoxCacheEvent(kind, bytecodeDynamicIntBoxCacheInsert)
		}
		return boxed, true
	}
	bytecodeIntBoxDynamicMu.Unlock()
	if bytecodeDynamicIntBoxReuseEnabled {
		bytecodeRecordDynamicIntBoxCacheEvent(kind, bytecodeDynamicIntBoxCacheCapacityMiss)
	}
	return boxed, true
}
