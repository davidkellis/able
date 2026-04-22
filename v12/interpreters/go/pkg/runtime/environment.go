package runtime

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// Environment provides lexical scoping for Able runtime values.
type Environment struct {
	values      map[string]Value
	singleName  string
	singleValue Value
	hasSingle   bool
	parent      *Environment
	mu          atomic.Pointer[sync.RWMutex]
	meta        *environmentMeta
	threadMode  *atomic.Bool
	version     uint64
}

type environmentMeta struct {
	structs map[string]*StructDefinitionValue
	data    any
}

// NewEnvironment creates a new environment, optionally nested under a parent.
func NewEnvironment(parent *Environment) *Environment {
	return NewEnvironmentWithValueCapacity(parent, 0)
}

// NewEnvironmentWithValueCapacity creates a new environment with an optional
// pre-sized value map for callers that know they will bind a fixed number of
// locals immediately.
func NewEnvironmentWithValueCapacity(parent *Environment, valueCapacity int) *Environment {
	var mode *atomic.Bool
	if parent != nil && parent.threadMode != nil {
		mode = parent.threadMode
	} else {
		mode = &atomic.Bool{}
	}
	var values map[string]Value
	if valueCapacity > 1 {
		values = make(map[string]Value, valueCapacity)
	}
	return &Environment{
		values:     values,
		parent:     parent,
		threadMode: mode,
	}
}

func (e *Environment) mutex() *sync.RWMutex {
	if e == nil {
		return nil
	}
	if mu := e.mu.Load(); mu != nil {
		return mu
	}
	mu := &sync.RWMutex{}
	if e.mu.CompareAndSwap(nil, mu) {
		return mu
	}
	return e.mu.Load()
}

func (e *Environment) currentValueCountNoLock() int {
	if e.values != nil {
		return len(e.values)
	}
	if e.hasSingle {
		return 1
	}
	return 0
}

func (e *Environment) lookupCurrentValueNoLock(name string) (Value, bool) {
	if e.values != nil {
		v, ok := e.values[name]
		return v, ok
	}
	if e.hasSingle && e.singleName == name {
		return e.singleValue, true
	}
	return nil, false
}

func (e *Environment) promoteSingleBindingNoLock(minCapacity int) {
	if e.values != nil {
		return
	}
	if minCapacity < 2 {
		minCapacity = 2
	}
	e.values = make(map[string]Value, minCapacity)
	if !e.hasSingle {
		return
	}
	e.values[e.singleName] = e.singleValue
	e.singleName = ""
	e.singleValue = nil
	e.hasSingle = false
}

func (e *Environment) setCurrentValueNoLock(name string, value Value) {
	if e.values != nil {
		e.values[name] = value
		return
	}
	if e.hasSingle {
		if e.singleName == name {
			e.singleValue = value
			return
		}
		e.promoteSingleBindingNoLock(2)
		e.values[name] = value
		return
	}
	e.singleName = name
	e.singleValue = value
	e.hasSingle = true
}

func (e *Environment) isSingleThread() bool {
	return e != nil && e.threadMode != nil && e.threadMode.Load()
}

func (e *Environment) ensureMetaNoLock() *environmentMeta {
	if e.meta == nil {
		e.meta = &environmentMeta{}
	}
	return e.meta
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
	mu := e.mutex()
	mu.RLock()
	parent := e.parent
	mu.RUnlock()
	return parent
}

// Snapshot returns a deterministic copy of the current bindings.
func (e *Environment) Snapshot() map[string]Value {
	if e.isSingleThread() {
		out := make(map[string]Value, e.currentValueCountNoLock())
		for k, v := range e.values {
			out[k] = v
		}
		if e.hasSingle {
			out[e.singleName] = e.singleValue
		}
		return out
	}
	mu := e.mutex()
	mu.RLock()
	out := make(map[string]Value, e.currentValueCountNoLock())
	for k, v := range e.values {
		out[k] = v
	}
	if e.hasSingle {
		out[e.singleName] = e.singleValue
	}
	mu.RUnlock()
	return out
}

// StructSnapshot returns a deterministic copy of the current struct bindings.
func (e *Environment) StructSnapshot() map[string]*StructDefinitionValue {
	var structs map[string]*StructDefinitionValue
	if e.isSingleThread() {
		if e.meta != nil {
			structs = e.meta.structs
		}
		out := make(map[string]*StructDefinitionValue, len(structs))
		for k, v := range structs {
			out[k] = v
		}
		return out
	}
	mu := e.mutex()
	mu.RLock()
	if e.meta != nil {
		structs = e.meta.structs
	}
	out := make(map[string]*StructDefinitionValue, len(structs))
	for k, v := range structs {
		out[k] = v
	}
	mu.RUnlock()
	return out
}

// Define inserts or shadows a binding in the current scope.
func (e *Environment) Define(name string, value Value) {
	if e.isSingleThread() {
		if existing, ok := e.lookupCurrentValueNoLock(name); ok {
			if merged, ok := MergeFunctionValues(existing, value); ok {
				e.setCurrentValueNoLock(name, merged)
				e.version++
				return
			}
		}
		e.setCurrentValueNoLock(name, value)
		e.version++
		return
	}
	mu := e.mutex()
	mu.Lock()
	if existing, ok := e.lookupCurrentValueNoLock(name); ok {
		if merged, ok := MergeFunctionValues(existing, value); ok {
			e.setCurrentValueNoLock(name, merged)
			e.version++
			mu.Unlock()
			return
		}
	}
	e.setCurrentValueNoLock(name, value)
	e.version++
	mu.Unlock()
}

