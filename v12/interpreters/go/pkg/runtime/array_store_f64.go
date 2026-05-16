package runtime

import "fmt"

func ArrayStoreMonoReadF64(handle int64, index int) (float64, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, err
	}
	if kind == monoArrayKindDynamic {
		value, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return float64FromValue(value)
	}
	if kind != monoArrayKindF64 {
		return 0, fmt.Errorf("array handle %d is not mono f64", handle)
	}
	state, ok := monoArrayF64States[handle]
	if !ok {
		return 0, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return 0, fmt.Errorf("index out of bounds")
	}
	return state.Values[index], nil
}

func ArrayStoreMonoWriteF64(handle int64, index int, value float64) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, f64ToValue(value))
	}
	if kind != monoArrayKindF64 {
		return fmt.Errorf("array handle %d is not mono f64", handle)
	}
	state, ok := monoArrayF64States[handle]
	if !ok {
		return fmt.Errorf("array handle %d is not defined", handle)
	}
	monoEnsureCapacity(state, index+1)
	if index >= len(state.Values) {
		monoSetLength(state, index+1)
	}
	state.Values[index] = value
	return nil
}

func ArrayStoreMonoF64ValuesIfAvailable(handle int64) ([]float64, bool, error) {
	if handle == 0 {
		return nil, false, nil
	}
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return nil, false, err
	}
	if kind != monoArrayKindF64 {
		return nil, false, nil
	}
	state, ok := monoArrayF64States[handle]
	if !ok {
		return nil, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state.Values, true, nil
}

func ArrayStoreAppendF64Promote(handle int64, value float64) (bool, error) {
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindF64 {
		state, ok := monoArrayF64States[handle]
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
	arrayHandleKinds[handle] = monoArrayKindF64
	return true, nil
}

func ArrayStoreAppendF64ValuesPromote(handle int64, values []float64) (bool, error) {
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return false, err
	}
	if len(values) == 0 {
		return kind == monoArrayKindF64 || kind == monoArrayKindDynamic, nil
	}
	if kind == monoArrayKindF64 {
		state, ok := monoArrayF64States[handle]
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
	monoArrayF64States[handle] = &monoArrayF64State{Values: converted, Capacity: cap(converted)}
	arrayHandleKinds[handle] = monoArrayKindF64
	return true, nil
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
}
