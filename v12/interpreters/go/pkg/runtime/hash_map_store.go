package runtime

import "fmt"

var hashMapStates map[int64]*HashMapValue
var hashMapNextHandle int64 = 1

func ensureHashMapStore() {
	if hashMapStates == nil {
		hashMapStates = make(map[int64]*HashMapValue)
	}
	if hashMapNextHandle == 0 {
		hashMapNextHandle = 1
	}
}

func HashMapStoreNew(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	ensureHashMapStore()
	handle := hashMapNextHandle
	hashMapNextHandle++
	hashMapStates[handle] = &HashMapValue{Entries: make([]HashMapEntry, 0, capacity)}
	return handle
}

func HashMapStoreNewWithCapacity(capacity int) int64 {
	return HashMapStoreNew(capacity)
}

func HashMapStoreState(handle int64) (*HashMapValue, error) {
	ensureHashMapStore()
	state, ok := hashMapStates[handle]
	if !ok || state == nil {
		return nil, fmt.Errorf("hash map handle %d is not defined", handle)
	}
	return state, nil
}

func HashMapStoreEnsureHandle(handle int64, capacityHint int) (*HashMapValue, error) {
	if handle == 0 {
		return nil, fmt.Errorf("hash map handle must be non-zero")
	}
	if capacityHint < 0 {
		capacityHint = 0
	}
	ensureHashMapStore()
	if state, ok := hashMapStates[handle]; ok && state != nil {
		return state, nil
	}
	state := &HashMapValue{Entries: make([]HashMapEntry, 0, capacityHint)}
	hashMapStates[handle] = state
	return state, nil
}

func HashMapStoreClone(handle int64) (int64, error) {
	state, err := HashMapStoreState(handle)
	if err != nil {
		return 0, err
	}
	cloned := make([]HashMapEntry, len(state.Entries))
	copy(cloned, state.Entries)
	newHandle := HashMapStoreNewWithCapacity(len(cloned))
	hashMapStates[newHandle].Entries = cloned
	return newHandle, nil
}
