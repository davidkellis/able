package runtime

import (
	"fmt"
	"sync"
)

type MutexState struct {
	Mu      sync.Mutex
	Cond    *sync.Cond
	Locked  bool
	Waiters int
}

var mutexStates map[int64]*MutexState
var mutexNextHandle int64 = 1
var mutexMu sync.Mutex

func ensureMutexStore() {
	if mutexStates == nil {
		mutexStates = make(map[int64]*MutexState)
	}
	if mutexNextHandle == 0 {
		mutexNextHandle = 1
	}
}

func MutexStoreNew() int64 {
	ensureMutexStore()
	mutexMu.Lock()
	handle := mutexNextHandle
	mutexNextHandle++
	state := &MutexState{}
	state.Cond = sync.NewCond(&state.Mu)
	mutexStates[handle] = state
	mutexMu.Unlock()
	return handle
}

func MutexStoreState(handle int64) (*MutexState, error) {
	if handle <= 0 {
		return nil, fmt.Errorf("mutex handle must be positive")
	}
	ensureMutexStore()
	mutexMu.Lock()
	state, ok := mutexStates[handle]
	mutexMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown mutex handle %d", handle)
	}
	return state, nil
}
