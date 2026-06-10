package runtime

import "sync"

type environmentMeta struct {
	structs map[string]*StructDefinitionValue
	data    any
}

type environmentState struct {
	mu   sync.RWMutex
	meta environmentMeta
}

func (e *Environment) mutex() *sync.RWMutex {
	if e == nil {
		return nil
	}
	if state := e.state.Load(); state != nil {
		return &state.mu
	}
	state := &environmentState{}
	if e.state.CompareAndSwap(nil, state) {
		return &state.mu
	}
	return &e.state.Load().mu
}

func (e *Environment) ensureMetaNoLock() *environmentMeta {
	if state := e.state.Load(); state != nil {
		return &state.meta
	}
	state := &environmentState{}
	if e.state.CompareAndSwap(nil, state) {
		return &state.meta
	}
	return &e.state.Load().meta
}

func (e *Environment) metaNoLock() *environmentMeta {
	if e == nil {
		return nil
	}
	state := e.state.Load()
	if state == nil {
		return nil
	}
	return &state.meta
}
