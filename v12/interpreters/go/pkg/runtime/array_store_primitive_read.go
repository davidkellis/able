package runtime

import (
	"fmt"
	"sync"
)

type ArrayStoreMonoPrimitiveReadKind uint8

const (
	ArrayStoreMonoPrimitiveReadNone ArrayStoreMonoPrimitiveReadKind = iota
	ArrayStoreMonoPrimitiveReadI32
	ArrayStoreMonoPrimitiveReadI64
	ArrayStoreMonoPrimitiveReadBool
	ArrayStoreMonoPrimitiveReadChar
	ArrayStoreMonoPrimitiveReadU8
	ArrayStoreMonoPrimitiveReadU32
	ArrayStoreMonoPrimitiveReadU64
	ArrayStoreMonoPrimitiveReadF64
)

type ArrayStoreMonoPrimitiveReadInfo struct {
	Kind     ArrayStoreMonoPrimitiveReadKind
	Size     int
	InBounds bool
	Int64    int64
	Uint64   uint64
	Float64  float64
	Bool     bool
}

var monoArrayI32ReadHotHandle int64
var monoArrayI32ReadHot *monoArrayI32State
var monoArrayI64ReadHotHandle int64
var monoArrayI64ReadHot *monoArrayI64State
var monoArrayBoolReadHotHandle int64
var monoArrayBoolReadHot *monoArrayBoolState
var monoArrayCharReadHotHandle int64
var monoArrayCharReadHot *monoArrayCharState
var monoArrayU8ReadHotHandle int64
var monoArrayU8ReadHot *monoArrayU8State
var monoArrayU32ReadHotHandle int64
var monoArrayU32ReadHot *monoArrayU32State
var monoArrayU64ReadHotHandle int64
var monoArrayU64ReadHot *monoArrayU64State
var monoArrayF64ReadHotHandle int64
var monoArrayF64ReadHot *monoArrayF64State
var monoArrayPrimitiveReadInfoHotHandle int64
var monoArrayPrimitiveReadInfoHotKind monoArrayKind
var monoArrayPrimitiveReadInfoHotOK bool
var monoArrayPrimitiveReadInfoHotI32 *monoArrayI32State
var monoArrayPrimitiveReadInfoHotI64 *monoArrayI64State
var monoArrayPrimitiveReadInfoHotBool *monoArrayBoolState
var monoArrayPrimitiveReadInfoHotChar *monoArrayCharState
var monoArrayPrimitiveReadInfoHotU8 *monoArrayU8State
var monoArrayPrimitiveReadInfoHotU32 *monoArrayU32State
var monoArrayPrimitiveReadInfoHotU64 *monoArrayU64State
var monoArrayPrimitiveReadInfoHotF64 *monoArrayF64State
var monoArrayPrimitiveReadMu sync.Mutex

func clearMonoArrayPrimitiveReadHot(handle int64, kind monoArrayKind) {
	if handle == 0 {
		return
	}
	if monoArrayPrimitiveReadInfoHotOK &&
		monoArrayPrimitiveReadInfoHotHandle == handle &&
		monoArrayPrimitiveReadInfoHotKind != kind {
		clearMonoArrayPrimitiveReadInfoHot()
	}
	if kind != monoArrayKindI32 && monoArrayI32ReadHotHandle == handle {
		monoArrayI32ReadHotHandle, monoArrayI32ReadHot = 0, nil
	}
	if kind != monoArrayKindI64 && monoArrayI64ReadHotHandle == handle {
		monoArrayI64ReadHotHandle, monoArrayI64ReadHot = 0, nil
	}
	if kind != monoArrayKindBool && monoArrayBoolReadHotHandle == handle {
		monoArrayBoolReadHotHandle, monoArrayBoolReadHot = 0, nil
	}
	if kind != monoArrayKindChar && monoArrayCharReadHotHandle == handle {
		monoArrayCharReadHotHandle, monoArrayCharReadHot = 0, nil
	}
	if kind != monoArrayKindChar {
		clearMonoArrayCharAppendCache(handle)
	}
	if kind != monoArrayKindU8 && monoArrayU8ReadHotHandle == handle {
		monoArrayU8ReadHotHandle, monoArrayU8ReadHot = 0, nil
	}
	if kind != monoArrayKindU32 && monoArrayU32ReadHotHandle == handle {
		monoArrayU32ReadHotHandle, monoArrayU32ReadHot = 0, nil
	}
	if kind != monoArrayKindU64 && monoArrayU64ReadHotHandle == handle {
		monoArrayU64ReadHotHandle, monoArrayU64ReadHot = 0, nil
	}
	if kind != monoArrayKindF64 && monoArrayF64ReadHotHandle == handle {
		monoArrayF64ReadHotHandle, monoArrayF64ReadHot = 0, nil
	}
}

