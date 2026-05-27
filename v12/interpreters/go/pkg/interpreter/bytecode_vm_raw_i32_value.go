package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeRawI32SlotValue int32

const (
	bytecodeRawI32SlotCacheMin = -1024
	bytecodeRawI32SlotCacheMax = 262143
)

var bytecodeRawI32SlotCache = func() [bytecodeRawI32SlotCacheMax - bytecodeRawI32SlotCacheMin + 1]runtime.Value {
	var cache [bytecodeRawI32SlotCacheMax - bytecodeRawI32SlotCacheMin + 1]runtime.Value
	for idx := range cache {
		cache[idx] = bytecodeRawI32SlotValue(int32(idx + bytecodeRawI32SlotCacheMin))
	}
	return cache
}()

func (v bytecodeRawI32SlotValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

func bytecodeRawI32SlotCachedValue(value int32) runtime.Value {
	if value >= bytecodeRawI32SlotCacheMin && value <= bytecodeRawI32SlotCacheMax {
		return bytecodeRawI32SlotCache[int(value-bytecodeRawI32SlotCacheMin)]
	}
	return bytecodeRawI32SlotValue(value)
}

func bytecodeBoxRawI32Value(value bytecodeRawI32SlotValue) runtime.Value {
	return bytecodeBoxedIntegerI32Value(int64(value))
}