// DefineWithoutMerge inserts or shadows a binding in the current scope without
// checking function-merge semantics. This is appropriate for plain local
// bindings such as pattern matches where merge behavior is never desired.
func (e *Environment) DefineWithoutMerge(name string, value Value) {
	if e.isSingleThread() {
		e.setCurrentValueNoLock(name, value)
		e.version++
		return
	}
	mu := e.mutex()
	mu.Lock()
	e.setCurrentValueNoLock(name, value)
	e.version++
	mu.Unlock()
}

// DefineStruct records a struct definition in the current scope.
func (e *Environment) DefineStruct(name string, def *StructDefinitionValue) {
	if def == nil {
		return
	}
	if e.isSingleThread() {
		meta := e.ensureMetaNoLock()
		if meta.structs == nil {
			meta.structs = make(map[string]*StructDefinitionValue)
		}
		meta.structs[name] = def
		return
	}
	mu := e.mutex()
	mu.Lock()
	meta := e.ensureMetaNoLock()
	if meta.structs == nil {
		meta.structs = make(map[string]*StructDefinitionValue)
	}
	meta.structs[name] = def
	mu.Unlock()
}

// StructDefinition retrieves a struct definition, searching outward through the scope chain.
func (e *Environment) StructDefinition(name string) (*StructDefinitionValue, bool) {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if cur.meta != nil {
				if v, ok := cur.meta.structs[name]; ok {
					return v, true
				}
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			var (
				v      *StructDefinitionValue
				ok     bool
				parent *Environment
			)
			if cur.meta != nil {
				v, ok = cur.meta.structs[name]
			}
			parent = cur.parent
			mu.RUnlock()
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
			if _, ok := cur.lookupCurrentValueNoLock(name); ok {
				cur.setCurrentValueNoLock(name, value)
				cur.version++
				return nil
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.Lock()
			if _, ok := cur.lookupCurrentValueNoLock(name); ok {
				cur.setCurrentValueNoLock(name, value)
				cur.version++
				mu.Unlock()
				return nil
			}
			parent := cur.parent
			mu.Unlock()
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
			if v, ok := cur.lookupCurrentValueNoLock(name); ok {
				return v, nil
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			v, ok := cur.lookupCurrentValueNoLock(name)
			parent := cur.parent
			mu.RUnlock()
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
			if v, ok := cur.lookupCurrentValueNoLock(name); ok {
				return v, true
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			v, ok := cur.lookupCurrentValueNoLock(name)
			parent := cur.parent
			mu.RUnlock()
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
			if v, ok := cur.lookupCurrentValueNoLock(name); ok {
				return v, cur, true
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			v, ok := cur.lookupCurrentValueNoLock(name)
			parent := cur.parent
			mu.RUnlock()
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
		return e.lookupCurrentValueNoLock(name)
	}
	mu := e.mutex()
	mu.RLock()
	v, ok := e.lookupCurrentValueNoLock(name)
	mu.RUnlock()
	return v, ok
}

// Keys returns the bindings in sorted order (useful for determinism in tests).
func (e *Environment) Keys() []string {
	if e.isSingleThread() {
		keys := make([]string, 0, e.currentValueCountNoLock())
		for k := range e.values {
			keys = append(keys, k)
		}
		if e.hasSingle {
			keys = append(keys, e.singleName)
		}
		sort.Strings(keys)
		return keys
	}
	mu := e.mutex()
	mu.RLock()
	keys := make([]string, 0, e.currentValueCountNoLock())
	for k := range e.values {
		keys = append(keys, k)
	}
	if e.hasSingle {
		keys = append(keys, e.singleName)
	}
	mu.RUnlock()
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
		e.ensureMetaNoLock().data = data
		return
	}
	mu := e.mutex()
	mu.Lock()
	e.ensureMetaNoLock().data = data
	mu.Unlock()
}

// RuntimeData returns the metadata associated with this environment, falling back to parents.
func (e *Environment) RuntimeData() any {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if cur.meta != nil && cur.meta.data != nil {
				return cur.meta.data
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			var data any
			if cur.meta != nil {
				data = cur.meta.data
			}
			parent := cur.parent
			mu.RUnlock()
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
			if _, ok := cur.lookupCurrentValueNoLock(name); ok {
				return true
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.RLock()
			_, ok := cur.lookupCurrentValueNoLock(name)
			parent := cur.parent
			mu.RUnlock()
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
		_, ok := e.lookupCurrentValueNoLock(name)
		return ok
	}
	mu := e.mutex()
	mu.RLock()
	_, ok := e.lookupCurrentValueNoLock(name)
	mu.RUnlock()
	return ok
}

// AssignExisting assigns a name if it exists anywhere in the scope chain.
// Returns true when the assignment succeeded.
func (e *Environment) AssignExisting(name string, value Value) bool {
	singleThread := e.isSingleThread()
	for cur := e; cur != nil; {
		if singleThread {
			if _, ok := cur.lookupCurrentValueNoLock(name); ok {
				cur.setCurrentValueNoLock(name, value)
				cur.version++
				return true
			}
			cur = cur.parent
		} else {
			mu := cur.mutex()
			mu.Lock()
			if _, ok := cur.lookupCurrentValueNoLock(name); ok {
				cur.setCurrentValueNoLock(name, value)
				cur.version++
				mu.Unlock()
				return true
			}
			parent := cur.parent
			mu.Unlock()
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
	mu := e.mutex()
	mu.RLock()
	version := e.version
	mu.RUnlock()
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
	mu := e.mutex()
	mu.RLock()
	version := e.version
	mu.RUnlock()
	return version
}
