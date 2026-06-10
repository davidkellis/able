package interpreter

import (
	"math"

	"able/interpreter-go/pkg/runtime"
)

func bytecodeArraySmallI32RawValues(values []runtime.Value) ([]int32, bool) {
	raws := make([]int32, len(values))
	for idx, value := range values {
		raw, ok := bytecodeDirectSmallI32Value(value)
		if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
			return nil, false
		}
		raws[idx] = int32(raw)
	}
	return raws, true
}

func clearTrackedArrayI32RawCache(state *arrayState) {
	if state == nil {
		return
	}
	state.CachedI32Values = nil
	state.CachedI32ValuesValid = nil
	state.CachedI32ValuesCount = 0
	state.CachedI32ValuesKnown = false
}

func trackedArrayI32RawCacheHoleValue(value runtime.Value) bool {
	if value == nil {
		return true
	}
	_, ok := unwrapInterfaceValue(value).(runtime.NilValue)
	return ok
}

func trackedArrayI32RawCacheStorage(state *arrayState, length int) ([]int32, []bool) {
	var raws []int32
	if state != nil && cap(state.CachedI32Values) >= length {
		raws = state.CachedI32Values[:length]
		clear(raws)
	} else {
		raws = make([]int32, length)
	}
	var valid []bool
	if state != nil && cap(state.CachedI32ValuesValid) >= length {
		valid = state.CachedI32ValuesValid[:length]
		clear(valid)
	} else {
		valid = make([]bool, length)
	}
	return raws, valid
}

func grownTrackedArrayI32RawCacheCapacity(current int, minimum int) int {
	if minimum <= current {
		return current
	}
	if current < 4 {
		current = 4
	}
	for current < minimum {
		if current < 4096 {
			current *= 2
		} else {
			current += current / 2
		}
	}
	return current
}

func resizeTrackedArrayI32RawCacheLength(state *arrayState, length int) bool {
	if state == nil {
		return false
	}
	oldLength := len(state.CachedI32Values)
	if oldLength != len(state.CachedI32ValuesValid) {
		clearTrackedArrayI32RawCache(state)
		return false
	}
	if length == oldLength {
		return true
	}
	if length < oldLength {
		removedValidCount := 0
		for _, valid := range state.CachedI32ValuesValid[length:oldLength] {
			if valid {
				removedValidCount++
			}
		}
		clear(state.CachedI32Values[length:oldLength])
		clear(state.CachedI32ValuesValid[length:oldLength])
		state.CachedI32Values = state.CachedI32Values[:length]
		state.CachedI32ValuesValid = state.CachedI32ValuesValid[:length]
		state.CachedI32ValuesCount -= removedValidCount
		if state.CachedI32ValuesCount < 0 {
			state.CachedI32ValuesCount = 0
		}
		state.CachedI32ValuesKnown = state.CachedI32ValuesCount == length
		return true
	}
	if cap(state.CachedI32Values) >= length && cap(state.CachedI32ValuesValid) >= length {
		prevLength := oldLength
		state.CachedI32Values = state.CachedI32Values[:length]
		state.CachedI32ValuesValid = state.CachedI32ValuesValid[:length]
		clear(state.CachedI32Values[prevLength:length])
		clear(state.CachedI32ValuesValid[prevLength:length])
	} else {
		currentCap := cap(state.CachedI32Values)
		if validCap := cap(state.CachedI32ValuesValid); validCap > currentCap {
			currentCap = validCap
		}
		grownCap := grownTrackedArrayI32RawCacheCapacity(currentCap, length)
		if valueCap := cap(state.Values); valueCap > grownCap && valueCap >= length {
			grownCap = valueCap
		}
		raws := make([]int32, length, grownCap)
		copy(raws, state.CachedI32Values)
		valid := make([]bool, length, grownCap)
		copy(valid, state.CachedI32ValuesValid)
		state.CachedI32Values = raws
		state.CachedI32ValuesValid = valid
	}
	if state.CachedI32ValuesCount > length {
		state.CachedI32ValuesCount = length
	}
	state.CachedI32ValuesKnown = false
	return true
}

