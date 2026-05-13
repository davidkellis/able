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

func ArrayStoreMonoNewU8() int64 {
	return ArrayStoreMonoNewWithCapacityU8(0)
}

func ArrayStoreMonoNewF64() int64 {
	return ArrayStoreMonoNewWithCapacityF64(0)
}

func ArrayStoreMonoNewWithCapacityI32(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	monoArrayI32States[handle] = &monoArrayI32State{
		Values:   make([]int32, 0, capacity),
		Capacity: capacity,
	}
	arrayHandleKinds[handle] = monoArrayKindI32
	return handle
}

func ArrayStoreMonoNewWithCapacityI64(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	monoArrayI64States[handle] = &monoArrayI64State{
		Values:   make([]int64, 0, capacity),
		Capacity: capacity,
	}
	arrayHandleKinds[handle] = monoArrayKindI64
	return handle
}

func ArrayStoreMonoNewWithCapacityBool(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	monoArrayBoolStates[handle] = &monoArrayBoolState{
		Values:   make([]bool, 0, capacity),
		Capacity: capacity,
	}
	arrayHandleKinds[handle] = monoArrayKindBool
	return handle
}

func ArrayStoreMonoNewWithCapacityU8(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	monoArrayU8States[handle] = &monoArrayU8State{
		Values:   make([]uint8, 0, capacity),
		Capacity: capacity,
	}
	arrayHandleKinds[handle] = monoArrayKindU8
	return handle
}

func ArrayStoreMonoNewWithCapacityF64(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	monoArrayF64States[handle] = &monoArrayF64State{
		Values:   make([]float64, 0, capacity),
		Capacity: capacity,
	}
	arrayHandleKinds[handle] = monoArrayKindF64
	return handle
}