func clearMonoArrayPrimitiveReadInfoHot() {
	monoArrayPrimitiveReadInfoHotHandle = 0
	monoArrayPrimitiveReadInfoHotKind = monoArrayKindDynamic
	monoArrayPrimitiveReadInfoHotOK = false
	monoArrayPrimitiveReadInfoHotI32 = nil
	monoArrayPrimitiveReadInfoHotI64 = nil
	monoArrayPrimitiveReadInfoHotBool = nil
	monoArrayPrimitiveReadInfoHotChar = nil
	monoArrayPrimitiveReadInfoHotU8 = nil
	monoArrayPrimitiveReadInfoHotU32 = nil
	monoArrayPrimitiveReadInfoHotU64 = nil
	monoArrayPrimitiveReadInfoHotF64 = nil
}

func cachedMonoArrayI32ReadState(handle int64) (*monoArrayI32State, bool) {
	if monoArrayI32ReadHotHandle == handle && monoArrayI32ReadHot != nil {
		return monoArrayI32ReadHot, true
	}
	state, ok := monoArrayI32States[handle]
	if ok {
		monoArrayI32ReadHotHandle, monoArrayI32ReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayI64ReadState(handle int64) (*monoArrayI64State, bool) {
	if monoArrayI64ReadHotHandle == handle && monoArrayI64ReadHot != nil {
		return monoArrayI64ReadHot, true
	}
	state, ok := monoArrayI64States[handle]
	if ok {
		monoArrayI64ReadHotHandle, monoArrayI64ReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayBoolReadState(handle int64) (*monoArrayBoolState, bool) {
	if monoArrayBoolReadHotHandle == handle && monoArrayBoolReadHot != nil {
		return monoArrayBoolReadHot, true
	}
	state, ok := monoArrayBoolStates[handle]
	if ok {
		monoArrayBoolReadHotHandle, monoArrayBoolReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayCharReadState(handle int64) (*monoArrayCharState, bool) {
	if monoArrayCharReadHotHandle == handle && monoArrayCharReadHot != nil {
		return monoArrayCharReadHot, true
	}
	state, ok := monoArrayCharStates[handle]
	if ok {
		monoArrayCharReadHotHandle, monoArrayCharReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayU8ReadState(handle int64) (*monoArrayU8State, bool) {
	if monoArrayU8ReadHotHandle == handle && monoArrayU8ReadHot != nil {
		return monoArrayU8ReadHot, true
	}
	state, ok := monoArrayU8States[handle]
	if ok {
		monoArrayU8ReadHotHandle, monoArrayU8ReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayU32ReadState(handle int64) (*monoArrayU32State, bool) {
	if monoArrayU32ReadHotHandle == handle && monoArrayU32ReadHot != nil {
		return monoArrayU32ReadHot, true
	}
	state, ok := monoArrayU32States[handle]
	if ok {
		monoArrayU32ReadHotHandle, monoArrayU32ReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayU64ReadState(handle int64) (*monoArrayU64State, bool) {
	if monoArrayU64ReadHotHandle == handle && monoArrayU64ReadHot != nil {
		return monoArrayU64ReadHot, true
	}
	state, ok := monoArrayU64States[handle]
	if ok {
		monoArrayU64ReadHotHandle, monoArrayU64ReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayF64ReadState(handle int64) (*monoArrayF64State, bool) {
	if monoArrayF64ReadHotHandle == handle && monoArrayF64ReadHot != nil {
		return monoArrayF64ReadHot, true
	}
	state, ok := monoArrayF64States[handle]
	if ok {
		monoArrayF64ReadHotHandle, monoArrayF64ReadHot = handle, state
	}
	return state, ok
}

func cachedMonoArrayPrimitiveReadInfoKind(handle int64) (monoArrayKind, bool) {
	if monoArrayPrimitiveReadInfoHotOK && monoArrayPrimitiveReadInfoHotHandle == handle {
		return monoArrayPrimitiveReadInfoHotKind, true
	}
	return monoArrayKindDynamic, false
}

func rememberMonoArrayPrimitiveReadInfoKind(handle int64, kind monoArrayKind) {
	if handle == 0 {
		return
	}
	if monoArrayPrimitiveReadInfoHotOK &&
		monoArrayPrimitiveReadInfoHotHandle == handle &&
		monoArrayPrimitiveReadInfoHotKind == kind {
		return
	}
	clearMonoArrayPrimitiveReadInfoHot()
	monoArrayPrimitiveReadInfoHotHandle = handle
	monoArrayPrimitiveReadInfoHotKind = kind
	monoArrayPrimitiveReadInfoHotOK = true
}

func rememberMonoArrayPrimitiveReadInfoI32State(handle int64, state *monoArrayI32State) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindI32)
	monoArrayPrimitiveReadInfoHotI32 = state
}

func rememberMonoArrayPrimitiveReadInfoI64State(handle int64, state *monoArrayI64State) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindI64)
	monoArrayPrimitiveReadInfoHotI64 = state
}

func rememberMonoArrayPrimitiveReadInfoBoolState(handle int64, state *monoArrayBoolState) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindBool)
	monoArrayPrimitiveReadInfoHotBool = state
}

func rememberMonoArrayPrimitiveReadInfoCharState(handle int64, state *monoArrayCharState) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindChar)
	monoArrayPrimitiveReadInfoHotChar = state
}

func rememberMonoArrayPrimitiveReadInfoU8State(handle int64, state *monoArrayU8State) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindU8)
	monoArrayPrimitiveReadInfoHotU8 = state
}

func rememberMonoArrayPrimitiveReadInfoU32State(handle int64, state *monoArrayU32State) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindU32)
	monoArrayPrimitiveReadInfoHotU32 = state
}

func rememberMonoArrayPrimitiveReadInfoU64State(handle int64, state *monoArrayU64State) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindU64)
	monoArrayPrimitiveReadInfoHotU64 = state
}

