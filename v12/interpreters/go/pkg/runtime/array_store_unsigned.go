package runtime

import (
	"fmt"
	"math/big"
)

func u32FromValue(value Value) (uint32, error) {
	raw, err := u64FromValue(value)
	if err != nil {
		return 0, err
	}
	if raw > uint64(^uint32(0)) {
		return 0, fmt.Errorf("array element is out of u32 range")
	}
	return uint32(raw), nil
}

func u64FromValue(value Value) (uint64, error) {
	switch v := value.(type) {
	case IntegerValue:
		return u64FromInteger(v)
	case *IntegerValue:
		if v == nil {
			return 0, fmt.Errorf("array element integer is nil")
		}
		return u64FromInteger(*v)
	default:
		return 0, fmt.Errorf("array element must be an integer")
	}
}

func u64FromInteger(value IntegerValue) (uint64, error) {
	if value.Sign() < 0 {
		return 0, fmt.Errorf("array element is out of u64 range")
	}
	if value.IsSmall() {
		return uint64(value.Int64Fast()), nil
	}
	raw := value.BigInt()
	if !raw.IsUint64() {
		return 0, fmt.Errorf("array element is out of u64 range")
	}
	return raw.Uint64(), nil
}

func u32ToValue(v uint32) Value {
	return NewSmallInt(int64(v), IntegerU32)
}

func u64ToValue(v uint64) Value {
	if v <= uint64(^uint64(0)>>1) {
		return NewSmallInt(int64(v), IntegerU64)
	}
	return NewBigIntValue(new(big.Int).SetUint64(v), IntegerU64)
}

func ArrayStoreMonoReadU32(handle int64, index int) (uint32, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindU32, monoArrayU32States)
	if err != nil || kind == monoArrayKindU32 {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return u32FromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono u32", handle)
}

func ArrayStoreMonoWriteU32(handle int64, index int, value uint32) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindU32, monoArrayU32States, false)
	if err != nil || kind == monoArrayKindU32 {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, u32ToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono u32", handle)
}

func ArrayStoreMonoReadU64(handle int64, index int) (uint64, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindU64, monoArrayU64States)
	if err != nil || kind == monoArrayKindU64 {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return u64FromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono u64", handle)
}

func ArrayStoreMonoWriteU64(handle int64, index int, value uint64) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindU64, monoArrayU64States, false)
	if err != nil || kind == monoArrayKindU64 {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, u64ToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono u64", handle)
}

func arrayStoreMonoU32State(handle int64) (*monoArrayU32State, bool, error) {
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, false, err
	}
	if kind != monoArrayKindU32 {
		return nil, false, nil
	}
	state, ok := monoArrayU32States[handle]
	if !ok {
		return nil, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state, true, nil
}

func ArrayStoreMonoReadU32IfAvailable(handle int64, index int) (uint32, bool, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	state, ok, err := arrayStoreMonoU32State(handle)
	if err != nil || !ok {
		return 0, ok, err
	}
	if index < 0 || index >= len(state.Values) {
		return 0, false, nil
	}
	return state.Values[index], true, nil
}

func ArrayStoreAppendU32Promote(handle int64, value uint32) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindU32 {
		state, ok := cachedMonoArrayU32ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoU32Value(state, value)
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
	values := make([]uint32, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := u32FromValue(current)
		if err != nil {
			return false, nil
		}
		values[idx] = raw
	}
	mono := &monoArrayU32State{Values: values, Capacity: capacity}
	appendMonoU32Value(mono, value)
	delete(arrayStates, handle)
	monoArrayU32States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindU32)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func arrayStoreMonoU64State(handle int64) (*monoArrayU64State, bool, error) {
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, false, err
	}
	if kind != monoArrayKindU64 {
		return nil, false, nil
	}
	state, ok := monoArrayU64States[handle]
	if !ok {
		return nil, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state, true, nil
}

func ArrayStoreMonoReadU64IfAvailable(handle int64, index int) (uint64, bool, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	state, ok, err := arrayStoreMonoU64State(handle)
	if err != nil || !ok {
		return 0, ok, err
	}
	if index < 0 || index >= len(state.Values) {
		return 0, false, nil
	}
	return state.Values[index], true, nil
}

func ArrayStoreAppendU64Promote(handle int64, value uint64) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindU64 {
		state, ok := cachedMonoArrayU64ReadState(handle)
		if !ok {
			return false, fmt.Errorf("array handle %d is not defined", handle)
		}
		appendMonoU64Value(state, value)
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
	values := make([]uint64, len(state.Values), capacity)
	for idx, current := range state.Values {
		raw, err := u64FromValue(current)
		if err != nil {
			return false, nil
		}
		values[idx] = raw
	}
	mono := &monoArrayU64State{Values: values, Capacity: capacity}
	appendMonoU64Value(mono, value)
	delete(arrayStates, handle)
	monoArrayU64States[handle] = mono
	recordArrayHandleKind(handle, monoArrayKindU64)
	cacheArrayHandleRevision(handle, &mono.Revision)
	return true, nil
}

func appendMonoU32Value(state *monoArrayU32State, value uint32) {
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

func appendMonoU64Value(state *monoArrayU64State, value uint64) {
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
