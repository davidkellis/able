package runtime

import "fmt"

func ArrayStoreMonoReadI32(handle int64, index int) (int32, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, err
	}
	if kind == monoArrayKindDynamic {
		value, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return int32FromValue(value)
	}
	if kind != monoArrayKindI32 {
		return 0, fmt.Errorf("array handle %d is not mono i32", handle)
	}
	state, ok := monoArrayI32States[handle]
	if !ok {
		return 0, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return 0, fmt.Errorf("index out of bounds")
	}
	return state.Values[index], nil
}

func ArrayStoreMonoWriteI32(handle int64, index int, value int32) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, i32ToValue(value))
	}
	if kind != monoArrayKindI32 {
		return fmt.Errorf("array handle %d is not mono i32", handle)
	}
	state, ok := monoArrayI32States[handle]
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

func ArrayStoreMonoReadI64(handle int64, index int) (int64, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, err
	}
	if kind == monoArrayKindDynamic {
		value, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return int64FromValue(value)
	}
	if kind != monoArrayKindI64 {
		return 0, fmt.Errorf("array handle %d is not mono i64", handle)
	}
	state, ok := monoArrayI64States[handle]
	if !ok {
		return 0, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return 0, fmt.Errorf("index out of bounds")
	}
	return state.Values[index], nil
}

func ArrayStoreMonoWriteI64(handle int64, index int, value int64) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, i64ToValue(value))
	}
	if kind != monoArrayKindI64 {
		return fmt.Errorf("array handle %d is not mono i64", handle)
	}
	state, ok := monoArrayI64States[handle]
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

func ArrayStoreMonoReadBool(handle int64, index int) (bool, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return false, err
	}
	if kind == monoArrayKindDynamic {
		value, err := ArrayStoreRead(handle, index)
		if err != nil {
			return false, err
		}
		return boolFromValue(value)
	}
	if kind != monoArrayKindBool {
		return false, fmt.Errorf("array handle %d is not mono bool", handle)
	}
	state, ok := monoArrayBoolStates[handle]
	if !ok {
		return false, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return false, fmt.Errorf("index out of bounds")
	}
	return state.Values[index], nil
}

func ArrayStoreMonoWriteBool(handle int64, index int, value bool) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, boolToValue(value))
	}
	if kind != monoArrayKindBool {
		return fmt.Errorf("array handle %d is not mono bool", handle)
	}
	state, ok := monoArrayBoolStates[handle]
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

func ArrayStoreMonoReadU8(handle int64, index int) (uint8, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, err
	}
	if kind == monoArrayKindDynamic {
		value, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return u8FromValue(value)
	}
	if kind != monoArrayKindU8 {
		return 0, fmt.Errorf("array handle %d is not mono u8", handle)
	}
	state, ok := monoArrayU8States[handle]
	if !ok {
		return 0, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return 0, fmt.Errorf("index out of bounds")
	}
	return state.Values[index], nil
}

func ArrayStoreMonoWriteU8(handle int64, index int, value uint8) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, u8ToValue(value))
	}
	if kind != monoArrayKindU8 {
		return fmt.Errorf("array handle %d is not mono u8", handle)
	}
	state, ok := monoArrayU8States[handle]
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
