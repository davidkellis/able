package runtime

import "fmt"

type ArrayState struct {
	Values   []Value
	Capacity int
}

var arrayStates map[int64]*ArrayState
var arrayNextHandle int64 = 1

func ensureArrayStore() {
	if arrayStates == nil {
		arrayStates = make(map[int64]*ArrayState)
	}
	if arrayNextHandle == 0 {
		arrayNextHandle = 1
	}
}

func ArrayEnsureCapacity(state *ArrayState, minimum int) bool {
	if state == nil {
		return false
	}
	if minimum <= state.Capacity {
		return false
	}
	newValues := make([]Value, len(state.Values), minimum)
	copy(newValues, state.Values)
	state.Values = newValues
	state.Capacity = minimum
	return true
}

func ArraySetLength(state *ArrayState, length int) {
	if state == nil || length < 0 {
		return
	}
	if length <= len(state.Values) {
		state.Values = state.Values[:length]
		if len(state.Values) > state.Capacity {
			state.Capacity = len(state.Values)
		}
		return
	}
	for len(state.Values) < length {
		state.Values = append(state.Values, NilValue{})
	}
	if len(state.Values) > state.Capacity {
		state.Capacity = len(state.Values)
	}
}

func ArrayStoreNew() int64 {
	ensureArrayStore()
	handle := arrayNextHandle
	arrayNextHandle++
	arrayStates[handle] = &ArrayState{Values: make([]Value, 0), Capacity: 0}
	return handle
}

func ArrayStoreNewWithCapacity(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	ensureArrayStore()
	handle := arrayNextHandle
	arrayNextHandle++
	arrayStates[handle] = &ArrayState{Values: make([]Value, 0, capacity), Capacity: capacity}
	return handle
}

func ArrayStoreState(handle int64) (*ArrayState, error) {
	ensureArrayStore()
	state, ok := arrayStates[handle]
	if !ok {
		return nil, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state, nil
}

func ArrayStoreEnsureHandle(handle int64, lengthHint int, capacityHint int) (*ArrayState, error) {
	if handle == 0 {
		return nil, fmt.Errorf("array handle must be non-zero")
	}
	ensureArrayStore()
	state, ok := arrayStates[handle]
	if !ok {
		if capacityHint < lengthHint {
			capacityHint = lengthHint
		}
		state = &ArrayState{Values: make([]Value, 0, capacityHint), Capacity: capacityHint}
		ArraySetLength(state, lengthHint)
		arrayStates[handle] = state
		return state, nil
	}
	updated := false
	if capacityHint > state.Capacity {
		updated = ArrayEnsureCapacity(state, capacityHint)
	}
	if lengthHint > len(state.Values) {
		ArraySetLength(state, lengthHint)
		updated = true
	}
	if updated && state.Capacity < len(state.Values) {
		state.Capacity = len(state.Values)
	}
	return state, nil
}

func ArrayStoreEnsure(arr *ArrayValue, capacityHint int) (*ArrayState, int64, error) {
	if arr == nil {
		return nil, 0, fmt.Errorf("array receiver is nil")
	}
	ensureArrayStore()
	handle := arr.Handle
	if handle != 0 {
		if state, ok := arrayStates[handle]; ok {
			if capacityHint > state.Capacity {
				ArrayEnsureCapacity(state, capacityHint)
			}
			arr.Elements = state.Values
			return state, handle, nil
		}
	}
	if handle == 0 {
		handle = arrayNextHandle
		arrayNextHandle++
	}
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
	state := &ArrayState{Values: values, Capacity: capacity}
	ArrayEnsureCapacity(state, capacity)
	arr.Elements = state.Values
	arr.Handle = handle
	arrayStates[handle] = state
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
	return arr, state, nil
}

func ArrayStoreSize(handle int64) (int, error) {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return 0, err
	}
	return len(state.Values), nil
}

func ArrayStoreCapacity(handle int64) (int, error) {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return 0, err
	}
	return state.Capacity, nil
}

func ArrayStoreSetLength(handle int64, length int) error {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return err
	}
	ArrayEnsureCapacity(state, length)
	ArraySetLength(state, length)
	return nil
}

func ArrayStoreRead(handle int64, index int) (Value, error) {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(state.Values) {
		return NilValue{}, nil
	}
	return state.Values[index], nil
}

func ArrayStoreWrite(handle int64, index int, value Value) error {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	ArrayEnsureCapacity(state, index+1)
	if index >= len(state.Values) {
		ArraySetLength(state, index+1)
	}
	state.Values[index] = value
	return nil
}

func ArrayStoreReserve(handle int64, capacity int) error {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return err
	}
	ArrayEnsureCapacity(state, capacity)
	return nil
}

func ArrayStoreClone(handle int64) (int64, error) {
	state, err := ArrayStoreState(handle)
	if err != nil {
		return 0, err
	}
	cloned := make([]Value, len(state.Values))
	copy(cloned, state.Values)
	ensureArrayStore()
	newHandle := arrayNextHandle
	arrayNextHandle++
	arrayStates[newHandle] = &ArrayState{Values: cloned, Capacity: state.Capacity}
	return newHandle, nil
}
