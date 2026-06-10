package bridge

import (
	"sync"

	"able/interpreter-go/pkg/runtime"
)

const (
	commonI32BoxMin int64 = -128
	commonI32BoxMax int64 = 4095
)

var (
	commonI32Boxes  [commonI32BoxMax - commonI32BoxMin + 1]lazyIntegerBox
	dynamicI64Boxes [commonI32BoxMax - commonI32BoxMin + 1]lazyIntegerBox
)

type lazyIntegerBox struct {
	once  sync.Once
	value runtime.Value
}

// commonI32Box returns an immutable runtime value for the common signed i32
// range. IntegerValue is stored by value, so callers cannot mutate the cached
// box through a returned runtime.Value interface.
func commonI32Box(value int64) (runtime.Value, bool) {
	if value < commonI32BoxMin || value > commonI32BoxMax {
		return nil, false
	}
	entry := &commonI32Boxes[int(value-commonI32BoxMin)]
	entry.once.Do(func() {
		entry.value = runtime.NewSmallInt(value, runtime.IntegerI32)
	})
	return entry.value, true
}

// ToDynamicI64 boxes an i64 that is crossing from generated native code into
// the dynamic runtime.Value boundary. Those values conservatively escape in
// Go, so reuse the immutable common-value boxes at this boundary only.
func ToDynamicI64(value int64) runtime.Value {
	if value < commonI32BoxMin || value > commonI32BoxMax {
		return runtime.NewSmallInt(value, runtime.IntegerI64)
	}
	entry := &dynamicI64Boxes[int(value-commonI32BoxMin)]
	// Initialize slots independently so a program touching one common value
	// does not allocate boxes for the entire range. Once also makes the first
	// access safe when generated calls reach the same slot concurrently.
	entry.once.Do(func() {
		entry.value = runtime.NewSmallInt(value, runtime.IntegerI64)
	})
	return entry.value
}
