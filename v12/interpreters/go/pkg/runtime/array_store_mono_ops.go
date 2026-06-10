package runtime

import "fmt"

func ArrayStoreMonoReadI32(handle int64, index int) (int32, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindI32, monoArrayI32States)
	if err != nil || kind == monoArrayKindI32 {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return int32FromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono i32", handle)
}

func ArrayStoreMonoWriteI32(handle int64, index int, value int32) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindI32, monoArrayI32States, false)
	if err != nil || kind == monoArrayKindI32 {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, i32ToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono i32", handle)
}

func ArrayStoreMonoReadI64(handle int64, index int) (int64, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindI64, monoArrayI64States)
	if err != nil || kind == monoArrayKindI64 {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return int64FromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono i64", handle)
}

func ArrayStoreMonoWriteI64(handle int64, index int, value int64) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindI64, monoArrayI64States, false)
	if err != nil || kind == monoArrayKindI64 {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, i64ToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono i64", handle)
}

func ArrayStoreMonoReadBool(handle int64, index int) (bool, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindBool, monoArrayBoolStates)
	if err != nil || kind == monoArrayKindBool {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return false, err
		}
		return boolFromValue(boxed)
	}
	return false, fmt.Errorf("array handle %d is not mono bool", handle)
}

func ArrayStoreMonoWriteBool(handle int64, index int, value bool) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindBool, monoArrayBoolStates, false)
	if err != nil || kind == monoArrayKindBool {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, boolToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono bool", handle)
}

func ArrayStoreMonoReadChar(handle int64, index int) (rune, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindChar, monoArrayCharStates)
	if err != nil || kind == monoArrayKindChar {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return charFromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono char", handle)
}

func ArrayStoreMonoWriteChar(handle int64, index int, value rune) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindChar, monoArrayCharStates, false)
	if err != nil || kind == monoArrayKindChar {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, charToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono char", handle)
}

func ArrayStoreMonoReadU8(handle int64, index int) (uint8, error) {
	value, kind, err := arrayStoreMonoReadValue(handle, index, monoArrayKindU8, monoArrayU8States)
	if err != nil || kind == monoArrayKindU8 {
		return value, err
	}
	if kind == monoArrayKindDynamic {
		boxed, err := ArrayStoreRead(handle, index)
		if err != nil {
			return 0, err
		}
		return u8FromValue(boxed)
	}
	return 0, fmt.Errorf("array handle %d is not mono u8", handle)
}

func ArrayStoreMonoWriteU8(handle int64, index int, value uint8) error {
	kind, err := arrayStoreMonoWriteValue(handle, index, value, monoArrayKindU8, monoArrayU8States, false)
	if err != nil || kind == monoArrayKindU8 {
		return err
	}
	if kind == monoArrayKindDynamic {
		return ArrayStoreWrite(handle, index, u8ToValue(value))
	}
	return fmt.Errorf("array handle %d is not mono u8", handle)
}
