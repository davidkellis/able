package runtime

import (
	"fmt"
)

type ArrayState struct {
	Values                []Value
	Capacity              int
	ElementTypeToken      uint16
	ElementTypeTokenKnown bool
}

type monoArrayKind uint8

const (
	monoArrayKindDynamic monoArrayKind = iota
	monoArrayKindI32
	monoArrayKindI64
	monoArrayKindBool
	monoArrayKindU8
)

type monoArrayState[T any] struct {
	Values   []T
	Capacity int
}

type monoArrayI32State = monoArrayState[int32]
type monoArrayI64State = monoArrayState[int64]
type monoArrayBoolState = monoArrayState[bool]
type monoArrayU8State = monoArrayState[uint8]

var arrayStates map[int64]*ArrayState
var monoArrayI32States map[int64]*monoArrayI32State
var monoArrayI64States map[int64]*monoArrayI64State
var monoArrayBoolStates map[int64]*monoArrayBoolState
var monoArrayU8States map[int64]*monoArrayU8State
var arrayHandleKinds map[int64]monoArrayKind
var arrayNextHandle int64 = 1

func ensureArrayStore() {
	if arrayStates == nil {
		arrayStates = make(map[int64]*ArrayState)
	}
	if monoArrayI32States == nil {
		monoArrayI32States = make(map[int64]*monoArrayI32State)
	}
	if monoArrayI64States == nil {
		monoArrayI64States = make(map[int64]*monoArrayI64State)
	}
	if monoArrayBoolStates == nil {
		monoArrayBoolStates = make(map[int64]*monoArrayBoolState)
	}
	if monoArrayU8States == nil {
		monoArrayU8States = make(map[int64]*monoArrayU8State)
	}
	if arrayHandleKinds == nil {
		arrayHandleKinds = make(map[int64]monoArrayKind)
	}
	if arrayNextHandle <= 0 {
		arrayNextHandle = 1
	}
}

func allocateArrayHandle() int64 {
	ensureArrayStore()
	handle := arrayNextHandle
	arrayNextHandle++
	return handle
}

func grownCapacity(current int, minimum int) int {
	if minimum <= 0 {
		return current
	}
	if current >= minimum {
		return current
	}
	if current < 4 {
		current = 4
	}
	for current < minimum {
		if current < 1024 {
			current *= 2
		} else {
			current += current / 2
		}
	}
	return current
}

func ArrayEnsureCapacity(state *ArrayState, minimum int) bool {
	if state == nil {
		return false
	}
	if minimum <= state.Capacity {
		return false
	}
	newCapacity := grownCapacity(state.Capacity, minimum)
	if newCapacity < minimum {
		newCapacity = minimum
	}
	newValues := make([]Value, len(state.Values), newCapacity)
	copy(newValues, state.Values)
	state.Values = newValues
	state.Capacity = newCapacity
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
	ArrayEnsureCapacity(state, length)
	for len(state.Values) < length {
		state.Values = append(state.Values, NilValue{})
	}
	if len(state.Values) > state.Capacity {
		state.Capacity = len(state.Values)
	}
}

func monoEnsureCapacity[T any](state *monoArrayState[T], minimum int) bool {
	if state == nil {
		return false
	}
	if minimum <= state.Capacity {
		return false
	}
	newCapacity := grownCapacity(state.Capacity, minimum)
	if newCapacity < minimum {
		newCapacity = minimum
	}
	newValues := make([]T, len(state.Values), newCapacity)
	copy(newValues, state.Values)
	state.Values = newValues
	state.Capacity = newCapacity
	return true
}

func monoSetLength[T any](state *monoArrayState[T], length int) {
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
	monoEnsureCapacity(state, length)
	for len(state.Values) < length {
		var zero T
		state.Values = append(state.Values, zero)
	}
	if len(state.Values) > state.Capacity {
		state.Capacity = len(state.Values)
	}
}

func arrayHandleKind(handle int64) (monoArrayKind, error) {
	ensureArrayStore()
	if handle == 0 {
		return monoArrayKindDynamic, fmt.Errorf("array handle must be non-zero")
	}
	if kind, ok := arrayHandleKinds[handle]; ok {
		return kind, nil
	}
	if _, ok := arrayStates[handle]; ok {
		arrayHandleKinds[handle] = monoArrayKindDynamic
		return monoArrayKindDynamic, nil
	}
	if _, ok := monoArrayI32States[handle]; ok {
		arrayHandleKinds[handle] = monoArrayKindI32
		return monoArrayKindI32, nil
	}
	if _, ok := monoArrayI64States[handle]; ok {
		arrayHandleKinds[handle] = monoArrayKindI64
		return monoArrayKindI64, nil
	}
	if _, ok := monoArrayBoolStates[handle]; ok {
		arrayHandleKinds[handle] = monoArrayKindBool
		return monoArrayKindBool, nil
	}
	if _, ok := monoArrayU8States[handle]; ok {
		arrayHandleKinds[handle] = monoArrayKindU8
		return monoArrayKindU8, nil
	}
	return monoArrayKindDynamic, fmt.Errorf("array handle %d is not defined", handle)
}

