package interpreter

import (
	"sync"

	"able/interpreter-go/pkg/runtime"
)

const (
	bytecodeSmallIntBoxMin int64 = -256
	bytecodeSmallIntBoxMax int64 = 16384
	// Bound dynamic boxed-int growth for out-of-range i32/i64/isize values.
	bytecodeIntBoxDynamicCacheLimit = 32768
)

var (
	bytecodeBoxedI32   []runtime.Value
	bytecodeBoxedI64   []runtime.Value
	bytecodeBoxedIsize []runtime.Value

	bytecodeIntBoxDynamicMu   sync.RWMutex
	bytecodeDynamicBoxedI32   map[int64]runtime.Value
	bytecodeDynamicBoxedI64   map[int64]runtime.Value
	bytecodeDynamicBoxedIsize map[int64]runtime.Value
)

func init() {
	initBytecodeSmallIntBoxCache()
}

func initBytecodeSmallIntBoxCache() {
	size := int(bytecodeSmallIntBoxMax-bytecodeSmallIntBoxMin) + 1
	bytecodeBoxedI32 = make([]runtime.Value, size)
	bytecodeBoxedI64 = make([]runtime.Value, size)
	bytecodeBoxedIsize = make([]runtime.Value, size)
	for cur := bytecodeSmallIntBoxMin; cur <= bytecodeSmallIntBoxMax; cur++ {
		idx := int(cur - bytecodeSmallIntBoxMin)
		bytecodeBoxedI32[idx] = runtime.NewSmallInt(cur, runtime.IntegerI32)
		bytecodeBoxedI64[idx] = runtime.NewSmallInt(cur, runtime.IntegerI64)
		bytecodeBoxedIsize[idx] = runtime.NewSmallInt(cur, runtime.IntegerIsize)
	}
}

func boxedSmallIntValue(kind runtime.IntegerType, value int64) (runtime.Value, bool) {
	if value < bytecodeSmallIntBoxMin || value > bytecodeSmallIntBoxMax {
		return nil, false
	}
	idx := int(value - bytecodeSmallIntBoxMin)
	switch kind {
	case runtime.IntegerI32:
		return bytecodeBoxedI32[idx], true
	case runtime.IntegerI64:
		return bytecodeBoxedI64[idx], true
	case runtime.IntegerIsize:
		return bytecodeBoxedIsize[idx], true
	default:
		return nil, false
	}
}

func bytecodeBoxedSmallIntValue(kind runtime.IntegerType, value int64) (runtime.Value, bool) {
	return boxedSmallIntValue(kind, value)
}

func bytecodeDynamicIntBoxCache(kind runtime.IntegerType) map[int64]runtime.Value {
	switch kind {
	case runtime.IntegerI32:
		return bytecodeDynamicBoxedI32
	case runtime.IntegerI64:
		return bytecodeDynamicBoxedI64
	case runtime.IntegerIsize:
		return bytecodeDynamicBoxedIsize
	default:
		return nil
	}
}

func setBytecodeDynamicIntBoxCache(kind runtime.IntegerType, cache map[int64]runtime.Value) {
	switch kind {
	case runtime.IntegerI32:
		bytecodeDynamicBoxedI32 = cache
	case runtime.IntegerI64:
		bytecodeDynamicBoxedI64 = cache
	case runtime.IntegerIsize:
		bytecodeDynamicBoxedIsize = cache
	}
}

// bytecodeBoxedIntegerValue returns cached boxed integers for supported hot
// integer kinds; values outside the fixed small-int cache use a bounded dynamic
// cache keyed by exact integer value.
func bytecodeBoxedIntegerValue(kind runtime.IntegerType, value int64) (runtime.Value, bool) {
	if boxed, ok := boxedSmallIntValue(kind, value); ok {
		return boxed, true
	}
	switch kind {
	case runtime.IntegerI32, runtime.IntegerI64, runtime.IntegerIsize:
		// supported kinds
	default:
		return nil, false
	}

	bytecodeIntBoxDynamicMu.RLock()
	cache := bytecodeDynamicIntBoxCache(kind)
	if cache != nil {
		if boxed, ok := cache[value]; ok {
			bytecodeIntBoxDynamicMu.RUnlock()
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
		return existing, true
	}
	if len(cache) < bytecodeIntBoxDynamicCacheLimit {
		cache[value] = boxed
	}
	bytecodeIntBoxDynamicMu.Unlock()
	return boxed, true
}
