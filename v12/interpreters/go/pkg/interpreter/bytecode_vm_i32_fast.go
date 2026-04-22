package interpreter

import (
	"math"

	"able/interpreter-go/pkg/runtime"
)

func bytecodeBoxedIntegerI32Value(value int64) runtime.Value {
	if value >= bytecodeSmallIntBoxMin && value <= bytecodeSmallIntBoxMax {
		return bytecodeBoxedI32[int(value-bytecodeSmallIntBoxMin)]
	}

	bytecodeIntBoxDynamicMu.RLock()
	if bytecodeDynamicBoxedI32 != nil {
		if boxed, ok := bytecodeDynamicBoxedI32[value]; ok {
			bytecodeIntBoxDynamicMu.RUnlock()
			return boxed
		}
	}
	bytecodeIntBoxDynamicMu.RUnlock()

	boxed := runtime.NewSmallInt(value, runtime.IntegerI32)

	bytecodeIntBoxDynamicMu.Lock()
	cache := bytecodeDynamicBoxedI32
	if cache == nil {
		cache = make(map[int64]runtime.Value, 256)
		bytecodeDynamicBoxedI32 = cache
	}
	if existing, ok := cache[value]; ok {
		bytecodeIntBoxDynamicMu.Unlock()
		return existing
	}
	if len(cache) < bytecodeIntBoxDynamicCacheLimit {
		cache[value] = boxed
	}
	bytecodeIntBoxDynamicMu.Unlock()
	return boxed
}

func bytecodeDirectSmallI32Value(val runtime.Value) (int64, bool) {
	switch iv := val.(type) {
	case runtime.IntegerValue:
		ivRef := &iv
		if iv.TypeSuffix == runtime.IntegerI32 && ivRef.IsSmallRef() {
			return ivRef.Int64FastRef(), true
		}
	case *runtime.IntegerValue:
		if iv != nil && iv.TypeSuffix == runtime.IntegerI32 && iv.IsSmallRef() {
			return iv.Int64FastRef(), true
		}
	}
	return 0, false
}

func bytecodeDirectSmallI32Pair(left runtime.Value, right runtime.Value) (int64, int64, bool) {
	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if lv.TypeSuffix != runtime.IntegerI32 || !lvRef.IsSmallRef() {
			return 0, 0, false
		}
		switch rv := right.(type) {
		case runtime.IntegerValue:
			rvRef := &rv
			if rv.TypeSuffix == runtime.IntegerI32 && rvRef.IsSmallRef() {
				return lvRef.Int64FastRef(), rvRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.TypeSuffix == runtime.IntegerI32 && rv.IsSmallRef() {
				return lvRef.Int64FastRef(), rv.Int64FastRef(), true
			}
		}
	case *runtime.IntegerValue:
		if lv == nil || lv.TypeSuffix != runtime.IntegerI32 || !lv.IsSmallRef() {
			return 0, 0, false
		}
		switch rv := right.(type) {
		case runtime.IntegerValue:
			rvRef := &rv
			if rv.TypeSuffix == runtime.IntegerI32 && rvRef.IsSmallRef() {
				return lv.Int64FastRef(), rvRef.Int64FastRef(), true
			}
		case *runtime.IntegerValue:
			if rv != nil && rv.TypeSuffix == runtime.IntegerI32 && rv.IsSmallRef() {
				return lv.Int64FastRef(), rv.Int64FastRef(), true
			}
		}
	}
	return 0, 0, false
}

func bytecodeAddSmallI32PairFast(left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	l, r, ok := bytecodeDirectSmallI32Pair(left, right)
	if !ok {
		return nil, false, nil
	}
	sum, overflow := addInt64Overflow(l, r)
	if overflow {
		return nil, false, nil
	}
	if sum < math.MinInt32 || sum > math.MaxInt32 {
		return nil, true, newOverflowError("integer overflow")
	}
	return bytecodeBoxedIntegerI32Value(sum), true, nil
}

func bytecodeSubtractSmallI32PairFast(left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	l, ok := bytecodeDirectSmallI32Value(left)
	if !ok {
		return nil, false, nil
	}
	r, ok := bytecodeDirectSmallI32Value(right)
	if !ok {
		return nil, false, nil
	}
	diff, overflow := subInt64Overflow(l, r)
	if overflow {
		return nil, false, nil
	}
	if diff < math.MinInt32 || diff > math.MaxInt32 {
		return nil, true, newOverflowError("integer overflow")
	}
	return bytecodeBoxedIntegerI32Value(diff), true, nil
}

func bytecodeSubtractIntegerImmediateI32Fast(left runtime.Value, right runtime.IntegerValue) (runtime.Value, bool, error) {
	rightRef := &right
	if right.TypeSuffix != runtime.IntegerI32 || !rightRef.IsSmallRef() {
		return nil, false, nil
	}
	leftVal, ok := bytecodeDirectSmallI32Value(left)
	if !ok {
		return nil, false, nil
	}
	diff, overflow := subInt64Overflow(leftVal, rightRef.Int64FastRef())
	if overflow {
		return nil, false, nil
	}
	if diff < math.MinInt32 || diff > math.MaxInt32 {
		return nil, true, newOverflowError("integer overflow")
	}
	return bytecodeBoxedIntegerI32Value(diff), true, nil
}