func int64FromValue(value Value) (int64, error) {
	switch v := value.(type) {
	case IntegerValue:
		if n, ok := v.ToInt64(); ok {
			return n, nil
		}
		return 0, fmt.Errorf("array element integer is out of range")
	case *IntegerValue:
		if v == nil {
			return 0, fmt.Errorf("array element integer is nil")
		}
		if n, ok := v.ToInt64(); ok {
			return n, nil
		}
		return 0, fmt.Errorf("array element integer is out of range")
	default:
		return 0, fmt.Errorf("array element must be an integer")
	}
}

func boolFromValue(value Value) (bool, error) {
	switch v := value.(type) {
	case BoolValue:
		return v.Val, nil
	case *BoolValue:
		if v == nil {
			return false, fmt.Errorf("array element must be a bool")
		}
		return v.Val, nil
	default:
		return false, fmt.Errorf("array element must be a bool")
	}
}

func int32FromValue(value Value) (int32, error) {
	raw, err := int64FromValue(value)
	if err != nil {
		return 0, err
	}
	if raw < -2147483648 || raw > 2147483647 {
		return 0, fmt.Errorf("array element is out of i32 range")
	}
	return int32(raw), nil
}

func u8FromValue(value Value) (uint8, error) {
	raw, err := int64FromValue(value)
	if err != nil {
		return 0, err
	}
	if raw < 0 || raw > 255 {
		return 0, fmt.Errorf("array element is out of u8 range")
	}
	return uint8(raw), nil
}

func i32ToValue(v int32) Value {
	return NewSmallInt(int64(v), IntegerI32)
}

func i64ToValue(v int64) Value {
	return NewSmallInt(v, IntegerI64)
}

func boolToValue(v bool) Value {
	return BoolValue{Val: v}
}

func u8ToValue(v uint8) Value {
	return NewSmallInt(int64(v), IntegerU8)
}

func deoptTypedArrayToDynamic(handle int64) (*ArrayState, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return nil, err
	}
	if kind == monoArrayKindDynamic {
		state, ok := arrayStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state, nil
	}
	var state *ArrayState
	switch kind {
	case monoArrayKindI32:
		mono, ok := monoArrayI32States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = i32ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity}
		delete(monoArrayI32States, handle)
	case monoArrayKindI64:
		mono, ok := monoArrayI64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = i64ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity}
		delete(monoArrayI64States, handle)
	case monoArrayKindBool:
		mono, ok := monoArrayBoolStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = boolToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity}
		delete(monoArrayBoolStates, handle)
	case monoArrayKindU8:
		mono, ok := monoArrayU8States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = u8ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity}
		delete(monoArrayU8States, handle)
	default:
		return nil, fmt.Errorf("array handle %d has unknown kind", handle)
	}
	arrayStates[handle] = state
	arrayHandleKinds[handle] = monoArrayKindDynamic
	return state, nil
}

func ArrayStoreNew() int64 {
	handle := allocateArrayHandle()
	arrayStates[handle] = &ArrayState{Values: make([]Value, 0), Capacity: 0}
	arrayHandleKinds[handle] = monoArrayKindDynamic
	return handle
}

func ArrayStoreNewWithCapacity(capacity int) int64 {
	if capacity < 0 {
		capacity = 0
	}
	handle := allocateArrayHandle()
	arrayStates[handle] = &ArrayState{Values: make([]Value, 0, capacity), Capacity: capacity}
	arrayHandleKinds[handle] = monoArrayKindDynamic
	return handle
}

func ArrayStoreMonoNewI32() int64 {
	return ArrayStoreMonoNewWithCapacityI32(0)
}

