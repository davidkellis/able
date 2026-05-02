package runtime

import "fmt"

func ArrayStoreMonoValueFromU8String(text string) *ArrayValue {
	ensureArrayStore()
	handle := allocateArrayHandle()
	values := make([]uint8, len(text))
	copy(values, text)
	monoArrayU8States[handle] = &monoArrayU8State{
		Values:   values,
		Capacity: len(values),
	}
	arrayHandleKinds[handle] = monoArrayKindU8
	return &ArrayValue{Handle: handle, TrackedHandle: handle}
}

func ArrayStoreMonoReadU8IfAvailable(handle int64, index int) (uint8, bool, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, false, err
	}
	if kind != monoArrayKindU8 {
		return 0, false, nil
	}
	state, ok := monoArrayU8States[handle]
	if !ok {
		return 0, false, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return 0, false, nil
	}
	return state.Values[index], true, nil
}
