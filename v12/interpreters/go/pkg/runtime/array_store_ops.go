package runtime

import "fmt"

func ArrayStoreReserve(handle int64, capacity int) error {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		ArrayEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, capacity)
		return nil
	default:
		return fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreClone(handle int64) (int64, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return 0, err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]Value, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &ArrayState{
			Values:             cloned,
			Capacity:           state.Capacity,
			ValuesMaterialized: state.ValuesMaterialized,
		}
		arrayStates[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindDynamic)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]int32, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayI32State{Values: cloned, Capacity: state.Capacity}
		monoArrayI32States[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindI32)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]int64, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayI64State{Values: cloned, Capacity: state.Capacity}
		monoArrayI64States[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindI64)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]bool, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayBoolState{Values: cloned, Capacity: state.Capacity}
		monoArrayBoolStates[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindBool)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]rune, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayCharState{Values: cloned, Capacity: state.Capacity}
		monoArrayCharStates[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindChar)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]uint8, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayU8State{Values: cloned, Capacity: state.Capacity}
		monoArrayU8States[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindU8)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]uint32, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayU32State{Values: cloned, Capacity: state.Capacity}
		monoArrayU32States[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindU32)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]uint64, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayU64State{Values: cloned, Capacity: state.Capacity}
		monoArrayU64States[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindU64)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]float64, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		newState := &monoArrayF64State{Values: cloned, Capacity: state.Capacity}
		monoArrayF64States[newHandle] = newState
		recordArrayHandleKind(newHandle, monoArrayKindF64)
		cacheArrayHandleRevision(newHandle, &newState.Revision)
		return newHandle, nil
	default:
		return 0, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}