func rememberMonoArrayPrimitiveReadInfoF64State(handle int64, state *monoArrayF64State) {
	rememberMonoArrayPrimitiveReadInfoKind(handle, monoArrayKindF64)
	monoArrayPrimitiveReadInfoHotF64 = state
}

func fillMonoArrayI32PrimitiveReadInfo(state *monoArrayI32State, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadI32
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Int64 = int64(state.Values[index])
	}
}

func fillMonoArrayI64PrimitiveReadInfo(state *monoArrayI64State, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadI64
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Int64 = state.Values[index]
	}
}

func fillMonoArrayBoolPrimitiveReadInfo(state *monoArrayBoolState, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadBool
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Bool = state.Values[index]
	}
}

func fillMonoArrayCharPrimitiveReadInfo(state *monoArrayCharState, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadChar
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Int64 = int64(state.Values[index])
	}
}

func fillMonoArrayU8PrimitiveReadInfo(state *monoArrayU8State, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadU8
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Uint64 = uint64(state.Values[index])
	}
}

func fillMonoArrayU32PrimitiveReadInfo(state *monoArrayU32State, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadU32
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Uint64 = uint64(state.Values[index])
	}
}

func fillMonoArrayU64PrimitiveReadInfo(state *monoArrayU64State, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadU64
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Uint64 = state.Values[index]
	}
}

