package runtime

import (
	"fmt"
	"sort"
	"sync"
)

// Environment provides lexical scoping for Able runtime values.
type Environment struct {
	values map[string]Value
	parent *Environment
	mu     sync.RWMutex
	data   any
}

// NewEnvironment creates a new environment, optionally nested under a parent.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		values: make(map[string]Value),
		parent: parent,
	}
}

// Parent exposes the lexical parent (nil when global).
func (e *Environment) Parent() *Environment {
	e.mu.RLock()
	parent := e.parent
	e.mu.RUnlock()
	return parent
}

// Snapshot returns a deterministic copy of the current bindings.
func (e *Environment) Snapshot() map[string]Value {
	e.mu.RLock()
	out := make(map[string]Value, len(e.values))
	for k, v := range e.values {
		out[k] = v
	}
	e.mu.RUnlock()
	return out
}

// Define inserts or shadows a binding in the current scope.
func (e *Environment) Define(name string, value Value) {
	e.mu.Lock()
	e.values[name] = value
	e.mu.Unlock()
}

// Assign updates an existing binding in the first scope where it appears.
func (e *Environment) Assign(name string, value Value) error {
	e.mu.Lock()
	if _, ok := e.values[name]; ok {
		e.values[name] = value
		e.mu.Unlock()
		return nil
	}
	parent := e.parent
	e.mu.Unlock()
	if parent != nil {
		return parent.Assign(name, value)
	}
	return fmt.Errorf("Undefined variable '%s'", name)
}

// Get retrieves a binding, searching outward through the scope chain.
func (e *Environment) Get(name string) (Value, error) {
	e.mu.RLock()
	if v, ok := e.values[name]; ok {
		e.mu.RUnlock()
		return v, nil
	}
	parent := e.parent
	e.mu.RUnlock()
	if parent != nil {
		return parent.Get(name)
	}
	return nil, fmt.Errorf("Undefined variable '%s'", name)
}

// Keys returns the bindings in sorted order (useful for determinism in tests).
func (e *Environment) Keys() []string {
	e.mu.RLock()
	keys := make([]string, 0, len(e.values))
	for k := range e.values {
		keys = append(keys, k)
	}
	e.mu.RUnlock()
	sort.Strings(keys)
	return keys
}

// Extend clones the current environment into a new child scope.
func (e *Environment) Extend() *Environment {
	return NewEnvironment(e)
}

// SetRuntimeData attaches interpreter-specific metadata to the environment.
func (e *Environment) SetRuntimeData(data any) {
	e.mu.Lock()
	e.data = data
	e.mu.Unlock()
}

// RuntimeData returns the metadata associated with this environment, falling back to parents.
func (e *Environment) RuntimeData() any {
	e.mu.RLock()
	data := e.data
	parent := e.parent
	e.mu.RUnlock()
	if data != nil {
		return data
	}
	if parent != nil {
		return parent.RuntimeData()
	}
	return nil
}
