package runtime

import "fmt"

func arrayStateCapacityForMonoPromotion(state *ArrayState) int {
	if state == nil {
		return 0
	}
	capacity := state.Capacity
	if capacity < len(state.Values) {
		capacity = len(state.Values)
	}
	return capacity
}

func ArrayStorePromoteHandleToMonoTypeIfPossible(handle int64, typeName string) (bool, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	if handle == 0 || typeName == "" {
		return false, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return false, err
	}
	switch typeName {
	case string(IntegerI32):
		if kind == monoArrayKindI32 {
			return true, nil
		}
	case string(IntegerI64):
		if kind == monoArrayKindI64 {
			return true, nil
		}
	case "bool":
		if kind == monoArrayKindBool {
			return true, nil
		}
	case "char":
		if kind == monoArrayKindChar {
			return true, nil
		}
	case string(IntegerU8):
		if kind == monoArrayKindU8 {
			return true, nil
		}
	case string(IntegerU32):
		if kind == monoArrayKindU32 {
			return true, nil
		}
	case string(IntegerU64):
		if kind == monoArrayKindU64 {
			return true, nil
		}
	case string(FloatF64):
		if kind == monoArrayKindF64 {
			return true, nil
		}
	default:
		return false, nil
	}
	if kind != monoArrayKindDynamic {
		return false, nil
	}
	state, ok := arrayStates[handle]
	if !ok {
		return false, fmt.Errorf("array handle %d is not defined", handle)
	}
	capacity := arrayStateCapacityForMonoPromotion(state)
	nextRevision := state.Revision + 1
	switch typeName {
	case string(IntegerI32):
		values := make([]int32, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := int32FromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayI32State{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayI32States[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindI32)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case string(IntegerI64):
		values := make([]int64, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := int64FromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayI64State{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayI64States[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindI64)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case "bool":
		values := make([]bool, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := boolFromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayBoolState{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayBoolStates[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindBool)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case "char":
		values := make([]rune, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := charFromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayCharState{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayCharStates[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindChar)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case string(IntegerU8):
		values := make([]uint8, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := u8FromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayU8State{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayU8States[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindU8)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case string(IntegerU32):
		values := make([]uint32, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := u32FromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayU32State{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayU32States[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindU32)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case string(IntegerU64):
		values := make([]uint64, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := u64FromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayU64State{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayU64States[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindU64)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	case string(FloatF64):
		values := make([]float64, len(state.Values), capacity)
		for idx, current := range state.Values {
			raw, err := float64FromValue(current)
			if err != nil {
				return false, nil
			}
			values[idx] = raw
		}
		delete(arrayStates, handle)
		mono := &monoArrayF64State{Values: values, Capacity: capacity, Revision: nextRevision}
		monoArrayF64States[handle] = mono
		recordArrayHandleKind(handle, monoArrayKindF64)
		cacheArrayHandleRevision(handle, &mono.Revision)
		return true, nil
	default:
		return false, nil
	}
}