func trackedArrayI32RawCacheGrowthHolesExcept(state *arrayState, start int, end int, skip int) bool {
	if state == nil || start < 0 || end < start || end > len(state.Values) {
		return false
	}
	for idx := start; idx < end; idx++ {
		if idx == skip {
			continue
		}
		if !trackedArrayI32RawCacheHoleValue(state.Values[idx]) {
			return false
		}
	}
	return true
}

func invalidateTrackedArrayI32RawCacheSlot(state *arrayState, idx int) {
	if state == nil || idx < 0 || idx >= len(state.CachedI32Values) || idx >= len(state.CachedI32ValuesValid) {
		return
	}
	if state.CachedI32ValuesValid[idx] {
		state.CachedI32Values[idx] = 0
		state.CachedI32ValuesValid[idx] = false
		state.CachedI32ValuesCount--
		if state.CachedI32ValuesCount < 0 {
			state.CachedI32ValuesCount = 0
		}
	}
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
}

func extendTrackedArrayI32RawCacheLengthExact(state *arrayState, length int) bool {
	if state == nil {
		return false
	}
	oldLength := len(state.CachedI32Values)
	if oldLength != len(state.CachedI32ValuesValid) || length < oldLength {
		return false
	}
	if cap(state.CachedI32Values) >= length && cap(state.CachedI32ValuesValid) >= length {
		state.CachedI32Values = state.CachedI32Values[:length]
		state.CachedI32ValuesValid = state.CachedI32ValuesValid[:length]
		clear(state.CachedI32Values[oldLength:length])
		clear(state.CachedI32ValuesValid[oldLength:length])
	} else {
		raws := make([]int32, length)
		copy(raws, state.CachedI32Values)
		valid := make([]bool, length)
		copy(valid, state.CachedI32ValuesValid)
		state.CachedI32Values = raws
		state.CachedI32ValuesValid = valid
	}
	if state.CachedI32ValuesCount > length {
		state.CachedI32ValuesCount = length
	}
	state.CachedI32ValuesKnown = false
	return true
}

func extendTrackedArrayI32RawCacheSuffix(state *arrayState, start int, end int, skip int) bool {
	if state == nil || start < 0 || end < start || end > len(state.Values) {
		return false
	}
	if start != len(state.CachedI32Values) || start != len(state.CachedI32ValuesValid) {
		return false
	}
	if trackedArrayI32RawCacheGrowthHolesExcept(state, start, end, skip) {
		return resizeTrackedArrayI32RawCacheLength(state, end)
	}
	if !extendTrackedArrayI32RawCacheLengthExact(state, end) {
		return false
	}
	if state.CachedI32ValuesCount < 0 {
		state.CachedI32ValuesCount = 0
	}
	if state.CachedI32ValuesCount > start {
		state.CachedI32ValuesCount = start
	}
	for idx := start; idx < end; idx++ {
		if idx == skip {
			continue
		}
		value := state.Values[idx]
		if trackedArrayI32RawCacheHoleValue(value) {
			continue
		}
		raw, ok := bytecodeDirectSmallI32Value(value)
		if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
			return false
		}
		if !state.CachedI32ValuesValid[idx] {
			state.CachedI32ValuesCount++
		}
		state.CachedI32Values[idx] = int32(raw)
		state.CachedI32ValuesValid[idx] = true
	}
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
	return true
}

func appendTrackedArrayI32RawCacheValue(state *arrayState, value runtime.Value) bool {
	if state == nil {
		return false
	}
	length := len(state.Values)
	if length <= 0 || len(state.CachedI32Values) != length-1 || len(state.CachedI32ValuesValid) != length-1 {
		return false
	}
	if !resizeTrackedArrayI32RawCacheLength(state, length) {
		return false
	}
	idx := length - 1
	if trackedArrayI32RawCacheHoleValue(value) {
		invalidateTrackedArrayI32RawCacheSlot(state, idx)
		return true
	}
	raw, ok := bytecodeDirectSmallI32Value(value)
	if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
		clearTrackedArrayI32RawCache(state)
		return true
	}
	if !state.CachedI32ValuesValid[idx] {
		state.CachedI32ValuesCount++
	}
	state.CachedI32Values[idx] = int32(raw)
	state.CachedI32ValuesValid[idx] = true
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
	return true
}

