package runtime

import (
	"fmt"
	"sort"
	"sync"
)

// Environment provides lexical scoping for Able runtime values.
type Environment struct {
	values       map[string]Value
	structs      map[string]*StructDefinitionValue
	parent       *Environment
	mu           sync.RWMutex
	data         any
	singleThread bool
}

// NewEnvironment creates a new environment, optionally nested under a parent.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		values:  make(map[string]Value),
		structs: make(map[string]*StructDefinitionValue),
		parent:  parent,
	}
}

// SetSingleThread marks the entire scope chain as single-threaded,
// allowing lock-free access. Call this at startup before any goroutines
// are spawned. Call SetMultiThread before the first spawn.
func (e *Environment) SetSingleThread() {
	for cur := e; cur != nil; cur = cur.parent {
		cur.singleThread = true
	}
}

// SetMultiThread reverts the entire scope chain to locked access.
func (e *Environment) SetMultiThread() {
	for cur := e; cur != nil; cur = cur.parent {
		cur.singleThread = false
	}
}

// Parent exposes the lexical parent (nil when global).
func (e *Environment) Parent() *Environment {
	if e.singleThread {
		return e.parent
	}
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

// StructSnapshot returns a deterministic copy of the current struct bindings.
func (e *Environment) StructSnapshot() map[string]*StructDefinitionValue {
	e.mu.RLock()
	out := make(map[string]*StructDefinitionValue, len(e.structs))
	for k, v := range e.structs {
		out[k] = v
	}
	e.mu.RUnlock()
	return out
}

// Define inserts or shadows a binding in the current scope.
func (e *Environment) Define(name string, value Value) {
	e.mu.Lock()
	if existing, ok := e.values[name]; ok {
		if merged, ok := MergeFunctionValues(existing, value); ok {
			e.values[name] = merged
			e.mu.Unlock()
			return
		}
	}
	e.values[name] = value
	e.mu.Unlock()
}

// DefineStruct records a struct definition in the current scope.
func (e *Environment) DefineStruct(name string, def *StructDefinitionValue) {
	if def == nil {
		return
	}
	e.mu.Lock()
	e.structs[name] = def
	e.mu.Unlock()
}

// StructDefinition retrieves a struct definition, searching outward through the scope chain.
func (e *Environment) StructDefinition(name string) (*StructDefinitionValue, bool) {
	for cur := e; cur != nil; {
		if cur.singleThread {
			if v, ok := cur.structs[name]; ok {
				return v, true
			}
			cur = cur.parent
		} else {
			cur.mu.RLock()
			v, ok := cur.structs[name]
			parent := cur.parent
			cur.mu.RUnlock()
			if ok {
				return v, true
			}
			cur = parent
		}
	}
	return nil, false
}

// Assign updates an existing binding in the first scope where it appears.
func (e *Environment) Assign(name string, value Value) error {
	for cur := e; cur != nil; {
		if cur.singleThread {
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				return nil
			}
			cur = cur.parent
		} else {
			cur.mu.Lock()
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				cur.mu.Unlock()
				return nil
			}
			parent := cur.parent
			cur.mu.Unlock()
			cur = parent
		}
	}
	return fmt.Errorf("Undefined variable '%s'", name)
}

// Get retrieves a binding, searching outward through the scope chain.
func (e *Environment) Get(name string) (Value, error) {
	for cur := e; cur != nil; {
		if cur.singleThread {
			if v, ok := cur.values[name]; ok {
				return v, nil
			}
			cur = cur.parent
		} else {
			cur.mu.RLock()
			v, ok := cur.values[name]
			parent := cur.parent
			cur.mu.RUnlock()
			if ok {
				return v, nil
			}
			cur = parent
		}
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
	for cur := e; cur != nil; {
		if cur.singleThread {
			if cur.data != nil {
				return cur.data
			}
			cur = cur.parent
		} else {
			cur.mu.RLock()
			data := cur.data
			parent := cur.parent
			cur.mu.RUnlock()
			if data != nil {
				return data
			}
			cur = parent
		}
	}
	return nil
}

// Has reports whether the binding exists anywhere in the scope chain.
func (e *Environment) Has(name string) bool {
	for cur := e; cur != nil; {
		if cur.singleThread {
			if _, ok := cur.values[name]; ok {
				return true
			}
			cur = cur.parent
		} else {
			cur.mu.RLock()
			_, ok := cur.values[name]
			parent := cur.parent
			cur.mu.RUnlock()
			if ok {
				return true
			}
			cur = parent
		}
	}
	return false
}

// HasInCurrentScope reports whether the binding exists in the current scope.
func (e *Environment) HasInCurrentScope(name string) bool {
	e.mu.RLock()
	_, ok := e.values[name]
	e.mu.RUnlock()
	return ok
}

// AssignExisting assigns a name if it exists anywhere in the scope chain.
// Returns true when the assignment succeeded.
func (e *Environment) AssignExisting(name string, value Value) bool {
	for cur := e; cur != nil; {
		if cur.singleThread {
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				return true
			}
			cur = cur.parent
		} else {
			cur.mu.Lock()
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				cur.mu.Unlock()
				return true
			}
			parent := cur.parent
			cur.mu.Unlock()
			cur = parent
		}
	}
	return false
}
