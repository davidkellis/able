package runtime

import (
	"fmt"
	"sort"
)

// Environment provides lexical scoping for Able runtime values.
type Environment struct {
	values map[string]Value
	parent *Environment
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
	return e.parent
}

// Snapshot returns a deterministic copy of the current bindings.
func (e *Environment) Snapshot() map[string]Value {
	out := make(map[string]Value, len(e.values))
	for k, v := range e.values {
		out[k] = v
	}
	return out
}

// Define inserts or shadows a binding in the current scope.
func (e *Environment) Define(name string, value Value) {
	e.values[name] = value
}

// Assign updates an existing binding in the first scope where it appears.
func (e *Environment) Assign(name string, value Value) error {
	if _, ok := e.values[name]; ok {
		e.values[name] = value
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(name, value)
	}
	return fmt.Errorf("Undefined variable '%s'", name)
}

// Get retrieves a binding, searching outward through the scope chain.
func (e *Environment) Get(name string) (Value, error) {
	if v, ok := e.values[name]; ok {
		return v, nil
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, fmt.Errorf("Undefined variable '%s'", name)
}

// Keys returns the bindings in sorted order (useful for determinism in tests).
func (e *Environment) Keys() []string {
	keys := make([]string, 0, len(e.values))
	for k := range e.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Extend clones the current environment into a new child scope.
func (e *Environment) Extend() *Environment {
	return NewEnvironment(e)
}