func ArrayStoreMonoNewI64() int64 {
	return ArrayStoreMonoNewWithCapacityI64(0)
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

func ArrayStoreMonoNewBool() int64 {
	return ArrayStoreMonoNewWithCapacityBool(0)
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

func ArrayStoreMonoNewU8() int64 {
	return ArrayStoreMonoNewWithCapacityU8(0)
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

func ArrayStoreState(handle int64) (*ArrayState, error) {
	ensureArrayStore()
	return deoptTypedArrayToDynamic(handle)
}

func ArrayStoreEnsureHandle(handle int64, lengthHint int, capacityHint int) (*ArrayState, error) {
	if handle == 0 {
		return nil, fmt.Errorf("array handle must be non-zero")
	}
	ensureArrayStore()
	if handle >= arrayNextHandle {
		arrayNextHandle = handle + 1
	}
	kind, err := arrayHandleKind(handle)
	if err != nil {
		if capacityHint < lengthHint {
			capacityHint = lengthHint
		}
		state := &ArrayState{Values: make([]Value, 0, capacityHint), Capacity: capacityHint}
		ArraySetLength(state, lengthHint)
		arrayStates[handle] = state
		arrayHandleKinds[handle] = monoArrayKindDynamic
		return state, nil
	}
	if kind != monoArrayKindDynamic {
		if _, err := deoptTypedArrayToDynamic(handle); err != nil {
			return nil, err
		}
	}
	state, ok := arrayStates[handle]
	if !ok {
		return nil, fmt.Errorf("array handle %d is not defined", handle)
	}
	if capacityHint > state.Capacity {
		ArrayEnsureCapacity(state, capacityHint)
	}
	if lengthHint > len(state.Values) {
		ArraySetLength(state, lengthHint)
	}
	if state.Capacity < len(state.Values) {
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
		lengthHint := len(arr.Elements)
		if state, err := ArrayStoreEnsureHandle(handle, lengthHint, capacityHint); err == nil {
			arr.Elements = state.Values
			return state, handle, nil
		}
	}
	if handle == 0 {
		handle = allocateArrayHandle()
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
	arrayHandleKinds[handle] = monoArrayKindDynamic
	if handle >= arrayNextHandle {
		arrayNextHandle = handle + 1
	}
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
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	default:
		return 0, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreCapacity(handle int64) (int, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return 0, err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	default:
		return 0, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreSetLength(handle int64, length int) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		ArrayEnsureCapacity(state, length)
		ArraySetLength(state, length)
		return nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, length)
		monoSetLength(state, length)
		return nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, length)
		monoSetLength(state, length)
		return nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, length)
		monoSetLength(state, length)
		return nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, length)
		monoSetLength(state, length)
		return nil
	default:
		return fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreRead(handle int64, index int) (Value, error) {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return nil, err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return state.Values[index], nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return i32ToValue(state.Values[index]), nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return i64ToValue(state.Values[index]), nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return boolToValue(state.Values[index]), nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return u8ToValue(state.Values[index]), nil
	default:
		return nil, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreWrite(handle int64, index int, value Value) error {
	kind, err := arrayHandleKind(handle)
	if err != nil {
		return err
	}
	if index < 0 {
		return fmt.Errorf("index must be non-negative")
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		length := len(state.Values)
		if index == length {
			if length == 0 && state.Capacity < 4 {
				ArrayEnsureCapacity(state, 4)
			}
			state.Values = append(state.Values, value)
			if state.Capacity < cap(state.Values) {
				state.Capacity = cap(state.Values)
			}
			return nil
		}
		ArrayEnsureCapacity(state, index+1)
		if index > length {
			ArraySetLength(state, index+1)
		}
		state.Values[index] = value
		return nil
	case monoArrayKindI32:
		typed, err := int32FromValue(value)
		if err != nil {
			return err
		}
		return ArrayStoreMonoWriteI32(handle, index, typed)
	case monoArrayKindI64:
		typed, err := int64FromValue(value)
		if err != nil {
			return err
		}
		return ArrayStoreMonoWriteI64(handle, index, typed)
	case monoArrayKindBool:
		typed, err := boolFromValue(value)
		if err != nil {
			return err
		}
		return ArrayStoreMonoWriteBool(handle, index, typed)
	case monoArrayKindU8:
		typed, err := u8FromValue(value)
		if err != nil {
			return err
		}
		return ArrayStoreMonoWriteU8(handle, index, typed)
	default:
		return fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreReserve(handle int64, capacity int) error {
	kind, err := arrayHandleKind(handle)
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
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
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
	kind, err := arrayHandleKind(handle)
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
		arrayStates[newHandle] = &ArrayState{Values: cloned, Capacity: state.Capacity}
		arrayHandleKinds[newHandle] = monoArrayKindDynamic
		return newHandle, nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]int32, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		monoArrayI32States[newHandle] = &monoArrayI32State{Values: cloned, Capacity: state.Capacity}
		arrayHandleKinds[newHandle] = monoArrayKindI32
		return newHandle, nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]int64, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		monoArrayI64States[newHandle] = &monoArrayI64State{Values: cloned, Capacity: state.Capacity}
		arrayHandleKinds[newHandle] = monoArrayKindI64
		return newHandle, nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]bool, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		monoArrayBoolStates[newHandle] = &monoArrayBoolState{Values: cloned, Capacity: state.Capacity}
		arrayHandleKinds[newHandle] = monoArrayKindBool
		return newHandle, nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		cloned := make([]uint8, len(state.Values))
		copy(cloned, state.Values)
		newHandle := allocateArrayHandle()
		monoArrayU8States[newHandle] = &monoArrayU8State{Values: cloned, Capacity: state.Capacity}
		arrayHandleKinds[newHandle] = monoArrayKindU8
		return newHandle, nil
	default:
		return 0, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}
