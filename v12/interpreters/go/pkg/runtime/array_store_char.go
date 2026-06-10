package runtime

import "fmt"

var monoArrayCharAppendHotHandle int64
var monoArrayCharAppendHot *monoArrayCharState
var monoArrayCharAppendCache []*monoArrayCharState

func clearMonoArrayCharAppendCache(handle int64) {
	if monoArrayCharAppendHotHandle == handle {
		monoArrayCharAppendHotHandle, monoArrayCharAppendHot = 0, nil
	}
	idx, ok := arrayHandleKindCacheIndex(handle)
	if !ok || idx >= len(monoArrayCharAppendCache) {
		return
	}
	monoArrayCharAppendCache[idx] = nil
}

func rememberMonoArrayCharAppendState(handle int64, state *monoArrayCharState) {
	if handle == 0 || state == nil {
		return
	}
	monoArrayCharAppendHotHandle, monoArrayCharAppendHot = handle, state
	idx, ok := arrayHandleKindCacheIndex(handle)
	if !ok {
		return
	}
	if idx >= len(monoArrayCharAppendCache) {
		growMonoArrayCharAppendCache(idx + 1)
	}
	monoArrayCharAppendCache[idx] = state
}

func growMonoArrayCharAppendCache(newLen int) {
	if newLen <= len(monoArrayCharAppendCache) {
		return
	}
	if newLen <= cap(monoArrayCharAppendCache) {
		monoArrayCharAppendCache = monoArrayCharAppendCache[:newLen]
		return
	}
	newCap := grownCapacity(cap(monoArrayCharAppendCache), newLen)
	next := make([]*monoArrayCharState, newLen, newCap)
	copy(next, monoArrayCharAppendCache)
	monoArrayCharAppendCache = next
}

func cachedMonoArrayCharAppendState(handle int64) (*monoArrayCharState, bool) {
	if monoArrayCharAppendHotHandle == handle && monoArrayCharAppendHot != nil {
		return monoArrayCharAppendHot, true
	}
	if idx, ok := arrayHandleKindCacheIndex(handle); ok && idx < len(monoArrayCharAppendCache) {
		if state := monoArrayCharAppendCache[idx]; state != nil {
			monoArrayCharAppendHotHandle, monoArrayCharAppendHot = handle, state
			return state, true
		}
	}
	state, ok := monoArrayCharStates[handle]
	if ok {
		rememberMonoArrayCharAppendState(handle, state)
	}
	return state, ok
}

func arrayStoreMonoCharState(handle int64) (*monoArrayCharState, bool, error) {
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, false, err
	}
	if kind != monoArrayKindChar {
		return nil, false, nil
	}
	state, ok := monoArrayCharStates[handle]
	if !ok {
		return nil, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state, true, nil
}

func ArrayStoreMonoReadCharIfAvailable(handle int64, index int) (rune, bool, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	state, ok, err := arrayStoreMonoCharState(handle)
	if err != nil || !ok {
		return 0, ok, err
	}
	if index < 0 || index >= len(state.Values) {
		return 0, false, nil
	}
	return state.Values[index], true, nil
}

func ArrayStoreAppendCharIfMono(handle int64, value rune) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	state, ok := cachedMonoArrayCharAppendState(handle)
	if !ok {
		return false, nil
	}
	appendMonoCharValue(state, value)
	return true, nil
}

func ArrayStoreAppendCharPromote(handle int64, value rune) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindChar {
		state, ok := cachedMonoArrayCharReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoCharValue(state, value)
		return true, nil
	}
	if kind != monoArrayKindDynamic {
		return false, nil
	}
	state, ok := arrayStates[handle]
	if !ok {
		return false, fmt.Errorf("array handle %d is not defined", handle)
	}
	capacity := state.Capacity
	if capacity < len(state.Values) {
		capacity = len(state.Values)
	}
	values := make([]rune, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := charFromValue(current)
		if err != nil {
			return false, nil
		}
		values[idx] = raw
	}
	mono := &monoArrayCharState{Values: values, Capacity: capacity}
	appendMonoCharValue(mono, value)
	delete(arrayStates, handle)
	monoArrayCharStates[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindChar)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func appendMonoCharValue(state *monoArrayCharState, value rune) {
	if state == nil {
		return
	}
	idx := len(state.Values)
	if idx+1 > state.Capacity || idx == cap(state.Values) {
		monoEnsureCapacity(state, idx+1)
	}
	state.Values = append(state.Values, value)
	if state.Capacity < cap(state.Values) {
		state.Capacity = cap(state.Values)
	}
	state.Revision++
}

func ArrayStoreValueViewFromHandle(handle int64, lengthHint int, capacityHint int) (*ArrayValue, *ArrayState, error) {
	if handle == 0 {
		return nil, nil, fmt.Errorf("array handle must be non-zero")
	}
	kind, err := arrayHandleKind(handle)
	if err == nil && kind != monoArrayKindDynamic {
		arr := &ArrayValue{Handle: handle, TrackedHandle: handle}
		if err := ArrayStoreTrackArrayValueLease(arr, handle); err != nil {
			return nil, nil, err
		}
		return arr, nil, nil
	}
	state, err := ArrayStoreEnsureHandle(handle, lengthHint, capacityHint)
	if err != nil {
		return nil, nil, err
	}
	arr := &ArrayValue{
		Elements:      state.Values,
		Handle:        handle,
		State:         state,
		TrackedHandle: handle,
	}
	if err := ArrayStoreTrackArrayValueLease(arr, handle); err != nil {
		return nil, nil, err
	}
	return arr, state, nil
}
