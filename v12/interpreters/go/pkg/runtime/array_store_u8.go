package runtime

import "fmt"

func ArrayStoreMonoValueFromU8Bytes(data []byte) *ArrayValue {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	ensureArrayStore()
	handle := allocateArrayHandle()
	values := make([]uint8, len(data))
	copy(values, data)
	state := &monoArrayU8State{
		Values:   values,
		Capacity: len(values),
	}
	monoArrayU8States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindU8)
	cacheArrayHandleRevision(handle, &state.Revision)
	arr := &ArrayValue{Handle: handle, TrackedHandle: handle}
	_ = arrayStoreTrackArrayValueLeaseLocked(arr, handle)
	return arr
}

func ArrayStoreMonoValueFromOwnedU8Bytes(data []byte) *ArrayValue {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	ensureArrayStore()
	handle := allocateArrayHandle()
	state := &monoArrayU8State{
		Values:   data,
		Capacity: len(data),
	}
	monoArrayU8States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindU8)
	cacheArrayHandleRevision(handle, &state.Revision)
	arr := &ArrayValue{Handle: handle, TrackedHandle: handle}
	_ = arrayStoreTrackArrayValueLeaseLocked(arr, handle)
	return arr
}

func ArrayStoreMonoValueFromU8String(text string) *ArrayValue {
	return ArrayStoreMonoValueFromU8Bytes([]byte(text))
}

func arrayStoreMonoU8State(handle int64) (*monoArrayU8State, bool, error) {
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, false, err
	}
	if kind != monoArrayKindU8 {
		return nil, false, nil
	}
	state, ok := monoArrayU8States[handle]
	if !ok {
		return nil, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state, true, nil
}

func ArrayStoreMonoReadU8IfAvailable(handle int64, index int) (uint8, bool, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	state, ok, err := arrayStoreMonoU8State(handle)
	if err != nil || !ok {
		return 0, ok, err
	}
	if index < 0 || index >= len(state.Values) {
		return 0, false, nil
	}
	return state.Values[index], true, nil
}

func ArrayStoreMonoBorrowedU8BytesIfAvailable(handle int64) ([]byte, bool, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	state, ok, err := arrayStoreMonoU8State(handle)
	if err != nil || !ok {
		return nil, ok, err
	}
	return state.Values, true, nil
}

func ArrayStoreMonoU8BytesIfAvailable(handle int64) ([]byte, bool, error) {
	values, ok, err := ArrayStoreMonoBorrowedU8BytesIfAvailable(handle)
	if err != nil || !ok {
		return nil, ok, err
	}
	bytes := make([]byte, len(values))
	copy(bytes, values)
	return bytes, true, nil
}

func ArrayStoreAppendU8Promote(handle int64, value uint8) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindU8 {
		state, ok := cachedMonoArrayU8ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoU8Value(state, value)
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
	values := make([]byte, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := u8FromValue(current)
		if err != nil {
			return false, nil
		}
		values[idx] = raw
	}
	mono := &monoArrayU8State{Values: values, Capacity: capacity}
	appendMonoU8Value(mono, value)
	delete(arrayStates, handle)
	monoArrayU8States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindU8)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func ArrayStoreAppendU8BytesPromote(handle int64, values []byte) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if len(values) == 0 {
		return kind == monoArrayKindU8 || kind == monoArrayKindDynamic, nil
	}
	if kind == monoArrayKindU8 {
		state, ok := cachedMonoArrayU8ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoU8Bytes(state, values)
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
	minCapacity := len(state.Values) + len(values)
	if capacity < minCapacity {
		capacity = grownCapacity(capacity, minCapacity)
	}
	converted := make([]byte, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := u8FromValue(current)
		if err != nil {
			return false, nil
		}
		converted[idx] = raw
	}
	converted = append(converted, values...)
	delete(arrayStates, handle)
	mono := &monoArrayU8State{Values: converted, Capacity: cap(converted)}
	monoArrayU8States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindU8)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func ArrayStoreAppendU8StringPromote(handle int64, text string) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if len(text) == 0 {
		return kind == monoArrayKindU8 || kind == monoArrayKindDynamic, nil
	}
	if kind == monoArrayKindU8 {
		state, ok := cachedMonoArrayU8ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoU8String(state, text)
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
	minCapacity := len(state.Values) + len(text)
	if capacity < minCapacity {
		capacity = grownCapacity(capacity, minCapacity)
	}
	converted := make([]byte, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := u8FromValue(current)
		if err != nil {
			return false, nil
		}
		converted[idx] = raw
	}
	converted = append(converted, text...)
	delete(arrayStates, handle)
	mono := &monoArrayU8State{Values: converted, Capacity: cap(converted)}
	monoArrayU8States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindU8)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func appendMonoU8Value(state *monoArrayU8State, value uint8) {
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

func appendMonoU8Bytes(state *monoArrayU8State, values []byte) {
	if state == nil || len(values) == 0 {
		return
	}
	minimum := len(state.Values) + len(values)
	if minimum > state.Capacity || minimum > cap(state.Values) {
		monoEnsureCapacity(state, minimum)
	}
	state.Values = append(state.Values, values...)
	if state.Capacity < cap(state.Values) {
		state.Capacity = cap(state.Values)
	}
	state.Revision++
}

func appendMonoU8String(state *monoArrayU8State, text string) {
	if state == nil || len(text) == 0 {
		return
	}
	minimum := len(state.Values) + len(text)
	if minimum > state.Capacity || minimum > cap(state.Values) {
		monoEnsureCapacity(state, minimum)
	}
	state.Values = append(state.Values, text...)
	if state.Capacity < cap(state.Values) {
		state.Capacity = cap(state.Values)
	}
	state.Revision++
}
