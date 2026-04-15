package runtime

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// Environment provides lexical scoping for Able runtime values.
type Environment struct {
	values     map[string]Value
	structs    map[string]*StructDefinitionValue
	parent     *Environment
	mu         sync.RWMutex
	data       any
	threadMode *atomic.Bool
	version    uint64
}

// NewEnvironment creates a new environment, optionally nested under a parent.
func NewEnvironment(parent *Environment) *Environment {
	var mode *atomic.Bool
	if parent != nil && parent.threadMode != nil {
		mode = parent.threadMode
	} else {
		mode = &atomic.Bool{}
	}
	return &Environment{
		parent:     parent,
		threadMode: mode,
	}
}

func (e *Environment) isSingleThread() bool {
	return e != nil && e.threadMode != nil && e.threadMode.Load()
}

// SetSingleThread marks the entire scope chain as single-threaded,
// allowing lock-free access. Call this at startup before any goroutines
// are spawned. Call SetMultiThread before the first spawn.
func (e *Environment) SetSingleThread() {
	if e == nil || e.threadMode == nil {
		return
	}
	e.threadMode.Store(true)
}

// SetMultiThread reverts the entire scope chain to locked access.
func (e *Environment) SetMultiThread() {
	if e == nil || e.threadMode == nil {
		return
	}
	e.threadMode.Store(false)
}

// Parent exposes the lexical parent (nil when global).
func (e *Environment) Parent() *Environment {
	if e.isSingleThread() {
		return e.parent
	}
	e.mu.RLock()
	parent := e.parent
	e.mu.RUnlock()
	return parent
}

// Snapshot returns a deterministic copy of the current bindings.
func (e *Environment) Snapshot() map[string]Value {
	if e.isSingleThread() {
		out := make(map[string]Value, len(e.values))
		for k, v := range e.values {
			out[k] = v
		}
		return out
	}
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
	if e.isSingleThread() {
		out := make(map[string]*StructDefinitionValue, len(e.structs))
		for k, v := range e.structs {
			out[k] = v
		}
		return out
	}
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
	if e.isSingleThread() {
		if e.values == nil {
			e.values = make(map[string]Value)
		}
		if existing, ok := e.values[name]; ok {
			if merged, ok := MergeFunctionValues(existing, value); ok {
				e.values[name] = merged
				e.version++
				return
			}
		}
		e.values[name] = value
		e.version++
		return
	}
	e.mu.Lock()
	if e.values == nil {
		e.values = make(map[string]Value)
	}
	if existing, ok := e.values[name]; ok {
		if merged, ok := MergeFunctionValues(existing, value); ok {
			e.values[name] = merged
			e.version++
			e.mu.Unlock()
			return
		}
	}
	e.values[name] = value
	e.version++
	e.mu.Unlock()
}

// DefineStruct records a struct definition in the current scope.
func (e *Environment) DefineStruct(name string, def *StructDefinitionValue) {
	if def == nil {
		return
	}
	if e.isSingleThread() {
		if e.structs == nil {
			e.structs = make(map[string]*StructDefinitionValue)
		}
		e.structs[name] = def
		return
	}
	e.mu.Lock()
	if e.structs == nil {
		e.structs = make(map[string]*StructDefinitionValue)
	}
	e.structs[name] = def
	e.mu.Unlock()
}

// StructDefinition retrieves a struct definition, searching outward through the scope chain.
func (e *Environment) StructDefinition(name string) (*StructDefinitionValue, bool) {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
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
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				cur.version++
				return nil
			}
			cur = cur.parent
		} else {
			cur.mu.Lock()
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				cur.version++
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
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
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

// Lookup retrieves a binding, searching outward through the scope chain.
// It avoids constructing an error on misses and is preferred in hot paths
// where absence is expected.
func (e *Environment) Lookup(name string) (Value, bool) {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if v, ok := cur.values[name]; ok {
				return v, true
			}
			cur = cur.parent
		} else {
			cur.mu.RLock()
			v, ok := cur.values[name]
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

// LookupWithOwner retrieves a binding plus the lexical scope that currently
// owns it. It avoids constructing an error on misses and is preferred in hot
// paths that want to cache parent/global hits without rewalking the chain.
func (e *Environment) LookupWithOwner(name string) (Value, *Environment, bool) {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if v, ok := cur.values[name]; ok {
				return v, cur, true
			}
			cur = cur.parent
		} else {
			cur.mu.RLock()
			v, ok := cur.values[name]
			parent := cur.parent
			cur.mu.RUnlock()
			if ok {
				return v, cur, true
			}
			cur = parent
		}
	}
	return nil, nil, false
}

// LookupInCurrentScope retrieves a binding only from the current scope.
// It avoids constructing an error on misses and does not walk lexical parents.
func (e *Environment) LookupInCurrentScope(name string) (Value, bool) {
	if e.isSingleThread() {
		v, ok := e.values[name]
		return v, ok
	}
	e.mu.RLock()
	v, ok := e.values[name]
	e.mu.RUnlock()
	return v, ok
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
	if e.isSingleThread() {
		e.data = data
		return
	}
	e.mu.Lock()
	e.data = data
	e.mu.Unlock()
}

// RuntimeData returns the metadata associated with this environment, falling back to parents.
func (e *Environment) RuntimeData() any {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
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
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
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
	if e.isSingleThread() {
		_, ok := e.values[name]
		return ok
	}
	e.mu.RLock()
	_, ok := e.values[name]
	e.mu.RUnlock()
	return ok
}

// AssignExisting assigns a name if it exists anywhere in the scope chain.
// Returns true when the assignment succeeded.
func (e *Environment) AssignExisting(name string, value Value) bool {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				cur.version++
				return true
			}
			cur = cur.parent
		} else {
			cur.mu.Lock()
			if _, ok := cur.values[name]; ok {
				cur.values[name] = value
				cur.version++
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

// Revision returns the mutation revision for this scope.
func (e *Environment) Revision() uint64 {
	if e.isSingleThread() {
		return e.version
	}
	e.mu.RLock()
	version := e.version
	e.mu.RUnlock()
	return version
}

// RevisionWithHint returns the mutation revision for this scope while letting
// the caller supply the already-known thread mode. This avoids repeating the
// shared thread-mode load on hot paths that already know execution is
// single-threaded.
func (e *Environment) RevisionWithHint(singleThread bool) uint64 {
	if e == nil {
		return 0
	}
	if singleThread {
		return e.version
	}
	e.mu.RLock()
	version := e.version
	e.mu.RUnlock()
	return version
}
