package runtime

func ArrayStoreMonoNewI32() int64 {
	return ArrayStoreMonoNewWithCapacityI32(0)
}

func ArrayStoreMonoNewI64() int64 {
	return ArrayStoreMonoNewWithCapacityI64(0)
}

func ArrayStoreMonoNewBool() int64 {
	return ArrayStoreMonoNewWithCapacityBool(0)
}

func ArrayStoreMonoNewChar() int64 {
	return ArrayStoreMonoNewWithCapacityChar(0)
}

func ArrayStoreMonoNewU8() int64 {
	return ArrayStoreMonoNewWithCapacityU8(0)
}

func ArrayStoreMonoNewU32() int64 {
	return ArrayStoreMonoNewWithCapacityU32(0)
}

func ArrayStoreMonoNewU64() int64 {
	return ArrayStoreMonoNewWithCapacityU64(0)
}

func ArrayStoreMonoNewF64() int64 {
	return ArrayStoreMonoNewWithCapacityF64(0)
}

func ArrayStoreMonoNewWithCapacityI32(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayI32State{
		Values:   make([]int32, 0, capacity),
		Capacity: capacity,
	}
	monoArrayI32States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindI32)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityI64(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayI64State{
		Values:   make([]int64, 0, capacity),
		Capacity: capacity,
	}
	monoArrayI64States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindI64)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityBool(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayBoolState{
		Values:   make([]bool, 0, capacity),
		Capacity: capacity,
	}
	monoArrayBoolStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindBool)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityChar(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayCharState{
		Values:   make([]rune, 0, capacity),
		Capacity: capacity,
	}
	monoArrayCharStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindChar)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityU8(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayU8State{
		Values:   make([]uint8, 0, capacity),
		Capacity: capacity,
	}
	monoArrayU8States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindU8)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityU32(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayU32State{
		Values:   make([]uint32, 0, capacity),
		Capacity: capacity,
	}
	monoArrayU32States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindU32)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityU64(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayU64State{
		Values:   make([]uint64, 0, capacity),
		Capacity: capacity,
	}
	monoArrayU64States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindU64)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}

func ArrayStoreMonoNewWithCapacityF64(capacity int) int64 {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	state := &monoArrayF64State{
		Values:   make([]float64, 0, capacity),
		Capacity: capacity,
	}
	monoArrayF64States[handle] = state
	recordArrayHandleKind(handle, monoArrayKindF64)
	cacheArrayHandleRevision(handle, &state.Revision)
	return handle
}
