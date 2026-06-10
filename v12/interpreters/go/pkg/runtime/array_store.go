package runtime

import (
	"fmt"
	"sync"
)

var arrayStates map[int64]*ArrayState
var monoArrayI32States map[int64]*monoArrayI32State
var monoArrayI64States map[int64]*monoArrayI64State
var monoArrayBoolStates map[int64]*monoArrayBoolState
var monoArrayCharStates map[int64]*monoArrayCharState
var monoArrayU8States map[int64]*monoArrayU8State
var monoArrayU32States map[int64]*monoArrayU32State
var monoArrayU64States map[int64]*monoArrayU64State
var monoArrayF64States map[int64]*monoArrayF64State
var arrayHandleKinds map[int64]monoArrayKind
var arrayHandleRevisions map[int64]*uint64
var arrayNextHandle int64 = 1

// arrayStoreMu protects the process-wide array-handle registry and the state
// reachable through it. Handles are shared by all interpreter instances, so a
// spawned tree-walker task can otherwise race another task merely by creating
// or reading an unrelated array.
var arrayStoreMu sync.RWMutex

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
	if monoArrayCharStates == nil {
		monoArrayCharStates = make(map[int64]*monoArrayCharState)
	}
	if monoArrayU8States == nil {
		monoArrayU8States = make(map[int64]*monoArrayU8State)
	}
	if monoArrayU32States == nil {
		monoArrayU32States = make(map[int64]*monoArrayU32State)
	}
	if monoArrayU64States == nil {
		monoArrayU64States = make(map[int64]*monoArrayU64State)
	}
	if monoArrayF64States == nil {
		monoArrayF64States = make(map[int64]*monoArrayF64State)
	}
	if arrayHandleKinds == nil {
		arrayHandleKinds = make(map[int64]monoArrayKind)
	}
	if arrayHandleRevisions == nil {
		arrayHandleRevisions = make(map[int64]*uint64)
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
		if current < 4096 {
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
	if minimum <= state.Capacity && minimum <= cap(state.Values) {
		return false
	}
	newCapacity := state.Capacity
	if newCapacity < len(state.Values) {
		newCapacity = len(state.Values)
	}
	if newCapacity < minimum {
		newCapacity = grownCapacity(newCapacity, minimum)
	}
	if newCapacity < minimum {
		newCapacity = minimum
	}
	if newCapacity < len(state.Values) {
		newCapacity = len(state.Values)
	}
	newValues := make([]Value, len(state.Values), newCapacity)
	copy(newValues, state.Values)
	state.Values = newValues
	if state.Capacity < newCapacity {
		state.Capacity = newCapacity
	}
	return true
}

func ArraySetLength(state *ArrayState, length int) {
	if state == nil || length < 0 {
		return
	}
	oldLength := len(state.Values)
	adjustArrayI32CacheLength(state, oldLength, length)
	if length <= oldLength {
		state.Values = state.Values[:length]
		if len(state.Values) > state.Capacity {
			state.Capacity = len(state.Values)
		}
		if length != oldLength {
			state.Revision++
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
	state.Revision++
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
	state.Revision++
	return true
}

func monoSetLength[T any](state *monoArrayState[T], length int) {
	if state == nil || length < 0 {
		return
	}
	if length <= len(state.Values) {
		oldLength := len(state.Values)
		state.Values = state.Values[:length]
		if len(state.Values) > state.Capacity {
			state.Capacity = len(state.Values)
		}
		if length != oldLength {
			state.Revision++
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
	state.Revision++
}

// arrayStoreMonoReadValue protects the shared primitive-handle registry while
// reading a concrete primitive array. A caller can route a dynamic handle
// through the boxed ArrayStore path after the registry lock is released.
func arrayStoreMonoReadValue[T any](handle int64, index int, expected monoArrayKind, states map[int64]*monoArrayState[T]) (T, monoArrayKind, error) {
	var zero T
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return zero, monoArrayKindDynamic, err
	}
	if kind != expected {
		return zero, kind, nil
	}
	state, ok := states[handle]
	if !ok {
		return zero, kind, fmt.Errorf("array handle %d is not defined", handle)
	}
	if index < 0 || index >= len(state.Values) {
		return zero, kind, fmt.Errorf("index out of bounds")
	}
	return state.Values[index], kind, nil
}

// arrayStoreMonoWriteValue serializes primitive array mutations with registry
// publication, promotion, and deoptimization. It returns a non-matching kind
// to let the typed wrapper preserve the boxed dynamic fallback.
func arrayStoreMonoWriteValue[T any](handle int64, index int, value T, expected monoArrayKind, states map[int64]*monoArrayState[T], incrementRevision bool) (monoArrayKind, error) {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	return arrayStoreMonoWriteValueLocked(handle, index, value, expected, states, incrementRevision)
}

func arrayStoreMonoWriteValueLocked[T any](handle int64, index int, value T, expected monoArrayKind, states map[int64]*monoArrayState[T], incrementRevision bool) (monoArrayKind, error) {
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return monoArrayKindDynamic, err
	}
	if kind != expected {
		return kind, nil
	}
	if index < 0 {
		return kind, fmt.Errorf("index must be non-negative")
	}
	state, ok := states[handle]
	if !ok {
		return kind, fmt.Errorf("array handle %d is not defined", handle)
	}
	monoEnsureCapacity(state, index+1)
	if index >= len(state.Values) {
		monoSetLength(state, index+1)
	}
	state.Values[index] = value
	if incrementRevision {
		state.Revision++
	}
	return kind, nil
}

func arrayHandleKind(handle int64) (monoArrayKind, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	return arrayHandleKindLocked(handle)
}

// arrayHandleKindLocked reads the handle registry while arrayStoreMu is held.
// Constructors and promotions publish a handle before it can be observed, so
// this lookup never needs to fill the cache on a read path.
func arrayHandleKindLocked(handle int64) (monoArrayKind, error) {
	if handle == 0 {
		return monoArrayKindDynamic, fmt.Errorf("array handle must be non-zero")
	}
	if kind, ok := cachedArrayHandleKind(handle); ok && arrayHandleHasKindBackingLocked(handle, kind) {
		return kind, nil
	}
	if kind, ok := arrayHandleKinds[handle]; ok && arrayHandleHasKindBackingLocked(handle, kind) {
		return kind, nil
	}
	if _, ok := arrayStates[handle]; ok {
		return monoArrayKindDynamic, nil
	}
	if _, ok := monoArrayI32States[handle]; ok {
		return monoArrayKindI32, nil
	}
	if _, ok := monoArrayI64States[handle]; ok {
		return monoArrayKindI64, nil
	}
	if _, ok := monoArrayBoolStates[handle]; ok {
		return monoArrayKindBool, nil
	}
	if _, ok := monoArrayCharStates[handle]; ok {
		return monoArrayKindChar, nil
	}
	if _, ok := monoArrayU8States[handle]; ok {
		return monoArrayKindU8, nil
	}
	if _, ok := monoArrayU32States[handle]; ok {
		return monoArrayKindU32, nil
	}
	if _, ok := monoArrayU64States[handle]; ok {
		return monoArrayKindU64, nil
	}
	if _, ok := monoArrayF64States[handle]; ok {
		return monoArrayKindF64, nil
	}
	return monoArrayKindDynamic, fmt.Errorf("array handle %d is not defined", handle)
}

func arrayHandleHasKindBackingLocked(handle int64, kind monoArrayKind) bool {
	switch kind {
	case monoArrayKindDynamic:
		_, ok := arrayStates[handle]
		return ok
	case monoArrayKindI32:
		_, ok := monoArrayI32States[handle]
		return ok
	case monoArrayKindI64:
		_, ok := monoArrayI64States[handle]
		return ok
	case monoArrayKindBool:
		_, ok := monoArrayBoolStates[handle]
		return ok
	case monoArrayKindChar:
		_, ok := monoArrayCharStates[handle]
		return ok
	case monoArrayKindU8:
		_, ok := monoArrayU8States[handle]
		return ok
	case monoArrayKindU32:
		_, ok := monoArrayU32States[handle]
		return ok
	case monoArrayKindU64:
		_, ok := monoArrayU64States[handle]
		return ok
	case monoArrayKindF64:
		_, ok := monoArrayF64States[handle]
		return ok
	default:
		return false
	}
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

func charFromValue(value Value) (rune, error) {
	switch v := value.(type) {
	case CharValue:
		return v.Val, nil
	case *CharValue:
		if v == nil {
			return 0, fmt.Errorf("array element must be a char")
		}
		return v.Val, nil
	default:
		return 0, fmt.Errorf("array element must be a char")
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

func float64FromValue(value Value) (float64, error) {
	switch v := value.(type) {
	case FloatValue:
		if v.TypeSuffix == FloatF64 {
			return v.Val, nil
		}
	case *FloatValue:
		if v != nil && v.TypeSuffix == FloatF64 {
			return v.Val, nil
		}
	}
	return 0, fmt.Errorf("array element must be an f64")
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

func charToValue(v rune) Value {
	return CharValue{Val: v}
}

func u8ToValue(v uint8) Value {
	return NewSmallInt(int64(v), IntegerU8)
}

func f64ToValue(v float64) Value {
	return FloatValue{Val: v, TypeSuffix: FloatF64}
}

func deoptTypedArrayToDynamic(handle int64) (*ArrayState, error) {
	kind, err := arrayHandleKindLocked(handle)
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
	var sourceRevision uint64
	switch kind {
	case monoArrayKindI32:
		mono, ok := monoArrayI32States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = i32ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		seedArrayI32CacheFromMonoValues(state, mono.Values)
		delete(monoArrayI32States, handle)
	case monoArrayKindI64:
		mono, ok := monoArrayI64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = i64ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		delete(monoArrayI64States, handle)
	case monoArrayKindBool:
		mono, ok := monoArrayBoolStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = boolToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		delete(monoArrayBoolStates, handle)
	case monoArrayKindChar:
		mono, ok := monoArrayCharStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = charToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		delete(monoArrayCharStates, handle)
	case monoArrayKindU8:
		mono, ok := monoArrayU8States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = u8ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		delete(monoArrayU8States, handle)
	case monoArrayKindU32:
		mono, ok := monoArrayU32States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = u32ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		delete(monoArrayU32States, handle)
	case monoArrayKindU64:
		mono, ok := monoArrayU64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = u64ToValue(value)
		}
		state = &ArrayState{Values: values, Capacity: mono.Capacity, ValuesMaterialized: true}
		delete(monoArrayU64States, handle)
	case monoArrayKindF64:
		mono, ok := monoArrayF64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		sourceRevision = mono.Revision
		values := make([]Value, len(mono.Values))
		for idx, value := range mono.Values {
			values[idx] = f64ToValue(value)
		}
		state = &ArrayState{
			Values:                values,
			Capacity:              mono.Capacity,
			ValuesMaterialized:    true,
			ElementTypeToken:      0,
			ElementTypeTokenKnown: false,
		}
		delete(monoArrayF64States, handle)
	default:
		return nil, fmt.Errorf("array handle %d has unknown kind", handle)
	}
	state.Revision = sourceRevision + 1
	arrayStates[handle] = state
	recordArrayHandleKind(handle, monoArrayKindDynamic)
	cacheArrayHandleRevision(handle, &state.Revision)
	return state, nil
}

func ArrayStoreSize(handle int64) (int, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
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
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
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
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return len(state.Values), nil
	default:
		return 0, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreCapacity(handle int64) (int, error) {
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
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
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
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
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
		if !ok {
			return 0, fmt.Errorf("array handle %d is not defined", handle)
		}
		return state.Capacity, nil
	default:
		return 0, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreSetLength(handle int64, length int) error {
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
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
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
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, length)
		monoSetLength(state, length)
		return nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return fmt.Errorf("array handle %d is not defined", handle)
		}
		monoEnsureCapacity(state, length)
		monoSetLength(state, length)
		return nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
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
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	kind, err := arrayHandleKindLocked(handle)
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
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return charToValue(state.Values[index]), nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return u8ToValue(state.Values[index]), nil
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return u32ToValue(state.Values[index]), nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return u64ToValue(state.Values[index]), nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
		if !ok {
			return nil, fmt.Errorf("array handle %d is not defined", handle)
		}
		if index < 0 || index >= len(state.Values) {
			return NilValue{}, nil
		}
		return f64ToValue(state.Values[index]), nil
	default:
		return nil, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}

func ArrayStoreWrite(handle int64, index int, value Value) error {
	arrayStoreMu.Lock()
	defer arrayStoreMu.Unlock()
	kind, err := arrayHandleKindLocked(handle)
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
			if length+1 > state.Capacity || length == cap(state.Values) {
				ArrayEnsureCapacity(state, length+1)
			}
			state.Values = append(state.Values, value)
			if state.Capacity < cap(state.Values) {
				state.Capacity = cap(state.Values)
			}
			updateArrayI32CacheForDynamicWrite(state, length, index, value)
			state.Revision++
			return nil
		}
		ArrayEnsureCapacity(state, index+1)
		if index > length {
			ArraySetLength(state, index+1)
		}
		state.Values[index] = value
		updateArrayI32CacheForDynamicWrite(state, length, index, value)
		state.Revision++
		return nil
	case monoArrayKindI32:
		typed, err := int32FromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindI32, monoArrayI32States, false)
		return err
	case monoArrayKindI64:
		typed, err := int64FromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindI64, monoArrayI64States, false)
		return err
	case monoArrayKindBool:
		typed, err := boolFromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindBool, monoArrayBoolStates, false)
		return err
	case monoArrayKindChar:
		typed, err := charFromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindChar, monoArrayCharStates, false)
		return err
	case monoArrayKindU8:
		typed, err := u8FromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindU8, monoArrayU8States, false)
		return err
	case monoArrayKindU32:
		typed, err := u32FromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindU32, monoArrayU32States, false)
		return err
	case monoArrayKindU64:
		typed, err := u64FromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindU64, monoArrayU64States, false)
		return err
	case monoArrayKindF64:
		typed, err := float64FromValue(value)
		if err != nil {
			return err
		}
		_, err = arrayStoreMonoWriteValueLocked(handle, index, typed, monoArrayKindF64, monoArrayF64States, true)
		return err
	default:
		return fmt.Errorf("array handle %d has unknown kind", handle)
	}
}
