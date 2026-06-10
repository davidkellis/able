package runtime

import "fmt"

func ArrayStoreMonoReadF64(handle int64, index int) (float64, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindF64, monoArrayF64States)
	if err != nil || kind == monoArrayKindF64 {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return float64FromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono f64", handle)
}

func ArrayStoreMonoWriteF64(handle int64, index int, value float64) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindF64, monoArrayF64States, true)
	if err != nil || kind == monoArrayKindF64 {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, f64ToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono f64", handle)
}

func ArrayStoreMonoF64ValuesIfAvailable(handle int64) ([]float64, bool, error) {
	values, _, ok, err := ArrayStoreMonoF64ValuesRevisionIfAvailable(handle)
	return values, ok, err
}

func ArrayStoreMonoF64ValuesRevisionIfAvailable(handle int64) ([]float64, uint64, bool, error) {
	if handle == 0 {
		return nil, 0, false, nil
	}
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, 0, false, err
	}
	if kind != monoArrayKindF64 {
		return nil, 0, false, nil
	}
	state, ok := monoArrayF64States[handle]
	if !ok {
		return nil, 0, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state.Values, state.Revision, true, nil
}

func ArrayStoreAppendF64Promote(handle int64, value float64) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindF64 {
		state, ok := cachedMonoArrayF64ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoF64Value(state, value)
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
	values := make([]float64, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := float64FromValue(current)
		if err != nil {
			return false, nil
		}
		values[idx] = raw
	}
	mono := &monoArrayF64State{Values: values, Capacity: capacity}
	appendMonoF64Value(mono, value)
	delete(arrayStates, handle)
	monoArrayF64States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindF64)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func ArrayStoreAppendF64ValuesPromote(handle int64, values []float64) (bool, error) {
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
		return kind == monoArrayKindF64 || kind == monoArrayKindDynamic, nil
	}
	if kind == monoArrayKindF64 {
		state, ok := cachedMonoArrayF64ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoF64Values(state, values)
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
	converted := make([]float64, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := float64FromValue(current)
		if err != nil {
			return false, nil
		}
		converted[idx] = raw
	}
	converted = append(converted, values...)
	delete(arrayStates, handle)
	mono := &monoArrayF64State{Values: converted, Capacity: cap(converted)}
	monoArrayF64States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindF64)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func ArrayStoreAppendF64UninitializedPromote(handle int64, count int) ([]float64, bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 || count < 0 {
		return nil, false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, false, err
	}
	if count == 0 {
		if kind == monoArrayKindF64 || kind == monoArrayKindDynamic {
			return []float64{}, true, nil
		}
		return nil, false, nil
	}
	if kind == monoArrayKindF64 {
		state, ok := monoArrayF64States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		oldLen := len(state.Values)
		minimum := oldLen + count
		if minimum > state.Capacity || minimum > cap(state.Values) {
			monoEnsureCapacity(state, minimum)
		}
		state.Values = state.Values[:minimum]
		if state.Capacity < cap(state.Values) {
			state.Capacity = cap(state.Values)
		}
		state.Revision++
		return state.Values[oldLen:minimum], true, nil
	}
	if kind != monoArrayKindDynamic {
		return nil, false, nil
	}
	state, ok := arrayStates[handle]
	if !ok {
		return nil, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	capacity := state.Capacity
	if capacity < len(state.Values) {
		capacity = len(state.Values)
	}
	minCapacity := len(state.Values) + count
	if capacity < minCapacity {
		capacity = grownCapacity(capacity, minCapacity)
	}
	converted := make([]float64, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := float64FromValue(current)
		if err != nil {
			return nil, false, nil
		}
		converted[idx] = raw
	}
	oldLen := len(converted)
	converted = converted[:minCapacity]
	delete(arrayStates, handle)
	mono := &monoArrayF64State{Values: converted, Capacity: cap(converted), Revision: state.Revision + 1}
	monoArrayF64States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindF64)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return converted[oldLen:minCapacity], true, nil
}

func appendMonoF64Value(state *monoArrayF64State, value float64) {
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

func appendMonoF64Values(state *monoArrayF64State, values []float64) {
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
