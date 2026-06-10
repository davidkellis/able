package runtime

import (
	"fmt"
	"sync"
)

var arrayHandleRevisionHotHandle int64
var arrayHandleRevisionHot *uint64
var arrayHandleRevisionMu sync.Mutex

func cacheArrayHandleRevision(handle int64, revision *uint64) {
	ensureArrayStore()
	if handle == 0 {
		return
	}
	if revision == nil {
		delete(arrayHandleRevisions, handle)
		if arrayHandleRevisionHotHandle == handle {
			arrayHandleRevisionHotHandle = 0
			arrayHandleRevisionHot = nil
		}
		return
	}
	arrayHandleRevisions[handle] = revision
	arrayHandleRevisionHotHandle = handle
	arrayHandleRevisionHot = revision
}

func arrayHandleRevisionPointer(handle int64) (*uint64, bool, error) {
	if handle == 0 {
		return nil, false, nil
	}
	if arrayHandleRevisionHotHandle == handle && arrayHandleRevisionHot != nil {
		return arrayHandleRevisionHot, true, nil
	}
	ensureArrayStore()
	if revision, ok := arrayHandleRevisions[handle]; ok && revision != nil {
		arrayHandleRevisionHotHandle = handle
		arrayHandleRevisionHot = revision
		return revision, true, nil
	}
	kind, err := arrayHandleKindLocked(handle)
	if err != nil {
		return nil, false, err
	}
	switch kind {
	case monoArrayKindDynamic:
		state, ok := arrayStates[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindI32:
		state, ok := monoArrayI32States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindI64:
		state, ok := monoArrayI64States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindBool:
		state, ok := monoArrayBoolStates[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindChar:
		state, ok := monoArrayCharStates[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindU8:
		state, ok := monoArrayU8States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindU32:
		state, ok := monoArrayU32States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindU64:
		state, ok := monoArrayU64States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	case monoArrayKindF64:
		state, ok := monoArrayF64States[handle]
		if !ok {
			return nil, false, fmt.Errorf("array handle %d is not defined", handle)
		}
		return &state.Revision, true, nil
	default:
		return nil, false, fmt.Errorf("array handle %d has unknown kind", handle)
	}
}