func refreshTrackedArrayI32RawCache(state *arrayState) {
	if state == nil || !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		clearTrackedArrayI32RawCache(state)
		return
	}
	raws, valid := trackedArrayI32RawCacheStorage(state, len(state.Values))
	validCount := 0
	for idx, value := range state.Values {
		if trackedArrayI32RawCacheHoleValue(value) {
			continue
		}
		raw, ok := bytecodeDirectSmallI32Value(value)
		if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
			clearTrackedArrayI32RawCache(state)
			return
		}
		raws[idx] = int32(raw)
		valid[idx] = true
		validCount++
	}
	state.CachedI32Values = raws
	state.CachedI32ValuesValid = valid
	state.CachedI32ValuesCount = validCount
	state.CachedI32ValuesKnown = validCount == len(state.Values)
}

func reconcileTrackedArrayI32RawCacheLength(state *arrayState, skip int) bool {
	if state == nil {
		return false
	}
	if len(state.CachedI32Values) == len(state.Values) && len(state.CachedI32ValuesValid) == len(state.Values) {
		if state.CachedI32ValuesCount < 0 {
			state.CachedI32ValuesCount = 0
		}
		if state.CachedI32ValuesCount > len(state.Values) {
			state.CachedI32ValuesCount = len(state.Values)
		}
		state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
		return true
	}
	cacheLength := len(state.CachedI32Values)
	if cacheLength != len(state.CachedI32ValuesValid) {
		clearTrackedArrayI32RawCache(state)
		return false
	}
	valueLength := len(state.Values)
	switch {
	case cacheLength > valueLength:
		if resizeTrackedArrayI32RawCacheLength(state, valueLength) {
			return true
		}
	case extendTrackedArrayI32RawCacheSuffix(state, cacheLength, valueLength, skip):
		return true
	}
	refreshTrackedArrayI32RawCache(state)
	return len(state.CachedI32Values) == len(state.Values) && len(state.CachedI32ValuesValid) == len(state.Values)
}

func updateTrackedArrayI32RawCacheForWrite(state *arrayState, idx int, value runtime.Value) {
	if state == nil || !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		clearTrackedArrayI32RawCache(state)
		return
	}
	if idx < 0 || idx >= len(state.Values) {
		clearTrackedArrayI32RawCache(state)
		return
	}
	if len(state.CachedI32Values) != len(state.Values) || len(state.CachedI32ValuesValid) != len(state.Values) {
		if idx == len(state.Values)-1 && appendTrackedArrayI32RawCacheValue(state, value) {
			return
		}
		if !reconcileTrackedArrayI32RawCacheLength(state, idx) {
			return
		}
	}
	if trackedArrayI32RawCacheHoleValue(value) {
		invalidateTrackedArrayI32RawCacheSlot(state, idx)
		return
	}
	raw, ok := bytecodeDirectSmallI32Value(value)
	if !ok || raw < math.MinInt32 || raw > math.MaxInt32 {
		clearTrackedArrayI32RawCache(state)
		return
	}
	if !state.CachedI32ValuesValid[idx] {
		state.CachedI32ValuesCount++
	}
	state.CachedI32Values[idx] = int32(raw)
	state.CachedI32ValuesValid[idx] = true
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.Values)
}

func trackedArrayCachedI32RawAt(state *runtime.ArrayState, idx int) (int64, bool) {
	if state == nil || idx < 0 || idx >= len(state.CachedI32Values) || idx >= len(state.CachedI32ValuesValid) {
		return 0, false
	}
	if !state.CachedI32ValuesValid[idx] {
		return 0, false
	}
	return int64(state.CachedI32Values[idx]), true
}

func swapTrackedArrayI32RawCache(state *runtime.ArrayState, first int, second int) {
	if state == nil || first < 0 || second < 0 {
		return
	}
	if first >= len(state.CachedI32Values) || second >= len(state.CachedI32Values) ||
		first >= len(state.CachedI32ValuesValid) || second >= len(state.CachedI32ValuesValid) ||
		first == second {
		return
	}
	state.CachedI32Values[first], state.CachedI32Values[second] = state.CachedI32Values[second], state.CachedI32Values[first]
	state.CachedI32ValuesValid[first], state.CachedI32ValuesValid[second] = state.CachedI32ValuesValid[second], state.CachedI32ValuesValid[first]
	state.CachedI32ValuesKnown = state.CachedI32ValuesCount == len(state.CachedI32ValuesValid)
}
