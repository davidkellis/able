package runtime

import (
	"fmt"
	"sync"
)

type ChannelState struct {
	Capacity int
	Ch       chan Value
	CloseCh  chan struct{}

	mu     sync.Mutex
	closed bool
}

var channelStates map[int64]*ChannelState
var channelNextHandle int64 = 1
var channelMu sync.Mutex

func ensureChannelStore() {
	if channelStates == nil {
		channelStates = make(map[int64]*ChannelState)
	}
	if channelNextHandle == 0 {
		channelNextHandle = 1
	}
}

func ChannelStoreNew(capacity int) (int64, error) {
	if capacity < 0 {
		return 0, fmt.Errorf("channel capacity must be non-negative")
	}
	ensureChannelStore()
	channelMu.Lock()
	handle := channelNextHandle
	channelNextHandle++
	state := &ChannelState{
		Capacity: capacity,
		Ch:       make(chan Value, capacity),
		CloseCh:  make(chan struct{}),
	}
	channelStates[handle] = state
	channelMu.Unlock()
	return handle, nil
}

func ChannelStoreState(handle int64) (*ChannelState, error) {
	if handle == 0 {
		return nil, fmt.Errorf("channel handle must be non-zero")
	}
	ensureChannelStore()
	channelMu.Lock()
	state, ok := channelStates[handle]
	channelMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("channel handle %d is not defined", handle)
	}
	return state, nil
}

func ChannelStoreIsClosed(handle int64) (bool, error) {
	state, err := ChannelStoreState(handle)
	if err != nil {
		return false, err
	}
	state.mu.Lock()
	closed := state.closed
	state.mu.Unlock()
	return closed, nil
}

func ChannelStoreClose(handle int64) (bool, error) {
	state, err := ChannelStoreState(handle)
	if err != nil {
		return false, err
	}
	state.mu.Lock()
	if state.closed {
		state.mu.Unlock()
		return false, nil
	}
	state.closed = true
	close(state.CloseCh)
	state.mu.Unlock()
	return true, nil
}
