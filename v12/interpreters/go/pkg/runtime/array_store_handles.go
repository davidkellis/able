package runtime

import "fmt"

func ArrayStoreNew() int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	handle := allocateArrayHandle()
	state := &ArrayState{Values: make([]Value, 0), Capacity: 0, ValuesMaterialized: true}
	arrayStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindDynamic)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreNewWithCapacity(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &ArrayState{Values: make([]Value, 0, capacity), Capacity: capacity, ValuesMaterialized: true}
	arrayStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindDynamic)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreNewReservedCapacity(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &ArrayState{Values: make([]Value, 0), Capacity: capacity, ValuesMaterialized: true}
	arrayStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindDynamic)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreState(handle int64) (*ArrayState, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	ensureArrayStore()
	return deoptTypedArrayToDynamic(handle)
}

func ArrayStoreEnsureHandle(handle int64, lengthHint int, capacityHint int) (*ArrayState, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	return arrayStoreEnsureExistingHandleLocked(handle, lengthHint, capacityHint)
}

// ArrayStoreAdoptHandle explicitly creates backing state for a host-provided
// handle that has never been published by ArrayStore. Normal runtime and
// compiler paths must use ArrayStoreEnsureHandle instead, so a stale released
// handle cannot be silently resurrected.
func ArrayStoreAdoptHandle(handle int64, lengthHint int, capacityHint int) (*ArrayState, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return nil, fmt.Errorf("array handle must be non-zero")
	}
	ensureArrayStore()
	if _, err := arrayHandleKindLocked(handle); err == nil {
		return arrayStoreEnsureExistingHandleLocked(handle, lengthHint, capacityHint)
	}
	if capacityHint < lengthHint {
		capacityHint = lengthHint
	}
	state := &ArrayState{Values: make([]Value, 0, capacityHint), Capacity: capacityHint, ValuesMaterialized: true}
	ArraySetLength(state, lengthHint)
	arrayStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindDynamic)
	cacheArrayHandleRevision(handle, &state.Revision)
	if handle >= arrayNextHandle {
		arrayNextHandle = handle + 1
	}
	return state, nil
}

func arrayStoreEnsureExistingHandleLocked(handle int64, lengthHint int, capacityHint int) (*ArrayState, error) {
	if handle == 0 {
		return nil, fmt.Errorf("array handle must be non-zero")
	}
	ensureArrayStore()
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, err
	}
	if kind != monoArrayKindDynamic {
		if _, err := deoptTypedArrayToDynamic(handle); err != nil {
			return nil, err
		}
	}
	state, ok := arrayStates[handle]
	if !ok {
		return nil, fmt.Errorf("array handle %d is not defined", handle)
	}
	if capacityHint > state.Capacity {
		ArrayEnsureCapacity(state, capacityHint)
	}
	if lengthHint > len(state.Values) {
		ArraySetLength(state, lengthHint)
	}
	if state.Capacity < len(state.Values) {
		state.Capacity = len(state.Values)
	}
	return state, nil
}

func ArrayStoreEnsure(arr *ArrayValue, capacityHint int) (*ArrayState, int64, error) {
	if arr == nil {
		return nil, 0, fmt.Errorf("array receiver is nil")
	}
	handle := arr.Handle
	if handle != 0 {
		lengthHint := len(arr.Elements)
		state, err := ArrayStoreEnsureHandle(handle, lengthHint, capacityHint)
		if err != nil {
			return nil, 0, err
		}
		arr.Elements = state.Values
		if err := ArrayStoreTrackArrayValueLease(arr, handle); err != nil {
			return nil, 0, err
		}
		return state, handle, nil
	}
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	ensureArrayStore()
	handle = allocateArrayHandle()
	values := arr.Elements
	if values == nil {
		values = make([]Value, 0)
	}
	capacity := len(values)
	if cap(values) > capacity {
		capacity = cap(values)
	}
	if capacityHint > capacity {
		capacity = capacityHint
	}
	state := &ArrayState{Values: values, Capacity: capacity, ValuesMaterialized: len(values) == 0}
	ArrayEnsureCapacity(state, capacity)
	arr.Elements = state.Values
	arr.Handle = handle
	arrayStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindDynamic)
	cacheArrayHandleRevision(handle, &state.Revision)
	if err := arrayStoreTrackArrayValueLeaseLocked(arr, handle); err != nil {
		return nil, 0, err
	}
	return state, handle, nil
}

func ArrayStoreValueFromHandle(handle int64, lengthHint int, capacityHint int) (*ArrayValue, *ArrayState, error) {
	if handle == 0 {
		return nil, nil, fmt.Errorf("array handle must be non-zero")
	}
	state, err := ArrayStoreEnsureHandle(handle, lengthHint, capacityHint)
	if err != nil {
		return nil, nil, err
	}
	arr := &ArrayValue{Handle: handle, Elements: state.Values}
	if err := ArrayStoreTrackArrayValueLease(arr, handle); err != nil {
		return nil, nil, err
	}
	return arr, state, nil
}