func fillMonoArrayF64PrimitiveReadInfo(state *monoArrayF64State, index int, info *ArrayStoreMonoPrimitiveReadInfo) {
	info.Kind = ArrayStoreMonoPrimitiveReadF64
	info.Size = len(state.Values)
	if index >= 0 && index < info.Size {
		info.InBounds = true
		info.Float64 = state.Values[index]
	}
}

// ArrayStoreMonoPrimitiveReadInfoInto reports direct mono primitive array read
// data into caller-owned storage without first materializing boxed values.
func ArrayStoreMonoPrimitiveReadInfoInto(handle int64, index int, info *ArrayStoreMonoPrimitiveReadInfo) (bool, error) {
	if info == nil {
		return false, fmt.Errorf("array primitive read info destination is nil")
	}
	*info = ArrayStoreMonoPrimitiveReadInfo{}
	return ArrayStoreMonoPrimitiveReadInfoIntoFresh(handle, index, info)
}

// ArrayStoreMonoPrimitiveReadInfoIntoFresh is the same direct mono primitive
// read path as ArrayStoreMonoPrimitiveReadInfoInto for callers that pass fresh
// zeroed storage. If it returns false, the destination is left as-is.
func ArrayStoreMonoPrimitiveReadInfoIntoFresh(handle int64, index int, info *ArrayStoreMonoPrimitiveReadInfo) (bool, error) {
	if info == nil {
		return false, fmt.Errorf("array primitive read info destination is nil")
	}
	if handle == 0 {
		return false, nil
	}
	arrayStoreMu.RLock()
	defer arrayStoreMu.RUnlock()
	monoArrayPrimitiveReadMu.Lock()
	defer monoArrayPrimitiveReadMu.Unlock()
	kind, ok := cachedMonoArrayPrimitiveReadInfoKind(handle)
	if !ok {
		var err error
		kind, err = arrayHandleKindLocked(handle)
		if err != nil {
			return false, err
		}
		rememberMonoArrayPrimitiveReadInfoKind(handle, kind)
	}
	switch kind {
	case monoArrayKindI32:
		state := monoArrayPrimitiveReadInfoHotI32
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayI32ReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoI32State(handle, state)
		}
		fillMonoArrayI32PrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindI64:
		state := monoArrayPrimitiveReadInfoHotI64
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayI64ReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoI64State(handle, state)
		}
		fillMonoArrayI64PrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindBool:
		state := monoArrayPrimitiveReadInfoHotBool
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayBoolReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoBoolState(handle, state)
		}
		fillMonoArrayBoolPrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindChar:
		state := monoArrayPrimitiveReadInfoHotChar
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayCharReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoCharState(handle, state)
		}
		fillMonoArrayCharPrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindU8:
		state := monoArrayPrimitiveReadInfoHotU8
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayU8ReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoU8State(handle, state)
		}
		fillMonoArrayU8PrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindU32:
		state := monoArrayPrimitiveReadInfoHotU32
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayU32ReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoU32State(handle, state)
		}
		fillMonoArrayU32PrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindU64:
		state := monoArrayPrimitiveReadInfoHotU64
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayU64ReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoU64State(handle, state)
		}
		fillMonoArrayU64PrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindF64:
		state := monoArrayPrimitiveReadInfoHotF64
		if state == nil {
			var ok bool
			state, ok = cachedMonoArrayF64ReadState(handle)
			if !ok {
				return false, fmt.Errorf("array handle %d is not defined", handle)
			}
			rememberMonoArrayPrimitiveReadInfoF64State(handle, state)
		}
		fillMonoArrayF64PrimitiveReadInfo(state, index, info)
		return true, nil
	case monoArrayKindDynamic:
		return false, nil
	default:
		return false, nil
	}
}

// ArrayStoreMonoPrimitiveReadInfoIfAvailable reports direct mono primitive
// array read data without first materializing boxed values or repeating handle
// kind lookups in the caller.
func ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle int64, index int) (ArrayStoreMonoPrimitiveReadInfo, bool, error) {
	var info ArrayStoreMonoPrimitiveReadInfo
	ok, err := ArrayStoreMonoPrimitiveReadInfoInto(handle, index, &info)
	return info, ok, err
}
